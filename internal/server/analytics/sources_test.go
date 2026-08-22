package analytics

import (
	"strings"
	"testing"
)

func validSourceConfig() SourceConfig {
	return SourceConfig{
		ID:        "omniroadmap-local",
		Name:      "OmniRoadmap Local",
		Connector: ConnectorOmniRoadmap,
		DSNRef:    "env://UIFORGE_OMNIROADMAP_DSN",
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
	if _, ok := LookupConnector(ConnectorOmniRoadmap); !ok {
		t.Fatal("omniroadmap connector not registered")
	}
	if _, ok := LookupConnector("does-not-exist"); ok {
		t.Fatal("unexpected connector registration")
	}
	found := false
	for _, name := range Connectors() {
		if name == ConnectorOmniRoadmap {
			found = true
		}
	}
	if !found {
		t.Fatalf("Connectors() = %v, missing %q", Connectors(), ConnectorOmniRoadmap)
	}
}

func TestRegisterConnectorPanics(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("expected panic on duplicate registration")
		}
	}()
	RegisterConnector(ConnectorOmniRoadmap, func(string) (QueryProvider, error) { return nil, nil })
}
