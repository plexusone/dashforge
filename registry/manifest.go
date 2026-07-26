// Package registry provides a component registry with manifest validation.
// ComponentSpec manifests describe what a component accepts (properties,
// data inputs, events, actions) and what constraints it has.
package registry

import "encoding/json"

// ComponentSpec is the manifest for a registered component.
type ComponentSpec struct {
	ID                string                  `json:"id"`
	Version           string                  `json:"version"`
	Category          string                  `json:"category"`
	Namespace         string                  `json:"namespace"`
	Runtime           string                  `json:"runtime"`
	Entrypoint        string                  `json:"entrypoint"`
	PropertiesSchema  json.RawMessage         `json:"propertiesSchema,omitempty"`
	DataInputs        map[string]DataInput    `json:"dataInputs,omitempty"`
	Events            map[string]EventDef     `json:"events,omitempty"`
	Actions           []string                `json:"actions,omitempty"`
	LayoutConstraints *LayoutConstraints      `json:"layoutConstraints,omitempty"`
	Capabilities      []string                `json:"capabilities,omitempty"`
	DesignSystem      *DesignSystemRef        `json:"designSystem,omitempty"`
}

// DataInput describes one named data slot a component can bind to.
type DataInput struct {
	Type        string `json:"type"`
	Description string `json:"description,omitempty"`
	Required    bool   `json:"required,omitempty"`
}

// EventDef describes an event a component can emit.
type EventDef struct {
	Description string          `json:"description,omitempty"`
	Schema      json.RawMessage `json:"schema,omitempty"`
}

// LayoutConstraints restrict where and how a component may be placed.
type LayoutConstraints struct {
	MinWidth  string   `json:"minWidth,omitempty"`
	MinHeight string   `json:"minHeight,omitempty"`
	MaxWidth  string   `json:"maxWidth,omitempty"`
	MaxHeight string   `json:"maxHeight,omitempty"`
	AllowedIn []string `json:"allowedIn,omitempty"`
}

// DesignSystemRef links a component to design tokens.
type DesignSystemRef struct {
	Tokens   []string `json:"tokens,omitempty"`
	Variants []string `json:"variants,omitempty"`
}
