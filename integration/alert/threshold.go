package alert

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/plexusone/dashforge/ent"
)

// ThresholdEvaluator evaluates threshold-based alerts.
// Threshold alerts fire when a metric value crosses a defined boundary.
type ThresholdEvaluator struct {
	client *ent.Client
	logger *slog.Logger
}

// NewThresholdEvaluator creates a new threshold evaluator.
func NewThresholdEvaluator(client *ent.Client, logger *slog.Logger) *ThresholdEvaluator {
	return &ThresholdEvaluator{
		client: client,
		logger: logger,
	}
}

// ThresholdConfig represents the configuration for a threshold trigger.
type ThresholdConfig struct {
	// MetricField is the field name to evaluate
	MetricField string `json:"metricField"`

	// Operator is the comparison operator: gt, gte, lt, lte, eq, neq
	Operator string `json:"operator"`

	// Value is the threshold value to compare against
	Value float64 `json:"value"`

	// DatasourceSlug is the datasource to query
	DatasourceSlug string `json:"datasourceSlug"`

	// Query is the SQL query to execute to get the metric value
	Query string `json:"query"`

	// Severity is the alert severity when triggered
	Severity string `json:"severity"`
}

// Evaluate evaluates a threshold alert.
func (e *ThresholdEvaluator) Evaluate(ctx context.Context, a *ent.Alert) (*EvaluationResult, error) {
	config := NewTriggerConfig(a)

	// Parse threshold config
	operator := config.GetString("operator")
	threshold := config.GetFloat("value")
	metricField := config.GetString("metricField")
	severity := config.GetString("severity")

	if operator == "" {
		return nil, &EvaluationError{Message: "threshold operator not configured"}
	}

	if metricField == "" {
		metricField = "value" // default field name
	}

	if severity == "" {
		severity = "warning"
	}

	// Get the current metric value
	// In a real implementation, this would query the datasource
	// For now, we'll look for a static value or simulate
	currentValue := config.GetFloat("currentValue")

	// If query is configured, we'd execute it here
	query := config.GetString("query")
	if query != "" {
		// TODO: Execute query against datasource and extract metric value
		// This would integrate with the datasource package
		e.logger.Debug("would execute query", "query", query)
	}

	// Evaluate the threshold
	triggered := e.compare(currentValue, operator, threshold)

	if !triggered {
		return &EvaluationResult{
			Triggered: false,
			Data: map[string]any{
				"currentValue": currentValue,
				"operator":     operator,
				"threshold":    threshold,
			},
		}, nil
	}

	return &EvaluationResult{
		Triggered: true,
		Message:   fmt.Sprintf("%s %s %s threshold (%.2f %s %.2f)", a.Name, metricField, e.operatorVerb(operator), currentValue, e.operatorSymbol(operator), threshold),
		Severity:  severity,
		Data: map[string]any{
			"currentValue": currentValue,
			"operator":     operator,
			"threshold":    threshold,
			"metricField":  metricField,
		},
	}, nil
}

// compare performs the threshold comparison.
func (e *ThresholdEvaluator) compare(value float64, operator string, threshold float64) bool {
	switch operator {
	case "gt", ">":
		return value > threshold
	case "gte", ">=":
		return value >= threshold
	case "lt", "<":
		return value < threshold
	case "lte", "<=":
		return value <= threshold
	case "eq", "==", "=":
		return value == threshold
	case "neq", "!=", "<>":
		return value != threshold
	default:
		return false
	}
}

// operatorVerb returns a verb describing the comparison.
func (e *ThresholdEvaluator) operatorVerb(operator string) string {
	switch operator {
	case "gt", ">":
		return "exceeded"
	case "gte", ">=":
		return "reached or exceeded"
	case "lt", "<":
		return "dropped below"
	case "lte", "<=":
		return "dropped to or below"
	case "eq", "==", "=":
		return "equals"
	case "neq", "!=", "<>":
		return "differs from"
	default:
		return "crossed"
	}
}

// operatorSymbol returns the symbol for the operator.
func (e *ThresholdEvaluator) operatorSymbol(operator string) string {
	switch operator {
	case "gt":
		return ">"
	case "gte":
		return ">="
	case "lt":
		return "<"
	case "lte":
		return "<="
	case "eq":
		return "=="
	case "neq":
		return "!="
	default:
		return operator
	}
}
