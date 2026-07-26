package uispec

const (
	LayoutResponsiveGrid  = "responsive-grid"
	LayoutStack           = "stack"
	LayoutSplitPane       = "split-pane"
	LayoutTabs            = "tabs"
	LayoutApplicationShell = "application-shell"
)

// LayoutSpec defines how components are arranged on a page.
type LayoutSpec struct {
	Type    string           `json:"type"`
	Config  *LayoutConfig    `json:"config,omitempty"`
	Regions []LayoutRegion   `json:"regions,omitempty"`
}

// LayoutConfig holds type-specific layout parameters.
type LayoutConfig struct {
	Columns     int               `json:"columns,omitempty"`
	Rows        int               `json:"rows,omitempty"`
	Gap         string            `json:"gap,omitempty"`
	Direction   string            `json:"direction,omitempty"`
	Breakpoints map[string]BreakpointConfig `json:"breakpoints,omitempty"`
	Sizes       []string          `json:"sizes,omitempty"`
	Resizable   bool              `json:"resizable,omitempty"`
}

// BreakpointConfig adjusts layout at different viewport widths.
type BreakpointConfig struct {
	MinWidth int `json:"minWidth"`
	Columns  int `json:"columns,omitempty"`
	Gap      string `json:"gap,omitempty"`
}

// LayoutRegion defines a named slot within a layout.
type LayoutRegion struct {
	Name     string      `json:"name"`
	Layout   *LayoutSpec `json:"layout,omitempty"`
	MinWidth string      `json:"minWidth,omitempty"`
	MaxWidth string      `json:"maxWidth,omitempty"`
	Default  string      `json:"default,omitempty"`
}
