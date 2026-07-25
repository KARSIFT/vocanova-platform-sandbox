package aifeedback

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/KARSIFT/vocanova-platform/apps/api/business/learning"
	"github.com/KARSIFT/vocanova-platform/apps/api/foundation/clock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type serviceFixture struct {
	userID          uuid.UUID
	userWordID      uuid.UUID
	reviewAttemptID uuid.UUID
	meaningID       uuid.UUID
	wordID          uuid.UUID
	repo            *MemoryRepository
	provider        *countingProvider
	service         *Service
}

func newServiceFixture(t *testing.T) *serviceFixture {
	userID := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	userWordID := uuid.MustParse("00000000-0000-0000-0000-000000000002")
	reviewAttemptID := uuid.MustParse("00000000-0000-0000-0000-000000000005")
	meaningID := uuid.MustParse("00000000-0000-0000-0000-000000000003")
	wordID := uuid.MustParse("00000000-0000-0000-0000-000000000004")

	repo := NewMemoryRepository(MemoryRepositoryData{
		UserWords: []MemoryUserWord{{
			ID:        userWordID,
			UserID:    userID,
			MeaningID: meaningID,
			Status:    "learning",
		}},
		ReviewAttempts: []MemoryReviewAttempt{{
			ID:         reviewAttemptID,
			UserID:     userID,
			UserWordID: userWordID,
			MeaningID:  meaningID,
		}},
		Meanings: []MemoryMeaning{{
			ID:              meaningID,
			WordID:          wordID,
			PartOfSpeech:    "verb",
			ShortDefinition: "to do a job or task",
			Status:          "active",
		}},
		Words: []MemoryWord{{
			ID:              wordID,
			Text:            "work",
			NormalizedText:  "work",
			WordType:        "word",
			DifficultyLevel: "a2",
			Status:          "active",
		}},
	})

	mockProvider := NewMockProvider()
	provider := &countingProvider{MockProvider: mockProvider}
	safety := NewProviderSafetyClassifier(mockProvider)
	clock := clock.Fixed{T: time.Date(2026, 7, 25, 12, 0, 0, 0, time.UTC)}
	service := NewService(
		repo,
		provider,
		safety,
		NewMemoryRateLimiter(DefaultRateLimitConfig(), clock),
		learning.NewMemoryIdempotencyStore(),
		NewStubMissionUpdater(),
		NewNoopTelemetryRecorder(),
		NewDefaultTaskBuilder(),
		NewDefaultOutputValidator(),
		clock,
		DefaultServiceConfig(),
	)

	return &serviceFixture{
		userID:          userID,
		userWordID:      userWordID,
		reviewAttemptID: reviewAttemptID,
		meaningID:       meaningID,
		wordID:          wordID,
		repo:            repo,
		provider:        provider,
		service:         service,
	}
}

type countingProvider struct {
	*MockProvider
	calls int
}

func (c *countingProvider) GenerateFeedback(ctx context.Context, task ProviderTask) (*ProviderFeedback, error) {
	c.calls++
	return c.MockProvider.GenerateFeedback(ctx, task)
}

func (f *serviceFixture) request(sentence string) SubmitSentenceFeedbackRequest {
	return SubmitSentenceFeedbackRequest{
		UserID:         f.userID,
		SentenceText:   sentence,
		Source:         SourceWordDetail,
		AttemptID:      f.userWordID,
		IdempotencyKey: uuid.New().String(),
	}
}

func TestServiceSubmitSentenceFeedbackSuccess(t *testing.T) {
	f := newServiceFixture(t)
	req := f.request("I work every day.")

	result, err := f.service.SubmitSentenceFeedback(t.Context(), req)
	require.NoError(t, err)
	assert.Equal(t, LearningStatusCorrect, result.Status)
	assert.Equal(t, req.SentenceText, result.OriginalSentence)
	assert.False(t, result.MissionCompleted)
	assert.False(t, result.CanRetry)
	assert.Equal(t, 1, f.provider.calls)
}

func TestServiceValidationFailureTooShort(t *testing.T) {
	f := newServiceFixture(t)
	req := f.request("I work")

	result, err := f.service.SubmitSentenceFeedback(t.Context(), req)
	require.NoError(t, err)
	assert.Equal(t, ValidationCodeTooShort, result.ErrorCode)
	assert.True(t, result.CanRetry)
	assert.Equal(t, 0, f.provider.calls)
}

func TestServiceValidationFailureMissingTarget(t *testing.T) {
	f := newServiceFixture(t)
	req := f.request("I play every day with my friends.")

	result, err := f.service.SubmitSentenceFeedback(t.Context(), req)
	require.NoError(t, err)
	assert.Equal(t, ValidationCodeMissingTarget, result.ErrorCode)
	assert.True(t, result.CanRetry)
	assert.Equal(t, 0, f.provider.calls)
}

func TestServiceRateLimitEnforcesPerMinuteCeiling(t *testing.T) {
	f := newServiceFixture(t)

	sentences := []string{
		"I work every day.",
		"I work hard every day.",
		"She works in the city.",
		"We worked yesterday.",
		"They are working now.",
	}

	for _, sentence := range sentences {
		req := f.request(sentence)
		_, err := f.service.SubmitSentenceFeedback(t.Context(), req)
		require.NoError(t, err)
	}
	assert.Equal(t, 5, f.provider.calls)

	result, err := f.service.SubmitSentenceFeedback(t.Context(), f.request("I will work tomorrow."))
	require.NoError(t, err)
	assert.Equal(t, ErrorCodeRateLimited, result.ErrorCode)
	assert.True(t, result.CanRetry)
	assert.Equal(t, 5, f.provider.calls)
}

func TestServiceSafetyBlocked(t *testing.T) {
	f := newServiceFixture(t)
	req := f.request("I work on blocked things every day.")

	result, err := f.service.SubmitSentenceFeedback(t.Context(), req)
	require.NoError(t, err)
	assert.Equal(t, ErrorCodeSafetyBlocked, result.ErrorCode)
	assert.False(t, result.CanRetry)
	assert.Equal(t, 0, f.provider.calls)
}

func TestServiceSafetySelfHarm(t *testing.T) {
	f := newServiceFixture(t)
	req := f.request("I work with people who self-harm.")

	result, err := f.service.SubmitSentenceFeedback(t.Context(), req)
	require.NoError(t, err)
	assert.Equal(t, ErrorCodeSafetySelfHarm, result.ErrorCode)
	assert.False(t, result.CanRetry)
	assert.Equal(t, 0, f.provider.calls)
}

func TestServiceSafetyModerationUnavailable(t *testing.T) {
	f := newServiceFixture(t)
	f.service.safety = &unavailableSafetyClassifier{}
	req := f.request("I work every day.")

	result, err := f.service.SubmitSentenceFeedback(t.Context(), req)
	require.NoError(t, err)
	assert.Equal(t, ErrorCodeSafetyModerationUnavailable, result.ErrorCode)
	assert.True(t, result.CanRetry)
	assert.Equal(t, 0, f.provider.calls)
}

type unavailableSafetyClassifier struct{}

func (u *unavailableSafetyClassifier) Classify(ctx context.Context, input ModerationInput) (*ModerationResult, error) {
	return &ModerationResult{Outcome: SafetyModerationUnavailable}, nil
}

func TestServiceDedupPreventsDuplicateProviderCall(t *testing.T) {
	f := newServiceFixture(t)
	key := uuid.New().String()
	req := SubmitSentenceFeedbackRequest{
		UserID:         f.userID,
		SentenceText:   "I work every day.",
		Source:         SourceWordDetail,
		AttemptID:      f.userWordID,
		IdempotencyKey: key,
	}

	result1, err := f.service.SubmitSentenceFeedback(t.Context(), req)
	require.NoError(t, err)
	assert.Equal(t, LearningStatusCorrect, result1.Status)
	assert.Equal(t, 1, f.provider.calls)

	result2, err := f.service.SubmitSentenceFeedback(t.Context(), req)
	require.NoError(t, err)
	assert.Equal(t, result1.Status, result2.Status)
	assert.Equal(t, result1.SentenceID, result2.SentenceID)
	assert.Equal(t, result1.AttemptID, result2.AttemptID)
	assert.Equal(t, 1, f.provider.calls)
}

func TestServiceReviewSourceSuccess(t *testing.T) {
	f := newServiceFixture(t)
	req := SubmitSentenceFeedbackRequest{
		UserID:         f.userID,
		SentenceText:   "I work every day.",
		Source:         SourceReview,
		AttemptID:      f.reviewAttemptID,
		IdempotencyKey: uuid.New().String(),
	}

	result, err := f.service.SubmitSentenceFeedback(t.Context(), req)
	require.NoError(t, err)
	assert.Equal(t, LearningStatusCorrect, result.Status)
	assert.False(t, result.MissionCompleted)
	assert.Equal(t, 1, f.provider.calls)
}

func TestServiceDailyMissionNotEligible(t *testing.T) {
	f := newServiceFixture(t)
	req := SubmitSentenceFeedbackRequest{
		UserID:         f.userID,
		SentenceText:   "I work every day.",
		Source:         SourceDailyMission,
		AttemptID:      f.userWordID,
		IdempotencyKey: uuid.New().String(),
	}

	result, err := f.service.SubmitSentenceFeedback(t.Context(), req)
	require.NoError(t, err)
	assert.Equal(t, ValidationCodeAttemptNotEligible, result.ErrorCode)
	assert.Equal(t, 0, f.provider.calls)
}

func TestServiceProviderFailureIsRetryable(t *testing.T) {
	f := newServiceFixture(t)
	f.service.provider = &failingProvider{}
	req := f.request("I work every day.")

	result, err := f.service.SubmitSentenceFeedback(t.Context(), req)
	require.NoError(t, err)
	assert.Equal(t, ErrorCodeTemporaryFailure, result.ErrorCode)
	assert.True(t, result.CanRetry)
	assert.NotEqual(t, uuid.Nil, result.SentenceID)
	assert.NotEqual(t, uuid.Nil, result.AttemptID)
}

type failingProvider struct{}

func (f *failingProvider) GenerateFeedback(ctx context.Context, task ProviderTask) (*ProviderFeedback, error) {
	return nil, errors.New("provider refused")
}

func TestServiceMissionStubIsNotWired(t *testing.T) {
	f := newServiceFixture(t)
	req := f.request("I work every day.")

	result, err := f.service.SubmitSentenceFeedback(t.Context(), req)
	require.NoError(t, err)
	assert.False(t, result.MissionCompleted)
}

func TestServiceIdempotencyKeyConflict(t *testing.T) {
	f := newServiceFixture(t)
	key := uuid.New().String()
	req1 := SubmitSentenceFeedbackRequest{
		UserID:         f.userID,
		SentenceText:   "I work every day.",
		Source:         SourceWordDetail,
		AttemptID:      f.userWordID,
		IdempotencyKey: key,
	}
	req2 := SubmitSentenceFeedbackRequest{
		UserID:         f.userID,
		SentenceText:   "I work hard every day.",
		Source:         SourceWordDetail,
		AttemptID:      f.userWordID,
		IdempotencyKey: key,
	}

	_, err := f.service.SubmitSentenceFeedback(t.Context(), req1)
	require.NoError(t, err)

	result, err := f.service.SubmitSentenceFeedback(t.Context(), req2)
	require.NoError(t, err)
	assert.Equal(t, ErrorCodeIdempotencyConflict, result.ErrorCode)
	assert.False(t, result.CanRetry)
}

func TestServiceRepairsInvalidOutput(t *testing.T) {
	f := newServiceFixture(t)
	f.service.provider = &invalidThenValidProvider{valid: NewMockProvider()}
	req := f.request("I work every day.")

	result, err := f.service.SubmitSentenceFeedback(t.Context(), req)
	require.NoError(t, err)
	assert.Equal(t, LearningStatusCorrect, result.Status)
	assert.Equal(t, 2, f.service.provider.(*invalidThenValidProvider).calls)
}

func TestServiceRepairFailureIsRetryable(t *testing.T) {
	f := newServiceFixture(t)
	f.service.provider = &invalidProvider{}
	req := f.request("I work every day.")

	result, err := f.service.SubmitSentenceFeedback(t.Context(), req)
	require.NoError(t, err)
	assert.Equal(t, ErrorCodeTemporaryFailure, result.ErrorCode)
	assert.True(t, result.CanRetry)
	assert.NotEqual(t, uuid.Nil, result.SentenceID)
	assert.NotEqual(t, uuid.Nil, result.AttemptID)
}

func TestServiceRecordsConfiguredProviderOnAttempt(t *testing.T) {
	f := newServiceFixture(t)
	f.service.config.Provider = ProviderOpenCode
	f.service.config.Model = DefaultOpenCodeModel
	req := f.request("I work every day.")

	result, err := f.service.SubmitSentenceFeedback(t.Context(), req)
	require.NoError(t, err)
	assert.Equal(t, LearningStatusCorrect, result.Status)

	assert.Len(t, f.repo.attempts, 1)
	assert.Equal(t, ProviderOpenCode, f.repo.attempts[0].Provider)
	assert.Equal(t, DefaultOpenCodeModel, f.repo.attempts[0].Model)
}

func TestOpenCodeServiceConfigWiresProviderAndModel(t *testing.T) {
	cfg := OpenCodeServiceConfig()
	assert.Equal(t, ProviderOpenCode, cfg.Provider)
	assert.Equal(t, DefaultOpenCodeModel, cfg.Model)
	assert.Equal(t, DefaultOpenCodeModel, cfg.OpenCode.Model)
	assert.Equal(t, DefaultOpenCodeConfig().BaseURL, cfg.OpenCode.BaseURL)
	assert.Empty(t, cfg.OpenCode.APIKey)
}

type invalidThenValidProvider struct {
	calls int
	valid FeedbackProvider
}

func (p *invalidThenValidProvider) GenerateFeedback(ctx context.Context, task ProviderTask) (*ProviderFeedback, error) {
	p.calls++
	if p.calls == 1 {
		return &ProviderFeedback{
			Status:                  LearningStatusCorrect,
			TargetWordUsedCorrectly: false,
			Explanation:             "Looks good.",
			RawJSON: map[string]any{
				"status":                     LearningStatusCorrect,
				"target_word_used_correctly": false,
			},
		}, nil
	}
	return p.valid.GenerateFeedback(ctx, task)
}

type invalidProvider struct{}

func (p *invalidProvider) GenerateFeedback(ctx context.Context, task ProviderTask) (*ProviderFeedback, error) {
	return &ProviderFeedback{
		Status:                  "not_a_status",
		TargetWordUsedCorrectly: false,
		Explanation:             "Invalid.",
		RawJSON: map[string]any{
			"status": "not_a_status",
		},
	}, nil
}
