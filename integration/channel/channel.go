// Package channel provides a plugin-style adapter architecture for notification channels.
// It follows the same registry pattern as the datasource package, allowing multiple
// notification backends (Slack, Email, WhatsApp, Webhook, etc.) to be used interchangeably.
package channel

import (
	"context"
	"time"
)

// Channel defines the interface all notification channels must implement.
// Channels are registered at init() time and used to send notifications.
type Channel interface {
	// Type returns the channel type identifier (e.g., "slack", "email").
	Type() string

	// Name returns a human-readable name for display.
	Name() string

	// Validate checks if the config is valid before sending.
	Validate(config Config) error

	// Send sends a message through this channel.
	Send(ctx context.Context, config Config, message Message) error

	// TestConnection tests the channel configuration.
	TestConnection(ctx context.Context, config Config) error

	// Capabilities returns what this channel supports.
	Capabilities() Capabilities
}

// Config holds channel-specific configuration.
// Different channel types use different fields.
type Config struct {
	// Common fields
	Name string `json:"name,omitempty"` // Display name for this config

	// Slack-specific
	BotToken  string `json:"botToken,omitempty"`  // Slack bot OAuth token
	ChannelID string `json:"channelId,omitempty"` // Slack channel ID
	Username  string `json:"username,omitempty"`  // Bot username override

	// Email-specific
	SMTPHost     string `json:"smtpHost,omitempty"`     // SMTP server host
	SMTPPort     int    `json:"smtpPort,omitempty"`     // SMTP server port
	SMTPUsername string `json:"smtpUsername,omitempty"` // SMTP auth username
	SMTPPassword string `json:"smtpPassword,omitempty"` // SMTP auth password
	FromAddress  string `json:"fromAddress,omitempty"`  // Sender email address
	FromName     string `json:"fromName,omitempty"`     // Sender display name
	ToAddresses  string `json:"toAddresses,omitempty"`  // Comma-separated recipients
	UseTLS       bool   `json:"useTls,omitempty"`       // Use TLS encryption
	SendGridKey  string `json:"sendGridKey,omitempty"`  // SendGrid API key (alternative to SMTP)

	// WhatsApp-specific
	WhatsAppBusinessID string `json:"whatsappBusinessId,omitempty"` // WhatsApp Business Account ID
	WhatsAppPhoneID    string `json:"whatsappPhoneId,omitempty"`    // WhatsApp Phone Number ID
	WhatsAppToken      string `json:"whatsappToken,omitempty"`      // Access token
	WhatsAppRecipient  string `json:"whatsappRecipient,omitempty"`  // Recipient phone number

	// Webhook-specific
	WebhookURL     string            `json:"webhookUrl,omitempty"`     // Target URL
	WebhookMethod  string            `json:"webhookMethod,omitempty"`  // HTTP method (POST, PUT, etc.)
	WebhookHeaders map[string]string `json:"webhookHeaders,omitempty"` // Custom headers
	WebhookAuth    string            `json:"webhookAuth,omitempty"`    // Auth type: none, basic, bearer
	WebhookSecret  string            `json:"webhookSecret,omitempty"`  // Auth secret/token

	// Extra contains channel-specific options not covered above.
	Extra map[string]any `json:"extra,omitempty"`
}

// Message represents a notification to be sent.
type Message struct {
	// Title is the message subject or headline.
	Title string `json:"title"`

	// Body is the main message content.
	Body string `json:"body"`

	// Severity indicates the importance level.
	Severity Severity `json:"severity"`

	// Source identifies where this notification came from.
	Source string `json:"source,omitempty"`

	// DashboardURL is an optional link to the dashboard.
	DashboardURL string `json:"dashboardUrl,omitempty"`

	// AlertID is the ID of the alert that triggered this.
	AlertID int `json:"alertId,omitempty"`

	// TriggerData contains the data that triggered the alert.
	TriggerData map[string]any `json:"triggerData,omitempty"`

	// Timestamp when the alert was triggered.
	Timestamp time.Time `json:"timestamp"`

	// Extra contains message-specific data for templating.
	Extra map[string]any `json:"extra,omitempty"`
}

// Severity represents the importance level of a notification.
type Severity string

const (
	SeverityInfo     Severity = "info"
	SeverityWarning  Severity = "warning"
	SeverityError    Severity = "error"
	SeverityCritical Severity = "critical"
)

// Capabilities describes what a channel supports.
type Capabilities struct {
	// SupportsRichText indicates if the channel supports formatted text.
	SupportsRichText bool `json:"supportsRichText"`

	// SupportsAttachments indicates if the channel supports file attachments.
	SupportsAttachments bool `json:"supportsAttachments"`

	// SupportsThreading indicates if the channel supports threaded replies.
	SupportsThreading bool `json:"supportsThreading"`

	// SupportsReactions indicates if the channel supports reactions/emojis.
	SupportsReactions bool `json:"supportsReactions"`

	// MaxMessageLength is the maximum message length (0 = no limit).
	MaxMessageLength int `json:"maxMessageLength"`

	// RequiresRecipient indicates if a recipient must be specified.
	RequiresRecipient bool `json:"requiresRecipient"`

	// SupportsBatching indicates if multiple messages can be batched.
	SupportsBatching bool `json:"supportsBatching"`
}

// SendResult contains the result of sending a message.
type SendResult struct {
	// Success indicates if the message was sent successfully.
	Success bool `json:"success"`

	// MessageID is the channel-specific message ID if available.
	MessageID string `json:"messageId,omitempty"`

	// Error contains the error message if Success is false.
	Error string `json:"error,omitempty"`

	// Timestamp is when the message was sent.
	Timestamp time.Time `json:"timestamp"`
}

// ConfigError represents a configuration validation error.
type ConfigError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
	Err     error  `json:"-"`
}

func (e *ConfigError) Error() string {
	if e.Err != nil {
		return "channel config error: " + e.Field + ": " + e.Message + ": " + e.Err.Error()
	}
	return "channel config error: " + e.Field + ": " + e.Message
}

func (e *ConfigError) Unwrap() error {
	return e.Err
}

// NewConfigError creates a new ConfigError.
func NewConfigError(field, message string, err error) *ConfigError {
	return &ConfigError{
		Field:   field,
		Message: message,
		Err:     err,
	}
}

// SendError represents an error sending a message.
type SendError struct {
	Channel string `json:"channel"`
	Message string `json:"message"`
	Err     error  `json:"-"`
}

func (e *SendError) Error() string {
	if e.Err != nil {
		return "channel send error [" + e.Channel + "]: " + e.Message + ": " + e.Err.Error()
	}
	return "channel send error [" + e.Channel + "]: " + e.Message
}

func (e *SendError) Unwrap() error {
	return e.Err
}

// NewSendError creates a new SendError.
func NewSendError(channel, message string, err error) *SendError {
	return &SendError{
		Channel: channel,
		Message: message,
		Err:     err,
	}
}
