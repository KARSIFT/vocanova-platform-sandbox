//go:build integration

package migrations_test

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

// TestVOC1445EnOnlyAppLanguageConstraintOnPostgres proves that the staged
// constraint aligns direct database writers with VOC-031-D06. The temporary
// schema lets this exercise both the original migration and its forward-only
// successor without mutating the shared validation database's migration state.
func TestVOC1445EnOnlyAppLanguageConstraintOnPostgres(t *testing.T) {
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

	schemaName := "voc1445_" + strings.ReplaceAll(uuid.NewString(), "-", "")
	if _, err := db.ExecContext(ctx, "CREATE SCHEMA "+pq.QuoteIdentifier(schemaName)); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_, _ = db.ExecContext(context.Background(), "DROP SCHEMA "+pq.QuoteIdentifier(schemaName)+" CASCADE")
	})
	conn, err := db.Conn(ctx)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = conn.Close() })
	if _, err := conn.ExecContext(ctx, "SET search_path TO "+pq.QuoteIdentifier(schemaName)); err != nil {
		t.Fatal(err)
	}

	applyMigrationFile(t, ctx, conn, "20260724210000_identity_foundation.sql")
	applyMigrationFile(t, ctx, conn, "20260725130000_voc030_p4_user_settings.sql")

	legacyUserID := insertVOC1445User(t, ctx, conn)
	legacySettingsID := uuid.New()
	now := time.Now().UTC().Truncate(time.Microsecond)
	if _, err := conn.ExecContext(ctx, `INSERT INTO user_settings (id, user_id, app_language, created_at, updated_at) VALUES ($1, $2, 'fr', $3, $3)`, legacySettingsID, legacyUserID, now); err != nil {
		t.Fatalf("precondition: original constraint rejected legacy app language: %v", err)
	}

	applyMigrationFile(t, ctx, conn, "20260908300000_voc1445_en_only_app_language.sql")

	var definition string
	var validated bool
	if err := conn.QueryRowContext(ctx, `SELECT pg_get_constraintdef(oid), convalidated FROM pg_constraint WHERE conname = 'user_settings_app_language_valid' AND conrelid = 'user_settings'::regclass`).Scan(&definition, &validated); err != nil {
		t.Fatalf("read app-language constraint catalog entry: %v", err)
	}
	if validated {
		t.Fatal("en-only constraint unexpectedly validated legacy rows")
	}
	if !strings.Contains(definition, "app_language = 'en'") {
		t.Fatalf("constraint definition = %q, want exact en-only predicate", definition)
	}

	var legacyLanguage string
	if err := conn.QueryRowContext(ctx, "SELECT app_language FROM user_settings WHERE id = $1", legacySettingsID).Scan(&legacyLanguage); err != nil {
		t.Fatal(err)
	}
	if legacyLanguage != "fr" {
		t.Fatalf("legacy app language = %q, want retained fr", legacyLanguage)
	}
	if _, err := conn.ExecContext(ctx, "UPDATE user_settings SET timezone = 'Europe/Paris' WHERE id = $1", legacySettingsID); err == nil {
		t.Fatal("legacy row update unexpectedly bypassed the staged en-only constraint")
	} else {
		assertVOC1445CheckViolation(t, err)
	}
	if _, err := conn.ExecContext(ctx, "UPDATE user_settings SET app_language = 'en' WHERE id = $1", legacySettingsID); err != nil {
		t.Fatalf("correct legacy language to en: %v", err)
	}

	exactUserID := insertVOC1445User(t, ctx, conn)
	exactSettingsID := uuid.New()
	if _, err := conn.ExecContext(ctx, `INSERT INTO user_settings (id, user_id, app_language, created_at, updated_at) VALUES ($1, $2, 'en', $3, $3)`, exactSettingsID, exactUserID, now); err != nil {
		t.Fatalf("insert exact en language: %v", err)
	}
	for _, value := range []string{"fr", "EN", " en "} {
		userID := insertVOC1445User(t, ctx, conn)
		if _, err := conn.ExecContext(ctx, "INSERT INTO user_settings (id, user_id, app_language, created_at, updated_at) VALUES ($1, $2, $3, $4, $4)", uuid.New(), userID, value, now); err == nil {
			t.Fatalf("insert app language %q unexpectedly succeeded", value)
		} else {
			assertVOC1445CheckViolation(t, err)
		}
		if _, err := conn.ExecContext(ctx, "UPDATE user_settings SET app_language = $1 WHERE id = $2", value, exactSettingsID); err == nil {
			t.Fatalf("update app language %q unexpectedly succeeded", value)
		} else {
			assertVOC1445CheckViolation(t, err)
		}
	}
}

func applyMigrationFile(t *testing.T, ctx context.Context, conn *sql.Conn, name string) {
	t.Helper()
	contents, err := os.ReadFile(name)
	if err != nil {
		t.Fatalf("read migration %s: %v", name, err)
	}
	if _, err := conn.ExecContext(ctx, string(contents)); err != nil {
		t.Fatalf("apply migration %s: %v", name, err)
	}
}

func insertVOC1445User(t *testing.T, ctx context.Context, conn *sql.Conn) uuid.UUID {
	t.Helper()
	userID := uuid.New()
	now := time.Now().UTC().Truncate(time.Microsecond)
	if _, err := conn.ExecContext(ctx, "INSERT INTO users (id, email, status, created_at, updated_at) VALUES ($1, $2, 'active', $3, $3)", userID, fmt.Sprintf("%s@example.test", userID), now); err != nil {
		t.Fatal(err)
	}
	return userID
}

func assertVOC1445CheckViolation(t *testing.T, err error) {
	t.Helper()
	var pqErr *pq.Error
	if !errors.As(err, &pqErr) || pqErr.Code != "23514" || pqErr.Constraint != "user_settings_app_language_valid" {
		t.Fatalf("error = %v, want user_settings_app_language_valid check violation", err)
	}
}
