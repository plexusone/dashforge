package api

import (
	"encoding/json"
	"errors"
	"net/http"

	serveranalytics "github.com/plexusone/uiforge/internal/server/analytics"
)

// setupSourceRoutes registers analytics source management endpoints.
func (h *AnalyticsHandler) setupSourceRoutes() {
	h.mux.HandleFunc("GET /api/v1/analytics/sources", h.listSources)
	h.mux.HandleFunc("POST /api/v1/analytics/sources", h.createSource)
	h.mux.HandleFunc("PUT /api/v1/analytics/sources/{id}", h.updateSource)
	h.mux.HandleFunc("DELETE /api/v1/analytics/sources/{id}", h.deleteSource)
	h.mux.HandleFunc("POST /api/v1/analytics/sources/test", h.testSource)
	h.mux.HandleFunc("GET /api/v1/analytics/connectors", h.listConnectors)
}

func (h *AnalyticsHandler) listSources(w http.ResponseWriter, r *http.Request) {
	if h.service == nil {
		h.jsonError(w, http.StatusServiceUnavailable, "analytics service not configured")
		return
	}
	sources, err := h.service.Sources(r.Context())
	if err != nil {
		if h.logger != nil {
			h.logger.Error("failed to list analytics sources", "error", err)
		}
		h.jsonError(w, http.StatusInternalServerError, "failed to list analytics sources")
		return
	}
	h.jsonResponse(w, http.StatusOK, map[string]any{"sources": sources})
}

func (h *AnalyticsHandler) createSource(w http.ResponseWriter, r *http.Request) {
	if h.service == nil {
		h.jsonError(w, http.StatusServiceUnavailable, "analytics service not configured")
		return
	}
	var cfg serveranalytics.SourceConfig
	if err := json.NewDecoder(r.Body).Decode(&cfg); err != nil {
		h.jsonError(w, http.StatusBadRequest, "invalid request: "+err.Error())
		return
	}
	saved, err := h.service.AddSource(r.Context(), cfg)
	if err != nil {
		h.jsonError(w, http.StatusBadRequest, err.Error())
		return
	}
	h.jsonResponse(w, http.StatusCreated, saved)
}

func (h *AnalyticsHandler) updateSource(w http.ResponseWriter, r *http.Request) {
	if h.service == nil {
		h.jsonError(w, http.StatusServiceUnavailable, "analytics service not configured")
		return
	}
	var cfg serveranalytics.SourceConfig
	if err := json.NewDecoder(r.Body).Decode(&cfg); err != nil {
		h.jsonError(w, http.StatusBadRequest, "invalid request: "+err.Error())
		return
	}
	cfg.ID = r.PathValue("id")
	saved, err := h.service.UpdateSource(r.Context(), cfg)
	if err != nil {
		if errors.Is(err, serveranalytics.ErrSourceNotFound) {
			h.jsonError(w, http.StatusNotFound, "analytics source not found")
			return
		}
		h.jsonError(w, http.StatusBadRequest, err.Error())
		return
	}
	h.jsonResponse(w, http.StatusOK, saved)
}

func (h *AnalyticsHandler) deleteSource(w http.ResponseWriter, r *http.Request) {
	if h.service == nil {
		h.jsonError(w, http.StatusServiceUnavailable, "analytics service not configured")
		return
	}
	if err := h.service.RemoveSource(r.Context(), r.PathValue("id")); err != nil {
		if errors.Is(err, serveranalytics.ErrSourceNotFound) {
			h.jsonError(w, http.StatusNotFound, "analytics source not found")
			return
		}
		if h.logger != nil {
			h.logger.Error("failed to delete analytics source", "error", err)
		}
		h.jsonError(w, http.StatusInternalServerError, "failed to delete analytics source")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *AnalyticsHandler) testSource(w http.ResponseWriter, r *http.Request) {
	if h.service == nil {
		h.jsonError(w, http.StatusServiceUnavailable, "analytics service not configured")
		return
	}
	var cfg serveranalytics.SourceConfig
	if err := json.NewDecoder(r.Body).Decode(&cfg); err != nil {
		h.jsonError(w, http.StatusBadRequest, "invalid request: "+err.Error())
		return
	}
	if err := h.service.TestSource(r.Context(), cfg); err != nil {
		h.jsonResponse(w, http.StatusOK, map[string]any{"ok": false, "error": err.Error()})
		return
	}
	h.jsonResponse(w, http.StatusOK, map[string]any{"ok": true})
}

func (h *AnalyticsHandler) listConnectors(w http.ResponseWriter, _ *http.Request) {
	h.jsonResponse(w, http.StatusOK, map[string]any{"connectors": serveranalytics.Connectors()})
}
