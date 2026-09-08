//go:build integration

package gamification_test

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/KARSIFT/vocanova-platform/apps/api/business/accounts"
	"github.com/KARSIFT/vocanova-platform/apps/api/business/gamification"
	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/stretchr/testify/require"
)

// TestLearningLedgersRejectMutationExceptScopedAccountPurge runs all committed
// migrations in an isolated real PostgreSQL schema. It proves that the DB,
// rather than only the normal Go writers, protects the immutable histories.
func TestLearningLedgersRejectMutationExceptScopedAccountPurge(t *testing.T) {
	db := ledgerWriterDB(t)
	ctx := t.Context()
	now := time.Now().UTC().Truncate(time.Microsecond)
	userID := uuid.New()
	insertLedgerUser(t, ctx, db, userID, now)
	insertLearningLedgerRows(t, ctx, db, userID, now)

	for _, statement := range []string{
		`UPDATE confidence_point_ledger SET amount = 99 WHERE user_id = $1`,
		`UPDATE grace_day_ledger SET amount = 99 WHERE user_id = $1`,
		`DELETE FROM confidence_point_ledger WHERE user_id = $1`,
		`DELETE FROM grace_day_ledger WHERE user_id = $1`,
	} {
		_, err := db.ExecContext(ctx, statement, userID)
		requireAppendOnlyViolation(t, err)
	}

	// The production purge enables the transaction-local setting before its
	// ledger deletes, then commits. It may remove a deleted learner's history
	// but cannot leave the exception enabled for a later transaction.
	counters, err := accounts.NewPostgreSQLRepository(db).AnonymizeUserData(ctx, userID)
	require.NoError(t, err)
	require.EqualValues(t, 1, counters.ConfidencePointLedger)
	require.EqualValues(t, 1, counters.GraceDayLedger)

	secondUserID := uuid.New()
	insertLedgerUser(t, ctx, db, secondUserID, now)
	insertLearningLedgerRows(t, ctx, db, secondUserID, now)
	_, err = db.ExecContext(ctx, `DELETE FROM confidence_point_ledger WHERE user_id = $1`, secondUserID)
	requireAppendOnlyViolation(t, err)
}

func TestLearningLedgerPurgeGateRollsBackWithAccountAnonymization(t *testing.T) {
	db := ledgerWriterDB(t)
	ctx := t.Context()
	now := time.Now().UTC().Truncate(time.Microsecond)
	userID := uuid.New()
	insertLedgerUser(t, ctx, db, userID, now)
	insertLearningLedgerRows(t, ctx, db, userID, now)

	// Fail after the confidence ledger has been deleted but while deleting the
	// grace ledger. This exercises the production rollback path with the gate
	// enabled, rather than only asserting that set_config was called in a mock.
	_, err := db.ExecContext(ctx, `
		CREATE FUNCTION vocanova_test_reject_grace_ledger_purge()
		RETURNS trigger
		LANGUAGE plpgsql
		AS $$
		BEGIN
			RAISE EXCEPTION 'injected grace ledger purge failure';
		END;
		$$;
		CREATE TRIGGER vocanova_test_reject_grace_ledger_purge
			AFTER DELETE ON grace_day_ledger
			FOR EACH ROW EXECUTE FUNCTION vocanova_test_reject_grace_ledger_purge();
	`)
	require.NoError(t, err)

	_, err = accounts.NewPostgreSQLRepository(db).AnonymizeUserData(ctx, userID)
	require.ErrorContains(t, err, "delete grace_day_ledger")

	var pointCount, graceCount int
	require.NoError(t, db.QueryRowContext(ctx,
		`SELECT count(*) FROM confidence_point_ledger WHERE user_id = $1`, userID,
	).Scan(&pointCount))
	require.NoError(t, db.QueryRowContext(ctx,
		`SELECT count(*) FROM grace_day_ledger WHERE user_id = $1`, userID,
	).Scan(&graceCount))
	require.Equal(t, 1, pointCount, "the failed purge must restore prior ledger deletes")
	require.Equal(t, 1, graceCount)

	// set_config(..., true) is transaction-local: rollback must not leave a
	// pooled connection able to delete ledger history in a later transaction.
	_, err = db.ExecContext(ctx, `DELETE FROM confidence_point_ledger WHERE user_id = $1`, userID)
	requireAppendOnlyViolation(t, err)
}

func TestLedgerIdempotencyReplayDoesNotMutateExistingRow(t *testing.T) {
	db := ledgerWriterDB(t)
	ctx := t.Context()
	now := time.Now().UTC().Truncate(time.Microsecond)
	userID := uuid.New()
	insertLedgerUser(t, ctx, db, userID, now)
	repo := gamification.NewRepository(db)
	key := gamification.ReviewAttemptRatedKey("append-only-replay")

	tx, err := db.BeginTx(ctx, nil)
	require.NoError(t, err)
	firstID, inserted, err := repo.InsertPointLedger(ctx, tx, userID, 5, 5, "review_correct", "review_attempt", nil, key, nil, now)
	require.NoError(t, err)
	require.True(t, inserted)
	require.NoError(t, tx.Commit())

	tx, err = db.BeginTx(ctx, nil)
	require.NoError(t, err)
	secondID, inserted, err := repo.InsertPointLedger(ctx, tx, userID, 999, 999, "review_correct", "review_attempt", nil, key, nil, now.Add(time.Hour))
	require.NoError(t, err)
	require.False(t, inserted)
	require.NoError(t, tx.Commit())
	require.Equal(t, firstID, secondID)

	var amount, balanceAfter, count int
	require.NoError(t, db.QueryRowContext(ctx, `SELECT amount, balance_after FROM confidence_point_ledger WHERE id = $1`, firstID).Scan(&amount, &balanceAfter))
	require.NoError(t, db.QueryRowContext(ctx, `SELECT count(*) FROM confidence_point_ledger WHERE user_id = $1`, userID).Scan(&count))
	require.Equal(t, 5, amount)
	require.Equal(t, 5, balanceAfter)
	require.Equal(t, 1, count)

	graceKey := gamification.StreakGraceDayEarnedKey("append-only-replay", "2026-09-08")
	tx, err = db.BeginTx(ctx, nil)
	require.NoError(t, err)
	firstGraceID, err := repo.InsertGraceLedger(ctx, tx, userID, 1, 1, "earned_by_streak", "streak", nil, now, "UTC", graceKey)
	require.NoError(t, err)
	require.NoError(t, tx.Commit())

	tx, err = db.BeginTx(ctx, nil)
	require.NoError(t, err)
	secondGraceID, err := repo.InsertGraceLedger(ctx, tx, userID, 99, 99, "earned_by_streak", "streak", nil, now.AddDate(0, 0, 1), "UTC", graceKey)
	require.NoError(t, err)
	require.NoError(t, tx.Commit())
	require.Equal(t, firstGraceID, secondGraceID)

	var graceAmount, graceBalance, graceCount int
	require.NoError(t, db.QueryRowContext(ctx, `SELECT amount, balance_after FROM grace_day_ledger WHERE id = $1`, firstGraceID).Scan(&graceAmount, &graceBalance))
	require.NoError(t, db.QueryRowContext(ctx, `SELECT count(*) FROM grace_day_ledger WHERE user_id = $1`, userID).Scan(&graceCount))
	require.Equal(t, 1, graceAmount)
	require.Equal(t, 1, graceBalance)
	require.Equal(t, 1, graceCount)
}

func insertLedgerUser(t *testing.T, ctx context.Context, db *sql.DB, userID uuid.UUID, now time.Time) {
	t.Helper()
	_, err := db.ExecContext(ctx, `INSERT INTO users (id, email, status, onboarding_status, created_at, updated_at)
		VALUES ($1, $2, 'deleted', 'completed', $3, $3)`, userID, userID.String()+"@example.test", now)
	require.NoError(t, err)
}

func insertLearningLedgerRows(t *testing.T, ctx context.Context, db *sql.DB, userID uuid.UUID, now time.Time) {
	t.Helper()
	_, err := db.ExecContext(ctx, `INSERT INTO confidence_point_ledger (id, user_id, amount, balance_after, reason, source_type, occurred_at, created_at, updated_at)
		VALUES ($1, $2, 5, 5, 'review_correct', 'review_attempt', $3, $3, $3)`, uuid.New(), userID, now)
	require.NoError(t, err)
	_, err = db.ExecContext(ctx, `INSERT INTO grace_day_ledger (id, user_id, amount, balance_after, reason, source_type, applied_to_local_date, timezone, created_at, updated_at)
		VALUES ($1, $2, 1, 1, 'earned_by_streak', 'streak', $3, 'UTC', $4, $4)`, uuid.New(), userID, now, now)
	require.NoError(t, err)
}

func requireAppendOnlyViolation(t *testing.T, err error) {
	t.Helper()
	require.Error(t, err)
	var pgErr *pq.Error
	require.True(t, errors.As(err, &pgErr), "want PostgreSQL error, got %v", err)
	require.Equal(t, pq.ErrorCode("55000"), pgErr.Code)
}
