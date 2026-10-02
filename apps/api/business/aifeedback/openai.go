package aifeedback

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"
	"unicode/utf8"
)

const (
	DefaultOpenAIModel        = "gpt-5-nano"
	openAIMaxOutputTokens     = 300
	openAINanoMaxOutputTokens = 1024
	openAIMaxResponseBytes    = 1 << 20
)

// OpenAIConfig configures Responses adapters. MaxRetries is clamped to zero or
// one for evaluation; runtime wiring and moderation use zero transport retries.
// The timeout applies per HTTP attempt, within the service request context.
type OpenAIConfig struct {
	BaseURL    string
	APIKey     string
	Model      string
	Timeout    time.Duration
	MaxRetries int
	OnFailure  func(OpenAIFailure)
}

// OpenAIFeedbackProvider supports explicit runtime and evaluator selection.
// The existing feedback interface does not expose usage;
// response usage is therefore not reported as measured billing or token data.
type OpenAIFeedbackProvider struct {
	config OpenAIConfig
	client *http.Client
}

func NewOpenAIFeedbackProvider(config OpenAIConfig) *OpenAIFeedbackProvider {
	if strings.TrimSpace(config.BaseURL) == "" {
		config.BaseURL = "https://api.openai.com/v1"
	}
	if strings.TrimSpace(config.Model) == "" {
		config.Model = DefaultOpenAIModel
	}
	if config.Timeout <= 0 {
		config.Timeout = 8 * time.Second
	}
	if config.MaxRetries < 0 {
		config.MaxRetries = 0
	}
	if config.MaxRetries > 1 {
		config.MaxRetries = 1
	}
	return &OpenAIFeedbackProvider{config: config, client: &http.Client{
		Timeout: config.Timeout,
		// Never forward the authorization header or learner payload on redirects.
		CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse },
	}}
}

func (p *OpenAIFeedbackProvider) EvaluationIdentity() (string, string) {
	return "openai", p.config.Model
}

func (p *OpenAIFeedbackProvider) GenerateFeedback(ctx context.Context, task ProviderTask) (feedback *ProviderFeedback, err error) {
	status := 0
	defer func() { p.recordFailure(OpenAIStageFeedback, status, err) }()
	body, err := buildOpenAIRequest(task, p.config.Model)
	if err != nil {
		return nil, err
	}
	data, status, err := p.sendResponse(ctx, body)
	if err != nil {
		return nil, err
	}
	return parseOpenAIResponse(data)
}

// sendResponse never returns raw provider errors or bodies on failure.
func (p *OpenAIFeedbackProvider) sendResponse(ctx context.Context, body []byte) ([]byte, int, error) {
	if err := ctx.Err(); err != nil {
		return nil, 0, err
	}
	if strings.TrimSpace(p.config.APIKey) == "" {
		return nil, 0, ErrProviderAuth
	}
	u, err := url.Parse(p.config.BaseURL)
	if err != nil || u.Host == "" || u.User != nil || u.RawQuery != "" || u.Fragment != "" ||
		(u.Scheme != "https" && !(u.Scheme == "http" && (u.Hostname() == "localhost" || u.Hostname() == "127.0.0.1" || u.Hostname() == "::1"))) {
		return nil, 0, ErrProviderInvalidInput
	}
	u.Path = strings.TrimRight(u.Path, "/") + "/responses"
	for attempt := 0; attempt <= p.config.MaxRetries; attempt++ {
		if err := ctx.Err(); err != nil {
			return nil, 0, err
		}
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, u.String(), bytes.NewReader(body))
		if err != nil {
			return nil, 0, ErrProviderInvalidInput
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+p.config.APIKey)
		resp, err := p.client.Do(req)
		if err != nil {
			if ctx.Err() != nil {
				return nil, 0, ctx.Err()
			}
			if attempt < p.config.MaxRetries {
				continue
			}
			return nil, 0, ErrProviderTimeout
		}
		data, readErr := io.ReadAll(io.LimitReader(resp.Body, openAIMaxResponseBytes+1))
		resp.Body.Close()
		if ctx.Err() != nil {
			return nil, 0, ctx.Err()
		}
		if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
			return nil, resp.StatusCode, ErrProviderAuth
		}
		if resp.StatusCode == http.StatusBadRequest || resp.StatusCode == http.StatusNotFound || resp.StatusCode == http.StatusUnprocessableEntity {
			return nil, resp.StatusCode, ErrProviderInvalidInput
		}
		if resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode >= 500 {
			if attempt < p.config.MaxRetries {
				continue
			}
			return nil, resp.StatusCode, ErrProviderTimeout
		}
		if errors.Is(readErr, context.DeadlineExceeded) {
			if attempt < p.config.MaxRetries {
				continue
			}
			return nil, 0, ErrProviderTimeout
		}
		if resp.StatusCode != http.StatusOK || readErr != nil || len(data) > openAIMaxResponseBytes {
			return nil, resp.StatusCode, ErrProviderInvalidResponse
		}
		return data, resp.StatusCode, nil
	}
	return nil, 0, ErrProviderTimeout
}

func buildOpenAIRequest(task ProviderTask, model string) ([]byte, error) {
	return buildOpenAIStructuredRequest(task, model, outputSchema(), "sentence_feedback")
}

// Only the two internal canonical contracts reach this helper. It copies the
// schema before strict adaptation and never changes the caller's task.
func buildOpenAIStructuredRequest(task ProviderTask, model string, expected map[string]any, name string) ([]byte, error) {
	effort, profileLimit, err := openAIModelProfile(model)
	if err != nil {
		return nil, err
	}
	if task.EnableTools || task.EnableWebSearch || task.EnableMemory {
		return nil, ErrProviderInvalidInput
	}
	// Reject any schema outside the selected internal contract.
	schemaJSON, err := json.Marshal(task.OutputSchema)
	canonical, canonicalErr := json.Marshal(expected)
	if err != nil || canonicalErr != nil || !bytes.Equal(schemaJSON, canonical) {
		return nil, ErrProviderInvalidInput
	}
	var schema map[string]any
	if json.Unmarshal(schemaJSON, &schema) != nil {
		return nil, ErrProviderInvalidInput
	}
	properties := schema["properties"].(map[string]any)
	required := make([]string, 0, len(properties))
	for name := range properties {
		required = append(required, name)
	}
	sort.Strings(required)
	schema["required"] = required
	schema["additionalProperties"] = false
	userJSON, err := json.Marshal(task.UserPayload)
	if err != nil {
		return nil, ErrProviderInvalidInput
	}
	// max_output_tokens includes hidden reasoning as well as the visible JSON.
	// The GPT-5 nano profile permits an explicit larger canary budget, but never
	// raises a positive caller ceiling. Canonical evaluation tasks remain at 300;
	// selecting this model alone does not establish a sufficient reasoning budget.
	maxTokens := task.MaxOutputTokens
	if maxTokens <= 0 {
		maxTokens = openAIMaxOutputTokens
	}
	if maxTokens > profileLimit {
		maxTokens = profileLimit
	}
	request := map[string]any{
		"model": model,
		"input": []map[string]string{
			{"role": "system", "content": task.SystemPrompt},
			{"role": "developer", "content": task.DeveloperPrompt},
			{"role": "user", "content": string(userJSON)},
		},
		"text": map[string]any{"format": map[string]any{
			"type": "json_schema", "name": name, "strict": true, "schema": schema,
		}},
		"store":             false,
		"tools":             []any{},
		"max_output_tokens": maxTokens,
	}
	if effort != "" {
		request["reasoning"] = map[string]string{"effort": effort}
	}
	body, err := json.Marshal(request)
	if err != nil {
		return nil, ErrProviderInvalidInput
	}
	return body, nil
}

// Only reviewed model families receive a request profile. Dated snapshots use
// the same profile, but syntax acceptance does not establish API availability.
func openAIModelProfile(model string) (string, int, error) {
	for _, profile := range []struct {
		model  string
		effort string
		limit  int
	}{
		{"gpt-5-nano", "minimal", openAINanoMaxOutputTokens},
		{"gpt-4.1-nano", "", openAIMaxOutputTokens},
		{"gpt-4o-mini", "", openAIMaxOutputTokens},
		{"gpt-6-luna", "none", openAIMaxOutputTokens},
	} {
		if model == profile.model {
			return profile.effort, profile.limit, nil
		}
		if suffix, found := strings.CutPrefix(model, profile.model+"-"); found {
			if date, err := time.Parse("2006-01-02", suffix); err == nil && date.Format("2006-01-02") == suffix {
				return profile.effort, profile.limit, nil
			}
		}
	}
	return "", 0, ErrProviderInvalidInput
}

func parseOpenAIOutputText(data []byte) (string, error) {
	var response struct {
		Status string          `json:"status"`
		Error  json.RawMessage `json:"error"`
		Output []struct {
			Type    string `json:"type"`
			Role    string `json:"role"`
			Status  string `json:"status"`
			Content []struct {
				Type string `json:"type"`
				Text string `json:"text"`
			} `json:"content"`
		} `json:"output"`
	}
	if json.Unmarshal(data, &response) != nil || response.Status != "completed" ||
		(len(response.Error) != 0 && string(response.Error) != "null") {
		return "", ErrProviderInvalidResponse
	}
	var text string
	for _, item := range response.Output {
		if item.Type == "reasoning" {
			continue
		}
		if item.Type != "message" || item.Role != "assistant" || item.Status != "completed" {
			return "", ErrProviderInvalidResponse
		}
		for _, part := range item.Content {
			if part.Type == "refusal" {
				return "", ErrProviderRefusal
			}
			if part.Type != "output_text" || text != "" || strings.TrimSpace(part.Text) == "" {
				return "", ErrProviderInvalidResponse
			}
			text = part.Text
		}
	}
	if text == "" {
		return "", ErrProviderInvalidResponse
	}
	return text, nil
}

func parseOpenAIResponse(data []byte) (*ProviderFeedback, error) {
	text, err := parseOpenAIOutputText(data)
	if err != nil {
		return nil, err
	}
	var raw map[string]any
	if json.Unmarshal([]byte(text), &raw) != nil || !validOpenAIFeedbackShape(raw) {
		return nil, ErrProviderInvalidResponse
	}
	feedback := &ProviderFeedback{
		Status:                  raw["status"].(string),
		TargetWordUsedCorrectly: raw["target_word_used_correctly"].(bool),
		GrammarAcceptable:       raw["grammar_acceptable"].(bool),
		MeaningClear:            raw["meaning_clear"].(bool),
		Naturalness:             raw["naturalness"].(string),
		Headline:                raw["headline"].(string),
		Explanation:             raw["explanation"].(string),
		RawJSON:                 raw,
	}
	if value, ok := raw["corrected_sentence"].(string); ok {
		feedback.CorrectedSentence = &value
	}
	if value, ok := raw["improvement_tip"].(string); ok {
		feedback.ImprovementTip = &value
	}
	// Semantic contradictions belong to the existing evaluator's validator so
	// its report can distinguish returned feedback from transport/schema failure.
	return feedback, nil
}

func validOpenAIFeedbackShape(raw map[string]any) bool {
	properties := outputSchema()["properties"].(map[string]any)
	if len(raw) != len(properties) {
		return false
	}
	for name, definition := range properties {
		value, present := raw[name]
		if !present {
			return false
		}
		if value == nil && (name == "corrected_sentence" || name == "improvement_tip") {
			continue
		}
		property := definition.(map[string]any)
		if property["type"] == "boolean" {
			if _, ok := value.(bool); !ok {
				return false
			}
			continue
		}
		text, ok := value.(string)
		if !ok {
			return false
		}
		if limit, ok := property["maxLength"].(int); ok && utf8.RuneCountInString(text) > limit {
			return false
		}
		if options, ok := property["enum"].([]string); ok {
			matched := false
			for _, option := range options {
				matched = matched || text == option
			}
			if !matched {
				return false
			}
		}
	}
	return true
}

var _ FeedbackProvider = (*OpenAIFeedbackProvider)(nil)
var _ EvaluationProviderIdentity = (*OpenAIFeedbackProvider)(nil)
