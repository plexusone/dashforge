// Package postgres provides a PostgreSQL implementation of the datasource.Provider interface.
package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"

	_ "github.com/lib/pq" // PostgreSQL driver

	"github.com/plexusone/dashforge/datasource"
)

func init() {
	datasource.Register(&Provider{})
}

// Provider implements datasource.Provider for PostgreSQL.
type Provider struct{}

// Name returns "postgres".
func (p *Provider) Name() string {
	return "postgres"
}

// Connect establishes a PostgreSQL connection.
func (p *Provider) Connect(ctx context.Context, config datasource.ConnectionConfig) (datasource.Connection, error) {
	connStr := buildConnectionString(config)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, datasource.NewConnectionError("postgres", config.Host, config.Port, config.Database, err)
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
		return nil, datasource.NewConnectionError("postgres", config.Host, config.Port, config.Database, err)
	}

	return &Connection{
		db:           db,
		config:       config,
		queryTimeout: config.QueryTimeout,
		readOnly:     config.ReadOnly,
	}, nil
}

// ValidateConfig checks if the PostgreSQL config is valid.
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

// Capabilities returns PostgreSQL capabilities.
func (p *Provider) Capabilities() datasource.Capabilities {
	return datasource.Capabilities{
		SupportsTransactions:  true,
		SupportsStreaming:     true,
		SupportsPreparedStmts: true,
		SupportsNamedParams:   false, // PostgreSQL uses $1, $2 positional
		MaxQuerySize:          0,     // No practical limit
		ParameterStyle:        datasource.ParameterStylePositionalDollar,
	}
}

// buildConnectionString builds a PostgreSQL connection string from config.
func buildConnectionString(config datasource.ConnectionConfig) string {
	// Check for connection URL from environment variable
	if config.ConnectionURLEnv != "" {
		if envURL := os.Getenv(config.ConnectionURLEnv); envURL != "" {
			return envURL
		}
	}

	// Use connection URL if provided directly
	if config.ConnectionURL != "" {
		return config.ConnectionURL
	}

	// Build from individual fields
	password := config.Password
	if config.PasswordEnv != "" {
		if envPass := os.Getenv(config.PasswordEnv); envPass != "" {
			password = envPass
		}
	}

	// Build connection URL
	u := &url.URL{
		Scheme: "postgres",
		Host:   config.Host,
		Path:   "/" + config.Database,
	}

	if config.Port > 0 {
		u.Host = fmt.Sprintf("%s:%d", config.Host, config.Port)
	}

	if config.Username != "" {
		if password != "" {
			u.User = url.UserPassword(config.Username, password)
		} else {
			u.User = url.User(config.Username)
		}
	}

	// Add query parameters
	q := u.Query()
	if config.SSLMode != "" {
		q.Set("sslmode", config.SSLMode)
	} else {
		q.Set("sslmode", "prefer") // Default to prefer
	}

	// Add any extra options
	for key, value := range config.Extra {
		if s, ok := value.(string); ok {
			q.Set(key, s)
		} else if i, ok := value.(int); ok {
			q.Set(key, strconv.Itoa(i))
		}
	}

	u.RawQuery = q.Encode()

	return u.String()
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
