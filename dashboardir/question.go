package dashboardir

import (
	"encoding/json"
	"time"
)

// SavedQuestion is UIForge metadata for a reusable analytical question. A
// question can later be placed on zero, one, or many dashboards.
type SavedQuestion struct {
	ID            string         `json:"id"`
	Name          string         `json:"name"`
	Description   string         `json:"description,omitempty"`
	SourceID      string         `json:"sourceId"`
	DatasetID     string         `json:"datasetId"`
	Dialect       string         `json:"dialect,omitempty"`
	Query         string         `json:"query"`
	CompiledQuery *CompiledQuery `json:"compiledQuery,omitempty"`
	Visualization map[string]any `json:"visualization,omitempty"`
	CreatedAt     time.Time      `json:"createdAt"`
	UpdatedAt     time.Time      `json:"updatedAt"`
}

// CompiledQuery captures the parsed, canonical, read-only query metadata used
// for validation, cache keys, and later execution planning. Query text remains
// the source of truth.
type CompiledQuery struct {
	Version     string          `json:"version"`
	AST         json.RawMessage `json:"ast,omitempty"`
	Fingerprint string          `json:"fingerprint"`
	Datasets    []string        `json:"datasets,omitempty"`
	Fields      []string        `json:"fields,omitempty"`
	ReadOnly    bool            `json:"readOnly"`
	Limit       int             `json:"limit,omitempty"`
}
