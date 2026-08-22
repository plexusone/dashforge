package analytics

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// ErrSourceNotFound is returned when a source config does not exist.
var ErrSourceNotFound = errors.New("analytics source not found")

// SourceStore persists analytics source configurations. Implementations hold
// secret references only, never resolved credentials.
type SourceStore interface {
	List(ctx context.Context) ([]SourceConfig, error)
	Get(ctx context.Context, id string) (SourceConfig, error)
	// Save creates or updates a config and returns the stored value with
	// timestamps applied.
	Save(ctx context.Context, cfg SourceConfig) (SourceConfig, error)
	Delete(ctx context.Context, id string) error
}

// DefaultSourceStorePath is the JSON file used when no metadata database is
// configured, alongside the saved-question store.
const DefaultSourceStorePath = ".uiforge/analytics-sources.json"

// sourceFileStore is a JSON-file SourceStore for deployments without a
// metadata database. Safe to keep on disk: it contains dsnRef references,
// not credentials.
type sourceFileStore struct {
	path string
	mu   sync.Mutex
}

// NewSourceFileStore creates a JSON-file-backed SourceStore at path,
// defaulting to DefaultSourceStorePath.
func NewSourceFileStore(path string) (SourceStore, error) {
	if path == "" {
		path = DefaultSourceStorePath
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, fmt.Errorf("creating analytics source store directory: %w", err)
	}
	return &sourceFileStore{path: path}, nil
}

func (s *sourceFileStore) List(_ context.Context) ([]SourceConfig, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.readLocked()
}

func (s *sourceFileStore) Get(_ context.Context, id string) (SourceConfig, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	sources, err := s.readLocked()
	if err != nil {
		return SourceConfig{}, err
	}
	for _, cfg := range sources {
		if cfg.ID == id {
			return cfg, nil
		}
	}
	return SourceConfig{}, ErrSourceNotFound
}

func (s *sourceFileStore) Save(_ context.Context, cfg SourceConfig) (SourceConfig, error) {
	if err := cfg.Validate(); err != nil {
		return SourceConfig{}, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	sources, err := s.readLocked()
	if err != nil {
		return SourceConfig{}, err
	}
	now := time.Now().UTC()
	cfg.UpdatedAt = now
	replaced := false
	for i, existing := range sources {
		if existing.ID == cfg.ID {
			cfg.CreatedAt = existing.CreatedAt
			sources[i] = cfg
			replaced = true
			break
		}
	}
	if !replaced {
		cfg.CreatedAt = now
		sources = append(sources, cfg)
	}
	if err := s.writeLocked(sources); err != nil {
		return SourceConfig{}, err
	}
	return cfg, nil
}

func (s *sourceFileStore) Delete(_ context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	sources, err := s.readLocked()
	if err != nil {
		return err
	}
	kept := sources[:0]
	for _, cfg := range sources {
		if cfg.ID != id {
			kept = append(kept, cfg)
		}
	}
	if len(kept) == len(sources) {
		return ErrSourceNotFound
	}
	return s.writeLocked(kept)
}

func (s *sourceFileStore) readLocked() ([]SourceConfig, error) {
	data, err := os.ReadFile(s.path)
	if err != nil {
		if os.IsNotExist(err) {
			return []SourceConfig{}, nil
		}
		return nil, fmt.Errorf("reading analytics source store: %w", err)
	}
	var sources []SourceConfig
	if err := json.Unmarshal(data, &sources); err != nil {
		return nil, fmt.Errorf("parsing analytics source store: %w", err)
	}
	return sources, nil
}

func (s *sourceFileStore) writeLocked(sources []SourceConfig) error {
	data, err := json.MarshalIndent(sources, "", "  ")
	if err != nil {
		return fmt.Errorf("encoding analytics source store: %w", err)
	}
	if err := os.WriteFile(s.path, append(data, '\n'), 0o644); err != nil {
		return fmt.Errorf("writing analytics source store: %w", err)
	}
	return nil
}
