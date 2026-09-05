package reviews

import (
	"testing"
	"time"

	"github.com/KARSIFT/vocanova-platform/apps/api/business/learning"
	"github.com/KARSIFT/vocanova-platform/apps/api/foundation/clock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestService(data MemoryRepositoryData) (*Service, *MemoryRepository) {
	repo := NewMemoryRepository(data)
	return NewService(repo, learning.NewMemoryIdempotencyStore(), nil), repo
}

func TestSubmitReviewValidatesRequiredFields(t *testing.T) {
	svc, _ := newTestService(MemoryRepositoryData{})
	userID := uuid.MustParse("00000000-0000-0000-0000-000000000001")

	_, err := svc.SubmitReview(t.Context(), SubmitReviewRequest{
		UserID:          userID,
		UserWordID:      uuid.MustParse("00000000-0000-0000-0000-000000000002"),
		MeaningID:       uuid.MustParse("00000000-0000-0000-0000-000000000003"),
		AttemptType:     AttemptTypeReview,
		PromptType:      PromptTypeMultipleChoice,
		Result:          ResultCorrect,
		Rating:          RatingGood,
		AnsweredAt:      time.Now().UTC(),
		ClientAttemptID: "ca",
	})
	assert.ErrorIs(t, err, ErrIdempotencyKeyRequired)
}

func TestSubmitReviewValidatesPromptType(t *testing.T) {
	svc, _ := newTestService(MemoryRepositoryData{})
	userID := uuid.MustParse("00000000-0000-0000-0000-000000000001")

	_, err := svc.SubmitReview(t.Context(), SubmitReviewRequest{
		UserID:          userID,
		UserWordID:      uuid.MustParse("00000000-0000-0000-0000-000000000002"),
		MeaningID:       uuid.MustParse("00000000-0000-0000-0000-000000000003"),
		AttemptType:     AttemptTypeReview,
		PromptType:      "typing",
		Result:          ResultCorrect,
		Rating:          RatingGood,
		AnsweredAt:      time.Now().UTC(),
		ClientAttemptID: "ca",
		IdempotencyKey:  "idem",
	})
	assert.ErrorIs(t, err, ErrInvalidPromptType)
}

func TestSubmitReviewInvalidRatingForIncorrect(t *testing.T) {
	svc, _ := newTestService(MemoryRepositoryData{})
	userID := uuid.MustParse("00000000-0000-0000-0000-000000000001")

	_, err := svc.SubmitReview(t.Context(), SubmitReviewRequest{
		UserID:          userID,
		UserWordID:      uuid.MustParse("00000000-0000-0000-0000-000000000002"),
		MeaningID:       uuid.MustParse("00000000-0000-0000-0000-000000000003"),
		AttemptType:     AttemptTypeReview,
		PromptType:      PromptTypeMultipleChoice,
		Result:          ResultIncorrect,
		Rating:          RatingGood,
		AnsweredAt:      time.Now().UTC(),
		ClientAttemptID: "ca",
		IdempotencyKey:  "idem",
	})
	assert.ErrorIs(t, err, ErrInvalidRatingForResult)
}

func TestSubmitReviewSchedulesAndUpdatesCounters(t *testing.T) {
	userID := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	wordID := uuid.MustParse("00000000-0000-0000-0000-000000000002")
	meaningID := uuid.MustParse("00000000-0000-0000-0000-000000000003")
	userWordID := uuid.MustParse("00000000-0000-0000-0000-000000000004")

	repo := NewMemoryRepository(MemoryRepositoryData{
		Words: []MemoryWord{{ID: wordID, Text: "boarding pass", NormalizedText: "boarding pass", Status: "active"}},
		Meanings: []MemoryMeaning{
			{ID: meaningID, WordID: wordID, PartOfSpeech: "noun", ShortDefinition: "A document.", Status: "active"},
		},
		UserWords: []MemoryUserWord{
			{ID: userWordID, UserID: userID, MeaningID: meaningID, Status: "new", Source: "journey", ReviewStep: 0},
		},
	})
	answeredAt := time.Date(2026, 7, 25, 12, 0, 0, 0, time.UTC)
	svc := NewService(repo, learning.NewMemoryIdempotencyStore(), clock.Fixed{T: answeredAt})
	selectedOptionMeaningID := meaningID
	attempt, err := svc.SubmitReview(t.Context(), SubmitReviewRequest{
		UserID:                  userID,
		UserWordID:              userWordID,
		MeaningID:               meaningID,
		AttemptType:             AttemptTypeReview,
		PromptType:              PromptTypeMultipleChoice,
		Result:                  ResultCorrect,
		Rating:                  RatingGood,
		SelectedOptionMeaningID: &selectedOptionMeaningID,
		AnsweredAt:              answeredAt,
		ClientAttemptID:         "ca-1",
		IdempotencyKey:          "idem-1",
	})
	require.NoError(t, err)
	assert.Equal(t, 1, attempt.ReviewStepAfter)
	assert.Equal(t, 1.0, attempt.NextReviewAt.Sub(answeredAt).Hours())
	assert.Equal(t, 1, repo.userWords[0].TotalReviewCount)
	assert.Equal(t, 1, repo.userWords[0].CorrectReviewCount)
	assert.Equal(t, 1, repo.userWords[0].ConsecutiveCorrectCount)
}

func TestSubmitReviewAnchorsScheduleToServerClock(t *testing.T) {
	userID := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	wordID := uuid.MustParse("00000000-0000-0000-0000-000000000002")
	meaningID := uuid.MustParse("00000000-0000-0000-0000-000000000003")
	userWordID := uuid.MustParse("00000000-0000-0000-0000-000000000004")
	serverNow := time.Date(2026, 9, 5, 12, 0, 0, 0, time.UTC)
	clientAnsweredAt := serverNow.Add(-30 * 24 * time.Hour)

	repo := NewMemoryRepository(MemoryRepositoryData{
		Words:     []MemoryWord{{ID: wordID, Text: "boarding pass", NormalizedText: "boarding pass", Status: "active"}},
		Meanings:  []MemoryMeaning{{ID: meaningID, WordID: wordID, PartOfSpeech: "noun", ShortDefinition: "A document.", Status: "active"}},
		UserWords: []MemoryUserWord{{ID: userWordID, UserID: userID, MeaningID: meaningID, Status: "new", Source: "journey", ReviewStep: 0}},
	})
	svc := NewService(repo, learning.NewMemoryIdempotencyStore(), clock.Fixed{T: serverNow})
	selectedOptionMeaningID := meaningID

	attempt, err := svc.SubmitReview(t.Context(), SubmitReviewRequest{
		UserID: userID, UserWordID: userWordID, MeaningID: meaningID,
		AttemptType: AttemptTypeReview, PromptType: PromptTypeMultipleChoice,
		Result: ResultCorrect, Rating: RatingGood, SelectedOptionMeaningID: &selectedOptionMeaningID,
		AnsweredAt: clientAnsweredAt, ClientAttemptID: "server-clock", IdempotencyKey: "server-clock",
	})
	require.NoError(t, err)
	assert.Equal(t, clientAnsweredAt, attempt.AnsweredAt, "attempt history retains the client-recorded answer time")
	assert.Equal(t, serverNow.Add(time.Hour), attempt.NextReviewAt, "the client timestamp cannot manipulate the schedule")
	assert.Equal(t, serverNow.Add(time.Hour), *repo.userWords[0].NextReviewAt)
}

func TestSubmitReviewRejectsForgedMultipleChoiceCorrectResult(t *testing.T) {
	userID := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	wordID := uuid.MustParse("00000000-0000-0000-0000-000000000002")
	meaningID := uuid.MustParse("00000000-0000-0000-0000-000000000003")
	userWordID := uuid.MustParse("00000000-0000-0000-0000-000000000004")
	wrongMeaningID := uuid.MustParse("00000000-0000-0000-0000-000000000005")
	repo := NewMemoryRepository(MemoryRepositoryData{
		Words:     []MemoryWord{{ID: wordID, Text: "boarding pass", NormalizedText: "boarding pass", Status: "active"}},
		Meanings:  []MemoryMeaning{{ID: meaningID, WordID: wordID, PartOfSpeech: "noun", ShortDefinition: "A document.", Status: "active"}},
		UserWords: []MemoryUserWord{{ID: userWordID, UserID: userID, MeaningID: meaningID, Status: "new", Source: "journey", ReviewStep: 0}},
	})
	svc := NewService(repo, learning.NewMemoryIdempotencyStore(), nil)

	_, err := svc.SubmitReview(t.Context(), SubmitReviewRequest{
		UserID: userID, UserWordID: userWordID, MeaningID: meaningID,
		AttemptType: AttemptTypeReview, PromptType: PromptTypeMultipleChoice,
		Result: ResultCorrect, Rating: RatingGood, SelectedOptionMeaningID: &wrongMeaningID,
		AnsweredAt:      time.Date(2026, 7, 25, 12, 0, 0, 0, time.UTC),
		ClientAttemptID: "forged-correct", IdempotencyKey: "forged-correct",
	})
	require.ErrorIs(t, err, ErrMultipleChoiceResultMismatch)
	assert.Empty(t, repo.attempts)
	assert.Equal(t, 0, repo.userWords[0].ReviewStep)
	assert.Zero(t, repo.userWords[0].TotalReviewCount)
}

func TestSubmitReviewIdempotencyReplay(t *testing.T) {
	userID := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	wordID := uuid.MustParse("00000000-0000-0000-0000-000000000002")
	meaningID := uuid.MustParse("00000000-0000-0000-0000-000000000003")
	userWordID := uuid.MustParse("00000000-0000-0000-0000-000000000004")

	repo := NewMemoryRepository(MemoryRepositoryData{
		Words: []MemoryWord{{ID: wordID, Text: "boarding pass", NormalizedText: "boarding pass", Status: "active"}},
		Meanings: []MemoryMeaning{
			{ID: meaningID, WordID: wordID, PartOfSpeech: "noun", ShortDefinition: "A document.", Status: "active"},
		},
		UserWords: []MemoryUserWord{
			{ID: userWordID, UserID: userID, MeaningID: meaningID, Status: "new", Source: "journey", ReviewStep: 0},
		},
	})
	svc := NewService(repo, learning.NewMemoryIdempotencyStore(), clock.Real{})
	selectedOptionMeaningID := meaningID

	req := SubmitReviewRequest{
		UserID:                  userID,
		UserWordID:              userWordID,
		MeaningID:               meaningID,
		AttemptType:             AttemptTypeReview,
		PromptType:              PromptTypeMultipleChoice,
		Result:                  ResultCorrect,
		Rating:                  RatingGood,
		SelectedOptionMeaningID: &selectedOptionMeaningID,
		AnsweredAt:              time.Date(2026, 7, 25, 12, 0, 0, 0, time.UTC),
		ClientAttemptID:         "ca-1",
		IdempotencyKey:          "idem-1",
	}
	first, err := svc.SubmitReview(t.Context(), req)
	require.NoError(t, err)
	second, err := svc.SubmitReview(t.Context(), req)
	require.NoError(t, err)
	assert.Equal(t, first.ID, second.ID)
	assert.Equal(t, 1, repo.userWords[0].TotalReviewCount)
}

func TestSubmitReviewIdempotencyKeyConflict(t *testing.T) {
	userID := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	wordID := uuid.MustParse("00000000-0000-0000-0000-000000000002")
	meaningID := uuid.MustParse("00000000-0000-0000-0000-000000000003")
	userWordID := uuid.MustParse("00000000-0000-0000-0000-000000000004")

	repo := NewMemoryRepository(MemoryRepositoryData{
		Words: []MemoryWord{{ID: wordID, Text: "boarding pass", NormalizedText: "boarding pass", Status: "active"}},
		Meanings: []MemoryMeaning{
			{ID: meaningID, WordID: wordID, PartOfSpeech: "noun", ShortDefinition: "A document.", Status: "active"},
		},
		UserWords: []MemoryUserWord{
			{ID: userWordID, UserID: userID, MeaningID: meaningID, Status: "new", Source: "journey", ReviewStep: 0},
		},
	})
	svc := NewService(repo, learning.NewMemoryIdempotencyStore(), nil)
	selectedOptionMeaningID := meaningID

	base := SubmitReviewRequest{
		UserID:                  userID,
		UserWordID:              userWordID,
		MeaningID:               meaningID,
		AttemptType:             AttemptTypeReview,
		PromptType:              PromptTypeMultipleChoice,
		Result:                  ResultCorrect,
		Rating:                  RatingGood,
		SelectedOptionMeaningID: &selectedOptionMeaningID,
		AnsweredAt:              time.Date(2026, 7, 25, 12, 0, 0, 0, time.UTC),
		ClientAttemptID:         "ca-1",
		IdempotencyKey:          "shared-key",
	}
	_, err := svc.SubmitReview(t.Context(), base)
	require.NoError(t, err)

	changed := base
	changed.Rating = RatingEasy
	_, err = svc.SubmitReview(t.Context(), changed)
	assert.ErrorIs(t, err, ErrIdempotencyConflict)
}

func TestSubmitReviewCrossUserIdempotentKey(t *testing.T) {
	userA := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	userB := uuid.MustParse("00000000-0000-0000-0000-000000000002")
	wordID := uuid.MustParse("00000000-0000-0000-0000-000000000003")
	meaningID := uuid.MustParse("00000000-0000-0000-0000-000000000004")
	userWordA := uuid.MustParse("00000000-0000-0000-0000-000000000005")
	userWordB := uuid.MustParse("00000000-0000-0000-0000-000000000006")

	repo := NewMemoryRepository(MemoryRepositoryData{
		Words: []MemoryWord{{ID: wordID, Text: "boarding pass", NormalizedText: "boarding pass", Status: "active"}},
		Meanings: []MemoryMeaning{
			{ID: meaningID, WordID: wordID, PartOfSpeech: "noun", ShortDefinition: "A document.", Status: "active"},
		},
		UserWords: []MemoryUserWord{
			{ID: userWordA, UserID: userA, MeaningID: meaningID, Status: "new", Source: "journey", ReviewStep: 0},
			{ID: userWordB, UserID: userB, MeaningID: meaningID, Status: "new", Source: "journey", ReviewStep: 0},
		},
	})
	svc := NewService(repo, learning.NewMemoryIdempotencyStore(), nil)
	selectedOptionMeaningID := meaningID

	base := SubmitReviewRequest{
		AttemptType:             AttemptTypeReview,
		PromptType:              PromptTypeMultipleChoice,
		Result:                  ResultCorrect,
		Rating:                  RatingGood,
		SelectedOptionMeaningID: &selectedOptionMeaningID,
		AnsweredAt:              time.Date(2026, 7, 25, 12, 0, 0, 0, time.UTC),
		ClientAttemptID:         "shared-ca",
		IdempotencyKey:          "shared-key",
	}
	_, err := svc.SubmitReview(t.Context(), SubmitReviewRequest{
		UserID: userA, UserWordID: userWordA, MeaningID: meaningID,
		AttemptType: base.AttemptType, PromptType: base.PromptType, Result: base.Result,
		Rating: base.Rating, SelectedOptionMeaningID: base.SelectedOptionMeaningID, AnsweredAt: base.AnsweredAt, ClientAttemptID: base.ClientAttemptID,
		IdempotencyKey: base.IdempotencyKey,
	})
	require.NoError(t, err)
	_, err = svc.SubmitReview(t.Context(), SubmitReviewRequest{
		UserID: userB, UserWordID: userWordB, MeaningID: meaningID,
		AttemptType: base.AttemptType, PromptType: base.PromptType, Result: base.Result,
		Rating: base.Rating, SelectedOptionMeaningID: base.SelectedOptionMeaningID, AnsweredAt: base.AnsweredAt, ClientAttemptID: base.ClientAttemptID,
		IdempotencyKey: base.IdempotencyKey,
	})
	require.NoError(t, err)
	assert.Equal(t, 1, repo.userWords[0].TotalReviewCount)
	assert.Equal(t, 1, repo.userWords[1].TotalReviewCount)
}

func TestSubmitReviewTwoConsecutiveIncorrectResetsStep(t *testing.T) {
	userID := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	wordID := uuid.MustParse("00000000-0000-0000-0000-000000000002")
	meaningID := uuid.MustParse("00000000-0000-0000-0000-000000000003")
	userWordID := uuid.MustParse("00000000-0000-0000-0000-000000000004")

	repo := NewMemoryRepository(MemoryRepositoryData{
		Words: []MemoryWord{{ID: wordID, Text: "boarding pass", NormalizedText: "boarding pass", Status: "active"}},
		Meanings: []MemoryMeaning{
			{ID: meaningID, WordID: wordID, PartOfSpeech: "noun", ShortDefinition: "A document.", Status: "active"},
		},
		UserWords: []MemoryUserWord{
			{ID: userWordID, UserID: userID, MeaningID: meaningID, Status: "learning", Source: "journey", ReviewStep: 3},
		},
	})
	svc := NewService(repo, learning.NewMemoryIdempotencyStore(), nil)

	base := SubmitReviewRequest{
		UserID: userID, UserWordID: userWordID, MeaningID: meaningID,
		AttemptType: AttemptTypeReview, PromptType: PromptTypeSelfCheck,
		Result: ResultIncorrect, Rating: RatingAgain,
		AnsweredAt: time.Date(2026, 7, 25, 12, 0, 0, 0, time.UTC),
	}
	_, err := svc.SubmitReview(t.Context(), SubmitReviewRequest{
		UserID: base.UserID, UserWordID: base.UserWordID, MeaningID: base.MeaningID,
		AttemptType: base.AttemptType, PromptType: base.PromptType, Result: base.Result,
		Rating: base.Rating, AnsweredAt: base.AnsweredAt, ClientAttemptID: "first", IdempotencyKey: "i1",
	})
	require.NoError(t, err)
	assert.Equal(t, 2, repo.userWords[0].ReviewStep)

	_, err = svc.SubmitReview(t.Context(), SubmitReviewRequest{
		UserID: base.UserID, UserWordID: base.UserWordID, MeaningID: base.MeaningID,
		AttemptType: base.AttemptType, PromptType: base.PromptType, Result: base.Result,
		Rating: base.Rating, AnsweredAt: base.AnsweredAt, ClientAttemptID: "second", IdempotencyKey: "i2",
	})
	require.NoError(t, err)
	assert.Equal(t, 0, repo.userWords[0].ReviewStep)
	assert.Equal(t, 2, repo.userWords[0].ConsecutiveIncorrectCount)
}
