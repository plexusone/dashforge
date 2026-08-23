package db

import (
	"context"
	"database/sql"
	"fmt"
	"net/url"
	"strings"
	"time"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	_ "github.com/go-sql-driver/mysql" // MySQL-wire driver (Dolt, MySQL)
	"github.com/grokify/godolt"

	"github.com/plexusone/uiforge/ent"
)

// openMySQL opens a MySQL-wire connection with Ent. This is the path for
// Dolt (`dolt sql-server`) as UIForge's local metadata database, and works
// against vanilla MySQL as well. The target database is created if it does
// not exist, matching the ecosystem's local-Dolt pattern.
func openMySQL(rawURL string) (Database, error) {
	dsn, err := mysqlURLToDSN(rawURL)
	if err != nil {
		return nil, err
	}
	dsn = godolt.EnsureParseTime(dsn)

	base, dbName, err := godolt.SplitDSN(dsn)
	if err != nil {
		return nil, err
	}
	serverDB, err := sql.Open("mysql", base)
	if err != nil {
		return nil, fmt.Errorf("opening mysql server connection: %w", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := godolt.CreateDatabase(ctx, serverDB, dbName); err != nil {
		if cerr := serverDB.Close(); cerr != nil {
			return nil, fmt.Errorf("%w (also closing server connection: %v)", err, cerr)
		}
		return nil, err
	}
	if err := serverDB.Close(); err != nil {
		return nil, fmt.Errorf("closing server connection: %w", err)
	}

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("opening mysql connection: %w", err)
	}
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	drv := entsql.OpenDB(dialect.MySQL, db)
	client := ent.NewClient(ent.Driver(drv))

	return &mysqlDB{
		dsn:    dsn,
		db:     db,
		client: client,
	}, nil
}

// mysqlURLToDSN converts a mysql:// URL into a go-sql-driver DSN
// (user:pass@tcp(host:port)/db?query). A string without the mysql://
// scheme is treated as an already-formed driver DSN.
func mysqlURLToDSN(rawURL string) (string, error) {
	if !strings.HasPrefix(rawURL, "mysql://") {
		return rawURL, nil
	}
	u, err := url.Parse(rawURL)
	if err != nil {
		return "", fmt.Errorf("parsing mysql URL: %w", err)
	}
	user := ""
	if u.User != nil {
		user = u.User.Username()
		if pass, ok := u.User.Password(); ok {
			user += ":" + pass
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
	dbName := strings.TrimPrefix(u.Path, "/")
	if dbName == "" {
		return "", fmt.Errorf("mysql URL %q has no database name", rawURL)
	}
	dsn := fmt.Sprintf("%s@tcp(%s:%s)/%s", user, host, port, dbName)
	if u.RawQuery != "" {
		dsn += "?" + u.RawQuery
	}
	return dsn, nil
}

// mysqlDB implements Database for MySQL-wire databases (Dolt, MySQL).
type mysqlDB struct {
	dsn    string
	db     *sql.DB
	client *ent.Client
}

func (m *mysqlDB) Client() *ent.Client {
	return m.client
}

func (m *mysqlDB) DB() *sql.DB {
	return m.db
}

func (m *mysqlDB) Query(ctx context.Context, query string, params map[string]any) (result *QueryResult, err error) {
	start := time.Now()

	// Convert named parameters (:key / @key) to MySQL's ? placeholders.
	args := make([]any, 0, len(params))
	processedQuery := query
	for key, value := range params {
		processedQuery = strings.ReplaceAll(processedQuery, ":"+key, "?")
		processedQuery = strings.ReplaceAll(processedQuery, "@"+key, "?")
		args = append(args, value)
	}

	rows, err := m.db.QueryContext(ctx, processedQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("executing query: %w", err)
	}
	defer func() {
		if cerr := rows.Close(); cerr != nil && err == nil {
			err = cerr
			result = nil
		}
	}()

	columns, err := rows.Columns()
	if err != nil {
		return nil, fmt.Errorf("getting columns: %w", err)
	}

	var results []map[string]any
	for rows.Next() {
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

func (m *mysqlDB) Ping(ctx context.Context) error {
	return m.db.PingContext(ctx)
}

func (m *mysqlDB) Close() error {
	if m.client != nil {
		if err := m.client.Close(); err != nil {
			return err
		}
	}
	return nil
}

func (m *mysqlDB) Type() string {
	return "mysql"
}

func (m *mysqlDB) Migrate(ctx context.Context) error {
	if err := m.client.Schema.Create(ctx); err != nil {
		return fmt.Errorf("running migrations: %w", err)
	}
	return nil
}

func (m *mysqlDB) MigrateRLS(context.Context) error {
	return fmt.Errorf("row level security requires PostgreSQL; the mysql/dolt metadata database does not support --enable-rls")
}
