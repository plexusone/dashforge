package api

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"

	"github.com/plexusone/dashforge/ent"
	entintegration "github.com/plexusone/dashforge/ent/integration"
	"github.com/plexusone/dashforge/integration/channel"
	"github.com/plexusone/dashforge/internal/server/db"
)

// MarketplaceHandler handles marketplace API requests.
type MarketplaceHandler struct {
	db     db.Database
	logger *slog.Logger
	mux    *http.ServeMux
}

// NewMarketplaceHandler creates a new MarketplaceHandler.
func NewMarketplaceHandler(database db.Database, logger *slog.Logger) *MarketplaceHandler {
	h := &MarketplaceHandler{
		db:     database,
		logger: logger,
		mux:    http.NewServeMux(),
	}
	h.setupRoutes()
	return h
}

func (h *MarketplaceHandler) setupRoutes() {
	h.mux.HandleFunc("GET /api/v1/marketplace/integrations", h.listMarketplaceIntegrations)
	h.mux.HandleFunc("GET /api/v1/marketplace/integrations/{slug}", h.getMarketplaceIntegration)
	h.mux.HandleFunc("POST /api/v1/marketplace/integrations/{slug}/install", h.installMarketplaceIntegration)
}

// ServeHTTP implements http.Handler.
func (h *MarketplaceHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.mux.ServeHTTP(w, r)
}

// client returns the Ent client.
func (h *MarketplaceHandler) client() *ent.Client {
	if h.db == nil {
		return nil
	}
	return h.db.Client()
}

// IntegrationDefinition represents a marketplace integration definition.
type IntegrationDefinition struct {
	Slug         string               `json:"slug"`
	Name         string               `json:"name"`
	Description  string               `json:"description"`
	ChannelType  string               `json:"channelType"`
	Category     string               `json:"category"`
	Icon         string               `json:"icon"`
	Version      string               `json:"version"`
	Author       string               `json:"author"`
	Source       string               `json:"source"`
	ConfigSchema map[string]any       `json:"configSchema"`
	Capabilities channel.Capabilities `json:"capabilities"`
}

// builtinIntegrations defines the curated list of built-in integrations.
var builtinIntegrations = []IntegrationDefinition{
	{
		Slug:        "slack",
		Name:        "Slack",
		Description: "Send alerts and notifications to Slack channels",
		ChannelType: "slack",
		Category:    "messaging",
		Icon:        "slack",
		Version:     "1.0.0",
		Author:      "Dashforge",
		Source:      "builtin",
		ConfigSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"botToken": map[string]any{
					"type":        "string",
					"title":       "Bot Token",
					"description": "Slack Bot OAuth Token (starts with xoxb-)",
					"secret":      true,
				},
				"channelId": map[string]any{
					"type":        "string",
					"title":       "Channel ID",
					"description": "Slack channel ID to send messages to",
				},
				"username": map[string]any{
					"type":        "string",
					"title":       "Bot Username",
					"description": "Optional custom bot username",
				},
			},
			"required": []string{"botToken", "channelId"},
		},
	},
	{
		Slug:        "email",
		Name:        "Email",
		Description: "Send alerts via email (SMTP or SendGrid)",
		ChannelType: "email",
		Category:    "messaging",
		Icon:        "mail",
		Version:     "1.0.0",
		Author:      "Dashforge",
		Source:      "builtin",
		ConfigSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"smtpHost": map[string]any{
					"type":        "string",
					"title":       "SMTP Host",
					"description": "SMTP server hostname",
				},
				"smtpPort": map[string]any{
					"type":        "integer",
					"title":       "SMTP Port",
					"description": "SMTP server port (587 for TLS, 465 for SSL)",
				},
				"smtpUsername": map[string]any{
					"type":        "string",
					"title":       "SMTP Username",
					"description": "SMTP authentication username",
					"secret":      true,
				},
				"smtpPassword": map[string]any{
					"type":        "string",
					"title":       "SMTP Password",
					"description": "SMTP authentication password",
					"secret":      true,
				},
				"fromAddress": map[string]any{
					"type":        "string",
					"title":       "From Address",
					"description": "Sender email address",
				},
				"fromName": map[string]any{
					"type":        "string",
					"title":       "From Name",
					"description": "Sender display name",
				},
				"toAddresses": map[string]any{
					"type":        "string",
					"title":       "To Addresses",
					"description": "Comma-separated recipient email addresses",
				},
				"sendGridKey": map[string]any{
					"type":        "string",
					"title":       "SendGrid API Key",
					"description": "Alternative to SMTP: SendGrid API key",
					"secret":      true,
				},
			},
			"required": []string{"toAddresses"},
		},
	},
	{
		Slug:        "whatsapp",
		Name:        "WhatsApp",
		Description: "Send alerts via WhatsApp Business API",
		ChannelType: "whatsapp",
		Category:    "messaging",
		Icon:        "message-circle",
		Version:     "1.0.0",
		Author:      "Dashforge",
		Source:      "builtin",
		ConfigSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"whatsappPhoneId": map[string]any{
					"type":        "string",
					"title":       "Phone Number ID",
					"description": "WhatsApp Business Phone Number ID",
				},
				"whatsappToken": map[string]any{
					"type":        "string",
					"title":       "Access Token",
					"description": "WhatsApp Business API access token",
					"secret":      true,
				},
				"whatsappRecipient": map[string]any{
					"type":        "string",
					"title":       "Recipient Phone",
					"description": "Recipient phone number with country code",
				},
			},
			"required": []string{"whatsappPhoneId", "whatsappToken", "whatsappRecipient"},
		},
	},
	{
		Slug:        "webhook",
		Name:        "Webhook",
		Description: "Send alerts to any HTTP endpoint",
		ChannelType: "webhook",
		Category:    "custom",
		Icon:        "globe",
		Version:     "1.0.0",
		Author:      "Dashforge",
		Source:      "builtin",
		ConfigSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"webhookUrl": map[string]any{
					"type":        "string",
					"title":       "Webhook URL",
					"description": "HTTP(S) endpoint URL",
				},
				"webhookMethod": map[string]any{
					"type":        "string",
					"title":       "HTTP Method",
					"description": "HTTP method (POST, PUT, etc.)",
					"default":     "POST",
				},
				"webhookAuth": map[string]any{
					"type":        "string",
					"title":       "Authentication",
					"description": "Authentication type: none, basic, bearer",
					"enum":        []string{"none", "basic", "bearer"},
					"default":     "none",
				},
				"webhookSecret": map[string]any{
					"type":        "string",
					"title":       "Auth Secret",
					"description": "Authentication secret (password or token)",
					"secret":      true,
				},
				"webhookHeaders": map[string]any{
					"type":        "object",
					"title":       "Custom Headers",
					"description": "Additional HTTP headers",
				},
			},
			"required": []string{"webhookUrl"},
		},
	},
}

// listMarketplaceIntegrations handles GET /api/v1/marketplace/integrations
func (h *MarketplaceHandler) listMarketplaceIntegrations(w http.ResponseWriter, r *http.Request) {
	category := r.URL.Query().Get("category")

	// Add capabilities from registered channels
	integrations := make([]IntegrationDefinition, 0, len(builtinIntegrations))
	for _, def := range builtinIntegrations {
		if category != "" && def.Category != category {
			continue
		}

		// Get capabilities from registered channel
		if ch, ok := channel.Get(def.ChannelType); ok {
			def.Capabilities = ch.Capabilities()
		}

		integrations = append(integrations, def)
	}

	h.jsonResponse(w, http.StatusOK, map[string]any{
		"integrations": integrations,
		"total":        len(integrations),
	})
}

// getMarketplaceIntegration handles GET /api/v1/marketplace/integrations/{slug}
func (h *MarketplaceHandler) getMarketplaceIntegration(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("slug")

	for _, def := range builtinIntegrations {
		if def.Slug == slug {
			// Get capabilities from registered channel
			if ch, ok := channel.Get(def.ChannelType); ok {
				def.Capabilities = ch.Capabilities()
			}
			h.jsonResponse(w, http.StatusOK, def)
			return
		}
	}

	h.jsonError(w, http.StatusNotFound, "integration not found")
}

// installMarketplaceIntegration handles POST /api/v1/marketplace/integrations/{slug}/install
func (h *MarketplaceHandler) installMarketplaceIntegration(w http.ResponseWriter, r *http.Request) {
	client := h.client()
	if client == nil {
		h.jsonError(w, http.StatusServiceUnavailable, "database not configured")
		return
	}

	marketplaceSlug := r.PathValue("slug")

	// Find the definition
	var def *IntegrationDefinition
	for _, d := range builtinIntegrations {
		if d.Slug == marketplaceSlug {
			def = &d
			break
		}
	}

	if def == nil {
		h.jsonError(w, http.StatusNotFound, "integration not found")
		return
	}

	var req struct {
		Name        string         `json:"name"`
		Slug        string         `json:"slug"`
		Config      map[string]any `json:"config"`
		Credentials map[string]any `json:"credentials"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.jsonError(w, http.StatusBadRequest, "invalid request: "+err.Error())
		return
	}

	// Validation
	if req.Name == "" {
		req.Name = def.Name
	}
	if req.Slug == "" {
		h.jsonError(w, http.StatusBadRequest, "slug is required")
		return
	}

	// Create the integration
	create := client.Integration.Create().
		SetName(req.Name).
		SetSlug(req.Slug).
		SetChannelType(entintegration.ChannelType(def.ChannelType)).
		SetSource(entintegration.Source(def.Source)).
		SetMarketplaceSlug(marketplaceSlug).
		SetStatus(entintegration.StatusInactive)

	if req.Config != nil {
		create = create.SetConfig(req.Config)
	}
	if req.Credentials != nil {
		create = create.SetCredentials(req.Credentials)
	}

	integration, err := create.Save(r.Context())
	if err != nil {
		if strings.Contains(err.Error(), "duplicate") || strings.Contains(err.Error(), "unique") {
			h.jsonError(w, http.StatusConflict, "integration with this slug already exists")
			return
		}
		h.logger.Error("failed to install integration", "error", err)
		h.jsonError(w, http.StatusInternalServerError, "failed to install integration")
		return
	}

	h.logger.Info("marketplace integration installed",
		"id", integration.ID,
		"name", integration.Name,
		"marketplaceSlug", marketplaceSlug,
	)

	h.jsonResponse(w, http.StatusCreated, map[string]any{
		"id":              integration.ID,
		"name":            integration.Name,
		"slug":            integration.Slug,
		"channelType":     integration.ChannelType,
		"status":          integration.Status,
		"source":          integration.Source,
		"marketplaceSlug": integration.MarketplaceSlug,
		"createdAt":       integration.CreatedAt,
	})
}

func (h *MarketplaceHandler) jsonResponse(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		h.logger.Error("failed to encode response", "error", err)
	}
}

func (h *MarketplaceHandler) jsonError(w http.ResponseWriter, status int, message string) {
	h.jsonResponse(w, status, map[string]string{"error": message})
}
