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

	"github.com/google/uuid"
	"github.com/lib/pq"
)

// TestVOC1414TimezoneCaptureIntegrityAgainstRealPostgres proves the
// forward-only migration keeps legacy malformed rows deployable while
// PostgreSQL enforces a usable timezone capture for every new or updated row.
func TestVOC1414TimezoneCaptureIntegrityAgainstRealPostgres(t *testing.T) {
	ctx := context.Background()
	db := isolatedTimezoneCapturePostgres(t)

	for _, table := range timezoneCaptureTables {
		if _, err := db.ExecContext(ctx, fmt.Sprintf("CREATE TABLE %s (id uuid PRIMARY KEY, timezone text NOT NULL)", table)); err != nil {
			t.Fatalf("create legacy %s table: %v", table, err)
		}
		if _, err := db.ExecContext(ctx, fmt.Sprintf("INSERT INTO %s (id, timezone) VALUES ($1, $2)", table), uuid.New(), " \t\n "); err != nil {
			t.Fatalf("insert legacy whitespace timezone into %s: %v", table, err)
		}
	}

	migration, err := os.ReadFile("20260908030000_voc1414_timezone_capture_integrity.sql")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.ExecContext(ctx, string(migration)); err != nil {
		t.Fatalf("apply timezone capture migration: %v", err)
	}

	for _, table := range timezoneCaptureTables {
		t.Run(table, func(t *testing.T) {
			assertTimezoneCaptureConstraintState(t, ctx, db, table)
			assertTimezoneCaptureLegacyRowPreserved(t, ctx, db, table)
			assertTimezoneCaptureWriteBehavior(t, ctx, db, table)
		})
	}
}

var timezoneCaptureTables = []string{
	"daily_mission_snapshots",
	"daily_activity_summaries",
	"streak_states",
	"grace_day_ledger",
}

func isolatedTimezoneCapturePostgres(t *testing.T) *sql.DB {
	t.Helper()
	dsn := os.Getenv("VOCANOVA_TEST_POSTGRES_DSN")
	if dsn == "" {
		t.Skip("VOCANOVA_TEST_POSTGRES_DSN is unset; real PostgreSQL test unavailable")
	}
	if strings.HasPrefix(dsn, "postgres://") || strings.HasPrefix(dsn, "postgresql://") {
		var err error
		dsn, err = pq.ParseURL(dsn)
		if err != nil {
			t.Fatal(err)
		}
	}
	admin, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatal(err)
	}
	schema := "timezone_capture_" + strings.ReplaceAll(uuid.NewString(), "-", "")
	if _, err := admin.ExecContext(t.Context(), "CREATE SCHEMA "+schema); err != nil {
		_ = admin.Close()
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if _, err := admin.ExecContext(context.Background(), "DROP SCHEMA "+schema+" CASCADE"); err != nil {
			t.Errorf("drop isolated schema: %v", err)
		}
		_ = admin.Close()
	})

	db, err := sql.Open("postgres", dsn+" search_path="+schema)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return db
}

func assertTimezoneCaptureConstraintState(t *testing.T, ctx context.Context, db *sql.DB, table string) {
	t.Helper()
	constraint := table + "_timezone_nonblank"
	var validated bool
	err := db.QueryRowContext(ctx, `
		SELECT c.convalidated
		FROM pg_constraint AS c
		JOIN pg_namespace AS n ON n.oid = c.connamespace
		WHERE n.nspname = current_schema() AND c.conname = $1`, constraint).Scan(&validated)
	if err != nil {
		t.Fatalf("read %s catalog state: %v", constraint, err)
	}
	if validated {
		t.Fatalf("%s must remain NOT VALID for rollout safety", constraint)
	}
}

func assertTimezoneCaptureLegacyRowPreserved(t *testing.T, ctx context.Context, db *sql.DB, table string) {
	t.Helper()
	var count int
	if err := db.QueryRowContext(ctx, fmt.Sprintf("SELECT count(*) FROM %s WHERE timezone = $1", table), " \t\n ").Scan(&count); err != nil {
		t.Fatalf("count legacy %s row: %v", table, err)
	}
	if count != 1 {
		t.Fatalf("legacy %s rows = %d, want 1", table, count)
	}
}

func assertTimezoneCaptureWriteBehavior(t *testing.T, ctx context.Context, db *sql.DB, table string) {
	t.Helper()
	if _, err := db.ExecContext(ctx, fmt.Sprintf("INSERT INTO %s (id, timezone) VALUES ($1, $2)", table), uuid.New(), "Asia/Tehran"); err != nil {
		t.Fatalf("insert valid timezone into %s: %v", table, err)
	}
	for _, invalid := range []string{"", " \t\n "} {
		_, err := db.ExecContext(ctx, fmt.Sprintf("INSERT INTO %s (id, timezone) VALUES ($1, $2)", table), uuid.New(), invalid)
		assertTimezoneCaptureCheckViolation(t, err, table+" insert")
	}

	var legacyID uuid.UUID
	if err := db.QueryRowContext(ctx, fmt.Sprintf("SELECT id FROM %s WHERE timezone = $1", table), " \t\n ").Scan(&legacyID); err != nil {
		t.Fatalf("select legacy %s row: %v", table, err)
	}
	_, err := db.ExecContext(ctx, fmt.Sprintf("UPDATE %s SET timezone = $1 WHERE id = $2", table), "", legacyID)
	assertTimezoneCaptureCheckViolation(t, err, table+" update")
}

func assertTimezoneCaptureCheckViolation(t *testing.T, err error, operation string) {
	t.Helper()
	if err == nil {
		t.Fatalf("%s unexpectedly accepted a blank timezone", operation)
	}
	var pqErr *pq.Error
	if !errors.As(err, &pqErr) || string(pqErr.Code) != "23514" {
		t.Fatalf("%s error = %v, want CHECK violation (23514)", operation, err)
	}
}
