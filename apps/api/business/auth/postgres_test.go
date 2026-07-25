package auth

import (
	"context"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPostgreSQLRepositoryCreateUserAndGetByEmail(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewPostgreSQLRepository(db)
	ctx := context.Background()
	now := time.Now().UTC()

	id := uuid.New()
	email := "user@example.com"

	mock.ExpectExec("INSERT INTO users").
		WithArgs(sqlmock.AnyArg(), email, &now, sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))

	user, err := repo.CreateUser(ctx, email, &now)
	require.NoError(t, err)
	assert.Equal(t, email, user.Email)
	assert.Equal(t, "active", user.Status)

	mock.ExpectQuery("SELECT id, email, status, email_verified_at, last_login_at, created_at, updated_at FROM users").
		WithArgs(email).
		WillReturnRows(sqlmock.NewRows([]string{"id", "email", "status", "email_verified_at", "last_login_at", "created_at", "updated_at"}).
			AddRow(id, email, "active", nil, nil, now, now))

	got, err := repo.GetUserByEmail(ctx, email)
	require.NoError(t, err)
	assert.Equal(t, email, got.Email)

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestPostgreSQLRepositoryCreateSessionAndGetByTokenHash(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewPostgreSQLRepository(db)
	ctx := context.Background()
	now := time.Now().UTC()
	userID := uuid.New()
	hash := []byte("01234567890123456789012345678901")
	expiresAt := now.Add(30 * 24 * time.Hour)

	mock.ExpectExec("INSERT INTO sessions").
		WithArgs(sqlmock.AnyArg(), userID, hash, now, expiresAt).
		WillReturnResult(sqlmock.NewResult(1, 1))

	session, err := repo.CreateSession(ctx, userID, hash, now, expiresAt)
	require.NoError(t, err)
	assert.Equal(t, userID, session.UserID)
	assert.Equal(t, expiresAt, session.ExpiresAt)

	mock.ExpectQuery("SELECT id, user_id, created_at, expires_at, revoked_at FROM sessions").
		WithArgs(hash).
		WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "created_at", "expires_at", "revoked_at"}).
			AddRow(uuid.New(), userID, now, expiresAt, nil))

	got, err := repo.GetSessionByTokenHash(ctx, hash)
	require.NoError(t, err)
	assert.Equal(t, userID, got.UserID)

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestPostgreSQLRepositoryCreateMagicLinkAndGetByTokenHash(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewPostgreSQLRepository(db)
	ctx := context.Background()
	now := time.Now().UTC()
	hash := []byte("01234567890123456789012345678901")
	expiresAt := now.Add(15 * time.Minute)

	mock.ExpectExec("INSERT INTO magic_links").
		WithArgs(sqlmock.AnyArg(), "user@example.com", hash, "test", now, expiresAt).
		WillReturnResult(sqlmock.NewResult(1, 1))

	link, err := repo.CreateMagicLink(ctx, "user@example.com", hash, "test", now, expiresAt)
	require.NoError(t, err)
	assert.Equal(t, "user@example.com", link.Email)
	assert.Equal(t, "test", link.Environment)

	mock.ExpectQuery("SELECT id, user_id, email, environment, created_at, expires_at, consumed_at, revoked_at FROM magic_links").
		WithArgs(hash).
		WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "email", "environment", "created_at", "expires_at", "consumed_at", "revoked_at"}).
			AddRow(uuid.New(), nil, "user@example.com", "test", now, expiresAt, nil, nil))

	got, err := repo.GetMagicLinkByTokenHash(ctx, hash)
	require.NoError(t, err)
	assert.Equal(t, "user@example.com", got.Email)

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestPostgreSQLRepositoryCleanupExpiredSessions(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewPostgreSQLRepository(db)
	ctx := context.Background()
	before := time.Now().UTC()

	mock.ExpectExec("DELETE FROM sessions").
		WithArgs(before).
		WillReturnResult(sqlmock.NewResult(0, 3))

	n, err := repo.CleanupExpiredSessions(ctx, before)
	require.NoError(t, err)
	assert.Equal(t, int64(3), n)

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestPostgreSQLRepositoryCleanupExpiredMagicLinks(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewPostgreSQLRepository(db)
	ctx := context.Background()
	before := time.Now().UTC()

	mock.ExpectExec("DELETE FROM magic_links").
		WithArgs(before).
		WillReturnResult(sqlmock.NewResult(0, 5))

	n, err := repo.CleanupExpiredMagicLinks(ctx, before)
	require.NoError(t, err)
	assert.Equal(t, int64(5), n)

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestPostgreSQLRepositoryCreateOAuthStateAndGetByTokenHash(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewPostgreSQLRepository(db)
	ctx := context.Background()
	now := time.Now().UTC()
	hash := []byte("01234567890123456789012345678901")
	expiresAt := now.Add(10 * time.Minute)

	mock.ExpectExec("INSERT INTO oauth_states").
		WithArgs(sqlmock.AnyArg(), hash, "test", "google", "https://test.example.com/app", now, expiresAt).
		WillReturnResult(sqlmock.NewResult(1, 1))

	state, err := repo.CreateOAuthState(ctx, hash, "test", "google", "https://test.example.com/app", now, expiresAt)
	require.NoError(t, err)
	assert.Equal(t, "test", state.Environment)
	assert.Equal(t, "google", state.Provider)

	mock.ExpectQuery("SELECT id, environment, provider, app_return_url, created_at, expires_at, consumed_at FROM oauth_states").
		WithArgs(hash).
		WillReturnRows(sqlmock.NewRows([]string{"id", "environment", "provider", "app_return_url", "created_at", "expires_at", "consumed_at"}).
			AddRow(uuid.New(), "test", "google", "https://test.example.com/app", now, expiresAt, nil))

	got, err := repo.GetOAuthStateByTokenHash(ctx, hash)
	require.NoError(t, err)
	assert.Equal(t, "test", got.Environment)

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestPostgreSQLRepositoryCreateExternalIdentityAndGet(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewPostgreSQLRepository(db)
	ctx := context.Background()
	userID := uuid.New()

	mock.ExpectExec("INSERT INTO external_identities").
		WithArgs(sqlmock.AnyArg(), userID, "google", "sub-123", "user@example.com", true, sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))

	ext, err := repo.CreateExternalIdentity(ctx, userID, "google", "sub-123", "user@example.com", true)
	require.NoError(t, err)
	assert.Equal(t, userID, ext.UserID)
	assert.Equal(t, "sub-123", ext.ProviderSubject)

	mock.ExpectQuery("SELECT id, user_id, provider, provider_subject, provider_email, provider_email_verified, created_at, updated_at FROM external_identities").
		WithArgs("google", "sub-123").
		WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "provider", "provider_subject", "provider_email", "provider_email_verified", "created_at", "updated_at"}).
			AddRow(uuid.New(), userID, "google", "sub-123", "user@example.com", true, time.Now().UTC(), time.Now().UTC()))

	got, err := repo.GetExternalIdentity(ctx, "google", "sub-123")
	require.NoError(t, err)
	assert.Equal(t, "sub-123", got.ProviderSubject)

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestPostgreSQLRepositoryCleanupExpiredOAuthStates(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewPostgreSQLRepository(db)
	ctx := context.Background()
	before := time.Now().UTC()

	mock.ExpectExec("DELETE FROM oauth_states").
		WithArgs(before).
		WillReturnResult(sqlmock.NewResult(0, 2))

	n, err := repo.CleanupExpiredOAuthStates(ctx, before)
	require.NoError(t, err)
	assert.Equal(t, int64(2), n)

	require.NoError(t, mock.ExpectationsWereMet())
}
