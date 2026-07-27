package accounts

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestPostgreSQLRepositoryCreateAccountDeletionRequestDeactivatesUser
// covers the SQL transaction the deactivation path runs:
// UPDATE users + UPDATE sessions + UPDATE magic_links +
// UPDATE email_change_links + INSERT INTO
// account_deletion_requests, all inside one transaction.
func TestPostgreSQLRepositoryCreateAccountDeletionRequestDeactivatesUser(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewPostgreSQLRepository(db)
	ctx := context.Background()
	now := time.Now().UTC().Truncate(time.Microsecond)
	uid := uuid.New()
	idemKey := "test-idem-key"
	purgeDelay := 30 * 24 * time.Hour
	purgeAfter := now.Add(purgeDelay)

	mock.ExpectBegin()
	mock.ExpectExec("UPDATE users SET status = 'deleted'").
		WithArgs(uid, now).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("UPDATE sessions SET revoked_at").
		WithArgs(uid, now).
		WillReturnResult(sqlmock.NewResult(0, 2))
	mock.ExpectExec("UPDATE magic_links SET revoked_at").
		WithArgs(uid, now).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("UPDATE email_change_links SET revoked_at").
		WithArgs(uid, now).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("INSERT INTO account_deletion_requests").
		WithArgs(sqlmock.AnyArg(), uid, now, purgeAfter, idemKey).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	row, err := repo.CreateAccountDeletionRequest(ctx, uid, idemKey, now, purgeDelay)
	require.NoError(t, err)
	require.NotNil(t, row)
	assert.Equal(t, uid, row.UserID)
	assert.Equal(t, "deactivated", row.Status)
	assert.Equal(t, idemKey, row.IdempotencyKey)
	assert.Equal(t, purgeAfter, row.PurgeAfter)
	require.NoError(t, mock.ExpectationsWereMet())
}

// TestPostgreSQLRepositoryCreateAccountDeletionRequestMissingUser
// covers the 404-path: the UPDATE users affects 0 rows, the
// transaction is rolled back, and the call returns
// ErrUserNotFound.
func TestPostgreSQLRepositoryCreateAccountDeletionRequestMissingUser(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewPostgreSQLRepository(db)
	uid := uuid.New()
	now := time.Now().UTC()

	mock.ExpectBegin()
	mock.ExpectExec("UPDATE users SET status = 'deleted'").
		WithArgs(uid, now).
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectRollback()

	_, err = repo.CreateAccountDeletionRequest(context.Background(), uid, "idem", now, 30*24*time.Hour)
	assert.ErrorIs(t, err, ErrUserNotFound)
	require.NoError(t, mock.ExpectationsWereMet())
}

// TestPostgreSQLRepositoryCreateAccountDeletionRequestAlreadyInFlight
// covers the (user_id) UNIQUE-violation translation: a second
// deactivation for the same user is translated to
// ErrAccountDeletionAlreadyInFlight (a stable 409), never a
// 500.
func TestPostgreSQLRepositoryCreateAccountDeletionRequestAlreadyInFlight(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewPostgreSQLRepository(db)
	uid := uuid.New()
	now := time.Now().UTC()

	mock.ExpectBegin()
	mock.ExpectExec("UPDATE users SET status = 'deleted'").
		WithArgs(uid, now).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("UPDATE sessions SET revoked_at").
		WithArgs(uid, now).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("UPDATE magic_links SET revoked_at").
		WithArgs(uid, now).
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec("UPDATE email_change_links SET revoked_at").
		WithArgs(uid, now).
		WillReturnResult(sqlmock.NewResult(0, 0))
	pqErr := &pq.Error{Code: "23505", Message: `duplicate key value violates unique constraint "account_deletion_requests_user_id_key"`}
	mock.ExpectExec("INSERT INTO account_deletion_requests").
		WithArgs(sqlmock.AnyArg(), uid, sqlmock.AnyArg(), sqlmock.AnyArg(), "idem-2").
		WillReturnError(pqErr)
	mock.ExpectRollback()

	_, err = repo.CreateAccountDeletionRequest(context.Background(), uid, "idem-2", now, 30*24*time.Hour)
	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrAccountDeletionAlreadyInFlight),
		"unique-violation on (user_id) must surface as ErrAccountDeletionAlreadyInFlight, not a 500")
	require.NoError(t, mock.ExpectationsWereMet())
}

// TestPostgreSQLRepositoryGetAccountDeletionRequestByUserID covers
// the read used by the replay path.
func TestPostgreSQLRepositoryGetAccountDeletionRequestByUserID(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewPostgreSQLRepository(db)
	uid := uuid.New()
	id := uuid.New()
	now := time.Now().UTC()

	row := sqlmock.NewRows([]string{
		"id", "user_id", "status", "requested_at", "purge_after",
		"completed_at", "idempotency_key", "created_at", "updated_at",
	}).AddRow(id, uid, "deactivated", now, now.Add(30*24*time.Hour), nil, "idem", now, now)

	mock.ExpectQuery("SELECT id, user_id, status, requested_at, purge_after").
		WithArgs(uid).
		WillReturnRows(row)

	got, err := repo.GetAccountDeletionRequestByUserID(context.Background(), uid)
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, id, got.ID)
	assert.Equal(t, "deactivated", got.Status)
	assert.Equal(t, "idem", got.IdempotencyKey)
	require.NoError(t, mock.ExpectationsWereMet())
}

// TestPostgreSQLRepositoryGetAccountDeletionRequestByUserIDNotFound
// covers the missing-row path.
func TestPostgreSQLRepositoryGetAccountDeletionRequestByUserIDNotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewPostgreSQLRepository(db)
	uid := uuid.New()

	mock.ExpectQuery("SELECT id, user_id, status, requested_at, purge_after").
		WithArgs(uid).
		WillReturnError(errors.New("sql: no rows in result set"))

	_, err = repo.GetAccountDeletionRequestByUserID(context.Background(), uid)
	require.Error(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

// TestPostgreSQLRepositoryListDeactivatedRequestsDueForPurge covers
// the sweep's "find candidates" read.
func TestPostgreSQLRepositoryListDeactivatedRequestsDueForPurge(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewPostgreSQLRepository(db)
	now := time.Now().UTC()
	uidA := uuid.New()
	uidB := uuid.New()

	rows := sqlmock.NewRows([]string{
		"id", "user_id", "status", "requested_at", "purge_after",
		"completed_at", "idempotency_key", "created_at", "updated_at",
	}).
		AddRow(uuid.New(), uidA, "deactivated", now.Add(-30*24*time.Hour), now.Add(-time.Hour), nil, "idem-A", now.Add(-30*24*time.Hour), now.Add(-30*24*time.Hour)).
		AddRow(uuid.New(), uidB, "deactivated", now.Add(-30*24*time.Hour), now.Add(-time.Minute), nil, "idem-B", now.Add(-30*24*time.Hour), now.Add(-30*24*time.Hour))

	mock.ExpectQuery("FROM account_deletion_requests").
		WithArgs(now, 100).
		WillReturnRows(rows)

	got, err := repo.ListDeactivatedRequestsDueForPurge(context.Background(), now, 100)
	require.NoError(t, err)
	require.Len(t, got, 2)
	assert.Equal(t, uidA, got[0].UserID)
	assert.Equal(t, uidB, got[1].UserID)
	require.NoError(t, mock.ExpectationsWereMet())
}

// TestPostgreSQLRepositoryClaimAccountDeletionRequestForAnonymization
// covers the atomic claim transition. A winning claim returns
// (true, nil); a losing claim (the row is no longer
// 'deactivated') returns (false, nil).
func TestPostgreSQLRepositoryClaimAccountDeletionRequestForAnonymization(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewPostgreSQLRepository(db)
	id := uuid.New()
	now := time.Now().UTC()

	// Winning claim.
	mock.ExpectExec("UPDATE account_deletion_requests SET status = 'anonymizing'").
		WithArgs(id, now).
		WillReturnResult(sqlmock.NewResult(0, 1))
	claimed, err := repo.ClaimAccountDeletionRequestForAnonymization(context.Background(), id, now)
	require.NoError(t, err)
	assert.True(t, claimed, "winning claim returns true")

	// Losing claim: 0 rows updated (already anonymizing/completed).
	mock.ExpectExec("UPDATE account_deletion_requests SET status = 'anonymizing'").
		WithArgs(id, now).
		WillReturnResult(sqlmock.NewResult(0, 0))
	claimed, err = repo.ClaimAccountDeletionRequestForAnonymization(context.Background(), id, now)
	require.NoError(t, err)
	assert.False(t, claimed, "losing claim returns false")

	require.NoError(t, mock.ExpectationsWereMet())
}

// TestPostgreSQLRepositoryMarkAccountDeletionRequestCompleted
// covers the terminal transition.
func TestPostgreSQLRepositoryMarkAccountDeletionRequestCompleted(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewPostgreSQLRepository(db)
	id := uuid.New()
	now := time.Now().UTC()

	mock.ExpectExec("UPDATE account_deletion_requests SET status = 'completed'").
		WithArgs(id, now).
		WillReturnResult(sqlmock.NewResult(0, 1))

	require.NoError(t, repo.MarkAccountDeletionRequestCompleted(context.Background(), id, now))
	require.NoError(t, mock.ExpectationsWereMet())
}

// TestPostgreSQLRepositoryCreateAccountDeletionRequestRequiresUserID
// covers the input-validation branch.
func TestPostgreSQLRepositoryCreateAccountDeletionRequestRequiresUserID(t *testing.T) {
	repo := NewPostgreSQLRepository(nil)
	_, err := repo.CreateAccountDeletionRequest(context.Background(), uuid.Nil, "idem", time.Now(), 30*24*time.Hour)
	require.Error(t, err)
	_, err = repo.CreateAccountDeletionRequest(context.Background(), uuid.New(), "", time.Now(), 30*24*time.Hour)
	require.Error(t, err)
}
