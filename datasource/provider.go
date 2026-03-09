// Package datasource provides a plugin-style provider architecture for external database connections.
// It follows the omniX pattern allowing multiple database backends (PostgreSQL, MySQL, ClickHouse, etc.)
// to be used interchangeably through a common interface.
package datasource

import (
	"context"
	"time"
)

// Provider defines the interface all data source providers must implement.
// Providers are registered at init() time and used to create connections.
type Provider interface {
	// Name returns the provider name (e.g., "postgres", "mysql").
	Name() string

	// Connect establishes a connection using the given config.
	Connect(ctx context.Context, config ConnectionConfig) (Connection, error)

	// ValidateConfig checks if the config is valid before connecting.
	ValidateConfig(config ConnectionConfig) error

	// Capabilities returns what this provider supports.
	Capabilities() Capabilities
}

// Connection represents an active database connection.
type Connection interface {
	// Query executes a query and returns results.
	Query(ctx context.Context, query string, params map[string]any) (*QueryResult, error)

	// Ping tests the connection.
	Ping(ctx context.Context) error

	// Close closes the connection.
	Close() error

	// Stats returns connection pool statistics.
	Stats() ConnectionStats
}

// ConnectionConfig holds connection parameters for a data source.
type ConnectionConfig struct {
	// Host is the database server host.
	Host string `json:"host,omitempty"`

	// Port is the database server port.
	Port int `json:"port,omitempty"`

	// Database is the database name to connect to.
	Database string `json:"database,omitempty"`

	// Username is the database user.
	Username string `json:"username,omitempty"`

	// Password is the database password (prefer PasswordEnv for production).
	Password string `json:"password,omitempty"`

	// PasswordEnv is an environment variable name containing the password.
	PasswordEnv string `json:"passwordEnv,omitempty"`

	// SSLMode specifies the SSL/TLS mode (e.g., "disable", "require", "verify-full").
	SSLMode string `json:"sslMode,omitempty"`

	// ConnectionURL is a complete connection string (alternative to individual fields).
	// When provided, takes precedence over individual fields.
	ConnectionURL string `json:"connectionUrl,omitempty"`

	// ConnectionURLEnv is an environment variable name containing the connection URL.
	ConnectionURLEnv string `json:"connectionUrlEnv,omitempty"`

	// MaxConnections is the maximum number of connections in the pool.
	MaxConnections int `json:"maxConnections,omitempty"`

	// MaxIdleConnections is the maximum number of idle connections in the pool.
	MaxIdleConnections int `json:"maxIdleConnections,omitempty"`

	// ConnectionMaxLifetime is the maximum lifetime of a connection.
	ConnectionMaxLifetime time.Duration `json:"connectionMaxLifetime,omitempty"`

	// QueryTimeout is the default timeout for queries.
	QueryTimeout time.Duration `json:"queryTimeout,omitempty"`

	// ReadOnly restricts the connection to read-only queries (SELECT only).
	ReadOnly bool `json:"readOnly,omitempty"`

	// Extra contains provider-specific options.
	Extra map[string]any `json:"extra,omitempty"`
}

// ParameterStyle describes how query parameters are specified.
type ParameterStyle string

const (
	// ParameterStylePositionalDollar uses $1, $2, $3 (PostgreSQL).
	ParameterStylePositionalDollar ParameterStyle = "positional_dollar"

	// ParameterStylePositionalQuestion uses ?, ?, ? (MySQL, SQLite).
	ParameterStylePositionalQuestion ParameterStyle = "positional_question"

	// ParameterStyleNamed uses :name, :value (Oracle, some drivers).
	ParameterStyleNamed ParameterStyle = "named"

	// ParameterStyleAtNamed uses @name, @value (SQL Server).
	ParameterStyleAtNamed ParameterStyle = "at_named"
)

// Capabilities describes what a provider supports.
type Capabilities struct {
	// SupportsTransactions indicates if the provider supports transactions.
	SupportsTransactions bool `json:"supportsTransactions"`

	// SupportsStreaming indicates if the provider supports streaming results.
	SupportsStreaming bool `json:"supportsStreaming"`

	// SupportsPreparedStmts indicates if the provider supports prepared statements.
	SupportsPreparedStmts bool `json:"supportsPreparedStmts"`

	// SupportsNamedParams indicates if the provider supports named parameters.
	SupportsNamedParams bool `json:"supportsNamedParams"`

	// MaxQuerySize is the maximum query size in bytes (0 = no limit).
	MaxQuerySize int64 `json:"maxQuerySize"`

	// ParameterStyle specifies how query parameters are formatted.
	ParameterStyle ParameterStyle `json:"parameterStyle"`
}

// ColumnInfo describes a column in query results.
type ColumnInfo struct {
	// Name is the column name.
	Name string `json:"name"`

	// Type is the database type name.
	Type string `json:"type"`

	// Nullable indicates if the column can be NULL.
	Nullable bool `json:"nullable"`

	// Length is the column length for string types (0 = unknown).
	Length int64 `json:"length,omitempty"`

	// Precision is the numeric precision for decimal types.
	Precision int64 `json:"precision,omitempty"`

	// Scale is the numeric scale for decimal types.
	Scale int64 `json:"scale,omitempty"`
}

// QueryResult holds query results.
type QueryResult struct {
	// Columns contains column metadata.
	Columns []ColumnInfo `json:"columns"`

	// Rows contains the data rows as maps.
	Rows []map[string]any `json:"rows"`

	// RowCount is the number of rows returned.
	RowCount int `json:"rowCount"`

	// AffectedRows is the number of rows affected (for INSERT/UPDATE/DELETE).
	AffectedRows int64 `json:"affectedRows,omitempty"`

	// ExecutionTimeMs is how long the query took in milliseconds.
	ExecutionTimeMs int64 `json:"executionTimeMs"`
}

// ConnectionStats contains connection pool statistics.
type ConnectionStats struct {
	// MaxOpen is the maximum number of open connections.
	MaxOpen int `json:"maxOpen"`

	// Open is the current number of open connections.
	Open int `json:"open"`

	// InUse is the number of connections currently in use.
	InUse int `json:"inUse"`

	// Idle is the number of idle connections.
	Idle int `json:"idle"`

	// WaitCount is the total number of connections waited for.
	WaitCount int64 `json:"waitCount"`

	// WaitDuration is the total time blocked waiting for a connection.
	WaitDuration time.Duration `json:"waitDuration"`

	// MaxIdleClosed is the total connections closed due to max idle.
	MaxIdleClosed int64 `json:"maxIdleClosed"`

	// MaxLifetimeClosed is the total connections closed due to max lifetime.
	MaxLifetimeClosed int64 `json:"maxLifetimeClosed"`
}

// TableInfo describes a database table.
type TableInfo struct {
	// Schema is the schema/namespace containing the table.
	Schema string `json:"schema"`

	// Name is the table name.
	Name string `json:"name"`

	// Type is "table" or "view".
	Type string `json:"type"`

	// Columns contains column information.
	Columns []ColumnInfo `json:"columns,omitempty"`

	// RowCount is an estimated row count (may be approximate).
	RowCount int64 `json:"rowCount,omitempty"`
}

// SchemaInfo describes the database schema.
type SchemaInfo struct {
	// Tables contains all tables and views.
	Tables []TableInfo `json:"tables"`
}
