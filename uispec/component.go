package uispec

import "encoding/json"

// ComponentInstance represents a placed component on a page.
type ComponentInstance struct {
	ID         string              `json:"id"`
	Type       string              `json:"type"`
	Version    string              `json:"version,omitempty"`
	Position   *Position           `json:"position,omitempty"`
	Properties map[string]any      `json:"properties,omitempty"`
	Data       map[string]Binding  `json:"data,omitempty"`
	Children   []ComponentInstance `json:"children,omitempty"`
	Visibility *VisibilityRule     `json:"visibility,omitempty"`
	Slot       string              `json:"slot,omitempty"`
	Style      map[string]string   `json:"style,omitempty"`
	RawConfig  json.RawMessage     `json:"rawConfig,omitempty"`
}

// Position places a component within a grid or flex layout.
type Position struct {
	Row     int    `json:"row,omitempty"`
	Col     int    `json:"col,omitempty"`
	RowSpan int    `json:"rowSpan,omitempty"`
	ColSpan int    `json:"colSpan,omitempty"`
	Order   int    `json:"order,omitempty"`
	Region  string `json:"region,omitempty"`
}

// VisibilityRule controls whether a component is rendered.
type VisibilityRule struct {
	Condition  string   `json:"condition"`
	Roles      []string `json:"roles,omitempty"`
	Capability string   `json:"capability,omitempty"`
}
