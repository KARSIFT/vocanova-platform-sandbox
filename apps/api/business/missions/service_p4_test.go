package missions

import (
	"database/sql"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/KARSIFT/vocanova-platform/apps/api/business/gamification"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newP4Updater builds a real missions.MissionUpdater backed by a sqlmock
// database, mirroring the construction pattern used by the P2 wiring tests
// in apps/api/business/reviews/postgres_p4_test.go.
func newP4Updater(t *testing.T, db *sql.DB) (*MissionUpdater, *gamification.Service) {
	t.Helper()
	gamSvc := gamification.NewService(gamification.NewRepository(db))
	missionsRepo := NewRepository(db)
	missionsSvc := NewService(missionsRepo, gamSvc)
	return NewMissionUpdater(missionsSvc, gamSvc), gamSvc
}

// TestUpdateForSentenceP4SuccessWiring exercises the happy path: a
// successful P3 (sentence feedback) call records the +3 sentence-submitted
// award, the +2 AI-feedback-received award, increments the activity
// counters, runs streak reconciliation, and returns missionCompleted=false
// (the structural P3 invariant: only the P2 review path increments
// reviews_completed and marks the snapshot completed).
func TestUpdateForSentenceP4SuccessWiring(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	now := time.Date(2026, 7, 26, 12, 0, 0, 0, time.UTC)
	updater, _ := newP4Updater(t, db)

	userID := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	sentenceID := uuid.MustParse("00000000-0000-0000-0000-00000000000a")
	day := time.Date(2026, 7, 26, 0, 0, 0, 0, time.UTC)

	// EnsureTodaySnapshot (idempotent insert-or-update).
	mock.ExpectBegin()
	mock.ExpectQuery("INSERT INTO daily_mission_snapshots").
		WithArgs(sqlmock.AnyArg(), userID, day, "UTC", 20, gamification.MissionPolicyVersion).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "user_id", "local_date", "timezone", "review_target", "reviews_completed",
			"new_word_target", "new_words_completed", "sentence_practice_target",
			"sentence_practices_completed", "policy_version", "status", "completed_at",
			"grace_applied", "grace_day_id",
		}).AddRow(
			uuid.New(), userID, day, "UTC", 20, 0,
			nil, nil, nil, nil,
			gamification.MissionPolicyVersion, "open", nil, false, nil,
		))
	// CurrentBalance (no rows yet -> empty).
	mock.ExpectQuery("SELECT balance_after FROM confidence_point_ledger").
		WithArgs(userID).
		WillReturnRows(sqlmock.NewRows([]string{"balance_after"}))
	// GrantPoint for +3 sentence-submitted.
	mock.ExpectQuery("INSERT INTO confidence_point_ledger").
		WithArgs(
			sqlmock.AnyArg(), userID, gamification.RewardSentenceSubmitted, gamification.RewardSentenceSubmitted,
			gamification.ReasonSentenceSubmitted, gamification.SourceLearnerSentence,
			sentenceID,
			gamification.LearnerSentenceSubmittedKey(sentenceID.String()),
			sqlmock.AnyArg(), now,
		).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(uuid.New()))
	// GrantPoint for +2 AI-feedback-received.
	mock.ExpectQuery("INSERT INTO confidence_point_ledger").
		WithArgs(
			sqlmock.AnyArg(), userID, gamification.RewardAIFeedbackGot,
			gamification.RewardSentenceSubmitted+gamification.RewardAIFeedbackGot,
			gamification.ReasonAIFeedbackReceived, gamification.SourceAIFeedbackAttempt,
			sentenceID,
			gamification.AIFeedbackAttemptReceivedKey(sentenceID.String()),
			sqlmock.AnyArg(), now,
		).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(uuid.New()))
	// IncrementSentenceSubmitted (D03 active=false, so no mission counter).
	mock.ExpectExec("INSERT INTO daily_activity_summaries").
		WithArgs(sqlmock.AnyArg(), userID, day, "UTC").
		WillReturnResult(sqlmock.NewResult(0, 1))
	// IncrementAIFeedbackReceived.
	mock.ExpectExec("INSERT INTO daily_activity_summaries").
		WithArgs(sqlmock.AnyArg(), userID, day, "UTC").
		WillReturnResult(sqlmock.NewResult(0, 1))
	// IncrementConfidencePointsEarned (+3 + +2 = +5).
	mock.ExpectExec("INSERT INTO daily_activity_summaries").
		WithArgs(sqlmock.AnyArg(), userID, day, "UTC", gamification.RewardSentenceSubmitted+gamification.RewardAIFeedbackGot).
		WillReturnResult(sqlmock.NewResult(0, 1))
	// ListRecentSnapshots for streak reconciliation.
	mock.ExpectQuery("SELECT id, user_id, local_date, timezone, review_target, reviews_completed").
		WithArgs(userID, 14).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "user_id", "local_date", "timezone", "review_target", "reviews_completed",
			"new_word_target", "new_words_completed", "sentence_practice_target",
			"sentence_practices_completed", "policy_version", "status", "completed_at",
			"grace_applied", "grace_day_id",
		}).AddRow(
			uuid.New(), userID, day, "UTC", 20, 0,
			nil, nil, nil, nil,
			gamification.MissionPolicyVersion, "open", nil, false, nil,
		))
	// CurrentGraceBalance.
	mock.ExpectQuery("SELECT balance_after FROM grace_day_ledger").
		WithArgs(userID).
		WillReturnRows(sqlmock.NewRows([]string{"balance_after"}))
	// GetStreakState.
	mock.ExpectQuery("SELECT user_id, current_streak_count, longest_streak_count").
		WithArgs(userID).
		WillReturnRows(sqlmock.NewRows([]string{
			"user_id", "current_streak_count", "longest_streak_count",
			"last_completed_local_date", "last_activity_local_date",
			"timezone", "status", "created_at", "updated_at",
		}))
	// UpsertStreakState.
	mock.ExpectExec("INSERT INTO streak_states").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	resolved := gamification.ResolvedSettings{Timezone: "UTC", DailyReviewTarget: 20}
	completed, err := updater.UpdateForSentence(t.Context(), userID, sentenceID, resolved, now, false)
	require.NoError(t, err)
	assert.False(t, completed, "P3 path never completes the mission (only the P2 review path increments reviews_completed)")
	require.NoError(t, mock.ExpectationsWereMet())
}

// TestUpdateForSentenceP4SentenceGoalActiveWhenD03Activated exercises the
// D03-activated code path: when the optional sentence-practice mission
// goal is enabled, the daily_mission_snapshots.sentence_practices_completed
// counter is also incremented in addition to the activity summary. Today
// (2026-07-26) D03 keeps this disabled at launch, but the wiring must be
// in place and exercised by a test so a future policy_version bump
// doesn't silently break the path.
func TestUpdateForSentenceP4SentenceGoalActiveWhenD03Activated(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	now := time.Date(2026, 7, 26, 12, 0, 0, 0, time.UTC)
	updater, _ := newP4Updater(t, db)

	userID := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	sentenceID := uuid.MustParse("00000000-0000-0000-0000-00000000000a")
	day := time.Date(2026, 7, 26, 0, 0, 0, 0, time.UTC)

	mock.ExpectBegin()
	mock.ExpectQuery("INSERT INTO daily_mission_snapshots").
		WithArgs(sqlmock.AnyArg(), userID, day, "UTC", 20, gamification.MissionPolicyVersion).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "user_id", "local_date", "timezone", "review_target", "reviews_completed",
			"new_word_target", "new_words_completed", "sentence_practice_target",
			"sentence_practices_completed", "policy_version", "status", "completed_at",
			"grace_applied", "grace_day_id",
		}).AddRow(
			uuid.New(), userID, day, "UTC", 20, 0,
			nil, nil, nil, nil,
			gamification.MissionPolicyVersion, "open", nil, false, nil,
		))
	mock.ExpectQuery("SELECT balance_after FROM confidence_point_ledger").
		WithArgs(userID).
		WillReturnRows(sqlmock.NewRows([]string{"balance_after"}))
	mock.ExpectQuery("INSERT INTO confidence_point_ledger").
		WithArgs(
			sqlmock.AnyArg(), userID, gamification.RewardSentenceSubmitted,
			gamification.RewardSentenceSubmitted,
			gamification.ReasonSentenceSubmitted, gamification.SourceLearnerSentence,
			sentenceID, sqlmock.AnyArg(), sqlmock.AnyArg(), now,
		).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(uuid.New()))
	mock.ExpectQuery("INSERT INTO confidence_point_ledger").
		WithArgs(
			sqlmock.AnyArg(), userID, gamification.RewardAIFeedbackGot,
			gamification.RewardSentenceSubmitted+gamification.RewardAIFeedbackGot,
			gamification.ReasonAIFeedbackReceived, gamification.SourceAIFeedbackAttempt,
			sentenceID, sqlmock.AnyArg(), sqlmock.AnyArg(), now,
		).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(uuid.New()))
	// IncrementSentenceSubmitted with includeSentenceGoal=true writes to BOTH
	// the activity summary AND the mission counter.
	mock.ExpectExec("INSERT INTO daily_activity_summaries").
		WithArgs(sqlmock.AnyArg(), userID, day, "UTC").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("UPDATE daily_mission_snapshots").
		WithArgs(userID, day).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("INSERT INTO daily_activity_summaries").
		WithArgs(sqlmock.AnyArg(), userID, day, "UTC").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("INSERT INTO daily_activity_summaries").
		WithArgs(sqlmock.AnyArg(), userID, day, "UTC", gamification.RewardSentenceSubmitted+gamification.RewardAIFeedbackGot).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery("SELECT id, user_id, local_date, timezone, review_target, reviews_completed").
		WithArgs(userID, 14).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "user_id", "local_date", "timezone", "review_target", "reviews_completed",
			"new_word_target", "new_words_completed", "sentence_practice_target",
			"sentence_practices_completed", "policy_version", "status", "completed_at",
			"grace_applied", "grace_day_id",
		}))
	mock.ExpectQuery("SELECT balance_after FROM grace_day_ledger").
		WithArgs(userID).
		WillReturnRows(sqlmock.NewRows([]string{"balance_after"}))
	mock.ExpectQuery("SELECT user_id, current_streak_count, longest_streak_count").
		WithArgs(userID).
		WillReturnRows(sqlmock.NewRows([]string{
			"user_id", "current_streak_count", "longest_streak_count",
			"last_completed_local_date", "last_activity_local_date",
			"timezone", "status", "created_at", "updated_at",
		}))
	mock.ExpectExec("INSERT INTO streak_states").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	resolved := gamification.ResolvedSettings{Timezone: "UTC", DailyReviewTarget: 20}
	completed, err := updater.UpdateForSentence(t.Context(), userID, sentenceID, resolved, now, true)
	require.NoError(t, err)
	assert.False(t, completed)
	require.NoError(t, mock.ExpectationsWereMet())
}

// TestUpdateForSentenceP4IdempotencyKeyDerivation proves the per-source-event
// idempotency keys are deterministic from the sentence ID. Two calls
// produce identical keys, which is what the (user_id, idempotency_key)
// partial unique index in confidence_point_ledger turns into a second
// line of defense against a retried/replayed transaction awarding the
// same reward twice.
func TestUpdateForSentenceP4IdempotencyKeyDerivation(t *testing.T) {
	sentenceID := uuid.MustParse("00000000-0000-0000-0000-00000000000a")
	gotSubmitted := gamification.LearnerSentenceSubmittedKey(sentenceID.String())
	gotFeedback := gamification.AIFeedbackAttemptReceivedKey(sentenceID.String())
	assert.Equal(t, "learner_sentence:00000000-0000-0000-0000-00000000000a:submitted", gotSubmitted.String())
	assert.Equal(t, "ai_feedback_attempt:00000000-0000-0000-0000-00000000000a:received", gotFeedback.String())
	// Two derivations produce identical strings (deterministic).
	assert.Equal(t, gotSubmitted.String(), gamification.LearnerSentenceSubmittedKey(sentenceID.String()).String())
}

// TestUpdateForSentenceP4RollbackOnError proves that a SQL error during
// the activity counter write rolls the entire transaction back: the
// +3/+2 ledger rows are not visible to subsequent reads and no partial
// state is persisted. This is the duplicate/failed-action safety
// guarantee from the spec's DOC-12 §5 P4 gate.
func TestUpdateForSentenceP4RollbackOnError(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	now := time.Date(2026, 7, 26, 12, 0, 0, 0, time.UTC)
	updater, _ := newP4Updater(t, db)

	userID := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	sentenceID := uuid.MustParse("00000000-0000-0000-0000-00000000000a")
	day := time.Date(2026, 7, 26, 0, 0, 0, 0, time.UTC)

	mock.ExpectBegin()
	mock.ExpectQuery("INSERT INTO daily_mission_snapshots").
		WithArgs(sqlmock.AnyArg(), userID, day, "UTC", 20, gamification.MissionPolicyVersion).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "user_id", "local_date", "timezone", "review_target", "reviews_completed",
			"new_word_target", "new_words_completed", "sentence_practice_target",
			"sentence_practices_completed", "policy_version", "status", "completed_at",
			"grace_applied", "grace_day_id",
		}).AddRow(
			uuid.New(), userID, day, "UTC", 20, 0,
			nil, nil, nil, nil,
			gamification.MissionPolicyVersion, "open", nil, false, nil,
		))
	mock.ExpectQuery("SELECT balance_after FROM confidence_point_ledger").
		WithArgs(userID).
		WillReturnRows(sqlmock.NewRows([]string{"balance_after"}))
	mock.ExpectQuery("INSERT INTO confidence_point_ledger").
		WithArgs(
			sqlmock.AnyArg(), userID, gamification.RewardSentenceSubmitted, gamification.RewardSentenceSubmitted,
			gamification.ReasonSentenceSubmitted, gamification.SourceLearnerSentence,
			sentenceID, sqlmock.AnyArg(), sqlmock.AnyArg(), now,
		).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(uuid.New()))
	mock.ExpectQuery("INSERT INTO confidence_point_ledger").
		WithArgs(
			sqlmock.AnyArg(), userID, gamification.RewardAIFeedbackGot,
			gamification.RewardSentenceSubmitted+gamification.RewardAIFeedbackGot,
			gamification.ReasonAIFeedbackReceived, gamification.SourceAIFeedbackAttempt,
			sentenceID, sqlmock.AnyArg(), sqlmock.AnyArg(), now,
		).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(uuid.New()))
	// IncrementSentenceSubmitted fails.
	mock.ExpectExec("INSERT INTO daily_activity_summaries").
		WithArgs(sqlmock.AnyArg(), userID, day, "UTC").
		WillReturnError(sqlmock.ErrCancelled)
	mock.ExpectRollback()

	resolved := gamification.ResolvedSettings{Timezone: "UTC", DailyReviewTarget: 20}
	completed, err := updater.UpdateForSentence(t.Context(), userID, sentenceID, resolved, now, false)
	require.Error(t, err)
	assert.False(t, completed, "rollback path returns false")
	require.NoError(t, mock.ExpectationsWereMet())
}

// TestUpdateForSentenceP4UpdateEntryPointWiresResolverAndDelegate is a
// smoke test for the top-level Update(ctx, userID, sentenceID) entry
// point on the real MissionUpdater. Update is the method the aifeedback
// module's MissionUpdater interface seam calls into, and it must:
//   - resolve settings (the D01 chain, applied with an empty
//     client-supplied IANA timezone since the aifeedback module passes no
//     request-time timezone here),
//   - delegate to UpdateForSentence with includeSentenceGoal=false
//     (VOC-030-D03 keeps the bonus sentence-practice mission goal
//     disabled at launch).
//
// The full SQL flow is exercised by the tests above; this case asserts
// the wiring inside Update itself stays correct.
func TestUpdateForSentenceP4UpdateEntryPointWiresResolverAndDelegate(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	updater, _ := newP4Updater(t, db)

	userID := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	sentenceID := uuid.MustParse("00000000-0000-0000-0000-00000000000a")
	// Update() reads time.Now() internally; we mirror that here so the
	// daily_mission_snapshots row matches what the implementation will
	// actually insert.
	day, err := gamification.LocalDate(time.Now(), "UTC")
	require.NoError(t, err)

	// GetSettings reads user_settings (empty -> no rows).
	mock.ExpectQuery("SELECT user_id, timezone, daily_review_target").
		WithArgs(userID).
		WillReturnRows(sqlmock.NewRows([]string{
			"user_id", "timezone", "daily_review_target", "review_interval_preset",
			"notifications_enabled", "marketing_emails_enabled", "app_language",
		}))
	// EnsureTodaySnapshot (the P3 Update path delegates to UpdateForSentence
	// which opens its own transaction).
	mock.ExpectBegin()
	mock.ExpectQuery("INSERT INTO daily_mission_snapshots").
		WithArgs(sqlmock.AnyArg(), userID, day, "UTC", 20, gamification.MissionPolicyVersion).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "user_id", "local_date", "timezone", "review_target", "reviews_completed",
			"new_word_target", "new_words_completed", "sentence_practice_target",
			"sentence_practices_completed", "policy_version", "status", "completed_at",
			"grace_applied", "grace_day_id",
		}).AddRow(
			uuid.New(), userID, day, "UTC", 20, 0,
			nil, nil, nil, nil,
			gamification.MissionPolicyVersion, "open", nil, false, nil,
		))
	mock.ExpectQuery("SELECT balance_after FROM confidence_point_ledger").
		WithArgs(userID).
		WillReturnRows(sqlmock.NewRows([]string{"balance_after"}))
	mock.ExpectQuery("INSERT INTO confidence_point_ledger").
		WithArgs(
			sqlmock.AnyArg(), userID, gamification.RewardSentenceSubmitted,
			gamification.RewardSentenceSubmitted,
			gamification.ReasonSentenceSubmitted, gamification.SourceLearnerSentence,
			sentenceID, sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
		).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(uuid.New()))
	mock.ExpectQuery("INSERT INTO confidence_point_ledger").
		WithArgs(
			sqlmock.AnyArg(), userID, gamification.RewardAIFeedbackGot,
			gamification.RewardSentenceSubmitted+gamification.RewardAIFeedbackGot,
			gamification.ReasonAIFeedbackReceived, gamification.SourceAIFeedbackAttempt,
			sentenceID, sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
		).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(uuid.New()))
	mock.ExpectExec("INSERT INTO daily_activity_summaries").
		WithArgs(sqlmock.AnyArg(), userID, day, "UTC").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("INSERT INTO daily_activity_summaries").
		WithArgs(sqlmock.AnyArg(), userID, day, "UTC").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("INSERT INTO daily_activity_summaries").
		WithArgs(sqlmock.AnyArg(), userID, day, "UTC", gamification.RewardSentenceSubmitted+gamification.RewardAIFeedbackGot).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery("SELECT id, user_id, local_date, timezone, review_target, reviews_completed").
		WithArgs(userID, 14).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "user_id", "local_date", "timezone", "review_target", "reviews_completed",
			"new_word_target", "new_words_completed", "sentence_practice_target",
			"sentence_practices_completed", "policy_version", "status", "completed_at",
			"grace_applied", "grace_day_id",
		}))
	mock.ExpectQuery("SELECT balance_after FROM grace_day_ledger").
		WithArgs(userID).
		WillReturnRows(sqlmock.NewRows([]string{"balance_after"}))
	mock.ExpectQuery("SELECT user_id, current_streak_count, longest_streak_count").
		WithArgs(userID).
		WillReturnRows(sqlmock.NewRows([]string{
			"user_id", "current_streak_count", "longest_streak_count",
			"last_completed_local_date", "last_activity_local_date",
			"timezone", "status", "created_at", "updated_at",
		}))
	mock.ExpectExec("INSERT INTO streak_states").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	completed, err := updater.Update(t.Context(), userID, sentenceID)
	require.NoError(t, err)
	assert.False(t, completed, "P3 Update path never completes the mission")
	require.NoError(t, mock.ExpectationsWereMet())
}
