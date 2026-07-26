package uispec

// Binding connects a component property to a data source.
type Binding struct {
	Source     string         `json:"source"`
	Operation  string         `json:"operation,omitempty"`
	Parameters map[string]any `json:"parameters,omitempty"`
	Transform  string         `json:"transform,omitempty"`
	Default    any            `json:"default,omitempty"`
}
