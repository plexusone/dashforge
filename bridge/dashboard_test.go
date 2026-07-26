package bridge

import (
	"encoding/json"
	"testing"

	"github.com/plexusone/uiforge/dashboardir"
	"github.com/plexusone/uiforge/uispec"
)

func TestDashboardToPageSpec(t *testing.T) {
	d := &dashboardir.Dashboard{
		ID:          "sales-overview",
		Title:       "Sales Overview",
		Description: "Q3 sales metrics",
		Version:     "1.0.0",
		Layout: dashboardir.Layout{
			Type:    dashboardir.LayoutTypeGrid,
			Columns: 12,
			Gap:     16,
		},
		Theme: &dashboardir.Theme{
			Mode:         dashboardir.ThemeModeDark,
			PrimaryColor: "#3b82f6",
		},
		Widgets: []dashboardir.Widget{
			{
				ID:           "revenue",
				Title:        "Revenue",
				Type:         dashboardir.WidgetTypeChart,
				Position:     dashboardir.Position{X: 0, Y: 0, W: 6, H: 4},
				DataSourceID: "sales-api",
				Config:       json.RawMessage(`{"chartType":"line"}`),
				DrillDown: &dashboardir.DrillDown{
					Type:   dashboardir.DrillDownTypeFilter,
					Target: "detail-table",
					Params: map[string]string{"month": "${event.month}"},
				},
			},
			{
				ID:           "total",
				Title:        "Total Sales",
				Type:         dashboardir.WidgetTypeMetric,
				Position:     dashboardir.Position{X: 6, Y: 0, W: 3, H: 2},
				DataSourceID: "sales-api",
				Config:       json.RawMessage(`{"valueField":"total"}`),
			},
			{
				ID:       "detail-table",
				Title:    "Details",
				Type:     dashboardir.WidgetTypeTable,
				Position: dashboardir.Position{X: 0, Y: 4, W: 12, H: 6},
				Config:   json.RawMessage(`{"columns":[]}`),
			},
		},
		Variables: []dashboardir.Variable{
			{ID: "period", Name: "Period", Type: "select", Default: "Q3"},
		},
	}

	page := DashboardToPageSpec(d)

	if page.APIVersion != uispec.APIVersion {
		t.Errorf("APIVersion = %q, want %q", page.APIVersion, uispec.APIVersion)
	}
	if page.Kind != uispec.KindPage {
		t.Errorf("Kind = %q, want %q", page.Kind, uispec.KindPage)
	}
	if page.Profile != uispec.ProfileDashboard {
		t.Errorf("Profile = %q, want %q", page.Profile, uispec.ProfileDashboard)
	}
	if page.Metadata.ID != "sales-overview" {
		t.Errorf("Metadata.ID = %q", page.Metadata.ID)
	}
	if page.Metadata.Title != "Sales Overview" {
		t.Errorf("Metadata.Title = %q", page.Metadata.Title)
	}

	if page.Layout.Type != uispec.LayoutResponsiveGrid {
		t.Errorf("Layout.Type = %q, want %q", page.Layout.Type, uispec.LayoutResponsiveGrid)
	}
	if page.Layout.Config.Columns != 12 {
		t.Errorf("Layout.Config.Columns = %d", page.Layout.Config.Columns)
	}

	if page.Theme == nil {
		t.Fatal("Theme is nil")
	}
	if page.Theme.Variant != "dark" {
		t.Errorf("Theme.Variant = %q", page.Theme.Variant)
	}
	if page.Theme.Tokens["color-primary"] != "#3b82f6" {
		t.Errorf("Theme.Tokens[color-primary] = %q", page.Theme.Tokens["color-primary"])
	}

	if len(page.Components) != 3 {
		t.Fatalf("Components len = %d, want 3", len(page.Components))
	}

	revenue := page.Components[0]
	if revenue.Type != "analytics.line-chart" {
		t.Errorf("revenue.Type = %q", revenue.Type)
	}
	if revenue.Position.Col != 0 || revenue.Position.Row != 0 {
		t.Errorf("revenue.Position = %+v", revenue.Position)
	}
	if revenue.Position.ColSpan != 6 || revenue.Position.RowSpan != 4 {
		t.Errorf("revenue.Position span = %d x %d", revenue.Position.ColSpan, revenue.Position.RowSpan)
	}
	if _, ok := revenue.Data["primary"]; !ok {
		t.Error("revenue missing primary data binding")
	}
	if revenue.RawConfig == nil {
		t.Error("revenue.RawConfig is nil")
	}

	total := page.Components[1]
	if total.Type != "analytics.metric" {
		t.Errorf("total.Type = %q", total.Type)
	}
	if total.Properties["title"] != "Total Sales" {
		t.Errorf("total.Properties[title] = %v", total.Properties["title"])
	}

	table := page.Components[2]
	if table.Type != "analytics.table" {
		t.Errorf("table.Type = %q", table.Type)
	}

	if len(page.Interactions) != 1 {
		t.Fatalf("Interactions len = %d, want 1", len(page.Interactions))
	}
	interaction := page.Interactions[0]
	if interaction.When.Component != "revenue" {
		t.Errorf("interaction.When.Component = %q", interaction.When.Component)
	}
	if interaction.Then[0].Target != "detail-table" {
		t.Errorf("interaction target = %q", interaction.Then[0].Target)
	}

	if page.Context["period"] != "Q3" {
		t.Errorf("Context[period] = %q", page.Context["period"])
	}
}

func TestDashboardToPageSpec_Minimal(t *testing.T) {
	d := &dashboardir.Dashboard{
		ID:    "empty",
		Title: "Empty Dashboard",
		Layout: dashboardir.Layout{
			Type:    dashboardir.LayoutTypeGrid,
			Columns: 6,
		},
	}

	page := DashboardToPageSpec(d)

	if page.Profile != uispec.ProfileDashboard {
		t.Errorf("Profile = %q", page.Profile)
	}
	if len(page.Components) != 0 {
		t.Errorf("Components len = %d", len(page.Components))
	}
	if page.Theme != nil {
		t.Error("Theme should be nil for dashboard without theme")
	}
}

func TestDashboardToPageSpec_RoundTrip(t *testing.T) {
	d := &dashboardir.Dashboard{
		ID:    "rt",
		Title: "Round Trip",
		Layout: dashboardir.Layout{
			Type:    dashboardir.LayoutTypeGrid,
			Columns: 12,
			Gap:     8,
		},
		Widgets: []dashboardir.Widget{
			{
				ID:       "w1",
				Type:     dashboardir.WidgetTypeText,
				Position: dashboardir.Position{X: 0, Y: 0, W: 12, H: 2},
				Config:   json.RawMessage(`{"content":"hello"}`),
			},
		},
	}

	page := DashboardToPageSpec(d)

	data, err := json.Marshal(page)
	if err != nil {
		t.Fatal(err)
	}

	var got uispec.PageSpec
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatal(err)
	}

	if got.Components[0].Type != "core.text" {
		t.Errorf("round-trip type = %q", got.Components[0].Type)
	}
}
