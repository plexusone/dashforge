// Package schema provides embedded JSON schemas for validation.
package schema

import (
	_ "embed"
)

// DashboardSchema is the JSON Schema for Dashboard, embedded at compile time.
//
//go:embed dashboard.schema.json
var DashboardSchema []byte

// PageSchema is the JSON Schema for PageSpec.
//
//go:embed page.schema.json
var PageSchema []byte

// ComponentSchema is the JSON Schema for ComponentSpec.
//
//go:embed component.schema.json
var ComponentSchema []byte
