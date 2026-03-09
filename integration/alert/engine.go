// Package alert provides the alert evaluation engine that monitors
// metric thresholds, schedules, and data changes to trigger notifications.
package alert

import (
	"context"
	"encoding/json"
	"log/slog"
	"sync"
	"time"

	"github.com/plexusone/dashforge/ent"
	"github.com/plexusone/dashforge/ent/alert"
	"github.com/plexusone/dashforge/ent/alertevent"
	"github.com/plexusone/dashforge/integration/channel"
)

// DefaultEvaluationInterval is the default time between alert evaluations.
const DefaultEvaluationInterval = 30 * time.Second

// MaxConsecutiveFailures is the max failures before an alert is disabled.
const MaxConsecutiveFailures = 10

// Engine is the alert evaluation engine that runs periodically to check alerts.
type Engine struct {
	client   *ent.Client
	logger   *slog.Logger
	interval time.Duration

	mu        sync.RWMutex
	running   bool
	stopCh    chan struct{}
	evaluator *Evaluator
}

// NewEngine creates a new alert engine.
func NewEngine(client *ent.Client, logger *slog.Logger) *Engine {
	if logger == nil {
		logger = slog.Default()
	}
	return &Engine{
		client:    client,
		logger:    logger,
		interval:  DefaultEvaluationInterval,
		evaluator: NewEvaluator(client, logger),
	}
}

// SetInterval sets the evaluation interval.
func (e *Engine) SetInterval(d time.Duration) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.interval = d
}

// Start begins the alert evaluation loop.
func (e *Engine) Start(ctx context.Context) error {
	e.mu.Lock()
	if e.running {
		e.mu.Unlock()
		return nil
	}
	e.running = true
	e.stopCh = make(chan struct{})
	e.mu.Unlock()

	e.logger.Info("starting alert engine", "interval", e.interval)

	go e.run(ctx)
	return nil
}

// Stop stops the alert evaluation loop.
func (e *Engine) Stop() {
	e.mu.Lock()
	defer e.mu.Unlock()

	if !e.running {
		return
	}

	close(e.stopCh)
	e.running = false
	e.logger.Info("stopped alert engine")
}

// IsRunning returns whether the engine is running.
func (e *Engine) IsRunning() bool {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.running
}

// run is the main evaluation loop.
func (e *Engine) run(ctx context.Context) {
	ticker := time.NewTicker(e.interval)
	defer ticker.Stop()

	// Run once immediately
	e.evaluateAll(ctx)

	for {
		select {
		case <-ctx.Done():
			e.mu.Lock()
			e.running = false
			e.mu.Unlock()
			return
		case <-e.stopCh:
			return
		case <-ticker.C:
			e.evaluateAll(ctx)
		}
	}
}

// evaluateAll evaluates all enabled alerts.
func (e *Engine) evaluateAll(ctx context.Context) {
	alerts, err := e.client.Alert.Query().
		Where(alert.EnabledEQ(true)).
		WithOrganization().
		WithChannels().
		All(ctx)
	if err != nil {
		e.logger.Error("failed to query alerts", "error", err)
		return
	}

	e.logger.Debug("evaluating alerts", "count", len(alerts))

	for _, a := range alerts {
		e.evaluateAlert(ctx, a)
	}
}

// evaluateAlert evaluates a single alert.
func (e *Engine) evaluateAlert(ctx context.Context, a *ent.Alert) {
	logger := e.logger.With("alert_id", a.ID, "alert_slug", a.Slug)

	// Check cooldown
	if a.LastTriggeredAt != nil {
		cooldownEnd := a.LastTriggeredAt.Add(time.Duration(a.CooldownSeconds) * time.Second)
		if time.Now().Before(cooldownEnd) {
			logger.Debug("alert in cooldown", "cooldown_ends", cooldownEnd)
			return
		}
	}

	// Get evaluator for trigger type
	result, err := e.evaluator.Evaluate(ctx, a)

	// Update last evaluated time
	now := time.Now()
	update := e.client.Alert.UpdateOne(a).SetLastEvaluatedAt(now)

	if err != nil {
		logger.Error("alert evaluation failed", "error", err)
		update = update.
			SetLastError(err.Error()).
			SetConsecutiveFailures(a.ConsecutiveFailures + 1)

		// Disable if too many failures
		if a.ConsecutiveFailures+1 >= MaxConsecutiveFailures {
			logger.Warn("disabling alert due to consecutive failures")
			update = update.SetEnabled(false)
		}

		if err := update.Exec(ctx); err != nil {
			logger.Error("failed to update alert after error", "error", err)
		}

		// Record error event
		e.recordEvent(ctx, a, alertevent.EventTypeError, nil, err.Error())
		return
	}

	// Clear error state on success
	update = update.SetLastError("").SetConsecutiveFailures(0)

	if !result.Triggered {
		if err := update.Exec(ctx); err != nil {
			logger.Error("failed to update alert", "error", err)
		}
		return
	}

	// Alert triggered!
	logger.Info("alert triggered", "trigger_data", result.Data)
	update = update.SetLastTriggeredAt(now)

	if err := update.Exec(ctx); err != nil {
		logger.Error("failed to update alert after trigger", "error", err)
	}

	// Send notifications
	e.sendNotifications(ctx, a, result)
}

// sendNotifications sends notifications to all configured channels.
func (e *Engine) sendNotifications(ctx context.Context, a *ent.Alert, result *EvaluationResult) {
	logger := e.logger.With("alert_id", a.ID)

	// Build message
	msg := channel.Message{
		Title:       a.Name,
		Body:        result.Message,
		Severity:    mapSeverity(result.Severity),
		Source:      "alert:" + a.Slug,
		AlertID:     a.ID,
		TriggerData: result.Data,
		Timestamp:   time.Now(),
	}

	// Get dashboard URL if associated
	if dashboard, err := a.QueryDashboard().Only(ctx); err == nil {
		msg.DashboardURL = "/dashboards/" + dashboard.Slug
	}

	// Get channels
	integrations, err := a.QueryChannels().All(ctx)
	if err != nil {
		logger.Error("failed to query channels", "error", err)
		return
	}

	var success, failed int
	var notifiedChannels []string

	for _, integration := range integrations {
		ch, ok := channel.Get(string(integration.ChannelType))
		if !ok {
			logger.Warn("unknown channel type", "type", integration.ChannelType)
			failed++
			continue
		}

		// Build config from integration
		config := e.buildChannelConfig(integration)

		if err := ch.Send(ctx, config, msg); err != nil {
			logger.Error("failed to send to channel",
				"channel_type", integration.ChannelType,
				"integration_id", integration.ID,
				"error", err)
			failed++
		} else {
			logger.Debug("sent notification", "channel_type", integration.ChannelType)
			success++
			notifiedChannels = append(notifiedChannels, integration.Slug)
		}
	}

	// Record trigger event
	e.recordEvent(ctx, a, alertevent.EventTypeTriggered, map[string]any{
		"trigger_data":      result.Data,
		"channels_notified": notifiedChannels,
		"channels_success":  success,
		"channels_failed":   failed,
	}, "")
}

// buildChannelConfig builds a channel.Config from an integration entity.
func (e *Engine) buildChannelConfig(integration *ent.Integration) channel.Config {
	config := channel.Config{
		Name: integration.Name,
	}

	// Copy config fields
	if integration.Config != nil {
		if v, ok := integration.Config["channelId"].(string); ok {
			config.ChannelID = v
		}
		if v, ok := integration.Config["username"].(string); ok {
			config.Username = v
		}
		if v, ok := integration.Config["smtpHost"].(string); ok {
			config.SMTPHost = v
		}
		if v, ok := integration.Config["smtpPort"].(float64); ok {
			config.SMTPPort = int(v)
		}
		if v, ok := integration.Config["fromAddress"].(string); ok {
			config.FromAddress = v
		}
		if v, ok := integration.Config["fromName"].(string); ok {
			config.FromName = v
		}
		if v, ok := integration.Config["toAddresses"].(string); ok {
			config.ToAddresses = v
		}
		if v, ok := integration.Config["useTls"].(bool); ok {
			config.UseTLS = v
		}
		if v, ok := integration.Config["whatsappPhoneId"].(string); ok {
			config.WhatsAppPhoneID = v
		}
		if v, ok := integration.Config["whatsappRecipient"].(string); ok {
			config.WhatsAppRecipient = v
		}
		if v, ok := integration.Config["webhookUrl"].(string); ok {
			config.WebhookURL = v
		}
		if v, ok := integration.Config["webhookMethod"].(string); ok {
			config.WebhookMethod = v
		}
		if v, ok := integration.Config["webhookAuth"].(string); ok {
			config.WebhookAuth = v
		}
		if v, ok := integration.Config["webhookHeaders"].(map[string]any); ok {
			config.WebhookHeaders = make(map[string]string)
			for k, val := range v {
				if s, ok := val.(string); ok {
					config.WebhookHeaders[k] = s
				}
			}
		}
	}

	// Copy credentials (sensitive fields)
	if integration.Credentials != nil {
		if v, ok := integration.Credentials["botToken"].(string); ok {
			config.BotToken = v
		}
		if v, ok := integration.Credentials["smtpUsername"].(string); ok {
			config.SMTPUsername = v
		}
		if v, ok := integration.Credentials["smtpPassword"].(string); ok {
			config.SMTPPassword = v
		}
		if v, ok := integration.Credentials["sendGridKey"].(string); ok {
			config.SendGridKey = v
		}
		if v, ok := integration.Credentials["whatsappToken"].(string); ok {
			config.WhatsAppToken = v
		}
		if v, ok := integration.Credentials["webhookSecret"].(string); ok {
			config.WebhookSecret = v
		}
	}

	return config
}

// recordEvent records an alert event in the database.
func (e *Engine) recordEvent(ctx context.Context, a *ent.Alert, eventType alertevent.EventType, data map[string]any, errMsg string) {
	builder := e.client.AlertEvent.Create().
		SetAlert(a).
		SetEventType(eventType)

	if data != nil {
		if td, ok := data["trigger_data"].(map[string]any); ok {
			builder = builder.SetTriggerData(td)
		}
		if cn, ok := data["channels_notified"].([]string); ok {
			builder = builder.SetChannelsNotified(cn)
		}
		if cs, ok := data["channels_success"].(int); ok {
			builder = builder.SetChannelsSuccess(cs)
		}
		if cf, ok := data["channels_failed"].(int); ok {
			builder = builder.SetChannelsFailed(cf)
		}
	}

	if errMsg != "" {
		builder = builder.SetErrorMessage(errMsg)
	}

	if _, err := builder.Save(ctx); err != nil {
		e.logger.Error("failed to record alert event", "error", err)
	}
}

// mapSeverity maps internal severity to channel.Severity.
func mapSeverity(s string) channel.Severity {
	switch s {
	case "warning":
		return channel.SeverityWarning
	case "error":
		return channel.SeverityError
	case "critical":
		return channel.SeverityCritical
	default:
		return channel.SeverityInfo
	}
}

// Evaluator handles the evaluation of different trigger types.
type Evaluator struct {
	client    *ent.Client
	logger    *slog.Logger
	threshold *ThresholdEvaluator
	schedule  *ScheduleEvaluator
	dataChange *DataChangeEvaluator
}

// NewEvaluator creates a new evaluator.
func NewEvaluator(client *ent.Client, logger *slog.Logger) *Evaluator {
	return &Evaluator{
		client:    client,
		logger:    logger,
		threshold: NewThresholdEvaluator(client, logger),
		schedule:  NewScheduleEvaluator(logger),
		dataChange: NewDataChangeEvaluator(client, logger),
	}
}

// EvaluationResult contains the result of evaluating an alert.
type EvaluationResult struct {
	Triggered bool           `json:"triggered"`
	Message   string         `json:"message,omitempty"`
	Severity  string         `json:"severity,omitempty"`
	Data      map[string]any `json:"data,omitempty"`
}

// Evaluate evaluates an alert based on its trigger type.
func (e *Evaluator) Evaluate(ctx context.Context, a *ent.Alert) (*EvaluationResult, error) {
	switch a.TriggerType {
	case alert.TriggerTypeThreshold:
		return e.threshold.Evaluate(ctx, a)
	case alert.TriggerTypeSchedule:
		return e.schedule.Evaluate(ctx, a)
	case alert.TriggerTypeDataChange:
		return e.dataChange.Evaluate(ctx, a)
	default:
		return nil, &EvaluationError{Message: "unknown trigger type: " + string(a.TriggerType)}
	}
}

// EvaluationError represents an evaluation error.
type EvaluationError struct {
	Message string
	Err     error
}

func (e *EvaluationError) Error() string {
	if e.Err != nil {
		return e.Message + ": " + e.Err.Error()
	}
	return e.Message
}

func (e *EvaluationError) Unwrap() error {
	return e.Err
}

// TriggerConfig is a helper to parse trigger configuration.
type TriggerConfig struct {
	raw map[string]any
}

// NewTriggerConfig creates a TriggerConfig from alert trigger_config.
func NewTriggerConfig(a *ent.Alert) *TriggerConfig {
	return &TriggerConfig{raw: a.TriggerConfig}
}

// Get retrieves a value by key.
func (c *TriggerConfig) Get(key string) (any, bool) {
	v, ok := c.raw[key]
	return v, ok
}

// GetString retrieves a string value.
func (c *TriggerConfig) GetString(key string) string {
	if v, ok := c.raw[key].(string); ok {
		return v
	}
	return ""
}

// GetFloat retrieves a float value.
func (c *TriggerConfig) GetFloat(key string) float64 {
	switch v := c.raw[key].(type) {
	case float64:
		return v
	case int:
		return float64(v)
	case json.Number:
		f, _ := v.Float64()
		return f
	}
	return 0
}

// GetInt retrieves an int value.
func (c *TriggerConfig) GetInt(key string) int {
	switch v := c.raw[key].(type) {
	case float64:
		return int(v)
	case int:
		return v
	case json.Number:
		i, _ := v.Int64()
		return int(i)
	}
	return 0
}

// GetBool retrieves a bool value.
func (c *TriggerConfig) GetBool(key string) bool {
	if v, ok := c.raw[key].(bool); ok {
		return v
	}
	return false
}
