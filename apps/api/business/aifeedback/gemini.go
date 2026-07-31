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
	defaultGeminiModel   = "gemini-2.5-flash"
	defaultGeminiBaseURL = "https://generativelanguage.googleapis.com"
)

// GeminiConfig configures the Gemini production adapters.
type GeminiConfig struct {
	APIKey     string
	Model      string
	BaseURL    string
	Timeout    time.Duration
	MaxRetries int
}

type geminiTransport struct {
	config GeminiConfig
	client *http.Client
}

func newGeminiTransport(config GeminiConfig) *geminiTransport {
	if strings.TrimSpace(config.Model) == "" {
		config.Model = defaultGeminiModel
	}
	if strings.TrimSpace(config.BaseURL) == "" {
		config.BaseURL = defaultGeminiBaseURL
	}
	if config.Timeout <= 0 {
		config.Timeout = 8 * time.Second
	}
	return &geminiTransport{
		config: config,
		client: &http.Client{Timeout: config.Timeout},
	}
}

type GeminiFeedbackProvider struct {
	*geminiTransport
}

func NewGeminiFeedbackProvider(config GeminiConfig) *GeminiFeedbackProvider {
	return &GeminiFeedbackProvider{
		geminiTransport: newGeminiTransport(config),
	}
}

func (p *GeminiFeedbackProvider) GenerateFeedback(ctx context.Context, task ProviderTask) (*ProviderFeedback, error) {
	body, err := buildGeminiGenerateContentBody(task.SystemPrompt, task.DeveloperPrompt, task.UserPayload, task.OutputSchema)
	if err != nil {
		return nil, fmt.Errorf("build request body: %w", err)
	}

	var result *ProviderFeedback
	parseErr := p.sendWithRetry(ctx, body, func(statusCode int, respBody []byte) error {
		raw, err := parseGeminiJSONResponse(statusCode, respBody)
		if err != nil {
			return err
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

type GeminiModerationProvider struct {
	*geminiTransport
}

func NewGeminiModerationProvider(config GeminiConfig) *GeminiModerationProvider {
	config.MaxRetries = 0
	return &GeminiModerationProvider{
		geminiTransport: newGeminiTransport(config),
	}
}

func (p *GeminiModerationProvider) Classify(ctx context.Context, input ModerationInput) (*ModerationResult, error) {
	if strings.TrimSpace(input.SentenceText) == "" {
		return nil, fmt.Errorf("moderation input missing learner sentence")
	}

	userPayload := map[string]any{
		"learner_sentence": input.SentenceText,
		"target_word":      input.TargetWord,
		"learner_level":    input.LearnerLevel,
	}
	body, err := buildGeminiGenerateContentBody(
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
		raw, err := parseGeminiJSONResponse(statusCode, respBody)
		if err != nil {
			return err
		}
		rawJSON, err := json.Marshal(raw)
		if err != nil {
			return fmt.Errorf("%w: %v", ErrProviderInvalidResponse, err)
		}

		openCodeBody, err := json.Marshal(openCodeMessageResponse{
			Parts: []openCodePart{{Type: "text", Text: string(rawJSON)}},
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

type geminiSendParser func(statusCode int, body []byte) error

func (t *geminiTransport) sendWithRetry(ctx context.Context, requestBody []byte, parse geminiSendParser) error {
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

func (t *geminiTransport) sendOnce(ctx context.Context, requestBody []byte) (int, []byte, error) {
	url := strings.TrimRight(t.config.BaseURL, "/") + "/v1beta/models/" + t.config.Model + ":generateContent"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(requestBody))
	if err != nil {
		return 0, nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if t.config.APIKey != "" {
		req.Header.Set("x-goog-api-key", t.config.APIKey)
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

func buildGeminiGenerateContentBody(
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

	req := geminiGenerateContentRequest{
		Contents: []geminiContent{
			{
				Role:  "user",
				Parts: []geminiPart{{Text: userText.String()}},
			},
		},
		SystemInstruction: geminiSystemInstruction{
			Parts: []geminiPart{{Text: systemPrompt}},
		},
		GenerationConfig: geminiGenerationConfig{
			ResponseMimeType: "application/json",
			ResponseSchema:   responseSchema,
		},
	}

	return json.Marshal(req)
}

func parseGeminiJSONResponse(statusCode int, body []byte) (map[string]any, error) {
	text, err := parseGeminiTextResponse(statusCode, body)
	if err != nil {
		return nil, err
	}

	text = extractJSON(text)
	if text == "" {
		return nil, ErrProviderInvalidResponse
	}

	var raw map[string]any
	if err := json.Unmarshal([]byte(text), &raw); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrProviderInvalidResponse, err)
	}

	return raw, nil
}

func parseGeminiTextResponse(statusCode int, body []byte) (string, error) {
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

	var resp geminiGenerateContentResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return "", fmt.Errorf("%w: %v", ErrProviderInvalidResponse, err)
	}

	if strings.TrimSpace(resp.PromptFeedback.BlockReason) != "" {
		return "", ErrProviderRefusal
	}
	if len(resp.Candidates) == 0 {
		return "", ErrProviderInvalidResponse
	}
	if strings.TrimSpace(resp.Candidates[0].FinishReason) != "STOP" {
		return "", ErrProviderRefusal
	}

	var textBuilder strings.Builder
	for _, part := range resp.Candidates[0].Content.Parts {
		if part.Text != "" {
			textBuilder.WriteString(part.Text)
		}
	}

	content := strings.TrimSpace(textBuilder.String())
	if content == "" {
		return "", ErrProviderInvalidResponse
	}
	return content, nil
}

type geminiGenerateContentRequest struct {
	Contents          []geminiContent         `json:"contents"`
	SystemInstruction geminiSystemInstruction `json:"systemInstruction"`
	GenerationConfig  geminiGenerationConfig  `json:"generationConfig"`
}

type geminiContent struct {
	Role  string       `json:"role"`
	Parts []geminiPart `json:"parts"`
}

type geminiPart struct {
	Text string `json:"text"`
}

type geminiSystemInstruction struct {
	Parts []geminiPart `json:"parts"`
}

type geminiGenerationConfig struct {
	ResponseMimeType string         `json:"responseMimeType"`
	ResponseSchema   map[string]any `json:"responseSchema"`
}

type geminiGenerateContentResponse struct {
	Candidates     []geminiCandidate `json:"candidates"`
	PromptFeedback struct {
		BlockReason string `json:"blockReason"`
	} `json:"promptFeedback"`
}

type geminiCandidate struct {
	Content struct {
		Parts []geminiPart `json:"parts"`
	} `json:"content"`
	FinishReason string `json:"finishReason"`
}

var _ FeedbackProvider = (*GeminiFeedbackProvider)(nil)
var _ ModerationProvider = (*GeminiModerationProvider)(nil)
