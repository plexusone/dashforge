package analytics

import (
	"context"
	"fmt"

	"github.com/plexusone/dashforge/ent"
	entanalyticssource "github.com/plexusone/dashforge/ent/analyticssource"
)

// sourceEntStore is a metadata-database SourceStore backed by the ent
// AnalyticsSource schema.
type sourceEntStore struct {
	client *ent.Client
}

// NewSourceEntStore creates a SourceStore backed by the DashForge metadata
// database.
func NewSourceEntStore(client *ent.Client) (SourceStore, error) {
	if client == nil {
		return nil, fmt.Errorf("ent client is required")
	}
	return &sourceEntStore{client: client}, nil
}

func (s *sourceEntStore) List(ctx context.Context) ([]SourceConfig, error) {
	rows, err := s.client.AnalyticsSource.Query().All(ctx)
	if err != nil {
		return nil, fmt.Errorf("listing analytics sources: %w", err)
	}
	sources := make([]SourceConfig, 0, len(rows))
	for _, row := range rows {
		sources = append(sources, entToSourceConfig(row))
	}
	return sources, nil
}

func (s *sourceEntStore) Get(ctx context.Context, id string) (SourceConfig, error) {
	row, err := s.client.AnalyticsSource.Query().
		Where(entanalyticssource.Slug(id)).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return SourceConfig{}, ErrSourceNotFound
		}
		return SourceConfig{}, fmt.Errorf("getting analytics source %q: %w", id, err)
	}
	return entToSourceConfig(row), nil
}

func (s *sourceEntStore) Save(ctx context.Context, cfg SourceConfig) (SourceConfig, error) {
	if err := cfg.Validate(); err != nil {
		return SourceConfig{}, err
	}
	existing, err := s.client.AnalyticsSource.Query().
		Where(entanalyticssource.Slug(cfg.ID)).
		Only(ctx)
	if err != nil && !ent.IsNotFound(err) {
		return SourceConfig{}, fmt.Errorf("looking up analytics source %q: %w", cfg.ID, err)
	}
	if existing != nil {
		row, err := existing.Update().
			SetName(cfg.Name).
			SetConnector(cfg.Connector).
			SetDsnRef(cfg.DSNRef).
			SetEnabled(cfg.Enabled).
			Save(ctx)
		if err != nil {
			return SourceConfig{}, fmt.Errorf("updating analytics source %q: %w", cfg.ID, err)
		}
		return entToSourceConfig(row), nil
	}
	row, err := s.client.AnalyticsSource.Create().
		SetSlug(cfg.ID).
		SetName(cfg.Name).
		SetConnector(cfg.Connector).
		SetDsnRef(cfg.DSNRef).
		SetEnabled(cfg.Enabled).
		Save(ctx)
	if err != nil {
		return SourceConfig{}, fmt.Errorf("creating analytics source %q: %w", cfg.ID, err)
	}
	return entToSourceConfig(row), nil
}

func (s *sourceEntStore) Delete(ctx context.Context, id string) error {
	n, err := s.client.AnalyticsSource.Delete().
		Where(entanalyticssource.Slug(id)).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("deleting analytics source %q: %w", id, err)
	}
	if n == 0 {
		return ErrSourceNotFound
	}
	return nil
}

func entToSourceConfig(row *ent.AnalyticsSource) SourceConfig {
	return SourceConfig{
		ID:        row.Slug,
		Name:      row.Name,
		Connector: row.Connector,
		DSNRef:    row.DsnRef,
		Enabled:   row.Enabled,
		CreatedAt: row.CreatedAt,
		UpdatedAt: row.UpdatedAt,
	}
}
