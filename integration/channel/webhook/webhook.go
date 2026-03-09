// Package webhook provides a generic HTTP webhook implementation of the channel.Channel interface.
package webhook

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/plexusone/dashforge/integration/channel"
)

func init() {
	channel.Register(&Channel{})
}

// Channel implements channel.Channel for generic HTTP webhooks.
type Channel struct{}

// Type returns "webhook".
func (c *Channel) Type() string {
	return "webhook"
}

// Name returns the display name.
func (c *Channel) Name() string {
	return "Webhook"
}

// Validate checks if the webhook config is valid.
func (c *Channel) Validate(config channel.Config) error {
	if config.WebhookURL == "" {
		return channel.NewConfigError("webhookUrl", "webhook URL is required", nil)
	}
	if !strings.HasPrefix(config.WebhookURL, "http://") && !strings.HasPrefix(config.WebhookURL, "https://") {
		return channel.NewConfigError("webhookUrl", "must be a valid HTTP(S) URL", nil)
	}

	if config.WebhookAuth != "" {
		switch config.WebhookAuth {
		case "none", "basic", "bearer":
			// Valid
		default:
			return channel.NewConfigError("webhookAuth", "must be 'none', 'basic', or 'bearer'", nil)
		}

		if config.WebhookAuth != "none" && config.WebhookSecret == "" {
			return channel.NewConfigError("webhookSecret", "secret is required when auth is enabled", nil)
		}
	}

	return nil
}

// Send sends a message via HTTP webhook.
func (c *Channel) Send(ctx context.Context, config channel.Config, message channel.Message) error {
	if err := c.Validate(config); err != nil {
		return err
	}

	payload := buildWebhookPayload(message)
	body, err := json.Marshal(payload)
	if err != nil {
		return channel.NewSendError("webhook", "failed to marshal payload", err)
	}

	method := config.WebhookMethod
	if method == "" {
		method = http.MethodPost
	}

	req, err := http.NewRequestWithContext(ctx, method, config.WebhookURL, bytes.NewReader(body))
	if err != nil {
		return channel.NewSendError("webhook", "failed to create request", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "Dashforge/1.0")

	// Add custom headers
	for key, value := range config.WebhookHeaders {
		req.Header.Set(key, value)
	}

	// Add authentication
	if err := addAuth(req, config); err != nil {
		return err
	}

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return channel.NewSendError("webhook", "failed to send request", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		respBody, _ := io.ReadAll(resp.Body)
		return channel.NewSendError("webhook", fmt.Sprintf("request failed (status %d): %s", resp.StatusCode, string(respBody)), nil)
	}

	return nil
}

// TestConnection tests the webhook configuration.
func (c *Channel) TestConnection(ctx context.Context, config channel.Config) error {
	if err := c.Validate(config); err != nil {
		return err
	}

	// Send a test payload
	testPayload := map[string]any{
		"type":      "test",
		"message":   "Connection test from Dashforge",
		"timestamp": time.Now().Format(time.RFC3339),
	}

	body, err := json.Marshal(testPayload)
	if err != nil {
		return channel.NewSendError("webhook", "failed to marshal test payload", err)
	}

	method := config.WebhookMethod
	if method == "" {
		method = http.MethodPost
	}

	req, err := http.NewRequestWithContext(ctx, method, config.WebhookURL, bytes.NewReader(body))
	if err != nil {
		return channel.NewSendError("webhook", "failed to create request", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "Dashforge/1.0")
	req.Header.Set("X-Dashforge-Test", "true")

	// Add custom headers
	for key, value := range config.WebhookHeaders {
		req.Header.Set(key, value)
	}

	// Add authentication
	if err := addAuth(req, config); err != nil {
		return err
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return channel.NewSendError("webhook", "failed to connect", err)
	}
	defer resp.Body.Close()

	// Accept any 2xx or 3xx status as success for test
	if resp.StatusCode >= 400 {
		respBody, _ := io.ReadAll(resp.Body)
		return channel.NewSendError("webhook", fmt.Sprintf("test failed (status %d): %s", resp.StatusCode, string(respBody)), nil)
	}

	return nil
}

// Capabilities returns webhook capabilities.
func (c *Channel) Capabilities() channel.Capabilities {
	return channel.Capabilities{
		SupportsRichText:    false, // Depends on receiver
		SupportsAttachments: false,
		SupportsThreading:   false,
		SupportsReactions:   false,
		MaxMessageLength:    0, // No limit (depends on receiver)
		RequiresRecipient:   false,
		SupportsBatching:    true,
	}
}

// WebhookPayload is the standard payload sent to webhooks.
type WebhookPayload struct {
	Type         string         `json:"type"`
	Title        string         `json:"title"`
	Body         string         `json:"body"`
	Severity     string         `json:"severity"`
	Source       string         `json:"source,omitempty"`
	DashboardURL string         `json:"dashboardUrl,omitempty"`
	AlertID      int            `json:"alertId,omitempty"`
	TriggerData  map[string]any `json:"triggerData,omitempty"`
	Timestamp    string         `json:"timestamp"`
	Extra        map[string]any `json:"extra,omitempty"`
}

// buildWebhookPayload constructs the webhook payload from the message.
func buildWebhookPayload(message channel.Message) WebhookPayload {
	return WebhookPayload{
		Type:         "alert",
		Title:        message.Title,
		Body:         message.Body,
		Severity:     string(message.Severity),
		Source:       message.Source,
		DashboardURL: message.DashboardURL,
		AlertID:      message.AlertID,
		TriggerData:  message.TriggerData,
		Timestamp:    message.Timestamp.Format(time.RFC3339),
		Extra:        message.Extra,
	}
}

// addAuth adds authentication to the request based on config.
func addAuth(req *http.Request, config channel.Config) error {
	switch config.WebhookAuth {
	case "basic":
		// Secret should be "username:password"
		encoded := base64.StdEncoding.EncodeToString([]byte(config.WebhookSecret))
		req.Header.Set("Authorization", "Basic "+encoded)
	case "bearer":
		req.Header.Set("Authorization", "Bearer "+config.WebhookSecret)
	}
	return nil
}
