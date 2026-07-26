// Dashforge Server - Full-featured dashboard server with database support.
package main

import (
	"os"

	"github.com/plexusone/uiforge/cmd/uiforge-server/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
