package analytics

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/plexusone/dashforge/dashboardir"
)

// fakeProvider is a QueryProvider whose catalog reports one source with an
// internal ID, to exercise re-tagging.
type fakeProvider struct {
	dsn    string
	closed bool
}

func (p *fakeProvider) Catalog(context.Context) (dashboardir.AnalyticsCatalog, error) {
	return dashboardir.AnalyticsCatalog{
		ID:   "fake",
		Name: "Fake",
		Sources: []dashboardir.AnalyticsSource{
			{ID: "fake-internal", Name: "Fake Internal"},
		},
	}, nil
}

func (p *fakeProvider) Query(_ context.Context, req dashboardir.AnalyticsQueryRequest) (dashboardir.AnalyticsQueryResult, error) {
	return dashboardir.AnalyticsQueryResult{Columns: []dashboardir.AnalyticsQueryColumn{{Name: p.dsn}}}, nil
}

func (p *fakeProvider) Close() error {
	p.closed = true
	return nil
}

var registerFakeConnector sync.Once

func setupFakeConnector(t *testing.T) {
	t.Helper()
	registerFakeConnector.Do(func() {
		RegisterConnector("fake", func(dsn string) (QueryProvider, error) {
			if dsn == "fail" {
				return nil, fmt.Errorf("dial failed")
			}
			return &fakeProvider{dsn: dsn}, nil
		})
	})
}

func newManagedServiceForTest(t *testing.T, storePath string) *Service {
	t.Helper()
	setupFakeConnector(t)
	store, err := NewSourceFileStore(storePath)
	if err != nil {
		t.Fatal(err)
	}
	resolver, err := NewDefaultResolver()
	if err != nil {
		t.Fatal(err)
	}
	return NewManagedService(store, resolver, slog.New(slog.NewTextHandler(os.Stderr, nil)))
}

func fakeSourceConfig(id, envVar string) SourceConfig {
	return SourceConfig{
		ID:        id,
		Name:      "Source " + id,
		Connector: "fake",
		DSNRef:    "env://" + envVar,
		Enabled:   true,
	}
}

func TestManagedServiceAddCatalogQueryRemove(t *testing.T) {
	ctx := context.Background()
	svc := newManagedServiceForTest(t, filepath.Join(t.TempDir(), "sources.json"))
	t.Setenv("DASHFORGE_TEST_FAKE_A", "dsn-a")
	t.Setenv("DASHFORGE_TEST_FAKE_B", "dsn-b")

	if svc.HasProviders() {
		t.Fatal("expected no providers before AddSource")
	}

	if _, err := svc.AddSource(ctx, fakeSourceConfig("alpha", "DASHFORGE_TEST_FAKE_A")); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.AddSource(ctx, fakeSourceConfig("beta", "DASHFORGE_TEST_FAKE_B")); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.AddSource(ctx, fakeSourceConfig("alpha", "DASHFORGE_TEST_FAKE_A")); err == nil {
		t.Fatal("expected duplicate ID to be rejected")
	}

	// Two sources on the same connector re-tag to unique configured IDs.
	catalog, err := svc.Catalog(ctx)
	if err != nil {
		t.Fatal(err)
	}
	ids := map[string]bool{}
	for _, src := range catalog.Sources {
		ids[src.ID] = true
	}
	if !ids["alpha"] || !ids["beta"] || len(ids) != 2 {
		t.Fatalf("expected re-tagged sources alpha and beta, got %v", ids)
	}

	// Query routes by configured source ID to the right provider (the fake
	// echoes its DSN as a column ID).
	result, err := svc.Query(ctx, dashboardir.AnalyticsQueryRequest{SourceID: "beta", Query: "q"})
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Columns) != 1 || result.Columns[0].Name != "dsn-b" {
		t.Fatalf("query routed to wrong provider: %+v", result)
	}

	if _, err := svc.Query(ctx, dashboardir.AnalyticsQueryRequest{SourceID: "gamma", Query: "q"}); !errors.Is(err, ErrQueryUnsupported) {
		t.Fatalf("expected ErrQueryUnsupported for unknown source, got %v", err)
	}

	statuses, err := svc.Sources(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(statuses) != 2 {
		t.Fatalf("expected 2 statuses, got %d", len(statuses))
	}
	for _, st := range statuses {
		if st.Status != "connected" {
			t.Fatalf("expected connected, got %+v", st)
		}
	}

	if err := svc.RemoveSource(ctx, "alpha"); err != nil {
		t.Fatal(err)
	}
	catalog, err = svc.Catalog(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(catalog.Sources) != 1 || catalog.Sources[0].ID != "beta" {
		t.Fatalf("expected only beta after removal, got %+v", catalog.Sources)
	}
	if err := svc.RemoveSource(ctx, "alpha"); !errors.Is(err, ErrSourceNotFound) {
		t.Fatalf("expected ErrSourceNotFound, got %v", err)
	}
}

func TestManagedServiceLoadAllReconnects(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "sources.json")
	t.Setenv("DASHFORGE_TEST_FAKE_A", "dsn-a")

	svc1 := newManagedServiceForTest(t, path)
	if _, err := svc1.AddSource(ctx, fakeSourceConfig("alpha", "DASHFORGE_TEST_FAKE_A")); err != nil {
		t.Fatal(err)
	}
	if err := svc1.Close(); err != nil {
		t.Fatal(err)
	}

	// A fresh service over the same store reconnects on LoadAll.
	svc2 := newManagedServiceForTest(t, path)
	if err := svc2.LoadAll(ctx); err != nil {
		t.Fatal(err)
	}
	if !svc2.HasProviders() {
		t.Fatal("expected provider after LoadAll")
	}
	if _, err := svc2.Query(ctx, dashboardir.AnalyticsQueryRequest{SourceID: "alpha", Query: "q"}); err != nil {
		t.Fatal(err)
	}
}

func TestManagedServiceConnectFailure(t *testing.T) {
	ctx := context.Background()
	svc := newManagedServiceForTest(t, filepath.Join(t.TempDir(), "sources.json"))
	t.Setenv("DASHFORGE_TEST_FAKE_FAIL", "fail")

	// AddSource refuses a config that cannot connect.
	if _, err := svc.AddSource(ctx, fakeSourceConfig("broken", "DASHFORGE_TEST_FAKE_FAIL")); err == nil {
		t.Fatal("expected AddSource to fail on dial error")
	}

	// A stored source that fails at LoadAll surfaces as status "error", and
	// the error never contains the resolved DSN.
	store, err := NewSourceFileStore(filepath.Join(t.TempDir(), "sources2.json"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := store.Save(ctx, fakeSourceConfig("broken", "DASHFORGE_TEST_FAKE_FAIL")); err != nil {
		t.Fatal(err)
	}
	resolver, err := NewDefaultResolver()
	if err != nil {
		t.Fatal(err)
	}
	svc2 := NewManagedService(store, resolver, slog.New(slog.NewTextHandler(os.Stderr, nil)))
	if err := svc2.LoadAll(ctx); err != nil {
		t.Fatal(err)
	}
	statuses, err := svc2.Sources(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(statuses) != 1 || statuses[0].Status != "error" {
		t.Fatalf("expected error status, got %+v", statuses)
	}
	if statuses[0].Error == "" {
		t.Fatal("expected sanitized error message")
	}
}

func TestManagedServiceTestSource(t *testing.T) {
	ctx := context.Background()
	svc := newManagedServiceForTest(t, filepath.Join(t.TempDir(), "sources.json"))
	t.Setenv("DASHFORGE_TEST_FAKE_A", "dsn-a")

	if err := svc.TestSource(ctx, fakeSourceConfig("probe", "DASHFORGE_TEST_FAKE_A")); err != nil {
		t.Fatal(err)
	}
	// Nothing persisted by a test.
	statuses, err := svc.Sources(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(statuses) != 0 {
		t.Fatalf("TestSource must not persist, got %+v", statuses)
	}
}

func TestManagedServiceUpdateSource(t *testing.T) {
	ctx := context.Background()
	svc := newManagedServiceForTest(t, filepath.Join(t.TempDir(), "sources.json"))
	t.Setenv("DASHFORGE_TEST_FAKE_A", "dsn-a")

	if _, err := svc.UpdateSource(ctx, fakeSourceConfig("alpha", "DASHFORGE_TEST_FAKE_A")); !errors.Is(err, ErrSourceNotFound) {
		t.Fatalf("expected ErrSourceNotFound for update of unknown source, got %v", err)
	}

	if _, err := svc.AddSource(ctx, fakeSourceConfig("alpha", "DASHFORGE_TEST_FAKE_A")); err != nil {
		t.Fatal(err)
	}
	cfg := fakeSourceConfig("alpha", "DASHFORGE_TEST_FAKE_A")
	cfg.Enabled = false
	updated, err := svc.UpdateSource(ctx, cfg)
	if err != nil {
		t.Fatal(err)
	}
	if updated.Enabled {
		t.Fatal("expected disabled source")
	}
	statuses, err := svc.Sources(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(statuses) != 1 || statuses[0].Status != "disabled" {
		t.Fatalf("expected disabled status, got %+v", statuses)
	}
	if svc.HasProviders() {
		t.Fatal("expected no live providers after disabling the only source")
	}
}
