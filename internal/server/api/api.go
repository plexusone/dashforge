// Package api provides the REST API handlers for Dashforge Server.
package api

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"

	"github.com/plexusone/dashforge/ent"
	"github.com/plexusone/dashforge/ent/dashboard"
	"github.com/plexusone/dashforge/internal/server/db"
)

// Handler handles API requests.
type Handler struct {
	db     db.Database
	logger *slog.Logger
	mux    *http.ServeMux
}

// NewHandler creates a new API handler.
func NewHandler(database db.Database, logger *slog.Logger) *Handler {
	h := &Handler{
		db:     database,
		logger: logger,
		mux:    http.NewServeMux(),
	}
	h.setupRoutes()
	return h
}

func (h *Handler) setupRoutes() {
	// API v1 routes
	// Dashboard CRUD
	h.mux.HandleFunc("GET /api/v1/dashboards", h.listDashboards)
	h.mux.HandleFunc("GET /api/v1/dashboards/{slug}", h.getDashboard)
	h.mux.HandleFunc("POST /api/v1/dashboards", h.createDashboard)
	h.mux.HandleFunc("PUT /api/v1/dashboards/{slug}", h.updateDashboard)
	h.mux.HandleFunc("DELETE /api/v1/dashboards/{slug}", h.deleteDashboard)

	// Query execution
	h.mux.HandleFunc("POST /api/v1/query", h.executeQuery)

	// Note: Data source routes are handled by DataSourceHandler in datasource.go

	// API info endpoint
	h.mux.HandleFunc("GET /api", h.apiInfo)
	h.mux.HandleFunc("GET /api/", h.apiInfo)
}

// ServeHTTP implements http.Handler.
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.mux.ServeHTTP(w, r)
}

// client returns the Ent client, or nil if not configured.
func (h *Handler) client() *ent.Client {
	if h.db == nil {
		return nil
	}
	return h.db.Client()
}

// Dashboard handlers

func (h *Handler) listDashboards(w http.ResponseWriter, r *http.Request) {
	client := h.client()
	if client == nil {
		h.jsonError(w, http.StatusServiceUnavailable, "database not configured")
		return
	}

	// Parse query parameters
	visibility := r.URL.Query().Get("visibility")
	archived := r.URL.Query().Get("archived") == "true"

	query := client.Dashboard.Query().
		Where(dashboard.ArchivedEQ(archived))

	if visibility != "" {
		query = query.Where(dashboard.VisibilityEQ(dashboard.Visibility(visibility)))
	}

	dashboards, err := query.
		Order(ent.Desc(dashboard.FieldUpdatedAt)).
		All(r.Context())

	if err != nil {
		h.logger.Error("failed to list dashboards", "error", err)
		h.jsonError(w, http.StatusInternalServerError, "failed to list dashboards")
		return
	}

	// Convert to response format
	result := make([]map[string]any, 0, len(dashboards))
	for _, d := range dashboards {
		result = append(result, map[string]any{
			"slug":        d.Slug,
			"title":       d.Title,
			"description": d.Description,
			"visibility":  d.Visibility,
			"version":     d.Version,
			"archived":    d.Archived,
			"createdAt":   d.CreatedAt,
			"updatedAt":   d.UpdatedAt,
		})
	}

	h.jsonResponse(w, http.StatusOK, map[string]any{
		"dashboards": result,
		"total":      len(result),
	})
}

func (h *Handler) getDashboard(w http.ResponseWriter, r *http.Request) {
	client := h.client()
	if client == nil {
		h.jsonError(w, http.StatusServiceUnavailable, "database not configured")
		return
	}

	slug := r.PathValue("slug")
	if slug == "" {
		h.jsonError(w, http.StatusBadRequest, "slug is required")
		return
	}

	d, err := client.Dashboard.Query().
		Where(dashboard.SlugEQ(slug)).
		Only(r.Context())

	if err != nil {
		if ent.IsNotFound(err) {
			h.jsonError(w, http.StatusNotFound, "dashboard not found")
			return
		}
		h.logger.Error("failed to get dashboard", "error", err, "slug", slug)
		h.jsonError(w, http.StatusInternalServerError, "failed to get dashboard")
		return
	}

	h.jsonResponse(w, http.StatusOK, map[string]any{
		"slug":        d.Slug,
		"title":       d.Title,
		"description": d.Description,
		"definition":  d.Definition,
		"visibility":  d.Visibility,
		"version":     d.Version,
		"archived":    d.Archived,
		"createdAt":   d.CreatedAt,
		"updatedAt":   d.UpdatedAt,
	})
}

func (h *Handler) createDashboard(w http.ResponseWriter, r *http.Request) {
	client := h.client()
	if client == nil {
		h.jsonError(w, http.StatusServiceUnavailable, "database not configured")
		return
	}

	var req struct {
		Slug        string         `json:"slug"`
		Title       string         `json:"title"`
		Description string         `json:"description"`
		Definition  map[string]any `json:"definition"`
		Visibility  string         `json:"visibility"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.jsonError(w, http.StatusBadRequest, "invalid request: "+err.Error())
		return
	}

	if req.Slug == "" || req.Title == "" {
		h.jsonError(w, http.StatusBadRequest, "slug and title are required")
		return
	}

	// Default visibility
	vis := dashboard.VisibilityPrivate
	if req.Visibility != "" {
		vis = dashboard.Visibility(req.Visibility)
	}

	d, err := client.Dashboard.Create().
		SetSlug(req.Slug).
		SetTitle(req.Title).
		SetDescription(req.Description).
		SetDefinition(req.Definition).
		SetVisibility(vis).
		Save(r.Context())

	if err != nil {
		if strings.Contains(err.Error(), "duplicate") || strings.Contains(err.Error(), "unique") {
			h.jsonError(w, http.StatusConflict, "dashboard with this slug already exists")
			return
		}
		h.logger.Error("failed to create dashboard", "error", err)
		h.jsonError(w, http.StatusInternalServerError, "failed to create dashboard")
		return
	}

	h.jsonResponse(w, http.StatusCreated, map[string]any{
		"slug":      d.Slug,
		"title":     d.Title,
		"version":   d.Version,
		"createdAt": d.CreatedAt,
	})
}

func (h *Handler) updateDashboard(w http.ResponseWriter, r *http.Request) {
	client := h.client()
	if client == nil {
		h.jsonError(w, http.StatusServiceUnavailable, "database not configured")
		return
	}

	slug := r.PathValue("slug")
	if slug == "" {
		h.jsonError(w, http.StatusBadRequest, "slug is required")
		return
	}

	var req struct {
		Title       *string        `json:"title"`
		Description *string        `json:"description"`
		Definition  map[string]any `json:"definition"`
		Visibility  *string        `json:"visibility"`
		Archived    *bool          `json:"archived"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.jsonError(w, http.StatusBadRequest, "invalid request: "+err.Error())
		return
	}

	// Get existing dashboard
	d, err := client.Dashboard.Query().
		Where(dashboard.SlugEQ(slug)).
		Only(r.Context())

	if err != nil {
		if ent.IsNotFound(err) {
			h.jsonError(w, http.StatusNotFound, "dashboard not found")
			return
		}
		h.logger.Error("failed to get dashboard", "error", err, "slug", slug)
		h.jsonError(w, http.StatusInternalServerError, "failed to get dashboard")
		return
	}

	// Build update
	update := client.Dashboard.UpdateOne(d)

	if req.Title != nil {
		update = update.SetTitle(*req.Title)
	}
	if req.Description != nil {
		update = update.SetDescription(*req.Description)
	}
	if req.Definition != nil {
		update = update.SetDefinition(req.Definition)
		update = update.SetVersion(d.Version + 1) // Increment version on definition change
	}
	if req.Visibility != nil {
		update = update.SetVisibility(dashboard.Visibility(*req.Visibility))
	}
	if req.Archived != nil {
		update = update.SetArchived(*req.Archived)
	}

	d, err = update.Save(r.Context())
	if err != nil {
		h.logger.Error("failed to update dashboard", "error", err, "slug", slug)
		h.jsonError(w, http.StatusInternalServerError, "failed to update dashboard")
		return
	}

	h.jsonResponse(w, http.StatusOK, map[string]any{
		"slug":      d.Slug,
		"title":     d.Title,
		"version":   d.Version,
		"updatedAt": d.UpdatedAt,
	})
}

func (h *Handler) deleteDashboard(w http.ResponseWriter, r *http.Request) {
	client := h.client()
	if client == nil {
		h.jsonError(w, http.StatusServiceUnavailable, "database not configured")
		return
	}

	slug := r.PathValue("slug")
	if slug == "" {
		h.jsonError(w, http.StatusBadRequest, "slug is required")
		return
	}

	_, err := client.Dashboard.Delete().
		Where(dashboard.SlugEQ(slug)).
		Exec(r.Context())

	if err != nil {
		h.logger.Error("failed to delete dashboard", "error", err, "slug", slug)
		h.jsonError(w, http.StatusInternalServerError, "failed to delete dashboard")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// Query handlers

func (h *Handler) executeQuery(w http.ResponseWriter, r *http.Request) {
	if h.db == nil {
		h.jsonError(w, http.StatusServiceUnavailable, "database not configured")
		return
	}

	var req struct {
		Query      string         `json:"query"`
		Parameters map[string]any `json:"parameters"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.jsonError(w, http.StatusBadRequest, "invalid request: "+err.Error())
		return
	}

	if req.Query == "" {
		h.jsonError(w, http.StatusBadRequest, "query is required")
		return
	}

	result, err := h.db.Query(r.Context(), req.Query, req.Parameters)
	if err != nil {
		h.logger.Error("query execution failed", "error", err, "query", req.Query)
		h.jsonError(w, http.StatusInternalServerError, "query failed: "+err.Error())
		return
	}

	h.jsonResponse(w, http.StatusOK, result)
}

// Note: Data source handlers have been moved to datasource.go for better organization

// API info handler

func (h *Handler) apiInfo(w http.ResponseWriter, _ *http.Request) {
	h.jsonResponse(w, http.StatusOK, map[string]any{
		"name":    "Dashforge API",
		"version": "v1",
		"endpoints": map[string]string{
			"dashboards":  "/api/v1/dashboards",
			"datasources": "/api/v1/datasources",
			"query":       "/api/v1/query",
		},
	})
}

// Helper methods

func (h *Handler) jsonResponse(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		h.logger.Error("failed to encode response", "error", err)
	}
}

func (h *Handler) jsonError(w http.ResponseWriter, status int, message string) {
	h.jsonResponse(w, status, map[string]string{"error": message})
}
