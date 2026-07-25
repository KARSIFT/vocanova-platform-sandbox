// Package aifeedback implements the AI-feedback domain boundaries: narrow
// FeedbackProvider and ModerationProvider interfaces, internal provider schema
// DTOs, the public SentenceFeedbackResult contract, and a deterministic mock
// provider so CI and orchestration tests never depend on a paid external model.
package aifeedback

import (
	"context"
	"errors"
	"strings"

	"github.com/google/uuid"
)

// Learning status values (DOC-09 §7). These describe the pedagogical outcome
// of a successful feedback attempt.
const (
	LearningStatusCorrect          = "correct"
	LearningStatusNeedsImprovement = "needs_improvement"
	LearningStatusIncorrect        = "incorrect"
)

// Operational attempt statuses for ai_feedback_attempts.
const (
	AttemptStatusPending   = "pending"
	AttemptStatusSucceeded = "succeeded"
	AttemptStatusFailed    = "failed"
	AttemptStatusCancelled = "cancelled"
)

// Learner-sentence lifecycle statuses.
const (
	SentenceStatusSubmitted      = "submitted"
	SentenceStatusFeedbackReady  = "feedback_ready"
	SentenceStatusFeedbackFailed = "feedback_failed"
	SentenceStatusArchived       = "archived"
)

// Safety/moderation outcomes (DOC-09 §15). These are internal and must never be
// shown directly to learners.
const (
	SafetyAllowed               = "allowed"
	SafetyAllowedSensitive      = "allowed_sensitive"
	SafetyBlocked               = "blocked"
	SafetySelfHarmIntervention  = "self_harm_intervention"
	SafetyModerationUnavailable = "moderation_unavailable"
)

// Prompt and schema versions (DOC-09 §14). Material prompt changes must create a
// new version; version strings live in version-controlled code.
const (
	PromptVersionSentenceFeedbackV1 = "sentence-feedback-v1"
	SchemaVersionFeedbackV1         = "feedback-schema-v1"
)

// Sources where a learner sentence may originate.
const (
	SourceWordDetail   = "word_detail"
	SourceReview       = "review"
	SourceDailyMission = "daily_mission"
	SourceFreePractice = "free_practice"
)

// Provider identifiers used in provider metadata and telemetry.
const (
	ProviderMock     = "mock"
	ProviderOpenCode = "opencode"
)

// Common errors surfaced by provider boundaries.
var (
	ErrMissingLearnerSentence = errors.New("provider task missing learner sentence")
	ErrProviderRefusal        = errors.New("provider refused to generate feedback")
)

// SentenceFeedbackResult is the public API contract for a sentence-feedback
// response (DOC-09 §9). It contains only the fields the frontend is permitted
// to display.
type SentenceFeedbackResult struct {
	SentenceID        uuid.UUID
	AttemptID         uuid.UUID
	Status            string
	OriginalSentence  string
	CorrectedSentence *string
	Explanation       string
	ImprovementTip    *string
	MissionCompleted  bool
	CanRetry          bool
	Reported          bool
}

// ProviderTask is the provider-neutral input built by the backend (DOC-09 §14).
// It separates the system prompt, developer prompt, and user task payload so
// learner input is always serialized as data, never concatenated into
// instruction text.
type ProviderTask struct {
	PromptVersion   string
	SchemaVersion   string
	SystemPrompt    string
	DeveloperPrompt string
	UserPayload     map[string]any
	OutputSchema    map[string]any
	MaxOutputTokens int
	Temperature     float64
	EnableWebSearch bool
	EnableTools     bool
	EnableMemory    bool
}

// ProviderFeedback is the structured, validated output returned by a feedback
// provider (DOC-09 §10).
type ProviderFeedback struct {
	Status                  string
	TargetWordUsedCorrectly bool
	CorrectedSentence       *string
	Explanation             string
	ImprovementTip          *string
	RawJSON                 map[string]any
}

// ModerationInput is the request scoped for a moderation provider.
type ModerationInput struct {
	SentenceText string
	TargetWord   string
	LearnerLevel string
}

// ModerationResult is the internal safety outcome. The Outcome field is one of
// the Safety* constants.
type ModerationResult struct {
	Outcome string
	Reason  string
}

// FeedbackProvider is the narrow boundary to a feedback-model adapter. Provider
// SDK types must stay inside the concrete adapter implementation.
type FeedbackProvider interface {
	GenerateFeedback(ctx context.Context, task ProviderTask) (*ProviderFeedback, error)
}

// ModerationProvider is the narrow boundary to a content-moderation adapter.
// Provider SDK types must stay inside the concrete adapter implementation.
type ModerationProvider interface {
	Classify(ctx context.Context, input ModerationInput) (*ModerationResult, error)
}

// MockProvider is a deterministic, schema-valid provider for tests and CI.
// It never calls an external service.
type MockProvider struct{}

// NewMockProvider returns a deterministic mock provider.
func NewMockProvider() *MockProvider {
	return &MockProvider{}
}

// GenerateFeedback returns a deterministic result based on the user payload.
// It treats the sentence as correct if it contains the target word (case-
// insensitive), otherwise incorrect with a minimal correction. Injection-style
// instructions are graded as text, never followed.
func (m *MockProvider) GenerateFeedback(ctx context.Context, task ProviderTask) (*ProviderFeedback, error) {
	payload, ok := task.UserPayload["learner_sentence"].(string)
	if !ok || strings.TrimSpace(payload) == "" {
		return nil, ErrMissingLearnerSentence
	}
	target, _ := task.UserPayload["target_word"].(string)

	normalized := strings.ToLower(strings.TrimSpace(payload))
	hasTarget := target != "" && strings.Contains(normalized, strings.ToLower(target))

	if hasTarget {
		return &ProviderFeedback{
			Status:                  LearningStatusCorrect,
			TargetWordUsedCorrectly: true,
			Explanation:             "The sentence uses the target word correctly.",
			RawJSON: map[string]any{
				"status":                     LearningStatusCorrect,
				"target_word_used_correctly": true,
			},
		}, nil
	}

	corrected := payload + " (target word missing)"
	return &ProviderFeedback{
		Status:                  LearningStatusIncorrect,
		TargetWordUsedCorrectly: false,
		CorrectedSentence:       &corrected,
		Explanation:             "The sentence does not include the target word.",
		ImprovementTip:          stringPtr("Try using the target word in your sentence."),
		RawJSON: map[string]any{
			"status":                     LearningStatusIncorrect,
			"target_word_used_correctly": false,
		},
	}, nil
}

// Classify returns a deterministic moderation outcome. It maps a small set of
// test markers to safety outcomes so the safety mapping in T03 can be exercised
// without a real provider.
func (m *MockProvider) Classify(ctx context.Context, input ModerationInput) (*ModerationResult, error) {
	normalized := strings.ToLower(strings.TrimSpace(input.SentenceText))
	if strings.Contains(normalized, "self-harm") || strings.Contains(normalized, "kill myself") {
		return &ModerationResult{Outcome: SafetySelfHarmIntervention, Reason: "test self-harm marker"}, nil
	}
	if strings.Contains(normalized, "blocked") || strings.Contains(normalized, "weapon") {
		return &ModerationResult{Outcome: SafetyBlocked, Reason: "test blocked marker"}, nil
	}
	if strings.Contains(normalized, "sensitive") {
		return &ModerationResult{Outcome: SafetyAllowedSensitive, Reason: "test sensitive marker"}, nil
	}
	return &ModerationResult{Outcome: SafetyAllowed, Reason: "mock allowed"}, nil
}

func stringPtr(s string) *string { return &s }
