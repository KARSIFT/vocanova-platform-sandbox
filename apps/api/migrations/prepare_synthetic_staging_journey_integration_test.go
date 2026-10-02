//go:build integration

package migrations_test

import (
	"context"
	"database/sql"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/KARSIFT/vocanova-platform/apps/api/business/gamification"
	"github.com/KARSIFT/vocanova-platform/apps/api/business/missions"
	"github.com/KARSIFT/vocanova-platform/apps/api/business/reviews"
	"github.com/KARSIFT/vocanova-platform/apps/api/foundation/clock"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/require"
)

const journeyEmail = "journey-proof@synthetic.vocanova.invalid"

func prepareJourneyCommand(t *testing.T, container, environment, email string) *exec.Cmd {
	t.Helper()
	contents, err := os.ReadFile(filepath.Join("..", "scripts", "prepare-synthetic-staging-journey.sql"))
	require.NoError(t, err)
	// Fixture setup uses Unix-socket local trust; readiness separately verifies TCP/SCRAM.
	cmd := exec.Command("docker", "exec", "-i", container, "psql", "-X", "--set=ON_ERROR_STOP=1",
		"--username", "vocanova", "--dbname", "vocanova", "--set=journey_environment="+environment,
		"--set=synthetic_email="+email, "--file", "-")
	cmd.Stdin = strings.NewReader(string(contents))
	return cmd
}

func prepareJourney(t *testing.T, container string) uuid.UUID {
	t.Helper()
	out, err := prepareJourneyCommand(t, container, "staging", journeyEmail).CombinedOutput()
	require.NoError(t, err, "preparation: %s", out)
	require.Contains(t, string(out), "prepared")
	return uuid.MustParse(querySingleValue(t, container, "SELECT id FROM users WHERE is_synthetic_test_account AND deleted_at IS NULL"))
}

func journeyDB(t *testing.T, container string) *sql.DB {
	t.Helper()
	out, err := exec.Command("docker", "port", container, "5432/tcp").Output()
	require.NoError(t, err)
	host, port, err := net.SplitHostPort(strings.TrimSpace(string(out)))
	require.NoError(t, err)
	require.Equal(t, "127.0.0.1", host)
	db, err := sql.Open("postgres", fmt.Sprintf("host=127.0.0.1 port=%s user=vocanova password=vocanova dbname=vocanova sslmode=disable", port))
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })
	require.NoError(t, db.PingContext(t.Context()))
	return db
}

// This records real service/repository reviews, not fabricated review counters.
// Only canonical content and initially saved meanings are arranged by SQL.
func submitJourneyReviews(t *testing.T, db *sql.DB, userID uuid.UUID, count int, now time.Time) {
	t.Helper()
	ctx, cancel := context.WithTimeout(t.Context(), time.Minute)
	defer cancel()
	gam := gamification.NewService(gamification.NewRepository(db))
	mission := missions.NewService(missions.NewRepository(db), gam)
	c := &clock.Fixed{T: now}
	repo := reviews.NewPostgreSQLRepository(db, c, reviews.WithGamificationService(gam), reviews.WithMissionsService(mission))
	service := reviews.NewService(repo, nil, c)
	for i := 0; i < count; i++ {
		wordID, meaningID, savedID := uuid.New(), uuid.New(), uuid.New()
		_, err := db.ExecContext(ctx, `INSERT INTO canonical_words (id,text,normalized_text,status,created_at,updated_at) VALUES ($1,$2,$2,'active',$3,$3)`, wordID, "fixture-"+wordID.String(), now)
		require.NoError(t, err)
		_, err = db.ExecContext(ctx, `INSERT INTO word_meanings (id,word_id,part_of_speech,short_definition,meaning_order,status,created_at,updated_at) VALUES ($1,$2,'noun','A synthetic fixture.',1,'active',$3,$3)`, meaningID, wordID, now)
		require.NoError(t, err)
		_, err = db.ExecContext(ctx, `INSERT INTO user_words (id,user_id,meaning_id,source,added_at,created_at,updated_at) VALUES ($1,$2,$3,'journey',$4,$4,$4)`, savedID, userID, meaningID, now)
		require.NoError(t, err)
		due, err := service.ListDueWords(ctx, reviews.ListDueWordsRequest{UserID: userID, Limit: 50})
		require.NoError(t, err)
		require.NotEmpty(t, due.Items)
		req := reviews.SubmitReviewRequest{UserID: userID, UserWordID: savedID, MeaningID: meaningID,
			PromptType: reviews.PromptTypeSelfCheck, Result: reviews.ResultCorrect, Rating: reviews.RatingGood,
			AnsweredAt: now, ClientAttemptID: uuid.NewString(), IdempotencyKey: uuid.NewString()}
		attempt, err := service.SubmitReview(ctx, req)
		require.NoError(t, err)
		require.Equal(t, userID, attempt.UserID)
		require.Equal(t, now.Add(time.Hour), attempt.NextReviewAt)
	}
}

func journeyHistory(t *testing.T, db *sql.DB, userID uuid.UUID) map[string]string {
	t.Helper()
	result := make(map[string]string)
	// Every column is compared, including ownership, source and timestamps.
	for _, table := range []string{"user_words", "review_attempts", "daily_mission_snapshots", "daily_activity_summaries", "confidence_point_ledger", "streak_states", "grace_day_ledger", "user_settings", "idempotency_keys"} {
		var rows string
		err := db.QueryRowContext(t.Context(), "SELECT COALESCE(jsonb_agg(to_jsonb(t) ORDER BY t.id)::text,'[]') FROM "+table+" t WHERE user_id=$1", userID).Scan(&rows)
		require.NoError(t, err, table)
		result[table] = rows
	}
	return result
}

func TestSyntheticStagingPreparationPreservesHistoryAndAllowsSameDayReview(t *testing.T) {
	container := startMigratedDisposablePostgres(t)
	db := journeyDB(t, container)
	oldID := prepareJourney(t, container) // absent reserved account is safe to create
	now := time.Now().UTC().Truncate(time.Microsecond)
	_, err := db.Exec(`INSERT INTO user_settings (id,user_id,timezone,daily_review_target,created_at,updated_at) VALUES ($1,$2,'UTC',5,$3,$3)`, uuid.New(), oldID, now)
	require.NoError(t, err)
	submitJourneyReviews(t, db, oldID, 5, now)
	require.Equal(t, "5|5|completed", querySingleValue(t, container, fmt.Sprintf("SELECT review_target||'|'||reviews_completed||'|'||status FROM daily_mission_snapshots WHERE user_id=%s", sqlLiteral(oldID.String()))))
	require.Equal(t, "35", querySingleValue(t, container, fmt.Sprintf("SELECT sum(amount) FROM confidence_point_ledger WHERE user_id=%s", sqlLiteral(oldID.String()))))
	require.Equal(t, "5", querySingleValue(t, container, fmt.Sprintf("SELECT reviews_attempted FROM daily_activity_summaries WHERE user_id=%s", sqlLiteral(oldID.String()))))
	oldHistory := journeyHistory(t, db, oldID)
	_, err = db.Exec(`INSERT INTO sessions (id,user_id,token_hash,created_at,expires_at) VALUES ($1,$2,decode(repeat('ab',32),'hex'),now()-interval '1 hour',now()+interval '1 day')`, uuid.New(), oldID)
	require.NoError(t, err)

	newID := prepareJourney(t, container)
	require.NotEqual(t, oldID, newID)
	require.Equal(t, oldHistory, journeyHistory(t, db, oldID))
	require.Equal(t, "deleted|true|true", querySingleValue(t, container, fmt.Sprintf("SELECT status||'|'||is_synthetic_test_account||'|'||(deleted_at IS NOT NULL) FROM users WHERE id=%s", sqlLiteral(oldID.String()))))
	require.Equal(t, "1", querySingleValue(t, container, fmt.Sprintf("SELECT count(*) FROM sessions WHERE user_id=%s AND revoked_at IS NOT NULL", sqlLiteral(oldID.String()))))
	for table, rows := range journeyHistory(t, db, newID) {
		require.Equal(t, "[]", rows, table)
	}
	assertSeededAccountShape(t, container, journeyEmail)

	gam := gamification.NewService(gamification.NewRepository(db))
	mission := missions.NewService(missions.NewRepository(db), gam)
	before, err := mission.GetDailyMissionView(t.Context(), newID, "UTC", now)
	require.NoError(t, err)
	require.Equal(t, 0, before.ReviewsCompleted)
	require.Positive(t, before.ReviewTarget)
	submitJourneyReviews(t, db, newID, 1, now)
	after, err := mission.GetDailyMissionView(t.Context(), newID, "UTC", now)
	require.NoError(t, err)
	require.Equal(t, before.LocalDate, after.LocalDate)
	require.Equal(t, 1, after.ReviewsCompleted)
	require.Equal(t, before.ReviewTarget, after.ReviewTarget)
	secondHistory := journeyHistory(t, db, newID)
	thirdID := prepareJourney(t, container)
	require.NotEqual(t, newID, thirdID)
	require.Equal(t, secondHistory, journeyHistory(t, db, newID))
	require.Equal(t, oldHistory, journeyHistory(t, db, oldID))
	require.Equal(t, "3|1", querySingleValue(t, container, "SELECT count(*)||'|'||count(*) FILTER (WHERE deleted_at IS NULL) FROM users WHERE is_synthetic_test_account"))

	// Ordinary seeding remains stable even with multiple historical markers.
	runSyntheticSeed(t, container, journeyEmail, "ordinary seed after preparations")
	runSyntheticSeed(t, container, journeyEmail, "ordinary seed repeat")
	require.Equal(t, thirdID.String(), querySingleValue(t, container, "SELECT id FROM users WHERE is_synthetic_test_account AND deleted_at IS NULL"))
	require.Equal(t, oldHistory, journeyHistory(t, db, oldID))
	require.Equal(t, secondHistory, journeyHistory(t, db, newID))
}

func TestSyntheticStagingPreparationRefusesUnsafeChangesAndRollsBack(t *testing.T) {
	container := startMigratedDisposablePostgres(t)
	db := journeyDB(t, container)
	userID := prepareJourney(t, container)
	_, err := db.Exec(`INSERT INTO sessions (id,user_id,token_hash,created_at,expires_at) VALUES ($1,$2,decode(repeat('cd',32),'hex'),now()-interval '1 hour',now()+interval '1 day')`, uuid.New(), userID)
	require.NoError(t, err)
	snapshot := func() string {
		return querySingleValue(t, container, `SELECT jsonb_build_object('users',(SELECT jsonb_agg(to_jsonb(u) ORDER BY id) FROM users u),'sessions',(SELECT jsonb_agg(to_jsonb(s) ORDER BY id) FROM sessions s))::text`)
	}
	for _, tc := range []struct{ name, env, email, message string }{
		{"blank core environment", "", journeyEmail, "requires staging"},
		{"production", "production", journeyEmail, "requires staging"},
		{"invalid core environment", "development", journeyEmail, "requires staging"},
		{"noncanonical core environment", "Staging", journeyEmail, "requires staging"},
		{"deliverable address", "staging", "journey@example.com", "canonical lowercase .invalid"},
		{"noncanonical address", "staging", " JOURNEY@synthetic.vocanova.invalid ", "canonical lowercase .invalid"},
		{"SQL-shaped address", "staging", "x'; SELECT 1;--@synthetic.vocanova.invalid", "canonical lowercase .invalid"},
		{"different marked identity", "staging", "other@synthetic.vocanova.invalid", "different active synthetic identity"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			before := snapshot()
			out, err := prepareJourneyCommand(t, container, tc.env, tc.email).CombinedOutput()
			require.Error(t, err)
			require.Contains(t, string(out), tc.message)
			require.Equal(t, before, snapshot())
		})
	}
	for _, change := range []string{"is_synthetic_test_account=false", "status='disabled'"} {
		execSQL(t, container, "UPDATE users SET "+change+" WHERE id="+sqlLiteral(userID.String()), "arrange refused identity")
		before := snapshot()
		out, err := prepareJourneyCommand(t, container, "staging", journeyEmail).CombinedOutput()
		require.Error(t, err)
		require.Contains(t, string(out), "unmarked or inactive reserved identity")
		require.Equal(t, before, snapshot())
		execSQL(t, container, "UPDATE users SET is_synthetic_test_account=true,status='active' WHERE id="+sqlLiteral(userID.String()), "restore guarded identity")
	}
	execSQL(t, container, `CREATE FUNCTION reject_journey_replacement() RETURNS trigger LANGUAGE plpgsql AS $$ BEGIN RAISE EXCEPTION 'forced replacement insert failure'; END $$;
 CREATE TRIGGER reject_journey_replacement BEFORE INSERT ON users FOR EACH ROW EXECUTE FUNCTION reject_journey_replacement()`, "force insert failure")
	before := snapshot()
	out, err := prepareJourneyCommand(t, container, "staging", journeyEmail).CombinedOutput()
	require.Error(t, err)
	require.Contains(t, string(out), "forced replacement insert failure")
	require.Equal(t, before, snapshot(), "retirement and session revocation roll back with the failed replacement")
	execSQL(t, container, "DROP TRIGGER reject_journey_replacement ON users; DROP FUNCTION reject_journey_replacement()", "remove failure injection")

	_, err = db.Exec(`INSERT INTO users(id,email,is_synthetic_test_account,status,onboarding_status,created_at,updated_at) VALUES ($1,'second@synthetic.vocanova.invalid',true,'active','completed',now(),now())`, uuid.New())
	require.Error(t, err)
	require.Contains(t, err.Error(), "users_single_synthetic_test_account_idx")
	require.Equal(t, before, snapshot(), "active-only singleton still rejects a second active marked identity")
}

func TestSyntheticStagingWrapperScopeGuards(t *testing.T) {
	script, err := filepath.Abs(filepath.Join("..", "scripts", "prepare-synthetic-staging-journey.sh"))
	require.NoError(t, err)
	for _, tc := range []struct {
		name    string
		env     []string
		allowed bool
	}{
		{"missing core environment", nil, false},
		{"Sentry alone cannot authorize staging", []string{"SENTRY_ENVIRONMENT=staging"}, false},
		{"blank core environment", []string{"ENVIRONMENT=", "SENTRY_ENVIRONMENT=staging"}, false},
		{"production despite staging Sentry", []string{"ENVIRONMENT=production", "SENTRY_ENVIRONMENT=staging"}, false},
		{"invalid core environment", []string{"ENVIRONMENT=development", "SENTRY_ENVIRONMENT=staging"}, false},
		{"noncanonical core environment", []string{"ENVIRONMENT=Staging", "SENTRY_ENVIRONMENT=staging"}, false},
		{"invalid address", []string{"ENVIRONMENT=staging", "SENTRY_ENVIRONMENT=staging", "VOCANOVA_SYNTHETIC_SMOKE_TEST_EMAIL=real@example.com"}, false},
		{"compose command", []string{"ENVIRONMENT=staging", "SENTRY_ENVIRONMENT=staging", "DOCKER_COMPOSE_CMD=docker compose -p production"}, false},
		{"compose file", []string{"ENVIRONMENT=staging", "SENTRY_ENVIRONMENT=staging", "COMPOSE_FILE=production.yml"}, false},
		{"compose project", []string{"ENVIRONMENT=staging", "SENTRY_ENVIRONMENT=staging", "COMPOSE_PROJECT_NAME=production"}, false},
		{"docker host", []string{"ENVIRONMENT=staging", "SENTRY_ENVIRONMENT=staging", "DOCKER_HOST=tcp://example.invalid:2375"}, false},
		{"docker context", []string{"ENVIRONMENT=staging", "SENTRY_ENVIRONMENT=staging", "DOCKER_CONTEXT=production"}, false},
		{"core staging without Sentry", []string{"ENVIRONMENT=staging"}, true},
		{"core staging with blank Sentry", []string{"ENVIRONMENT=staging", "SENTRY_ENVIRONMENT="}, true},
		{"core staging with unrelated Sentry label", []string{"ENVIRONMENT=staging", "SENTRY_ENVIRONMENT=production"}, true},
		{"core staging with staging Sentry", []string{"ENVIRONMENT=staging", "SENTRY_ENVIRONMENT=staging"}, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			record := filepath.Join(dir, "docker-args")
			require.NoError(t, os.WriteFile(filepath.Join(dir, "docker"), []byte("#!/bin/sh\nprintf '%s\\n' \"$@\" > \"$JOURNEY_DOCKER_RECORD\"\ncat > /dev/null\n"), 0700))
			cmd := exec.Command("sh", script)
			cmd.Env = append([]string{"PATH=" + dir + ":/usr/bin:/bin", "JOURNEY_DOCKER_RECORD=" + record}, tc.env...)
			out, err := cmd.CombinedOutput()
			if !tc.allowed {
				require.Error(t, err, "%s", out)
				_, statErr := os.Stat(record)
				require.True(t, os.IsNotExist(statErr), "must reject before invoking Docker")
				return
			}
			require.NoError(t, err, "%s", out)
			args, err := os.ReadFile(record)
			require.NoError(t, err)
			require.Contains(t, string(args), "--context\ndefault\ncompose\n--project-name\nvocanova-staging\n--file\n")
			require.Contains(t, string(args), "/infra/docker-compose.yml\nexec\n-T\npostgres\npsql\n-X\n")
			require.Contains(t, string(args), "--set=synthetic_email=smoke-test-bot@synthetic.vocanova.invalid")
			require.Contains(t, string(args), "--set=journey_environment=staging\n")
			require.NotContains(t, string(args), "--set=api_environment=")
		})
	}
}
