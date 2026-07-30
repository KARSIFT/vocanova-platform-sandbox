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

// openCodeTransport is the unexported, behavior-preserving session/HTTP layer
// shared by OpenCodeFeedbackProvider and OpenCodeModerationProvider (VOC-034-D01).
// It owns session creation, the retry loop, and the network/HTTP-status error
// mapping so both providers get identical fail-closed transport behavior rather
// than an independently-maintained copy that could drift (a correctness risk
// for the moderation adapter in particular, where a divergent timeout/auth
// mapping would silently change the safety outcome).
type openCodeTransport struct {
	config     OpenCodeConfig
	client     *http.Client
	providerID string
	modelID    string
}

func newOpenCodeTransport(config OpenCodeConfig) *openCodeTransport {
	if config.Model == "" {
		config.Model = DefaultOpenCodeModel
	}
	if config.Timeout <= 0 {
		config.Timeout = 8 * time.Second
	}
	providerID, modelID := splitOpenCodeModel(config.Model)
	return &openCodeTransport{
		config:     config,
		client:     &http.Client{Timeout: config.Timeout},
		providerID: providerID,
		modelID:    modelID,
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
	*openCodeTransport
}

// NewOpenCodeFeedbackProvider creates a production adapter. The API key is kept
// in the config struct and used only in request headers; it is never logged.
func NewOpenCodeFeedbackProvider(config OpenCodeConfig) *OpenCodeFeedbackProvider {
	return &OpenCodeFeedbackProvider{
		openCodeTransport: newOpenCodeTransport(config),
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

	var result *ProviderFeedback
	parseErr := p.sendWithRetry(ctx, body, func(statusCode int, respBody []byte) error {
		feedback, perr := p.parseMessageResponse(statusCode, respBody)
		if perr != nil {
			return perr
		}
		result = feedback
		return nil
	})
	if parseErr != nil {
		return nil, parseErr
	}
	return result, nil
}

// openCodeSendParser is the per-attempt response parser passed to
// openCodeTransport.sendWithRetry. It receives the HTTP status code and the
// already-read response body, performs provider-specific parsing, and returns
// any error. A non-nil error is treated by the transport as a candidate for
// retry when the HTTP status is in the retryable set.
type openCodeSendParser func(statusCode int, body []byte) error

// sendWithRetry is the shared retry loop. It creates a fresh session per
// attempt, sends messageBody, reads the response, and invokes parse for
// provider-specific parsing. A network error or a parse error on a retryable
// HTTP status retries up to MaxRetries times. The retry budget is the same
// field the feedback adapter has always used (DOC-09 §18); moderation sets
// MaxRetries to 0 at construction time to make its own budget explicit
// (VOC-034-D03).
func (t *openCodeTransport) sendWithRetry(ctx context.Context, messageBody []byte, parse openCodeSendParser) error {
	var lastErr error
	maxAttempts := t.config.MaxRetries + 1
	if maxAttempts < 1 {
		maxAttempts = 1
	}

	for attempt := 0; attempt < maxAttempts; attempt++ {
		statusCode, respBody, err := t.sendOnce(ctx, messageBody)
		if err != nil {
			lastErr = err
			if attempt < maxAttempts-1 && isRetryableError(err) {
				continue
			}
			return lastErr
		}
		if perr := parse(statusCode, respBody); perr != nil {
			lastErr = perr
			if attempt < maxAttempts-1 && isRetryableHTTPStatus(statusCode) {
				continue
			}
			return lastErr
		}
		return nil
	}

	return lastErr
}

// sendOnce performs a single transport attempt: create a session, send the
// message, read the response body. It does not retry, and it does not parse -
// the caller decides what the response means via sendWithRetry's parser.
func (t *openCodeTransport) sendOnce(ctx context.Context, messageBody []byte) (int, []byte, error) {
	sessionID, err := t.createSession(ctx)
	if err != nil {
		return 0, nil, err
	}

	url := strings.TrimRight(t.config.BaseURL, "/") + "/session/" + sessionID + "/message"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(messageBody))
	if err != nil {
		return 0, nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if t.config.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+t.config.APIKey)
	}

	resp, err := t.client.Do(req)
	if err != nil {
		return 0, nil, mapNetworkError(err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return resp.StatusCode, nil, fmt.Errorf("read response body: %w", err)
	}
	return resp.StatusCode, respBody, nil
}

func (t *openCodeTransport) createSession(ctx context.Context) (string, error) {
	url := strings.TrimRight(t.config.BaseURL, "/") + "/session"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader([]byte("{}")))
	if err != nil {
		return "", fmt.Errorf("create session request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if t.config.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+t.config.APIKey)
	}

	resp, err := t.client.Do(req)
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
