// Package slack provides a Slack implementation of the channel.Channel interface.
package slack

import (
	"bytes"
	"context"
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

// Channel implements channel.Channel for Slack.
type Channel struct{}

// Type returns "slack".
func (c *Channel) Type() string {
	return "slack"
}

// Name returns the display name.
func (c *Channel) Name() string {
	return "Slack"
}

// Validate checks if the Slack config is valid.
func (c *Channel) Validate(config channel.Config) error {
	if config.BotToken == "" {
		return channel.NewConfigError("botToken", "bot token is required", nil)
	}
	if !strings.HasPrefix(config.BotToken, "xoxb-") {
		return channel.NewConfigError("botToken", "must be a bot token (starts with xoxb-)", nil)
	}
	if config.ChannelID == "" {
		return channel.NewConfigError("channelId", "channel ID is required", nil)
	}
	return nil
}

// Send sends a message to Slack.
func (c *Channel) Send(ctx context.Context, config channel.Config, message channel.Message) error {
	if err := c.Validate(config); err != nil {
		return err
	}

	payload := buildSlackMessage(config, message)
	body, err := json.Marshal(payload)
	if err != nil {
		return channel.NewSendError("slack", "failed to marshal message", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://slack.com/api/chat.postMessage", bytes.NewReader(body))
	if err != nil {
		return channel.NewSendError("slack", "failed to create request", err)
	}

	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	req.Header.Set("Authorization", "Bearer "+config.BotToken)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return channel.NewSendError("slack", "failed to send request", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return channel.NewSendError("slack", "failed to read response", err)
	}

	var slackResp slackResponse
	if err := json.Unmarshal(respBody, &slackResp); err != nil {
		return channel.NewSendError("slack", "failed to parse response", err)
	}

	if !slackResp.OK {
		return channel.NewSendError("slack", "API error: "+slackResp.Error, nil)
	}

	return nil
}

// TestConnection tests the Slack configuration.
func (c *Channel) TestConnection(ctx context.Context, config channel.Config) error {
	if err := c.Validate(config); err != nil {
		return err
	}

	// Test by calling auth.test to verify the token
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://slack.com/api/auth.test", nil)
	if err != nil {
		return channel.NewSendError("slack", "failed to create request", err)
	}

	req.Header.Set("Authorization", "Bearer "+config.BotToken)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return channel.NewSendError("slack", "failed to send request", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return channel.NewSendError("slack", "failed to read response", err)
	}

	var slackResp slackResponse
	if err := json.Unmarshal(respBody, &slackResp); err != nil {
		return channel.NewSendError("slack", "failed to parse response", err)
	}

	if !slackResp.OK {
		return channel.NewSendError("slack", "auth failed: "+slackResp.Error, nil)
	}

	return nil
}

// Capabilities returns Slack capabilities.
func (c *Channel) Capabilities() channel.Capabilities {
	return channel.Capabilities{
		SupportsRichText:    true,
		SupportsAttachments: true,
		SupportsThreading:   true,
		SupportsReactions:   true,
		MaxMessageLength:    40000, // Slack limit
		RequiresRecipient:   true,  // Channel ID required
		SupportsBatching:    false,
	}
}

// slackResponse represents a Slack API response.
type slackResponse struct {
	OK    bool   `json:"ok"`
	Error string `json:"error,omitempty"`
	TS    string `json:"ts,omitempty"`
}

// slackMessage represents a Slack message payload.
type slackMessage struct {
	Channel     string            `json:"channel"`
	Text        string            `json:"text,omitempty"`
	Username    string            `json:"username,omitempty"`
	Attachments []slackAttachment `json:"attachments,omitempty"`
	Blocks      []slackBlock      `json:"blocks,omitempty"`
}

type slackAttachment struct {
	Color      string `json:"color,omitempty"`
	Title      string `json:"title,omitempty"`
	TitleLink  string `json:"title_link,omitempty"`
	Text       string `json:"text,omitempty"`
	Footer     string `json:"footer,omitempty"`
	FooterIcon string `json:"footer_icon,omitempty"`
	Ts         int64  `json:"ts,omitempty"`
}

type slackBlock struct {
	Type string `json:"type"`
	Text any    `json:"text,omitempty"`
}

// buildSlackMessage constructs a Slack message from the notification.
func buildSlackMessage(config channel.Config, message channel.Message) slackMessage {
	color := severityToColor(message.Severity)

	attachment := slackAttachment{
		Color:     color,
		Title:     message.Title,
		TitleLink: message.DashboardURL,
		Text:      message.Body,
		Footer:    "Dashforge Alert",
		Ts:        message.Timestamp.Unix(),
	}

	if message.Source != "" {
		attachment.Footer = fmt.Sprintf("Dashforge Alert | %s", message.Source)
	}

	msg := slackMessage{
		Channel:     config.ChannelID,
		Attachments: []slackAttachment{attachment},
	}

	if config.Username != "" {
		msg.Username = config.Username
	}

	// Add fallback text for notifications
	msg.Text = fmt.Sprintf("[%s] %s", strings.ToUpper(string(message.Severity)), message.Title)

	return msg
}

// severityToColor maps severity levels to Slack colors.
func severityToColor(severity channel.Severity) string {
	switch severity {
	case channel.SeverityInfo:
		return "#36a64f" // green
	case channel.SeverityWarning:
		return "#ffa500" // orange
	case channel.SeverityError:
		return "#ff0000" // red
	case channel.SeverityCritical:
		return "#8b0000" // dark red
	default:
		return "#808080" // gray
	}
}
