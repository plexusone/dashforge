package analytics

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"
)

// SourceConfig describes one persisted analytics source. Only the secret
// reference (DSNRef) is ever stored; raw DSNs are rejected by Validate so
// credentials never reach a store or an API response.
type SourceConfig struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Connector string    `json:"connector"`
	DSNRef    string    `json:"dsnRef"`
	Enabled   bool      `json:"enabled"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

var sourceIDPattern = regexp.MustCompile(`^[a-z0-9][a-z0-9-]*$`)

// secretRefPattern requires an explicit scheme://path form. This is stricter
// than omnivault.IsSecretRef, which accepts anything before a bare colon as a
// scheme and would therefore let MySQL-style DSNs (user:pass@tcp(...)/db)
// through. Scheme existence is checked again at resolve time, which also
// rejects URL-style raw DSNs such as mysql://... because no vault provider is
// registered for that scheme.
var secretRefPattern = regexp.MustCompile(`^[a-z][a-z0-9+.-]*://\S+$`)

// IsSecretRefStrict reports whether s is an explicit scheme://path secret
// reference.
func IsSecretRefStrict(s string) bool {
	return secretRefPattern.MatchString(s)
}

// Validate checks that the config is well formed and safe to persist.
func (c SourceConfig) Validate() error {
	if strings.TrimSpace(c.ID) == "" {
		return fmt.Errorf("source id is required")
	}
	if !sourceIDPattern.MatchString(c.ID) {
		return fmt.Errorf("source id %q must be a lowercase slug (a-z, 0-9, -)", c.ID)
	}
	if strings.TrimSpace(c.Name) == "" {
		return fmt.Errorf("source name is required")
	}
	if strings.TrimSpace(c.Connector) == "" {
		return fmt.Errorf("source connector is required")
	}
	if _, ok := LookupConnector(c.Connector); !ok {
		return fmt.Errorf("unknown connector %q (registered: %s)", c.Connector, strings.Join(Connectors(), ", "))
	}
	if strings.TrimSpace(c.DSNRef) == "" {
		return fmt.Errorf("source dsnRef is required")
	}
	if !IsSecretRefStrict(c.DSNRef) {
		return fmt.Errorf("dsnRef must be a secret reference such as env://VAR_NAME; raw DSNs are not accepted")
	}
	return nil
}

// ConnectorFactory opens a QueryProvider from a resolved DSN. The DSN is only
// held in memory for the duration of the dial.
type ConnectorFactory func(dsn string) (QueryProvider, error)

var (
	connectorMu sync.RWMutex
	connectors  = map[string]ConnectorFactory{}
)

// RegisterConnector registers a connector factory under a stable name.
// Registering a duplicate name panics: connector names are compile-time
// wiring, so a collision is a programming error.
func RegisterConnector(name string, factory ConnectorFactory) {
	if name == "" || factory == nil {
		panic("analytics: RegisterConnector requires a name and factory")
	}
	connectorMu.Lock()
	defer connectorMu.Unlock()
	if _, exists := connectors[name]; exists {
		panic(fmt.Sprintf("analytics: connector %q already registered", name))
	}
	connectors[name] = factory
}

// LookupConnector returns the factory for a connector name.
func LookupConnector(name string) (ConnectorFactory, bool) {
	connectorMu.RLock()
	defer connectorMu.RUnlock()
	factory, ok := connectors[name]
	return factory, ok
}

// Connectors returns the sorted names of all registered connectors.
func Connectors() []string {
	connectorMu.RLock()
	defer connectorMu.RUnlock()
	names := make([]string, 0, len(connectors))
	for name := range connectors {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}
