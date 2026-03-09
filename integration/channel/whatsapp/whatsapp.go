// Package whatsapp provides a WhatsApp Business API implementation of the channel.Channel interface.
package whatsapp

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

// Channel implements channel.Channel for WhatsApp Business API.
type Channel struct{}

// Type returns "whatsapp".
func (c *Channel) Type() string {
	return "whatsapp"
}

// Name returns the display name.
func (c *Channel) Name() string {
	return "WhatsApp"
}

// Validate checks if the WhatsApp config is valid.
func (c *Channel) Validate(config channel.Config) error {
	if config.WhatsAppPhoneID == "" {
		return channel.NewConfigError("whatsappPhoneId", "WhatsApp Phone Number ID is required", nil)
	}
	if config.WhatsAppToken == "" {
		return channel.NewConfigError("whatsappToken", "access token is required", nil)
	}
	if config.WhatsAppRecipient == "" {
		return channel.NewConfigError("whatsappRecipient", "recipient phone number is required", nil)
	}
	return nil
}

// Send sends a message via WhatsApp Business API.
func (c *Channel) Send(ctx context.Context, config channel.Config, message channel.Message) error {
	if err := c.Validate(config); err != nil {
		return err
	}

	// Build the message text
	text := buildWhatsAppMessage(message)

	// WhatsApp Cloud API payload
	payload := map[string]any{
		"messaging_product": "whatsapp",
		"recipient_type":    "individual",
		"to":                normalizePhoneNumber(config.WhatsAppRecipient),
		"type":              "text",
		"text": map[string]string{
			"body": text,
		},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return channel.NewSendError("whatsapp", "failed to marshal payload", err)
	}

	url := fmt.Sprintf("https://graph.facebook.com/v18.0/%s/messages", config.WhatsAppPhoneID)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return channel.NewSendError("whatsapp", "failed to create request", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+config.WhatsAppToken)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return channel.NewSendError("whatsapp", "failed to send request", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return channel.NewSendError("whatsapp", "failed to read response", err)
	}

	if resp.StatusCode >= 400 {
		var errResp whatsappErrorResponse
		if err := json.Unmarshal(respBody, &errResp); err == nil && errResp.Error.Message != "" {
			return channel.NewSendError("whatsapp", fmt.Sprintf("API error: %s", errResp.Error.Message), nil)
		}
		return channel.NewSendError("whatsapp", fmt.Sprintf("API error (status %d): %s", resp.StatusCode, string(respBody)), nil)
	}

	return nil
}

// TestConnection tests the WhatsApp configuration.
func (c *Channel) TestConnection(ctx context.Context, config channel.Config) error {
	if err := c.Validate(config); err != nil {
		return err
	}

	// Verify the phone number ID by fetching its details
	url := fmt.Sprintf("https://graph.facebook.com/v18.0/%s", config.WhatsAppPhoneID)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return channel.NewSendError("whatsapp", "failed to create request", err)
	}

	req.Header.Set("Authorization", "Bearer "+config.WhatsAppToken)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return channel.NewSendError("whatsapp", "failed to send request", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		respBody, _ := io.ReadAll(resp.Body)
		return channel.NewSendError("whatsapp", fmt.Sprintf("API error (status %d): %s", resp.StatusCode, string(respBody)), nil)
	}

	return nil
}

// Capabilities returns WhatsApp capabilities.
func (c *Channel) Capabilities() channel.Capabilities {
	return channel.Capabilities{
		SupportsRichText:    false, // WhatsApp has limited formatting
		SupportsAttachments: true,  // Can send media
		SupportsThreading:   false,
		SupportsReactions:   false,
		MaxMessageLength:    4096,
		RequiresRecipient:   true,
		SupportsBatching:    false,
	}
}

// whatsappErrorResponse represents a WhatsApp API error response.
type whatsappErrorResponse struct {
	Error struct {
		Message   string `json:"message"`
		Type      string `json:"type"`
		Code      int    `json:"code"`
		FBTraceID string `json:"fbtrace_id"`
	} `json:"error"`
}

// buildWhatsAppMessage constructs a text message for WhatsApp.
func buildWhatsAppMessage(message channel.Message) string {
	var sb strings.Builder

	// Severity emoji
	switch message.Severity {
	case channel.SeverityInfo:
		sb.WriteString("ℹ️ ")
	case channel.SeverityWarning:
		sb.WriteString("⚠️ ")
	case channel.SeverityError:
		sb.WriteString("🔴 ")
	case channel.SeverityCritical:
		sb.WriteString("🚨 ")
	}

	sb.WriteString("*")
	sb.WriteString(message.Title)
	sb.WriteString("*\n\n")
	sb.WriteString(message.Body)

	if message.DashboardURL != "" {
		sb.WriteString("\n\n🔗 ")
		sb.WriteString(message.DashboardURL)
	}

	if message.Source != "" {
		sb.WriteString("\n\n_Source: ")
		sb.WriteString(message.Source)
		sb.WriteString("_")
	}

	return sb.String()
}

// normalizePhoneNumber ensures the phone number is in the correct format.
func normalizePhoneNumber(phone string) string {
	// Remove common formatting characters
	phone = strings.ReplaceAll(phone, " ", "")
	phone = strings.ReplaceAll(phone, "-", "")
	phone = strings.ReplaceAll(phone, "(", "")
	phone = strings.ReplaceAll(phone, ")", "")

	// Remove leading + if present (WhatsApp API doesn't want it)
	phone = strings.TrimPrefix(phone, "+")

	return phone
}
