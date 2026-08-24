package sqlsource

import (
	"context"
	"database/sql"
	"testing"

	"github.com/plexusone/dashforge/analytics"
	"github.com/plexusone/dashforge/dashboardir"
)

func TestConnectorRegistered(t *testing.T) {
	if _, ok := analytics.LookupConnector(ConnectorName); !ok {
		t.Fatalf("connector %q not registered", ConnectorName)
	}
}

func TestDialectFor(t *testing.T) {
	cases := map[string]string{
		"postgres://u@h/db":              "postgres",
		"postgresql://u@h/db":            "postgres",
		"mysql://root@127.0.0.1:3306/db": "mysql",
		"root:@tcp(127.0.0.1:13307)/db":  "mysql",
		"file:test.db?mode=memory":       "sqlite3",
		"sqlite:///tmp/x.db":             "sqlite3",
		"/var/data/app.sqlite":           "sqlite3",
	}
	for dsn, want := range cases {
		d, err := dialectFor(dsn)
		if err != nil {
			t.Errorf("dialectFor(%q): %v", dsn, err)
			continue
		}
		if d.driver() != want {
			t.Errorf("dialectFor(%q).driver() = %q, want %q", dsn, d.driver(), want)
		}
	}
	if _, err := dialectFor("bogus-dsn"); err == nil {
		t.Error("expected error for unrecognizable DSN")
	}
}

func TestCatalogIntrospectionSQLite(t *testing.T) {
	// In-memory SQLite: no external server needed, runs everywhere.
	db, err := sql.Open("sqlite3", "file:catalog_test?mode=memory&cache=shared")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	ctx := context.Background()
	mustExec(t, db, `CREATE TABLE initiatives (
		id INTEGER PRIMARY KEY,
		name TEXT NOT NULL,
		score REAL,
		created_at DATETIME,
		active BOOLEAN
	)`)
	mustExec(t, db, `CREATE TABLE items (id INTEGER PRIMARY KEY, title TEXT)`)

	p := &Provider{db: db, dialect: sqliteDialect{}}
	catalog, err := p.Catalog(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(catalog.Sources) != 1 {
		t.Fatalf("expected 1 source, got %d", len(catalog.Sources))
	}
	src := catalog.Sources[0]
	if src.Type != dashboardir.AnalyticsSourceTypeSQL {
		t.Fatalf("source type = %q, want sql", src.Type)
	}
	ds := datasetByName(src.Datasets, "initiatives")
	if ds == nil {
		t.Fatalf("initiatives dataset missing; got %v", datasetNames(src.Datasets))
	}
	if len(ds.Fields) != 5 {
		t.Fatalf("initiatives fields = %d, want 5", len(ds.Fields))
	}
	// Type + role mapping spot-checks.
	want := map[string]struct{ typ, role string }{
		"name":       {dashboardir.AnalyticsFieldTypeString, dashboardir.AnalyticsFieldRoleDimension},
		"score":      {dashboardir.AnalyticsFieldTypeNumber, dashboardir.AnalyticsFieldRoleMeasure},
		"created_at": {dashboardir.AnalyticsFieldTypeDate, dashboardir.AnalyticsFieldRoleTime},
		"active":     {dashboardir.AnalyticsFieldTypeBool, dashboardir.AnalyticsFieldRoleDimension},
	}
	for _, f := range ds.Fields {
		if w, ok := want[f.QueryName]; ok {
			if f.Type != w.typ || f.Role != w.role {
				t.Errorf("field %s: type=%s role=%s, want type=%s role=%s", f.ID, f.Type, f.Role, w.typ, w.role)
			}
		}
		if !f.Selectable || !f.Filterable || !f.Sortable {
			t.Errorf("field %s should be selectable/filterable/sortable", f.ID)
		}
	}
	// NOT NULL column reflected.
	for _, f := range ds.Fields {
		if f.ID == "name" && f.Nullable {
			t.Error("name is NOT NULL but reported nullable")
		}
	}
}

func TestQueryReadOnlyGuard(t *testing.T) {
	db, err := sql.Open("sqlite3", "file:q_test?mode=memory&cache=shared")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	mustExec(t, db, `CREATE TABLE t (id INTEGER, name TEXT)`)
	mustExec(t, db, `INSERT INTO t VALUES (1,'a'),(2,'b')`)
	p := &Provider{db: db, dialect: sqliteDialect{}}
	ctx := context.Background()

	res, err := p.Query(ctx, dashboardir.AnalyticsQueryRequest{Query: "SELECT name FROM t ORDER BY id"})
	if err != nil {
		t.Fatal(err)
	}
	if res.RowCount != 2 || len(res.Columns) != 1 || res.Rows[0]["name"] != "a" {
		t.Fatalf("unexpected result: %+v", res)
	}
	if _, err := p.Query(ctx, dashboardir.AnalyticsQueryRequest{Query: "DELETE FROM t"}); err == nil {
		t.Error("expected non-SELECT to be rejected")
	}
}

// helpers

func mustExec(t *testing.T, db *sql.DB, q string) {
	t.Helper()
	if _, err := db.Exec(q); err != nil {
		t.Fatalf("exec %q: %v", q, err)
	}
}

func datasetByName(datasets []dashboardir.AnalyticsDataset, name string) *dashboardir.AnalyticsDataset {
	for i := range datasets {
		if datasets[i].Name == name {
			return &datasets[i]
		}
	}
	return nil
}

func datasetNames(datasets []dashboardir.AnalyticsDataset) []string {
	out := make([]string, len(datasets))
	for i, d := range datasets {
		out[i] = d.Name
	}
	return out
}
