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

func TestPostgreSQLRepositoryCreateEmailChangeLink(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewPostgreSQLRepository(db)
	ctx := context.Background()
	now := time.Now().UTC()
	uid := uuid.New()
	hash := []byte("01234567890123456789012345678901")
	expires := now.Add(15 * time.Minute)

	mock.ExpectExec("INSERT INTO email_change_links").
		WithArgs(sqlmock.AnyArg(), uid, "new@example.com", hash, "test", now, expires).
		WillReturnResult(sqlmock.NewResult(1, 1))

	link, err := repo.CreateEmailChangeLink(ctx, uid, "new@example.com", hash, "test", now, expires)
	require.NoError(t, err)
	assert.Equal(t, uid, link.UserID)
	assert.Equal(t, "new@example.com", link.NewEmail)
	assert.Equal(t, "test", link.Environment)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestPostgreSQLRepositoryCreateEmailChangeLinkRequiresUserID(t *testing.T) {
	db, _, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()
	repo := NewPostgreSQLRepository(db)
	_, err = repo.CreateEmailChangeLink(context.Background(), uuid.Nil, "new@example.com",
		[]byte("01234567890123456789012345678901"), "test", time.Now(), time.Now().Add(time.Hour))
	require.Error(t, err)
}

func TestPostgreSQLRepositoryGetEmailChangeLinkByTokenHash(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewPostgreSQLRepository(db)
	ctx := context.Background()
	now := time.Now().UTC()
	uid := uuid.New()
	id := uuid.New()
	hash := []byte("01234567890123456789012345678901")

	mock.ExpectQuery("SELECT id, user_id, new_email, environment, created_at, expires_at, consumed_at, revoked_at FROM email_change_links WHERE token_hash").
		WithArgs(hash).
		WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "new_email", "environment", "created_at", "expires_at", "consumed_at", "revoked_at"}).
			AddRow(id, uid, "new@example.com", "test", now, now.Add(15*time.Minute), nil, nil))

	got, err := repo.GetEmailChangeLinkByTokenHash(ctx, hash)
	require.NoError(t, err)
	assert.Equal(t, id, got.ID)
	assert.Equal(t, uid, got.UserID)
	assert.Equal(t, "new@example.com", got.NewEmail)
	assert.Nil(t, got.ConsumedAt)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestPostgreSQLRepositoryGetEmailChangeLinkNotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewPostgreSQLRepository(db)
	hash := []byte("01234567890123456789012345678901")

	mock.ExpectQuery("SELECT id, user_id, new_email").
		WithArgs(hash).
		WillReturnError(errors.New("sql: no rows in result set"))

	_, err = repo.GetEmailChangeLinkByTokenHash(context.Background(), hash)
	require.Error(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestPostgreSQLRepositoryConsumeEmailChangeLink(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewPostgreSQLRepository(db)
	id := uuid.New()
	now := time.Now().UTC()

	mock.ExpectExec("UPDATE email_change_links SET consumed_at").
		WithArgs(now, id).
		WillReturnResult(sqlmock.NewResult(0, 1))

	require.NoError(t, repo.ConsumeEmailChangeLink(context.Background(), id, now))
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestPostgreSQLRepositoryRevokeAllEmailChangeLinksForUser(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewPostgreSQLRepository(db)
	uid := uuid.New()
	now := time.Now().UTC()

	mock.ExpectExec("UPDATE email_change_links SET revoked_at").
		WithArgs(now, uid).
		WillReturnResult(sqlmock.NewResult(0, 2))

	n, err := repo.RevokeAllEmailChangeLinksForUser(context.Background(), uid, now)
	require.NoError(t, err)
	assert.Equal(t, int64(2), n)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestPostgreSQLRepositoryUpdateUserEmailSuccess(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewPostgreSQLRepository(db)
	uid := uuid.New()
	now := time.Now().UTC()

	mock.ExpectExec("UPDATE users SET email").
		WithArgs(uid, "new@example.com", now).
		WillReturnResult(sqlmock.NewResult(0, 1))

	require.NoError(t, repo.UpdateUserEmail(context.Background(), uid, "new@example.com", now))
	require.NoError(t, mock.ExpectationsWereMet())
}

// TestPostgreSQLRepositoryUpdateUserEmailUniqueViolationIsStableConflict
// covers VOC-031-R02: a partial-unique-index violation on users.email
// must be translated to ErrEmailAlreadyRegistered, never a 500.
func TestPostgreSQLRepositoryUpdateUserEmailUniqueViolationIsStableConflict(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewPostgreSQLRepository(db)
	uid := uuid.New()
	now := time.Now().UTC()

	pqErr := &pq.Error{
		Code:    "23505",
		Message: `duplicate key value violates unique constraint "users_active_email_key"`,
	}
	mock.ExpectExec("UPDATE users SET email").
		WithArgs(uid, "taken@example.com", now).
		WillReturnError(pqErr)

	err = repo.UpdateUserEmail(context.Background(), uid, "taken@example.com", now)
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrEmailAlreadyRegistered, "unique_violation must surface as ErrEmailAlreadyRegistered")
	require.NoError(t, mock.ExpectationsWereMet())
}

// TestPostgreSQLRepositoryUpdateUserEmailOtherErrorPropagates covers
// the non-unique-violation case: a generic SQL error must not be
// misclassified as a duplicate-email conflict.
func TestPostgreSQLRepositoryUpdateUserEmailOtherErrorPropagates(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewPostgreSQLRepository(db)
	uid := uuid.New()
	now := time.Now().UTC()

	mock.ExpectExec("UPDATE users SET email").
		WithArgs(uid, "new@example.com", now).
		WillReturnError(errors.New("connection reset by peer"))

	err = repo.UpdateUserEmail(context.Background(), uid, "new@example.com", now)
	require.Error(t, err)
	assert.False(t, errors.Is(err, ErrEmailAlreadyRegistered), "non-unique-violation must not surface as ErrEmailAlreadyRegistered")
	require.NoError(t, mock.ExpectationsWereMet())
}
