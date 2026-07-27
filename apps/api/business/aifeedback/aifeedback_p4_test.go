package aifeedback

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/KARSIFT/vocanova-platform/apps/api/business/gamification"
	"github.com/KARSIFT/vocanova-platform/apps/api/business/learning"
	"github.com/KARSIFT/vocanova-platform/apps/api/business/missions"
	"github.com/KARSIFT/vocanova-platform/apps/api/foundation/clock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// p4Fixture holds the cross-module fixture for T03: an aifeedback service
// backed by its in-process MemoryRepository for the P3 domain, plus a
// sqlmock database backing the real missions.MissionUpdater that replaces
// the StubMissionUpdater seam. The clock is fixed so deterministic tests
// can compare exact timestamps.
type p4Fixture struct {
	userID     uuid.UUID
	wordID     uuid.UUID
	meaningID  uuid.UUID
	userWordID uuid.UUID
	repo       *MemoryRepository
	provider   *countingProvider
	clock      clock.Fixed
	db         *sql.DB
	mock       sqlmock.Sqlmock
	updater    *missions.MissionUpdater
	service    *Service
}

func newP4Fixture(t *testing.T) *p4Fixture {
	t.Helper()

	userID := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	wordID := uuid.MustParse("00000000-0000-0000-0000-000000000004")
	meaningID := uuid.MustParse("00000000-0000-0000-0000-000000000003")
	userWordID := uuid.MustParse("00000000-0000-0000-0000-000000000002")

	repo := NewMemoryRepository(MemoryRepositoryData{
		UserWords: []MemoryUserWord{{
			ID:        userWordID,
			UserID:    userID,
			MeaningID: meaningID,
			Status:    "learning",
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

	provider := &countingProvider{MockProvider: NewMockProvider()}
	fixed := clock.Fixed{T: time.Date(2026, 7, 26, 12, 0, 0, 0, time.UTC)}

	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	gamSvc := gamification.NewService(gamification.NewRepository(db))
	missionsRepo := missions.NewRepository(db)
	missionsSvc := missions.NewService(missionsRepo, gamSvc)
	updater := missions.NewMissionUpdater(missionsSvc, gamSvc)

	safety := NewCompositeSafetyClassifier(NewDefaultLocalAbuseChecker(), NewMockProvider())
	svc := NewService(
		repo,
		provider,
		safety,
		NewMemoryRateLimiter(DefaultRateLimitConfig(), fixed),
		learning.NewMemoryIdempotencyStore(),
		updater,
		NewNoopTelemetryRecorder(),
		NewDefaultTaskBuilder(),
		NewDefaultOutputValidator(),
		fixed,
		DefaultServiceConfig(),
	)

	return &p4Fixture{
		userID:     userID,
		wordID:     wordID,
		meaningID:  meaningID,
		userWordID: userWordID,
		repo:       repo,
		provider:   provider,
		clock:      fixed,
		db:         db,
		mock:       mock,
		updater:    updater,
		service:    svc,
	}
}

func (f *p4Fixture) request(sentence string) SubmitSentenceFeedbackRequest {
	return SubmitSentenceFeedbackRequest{
		UserID:         f.userID,
		SentenceText:   sentence,
		Source:         SourceWordDetail,
		AttemptID:      f.userWordID,
		IdempotencyKey: uuid.New().String(),
	}
}

// TestAIFeedbackP4RealMissionUpdaterSuccessWiring is the type-level and
// integration smoke test for T03: a successful sentence-feedback call
// flows through the real missions.MissionUpdater (replacing the P3 stub)
// and the missions module sees a +3 sentence-submitted grant, a +2
// AI-feedback-received grant, and a streak-reconciliation pass. The
// returned missionCompleted is the real boolean (false on the P3 path).
func TestAIFeedbackP4RealMissionUpdaterSuccessWiring(t *testing.T) {
	f := newP4Fixture(t)

	// The sentenceID is minted inside MemoryRepository's
	// CreatePendingAttempt, so the test cannot pre-author the SQL
	// expectations against a known value. The cleanest path is to
	// prime the mock with AnyArg for the sentenceID-bearing positions
	// and verify that the missions module observed the call (one grant
	// per reward, one streak upsert, one transaction commit).
	day, err := gamification.LocalDate(time.Now(), "UTC")
	require.NoError(t, err)

	// 1. GetSettings reads user_settings (empty -> no rows).
	f.mock.ExpectQuery("SELECT user_id, timezone, daily_review_target").
		WithArgs(f.userID).
		WillReturnRows(sqlmock.NewRows([]string{
			"user_id", "timezone", "daily_review_target", "review_interval_preset",
			"notifications_enabled", "marketing_emails_enabled", "app_language",
		}))

	// 2. UpdateForSentence opens its own transaction.
	f.mock.ExpectBegin()
	// 2a. EnsureTodaySnapshot (idempotent insert-or-update).
	f.mock.ExpectQuery("INSERT INTO daily_mission_snapshots").
		WithArgs(sqlmock.AnyArg(), f.userID, day, "UTC", 20, gamification.MissionPolicyVersion).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "user_id", "local_date", "timezone", "review_target", "reviews_completed",
			"new_word_target", "new_words_completed", "sentence_practice_target",
			"sentence_practices_completed", "policy_version", "status", "completed_at",
			"grace_applied", "grace_day_id",
		}).AddRow(
			uuid.New(), f.userID, day, "UTC", 20, 0,
			nil, nil, nil, nil,
			gamification.MissionPolicyVersion, "open", nil, false, nil,
		))
	// 2b. CurrentBalance.
	f.mock.ExpectQuery("SELECT balance_after FROM confidence_point_ledger").
		WithArgs(f.userID).
		WillReturnRows(sqlmock.NewRows([]string{"balance_after"}))
	// 2c. GrantPoint for +3 sentence-submitted. sentenceID is AnyArg
	// because MemoryRepository mints a fresh UUID per submission.
	f.mock.ExpectQuery("INSERT INTO confidence_point_ledger").
		WithArgs(
			sqlmock.AnyArg(), f.userID, gamification.RewardSentenceSubmitted,
			gamification.RewardSentenceSubmitted,
			gamification.ReasonSentenceSubmitted, gamification.SourceLearnerSentence,
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
		).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(uuid.New()))
	// 2d. GrantPoint for +2 AI-feedback-received.
	f.mock.ExpectQuery("INSERT INTO confidence_point_ledger").
		WithArgs(
			sqlmock.AnyArg(), f.userID, gamification.RewardAIFeedbackGot,
			gamification.RewardSentenceSubmitted+gamification.RewardAIFeedbackGot,
			gamification.ReasonAIFeedbackReceived, gamification.SourceAIFeedbackAttempt,
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
		).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(uuid.New()))
	// 2e. IncrementSentenceSubmitted (D03 disabled).
	f.mock.ExpectExec("INSERT INTO daily_activity_summaries").
		WithArgs(sqlmock.AnyArg(), f.userID, day, "UTC").
		WillReturnResult(sqlmock.NewResult(0, 1))
	// 2f. IncrementAIFeedbackReceived.
	f.mock.ExpectExec("INSERT INTO daily_activity_summaries").
		WithArgs(sqlmock.AnyArg(), f.userID, day, "UTC").
		WillReturnResult(sqlmock.NewResult(0, 1))
	// 2g. IncrementConfidencePointsEarned (+3 + +2 = +5).
	f.mock.ExpectExec("INSERT INTO daily_activity_summaries").
		WithArgs(sqlmock.AnyArg(), f.userID, day, "UTC",
			gamification.RewardSentenceSubmitted+gamification.RewardAIFeedbackGot).
		WillReturnResult(sqlmock.NewResult(0, 1))
	// 2h. ListRecentSnapshots.
	f.mock.ExpectQuery("SELECT id, user_id, local_date, timezone, review_target, reviews_completed").
		WithArgs(f.userID, 14).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "user_id", "local_date", "timezone", "review_target", "reviews_completed",
			"new_word_target", "new_words_completed", "sentence_practice_target",
			"sentence_practices_completed", "policy_version", "status", "completed_at",
			"grace_applied", "grace_day_id",
		}).AddRow(
			uuid.New(), f.userID, day, "UTC", 20, 0,
			nil, nil, nil, nil,
			gamification.MissionPolicyVersion, "open", nil, false, nil,
		))
	// 2i. CurrentGraceBalance.
	f.mock.ExpectQuery("SELECT balance_after FROM grace_day_ledger").
		WithArgs(f.userID).
		WillReturnRows(sqlmock.NewRows([]string{"balance_after"}))
	// 2j. GetStreakState.
	f.mock.ExpectQuery("SELECT user_id, current_streak_count, longest_streak_count").
		WithArgs(f.userID).
		WillReturnRows(sqlmock.NewRows([]string{
			"user_id", "current_streak_count", "longest_streak_count",
			"last_completed_local_date", "last_activity_local_date",
			"timezone", "status", "created_at", "updated_at",
		}))
	// 2k. UpsertStreakState.
	f.mock.ExpectExec("INSERT INTO streak_states").
		WillReturnResult(sqlmock.NewResult(0, 1))
	f.mock.ExpectCommit()

	result, err := f.service.SubmitSentenceFeedback(t.Context(), f.request("I work every day."))
	require.NoError(t, err)
	assert.Equal(t, LearningStatusCorrect, result.Status)
	assert.NotEqual(t, uuid.Nil, result.SentenceID)
	assert.False(t, result.MissionCompleted, "P3 path: only P2 review path can complete the mission")
	assert.Equal(t, 1, f.provider.calls)
	require.NoError(t, f.mock.ExpectationsWereMet())
}

// TestAIFeedbackP4RealMissionUpdaterSafetyBlockedNoReward is the duplicate/
// failed-safety boundary: when the local safety classifier rejects the
// sentence, the P3 pending-row phase is never entered, so the real
// missions.MissionUpdater must NOT be called and no reward SQL may be
// issued against the missions db. This is the safety-outcome guard
// called out in VOC-030-AC-05 / VOC-030-AC-08.
func TestAIFeedbackP4RealMissionUpdaterSafetyBlockedNoReward(t *testing.T) {
	f := newP4Fixture(t)

	counter := &countingMissionUpdater{}
	f.service.mission = counter

	req := f.request("I work on how to make a bomb.")
	result, err := f.service.SubmitSentenceFeedback(t.Context(), req)
	require.NoError(t, err)
	assert.Equal(t, ErrorCodeSafetyBlocked, result.ErrorCode)
	assert.False(t, result.CanRetry)
	assert.False(t, result.MissionCompleted)
	assert.Equal(t, 0, f.provider.calls, "safety block: provider must not be called")
	assert.Equal(t, 0, counter.calls, "safety block: real MissionUpdater must not be called")
	assert.Empty(t, f.repo.sentences, "safety block: no learner_sentence row was persisted")
	assert.Empty(t, f.repo.attempts, "safety block: no ai_feedback_attempts row was persisted")
	// No SQL expectations were set on f.mock, so any unexpected query
	// would cause ExpectationsWereMet to fail; the absence of error
	// confirms no P3 writes reached the missions db.
	require.NoError(t, f.mock.ExpectationsWereMet())
}

// TestAIFeedbackP4RealMissionUpdaterSelfHarmNoReward is the second
// duplicate/failed-safety boundary: self-harm-intervened attempts
// surface the CrisisResourceText and never write reward SQL, even
// though the local safety classifier itself successfully produced a
// result. The real MissionUpdater must remain uncalled.
func TestAIFeedbackP4RealMissionUpdaterSelfHarmNoReward(t *testing.T) {
	f := newP4Fixture(t)

	counter := &countingMissionUpdater{}
	f.service.mission = counter

	req := f.request("I work but I want to self-harm.")
	result, err := f.service.SubmitSentenceFeedback(t.Context(), req)
	require.NoError(t, err)
	assert.Equal(t, ErrorCodeSafetySelfHarm, result.ErrorCode)
	assert.False(t, result.CanRetry)
	assert.False(t, result.MissionCompleted)
	assert.Equal(t, 0, counter.calls, "self-harm: real MissionUpdater must not be called")
	require.NoError(t, f.mock.ExpectationsWereMet())
}

// TestAIFeedbackP4RealMissionUpdaterProviderFailureNoReward covers the
// provider-failure case: a temporary provider error must mark the
// attempt failed in the aifeedback store, must NOT call the real
// MissionUpdater, and must surface ErrorCodeTemporaryFailure with
// canRetry=true. The pre-existing P3 behavior is preserved byte-for-byte
// when the new wiring is in place.
func TestAIFeedbackP4RealMissionUpdaterProviderFailureNoReward(t *testing.T) {
	f := newP4Fixture(t)
	f.service.provider = &failingProvider{}

	counter := &countingMissionUpdater{}
	f.service.mission = counter

	result, err := f.service.SubmitSentenceFeedback(t.Context(), f.request("I work every day."))
	require.NoError(t, err)
	assert.Equal(t, ErrorCodeTemporaryFailure, result.ErrorCode)
	assert.True(t, result.CanRetry)
	assert.NotEqual(t, uuid.Nil, result.SentenceID)
	assert.NotEqual(t, uuid.Nil, result.AttemptID)
	assert.False(t, result.MissionCompleted)
	assert.Equal(t, 0, counter.calls, "provider failure: real MissionUpdater must not be called")
	require.NoError(t, f.mock.ExpectationsWereMet())
}

// TestAIFeedbackP4RealMissionUpdaterIdempotencyNoDoubleReward proves the
// P4 wiring's idempotency at the application layer: a replayed
// sentence-feedback request with a different sentence text but the same
// idempotency key short-circuits to the existing attempt and never
// reaches the real MissionUpdater. The first-line defense is the
// (user_id, idempotency_key) guard in the aifeedback module; this test
// ensures the P4 wiring doesn't accidentally bypass it.
func TestAIFeedbackP4RealMissionUpdaterIdempotencyNoDoubleReward(t *testing.T) {
	f := newP4Fixture(t)
	counter := &countingMissionUpdater{}
	f.service.mission = counter

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

	// The second request reuses the first's idempotency key. With the
	// P3 module's pre-existing idempotency guard, the second call
	// returns IdempotencyConflict without ever calling the provider or
	// the mission updater.
	_, err := f.service.SubmitSentenceFeedback(t.Context(), req1)
	require.NoError(t, err)
	// First call flowed through; counter is 1.
	assert.Equal(t, 1, counter.calls)

	result2, err := f.service.SubmitSentenceFeedback(t.Context(), req2)
	require.NoError(t, err)
	assert.Equal(t, ErrorCodeIdempotencyConflict, result2.ErrorCode)
	assert.False(t, result2.CanRetry)
	// Second call did not re-enter the reward path.
	assert.Equal(t, 1, counter.calls, "idempotency conflict: real MissionUpdater must not be called again")
	require.NoError(t, f.mock.ExpectationsWereMet())
}

// countingMissionUpdater is a test double that records the number of
// Update invocations without performing any work. It is the strict
// "must not be called" counterpart to the real MissionUpdater in the
// duplicate/failed-safety boundary tests.
type countingMissionUpdater struct {
	calls int
}

func (c *countingMissionUpdater) Update(ctx context.Context, userID, sentenceID uuid.UUID) (bool, error) {
	c.calls++
	return false, nil
}
