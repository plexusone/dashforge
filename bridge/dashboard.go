// Package bridge converts legacy DashboardIR definitions to UISpec PageSpecs.
package bridge

import (
	"fmt"

	"github.com/plexusone/uiforge/dashboardir"
	"github.com/plexusone/uiforge/uispec"
)

// DashboardToPageSpec converts a DashboardIR Dashboard to a UISpec PageSpec
// with the dashboard profile applied.
func DashboardToPageSpec(d *dashboardir.Dashboard) *uispec.PageSpec {
	page := &uispec.PageSpec{
		APIVersion: uispec.APIVersion,
		Kind:       uispec.KindPage,
		Metadata: uispec.PageMetadata{
			ID:          d.ID,
			Name:        d.ID,
			Title:       d.Title,
			Description: d.Description,
			Version:     d.Version,
		},
		Profile: uispec.ProfileDashboard,
		Layout:  convertLayout(&d.Layout),
	}

	if d.Theme != nil {
		page.Theme = &uispec.ThemeRef{
			ID:      "default",
			Variant: d.Theme.Mode,
			Tokens:  convertThemeTokens(d.Theme),
		}
	}

	for _, w := range d.Widgets {
		page.Components = append(page.Components, convertWidget(&w))
	}

	for _, w := range d.Widgets {
		if w.DrillDown != nil && w.DrillDown.Type == dashboardir.DrillDownTypeFilter {
			page.Interactions = append(page.Interactions, convertDrillDown(&w))
		}
	}

	if len(d.Variables) > 0 {
		if page.Context == nil {
			page.Context = make(map[string]string)
		}
		for _, v := range d.Variables {
			page.Context[v.ID] = v.Default
		}
	}

	return page
}

func convertLayout(l *dashboardir.Layout) uispec.LayoutSpec {
	return uispec.LayoutSpec{
		Type: uispec.LayoutResponsiveGrid,
		Config: &uispec.LayoutConfig{
			Columns: l.Columns,
			Gap:     fmt.Sprintf("%dpx", l.Gap),
		},
	}
}

func convertWidget(w *dashboardir.Widget) uispec.ComponentInstance {
	comp := uispec.ComponentInstance{
		ID:   w.ID,
		Type: widgetTypeToComponentType(w.Type),
		Position: &uispec.Position{
			Col:     w.Position.X,
			Row:     w.Position.Y,
			ColSpan: w.Position.W,
			RowSpan: w.Position.H,
		},
		RawConfig: w.Config,
	}

	props := map[string]any{}
	if w.Title != "" {
		props["title"] = w.Title
	}
	if w.Description != "" {
		props["description"] = w.Description
	}
	if len(props) > 0 {
		comp.Properties = props
	}

	if w.DataSourceID != "" {
		comp.Data = map[string]uispec.Binding{
			"primary": {
				Source:    w.DataSourceID,
				Operation: "query",
			},
		}
	}

	if w.Visible != nil && !*w.Visible {
		comp.Visibility = &uispec.VisibilityRule{
			Condition: "false",
		}
	}

	return comp
}

func widgetTypeToComponentType(wt string) string {
	switch wt {
	case dashboardir.WidgetTypeChart:
		return "analytics.line-chart"
	case dashboardir.WidgetTypeTable:
		return "analytics.table"
	case dashboardir.WidgetTypeMetric:
		return "analytics.metric"
	case dashboardir.WidgetTypeText:
		return "core.text"
	case dashboardir.WidgetTypeImage:
		return "core.image"
	default:
		return "core.unknown"
	}
}

func convertThemeTokens(t *dashboardir.Theme) map[string]string {
	tokens := map[string]string{}
	if t.PrimaryColor != "" {
		tokens["color-primary"] = t.PrimaryColor
	}
	if t.BackgroundColor != "" {
		tokens["color-background"] = t.BackgroundColor
	}
	if t.FontFamily != "" {
		tokens["font-family"] = t.FontFamily
	}
	return tokens
}

func convertDrillDown(w *dashboardir.Widget) uispec.Interaction {
	actions := []uispec.InteractionAction{
		{
			Target: w.DrillDown.Target,
			Action: "filter",
			Params: anyMap(w.DrillDown.Params),
		},
	}
	return uispec.Interaction{
		When: uispec.InteractionTrigger{
			Component: w.ID,
			Event:     "click",
		},
		Then: actions,
	}
}

func anyMap(m map[string]string) map[string]any {
	if m == nil {
		return nil
	}
	result := make(map[string]any, len(m))
	for k, v := range m {
		result[k] = v
	}
	return result
}
