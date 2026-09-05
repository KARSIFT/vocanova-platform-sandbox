//go:build integration

package users

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/require"
)

// TestPostgreSQLRepositoryUpdateSettingsConcurrentPartialUpdatesPreserveBoth
// uses a PostgreSQL advisory-lock trigger to ensure both calls reach their
// INSERT before either can complete. It covers both an existing settings row
// and the concurrent first-write path. Before the conditional conflict update,
// the later request restored the other's stale default value.
func TestPostgreSQLRepositoryUpdateSettingsConcurrentPartialUpdatesPreserveBoth(t *testing.T) {
	dsn := os.Getenv("VOCANOVA_TEST_POSTGRES_DSN")
	if dsn == "" {
		t.Skip("VOCANOVA_TEST_POSTGRES_DSN is unset; real PostgreSQL concurrency test unavailable")
	}

	for _, existingRow := range []bool{true, false} {
		t.Run(fmt.Sprintf("existing_row=%t", existingRow), func(t *testing.T) {
			testConcurrentPartialSettingsUpdates(t, dsn, existingRow)
		})
	}
}

func testConcurrentPartialSettingsUpdates(t *testing.T, dsn string, existingRow bool) {
	t.Helper()
	ctx := t.Context()
	schema := "settings_concurrency_" + strings.ReplaceAll(uuid.NewString(), "-", "")
	admin, err := sql.Open("postgres", dsn)
	require.NoError(t, err)
	t.Cleanup(func() {
		_, _ = admin.ExecContext(context.Background(), "DROP SCHEMA "+schema+" CASCADE")
		_ = admin.Close()
	})
	require.NoError(t, admin.PingContext(ctx))

	_, err = admin.ExecContext(ctx, fmt.Sprintf(`
		CREATE SCHEMA %s;
		CREATE TABLE %s.users (id uuid PRIMARY KEY, display_name text, deleted_at timestamptz, updated_at timestamptz NOT NULL);
		CREATE TABLE %s.user_settings (
			id uuid PRIMARY KEY, user_id uuid NOT NULL UNIQUE REFERENCES %s.users(id), timezone text NOT NULL DEFAULT 'UTC',
			daily_review_target integer NOT NULL DEFAULT 20, review_interval_preset text NOT NULL DEFAULT 'vocanova_default',
			notifications_enabled boolean NOT NULL DEFAULT true, marketing_emails_enabled boolean NOT NULL DEFAULT false,
			app_language text NOT NULL DEFAULT 'en', created_at timestamptz NOT NULL, updated_at timestamptz NOT NULL
		);
		CREATE FUNCTION %s.wait_for_settings_writers() RETURNS trigger LANGUAGE plpgsql AS $$
		BEGIN PERFORM pg_advisory_xact_lock(1262); RETURN NEW; END $$;
		CREATE TRIGGER wait_for_settings_writers BEFORE INSERT ON %s.user_settings
		FOR EACH ROW EXECUTE FUNCTION %s.wait_for_settings_writers();`, schema, schema, schema, schema, schema, schema, schema))
	require.NoError(t, err)

	userID := uuid.New()
	_, err = admin.ExecContext(ctx, fmt.Sprintf("INSERT INTO %s.users (id, updated_at) VALUES ($1, NOW())", schema), userID)
	require.NoError(t, err)
	if existingRow {
		_, err = admin.ExecContext(ctx, fmt.Sprintf("INSERT INTO %s.user_settings (id, user_id, created_at, updated_at) VALUES ($1, $2, NOW(), NOW())", schema), uuid.New(), userID)
		require.NoError(t, err)
	}

	blocker, err := sql.Open("postgres", dsn)
	require.NoError(t, err)
	defer blocker.Close()
	require.NoError(t, setSearchPath(ctx, blocker, schema))
	_, err = blocker.ExecContext(ctx, "SELECT pg_advisory_lock(1262)")
	require.NoError(t, err)
	defer blocker.ExecContext(context.Background(), "SELECT pg_advisory_unlock(1262)")

	dbA := newSchemaDB(t, dsn, schema)
	dbB := newSchemaDB(t, dsn, schema)
	target := 30
	notifications := false
	errs := make(chan error, 2)
	go func() {
		_, err := NewPostgreSQLRepository(dbA).UpdateSettings(ctx, userID, SettingsUpdate{DailyReviewTarget: &target}, time.Now().UTC())
		errs <- err
	}()
	go func() {
		_, err := NewPostgreSQLRepository(dbB).UpdateSettings(ctx, userID, SettingsUpdate{NotificationsEnabled: &notifications}, time.Now().UTC())
		errs <- err
	}()

	require.Eventually(t, func() bool {
		var waiting int
		err := admin.QueryRowContext(ctx, "SELECT count(*) FROM pg_locks WHERE locktype = 'advisory' AND NOT granted").Scan(&waiting)
		return err == nil && waiting >= 2
	}, 5*time.Second, 10*time.Millisecond, "both PATCHes must be paused after their write statement begins")
	_, err = blocker.ExecContext(ctx, "SELECT pg_advisory_unlock(1262)")
	require.NoError(t, err)

	require.NoError(t, <-errs)
	require.NoError(t, <-errs)
	var gotTarget int
	var gotNotifications bool
	err = admin.QueryRowContext(ctx, fmt.Sprintf("SELECT daily_review_target, notifications_enabled FROM %s.user_settings WHERE user_id = $1", schema), userID).Scan(&gotTarget, &gotNotifications)
	require.NoError(t, err)
	require.Equal(t, target, gotTarget)
	require.Equal(t, notifications, gotNotifications)
}

func newSchemaDB(t *testing.T, dsn, schema string) *sql.DB {
	t.Helper()
	db, err := sql.Open("postgres", dsn)
	require.NoError(t, err)
	db.SetMaxOpenConns(1)
	require.NoError(t, setSearchPath(t.Context(), db, schema))
	t.Cleanup(func() { _ = db.Close() })
	return db
}

func setSearchPath(ctx context.Context, db *sql.DB, schema string) error {
	_, err := db.ExecContext(ctx, "SET search_path TO "+schema)
	return err
}
