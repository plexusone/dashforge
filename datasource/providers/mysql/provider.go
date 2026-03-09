// Package mysql provides a MySQL implementation of the datasource.Provider interface.
package mysql

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql" // MySQL driver

	"github.com/plexusone/dashforge/datasource"
)

func init() {
	datasource.Register(&Provider{})
}

// Provider implements datasource.Provider for MySQL.
type Provider struct{}

// Name returns "mysql".
func (p *Provider) Name() string {
	return "mysql"
}

// Connect establishes a MySQL connection.
func (p *Provider) Connect(ctx context.Context, config datasource.ConnectionConfig) (datasource.Connection, error) {
	dsn, err := buildDSN(config)
	if err != nil {
		return nil, err
	}

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, datasource.NewConnectionError("mysql", config.Host, config.Port, config.Database, err)
	}

	// Configure connection pool
	if config.MaxConnections > 0 {
		db.SetMaxOpenConns(config.MaxConnections)
	} else {
		db.SetMaxOpenConns(10) // Default
	}

	if config.MaxIdleConnections > 0 {
		db.SetMaxIdleConns(config.MaxIdleConnections)
	} else {
		db.SetMaxIdleConns(5) // Default
	}

	if config.ConnectionMaxLifetime > 0 {
		db.SetConnMaxLifetime(config.ConnectionMaxLifetime)
	}

	// Test connection
	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return nil, datasource.NewConnectionError("mysql", config.Host, config.Port, config.Database, err)
	}

	return &Connection{
		db:           db,
		config:       config,
		queryTimeout: config.QueryTimeout,
		readOnly:     config.ReadOnly,
	}, nil
}

// ValidateConfig checks if the MySQL config is valid.
func (p *Provider) ValidateConfig(config datasource.ConnectionConfig) error {
	// Need either connection URL or host/database
	hasURL := config.ConnectionURL != "" || config.ConnectionURLEnv != ""
	hasFields := config.Host != "" && config.Database != ""

	if !hasURL && !hasFields {
		return datasource.NewConfigError("connection", "either connectionUrl/connectionUrlEnv or host+database required", nil)
	}

	if config.Port < 0 || config.Port > 65535 {
		return datasource.NewConfigError("port", "must be between 0 and 65535", nil)
	}

	if config.MaxConnections < 0 {
		return datasource.NewConfigError("maxConnections", "must be non-negative", nil)
	}

	if config.QueryTimeout < 0 {
		return datasource.NewConfigError("queryTimeout", "must be non-negative", nil)
	}

	return nil
}

// Capabilities returns MySQL capabilities.
func (p *Provider) Capabilities() datasource.Capabilities {
	return datasource.Capabilities{
		SupportsTransactions:  true,
		SupportsStreaming:     true,
		SupportsPreparedStmts: true,
		SupportsNamedParams:   false, // MySQL uses ? positional
		MaxQuerySize:          0,     // Configurable via max_allowed_packet
		ParameterStyle:        datasource.ParameterStylePositionalQuestion,
	}
}

// buildDSN builds a MySQL DSN from config.
// Format: [username[:password]@][protocol[(address)]]/dbname[?param=value]
func buildDSN(config datasource.ConnectionConfig) (string, error) {
	// Check for connection URL from environment variable
	if config.ConnectionURLEnv != "" {
		if envURL := os.Getenv(config.ConnectionURLEnv); envURL != "" {
			return convertURLToDSN(envURL), nil
		}
	}

	// Use connection URL if provided directly
	if config.ConnectionURL != "" {
		return convertURLToDSN(config.ConnectionURL), nil
	}

	// Build from individual fields
	password := config.Password
	if config.PasswordEnv != "" {
		if envPass := os.Getenv(config.PasswordEnv); envPass != "" {
			password = envPass
		}
	}

	// Build DSN string
	var dsn strings.Builder

	// User and password
	if config.Username != "" {
		dsn.WriteString(config.Username)
		if password != "" {
			dsn.WriteString(":")
			dsn.WriteString(password)
		}
		dsn.WriteString("@")
	}

	// Protocol and address
	dsn.WriteString("tcp(")
	dsn.WriteString(config.Host)
	if config.Port > 0 {
		_, _ = fmt.Fprintf(&dsn, ":%d", config.Port)
	} else {
		dsn.WriteString(":3306") // Default MySQL port
	}
	dsn.WriteString(")/")

	// Database name
	dsn.WriteString(config.Database)

	// Parameters
	params := make([]string, 0)
	params = append(params, "parseTime=true") // Always parse TIME/DATE values

	// SSL mode
	if config.SSLMode != "" && config.SSLMode != "disable" {
		params = append(params, "tls="+config.SSLMode)
	}

	// Timeout
	if config.QueryTimeout > 0 {
		params = append(params, fmt.Sprintf("timeout=%s", config.QueryTimeout))
		params = append(params, fmt.Sprintf("readTimeout=%s", config.QueryTimeout))
		params = append(params, fmt.Sprintf("writeTimeout=%s", config.QueryTimeout))
	}

	// Extra parameters
	for key, value := range config.Extra {
		params = append(params, fmt.Sprintf("%s=%v", key, value))
	}

	if len(params) > 0 {
		dsn.WriteString("?")
		dsn.WriteString(strings.Join(params, "&"))
	}

	return dsn.String(), nil
}

// convertURLToDSN converts a mysql:// URL to a DSN string.
func convertURLToDSN(url string) string {
	// If it's already a DSN (doesn't start with mysql://), return as-is
	if !strings.HasPrefix(url, "mysql://") {
		return url
	}

	// Strip mysql:// prefix
	url = strings.TrimPrefix(url, "mysql://")

	// The go-sql-driver/mysql expects DSN format, not URL format
	// URL: mysql://user:pass@host:port/db?params
	// DSN: user:pass@tcp(host:port)/db?params

	// Find the @ separator for credentials
	atIdx := strings.Index(url, "@")
	if atIdx == -1 {
		// No credentials, just host/db
		// Convert host:port/db to tcp(host:port)/db
		slashIdx := strings.Index(url, "/")
		if slashIdx == -1 {
			return "tcp(" + url + ")/"
		}
		host := url[:slashIdx]
		rest := url[slashIdx:]
		return "tcp(" + host + ")" + rest
	}

	creds := url[:atIdx]
	rest := url[atIdx+1:]

	// Find where host:port ends
	slashIdx := strings.Index(rest, "/")
	if slashIdx == -1 {
		return creds + "@tcp(" + rest + ")/"
	}

	host := rest[:slashIdx]
	dbAndParams := rest[slashIdx:]

	return creds + "@tcp(" + host + ")" + dbAndParams
}

// isWriteQuery checks if a query is a write operation.
func isWriteQuery(query string) bool {
	q := strings.TrimSpace(strings.ToUpper(query))
	writeKeywords := []string{"INSERT", "UPDATE", "DELETE", "DROP", "CREATE", "ALTER", "TRUNCATE", "GRANT", "REVOKE"}
	for _, kw := range writeKeywords {
		if strings.HasPrefix(q, kw) {
			return true
		}
	}
	return false
}

// queryTimeoutContext returns a context with query timeout if configured.
func queryTimeoutContext(ctx context.Context, timeout time.Duration) (context.Context, context.CancelFunc) {
	if timeout > 0 {
		return context.WithTimeout(ctx, timeout)
	}
	return ctx, func() {}
}
