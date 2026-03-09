package datasource

import (
	"sort"
	"sync"
)

var (
	providersMu sync.RWMutex
	providers   = make(map[string]Provider)
)

// Register adds a provider to the registry.
// It should be called from init() in provider packages.
// If a provider with the same name already exists, it is replaced.
func Register(p Provider) {
	providersMu.Lock()
	defer providersMu.Unlock()
	providers[p.Name()] = p
}

// Get retrieves a provider by name.
// Returns the provider and true if found, nil and false otherwise.
func Get(name string) (Provider, bool) {
	providersMu.RLock()
	defer providersMu.RUnlock()
	p, ok := providers[name]
	return p, ok
}

// MustGet retrieves a provider by name or panics if not found.
func MustGet(name string) Provider {
	p, ok := Get(name)
	if !ok {
		panic("datasource: unknown provider " + name)
	}
	return p
}

// Available returns all registered provider names in sorted order.
func Available() []string {
	providersMu.RLock()
	defer providersMu.RUnlock()

	names := make([]string, 0, len(providers))
	for name := range providers {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// Registered returns a copy of all registered providers.
func Registered() map[string]Provider {
	providersMu.RLock()
	defer providersMu.RUnlock()

	result := make(map[string]Provider, len(providers))
	for name, p := range providers {
		result[name] = p
	}
	return result
}

// Unregister removes a provider from the registry.
// This is primarily useful for testing.
func Unregister(name string) {
	providersMu.Lock()
	defer providersMu.Unlock()
	delete(providers, name)
}

// Reset clears all registered providers.
// This is primarily useful for testing.
func Reset() {
	providersMu.Lock()
	defer providersMu.Unlock()
	providers = make(map[string]Provider)
}
