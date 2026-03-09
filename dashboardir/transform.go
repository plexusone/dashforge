package dashboardir

import "encoding/json"

// Transform defines a data transformation operation.
// Transformations are applied in order to shape data for visualization.
type Transform struct {
	// Type is the transformation type.
	Type string `json:"type"`

	// Config contains type-specific configuration.
	// Using json.RawMessage for flexibility while keeping IR non-polymorphic.
	Config json.RawMessage `json:"config,omitempty"`
}

// ExtractConfig extracts a nested path from the data.
type ExtractConfig struct {
	// Path is the dot-notation path to extract: "summary.byLanguage"
	Path string `json:"path"`
}

// FilterConfig filters rows based on conditions.
type FilterConfig struct {
	// Field is the field to filter on.
	Field string `json:"field"`

	// Operator is the comparison: "eq", "ne", "gt", "lt", "gte", "lte", "contains", "in", "between".
	Operator string `json:"operator"`

	// Value is the comparison value. Can reference variables with "${var:variableId}".
	Value string `json:"value"`

	// Values is for "in" operator (multiple values).
	Values []string `json:"values,omitempty"`

	// Variable indicates this filter uses a dashboard variable (for documentation/tooling).
	Variable string `json:"variable,omitempty"`
}

// AggregateConfig groups and aggregates data.
type AggregateConfig struct {
	// GroupBy is the field(s) to group by.
	GroupBy []string `json:"groupBy"`

	// Aggregations defines the aggregation operations.
	Aggregations []Aggregation `json:"aggregations"`
}

// Aggregation defines a single aggregation operation.
type Aggregation struct {
	// Field is the field to aggregate.
	Field string `json:"field"`

	// Function is the aggregation: "sum", "avg", "min", "max", "count".
	Function string `json:"function"`

	// As is the output field name.
	As string `json:"as"`
}

// SortConfig sorts data by field(s).
type SortConfig struct {
	// Field is the field to sort by.
	Field string `json:"field"`

	// Direction is "asc" or "desc".
	Direction string `json:"direction"`
}

// LimitConfig limits the number of rows.
type LimitConfig struct {
	// Count is the maximum number of rows.
	Count int `json:"count"`

	// Offset skips the first N rows.
	Offset int `json:"offset,omitempty"`
}

// SelectConfig selects specific fields.
type SelectConfig struct {
	// Fields is the list of fields to include.
	Fields []string `json:"fields"`
}

// RenameConfig renames fields.
type RenameConfig struct {
	// Mapping is old name -> new name.
	Mapping map[string]string `json:"mapping"`
}

// ComputeConfig adds computed fields.
type ComputeConfig struct {
	// Field is the new field name.
	Field string `json:"field"`

	// Expression is the computation expression.
	// Simple expressions: "field1 + field2", "field1 * 100"
	Expression string `json:"expression"`
}

// Transform type constants.
const (
	TransformTypeExtract   = "extract"
	TransformTypeFilter    = "filter"
	TransformTypeAggregate = "aggregate"
	TransformTypeSort      = "sort"
	TransformTypeLimit     = "limit"
	TransformTypeSelect    = "select"
	TransformTypeRename    = "rename"
	TransformTypeCompute   = "compute"
)

// Filter operator constants.
const (
	FilterOpEqual       = "eq"
	FilterOpNotEqual    = "ne"
	FilterOpGreaterThan = "gt"
	FilterOpLessThan    = "lt"
	FilterOpGTE         = "gte"
	FilterOpLTE         = "lte"
	FilterOpContains    = "contains"
	FilterOpIn          = "in"
	FilterOpBetween     = "between"
)

// Aggregation function constants.
const (
	AggFuncSum   = "sum"
	AggFuncAvg   = "avg"
	AggFuncMin   = "min"
	AggFuncMax   = "max"
	AggFuncCount = "count"
)

// Sort direction constants.
const (
	SortDirectionAsc  = "asc"
	SortDirectionDesc = "desc"
)
