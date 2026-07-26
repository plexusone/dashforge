package uispec

import (
	"encoding/json"
	"testing"
)

func TestPageSpecRoundTrip(t *testing.T) {
	page := PageSpec{
		APIVersion: APIVersion,
		Kind:       KindPage,
		Metadata: PageMetadata{
			ID:     "test-dashboard",
			Name:   "test-dashboard",
			Title:  "Test Dashboard",
			Labels: map[string]string{"env": "dev"},
		},
		Profile: ProfileDashboard,
		Context: map[string]string{"entityType": "customer"},
		Layout: LayoutSpec{
			Type: LayoutResponsiveGrid,
			Config: &LayoutConfig{
				Columns: 12,
				Gap:     "16px",
			},
			Regions: []LayoutRegion{
				{Name: "header"},
				{Name: "main"},
			},
		},
		Components: []ComponentInstance{
			{
				ID:   "revenue-chart",
				Type: "analytics.line-chart",
				Position: &Position{
					Row: 0, Col: 0, ColSpan: 6, RowSpan: 4,
				},
				Properties: map[string]any{
					"title": "Revenue Over Time",
				},
				Data: map[string]Binding{
					"series": {
						Source:    "revenue-api",
						Operation: "getTimeSeries",
						Parameters: map[string]any{
							"period": "${context.period}",
						},
					},
				},
			},
			{
				ID:   "metric-card",
				Type: "analytics.metric",
				Position: &Position{
					Row: 0, Col: 6, ColSpan: 3, RowSpan: 2,
				},
				Visibility: &VisibilityRule{
					Condition: "${user.role == 'admin'}",
				},
			},
		},
		Interactions: []Interaction{
			{
				When: InteractionTrigger{
					Component: "revenue-chart",
					Event:     "pointSelected",
				},
				Then: []InteractionAction{
					{
						Target: "detail-table",
						Action: "filter",
						Params: map[string]any{
							"month": "${event.month}",
						},
					},
				},
			},
		},
		Theme: &ThemeRef{
			ID:      "default",
			Variant: "dark",
		},
	}

	data, err := json.Marshal(page)
	if err != nil {
		t.Fatal(err)
	}

	var got PageSpec
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatal(err)
	}

	if got.APIVersion != APIVersion {
		t.Errorf("APIVersion = %q, want %q", got.APIVersion, APIVersion)
	}
	if got.Kind != KindPage {
		t.Errorf("Kind = %q, want %q", got.Kind, KindPage)
	}
	if got.Metadata.ID != "test-dashboard" {
		t.Errorf("Metadata.ID = %q, want %q", got.Metadata.ID, "test-dashboard")
	}
	if got.Profile != ProfileDashboard {
		t.Errorf("Profile = %q, want %q", got.Profile, ProfileDashboard)
	}
	if got.Layout.Type != LayoutResponsiveGrid {
		t.Errorf("Layout.Type = %q, want %q", got.Layout.Type, LayoutResponsiveGrid)
	}
	if got.Layout.Config.Columns != 12 {
		t.Errorf("Layout.Config.Columns = %d, want 12", got.Layout.Config.Columns)
	}
	if len(got.Layout.Regions) != 2 {
		t.Errorf("Layout.Regions len = %d, want 2", len(got.Layout.Regions))
	}
	if len(got.Components) != 2 {
		t.Errorf("Components len = %d, want 2", len(got.Components))
	}
	if got.Components[0].Type != "analytics.line-chart" {
		t.Errorf("Components[0].Type = %q, want %q", got.Components[0].Type, "analytics.line-chart")
	}
	if got.Components[0].Position.ColSpan != 6 {
		t.Errorf("Components[0].Position.ColSpan = %d, want 6", got.Components[0].Position.ColSpan)
	}
	if _, ok := got.Components[0].Data["series"]; !ok {
		t.Error("Components[0].Data missing 'series' binding")
	}
	if got.Components[1].Visibility == nil {
		t.Error("Components[1].Visibility is nil")
	}
	if len(got.Interactions) != 1 {
		t.Errorf("Interactions len = %d, want 1", len(got.Interactions))
	}
	if got.Theme.Variant != "dark" {
		t.Errorf("Theme.Variant = %q, want %q", got.Theme.Variant, "dark")
	}
}

func TestPageSpecFromJSON(t *testing.T) {
	raw := `{
		"apiVersion": "ui.plexusone.dev/v1",
		"kind": "Page",
		"metadata": {"id": "agent-workspace", "name": "agent-workspace"},
		"profile": "agent",
		"layout": {
			"type": "split-pane",
			"config": {"sizes": ["25%", "50%", "25%"], "resizable": true},
			"regions": [
				{"name": "sidebar"},
				{"name": "main"},
				{"name": "inspector"}
			]
		},
		"components": [
			{
				"id": "thread-list",
				"type": "assistant.thread-list",
				"slot": "sidebar"
			},
			{
				"id": "thread",
				"type": "assistant.thread",
				"slot": "main",
				"data": {
					"messages": {
						"source": "agentos-api",
						"operation": "getMessages"
					}
				}
			}
		],
		"navigation": {
			"items": [{"id": "chat", "label": "Chat", "icon": "message-circle"}],
			"position": "left"
		}
	}`

	var page PageSpec
	if err := json.Unmarshal([]byte(raw), &page); err != nil {
		t.Fatal(err)
	}

	if page.Profile != ProfileAgent {
		t.Errorf("Profile = %q, want %q", page.Profile, ProfileAgent)
	}
	if page.Layout.Type != LayoutSplitPane {
		t.Errorf("Layout.Type = %q, want %q", page.Layout.Type, LayoutSplitPane)
	}
	if len(page.Layout.Config.Sizes) != 3 {
		t.Errorf("Layout.Config.Sizes len = %d, want 3", len(page.Layout.Config.Sizes))
	}
	if !page.Layout.Config.Resizable {
		t.Error("Layout.Config.Resizable = false, want true")
	}
	if len(page.Layout.Regions) != 3 {
		t.Errorf("Layout.Regions len = %d, want 3", len(page.Layout.Regions))
	}
	if len(page.Components) != 2 {
		t.Errorf("Components len = %d, want 2", len(page.Components))
	}
	if page.Components[0].Slot != "sidebar" {
		t.Errorf("Components[0].Slot = %q, want %q", page.Components[0].Slot, "sidebar")
	}
	if page.Navigation == nil {
		t.Fatal("Navigation is nil")
	}
	if page.Navigation.Position != "left" {
		t.Errorf("Navigation.Position = %q, want %q", page.Navigation.Position, "left")
	}
}
