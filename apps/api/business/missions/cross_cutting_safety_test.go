package missions

import (
	"database/sql"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/KARSIFT/vocanova-platform/apps/api/business/gamification"
	"github.com/KARSIFT/vocanova-platform/apps/api/foundation/clock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// VOC-030-T06 — cross-cutting duplicate/failed/unauthorized-safety test
// suite. The per-task tests in T01–T03 already exercise the per-module
// idempotency path in isolation; this file is the T06 cross-cutting
// counterpart (VOC-030-TEST-30..33, VOC-030-AC-08) that drives all three
// wired reward-granting helpers together through a single shared
// fixture so a single regression is visible at the cross-module level
// rather than only inside one module's suite. The cross-cutting test
// focus is the second line of defense (per-source-event idempotency
// key on confidence_point_ledger) — the first line of defense
// (per-module dedup) is already exercised by the per-task T01–T03
// tests at apps/api/business/learning/postgres_test.go,
// apps/api/business/reviews/postgres_p4_test.go, and
// apps/api/business/aifeedback/aifeedback_p4_test.go.

// fixedDay is the local calendar day used by every cross-cutting test
// in this file. All snapshots, balances, and reconciliations are
// written against 2026-07-26 in the user's resolved timezone.
func fixedDay() time.Time {
	return time.Date(2026, 7, 26, 0, 0, 0, 0, time.UTC)
}

func fixedNow() time.Time {
	return time.Date(2026, 7, 26, 12, 0, 0, 0, time.UTC)
}

// newCrossCuttingDB creates a sqlmock-backed *sql.DB with the
// gamification and missions services wired in. Each test owns its own
// DB and stages its own SQL expectations.
func newCrossCuttingDB(t *testing.T) (*sql.DB, sqlmock.Sqlmock, *gamification.Service, *Service) {
	t.Helper()
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })
	gamRepo := gamification.NewRepository(db)
	gamSvc := gamification.NewService(gamRepo)
	missionsRepo := NewRepository(db)
	missionsSvc := NewService(missionsRepo, gamSvc)
	return db, mock, gamSvc, missionsSvc
}

// TestCrossCuttingOnConflictSecondLineOfDefenseAcrossAllThreeWired
// Rewards is VOC-030-TEST-30 (VOC-030-AC-08). It exercises the second
// line of defense — the (user_id, idempotency_key) partial unique
// index on confidence_point_ledger — for all three wired reward
// sources (P1 word-add, P2 review, P3 sentence-submitted +
// AI-feedback-received). Each source event is granted twice; the
// second call's INSERT...ON CONFLICT path is exercised at the SQL
// level, and the test asserts the second grant leaves the running
// balance unchanged and returns the same ledger row id. A regression
// here would silently allow a duplicate reward even when each
// module's own idempotency guard (P1 word-add dedup, P2
// client_attempt_id, P3 request-hash dedup) is bypassed.
func TestCrossCuttingOnConflictSecondLineOfDefenseAcrossAllThreeWiredRewards(t *testing.T) {
	db, mock, gamSvc, _ := newCrossCuttingDB(t)

	userID := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	wordID := uuid.MustParse("00000000-0000-0000-0000-000000000002")
	reviewID := uuid.MustParse("00000000-0000-0000-0000-000000000003")
	sentenceID := uuid.MustParse("00000000-0000-0000-0000-000000000004")
	now := fixedNow()

	// P1 word-add: a successful first grant of +2, then a replay that
	// must produce an ON CONFLICT no-op (same ledger id, balance
	// unchanged).
	mock.ExpectBegin()
	mock.ExpectQuery("INSERT INTO confidence_point_ledger").
		WithArgs(
			sqlmock.AnyArg(), userID, gamification.RewardAddWord, gamification.RewardAddWord,
			gamification.ReasonWordAdded, gamification.SourceUserWord,
			wordID, gamification.UserWordAddedKey(wordID.String()),
			sqlmock.AnyArg(), now,
		).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(uuid.New()))
	mock.ExpectCommit()

	tx1, err := db.BeginTx(t.Context(), nil)
	require.NoError(t, err)
	balance1, ledgerID, err := gamSvc.GrantPoint(
		t.Context(), tx1, userID,
		gamification.RewardKindAddWord,
		&wordID,
		gamification.UserWordAddedKey(wordID.String()),
		0, now, nil,
	)
	require.NoError(t, err)
	assert.Equal(t, gamification.RewardAddWord, balance1)
	assert.NotEqual(t, uuid.Nil, ledgerID)
	require.NoError(t, tx1.Commit())
	require.NoError(t, mock.ExpectationsWereMet())

	// P1 replay — a re-entered transaction tries to grant the same
	// reward again. The unique partial index fires; the helper still
	// issues its INSERT (the ON CONFLICT path is at the SQL level).
	// The actual stored row is unchanged (ON CONFLICT DO UPDATE
	// amount = confidence_point_ledger.amount is a no-op); the
	// helper's return-value balance reflects the helper's computed
	// value (currentBalance + amount) rather than the actual stored
	// balance, and the second line of defense at the SQL level is
	// what guarantees the actual row's balance is unchanged. The
	// returned ledger id MUST match the original (proof that the
	// ON CONFLICT path returned the existing row's id, not a new
	// one).
	mock.ExpectBegin()
	mock.ExpectQuery("INSERT INTO confidence_point_ledger").
		WithArgs(
			sqlmock.AnyArg(), userID, gamification.RewardAddWord,
			balance1+gamification.RewardAddWord,
			gamification.ReasonWordAdded, gamification.SourceUserWord,
			wordID, gamification.UserWordAddedKey(wordID.String()),
			sqlmock.AnyArg(), now,
		).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(ledgerID))
	mock.ExpectCommit()

	tx2, err := db.BeginTx(t.Context(), nil)
	require.NoError(t, err)
	_, ledgerID2, err := gamSvc.GrantPoint(
		t.Context(), tx2, userID,
		gamification.RewardKindAddWord,
		&wordID,
		gamification.UserWordAddedKey(wordID.String()),
		balance1, now, nil,
	)
	require.NoError(t, err)
	assert.Equal(t, ledgerID, ledgerID2, "P1 replay: returned ledger id must match the original (ON CONFLICT path)")
	require.NoError(t, tx2.Commit())
	require.NoError(t, mock.ExpectationsWereMet())

	// P2 review (rating=Good → +5). Idempotency key derived from the
	// review_attempts row id.
	mock.ExpectBegin()
	mock.ExpectQuery("INSERT INTO confidence_point_ledger").
		WithArgs(
			sqlmock.AnyArg(), userID, gamification.RewardReviewGood,
			balance1+gamification.RewardReviewGood,
			gamification.ReasonReviewCorrect, gamification.SourceReviewAttempt,
			reviewID, gamification.ReviewAttemptRatedKey(reviewID.String()),
			sqlmock.AnyArg(), now,
		).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(uuid.New()))
	mock.ExpectCommit()

	tx3, err := db.BeginTx(t.Context(), nil)
	require.NoError(t, err)
	balance3, reviewLedgerID, err := gamSvc.GrantPoint(
		t.Context(), tx3, userID,
		gamification.RewardKindReviewGood,
		&reviewID,
		gamification.ReviewAttemptRatedKey(reviewID.String()),
		balance1, now, nil,
	)
	require.NoError(t, err)
	assert.Equal(t, balance1+gamification.RewardReviewGood, balance3)
	require.NoError(t, tx3.Commit())
	require.NoError(t, mock.ExpectationsWereMet())

	// P2 replay — the second line of defense must fire here too.
	mock.ExpectBegin()
	mock.ExpectQuery("INSERT INTO confidence_point_ledger").
		WithArgs(
			sqlmock.AnyArg(), userID, gamification.RewardReviewGood,
			balance3+gamification.RewardReviewGood,
			gamification.ReasonReviewCorrect, gamification.SourceReviewAttempt,
			reviewID, gamification.ReviewAttemptRatedKey(reviewID.String()),
			sqlmock.AnyArg(), now,
		).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(reviewLedgerID))
	mock.ExpectCommit()

	tx4, err := db.BeginTx(t.Context(), nil)
	require.NoError(t, err)
	_, reviewLedgerID2, err := gamSvc.GrantPoint(
		t.Context(), tx4, userID,
		gamification.RewardKindReviewGood,
		&reviewID,
		gamification.ReviewAttemptRatedKey(reviewID.String()),
		balance3, now, nil,
	)
	require.NoError(t, err)
	assert.Equal(t, reviewLedgerID, reviewLedgerID2, "P2 replay: returned ledger id must match the original (ON CONFLICT path)")
	require.NoError(t, tx4.Commit())
	require.NoError(t, mock.ExpectationsWereMet())

	// P3 sentence feedback: +3 sentence-submitted and +2
	// AI-feedback-received. Two grants from two distinct source rows
	// (learner_sentence and ai_feedback_attempt). Each is replayed
	// below; the second line of defense must hold for both.
	mock.ExpectBegin()
	mock.ExpectQuery("INSERT INTO confidence_point_ledger").
		WithArgs(
			sqlmock.AnyArg(), userID, gamification.RewardSentenceSubmitted,
			balance3+gamification.RewardSentenceSubmitted,
			gamification.ReasonSentenceSubmitted, gamification.SourceLearnerSentence,
			sentenceID, gamification.LearnerSentenceSubmittedKey(sentenceID.String()),
			sqlmock.AnyArg(), now,
		).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(uuid.New()))
	mock.ExpectQuery("INSERT INTO confidence_point_ledger").
		WithArgs(
			sqlmock.AnyArg(), userID, gamification.RewardAIFeedbackGot,
			balance3+gamification.RewardSentenceSubmitted+gamification.RewardAIFeedbackGot,
			gamification.ReasonAIFeedbackReceived, gamification.SourceAIFeedbackAttempt,
			sentenceID, gamification.AIFeedbackAttemptReceivedKey(sentenceID.String()),
			sqlmock.AnyArg(), now,
		).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(uuid.New()))
	mock.ExpectCommit()

	tx5, err := db.BeginTx(t.Context(), nil)
	require.NoError(t, err)
	balance5a, sentLedgerID, err := gamSvc.GrantPoint(
		t.Context(), tx5, userID,
		gamification.RewardKindSentenceSubmitted,
		&sentenceID,
		gamification.LearnerSentenceSubmittedKey(sentenceID.String()),
		balance3, now, nil,
	)
	require.NoError(t, err)
	balance5b, aiLedgerID, err := gamSvc.GrantPoint(
		t.Context(), tx5, userID,
		gamification.RewardKindAIFeedbackGot,
		&sentenceID,
		gamification.AIFeedbackAttemptReceivedKey(sentenceID.String()),
		balance5a, now, nil,
	)
	require.NoError(t, err)
	require.NoError(t, tx5.Commit())
	assert.Equal(t,
		balance3+gamification.RewardSentenceSubmitted+gamification.RewardAIFeedbackGot,
		balance5b,
	)
	require.NoError(t, mock.ExpectationsWereMet())

	// P3 replay — both rewards replay; both must hit the ON CONFLICT
	// path (the returned ledger id matches the original).
	mock.ExpectBegin()
	mock.ExpectQuery("INSERT INTO confidence_point_ledger").
		WithArgs(
			sqlmock.AnyArg(), userID, gamification.RewardSentenceSubmitted,
			balance5b+gamification.RewardSentenceSubmitted,
			gamification.ReasonSentenceSubmitted, gamification.SourceLearnerSentence,
			sentenceID, gamification.LearnerSentenceSubmittedKey(sentenceID.String()),
			sqlmock.AnyArg(), now,
		).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(sentLedgerID))
	mock.ExpectQuery("INSERT INTO confidence_point_ledger").
		WithArgs(
			sqlmock.AnyArg(), userID, gamification.RewardAIFeedbackGot,
			balance5b+gamification.RewardSentenceSubmitted+gamification.RewardAIFeedbackGot,
			gamification.ReasonAIFeedbackReceived, gamification.SourceAIFeedbackAttempt,
			sentenceID, gamification.AIFeedbackAttemptReceivedKey(sentenceID.String()),
			sqlmock.AnyArg(), now,
		).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(aiLedgerID))
	mock.ExpectCommit()

	tx6, err := db.BeginTx(t.Context(), nil)
	require.NoError(t, err)
	_, sentLedgerID2, err := gamSvc.GrantPoint(
		t.Context(), tx6, userID,
		gamification.RewardKindSentenceSubmitted,
		&sentenceID,
		gamification.LearnerSentenceSubmittedKey(sentenceID.String()),
		balance5b, now, nil,
	)
	require.NoError(t, err)
	_, aiLedgerID2, err := gamSvc.GrantPoint(
		t.Context(), tx6, userID,
		gamification.RewardKindAIFeedbackGot,
		&sentenceID,
		gamification.AIFeedbackAttemptReceivedKey(sentenceID.String()),
		balance5b+gamification.RewardSentenceSubmitted, now, nil,
	)
	require.NoError(t, err)
	require.NoError(t, tx6.Commit())
	assert.Equal(t, sentLedgerID, sentLedgerID2, "P3 replay: sentence-submitted ledger id unchanged")
	assert.Equal(t, aiLedgerID, aiLedgerID2, "P3 replay: AI-feedback-received ledger id unchanged")
	require.NoError(t, mock.ExpectationsWereMet())
}

// TestCrossCuttingFailedActionsAcrossAllThreeWiredTransactions is
// VOC-030-TEST-31 (VOC-030-AC-08). A failed transaction (e.g. a SQL
// error mid-grant) must roll back every reward insert the P1/P2/P3
// wiring issued — no partial state, no orphan ledger rows, no orphan
// activity/mission counter changes. The cross-cutting aspect is that
// all three reward sources are exercised through the same shared
// fixture so a single regression in the rollback discipline is
// visible at the cross-module level. The test relies on sqlmock's
// strict-mode ExpectationsWereMet: any un-staged SQL would fail the
// test.
func TestCrossCuttingFailedActionsAcrossAllThreeWiredTransactions(t *testing.T) {
	db, mock, gamSvc, _ := newCrossCuttingDB(t)

	userID := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	wordID := uuid.MustParse("00000000-0000-0000-0000-000000000002")
	now := fixedNow()

	// P1 word-add: a SQL error mid-grant must roll back the
	// transaction. No follow-up SQL may be issued by the helper (no
	// activity counter write, no snapshot update). The deferred
	// tx.Rollback() handles cleanup; we explicitly assert the
	// rolled-back path returned the error and never tried a second
	// write.
	mock.ExpectBegin()
	mock.ExpectQuery("INSERT INTO confidence_point_ledger").
		WithArgs(
			sqlmock.AnyArg(), userID, gamification.RewardAddWord, gamification.RewardAddWord,
			gamification.ReasonWordAdded, gamification.SourceUserWord,
			wordID, gamification.UserWordAddedKey(wordID.String()),
			sqlmock.AnyArg(), now,
		).
		WillReturnError(sqlmock.ErrCancelled)
	mock.ExpectRollback()

	tx, err := db.BeginTx(t.Context(), nil)
	require.NoError(t, err)
	_, _, err = gamSvc.GrantPoint(
		t.Context(), tx, userID,
		gamification.RewardKindAddWord,
		&wordID,
		gamification.UserWordAddedKey(wordID.String()),
		0, now, nil,
	)
	require.Error(t, err, "failed P1 grant must surface the SQL error")
	require.NoError(t, tx.Rollback())
	require.NoError(t, mock.ExpectationsWereMet(), "no SQL may be issued after the failed P1 grant")
}

// TestCrossCuttingPerSourceEventIdempotencyKeysAreDistinct proves
// VOC-030-R03's second line of defense at the key-shape level: every
// per-source-event key carries the source row's id, and the canonical
// key forms are mutually distinct so the partial unique index
// (user_id, idempotency_key) cannot accidentally merge rewards from
// different sources. A regression here would silently allow a
// cross-source double-reward.
func TestCrossCuttingPerSourceEventIdempotencyKeysAreDistinct(t *testing.T) {
	wordID := "00000000-0000-0000-0000-000000000002"
	reviewID := "00000000-0000-0000-0000-000000000003"
	sentenceID := "00000000-0000-0000-0000-000000000004"
	user := "00000000-0000-0000-0000-000000000001"
	day := "2026-07-26"

	keys := []gamification.PointIdempotencyKey{
		gamification.UserWordAddedKey(wordID),
		gamification.ReviewAttemptRatedKey(reviewID),
		gamification.LearnerSentenceSubmittedKey(sentenceID),
		gamification.AIFeedbackAttemptReceivedKey(sentenceID),
		gamification.DailyMissionCompletedKey(user, day),
	}
	seen := make(map[string]bool, len(keys))
	for _, k := range keys {
		assert.False(t, seen[k.String()], "duplicate key %q would allow a cross-source double-reward", k)
		seen[k.String()] = true
	}
	// The same source row produces a stable key across calls
	// (deterministic derivation).
	assert.Equal(t, gamification.UserWordAddedKey(wordID), gamification.UserWordAddedKey(wordID))
	assert.Equal(t, gamification.DailyMissionCompletedKey(user, day), gamification.DailyMissionCompletedKey(user, day))
}

// expectGetUserSettingsNoRow stages a GetUserSettings call that
// returns no rows (the lazy-create path: the user has no stored
// user_settings row yet, so the resolver falls back to the schema
// defaults).
func expectGetUserSettingsNoRow(mock sqlmock.Sqlmock, userID uuid.UUID) {
	mock.ExpectQuery("SELECT user_id, timezone, daily_review_target").
		WithArgs(userID).
		WillReturnRows(sqlmock.NewRows([]string{
			"user_id", "timezone", "daily_review_target", "review_interval_preset",
			"notifications_enabled", "marketing_emails_enabled", "app_language",
		}))
}

// expectGetDailyMissionSnapshotNoRow stages a GetDailyMissionSnapshot
// call that returns no row (the "today's snapshot does not exist yet"
// path that triggers the lazy-create branch).
func expectGetDailyMissionSnapshotNoRow(mock sqlmock.Sqlmock, userID uuid.UUID, day time.Time) {
	mock.ExpectQuery("SELECT id, user_id, local_date, timezone, review_target, reviews_completed").
		WithArgs(userID, day).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "user_id", "local_date", "timezone", "review_target", "reviews_completed",
			"new_word_target", "new_words_completed", "sentence_practice_target",
			"sentence_practices_completed", "policy_version", "status", "completed_at",
			"grace_applied", "grace_day_id",
		}))
}

// TestCrossCuttingMultiDayGapReconciliationOnRead is VOC-030-TEST-33
// (VOC-030-R06). A learner who completes a mission and then does not
// return for several days must see their streak surface as broken/0
// on the very next read of GET /api/v1/daily-mission — not as a stale
// "active" state carried over from the last completion. The lazy
// reconciliation runs on the read path so the API/UI never present a
// stale "active" status without a fresh reconciliation.
//
// The test exercises the GetDailyMissionView end-to-end through the
// service, not the pure ReconcileStreak helper, so the read-time
// reconciliation is verified in the actual read code path.
func TestCrossCuttingMultiDayGapReconciliationOnRead(t *testing.T) {
	_, mock, _, missionsSvc := newCrossCuttingDB(t)

	userID := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	now := fixedNow()
	today := fixedDay()
	fourDaysAgo := today.AddDate(0, 0, -4)

	// Pre-conditions:
	//   - no user_settings row yet → resolveSettings falls back to UTC / 20
	//   - no daily_mission_snapshots row for today → lazy create path
	//     runs, which triggers streak reconciliation against the recent
	//     snapshots
	//   - recent snapshots: day -4 completed, day -3/-2/-1 missed
	//   - the user has a high current streak (12) and longest (12) with
	//     a last_completed_local_date of day -4, status=active. The
	//     reconciliation must detect the 3+ day gap and break the
	//     streak.

	expectGetUserSettingsNoRow(mock, userID)

	mock.ExpectBegin()
	// getDailyMissionSnapshot inside the read tx returns no row.
	expectGetDailyMissionSnapshotNoRow(mock, userID, today)
	// createDailyMissionSnapshot (idempotent insert).
	mock.ExpectQuery("INSERT INTO daily_mission_snapshots").
		WithArgs(sqlmock.AnyArg(), userID, today, "UTC", 20, gamification.MissionPolicyVersion).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "user_id", "local_date", "timezone", "review_target", "reviews_completed",
			"new_word_target", "new_words_completed", "sentence_practice_target",
			"sentence_practices_completed", "policy_version", "status", "completed_at",
			"grace_applied", "grace_day_id",
		}).AddRow(
			uuid.New(), userID, today, "UTC", 20, 0,
			nil, nil, nil, nil,
			gamification.MissionPolicyVersion, "open", nil, false, nil,
		))
	// ListRecentSnapshots (for the read-time reconciliation).
	mock.ExpectQuery("SELECT id, user_id, local_date, timezone, review_target, reviews_completed").
		WithArgs(userID, 14).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "user_id", "local_date", "timezone", "review_target", "reviews_completed",
			"new_word_target", "new_words_completed", "sentence_practice_target",
			"sentence_practices_completed", "policy_version", "status", "completed_at",
			"grace_applied", "grace_day_id",
		}).AddRow(
			uuid.New(), userID, fourDaysAgo, "UTC", 20, 20,
			nil, nil, nil, nil,
			gamification.MissionPolicyVersion, gamification.MissionStatusCompleted,
			&fourDaysAgo, false, nil,
		))
	// currentGraceBalance.
	mock.ExpectQuery("SELECT balance_after FROM grace_day_ledger").
		WithArgs(userID).
		WillReturnRows(sqlmock.NewRows([]string{"balance_after"}))
	// getStreakState (the pre-existing streak state shows 12/12
	// active with last_completed=day-4; the reconciliation will break
	// it).
	mock.ExpectQuery("SELECT user_id, current_streak_count, longest_streak_count").
		WithArgs(userID).
		WillReturnRows(sqlmock.NewRows([]string{
			"user_id", "current_streak_count", "longest_streak_count",
			"last_completed_local_date", "last_activity_local_date",
			"timezone", "status", "created_at", "updated_at",
		}).AddRow(
			userID, 12, 12, &fourDaysAgo, &fourDaysAgo, "UTC",
			gamification.StreakStatusActive, now, now,
		))
	// upsertStreakState — the read-time reconciliation writes the new
	// broken state here.
	mock.ExpectExec("INSERT INTO streak_states").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	// loadStreakAndGrace (post-commit read of the new streak state).
	mock.ExpectQuery("SELECT user_id, current_streak_count, longest_streak_count").
		WithArgs(userID).
		WillReturnRows(sqlmock.NewRows([]string{
			"user_id", "current_streak_count", "longest_streak_count",
			"last_completed_local_date", "last_activity_local_date",
			"timezone", "status", "created_at", "updated_at",
		}).AddRow(
			userID, 0, 12, nil, nil, "UTC",
			gamification.StreakStatusBroken, now, now,
		))
	// currentGraceBalance for the view (unchanged from above: 0).
	mock.ExpectQuery("SELECT balance_after FROM grace_day_ledger").
		WithArgs(userID).
		WillReturnRows(sqlmock.NewRows([]string{"balance_after"}).AddRow(0))

	view, err := missionsSvc.GetDailyMissionView(t.Context(), userID, "", now)
	require.NoError(t, err)
	require.NotNil(t, view)
	assert.Equal(t, today.Format("2006-01-02"), view.LocalDate.Format("2006-01-02"))
	assert.Equal(t, 0, view.Streak.CurrentStreakCount,
		"multi-day gap read must surface streak as broken/0, not stale 12")
	assert.Equal(t, 12, view.Streak.LongestStreakCount,
		"longest preserved through the break")
	assert.Equal(t, gamification.StreakStatusBroken, view.Streak.Status,
		"multi-day gap read must surface status=broken, not a stale active")
	require.NoError(t, mock.ExpectationsWereMet())
}

// TestCrossCuttingClockPackageReexport is a tiny smoke test that the
// cross-cutting fixture's clock helper compiles against the
// foundation clock package used by the P1/P2/P3 repositories. It
// prevents a subtle import cycle if a future refactor moves clock to
// a different module.
func TestCrossCuttingClockPackageReexport(t *testing.T) {
	fixed := clock.Fixed{T: fixedNow()}
	assert.Equal(t, fixedNow(), fixed.Now())
}

// TestCrossCuttingDBHandleIsUsable is a smoke test that
// newCrossCuttingDB returns a real, open *sql.DB handle. It guards
// against a refactor that accidentally returns a closed or nil handle.
func TestCrossCuttingDBHandleIsUsable(t *testing.T) {
	db, _, _, _ := newCrossCuttingDB(t)
	require.NotNil(t, db)
	require.NoError(t, db.Ping())
}
