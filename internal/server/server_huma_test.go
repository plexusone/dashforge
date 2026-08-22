package server

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHumaOpenAPIAndQuestionRoutes(t *testing.T) {
	t.Parallel()

	srv, err := New(Config{
		QuestionStorePath: t.TempDir() + "/questions.json",
	}, slog.Default())
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	openAPIResp := httptest.NewRecorder()
	openAPIReq := httptest.NewRequest(http.MethodGet, "/api/openapi.json", nil)
	srv.Router().ServeHTTP(openAPIResp, openAPIReq)
	if openAPIResp.Code != http.StatusOK {
		t.Fatalf("GET /api/openapi.json status = %d body = %s", openAPIResp.Code, openAPIResp.Body.String())
	}

	var spec struct {
		Paths map[string]json.RawMessage `json:"paths"`
	}
	if err := json.Unmarshal(openAPIResp.Body.Bytes(), &spec); err != nil {
		t.Fatalf("openapi response is not JSON: %v", err)
	}
	if _, ok := spec.Paths["/api/v1/questions"]; !ok {
		t.Fatalf("openapi spec missing /api/v1/questions path")
	}
	if _, ok := spec.Paths["/api/v1/questions/{id}"]; !ok {
		t.Fatalf("openapi spec missing /api/v1/questions/{id} path")
	}

	listResp := httptest.NewRecorder()
	listReq := httptest.NewRequest(http.MethodGet, "/api/v1/questions", nil)
	srv.Router().ServeHTTP(listResp, listReq)
	if listResp.Code != http.StatusOK {
		t.Fatalf("GET /api/v1/questions status = %d body = %s", listResp.Code, listResp.Body.String())
	}
	var list struct {
		Questions []any `json:"questions"`
		Total     int   `json:"total"`
	}
	if err := json.Unmarshal(listResp.Body.Bytes(), &list); err != nil {
		t.Fatalf("questions response is not JSON: %v", err)
	}
	if list.Total != 0 || len(list.Questions) != 0 {
		t.Fatalf("empty question store response = %+v", list)
	}

	createBody := []byte(`{
		"id": "question-test",
		"name": "Test question",
		"sourceId": "default",
		"datasetId": "items",
		"dialect": "grokifyql",
		"query": "SELECT name FROM items LIMIT 10",
		"visualization": {"type": "table"}
	}`)
	createResp := httptest.NewRecorder()
	createReq := httptest.NewRequest(http.MethodPost, "/api/v1/questions", bytes.NewReader(createBody))
	createReq.Header.Set("Content-Type", "application/json")
	srv.Router().ServeHTTP(createResp, createReq)
	if createResp.Code != http.StatusOK {
		t.Fatalf("POST /api/v1/questions status = %d body = %s", createResp.Code, createResp.Body.String())
	}

	updateBody := []byte(`{
		"name": "Updated question",
		"sourceId": "default",
		"datasetId": "items",
		"dialect": "grokifyql",
		"query": "SELECT name FROM items LIMIT 25",
		"visualization": {"type": "table"}
	}`)
	updateResp := httptest.NewRecorder()
	updateReq := httptest.NewRequest(http.MethodPut, "/api/v1/questions/question-test", bytes.NewReader(updateBody))
	updateReq.Header.Set("Content-Type", "application/json")
	srv.Router().ServeHTTP(updateResp, updateReq)
	if updateResp.Code != http.StatusOK {
		t.Fatalf("PUT /api/v1/questions/question-test status = %d body = %s", updateResp.Code, updateResp.Body.String())
	}
	var updated struct {
		ID    string `json:"id"`
		Name  string `json:"name"`
		Query string `json:"query"`
	}
	if err := json.Unmarshal(updateResp.Body.Bytes(), &updated); err != nil {
		t.Fatalf("updated question response is not JSON: %v", err)
	}
	if updated.ID != "question-test" || updated.Name != "Updated question" || updated.Query != "SELECT name FROM items LIMIT 25" {
		t.Fatalf("updated question mismatch: %+v", updated)
	}
}
