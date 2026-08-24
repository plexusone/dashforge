package sqlsource

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/plexusone/dashforge/dashboardir"
)

// columnInfo is one introspected column, grouped into datasets by table.
type columnInfo struct {
	table    string
	name     string
	dataType string
	nullable bool
}

// dialect abstracts the per-driver differences in connecting, introspecting
// the schema, and mapping native column types to dashboardir field types.
// Adding a database means adding a dialect, not touching the connector.
type dialect interface {
	// driver is the database/sql driver name to open with.
	driver() string
	// columns returns every column of every user table, ordered by table then
	// column position, so the connector can group them into datasets.
	columns(ctx context.Context, db *sql.DB) ([]columnInfo, error)
	// fieldType maps a native SQL type name to a dashboardir field type.
	fieldType(nativeType string) string
	// quoteIdent quotes an identifier for safe use in generated queries.
	quoteIdent(ident string) string
}

// dialectFor selects a dialect from a resolved DSN. It recognizes explicit
// URL schemes (postgres://, mysql://) and the common driver-DSN and file
// forms, defaulting to sqlite for file paths.
func dialectFor(dsn string) (dialect, error) {
	switch {
	case strings.HasPrefix(dsn, "postgres://"), strings.HasPrefix(dsn, "postgresql://"):
		return postgresDialect{}, nil
	case strings.HasPrefix(dsn, "mysql://"), strings.Contains(dsn, "@tcp("), strings.Contains(dsn, "@unix("):
		return mysqlDialect{}, nil
	case strings.HasPrefix(dsn, "sqlite://"), strings.HasPrefix(dsn, "file:"),
		strings.HasSuffix(dsn, ".db"), strings.HasSuffix(dsn, ".sqlite"):
		return sqliteDialect{}, nil
	default:
		return nil, fmt.Errorf("cannot determine SQL dialect from DSN (expected postgres://, mysql://, a go-sql-driver DSN, or a sqlite file/URL)")
	}
}

// ---- shared type-mapping helper ----

// mapCommonType maps a lowercased native type family to a dashboardir field
// type using substring rules shared across dialects.
func mapCommonType(nativeType string) string {
	t := strings.ToLower(strings.TrimSpace(nativeType))
	switch {
	case strings.Contains(t, "bool"):
		return dashboardir.AnalyticsFieldTypeBool
	case strings.Contains(t, "json"):
		return dashboardir.AnalyticsFieldTypeJSON
	case strings.Contains(t, "date"), strings.Contains(t, "time"):
		return dashboardir.AnalyticsFieldTypeDate
	case strings.Contains(t, "int"), strings.Contains(t, "dec"), strings.Contains(t, "num"),
		strings.Contains(t, "real"), strings.Contains(t, "floa"), strings.Contains(t, "doub"), //nolint:misspell // "doub" matches the "double" type family
		strings.Contains(t, "serial"), strings.Contains(t, "money"):
		return dashboardir.AnalyticsFieldTypeNumber
	case strings.Contains(t, "char"), strings.Contains(t, "text"), strings.Contains(t, "clob"),
		strings.Contains(t, "uuid"), strings.Contains(t, "enum"):
		return dashboardir.AnalyticsFieldTypeString
	default:
		return dashboardir.AnalyticsFieldTypeString
	}
}

// scanColumns runs the given introspection query (table, column, dataType,
// isNullable) and collects the rows. The nullable column is interpreted from
// either a boolean-ish or "YES"/"NO" value.
func scanColumns(ctx context.Context, db *sql.DB, query string, args ...any) ([]columnInfo, error) {
	rows, err := db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var out []columnInfo
	for rows.Next() {
		var table, name, dataType, nullable string
		if err := rows.Scan(&table, &name, &dataType, &nullable); err != nil {
			return nil, err
		}
		out = append(out, columnInfo{
			table:    table,
			name:     name,
			dataType: dataType,
			nullable: isNullable(nullable),
		})
	}
	return out, rows.Err()
}

func isNullable(v string) bool {
	switch strings.ToLower(strings.TrimSpace(v)) {
	case "yes", "true", "1", "t":
		return true
	default:
		return false
	}
}

// ---- MySQL / Dolt ----

type mysqlDialect struct{}

func (mysqlDialect) driver() string { return "mysql" }

func (mysqlDialect) columns(ctx context.Context, db *sql.DB) ([]columnInfo, error) {
	return scanColumns(ctx, db, `
SELECT TABLE_NAME, COLUMN_NAME, DATA_TYPE, IS_NULLABLE
FROM INFORMATION_SCHEMA.COLUMNS
WHERE TABLE_SCHEMA = DATABASE()
ORDER BY TABLE_NAME, ORDINAL_POSITION`)
}

func (mysqlDialect) fieldType(nativeType string) string { return mapCommonType(nativeType) }

func (mysqlDialect) quoteIdent(ident string) string {
	return "`" + strings.ReplaceAll(ident, "`", "``") + "`"
}

// ---- PostgreSQL ----

type postgresDialect struct{}

func (postgresDialect) driver() string { return "postgres" }

func (postgresDialect) columns(ctx context.Context, db *sql.DB) ([]columnInfo, error) {
	return scanColumns(ctx, db, `
SELECT table_name, column_name, data_type, is_nullable
FROM information_schema.columns
WHERE table_schema = current_schema()
ORDER BY table_name, ordinal_position`)
}

func (postgresDialect) fieldType(nativeType string) string { return mapCommonType(nativeType) }

func (postgresDialect) quoteIdent(ident string) string {
	return `"` + strings.ReplaceAll(ident, `"`, `""`) + `"`
}

// ---- SQLite ----

type sqliteDialect struct{}

func (sqliteDialect) driver() string { return "sqlite3" }

// columns introspects SQLite via sqlite_master + PRAGMA table_info, since
// SQLite has no information_schema.
func (sqliteDialect) columns(ctx context.Context, db *sql.DB) ([]columnInfo, error) {
	tableRows, err := db.QueryContext(ctx,
		`SELECT name FROM sqlite_master WHERE type='table' AND name NOT LIKE 'sqlite_%' ORDER BY name`)
	if err != nil {
		return nil, err
	}
	var tables []string
	for tableRows.Next() {
		var name string
		if err := tableRows.Scan(&name); err != nil {
			_ = tableRows.Close()
			return nil, err
		}
		tables = append(tables, name)
	}
	if err := tableRows.Err(); err != nil {
		_ = tableRows.Close()
		return nil, err
	}
	_ = tableRows.Close()

	var out []columnInfo
	for _, table := range tables {
		// PRAGMA table_info cannot be parameterized; the identifier comes from
		// sqlite_master, not user input, and is quoted.
		q := fmt.Sprintf("PRAGMA table_info(%s)", sqliteDialect{}.quoteIdent(table))
		colRows, err := db.QueryContext(ctx, q) //nolint:rowserrcheck // closed below
		if err != nil {
			return nil, err
		}
		for colRows.Next() {
			var cid int
			var name, dataType string
			var notnull, pk int
			var dflt sql.NullString
			if err := colRows.Scan(&cid, &name, &dataType, &notnull, &dflt, &pk); err != nil {
				_ = colRows.Close()
				return nil, err
			}
			out = append(out, columnInfo{
				table:    table,
				name:     name,
				dataType: dataType,
				nullable: notnull == 0,
			})
		}
		if err := colRows.Err(); err != nil {
			_ = colRows.Close()
			return nil, err
		}
		_ = colRows.Close()
	}
	return out, nil
}

func (sqliteDialect) fieldType(nativeType string) string { return mapCommonType(nativeType) }

func (sqliteDialect) quoteIdent(ident string) string {
	return `"` + strings.ReplaceAll(ident, `"`, `""`) + `"`
}
