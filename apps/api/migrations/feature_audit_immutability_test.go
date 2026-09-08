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

const featureAuditImmutabilityMigration = "20260908260000_voc1437_feature_audit_immutability.sql"

func TestFeatureAuditImmutabilityMigrationIsNarrowlyScoped(t *testing.T) {
	body, err := os.ReadFile(featureAuditImmutabilityMigration)
	require.NoError(t, err)
	text := string(body)
	for _, invariant := range []string{
		"CREATE FUNCTION vocanova_guard_feature_audit_mutation()",
		"TG_OP = 'UPDATE'",
		"current_setting('vocanova.ledger_purge', true) = 'on'",
		"NEW.user_id IS NULL",
		"NEW.action = OLD.action",
		"NEW.entity_type = OLD.entity_type",
		"NEW.entity_id IS NULL",
		"NEW.request_id IS NOT DISTINCT FROM OLD.request_id",
		"NEW.actor_type = OLD.actor_type",
		"NEW.actor_id IS NULL",
		"NEW.metadata = '{}'::jsonb",
		"NEW.created_at = OLD.created_at",
		"ERRCODE = '55000'",
		"BEFORE UPDATE OR DELETE ON feature_audit_logs",
	} {
		assert.Contains(t, text, invariant)
	}

	accountRepository, err := os.ReadFile("../business/accounts/postgres.go")
	require.NoError(t, err)
	assert.Contains(t, string(accountRepository), "SELECT set_config('vocanova.ledger_purge', 'on', true)")
	assert.Contains(t, string(accountRepository), "SET user_id = NULL, actor_id = NULL, entity_id = NULL,")
}

func TestFeatureAuditImmutabilityAgainstPostgreSQL(t *testing.T) {
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
	schema := "vocanova_feature_audit_immutable_" + strings.ReplaceAll(uuid.NewString(), "-", "")
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
		"20260908020000_voc1352_feature_audit_logs.sql",
		featureAuditImmutabilityMigration,
	} {
		migration, readErr := os.ReadFile(filename)
		require.NoError(t, readErr)
		_, err = db.ExecContext(ctx, string(migration))
		require.NoError(t, err)
	}

	now := time.Now().UTC().Truncate(time.Microsecond)
	userID := uuid.New()
	auditID := uuid.New()
	entityID := uuid.New()
	_, err = db.ExecContext(ctx, `INSERT INTO users
		(id, email, status, onboarding_status, created_at, updated_at)
		VALUES ($1, 'audit@example.test', 'active', 'completed', $2, $2)`, userID, now)
	require.NoError(t, err)
	_, err = db.ExecContext(ctx, `INSERT INTO feature_audit_logs
		(id, user_id, action, entity_type, entity_id, request_id, actor_type, actor_id, metadata, created_at, updated_at)
		VALUES ($1, $2, 'user_word_saved', 'user_word', $3, 'request-1', 'user', $2, '{"source":"journey"}', $4, $4)`,
		auditID, userID, entityID, now)
	require.NoError(t, err, "normal immutable-history inserts remain allowed")

	_, err = db.ExecContext(ctx, `UPDATE feature_audit_logs SET action = 'rewritten' WHERE id = $1`, auditID)
	requireFeatureAuditImmutabilityViolation(t, err)
	_, err = db.ExecContext(ctx, `DELETE FROM feature_audit_logs WHERE id = $1`, auditID)
	requireFeatureAuditImmutabilityViolation(t, err)

	rollbackTx, err := db.BeginTx(ctx, nil)
	require.NoError(t, err)
	_, err = rollbackTx.ExecContext(ctx, `SELECT set_config('vocanova.ledger_purge', 'on', true)`)
	require.NoError(t, err)
	_, err = rollbackTx.ExecContext(ctx, `UPDATE feature_audit_logs
		SET user_id = NULL, actor_id = NULL, entity_id = NULL,
		    metadata = '{}'::jsonb, updated_at = $2
		WHERE id = $1`, auditID, now.Add(time.Second))
	require.NoError(t, err, "the exact account-deletion de-identification is allowed inside the gate")
	require.NoError(t, rollbackTx.Rollback())

	_, err = db.ExecContext(ctx, `UPDATE feature_audit_logs SET updated_at = $2 WHERE id = $1`, auditID, now.Add(2*time.Second))
	requireFeatureAuditImmutabilityViolation(t, err)

	misuseUpdateTx, err := db.BeginTx(ctx, nil)
	require.NoError(t, err)
	_, err = misuseUpdateTx.ExecContext(ctx, `SELECT set_config('vocanova.ledger_purge', 'on', true)`)
	require.NoError(t, err)
	_, err = misuseUpdateTx.ExecContext(ctx, `UPDATE feature_audit_logs
		SET user_id = NULL, actor_id = NULL, entity_id = NULL,
		    metadata = '{}'::jsonb, action = 'rewritten', updated_at = $2
		WHERE id = $1`, auditID, now.Add(time.Second))
	requireFeatureAuditImmutabilityViolation(t, err)
	require.NoError(t, misuseUpdateTx.Rollback())

	misuseDeleteTx, err := db.BeginTx(ctx, nil)
	require.NoError(t, err)
	_, err = misuseDeleteTx.ExecContext(ctx, `SELECT set_config('vocanova.ledger_purge', 'on', true)`)
	require.NoError(t, err)
	_, err = misuseDeleteTx.ExecContext(ctx, `DELETE FROM feature_audit_logs WHERE id = $1`, auditID)
	requireFeatureAuditImmutabilityViolation(t, err)
	require.NoError(t, misuseDeleteTx.Rollback())

	commitTx, err := db.BeginTx(ctx, nil)
	require.NoError(t, err)
	_, err = commitTx.ExecContext(ctx, `SELECT set_config('vocanova.ledger_purge', 'on', true)`)
	require.NoError(t, err)
	_, err = commitTx.ExecContext(ctx, `UPDATE feature_audit_logs
		SET user_id = NULL, actor_id = NULL, entity_id = NULL,
		    metadata = '{}'::jsonb, updated_at = $2
		WHERE id = $1`, auditID, now.Add(time.Second))
	require.NoError(t, err)
	require.NoError(t, commitTx.Commit())

	var userIDAfter, actorIDAfter, entityIDAfter uuid.NullUUID
	var action, entityType, requestID, actorType, metadata string
	var createdAt time.Time
	require.NoError(t, db.QueryRowContext(ctx, `SELECT user_id, actor_id, entity_id,
		action, entity_type, request_id, actor_type, metadata::text, created_at
		FROM feature_audit_logs WHERE id = $1`, auditID).Scan(
		&userIDAfter, &actorIDAfter, &entityIDAfter, &action, &entityType, &requestID, &actorType, &metadata, &createdAt))
	assert.False(t, userIDAfter.Valid)
	assert.False(t, actorIDAfter.Valid)
	assert.False(t, entityIDAfter.Valid)
	assert.Equal(t, "user_word_saved", action)
	assert.Equal(t, "user_word", entityType)
	assert.Equal(t, "request-1", requestID)
	assert.Equal(t, "user", actorType)
	assert.Equal(t, "{}", metadata)
	assert.True(t, createdAt.Equal(now), "creation instant must remain unchanged")

	_, err = db.ExecContext(ctx, `UPDATE feature_audit_logs SET updated_at = $2 WHERE id = $1`, auditID, now.Add(3*time.Second))
	requireFeatureAuditImmutabilityViolation(t, err, "the purge gate must reset after commit")

	var triggerDefinition string
	require.NoError(t, db.QueryRowContext(ctx, `
		SELECT pg_get_triggerdef(oid)
		FROM pg_trigger
		WHERE tgrelid = 'feature_audit_logs'::regclass
		  AND tgname = 'feature_audit_logs_immutable'
		  AND NOT tgisinternal`).Scan(&triggerDefinition))
	assert.Contains(t, triggerDefinition, "BEFORE DELETE OR UPDATE")
	assert.Contains(t, triggerDefinition, "vocanova_guard_feature_audit_mutation")
}

func requireFeatureAuditImmutabilityViolation(t *testing.T, err error, messageAndArgs ...any) {
	t.Helper()
	require.Error(t, err, messageAndArgs...)
	var pqErr *pq.Error
	require.True(t, errors.As(err, &pqErr), "expected PostgreSQL error, got %T: %v", err, err)
	assert.Equal(t, pq.ErrorCode("55000"), pqErr.Code)
}
