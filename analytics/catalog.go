// Package analytics combines queryable-source catalogs for DashForge server mode.
package analytics

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sync"

	"github.com/plexusone/dashforge/dashboardir"
	"github.com/plexusone/omnivault"
)

// ErrQueryUnsupported is returned when no configured analytics provider can
// execute query requests yet.
var ErrQueryUnsupported = errors.New("analytics query execution not configured")

// CatalogProvider supplies one source catalog to DashForge's analytics layer.
// Application connectors such as OmniRoadmap and VisionStudio should implement
// this interface without depending on DashForge dashboard storage.
type CatalogProvider interface {
	Catalog(ctx context.Context) (dashboardir.AnalyticsCatalog, error)
	Close() error
}

// QueryProvider executes read-only analytics queries for one or more sources.
type QueryProvider interface {
	CatalogProvider
	Query(ctx context.Context, req dashboardir.AnalyticsQueryRequest) (dashboardir.AnalyticsQueryResult, error)
}

// SourceStatus reports a managed source's configuration and runtime state.
type SourceStatus struct {
	SourceConfig
	// Status is "connected", "error", or "disabled".
	Status string `json:"status"`
	// Error carries a sanitized connect error when Status is "error".
	Error string `json:"error,omitempty"`
}

// managedSource pairs a stored config with its live provider (nil when
// disabled or errored).
type managedSource struct {
	config   SourceConfig
	provider QueryProvider
	err      error
}

// Service serves a combined analytics catalog. Sources come from two places:
// static providers fixed at construction (tests, embedded deployments), and
// managed sources persisted in a SourceStore and connected through the
// connector registry with OmniVault-resolved DSNs.
type Service struct {
	mu       sync.RWMutex
	static   []CatalogProvider
	resolver *omnivault.Resolver
	store    SourceStore
	logger   *slog.Logger
	sources  map[string]*managedSource
}

// NewService creates an analytics catalog service over fixed providers.
func NewService(providers ...CatalogProvider) *Service {
	return &Service{static: providers, sources: map[string]*managedSource{}}
}

// NewManagedService creates a service whose sources are persisted in store and
// connected via the connector registry, resolving dsnRefs through resolver.
func NewManagedService(store SourceStore, resolver *omnivault.Resolver, logger *slog.Logger) *Service {
	return &Service{
		store:    store,
		resolver: resolver,
		logger:   logger,
		sources:  map[string]*managedSource{},
	}
}

// LoadAll connects every enabled stored source. A source that fails to connect
// is recorded with its error and logged, not fatal: the server starts and the
// failure surfaces in Sources() status.
func (s *Service) LoadAll(ctx context.Context) error {
	if s == nil || s.store == nil {
		return nil
	}
	configs, err := s.store.List(ctx)
	if err != nil {
		return fmt.Errorf("loading analytics sources: %w", err)
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, cfg := range configs {
		s.connectLocked(ctx, cfg)
	}
	return nil
}

// connectLocked (re)connects one source and records the outcome. Caller holds
// the write lock.
func (s *Service) connectLocked(ctx context.Context, cfg SourceConfig) {
	if existing, ok := s.sources[cfg.ID]; ok && existing.provider != nil {
		if err := existing.provider.Close(); err != nil && s.logger != nil {
			s.logger.Error("closing analytics source", "source", cfg.ID, "error", err)
		}
	}
	ms := &managedSource{config: cfg}
	s.sources[cfg.ID] = ms
	if !cfg.Enabled {
		return
	}
	provider, err := s.openProvider(ctx, cfg)
	if err != nil {
		ms.err = err
		if s.logger != nil {
			s.logger.Warn("analytics source unavailable", "source", cfg.ID, "connector", cfg.Connector, "error", err)
		}
		return
	}
	ms.provider = provider
	if s.logger != nil {
		s.logger.Info("analytics source connected", "source", cfg.ID, "connector", cfg.Connector)
	}
}

// openProvider resolves the dsnRef and dials the connector. The resolved DSN
// stays in this function's scope.
func (s *Service) openProvider(ctx context.Context, cfg SourceConfig) (QueryProvider, error) {
	factory, ok := LookupConnector(cfg.Connector)
	if !ok {
		return nil, fmt.Errorf("unknown connector %q", cfg.Connector)
	}
	dsn, err := ResolveDSN(ctx, s.resolver, cfg.DSNRef)
	if err != nil {
		return nil, err
	}
	provider, err := factory(dsn)
	if err != nil {
		// Deliberately not wrapping the driver error into API-visible state:
		// driver errors can echo the DSN. Callers see a generic message; the
		// full error is only logged.
		if s.logger != nil {
			s.logger.Error("connecting analytics source", "source", cfg.ID, "connector", cfg.Connector, "error", err)
		}
		return nil, fmt.Errorf("connecting %s source %q failed (see server log)", cfg.Connector, cfg.ID)
	}
	return provider, nil
}

// AddSource validates, connects, persists, and registers a new source.
func (s *Service) AddSource(ctx context.Context, cfg SourceConfig) (SourceConfig, error) {
	if s.store == nil {
		return SourceConfig{}, fmt.Errorf("analytics source management not configured")
	}
	if err := cfg.Validate(); err != nil {
		return SourceConfig{}, err
	}
	if _, err := s.store.Get(ctx, cfg.ID); err == nil {
		return SourceConfig{}, fmt.Errorf("analytics source %q already exists", cfg.ID)
	} else if !errors.Is(err, ErrSourceNotFound) {
		return SourceConfig{}, err
	}
	if cfg.Enabled {
		provider, err := s.openProvider(ctx, cfg)
		if err != nil {
			return SourceConfig{}, err
		}
		if err := provider.Close(); err != nil && s.logger != nil {
			s.logger.Error("closing probe connection", "source", cfg.ID, "error", err)
		}
	}
	saved, err := s.store.Save(ctx, cfg)
	if err != nil {
		return SourceConfig{}, err
	}
	s.mu.Lock()
	s.connectLocked(ctx, saved)
	s.mu.Unlock()
	return saved, nil
}

// UpdateSource persists changes to an existing source and reconnects it. The
// source ID is immutable; unknown IDs return ErrSourceNotFound.
func (s *Service) UpdateSource(ctx context.Context, cfg SourceConfig) (SourceConfig, error) {
	if s.store == nil {
		return SourceConfig{}, fmt.Errorf("analytics source management not configured")
	}
	if err := cfg.Validate(); err != nil {
		return SourceConfig{}, err
	}
	if _, err := s.store.Get(ctx, cfg.ID); err != nil {
		return SourceConfig{}, err
	}
	saved, err := s.store.Save(ctx, cfg)
	if err != nil {
		return SourceConfig{}, err
	}
	s.mu.Lock()
	s.connectLocked(ctx, saved)
	s.mu.Unlock()
	return saved, nil
}

// RemoveSource closes and deletes a source.
func (s *Service) RemoveSource(ctx context.Context, id string) error {
	if s.store == nil {
		return fmt.Errorf("analytics source management not configured")
	}
	if err := s.store.Delete(ctx, id); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if ms, ok := s.sources[id]; ok {
		if ms.provider != nil {
			if err := ms.provider.Close(); err != nil && s.logger != nil {
				s.logger.Error("closing analytics source", "source", id, "error", err)
			}
		}
		delete(s.sources, id)
	}
	return nil
}

// TestSource resolves, connects, and fetches the catalog for a candidate
// config without persisting anything.
func (s *Service) TestSource(ctx context.Context, cfg SourceConfig) error {
	if err := cfg.Validate(); err != nil {
		return err
	}
	provider, err := s.openProvider(ctx, cfg)
	if err != nil {
		return err
	}
	defer func() {
		if cerr := provider.Close(); cerr != nil && s.logger != nil {
			s.logger.Error("closing test connection", "source", cfg.ID, "error", cerr)
		}
	}()
	if _, err := provider.Catalog(ctx); err != nil {
		if s.logger != nil {
			s.logger.Error("testing analytics source catalog", "source", cfg.ID, "error", err)
		}
		return fmt.Errorf("connected, but loading catalog failed (see server log)")
	}
	return nil
}

// Sources returns all managed source configs with runtime status.
func (s *Service) Sources(ctx context.Context) ([]SourceStatus, error) {
	if s == nil || s.store == nil {
		return []SourceStatus{}, nil
	}
	configs, err := s.store.List(ctx)
	if err != nil {
		return nil, err
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	statuses := make([]SourceStatus, 0, len(configs))
	for _, cfg := range configs {
		status := SourceStatus{SourceConfig: cfg, Status: "disabled"}
		if ms, ok := s.sources[cfg.ID]; ok {
			switch {
			case ms.provider != nil:
				status.Status = "connected"
			case ms.err != nil:
				status.Status = "error"
				status.Error = ms.err.Error()
			}
		}
		statuses = append(statuses, status)
	}
	return statuses, nil
}

// HasProviders reports whether the service can return any connected source.
func (s *Service) HasProviders() bool {
	if s == nil {
		return false
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	if len(s.static) > 0 {
		return true
	}
	for _, ms := range s.sources {
		if ms.provider != nil {
			return true
		}
	}
	return false
}

// Catalog returns a combined catalog for all connected analytics sources.
// Managed sources are re-tagged with their configured ID and name so IDs stay
// stable and unique even with several sources on the same connector.
func (s *Service) Catalog(ctx context.Context) (dashboardir.AnalyticsCatalog, error) {
	if s == nil || !s.HasProviders() {
		return dashboardir.AnalyticsCatalog{
			ID:          "dashforge",
			Name:        "DashForge",
			Description: "No analytics sources configured.",
			Sources:     []dashboardir.AnalyticsSource{},
		}, nil
	}

	out := dashboardir.AnalyticsCatalog{
		ID:          "dashforge",
		Name:        "DashForge Analytics",
		Description: "Combined catalog for connected analytics sources.",
	}
	s.mu.RLock()
	static := make([]CatalogProvider, len(s.static))
	copy(static, s.static)
	managed := make([]*managedSource, 0, len(s.sources))
	for _, ms := range s.sources {
		if ms.provider != nil {
			managed = append(managed, ms)
		}
	}
	s.mu.RUnlock()

	for i, provider := range static {
		catalog, err := provider.Catalog(ctx)
		if err != nil {
			return dashboardir.AnalyticsCatalog{}, fmt.Errorf("loading analytics catalog %d: %w", i, err)
		}
		out.Sources = append(out.Sources, catalog.Sources...)
	}
	for _, ms := range managed {
		catalog, err := ms.provider.Catalog(ctx)
		if err != nil {
			return dashboardir.AnalyticsCatalog{}, fmt.Errorf("loading analytics catalog %q: %w", ms.config.ID, err)
		}
		out.Sources = append(out.Sources, retagSources(catalog.Sources, ms.config)...)
	}
	return out, nil
}

// retagSources stamps the configured source ID and name onto a provider's
// catalog. A single-source catalog takes the config identity directly; a
// multi-source catalog gets the config ID as a prefix to stay unique.
func retagSources(sources []dashboardir.AnalyticsSource, cfg SourceConfig) []dashboardir.AnalyticsSource {
	if len(sources) == 1 {
		sources[0].ID = cfg.ID
		sources[0].Name = cfg.Name
		return sources
	}
	for i := range sources {
		sources[i].ID = cfg.ID + ":" + sources[i].ID
		sources[i].Name = cfg.Name + " — " + sources[i].Name
	}
	return sources
}

// Query executes a read-only analytics query against the matching provider.
func (s *Service) Query(ctx context.Context, req dashboardir.AnalyticsQueryRequest) (dashboardir.AnalyticsQueryResult, error) {
	if s == nil {
		return dashboardir.AnalyticsQueryResult{}, ErrQueryUnsupported
	}

	s.mu.RLock()
	var target QueryProvider
	for _, ms := range s.sources {
		if ms.provider == nil {
			continue
		}
		if ms.config.ID == req.SourceID {
			target = ms.provider
			break
		}
	}
	static := make([]CatalogProvider, len(s.static))
	copy(static, s.static)
	s.mu.RUnlock()

	if target != nil {
		return target.Query(ctx, req)
	}

	for _, provider := range static {
		queryProvider, ok := provider.(QueryProvider)
		if !ok {
			continue
		}
		catalog, err := provider.Catalog(ctx)
		if err != nil {
			return dashboardir.AnalyticsQueryResult{}, err
		}
		for _, source := range catalog.Sources {
			if source.ID == req.SourceID {
				return queryProvider.Query(ctx, req)
			}
		}
	}
	return dashboardir.AnalyticsQueryResult{}, ErrQueryUnsupported
}

// Close releases all provider resources.
func (s *Service) Close() error {
	if s == nil {
		return nil
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	var firstErr error
	for _, provider := range s.static {
		if err := provider.Close(); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	for _, ms := range s.sources {
		if ms.provider == nil {
			continue
		}
		if err := ms.provider.Close(); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	return firstErr
}
