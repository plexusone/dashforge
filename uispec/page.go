// Package uispec defines the canonical UISpec type system — the JSON IR
// for declarative UI composition. Go structs are the source of truth;
// JSON Schemas are generated from these types.
package uispec

const (
	APIVersion = "ui.plexusone.dev/v1"
	KindPage   = "Page"
)

// PageSpec is the top-level document describing a single composable page.
type PageSpec struct {
	APIVersion   string              `json:"apiVersion"`
	Kind         string              `json:"kind"`
	Metadata     PageMetadata        `json:"metadata"`
	Profile      string              `json:"profile,omitempty"`
	Context      map[string]string   `json:"context,omitempty"`
	Layout       LayoutSpec          `json:"layout"`
	Components   []ComponentInstance `json:"components"`
	Interactions []Interaction       `json:"interactions,omitempty"`
	Navigation   *NavigationSpec     `json:"navigation,omitempty"`
	Theme        *ThemeRef           `json:"theme,omitempty"`
}

// PageMetadata holds identity and descriptive fields.
type PageMetadata struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Title       string            `json:"title,omitempty"`
	Description string            `json:"description,omitempty"`
	Version     string            `json:"version,omitempty"`
	Labels      map[string]string `json:"labels,omitempty"`
}
