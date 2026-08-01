package aifeedback

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const (
	defaultCloudflareModel   = "@cf/meta/llama-3.3-70b-instruct-fp8-fast"
	defaultCloudflareBaseURL = "https://api.cloudflare.com/client/v4"
)

// CloudflareConfig configures the Cloudflare Workers AI adapters.
type CloudflareConfig struct {
	APIToken   string
	AccountID  string
	Model      string
	BaseURL    string
	Timeout    time.Duration
	MaxRetries int
}

type cloudflareTransport struct {
	config CloudflareConfig
	client *http.Client
}

func newCloudflareTransport(config CloudflareConfig) *cloudflareTransport {
	if strings.TrimSpace(config.Model) == "" {
		config.Model = defaultCloudflareModel
	}
	if strings.TrimSpace(config.BaseURL) == "" {
		config.BaseURL = defaultCloudflareBaseURL
	}
	if config.Timeout <= 0 {
		config.Timeout = 8 * time.Second
	}
	return &cloudflareTransport{
		config: config,
		client: &http.Client{Timeout: config.Timeout},
	}
}

type CloudflareFeedbackProvider struct {
	*cloudflareTransport
}

func NewCloudflareFeedbackProvider(config CloudflareConfig) *CloudflareFeedbackProvider {
	return &CloudflareFeedbackProvider{
		cloudflareTransport: newCloudflareTransport(config),
	}
}

func (p *CloudflareFeedbackProvider) GenerateFeedback(ctx context.Context, task ProviderTask) (*ProviderFeedback, error) {
	body, err := buildCloudflareRequestBody(task.SystemPrompt, task.DeveloperPrompt, task.UserPayload, task.OutputSchema)
	if err != nil {
		return nil, fmt.Errorf("build request body: %w", err)
	}

	var result *ProviderFeedback
	parseErr := p.sendWithRetry(ctx, body, func(statusCode int, respBody []byte) error {
		content, err := parseCloudflareTextResponse(statusCode, respBody)
		if err != nil {
			return err
		}

		content = extractJSON(content)
		if content == "" {
			return ErrProviderInvalidResponse
		}

		var raw map[string]any
		if err := json.Unmarshal([]byte(content), &raw); err != nil {
			return fmt.Errorf("%w: %v", ErrProviderInvalidResponse, err)
		}
		feedback, err := mapOpenCodeFeedback(raw)
		if err != nil {
			return err
		}
		result = feedback
		return nil
	})
	if parseErr != nil {
		return nil, parseErr
	}

	return result, nil
}

type CloudflareModerationProvider struct {
	*cloudflareTransport
}

func NewCloudflareModerationProvider(config CloudflareConfig) *CloudflareModerationProvider {
	config.MaxRetries = 0
	return &CloudflareModerationProvider{
		cloudflareTransport: newCloudflareTransport(config),
	}
}

func (p *CloudflareModerationProvider) Classify(ctx context.Context, input ModerationInput) (*ModerationResult, error) {
	if strings.TrimSpace(input.SentenceText) == "" {
		return nil, fmt.Errorf("moderation input missing learner sentence")
	}

	userPayload := map[string]any{
		"learner_sentence": input.SentenceText,
		"target_word":      input.TargetWord,
		"learner_level":    input.LearnerLevel,
	}
	body, err := buildCloudflareRequestBody(
		moderationSystemPrompt(),
		moderationDeveloperPrompt(),
		userPayload,
		moderationOutputSchema(),
	)
	if err != nil {
		return nil, fmt.Errorf("build moderation request body: %w", err)
	}

	var result *ModerationResult
	parseErr := p.sendWithRetry(ctx, body, func(statusCode int, respBody []byte) error {
		content, err := parseCloudflareTextResponse(statusCode, respBody)
		if err != nil {
			return err
		}

		content = extractJSON(content)
		if content == "" {
			return ErrProviderInvalidResponse
		}

		openCodeBody, err := json.Marshal(openCodeMessageResponse{
			Parts: []openCodePart{{Type: "text", Text: content}},
		})
		if err != nil {
			return fmt.Errorf("%w: %v", ErrProviderInvalidResponse, err)
		}

		parsed, err := parseModerationResponse(http.StatusOK, openCodeBody)
		if err != nil {
			return err
		}
		result = parsed
		return nil
	})
	if parseErr != nil {
		return nil, parseErr
	}

	return result, nil
}

type cloudflareSendParser func(statusCode int, body []byte) error

func (t *cloudflareTransport) sendWithRetry(ctx context.Context, requestBody []byte, parse cloudflareSendParser) error {
	var lastErr error
	maxAttempts := t.config.MaxRetries + 1
	if maxAttempts < 1 {
		maxAttempts = 1
	}

	for attempt := 0; attempt < maxAttempts; attempt++ {
		statusCode, respBody, err := t.sendOnce(ctx, requestBody)
		if err != nil {
			lastErr = err
			if attempt < maxAttempts-1 && isRetryableError(err) {
				continue
			}
			return lastErr
		}

		if err := parse(statusCode, respBody); err != nil {
			lastErr = err
			if attempt < maxAttempts-1 && isRetryableHTTPStatus(statusCode) {
				continue
			}
			return lastErr
		}
		return nil
	}

	return lastErr
}

func (t *cloudflareTransport) sendOnce(ctx context.Context, requestBody []byte) (int, []byte, error) {
	accountID := strings.TrimSpace(t.config.AccountID)
	if accountID == "" {
		return 0, nil, fmt.Errorf("%w: missing cloudflare account id", ErrProviderInvalidInput)
	}

	baseURL := strings.TrimRight(t.config.BaseURL, "/")
	model := strings.TrimSpace(t.config.Model)
	url := baseURL + "/accounts/" + accountID + "/ai/run/" + model

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(requestBody))
	if err != nil {
		return 0, nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if t.config.APIToken != "" {
		req.Header.Set("Authorization", "Bearer "+t.config.APIToken)
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

func buildCloudflareRequestBody(
	systemPrompt string,
	developerPrompt string,
	userPayload map[string]any,
	responseSchema map[string]any,
) ([]byte, error) {
	payloadJSON, err := json.Marshal(userPayload)
	if err != nil {
		return nil, fmt.Errorf("marshal user payload: %w", err)
	}

	var userText strings.Builder
	userText.WriteString(developerPrompt)
	userText.WriteString("\n\nTask data (JSON):\n")
	userText.Write(payloadJSON)

	req := cloudflareRunRequest{
		Messages: []cloudflareMessage{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: userText.String()},
		},
		ResponseFormat: cloudflareResponseFormat{
			Type:       "json_schema",
			JSONSchema: responseSchema,
		},
	}
	return json.Marshal(req)
}

func parseCloudflareTextResponse(statusCode int, body []byte) (string, error) {
	if statusCode == http.StatusUnauthorized || statusCode == http.StatusForbidden {
		return "", ErrProviderAuth
	}
	if statusCode == http.StatusBadRequest {
		return "", ErrProviderInvalidInput
	}
	if statusCode >= 500 || statusCode == http.StatusTooManyRequests {
		return "", ErrProviderTimeout
	}
	if statusCode != http.StatusOK {
		return "", fmt.Errorf("%w: unexpected status %d", ErrProviderInvalidResponse, statusCode)
	}

	var resp cloudflareRunResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return "", fmt.Errorf("%w: %v", ErrProviderInvalidResponse, err)
	}

	if !resp.Success {
		if len(resp.Errors) > 0 && strings.TrimSpace(resp.Errors[0].Message) != "" {
			return "", fmt.Errorf("%w: %s", ErrProviderInvalidResponse, strings.TrimSpace(resp.Errors[0].Message))
		}
		return "", fmt.Errorf("%w: cloudflare response success=false", ErrProviderInvalidResponse)
	}

	content := strings.TrimSpace(string(resp.Result.Response))
	if content == "" || content == "null" {
		return "", ErrProviderInvalidResponse
	}
	if isProviderRefusal(content) {
		return "", ErrProviderRefusal
	}

	return content, nil
}

type cloudflareRunRequest struct {
	Messages       []cloudflareMessage      `json:"messages"`
	ResponseFormat cloudflareResponseFormat `json:"response_format"`
}

type cloudflareMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type cloudflareResponseFormat struct {
	Type       string         `json:"type"`
	JSONSchema map[string]any `json:"json_schema"`
}

// cloudflareRunResponse's Result.Response is json.RawMessage, not string.
// Cloudflare's `response_format: {type: "json_schema", ...}` mode returns
// `result.response` as an actual parsed JSON object matching the schema
// (e.g. `{"status":"correct","explanation":"..."}`), not a JSON-encoded
// string containing that object - confirmed live against the real API
// (VOC-036-T03's first live-evaluation attempt: every one of 56 calls
// failed with ProviderCalled=0 because `Response string` could never
// successfully unmarshal an object). json.RawMessage captures the field's
// raw bytes verbatim regardless of whether the underlying JSON is a string
// or an object, which is exactly what downstream extractJSON/json.Unmarshal
// already expects to receive - no other parsing logic changes needed.
type cloudflareRunResponse struct {
	Result struct {
		Response json.RawMessage `json:"response"`
	} `json:"result"`
	Success bool `json:"success"`
	Errors  []struct {
		Message string `json:"message"`
	} `json:"errors"`
}

var _ FeedbackProvider = (*CloudflareFeedbackProvider)(nil)
var _ ModerationProvider = (*CloudflareModerationProvider)(nil)
