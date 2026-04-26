package datasource

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"
)

// QueryExecutor executes queries against data sources with additional features
// like timeout handling, result transformations, and logging.
type QueryExecutor struct {
	manager *Manager
	logger  *slog.Logger

	// DefaultTimeout is the default query timeout if not specified.
	DefaultTimeout time.Duration

	// MaxRowsDefault is the default maximum rows to return.
	MaxRowsDefault int

	// MaxRowsLimit is the absolute maximum rows allowed.
	MaxRowsLimit int
}

// QueryExecutorConfig configures the QueryExecutor.
type QueryExecutorConfig struct {
	Manager        *Manager
	Logger         *slog.Logger
	DefaultTimeout time.Duration
	MaxRowsDefault int
	MaxRowsLimit   int
}

// NewQueryExecutor creates a new QueryExecutor.
func NewQueryExecutor(cfg QueryExecutorConfig) *QueryExecutor {
	if cfg.DefaultTimeout == 0 {
		cfg.DefaultTimeout = 30 * time.Second
	}
	if cfg.MaxRowsDefault == 0 {
		cfg.MaxRowsDefault = 1000
	}
	if cfg.MaxRowsLimit == 0 {
		cfg.MaxRowsLimit = 10000
	}

	return &QueryExecutor{
		manager:        cfg.Manager,
		logger:         cfg.Logger,
		DefaultTimeout: cfg.DefaultTimeout,
		MaxRowsDefault: cfg.MaxRowsDefault,
		MaxRowsLimit:   cfg.MaxRowsLimit,
	}
}

// QueryRequest represents a query execution request.
type QueryRequest struct {
	// DataSource is the data source configuration.
	DataSource DataSourceConfig

	// Query is the SQL query to execute.
	Query string

	// Parameters are the query parameters.
	Parameters map[string]any

	// Timeout overrides the default query timeout.
	Timeout time.Duration

	// MaxRows limits the number of rows returned.
	MaxRows int
}

// Execute executes a query against a data source.
func (e *QueryExecutor) Execute(ctx context.Context, req QueryRequest) (*QueryResult, error) {
	start := time.Now()

	// Apply timeout
	timeout := e.DefaultTimeout
	if req.Timeout > 0 {
		timeout = req.Timeout
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// Get connection
	conn, err := e.manager.GetConnection(ctx, req.DataSource)
	if err != nil {
		return nil, fmt.Errorf("getting connection: %w", err)
	}

	// Apply row limit
	query := req.Query
	maxRows := e.MaxRowsDefault
	if req.MaxRows > 0 {
		maxRows = req.MaxRows
	}
	if maxRows > e.MaxRowsLimit {
		maxRows = e.MaxRowsLimit
	}
	query = applyRowLimit(query, maxRows)

	// Execute query
	result, err := conn.Query(ctx, query, req.Parameters)
	if err != nil {
		e.logQuery(req, time.Since(start), err)
		return nil, err
	}

	e.logQuery(req, time.Since(start), nil)

	return result, nil
}

// ExecuteRaw executes a query without any transformations.
func (e *QueryExecutor) ExecuteRaw(ctx context.Context, req QueryRequest) (*QueryResult, error) {
	// Apply timeout
	timeout := e.DefaultTimeout
	if req.Timeout > 0 {
		timeout = req.Timeout
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// Get connection
	conn, err := e.manager.GetConnection(ctx, req.DataSource)
	if err != nil {
		return nil, fmt.Errorf("getting connection: %w", err)
	}

	// Execute query as-is
	return conn.Query(ctx, req.Query, req.Parameters)
}

// logQuery logs query execution details.
func (e *QueryExecutor) logQuery(req QueryRequest, duration time.Duration, err error) {
	if e.logger == nil {
		return
	}

	// Truncate query for logging
	query := req.Query
	if len(query) > 200 {
		query = query[:200] + "..."
	}

	if err != nil {
		e.logger.Error("query failed",
			"datasource_id", req.DataSource.ID,
			"datasource_name", req.DataSource.Name,
			"query", query,
			"duration_ms", duration.Milliseconds(),
			"error", err,
		)
	} else {
		e.logger.Info("query executed",
			"datasource_id", req.DataSource.ID,
			"datasource_name", req.DataSource.Name,
			"query", query,
			"duration_ms", duration.Milliseconds(),
		)
	}
}

// applyRowLimit adds a LIMIT clause to the query if not present.
func applyRowLimit(query string, maxRows int) string {
	// Simple check - doesn't handle all cases but works for most
	upperQuery := strings.ToUpper(query)
	if strings.Contains(upperQuery, " LIMIT ") {
		return query
	}

	// Don't add LIMIT to non-SELECT queries
	trimmed := strings.TrimSpace(upperQuery)
	if !strings.HasPrefix(trimmed, "SELECT") {
		return query
	}

	return fmt.Sprintf("%s LIMIT %d", strings.TrimSuffix(query, ";"), maxRows)
}

// SchemaRequest represents a request to get database schema.
type SchemaRequest struct {
	// DataSource is the data source configuration.
	DataSource DataSourceConfig

	// Schema filters to a specific schema (optional).
	Schema string

	// IncludeColumns includes column information.
	IncludeColumns bool

	// TableFilter filters tables by name pattern (optional).
	TableFilter string
}

// GetSchema retrieves database schema information.
func (e *QueryExecutor) GetSchema(ctx context.Context, req SchemaRequest) (*SchemaInfo, error) {
	conn, err := e.manager.GetConnection(ctx, req.DataSource)
	if err != nil {
		return nil, fmt.Errorf("getting connection: %w", err)
	}

	// Get the provider to determine the schema query
	provider := e.manager.getProvider(req.DataSource.Type)
	if provider == nil {
		return nil, fmt.Errorf("%w: %s", ErrProviderNotFound, req.DataSource.Type)
	}

	// Query depends on database type
	var tables []TableInfo

	switch req.DataSource.Type {
	case "postgres":
		tables, err = e.getPostgresSchema(ctx, conn, req)
	case "mysql":
		tables, err = e.getMySQLSchema(ctx, conn, req)
	default:
		return nil, fmt.Errorf("schema introspection not supported for %s", req.DataSource.Type)
	}

	if err != nil {
		return nil, err
	}

	return &SchemaInfo{Tables: tables}, nil
}

// getPostgresSchema retrieves schema for PostgreSQL.
func (e *QueryExecutor) getPostgresSchema(ctx context.Context, conn Connection, req SchemaRequest) ([]TableInfo, error) {
	schema := req.Schema
	if schema == "" {
		schema = "public"
	}

	// Query tables
	query := `
		SELECT table_schema, table_name, table_type
		FROM information_schema.tables
		WHERE table_schema = $1
		ORDER BY table_name
	`
	params := map[string]any{"schema": schema}

	// Apply table filter if specified
	if req.TableFilter != "" {
		query = `
			SELECT table_schema, table_name, table_type
			FROM information_schema.tables
			WHERE table_schema = $1 AND table_name LIKE $2
			ORDER BY table_name
		`
		params["filter"] = req.TableFilter
	}

	// Use raw query to avoid LIMIT injection
	result, err := conn.Query(ctx, query, params)
	if err != nil {
		return nil, fmt.Errorf("querying tables: %w", err)
	}

	tables := make([]TableInfo, 0, len(result.Rows))
	for _, row := range result.Rows {
		tableType := "table"
		if t, ok := row["table_type"].(string); ok && t == "VIEW" {
			tableType = "view"
		}

		table := TableInfo{
			Schema: fmt.Sprintf("%v", row["table_schema"]),
			Name:   fmt.Sprintf("%v", row["table_name"]),
			Type:   tableType,
		}

		// Get columns if requested
		if req.IncludeColumns {
			cols, err := e.getPostgresColumns(ctx, conn, table.Schema, table.Name)
			if err != nil {
				return nil, err
			}
			table.Columns = cols
		}

		tables = append(tables, table)
	}

	return tables, nil
}

// getPostgresColumns retrieves column info for a PostgreSQL table.
func (e *QueryExecutor) getPostgresColumns(ctx context.Context, conn Connection, schema, table string) ([]ColumnInfo, error) {
	query := `
		SELECT column_name, data_type, is_nullable, character_maximum_length,
		       numeric_precision, numeric_scale
		FROM information_schema.columns
		WHERE table_schema = $1 AND table_name = $2
		ORDER BY ordinal_position
	`
	return e.getColumnsWithQuery(ctx, conn, query, schema, table)
}

// getColumnsWithQuery executes a column info query and parses the results.
func (e *QueryExecutor) getColumnsWithQuery(ctx context.Context, conn Connection, query, schema, table string) ([]ColumnInfo, error) {
	params := map[string]any{"schema": schema, "table": table}

	result, err := conn.Query(ctx, query, params)
	if err != nil {
		return nil, fmt.Errorf("querying columns: %w", err)
	}

	columns := make([]ColumnInfo, 0, len(result.Rows))
	for _, row := range result.Rows {
		col := ColumnInfo{
			Name:     fmt.Sprintf("%v", row["column_name"]),
			Type:     fmt.Sprintf("%v", row["data_type"]),
			Nullable: row["is_nullable"] == "YES",
		}

		if length, ok := row["character_maximum_length"].(int64); ok {
			col.Length = length
		}
		if precision, ok := row["numeric_precision"].(int64); ok {
			col.Precision = precision
		}
		if scale, ok := row["numeric_scale"].(int64); ok {
			col.Scale = scale
		}

		columns = append(columns, col)
	}

	return columns, nil
}

// getMySQLSchema retrieves schema for MySQL.
func (e *QueryExecutor) getMySQLSchema(ctx context.Context, conn Connection, req SchemaRequest) ([]TableInfo, error) {
	// MySQL uses database name instead of schema
	database := req.Schema
	if database == "" {
		// Get current database
		result, err := conn.Query(ctx, "SELECT DATABASE()", nil)
		if err != nil {
			return nil, fmt.Errorf("getting current database: %w", err)
		}
		if len(result.Rows) > 0 {
			for _, v := range result.Rows[0] {
				database = fmt.Sprintf("%v", v)
				break
			}
		}
	}

	// Query tables
	query := `
		SELECT table_schema, table_name, table_type
		FROM information_schema.tables
		WHERE table_schema = ?
		ORDER BY table_name
	`
	params := map[string]any{"schema": database}

	result, err := conn.Query(ctx, query, params)
	if err != nil {
		return nil, fmt.Errorf("querying tables: %w", err)
	}

	tables := make([]TableInfo, 0, len(result.Rows))
	for _, row := range result.Rows {
		tableType := "table"
		if t, ok := row["table_type"].(string); ok && t == "VIEW" {
			tableType = "view"
		}

		table := TableInfo{
			Schema: fmt.Sprintf("%v", row["table_schema"]),
			Name:   fmt.Sprintf("%v", row["table_name"]),
			Type:   tableType,
		}

		// Get columns if requested
		if req.IncludeColumns {
			cols, err := e.getMySQLColumns(ctx, conn, table.Schema, table.Name)
			if err != nil {
				return nil, err
			}
			table.Columns = cols
		}

		tables = append(tables, table)
	}

	return tables, nil
}

// getMySQLColumns retrieves column info for a MySQL table.
func (e *QueryExecutor) getMySQLColumns(ctx context.Context, conn Connection, schema, table string) ([]ColumnInfo, error) {
	query := `
		SELECT column_name, data_type, is_nullable, character_maximum_length,
		       numeric_precision, numeric_scale
		FROM information_schema.columns
		WHERE table_schema = ? AND table_name = ?
		ORDER BY ordinal_position
	`
	return e.getColumnsWithQuery(ctx, conn, query, schema, table)
}
