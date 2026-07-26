package registry

import "encoding/json"

// RegisterAnalyticsComponents adds the analytics namespace components to the registry:
// line-chart, bar-chart, metric, table, filter, gauge.
func RegisterAnalyticsComponents(r *Registry) error {
	specs := []*ComponentSpec{
		{
			ID:       "analytics.line-chart",
			Version:  "1.0.0",
			Category: "visualization",
			Runtime:  "react",
			PropertiesSchema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"title":      {"type": "string"},
					"xAxis":      {"type": "string"},
					"yAxis":      {"type": "string"},
					"showLegend": {"type": "boolean"},
					"stacked":    {"type": "boolean"},
					"smooth":     {"type": "boolean"}
				}
			}`),
			DataInputs: map[string]DataInput{
				"primary": {Type: "timeseries", Description: "Time-series data to plot", Required: true},
			},
			Events: map[string]EventDef{
				"click":     {Description: "Data point clicked"},
				"brushEnd":  {Description: "Selection brush completed"},
				"zoomReset": {Description: "Zoom reset to default"},
			},
			LayoutConstraints: &LayoutConstraints{MinWidth: "300px", MinHeight: "200px"},
			DesignSystem: &DesignSystemRef{
				Tokens: []string{"color-accent", "color-series-1", "color-series-2", "color-series-3"},
			},
		},
		{
			ID:       "analytics.bar-chart",
			Version:  "1.0.0",
			Category: "visualization",
			Runtime:  "react",
			PropertiesSchema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"title":      {"type": "string"},
					"orientation": {"type": "string", "enum": ["vertical","horizontal"]},
					"showLegend": {"type": "boolean"},
					"stacked":    {"type": "boolean"}
				}
			}`),
			DataInputs: map[string]DataInput{
				"primary": {Type: "categorical", Description: "Categorical data for bars", Required: true},
			},
			Events: map[string]EventDef{
				"click": {Description: "Bar clicked"},
			},
			LayoutConstraints: &LayoutConstraints{MinWidth: "250px", MinHeight: "200px"},
			DesignSystem: &DesignSystemRef{
				Tokens: []string{"color-accent", "color-series-1", "color-series-2", "color-series-3"},
			},
		},
		{
			ID:       "analytics.metric",
			Version:  "1.0.0",
			Category: "visualization",
			Runtime:  "react",
			PropertiesSchema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"title":        {"type": "string"},
					"valueField":   {"type": "string"},
					"format":       {"type": "string"},
					"prefix":       {"type": "string"},
					"suffix":       {"type": "string"},
					"trend":        {"type": "boolean"},
					"sparkline":    {"type": "boolean"}
				}
			}`),
			DataInputs: map[string]DataInput{
				"primary": {Type: "scalar", Description: "Scalar metric value", Required: true},
			},
			Events: map[string]EventDef{
				"click": {Description: "Metric card clicked"},
			},
			LayoutConstraints: &LayoutConstraints{MinWidth: "140px", MinHeight: "80px"},
			DesignSystem: &DesignSystemRef{
				Tokens:   []string{"color-success", "color-danger", "font-size-lg"},
				Variants: []string{"default", "compact", "sparkline"},
			},
		},
		{
			ID:       "analytics.table",
			Version:  "1.0.0",
			Category: "visualization",
			Runtime:  "react",
			PropertiesSchema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"title":      {"type": "string"},
					"columns":    {"type": "array"},
					"sortable":   {"type": "boolean"},
					"filterable": {"type": "boolean"},
					"pageSize":   {"type": "integer"}
				}
			}`),
			DataInputs: map[string]DataInput{
				"primary": {Type: "tabular", Description: "Tabular row data", Required: true},
			},
			Events: map[string]EventDef{
				"rowClick": {Description: "Table row clicked"},
				"sort":     {Description: "Column sort changed"},
				"filter":   {Description: "Column filter applied"},
			},
			LayoutConstraints: &LayoutConstraints{MinWidth: "400px"},
			DesignSystem: &DesignSystemRef{
				Tokens: []string{"color-border", "color-surface"},
			},
		},
		{
			ID:       "analytics.filter",
			Version:  "1.0.0",
			Category: "control",
			Runtime:  "react",
			PropertiesSchema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"label":       {"type": "string"},
					"field":       {"type": "string"},
					"filterType":  {"type": "string", "enum": ["select","multi-select","date-range","search"]},
					"options":     {"type": "array"},
					"placeholder": {"type": "string"}
				},
				"required": ["field","filterType"]
			}`),
			Events: map[string]EventDef{
				"change": {Description: "Filter value changed"},
				"clear":  {Description: "Filter cleared"},
			},
			DesignSystem: &DesignSystemRef{
				Tokens: []string{"color-accent", "color-border"},
			},
		},
		{
			ID:       "analytics.gauge",
			Version:  "1.0.0",
			Category: "visualization",
			Runtime:  "react",
			PropertiesSchema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"title":    {"type": "string"},
					"min":      {"type": "number"},
					"max":      {"type": "number"},
					"thresholds": {"type": "array"},
					"format":   {"type": "string"},
					"size":     {"type": "string", "enum": ["sm","md","lg"]}
				}
			}`),
			DataInputs: map[string]DataInput{
				"primary": {Type: "scalar", Description: "Current gauge value", Required: true},
			},
			LayoutConstraints: &LayoutConstraints{MinWidth: "120px", MinHeight: "120px"},
			DesignSystem: &DesignSystemRef{
				Tokens: []string{"color-success", "color-warning", "color-danger"},
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
