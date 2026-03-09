package alert

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/plexusone/dashforge/ent"
)

// DataChangeEvaluator evaluates data change alerts.
// These alerts fire when query results change from the previous evaluation.
type DataChangeEvaluator struct {
	client *ent.Client
	logger *slog.Logger

	// Cache of previous data hashes (in production, this would be persisted)
	previousHashes map[int]string
}

// NewDataChangeEvaluator creates a new data change evaluator.
func NewDataChangeEvaluator(client *ent.Client, logger *slog.Logger) *DataChangeEvaluator {
	return &DataChangeEvaluator{
		client:         client,
		logger:         logger,
		previousHashes: make(map[int]string),
	}
}

// DataChangeConfig represents the configuration for a data change trigger.
type DataChangeConfig struct {
	// DatasourceSlug is the datasource to query
	DatasourceSlug string `json:"datasourceSlug"`

	// Query is the SQL query to execute
	Query string `json:"query"`

	// ChangeType specifies what type of change to detect: any, increase, decrease, new_rows, deleted_rows
	ChangeType string `json:"changeType"`

	// CompareField is the field to compare for increase/decrease detection
	CompareField string `json:"compareField"`

	// Severity is the alert severity when triggered
	Severity string `json:"severity"`

	// Message template for the notification
	Message string `json:"message"`
}

// Evaluate evaluates a data change alert.
func (e *DataChangeEvaluator) Evaluate(ctx context.Context, a *ent.Alert) (*EvaluationResult, error) {
	config := NewTriggerConfig(a)

	changeType := config.GetString("changeType")
	if changeType == "" {
		changeType = "any"
	}

	severity := config.GetString("severity")
	if severity == "" {
		severity = "info"
	}

	message := config.GetString("message")
	if message == "" {
		message = fmt.Sprintf("Data change detected for alert: %s", a.Name)
	}

	// In a real implementation, this would:
	// 1. Execute the query against the datasource
	// 2. Compare with previous results
	// 3. Detect the specified type of change

	query := config.GetString("query")
	if query == "" {
		return nil, &EvaluationError{Message: "query not configured for data change alert"}
	}

	// Simulate getting current data
	// In production, this would execute the query
	currentData := config.GetString("currentData")
	if currentData == "" {
		// Use query as placeholder for simulation
		currentData = query
	}

	// Calculate hash of current data
	currentHash := e.hashData(currentData)

	// Get previous hash
	previousHash, hasPrevious := e.previousHashes[a.ID]

	// Store current hash for next evaluation
	e.previousHashes[a.ID] = currentHash

	// First evaluation - no change to detect
	if !hasPrevious {
		e.logger.Debug("first evaluation, storing baseline", "alert_id", a.ID)
		return &EvaluationResult{
			Triggered: false,
			Data: map[string]any{
				"changeType": changeType,
				"baseline":   true,
				"hash":       currentHash,
			},
		}, nil
	}

	// No change detected
	if currentHash == previousHash {
		return &EvaluationResult{
			Triggered: false,
			Data: map[string]any{
				"changeType": changeType,
				"changed":    false,
			},
		}, nil
	}

	// Change detected!
	e.logger.Info("data change detected",
		"alert_id", a.ID,
		"previous_hash", previousHash,
		"current_hash", currentHash)

	return &EvaluationResult{
		Triggered: true,
		Message:   message,
		Severity:  severity,
		Data: map[string]any{
			"changeType":   changeType,
			"previousHash": previousHash,
			"currentHash":  currentHash,
		},
	}, nil
}

// hashData creates a SHA256 hash of the data for comparison.
func (e *DataChangeEvaluator) hashData(data any) string {
	var bytes []byte

	switch v := data.(type) {
	case string:
		bytes = []byte(v)
	case []byte:
		bytes = v
	default:
		// JSON encode for complex types
		b, err := json.Marshal(data)
		if err != nil {
			return ""
		}
		bytes = b
	}

	hash := sha256.Sum256(bytes)
	return hex.EncodeToString(hash[:])
}

// ClearBaseline clears the stored baseline for an alert.
// This is useful when the alert configuration changes.
func (e *DataChangeEvaluator) ClearBaseline(alertID int) {
	delete(e.previousHashes, alertID)
}

// ClearAllBaselines clears all stored baselines.
func (e *DataChangeEvaluator) ClearAllBaselines() {
	e.previousHashes = make(map[int]string)
}
