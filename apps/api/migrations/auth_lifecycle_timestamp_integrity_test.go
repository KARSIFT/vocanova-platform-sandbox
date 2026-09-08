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

const authLifecycleTimestampMigration = "20260908190000_voc1415_auth_lifecycle_timestamps.sql"

var authLifecycleConstraintNames = []string{
	"sessions_revoked_at_not_before_created",
	"magic_links_consumed_at_within_lifetime",
	"magic_links_revoked_at_not_before_created",
	"oauth_states_consumed_at_within_lifetime",
	"email_change_links_consumed_at_within_lifetime",
	"email_change_links_revoked_at_not_before_created",
}

func TestAuthLifecycleTimestampMigrationCarriesDatabaseInvariants(t *testing.T) {
	body, err := os.ReadFile(authLifecycleTimestampMigration)
	require.NoError(t, err)
	text := string(body)
	for _, name := range authLifecycleConstraintNames {
		assert.Contains(t, text, name)
	}
	assert.Equal(t, 6, strings.Count(text, "\n  NOT VALID;"))
	assert.Equal(t, 3, strings.Count(text, "consumed_at >= created_at AND consumed_at < expires_at"))
	assert.Equal(t, 3, strings.Count(text, "revoked_at IS NULL OR revoked_at >= created_at"))

	for _, schemaFile := range []string{"session.go", "magiclink.go", "emailchangelink.go"} {
		schemaBody, readErr := os.ReadFile("../ent/schema/" + schemaFile)
		require.NoError(t, readErr)
		assert.Contains(t, string(schemaBody), "_at_not_before_created")
	}
	for _, schemaFile := range []string{"magiclink.go", "emailchangelink.go"} {
		schemaBody, readErr := os.ReadFile("../ent/schema/" + schemaFile)
		require.NoError(t, readErr)
		assert.Contains(t, string(schemaBody), "consumed_at >= created_at AND consumed_at < expires_at")
	}
	// oauth_states is deliberately owned by the SQL auth repository rather
	// than Ent; the migration and the real-PostgreSQL cases below cover it.
}

func TestAuthLifecycleTimestampIntegrityAgainstPostgreSQL(t *testing.T) {
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
	schema := "vocanova_auth_lifecycle_" + strings.ReplaceAll(uuid.NewString(), "-", "")
	_, err = db.ExecContext(ctx, "CREATE SCHEMA "+pq.QuoteIdentifier(schema))
	require.NoError(t, err)
	t.Cleanup(func() {
		_, dropErr := db.ExecContext(context.Background(), "DROP SCHEMA IF EXISTS "+pq.QuoteIdentifier(schema)+" CASCADE")
		assert.NoError(t, dropErr)
	})
	_, err = db.ExecContext(ctx, "SET search_path TO "+pq.QuoteIdentifier(schema))
	require.NoError(t, err)

	_, err = db.ExecContext(ctx, `
		CREATE TABLE sessions (
			id uuid PRIMARY KEY,
			created_at timestamptz NOT NULL,
			expires_at timestamptz NOT NULL CHECK (expires_at > created_at),
			revoked_at timestamptz,
			marker integer NOT NULL DEFAULT 0
		);
		CREATE TABLE magic_links (
			id uuid PRIMARY KEY,
			created_at timestamptz NOT NULL,
			expires_at timestamptz NOT NULL CHECK (expires_at > created_at),
			consumed_at timestamptz,
			revoked_at timestamptz,
			marker integer NOT NULL DEFAULT 0,
			CHECK (consumed_at IS NULL OR revoked_at IS NULL)
		);
		CREATE TABLE oauth_states (
			id uuid PRIMARY KEY,
			created_at timestamptz NOT NULL,
			expires_at timestamptz NOT NULL CHECK (expires_at > created_at),
			consumed_at timestamptz,
			marker integer NOT NULL DEFAULT 0
		);
		CREATE TABLE email_change_links (
			id uuid PRIMARY KEY,
			created_at timestamptz NOT NULL,
			expires_at timestamptz NOT NULL CHECK (expires_at > created_at),
			consumed_at timestamptz,
			revoked_at timestamptz,
			marker integer NOT NULL DEFAULT 0,
			CHECK (consumed_at IS NULL OR revoked_at IS NULL)
		)`)
	require.NoError(t, err)

	now := time.Now().UTC().Truncate(time.Microsecond)
	legacy := map[string]uuid.UUID{
		"sessions":           uuid.New(),
		"magic_links":        uuid.New(),
		"oauth_states":       uuid.New(),
		"email_change_links": uuid.New(),
	}
	_, err = db.ExecContext(ctx, `INSERT INTO sessions (id, created_at, expires_at, revoked_at) VALUES ($1, $2, $3, $4)`, legacy["sessions"], now, now.Add(30*24*time.Hour), now.Add(-time.Second))
	require.NoError(t, err)
	_, err = db.ExecContext(ctx, `INSERT INTO magic_links (id, created_at, expires_at, consumed_at) VALUES ($1, $2, $3, $4)`, legacy["magic_links"], now, now.Add(15*time.Minute), now.Add(time.Hour))
	require.NoError(t, err)
	_, err = db.ExecContext(ctx, `INSERT INTO oauth_states (id, created_at, expires_at, consumed_at) VALUES ($1, $2, $3, $4)`, legacy["oauth_states"], now, now.Add(10*time.Minute), now.Add(-time.Second))
	require.NoError(t, err)
	_, err = db.ExecContext(ctx, `INSERT INTO email_change_links (id, created_at, expires_at, revoked_at) VALUES ($1, $2, $3, $4)`, legacy["email_change_links"], now, now.Add(15*time.Minute), now.Add(-time.Second))
	require.NoError(t, err)

	migration, err := os.ReadFile(authLifecycleTimestampMigration)
	require.NoError(t, err)
	_, err = db.ExecContext(ctx, string(migration))
	require.NoError(t, err, "NOT VALID constraints must preserve legacy chronology contradictions")

	type lifecycleWrite struct {
		name      string
		statement string
		args      []any
	}
	validWrites := []lifecycleWrite{
		{name: "session revoked at creation", statement: `INSERT INTO sessions (id, created_at, expires_at, revoked_at) VALUES ($1, $2, $3, $2)`, args: []any{uuid.New(), now, now.Add(30 * 24 * time.Hour)}},
		{name: "session revoked after expiry", statement: `INSERT INTO sessions (id, created_at, expires_at, revoked_at) VALUES ($1, $2, $3, $4)`, args: []any{uuid.New(), now, now.Add(30 * 24 * time.Hour), now.Add(31 * 24 * time.Hour)}},
		{name: "magic consumed at creation", statement: `INSERT INTO magic_links (id, created_at, expires_at, consumed_at) VALUES ($1, $2, $3, $2)`, args: []any{uuid.New(), now, now.Add(15 * time.Minute)}},
		{name: "magic consumed immediately before expiry", statement: `INSERT INTO magic_links (id, created_at, expires_at, consumed_at) VALUES ($1, $2, $3, $4)`, args: []any{uuid.New(), now, now.Add(15 * time.Minute), now.Add(15*time.Minute - time.Microsecond)}},
		{name: "magic revoked after expiry", statement: `INSERT INTO magic_links (id, created_at, expires_at, revoked_at) VALUES ($1, $2, $3, $4)`, args: []any{uuid.New(), now, now.Add(15 * time.Minute), now.Add(time.Hour)}},
		{name: "oauth consumed at creation", statement: `INSERT INTO oauth_states (id, created_at, expires_at, consumed_at) VALUES ($1, $2, $3, $2)`, args: []any{uuid.New(), now, now.Add(10 * time.Minute)}},
		{name: "oauth consumed immediately before expiry", statement: `INSERT INTO oauth_states (id, created_at, expires_at, consumed_at) VALUES ($1, $2, $3, $4)`, args: []any{uuid.New(), now, now.Add(10 * time.Minute), now.Add(10*time.Minute - time.Microsecond)}},
		{name: "email change consumed at creation", statement: `INSERT INTO email_change_links (id, created_at, expires_at, consumed_at) VALUES ($1, $2, $3, $2)`, args: []any{uuid.New(), now, now.Add(15 * time.Minute)}},
		{name: "email change consumed immediately before expiry", statement: `INSERT INTO email_change_links (id, created_at, expires_at, consumed_at) VALUES ($1, $2, $3, $4)`, args: []any{uuid.New(), now, now.Add(15 * time.Minute), now.Add(15*time.Minute - time.Microsecond)}},
		{name: "email change revoked after expiry", statement: `INSERT INTO email_change_links (id, created_at, expires_at, revoked_at) VALUES ($1, $2, $3, $4)`, args: []any{uuid.New(), now, now.Add(15 * time.Minute), now.Add(time.Hour)}},
	}
	for _, write := range validWrites {
		t.Run("valid "+write.name, func(t *testing.T) {
			_, insertErr := db.ExecContext(ctx, write.statement, write.args...)
			require.NoError(t, insertErr)
		})
	}

	invalidWrites := []lifecycleWrite{
		{name: "session revoked before creation", statement: `INSERT INTO sessions (id, created_at, expires_at, revoked_at) VALUES ($1, $2, $3, $4)`, args: []any{uuid.New(), now, now.Add(30 * 24 * time.Hour), now.Add(-time.Second)}},
		{name: "magic consumed before creation", statement: `INSERT INTO magic_links (id, created_at, expires_at, consumed_at) VALUES ($1, $2, $3, $4)`, args: []any{uuid.New(), now, now.Add(15 * time.Minute), now.Add(-time.Second)}},
		{name: "magic consumed after expiry", statement: `INSERT INTO magic_links (id, created_at, expires_at, consumed_at) VALUES ($1, $2, $3, $4)`, args: []any{uuid.New(), now, now.Add(15 * time.Minute), now.Add(time.Hour)}},
		{name: "magic consumed at expiry", statement: `INSERT INTO magic_links (id, created_at, expires_at, consumed_at) VALUES ($1, $2, $3, $3)`, args: []any{uuid.New(), now, now.Add(15 * time.Minute)}},
		{name: "magic revoked before creation", statement: `INSERT INTO magic_links (id, created_at, expires_at, revoked_at) VALUES ($1, $2, $3, $4)`, args: []any{uuid.New(), now, now.Add(15 * time.Minute), now.Add(-time.Second)}},
		{name: "oauth consumed before creation", statement: `INSERT INTO oauth_states (id, created_at, expires_at, consumed_at) VALUES ($1, $2, $3, $4)`, args: []any{uuid.New(), now, now.Add(10 * time.Minute), now.Add(-time.Second)}},
		{name: "oauth consumed after expiry", statement: `INSERT INTO oauth_states (id, created_at, expires_at, consumed_at) VALUES ($1, $2, $3, $4)`, args: []any{uuid.New(), now, now.Add(10 * time.Minute), now.Add(time.Hour)}},
		{name: "oauth consumed at expiry", statement: `INSERT INTO oauth_states (id, created_at, expires_at, consumed_at) VALUES ($1, $2, $3, $3)`, args: []any{uuid.New(), now, now.Add(10 * time.Minute)}},
		{name: "email change consumed before creation", statement: `INSERT INTO email_change_links (id, created_at, expires_at, consumed_at) VALUES ($1, $2, $3, $4)`, args: []any{uuid.New(), now, now.Add(15 * time.Minute), now.Add(-time.Second)}},
		{name: "email change consumed after expiry", statement: `INSERT INTO email_change_links (id, created_at, expires_at, consumed_at) VALUES ($1, $2, $3, $4)`, args: []any{uuid.New(), now, now.Add(15 * time.Minute), now.Add(time.Hour)}},
		{name: "email change consumed at expiry", statement: `INSERT INTO email_change_links (id, created_at, expires_at, consumed_at) VALUES ($1, $2, $3, $3)`, args: []any{uuid.New(), now, now.Add(15 * time.Minute)}},
		{name: "email change revoked before creation", statement: `INSERT INTO email_change_links (id, created_at, expires_at, revoked_at) VALUES ($1, $2, $3, $4)`, args: []any{uuid.New(), now, now.Add(15 * time.Minute), now.Add(-time.Second)}},
	}
	for _, write := range invalidWrites {
		t.Run("invalid "+write.name, func(t *testing.T) {
			_, insertErr := db.ExecContext(ctx, write.statement, write.args...)
			requireAuthLifecycleCheckViolation(t, insertErr)
		})
	}

	for table, id := range legacy {
		t.Run("legacy update "+table, func(t *testing.T) {
			_, updateErr := db.ExecContext(ctx,
				"UPDATE "+pq.QuoteIdentifier(table)+" SET marker = marker + 1 WHERE id = $1", id)
			requireAuthLifecycleCheckViolation(t, updateErr)
		})
	}

	var legacyRows int
	for table, id := range legacy {
		require.NoError(t, db.QueryRowContext(ctx,
			"SELECT count(*) FROM "+pq.QuoteIdentifier(table)+" WHERE id = $1", id).Scan(&legacyRows))
		assert.Equal(t, 1, legacyRows)
	}

	rows, err := db.QueryContext(ctx, `
		SELECT conname, convalidated
		FROM pg_constraint
		WHERE conname = ANY($1)
		ORDER BY conname`, pq.Array(authLifecycleConstraintNames))
	require.NoError(t, err)
	defer rows.Close()
	seen := map[string]bool{}
	for rows.Next() {
		var name string
		var validated bool
		require.NoError(t, rows.Scan(&name, &validated))
		assert.False(t, validated)
		seen[name] = true
	}
	require.NoError(t, rows.Err())
	assert.Len(t, seen, len(authLifecycleConstraintNames))
}

func requireAuthLifecycleCheckViolation(t *testing.T, err error) {
	t.Helper()
	require.Error(t, err)
	var pqErr *pq.Error
	require.True(t, errors.As(err, &pqErr), "expected PostgreSQL error, got %T: %v", err, err)
	assert.Equal(t, pq.ErrorCode("23514"), pqErr.Code)
}
