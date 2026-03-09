// Package server provides the Dashforge HTTP server with database support.
package server

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/grokify/coreforge/authz"
	"github.com/grokify/coreforge/authz/spicedb"
	"github.com/plexusone/dashforge/builder"
	"github.com/plexusone/dashforge/datasource"
	// Import providers for registration via init()
	_ "github.com/plexusone/dashforge/datasource/providers/mysql"
	_ "github.com/plexusone/dashforge/datasource/providers/postgres"
	// Import channel adapters for registration via init()
	_ "github.com/plexusone/dashforge/integration/channel/email"
	_ "github.com/plexusone/dashforge/integration/channel/slack"
	_ "github.com/plexusone/dashforge/integration/channel/webhook"
	_ "github.com/plexusone/dashforge/integration/channel/whatsapp"
	localAuthz "github.com/plexusone/dashforge/internal/authz"
	"github.com/plexusone/dashforge/internal/server/api"
	"github.com/plexusone/dashforge/internal/server/auth"
	"github.com/plexusone/dashforge/internal/server/db"
	"github.com/plexusone/dashforge/viewer"
)

// Config holds server configuration.
type Config struct {
	// Port to listen on.
	Port int

	// ConfigPath is the path to a YAML config file.
	ConfigPath string

	// DatabaseURL is the connection string for the primary database.
	DatabaseURL string

	// DashboardDir is the directory containing dashboard JSON files.
	DashboardDir string

	// DisableAuth disables authentication (for development).
	DisableAuth bool

	// AutoMigrate runs database migrations on startup.
	AutoMigrate bool

	// EnableRLS enables Row Level Security for multi-tenancy.
	EnableRLS bool

	// BaseURL is the public URL of the server (for OAuth callbacks).
	BaseURL string

	// JWTSecret is the secret for signing JWT tokens.
	JWTSecret string

	// OAuth provider credentials
	GitHubClientID     string
	GitHubClientSecret string
	GoogleClientID     string
	GoogleClientSecret string

	// CoreControl (CoreAuth) settings
	CoreControlURL          string
	CoreControlClientID     string
	CoreControlClientSecret string
	CoreControlCallbackURL  string
	CoreControlScopes       []string

	// Authorization settings
	AuthZMode         string // "simple" or "spicedb"
	SpiceDBEndpoint   string
	SpiceDBToken      string
	SpiceDBInsecure   bool
}

// Server is the Dashforge HTTP server.
type Server struct {
	config       Config
	logger       *slog.Logger
	db           db.Database
	mux          *http.ServeMux
	jwtService   *auth.JWTService
	oauthHandler *auth.OAuthHandler
	authzService *localAuthz.Service

	// Data source management
	dsManager  *datasource.Manager
	dsExecutor *datasource.QueryExecutor

	// AI handler
	aiHandler *api.AIHandler
}

// New creates a new Dashforge server.
func New(cfg Config, logger *slog.Logger) (*Server, error) {
	s := &Server{
		config: cfg,
		logger: logger,
		mux:    http.NewServeMux(),
	}

	// Initialize database if URL provided
	if cfg.DatabaseURL != "" {
		database, err := db.Open(cfg.DatabaseURL)
		if err != nil {
			return nil, fmt.Errorf("opening database: %w", err)
		}
		s.db = database

		// Verify connection
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := database.Ping(ctx); err != nil {
			return nil, fmt.Errorf("connecting to database: %w", err)
		}
		logger.Info("database connected", "type", database.Type())

		// Run migrations if enabled
		if cfg.AutoMigrate {
			logger.Info("running database migrations")
			if err := database.Migrate(ctx); err != nil {
				return nil, fmt.Errorf("running migrations: %w", err)
			}
			logger.Info("migrations complete")

			// Apply RLS policies if enabled
			if cfg.EnableRLS {
				logger.Info("applying row level security policies")
				if err := database.MigrateRLS(ctx); err != nil {
					return nil, fmt.Errorf("applying RLS: %w", err)
				}
				logger.Info("RLS policies applied")
			}
		}
	}

	// Initialize data source manager (always, even without primary DB)
	s.dsManager = datasource.NewManager(datasource.ManagerConfig{
		Logger: logger,
	})

	// Initialize query executor
	s.dsExecutor = datasource.NewQueryExecutor(datasource.QueryExecutorConfig{
		Manager:        s.dsManager,
		Logger:         logger,
		DefaultTimeout: 30 * time.Second,
		MaxRowsDefault: 1000,
		MaxRowsLimit:   10000,
	})

	// Initialize JWT service if secret provided
	if cfg.JWTSecret != "" {
		jwtSvc, err := auth.NewJWTService(auth.JWTConfig{
			Secret:          cfg.JWTSecret,
			AccessTokenTTL:  15 * time.Minute,
			RefreshTokenTTL: 7 * 24 * time.Hour,
			Issuer:          "dashforge",
		})
		if err != nil {
			return nil, fmt.Errorf("creating JWT service: %w", err)
		}
		s.jwtService = jwtSvc

		// Initialize OAuth handler if we have an Ent client
		if s.db != nil {
			oauthCfg := auth.NewOAuthConfig(auth.OAuthProviderConfig{
				GitHubClientID:          cfg.GitHubClientID,
				GitHubClientSecret:      cfg.GitHubClientSecret,
				GoogleClientID:          cfg.GoogleClientID,
				GoogleClientSecret:      cfg.GoogleClientSecret,
				CoreControlURL:          cfg.CoreControlURL,
				CoreControlClientID:     cfg.CoreControlClientID,
				CoreControlClientSecret: cfg.CoreControlClientSecret,
				CoreControlCallbackURL:  cfg.CoreControlCallbackURL,
				CoreControlScopes:       cfg.CoreControlScopes,
				BaseURL:                 cfg.BaseURL,
			})
			s.oauthHandler = auth.NewOAuthHandler(oauthCfg, jwtSvc, s.db.Client(), logger, cfg.BaseURL)
		}
	}

	// Initialize authorization service
	if s.db != nil {
		var authzOpts []localAuthz.ServiceOption
		if cfg.AuthZMode == "spicedb" && cfg.SpiceDBEndpoint != "" {
			spiceCfg := spicedb.Config{
				Mode:     "remote",
				Endpoint: cfg.SpiceDBEndpoint,
				Token:    cfg.SpiceDBToken,
				Insecure: cfg.SpiceDBInsecure,
			}

			spiceClient, err := spicedb.NewClient(context.Background(), spiceCfg, logger)
			if err != nil {
				return nil, fmt.Errorf("creating SpiceDB client: %w", err)
			}

			// Write schema on startup
			if err := spiceClient.WriteSchema(context.Background(), localAuthz.DashForgeSchema); err != nil {
				return nil, fmt.Errorf("writing SpiceDB schema: %w", err)
			}

			spiceProvider := spicedb.NewProvider(spiceClient)
			authzOpts = append(authzOpts,
				localAuthz.WithMode(localAuthz.ModeSpiceDB),
				localAuthz.WithSpiceDBProvider(spiceProvider),
				localAuthz.WithSyncMode(authz.SyncModeEventual),
			)
			logger.Info("SpiceDB authorization initialized", "endpoint", cfg.SpiceDBEndpoint)
		} else {
			logger.Info("using simple role hierarchy authorization")
		}
		s.authzService = localAuthz.NewService(s.db.Client(), authzOpts...)
	}

	// Initialize AI handler (reads API keys from environment)
	aiHandler, err := api.NewAIHandler(api.AIConfig{
		EnableFallback: true,
		Timeout:        60 * time.Second,
	}, logger)
	if err != nil {
		logger.Warn("AI handler initialization failed", "error", err)
	} else {
		s.aiHandler = aiHandler
	}

	// Set up routes
	s.setupRoutes()

	return s, nil
}

func (s *Server) setupRoutes() {
	// Health check
	s.mux.HandleFunc("GET /health", s.handleHealth)

	// Embedded viewer
	s.mux.Handle("GET /viewer/", http.StripPrefix("/viewer/", http.FileServer(http.FS(viewer.FS()))))

	// Embedded builder (dashboard editor)
	s.mux.Handle("GET /builder/", http.StripPrefix("/builder/", http.FileServer(http.FS(builder.FS()))))

	// OAuth/Auth routes (if configured)
	if s.oauthHandler != nil {
		s.mux.Handle("/api/v1/auth/", s.oauthHandler)
	}

	// API routes
	apiHandler := api.NewHandler(s.db, s.logger)
	s.mux.Handle("/api/", apiHandler)

	// Data source API routes (separate handler for cleaner organization)
	dsHandler := api.NewDataSourceHandler(s.db, s.dsManager, s.dsExecutor, s.logger)
	s.mux.Handle("/api/v1/datasources", dsHandler)
	s.mux.Handle("/api/v1/datasources/", dsHandler)

	// AI API routes
	if s.aiHandler != nil {
		s.mux.Handle("/api/v1/ai/", s.aiHandler)
	}

	// Integration API routes
	integrationHandler := api.NewIntegrationHandler(s.db, s.logger)
	s.mux.Handle("/api/v1/integrations", integrationHandler)
	s.mux.Handle("/api/v1/integrations/", integrationHandler)

	// Alert API routes
	alertHandler := api.NewAlertHandler(s.db, s.logger)
	s.mux.Handle("/api/v1/alerts", alertHandler)
	s.mux.Handle("/api/v1/alerts/", alertHandler)

	// Marketplace API routes
	marketplaceHandler := api.NewMarketplaceHandler(s.db, s.logger)
	s.mux.Handle("/api/v1/marketplace/", marketplaceHandler)

	// Static dashboard files
	if s.config.DashboardDir != "" {
		s.mux.Handle("GET /dashboards/", http.StripPrefix("/dashboards/",
			http.FileServer(http.Dir(s.config.DashboardDir))))
	}

	// Root redirect to viewer
	s.mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			http.Redirect(w, r, "/viewer/", http.StatusTemporaryRedirect)
			return
		}
		http.NotFound(w, r)
	})
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Check database if configured
	if s.db != nil {
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		if err := s.db.Ping(ctx); err != nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			_, _ = w.Write([]byte(`{"status":"unhealthy","database":"disconnected"}`))
			return
		}
	}

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"status":"ok"}`))
}

// ListenAndServe starts the HTTP server.
func (s *Server) ListenAndServe() error {
	addr := fmt.Sprintf(":%d", s.config.Port)

	server := &http.Server{
		Addr:              addr,
		Handler:           s.mux,
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      60 * time.Second,
		IdleTimeout:       120 * time.Second,
	}

	s.logger.Info("starting server",
		"addr", addr,
		"dashboard_dir", s.config.DashboardDir,
		"db_configured", s.db != nil,
	)

	fmt.Printf("\nDashforge Server starting...\n")
	fmt.Printf("  Builder:    http://localhost:%d/builder/\n", s.config.Port)
	fmt.Printf("  Viewer:     http://localhost:%d/viewer/\n", s.config.Port)
	fmt.Printf("  API:        http://localhost:%d/api/v1/\n", s.config.Port)
	fmt.Printf("  Health:     http://localhost:%d/health\n", s.config.Port)
	if s.config.DashboardDir != "" {
		fmt.Printf("  Dashboards: http://localhost:%d/dashboards/\n", s.config.Port)
	}
	if s.db != nil {
		fmt.Printf("  Database:   %s (connected)\n", s.db.Type())
	}
	if s.oauthHandler != nil {
		fmt.Printf("  Auth:       OAuth enabled (GitHub: %t, Google: %t, CoreControl: %t)\n",
			s.config.GitHubClientID != "", s.config.GoogleClientID != "",
			s.config.CoreControlClientID != "")
	}
	if s.aiHandler != nil {
		fmt.Printf("  AI:         Enabled (set ANTHROPIC_API_KEY, OPENAI_API_KEY, etc.)\n")
	} else {
		fmt.Printf("  AI:         Disabled (no API keys configured)\n")
	}
	fmt.Println("\nPress Ctrl+C to stop")

	return server.ListenAndServe()
}

// Shutdown gracefully shuts down the server.
func (s *Server) Shutdown(ctx context.Context) error {
	// Close all data source connections
	if s.dsManager != nil {
		if err := s.dsManager.CloseAll(); err != nil {
			s.logger.Error("error closing data source connections", "error", err)
		}
	}

	// Close primary database
	if s.db != nil {
		if err := s.db.Close(); err != nil {
			s.logger.Error("error closing database", "error", err)
		}
	}
	return nil
}
