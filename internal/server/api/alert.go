package api

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"github.com/plexusone/dashforge/ent"
	entalert "github.com/plexusone/dashforge/ent/alert"
	"github.com/plexusone/dashforge/ent/alertevent"
	entdashboard "github.com/plexusone/dashforge/ent/dashboard"
	"github.com/plexusone/dashforge/internal/server/db"
)

// AlertHandler handles alert API requests.
type AlertHandler struct {
	db     db.Database
	logger *slog.Logger
	mux    *http.ServeMux
}

// NewAlertHandler creates a new AlertHandler.
func NewAlertHandler(database db.Database, logger *slog.Logger) *AlertHandler {
	h := &AlertHandler{
		db:     database,
		logger: logger,
		mux:    http.NewServeMux(),
	}
	h.setupRoutes()
	return h
}

func (h *AlertHandler) setupRoutes() {
	// CRUD operations
	h.mux.HandleFunc("GET /api/v1/alerts", h.listAlerts)
	h.mux.HandleFunc("GET /api/v1/alerts/{id}", h.getAlert)
	h.mux.HandleFunc("POST /api/v1/alerts", h.createAlert)
	h.mux.HandleFunc("PUT /api/v1/alerts/{id}", h.updateAlert)
	h.mux.HandleFunc("DELETE /api/v1/alerts/{id}", h.deleteAlert)

	// Operations
	h.mux.HandleFunc("POST /api/v1/alerts/{id}/enable", h.enableAlert)
	h.mux.HandleFunc("POST /api/v1/alerts/{id}/disable", h.disableAlert)
	h.mux.HandleFunc("GET /api/v1/alerts/{id}/events", h.getAlertEvents)

	// Dashboard binding
	h.mux.HandleFunc("GET /api/v1/dashboards/{id}/alerts", h.getDashboardAlerts)
}

// ServeHTTP implements http.Handler.
func (h *AlertHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.mux.ServeHTTP(w, r)
}

// client returns the Ent client.
func (h *AlertHandler) client() *ent.Client {
	if h.db == nil {
		return nil
	}
	return h.db.Client()
}

// listAlerts handles GET /api/v1/alerts
func (h *AlertHandler) listAlerts(w http.ResponseWriter, r *http.Request) {
	client := h.client()
	if client == nil {
		h.jsonError(w, http.StatusServiceUnavailable, "database not configured")
		return
	}

	// Parse query parameters
	triggerType := r.URL.Query().Get("triggerType")
	enabled := r.URL.Query().Get("enabled")

	query := client.Alert.Query()

	if triggerType != "" {
		query = query.Where(entalert.TriggerTypeEQ(entalert.TriggerType(triggerType)))
	}
	if enabled != "" {
		query = query.Where(entalert.EnabledEQ(enabled == "true"))
	}

	alerts, err := query.
		Order(ent.Asc(entalert.FieldName)).
		WithChannels().
		All(r.Context())

	if err != nil {
		h.logger.Error("failed to list alerts", "error", err)
		h.jsonError(w, http.StatusInternalServerError, "failed to list alerts")
		return
	}

	result := make([]map[string]any, 0, len(alerts))
	for _, a := range alerts {
		result = append(result, h.alertToResponse(a))
	}

	h.jsonResponse(w, http.StatusOK, map[string]any{
		"alerts": result,
		"total":  len(result),
	})
}

// getAlert handles GET /api/v1/alerts/{id}
func (h *AlertHandler) getAlert(w http.ResponseWriter, r *http.Request) {
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

	alert, err := client.Alert.Query().
		Where(entalert.IDEQ(id)).
		WithChannels().
		Only(r.Context())
	if err != nil {
		if ent.IsNotFound(err) {
			h.jsonError(w, http.StatusNotFound, "alert not found")
			return
		}
		h.logger.Error("failed to get alert", "error", err, "id", id)
		h.jsonError(w, http.StatusInternalServerError, "failed to get alert")
		return
	}

	h.jsonResponse(w, http.StatusOK, h.alertToResponse(alert))
}

// createAlert handles POST /api/v1/alerts
func (h *AlertHandler) createAlert(w http.ResponseWriter, r *http.Request) {
	client := h.client()
	if client == nil {
		h.jsonError(w, http.StatusServiceUnavailable, "database not configured")
		return
	}

	var req struct {
		Name            string         `json:"name"`
		Slug            string         `json:"slug"`
		Description     string         `json:"description"`
		TriggerType     string         `json:"triggerType"`
		TriggerConfig   map[string]any `json:"triggerConfig"`
		CooldownSeconds int            `json:"cooldownSeconds"`
		Enabled         bool           `json:"enabled"`
		DashboardID     *int           `json:"dashboardId"`
		ChannelIDs      []int          `json:"channelIds"`
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
	if req.TriggerType == "" {
		h.jsonError(w, http.StatusBadRequest, "triggerType is required")
		return
	}
	if req.TriggerConfig == nil {
		h.jsonError(w, http.StatusBadRequest, "triggerConfig is required")
		return
	}

	// Set defaults
	if req.CooldownSeconds == 0 {
		req.CooldownSeconds = 300
	}

	// Create alert
	create := client.Alert.Create().
		SetName(req.Name).
		SetSlug(req.Slug).
		SetTriggerType(entalert.TriggerType(req.TriggerType)).
		SetTriggerConfig(req.TriggerConfig).
		SetCooldownSeconds(req.CooldownSeconds).
		SetEnabled(req.Enabled)

	if req.Description != "" {
		create = create.SetDescription(req.Description)
	}

	// Add channels
	if len(req.ChannelIDs) > 0 {
		create = create.AddChannelIDs(req.ChannelIDs...)
	}

	alert, err := create.Save(r.Context())
	if err != nil {
		if strings.Contains(err.Error(), "duplicate") || strings.Contains(err.Error(), "unique") {
			h.jsonError(w, http.StatusConflict, "alert with this slug already exists")
			return
		}
		h.logger.Error("failed to create alert", "error", err)
		h.jsonError(w, http.StatusInternalServerError, "failed to create alert")
		return
	}

	// Re-fetch with channels
	alert, _ = client.Alert.Query().
		Where(entalert.IDEQ(alert.ID)).
		WithChannels().
		Only(r.Context())

	h.logger.Info("alert created",
		"id", alert.ID,
		"name", alert.Name,
		"triggerType", alert.TriggerType,
	)

	h.jsonResponse(w, http.StatusCreated, h.alertToResponse(alert))
}

// updateAlert handles PUT /api/v1/alerts/{id}
func (h *AlertHandler) updateAlert(w http.ResponseWriter, r *http.Request) {
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
		Name            *string         `json:"name"`
		Description     *string         `json:"description"`
		TriggerConfig   *map[string]any `json:"triggerConfig"`
		CooldownSeconds *int            `json:"cooldownSeconds"`
		Enabled         *bool           `json:"enabled"`
		ChannelIDs      *[]int          `json:"channelIds"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.jsonError(w, http.StatusBadRequest, "invalid request: "+err.Error())
		return
	}

	// Get existing alert
	alert, err := client.Alert.Get(r.Context(), id)
	if err != nil {
		if ent.IsNotFound(err) {
			h.jsonError(w, http.StatusNotFound, "alert not found")
			return
		}
		h.logger.Error("failed to get alert", "error", err, "id", id)
		h.jsonError(w, http.StatusInternalServerError, "failed to get alert")
		return
	}

	// Build update
	update := client.Alert.UpdateOne(alert)

	if req.Name != nil {
		update = update.SetName(*req.Name)
	}
	if req.Description != nil {
		update = update.SetDescription(*req.Description)
	}
	if req.TriggerConfig != nil {
		update = update.SetTriggerConfig(*req.TriggerConfig)
	}
	if req.CooldownSeconds != nil {
		update = update.SetCooldownSeconds(*req.CooldownSeconds)
	}
	if req.Enabled != nil {
		update = update.SetEnabled(*req.Enabled)
	}
	if req.ChannelIDs != nil {
		update = update.ClearChannels().AddChannelIDs(*req.ChannelIDs...)
	}

	alert, err = update.Save(r.Context())
	if err != nil {
		h.logger.Error("failed to update alert", "error", err, "id", id)
		h.jsonError(w, http.StatusInternalServerError, "failed to update alert")
		return
	}

	// Re-fetch with channels
	alert, _ = client.Alert.Query().
		Where(entalert.IDEQ(alert.ID)).
		WithChannels().
		Only(r.Context())

	h.logger.Info("alert updated", "id", alert.ID, "name", alert.Name)

	h.jsonResponse(w, http.StatusOK, h.alertToResponse(alert))
}

// deleteAlert handles DELETE /api/v1/alerts/{id}
func (h *AlertHandler) deleteAlert(w http.ResponseWriter, r *http.Request) {
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

	err = client.Alert.DeleteOneID(id).Exec(r.Context())
	if err != nil {
		if ent.IsNotFound(err) {
			h.jsonError(w, http.StatusNotFound, "alert not found")
			return
		}
		h.logger.Error("failed to delete alert", "error", err, "id", id)
		h.jsonError(w, http.StatusInternalServerError, "failed to delete alert")
		return
	}

	h.logger.Info("alert deleted", "id", id)

	w.WriteHeader(http.StatusNoContent)
}

// enableAlert handles POST /api/v1/alerts/{id}/enable
func (h *AlertHandler) enableAlert(w http.ResponseWriter, r *http.Request) {
	h.setAlertEnabled(w, r, true)
}

// disableAlert handles POST /api/v1/alerts/{id}/disable
func (h *AlertHandler) disableAlert(w http.ResponseWriter, r *http.Request) {
	h.setAlertEnabled(w, r, false)
}

func (h *AlertHandler) setAlertEnabled(w http.ResponseWriter, r *http.Request, enabled bool) {
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

	alert, err := client.Alert.Get(r.Context(), id)
	if err != nil {
		if ent.IsNotFound(err) {
			h.jsonError(w, http.StatusNotFound, "alert not found")
			return
		}
		h.logger.Error("failed to get alert", "error", err, "id", id)
		h.jsonError(w, http.StatusInternalServerError, "failed to get alert")
		return
	}

	alert, err = client.Alert.UpdateOne(alert).
		SetEnabled(enabled).
		SetConsecutiveFailures(0).
		SetLastError("").
		Save(r.Context())
	if err != nil {
		h.logger.Error("failed to update alert", "error", err, "id", id)
		h.jsonError(w, http.StatusInternalServerError, "failed to update alert")
		return
	}

	// Re-fetch with channels
	alert, _ = client.Alert.Query().
		Where(entalert.IDEQ(alert.ID)).
		WithChannels().
		Only(r.Context())

	action := "disabled"
	if enabled {
		action = "enabled"
	}
	h.logger.Info("alert "+action, "id", alert.ID, "name", alert.Name)

	h.jsonResponse(w, http.StatusOK, h.alertToResponse(alert))
}

// getAlertEvents handles GET /api/v1/alerts/{id}/events
func (h *AlertHandler) getAlertEvents(w http.ResponseWriter, r *http.Request) {
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

	// Parse pagination
	limit := 50
	if l := r.URL.Query().Get("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	events, err := client.AlertEvent.Query().
		Where(alertevent.HasAlertWith(entalert.IDEQ(id))).
		Order(ent.Desc(alertevent.FieldCreatedAt)).
		Limit(limit).
		All(r.Context())

	if err != nil {
		h.logger.Error("failed to get alert events", "error", err, "alert_id", id)
		h.jsonError(w, http.StatusInternalServerError, "failed to get events")
		return
	}

	result := make([]map[string]any, 0, len(events))
	for _, e := range events {
		result = append(result, map[string]any{
			"id":               e.ID,
			"eventType":        e.EventType,
			"triggerData":      e.TriggerData,
			"channelsNotified": e.ChannelsNotified,
			"channelsSuccess":  e.ChannelsSuccess,
			"channelsFailed":   e.ChannelsFailed,
			"errorMessage":     e.ErrorMessage,
			"acknowledgedBy":   e.AcknowledgedBy,
			"createdAt":        e.CreatedAt,
		})
	}

	h.jsonResponse(w, http.StatusOK, map[string]any{
		"events": result,
		"total":  len(result),
	})
}

// getDashboardAlerts handles GET /api/v1/dashboards/{id}/alerts
func (h *AlertHandler) getDashboardAlerts(w http.ResponseWriter, r *http.Request) {
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

	alerts, err := client.Alert.Query().
		Where(entalert.HasDashboardWith(entdashboard.IDEQ(id))).
		WithChannels().
		All(r.Context())

	if err != nil {
		h.logger.Error("failed to get dashboard alerts", "error", err, "dashboard_id", id)
		h.jsonError(w, http.StatusInternalServerError, "failed to get alerts")
		return
	}

	result := make([]map[string]any, 0, len(alerts))
	for _, a := range alerts {
		result = append(result, h.alertToResponse(a))
	}

	h.jsonResponse(w, http.StatusOK, map[string]any{
		"alerts": result,
		"total":  len(result),
	})
}

// Helper methods

func (h *AlertHandler) parseID(r *http.Request) (int, error) {
	idStr := r.PathValue("id")
	return strconv.Atoi(idStr)
}

func (h *AlertHandler) alertToResponse(a *ent.Alert) map[string]any {
	resp := map[string]any{
		"id":                  a.ID,
		"name":                a.Name,
		"slug":                a.Slug,
		"triggerType":         a.TriggerType,
		"triggerConfig":       a.TriggerConfig,
		"enabled":             a.Enabled,
		"cooldownSeconds":     a.CooldownSeconds,
		"consecutiveFailures": a.ConsecutiveFailures,
		"createdAt":           a.CreatedAt,
		"updatedAt":           a.UpdatedAt,
	}

	if a.Description != "" {
		resp["description"] = a.Description
	}

	if a.LastTriggeredAt != nil {
		resp["lastTriggeredAt"] = a.LastTriggeredAt
	}

	if a.LastEvaluatedAt != nil {
		resp["lastEvaluatedAt"] = a.LastEvaluatedAt
	}

	if a.LastError != "" {
		resp["lastError"] = a.LastError
	}

	// Include channels if loaded
	if a.Edges.Channels != nil {
		channels := make([]map[string]any, 0, len(a.Edges.Channels))
		for _, ch := range a.Edges.Channels {
			channels = append(channels, map[string]any{
				"id":          ch.ID,
				"name":        ch.Name,
				"slug":        ch.Slug,
				"channelType": ch.ChannelType,
				"status":      ch.Status,
			})
		}
		resp["channels"] = channels
	}

	return resp
}

func (h *AlertHandler) jsonResponse(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		h.logger.Error("failed to encode response", "error", err)
	}
}

func (h *AlertHandler) jsonError(w http.ResponseWriter, status int, message string) {
	h.jsonResponse(w, status, map[string]string{"error": message})
}
