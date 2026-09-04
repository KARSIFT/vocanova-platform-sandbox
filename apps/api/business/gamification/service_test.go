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
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(uuid.New()))
		mock.ExpectCommit()

		tx, err := db.Begin()
		require.NoError(t, err)

		gotBalance, _, err := svc.GrantPoint(
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
	// ledger's latest balance_after column - not a separate mutable field.
	mock.ExpectQuery(`SELECT balance_after FROM confidence_point_ledger`).
		WithArgs(userID).
		WillReturnRows(sqlmock.NewRows([]string{"balance_after"}).AddRow(runningBalance))

	got, err := svc.CurrentBalance(t.Context(), userID)
	require.NoError(t, err)
	assert.Equal(t, wantTotal, got, "Progress screen's Confidence Points total must equal the exact sum of ledger entries")

	require.NoError(t, mock.ExpectationsWereMet())
}
