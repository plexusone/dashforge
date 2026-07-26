package registry

import (
	"fmt"
	"strings"
	"sync"
)

// Registry is an in-memory store of ComponentSpec manifests.
type Registry struct {
	mu         sync.RWMutex
	components map[string]*ComponentSpec // keyed by component ID
}

// New creates an empty registry.
func New() *Registry {
	return &Registry{
		components: make(map[string]*ComponentSpec),
	}
}

// Register adds or replaces a component manifest in the registry.
// The component ID must be in "namespace.name" format.
func (r *Registry) Register(spec *ComponentSpec) error {
	if spec.ID == "" {
		return fmt.Errorf("component ID is required")
	}
	if !strings.Contains(spec.ID, ".") {
		return fmt.Errorf("component ID %q must be in namespace.name format", spec.ID)
	}
	if spec.Version == "" {
		return fmt.Errorf("component %q: version is required", spec.ID)
	}

	parts := strings.SplitN(spec.ID, ".", 2)
	spec.Namespace = parts[0]

	r.mu.Lock()
	r.components[spec.ID] = spec
	r.mu.Unlock()
	return nil
}

// Get retrieves a component manifest by ID. Returns nil if not found.
func (r *Registry) Get(id string) *ComponentSpec {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.components[id]
}

// List returns all registered component manifests.
func (r *Registry) List() []*ComponentSpec {
	r.mu.RLock()
	defer r.mu.RUnlock()
	result := make([]*ComponentSpec, 0, len(r.components))
	for _, spec := range r.components {
		result = append(result, spec)
	}
	return result
}

// ListByNamespace returns components matching the given namespace.
func (r *Registry) ListByNamespace(ns string) []*ComponentSpec {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var result []*ComponentSpec
	for _, spec := range r.components {
		if spec.Namespace == ns {
			result = append(result, spec)
		}
	}
	return result
}

// Has returns true if a component with the given ID is registered.
func (r *Registry) Has(id string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	_, ok := r.components[id]
	return ok
}

// Count returns the number of registered components.
func (r *Registry) Count() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.components)
}
