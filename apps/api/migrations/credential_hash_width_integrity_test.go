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

const credentialHashCheck = "CHECK (octet_length(token_hash) = 32)"

var credentialHashMigrationSources = map[string]string{
	"sessions":           "20260724210000_identity_foundation.sql",
	"magic_links":        "20260724210000_identity_foundation.sql",
	"oauth_states":       "20260724210001_oauth_state.sql",
	"email_change_links": "20260725140001_voc031_p5_email_change_links.sql",
}

func TestCredentialHashWidthDatabaseAndEntInvariants(t *testing.T) {
	for table, migration := range credentialHashMigrationSources {
		body, err := os.ReadFile(migration)
		require.NoError(t, err)
		assert.Contains(t, string(body), "CREATE TABLE "+table)
		assert.Contains(t, string(body), credentialHashCheck)
	}

	for _, schemaFile := range []string{"session.go", "magiclink.go", "emailchangelink.go"} {
		body, err := os.ReadFile("../ent/schema/" + schemaFile)
		require.NoError(t, err)
		assert.Contains(t, string(body), `"token_hash_sha256_width": "octet_length(token_hash) = 32"`)
	}
}

func TestCredentialHashWidthIntegrityAgainstPostgreSQL(t *testing.T) {
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
	schema := "vocanova_credential_hash_" + strings.ReplaceAll(uuid.NewString(), "-", "")
	_, err = db.ExecContext(ctx, "CREATE SCHEMA "+pq.QuoteIdentifier(schema))
	require.NoError(t, err)
	t.Cleanup(func() {
		_, dropErr := db.ExecContext(context.Background(), "DROP SCHEMA IF EXISTS "+pq.QuoteIdentifier(schema)+" CASCADE")
		assert.NoError(t, dropErr)
	})
	_, err = db.ExecContext(ctx, "SET search_path TO "+pq.QuoteIdentifier(schema))
	require.NoError(t, err)

	for _, migration := range []string{
		"20260724210000_identity_foundation.sql",
		"20260724210001_oauth_state.sql",
		"20260725140001_voc031_p5_email_change_links.sql",
	} {
		body, readErr := os.ReadFile(migration)
		require.NoError(t, readErr)
		_, execErr := db.ExecContext(ctx, string(body))
		require.NoError(t, execErr, "apply historical migration %s", migration)
	}

	now := time.Now().UTC().Truncate(time.Microsecond)
	userID := uuid.New()
	_, err = db.ExecContext(ctx,
		"INSERT INTO users (id, created_at, updated_at) VALUES ($1, $2, $2)", userID, now)
	require.NoError(t, err)

	validHash := bytes.Repeat([]byte{0x5a}, 32)
	invalidHashes := []struct {
		name string
		hash []byte
	}{
		{name: "zero", hash: []byte{}},
		{name: "one", hash: bytes.Repeat([]byte{0x01}, 1)},
		{name: "short", hash: bytes.Repeat([]byte{0x1f}, 31)},
		{name: "long", hash: bytes.Repeat([]byte{0x21}, 33)},
	}

	for _, table := range credentialHashTables(userID, now) {
		t.Run("table "+table.name, func(t *testing.T) {
			validID := uuid.New()
			_, insertErr := db.ExecContext(ctx, table.insertSQL, table.args(validID, validHash)...)
			require.NoError(t, insertErr)

			for _, tc := range invalidHashes {
				t.Run("reject "+tc.name, func(t *testing.T) {
					_, invalidErr := db.ExecContext(ctx, table.insertSQL, table.args(uuid.New(), tc.hash)...)
					requireCredentialHashWidthViolation(t, invalidErr)
				})
			}

			_, updateErr := db.ExecContext(ctx,
				"UPDATE "+pq.QuoteIdentifier(table.name)+" SET token_hash = $1 WHERE id = $2",
				bytes.Repeat([]byte{0x02}, 31), validID)
			requireCredentialHashWidthViolation(t, updateErr)
			_, updateErr = db.ExecContext(ctx,
				"UPDATE "+pq.QuoteIdentifier(table.name)+" SET expires_at = expires_at WHERE id = $1", validID)
			require.NoError(t, updateErr, "the hash check must not alter credential lifecycle updates")

			var constraints int
			require.NoError(t, db.QueryRowContext(ctx, `
				SELECT count(*)
				FROM pg_constraint
				WHERE conrelid = $1::regclass
				  AND contype = 'c'
				  AND pg_get_constraintdef(oid) LIKE '%octet_length(token_hash)%32%'
				  AND convalidated`, table.name).Scan(&constraints))
			assert.Equal(t, 1, constraints, "the historical migration carries one validated width check")
		})
	}
}

type credentialHashTable struct {
	name      string
	insertSQL string
	args      func(uuid.UUID, []byte) []any
}

func credentialHashTables(userID uuid.UUID, now time.Time) []credentialHashTable {
	expiresAt := now.Add(time.Minute)
	return []credentialHashTable{
		{
			name:      "sessions",
			insertSQL: "INSERT INTO sessions (id, user_id, token_hash, created_at, expires_at) VALUES ($1, $2, $3, $4, $5)",
			args: func(id uuid.UUID, hash []byte) []any {
				return []any{id, userID, hash, now, expiresAt}
			},
		},
		{
			name:      "magic_links",
			insertSQL: "INSERT INTO magic_links (id, email, token_hash, environment, created_at, expires_at) VALUES ($1, $2, $3, $4, $5, $6)",
			args: func(id uuid.UUID, hash []byte) []any {
				return []any{id, "learner@example.test", hash, "test", now, expiresAt}
			},
		},
		{
			name:      "oauth_states",
			insertSQL: "INSERT INTO oauth_states (id, token_hash, environment, provider, app_return_url, created_at, expires_at) VALUES ($1, $2, $3, $4, $5, $6, $7)",
			args: func(id uuid.UUID, hash []byte) []any {
				return []any{id, hash, "test", "google", "https://app.example.test/return", now, expiresAt}
			},
		},
		{
			name:      "email_change_links",
			insertSQL: "INSERT INTO email_change_links (id, user_id, new_email, token_hash, environment, created_at, expires_at) VALUES ($1, $2, $3, $4, $5, $6, $7)",
			args: func(id uuid.UUID, hash []byte) []any {
				return []any{id, userID, "new-learner@example.test", hash, "test", now, expiresAt}
			},
		},
	}
}

func requireCredentialHashWidthViolation(t *testing.T, err error) {
	t.Helper()
	require.Error(t, err)
	var pqErr *pq.Error
	require.True(t, errors.As(err, &pqErr), "expected PostgreSQL error, got %T: %v", err, err)
	assert.Equal(t, pq.ErrorCode("23514"), pqErr.Code)
}
