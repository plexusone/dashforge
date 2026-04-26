// Package email provides an Email implementation of the channel.Channel interface.
// Supports both SMTP and SendGrid for sending emails.
package email

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"net/smtp"
	"strings"
	"time"

	"github.com/plexusone/dashforge/integration/channel"
)

func init() {
	channel.Register(&Channel{})
}

// Channel implements channel.Channel for Email.
type Channel struct{}

// Type returns "email".
func (c *Channel) Type() string {
	return "email"
}

// Name returns the display name.
func (c *Channel) Name() string {
	return "Email"
}

// Validate checks if the Email config is valid.
func (c *Channel) Validate(config channel.Config) error {
	// Either SMTP or SendGrid must be configured
	hasSMTP := config.SMTPHost != ""
	hasSendGrid := config.SendGridKey != ""

	if !hasSMTP && !hasSendGrid {
		return channel.NewConfigError("smtp/sendgrid", "either SMTP or SendGrid must be configured", nil)
	}

	if hasSMTP {
		if config.SMTPPort == 0 {
			return channel.NewConfigError("smtpPort", "SMTP port is required", nil)
		}
		if config.FromAddress == "" {
			return channel.NewConfigError("fromAddress", "sender address is required", nil)
		}
	}

	if config.ToAddresses == "" {
		return channel.NewConfigError("toAddresses", "recipient addresses are required", nil)
	}

	return nil
}

// Send sends an email.
func (c *Channel) Send(ctx context.Context, config channel.Config, message channel.Message) error {
	if err := c.Validate(config); err != nil {
		return err
	}

	// Use SendGrid if configured
	if config.SendGridKey != "" {
		return c.sendViaSendGrid(ctx, config, message)
	}

	// Fall back to SMTP
	return c.sendViaSMTP(ctx, config, message)
}

// sendViaSMTP sends email using SMTP.
func (c *Channel) sendViaSMTP(_ context.Context, config channel.Config, message channel.Message) error {
	addr := fmt.Sprintf("%s:%d", config.SMTPHost, config.SMTPPort)

	// Build email headers and body
	from := config.FromAddress
	if config.FromName != "" {
		from = fmt.Sprintf("%s <%s>", config.FromName, config.FromAddress)
	}

	recipients := strings.Split(config.ToAddresses, ",")
	for i := range recipients {
		recipients[i] = strings.TrimSpace(recipients[i])
	}

	subject := fmt.Sprintf("[%s] %s", strings.ToUpper(string(message.Severity)), message.Title)

	// Build HTML body
	htmlBody := buildEmailHTML(message)

	// Construct the email
	var emailBuf bytes.Buffer
	emailBuf.WriteString(fmt.Sprintf("From: %s\r\n", from))
	emailBuf.WriteString(fmt.Sprintf("To: %s\r\n", strings.Join(recipients, ", ")))
	emailBuf.WriteString(fmt.Sprintf("Subject: %s\r\n", subject))
	emailBuf.WriteString("MIME-Version: 1.0\r\n")
	emailBuf.WriteString("Content-Type: text/html; charset=UTF-8\r\n")
	emailBuf.WriteString("\r\n")
	emailBuf.WriteString(htmlBody)

	// Set up authentication if credentials provided
	var auth smtp.Auth
	if config.SMTPUsername != "" && config.SMTPPassword != "" {
		auth = smtp.PlainAuth("", config.SMTPUsername, config.SMTPPassword, config.SMTPHost)
	}

	// Send the email
	err := smtp.SendMail(addr, auth, config.FromAddress, recipients, emailBuf.Bytes())
	if err != nil {
		return channel.NewSendError("email", "failed to send via SMTP", err)
	}

	return nil
}

// sendViaSendGrid sends email using SendGrid API.
func (c *Channel) sendViaSendGrid(ctx context.Context, config channel.Config, message channel.Message) error {
	recipients := strings.Split(config.ToAddresses, ",")
	toList := make([]map[string]string, len(recipients))
	for i, r := range recipients {
		toList[i] = map[string]string{"email": strings.TrimSpace(r)}
	}

	from := map[string]string{"email": config.FromAddress}
	if config.FromName != "" {
		from["name"] = config.FromName
	}

	subject := fmt.Sprintf("[%s] %s", strings.ToUpper(string(message.Severity)), message.Title)

	payload := map[string]any{
		"personalizations": []map[string]any{
			{"to": toList},
		},
		"from":    from,
		"subject": subject,
		"content": []map[string]string{
			{
				"type":  "text/html",
				"value": buildEmailHTML(message),
			},
		},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return channel.NewSendError("email", "failed to marshal payload", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.sendgrid.com/v3/mail/send", bytes.NewReader(body))
	if err != nil {
		return channel.NewSendError("email", "failed to create request", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+config.SendGridKey)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return channel.NewSendError("email", "failed to send request", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		respBody, _ := io.ReadAll(resp.Body)
		return channel.NewSendError("email", fmt.Sprintf("SendGrid error (status %d): %s", resp.StatusCode, string(respBody)), nil)
	}

	return nil
}

// TestConnection tests the email configuration.
func (c *Channel) TestConnection(ctx context.Context, config channel.Config) error {
	if err := c.Validate(config); err != nil {
		return err
	}

	// For SendGrid, we can't easily test without sending
	// For SMTP, try to connect
	if config.SMTPHost != "" {
		addr := fmt.Sprintf("%s:%d", config.SMTPHost, config.SMTPPort)
		conn, err := smtp.Dial(addr)
		if err != nil {
			return channel.NewSendError("email", "failed to connect to SMTP server", err)
		}
		defer conn.Close()

		if err := conn.Hello("dashforge"); err != nil {
			return channel.NewSendError("email", "SMTP HELO failed", err)
		}

		return nil
	}

	// For SendGrid, assume configuration is valid if key is present
	return nil
}

// Capabilities returns Email capabilities.
func (c *Channel) Capabilities() channel.Capabilities {
	return channel.Capabilities{
		SupportsRichText:    true,
		SupportsAttachments: false, // Could be added later
		SupportsThreading:   false,
		SupportsReactions:   false,
		MaxMessageLength:    0, // No practical limit
		RequiresRecipient:   true,
		SupportsBatching:    true,
	}
}

// buildEmailHTML creates an HTML email body.
func buildEmailHTML(message channel.Message) string {
	tmpl := `<!DOCTYPE html>
<html>
<head>
    <style>
        body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; margin: 0; padding: 20px; background: #f5f5f5; }
        .container { max-width: 600px; margin: 0 auto; background: white; border-radius: 8px; overflow: hidden; box-shadow: 0 2px 4px rgba(0,0,0,0.1); }
        .header { padding: 20px; background: {{.HeaderColor}}; color: white; }
        .header h1 { margin: 0; font-size: 18px; }
        .content { padding: 20px; }
        .content p { margin: 0 0 15px; line-height: 1.6; color: #333; }
        .metadata { background: #f9f9f9; padding: 15px 20px; border-top: 1px solid #eee; font-size: 12px; color: #666; }
        .button { display: inline-block; padding: 10px 20px; background: #0066cc; color: white; text-decoration: none; border-radius: 4px; margin-top: 15px; }
        .severity { display: inline-block; padding: 4px 8px; border-radius: 4px; font-size: 12px; font-weight: bold; text-transform: uppercase; }
        .severity-info { background: #d4edda; color: #155724; }
        .severity-warning { background: #fff3cd; color: #856404; }
        .severity-error { background: #f8d7da; color: #721c24; }
        .severity-critical { background: #721c24; color: white; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>{{.Title}}</h1>
        </div>
        <div class="content">
            <span class="severity severity-{{.SeverityClass}}">{{.Severity}}</span>
            <p style="margin-top: 15px;">{{.Body}}</p>
            {{if .DashboardURL}}
            <a href="{{.DashboardURL}}" class="button">View Dashboard</a>
            {{end}}
        </div>
        <div class="metadata">
            <p>Alert triggered at {{.Timestamp}}{{if .Source}} | Source: {{.Source}}{{end}}</p>
            <p>Sent by Dashforge</p>
        </div>
    </div>
</body>
</html>`

	severityClass := string(message.Severity)
	headerColor := "#0066cc"
	switch message.Severity {
	case channel.SeverityWarning:
		headerColor = "#ffa500"
	case channel.SeverityError:
		headerColor = "#dc3545"
	case channel.SeverityCritical:
		headerColor = "#721c24"
	}

	data := map[string]any{
		"Title":         message.Title,
		"Body":          message.Body,
		"Severity":      string(message.Severity),
		"SeverityClass": severityClass,
		"HeaderColor":   headerColor,
		"DashboardURL":  message.DashboardURL,
		"Source":        message.Source,
		"Timestamp":     message.Timestamp.Format(time.RFC1123),
	}

	t, err := template.New("email").Parse(tmpl)
	if err != nil {
		return message.Body // Fallback to plain text
	}

	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		return message.Body // Fallback to plain text
	}

	return buf.String()
}
