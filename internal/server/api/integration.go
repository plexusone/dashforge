package api

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/plexusone/dashforge/ent"
	entintegration "github.com/plexusone/dashforge/ent/integration"
	"github.com/plexusone/dashforge/integration/channel"
	"github.com/plexusone/dashforge/internal/server/db"
)

// IntegrationHandler handles integration API requests.
type IntegrationHandler struct {
	db     db.Database
	logger *slog.Logger
	mux    *http.ServeMux
}

// NewIntegrationHandler creates a new IntegrationHandler.
func NewIntegrationHandler(database db.Database, logger *slog.Logger) *IntegrationHandler {
	h := &IntegrationHandler{
		db:     database,
		logger: logger,
		mux:    http.NewServeMux(),
	}
	h.setupRoutes()
	return h
}

func (h *IntegrationHandler) setupRoutes() {
	// CRUD operations
	h.mux.HandleFunc("GET /api/v1/integrations", h.listIntegrations)
	h.mux.HandleFunc("GET /api/v1/integrations/{id}", h.getIntegration)
	h.mux.HandleFunc("POST /api/v1/integrations", h.createIntegration)
	h.mux.HandleFunc("PUT /api/v1/integrations/{id}", h.updateIntegration)
	h.mux.HandleFunc("DELETE /api/v1/integrations/{id}", h.deleteIntegration)

	// Operations
	h.mux.HandleFunc("POST /api/v1/integrations/{id}/test", h.testIntegration)

	// Channel info
	h.mux.HandleFunc("GET /api/v1/integrations/channels", h.listChannels)
}

// ServeHTTP implements http.Handler.
func (h *IntegrationHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.mux.ServeHTTP(w, r)
}

// client returns the Ent client.
func (h *IntegrationHandler) client() *ent.Client {
	if h.db == nil {
		return nil
	}
	return h.db.Client()
}

// listIntegrations handles GET /api/v1/integrations
func (h *IntegrationHandler) listIntegrations(w http.ResponseWriter, r *http.Request) {
	client := h.client()
	if client == nil {
		h.jsonError(w, http.StatusServiceUnavailable, "database not configured")
		return
	}

	// Parse query parameters
	channelType := r.URL.Query().Get("channelType")
	status := r.URL.Query().Get("status")
	source := r.URL.Query().Get("source")

	query := client.Integration.Query()

	if channelType != "" {
		query = query.Where(entintegration.ChannelTypeEQ(entintegration.ChannelType(channelType)))
	}
	if status != "" {
		query = query.Where(entintegration.StatusEQ(entintegration.Status(status)))
	}
	if source != "" {
		query = query.Where(entintegration.SourceEQ(entintegration.Source(source)))
	}

	integrations, err := query.
		Order(ent.Asc(entintegration.FieldName)).
		All(r.Context())

	if err != nil {
		h.logger.Error("failed to list integrations", "error", err)
		h.jsonError(w, http.StatusInternalServerError, "failed to list integrations")
		return
	}

	result := make([]map[string]any, 0, len(integrations))
	for _, i := range integrations {
		result = append(result, h.integrationToResponse(i, false))
	}

	h.jsonResponse(w, http.StatusOK, map[string]any{
		"integrations": result,
		"total":        len(result),
	})
}

// getIntegration handles GET /api/v1/integrations/{id}
func (h *IntegrationHandler) getIntegration(w http.ResponseWriter, r *http.Request) {
	client := h.client()
	if client == nil {
		h.jsonError(w, http.StatusServiceUnavailable, "database not configured")
		return
	}

	id, err := h.parseID(r)
	if err != nil {
		h.jsonError(w, http.StatusBadRequest, "invalid id")
		return
	}

	integration, err := client.Integration.Get(r.Context(), id)
	if err != nil {
		if ent.IsNotFound(err) {
			h.jsonError(w, http.StatusNotFound, "integration not found")
			return
		}
		h.logger.Error("failed to get integration", "error", err, "id", id)
		h.jsonError(w, http.StatusInternalServerError, "failed to get integration")
		return
	}

	h.jsonResponse(w, http.StatusOK, h.integrationToResponse(integration, false))
}

// createIntegration handles POST /api/v1/integrations
func (h *IntegrationHandler) createIntegration(w http.ResponseWriter, r *http.Request) {
	client := h.client()
	if client == nil {
		h.jsonError(w, http.StatusServiceUnavailable, "database not configured")
		return
	}

	var req struct {
		Name        string         `json:"name"`
		Slug        string         `json:"slug"`
		ChannelType string         `json:"channelType"`
		Config      map[string]any `json:"config"`
		Credentials map[string]any `json:"credentials"`
		Source      string         `json:"source"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.jsonError(w, http.StatusBadRequest, "invalid request: "+err.Error())
		return
	}

	// Validation
	if req.Name == "" || req.Slug == "" {
		h.jsonError(w, http.StatusBadRequest, "name and slug are required")
		return
	}
	if req.ChannelType == "" {
		h.jsonError(w, http.StatusBadRequest, "channelType is required")
		return
	}

	// Check if channel type exists
	if _, ok := channel.Get(req.ChannelType); !ok {
		h.jsonError(w, http.StatusBadRequest, "unknown channel type: "+req.ChannelType)
		return
	}

	// Set defaults
	source := entintegration.SourceBuiltin
	if req.Source != "" {
		source = entintegration.Source(req.Source)
	}

	// Create integration
	create := client.Integration.Create().
		SetName(req.Name).
		SetSlug(req.Slug).
		SetChannelType(entintegration.ChannelType(req.ChannelType)).
		SetSource(source).
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
		h.logger.Error("failed to create integration", "error", err)
		h.jsonError(w, http.StatusInternalServerError, "failed to create integration")
		return
	}

	h.logger.Info("integration created",
		"id", integration.ID,
		"name", integration.Name,
		"channelType", integration.ChannelType,
	)

	h.jsonResponse(w, http.StatusCreated, h.integrationToResponse(integration, false))
}

// updateIntegration handles PUT /api/v1/integrations/{id}
func (h *IntegrationHandler) updateIntegration(w http.ResponseWriter, r *http.Request) {
	client := h.client()
	if client == nil {
		h.jsonError(w, http.StatusServiceUnavailable, "database not configured")
		return
	}

	id, err := h.parseID(r)
	if err != nil {
		h.jsonError(w, http.StatusBadRequest, "invalid id")
		return
	}

	var req struct {
		Name        *string         `json:"name"`
		Config      *map[string]any `json:"config"`
		Credentials *map[string]any `json:"credentials"`
		Status      *string         `json:"status"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.jsonError(w, http.StatusBadRequest, "invalid request: "+err.Error())
		return
	}

	// Get existing integration
	integration, err := client.Integration.Get(r.Context(), id)
	if err != nil {
		if ent.IsNotFound(err) {
			h.jsonError(w, http.StatusNotFound, "integration not found")
			return
		}
		h.logger.Error("failed to get integration", "error", err, "id", id)
		h.jsonError(w, http.StatusInternalServerError, "failed to get integration")
		return
	}

	// Build update
	update := client.Integration.UpdateOne(integration)

	if req.Name != nil {
		update = update.SetName(*req.Name)
	}
	if req.Config != nil {
		update = update.SetConfig(*req.Config)
	}
	if req.Credentials != nil {
		update = update.SetCredentials(*req.Credentials)
	}
	if req.Status != nil {
		update = update.SetStatus(entintegration.Status(*req.Status))
	}

	integration, err = update.Save(r.Context())
	if err != nil {
		h.logger.Error("failed to update integration", "error", err, "id", id)
		h.jsonError(w, http.StatusInternalServerError, "failed to update integration")
		return
	}

	h.logger.Info("integration updated", "id", integration.ID, "name", integration.Name)

	h.jsonResponse(w, http.StatusOK, h.integrationToResponse(integration, false))
}

// deleteIntegration handles DELETE /api/v1/integrations/{id}
func (h *IntegrationHandler) deleteIntegration(w http.ResponseWriter, r *http.Request) {
	client := h.client()
	if client == nil {
		h.jsonError(w, http.StatusServiceUnavailable, "database not configured")
		return
	}

	id, err := h.parseID(r)
	if err != nil {
		h.jsonError(w, http.StatusBadRequest, "invalid id")
		return
	}

	err = client.Integration.DeleteOneID(id).Exec(r.Context())
	if err != nil {
		if ent.IsNotFound(err) {
			h.jsonError(w, http.StatusNotFound, "integration not found")
			return
		}
		h.logger.Error("failed to delete integration", "error", err, "id", id)
		h.jsonError(w, http.StatusInternalServerError, "failed to delete integration")
		return
	}

	h.logger.Info("integration deleted", "id", id)

	w.WriteHeader(http.StatusNoContent)
}

// testIntegration handles POST /api/v1/integrations/{id}/test
func (h *IntegrationHandler) testIntegration(w http.ResponseWriter, r *http.Request) {
	client := h.client()
	if client == nil {
		h.jsonError(w, http.StatusServiceUnavailable, "database not configured")
		return
	}

	id, err := h.parseID(r)
	if err != nil {
		h.jsonError(w, http.StatusBadRequest, "invalid id")
		return
	}

	integration, err := client.Integration.Get(r.Context(), id)
	if err != nil {
		if ent.IsNotFound(err) {
			h.jsonError(w, http.StatusNotFound, "integration not found")
			return
		}
		h.logger.Error("failed to get integration", "error", err, "id", id)
		h.jsonError(w, http.StatusInternalServerError, "failed to get integration")
		return
	}

	// Get channel
	ch, ok := channel.Get(string(integration.ChannelType))
	if !ok {
		h.jsonError(w, http.StatusBadRequest, "unknown channel type: "+string(integration.ChannelType))
		return
	}

	// Build config
	config := h.buildChannelConfig(integration)

	// Test connection
	start := time.Now()
	err = ch.TestConnection(r.Context(), config)
	elapsed := time.Since(start)

	if err != nil {
		h.logger.Warn("integration connection test failed",
			"id", id,
			"name", integration.Name,
			"error", err,
		)

		// Update status
		_, _ = client.Integration.UpdateOne(integration).
			SetStatus(entintegration.StatusError).
			SetStatusMessage(err.Error()).
			Save(r.Context())

		h.jsonResponse(w, http.StatusOK, map[string]any{
			"success":    false,
			"error":      err.Error(),
			"durationMs": elapsed.Milliseconds(),
		})
		return
	}

	// Update status to active
	_, _ = client.Integration.UpdateOne(integration).
		SetStatus(entintegration.StatusActive).
		SetStatusMessage("").
		Save(r.Context())

	h.jsonResponse(w, http.StatusOK, map[string]any{
		"success":    true,
		"durationMs": elapsed.Milliseconds(),
	})
}

// listChannels handles GET /api/v1/integrations/channels
func (h *IntegrationHandler) listChannels(w http.ResponseWriter, _ *http.Request) {
	channels := channel.ListChannels()

	h.jsonResponse(w, http.StatusOK, map[string]any{
		"channels": channels,
	})
}

// Helper methods

func (h *IntegrationHandler) parseID(r *http.Request) (int, error) {
	idStr := r.PathValue("id")
	return strconv.Atoi(idStr)
}

func (h *IntegrationHandler) buildChannelConfig(integration *ent.Integration) channel.Config {
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

//nolint:unparam // includeSecrets is for future use when admin endpoints need to expose credentials
func (h *IntegrationHandler) integrationToResponse(i *ent.Integration, includeSecrets bool) map[string]any {
	resp := map[string]any{
		"id":          i.ID,
		"name":        i.Name,
		"slug":        i.Slug,
		"channelType": i.ChannelType,
		"status":      i.Status,
		"source":      i.Source,
		"createdAt":   i.CreatedAt,
		"updatedAt":   i.UpdatedAt,
	}

	if i.Config != nil {
		resp["config"] = i.Config
	}

	if i.StatusMessage != "" {
		resp["statusMessage"] = i.StatusMessage
	}

	if i.LastUsedAt != nil {
		resp["lastUsedAt"] = i.LastUsedAt
	}

	if i.MarketplaceSlug != "" {
		resp["marketplaceSlug"] = i.MarketplaceSlug
	}

	// Only include credentials if explicitly requested
	if includeSecrets && i.Credentials != nil {
		resp["credentials"] = i.Credentials
	}

	return resp
}

func (h *IntegrationHandler) jsonResponse(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		h.logger.Error("failed to encode response", "error", err)
	}
}

func (h *IntegrationHandler) jsonError(w http.ResponseWriter, status int, message string) {
	h.jsonResponse(w, status, map[string]string{"error": message})
}
