//go:build integration

package migrations_test

import (
	"context"
	"database/sql"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

// TestVOC1352FeatureAuditLogsMigrationOnPostgres proves the migration accepts
// a documented event while preserving the restrictive ownership constraint.
func TestVOC1352FeatureAuditLogsMigrationOnPostgres(t *testing.T) {
	dsn := os.Getenv("VOCANOVA_TEST_POSTGRES_DSN")
	if dsn == "" {
		t.Skip("VOCANOVA_TEST_POSTGRES_DSN is not set")
	}
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	ctx := context.Background()
	if err := db.PingContext(ctx); err != nil {
		t.Fatal(err)
	}

	var table sql.NullString
	if err := db.QueryRowContext(ctx, `SELECT to_regclass('feature_audit_logs')`).Scan(&table); err != nil {
		t.Fatal(err)
	}
	if !table.Valid {
		migration, err := os.ReadFile("20260908020000_voc1352_feature_audit_logs.sql")
		if err != nil {
			t.Fatal(err)
		}
		if _, err := db.ExecContext(ctx, string(migration)); err != nil {
			t.Fatalf("apply feature audit migration: %v", err)
		}
	}

	now := time.Now().UTC().Truncate(time.Microsecond)
	userID, auditID := uuid.New(), uuid.New()
	if _, err := db.ExecContext(ctx, `INSERT INTO users (id, email, status, created_at, updated_at) VALUES ($1, $2, 'active', $3, $3)`, userID, userID.String()+"@example.test", now); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_, _ = db.ExecContext(context.Background(), `DELETE FROM feature_audit_logs WHERE id = $1`, auditID)
		_, _ = db.ExecContext(context.Background(), `DELETE FROM users WHERE id = $1`, userID)
	})
	if _, err := db.ExecContext(ctx, `INSERT INTO feature_audit_logs (id, user_id, action, entity_type, entity_id, actor_type, actor_id, metadata, created_at, updated_at) VALUES ($1, $2, 'user_word_saved', 'user_word', $3, 'user', $2, '{"source":"integration"}', $4, $4)`, auditID, userID, uuid.New(), now); err != nil {
		t.Fatal(err)
	}
	if _, err := db.ExecContext(ctx, `DELETE FROM users WHERE id = $1`, userID); err == nil {
		t.Fatal("feature audit log user FK unexpectedly allowed deletion")
	} else {
		var pqErr *pq.Error
		if !errors.As(err, &pqErr) || string(pqErr.Code) != "23503" {
			t.Fatalf("expected foreign-key violation, got %v", err)
		}
	}
}
