// Package dashboardir provides the Dashboard Intermediate Representation (IR)
// for defining dashboards in a non-polymorphic, AI-friendly JSON format.
package dashboardir

// Dashboard is the top-level container for a dashboard definition.
type Dashboard struct {
	// Schema is the JSON Schema version for validation.
	Schema string `json:"$schema,omitempty"`

	// ID is the unique identifier for this dashboard.
	ID string `json:"id"`

	// Title is the display title for the dashboard.
	Title string `json:"title"`

	// Description provides additional context about the dashboard.
	Description string `json:"description,omitempty"`

	// Version tracks dashboard definition changes.
	Version string `json:"version,omitempty"`

	// Layout defines how widgets are arranged.
	Layout Layout `json:"layout"`

	// DataSources defines where data comes from.
	DataSources []DataSource `json:"dataSources"`

	// Widgets are the visual components on the dashboard.
	Widgets []Widget `json:"widgets"`

	// Theme customizes the visual appearance.
	Theme *Theme `json:"theme,omitempty"`

	// Variables define user-configurable parameters.
	Variables []Variable `json:"variables,omitempty"`
}

// Layout defines the dashboard layout system.
type Layout struct {
	// Type is the layout system: "grid" or "flex".
	Type string `json:"type"`

	// Columns is the number of grid columns (for grid layout).
	Columns int `json:"columns,omitempty"`

	// RowHeight is the height of each grid row in pixels (for grid layout).
	RowHeight int `json:"rowHeight,omitempty"`

	// Gap is the spacing between widgets in pixels.
	Gap int `json:"gap,omitempty"`

	// Padding is the dashboard edge padding in pixels.
	Padding int `json:"padding,omitempty"`
}

// Theme defines visual styling for the dashboard.
type Theme struct {
	// Mode is "light" or "dark".
	Mode string `json:"mode,omitempty"`

	// PrimaryColor is the main accent color (hex).
	PrimaryColor string `json:"primaryColor,omitempty"`

	// BackgroundColor is the dashboard background (hex).
	BackgroundColor string `json:"backgroundColor,omitempty"`

	// FontFamily is the default font.
	FontFamily string `json:"fontFamily,omitempty"`
}

// Variable defines a user-configurable parameter.
type Variable struct {
	// ID is the unique identifier for this variable.
	ID string `json:"id"`

	// Name is the display name.
	Name string `json:"name"`

	// Type is the variable type: "select", "text", "date", "daterange".
	Type string `json:"type"`

	// Default is the default value.
	Default string `json:"default,omitempty"`

	// Options are the available choices (for select type).
	Options []VariableOption `json:"options,omitempty"`

	// DataSourceID references a data source for dynamic options.
	DataSourceID string `json:"dataSourceId,omitempty"`

	// LabelField is the field to display (for dynamic options).
	LabelField string `json:"labelField,omitempty"`

	// ValueField is the field to use as value (for dynamic options).
	ValueField string `json:"valueField,omitempty"`
}

// VariableOption is a single option for a select variable.
type VariableOption struct {
	Label string `json:"label"`
	Value string `json:"value"`
}

// Layout type constants.
const (
	LayoutTypeGrid = "grid"
	LayoutTypeFlex = "flex"
)

// Theme mode constants.
const (
	ThemeModeLight = "light"
	ThemeModeDark  = "dark"
)

// Variable type constants.
const (
	VariableTypeSelect    = "select"
	VariableTypeText      = "text"
	VariableTypeDate      = "date"
	VariableTypeDateRange = "daterange"
)
