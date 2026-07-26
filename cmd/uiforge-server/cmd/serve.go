package cmd

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/spf13/cobra"

	"github.com/plexusone/uiforge/internal/server"
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the Dashforge server",
	Long: `Start the Dashforge server with database support.

The server provides:
  - Static dashboard serving
  - REST API for dashboard CRUD
  - Database query execution
  - Optional authentication

Example:
  uiforge-server serve --port 8080 --config config.yaml
  uiforge-server serve --db-url postgres://localhost/analytics`,
	RunE: runServe,
}

func init() {
	rootCmd.AddCommand(serveCmd)

	serveCmd.Flags().IntP("port", "p", 8080, "Port to listen on")
	serveCmd.Flags().StringP("config", "c", "", "Config file path")
	serveCmd.Flags().String("db-url", "", "Database connection URL")
	serveCmd.Flags().String("dashboards", "./dashboards", "Dashboard directory")
	serveCmd.Flags().Bool("no-auth", false, "Disable authentication (development only)")
	serveCmd.Flags().Bool("auto-migrate", false, "Run database migrations on startup")
	serveCmd.Flags().Bool("enable-rls", false, "Enable Row Level Security for multi-tenancy")
}

func runServe(cmd *cobra.Command, _ []string) error {
	port, _ := cmd.Flags().GetInt("port")
	configPath, _ := cmd.Flags().GetString("config")
	dbURL, _ := cmd.Flags().GetString("db-url")
	dashboardDir, _ := cmd.Flags().GetString("dashboards")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	autoMigrate, _ := cmd.Flags().GetBool("auto-migrate")
	enableRLS, _ := cmd.Flags().GetBool("enable-rls")

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	cfg := server.Config{
		Port:         port,
		ConfigPath:   configPath,
		DatabaseURL:  dbURL,
		DashboardDir: dashboardDir,
		DisableAuth:  noAuth,
		AutoMigrate:  autoMigrate,
		EnableRLS:    enableRLS,
	}

	srv, err := server.New(cfg, logger)
	if err != nil {
		return fmt.Errorf("creating server: %w", err)
	}

	return srv.ListenAndServe()
}
