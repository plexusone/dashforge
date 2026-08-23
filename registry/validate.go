package registry

import (
	"fmt"
	"strings"

	"github.com/plexusone/dashforge/uispec"
)

// ValidationError collects all problems found during validation.
type ValidationError struct {
	Errors []string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation failed with %d error(s):\n  %s",
		len(e.Errors), strings.Join(e.Errors, "\n  "))
}

func (e *ValidationError) add(msg string, args ...any) {
	e.Errors = append(e.Errors, fmt.Sprintf(msg, args...))
}

// ValidatePage checks that a PageSpec references only registered components
// and that data bindings reference declared data inputs.
func (r *Registry) ValidatePage(page *uispec.PageSpec) error {
	ve := &ValidationError{}

	if page.APIVersion != uispec.APIVersion {
		ve.add("apiVersion %q is not supported (expected %q)", page.APIVersion, uispec.APIVersion)
	}
	if page.Kind != uispec.KindPage {
		ve.add("kind %q is not supported (expected %q)", page.Kind, uispec.KindPage)
	}
	if page.Metadata.ID == "" {
		ve.add("metadata.id is required")
	}
	if page.Metadata.Name == "" {
		ve.add("metadata.name is required")
	}

	validateLayout(ve, &page.Layout)

	seen := map[string]bool{}
	for i, comp := range page.Components {
		if comp.ID == "" {
			ve.add("components[%d]: id is required", i)
		} else if seen[comp.ID] {
			ve.add("components[%d]: duplicate id %q", i, comp.ID)
		}
		seen[comp.ID] = true

		r.validateComponent(ve, &comp, i)
	}

	for i, interaction := range page.Interactions {
		if interaction.When.Component == "" {
			ve.add("interactions[%d].when.component is required", i)
		} else if !seen[interaction.When.Component] {
			ve.add("interactions[%d].when.component %q not found in page components", i, interaction.When.Component)
		}
		if interaction.When.Event == "" {
			ve.add("interactions[%d].when.event is required", i)
		}
		for j, action := range interaction.Then {
			if action.Target == "" {
				ve.add("interactions[%d].then[%d].target is required", i, j)
			}
			if action.Action == "" {
				ve.add("interactions[%d].then[%d].action is required", i, j)
			}
		}
	}

	if page.Profile != "" {
		if err := ValidateProfile(page); err != nil {
			if pve, ok := err.(*ValidationError); ok {
				ve.Errors = append(ve.Errors, pve.Errors...)
			} else {
				ve.add("profile: %v", err)
			}
		}
	}

	if len(ve.Errors) > 0 {
		return ve
	}
	return nil
}

func (r *Registry) validateComponent(ve *ValidationError, comp *uispec.ComponentInstance, idx int) {
	spec := r.Get(comp.Type)
	if spec == nil {
		ve.add("components[%d] (%s): type %q is not registered", idx, comp.ID, comp.Type)
		return
	}

	for bindingName := range comp.Data {
		if spec.DataInputs == nil {
			ve.add("components[%d] (%s): data binding %q but component declares no data inputs",
				idx, comp.ID, bindingName)
			continue
		}
		if _, ok := spec.DataInputs[bindingName]; !ok {
			ve.add("components[%d] (%s): data binding %q not declared in component data inputs",
				idx, comp.ID, bindingName)
		}
	}

	if spec.DataInputs != nil {
		for name, input := range spec.DataInputs {
			if input.Required {
				if _, ok := comp.Data[name]; !ok {
					ve.add("components[%d] (%s): required data input %q is not bound",
						idx, comp.ID, name)
				}
			}
		}
	}

	for _, child := range comp.Children {
		r.validateComponent(ve, &child, idx)
	}
}

var validLayoutTypes = map[string]bool{
	uispec.LayoutResponsiveGrid:   true,
	uispec.LayoutStack:            true,
	uispec.LayoutSplitPane:        true,
	uispec.LayoutTabs:             true,
	uispec.LayoutApplicationShell: true,
}

func validateLayout(ve *ValidationError, layout *uispec.LayoutSpec) {
	if layout.Type == "" {
		ve.add("layout.type is required")
		return
	}
	if !validLayoutTypes[layout.Type] {
		ve.add("layout.type %q is not a valid layout primitive", layout.Type)
	}

	switch layout.Type {
	case uispec.LayoutResponsiveGrid:
		validateGridLayout(ve, layout)
	case uispec.LayoutSplitPane:
		validateSplitPaneLayout(ve, layout)
	case uispec.LayoutTabs:
		validateTabsLayout(ve, layout)
	case uispec.LayoutApplicationShell:
		validateAppShellLayout(ve, layout)
	}

	for i, region := range layout.Regions {
		if region.Name == "" {
			ve.add("layout.regions[%d]: name is required", i)
		}
		if region.Layout != nil {
			validateLayout(ve, region.Layout)
		}
	}
}

func validateGridLayout(ve *ValidationError, layout *uispec.LayoutSpec) {
	if layout.Config != nil && layout.Config.Columns < 0 {
		ve.add("responsive-grid: columns must be >= 0")
	}
}

func validateSplitPaneLayout(ve *ValidationError, layout *uispec.LayoutSpec) {
	if len(layout.Regions) < 2 {
		ve.add("split-pane: requires at least 2 regions")
	}
	if layout.Config != nil && len(layout.Config.Sizes) > 0 {
		if len(layout.Config.Sizes) != len(layout.Regions) {
			ve.add("split-pane: sizes count (%d) must match regions count (%d)",
				len(layout.Config.Sizes), len(layout.Regions))
		}
	}
}

func validateTabsLayout(ve *ValidationError, layout *uispec.LayoutSpec) {
	if len(layout.Regions) < 1 {
		ve.add("tabs: requires at least 1 region (tab)")
	}
}

func validateAppShellLayout(ve *ValidationError, layout *uispec.LayoutSpec) {
	required := map[string]bool{"main": false}
	for _, r := range layout.Regions {
		if r.Name == "main" {
			required["main"] = true
		}
	}
	if !required["main"] {
		ve.add("application-shell: requires a 'main' region")
	}
}
