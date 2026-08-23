package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/plexusone/dashforge/dashboardir"
)

var validateCmd = &cobra.Command{
	Use:   "validate <file>",
	Short: "Validate a dashboard JSON file",
	Long: `Validate a dashboard JSON file against the DashboardIR schema.

Example:
  dashforge validate dashboard.json
  dashforge validate ./dashboards/*.json`,
	Args: cobra.MinimumNArgs(1),
	RunE: runValidate,
}

func init() {
	rootCmd.AddCommand(validateCmd)
}

func runValidate(_ *cobra.Command, args []string) error {
	hasErrors := false

	for _, file := range args {
		if err := validateFile(file); err != nil {
			fmt.Fprintf(os.Stderr, "❌ %s: %v\n", file, err)
			hasErrors = true
		} else {
			fmt.Printf("✅ %s: valid\n", file)
		}
	}

	if hasErrors {
		return fmt.Errorf("validation failed")
	}
	return nil
}

func validateFile(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("reading file: %w", err)
	}

	var dashboard dashboardir.Dashboard
	if err := json.Unmarshal(data, &dashboard); err != nil {
		return fmt.Errorf("parsing JSON: %w", err)
	}

	// Basic validation
	if dashboard.ID == "" {
		return fmt.Errorf("missing required field: id")
	}
	if dashboard.Title == "" {
		return fmt.Errorf("missing required field: title")
	}
	if len(dashboard.Widgets) == 0 {
		return fmt.Errorf("dashboard has no widgets")
	}

	// Validate widget references
	dsIDs := make(map[string]bool)
	for _, ds := range dashboard.DataSources {
		dsIDs[ds.ID] = true
	}

	for _, widget := range dashboard.Widgets {
		if widget.ID == "" {
			return fmt.Errorf("widget missing id")
		}
		if widget.DataSourceID != "" && !dsIDs[widget.DataSourceID] {
			return fmt.Errorf("widget %q references unknown data source %q", widget.ID, widget.DataSourceID)
		}
	}

	return nil
}
