package api

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/grokify/grokifyql"
	"github.com/plexusone/uiforge/dashboardir"
)

// SavedQuestionHandler serves UIForge question metadata. The current
// implementation is file-backed so local server mode can persist questions even
// when the UIForge metadata DB is not configured.
type SavedQuestionHandler struct {
	store          *questionFileStore
	logger         *slog.Logger
	mux            *http.ServeMux
	policyProvider GrokifyQLPolicyProvider
}

// NewSavedQuestionHandler creates a saved-question API handler.
func NewSavedQuestionHandler(path string, logger *slog.Logger, policyProvider GrokifyQLPolicyProvider) (*SavedQuestionHandler, error) {
	store, err := newQuestionFileStore(path)
	if err != nil {
		return nil, err
	}
	if policyProvider == nil {
		policyProvider = staticGrokifyQLPolicyProvider{}
	}
	h := &SavedQuestionHandler{
		store:          store,
		logger:         logger,
		mux:            http.NewServeMux(),
		policyProvider: policyProvider,
	}
	h.setupRoutes()
	return h, nil
}

func (h *SavedQuestionHandler) setupRoutes() {
	h.mux.HandleFunc("GET /api/v1/questions", h.listQuestions)
	h.mux.HandleFunc("GET /api/v1/questions/{id}", h.getQuestion)
	h.mux.HandleFunc("POST /api/v1/questions", h.saveQuestion)
	h.mux.HandleFunc("PUT /api/v1/questions/{id}", h.updateQuestion)
	h.mux.HandleFunc("DELETE /api/v1/questions/{id}", h.deleteQuestion)
}

// ServeHTTP implements http.Handler.
func (h *SavedQuestionHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.mux.ServeHTTP(w, r)
}

func (h *SavedQuestionHandler) listQuestions(w http.ResponseWriter, r *http.Request) {
	questions, err := h.store.list()
	if err != nil {
		h.jsonError(w, http.StatusInternalServerError, "failed to load questions")
		return
	}
	h.jsonResponse(w, http.StatusOK, map[string]any{
		"questions": questions,
		"total":     len(questions),
	})
}

func (h *SavedQuestionHandler) getQuestion(w http.ResponseWriter, r *http.Request) {
	question, err := h.store.get(r.PathValue("id"))
	if err != nil {
		if errors.Is(err, errQuestionNotFound) {
			h.jsonError(w, http.StatusNotFound, "question not found")
			return
		}
		h.jsonError(w, http.StatusInternalServerError, "failed to load question")
		return
	}
	h.jsonResponse(w, http.StatusOK, question)
}

func (h *SavedQuestionHandler) saveQuestion(w http.ResponseWriter, r *http.Request) {
	var req dashboardir.SavedQuestion
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.jsonError(w, http.StatusBadRequest, "invalid request: "+err.Error())
		return
	}
	question, err := h.store.save(r.Context(), req, h.policyProvider)
	if err != nil {
		h.jsonError(w, http.StatusBadRequest, err.Error())
		return
	}
	h.jsonResponse(w, http.StatusOK, question)
}

func (h *SavedQuestionHandler) updateQuestion(w http.ResponseWriter, r *http.Request) {
	var req dashboardir.SavedQuestion
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.jsonError(w, http.StatusBadRequest, "invalid request: "+err.Error())
		return
	}
	req.ID = r.PathValue("id")
	question, err := h.store.save(r.Context(), req, h.policyProvider)
	if err != nil {
		h.jsonError(w, http.StatusBadRequest, err.Error())
		return
	}
	h.jsonResponse(w, http.StatusOK, question)
}

func (h *SavedQuestionHandler) deleteQuestion(w http.ResponseWriter, r *http.Request) {
	if err := h.store.delete(r.PathValue("id")); err != nil {
		if errors.Is(err, errQuestionNotFound) {
			h.jsonError(w, http.StatusNotFound, "question not found")
			return
		}
		h.jsonError(w, http.StatusInternalServerError, "failed to delete question")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *SavedQuestionHandler) jsonResponse(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil && h.logger != nil {
		h.logger.Error("failed to write response", "error", err)
	}
}

func (h *SavedQuestionHandler) jsonError(w http.ResponseWriter, status int, message string) {
	h.jsonResponse(w, status, map[string]string{"error": message})
}

var errQuestionNotFound = errors.New("question not found")

type questionFileStore struct {
	path string
	mu   sync.Mutex
}

func newQuestionFileStore(path string) (*questionFileStore, error) {
	if strings.TrimSpace(path) == "" {
		path = ".uiforge/questions.json"
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, fmt.Errorf("creating question store directory: %w", err)
	}
	return &questionFileStore{path: path}, nil
}

func (s *questionFileStore) list() ([]dashboardir.SavedQuestion, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.readLocked()
}

func (s *questionFileStore) get(id string) (dashboardir.SavedQuestion, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	questions, err := s.readLocked()
	if err != nil {
		return dashboardir.SavedQuestion{}, err
	}
	for _, question := range questions {
		if question.ID == id {
			return question, nil
		}
	}
	return dashboardir.SavedQuestion{}, errQuestionNotFound
}

func (s *questionFileStore) save(ctx context.Context, question dashboardir.SavedQuestion, policyProvider GrokifyQLPolicyProvider) (dashboardir.SavedQuestion, error) {
	if strings.TrimSpace(question.Name) == "" {
		return dashboardir.SavedQuestion{}, fmt.Errorf("name is required")
	}
	if strings.TrimSpace(question.SourceID) == "" {
		return dashboardir.SavedQuestion{}, fmt.Errorf("sourceId is required")
	}
	if strings.TrimSpace(question.DatasetID) == "" {
		return dashboardir.SavedQuestion{}, fmt.Errorf("datasetId is required")
	}
	if strings.TrimSpace(question.Query) == "" {
		return dashboardir.SavedQuestion{}, fmt.Errorf("query is required")
	}
	if question.ID == "" {
		question.ID = fmt.Sprintf("question-%d", time.Now().UnixNano())
	}
	if question.Dialect == "" {
		question.Dialect = "grokifyql"
	}
	if strings.ToLower(strings.TrimSpace(question.Dialect)) != "grokifyql" {
		return dashboardir.SavedQuestion{}, fmt.Errorf("questions only support grokifyql")
	}
	if policyProvider == nil {
		policyProvider = staticGrokifyQLPolicyProvider{}
	}
	policy, err := policyProvider.Policy(ctx, question.SourceID)
	if err != nil {
		return dashboardir.SavedQuestion{}, err
	}
	compiled, err := compileGrokifyQLQuestion(question.Query, policy)
	if err != nil {
		return dashboardir.SavedQuestion{}, err
	}
	if !sameQueryName(compiled.Datasets, question.DatasetID) {
		return dashboardir.SavedQuestion{}, fmt.Errorf("query FROM must match datasetId %q", question.DatasetID)
	}
	question.CompiledQuery = compiled
	now := time.Now().UTC()
	s.mu.Lock()
	defer s.mu.Unlock()
	questions, err := s.readLocked()
	if err != nil {
		return dashboardir.SavedQuestion{}, err
	}
	found := false
	for i := range questions {
		if questions[i].ID != question.ID {
			continue
		}
		question.CreatedAt = questions[i].CreatedAt
		question.UpdatedAt = now
		questions[i] = question
		found = true
		break
	}
	if !found {
		question.CreatedAt = now
		question.UpdatedAt = now
		questions = append([]dashboardir.SavedQuestion{question}, questions...)
	}
	if err := s.writeLocked(questions); err != nil {
		return dashboardir.SavedQuestion{}, err
	}
	return question, nil
}

func sameQueryName(datasets []string, datasetID string) bool {
	if len(datasets) == 0 {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(datasets[0]), strings.TrimSpace(datasetID))
}

func compileGrokifyQLQuestion(input string, policy grokifyql.Policy) (*dashboardir.CompiledQuery, error) {
	q, err := grokifyql.Parse(input)
	if err != nil {
		return nil, err
	}
	issues := grokifyql.CheckPolicy(q, policy)
	if len(issues) > 0 {
		return nil, fmt.Errorf("invalid query policy: %s", issues[0].Message)
	}
	ast, err := json.Marshal(q)
	if err != nil {
		return nil, err
	}
	sum := sha256.Sum256(ast)
	fields := compiledFields(q)
	return &dashboardir.CompiledQuery{
		Version:     "grokifyql/v1",
		AST:         ast,
		Fingerprint: "sha256:" + hex.EncodeToString(sum[:]),
		Datasets:    []string{q.From},
		Fields:      fields,
		ReadOnly:    true,
		Limit:       q.Limit,
	}, nil
}

func compiledFields(q *grokifyql.Query) []string {
	if q == nil {
		return nil
	}
	seen := map[string]struct{}{}
	var fields []string
	add := func(field string) {
		field = strings.TrimSpace(strings.ToLower(field))
		if field == "" {
			return
		}
		if _, ok := seen[field]; ok {
			return
		}
		seen[field] = struct{}{}
		fields = append(fields, field)
	}
	for _, item := range q.Select {
		if item.Aggregate != nil {
			if !item.Aggregate.Star {
				add(item.Aggregate.Field)
			}
			add(grokifyql.OutputName(item))
			continue
		}
		if !item.Star {
			add(item.Field)
		}
	}
	for _, field := range q.GroupBy {
		add(field)
	}
	for _, order := range q.OrderBy {
		add(order.Field)
	}
	collectExprFields(q.Where, add)
	sort.Strings(fields)
	return fields
}

func collectExprFields(expr grokifyql.Expr, add func(string)) {
	switch e := expr.(type) {
	case nil:
	case *grokifyql.CompareExpr:
		add(e.Field)
	case *grokifyql.LogicalExpr:
		collectExprFields(e.Left, add)
		collectExprFields(e.Right, add)
	case *grokifyql.NotExpr:
		collectExprFields(e.Expr, add)
	}
}

func (s *questionFileStore) delete(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	questions, err := s.readLocked()
	if err != nil {
		return err
	}
	next := questions[:0]
	found := false
	for _, question := range questions {
		if question.ID == id {
			found = true
			continue
		}
		next = append(next, question)
	}
	if !found {
		return errQuestionNotFound
	}
	return s.writeLocked(next)
}

func (s *questionFileStore) readLocked() ([]dashboardir.SavedQuestion, error) {
	data, err := os.ReadFile(s.path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return []dashboardir.SavedQuestion{}, nil
		}
		return nil, err
	}
	var questions []dashboardir.SavedQuestion
	if err := json.Unmarshal(data, &questions); err != nil {
		return nil, err
	}
	sort.SliceStable(questions, func(i, j int) bool {
		return questions[i].UpdatedAt.After(questions[j].UpdatedAt)
	})
	return questions, nil
}

func (s *questionFileStore) writeLocked(questions []dashboardir.SavedQuestion) error {
	data, err := json.MarshalIndent(questions, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.path, append(data, '\n'), 0o644)
}
