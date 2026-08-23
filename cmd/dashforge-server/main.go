package main

import (
	"os"

	"github.com/plexusone/dashforge/cmd/dashforge-server/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
