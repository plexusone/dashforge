package diff

import (
	"testing"

	"github.com/plexusone/uiforge/uispec"
)

func basePage() *uispec.PageSpec {
	return &uispec.PageSpec{
		APIVersion: uispec.APIVersion,
		Kind:       uispec.KindPage,
		Metadata: uispec.PageMetadata{
			ID:   "test",
			Name: "test",
		},
		Profile: uispec.ProfileDashboard,
		Layout: uispec.LayoutSpec{
			Type:   uispec.LayoutResponsiveGrid,
			Config: &uispec.LayoutConfig{Columns: 12},
		},
		Components: []uispec.ComponentInstance{
			{ID: "chart-1", Type: "analytics.line-chart"},
			{ID: "table-1", Type: "analytics.table"},
		},
		Interactions: []uispec.Interaction{
			{
				When: uispec.InteractionTrigger{Component: "chart-1", Event: "click"},
				Then: []uispec.InteractionAction{
					{Action: "state.set", Params: map[string]any{"path": "x", "value": 1}},
				},
			},
		},
	}
}

func TestNoChanges(t *testing.T) {
	a := basePage()
	b := basePage()
	result := PageSpecs(a, b)
	if result.HasChanges() {
		t.Errorf("expected no changes, got %d: %s", len(result.Changes), result.Summary())
	}
}

func TestBothNil(t *testing.T) {
	result := PageSpecs(nil, nil)
	if result.HasChanges() {
		t.Error("expected no changes for nil,nil")
	}
}

func TestAddedPage(t *testing.T) {
	result := PageSpecs(nil, basePage())
	if !result.HasChanges() {
		t.Fatal("expected changes")
	}
	if result.Changes[0].Type != Added {
		t.Errorf("type = %s, want added", result.Changes[0].Type)
	}
}

func TestRemovedPage(t *testing.T) {
	result := PageSpecs(basePage(), nil)
	if !result.HasChanges() {
		t.Fatal("expected changes")
	}
	if result.Changes[0].Type != Removed {
		t.Errorf("type = %s, want removed", result.Changes[0].Type)
	}
}

func TestComponentAdded(t *testing.T) {
	a := basePage()
	b := basePage()
	b.Components = append(b.Components, uispec.ComponentInstance{ID: "new-widget", Type: "core.card"})

	result := PageSpecs(a, b)
	if !result.HasChanges() {
		t.Fatal("expected changes")
	}

	var found bool
	for _, c := range result.Changes {
		if c.Path == "components[new-widget]" && c.Type == Added {
			found = true
		}
	}
	if !found {
		t.Error("expected 'added' change for new-widget")
	}
}

func TestComponentRemoved(t *testing.T) {
	a := basePage()
	b := basePage()
	b.Components = b.Components[:1] // remove table-1

	result := PageSpecs(a, b)
	if !result.HasChanges() {
		t.Fatal("expected changes")
	}

	var found bool
	for _, c := range result.Changes {
		if c.Path == "components[table-1]" && c.Type == Removed {
			found = true
		}
	}
	if !found {
		t.Error("expected 'removed' change for table-1")
	}
}

func TestComponentModified(t *testing.T) {
	a := basePage()
	b := basePage()
	b.Components[0].Type = "analytics.bar-chart"

	result := PageSpecs(a, b)
	if !result.HasChanges() {
		t.Fatal("expected changes")
	}

	var found bool
	for _, c := range result.Changes {
		if c.Path == "components[chart-1].type" && c.Type == Modified {
			found = true
		}
	}
	if !found {
		t.Errorf("expected 'modified' change for chart-1.type, got: %v", result.Changes)
	}
}

func TestLayoutChange(t *testing.T) {
	a := basePage()
	b := basePage()
	b.Layout.Type = uispec.LayoutStack

	result := PageSpecs(a, b)
	if !result.HasChanges() {
		t.Fatal("expected changes")
	}

	var found bool
	for _, c := range result.Changes {
		if c.Path == "layout.type" && c.Type == Modified {
			found = true
		}
	}
	if !found {
		t.Error("expected layout.type change")
	}
}

func TestInteractionChange(t *testing.T) {
	a := basePage()
	b := basePage()
	b.Interactions = nil

	result := PageSpecs(a, b)
	if !result.HasChanges() {
		t.Fatal("expected changes")
	}

	var found bool
	for _, c := range result.Changes {
		if c.Path == "interactions" && c.Type == Modified {
			found = true
		}
	}
	if !found {
		t.Error("expected interactions change")
	}
}

func TestProfileChange(t *testing.T) {
	a := basePage()
	b := basePage()
	b.Profile = uispec.ProfileAgent

	result := PageSpecs(a, b)
	var found bool
	for _, c := range result.Changes {
		if c.Path == "profile" && c.Type == Modified {
			found = true
		}
	}
	if !found {
		t.Error("expected profile change")
	}
}

func TestThemeAddedRemoved(t *testing.T) {
	a := basePage()
	b := basePage()
	b.Theme = &uispec.ThemeRef{ID: "dark", Variant: "high-contrast"}

	result := PageSpecs(a, b)
	var found bool
	for _, c := range result.Changes {
		if c.Path == "theme" && c.Type == Added {
			found = true
		}
	}
	if !found {
		t.Error("expected theme added")
	}

	// Reverse: theme removed
	result2 := PageSpecs(b, a)
	found = false
	for _, c := range result2.Changes {
		if c.Path == "theme" && c.Type == Removed {
			found = true
		}
	}
	if !found {
		t.Error("expected theme removed")
	}
}

func TestSummary(t *testing.T) {
	r := &DiffResult{}
	if r.Summary() != "No changes" {
		t.Errorf("got %q, want 'No changes'", r.Summary())
	}

	r.Changes = []Change{
		{Type: Added},
		{Type: Removed},
		{Type: Modified},
		{Type: Modified},
	}
	s := r.Summary()
	if s != "4 changes: 1 added, 1 removed, 2 modified" {
		t.Errorf("got %q", s)
	}
}

func TestMetadataChange(t *testing.T) {
	a := basePage()
	b := basePage()
	b.Metadata.Title = "Updated Title"

	result := PageSpecs(a, b)
	var found bool
	for _, c := range result.Changes {
		if c.Path == "metadata.title" && c.Type == Modified {
			found = true
		}
	}
	if !found {
		t.Error("expected metadata.title change")
	}
}
