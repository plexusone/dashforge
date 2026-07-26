package registry

import (
	"testing"

	"github.com/plexusone/uiforge/uispec"
)

func TestValidateProfile_Dashboard(t *testing.T) {
	page := &uispec.PageSpec{
		APIVersion: uispec.APIVersion,
		Kind:       uispec.KindPage,
		Metadata:   uispec.PageMetadata{ID: "d", Name: "d"},
		Profile:    uispec.ProfileDashboard,
		Layout:     uispec.LayoutSpec{Type: uispec.LayoutResponsiveGrid},
		Components: []uispec.ComponentInstance{
			{ID: "chart", Type: "analytics.line-chart"},
			{ID: "card", Type: "core.card"},
		},
	}

	if err := ValidateProfile(page); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestValidateProfile_DashboardRejectsAssistant(t *testing.T) {
	page := &uispec.PageSpec{
		APIVersion: uispec.APIVersion,
		Kind:       uispec.KindPage,
		Metadata:   uispec.PageMetadata{ID: "d", Name: "d"},
		Profile:    uispec.ProfileDashboard,
		Layout:     uispec.LayoutSpec{Type: uispec.LayoutResponsiveGrid},
		Components: []uispec.ComponentInstance{
			{ID: "thread", Type: "assistant.thread"},
		},
	}

	if err := ValidateProfile(page); err == nil {
		t.Error("expected error: assistant namespace not allowed in dashboard profile")
	}
}

func TestValidateProfile_DashboardRejectsSplitPane(t *testing.T) {
	page := &uispec.PageSpec{
		APIVersion: uispec.APIVersion,
		Kind:       uispec.KindPage,
		Metadata:   uispec.PageMetadata{ID: "d", Name: "d"},
		Profile:    uispec.ProfileDashboard,
		Layout: uispec.LayoutSpec{
			Type:    uispec.LayoutSplitPane,
			Regions: []uispec.LayoutRegion{{Name: "a"}, {Name: "b"}},
		},
	}

	if err := ValidateProfile(page); err == nil {
		t.Error("expected error: split-pane not allowed in dashboard profile")
	}
}

func TestValidateProfile_Agent(t *testing.T) {
	page := &uispec.PageSpec{
		APIVersion: uispec.APIVersion,
		Kind:       uispec.KindPage,
		Metadata:   uispec.PageMetadata{ID: "a", Name: "a"},
		Profile:    uispec.ProfileAgent,
		Layout: uispec.LayoutSpec{
			Type: uispec.LayoutSplitPane,
			Regions: []uispec.LayoutRegion{
				{Name: "sidebar"},
				{Name: "main"},
			},
		},
		Components: []uispec.ComponentInstance{
			{ID: "tl", Type: "assistant.thread-list"},
			{ID: "t", Type: "assistant.thread"},
			{ID: "c", Type: "core.card"},
		},
	}

	if err := ValidateProfile(page); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestValidateProfile_AgentRejectsAnalytics(t *testing.T) {
	page := &uispec.PageSpec{
		APIVersion: uispec.APIVersion,
		Kind:       uispec.KindPage,
		Metadata:   uispec.PageMetadata{ID: "a", Name: "a"},
		Profile:    uispec.ProfileAgent,
		Layout:     uispec.LayoutSpec{Type: uispec.LayoutStack},
		Components: []uispec.ComponentInstance{
			{ID: "chart", Type: "analytics.line-chart"},
		},
	}

	if err := ValidateProfile(page); err == nil {
		t.Error("expected error: analytics namespace not allowed in agent profile")
	}
}

func TestValidateProfile_ApplicationRequiresMainSlot(t *testing.T) {
	page := &uispec.PageSpec{
		APIVersion: uispec.APIVersion,
		Kind:       uispec.KindPage,
		Metadata:   uispec.PageMetadata{ID: "a", Name: "a"},
		Profile:    uispec.ProfileApplication,
		Layout: uispec.LayoutSpec{
			Type:    uispec.LayoutApplicationShell,
			Regions: []uispec.LayoutRegion{{Name: "nav"}, {Name: "sidebar"}},
		},
	}

	if err := ValidateProfile(page); err == nil {
		t.Error("expected error: application profile requires 'main' slot")
	}
}

func TestValidateProfile_EmbeddedMaxDepth(t *testing.T) {
	page := &uispec.PageSpec{
		APIVersion: uispec.APIVersion,
		Kind:       uispec.KindPage,
		Metadata:   uispec.PageMetadata{ID: "e", Name: "e"},
		Profile:    uispec.ProfileEmbedded,
		Layout: uispec.LayoutSpec{
			Type: uispec.LayoutStack,
			Regions: []uispec.LayoutRegion{
				{
					Name: "inner",
					Layout: &uispec.LayoutSpec{
						Type: uispec.LayoutStack,
					},
				},
			},
		},
	}

	if err := ValidateProfile(page); err == nil {
		t.Error("expected error: embedded profile max depth is 1")
	}
}

func TestValidateProfile_NoProfile(t *testing.T) {
	page := &uispec.PageSpec{
		APIVersion: uispec.APIVersion,
		Kind:       uispec.KindPage,
		Metadata:   uispec.PageMetadata{ID: "x", Name: "x"},
		Layout:     uispec.LayoutSpec{Type: uispec.LayoutStack},
	}

	if err := ValidateProfile(page); err != nil {
		t.Errorf("unexpected error for no profile: %v", err)
	}
}

func TestValidateProfile_Unknown(t *testing.T) {
	page := &uispec.PageSpec{
		APIVersion: uispec.APIVersion,
		Kind:       uispec.KindPage,
		Metadata:   uispec.PageMetadata{ID: "x", Name: "x"},
		Profile:    "fantasy",
		Layout:     uispec.LayoutSpec{Type: uispec.LayoutStack},
	}

	if err := ValidateProfile(page); err == nil {
		t.Error("expected error for unknown profile")
	}
}
