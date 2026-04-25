// Package api provides REST API handlers for Dashforge.
package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"

	omnillm "github.com/plexusone/omnillm-core"
	"github.com/plexusone/omnillm-core/models"
	"github.com/plexusone/omnillm-core/provider"
)

// AIHandler handles AI generation requests using omnillm.
type AIHandler struct {
	client *omnillm.ChatClient
	logger *slog.Logger
}

// AIConfig holds configuration for the AI handler.
type AIConfig struct {
	// Provider API keys (read from environment if not set)
	AnthropicAPIKey string
	OpenAIAPIKey    string
	GeminiAPIKey    string
	GrokAPIKey      string

	// Default model to use
	DefaultModel string

	// Enable fallback to secondary providers
	EnableFallback bool

	// Request timeout
	Timeout time.Duration
}

// NewAIHandler creates a new AI handler with the given configuration.
func NewAIHandler(cfg AIConfig, logger *slog.Logger) (*AIHandler, error) {
	if logger == nil {
		logger = slog.Default()
	}

	// Build provider list from available API keys
	var providers []omnillm.ProviderConfig

	// Try to get API keys from config or environment
	anthropicKey := cfg.AnthropicAPIKey
	if anthropicKey == "" {
		anthropicKey = os.Getenv("ANTHROPIC_API_KEY")
	}
	if anthropicKey != "" {
		providers = append(providers, omnillm.ProviderConfig{
			Provider: omnillm.ProviderNameAnthropic,
			APIKey:   anthropicKey,
		})
		logger.Info("AI provider configured", "provider", "anthropic")
	}

	openaiKey := cfg.OpenAIAPIKey
	if openaiKey == "" {
		openaiKey = os.Getenv("OPENAI_API_KEY")
	}
	if openaiKey != "" {
		providers = append(providers, omnillm.ProviderConfig{
			Provider: omnillm.ProviderNameOpenAI,
			APIKey:   openaiKey,
		})
		logger.Info("AI provider configured", "provider", "openai")
	}

	geminiKey := cfg.GeminiAPIKey
	if geminiKey == "" {
		geminiKey = os.Getenv("GEMINI_API_KEY")
	}
	if geminiKey != "" {
		providers = append(providers, omnillm.ProviderConfig{
			Provider: omnillm.ProviderNameGemini,
			APIKey:   geminiKey,
		})
		logger.Info("AI provider configured", "provider", "gemini")
	}

	grokKey := cfg.GrokAPIKey
	if grokKey == "" {
		grokKey = os.Getenv("GROK_API_KEY")
	}
	if grokKey != "" {
		providers = append(providers, omnillm.ProviderConfig{
			Provider: omnillm.ProviderNameXAI,
			APIKey:   grokKey,
		})
		logger.Info("AI provider configured", "provider", "xai")
	}

	if len(providers) == 0 {
		logger.Warn("no AI providers configured - AI features will be disabled")
		return &AIHandler{logger: logger}, nil
	}

	// Create omnillm client
	clientConfig := omnillm.ClientConfig{
		Providers: providers,
	}

	// Enable circuit breaker for fallback
	if cfg.EnableFallback && len(providers) > 1 {
		clientConfig.CircuitBreakerConfig = &omnillm.CircuitBreakerConfig{
			FailureThreshold: 3,
			Timeout:          30 * time.Second,
		}
	}

	client, err := omnillm.NewClient(clientConfig)
	if err != nil {
		return nil, fmt.Errorf("creating AI client: %w", err)
	}

	return &AIHandler{
		client: client,
		logger: logger,
	}, nil
}

// ServeHTTP implements http.Handler.
func (h *AIHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Route based on path
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/ai")

	switch {
	case path == "/generate" && r.Method == http.MethodPost:
		h.handleGenerate(w, r)
	case path == "/stream" && r.Method == http.MethodPost:
		h.handleStream(w, r)
	case path == "/models" && r.Method == http.MethodGet:
		h.handleListModels(w, r)
	case path == "/status" && r.Method == http.MethodGet:
		h.handleStatus(w, r)
	default:
		http.NotFound(w, r)
	}
}

// GenerateRequest is the request body for AI generation.
type GenerateRequest struct {
	// Prompt is the user's request
	Prompt string `json:"prompt"`

	// SystemPrompt overrides the default system prompt
	SystemPrompt string `json:"systemPrompt,omitempty"`

	// Type is what to generate: "dashboard", "widget", "query", "modify"
	Type string `json:"type,omitempty"`

	// Context provides additional context (existing widgets, schema, etc.)
	Context *GenerateContext `json:"context,omitempty"`

	// Model to use (optional, uses default if not specified)
	Model string `json:"model,omitempty"`

	// Temperature for generation (0.0-1.0)
	Temperature float64 `json:"temperature,omitempty"`

	// MaxTokens limits response length
	MaxTokens int `json:"maxTokens,omitempty"`
}

// GenerateContext provides context for generation.
type GenerateContext struct {
	// ExistingWidgets for positioning new widgets
	ExistingWidgets []WidgetPosition `json:"existingWidgets,omitempty"`

	// Schema information from Cube.js
	Schema *SchemaContext `json:"schema,omitempty"`

	// CurrentWidget for modification requests
	CurrentWidget json.RawMessage `json:"currentWidget,omitempty"`
}

// WidgetPosition describes a widget's position.
type WidgetPosition struct {
	ID       string `json:"id"`
	Type     string `json:"type"`
	Title    string `json:"title"`
	Position struct {
		X int `json:"x"`
		Y int `json:"y"`
		W int `json:"w"`
		H int `json:"h"`
	} `json:"position"`
}

// SchemaContext provides Cube.js schema context.
type SchemaContext struct {
	Cubes []CubeInfo `json:"cubes"`
}

// CubeInfo describes a Cube.js cube.
type CubeInfo struct {
	Name        string       `json:"name"`
	Title       string       `json:"title"`
	Description string       `json:"description,omitempty"`
	Measures    []MemberInfo `json:"measures"`
	Dimensions  []MemberInfo `json:"dimensions"`
}

// MemberInfo describes a measure or dimension.
type MemberInfo struct {
	Name        string `json:"name"`
	Title       string `json:"title"`
	Type        string `json:"type"`
	Description string `json:"description,omitempty"`
}

// GenerateResponse is the response from AI generation.
type GenerateResponse struct {
	Success  bool            `json:"success"`
	Data     json.RawMessage `json:"data,omitempty"`
	Text     string          `json:"text,omitempty"`
	Error    string          `json:"error,omitempty"`
	Warnings []string        `json:"warnings,omitempty"`
	Usage    *UsageInfo      `json:"usage,omitempty"`
	Model    string          `json:"model,omitempty"`
	Provider string          `json:"provider,omitempty"`
}

// UsageInfo contains token usage information.
type UsageInfo struct {
	PromptTokens     int `json:"promptTokens"`
	CompletionTokens int `json:"completionTokens"`
	TotalTokens      int `json:"totalTokens"`
}

func (h *AIHandler) handleGenerate(w http.ResponseWriter, r *http.Request) {
	if h.client == nil {
		writeJSON(w, http.StatusServiceUnavailable, GenerateResponse{
			Success: false,
			Error:   "AI service not configured. Set ANTHROPIC_API_KEY, OPENAI_API_KEY, or another provider's API key.",
		})
		return
	}

	var req GenerateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, GenerateResponse{
			Success: false,
			Error:   "Invalid request body: " + err.Error(),
		})
		return
	}

	if req.Prompt == "" {
		writeJSON(w, http.StatusBadRequest, GenerateResponse{
			Success: false,
			Error:   "Prompt is required",
		})
		return
	}

	// Build system prompt based on request type
	systemPrompt := req.SystemPrompt
	if systemPrompt == "" {
		systemPrompt = h.buildSystemPrompt(req.Type, req.Context)
	}

	// Build messages
	messages := []provider.Message{
		{Role: provider.RoleSystem, Content: systemPrompt},
		{Role: provider.RoleUser, Content: req.Prompt},
	}

	// Determine model
	model := req.Model
	if model == "" {
		model = models.Claude3_7Sonnet // Default to Claude 3.7 Sonnet
	}

	// Set defaults
	temperature := req.Temperature
	if temperature == 0 {
		temperature = 0.7
	}
	maxTokens := req.MaxTokens
	if maxTokens == 0 {
		maxTokens = 4096
	}

	// Create completion request
	completionReq := &provider.ChatCompletionRequest{
		Model:       model,
		Messages:    messages,
		Temperature: &temperature,
		MaxTokens:   &maxTokens,
	}

	// Execute request with timeout
	ctx, cancel := context.WithTimeout(r.Context(), 60*time.Second)
	defer cancel()

	resp, err := h.client.CreateChatCompletion(ctx, completionReq)
	if err != nil {
		h.logger.Error("AI generation failed", "error", err)
		writeJSON(w, http.StatusInternalServerError, GenerateResponse{
			Success: false,
			Error:   "AI generation failed: " + err.Error(),
		})
		return
	}

	if len(resp.Choices) == 0 {
		writeJSON(w, http.StatusInternalServerError, GenerateResponse{
			Success: false,
			Error:   "No response from AI",
		})
		return
	}

	content := resp.Choices[0].Message.Content

	// Try to parse as JSON
	var jsonData json.RawMessage
	if err := json.Unmarshal([]byte(content), &jsonData); err != nil {
		// Try to extract JSON from markdown code blocks
		jsonData = extractJSON(content)
	}

	response := GenerateResponse{
		Success: true,
		Text:    content,
		Model:   resp.Model,
	}

	if jsonData != nil {
		response.Data = jsonData
	}

	response.Usage = &UsageInfo{
		PromptTokens:     resp.Usage.PromptTokens,
		CompletionTokens: resp.Usage.CompletionTokens,
		TotalTokens:      resp.Usage.TotalTokens,
	}

	writeJSON(w, http.StatusOK, response)
}

func (h *AIHandler) handleStream(w http.ResponseWriter, r *http.Request) {
	if h.client == nil {
		writeJSON(w, http.StatusServiceUnavailable, GenerateResponse{
			Success: false,
			Error:   "AI service not configured",
		})
		return
	}

	var req GenerateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, GenerateResponse{
			Success: false,
			Error:   "Invalid request body: " + err.Error(),
		})
		return
	}

	if req.Prompt == "" {
		writeJSON(w, http.StatusBadRequest, GenerateResponse{
			Success: false,
			Error:   "Prompt is required",
		})
		return
	}

	// Set up SSE headers
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming not supported", http.StatusInternalServerError)
		return
	}

	// Build system prompt
	systemPrompt := req.SystemPrompt
	if systemPrompt == "" {
		systemPrompt = h.buildSystemPrompt(req.Type, req.Context)
	}

	messages := []provider.Message{
		{Role: provider.RoleSystem, Content: systemPrompt},
		{Role: provider.RoleUser, Content: req.Prompt},
	}

	model := req.Model
	if model == "" {
		model = models.Claude3_7Sonnet
	}

	temperature := req.Temperature
	if temperature == 0 {
		temperature = 0.7
	}
	maxTokens := req.MaxTokens
	if maxTokens == 0 {
		maxTokens = 4096
	}

	completionReq := &provider.ChatCompletionRequest{
		Model:       model,
		Messages:    messages,
		Temperature: &temperature,
		MaxTokens:   &maxTokens,
	}

	ctx, cancel := context.WithTimeout(r.Context(), 120*time.Second)
	defer cancel()

	stream, err := h.client.CreateChatCompletionStream(ctx, completionReq)
	if err != nil {
		fmt.Fprintf(w, "data: {\"error\": %q}\n\n", err.Error())
		flusher.Flush()
		return
	}
	defer stream.Close()

	for {
		chunk, err := stream.Recv()
		if err == io.EOF {
			fmt.Fprintf(w, "data: [DONE]\n\n")
			flusher.Flush()
			break
		}
		if err != nil {
			fmt.Fprintf(w, "data: {\"error\": %q}\n\n", err.Error())
			flusher.Flush()
			break
		}

		if len(chunk.Choices) > 0 && chunk.Choices[0].Delta != nil && chunk.Choices[0].Delta.Content != "" {
			data, _ := json.Marshal(map[string]string{
				"content": chunk.Choices[0].Delta.Content,
			})
			fmt.Fprintf(w, "data: %s\n\n", data)
			flusher.Flush()
		}
	}
}

func (h *AIHandler) handleListModels(w http.ResponseWriter, r *http.Request) {
	// Return available models grouped by provider
	modelList := map[string][]string{
		"anthropic": {
			models.ClaudeOpus4_1,
			models.ClaudeOpus4,
			models.ClaudeSonnet4,
			models.Claude3_7Sonnet,
			models.Claude3_5Haiku,
		},
		"openai": {
			models.GPT5,
			models.GPT4_1,
			models.GPT4o,
			models.GPT4oMini,
			models.GPT4Turbo,
		},
		"gemini": {
			models.Gemini2_5Pro,
			models.Gemini2_5Flash,
			models.Gemini1_5Pro,
			models.Gemini1_5Flash,
		},
		"xai": {
			models.Grok4_1FastReasoning,
			models.Grok4_0709,
			models.GrokCodeFast1,
		},
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"models":  modelList,
		"default": models.Claude3_7Sonnet,
	})
}

func (h *AIHandler) handleStatus(w http.ResponseWriter, r *http.Request) {
	status := map[string]any{
		"enabled": h.client != nil,
	}

	if h.client != nil {
		status["providers"] = []string{} // Would need to expose this from omnillm
	}

	writeJSON(w, http.StatusOK, status)
}

func (h *AIHandler) buildSystemPrompt(reqType string, ctx *GenerateContext) string {
	basePrompt := `You are a dashboard design assistant. You help users create data dashboards by generating JSON configurations.

You output JSON that follows the DashboardIR specification. Key rules:
- Use a 12-column grid layout
- Position widgets using x, y, w, h coordinates (x and y start at 0)
- Common widget sizes: metrics (2x2), charts (4x3 or 6x3), tables (6x4)
- Align widgets to avoid overlap
- Use descriptive titles for widgets
- Connect widgets to appropriate data sources

Chart types available: line, bar, pie, scatter, area
Widget types available: chart, metric, table, text, image

When creating charts:
- For time series: use line or area charts with date on x-axis
- For comparisons: use bar charts
- For distributions: use pie charts
- For correlations: use scatter plots

Always respond with valid JSON only. No explanations or markdown code blocks.`

	var contextInfo strings.Builder

	if ctx != nil {
		// Add existing widgets context
		if len(ctx.ExistingWidgets) > 0 {
			contextInfo.WriteString("\n\nExisting widgets in dashboard (avoid overlapping):\n")
			for _, w := range ctx.ExistingWidgets {
				fmt.Fprintf(&contextInfo, "- %s (%s): x=%d, y=%d, w=%d, h=%d\n",
					w.Title, w.Type, w.Position.X, w.Position.Y, w.Position.W, w.Position.H)
			}
		}

		// Add schema context
		if ctx.Schema != nil && len(ctx.Schema.Cubes) > 0 {
			contextInfo.WriteString("\n\nAvailable data (Cube.js schema):\n")
			for _, cube := range ctx.Schema.Cubes {
				fmt.Fprintf(&contextInfo, "\n%s (%s):\n", cube.Name, cube.Title)
				if len(cube.Measures) > 0 {
					contextInfo.WriteString("  Measures: ")
					var measureNames []string
					for _, m := range cube.Measures {
						measureNames = append(measureNames, m.Name)
					}
					contextInfo.WriteString(strings.Join(measureNames, ", "))
					contextInfo.WriteString("\n")
				}
				if len(cube.Dimensions) > 0 {
					contextInfo.WriteString("  Dimensions: ")
					var dimNames []string
					for _, d := range cube.Dimensions {
						dimNames = append(dimNames, d.Name)
					}
					contextInfo.WriteString(strings.Join(dimNames, ", "))
					contextInfo.WriteString("\n")
				}
			}
		}

		// Add current widget for modification
		if ctx.CurrentWidget != nil {
			contextInfo.WriteString("\n\nCurrent widget to modify:\n")
			contextInfo.Write(ctx.CurrentWidget)
			contextInfo.WriteString("\n")
		}
	}

	// Add type-specific instructions
	var typeInstructions string
	switch reqType {
	case "dashboard":
		typeInstructions = `

Generate a complete dashboard JSON with this structure:
{
  "title": "Dashboard Title",
  "layout": { "type": "grid", "columns": 12, "rowHeight": 80 },
  "widgets": [
    {
      "id": "unique-id",
      "type": "chart|metric|table|text",
      "title": "Widget Title",
      "position": { "x": 0, "y": 0, "w": 4, "h": 3 },
      "config": { ... }
    }
  ]
}`

	case "widget":
		typeInstructions = `

Generate a single widget JSON:
{
  "type": "chart|metric|table|text",
  "title": "Widget Title",
  "position": { "x": 0, "y": 0, "w": 4, "h": 3 },
  "config": {
    "geometry": "line|bar|pie|scatter|area",
    "encodings": { "x": "field", "y": "field" },
    "style": { "showLegend": true }
  }
}`

	case "query":
		typeInstructions = `

Generate a Cube.js query JSON:
{
  "measures": ["CubeName.measureName"],
  "dimensions": ["CubeName.dimensionName"],
  "timeDimensions": [{ "dimension": "CubeName.date", "granularity": "month" }],
  "filters": [{ "member": "CubeName.field", "operator": "equals", "values": ["value"] }],
  "order": { "CubeName.measure": "desc" },
  "limit": 100
}`

	case "modify":
		typeInstructions = `

Modify the provided widget based on the user's request. Return the complete modified widget JSON.
Only change what the user requested, keep other settings the same.`
	}

	return basePrompt + contextInfo.String() + typeInstructions
}

// extractJSON attempts to extract JSON from markdown code blocks.
func extractJSON(content string) json.RawMessage {
	// Try to find JSON in code blocks
	start := strings.Index(content, "```json")
	if start == -1 {
		start = strings.Index(content, "```")
	}
	if start != -1 {
		// Find the start of actual JSON
		start = strings.Index(content[start:], "\n") + start + 1
		end := strings.Index(content[start:], "```")
		if end != -1 {
			jsonStr := strings.TrimSpace(content[start : start+end])
			var data json.RawMessage
			if json.Unmarshal([]byte(jsonStr), &data) == nil {
				return data
			}
		}
	}

	// Try to find JSON object directly
	start = strings.Index(content, "{")
	if start != -1 {
		// Find matching closing brace
		depth := 0
		for i := start; i < len(content); i++ {
			switch content[i] {
			case '{':
				depth++
			case '}':
				depth--
				if depth == 0 {
					jsonStr := content[start : i+1]
					var data json.RawMessage
					if json.Unmarshal([]byte(jsonStr), &data) == nil {
						return data
					}
					break
				}
			}
		}
	}

	return nil
}

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}
