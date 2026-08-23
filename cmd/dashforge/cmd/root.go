// Package cmd provides the CLI commands for Dashforge.
package cmd

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "dashforge",
	Short: "Dashboard framework for static sites and databases",
	Long: `Dashforge is a JSON-first dashboard framework that starts simple
with static hosting (GitHub Pages) and grows into a full analytics platform.

Features:
  - JSON IR for dashboard definitions
  - Static viewer for GitHub Pages
  - Local development server with hot reload
  - PostgreSQL data source support (coming soon)
  - ChartIR integration via echartify`,
}

// Execute runs the root command.
func Execute() error {
	return rootCmd.Execute()
}
