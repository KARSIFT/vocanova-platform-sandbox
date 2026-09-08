package gamification

import (
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestCurrentBalanceMatchesSumOfLedgerEntries is VOC-1178's verification that
// the Confidence Points total the Progress screen reads
// (missions.Service.GetProgressView -> gamification.Service.CurrentBalance)
// is exactly the sum of the confidence_point_ledger entries granted for a
// user - never a value that could drift from a separately maintained
// mutable balance.
//
// It exercises the same sequence a real caller follows: read the current
// balance, grant a point (which persists currentBalance+amount as the new
// row's balance_after), repeat. After several grants of different
// RewardKinds, it reads the balance back via CurrentBalance (the exact call
// GetProgressView makes) and asserts it equals the arithmetic sum of every
// granted amount - not a hard-coded expectation, so the test would fail if
// GrantPoint's running-balance arithmetic, or CurrentBalance's read query,
// ever diverged from "sum of the ledger".
func TestCurrentBalanceMatchesSumOfLedgerEntries(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	svc := NewService(NewRepository(db))
	userID := uuid.MustParse("00000000-0000-0000-0000-0000000000aa")
	now := time.Date(2026, 9, 1, 12, 0, 0, 0, time.UTC)

	grants := []struct {
		kind PointIdempotencyKey
		rk   RewardKind
	}{
		{ReviewAttemptRatedKey("r1"), RewardKindReviewGood},
		{UserWordAddedKey("w1"), RewardKindAddWord},
		{LearnerSentenceSubmittedKey("s1"), RewardKindSentenceSubmitted},
		{ReviewAttemptRatedKey("r2"), RewardKindReviewEasy},
		{DailyMissionCompletedKey(userID.String(), "2026-09-01"), RewardKindDailyMissionDone},
	}

	wantTotal := 0
	runningBalance := 0
	for _, g := range grants {
		outcome, err := RewardFor(g.rk)
		require.NoError(t, err)
		wantTotal += outcome.Amount

		mock.ExpectBegin()
		newBalanceAfterInsert := runningBalance + outcome.Amount
		mock.ExpectQuery(pointLedgerInsertColumnsPattern).
			WithArgs(
				sqlmock.AnyArg(), userID, outcome.Amount, newBalanceAfterInsert,
				outcome.Reason, outcome.SourceType, sqlmock.AnyArg(),
				g.kind.String(), sqlmock.AnyArg(), now,
			).
			WillReturnRows(sqlmock.NewRows([]string{"id", "inserted"}).AddRow(uuid.New(), true))
		mock.ExpectCommit()

		tx, err := db.Begin()
		require.NoError(t, err)

		gotBalance, _, _, err := svc.GrantPoint(
			t.Context(), tx, userID, g.rk, nil, g.kind, runningBalance, now, nil,
		)
		require.NoError(t, err)
		require.NoError(t, tx.Commit())

		runningBalance = gotBalance
	}

	// runningBalance is exactly the sum of every grant's Amount: nothing in
	// GrantPoint's arithmetic can silently diverge from the ledger.
	assert.Equal(t, wantTotal, runningBalance)

	// The Progress screen's read path (missions.Service.GetProgressView)
	// calls exactly this method. It must read the same total back from the
	// sum of immutable ledger amounts - not a separate mutable field.
	mock.ExpectQuery(`SELECT COALESCE\(SUM\(amount\), 0\) FROM confidence_point_ledger`).
		WithArgs(userID).
		WillReturnRows(sqlmock.NewRows([]string{"sum"}).AddRow(runningBalance))

	got, err := svc.CurrentBalance(t.Context(), userID)
	require.NoError(t, err)
	assert.Equal(t, wantTotal, got, "Progress screen's Confidence Points total must equal the exact sum of ledger entries")

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestGrantPointDuplicateKeepsCurrentBalance(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	svc := NewService(NewRepository(db))
	userID := uuid.MustParse("00000000-0000-0000-0000-0000000000ac")
	now := time.Date(2026, 9, 8, 12, 0, 0, 0, time.UTC)
	key := LearnerSentenceSubmittedKey("sentence-1")

	mock.ExpectBegin()
	mock.ExpectQuery(pointLedgerInsertColumnsPattern).
		WithArgs(
			sqlmock.AnyArg(), userID, RewardSentenceSubmitted, 8,
			ReasonSentenceSubmitted, SourceLearnerSentence, sqlmock.AnyArg(),
			key.String(), sqlmock.AnyArg(), now,
		).
		WillReturnRows(sqlmock.NewRows([]string{"id", "inserted"}).AddRow(uuid.New(), false))
	mock.ExpectCommit()

	tx, err := db.Begin()
	require.NoError(t, err)
	balance, _, created, err := svc.GrantPoint(
		t.Context(), tx, userID, RewardKindSentenceSubmitted, nil, key, 5, now, nil,
	)
	require.NoError(t, err)
	require.NoError(t, tx.Commit())
	assert.False(t, created)
	assert.Equal(t, 5, balance, "a replay must not claim a second +3 award")
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestReconcileAndAdvanceUsesResolvedTimezoneForExistingStreakState(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	svc := NewService(NewRepository(db))
	userID := uuid.MustParse("00000000-0000-0000-0000-0000000000bb")
	now := time.Date(2026, 9, 4, 15, 30, 0, 0, time.UTC) // Sep 5 in Tokyo, Sep 4 in UTC.
	localDay := time.Date(2026, 9, 5, 0, 0, 0, 0, time.UTC)

	// A pre-existing state was written before the learner changed their
	// settings from UTC to Asia/Tokyo. Reconciliation must use and persist the
	// resolved timezone, rather than keeping its obsolete timezone forever.
	mock.ExpectBegin()
	mock.ExpectExec("SELECT pg_advisory_xact_lock").WithArgs(userID.String()).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery("SELECT COALESCE\\(SUM\\(amount\\), 0\\) FROM grace_day_ledger").WithArgs(userID).
		WillReturnRows(sqlmock.NewRows([]string{"balance_after"}).AddRow(0))
	mock.ExpectQuery("SELECT user_id, current_streak_count, longest_streak_count").
		WithArgs(userID).
		WillReturnRows(sqlmock.NewRows([]string{
			"user_id", "current_streak_count", "longest_streak_count",
			"last_completed_local_date", "last_activity_local_date",
			"timezone", "status", "created_at", "updated_at",
		}).AddRow(userID, 4, 4, nil, nil, "UTC", StreakStatusActive, now, now))
	mock.ExpectExec("INSERT INTO streak_states").
		WithArgs(sqlmock.AnyArg(), userID, 4, 4, nil, localDay, "Asia/Tokyo", StreakStatusActive).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	tx, err := db.BeginTx(t.Context(), nil)
	require.NoError(t, err)
	_, err = svc.ReconcileAndAdvance(t.Context(), tx, userID, now, "Asia/Tokyo", nil, false)
	require.NoError(t, err)
	require.NoError(t, tx.Commit())
	require.NoError(t, mock.ExpectationsWereMet())
}
