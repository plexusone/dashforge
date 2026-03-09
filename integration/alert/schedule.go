package alert

import (
	"context"
	"log/slog"
	"strings"
	"time"

	"github.com/plexusone/dashforge/ent"
)

// ScheduleEvaluator evaluates schedule-based alerts.
// Schedule alerts fire at specific times based on a cron expression.
type ScheduleEvaluator struct {
	logger *slog.Logger
}

// NewScheduleEvaluator creates a new schedule evaluator.
func NewScheduleEvaluator(logger *slog.Logger) *ScheduleEvaluator {
	return &ScheduleEvaluator{
		logger: logger,
	}
}

// ScheduleConfig represents the configuration for a schedule trigger.
type ScheduleConfig struct {
	// Cron is a cron expression (minute hour day month weekday)
	Cron string `json:"cron"`

	// Timezone is the timezone for evaluation (e.g., "America/New_York")
	Timezone string `json:"timezone"`

	// Message is the notification message to send
	Message string `json:"message"`

	// Severity is the alert severity
	Severity string `json:"severity"`
}

// Evaluate evaluates a schedule alert.
func (e *ScheduleEvaluator) Evaluate(ctx context.Context, a *ent.Alert) (*EvaluationResult, error) {
	config := NewTriggerConfig(a)

	cron := config.GetString("cron")
	if cron == "" {
		return nil, &EvaluationError{Message: "cron expression not configured"}
	}

	timezone := config.GetString("timezone")
	message := config.GetString("message")
	severity := config.GetString("severity")

	if message == "" {
		message = "Scheduled alert: " + a.Name
	}

	if severity == "" {
		severity = "info"
	}

	// Determine the time to check
	now := time.Now()
	if timezone != "" {
		loc, err := time.LoadLocation(timezone)
		if err != nil {
			return nil, &EvaluationError{Message: "invalid timezone: " + timezone, Err: err}
		}
		now = now.In(loc)
	}

	// Check if the cron matches
	triggered, err := e.matchesCron(cron, now)
	if err != nil {
		return nil, &EvaluationError{Message: "invalid cron expression", Err: err}
	}

	if !triggered {
		return &EvaluationResult{
			Triggered: false,
			Data: map[string]any{
				"cron":     cron,
				"timezone": timezone,
				"checked":  now.Format(time.RFC3339),
			},
		}, nil
	}

	return &EvaluationResult{
		Triggered: true,
		Message:   message,
		Severity:  severity,
		Data: map[string]any{
			"cron":        cron,
			"timezone":    timezone,
			"triggeredAt": now.Format(time.RFC3339),
		},
	}, nil
}

// matchesCron checks if the current time matches the cron expression.
// Simplified cron parser supporting: minute hour day month weekday
// Supports: * (any), specific values, ranges (1-5), lists (1,3,5)
func (e *ScheduleEvaluator) matchesCron(cron string, t time.Time) (bool, error) {
	parts := strings.Fields(cron)
	if len(parts) != 5 {
		return false, &EvaluationError{Message: "cron must have 5 fields: minute hour day month weekday"}
	}

	minute := t.Minute()
	hour := t.Hour()
	day := t.Day()
	month := int(t.Month())
	weekday := int(t.Weekday()) // 0 = Sunday

	if !e.matchField(parts[0], minute, 0, 59) {
		return false, nil
	}
	if !e.matchField(parts[1], hour, 0, 23) {
		return false, nil
	}
	if !e.matchField(parts[2], day, 1, 31) {
		return false, nil
	}
	if !e.matchField(parts[3], month, 1, 12) {
		return false, nil
	}
	if !e.matchField(parts[4], weekday, 0, 6) {
		return false, nil
	}

	return true, nil
}

// matchField checks if a value matches a cron field pattern.
func (e *ScheduleEvaluator) matchField(pattern string, value, min, max int) bool {
	if pattern == "*" {
		return true
	}

	// Handle lists (1,3,5)
	if strings.Contains(pattern, ",") {
		for _, part := range strings.Split(pattern, ",") {
			if e.matchField(part, value, min, max) {
				return true
			}
		}
		return false
	}

	// Handle ranges (1-5)
	if strings.Contains(pattern, "-") {
		parts := strings.Split(pattern, "-")
		if len(parts) != 2 {
			return false
		}
		start := e.parseInt(parts[0])
		end := e.parseInt(parts[1])
		return value >= start && value <= end
	}

	// Handle step values (*/5)
	if strings.HasPrefix(pattern, "*/") {
		step := e.parseInt(pattern[2:])
		if step == 0 {
			return false
		}
		return value%step == 0
	}

	// Exact value
	return value == e.parseInt(pattern)
}

// parseInt parses an integer from string, returning 0 on error.
func (e *ScheduleEvaluator) parseInt(s string) int {
	var n int
	for _, c := range s {
		if c >= '0' && c <= '9' {
			n = n*10 + int(c-'0')
		}
	}
	return n
}
