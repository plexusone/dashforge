package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// Version information (set via ldflags)
var (
	Version   = "dev"
	GitCommit = "unknown"
	BuildDate = "unknown"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	Run: func(_ *cobra.Command, _ []string) {
		fmt.Printf("uiforge %s\n", Version)
		fmt.Printf("  commit: %s\n", GitCommit)
		fmt.Printf("  built:  %s\n", BuildDate)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
