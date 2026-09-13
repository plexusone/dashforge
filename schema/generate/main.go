//go:build ignore

// Package main generates JSON Schema from Go structs.
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/invopop/jsonschema"

	"github.com/plexusone/dashforge/dashboardir"
)

type schemaTarget struct {
	Type        any
	ID          string
	Title       string
	Description string
	OutputFile  string
}

func main() {
	targets := []schemaTarget{
		{
			Type:        &dashboardir.Dashboard{},
			ID:          "https://github.com/plexusone/dashforge/schema/dashboard.schema.json",
			Title:       "Dashboard",
			Description: "DashForge dashboard definition (DashboardIR)",
			OutputFile:  "schema/dashboard.schema.json",
		},
	}

	r := &jsonschema.Reflector{
		RequiredFromJSONSchemaTags: true,
	}

	for _, t := range targets {
		schema := r.Reflect(t.Type)
		schema.Version = "https://json-schema.org/draft/2020-12/schema"
		schema.ID = jsonschema.ID(t.ID)
		schema.Title = t.Title
		schema.Description = t.Description

		data, err := json.MarshalIndent(schema, "", "  ")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error marshaling %s: %v\n", t.Title, err)
			os.Exit(1)
		}

		dir := filepath.Dir(t.OutputFile)
		if err := os.MkdirAll(dir, 0755); err != nil {
			fmt.Fprintf(os.Stderr, "Error creating directory %s: %v\n", dir, err)
			os.Exit(1)
		}

		if err := os.WriteFile(t.OutputFile, data, 0644); err != nil {
			fmt.Fprintf(os.Stderr, "Error writing %s: %v\n", t.OutputFile, err)
			os.Exit(1)
		}

		fmt.Printf("Schema written to %s\n", t.OutputFile)
	}
}
