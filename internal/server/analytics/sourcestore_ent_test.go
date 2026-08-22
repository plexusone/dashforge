package analytics

import (
	"context"
	"errors"
	"testing"

	_ "github.com/mattn/go-sqlite3"

	"github.com/plexusone/uiforge/ent/enttest"
)

func newTestEntStore(t *testing.T) SourceStore {
	t.Helper()
	client := enttest.Open(t, "sqlite3", "file:ent?mode=memory&cache=shared&_fk=1")
	t.Cleanup(func() {
		if err := client.Close(); err != nil {
			t.Errorf("closing ent client: %v", err)
		}
	})
	store, err := NewSourceEntStore(client)
	if err != nil {
		t.Fatal(err)
	}
	return store
}

func TestSourceEntStoreCRUD(t *testing.T) {
	ctx := context.Background()
	store := newTestEntStore(t)

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
	if got.Name != cfg.Name || got.Connector != cfg.Connector || got.DSNRef != cfg.DSNRef || !got.Enabled {
		t.Fatalf("round-trip mismatch: %+v", got)
	}

	// Update by slug, not insert.
	got.Name = "Renamed"
	got.Enabled = false
	if _, err := store.Save(ctx, got); err != nil {
		t.Fatal(err)
	}
	sources, err := store.List(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(sources) != 1 {
		t.Fatalf("expected 1 source after update, got %d", len(sources))
	}
	if sources[0].Name != "Renamed" || sources[0].Enabled {
		t.Fatalf("update not applied: %+v", sources[0])
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

func TestSourceEntStoreRejectsInvalid(t *testing.T) {
	ctx := context.Background()
	store := newTestEntStore(t)

	cfg := validSourceConfig()
	cfg.DSNRef = "root:@tcp(127.0.0.1:13307)/db"
	if _, err := store.Save(ctx, cfg); err == nil {
		t.Fatal("expected raw DSN to be rejected")
	}
}
