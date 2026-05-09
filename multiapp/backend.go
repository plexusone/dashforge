// Package multiapp provides an AppBackend adapter for DashForge.
// This enables DashForge to run in both standalone and multi-app modes.
//
// Standalone mode (existing behavior):
//
//	go run ./cmd/dashforge-server serve
//
// Multi-app mode (via systemforge multi-app server):
//
//	import "github.com/plexusone/dashforge/multiapp"
//	server.RegisterApp(multiapp.NewBackend(nil))
package multiapp

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"os"

	"entgo.io/ent"
	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	"github.com/go-chi/chi/v5"
	cfmultiapp "github.com/grokify/systemforge/multiapp"
	"github.com/jackc/pgx/v5/stdlib"
	dashent "github.com/plexusone/dashforge/ent"
	"github.com/plexusone/dashforge/internal/server"
	"github.com/plexusone/dashforge/internal/server/db"
)

// Backend implements multiapp.AppBackend for DashForge.
// It adapts the existing DashForge server to work with the multi-app framework.
type Backend struct {
	cfg *server.Config

	// Runtime dependencies (set during Routes())
	deps   cfmultiapp.Dependencies
	db     db.Database
	server *server.Server
}

// NewBackend creates a new DashForge backend for multi-app deployment.
// Pass nil to load configuration from environment variables.
func NewBackend(cfg *server.Config) *Backend {
	return &Backend{cfg: cfg}
}

// Slug returns the app identifier.
func (b *Backend) Slug() string {
	return "dashforge"
}

// Name returns the display name.
func (b *Backend) Name() string {
	return "DashForge"
}

// EntSchemas returns the Ent schemas for this app.
// DashForge uses Ent migrations via the database interface.
func (b *Backend) EntSchemas() []ent.Schema {
	return nil
}

// Migrations returns database migrations for this app.
// DashForge uses Ent's auto-migration, which is handled when creating
// the database connection.
func (b *Backend) Migrations() []cfmultiapp.Migration {
	return nil
}

// Routes returns the HTTP routes for DashForge.
// This creates all necessary services and returns the router.
func (b *Backend) Routes(deps cfmultiapp.Dependencies) chi.Router {
	b.deps = deps
	ctx := context.Background()

	// Load config if not provided
	if b.cfg == nil {
		b.cfg = b.loadConfigFromEnv()
	}

	// Create database connection using multiapp's schema-aware connection
	if err := b.setupDatabase(ctx); err != nil {
		deps.Logger.Error("failed to setup database", "error", err)
		return b.errorRouter("database setup error")
	}

	// Create server using existing server package with our database
	srv, err := server.NewServerWithDatabase(*b.cfg, deps.Logger, b.db)
	if err != nil {
		deps.Logger.Error("failed to create server", "error", err)
		return b.errorRouter("server creation error")
	}
	b.server = srv

	// Wrap the ServeMux in a chi.Router for compatibility
	r := chi.NewRouter()
	r.Mount("/", srv.Router())

	return r
}

// OnRegister is called when the app is registered with the server.
func (b *Backend) OnRegister(ctx context.Context, cfg *cfmultiapp.AppConfig) error {
	b.deps.Logger.Info("DashForge registered",
		"app_id", cfg.AppID,
		"schema", cfg.DatabaseSchema,
	)
	return nil
}

// OnShutdown is called during graceful shutdown.
func (b *Backend) OnShutdown(ctx context.Context) error {
	b.deps.Logger.Info("DashForge shutting down")

	if b.server != nil {
		if err := b.server.Shutdown(ctx); err != nil {
			return fmt.Errorf("failed to shutdown server: %w", err)
		}
	}

	return nil
}

// setupDatabase creates a database connection scoped to the app's schema.
func (b *Backend) setupDatabase(ctx context.Context) error {
	// Get the schema-scoped database connection from multiapp
	schemaDB := b.deps.DB

	// Create a *sql.DB that uses the schema's search_path
	pool := schemaDB.Pool()
	poolConfig := pool.Config().ConnConfig.Copy()

	// Set search_path in connection parameters
	if poolConfig.RuntimeParams == nil {
		poolConfig.RuntimeParams = make(map[string]string)
	}
	poolConfig.RuntimeParams["search_path"] = schemaDB.Schema() + ", public"

	// Register the config and get a connection string
	connStr := stdlib.RegisterConnConfig(poolConfig)

	// Open standard database connection
	sqlDB, err := sql.Open("pgx", connStr)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}

	// Create Ent client
	drv := entsql.OpenDB(dialect.Postgres, sqlDB)
	client := dashent.NewClient(dashent.Driver(drv))

	// Run Ent auto-migration
	if err := client.Schema.Create(ctx); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	// Wrap in our Database interface
	b.db = &schemaDatabase{
		client: client,
		sqlDB:  sqlDB,
	}

	return nil
}

// loadConfigFromEnv loads configuration from environment variables.
func (b *Backend) loadConfigFromEnv() *server.Config {
	return &server.Config{
		JWTSecret:               os.Getenv("JWT_SECRET"),
		BaseURL:                 os.Getenv("BASE_URL"),
		GitHubClientID:          os.Getenv("GITHUB_CLIENT_ID"),
		GitHubClientSecret:      os.Getenv("GITHUB_CLIENT_SECRET"),
		GoogleClientID:          os.Getenv("GOOGLE_CLIENT_ID"),
		GoogleClientSecret:      os.Getenv("GOOGLE_CLIENT_SECRET"),
		CoreControlURL:          os.Getenv("CORECONTROL_URL"),
		CoreControlClientID:     os.Getenv("CORECONTROL_CLIENT_ID"),
		CoreControlClientSecret: os.Getenv("CORECONTROL_CLIENT_SECRET"),
		CoreControlCallbackURL:  os.Getenv("CORECONTROL_CALLBACK_URL"),
		AuthZMode:               os.Getenv("AUTHZ_MODE"),
		SpiceDBEndpoint:         os.Getenv("SPICEDB_ENDPOINT"),
		SpiceDBToken:            os.Getenv("SPICEDB_TOKEN"),
		SpiceDBInsecure:         os.Getenv("SPICEDB_INSECURE") == "true",
		DashboardDir:            os.Getenv("DASHBOARD_DIR"),
	}
}

// errorRouter returns a router that returns an error for all requests.
func (b *Backend) errorRouter(msg string) chi.Router {
	r := chi.NewRouter()
	r.HandleFunc("/*", func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, msg, http.StatusInternalServerError)
	})
	return r
}

// schemaDatabase implements db.Database for schema-isolated connections.
type schemaDatabase struct {
	client *dashent.Client
	sqlDB  *sql.DB
}

func (d *schemaDatabase) Client() *dashent.Client {
	return d.client
}

func (d *schemaDatabase) DB() *sql.DB {
	return d.sqlDB
}

func (d *schemaDatabase) Query(ctx context.Context, query string, params map[string]any) (*db.QueryResult, error) {
	// Simplified implementation - the full implementation is in db package
	return nil, fmt.Errorf("raw queries not implemented in multiapp mode")
}

func (d *schemaDatabase) Ping(ctx context.Context) error {
	return d.sqlDB.PingContext(ctx)
}

func (d *schemaDatabase) Close() error {
	if d.client != nil {
		if err := d.client.Close(); err != nil {
			return err
		}
	}
	return d.sqlDB.Close()
}

func (d *schemaDatabase) Type() string {
	return "postgres"
}

func (d *schemaDatabase) Migrate(ctx context.Context) error {
	return d.client.Schema.Create(ctx)
}

func (d *schemaDatabase) MigrateRLS(ctx context.Context) error {
	// RLS is managed by the multiapp framework at the schema level
	return nil
}
