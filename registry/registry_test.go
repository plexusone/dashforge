package registry

import (
	"testing"

	"github.com/plexusone/dashforge/uispec"
)

func newTestRegistry() *Registry {
	r := New()
	_ = r.Register(&ComponentSpec{
		ID:       "analytics.line-chart",
		Version:  "1.0.0",
		Category: "visualization",
		Runtime:  "react",
		DataInputs: map[string]DataInput{
			"series": {Type: "timeseries", Required: true},
		},
		Events: map[string]EventDef{
			"pointSelected": {Description: "fired when a data point is clicked"},
		},
	})
	_ = r.Register(&ComponentSpec{
		ID:       "analytics.metric",
		Version:  "1.0.0",
		Category: "visualization",
		Runtime:  "react",
		DataInputs: map[string]DataInput{
			"value": {Type: "scalar", Required: true},
		},
	})
	_ = r.Register(&ComponentSpec{
		ID:       "core.card",
		Version:  "1.0.0",
		Category: "layout",
		Runtime:  "react",
	})
	_ = r.Register(&ComponentSpec{
		ID:       "assistant.thread",
		Version:  "1.0.0",
		Category: "conversation",
		Runtime:  "react",
		DataInputs: map[string]DataInput{
			"messages": {Type: "message-list", Required: true},
		},
	})
	_ = r.Register(&ComponentSpec{
		ID:       "assistant.thread-list",
		Version:  "1.0.0",
		Category: "conversation",
		Runtime:  "react",
	})
	return r
}

func TestRegistration(t *testing.T) {
	r := New()

	if err := r.Register(&ComponentSpec{ID: "", Version: "1.0.0"}); err == nil {
		t.Error("expected error for empty ID")
	}
	if err := r.Register(&ComponentSpec{ID: "noDot", Version: "1.0.0"}); err == nil {
		t.Error("expected error for ID without namespace")
	}
	if err := r.Register(&ComponentSpec{ID: "core.button", Version: ""}); err == nil {
		t.Error("expected error for empty version")
	}

	err := r.Register(&ComponentSpec{ID: "core.button", Version: "1.0.0", Runtime: "react"})
	if err != nil {
		t.Fatal(err)
	}
	if r.Count() != 1 {
		t.Errorf("Count = %d, want 1", r.Count())
	}
	if !r.Has("core.button") {
		t.Error("Has(core.button) = false")
	}

	spec := r.Get("core.button")
	if spec == nil {
		t.Fatal("Get(core.button) = nil")
	}
	if spec.Namespace != "core" {
		t.Errorf("Namespace = %q, want %q", spec.Namespace, "core")
	}
}

func TestListByNamespace(t *testing.T) {
	r := newTestRegistry()

	analytics := r.ListByNamespace("analytics")
	if len(analytics) != 2 {
		t.Errorf("analytics namespace count = %d, want 2", len(analytics))
	}

	assistant := r.ListByNamespace("assistant")
	if len(assistant) != 2 {
		t.Errorf("assistant namespace count = %d, want 2", len(assistant))
	}

	core := r.ListByNamespace("core")
	if len(core) != 1 {
		t.Errorf("core namespace count = %d, want 1", len(core))
	}
}

func TestValidatePage_Valid(t *testing.T) {
	r := newTestRegistry()

	page := &uispec.PageSpec{
		APIVersion: uispec.APIVersion,
		Kind:       uispec.KindPage,
		Metadata:   uispec.PageMetadata{ID: "test", Name: "test"},
		Layout: uispec.LayoutSpec{
			Type:   uispec.LayoutResponsiveGrid,
			Config: &uispec.LayoutConfig{Columns: 12},
		},
		Components: []uispec.ComponentInstance{
			{
				ID:   "chart",
				Type: "analytics.line-chart",
				Data: map[string]uispec.Binding{
					"series": {Source: "api", Operation: "getData"},
				},
			},
			{
				ID:   "metric",
				Type: "analytics.metric",
				Data: map[string]uispec.Binding{
					"value": {Source: "api", Operation: "getTotal"},
				},
			},
		},
		Interactions: []uispec.Interaction{
			{
				When: uispec.InteractionTrigger{Component: "chart", Event: "pointSelected"},
				Then: []uispec.InteractionAction{
					{Target: "metric", Action: "refresh"},
				},
			},
		},
	}

	if err := r.ValidatePage(page); err != nil {
		t.Errorf("unexpected validation error: %v", err)
	}
}

func TestValidatePage_UnregisteredComponent(t *testing.T) {
	r := newTestRegistry()

	page := &uispec.PageSpec{
		APIVersion: uispec.APIVersion,
		Kind:       uispec.KindPage,
		Metadata:   uispec.PageMetadata{ID: "test", Name: "test"},
		Layout:     uispec.LayoutSpec{Type: uispec.LayoutStack},
		Components: []uispec.ComponentInstance{
			{ID: "widget", Type: "custom.nonexistent"},
		},
	}

	err := r.ValidatePage(page)
	if err == nil {
		t.Fatal("expected validation error for unregistered component")
	}
	ve, ok := err.(*ValidationError)
	if !ok {
		t.Fatalf("expected *ValidationError, got %T", err)
	}
	if len(ve.Errors) != 1 {
		t.Errorf("expected 1 error, got %d: %v", len(ve.Errors), ve.Errors)
	}
}

func TestValidatePage_MissingRequiredBinding(t *testing.T) {
	r := newTestRegistry()

	page := &uispec.PageSpec{
		APIVersion: uispec.APIVersion,
		Kind:       uispec.KindPage,
		Metadata:   uispec.PageMetadata{ID: "test", Name: "test"},
		Layout:     uispec.LayoutSpec{Type: uispec.LayoutStack},
		Components: []uispec.ComponentInstance{
			{ID: "chart", Type: "analytics.line-chart"},
		},
	}

	err := r.ValidatePage(page)
	if err == nil {
		t.Fatal("expected validation error for missing required data binding")
	}
}

func TestValidatePage_UndeclaredBinding(t *testing.T) {
	r := newTestRegistry()

	page := &uispec.PageSpec{
		APIVersion: uispec.APIVersion,
		Kind:       uispec.KindPage,
		Metadata:   uispec.PageMetadata{ID: "test", Name: "test"},
		Layout:     uispec.LayoutSpec{Type: uispec.LayoutStack},
		Components: []uispec.ComponentInstance{
			{
				ID:   "card",
				Type: "core.card",
				Data: map[string]uispec.Binding{
					"unknown": {Source: "api"},
				},
			},
		},
	}

	err := r.ValidatePage(page)
	if err == nil {
		t.Fatal("expected validation error for undeclared data binding")
	}
}

func TestValidatePage_DuplicateComponentID(t *testing.T) {
	r := newTestRegistry()

	page := &uispec.PageSpec{
		APIVersion: uispec.APIVersion,
		Kind:       uispec.KindPage,
		Metadata:   uispec.PageMetadata{ID: "test", Name: "test"},
		Layout:     uispec.LayoutSpec{Type: uispec.LayoutStack},
		Components: []uispec.ComponentInstance{
			{ID: "card", Type: "core.card"},
			{ID: "card", Type: "core.card"},
		},
	}

	err := r.ValidatePage(page)
	if err == nil {
		t.Fatal("expected validation error for duplicate component ID")
	}
}

func TestValidatePage_InteractionRefsMissingComponent(t *testing.T) {
	r := newTestRegistry()

	page := &uispec.PageSpec{
		APIVersion: uispec.APIVersion,
		Kind:       uispec.KindPage,
		Metadata:   uispec.PageMetadata{ID: "test", Name: "test"},
		Layout:     uispec.LayoutSpec{Type: uispec.LayoutStack},
		Components: []uispec.ComponentInstance{
			{ID: "card", Type: "core.card"},
		},
		Interactions: []uispec.Interaction{
			{
				When: uispec.InteractionTrigger{Component: "ghost", Event: "click"},
				Then: []uispec.InteractionAction{{Target: "card", Action: "show"}},
			},
		},
	}

	err := r.ValidatePage(page)
	if err == nil {
		t.Fatal("expected validation error for interaction referencing missing component")
	}
}

func TestValidateLayout_SplitPane(t *testing.T) {
	r := newTestRegistry()

	page := &uispec.PageSpec{
		APIVersion: uispec.APIVersion,
		Kind:       uispec.KindPage,
		Metadata:   uispec.PageMetadata{ID: "test", Name: "test"},
		Layout: uispec.LayoutSpec{
			Type: uispec.LayoutSplitPane,
			Config: &uispec.LayoutConfig{
				Sizes:     []string{"25%", "50%", "25%"},
				Resizable: true,
			},
			Regions: []uispec.LayoutRegion{
				{Name: "sidebar"},
				{Name: "main"},
				{Name: "inspector"},
			},
		},
		Components: []uispec.ComponentInstance{
			{ID: "tl", Type: "assistant.thread-list", Slot: "sidebar"},
			{
				ID:   "thread",
				Type: "assistant.thread",
				Slot: "main",
				Data: map[string]uispec.Binding{
					"messages": {Source: "api", Operation: "getMessages"},
				},
			},
		},
	}

	if err := r.ValidatePage(page); err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	page.Layout.Regions = page.Layout.Regions[:1]
	err := r.ValidatePage(page)
	if err == nil {
		t.Error("expected error for split-pane with < 2 regions")
	}
}

func TestValidateLayout_Tabs(t *testing.T) {
	r := newTestRegistry()

	page := &uispec.PageSpec{
		APIVersion: uispec.APIVersion,
		Kind:       uispec.KindPage,
		Metadata:   uispec.PageMetadata{ID: "test", Name: "test"},
		Layout: uispec.LayoutSpec{
			Type:    uispec.LayoutTabs,
			Regions: []uispec.LayoutRegion{{Name: "general"}, {Name: "advanced"}},
		},
		Components: []uispec.ComponentInstance{
			{ID: "card", Type: "core.card", Slot: "general"},
		},
	}

	if err := r.ValidatePage(page); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestValidateLayout_ApplicationShell(t *testing.T) {
	r := newTestRegistry()

	page := &uispec.PageSpec{
		APIVersion: uispec.APIVersion,
		Kind:       uispec.KindPage,
		Metadata:   uispec.PageMetadata{ID: "test", Name: "test"},
		Layout: uispec.LayoutSpec{
			Type: uispec.LayoutApplicationShell,
			Regions: []uispec.LayoutRegion{
				{Name: "nav"},
				{Name: "main"},
				{Name: "footer"},
			},
		},
		Components: []uispec.ComponentInstance{
			{ID: "card", Type: "core.card", Slot: "main"},
		},
	}

	if err := r.ValidatePage(page); err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	page.Layout.Regions = []uispec.LayoutRegion{{Name: "nav"}, {Name: "footer"}}
	err := r.ValidatePage(page)
	if err == nil {
		t.Error("expected error for application-shell missing 'main' region")
	}
}

func TestValidateLayout_InvalidType(t *testing.T) {
	r := newTestRegistry()

	page := &uispec.PageSpec{
		APIVersion: uispec.APIVersion,
		Kind:       uispec.KindPage,
		Metadata:   uispec.PageMetadata{ID: "test", Name: "test"},
		Layout:     uispec.LayoutSpec{Type: "carousel"},
		Components: []uispec.ComponentInstance{},
	}

	err := r.ValidatePage(page)
	if err == nil {
		t.Error("expected error for invalid layout type")
	}
}

func TestValidateLayout_NestedRegions(t *testing.T) {
	r := newTestRegistry()

	page := &uispec.PageSpec{
		APIVersion: uispec.APIVersion,
		Kind:       uispec.KindPage,
		Metadata:   uispec.PageMetadata{ID: "test", Name: "test"},
		Layout: uispec.LayoutSpec{
			Type: uispec.LayoutApplicationShell,
			Regions: []uispec.LayoutRegion{
				{Name: "nav"},
				{
					Name: "main",
					Layout: &uispec.LayoutSpec{
						Type:   uispec.LayoutResponsiveGrid,
						Config: &uispec.LayoutConfig{Columns: 12},
					},
				},
			},
		},
		Components: []uispec.ComponentInstance{
			{ID: "card", Type: "core.card", Slot: "main"},
		},
	}

	if err := r.ValidatePage(page); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}
