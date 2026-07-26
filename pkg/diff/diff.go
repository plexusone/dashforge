// Package diff provides structural comparison of UISpec PageSpecs,
// producing a list of changes suitable for code review or audit trails.
package diff

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/plexusone/uiforge/uispec"
)

// ChangeType classifies a diff entry.
const (
	Added    = "added"
	Removed  = "removed"
	Modified = "modified"
)

// Change represents a single structural difference.
type Change struct {
	Path string `json:"path"`
	Type string `json:"type"`
	Old  any    `json:"old,omitempty"`
	New  any    `json:"new,omitempty"`
}

// DiffResult holds the complete set of changes between two PageSpecs.
type DiffResult struct {
	Changes []Change `json:"changes"`
}

// HasChanges reports whether any differences were found.
func (r *DiffResult) HasChanges() bool {
	return len(r.Changes) > 0
}

// Summary returns a human-readable summary of the diff.
func (r *DiffResult) Summary() string {
	if !r.HasChanges() {
		return "No changes"
	}

	var added, removed, modified int
	for _, c := range r.Changes {
		switch c.Type {
		case Added:
			added++
		case Removed:
			removed++
		case Modified:
			modified++
		}
	}

	parts := make([]string, 0, 3)
	if added > 0 {
		parts = append(parts, fmt.Sprintf("%d added", added))
	}
	if removed > 0 {
		parts = append(parts, fmt.Sprintf("%d removed", removed))
	}
	if modified > 0 {
		parts = append(parts, fmt.Sprintf("%d modified", modified))
	}
	return fmt.Sprintf("%d changes: %s", len(r.Changes), strings.Join(parts, ", "))
}

// PageSpecs compares two PageSpecs and returns all structural differences.
func PageSpecs(a, b *uispec.PageSpec) *DiffResult {
	result := &DiffResult{}

	if a == nil && b == nil {
		return result
	}
	if a == nil {
		result.Changes = append(result.Changes, Change{Path: "", Type: Added, New: "entire page"})
		return result
	}
	if b == nil {
		result.Changes = append(result.Changes, Change{Path: "", Type: Removed, Old: "entire page"})
		return result
	}

	diffMetadata(a, b, result)
	diffScalar(a.Profile, b.Profile, "profile", result)
	diffLayout(a, b, result)
	diffComponents(a.Components, b.Components, result)
	diffInteractions(a.Interactions, b.Interactions, result)
	diffNavigation(a.Navigation, b.Navigation, result)
	diffTheme(a.Theme, b.Theme, result)
	diffContext(a.Context, b.Context, result)

	return result
}

func diffMetadata(a, b *uispec.PageSpec, r *DiffResult) {
	if a.Metadata.ID != b.Metadata.ID {
		r.Changes = append(r.Changes, Change{Path: "metadata.id", Type: Modified, Old: a.Metadata.ID, New: b.Metadata.ID})
	}
	if a.Metadata.Name != b.Metadata.Name {
		r.Changes = append(r.Changes, Change{Path: "metadata.name", Type: Modified, Old: a.Metadata.Name, New: b.Metadata.Name})
	}
	if a.Metadata.Title != b.Metadata.Title {
		r.Changes = append(r.Changes, Change{Path: "metadata.title", Type: Modified, Old: a.Metadata.Title, New: b.Metadata.Title})
	}
	if a.Metadata.Description != b.Metadata.Description {
		r.Changes = append(r.Changes, Change{Path: "metadata.description", Type: Modified, Old: a.Metadata.Description, New: b.Metadata.Description})
	}
	if a.Metadata.Version != b.Metadata.Version {
		r.Changes = append(r.Changes, Change{Path: "metadata.version", Type: Modified, Old: a.Metadata.Version, New: b.Metadata.Version})
	}
}

func diffScalar(a, b any, path string, r *DiffResult) {
	if !reflect.DeepEqual(a, b) {
		r.Changes = append(r.Changes, Change{Path: path, Type: Modified, Old: a, New: b})
	}
}

func diffLayout(a, b *uispec.PageSpec, r *DiffResult) {
	if a.Layout.Type != b.Layout.Type {
		r.Changes = append(r.Changes, Change{Path: "layout.type", Type: Modified, Old: a.Layout.Type, New: b.Layout.Type})
	}
	if !reflect.DeepEqual(a.Layout.Config, b.Layout.Config) {
		r.Changes = append(r.Changes, Change{Path: "layout.config", Type: Modified, Old: a.Layout.Config, New: b.Layout.Config})
	}
	if !reflect.DeepEqual(a.Layout.Regions, b.Layout.Regions) {
		r.Changes = append(r.Changes, Change{Path: "layout.regions", Type: Modified, Old: len(a.Layout.Regions), New: len(b.Layout.Regions)})
	}
}

func diffComponents(aComps, bComps []uispec.ComponentInstance, r *DiffResult) {
	aMap := make(map[string]uispec.ComponentInstance, len(aComps))
	for _, c := range aComps {
		aMap[c.ID] = c
	}
	bMap := make(map[string]uispec.ComponentInstance, len(bComps))
	for _, c := range bComps {
		bMap[c.ID] = c
	}

	for id, ac := range aMap {
		bc, ok := bMap[id]
		if !ok {
			r.Changes = append(r.Changes, Change{
				Path: fmt.Sprintf("components[%s]", id),
				Type: Removed,
				Old:  ac.Type,
			})
			continue
		}
		diffComponent(ac, bc, id, r)
	}

	for id, bc := range bMap {
		if _, ok := aMap[id]; !ok {
			r.Changes = append(r.Changes, Change{
				Path: fmt.Sprintf("components[%s]", id),
				Type: Added,
				New:  bc.Type,
			})
		}
	}
}

func diffComponent(a, b uispec.ComponentInstance, id string, r *DiffResult) {
	prefix := fmt.Sprintf("components[%s]", id)
	if a.Type != b.Type {
		r.Changes = append(r.Changes, Change{Path: prefix + ".type", Type: Modified, Old: a.Type, New: b.Type})
	}
	if !reflect.DeepEqual(a.Properties, b.Properties) {
		r.Changes = append(r.Changes, Change{Path: prefix + ".properties", Type: Modified})
	}
	if !reflect.DeepEqual(a.Data, b.Data) {
		r.Changes = append(r.Changes, Change{Path: prefix + ".data", Type: Modified})
	}
	if !reflect.DeepEqual(a.Position, b.Position) {
		r.Changes = append(r.Changes, Change{Path: prefix + ".position", Type: Modified})
	}
	if a.Slot != b.Slot {
		r.Changes = append(r.Changes, Change{Path: prefix + ".slot", Type: Modified, Old: a.Slot, New: b.Slot})
	}
	if !reflect.DeepEqual(a.Style, b.Style) {
		r.Changes = append(r.Changes, Change{Path: prefix + ".style", Type: Modified})
	}
	if !reflect.DeepEqual(a.Visibility, b.Visibility) {
		r.Changes = append(r.Changes, Change{Path: prefix + ".visibility", Type: Modified})
	}
	if !reflect.DeepEqual(a.Children, b.Children) {
		r.Changes = append(r.Changes, Change{Path: prefix + ".children", Type: Modified, Old: len(a.Children), New: len(b.Children)})
	}
}

func diffInteractions(aIx, bIx []uispec.Interaction, r *DiffResult) {
	if len(aIx) != len(bIx) {
		r.Changes = append(r.Changes, Change{
			Path: "interactions",
			Type: Modified,
			Old:  len(aIx),
			New:  len(bIx),
		})
		return
	}
	for i := range aIx {
		if !reflect.DeepEqual(aIx[i], bIx[i]) {
			r.Changes = append(r.Changes, Change{
				Path: fmt.Sprintf("interactions[%d]", i),
				Type: Modified,
			})
		}
	}
}

func diffNavigation(a, b *uispec.NavigationSpec, r *DiffResult) {
	if !reflect.DeepEqual(a, b) {
		if a == nil {
			r.Changes = append(r.Changes, Change{Path: "navigation", Type: Added})
		} else if b == nil {
			r.Changes = append(r.Changes, Change{Path: "navigation", Type: Removed})
		} else {
			r.Changes = append(r.Changes, Change{Path: "navigation", Type: Modified})
		}
	}
}

func diffTheme(a, b *uispec.ThemeRef, r *DiffResult) {
	if !reflect.DeepEqual(a, b) {
		if a == nil {
			r.Changes = append(r.Changes, Change{Path: "theme", Type: Added})
		} else if b == nil {
			r.Changes = append(r.Changes, Change{Path: "theme", Type: Removed})
		} else {
			r.Changes = append(r.Changes, Change{Path: "theme", Type: Modified})
		}
	}
}

func diffContext(a, b map[string]string, r *DiffResult) {
	if !reflect.DeepEqual(a, b) {
		r.Changes = append(r.Changes, Change{Path: "context", Type: Modified})
	}
}
