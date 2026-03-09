// Package db provides database abstractions for Dashforge Server.
package db

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	_ "github.com/lib/pq" // PostgreSQL driver

	"github.com/plexusone/dashforge/ent"
)

// Database is the interface for database operations.
type Database interface {
	// Client returns the Ent client for ORM operations.
	Client() *ent.Client

	// DB returns the underlying sql.DB for raw operations (RLS, etc.).
	DB() *sql.DB

	// Query executes a raw SQL query and returns the results.
	Query(ctx context.Context, query string, params map[string]any) (*QueryResult, error)

	// Ping checks the database connection.
	Ping(ctx context.Context) error

	// Close closes the database connection.
	Close() error

	// Type returns the database type (postgres, mysql, sqlite).
	Type() string

	// Migrate runs database migrations.
	Migrate(ctx context.Context) error

	// MigrateRLS applies Row Level Security policies (PostgreSQL only).
	MigrateRLS(ctx context.Context) error
}

// QueryResult holds the results of a database query.
type QueryResult struct {
	// Columns contains the column names.
	Columns []string `json:"columns"`

	// Rows contains the data rows.
	Rows []map[string]any `json:"rows"`

	// RowCount is the number of rows returned.
	RowCount int `json:"rowCount"`

	// ExecutionTimeMs is how long the query took.
	ExecutionTimeMs int64 `json:"executionTimeMs"`
}

// Open opens a database connection based on the URL scheme.
func Open(url string) (Database, error) {
	switch {
	case strings.HasPrefix(url, "postgres://"), strings.HasPrefix(url, "postgresql://"):
		return openPostgres(url)
	case strings.HasPrefix(url, "mysql://"):
		return nil, fmt.Errorf("mysql not yet implemented")
	case strings.HasPrefix(url, "sqlite://"), strings.HasPrefix(url, "file:"):
		return nil, fmt.Errorf("sqlite not yet implemented")
	default:
		return nil, fmt.Errorf("unsupported database URL scheme: %s", url)
	}
}

// openPostgres opens a PostgreSQL connection with Ent.
func openPostgres(url string) (Database, error) {
	// Open raw SQL connection
	db, err := sql.Open("postgres", url)
	if err != nil {
		return nil, fmt.Errorf("opening postgres connection: %w", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	// Wrap with Ent driver
	drv := entsql.OpenDB(dialect.Postgres, db)

	// Create Ent client
	client := ent.NewClient(ent.Driver(drv))

	return &postgresDB{
		url:    url,
		db:     db,
		client: client,
	}, nil
}

// postgresDB implements Database for PostgreSQL.
type postgresDB struct {
	url    string
	db     *sql.DB
	client *ent.Client
}

func (p *postgresDB) Client() *ent.Client {
	return p.client
}

func (p *postgresDB) DB() *sql.DB {
	return p.db
}

func (p *postgresDB) Query(ctx context.Context, query string, params map[string]any) (*QueryResult, error) {
	start := time.Now()

	// Convert named parameters to positional for PostgreSQL
	// For simplicity, we'll use positional parameters ($1, $2, etc.)
	args := make([]any, 0, len(params))
	processedQuery := query

	// If params is a map, convert to ordered slice
	// This is a simple implementation - for production, use proper parameter binding
	i := 1
	for key, value := range params {
		placeholder := fmt.Sprintf("$%d", i)
		processedQuery = strings.ReplaceAll(processedQuery, ":"+key, placeholder)
		processedQuery = strings.ReplaceAll(processedQuery, "@"+key, placeholder)
		args = append(args, value)
		i++
	}

	rows, err := p.db.QueryContext(ctx, processedQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("executing query: %w", err)
	}
	defer func() {
		_ = rows.Close()
	}()

	// Get column names
	columns, err := rows.Columns()
	if err != nil {
		return nil, fmt.Errorf("getting columns: %w", err)
	}

	// Scan all rows
	var results []map[string]any
	for rows.Next() {
		// Create a slice of interface{} to hold column values
		values := make([]any, len(columns))
		valuePtrs := make([]any, len(columns))
		for i := range values {
			valuePtrs[i] = &values[i]
		}

		if err := rows.Scan(valuePtrs...); err != nil {
			return nil, fmt.Errorf("scanning row: %w", err)
		}

		// Convert to map
		row := make(map[string]any)
		for i, col := range columns {
			val := values[i]
			// Handle []byte -> string conversion
			if b, ok := val.([]byte); ok {
				row[col] = string(b)
			} else {
				row[col] = val
			}
		}
		results = append(results, row)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating rows: %w", err)
	}

	return &QueryResult{
		Columns:         columns,
		Rows:            results,
		RowCount:        len(results),
		ExecutionTimeMs: time.Since(start).Milliseconds(),
	}, nil
}

func (p *postgresDB) Ping(ctx context.Context) error {
	return p.db.PingContext(ctx)
}

func (p *postgresDB) Close() error {
	if p.client != nil {
		if err := p.client.Close(); err != nil {
			return err
		}
	}
	return nil
}

func (p *postgresDB) Type() string {
	return "postgres"
}

func (p *postgresDB) Migrate(ctx context.Context) error {
	// Run Ent auto-migration
	if err := p.client.Schema.Create(ctx); err != nil {
		return fmt.Errorf("running migrations: %w", err)
	}
	return nil
}

func (p *postgresDB) MigrateRLS(ctx context.Context) error {
	return ApplyRLS(ctx, p.db)
}
