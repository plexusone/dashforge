package analytics

import (
	"strings"
	"testing"
)

// testConnector is a stand-in connector registered for tests in this package.
// The engine core no longer bundles any specific connector (OmniRoadmap lives
// in connectors/omniroadmap), so the contract tests register their own.
const testConnector = "test-connector"

func init() {
	RegisterConnector(testConnector, func(string) (QueryProvider, error) { return nil, nil })
}

func validSourceConfig() SourceConfig {
	return SourceConfig{
		ID:        "omniroadmap-local",
		Name:      "OmniRoadmap Local",
		Connector: testConnector,
		DSNRef:    "env://DASHFORGE_OMNIROADMAP_DSN",
		Enabled:   true,
	}
}

func TestSourceConfigValidate(t *testing.T) {
	if err := validSourceConfig().Validate(); err != nil {
		t.Fatalf("valid config rejected: %v", err)
	}

	tests := []struct {
		name    string
		mutate  func(*SourceConfig)
		wantSub string
	}{
		{"missing id", func(c *SourceConfig) { c.ID = "" }, "id is required"},
		{"bad slug", func(c *SourceConfig) { c.ID = "Bad_Slug!" }, "lowercase slug"},
		{"missing name", func(c *SourceConfig) { c.Name = " " }, "name is required"},
		{"missing connector", func(c *SourceConfig) { c.Connector = "" }, "connector is required"},
		{"unknown connector", func(c *SourceConfig) { c.Connector = "nope" }, "unknown connector"},
		{"missing dsnRef", func(c *SourceConfig) { c.DSNRef = "" }, "dsnRef is required"},
		{"raw dsn rejected", func(c *SourceConfig) { c.DSNRef = "root:@tcp(127.0.0.1:13307)/omniroadmap" }, "secret reference"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := validSourceConfig()
			tt.mutate(&cfg)
			err := cfg.Validate()
			if err == nil {
				t.Fatalf("expected error containing %q, got nil", tt.wantSub)
			}
			if !strings.Contains(err.Error(), tt.wantSub) {
				t.Fatalf("expected error containing %q, got %q", tt.wantSub, err.Error())
			}
		})
	}
}

func TestConnectorRegistry(t *testing.T) {
	if _, ok := LookupConnector(testConnector); !ok {
		t.Fatal("omniroadmap connector not registered")
	}
	if _, ok := LookupConnector("does-not-exist"); ok {
		t.Fatal("unexpected connector registration")
	}
	found := false
	for _, name := range Connectors() {
		if name == testConnector {
			found = true
		}
	}
	if !found {
		t.Fatalf("Connectors() = %v, missing %q", Connectors(), testConnector)
	}
}

func TestRegisterConnectorPanics(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("expected panic on duplicate registration")
		}
	}()
	RegisterConnector(testConnector, func(string) (QueryProvider, error) { return nil, nil })
}
