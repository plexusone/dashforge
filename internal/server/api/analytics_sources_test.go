package api

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/plexusone/uiforge/dashboardir"
	serveranalytics "github.com/plexusone/uiforge/internal/server/analytics"
)

type apiFakeProvider struct{ dsn string }

func (p *apiFakeProvider) Catalog(context.Context) (dashboardir.AnalyticsCatalog, error) {
	return dashboardir.AnalyticsCatalog{
		Sources: []dashboardir.AnalyticsSource{{ID: "internal", Name: "Internal"}},
	}, nil
}

func (p *apiFakeProvider) Query(context.Context, dashboardir.AnalyticsQueryRequest) (dashboardir.AnalyticsQueryResult, error) {
	return dashboardir.AnalyticsQueryResult{}, nil
}

func (p *apiFakeProvider) Close() error { return nil }

var registerAPIFakeConnector sync.Once

func newSourcesTestHandler(t *testing.T) *AnalyticsHandler {
	t.Helper()
	registerAPIFakeConnector.Do(func() {
		serveranalytics.RegisterConnector("api-fake", func(dsn string) (serveranalytics.QueryProvider, error) {
			if dsn == "fail" {
				return nil, fmt.Errorf("dial failed")
			}
			return &apiFakeProvider{dsn: dsn}, nil
		})
	})
	store, err := serveranalytics.NewSourceFileStore(filepath.Join(t.TempDir(), "sources.json"))
	if err != nil {
		t.Fatal(err)
	}
	resolver, err := serveranalytics.NewDefaultResolver()
	if err != nil {
		t.Fatal(err)
	}
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	service := serveranalytics.NewManagedService(store, resolver, logger)
	return NewAnalyticsHandler(service, logger, nil)
}

func TestAnalyticsSourcesAPI(t *testing.T) {
	handler := newSourcesTestHandler(t)
	t.Setenv("UIFORGE_TEST_API_DSN", "dsn-value")

	body := `{"id":"demo","name":"Demo","connector":"api-fake","dsnRef":"env://UIFORGE_TEST_API_DSN","enabled":true}`

	// Create.
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequest("POST", "/api/v1/analytics/sources", strings.NewReader(body)))
	if rec.Code != 201 {
		t.Fatalf("create: expected 201, got %d: %s", rec.Code, rec.Body.String())
	}
	if strings.Contains(rec.Body.String(), "dsn-value") {
		t.Fatal("resolved DSN leaked into create response")
	}

	// List shows connected status and the dsnRef (a reference is not a secret).
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequest("GET", "/api/v1/analytics/sources", nil))
	if rec.Code != 200 {
		t.Fatalf("list: expected 200, got %d", rec.Code)
	}
	var listResp struct {
		Sources []serveranalytics.SourceStatus `json:"sources"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &listResp); err != nil {
		t.Fatal(err)
	}
	if len(listResp.Sources) != 1 || listResp.Sources[0].Status != "connected" {
		t.Fatalf("unexpected list response: %+v", listResp)
	}
	if listResp.Sources[0].DSNRef != "env://UIFORGE_TEST_API_DSN" {
		t.Fatalf("expected dsnRef in response, got %+v", listResp.Sources[0])
	}

	// Catalog includes the re-tagged source.
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequest("GET", "/api/v1/analytics/catalog", nil))
	if rec.Code != 200 || !strings.Contains(rec.Body.String(), `"demo"`) {
		t.Fatalf("catalog: expected re-tagged source, got %d: %s", rec.Code, rec.Body.String())
	}

	// Update.
	rec = httptest.NewRecorder()
	updated := strings.Replace(body, `"Demo"`, `"Renamed"`, 1)
	handler.ServeHTTP(rec, httptest.NewRequest("PUT", "/api/v1/analytics/sources/demo", strings.NewReader(updated)))
	if rec.Code != 200 || !strings.Contains(rec.Body.String(), "Renamed") {
		t.Fatalf("update: got %d: %s", rec.Code, rec.Body.String())
	}

	// Update of unknown ID is 404.
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequest("PUT", "/api/v1/analytics/sources/nope", strings.NewReader(body)))
	if rec.Code != 404 {
		t.Fatalf("update unknown: expected 404, got %d", rec.Code)
	}

	// Connectors list.
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequest("GET", "/api/v1/analytics/connectors", nil))
	if rec.Code != 200 || !strings.Contains(rec.Body.String(), "api-fake") {
		t.Fatalf("connectors: got %d: %s", rec.Code, rec.Body.String())
	}

	// Delete.
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequest("DELETE", "/api/v1/analytics/sources/demo", nil))
	if rec.Code != 204 {
		t.Fatalf("delete: expected 204, got %d", rec.Code)
	}
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequest("DELETE", "/api/v1/analytics/sources/demo", nil))
	if rec.Code != 404 {
		t.Fatalf("double delete: expected 404, got %d", rec.Code)
	}
}

func TestAnalyticsSourcesAPITest(t *testing.T) {
	handler := newSourcesTestHandler(t)
	t.Setenv("UIFORGE_TEST_API_DSN", "dsn-value")
	t.Setenv("UIFORGE_TEST_API_FAIL", "fail")

	rec := httptest.NewRecorder()
	ok := `{"id":"probe","name":"Probe","connector":"api-fake","dsnRef":"env://UIFORGE_TEST_API_DSN","enabled":true}`
	handler.ServeHTTP(rec, httptest.NewRequest("POST", "/api/v1/analytics/sources/test", strings.NewReader(ok)))
	if rec.Code != 200 || !strings.Contains(rec.Body.String(), `"ok":true`) {
		t.Fatalf("test ok: got %d: %s", rec.Code, rec.Body.String())
	}

	rec = httptest.NewRecorder()
	bad := `{"id":"probe","name":"Probe","connector":"api-fake","dsnRef":"env://UIFORGE_TEST_API_FAIL","enabled":true}`
	handler.ServeHTTP(rec, httptest.NewRequest("POST", "/api/v1/analytics/sources/test", strings.NewReader(bad)))
	if rec.Code != 200 || !strings.Contains(rec.Body.String(), `"ok":false`) {
		t.Fatalf("test fail: got %d: %s", rec.Code, rec.Body.String())
	}
	// Sanitized: dial errors must not surface driver details or DSNs.
	if strings.Contains(rec.Body.String(), "fail\"") && strings.Contains(rec.Body.String(), "dial failed") {
		t.Fatalf("unsanitized dial error in response: %s", rec.Body.String())
	}

	// Raw DSN rejected outright.
	rec = httptest.NewRecorder()
	raw := `{"id":"probe","name":"Probe","connector":"api-fake","dsnRef":"root:@tcp(127.0.0.1:13307)/db","enabled":true}`
	handler.ServeHTTP(rec, httptest.NewRequest("POST", "/api/v1/analytics/sources/test", strings.NewReader(raw)))
	if rec.Code != 200 || !strings.Contains(rec.Body.String(), `"ok":false`) {
		t.Fatalf("raw dsn: got %d: %s", rec.Code, rec.Body.String())
	}
}
