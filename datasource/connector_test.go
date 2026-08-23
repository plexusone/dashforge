package datasource

import (
	"context"
	"fmt"
	"testing"

	"github.com/plexusone/dashforge/uispec"
)

type mockConnector struct {
	id      string
	results map[string]any
}

func (m *mockConnector) ConnectorID() string { return m.id }

func (m *mockConnector) Execute(_ context.Context, operation string, _ map[string]any) (any, error) {
	result, ok := m.results[operation]
	if !ok {
		return nil, fmt.Errorf("unknown operation: %s", operation)
	}
	return result, nil
}

func TestConnectorRegistryRegisterAndGet(t *testing.T) {
	reg := NewConnectorRegistry()
	c := &mockConnector{id: "test-api", results: map[string]any{}}

	if err := reg.RegisterConnector(c); err != nil {
		t.Fatal(err)
	}

	got, ok := reg.GetConnector("test-api")
	if !ok {
		t.Fatal("expected connector to exist")
	}
	if got.ConnectorID() != "test-api" {
		t.Errorf("got ID %q, want test-api", got.ConnectorID())
	}
}

func TestConnectorRegistryMissing(t *testing.T) {
	reg := NewConnectorRegistry()
	_, ok := reg.GetConnector("nope")
	if ok {
		t.Error("expected connector to be missing")
	}
}

func TestConnectorRegistryEmptyID(t *testing.T) {
	reg := NewConnectorRegistry()
	err := reg.RegisterConnector(&mockConnector{id: ""})
	if err == nil {
		t.Fatal("expected error for empty ID")
	}
}

func TestResolveBindings(t *testing.T) {
	reg := NewConnectorRegistry()
	if err := reg.RegisterConnector(&mockConnector{
		id: "revenue-api",
		results: map[string]any{
			"getTimeSeries": []any{
				map[string]any{"month": "Jan", "value": 100},
				map[string]any{"month": "Feb", "value": 200},
			},
		},
	}); err != nil {
		t.Fatal(err)
	}

	bindings := map[string]uispec.Binding{
		"series": {
			Source:    "revenue-api",
			Operation: "getTimeSeries",
			Parameters: map[string]any{
				"period": "${context.period}",
			},
		},
	}

	exprCtx := map[string]any{
		"context": map[string]any{"period": "monthly"},
	}

	results, err := reg.Resolve(context.Background(), bindings, exprCtx)
	if err != nil {
		t.Fatal(err)
	}

	series, ok := results["series"]
	if !ok {
		t.Fatal("expected 'series' in results")
	}

	arr, ok := series.([]any)
	if !ok {
		t.Fatalf("expected []any, got %T", series)
	}
	if len(arr) != 2 {
		t.Errorf("expected 2 items, got %d", len(arr))
	}
}

func TestResolveBindingsUnknownSource(t *testing.T) {
	reg := NewConnectorRegistry()
	bindings := map[string]uispec.Binding{
		"data": {Source: "unknown-api", Operation: "get"},
	}

	_, err := reg.Resolve(context.Background(), bindings, nil)
	if err == nil {
		t.Fatal("expected error for unknown source")
	}
}

func TestResolveBindingsDefault(t *testing.T) {
	reg := NewConnectorRegistry()
	if err := reg.RegisterConnector(&mockConnector{
		id:      "api",
		results: map[string]any{"get": nil},
	}); err != nil {
		t.Fatal(err)
	}

	bindings := map[string]uispec.Binding{
		"data": {
			Source:    "api",
			Operation: "get",
			Default:   "fallback",
		},
	}

	results, err := reg.Resolve(context.Background(), bindings, nil)
	if err != nil {
		t.Fatal(err)
	}

	if results["data"] != "fallback" {
		t.Errorf("got %v, want fallback", results["data"])
	}
}
