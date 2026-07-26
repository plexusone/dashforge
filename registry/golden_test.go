package registry

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/plexusone/uiforge/uispec"
)

func TestGoldenPageSpecs(t *testing.T) {
	r, err := NewWithBuiltins()
	if err != nil {
		t.Fatal(err)
	}

	files, err := filepath.Glob("../testdata/pagespecs/*.json")
	if err != nil {
		t.Fatal(err)
	}
	if len(files) == 0 {
		t.Fatal("no golden PageSpec files found in testdata/pagespecs/")
	}

	for _, f := range files {
		name := filepath.Base(f)
		t.Run(name, func(t *testing.T) {
			data, err := os.ReadFile(f)
			if err != nil {
				t.Fatal(err)
			}

			var page uispec.PageSpec
			if err := json.Unmarshal(data, &page); err != nil {
				t.Fatalf("invalid JSON: %v", err)
			}

			if page.APIVersion != uispec.APIVersion {
				t.Errorf("apiVersion = %q, want %q", page.APIVersion, uispec.APIVersion)
			}
			if page.Kind != uispec.KindPage {
				t.Errorf("kind = %q, want %q", page.Kind, uispec.KindPage)
			}
			if page.Metadata.ID == "" {
				t.Error("metadata.id is empty")
			}

			if err := r.ValidatePage(&page); err != nil {
				t.Errorf("validation failed:\n%v", err)
			}
		})
	}
}

func TestGoldenPageSpec_ProfileEnforcement(t *testing.T) {
	r, err := NewWithBuiltins()
	if err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile("../testdata/pagespecs/sales-dashboard.json")
	if err != nil {
		t.Fatal(err)
	}

	var page uispec.PageSpec
	if err := json.Unmarshal(data, &page); err != nil {
		t.Fatal(err)
	}

	if page.Profile != uispec.ProfileDashboard {
		t.Fatalf("expected dashboard profile, got %q", page.Profile)
	}

	if err := r.ValidatePage(&page); err != nil {
		t.Errorf("valid dashboard should pass: %v", err)
	}

	page.Components = append(page.Components, uispec.ComponentInstance{
		ID:   "bad-component",
		Type: "assistant.chat-panel",
	})
	if err := r.ValidatePage(&page); err == nil {
		t.Error("expected error for assistant namespace in dashboard profile")
	}
}

func TestGoldenPageSpec_RoundTrip(t *testing.T) {
	files, err := filepath.Glob("../testdata/pagespecs/*.json")
	if err != nil {
		t.Fatal(err)
	}

	for _, f := range files {
		name := filepath.Base(f)
		t.Run(name, func(t *testing.T) {
			data, err := os.ReadFile(f)
			if err != nil {
				t.Fatal(err)
			}

			var page uispec.PageSpec
			if err := json.Unmarshal(data, &page); err != nil {
				t.Fatal(err)
			}

			out, err := json.Marshal(&page)
			if err != nil {
				t.Fatal(err)
			}

			var page2 uispec.PageSpec
			if err := json.Unmarshal(out, &page2); err != nil {
				t.Fatalf("re-parse failed: %v", err)
			}

			if page2.Metadata.ID != page.Metadata.ID {
				t.Errorf("ID mismatch: %q vs %q", page2.Metadata.ID, page.Metadata.ID)
			}
			if len(page2.Components) != len(page.Components) {
				t.Errorf("component count mismatch: %d vs %d", len(page2.Components), len(page.Components))
			}
		})
	}
}

func TestBuiltinRegistration(t *testing.T) {
	r, err := NewWithBuiltins()
	if err != nil {
		t.Fatal(err)
	}

	coreExpected := []string{"core.card", "core.text", "core.tabs", "core.button", "core.modal"}
	for _, id := range coreExpected {
		if !r.Has(id) {
			t.Errorf("missing core component: %s", id)
		}
	}

	analyticsExpected := []string{
		"analytics.line-chart", "analytics.bar-chart", "analytics.metric",
		"analytics.table", "analytics.filter", "analytics.gauge",
	}
	for _, id := range analyticsExpected {
		if !r.Has(id) {
			t.Errorf("missing analytics component: %s", id)
		}
	}

	assistantExpected := []string{
		"assistant.thread", "assistant.composer", "assistant.thread-list",
		"assistant.tool-call", "assistant.run-status",
	}
	for _, id := range assistantExpected {
		if !r.Has(id) {
			t.Errorf("missing assistant component: %s", id)
		}
	}

	if got := r.Count(); got != 16 {
		t.Errorf("expected 16 builtins, got %d", got)
	}

	core := r.ListByNamespace("core")
	if len(core) != 5 {
		t.Errorf("expected 5 core components, got %d", len(core))
	}
	analytics := r.ListByNamespace("analytics")
	if len(analytics) != 6 {
		t.Errorf("expected 6 analytics components, got %d", len(analytics))
	}
	assistant := r.ListByNamespace("assistant")
	if len(assistant) != 5 {
		t.Errorf("expected 5 assistant components, got %d", len(assistant))
	}
}
