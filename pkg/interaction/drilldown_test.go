package interaction

import (
	"context"
	"encoding/json"
	"os"
	"testing"

	"github.com/plexusone/uiforge/pkg/state"
	"github.com/plexusone/uiforge/uispec"
)

func TestDashboardDrillDown(t *testing.T) {
	data, err := os.ReadFile("../../examples/dashboard-drilldown.page.json")
	if err != nil {
		t.Fatal(err)
	}
	var page uispec.PageSpec
	if err := json.Unmarshal(data, &page); err != nil {
		t.Fatal(err)
	}

	if len(page.Interactions) != 3 {
		t.Fatalf("expected 3 interactions, got %d", len(page.Interactions))
	}

	store := state.New()
	engine := New(store)

	var refreshed []string
	engine.RegisterHandler("component.refresh", func(_ context.Context, action uispec.InteractionAction, _ map[string]any) error {
		refreshed = append(refreshed, action.Target)
		return nil
	})

	// Simulate chart point selection
	evt := Event{
		ComponentID: "revenue-chart",
		EventName:   "pointSelected",
		Data:        map[string]any{"month": "2026-03"},
	}
	if err := engine.Dispatch(context.Background(), page.Interactions, evt); err != nil {
		t.Fatal(err)
	}

	got, ok := store.Get("selectedMonth")
	if !ok {
		t.Fatal("selectedMonth not set in state")
	}
	if got != "2026-03" {
		t.Errorf("selectedMonth = %v, want 2026-03", got)
	}

	if len(refreshed) != 4 {
		t.Errorf("expected 4 component refreshes, got %d: %v", len(refreshed), refreshed)
	}

	// Simulate category bar selection
	refreshed = nil
	evt2 := Event{
		ComponentID: "category-chart",
		EventName:   "barSelected",
		Data:        map[string]any{"category": "SaaS"},
	}
	if err := engine.Dispatch(context.Background(), page.Interactions, evt2); err != nil {
		t.Fatal(err)
	}

	got2, ok := store.Get("selectedCategory")
	if !ok {
		t.Fatal("selectedCategory not set in state")
	}
	if got2 != "SaaS" {
		t.Errorf("selectedCategory = %v, want SaaS", got2)
	}
	if len(refreshed) != 1 || refreshed[0] != "transaction-table" {
		t.Errorf("expected [transaction-table] refresh, got %v", refreshed)
	}
}
