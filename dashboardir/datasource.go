package dashboardir

import "encoding/json"

// DataSource defines where dashboard data comes from.
// Supports static files (URL), inline data, and future database connections.
type DataSource struct {
	// ID is the unique identifier for this data source.
	ID string `json:"id"`

	// Name is an optional display name.
	Name string `json:"name,omitempty"`

	// Type is the data source type: "url", "inline", "postgres", "derived".
	Type string `json:"type"`

	// URL is the data location (for url type).
	// Supports relative paths for GitHub Pages: "./data/results.json"
	URL string `json:"url,omitempty"`

	// Format is the data format: "json", "csv", "ndjson".
	Format string `json:"format,omitempty"`

	// Data is inline data (for inline type).
	// Use json.RawMessage to preserve structure.
	Data json.RawMessage `json:"data,omitempty"`

	// Connection is database connection config (for postgres type).
	// Future: will contain connection string or reference.
	Connection *ConnectionConfig `json:"connection,omitempty"`

	// Query is the data query (for postgres type).
	Query string `json:"query,omitempty"`

	// DerivedFrom references another data source (for derived type).
	DerivedFrom string `json:"derivedFrom,omitempty"`

	// Transform applies transformations to the data.
	Transform []Transform `json:"transform,omitempty"`

	// Refresh configures automatic data refresh.
	Refresh *RefreshConfig `json:"refresh,omitempty"`

	// Cache configures data caching.
	Cache *CacheConfig `json:"cache,omitempty"`
}

// ConnectionConfig defines database connection settings.
// Placeholder for future PostgreSQL support.
type ConnectionConfig struct {
	// Type is the database type: "postgres", "mysql", "sqlite".
	Type string `json:"type"`

	// Host is the database host.
	Host string `json:"host,omitempty"`

	// Port is the database port.
	Port int `json:"port,omitempty"`

	// Database is the database name.
	Database string `json:"database,omitempty"`

	// Username is the database user.
	Username string `json:"username,omitempty"`

	// PasswordEnv is the environment variable containing the password.
	// Never store passwords directly in the IR.
	PasswordEnv string `json:"passwordEnv,omitempty"`

	// SSLMode is the SSL connection mode.
	SSLMode string `json:"sslMode,omitempty"`

	// ConnectionString is an alternative to individual fields.
	// Can reference environment variable: "${DATABASE_URL}"
	ConnectionString string `json:"connectionString,omitempty"`
}

// RefreshConfig defines automatic data refresh settings.
type RefreshConfig struct {
	// Enabled toggles automatic refresh.
	Enabled bool `json:"enabled"`

	// IntervalSeconds is the refresh interval.
	IntervalSeconds int `json:"intervalSeconds,omitempty"`

	// OnFocus refreshes when the browser tab gains focus.
	OnFocus bool `json:"onFocus,omitempty"`
}

// CacheConfig defines data caching settings.
type CacheConfig struct {
	// Enabled toggles caching.
	Enabled bool `json:"enabled"`

	// TTLSeconds is the cache time-to-live.
	TTLSeconds int `json:"ttlSeconds,omitempty"`

	// Key is a custom cache key (defaults to data source ID).
	Key string `json:"key,omitempty"`
}

// DataSource type constants.
const (
	DataSourceTypeURL      = "url"
	DataSourceTypeInline   = "inline"
	DataSourceTypePostgres = "postgres"
	DataSourceTypeDerived  = "derived"
)

// Data format constants.
const (
	DataFormatJSON   = "json"
	DataFormatCSV    = "csv"
	DataFormatNDJSON = "ndjson"
)

// Database type constants.
const (
	DatabaseTypePostgres = "postgres"
	DatabaseTypeMySQL    = "mysql"
	DatabaseTypeSQLite   = "sqlite"
)
