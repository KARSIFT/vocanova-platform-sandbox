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
// they are never hard-coded or committed to source. Model, temperature, and
// max-output tokens are controlled by the version-controlled prompt architecture
// (ProviderTask); the adapter only supplies the runtime connection details.
type OpenCodeConfig struct {
	BaseURL    string
	APIKey     string
	Model      string
	Endpoint   string
	Timeout    time.Duration
	MaxRetries int
}

// DefaultOpenCodeConfig returns the default adapter configuration. Production
// callers must override BaseURL and APIKey from backend-only secrets.
func DefaultOpenCodeConfig() OpenCodeConfig {
	return OpenCodeConfig{
		BaseURL:    "http://127.0.0.1:3000",
		Endpoint:   "/v1/chat/completions",
		Model:      DefaultOpenCodeModel,
		Timeout:    8 * time.Second,
		MaxRetries: 1,
	}
}

// OpenCodeFeedbackProvider implements FeedbackProvider against the OpenCode
// serve HTTP endpoint. Provider SDK/network types are confined to this adapter.
type OpenCodeFeedbackProvider struct {
	config OpenCodeConfig
	client *http.Client
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
	if config.Endpoint == "" {
		config.Endpoint = "/v1/chat/completions"
	}
	return &OpenCodeFeedbackProvider{
		config: config,
		client: &http.Client{Timeout: config.Timeout},
	}
}

// GenerateFeedback sends a structured chat request to the configured OpenCode
// endpoint and returns the parsed ProviderFeedback. It makes at most one
// transport retry for clearly transient failures (DOC-09 §18).
func (p *OpenCodeFeedbackProvider) GenerateFeedback(ctx context.Context, task ProviderTask) (*ProviderFeedback, error) {
	body, err := p.buildRequestBody(task)
	if err != nil {
		return nil, fmt.Errorf("build request body: %w", err)
	}

	url := strings.TrimRight(p.config.BaseURL, "/") + p.config.Endpoint
	var lastErr error
	maxAttempts := p.config.MaxRetries + 1
	if maxAttempts < 1 {
		maxAttempts = 1
	}

	for attempt := 0; attempt < maxAttempts; attempt++ {
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
		if err != nil {
			return nil, fmt.Errorf("create request: %w", err)
		}
		req.Header.Set("Content-Type", "application/json")
		if p.config.APIKey != "" {
			req.Header.Set("Authorization", "Bearer "+p.config.APIKey)
		}

		resp, err := p.client.Do(req)
		if err != nil {
			lastErr = mapNetworkError(err)
			if attempt < maxAttempts-1 && isRetryableError(lastErr) {
				continue
			}
			return nil, lastErr
		}

		respBody, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
		resp.Body.Close()
		if err != nil {
			lastErr = fmt.Errorf("read response body: %w", err)
			if attempt < maxAttempts-1 && isRetryableHTTPStatus(resp.StatusCode) {
				continue
			}
			return nil, lastErr
		}

		feedback, err := p.parseResponse(resp.StatusCode, respBody)
		if err != nil {
			lastErr = err
			if attempt < maxAttempts-1 && isRetryableHTTPStatus(resp.StatusCode) {
				continue
			}
			return nil, lastErr
		}
		return feedback, nil
	}

	return nil, lastErr
}

func (p *OpenCodeFeedbackProvider) buildRequestBody(task ProviderTask) ([]byte, error) {
	payloadJSON, err := json.Marshal(task.UserPayload)
	if err != nil {
		return nil, fmt.Errorf("marshal user payload: %w", err)
	}

	messages := []openCodeMessage{
		{Role: "system", Content: task.SystemPrompt},
		{Role: "developer", Content: task.DeveloperPrompt},
		{Role: "user", Content: string(payloadJSON)},
	}

	req := openCodeRequest{
		Model:       p.config.Model,
		Messages:    messages,
		Temperature: task.Temperature,
		MaxTokens:   task.MaxOutputTokens,
		Stream:      false,
	}

	if task.OutputSchema != nil {
		req.ResponseFormat = map[string]any{
			"type": "json_schema",
			"json_schema": map[string]any{
				"name":   "sentence_feedback",
				"strict": true,
				"schema": task.OutputSchema,
			},
		}
	}

	return json.Marshal(req)
}

func (p *OpenCodeFeedbackProvider) parseResponse(statusCode int, body []byte) (*ProviderFeedback, error) {
	var resp openCodeResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrProviderInvalidResponse, err)
	}

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

	if resp.Error != nil {
		msg := strings.ToLower(resp.Error.Message)
		if strings.Contains(msg, "auth") || strings.Contains(msg, "unauthorized") || strings.Contains(msg, "forbidden") || strings.Contains(msg, "api key") {
			return nil, ErrProviderAuth
		}
		if strings.Contains(msg, "refus") || strings.Contains(msg, "cannot") || strings.Contains(msg, "unable") || strings.Contains(msg, "content") {
			return nil, ErrProviderRefusal
		}
		return nil, fmt.Errorf("%w: %s", ErrProviderInvalidResponse, resp.Error.Message)
	}

	if len(resp.Choices) == 0 {
		return nil, ErrProviderInvalidResponse
	}

	content := strings.TrimSpace(resp.Choices[0].Message.Content)
	finishReason := strings.ToLower(resp.Choices[0].FinishReason)

	if finishReason == "content_filter" || finishReason == "content_filter_finish" {
		return nil, ErrProviderRefusal
	}
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

type openCodeMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type openCodeRequest struct {
	Model          string            `json:"model"`
	Messages       []openCodeMessage `json:"messages"`
	ResponseFormat map[string]any    `json:"response_format,omitempty"`
	Temperature    float64           `json:"temperature"`
	MaxTokens      int               `json:"max_tokens"`
	Stream         bool              `json:"stream"`
}

type openCodeResponse struct {
	Choices []struct {
		Message      openCodeMessage `json:"message"`
		FinishReason string          `json:"finish_reason"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
		Type    string `json:"type"`
		Code    string `json:"code"`
	} `json:"error"`
}

// compile-time interface check.
var _ FeedbackProvider = (*OpenCodeFeedbackProvider)(nil)
