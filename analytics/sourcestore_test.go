package analytics

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func newTestFileStore(t *testing.T) SourceStore {
	t.Helper()
	store, err := NewSourceFileStore(filepath.Join(t.TempDir(), "analytics-sources.json"))
	if err != nil {
		t.Fatal(err)
	}
	return store
}

func TestSourceFileStoreCRUD(t *testing.T) {
	ctx := context.Background()
	store := newTestFileStore(t)

	sources, err := store.List(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(sources) != 0 {
		t.Fatalf("expected empty store, got %d sources", len(sources))
	}

	cfg := validSourceConfig()
	saved, err := store.Save(ctx, cfg)
	if err != nil {
		t.Fatal(err)
	}
	if saved.CreatedAt.IsZero() || saved.UpdatedAt.IsZero() {
		t.Fatal("expected timestamps to be set")
	}

	got, err := store.Get(ctx, cfg.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.Name != cfg.Name || got.Connector != cfg.Connector || got.DSNRef != cfg.DSNRef {
		t.Fatalf("round-trip mismatch: %+v", got)
	}

	// Update keeps CreatedAt.
	got.Name = "Renamed"
	updated, err := store.Save(ctx, got)
	if err != nil {
		t.Fatal(err)
	}
	if !updated.CreatedAt.Equal(saved.CreatedAt) {
		t.Fatal("update must preserve CreatedAt")
	}
	if updated.Name != "Renamed" {
		t.Fatalf("update not applied: %+v", updated)
	}

	sources, err = store.List(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(sources) != 1 {
		t.Fatalf("expected 1 source after update, got %d", len(sources))
	}

	if err := store.Delete(ctx, cfg.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := store.Get(ctx, cfg.ID); !errors.Is(err, ErrSourceNotFound) {
		t.Fatalf("expected ErrSourceNotFound, got %v", err)
	}
	if err := store.Delete(ctx, cfg.ID); !errors.Is(err, ErrSourceNotFound) {
		t.Fatalf("expected ErrSourceNotFound on double delete, got %v", err)
	}
}

func TestSourceFileStoreRejectsInvalid(t *testing.T) {
	ctx := context.Background()
	store := newTestFileStore(t)

	cfg := validSourceConfig()
	cfg.DSNRef = "root:@tcp(127.0.0.1:13307)/db"
	if _, err := store.Save(ctx, cfg); err == nil {
		t.Fatal("expected raw DSN to be rejected")
	}
}

func TestSourceFileStorePersistsAcrossInstances(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "analytics-sources.json")

	store1, err := NewSourceFileStore(path)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := store1.Save(ctx, validSourceConfig()); err != nil {
		t.Fatal(err)
	}

	// A raw DSN must never appear on disk.
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "env://UIFORGE_OMNIROADMAP_DSN") {
		t.Fatal("expected dsnRef in store file")
	}

	store2, err := NewSourceFileStore(path)
	if err != nil {
		t.Fatal(err)
	}
	sources, err := store2.List(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(sources) != 1 || sources[0].ID != "omniroadmap-local" {
		t.Fatalf("expected persisted source, got %+v", sources)
	}
}
