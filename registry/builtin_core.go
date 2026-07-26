package registry

import "encoding/json"

// RegisterCoreComponents adds the core namespace components to the registry:
// card, text, tabs, button, modal.
func RegisterCoreComponents(r *Registry) error {
	specs := []*ComponentSpec{
		{
			ID:       "core.card",
			Version:  "1.0.0",
			Category: "layout",
			Runtime:  "react",
			PropertiesSchema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"title":    {"type": "string"},
					"subtitle": {"type": "string"},
					"padding":  {"type": "string"},
					"elevated": {"type": "boolean"}
				}
			}`),
			Events: map[string]EventDef{
				"click": {Description: "Card body clicked"},
			},
			LayoutConstraints: &LayoutConstraints{MinWidth: "200px"},
			DesignSystem: &DesignSystemRef{
				Tokens:   []string{"color-surface", "color-border", "radius-md", "shadow-sm"},
				Variants: []string{"default", "outlined", "elevated"},
			},
		},
		{
			ID:       "core.text",
			Version:  "1.0.0",
			Category: "content",
			Runtime:  "react",
			PropertiesSchema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"content":  {"type": "string"},
					"variant":  {"type": "string", "enum": ["body","heading","caption","code"]},
					"markdown": {"type": "boolean"}
				}
			}`),
			DesignSystem: &DesignSystemRef{
				Tokens:   []string{"font-family", "font-size-base", "color-text"},
				Variants: []string{"body", "heading", "caption", "code"},
			},
		},
		{
			ID:       "core.tabs",
			Version:  "1.0.0",
			Category: "navigation",
			Runtime:  "react",
			PropertiesSchema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"defaultTab": {"type": "string"},
					"variant":    {"type": "string", "enum": ["line","enclosed","pill"]}
				}
			}`),
			Events: map[string]EventDef{
				"change": {Description: "Active tab changed", Schema: json.RawMessage(`{"type":"object","properties":{"tab":{"type":"string"}}}`)},
			},
			DesignSystem: &DesignSystemRef{
				Tokens:   []string{"color-accent", "color-border"},
				Variants: []string{"line", "enclosed", "pill"},
			},
		},
		{
			ID:       "core.button",
			Version:  "1.0.0",
			Category: "action",
			Runtime:  "react",
			PropertiesSchema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"label":    {"type": "string"},
					"variant":  {"type": "string", "enum": ["primary","secondary","ghost","danger"]},
					"size":     {"type": "string", "enum": ["sm","md","lg"]},
					"disabled": {"type": "boolean"},
					"icon":     {"type": "string"}
				},
				"required": ["label"]
			}`),
			Events: map[string]EventDef{
				"click": {Description: "Button clicked"},
			},
			DesignSystem: &DesignSystemRef{
				Tokens:   []string{"color-accent", "color-danger", "radius-sm"},
				Variants: []string{"primary", "secondary", "ghost", "danger"},
			},
		},
		{
			ID:       "core.modal",
			Version:  "1.0.0",
			Category: "overlay",
			Runtime:  "react",
			PropertiesSchema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"title":      {"type": "string"},
					"size":       {"type": "string", "enum": ["sm","md","lg","full"]},
					"closable":   {"type": "boolean"},
					"overlayDim": {"type": "boolean"}
				}
			}`),
			Events: map[string]EventDef{
				"close":   {Description: "Modal closed"},
				"confirm": {Description: "Confirm action triggered"},
			},
			Actions: []string{"open", "close"},
			DesignSystem: &DesignSystemRef{
				Tokens:   []string{"color-surface", "shadow-lg", "radius-lg"},
				Variants: []string{"sm", "md", "lg", "full"},
			},
		},
	}

	for _, s := range specs {
		if err := r.Register(s); err != nil {
			return err
		}
	}
	return nil
}
