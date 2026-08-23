// Package main is the entry point for the dashforge CLI.
package main

import (
	"os"

	"github.com/plexusone/dashforge/cmd/dashforge/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
