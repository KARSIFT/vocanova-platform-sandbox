package gamification

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	pointLedgerInsertColumnsPattern = `INSERT INTO confidence_point_ledger \(\s*id,\s*user_id,\s*amount,\s*balance_after,\s*reason,\s*source_type,\s*source_id,\s*idempotency_key,\s*metadata,\s*occurred_at,\s*created_at,\s*updated_at\s*\)`
	graceLedgerInsertColumnsPattern = `INSERT INTO grace_day_ledger \(\s*id,\s*user_id,\s*amount,\s*balance_after,\s*reason,\s*source_type,\s*source_id,\s*applied_to_local_date,\s*timezone,\s*idempotency_key,\s*created_at,\s*updated_at\s*\)`
	streakStateInsertColumnsPattern = `INSERT INTO streak_states \(\s*id,\s*user_id,\s*current_streak_count,\s*longest_streak_count,\s*last_completed_local_date,\s*last_activity_local_date,\s*timezone,\s*status,\s*created_at,\s*updated_at\s*\)`
)

// TestRepositoryInsertPointLedgerIdempotent exercises the
// (user_id, idempotency_key) ON CONFLICT branch. A retried point award
// with the same idempotency key must not create a second row.
func TestRepositoryInsertPointLedgerIdempotent(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewRepository(db)
	userID := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	now := time.Date(2026, 7, 26, 12, 0, 0, 0, time.UTC)
	meta := json.RawMessage(`{"k":"v"}`)

	mock.ExpectBegin()
	mock.ExpectQuery(pointLedgerInsertColumnsPattern).
		WithArgs(
			sqlmock.AnyArg(), userID, 5, 5, "review_correct", "review_attempt",
			sqlmock.AnyArg(), "review_attempt:abc:rated", []byte(meta), now,
		).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(uuid.New()))
	mock.ExpectCommit()

	tx, err := db.Begin()
	require.NoError(t, err)
	id, err := repo.InsertPointLedger(
		t.Context(), tx, userID, 5, 5,
		"review_correct", "review_attempt", nil,
		ReviewAttemptRatedKey("abc"), meta, now,
	)
	require.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, id)
	require.NoError(t, tx.Commit())
	require.NoError(t, mock.ExpectationsWereMet())
}

// TestRepositoryInsertGraceLedgerIdempotent is the grace-day equivalent of
// the point-ledger idempotency test.
func TestRepositoryInsertGraceLedgerIdempotent(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewRepository(db)
	userID := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	day := time.Date(2026, 7, 26, 0, 0, 0, 0, time.UTC)

	mock.ExpectBegin()
	mock.ExpectQuery(graceLedgerInsertColumnsPattern).
		WithArgs(
			sqlmock.AnyArg(), userID, 1, 1, "earned_by_streak", "streak",
			sqlmock.AnyArg(), day, "UTC", "streak:u1:2026-07-26:grace_day_earned",
		).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(uuid.New()))
	mock.ExpectCommit()

	tx, err := db.Begin()
	require.NoError(t, err)
	id, err := repo.InsertGraceLedger(
		t.Context(), tx, userID, 1, 1,
		"earned_by_streak", "streak", nil,
		day, "UTC",
		StreakGraceDayEarnedKey("u1", "2026-07-26"),
	)
	require.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, id)
	require.NoError(t, tx.Commit())
	require.NoError(t, mock.ExpectationsWereMet())
}

// userSettingsInsertColumnsPattern matches the user_settings INSERT column
// list only: the regression this guards is the omission of
// created_at/updated_at, not the placeholder numbering or timestamp source the
// fix chooses for them. sqlmock collapses each whitespace run in the actual
// statement to a single space, so the pattern tolerates optional spaces around
// the parentheses rather than assuming how the SQL literal is wrapped.
const userSettingsInsertColumnsPattern = `INSERT INTO user_settings \(\s*id,\s*user_id,\s*timezone,\s*daily_review_target,\s*created_at,\s*updated_at\s*\)`

func TestRepositoryUpsertUserSettingsFreshInsertSuppliesTimestamps(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewRepository(db)
	userID := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	timezone := "Europe/Madrid"
	dailyReviewTarget := 30

	mock.ExpectBegin()
	mock.ExpectQuery(userSettingsInsertColumnsPattern).
		WillReturnRows(sqlmock.NewRows([]string{
			"user_id",
			"timezone",
			"daily_review_target",
			"review_interval_preset",
			"notifications_enabled",
			"marketing_emails_enabled",
			"app_language",
		}).AddRow(
			userID,
			timezone,
			dailyReviewTarget,
			"vocanova_default",
			true,
			false,
			"en",
		))
	mock.ExpectCommit()

	tx, err := db.Begin()
	require.NoError(t, err)
	got, err := repo.UpsertUserSettings(t.Context(), tx, userID, timezone, dailyReviewTarget)
	require.NoError(t, err)
	require.NoError(t, tx.Commit())

	require.NotNil(t, got)
	assert.Equal(t, userID, got.UserID)
	assert.Equal(t, timezone, got.Timezone)
	assert.Equal(t, dailyReviewTarget, got.DailyReviewTarget)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestRepositoryUpsertStreakStateFreshInsertSuppliesTimestamps(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewRepository(db)
	userID := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	now := time.Date(2026, 8, 7, 12, 0, 0, 0, time.UTC)

	mock.ExpectBegin()
	mock.ExpectExec(streakStateInsertColumnsPattern).
		WithArgs(sqlmock.AnyArg(), userID, 3, 5, now, now, "UTC", "active").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	tx, err := db.Begin()
	require.NoError(t, err)
	err = repo.UpsertStreakState(t.Context(), tx, userID, 3, 5, &now, &now, "UTC", "active")
	require.NoError(t, err)
	require.NoError(t, tx.Commit())
	require.NoError(t, mock.ExpectationsWereMet())
}
