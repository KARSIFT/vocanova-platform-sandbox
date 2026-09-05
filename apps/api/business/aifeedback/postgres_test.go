package aifeedback

import (
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPostgreSQLRepositoryLoadTargetFromUserWord(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewPostgreSQLRepository(db, nil)
	userID := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	userWordID := uuid.MustParse("00000000-0000-0000-0000-000000000002")
	meaningID := uuid.MustParse("00000000-0000-0000-0000-000000000003")
	wordID := uuid.MustParse("00000000-0000-0000-0000-000000000004")

	mock.ExpectQuery("SELECT cw.id, cw.text, cw.normalized_text, cw.word_type, cw.difficulty_level").
		WithArgs(userWordID, userID).
		WillReturnRows(sqlmock.NewRows([]string{"cw.id", "cw.text", "cw.normalized_text", "cw.word_type", "cw.difficulty_level", "wm.id", "wm.part_of_speech", "wm.short_definition", "uw.id"}).
			AddRow(wordID, "work", "work", "word", "a2", meaningID, "verb", "to do a job", userWordID))

	target, err := repo.LoadTarget(t.Context(), LoadTargetRequest{UserID: userID, Source: SourceWordDetail, AttemptID: userWordID})
	require.NoError(t, err)
	assert.Equal(t, "work", target.NormalizedWord)
	assert.Equal(t, "verb", target.PartOfSpeech)
	assert.Equal(t, "a2", target.LearnerLevel)
	assert.Contains(t, target.AcceptedForms, "worked")
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestPostgreSQLRepositoryLoadTargetNotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewPostgreSQLRepository(db, nil)
	userID := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	userWordID := uuid.MustParse("00000000-0000-0000-0000-000000000002")

	mock.ExpectQuery("SELECT cw.id").
		WithArgs(userWordID, userID).
		WillReturnRows(sqlmock.NewRows([]string{"cw.id"}))

	_, err = repo.LoadTarget(t.Context(), LoadTargetRequest{UserID: userID, Source: SourceWordDetail, AttemptID: userWordID})
	assert.ErrorIs(t, err, ErrTargetNotFound)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestPostgreSQLRepositoryCreatePendingAttempt(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewPostgreSQLRepository(db, nil)
	userID := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	userWordID := uuid.MustParse("00000000-0000-0000-0000-000000000002")
	meaningID := uuid.MustParse("00000000-0000-0000-0000-000000000003")
	wordID := uuid.MustParse("00000000-0000-0000-0000-000000000004")
	now := time.Date(2026, 7, 25, 12, 0, 0, 0, time.UTC)

	target := &Target{
		WordID:         wordID,
		MeaningID:      meaningID,
		UserWordID:     userWordID,
		NormalizedWord: "work",
		WordType:       "word",
		PartOfSpeech:   "verb",
		AcceptedForms:  BuildAcceptedForms("work", "word", "verb"),
	}
	req := SubmitSentenceFeedbackRequest{
		UserID:       userID,
		SentenceText: "I work every day.",
		Source:       SourceWordDetail,
		AttemptID:    userWordID,
	}
	requestHash := RequestHash(userID, userWordID, "work", "i work every day.", PromptVersionSentenceFeedbackV1)

	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO learner_sentences").
		WithArgs(sqlmock.AnyArg(), userID, meaningID, userWordID, req.SentenceText, "i work every day.", SourceWordDetail, SentenceStatusSubmitted, now, now).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("INSERT INTO ai_feedback_attempts").
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), AttemptStatusPending, ProviderMock, "mock", PromptVersionSentenceFeedbackV1, requestHash, now).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	pending, err := repo.CreatePendingAttempt(t.Context(), req, target, "i work every day.", requestHash, ProviderMock, "mock", now)
	require.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, pending.SentenceID)
	assert.NotEqual(t, uuid.Nil, pending.AttemptID)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestPostgreSQLRepositoryCompleteFeedbackAttemptSuccess(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewPostgreSQLRepository(db, nil)
	now := time.Date(2026, 7, 25, 12, 0, 0, 0, time.UTC)
	sentenceID := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	attemptID := uuid.MustParse("00000000-0000-0000-0000-000000000002")
	feedback := &ProviderFeedback{
		Status:                  LearningStatusCorrect,
		TargetWordUsedCorrectly: true,
		Explanation:             "Good.",
		RawJSON: map[string]any{
			"status":                     LearningStatusCorrect,
			"target_word_used_correctly": true,
		},
	}

	mock.ExpectBegin()
	mock.ExpectExec("UPDATE ai_feedback_attempts").
		WithArgs(AttemptStatusSucceeded, sqlmock.AnyArg(), feedback.Explanation, now, now, attemptID, AttemptStatusPending).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("UPDATE learner_sentences").
		WithArgs(SentenceStatusFeedbackReady, now, sentenceID).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	err = repo.CompleteFeedbackAttempt(t.Context(), PendingAttempt{SentenceID: sentenceID, AttemptID: attemptID}, feedback, "", "", now)
	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestPostgreSQLRepositoryCompleteFeedbackAttemptDoesNotClobberSettledAttempt(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewPostgreSQLRepository(db, nil)
	now := time.Date(2026, 7, 25, 12, 0, 0, 0, time.UTC)
	sentenceID := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	attemptID := uuid.MustParse("00000000-0000-0000-0000-000000000002")

	mock.ExpectBegin()
	mock.ExpectExec("UPDATE ai_feedback_attempts").
		WithArgs(AttemptStatusFailed, ErrorCodeTemporaryFailure, "late cleanup", now, now, attemptID, AttemptStatusPending).
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectCommit()

	err = repo.CompleteFeedbackAttempt(t.Context(), PendingAttempt{SentenceID: sentenceID, AttemptID: attemptID}, nil, ErrorCodeTemporaryFailure, "late cleanup", now)
	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestPostgreSQLRepositoryCreateRetryAttempt(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewPostgreSQLRepository(db, nil)
	now := time.Date(2026, 7, 25, 12, 0, 0, 0, time.UTC)
	sentenceID := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	failed := &StoredFeedbackAttempt{
		ID:                uuid.MustParse("00000000-0000-0000-0000-000000000002"),
		LearnerSentenceID: sentenceID,
		Status:            AttemptStatusFailed,
		RequestHash:       "retry-hash",
	}

	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO ai_feedback_attempts").
		WithArgs(sqlmock.AnyArg(), sentenceID, AttemptStatusPending, ProviderMock, "mock", PromptVersionSentenceFeedbackV1, "retry-hash", now).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("UPDATE learner_sentences").
		WithArgs(SentenceStatusSubmitted, now, sentenceID).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	retry, err := repo.CreateRetryAttempt(t.Context(), failed, ProviderMock, "mock", now)
	require.NoError(t, err)
	require.NotNil(t, retry.Pending)
	assert.Nil(t, retry.Existing)
	assert.Equal(t, sentenceID, retry.Pending.SentenceID)
	assert.NotEqual(t, uuid.Nil, retry.Pending.AttemptID)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestPostgreSQLRepositoryCreateRetryAttemptReturnsActiveGenerationAfterConflict(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewPostgreSQLRepository(db, nil)
	now := time.Date(2026, 7, 25, 12, 0, 0, 0, time.UTC)
	sentenceID := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	failed := &StoredFeedbackAttempt{
		ID:                uuid.MustParse("00000000-0000-0000-0000-000000000002"),
		LearnerSentenceID: sentenceID,
		Status:            AttemptStatusFailed,
		RequestHash:       "retry-hash",
	}
	activeID := uuid.MustParse("00000000-0000-0000-0000-000000000003")

	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO ai_feedback_attempts").
		WithArgs(sqlmock.AnyArg(), sentenceID, AttemptStatusPending, ProviderMock, "mock", PromptVersionSentenceFeedbackV1, "retry-hash", now).
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectQuery("SELECT id, learner_sentence_id, status").
		WithArgs("retry-hash").
		WillReturnRows(sqlmock.NewRows([]string{"id", "learner_sentence_id", "status", "provider", "model", "prompt_version", "request_hash", "feedback_json", "feedback_text", "error_code", "error_message", "reported"}).
			AddRow(activeID, sentenceID, AttemptStatusPending, ProviderMock, "mock", PromptVersionSentenceFeedbackV1, "retry-hash", nil, nil, nil, nil, false))
	mock.ExpectCommit()

	retry, err := repo.CreateRetryAttempt(t.Context(), failed, ProviderMock, "mock", now)
	require.NoError(t, err)
	assert.Nil(t, retry.Pending)
	require.NotNil(t, retry.Existing)
	assert.Equal(t, activeID, retry.Existing.ID)
	assert.Equal(t, AttemptStatusPending, retry.Existing.Status)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestPostgreSQLRepositoryGetFeedbackAttemptByRequestHash(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewPostgreSQLRepository(db, nil)
	attemptID := uuid.MustParse("00000000-0000-0000-0000-000000000002")
	sentenceID := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	requestHash := "abc123"

	mock.ExpectQuery("SELECT id, learner_sentence_id").
		WithArgs(requestHash).
		WillReturnRows(sqlmock.NewRows([]string{"id", "learner_sentence_id", "status", "provider", "model", "prompt_version", "request_hash", "feedback_json", "feedback_text", "error_code", "error_message", "reported"}).
			AddRow(attemptID, sentenceID, AttemptStatusSucceeded, ProviderMock, "mock", PromptVersionSentenceFeedbackV1, requestHash, []byte(`{"status":"correct"}`), "Good.", nil, nil, true))

	stored, err := repo.GetFeedbackAttemptByRequestHash(t.Context(), requestHash)
	require.NoError(t, err)
	require.NotNil(t, stored)
	assert.Equal(t, AttemptStatusSucceeded, stored.Status)
	assert.Equal(t, "correct", stored.FeedbackJSON["status"])
	assert.True(t, stored.Reported)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestPostgreSQLRepositoryRequestHashNotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewPostgreSQLRepository(db, nil)
	requestHash := "missing"

	mock.ExpectQuery("SELECT id, learner_sentence_id").
		WithArgs(requestHash).
		WillReturnRows(sqlmock.NewRows([]string{"id"}))

	stored, err := repo.GetFeedbackAttemptByRequestHash(t.Context(), requestHash)
	require.NoError(t, err)
	assert.Nil(t, stored)
	require.NoError(t, mock.ExpectationsWereMet())
}
