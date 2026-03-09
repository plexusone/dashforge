package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/plexusone/dashforge/datasource"
)

// Connection implements datasource.Connection for PostgreSQL.
type Connection struct {
	db           *sql.DB
	config       datasource.ConnectionConfig
	queryTimeout time.Duration
	readOnly     bool
}

// Query executes a query and returns results.
func (c *Connection) Query(ctx context.Context, query string, params map[string]any) (*datasource.QueryResult, error) {
	// Check read-only mode
	if c.readOnly && isWriteQuery(query) {
		return nil, datasource.ErrReadOnlyViolation
	}

	// Apply query timeout if configured
	if c.queryTimeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, c.queryTimeout)
		defer cancel()
	}

	start := time.Now()

	// Convert named parameters to positional ($1, $2, etc.)
	processedQuery, args := convertParams(query, params)

	rows, err := c.db.QueryContext(ctx, processedQuery, args...)
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return nil, datasource.ErrQueryTimeout
		}
		return nil, datasource.NewQueryError(query, err)
	}
	defer func() {
		_ = rows.Close()
	}()

	// Get column info
	columns, err := c.getColumnInfo(rows)
	if err != nil {
		return nil, datasource.NewQueryError(query, err)
	}

	// Scan all rows
	var results []map[string]any
	for rows.Next() {
		row, err := c.scanRow(rows, columns)
		if err != nil {
			return nil, datasource.NewQueryError(query, err)
		}
		results = append(results, row)
	}

	if err := rows.Err(); err != nil {
		return nil, datasource.NewQueryError(query, err)
	}

	return &datasource.QueryResult{
		Columns:         columns,
		Rows:            results,
		RowCount:        len(results),
		ExecutionTimeMs: time.Since(start).Milliseconds(),
	}, nil
}

// Ping tests the connection.
func (c *Connection) Ping(ctx context.Context) error {
	return c.db.PingContext(ctx)
}

// Close closes the connection.
func (c *Connection) Close() error {
	return c.db.Close()
}

// Stats returns connection pool statistics.
func (c *Connection) Stats() datasource.ConnectionStats {
	stats := c.db.Stats()
	return datasource.ConnectionStats{
		MaxOpen:           stats.MaxOpenConnections,
		Open:              stats.OpenConnections,
		InUse:             stats.InUse,
		Idle:              stats.Idle,
		WaitCount:         stats.WaitCount,
		WaitDuration:      stats.WaitDuration,
		MaxIdleClosed:     stats.MaxIdleClosed,
		MaxLifetimeClosed: stats.MaxLifetimeClosed,
	}
}

// getColumnInfo extracts column metadata from rows.
func (c *Connection) getColumnInfo(rows *sql.Rows) ([]datasource.ColumnInfo, error) {
	columnNames, err := rows.Columns()
	if err != nil {
		return nil, fmt.Errorf("getting columns: %w", err)
	}

	columnTypes, err := rows.ColumnTypes()
	if err != nil {
		return nil, fmt.Errorf("getting column types: %w", err)
	}

	columns := make([]datasource.ColumnInfo, len(columnNames))
	for i, name := range columnNames {
		col := datasource.ColumnInfo{
			Name: name,
		}

		if i < len(columnTypes) {
			ct := columnTypes[i]
			col.Type = ct.DatabaseTypeName()

			nullable, ok := ct.Nullable()
			if ok {
				col.Nullable = nullable
			}

			length, ok := ct.Length()
			if ok {
				col.Length = length
			}

			precision, scale, ok := ct.DecimalSize()
			if ok {
				col.Precision = precision
				col.Scale = scale
			}
		}

		columns[i] = col
	}

	return columns, nil
}

// scanRow scans a single row into a map.
func (c *Connection) scanRow(rows *sql.Rows, columns []datasource.ColumnInfo) (map[string]any, error) {
	values := make([]any, len(columns))
	valuePtrs := make([]any, len(columns))
	for i := range values {
		valuePtrs[i] = &values[i]
	}

	if err := rows.Scan(valuePtrs...); err != nil {
		return nil, fmt.Errorf("scanning row: %w", err)
	}

	row := make(map[string]any)
	for i, col := range columns {
		val := values[i]
		// Convert []byte to string for text types
		if b, ok := val.([]byte); ok {
			row[col.Name] = string(b)
		} else {
			row[col.Name] = convertValue(val)
		}
	}

	return row, nil
}

// convertParams converts named parameters (:name or @name) to PostgreSQL positional ($1, $2).
func convertParams(query string, params map[string]any) (string, []any) {
	if len(params) == 0 {
		return query, nil
	}

	args := make([]any, 0, len(params))
	processedQuery := query
	i := 1

	for key, value := range params {
		placeholder := fmt.Sprintf("$%d", i)
		// Replace both :name and @name style parameters
		processedQuery = strings.ReplaceAll(processedQuery, ":"+key, placeholder)
		processedQuery = strings.ReplaceAll(processedQuery, "@"+key, placeholder)
		args = append(args, value)
		i++
	}

	return processedQuery, args
}

// convertValue converts database values to JSON-safe types.
func convertValue(v any) any {
	if v == nil {
		return nil
	}

	switch val := v.(type) {
	case time.Time:
		return val.Format(time.RFC3339)
	case []byte:
		return string(val)
	case int64, int32, int16, int8, int:
		return val
	case uint64, uint32, uint16, uint8, uint:
		return val
	case float64, float32:
		return val
	case bool:
		return val
	case string:
		return val
	default:
		// For complex types, try to use reflect
		rv := reflect.ValueOf(v)
		switch rv.Kind() {
		case reflect.Slice, reflect.Array:
			// PostgreSQL arrays
			result := make([]any, rv.Len())
			for i := 0; i < rv.Len(); i++ {
				result[i] = convertValue(rv.Index(i).Interface())
			}
			return result
		case reflect.Map:
			// JSONB/hstore
			result := make(map[string]any)
			iter := rv.MapRange()
			for iter.Next() {
				key := fmt.Sprintf("%v", iter.Key().Interface())
				result[key] = convertValue(iter.Value().Interface())
			}
			return result
		default:
			// Fall back to string representation
			return fmt.Sprintf("%v", v)
		}
	}
}
