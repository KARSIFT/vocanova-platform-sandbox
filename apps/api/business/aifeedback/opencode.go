package aifeedback

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// OpenCodeConfig configures the OpenCode production adapter (VOC-028-D02).
// BaseURL and APIKey must be supplied from configuration/secrets at runtime;
// they are never hard-coded or committed to source. Model is a
// "providerID/modelID" pair (e.g. "opencode-go/deepseek-v4-pro"), matching
// this project's own AI-role configuration convention (config/roles.yml).
type OpenCodeConfig struct {
	BaseURL    string
	APIKey     string
	Model      string
	Timeout    time.Duration
	MaxRetries int
}

// DefaultOpenCodeConfig returns the default adapter configuration. Production
// callers must override BaseURL and APIKey from backend-only secrets.
func DefaultOpenCodeConfig() OpenCodeConfig {
	return OpenCodeConfig{
		BaseURL:    "http://127.0.0.1:4096",
		Model:      DefaultOpenCodeModel,
		Timeout:    8 * time.Second,
		MaxRetries: 1,
	}
}

// OpenCodeFeedbackProvider implements FeedbackProvider against a real
// `opencode serve` instance (VOC-028-D02: the founder's own OpenCode Go
// account). `opencode serve` exposes its own session/message HTTP API - not
// an OpenAI-compatible completions endpoint - confirmed live against its own
// committed OpenAPI document at GET /doc (there is no /v1/chat/completions
// route). One session is created per feedback request (POST /session) and
// one prompt is sent to it (POST /session/{id}/message); despite that
// endpoint's own summary saying "streaming the AI response", its 200
// response is the complete assistant message, not an SSE stream, so this
// stays a single synchronous HTTP round trip per DOC-09 §18's 8s provider
// timeout. Provider SDK/network types are confined to this adapter.
type OpenCodeFeedbackProvider struct {
	config     OpenCodeConfig
	client     *http.Client
	providerID string
	modelID    string
}

// NewOpenCodeFeedbackProvider creates a production adapter. The API key is kept
// in the config struct and used only in request headers; it is never logged.
func NewOpenCodeFeedbackProvider(config OpenCodeConfig) *OpenCodeFeedbackProvider {
	if config.Model == "" {
		config.Model = DefaultOpenCodeModel
	}
	if config.Timeout <= 0 {
		config.Timeout = 8 * time.Second
	}
	providerID, modelID := splitOpenCodeModel(config.Model)
	return &OpenCodeFeedbackProvider{
		config:     config,
		client:     &http.Client{Timeout: config.Timeout},
		providerID: providerID,
		modelID:    modelID,
	}
}

// splitOpenCodeModel splits a "providerID/modelID" string. A model string
// with no slash is used as both (a defensive fallback, not the expected
// production shape).
func splitOpenCodeModel(model string) (providerID, modelID string) {
	if idx := strings.Index(model, "/"); idx >= 0 {
		return model[:idx], model[idx+1:]
	}
	return model, model
}

// GenerateFeedback creates a fresh OpenCode session and sends one prompt
// message to it. It makes at most one transport retry for clearly transient
// failures (DOC-09 §18); a failed session creation is not retried
// separately (kept as one logical attempt with the message send).
func (p *OpenCodeFeedbackProvider) GenerateFeedback(ctx context.Context, task ProviderTask) (*ProviderFeedback, error) {
	body, err := p.buildMessageRequestBody(task)
	if err != nil {
		return nil, fmt.Errorf("build request body: %w", err)
	}

	var lastErr error
	maxAttempts := p.config.MaxRetries + 1
	if maxAttempts < 1 {
		maxAttempts = 1
	}

	for attempt := 0; attempt < maxAttempts; attempt++ {
		feedback, retryable, err := p.attempt(ctx, body)
		if err == nil {
			return feedback, nil
		}
		lastErr = err
		if attempt < maxAttempts-1 && retryable {
			continue
		}
		return nil, lastErr
	}

	return nil, lastErr
}

func (p *OpenCodeFeedbackProvider) attempt(ctx context.Context, messageBody []byte) (*ProviderFeedback, bool, error) {
	sessionID, err := p.createSession(ctx)
	if err != nil {
		return nil, errors.Is(err, ErrProviderTimeout), err
	}

	url := strings.TrimRight(p.config.BaseURL, "/") + "/session/" + sessionID + "/message"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(messageBody))
	if err != nil {
		return nil, false, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if p.config.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+p.config.APIKey)
	}

	resp, err := p.client.Do(req)
	if err != nil {
		mapped := mapNetworkError(err)
		return nil, isRetryableError(mapped), mapped
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return nil, isRetryableHTTPStatus(resp.StatusCode), fmt.Errorf("read response body: %w", err)
	}

	feedback, err := p.parseMessageResponse(resp.StatusCode, respBody)
	if err != nil {
		return nil, isRetryableHTTPStatus(resp.StatusCode), err
	}
	return feedback, false, nil
}

func (p *OpenCodeFeedbackProvider) createSession(ctx context.Context) (string, error) {
	url := strings.TrimRight(p.config.BaseURL, "/") + "/session"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader([]byte("{}")))
	if err != nil {
		return "", fmt.Errorf("create session request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if p.config.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+p.config.APIKey)
	}

	resp, err := p.client.Do(req)
	if err != nil {
		return "", mapNetworkError(err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<16))
	if err != nil {
		return "", fmt.Errorf("read session response: %w", err)
	}
	if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
		return "", ErrProviderAuth
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("%w: session create status %d", ErrProviderInvalidResponse, resp.StatusCode)
	}

	var session struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(body, &session); err != nil || session.ID == "" {
		return "", fmt.Errorf("%w: invalid session response", ErrProviderInvalidResponse)
	}
	return session.ID, nil
}

func (p *OpenCodeFeedbackProvider) buildMessageRequestBody(task ProviderTask) ([]byte, error) {
	payloadJSON, err := json.Marshal(task.UserPayload)
	if err != nil {
		return nil, fmt.Errorf("marshal user payload: %w", err)
	}

	var text strings.Builder
	text.WriteString(task.DeveloperPrompt)
	text.WriteString("\n\nTask data (JSON):\n")
	text.Write(payloadJSON)
	if task.OutputSchema != nil {
		schemaJSON, err := json.Marshal(task.OutputSchema)
		if err == nil {
			text.WriteString("\n\nRespond with a single JSON object matching this schema exactly, and nothing else:\n")
			text.Write(schemaJSON)
		}
	}

	req := openCodeMessageRequest{
		System: task.SystemPrompt,
		Model: &openCodeModelRef{
			ProviderID: p.providerID,
			ModelID:    p.modelID,
		},
		Parts: []openCodePart{
			{Type: "text", Text: text.String()},
		},
	}

	return json.Marshal(req)
}

func (p *OpenCodeFeedbackProvider) parseMessageResponse(statusCode int, body []byte) (*ProviderFeedback, error) {
	if statusCode == http.StatusUnauthorized || statusCode == http.StatusForbidden {
		return nil, ErrProviderAuth
	}
	if statusCode == http.StatusBadRequest {
		return nil, ErrProviderInvalidInput
	}
	if statusCode >= 500 || statusCode == http.StatusTooManyRequests {
		return nil, ErrProviderTimeout
	}
	if statusCode != http.StatusOK {
		return nil, fmt.Errorf("%w: unexpected status %d", ErrProviderInvalidResponse, statusCode)
	}

	var resp openCodeMessageResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrProviderInvalidResponse, err)
	}

	var textBuilder strings.Builder
	for _, part := range resp.Parts {
		if part.Type == "text" && part.Text != "" {
			textBuilder.WriteString(part.Text)
		}
	}
	content := strings.TrimSpace(textBuilder.String())
	if content == "" {
		return nil, ErrProviderInvalidResponse
	}

	if isProviderRefusal(content) {
		return nil, ErrProviderRefusal
	}

	content = extractJSON(content)
	if content == "" {
		return nil, ErrProviderInvalidResponse
	}

	var raw map[string]any
	if err := json.Unmarshal([]byte(content), &raw); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrProviderInvalidResponse, err)
	}

	return mapOpenCodeFeedback(raw)
}

func mapOpenCodeFeedback(raw map[string]any) (*ProviderFeedback, error) {
	status, _ := raw["status"].(string)
	if status == "" {
		return nil, fmt.Errorf("%w: missing status", ErrProviderInvalidResponse)
	}

	fb := ProviderFeedback{
		Status:  status,
		RawJSON: raw,
	}

	if v, ok := raw["target_word_used_correctly"].(bool); ok {
		fb.TargetWordUsedCorrectly = v
	}
	if v, ok := raw["explanation"].(string); ok {
		fb.Explanation = v
	}
	if v, ok := raw["corrected_sentence"].(string); ok && v != "" {
		fb.CorrectedSentence = &v
	}
	if v, ok := raw["improvement_tip"].(string); ok && v != "" {
		fb.ImprovementTip = &v
	}

	return &fb, nil
}

func isProviderRefusal(content string) bool {
	lower := strings.ToLower(content)
	phrases := []string{
		"i cannot", "i can't", "i'm sorry", "i am sorry", "i'm not able",
		"i am not able", "i cannot fulfill", "i can't fulfill", "i refuse",
		"as an ai", "i'm unable to", "i am unable to", "content policy",
	}
	for _, p := range phrases {
		if strings.Contains(lower, p) {
			return true
		}
	}
	return false
}

func extractJSON(content string) string {
	content = strings.TrimSpace(content)

	// Strip a markdown JSON fence if present.
	if strings.HasPrefix(content, "```") {
		content = strings.TrimPrefix(content, "```json")
		content = strings.TrimPrefix(content, "```")
		content = strings.TrimSuffix(content, "```")
		content = strings.TrimSpace(content)
	}

	// If the content contains extra text before/after the JSON object, extract
	// the first balanced JSON object.
	start := strings.Index(content, "{")
	end := strings.LastIndex(content, "}")
	if start >= 0 && end > start {
		return content[start : end+1]
	}

	return content
}

func mapNetworkError(err error) error {
	if errors.Is(err, context.DeadlineExceeded) {
		return ErrProviderTimeout
	}
	var netErr interface{ Timeout() bool }
	if errors.As(err, &netErr) && netErr.Timeout() {
		return ErrProviderTimeout
	}
	return fmt.Errorf("%w: %v", ErrProviderTimeout, err)
}

func isRetryableError(err error) bool {
	if errors.Is(err, ErrProviderTimeout) {
		return true
	}
	return false
}

func isRetryableHTTPStatus(code int) bool {
	switch code {
	case http.StatusTooManyRequests, http.StatusBadGateway, http.StatusServiceUnavailable, http.StatusGatewayTimeout:
		return true
	}
	return false
}

// openCodeModelRef, openCodePart, openCodeMessageRequest, and
// openCodeMessageResponse mirror `opencode serve`'s real
// POST /session/{sessionID}/message request/response shape (confirmed live
// against its own OpenAPI document, GET /doc) - not an OpenAI-compatible
// chat-completions schema.
type openCodeModelRef struct {
	ProviderID string `json:"providerID"`
	ModelID    string `json:"modelID"`
}

type openCodePart struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type openCodeMessageRequest struct {
	System string            `json:"system,omitempty"`
	Model  *openCodeModelRef `json:"model,omitempty"`
	Parts  []openCodePart    `json:"parts"`
}

type openCodeMessageResponse struct {
	Info  json.RawMessage `json:"info"`
	Parts []openCodePart  `json:"parts"`
}

// compile-time interface check.
var _ FeedbackProvider = (*OpenCodeFeedbackProvider)(nil)
