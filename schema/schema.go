// Package schema provides embedded JSON schemas for validation.
package schema

import (
	_ "embed"
)

// DashboardSchema is the JSON Schema for Dashboard, embedded at compile time.
//
//go:embed dashboard.schema.json
var DashboardSchema []byte
