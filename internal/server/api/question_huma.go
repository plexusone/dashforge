package api

import (
	"context"
	"errors"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/plexusone/dashforge/dashboardir"
)

// RegisterSavedQuestionRoutes registers saved-question endpoints with Huma so
// they are included in the generated OpenAPI document.
func RegisterSavedQuestionRoutes(api huma.API, h *SavedQuestionHandler) {
	if api == nil || h == nil {
		return
	}

	huma.Register(api, huma.Operation{
		OperationID: "list-saved-questions",
		Method:      http.MethodGet,
		Path:        "/api/v1/questions",
		Summary:     "List saved questions",
		Tags:        []string{"Questions"},
		Errors:      []int{http.StatusInternalServerError},
	}, h.humaListQuestions)

	huma.Register(api, huma.Operation{
		OperationID: "get-saved-question",
		Method:      http.MethodGet,
		Path:        "/api/v1/questions/{id}",
		Summary:     "Get a saved question",
		Tags:        []string{"Questions"},
		Errors:      []int{http.StatusNotFound, http.StatusInternalServerError},
	}, h.humaGetQuestion)

	huma.Register(api, huma.Operation{
		OperationID: "create-saved-question",
		Method:      http.MethodPost,
		Path:        "/api/v1/questions",
		Summary:     "Create a saved question",
		Tags:        []string{"Questions"},
		Errors:      []int{http.StatusBadRequest},
	}, h.humaSaveQuestion)

	huma.Register(api, huma.Operation{
		OperationID: "update-saved-question",
		Method:      http.MethodPut,
		Path:        "/api/v1/questions/{id}",
		Summary:     "Update a saved question",
		Tags:        []string{"Questions"},
		Errors:      []int{http.StatusBadRequest},
	}, h.humaUpdateQuestion)

	huma.Register(api, huma.Operation{
		OperationID:   "delete-saved-question",
		Method:        http.MethodDelete,
		Path:          "/api/v1/questions/{id}",
		Summary:       "Delete a saved question",
		Tags:          []string{"Questions"},
		DefaultStatus: http.StatusNoContent,
		Errors:        []int{http.StatusNotFound, http.StatusInternalServerError},
	}, h.humaDeleteQuestion)
}

type listSavedQuestionsInput struct{}

type listSavedQuestionsOutput struct {
	Body struct {
		Questions []dashboardir.SavedQuestion `json:"questions" doc:"Saved questions sorted by most recently updated"`
		Total     int                         `json:"total" example:"3" doc:"Total saved question count"`
	}
}

type getSavedQuestionInput struct {
	ID string `path:"id" doc:"Saved question ID"`
}

type savedQuestionOutput struct {
	Body dashboardir.SavedQuestion
}

type savedQuestionRequest struct {
	ID            string         `json:"id,omitempty" doc:"Optional saved question ID"`
	Name          string         `json:"name" doc:"Saved question name"`
	Description   string         `json:"description,omitempty" doc:"Optional saved question description"`
	SourceID      string         `json:"sourceId" doc:"Analytics source ID"`
	DatasetID     string         `json:"datasetId" doc:"Analytics dataset ID"`
	Dialect       string         `json:"dialect,omitempty" doc:"Query dialect, currently grokifyql"`
	Query         string         `json:"query" doc:"GrokifyQL query text"`
	Visualization map[string]any `json:"visualization,omitempty" doc:"Question visualization settings"`
}

func (r savedQuestionRequest) savedQuestion() dashboardir.SavedQuestion {
	return dashboardir.SavedQuestion{
		ID:            r.ID,
		Name:          r.Name,
		Description:   r.Description,
		SourceID:      r.SourceID,
		DatasetID:     r.DatasetID,
		Dialect:       r.Dialect,
		Query:         r.Query,
		Visualization: r.Visualization,
	}
}

type saveSavedQuestionInput struct {
	Body savedQuestionRequest
}

type updateSavedQuestionInput struct {
	ID   string `path:"id" doc:"Saved question ID"`
	Body savedQuestionRequest
}

type deleteSavedQuestionInput struct {
	ID string `path:"id" doc:"Saved question ID"`
}

type deleteSavedQuestionOutput struct{}

func (h *SavedQuestionHandler) humaListQuestions(context.Context, *listSavedQuestionsInput) (*listSavedQuestionsOutput, error) {
	questions, err := h.store.list()
	if err != nil {
		return nil, huma.Error500InternalServerError("failed to load questions", err)
	}
	out := &listSavedQuestionsOutput{}
	out.Body.Questions = questions
	out.Body.Total = len(questions)
	return out, nil
}

func (h *SavedQuestionHandler) humaGetQuestion(_ context.Context, input *getSavedQuestionInput) (*savedQuestionOutput, error) {
	question, err := h.store.get(input.ID)
	if err != nil {
		if errors.Is(err, errQuestionNotFound) {
			return nil, huma.Error404NotFound("question not found", err)
		}
		return nil, huma.Error500InternalServerError("failed to load question", err)
	}
	return &savedQuestionOutput{Body: question}, nil
}

func (h *SavedQuestionHandler) humaSaveQuestion(ctx context.Context, input *saveSavedQuestionInput) (*savedQuestionOutput, error) {
	question, err := h.store.save(ctx, input.Body.savedQuestion(), h.policyProvider)
	if err != nil {
		return nil, huma.Error400BadRequest(err.Error(), err)
	}
	return &savedQuestionOutput{Body: question}, nil
}

func (h *SavedQuestionHandler) humaUpdateQuestion(ctx context.Context, input *updateSavedQuestionInput) (*savedQuestionOutput, error) {
	questionReq := input.Body
	questionReq.ID = input.ID
	question, err := h.store.save(ctx, questionReq.savedQuestion(), h.policyProvider)
	if err != nil {
		return nil, huma.Error400BadRequest(err.Error(), err)
	}
	return &savedQuestionOutput{Body: question}, nil
}

func (h *SavedQuestionHandler) humaDeleteQuestion(_ context.Context, input *deleteSavedQuestionInput) (*deleteSavedQuestionOutput, error) {
	if err := h.store.delete(input.ID); err != nil {
		if errors.Is(err, errQuestionNotFound) {
			return nil, huma.Error404NotFound("question not found", err)
		}
		return nil, huma.Error500InternalServerError("failed to delete question", err)
	}
	return &deleteSavedQuestionOutput{}, nil
}
