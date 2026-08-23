package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"github.com/grokify/guardsql"
	serveranalytics "github.com/plexusone/uiforge/analytics"
	"github.com/plexusone/uiforge/dashboardir"
)

// AnalyticsHandler serves generic analytics metadata and, over time, query
// execution endpoints backed by configured source providers.
type AnalyticsHandler struct {
	service        *serveranalytics.Service
	logger         *slog.Logger
	mux            *http.ServeMux
	policyProvider GrokifyQLPolicyProvider
}

// NewAnalyticsHandler creates an AnalyticsHandler.
func NewAnalyticsHandler(service *serveranalytics.Service, logger *slog.Logger, policyProvider GrokifyQLPolicyProvider) *AnalyticsHandler {
	if policyProvider == nil {
		policyProvider = staticGrokifyQLPolicyProvider{}
	}
	h := &AnalyticsHandler{
		service:        service,
		logger:         logger,
		mux:            http.NewServeMux(),
		policyProvider: policyProvider,
	}
	h.setupRoutes()
	return h
}

func (h *AnalyticsHandler) setupRoutes() {
	h.mux.HandleFunc("GET /api/v1/analytics/catalog", h.getCatalog)
	h.mux.HandleFunc("POST /api/v1/analytics/query", h.executeQuery)
	h.setupSourceRoutes()
}

// ServeHTTP implements http.Handler.
func (h *AnalyticsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.mux.ServeHTTP(w, r)
}

func (h *AnalyticsHandler) executeQuery(w http.ResponseWriter, r *http.Request) {
	if h.service == nil {
		h.jsonError(w, http.StatusServiceUnavailable, "analytics service not configured")
		return
	}
	var req dashboardir.AnalyticsQueryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.jsonError(w, http.StatusBadRequest, "invalid request: "+err.Error())
		return
	}
	if req.SourceID == "" || req.Query == "" {
		h.jsonError(w, http.StatusBadRequest, "sourceId and query are required")
		return
	}
	if req.Dialect == "" {
		req.Dialect = "grokifyql"
	}
	if strings.EqualFold(req.Dialect, "grokifyql") {
		if err := h.checkGrokifyQLPolicy(r, req); err != nil {
			h.jsonError(w, http.StatusForbidden, err.Error())
			return
		}
	}
	result, err := h.service.Query(r.Context(), req)
	if err != nil {
		if errors.Is(err, serveranalytics.ErrQueryUnsupported) {
			h.jsonError(w, http.StatusNotImplemented, "analytics query execution not configured")
			return
		}
		if h.logger != nil {
			h.logger.Error("failed to execute analytics query", "error", err)
		}
		h.jsonError(w, http.StatusInternalServerError, "failed to execute analytics query")
		return
	}
	h.jsonResponse(w, http.StatusOK, result)
}

func (h *AnalyticsHandler) checkGrokifyQLPolicy(r *http.Request, req dashboardir.AnalyticsQueryRequest) error {
	q, err := guardsql.Parse(req.Query)
	if err != nil {
		return err
	}
	policy, err := h.policyProvider.Policy(r.Context(), req.SourceID)
	if err != nil {
		return err
	}
	if issues := guardsql.CheckPolicy(q, policy); len(issues) > 0 {
		return fmt.Errorf("invalid query policy: %s", issues[0].Message)
	}
	return nil
}

func (h *AnalyticsHandler) getCatalog(w http.ResponseWriter, r *http.Request) {
	if h.service == nil {
		h.jsonError(w, http.StatusServiceUnavailable, "analytics service not configured")
		return
	}
	catalog, err := h.service.Catalog(r.Context())
	if err != nil {
		if h.logger != nil {
			h.logger.Error("failed to load analytics catalog", "error", err)
		}
		h.jsonError(w, http.StatusInternalServerError, "failed to load analytics catalog")
		return
	}
	h.jsonResponse(w, http.StatusOK, catalog)
}

func (h *AnalyticsHandler) jsonResponse(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil && h.logger != nil {
		h.logger.Error("failed to write response", "error", err)
	}
}

func (h *AnalyticsHandler) jsonError(w http.ResponseWriter, status int, message string) {
	h.jsonResponse(w, status, map[string]any{
		"error": message,
	})
}
