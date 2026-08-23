package registry

import (
	"fmt"
	"strings"

	"github.com/plexusone/dashforge/uispec"
)

// BuiltinProfiles returns the default profile constraint definitions.
func BuiltinProfiles() map[string]*uispec.ProfileConstraints {
	return map[string]*uispec.ProfileConstraints{
		uispec.ProfileDashboard: {
			Name:              uispec.ProfileDashboard,
			AllowedLayouts:    []string{uispec.LayoutResponsiveGrid, uispec.LayoutStack, uispec.LayoutTabs},
			AllowedNamespaces: []string{"core", "analytics"},
			MaxDepth:          2,
		},
		uispec.ProfileApplication: {
			Name:              uispec.ProfileApplication,
			AllowedLayouts:    []string{uispec.LayoutApplicationShell, uispec.LayoutStack, uispec.LayoutTabs, uispec.LayoutResponsiveGrid},
			AllowedNamespaces: []string{"core", "analytics", "application"},
			RequiredSlots:     []string{"main"},
			MaxDepth:          3,
		},
		uispec.ProfileAgent: {
			Name:              uispec.ProfileAgent,
			AllowedLayouts:    []string{uispec.LayoutSplitPane, uispec.LayoutApplicationShell, uispec.LayoutStack, uispec.LayoutTabs},
			AllowedNamespaces: []string{"core", "assistant", "agentos"},
			MaxDepth:          3,
		},
		uispec.ProfilePortal: {
			Name:              uispec.ProfilePortal,
			AllowedLayouts:    []string{uispec.LayoutApplicationShell, uispec.LayoutStack, uispec.LayoutTabs, uispec.LayoutResponsiveGrid},
			AllowedNamespaces: []string{"core", "analytics", "application", "assistant"},
			RequiredSlots:     []string{"main"},
			MaxDepth:          4,
		},
		uispec.ProfileEmbedded: {
			Name:              uispec.ProfileEmbedded,
			AllowedLayouts:    []string{uispec.LayoutStack},
			AllowedNamespaces: []string{"core", "analytics"},
			MaxDepth:          1,
		},
	}
}

// ValidateProfile checks that a PageSpec conforms to its declared profile.
// Returns nil if no profile is set (unconstrained).
func ValidateProfile(page *uispec.PageSpec) error {
	if page.Profile == "" {
		return nil
	}

	profiles := BuiltinProfiles()
	pc, ok := profiles[page.Profile]
	if !ok {
		return fmt.Errorf("unknown profile %q", page.Profile)
	}

	ve := &ValidationError{}

	validateLayoutAllowed(ve, &page.Layout, pc)

	for i, comp := range page.Components {
		ns := componentNamespace(comp.Type)
		if !contains(pc.AllowedNamespaces, ns) {
			ve.add("components[%d] (%s): namespace %q not allowed in %s profile (allowed: %s)",
				i, comp.ID, ns, pc.Name, strings.Join(pc.AllowedNamespaces, ", "))
		}
	}

	if len(pc.RequiredSlots) > 0 {
		regionNames := collectRegionNames(&page.Layout)
		for _, slot := range pc.RequiredSlots {
			if !regionNames[slot] {
				ve.add("profile %s requires slot %q but it is not in the layout regions", pc.Name, slot)
			}
		}
	}

	if pc.MaxDepth > 0 {
		depth := layoutDepth(&page.Layout)
		if depth > pc.MaxDepth {
			ve.add("layout nesting depth %d exceeds profile %s maximum of %d", depth, pc.Name, pc.MaxDepth)
		}
	}

	if len(ve.Errors) > 0 {
		return ve
	}
	return nil
}

func validateLayoutAllowed(ve *ValidationError, layout *uispec.LayoutSpec, pc *uispec.ProfileConstraints) {
	if layout.Type != "" && !contains(pc.AllowedLayouts, layout.Type) {
		ve.add("layout type %q not allowed in %s profile (allowed: %s)",
			layout.Type, pc.Name, strings.Join(pc.AllowedLayouts, ", "))
	}
	for _, region := range layout.Regions {
		if region.Layout != nil {
			validateLayoutAllowed(ve, region.Layout, pc)
		}
	}
}

func componentNamespace(compType string) string {
	if idx := strings.Index(compType, "."); idx >= 0 {
		return compType[:idx]
	}
	return compType
}

func collectRegionNames(layout *uispec.LayoutSpec) map[string]bool {
	names := map[string]bool{}
	for _, r := range layout.Regions {
		names[r.Name] = true
		if r.Layout != nil {
			for k, v := range collectRegionNames(r.Layout) {
				names[k] = v
			}
		}
	}
	return names
}

func layoutDepth(layout *uispec.LayoutSpec) int {
	max := 1
	for _, r := range layout.Regions {
		if r.Layout != nil {
			d := 1 + layoutDepth(r.Layout)
			if d > max {
				max = d
			}
		}
	}
	return max
}

func contains(ss []string, s string) bool {
	for _, v := range ss {
		if v == s {
			return true
		}
	}
	return false
}
