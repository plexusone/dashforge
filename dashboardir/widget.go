package dashboardir

import "encoding/json"

// Widget is a visual component on the dashboard.
type Widget struct {
	// ID is the unique identifier for this widget.
	ID string `json:"id"`

	// Title is the display title shown above the widget.
	Title string `json:"title,omitempty"`

	// Description provides additional context.
	Description string `json:"description,omitempty"`

	// Type is the widget type: "chart", "table", "metric", "text", "image".
	Type string `json:"type"`

	// Position defines where the widget appears in the grid.
	Position Position `json:"position"`

	// DataSourceID references the data source for this widget.
	DataSourceID string `json:"dataSourceId,omitempty"`

	// Transform applies widget-specific data transformations.
	// Applied after the data source's transforms.
	Transform []Transform `json:"transform,omitempty"`

	// Config contains type-specific widget configuration.
	// For "chart": ChartIR from echartify
	// For "table": TableConfig
	// For "metric": MetricConfig
	// For "text": TextConfig
	// For "image": ImageConfig
	Config json.RawMessage `json:"config"`

	// Visible controls widget visibility.
	Visible *bool `json:"visible,omitempty"`

	// RefreshOverride overrides the data source refresh for this widget.
	RefreshOverride *RefreshConfig `json:"refreshOverride,omitempty"`

	// DrillDown defines drill-down behavior when clicking.
	DrillDown *DrillDown `json:"drillDown,omitempty"`
}

// Position defines widget placement in the grid layout.
type Position struct {
	// X is the column position (0-indexed).
	X int `json:"x"`

	// Y is the row position (0-indexed).
	Y int `json:"y"`

	// W is the width in grid columns.
	W int `json:"w"`

	// H is the height in grid rows.
	H int `json:"h"`

	// MinW is the minimum width (for responsive).
	MinW int `json:"minW,omitempty"`

	// MinH is the minimum height (for responsive).
	MinH int `json:"minH,omitempty"`
}

// DrillDown defines navigation when a widget element is clicked.
type DrillDown struct {
	// Type is the drill-down action: "dashboard", "url", "filter".
	Type string `json:"type"`

	// Target is the destination dashboard ID or URL.
	Target string `json:"target"`

	// Params maps clicked data fields to target parameters.
	Params map[string]string `json:"params,omitempty"`
}

// TableConfig configures a table widget.
type TableConfig struct {
	// Columns defines which data fields to display.
	Columns []TableColumn `json:"columns"`

	// Pagination enables table pagination.
	Pagination *TablePagination `json:"pagination,omitempty"`

	// Sortable enables column sorting.
	Sortable bool `json:"sortable,omitempty"`

	// Striped enables alternating row colors.
	Striped bool `json:"striped,omitempty"`

	// Compact uses reduced padding.
	Compact bool `json:"compact,omitempty"`
}

// TableColumn defines a table column.
type TableColumn struct {
	// Field is the data field to display.
	Field string `json:"field"`

	// Header is the column header text.
	Header string `json:"header,omitempty"`

	// Width is the column width (e.g., "100px", "20%").
	Width string `json:"width,omitempty"`

	// Align is the text alignment: "left", "center", "right".
	Align string `json:"align,omitempty"`

	// Format is the display format: "number", "percent", "currency", "date".
	Format string `json:"format,omitempty"`

	// FormatOptions contains format-specific options.
	FormatOptions *FormatOptions `json:"formatOptions,omitempty"`

	// Link makes the cell a clickable link.
	Link *ColumnLink `json:"link,omitempty"`
}

// FormatOptions provides format-specific settings.
type FormatOptions struct {
	// Decimals is the number of decimal places (for numbers).
	Decimals int `json:"decimals,omitempty"`

	// Prefix is text before the value (e.g., "$").
	Prefix string `json:"prefix,omitempty"`

	// Suffix is text after the value (e.g., "%").
	Suffix string `json:"suffix,omitempty"`

	// DateFormat is the date format string.
	DateFormat string `json:"dateFormat,omitempty"`

	// Locale is the locale for number formatting.
	Locale string `json:"locale,omitempty"`
}

// ColumnLink makes a table cell clickable.
type ColumnLink struct {
	// URLField is the data field containing the URL.
	URLField string `json:"urlField,omitempty"`

	// URLTemplate is a URL template with {field} placeholders.
	URLTemplate string `json:"urlTemplate,omitempty"`

	// External opens in a new tab.
	External bool `json:"external,omitempty"`
}

// TablePagination configures table pagination.
type TablePagination struct {
	// Enabled toggles pagination.
	Enabled bool `json:"enabled"`

	// PageSize is rows per page.
	PageSize int `json:"pageSize,omitempty"`

	// PageSizeOptions are selectable page sizes.
	PageSizeOptions []int `json:"pageSizeOptions,omitempty"`
}

// MetricConfig configures a single-value metric widget.
type MetricConfig struct {
	// ValueField is the data field to display.
	ValueField string `json:"valueField"`

	// LabelField is an optional label field.
	LabelField string `json:"labelField,omitempty"`

	// Format is the display format: "number", "percent", "currency".
	Format string `json:"format,omitempty"`

	// FormatOptions contains format settings.
	FormatOptions *FormatOptions `json:"formatOptions,omitempty"`

	// Comparison shows change vs previous value.
	Comparison *MetricComparison `json:"comparison,omitempty"`

	// Thresholds define color ranges.
	Thresholds []MetricThreshold `json:"thresholds,omitempty"`

	// Icon is an optional icon name.
	Icon string `json:"icon,omitempty"`

	// Sparkline shows a mini trend chart.
	Sparkline *SparklineConfig `json:"sparkline,omitempty"`
}

// MetricComparison shows comparison to previous value.
type MetricComparison struct {
	// Enabled toggles comparison display.
	Enabled bool `json:"enabled"`

	// Field is the comparison value field.
	Field string `json:"field,omitempty"`

	// Mode is "absolute" or "percent".
	Mode string `json:"mode,omitempty"`

	// InvertColors swaps green/red meanings.
	InvertColors bool `json:"invertColors,omitempty"`
}

// MetricThreshold defines a color threshold.
type MetricThreshold struct {
	// Value is the threshold value.
	Value float64 `json:"value"`

	// Color is the color when value exceeds threshold.
	Color string `json:"color"`
}

// SparklineConfig configures a mini trend chart.
type SparklineConfig struct {
	// Enabled toggles sparkline display.
	Enabled bool `json:"enabled"`

	// Field is the data field for sparkline values.
	Field string `json:"field"`

	// Color is the sparkline color.
	Color string `json:"color,omitempty"`

	// Height is the sparkline height in pixels.
	Height int `json:"height,omitempty"`
}

// TextConfig configures a text/markdown widget.
type TextConfig struct {
	// Content is the text or markdown content.
	Content string `json:"content"`

	// Format is "text", "markdown", or "html".
	Format string `json:"format,omitempty"`

	// Variables can be interpolated: "Total repos: {{summary.totalRepos}}"
	Variables bool `json:"variables,omitempty"`
}

// ImageConfig configures an image widget.
type ImageConfig struct {
	// URL is the image source.
	URL string `json:"url"`

	// Alt is the alt text.
	Alt string `json:"alt,omitempty"`

	// Fit is the object-fit: "contain", "cover", "fill".
	Fit string `json:"fit,omitempty"`
}

// Widget type constants.
const (
	WidgetTypeChart  = "chart"
	WidgetTypeTable  = "table"
	WidgetTypeMetric = "metric"
	WidgetTypeText   = "text"
	WidgetTypeImage  = "image"
)

// DrillDown type constants.
const (
	DrillDownTypeDashboard = "dashboard"
	DrillDownTypeURL       = "url"
	DrillDownTypeFilter    = "filter"
)

// Text format constants.
const (
	TextFormatPlain    = "text"
	TextFormatMarkdown = "markdown"
	TextFormatHTML     = "html"
)

// Image fit constants.
const (
	ImageFitContain = "contain"
	ImageFitCover   = "cover"
	ImageFitFill    = "fill"
)
