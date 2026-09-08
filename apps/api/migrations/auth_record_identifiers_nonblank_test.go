package migrations_test

import (
	"bytes"
	"context"
	"database/sql"
	"errors"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const authRecordIdentifiersNonblankMigration = "20260908290000_voc1444_auth_record_identifiers_nonblank.sql"

func TestAuthRecordIdentifiersNonblankMigrationCarriesDatabaseAndEntInvariants(t *testing.T) {
	body, err := os.ReadFile(authRecordIdentifiersNonblankMigration)
	require.NoError(t, err)
	text := string(body)
	for _, invariant := range []string{
		"external_identities_provider_subject_nonblank",
		"magic_links_email_nonblank",
		"magic_links_environment_nonblank",
		"email_change_links_new_email_nonblank",
		"email_change_links_environment_nonblank",
		`U&'[^\0009-\000D\0020\0085\00A0\1680\2000-\200A\2028\2029\202F\205F\3000]'`,
		"NOT VALID",
	} {
		assert.Contains(t, text, invariant)
	}

	for filename, checks := range map[string][]string{
		"../ent/schema/externalidentity.go": {
			`"provider_subject_nonblank": authRecordIdentifierNonblankCheck("provider_subject")`,
			`field.String("provider_subject").NotEmpty().Validate(validateAuthRecordIdentifierNonblank)`,
		},
		"../ent/schema/magiclink.go": {
			`"email_nonblank":       authRecordIdentifierNonblankCheck("email")`,
			`"environment_nonblank": authRecordIdentifierNonblankCheck("environment")`,
			`field.String("email").NotEmpty().Validate(validateAuthRecordIdentifierNonblank).Immutable()`,
		},
		"../ent/schema/emailchangelink.go": {
			`"new_email_nonblank":   authRecordIdentifierNonblankCheck("new_email")`,
			`"environment_nonblank": authRecordIdentifierNonblankCheck("environment")`,
			`field.String("new_email").NotEmpty().Validate(validateAuthRecordIdentifierNonblank).Immutable()`,
		},
	} {
		schemaBody, readErr := os.ReadFile(filename)
		require.NoError(t, readErr)
		for _, check := range checks {
			assert.Contains(t, string(schemaBody), check)
		}
	}
}

func TestAuthRecordIdentifiersNonblankAgainstPostgreSQL(t *testing.T) {
	dsn := os.Getenv("VOCANOVA_TEST_POSTGRES_DSN")
	if dsn == "" {
		t.Skip("VOCANOVA_TEST_POSTGRES_DSN is unset; real PostgreSQL migration test unavailable")
	}

	db, err := sql.Open("postgres", dsn)
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, db.Close()) })
	db.SetMaxOpenConns(1)

	ctx, cancel := context.WithTimeout(t.Context(), 30*time.Second)
	defer cancel()
	schema := "vocanova_auth_nonblank_" + strings.ReplaceAll(uuid.NewString(), "-", "")
	_, err = db.ExecContext(ctx, "CREATE SCHEMA "+pq.QuoteIdentifier(schema))
	require.NoError(t, err)
	t.Cleanup(func() {
		_, dropErr := db.ExecContext(context.Background(), "DROP SCHEMA IF EXISTS "+pq.QuoteIdentifier(schema)+" CASCADE")
		assert.NoError(t, dropErr)
	})
	_, err = db.ExecContext(ctx, "SET search_path TO "+pq.QuoteIdentifier(schema))
	require.NoError(t, err)

	for _, filename := range []string{
		"20260724210000_identity_foundation.sql",
		"20260725140001_voc031_p5_email_change_links.sql",
	} {
		migration, readErr := os.ReadFile(filename)
		require.NoError(t, readErr)
		_, err = db.ExecContext(ctx, string(migration))
		require.NoError(t, err)
	}

	now := time.Now().UTC().Truncate(time.Microsecond)
	expiresAt := now.Add(10 * time.Minute)
	userID := uuid.New()
	_, err = db.ExecContext(ctx, `INSERT INTO users
		(id, email, status, onboarding_status, created_at, updated_at)
		VALUES ($1, 'auth@example.test', 'active', 'completed', $2, $2)`, userID, now)
	require.NoError(t, err)

	legacyIdentityID := uuid.New()
	legacyMagicID := uuid.New()
	legacyEmailChangeID := uuid.New()
	_, err = db.ExecContext(ctx, `INSERT INTO external_identities
		(id, user_id, provider, provider_subject, created_at, updated_at)
		VALUES ($1, $2, 'google', $3, $4, $4)`, legacyIdentityID, userID, "\u00a0\u2007", now)
	require.NoError(t, err)
	_, err = db.ExecContext(ctx, `INSERT INTO magic_links
		(id, user_id, email, token_hash, environment, created_at, expires_at)
		VALUES ($1, $2, E' \t\n', $3, $4, $5, $6)`, legacyMagicID, userID, authHash(1), "\u202f", now, expiresAt)
	require.NoError(t, err)
	_, err = db.ExecContext(ctx, `INSERT INTO email_change_links
		(id, user_id, new_email, token_hash, environment, created_at, expires_at)
		VALUES ($1, $2, $3, $4, E' \t\n', $5, $6)`, legacyEmailChangeID, userID, "\u3000", authHash(2), now, expiresAt)
	require.NoError(t, err)

	migration, err := os.ReadFile(authRecordIdentifiersNonblankMigration)
	require.NoError(t, err)
	_, err = db.ExecContext(ctx, string(migration))
	require.NoError(t, err, "NOT VALID constraints must preserve legacy auth records")

	validIdentityID := uuid.New()
	validMagicID := uuid.New()
	validEmailChangeID := uuid.New()
	_, err = db.ExecContext(ctx, `INSERT INTO external_identities
		(id, user_id, provider, provider_subject, created_at, updated_at)
		VALUES ($1, $2, 'google', 'provider-subject', $3, $3)`, validIdentityID, userID, now)
	require.NoError(t, err)
	_, err = db.ExecContext(ctx, `INSERT INTO magic_links
		(id, user_id, email, token_hash, environment, created_at, expires_at)
		VALUES ($1, $2, 'learner@example.test', $3, 'test', $4, $5)`, validMagicID, userID, authHash(3), now, expiresAt)
	require.NoError(t, err)
	_, err = db.ExecContext(ctx, `INSERT INTO email_change_links
		(id, user_id, new_email, token_hash, environment, created_at, expires_at)
		VALUES ($1, $2, 'new@example.test', $3, 'test', $4, $5)`, validEmailChangeID, userID, authHash(4), now, expiresAt)
	require.NoError(t, err)

	invalidInserts := []struct {
		name  string
		query string
		args  []any
	}{
		{name: "provider subject", query: `INSERT INTO external_identities (id, user_id, provider, provider_subject, created_at, updated_at) VALUES ($1, $2, 'google', $3, $4, $4)`, args: []any{uuid.New(), userID, "\u00a0\u202f", now}},
		{name: "magic email", query: `INSERT INTO magic_links (id, user_id, email, token_hash, environment, created_at, expires_at) VALUES ($1, $2, E' \t\n', $3, 'test', $4, $5)`, args: []any{uuid.New(), userID, authHash(10), now, expiresAt}},
		{name: "magic environment", query: `INSERT INTO magic_links (id, user_id, email, token_hash, environment, created_at, expires_at) VALUES ($1, $2, 'valid@example.test', $3, $4, $5, $6)`, args: []any{uuid.New(), userID, authHash(11), "\u3000", now, expiresAt}},
		{name: "email change address", query: `INSERT INTO email_change_links (id, user_id, new_email, token_hash, environment, created_at, expires_at) VALUES ($1, $2, $3, $4, 'test', $5, $6)`, args: []any{uuid.New(), userID, "\u2007", authHash(12), now, expiresAt}},
		{name: "email change environment", query: `INSERT INTO email_change_links (id, user_id, new_email, token_hash, environment, created_at, expires_at) VALUES ($1, $2, 'valid@example.test', $3, E' \t\n', $4, $5)`, args: []any{uuid.New(), userID, authHash(13), now, expiresAt}},
	}
	for _, tc := range invalidInserts {
		t.Run("reject insert "+tc.name, func(t *testing.T) {
			_, insertErr := db.ExecContext(ctx, tc.query, tc.args...)
			requireAuthIdentifierCheckViolation(t, insertErr)
		})
	}

	invalidUpdates := []struct {
		name  string
		query string
		id    uuid.UUID
	}{
		{name: "provider subject", query: `UPDATE external_identities SET provider_subject = $2 WHERE id = $1`, id: validIdentityID},
		{name: "magic email", query: `UPDATE magic_links SET email = $2 WHERE id = $1`, id: validMagicID},
		{name: "magic environment", query: `UPDATE magic_links SET environment = $2 WHERE id = $1`, id: validMagicID},
		{name: "email change address", query: `UPDATE email_change_links SET new_email = $2 WHERE id = $1`, id: validEmailChangeID},
		{name: "email change environment", query: `UPDATE email_change_links SET environment = $2 WHERE id = $1`, id: validEmailChangeID},
	}
	for _, tc := range invalidUpdates {
		t.Run("reject update "+tc.name, func(t *testing.T) {
			_, updateErr := db.ExecContext(ctx, tc.query, tc.id, "\u00a0\u2007\u202f\u3000")
			requireAuthIdentifierCheckViolation(t, updateErr)
		})
	}

	_, err = db.ExecContext(ctx, `UPDATE external_identities SET provider_email_verified = true WHERE id = $1`, legacyIdentityID)
	requireAuthIdentifierCheckViolation(t, err)
	_, err = db.ExecContext(ctx, `UPDATE magic_links SET revoked_at = $2 WHERE id = $1`, legacyMagicID, now)
	requireAuthIdentifierCheckViolation(t, err)
	_, err = db.ExecContext(ctx, `UPDATE email_change_links SET revoked_at = $2 WHERE id = $1`, legacyEmailChangeID, now)
	requireAuthIdentifierCheckViolation(t, err)

	for _, name := range []string{
		"external_identities_provider_subject_nonblank",
		"magic_links_email_nonblank",
		"magic_links_environment_nonblank",
		"email_change_links_new_email_nonblank",
		"email_change_links_environment_nonblank",
	} {
		var validated bool
		var definition string
		require.NoError(t, db.QueryRowContext(ctx, `
			SELECT convalidated, pg_get_constraintdef(oid)
			FROM pg_constraint
			WHERE connamespace = current_schema()::regnamespace AND conname = $1`, name).Scan(&validated, &definition))
		assert.False(t, validated, name)
		assert.Contains(t, definition, "~ '[^", name)
		assert.Contains(t, definition, "\u00a0", name)
		assert.Contains(t, definition, "\u3000", name)
	}
}

func authHash(seed byte) []byte {
	return bytes.Repeat([]byte{seed}, 32)
}

func requireAuthIdentifierCheckViolation(t *testing.T, err error) {
	t.Helper()
	require.Error(t, err)
	var pqErr *pq.Error
	require.True(t, errors.As(err, &pqErr), "expected PostgreSQL error, got %T: %v", err, err)
	assert.Equal(t, pq.ErrorCode("23514"), pqErr.Code)
}
