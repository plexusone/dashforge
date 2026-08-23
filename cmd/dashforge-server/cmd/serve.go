package cmd

import (
	"fmt"
	"log/slog"
	"net"
	"os"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/plexusone/dashforge/internal/server"
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
  dashforge-server serve --port 8080 --config config.yaml
  dashforge-server serve --db-url postgres://localhost/analytics`,
	RunE: runServe,
}

func init() {
	rootCmd.AddCommand(serveCmd)

	serveCmd.Flags().String("address", "", "Listen address as host:port (overrides --host and --port)")
	serveCmd.Flags().String("host", "127.0.0.1", "Host to listen on")
	serveCmd.Flags().IntP("port", "p", 8080, "Port to listen on")
	serveCmd.Flags().StringP("config", "c", "", "Config file path")
	serveCmd.Flags().String("db-url", "", "Database connection URL")
	serveCmd.Flags().String("dashboards", "examples", "Dashboard directory")
	serveCmd.Flags().String("question-store", ".dashforge/questions.json", "Saved question metadata JSON file")
	serveCmd.Flags().Bool("no-auth", false, "Disable authentication (development only)")
	serveCmd.Flags().Bool("auto-migrate", false, "Run database migrations on startup")
	serveCmd.Flags().Bool("enable-rls", false, "Enable Row Level Security for multi-tenancy")
	serveCmd.Flags().String("analytics-source-store", "", "Analytics source config JSON file when no metadata DB is set (default .dashforge/analytics-sources.json)")
	serveCmd.Flags().Bool("enable-ollama", false, "Enable local Ollama as an AI provider")
	serveCmd.Flags().String("ollama-url", "", "Ollama base URL, e.g. http://localhost:11434")
}

func runServe(cmd *cobra.Command, _ []string) error {
	address, _ := cmd.Flags().GetString("address")
	host, _ := cmd.Flags().GetString("host")
	port, _ := cmd.Flags().GetInt("port")
	configPath, _ := cmd.Flags().GetString("config")
	dbURL, _ := cmd.Flags().GetString("db-url")
	dashboardDir, _ := cmd.Flags().GetString("dashboards")
	questionStore, _ := cmd.Flags().GetString("question-store")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	autoMigrate, _ := cmd.Flags().GetBool("auto-migrate")
	enableRLS, _ := cmd.Flags().GetBool("enable-rls")
	analyticsSourceStore, _ := cmd.Flags().GetString("analytics-source-store")
	enableOllama, _ := cmd.Flags().GetBool("enable-ollama")
	ollamaURL, _ := cmd.Flags().GetString("ollama-url")

	if address != "" {
		parsedHost, parsedPort, err := parseAddress(address)
		if err != nil {
			return err
		}
		host = parsedHost
		port = parsedPort
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	cfg := server.Config{
		Host:                     host,
		Port:                     port,
		ConfigPath:               configPath,
		DatabaseURL:              dbURL,
		DashboardDir:             dashboardDir,
		QuestionStorePath:        questionStore,
		DisableAuth:              noAuth,
		AutoMigrate:              autoMigrate,
		EnableRLS:                enableRLS,
		AnalyticsSourceStorePath: analyticsSourceStore,
		EnableOllama:             enableOllama,
		OllamaBaseURL:            ollamaURL,
	}

	srv, err := server.New(cfg, logger)
	if err != nil {
		return fmt.Errorf("creating server: %w", err)
	}

	return srv.ListenAndServe()
}

func parseAddress(address string) (string, int, error) {
	host, portString, err := net.SplitHostPort(address)
	if err != nil {
		return "", 0, fmt.Errorf("invalid address %q: %w", address, err)
	}
	if host == "" {
		host = "127.0.0.1"
	}
	port, err := strconv.Atoi(portString)
	if err != nil || port <= 0 {
		return "", 0, fmt.Errorf("invalid address %q: invalid port", address)
	}
	return host, port, nil
}
