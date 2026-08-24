// Package sqlsource is DashForge's generic Tier-1 SQL connector (ADR-0002):
// any Postgres, MySQL/Dolt, or SQLite database becomes a queryable analytics
// source through configuration alone, via INFORMATION_SCHEMA/PRAGMA
// introspection. It is domain-free and ships in the engine core; stock
// dashforge-server registers it. Registry name: "sql".
//
// Import for its side effect to register the connector, then create sources
// with connector "sql" and a dsnRef to the database.
package sqlsource

import (
	"context"
	"database/sql"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/plexusone/dashforge/analytics"
	"github.com/plexusone/dashforge/dashboardir"
)

// ConnectorName is the DashForge connector-registry name.
const ConnectorName = "sql"

func init() {
	analytics.RegisterConnector(ConnectorName, func(dsn string) (analytics.QueryProvider, error) {
		return New(dsn)
	})
}

// Provider is a SQL-backed analytics source. It holds an open *sql.DB and the
// dialect selected from the DSN.
type Provider struct {
	db      *sql.DB
	dialect dialect
}

// New opens a SQL database from a resolved DSN, selecting the driver from the
// DSN form. The DSN is already resolved by the engine (never a secret
// reference) by the time it reaches here.
func New(dsn string) (*Provider, error) {
	if strings.TrimSpace(dsn) == "" {
		return nil, fmt.Errorf("sql connector: DSN is required")
	}
	d, err := dialectFor(dsn)
	if err != nil {
		return nil, fmt.Errorf("sql connector: %w", err)
	}
	open := dsn
	if d.driver() == "mysql" {
		open = mysqlURLToDSN(dsn)
	}
	db, err := sql.Open(d.driver(), open)
	if err != nil {
		return nil, fmt.Errorf("sql connector: opening %s: %w", d.driver(), err)
	}
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(2)
	db.SetConnMaxLifetime(5 * time.Minute)
	return &Provider{db: db, dialect: d}, nil
}

// mysqlURLToDSN converts a mysql:// URL into a go-sql-driver DSN. A string
// already in driver-DSN form is returned unchanged. Mirrors the metadata-DB
// layer's conversion.
func mysqlURLToDSN(raw string) string {
	if !strings.HasPrefix(raw, "mysql://") {
		return raw
	}
	u, err := url.Parse(raw)
	if err != nil {
		return raw
	}
	user := ""
	if u.User != nil {
		user = u.User.Username()
		if pw, ok := u.User.Password(); ok {
			user += ":" + pw
		}
	}
	host := u.Hostname()
	if host == "" {
		host = "127.0.0.1"
	}
	port := u.Port()
	if port == "" {
		port = "3306"
	}
	dsn := fmt.Sprintf("%s@tcp(%s:%s)/%s", user, host, port, strings.TrimPrefix(u.Path, "/"))
	if u.RawQuery != "" {
		dsn += "?" + u.RawQuery
	}
	return dsn
}

// Catalog introspects the database schema into a dashboardir catalog. It emits
// a single source with the generic ID "sql"; the engine re-tags it with the
// configured source ID/name, so multiple SQL sources stay distinct.
func (p *Provider) Catalog(ctx context.Context) (dashboardir.AnalyticsCatalog, error) {
	cols, err := p.dialect.columns(ctx, p.db)
	if err != nil {
		return dashboardir.AnalyticsCatalog{}, fmt.Errorf("sql connector: introspecting schema: %w", err)
	}

	// Group columns into datasets, preserving first-seen table order.
	order := make([]string, 0)
	byTable := make(map[string][]columnInfo)
	for _, c := range cols {
		if _, ok := byTable[c.table]; !ok {
			order = append(order, c.table)
		}
		byTable[c.table] = append(byTable[c.table], c)
	}

	datasets := make([]dashboardir.AnalyticsDataset, 0, len(order))
	for _, table := range order {
		tcols := byTable[table]
		fields := make([]dashboardir.AnalyticsField, 0, len(tcols))
		for _, c := range tcols {
			ftype := p.dialect.fieldType(c.dataType)
			fields = append(fields, dashboardir.AnalyticsField{
				ID:         c.name,
				Name:       c.name,
				QueryName:  c.name,
				Type:       ftype,
				Source:     dashboardir.AnalyticsFieldSourceStandard,
				Role:       roleFor(ftype),
				Nullable:   c.nullable,
				Selectable: true,
				Filterable: true,
				Sortable:   true,
			})
		}
		datasets = append(datasets, dashboardir.AnalyticsDataset{
			ID:        table,
			Name:      table,
			QueryName: table,
			Fields:    fields,
		})
	}

	return dashboardir.AnalyticsCatalog{
		ID:   ConnectorName,
		Name: "SQL",
		Sources: []dashboardir.AnalyticsSource{{
			ID:       ConnectorName,
			Name:     "SQL",
			Type:     dashboardir.AnalyticsSourceTypeSQL,
			Datasets: datasets,
		}},
	}, nil
}

// roleFor assigns a default semantic role so chart builders have a sensible
// starting point: numbers measure, dates are time, everything else dimensions.
func roleFor(fieldType string) string {
	switch fieldType {
	case dashboardir.AnalyticsFieldTypeNumber:
		return dashboardir.AnalyticsFieldRoleMeasure
	case dashboardir.AnalyticsFieldTypeDate:
		return dashboardir.AnalyticsFieldRoleTime
	default:
		return dashboardir.AnalyticsFieldRoleDimension
	}
}

// Query executes a read-only SQL query and returns the neutral tabular result.
//
// NOTE: this is the baseline executor. RMI-DASHFORGE-009 hardens it with a
// read-only statement allowlist, LIMIT clamping, and error sanitization; for
// now it enforces only a single leading SELECT and scans the result.
func (p *Provider) Query(ctx context.Context, req dashboardir.AnalyticsQueryRequest) (dashboardir.AnalyticsQueryResult, error) {
	q := strings.TrimSpace(req.Query)
	if !strings.HasPrefix(strings.ToUpper(q), "SELECT") {
		return dashboardir.AnalyticsQueryResult{}, fmt.Errorf("sql connector: only read-only SELECT queries are supported")
	}
	start := time.Now()
	rows, err := p.db.QueryContext(ctx, q)
	if err != nil {
		return dashboardir.AnalyticsQueryResult{}, fmt.Errorf("sql connector: query failed: %w", err)
	}
	defer func() { _ = rows.Close() }()

	colNames, err := rows.Columns()
	if err != nil {
		return dashboardir.AnalyticsQueryResult{}, err
	}
	var out []map[string]any
	for rows.Next() {
		vals := make([]any, len(colNames))
		ptrs := make([]any, len(colNames))
		for i := range vals {
			ptrs[i] = &vals[i]
		}
		if err := rows.Scan(ptrs...); err != nil {
			return dashboardir.AnalyticsQueryResult{}, err
		}
		row := make(map[string]any, len(colNames))
		for i, name := range colNames {
			if b, ok := vals[i].([]byte); ok {
				row[name] = string(b)
			} else {
				row[name] = vals[i]
			}
		}
		out = append(out, row)
	}
	if err := rows.Err(); err != nil {
		return dashboardir.AnalyticsQueryResult{}, err
	}

	columns := make([]dashboardir.AnalyticsQueryColumn, len(colNames))
	for i, name := range colNames {
		columns[i] = dashboardir.AnalyticsQueryColumn{Name: name}
	}
	return dashboardir.AnalyticsQueryResult{
		Columns:       columns,
		Rows:          out,
		RowCount:      len(out),
		ExecutionTime: time.Since(start).Milliseconds(),
	}, nil
}

// Close closes the underlying database.
func (p *Provider) Close() error {
	if p == nil || p.db == nil {
		return nil
	}
	return p.db.Close()
}
