// Package state provides page-level shared state management with
// dot-path access and subscription-based change notification.
package state

import (
	"fmt"
	"strings"
	"sync"
)

// Store is a concurrency-safe key-value store with dot-path access.
type Store struct {
	mu       sync.RWMutex
	values   map[string]any
	watchers []watcher
	nextID   int
}

type watcher struct {
	id     int
	prefix string
	fn     func(key string, value any)
}

// New creates an empty state store.
func New() *Store {
	return &Store{
		values: make(map[string]any),
	}
}

// Get retrieves a value by dot-path. Returns the value and whether it was found.
func (s *Store) Get(path string) (any, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return getPath(s.values, path)
}

// Set writes a value at the given dot-path, creating intermediate maps as needed.
// Notifies all matching watchers.
func (s *Store) Set(path string, value any) {
	s.mu.Lock()
	setPath(s.values, path, value)
	watchers := s.matchingWatchers(path)
	s.mu.Unlock()

	for _, w := range watchers {
		w.fn(path, value)
	}
}

// Toggle flips a boolean value at the given path.
func (s *Store) Toggle(path string) error {
	s.mu.Lock()
	val, ok := getPath(s.values, path)
	if !ok {
		s.mu.Unlock()
		return fmt.Errorf("state: path %q not found", path)
	}
	b, ok := val.(bool)
	if !ok {
		s.mu.Unlock()
		return fmt.Errorf("state: path %q is %T, not bool", path, val)
	}
	newVal := !b
	setPath(s.values, path, newVal)
	watchers := s.matchingWatchers(path)
	s.mu.Unlock()

	for _, w := range watchers {
		w.fn(path, newVal)
	}
	return nil
}

// Subscribe registers a callback for changes at or under pathPrefix.
// Returns an unsubscribe function.
func (s *Store) Subscribe(pathPrefix string, fn func(key string, value any)) func() {
	s.mu.Lock()
	s.nextID++
	id := s.nextID
	s.watchers = append(s.watchers, watcher{id: id, prefix: pathPrefix, fn: fn})
	s.mu.Unlock()

	return func() {
		s.mu.Lock()
		defer s.mu.Unlock()
		for i, w := range s.watchers {
			if w.id == id {
				s.watchers = append(s.watchers[:i], s.watchers[i+1:]...)
				return
			}
		}
	}
}

// Snapshot returns a deep copy of all state values.
func (s *Store) Snapshot() map[string]any {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return deepCopy(s.values)
}

// Load bulk-loads initial state, replacing any existing values.
func (s *Store) Load(initial map[string]any) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.values = deepCopy(initial)
}

// AsContext returns the values map for use as an expression context.
// The returned map is the live reference (caller must hold no write lock).
func (s *Store) AsContext() map[string]any {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return deepCopy(s.values)
}

// matchingWatchers returns watchers whose prefix matches path. Caller must hold mu.
func (s *Store) matchingWatchers(path string) []watcher {
	var matched []watcher
	for _, w := range s.watchers {
		if w.prefix == "" || path == w.prefix || strings.HasPrefix(path, w.prefix+".") {
			matched = append(matched, w)
		}
	}
	return matched
}

func getPath(m map[string]any, path string) (any, bool) {
	parts := strings.Split(path, ".")
	var current any = m
	for _, part := range parts {
		cm, ok := current.(map[string]any)
		if !ok {
			return nil, false
		}
		current, ok = cm[part]
		if !ok {
			return nil, false
		}
	}
	return current, true
}

func setPath(m map[string]any, path string, value any) {
	parts := strings.Split(path, ".")
	current := m
	for _, part := range parts[:len(parts)-1] {
		next, ok := current[part]
		if !ok {
			next = make(map[string]any)
			current[part] = next
		}
		nm, ok := next.(map[string]any)
		if !ok {
			nm = make(map[string]any)
			current[part] = nm
		}
		current = nm
	}
	current[parts[len(parts)-1]] = value
}

func deepCopy(src map[string]any) map[string]any {
	if src == nil {
		return nil
	}
	dst := make(map[string]any, len(src))
	for k, v := range src {
		switch val := v.(type) {
		case map[string]any:
			dst[k] = deepCopy(val)
		case []any:
			cp := make([]any, len(val))
			copy(cp, val)
			dst[k] = cp
		default:
			dst[k] = v
		}
	}
	return dst
}
