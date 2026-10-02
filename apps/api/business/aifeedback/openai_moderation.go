package aifeedback

import (
	"context"
	"encoding/json"
	"strings"
	"unicode/utf8"
)

// OpenAIModerationProvider uses the same four-outcome learning policy as the
// other runtime adapters. It does not substitute a generic flagged boolean.
type OpenAIModerationProvider struct {
	transport *OpenAIFeedbackProvider
}

func NewOpenAIModerationProvider(config OpenAIConfig) *OpenAIModerationProvider {
	config.MaxRetries = 0
	return &OpenAIModerationProvider{transport: NewOpenAIFeedbackProvider(config)}
}

func (p *OpenAIModerationProvider) Classify(ctx context.Context, input ModerationInput) (result *ModerationResult, err error) {
	status := 0
	defer func() { p.transport.recordFailure(OpenAIStageModeration, status, err) }()
	if strings.TrimSpace(input.SentenceText) == "" {
		return nil, ErrProviderInvalidInput
	}
	task := ProviderTask{
		SystemPrompt: moderationSystemPrompt(), DeveloperPrompt: moderationDeveloperPrompt(),
		UserPayload:  map[string]any{"learner_sentence": input.SentenceText, "target_word": input.TargetWord, "learner_level": input.LearnerLevel},
		OutputSchema: moderationOutputSchema(), MaxOutputTokens: openAIMaxOutputTokens,
	}
	body, err := buildOpenAIStructuredRequest(task, p.transport.config.Model, moderationOutputSchema(), "sentence_moderation")
	if err != nil {
		return nil, err
	}
	data, status, err := p.transport.sendResponse(ctx, body)
	if err != nil {
		return nil, err
	}
	text, err := parseOpenAIOutputText(data)
	if err != nil {
		return nil, err
	}
	var raw map[string]any
	if json.Unmarshal([]byte(text), &raw) != nil || len(raw) != 2 {
		return nil, ErrProviderInvalidResponse
	}
	outcome, ok := raw["outcome"].(string)
	if !ok {
		return nil, ErrProviderInvalidResponse
	}
	reason, ok := raw["reason"].(string)
	if !ok || utf8.RuneCountInString(reason) > 200 {
		return nil, ErrProviderInvalidResponse
	}
	switch outcome {
	case SafetyAllowed, SafetyAllowedSensitive, SafetyBlocked, SafetySelfHarmIntervention:
		return &ModerationResult{Outcome: outcome, Reason: reason}, nil
	default:
		return nil, ErrProviderInvalidResponse
	}
}

var _ ModerationProvider = (*OpenAIModerationProvider)(nil)
