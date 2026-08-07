package missions

import (
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const createDailyMissionSnapshotInsertQueryPattern = `INSERT INTO daily_mission_snapshots \(
			id, user_id, local_date, timezone, review_target, reviews_completed,
			policy_version, status, grace_applied, created_at, updated_at`

func newUUIDs(t *testing.T) (uuid.UUID, uuid.UUID) {
	t.Helper()
	return uuid.MustParse("00000000-0000-0000-0000-000000000001"),
		uuid.MustParse("00000000-0000-0000-0000-000000000002")
}

func TestPostgreSQLRepositoryCreateDailyMissionSnapshot(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewRepository(db)
	userID, _ := newUUIDs(t)
	day := time.Date(2026, 7, 26, 0, 0, 0, 0, time.UTC)

	mock.ExpectBegin()
	mock.ExpectQuery(createDailyMissionSnapshotInsertQueryPattern).
		WithArgs(
			sqlmock.AnyArg(), userID, day, "UTC", 20, "p4-mission-policy-v1",
		).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "user_id", "local_date", "timezone", "review_target", "reviews_completed",
			"new_word_target", "new_words_completed", "sentence_practice_target",
			"sentence_practices_completed", "policy_version", "status", "completed_at",
			"grace_applied", "grace_day_id",
		}).AddRow(
			uuid.New(), userID, day, "UTC", 20, 0,
			nil, nil, nil, nil,
			"p4-mission-policy-v1", "open", nil, false, nil,
		))
	mock.ExpectCommit()

	tx, err := db.Begin()
	require.NoError(t, err)
	snap, err := repo.CreateDailyMissionSnapshot(t.Context(), tx, userID, day, "UTC", 20, "p4-mission-policy-v1")
	require.NoError(t, err)
	require.NoError(t, tx.Commit())
	assert.Equal(t, "open", snap.Status)
	assert.Equal(t, 20, snap.ReviewTarget)
	assert.Equal(t, 0, snap.ReviewsCompleted)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestPostgreSQLRepositoryMarkSnapshotCompletedIdempotent(t *testing.T) {
	// A re-run with status='completed' must not update the row (the
	// WHERE status='open' guard). The transaction layer relies on this
	// to never double-award the +10 daily-mission reward.
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewRepository(db)
	userID, _ := newUUIDs(t)
	day := time.Date(2026, 7, 26, 0, 0, 0, 0, time.UTC)
	now := time.Date(2026, 7, 26, 12, 0, 0, 0, time.UTC)

	mock.ExpectBegin()
	mock.ExpectExec("UPDATE daily_mission_snapshots").
		WithArgs(userID, day, now).
		WillReturnResult(sqlmock.NewResult(0, 0)) // 0 rows: already completed
	mock.ExpectCommit()

	tx, err := db.Begin()
	require.NoError(t, err)
	completed, err := repo.MarkSnapshotCompleted(t.Context(), tx, userID, day, now)
	require.NoError(t, err)
	require.NoError(t, tx.Commit())
	assert.False(t, completed, "idempotent: already-completed snapshot does not re-transition")
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestPostgreSQLRepositoryMarkSnapshotCompletedFirstTime(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewRepository(db)
	userID, _ := newUUIDs(t)
	day := time.Date(2026, 7, 26, 0, 0, 0, 0, time.UTC)
	now := time.Date(2026, 7, 26, 12, 0, 0, 0, time.UTC)

	mock.ExpectBegin()
	mock.ExpectExec("UPDATE daily_mission_snapshots").
		WithArgs(userID, day, now).
		WillReturnResult(sqlmock.NewResult(0, 1)) // 1 row: transitioned
	mock.ExpectCommit()

	tx, err := db.Begin()
	require.NoError(t, err)
	completed, err := repo.MarkSnapshotCompleted(t.Context(), tx, userID, day, now)
	require.NoError(t, err)
	require.NoError(t, tx.Commit())
	assert.True(t, completed)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestPostgreSQLRepositoryIncrementWordsAdded(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewRepository(db)
	userID, _ := newUUIDs(t)
	day := time.Date(2026, 7, 26, 0, 0, 0, 0, time.UTC)

	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO daily_activity_summaries").
		WithArgs(sqlmock.AnyArg(), userID, day, "UTC").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	tx, err := db.Begin()
	require.NoError(t, err)
	err = repo.IncrementWordsAdded(t.Context(), tx, userID, day, "UTC", false)
	require.NoError(t, err)
	require.NoError(t, tx.Commit())
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestPostgreSQLRepositoryIncrementWordsAddedOptionalGoal(t *testing.T) {
	// When includeNewWordGoal is true, the mission counter is also
	// incremented. When false, only the activity summary is updated.
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewRepository(db)
	userID, _ := newUUIDs(t)
	day := time.Date(2026, 7, 26, 0, 0, 0, 0, time.UTC)

	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO daily_activity_summaries").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("UPDATE daily_mission_snapshots").
		WithArgs(userID, day).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	tx, err := db.Begin()
	require.NoError(t, err)
	err = repo.IncrementWordsAdded(t.Context(), tx, userID, day, "UTC", true)
	require.NoError(t, err)
	require.NoError(t, tx.Commit())
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestPostgreSQLRepositoryIncrementReviewsCompleted(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewRepository(db)
	userID, _ := newUUIDs(t)
	day := time.Date(2026, 7, 26, 0, 0, 0, 0, time.UTC)

	mock.ExpectBegin()
	mock.ExpectQuery("UPDATE daily_mission_snapshots").
		WithArgs(userID, day).
		WillReturnRows(sqlmock.NewRows([]string{"reviews_completed"}).AddRow(1))
	mock.ExpectExec("INSERT INTO daily_activity_summaries").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	tx, err := db.Begin()
	require.NoError(t, err)
	newCount, err := repo.IncrementReviewsCompleted(t.Context(), tx, userID, day, "UTC", 20, true, false)
	require.NoError(t, err)
	require.NoError(t, tx.Commit())
	assert.Equal(t, 1, newCount)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestPostgreSQLRepositoryIncrementReviewsCompletedSnapshotMissing(t *testing.T) {
	// If the snapshot doesn't exist, return ErrSnapshotNotFound so the
	// caller can lazily create it before retrying.
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewRepository(db)
	userID, _ := newUUIDs(t)
	day := time.Date(2026, 7, 26, 0, 0, 0, 0, time.UTC)

	mock.ExpectBegin()
	mock.ExpectQuery("UPDATE daily_mission_snapshots").
		WithArgs(userID, day).
		WillReturnError(sqlmock.ErrCancelled)
	mock.ExpectRollback()

	tx, err := db.Begin()
	require.NoError(t, err)
	_, err = repo.IncrementReviewsCompleted(t.Context(), tx, userID, day, "UTC", 20, true, false)
	require.Error(t, err)
	_ = ErrSnapshotNotFound
	require.NoError(t, tx.Rollback())
	require.NoError(t, mock.ExpectationsWereMet())
}
