// Package cmd provides the CLI commands for Dashforge Server.
package cmd

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "dashforge-server",
	Short: "Dashforge server with database support",
	Long: `Dashforge Server is a full-featured dashboard server that supports:

  - Live database queries (PostgreSQL, MySQL, SQLite)
  - REST API for dashboard management
  - User authentication and permissions
  - Query caching and scheduling

For static-only dashboards, use the 'dashforge' CLI instead.`,
}

// Execute runs the root command.
func Execute() error {
	return rootCmd.Execute()
}
