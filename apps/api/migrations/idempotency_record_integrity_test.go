package migrations_test

import (
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

const idempotencyRecordIntegrityMigration = "20260909142063_voc1411_idempotency_record_integrity.sql"

func TestIdempotencyRecordIntegrityMigrationCarriesDatabaseInvariants(t *testing.T) {
	body, err := os.ReadFile(idempotencyRecordIntegrityMigration)
	require.NoError(t, err)
	text := string(body)
	assert.Contains(t, text, "idempotency_keys_user_id_fkey")
	assert.Contains(t, text, "FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE RESTRICT")
	assert.Contains(t, text, "idempotency_keys_fingerprint_nonempty")
	assert.Contains(t, text, "CHECK (char_length(fingerprint) > 0)")
	assert.Equal(t, 2, strings.Count(text, "\n  NOT VALID;"))
}

func TestIdempotencyRecordIntegrityAgainstPostgreSQL(t *testing.T) {
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
	schema := "vocanova_idempotency_integrity_" + strings.ReplaceAll(uuid.NewString(), "-", "")
	_, err = db.ExecContext(ctx, "CREATE SCHEMA "+pq.QuoteIdentifier(schema))
	require.NoError(t, err)
	t.Cleanup(func() {
		_, dropErr := db.ExecContext(context.Background(), "DROP SCHEMA IF EXISTS "+pq.QuoteIdentifier(schema)+" CASCADE")
		assert.NoError(t, dropErr)
	})
	_, err = db.ExecContext(ctx, "SET search_path TO "+pq.QuoteIdentifier(schema))
	require.NoError(t, err)

	_, err = db.ExecContext(ctx, `
		CREATE TABLE users (id uuid PRIMARY KEY);
		CREATE TABLE idempotency_keys (
			id uuid PRIMARY KEY,
			user_id uuid NOT NULL,
			operation text NOT NULL CHECK (operation <> ''),
			key text NOT NULL CHECK (key <> ''),
			fingerprint text NOT NULL,
			created_at timestamptz NOT NULL
		)`)
	require.NoError(t, err)

	ownerID := uuid.New()
	legacyOrphanID := uuid.New()
	legacyEmptyID := uuid.New()
	_, err = db.ExecContext(ctx, `INSERT INTO users (id) VALUES ($1)`, ownerID)
	require.NoError(t, err)
	_, err = db.ExecContext(ctx, `
		INSERT INTO idempotency_keys (id, user_id, operation, key, fingerprint, created_at)
		VALUES
			($1, $2, 'word_addition', 'legacy-orphan', 'fingerprint', NOW()),
			($3, $4, 'word_addition', 'legacy-empty', '', NOW())`,
		legacyOrphanID, uuid.New(), legacyEmptyID, ownerID)
	require.NoError(t, err, "the original table permits orphan and empty-fingerprint rows")

	migration, err := os.ReadFile(idempotencyRecordIntegrityMigration)
	require.NoError(t, err)
	_, err = db.ExecContext(ctx, string(migration))
	require.NoError(t, err, "NOT VALID constraints must preserve legacy contradictions")

	validID := uuid.New()
	_, err = db.ExecContext(ctx, `
		INSERT INTO idempotency_keys (id, user_id, operation, key, fingerprint, created_at)
		VALUES ($1, $2, 'word_addition', 'valid', 'meaning|journey', NOW())`, validID, ownerID)
	require.NoError(t, err)

	_, err = db.ExecContext(ctx, `
		INSERT INTO idempotency_keys (id, user_id, operation, key, fingerprint, created_at)
		VALUES ($1, $2, 'word_addition', 'orphan', 'fingerprint', NOW())`, uuid.New(), uuid.New())
	requirePostgreSQLErrorCode(t, err, "23503")

	_, err = db.ExecContext(ctx, `
		INSERT INTO idempotency_keys (id, user_id, operation, key, fingerprint, created_at)
		VALUES ($1, $2, 'word_addition', 'empty', '', NOW())`, uuid.New(), ownerID)
	requirePostgreSQLErrorCode(t, err, "23514")

	_, err = db.ExecContext(ctx, `UPDATE idempotency_keys SET user_id = $1 WHERE id = $2`, uuid.New(), validID)
	requirePostgreSQLErrorCode(t, err, "23503")
	_, err = db.ExecContext(ctx, `UPDATE idempotency_keys SET fingerprint = '' WHERE id = $1`, validID)
	requirePostgreSQLErrorCode(t, err, "23514")
	_, err = db.ExecContext(ctx, `UPDATE idempotency_keys SET user_id = $1 WHERE id = $2`, uuid.New(), legacyOrphanID)
	requirePostgreSQLErrorCode(t, err, "23503")
	// PostgreSQL's FK trigger is UPDATE OF user_id, so NOT VALID deliberately
	// leaves a legacy orphan writable through unrelated columns. This is not a
	// hole for new claims: inserts and user_id changes are enforced above.
	_, err = db.ExecContext(ctx, `UPDATE idempotency_keys SET created_at = created_at WHERE id = $1`, legacyOrphanID)
	require.NoError(t, err)
	// CHECK constraints, unlike the foreign-key trigger, are evaluated for all
	// UPDATEs. An unrelated update cannot perpetuate the legacy empty value.
	_, err = db.ExecContext(ctx, `UPDATE idempotency_keys SET created_at = created_at WHERE id = $1`, legacyEmptyID)
	requirePostgreSQLErrorCode(t, err, "23514")

	var legacyRows int
	require.NoError(t, db.QueryRowContext(ctx, `
		SELECT count(*) FROM idempotency_keys WHERE id IN ($1, $2)`,
		legacyOrphanID, legacyEmptyID).Scan(&legacyRows))
	assert.Equal(t, 2, legacyRows)

	rows, err := db.QueryContext(ctx, `
		SELECT conname, contype, convalidated, confdeltype,
		       pg_get_constraintdef(oid)
		FROM pg_constraint
		WHERE conrelid = 'idempotency_keys'::regclass
		  AND conname IN ('idempotency_keys_user_id_fkey', 'idempotency_keys_fingerprint_nonempty')
		ORDER BY conname`)
	require.NoError(t, err)
	defer rows.Close()
	constraints := map[string]struct {
		kind       string
		validated  bool
		delete     string
		definition string
	}{}
	for rows.Next() {
		var name, kind, deleteAction, definition string
		var validated bool
		require.NoError(t, rows.Scan(&name, &kind, &validated, &deleteAction, &definition))
		constraints[name] = struct {
			kind       string
			validated  bool
			delete     string
			definition string
		}{kind: kind, validated: validated, delete: deleteAction, definition: definition}
	}
	require.NoError(t, rows.Err())
	require.Len(t, constraints, 2)
	assert.Equal(t, "f", constraints["idempotency_keys_user_id_fkey"].kind)
	assert.False(t, constraints["idempotency_keys_user_id_fkey"].validated)
	assert.Equal(t, "r", constraints["idempotency_keys_user_id_fkey"].delete)
	assert.Contains(t, constraints["idempotency_keys_user_id_fkey"].definition, "FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE RESTRICT")
	assert.Equal(t, "c", constraints["idempotency_keys_fingerprint_nonempty"].kind)
	assert.False(t, constraints["idempotency_keys_fingerprint_nonempty"].validated)
	assert.Contains(t, constraints["idempotency_keys_fingerprint_nonempty"].definition, "CHECK ((char_length(fingerprint) > 0))")

	_, err = db.ExecContext(ctx, `DELETE FROM users WHERE id = $1`, ownerID)
	requirePostgreSQLErrorCode(t, err, "23503")
	_, err = db.ExecContext(ctx, `DELETE FROM idempotency_keys WHERE user_id = $1`, ownerID)
	require.NoError(t, err)
	_, err = db.ExecContext(ctx, `DELETE FROM users WHERE id = $1`, ownerID)
	require.NoError(t, err, "the account purge's child-before-parent order remains valid")
}

func requirePostgreSQLErrorCode(t *testing.T, err error, code string) {
	t.Helper()
	require.Error(t, err)
	var pqErr *pq.Error
	require.True(t, errors.As(err, &pqErr), "expected PostgreSQL error, got %T: %v", err, err)
	assert.Equal(t, pq.ErrorCode(code), pqErr.Code)
}
