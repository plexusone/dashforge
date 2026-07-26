// Package main is the entry point for the uiforge CLI.
package main

import (
	"os"

	"github.com/plexusone/uiforge/cmd/uiforge/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
