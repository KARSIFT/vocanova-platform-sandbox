package reviews

import (
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/KARSIFT/vocanova-platform/apps/api/business/gamification"
	"github.com/KARSIFT/vocanova-platform/apps/api/business/missions"
	"github.com/KARSIFT/vocanova-platform/apps/api/foundation/clock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestReviewRewardKind is a pure unit test for the (result, rating) ->
// (correct, skipped, reward kind, ok) mapping. It documents the
// DOC-06 §11 reward table mirror used by the P4 wiring and catches any
// silent drift if the underlying gamification reward kinds change.
func TestReviewRewardKind(t *testing.T) {
	cases := []struct {
		name        string
		result      string
		rating      string
		wantCorrect bool
		wantSkipped bool
		wantKind    gamification.RewardKind
		wantOK      bool
	}{
		{"correct hard", ResultCorrect, RatingHard, true, false, gamification.RewardKindReviewHard, true},
		{"correct good", ResultCorrect, RatingGood, true, false, gamification.RewardKindReviewGood, true},
		{"correct easy", ResultCorrect, RatingEasy, true, false, gamification.RewardKindReviewEasy, true},
		{"correct default to good", ResultCorrect, "", true, false, gamification.RewardKindReviewGood, true},
		{"incorrect again", ResultIncorrect, RatingAgain, false, false, gamification.RewardKindReviewAgain, true},
		{"skipped no rating", ResultSkipped, "", false, true, "", true},
		{"correct again invalid", ResultCorrect, RatingAgain, false, false, "", false},
		{"incorrect hard invalid", ResultIncorrect, RatingHard, false, false, "", false},
		{"unknown result", "garbage", RatingGood, false, false, "", false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			correct, skipped, kind, ok := reviewRewardKind(c.result, c.rating)
			assert.Equal(t, c.wantCorrect, correct)
			assert.Equal(t, c.wantSkipped, skipped)
			assert.Equal(t, c.wantKind, kind)
			assert.Equal(t, c.wantOK, ok)
		})
	}
}

// newP4Repo constructs a PostgreSQLRepository with the P4 wiring enabled
// against the supplied sqlmock-backed db. It centralizes the option
// wiring so the individual test bodies can focus on the SQL expectations.
func newP4Repo(db *sqlmock.Sqlmock, now time.Time, gamSvc *gamification.Service, missionsSvc *missions.Service) *PostgreSQLRepository {
	// The variadic options don't depend on the actual *sql.DB type at
	// this call site; sqlmock's underlying *sql.DB satisfies what we need.
	return NewPostgreSQLRepository(nil, clock.Fixed{T: now},
		WithGamificationService(gamSvc),
		WithMissionsService(missionsSvc),
	)
}

// p4TestID is a tiny helper to build a user_id UUID with all-zero
// padding except the last byte — kept inline to avoid a package-level
// helper that the production code would have to ignore.
func p4TestID(t *testing.T, s string) uuid.UUID {
	t.Helper()
	id, err := uuid.Parse(s)
	require.NoError(t, err)
	return id
}

// TestPostgreSQLRepositorySubmitReviewP4NilDependenciesNoP4Wiring proves the
// pre-existing P2 submission behavior is byte-for-byte unchanged with the
// new wiring in place. When the gamification or missions service is nil
// (the legacy configuration), no P4 SQL is issued: no user_settings read,
// no daily_mission_snapshots read/write, no confidence_point_ledger
// insert, no streak_states upsert, no grace_day_ledger activity.
func TestPostgreSQLRepositorySubmitReviewP4NilDependenciesNoP4Wiring(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	now := time.Date(2026, 7, 25, 12, 0, 0, 0, time.UTC)
	repo := NewPostgreSQLRepository(db, clock.Fixed{T: now})

	userID := p4TestID(t, "00000000-0000-0000-0000-000000000001")
	userWordID := p4TestID(t, "00000000-0000-0000-0000-000000000002")
	meaningID := p4TestID(t, "00000000-0000-0000-0000-000000000003")

	mock.ExpectBegin()
	mock.ExpectQuery("SELECT review_step, meaning_id").
		WithArgs(userWordID, userID).
		WillReturnRows(sqlmock.NewRows([]string{"review_step", "meaning_id", "total_review_count", "correct_review_count", "consecutive_correct_count", "consecutive_incorrect_count"}).
			AddRow(0, meaningID, 0, 0, 0, 0))
	expectReviewKeyClaim(mock)
	mock.ExpectQuery("SELECT ra.id, ra.user_word_id").
		WithArgs(userID, "ca-1").
		WillReturnRows(sqlmock.NewRows([]string{"id", "user_word_id", "meaning_id", "attempt_type", "prompt_type", "result", "rating", "review_step_before", "review_step_after", "answered_at", "response_time_ms", "selected_option_meaning_id", "typed_answer", "was_hint_used", "source", "client_attempt_id", "metadata", "next_review_at"}))
	mock.ExpectExec("INSERT INTO review_attempts").
		WithArgs(sqlmock.AnyArg(), userID, userWordID, meaningID, "review", "multiple_choice", "correct", "good", 0, 1, now, 0, nil, nil, false, "review", "ca-1", nil, now, now).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("UPDATE user_words").
		WithArgs(1, sqlmock.AnyArg(), now, "correct", "good", 1, 1, 1, 0, now, userWordID, userID).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	req := SubmitReviewRequest{
		UserID:          userID,
		UserWordID:      userWordID,
		MeaningID:       meaningID,
		AttemptType:     AttemptTypeReview,
		PromptType:      PromptTypeMultipleChoice,
		Result:          ResultCorrect,
		Rating:          RatingGood,
		AnsweredAt:      now,
		ResponseTimeMs:  0,
		Source:          SourceReview,
		ClientAttemptID: "ca-1",
		IdempotencyKey:  "review-test-key",
	}
	attempt, err := repo.SubmitReview(t.Context(), req)
	require.NoError(t, err)
	assert.Equal(t, 1, attempt.ReviewStepAfter)
	require.NoError(t, mock.ExpectationsWereMet())
	// sqlmock would have failed ExpectationsWereMet if any extra P4
	// SQL had been issued.
}

// TestPostgreSQLRepositorySubmitReviewP4RatingGoodWiring exercises the P4
// reward wiring on a non-completing review: a single Good rating should
// grant the +5 review_correct point, bump the daily mission's
// reviews_completed to 1 (under the target of 20), update the activity
// summary, and write a fresh streak_states row. The daily-mission
// completion (+10) and the streak advance must NOT fire because the
// target is not yet met.
func TestPostgreSQLRepositorySubmitReviewP4RatingGoodWiring(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	now := time.Date(2026, 7, 26, 12, 0, 0, 0, time.UTC)
	gamSvc := gamification.NewService(gamification.NewRepository(db))
	missionsSvc := missions.NewService(missions.NewRepository(db), gamSvc)
	repo := NewPostgreSQLRepository(db, clock.Fixed{T: now},
		WithGamificationService(gamSvc),
		WithMissionsService(missionsSvc),
	)

	userID := p4TestID(t, "00000000-0000-0000-0000-000000000001")
	userWordID := p4TestID(t, "00000000-0000-0000-0000-000000000002")
	meaningID := p4TestID(t, "00000000-0000-0000-0000-000000000003")

	day := time.Date(2026, 7, 26, 0, 0, 0, 0, time.UTC)

	// Phase 1: existing P2 SQL (lock user word, idempotency, insert attempt,
	// update user_words).
	mock.ExpectBegin()
	mock.ExpectQuery("SELECT review_step, meaning_id").
		WithArgs(userWordID, userID).
		WillReturnRows(sqlmock.NewRows([]string{"review_step", "meaning_id", "total_review_count", "correct_review_count", "consecutive_correct_count", "consecutive_incorrect_count"}).
			AddRow(0, meaningID, 0, 0, 0, 0))
	expectReviewKeyClaim(mock)
	mock.ExpectQuery("SELECT ra.id, ra.user_word_id").
		WithArgs(userID, "ca-1").
		WillReturnRows(sqlmock.NewRows([]string{"id", "user_word_id", "meaning_id", "attempt_type", "prompt_type", "result", "rating", "review_step_before", "review_step_after", "answered_at", "response_time_ms", "selected_option_meaning_id", "typed_answer", "was_hint_used", "source", "client_attempt_id", "metadata", "next_review_at"}))
	mock.ExpectExec("INSERT INTO review_attempts").
		WithArgs(sqlmock.AnyArg(), userID, userWordID, meaningID, "review", "multiple_choice", "correct", "good", 0, 1, now, 0, nil, nil, false, "review", "ca-1", nil, now, now).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("UPDATE user_words").
		WithArgs(1, sqlmock.AnyArg(), now, "correct", "good", 1, 1, 1, 0, now, userWordID, userID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	// Phase 2: P4 wiring SQL (still inside the same tx).
	// 2a. GetSettings reads user_settings (on db, not tx).
	mock.ExpectQuery("SELECT user_id, timezone, daily_review_target").
		WithArgs(userID).
		WillReturnRows(sqlmock.NewRows([]string{"user_id", "timezone", "daily_review_target", "review_interval_preset", "notifications_enabled", "marketing_emails_enabled", "app_language"}).
			AddRow(userID, "UTC", 20, "vocanova_default", false, false, "en"))
	// 2b. EnsureTodaySnapshot -> CreateDailyMissionSnapshot ON CONFLICT
	// UPDATE/RETURNING.
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
	// 2c. IncrementReviewsCompleted: UPDATE daily_mission_snapshots
	// RETURNING reviews_completed, then INSERT into daily_activity_summaries.
	mock.ExpectQuery("UPDATE daily_mission_snapshots").
		WithArgs(userID, day).
		WillReturnRows(sqlmock.NewRows([]string{"reviews_completed"}).AddRow(1))
	mock.ExpectExec("INSERT INTO daily_activity_summaries").
		WillReturnResult(sqlmock.NewResult(0, 1))
	// 2d. getLatestPointBalanceTx (no rows yet -> empty).
	mock.ExpectQuery("SELECT COALESCE\\(balance_after, 0\\) FROM confidence_point_ledger").
		WithArgs(userID).
		WillReturnRows(sqlmock.NewRows([]string{"balance_after"}))
	// 2e. GrantPoint -> InsertPointLedger for the +5 Good rating reward.
	mock.ExpectQuery("INSERT INTO confidence_point_ledger").
		WithArgs(
			sqlmock.AnyArg(), userID, gamification.RewardReviewGood, gamification.RewardReviewGood,
			gamification.ReasonReviewCorrect, gamification.SourceReviewAttempt,
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), now,
		).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(uuid.New()))
	// 2f. IncrementConfidencePointsEarned for the +5.
	mock.ExpectExec("INSERT INTO daily_activity_summaries").
		WithArgs(sqlmock.AnyArg(), userID, day, "UTC", gamification.RewardReviewGood).
		WillReturnResult(sqlmock.NewResult(0, 1))
	// 2g. Mission not yet completed (1 < 20) — no MarkSnapshotCompleted
	// or second GrantPoint.
	// 2h. fetchStreakSnapshotsTx: SELECT recent snapshots.
	mock.ExpectQuery("SELECT local_date, status, completed_at, grace_applied, grace_day_id FROM daily_mission_snapshots").
		WithArgs(userID, 14).
		WillReturnRows(sqlmock.NewRows([]string{"local_date", "status", "completed_at", "grace_applied", "grace_day_id"}))
	// 2i. getLatestGraceBalanceTx (no rows yet -> empty).
	mock.ExpectQuery("SELECT COALESCE\\(balance_after, 0\\) FROM grace_day_ledger").
		WithArgs(userID).
		WillReturnRows(sqlmock.NewRows([]string{"balance_after"}))
	// 2j. ReconcileAndAdvance: GetStreakState on db, then UpsertStreakState
	// on tx.
	mock.ExpectQuery("SELECT user_id, current_streak_count, longest_streak_count").
		WithArgs(userID).
		WillReturnRows(sqlmock.NewRows([]string{"user_id", "current_streak_count", "longest_streak_count", "last_completed_local_date", "last_activity_local_date", "timezone", "status", "created_at", "updated_at"}))
	mock.ExpectExec("INSERT INTO streak_states").
		WillReturnResult(sqlmock.NewResult(0, 1))

	mock.ExpectCommit()

	req := SubmitReviewRequest{
		UserID:          userID,
		UserWordID:      userWordID,
		MeaningID:       meaningID,
		AttemptType:     AttemptTypeReview,
		PromptType:      PromptTypeMultipleChoice,
		Result:          ResultCorrect,
		Rating:          RatingGood,
		AnsweredAt:      now,
		ResponseTimeMs:  0,
		Source:          SourceReview,
		ClientAttemptID: "ca-1",
		IdempotencyKey:  "review-test-key",
	}
	attempt, err := repo.SubmitReview(t.Context(), req)
	require.NoError(t, err)
	assert.Equal(t, 1, attempt.ReviewStepAfter)
	require.NoError(t, mock.ExpectationsWereMet())
}

// TestPostgreSQLRepositorySubmitReviewP4MissionCompletion exercises the
// mission-completion path: a review that lands exactly on the target
// (20th review) must transition today's snapshot to status='completed',
// grant the +10 daily-mission reward, and signal the streak to advance
// exactly once. The mission-completion grant uses the
// daily_mission:<user>:<date>:completed idempotency key, so a retried
// transaction can never double-award.
func TestPostgreSQLRepositorySubmitReviewP4MissionCompletion(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	now := time.Date(2026, 7, 26, 12, 0, 0, 0, time.UTC)
	gamSvc := gamification.NewService(gamification.NewRepository(db))
	missionsSvc := missions.NewService(missions.NewRepository(db), gamSvc)
	repo := NewPostgreSQLRepository(db, clock.Fixed{T: now},
		WithGamificationService(gamSvc),
		WithMissionsService(missionsSvc),
	)

	userID := p4TestID(t, "00000000-0000-0000-0000-000000000001")
	userWordID := p4TestID(t, "00000000-0000-0000-0000-000000000002")
	meaningID := p4TestID(t, "00000000-0000-0000-0000-000000000003")
	day := time.Date(2026, 7, 26, 0, 0, 0, 0, time.UTC)

	// Phase 1: existing P2 SQL.
	mock.ExpectBegin()
	mock.ExpectQuery("SELECT review_step, meaning_id").
		WithArgs(userWordID, userID).
		WillReturnRows(sqlmock.NewRows([]string{"review_step", "meaning_id", "total_review_count", "correct_review_count", "consecutive_correct_count", "consecutive_incorrect_count"}).
			AddRow(0, meaningID, 0, 0, 0, 0))
	expectReviewKeyClaim(mock)
	mock.ExpectQuery("SELECT ra.id, ra.user_word_id").
		WithArgs(userID, "ca-20").
		WillReturnRows(sqlmock.NewRows([]string{"id", "user_word_id", "meaning_id", "attempt_type", "prompt_type", "result", "rating", "review_step_before", "review_step_after", "answered_at", "response_time_ms", "selected_option_meaning_id", "typed_answer", "was_hint_used", "source", "client_attempt_id", "metadata", "next_review_at"}))
	mock.ExpectExec("INSERT INTO review_attempts").
		WithArgs(sqlmock.AnyArg(), userID, userWordID, meaningID, "review", "multiple_choice", "correct", "good", 0, 1, now, 0, nil, nil, false, "review", "ca-20", nil, now, now).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("UPDATE user_words").
		WithArgs(1, sqlmock.AnyArg(), now, "correct", "good", 1, 1, 1, 0, now, userWordID, userID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	// Phase 2: P4 wiring.
	mock.ExpectQuery("SELECT user_id, timezone, daily_review_target").
		WithArgs(userID).
		WillReturnRows(sqlmock.NewRows([]string{"user_id", "timezone", "daily_review_target", "review_interval_preset", "notifications_enabled", "marketing_emails_enabled", "app_language"}).
			AddRow(userID, "UTC", 20, "vocanova_default", false, false, "en"))
	// Snapshot already exists; EnsureTodaySnapshot is idempotent.
	mock.ExpectQuery("INSERT INTO daily_mission_snapshots").
		WithArgs(sqlmock.AnyArg(), userID, day, "UTC", 20, gamification.MissionPolicyVersion).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "user_id", "local_date", "timezone", "review_target", "reviews_completed",
			"new_word_target", "new_words_completed", "sentence_practice_target",
			"sentence_practices_completed", "policy_version", "status", "completed_at",
			"grace_applied", "grace_day_id",
		}).AddRow(
			uuid.New(), userID, day, "UTC", 20, 19,
			nil, nil, nil, nil,
			gamification.MissionPolicyVersion, "open", nil, false, nil,
		))
	// IncrementReviewsCompleted: 19 -> 20 (target met).
	mock.ExpectQuery("UPDATE daily_mission_snapshots").
		WithArgs(userID, day).
		WillReturnRows(sqlmock.NewRows([]string{"reviews_completed"}).AddRow(20))
	mock.ExpectExec("INSERT INTO daily_activity_summaries").
		WillReturnResult(sqlmock.NewResult(0, 1))
	// Balance before the rating-tiered grant: 0 (clean start).
	mock.ExpectQuery("SELECT COALESCE\\(balance_after, 0\\) FROM confidence_point_ledger").
		WithArgs(userID).
		WillReturnRows(sqlmock.NewRows([]string{"balance_after"}))
	// +5 Good reward.
	mock.ExpectQuery("INSERT INTO confidence_point_ledger").
		WithArgs(
			sqlmock.AnyArg(), userID, gamification.RewardReviewGood, gamification.RewardReviewGood,
			gamification.ReasonReviewCorrect, gamification.SourceReviewAttempt,
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), now,
		).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(uuid.New()))
	// Points-earned summary for the +5.
	mock.ExpectExec("INSERT INTO daily_activity_summaries").
		WithArgs(sqlmock.AnyArg(), userID, day, "UTC", gamification.RewardReviewGood).
		WillReturnResult(sqlmock.NewResult(0, 1))
	// Mission completion: MarkSnapshotCompleted transitions open -> completed.
	mock.ExpectExec("UPDATE daily_mission_snapshots").
		WithArgs(userID, day, now).
		WillReturnResult(sqlmock.NewResult(0, 1))
	// Balance before the +10 daily-mission grant: now 5 (the +5 was just written).
	mock.ExpectQuery("SELECT COALESCE\\(balance_after, 0\\) FROM confidence_point_ledger").
		WithArgs(userID).
		WillReturnRows(sqlmock.NewRows([]string{"balance_after"}).AddRow(gamification.RewardReviewGood))
	// +10 daily-mission-completion reward.
	mock.ExpectQuery("INSERT INTO confidence_point_ledger").
		WithArgs(
			sqlmock.AnyArg(), userID, gamification.RewardDailyMissionDone,
			gamification.RewardReviewGood+gamification.RewardDailyMissionDone,
			gamification.ReasonDailyMissionCompleted, gamification.SourceDailyMission,
			nil,
			sqlmock.AnyArg(), sqlmock.AnyArg(), now,
		).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(uuid.New()))
	// Points-earned summary for the +10.
	mock.ExpectExec("INSERT INTO daily_activity_summaries").
		WithArgs(sqlmock.AnyArg(), userID, day, "UTC", gamification.RewardDailyMissionDone).
		WillReturnResult(sqlmock.NewResult(0, 1))
	// Streak reconciliation: fetch recent snapshots (includes today just
	// marked completed, mirroring applyP4ReviewWiring's mark-then-fetch
	// order) and current grace balance.
	yesterday := time.Date(2026, 7, 25, 0, 0, 0, 0, time.UTC)
	mock.ExpectQuery("SELECT local_date, status, completed_at, grace_applied, grace_day_id FROM daily_mission_snapshots").
		WithArgs(userID, 14).
		WillReturnRows(sqlmock.NewRows([]string{"local_date", "status", "completed_at", "grace_applied", "grace_day_id"}).
			AddRow(yesterday, "completed", now, false, nil).
			AddRow(day, "completed", now, false, nil))
	mock.ExpectQuery("SELECT COALESCE\\(balance_after, 0\\) FROM grace_day_ledger").
		WithArgs(userID).
		WillReturnRows(sqlmock.NewRows([]string{"balance_after"}))
	// ReconcileAndAdvance: GetStreakState on db.
	mock.ExpectQuery("SELECT user_id, current_streak_count, longest_streak_count").
		WithArgs(userID).
		WillReturnRows(sqlmock.NewRows([]string{"user_id", "current_streak_count", "longest_streak_count", "last_completed_local_date", "last_activity_local_date", "timezone", "status", "created_at", "updated_at"}).
			AddRow(userID, 19, 19, yesterday, yesterday, "UTC", "active", now, now))
	// UpsertStreakState on tx.
	mock.ExpectExec("INSERT INTO streak_states").
		WillReturnResult(sqlmock.NewResult(0, 1))

	mock.ExpectCommit()

	req := SubmitReviewRequest{
		UserID:          userID,
		UserWordID:      userWordID,
		MeaningID:       meaningID,
		AttemptType:     AttemptTypeReview,
		PromptType:      PromptTypeMultipleChoice,
		Result:          ResultCorrect,
		Rating:          RatingGood,
		AnsweredAt:      now,
		ResponseTimeMs:  0,
		Source:          SourceReview,
		ClientAttemptID: "ca-20",
		IdempotencyKey:  "review-test-key",
	}
	attempt, err := repo.SubmitReview(t.Context(), req)
	require.NoError(t, err)
	assert.Equal(t, 1, attempt.ReviewStepAfter)
	require.NoError(t, mock.ExpectationsWereMet())
}

// TestPostgreSQLRepositorySubmitReviewP4SkippedNoRatingReward proves that a
// skipped review (Result=skipped, no rating) does not award a rating-tiered
// point. The activity summary still gets a reviews_skipped counter so the
// read APIs can show "skipped today" honestly, but the
// confidence_point_ledger and the daily-mission reward path are untouched.
func TestPostgreSQLRepositorySubmitReviewP4SkippedNoRatingReward(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	now := time.Date(2026, 7, 26, 12, 0, 0, 0, time.UTC)
	gamSvc := gamification.NewService(gamification.NewRepository(db))
	missionsSvc := missions.NewService(missions.NewRepository(db), gamSvc)
	repo := NewPostgreSQLRepository(db, clock.Fixed{T: now},
		WithGamificationService(gamSvc),
		WithMissionsService(missionsSvc),
	)

	userID := p4TestID(t, "00000000-0000-0000-0000-000000000001")
	userWordID := p4TestID(t, "00000000-0000-0000-0000-000000000002")
	meaningID := p4TestID(t, "00000000-0000-0000-0000-000000000003")
	day := time.Date(2026, 7, 26, 0, 0, 0, 0, time.UTC)

	// Phase 1: existing P2 SQL.
	mock.ExpectBegin()
	mock.ExpectQuery("SELECT review_step, meaning_id").
		WithArgs(userWordID, userID).
		WillReturnRows(sqlmock.NewRows([]string{"review_step", "meaning_id", "total_review_count", "correct_review_count", "consecutive_correct_count", "consecutive_incorrect_count"}).
			AddRow(0, meaningID, 0, 0, 0, 0))
	expectReviewKeyClaim(mock)
	mock.ExpectQuery("SELECT ra.id, ra.user_word_id").
		WithArgs(userID, "ca-skip").
		WillReturnRows(sqlmock.NewRows([]string{"id", "user_word_id", "meaning_id", "attempt_type", "prompt_type", "result", "rating", "review_step_before", "review_step_after", "answered_at", "response_time_ms", "selected_option_meaning_id", "typed_answer", "was_hint_used", "source", "client_attempt_id", "metadata", "next_review_at"}))
	mock.ExpectExec("INSERT INTO review_attempts").
		WithArgs(sqlmock.AnyArg(), userID, userWordID, meaningID, "review", "multiple_choice", "skipped", nil, 0, 0, now, 0, nil, nil, false, "review", "ca-skip", nil, now, now).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("UPDATE user_words").
		WithArgs(0, sqlmock.AnyArg(), now, "skipped", sqlmock.AnyArg(), 1, 0, 0, 0, now, userWordID, userID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	// Phase 2: P4 wiring.
	mock.ExpectQuery("SELECT user_id, timezone, daily_review_target").
		WithArgs(userID).
		WillReturnRows(sqlmock.NewRows([]string{"user_id", "timezone", "daily_review_target", "review_interval_preset", "notifications_enabled", "marketing_emails_enabled", "app_language"}).
			AddRow(userID, "UTC", 20, "vocanova_default", false, false, "en"))
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
	// reviews_attempted and reviews_skipped both bump by 1; reviews_completed
	// also increments toward the target (every submitted attempt counts).
	mock.ExpectQuery("UPDATE daily_mission_snapshots").
		WithArgs(userID, day).
		WillReturnRows(sqlmock.NewRows([]string{"reviews_completed"}).AddRow(1))
	mock.ExpectExec("INSERT INTO daily_activity_summaries").
		WillReturnResult(sqlmock.NewResult(0, 1))
	// NO getLatestPointBalanceTx, NO InsertPointLedger, NO
	// IncrementConfidencePointsEarned: the skipped review's path skips
	// the rating-tiered reward.
	// Mission not yet completed (1 < 20) — no MarkSnapshotCompleted.
	// Streak reconciliation still runs (read-only update of streak state).
	mock.ExpectQuery("SELECT local_date, status, completed_at, grace_applied, grace_day_id FROM daily_mission_snapshots").
		WithArgs(userID, 14).
		WillReturnRows(sqlmock.NewRows([]string{"local_date", "status", "completed_at", "grace_applied", "grace_day_id"}))
	mock.ExpectQuery("SELECT COALESCE\\(balance_after, 0\\) FROM grace_day_ledger").
		WithArgs(userID).
		WillReturnRows(sqlmock.NewRows([]string{"balance_after"}))
	mock.ExpectQuery("SELECT user_id, current_streak_count, longest_streak_count").
		WithArgs(userID).
		WillReturnRows(sqlmock.NewRows([]string{"user_id", "current_streak_count", "longest_streak_count", "last_completed_local_date", "last_activity_local_date", "timezone", "status", "created_at", "updated_at"}))
	mock.ExpectExec("INSERT INTO streak_states").
		WillReturnResult(sqlmock.NewResult(0, 1))

	mock.ExpectCommit()

	req := SubmitReviewRequest{
		UserID:          userID,
		UserWordID:      userWordID,
		MeaningID:       meaningID,
		AttemptType:     AttemptTypeReview,
		PromptType:      PromptTypeMultipleChoice,
		Result:          ResultSkipped,
		AnsweredAt:      now,
		ResponseTimeMs:  0,
		Source:          SourceReview,
		ClientAttemptID: "ca-skip",
		IdempotencyKey:  "review-test-key",
	}
	attempt, err := repo.SubmitReview(t.Context(), req)
	require.NoError(t, err)
	assert.Equal(t, 0, attempt.ReviewStepAfter)
	require.NoError(t, mock.ExpectationsWereMet())
}

// TestPostgreSQLRepositorySubmitReviewP4IdempotentMatchNoP4Wiring proves
// the duplicate-safety guarantee at the application layer: a replayed
// submission with the same client_attempt_id and same body short-circuits
// to the existing attempt and never enters the P4 wiring block. This is
// the first of the two lines of defense against a double reward: the
// application-level idempotency guard on (user_id, client_attempt_id).
func TestPostgreSQLRepositorySubmitReviewP4IdempotentMatchNoP4Wiring(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	now := time.Date(2026, 7, 26, 12, 0, 0, 0, time.UTC)
	gamSvc := gamification.NewService(gamification.NewRepository(db))
	missionsSvc := missions.NewService(missions.NewRepository(db), gamSvc)
	repo := NewPostgreSQLRepository(db, clock.Fixed{T: now},
		WithGamificationService(gamSvc),
		WithMissionsService(missionsSvc),
	)

	userID := p4TestID(t, "00000000-0000-0000-0000-000000000001")
	userWordID := p4TestID(t, "00000000-0000-0000-0000-000000000002")
	meaningID := p4TestID(t, "00000000-0000-0000-0000-000000000003")
	existingAttemptID := p4TestID(t, "00000000-0000-0000-0000-000000000004")

	mock.ExpectBegin()
	mock.ExpectQuery("SELECT review_step, meaning_id").
		WithArgs(userWordID, userID).
		WillReturnRows(sqlmock.NewRows([]string{"review_step", "meaning_id", "total_review_count", "correct_review_count", "consecutive_correct_count", "consecutive_incorrect_count"}).
			AddRow(0, meaningID, 0, 0, 0, 0))
	// fetchAttemptByClientAttemptID returns the existing attempt (same body).
	expectReviewKeyClaim(mock)
	mock.ExpectQuery("SELECT ra.id, ra.user_word_id").
		WithArgs(userID, "ca-1").
		WillReturnRows(sqlmock.NewRows([]string{"id", "user_word_id", "meaning_id", "attempt_type", "prompt_type", "result", "rating", "review_step_before", "review_step_after", "answered_at", "response_time_ms", "selected_option_meaning_id", "typed_answer", "was_hint_used", "source", "client_attempt_id", "metadata", "next_review_at"}).
			AddRow(existingAttemptID, userWordID, meaningID, "review", "multiple_choice", "correct", "good", 0, 1, now, 0, nil, nil, false, "review", "ca-1", nil, now))
	mock.ExpectCommit()
	// No P4 SQL — the existing-attempt short-circuit returns before the
	// gamification/missions block.

	req := SubmitReviewRequest{
		UserID:          userID,
		UserWordID:      userWordID,
		MeaningID:       meaningID,
		AttemptType:     AttemptTypeReview,
		PromptType:      PromptTypeMultipleChoice,
		Result:          ResultCorrect,
		Rating:          RatingGood,
		AnsweredAt:      now,
		Source:          SourceReview,
		ClientAttemptID: "ca-1",
		IdempotencyKey:  "review-test-key",
	}
	attempt, err := repo.SubmitReview(t.Context(), req)
	require.NoError(t, err)
	assert.Equal(t, existingAttemptID, attempt.ID)
	require.NoError(t, mock.ExpectationsWereMet())
}

// TestPostgreSQLRepositorySubmitReviewP4AlreadyCompletedSnapshotNoDoubleReward
// proves the second line of defense against a double reward: when the
// mission was already completed by an earlier call in the same local
// day, the second call's MarkSnapshotCompleted update affects 0 rows
// (the WHERE status='open' guard) and the +10 daily-mission reward is
// not re-issued. The ledger's (user_id, idempotency_key) partial unique
// index would also catch it at the SQL level, but the snapshot guard
// keeps the +10 grant from being attempted at all.
func TestPostgreSQLRepositorySubmitReviewP4AlreadyCompletedSnapshotNoDoubleReward(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	now := time.Date(2026, 7, 26, 12, 0, 0, 0, time.UTC)
	gamSvc := gamification.NewService(gamification.NewRepository(db))
	missionsSvc := missions.NewService(missions.NewRepository(db), gamSvc)
	repo := NewPostgreSQLRepository(db, clock.Fixed{T: now},
		WithGamificationService(gamSvc),
		WithMissionsService(missionsSvc),
	)

	userID := p4TestID(t, "00000000-0000-0000-0000-000000000001")
	userWordID := p4TestID(t, "00000000-0000-0000-0000-000000000002")
	meaningID := p4TestID(t, "00000000-0000-0000-0000-000000000003")
	day := time.Date(2026, 7, 26, 0, 0, 0, 0, time.UTC)

	mock.ExpectBegin()
	mock.ExpectQuery("SELECT review_step, meaning_id").
		WithArgs(userWordID, userID).
		WillReturnRows(sqlmock.NewRows([]string{"review_step", "meaning_id", "total_review_count", "correct_review_count", "consecutive_correct_count", "consecutive_incorrect_count"}).
			AddRow(0, meaningID, 0, 0, 0, 0))
	expectReviewKeyClaim(mock)
	mock.ExpectQuery("SELECT ra.id, ra.user_word_id").
		WithArgs(userID, "ca-2nd").
		WillReturnRows(sqlmock.NewRows([]string{"id", "user_word_id", "meaning_id", "attempt_type", "prompt_type", "result", "rating", "review_step_before", "review_step_after", "answered_at", "response_time_ms", "selected_option_meaning_id", "typed_answer", "was_hint_used", "source", "client_attempt_id", "metadata", "next_review_at"}))
	mock.ExpectExec("INSERT INTO review_attempts").
		WithArgs(sqlmock.AnyArg(), userID, userWordID, meaningID, "review", "multiple_choice", "correct", "good", 0, 1, now, 0, nil, nil, false, "review", "ca-2nd", nil, now, now).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("UPDATE user_words").
		WithArgs(1, sqlmock.AnyArg(), now, "correct", "good", 1, 1, 1, 0, now, userWordID, userID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	mock.ExpectQuery("SELECT user_id, timezone, daily_review_target").
		WithArgs(userID).
		WillReturnRows(sqlmock.NewRows([]string{"user_id", "timezone", "daily_review_target", "review_interval_preset", "notifications_enabled", "marketing_emails_enabled", "app_language"}).
			AddRow(userID, "UTC", 20, "vocanova_default", false, false, "en"))
	// Snapshot is already at target and status='completed'.
	mock.ExpectQuery("INSERT INTO daily_mission_snapshots").
		WithArgs(sqlmock.AnyArg(), userID, day, "UTC", 20, gamification.MissionPolicyVersion).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "user_id", "local_date", "timezone", "review_target", "reviews_completed",
			"new_word_target", "new_words_completed", "sentence_practice_target",
			"sentence_practices_completed", "policy_version", "status", "completed_at",
			"grace_applied", "grace_day_id",
		}).AddRow(
			uuid.New(), userID, day, "UTC", 20, 20,
			nil, nil, nil, nil,
			gamification.MissionPolicyVersion, "completed", now, false, nil,
		))
	mock.ExpectQuery("UPDATE daily_mission_snapshots").
		WithArgs(userID, day).
		WillReturnRows(sqlmock.NewRows([]string{"reviews_completed"}).AddRow(20))
	mock.ExpectExec("INSERT INTO daily_activity_summaries").
		WillReturnResult(sqlmock.NewResult(0, 1))
	// +5 Good reward (still awarded — every individual review attempt earns
	// its rating-tiered point, even on a day that's already complete).
	mock.ExpectQuery("SELECT COALESCE\\(balance_after, 0\\) FROM confidence_point_ledger").
		WithArgs(userID).
		WillReturnRows(sqlmock.NewRows([]string{"balance_after"}))
	mock.ExpectQuery("INSERT INTO confidence_point_ledger").
		WithArgs(
			sqlmock.AnyArg(), userID, gamification.RewardReviewGood, gamification.RewardReviewGood,
			gamification.ReasonReviewCorrect, gamification.SourceReviewAttempt,
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), now,
		).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(uuid.New()))
	mock.ExpectExec("INSERT INTO daily_activity_summaries").
		WithArgs(sqlmock.AnyArg(), userID, day, "UTC", gamification.RewardReviewGood).
		WillReturnResult(sqlmock.NewResult(0, 1))
	// NO MarkSnapshotCompleted (reviews_completed >= target but snapshot is
	// already 'completed' so the comparison is false). NO second GrantPoint
	// for the +10. NO daily-mission-completion idempotency key grant.
	// Streak reconciliation still runs (read-only update).
	mock.ExpectQuery("SELECT local_date, status, completed_at, grace_applied, grace_day_id FROM daily_mission_snapshots").
		WithArgs(userID, 14).
		WillReturnRows(sqlmock.NewRows([]string{"local_date", "status", "completed_at", "grace_applied", "grace_day_id"}).
			AddRow(day, "completed", now, false, nil))
	mock.ExpectQuery("SELECT COALESCE\\(balance_after, 0\\) FROM grace_day_ledger").
		WithArgs(userID).
		WillReturnRows(sqlmock.NewRows([]string{"balance_after"}))
	mock.ExpectQuery("SELECT user_id, current_streak_count, longest_streak_count").
		WithArgs(userID).
		WillReturnRows(sqlmock.NewRows([]string{"user_id", "current_streak_count", "longest_streak_count", "last_completed_local_date", "last_activity_local_date", "timezone", "status", "created_at", "updated_at"}).
			AddRow(userID, 1, 1, day, day, "UTC", "active", now, now))
	mock.ExpectExec("INSERT INTO streak_states").
		WillReturnResult(sqlmock.NewResult(0, 1))

	mock.ExpectCommit()

	req := SubmitReviewRequest{
		UserID:          userID,
		UserWordID:      userWordID,
		MeaningID:       meaningID,
		AttemptType:     AttemptTypeReview,
		PromptType:      PromptTypeMultipleChoice,
		Result:          ResultCorrect,
		Rating:          RatingGood,
		AnsweredAt:      now,
		ResponseTimeMs:  0,
		Source:          SourceReview,
		ClientAttemptID: "ca-2nd",
		IdempotencyKey:  "review-test-key",
	}
	attempt, err := repo.SubmitReview(t.Context(), req)
	require.NoError(t, err)
	assert.Equal(t, 1, attempt.ReviewStepAfter)
	require.NoError(t, mock.ExpectationsWereMet())
}
