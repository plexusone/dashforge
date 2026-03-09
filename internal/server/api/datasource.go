package api

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/plexusone/dashforge/datasource"
	"github.com/plexusone/dashforge/ent"
	entdatasource "github.com/plexusone/dashforge/ent/datasource"
	"github.com/plexusone/dashforge/internal/server/db"
)

// DataSourceHandler handles data source API requests.
type DataSourceHandler struct {
	db       db.Database
	manager  *datasource.Manager
	executor *datasource.QueryExecutor
	logger   *slog.Logger
	mux      *http.ServeMux
}

// NewDataSourceHandler creates a new DataSourceHandler.
func NewDataSourceHandler(database db.Database, manager *datasource.Manager, executor *datasource.QueryExecutor, logger *slog.Logger) *DataSourceHandler {
	h := &DataSourceHandler{
		db:       database,
		manager:  manager,
		executor: executor,
		logger:   logger,
		mux:      http.NewServeMux(),
	}
	h.setupRoutes()
	return h
}

func (h *DataSourceHandler) setupRoutes() {
	// CRUD operations
	h.mux.HandleFunc("GET /api/v1/datasources", h.listDataSources)
	h.mux.HandleFunc("GET /api/v1/datasources/{id}", h.getDataSource)
	h.mux.HandleFunc("POST /api/v1/datasources", h.createDataSource)
	h.mux.HandleFunc("PUT /api/v1/datasources/{id}", h.updateDataSource)
	h.mux.HandleFunc("DELETE /api/v1/datasources/{id}", h.deleteDataSource)

	// Operations
	h.mux.HandleFunc("POST /api/v1/datasources/{id}/test", h.testDataSource)
	h.mux.HandleFunc("POST /api/v1/datasources/{id}/query", h.queryDataSource)
	h.mux.HandleFunc("GET /api/v1/datasources/{id}/schema", h.getDataSourceSchema)

	// Providers
	h.mux.HandleFunc("GET /api/v1/datasources/providers", h.listProviders)
}

// ServeHTTP implements http.Handler.
func (h *DataSourceHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.mux.ServeHTTP(w, r)
}

// client returns the Ent client.
func (h *DataSourceHandler) client() *ent.Client {
	if h.db == nil {
		return nil
	}
	return h.db.Client()
}

// listDataSources handles GET /api/v1/datasources
func (h *DataSourceHandler) listDataSources(w http.ResponseWriter, r *http.Request) {
	client := h.client()
	if client == nil {
		h.jsonError(w, http.StatusServiceUnavailable, "database not configured")
		return
	}

	// Parse query parameters
	activeOnly := r.URL.Query().Get("active") == "true"
	dsType := r.URL.Query().Get("type")

	query := client.DataSource.Query()

	if activeOnly {
		query = query.Where(entdatasource.ActiveEQ(true))
	}
	if dsType != "" {
		query = query.Where(entdatasource.TypeEQ(entdatasource.Type(dsType)))
	}

	datasources, err := query.
		Order(ent.Asc(entdatasource.FieldName)).
		All(r.Context())

	if err != nil {
		h.logger.Error("failed to list data sources", "error", err)
		h.jsonError(w, http.StatusInternalServerError, "failed to list data sources")
		return
	}

	// Convert to response format (exclude sensitive fields)
	result := make([]map[string]any, 0, len(datasources))
	for _, ds := range datasources {
		result = append(result, h.dataSourceToResponse(ds, false))
	}

	h.jsonResponse(w, http.StatusOK, map[string]any{
		"dataSources": result,
		"total":       len(result),
	})
}

// getDataSource handles GET /api/v1/datasources/{id}
func (h *DataSourceHandler) getDataSource(w http.ResponseWriter, r *http.Request) {
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

	ds, err := client.DataSource.Get(r.Context(), id)
	if err != nil {
		if ent.IsNotFound(err) {
			h.jsonError(w, http.StatusNotFound, "data source not found")
			return
		}
		h.logger.Error("failed to get data source", "error", err, "id", id)
		h.jsonError(w, http.StatusInternalServerError, "failed to get data source")
		return
	}

	h.jsonResponse(w, http.StatusOK, h.dataSourceToResponse(ds, false))
}

// createDataSource handles POST /api/v1/datasources
func (h *DataSourceHandler) createDataSource(w http.ResponseWriter, r *http.Request) {
	client := h.client()
	if client == nil {
		h.jsonError(w, http.StatusServiceUnavailable, "database not configured")
		return
	}

	var req struct {
		Name                string `json:"name"`
		Slug                string `json:"slug"`
		Type                string `json:"type"`
		ConnectionURL       string `json:"connectionUrl"`
		ConnectionURLEnv    string `json:"connectionUrlEnv"`
		MaxConnections      int    `json:"maxConnections"`
		QueryTimeoutSeconds int    `json:"queryTimeoutSeconds"`
		ReadOnly            bool   `json:"readOnly"`
		SSLEnabled          bool   `json:"sslEnabled"`
		Active              bool   `json:"active"`
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
	if req.Type == "" {
		h.jsonError(w, http.StatusBadRequest, "type is required")
		return
	}
	if req.ConnectionURL == "" && req.ConnectionURLEnv == "" {
		h.jsonError(w, http.StatusBadRequest, "connectionUrl or connectionUrlEnv is required")
		return
	}

	// Check if provider exists
	if _, ok := datasource.Get(req.Type); !ok {
		h.jsonError(w, http.StatusBadRequest, "unknown provider type: "+req.Type)
		return
	}

	// Set defaults
	if req.MaxConnections == 0 {
		req.MaxConnections = 10
	}
	if req.QueryTimeoutSeconds == 0 {
		req.QueryTimeoutSeconds = 30
	}

	// Create data source
	create := client.DataSource.Create().
		SetName(req.Name).
		SetSlug(req.Slug).
		SetType(entdatasource.Type(req.Type)).
		SetConnectionURL(req.ConnectionURL).
		SetMaxConnections(req.MaxConnections).
		SetQueryTimeoutSeconds(req.QueryTimeoutSeconds).
		SetReadOnly(req.ReadOnly).
		SetSslEnabled(req.SSLEnabled).
		SetActive(req.Active)

	if req.ConnectionURLEnv != "" {
		create = create.SetConnectionURLEnv(req.ConnectionURLEnv)
	}

	ds, err := create.Save(r.Context())
	if err != nil {
		if strings.Contains(err.Error(), "duplicate") || strings.Contains(err.Error(), "unique") {
			h.jsonError(w, http.StatusConflict, "data source with this slug already exists")
			return
		}
		h.logger.Error("failed to create data source", "error", err)
		h.jsonError(w, http.StatusInternalServerError, "failed to create data source")
		return
	}

	h.logger.Info("data source created",
		"id", ds.ID,
		"name", ds.Name,
		"type", ds.Type,
	)

	h.jsonResponse(w, http.StatusCreated, h.dataSourceToResponse(ds, false))
}

// updateDataSource handles PUT /api/v1/datasources/{id}
func (h *DataSourceHandler) updateDataSource(w http.ResponseWriter, r *http.Request) {
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
		Name                *string `json:"name"`
		ConnectionURL       *string `json:"connectionUrl"`
		ConnectionURLEnv    *string `json:"connectionUrlEnv"`
		MaxConnections      *int    `json:"maxConnections"`
		QueryTimeoutSeconds *int    `json:"queryTimeoutSeconds"`
		ReadOnly            *bool   `json:"readOnly"`
		SSLEnabled          *bool   `json:"sslEnabled"`
		Active              *bool   `json:"active"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.jsonError(w, http.StatusBadRequest, "invalid request: "+err.Error())
		return
	}

	// Get existing data source
	ds, err := client.DataSource.Get(r.Context(), id)
	if err != nil {
		if ent.IsNotFound(err) {
			h.jsonError(w, http.StatusNotFound, "data source not found")
			return
		}
		h.logger.Error("failed to get data source", "error", err, "id", id)
		h.jsonError(w, http.StatusInternalServerError, "failed to get data source")
		return
	}

	// Build update
	update := client.DataSource.UpdateOne(ds)

	if req.Name != nil {
		update = update.SetName(*req.Name)
	}
	if req.ConnectionURL != nil {
		update = update.SetConnectionURL(*req.ConnectionURL)
	}
	if req.ConnectionURLEnv != nil {
		update = update.SetConnectionURLEnv(*req.ConnectionURLEnv)
	}
	if req.MaxConnections != nil {
		update = update.SetMaxConnections(*req.MaxConnections)
	}
	if req.QueryTimeoutSeconds != nil {
		update = update.SetQueryTimeoutSeconds(*req.QueryTimeoutSeconds)
	}
	if req.ReadOnly != nil {
		update = update.SetReadOnly(*req.ReadOnly)
	}
	if req.SSLEnabled != nil {
		update = update.SetSslEnabled(*req.SSLEnabled)
	}
	if req.Active != nil {
		update = update.SetActive(*req.Active)
	}

	ds, err = update.Save(r.Context())
	if err != nil {
		h.logger.Error("failed to update data source", "error", err, "id", id)
		h.jsonError(w, http.StatusInternalServerError, "failed to update data source")
		return
	}

	// Close existing connection to force reconnect with new settings
	_ = h.manager.CloseConnection(ds.ID)

	h.logger.Info("data source updated", "id", ds.ID, "name", ds.Name)

	h.jsonResponse(w, http.StatusOK, h.dataSourceToResponse(ds, false))
}

// deleteDataSource handles DELETE /api/v1/datasources/{id}
func (h *DataSourceHandler) deleteDataSource(w http.ResponseWriter, r *http.Request) {
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

	// Close connection first
	_ = h.manager.CloseConnection(id)

	// Delete from database
	err = client.DataSource.DeleteOneID(id).Exec(r.Context())
	if err != nil {
		if ent.IsNotFound(err) {
			h.jsonError(w, http.StatusNotFound, "data source not found")
			return
		}
		h.logger.Error("failed to delete data source", "error", err, "id", id)
		h.jsonError(w, http.StatusInternalServerError, "failed to delete data source")
		return
	}

	h.logger.Info("data source deleted", "id", id)

	w.WriteHeader(http.StatusNoContent)
}

// testDataSource handles POST /api/v1/datasources/{id}/test
func (h *DataSourceHandler) testDataSource(w http.ResponseWriter, r *http.Request) {
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

	// Get data source
	ds, err := client.DataSource.Get(r.Context(), id)
	if err != nil {
		if ent.IsNotFound(err) {
			h.jsonError(w, http.StatusNotFound, "data source not found")
			return
		}
		h.logger.Error("failed to get data source", "error", err, "id", id)
		h.jsonError(w, http.StatusInternalServerError, "failed to get data source")
		return
	}

	// Test connection
	start := time.Now()
	err = h.manager.TestConnection(r.Context(), h.entToConfig(ds))
	elapsed := time.Since(start)

	if err != nil {
		h.logger.Warn("data source connection test failed",
			"id", id,
			"name", ds.Name,
			"error", err,
		)
		h.jsonResponse(w, http.StatusOK, map[string]any{
			"success":    false,
			"error":      err.Error(),
			"durationMs": elapsed.Milliseconds(),
		})
		return
	}

	// Update last connected timestamp
	_, _ = client.DataSource.UpdateOne(ds).
		SetLastConnectedAt(time.Now()).
		Save(r.Context())

	h.jsonResponse(w, http.StatusOK, map[string]any{
		"success":    true,
		"durationMs": elapsed.Milliseconds(),
	})
}

// queryDataSource handles POST /api/v1/datasources/{id}/query
func (h *DataSourceHandler) queryDataSource(w http.ResponseWriter, r *http.Request) {
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
		Query      string         `json:"query"`
		Parameters map[string]any `json:"parameters"`
		MaxRows    int            `json:"maxRows"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.jsonError(w, http.StatusBadRequest, "invalid request: "+err.Error())
		return
	}

	if req.Query == "" {
		h.jsonError(w, http.StatusBadRequest, "query is required")
		return
	}

	// Get data source
	ds, err := client.DataSource.Get(r.Context(), id)
	if err != nil {
		if ent.IsNotFound(err) {
			h.jsonError(w, http.StatusNotFound, "data source not found")
			return
		}
		h.logger.Error("failed to get data source", "error", err, "id", id)
		h.jsonError(w, http.StatusInternalServerError, "failed to get data source")
		return
	}

	if !ds.Active {
		h.jsonError(w, http.StatusBadRequest, "data source is not active")
		return
	}

	// Execute query
	result, err := h.executor.Execute(r.Context(), datasource.QueryRequest{
		DataSource: h.entToConfig(ds),
		Query:      req.Query,
		Parameters: req.Parameters,
		MaxRows:    req.MaxRows,
	})

	if err != nil {
		h.logger.Warn("query execution failed",
			"datasource_id", id,
			"error", err,
		)
		h.jsonError(w, http.StatusInternalServerError, "query failed: "+err.Error())
		return
	}

	h.jsonResponse(w, http.StatusOK, result)
}

// getDataSourceSchema handles GET /api/v1/datasources/{id}/schema
func (h *DataSourceHandler) getDataSourceSchema(w http.ResponseWriter, r *http.Request) {
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

	// Get data source
	ds, err := client.DataSource.Get(r.Context(), id)
	if err != nil {
		if ent.IsNotFound(err) {
			h.jsonError(w, http.StatusNotFound, "data source not found")
			return
		}
		h.logger.Error("failed to get data source", "error", err, "id", id)
		h.jsonError(w, http.StatusInternalServerError, "failed to get data source")
		return
	}

	// Parse query parameters
	schema := r.URL.Query().Get("schema")
	includeColumns := r.URL.Query().Get("columns") == "true"
	tableFilter := r.URL.Query().Get("filter")

	// Get schema
	result, err := h.executor.GetSchema(r.Context(), datasource.SchemaRequest{
		DataSource:     h.entToConfig(ds),
		Schema:         schema,
		IncludeColumns: includeColumns,
		TableFilter:    tableFilter,
	})

	if err != nil {
		h.logger.Warn("schema retrieval failed",
			"datasource_id", id,
			"error", err,
		)
		h.jsonError(w, http.StatusInternalServerError, "failed to get schema: "+err.Error())
		return
	}

	h.jsonResponse(w, http.StatusOK, result)
}

// listProviders handles GET /api/v1/datasources/providers
func (h *DataSourceHandler) listProviders(w http.ResponseWriter, _ *http.Request) {
	providers := datasource.Available()

	result := make([]map[string]any, 0, len(providers))
	for _, name := range providers {
		p, _ := datasource.Get(name)
		caps := p.Capabilities()
		result = append(result, map[string]any{
			"name":         name,
			"capabilities": caps,
		})
	}

	h.jsonResponse(w, http.StatusOK, map[string]any{
		"providers": result,
	})
}

// Helper methods

func (h *DataSourceHandler) parseID(r *http.Request) (int, error) {
	idStr := r.PathValue("id")
	return strconv.Atoi(idStr)
}

func (h *DataSourceHandler) entToConfig(ds *ent.DataSource) datasource.DataSourceConfig {
	return datasource.DataSourceConfig{
		ID:                  ds.ID,
		Name:                ds.Name,
		Slug:                ds.Slug,
		Type:                string(ds.Type),
		ConnectionURL:       ds.ConnectionURL,
		ConnectionURLEnv:    ds.ConnectionURLEnv,
		MaxConnections:      ds.MaxConnections,
		QueryTimeoutSeconds: ds.QueryTimeoutSeconds,
		ReadOnly:            ds.ReadOnly,
		SSLEnabled:          ds.SslEnabled,
	}
}

func (h *DataSourceHandler) dataSourceToResponse(ds *ent.DataSource, includeSecrets bool) map[string]any {
	resp := map[string]any{
		"id":                  ds.ID,
		"name":                ds.Name,
		"slug":                ds.Slug,
		"type":                ds.Type,
		"maxConnections":      ds.MaxConnections,
		"queryTimeoutSeconds": ds.QueryTimeoutSeconds,
		"readOnly":            ds.ReadOnly,
		"sslEnabled":          ds.SslEnabled,
		"active":              ds.Active,
		"createdAt":           ds.CreatedAt,
		"updatedAt":           ds.UpdatedAt,
	}

	if ds.LastConnectedAt != nil {
		resp["lastConnectedAt"] = ds.LastConnectedAt
	}

	if ds.ConnectionURLEnv != "" {
		resp["connectionUrlEnv"] = ds.ConnectionURLEnv
	}

	// Only include sensitive fields if explicitly requested
	if includeSecrets {
		resp["connectionUrl"] = ds.ConnectionURL
	}

	return resp
}

func (h *DataSourceHandler) jsonResponse(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		h.logger.Error("failed to encode response", "error", err)
	}
}

func (h *DataSourceHandler) jsonError(w http.ResponseWriter, status int, message string) {
	h.jsonResponse(w, status, map[string]string{"error": message})
}
