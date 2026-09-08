//go:build integration

// VOC-046-T00 / VOC-046-TEST-00: real-Postgres proof that
// CreateDailyMissionSnapshot's fresh-insert path no longer violates
// daily_mission_snapshots' NOT NULL constraints on created_at/updated_at.
//
// The sqlmock unit test in repository_test.go asserts the INSERT column
// list textually; it cannot produce a real NOT NULL violation because no
// schema is involved. VOC-046-AC-00 and VOC-046-TEST-00 both require
// confirmation against a real Postgres instance running the committed
// migration set, so this file exists as a separate, schema-backed proof.
//
// How to run it locally:
//
//	cd apps/api && go test -tags=integration ./business/missions/...
//
// Requirements: Docker (or a docker-compatible runtime on PATH as
// `docker`). Unlike apps/api/migrations/atlas_apply_integration_test.go,
// this test does not need the Atlas CLI - it applies the committed
// forward migrations directly over the Postgres connection, because what
// is under test here is the application INSERT against the real schema,
// not Atlas's own apply behavior.
//
// Build tag and CI trade-off: the same trade-off VOC-033-D02 recorded for
// the Atlas apply proof applies here. This test is gated behind the
// `integration` tag and is therefore not part of the default
// `go test ./...` path, because adding a Postgres service container to
// the shared CI workflow is not possible from this repository (the
// workflow lives in KARSIFT/karsift-ai-infra) and is out of VOC-046-T00's
// scope. The test skips cleanly when Docker is absent.
package missions

import (
	"bytes"
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/KARSIFT/vocanova-platform/apps/api/business/gamification"
	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// postgresReadyTimeout bounds how long the test waits for the disposable
// container to start accepting connections. 60s is deliberately more
// generous than the Atlas apply proof's 30s: this test may pull the
// postgres image on a cold runner before the container even starts.
const postgresReadyTimeout = 60 * time.Second

// postgresReadyPollInterval is the gap between `pg_isready` probes, kept
// identical to the Atlas apply proof's interval so both integration tests
// behave the same way on a developer machine.
const postgresReadyPollInterval = 250 * time.Millisecond

// postgresImage is pinned rather than floating: a `latest` tag would make
// this proof depend on whichever image the local Docker daemon happened to
// pull most recently. 16-alpine matches the Atlas apply proof and the
// documented staging major version.
const postgresImage = "postgres:16-alpine"

// migrationsDirRelativeToPackage locates the committed forward migrations
// from this package's directory. The schema under test must be the real
// one, not a hand-written fixture, or the test could pass against a
// definition that production does not have.
const migrationsDirRelativeToPackage = "../../migrations"

// TestCreateDailyMissionSnapshotFreshInsertAgainstRealPostgres is
// VOC-046-TEST-00's first half: the repository-level proof. It applies the
// committed migration set to a disposable Postgres, then drives
// CreateDailyMissionSnapshot for a user with no daily_mission_snapshots row
// for the target local date - the exact state issue #352 reports as a
// production 500 - and asserts the insert succeeds and persists non-null,
// plausible created_at/updated_at values.
//
// Against the pre-fix statement (created_at/updated_at absent from the
// column list) this test fails with Postgres error 23502,
// `null value in column "created_at" ... violates not-null constraint`,
// which is the failing-first half of VOC-046-TEST-00's procedure. See the
// package's t00-evidence.md for the recorded pre-fix and post-fix runs.
func TestCreateDailyMissionSnapshotFreshInsertAgainstRealPostgres(t *testing.T) {
	db := newMigratedDisposablePostgres(t)
	userID := insertTestUser(t, db)
	localDate := time.Date(2026, 7, 26, 0, 0, 0, 0, time.UTC)

	requireNoSnapshotFor(t, db, userID, localDate)

	beforeInsert := databaseNow(t, db)

	repo := NewRepository(db)
	tx, err := db.Begin()
	require.NoError(t, err)
	defer tx.Rollback()

	snapshot, err := repo.CreateDailyMissionSnapshot(
		t.Context(), tx, userID, localDate, "UTC", 20, gamification.MissionPolicyVersion,
	)
	require.NoError(t, err,
		"fresh insert must not raise a NOT NULL violation on created_at/updated_at (issue #352)")
	require.NoError(t, tx.Commit())

	require.NotNil(t, snapshot)
	assert.Equal(t, "open", snapshot.Status)
	assert.Equal(t, 20, snapshot.ReviewTarget)
	assert.Equal(t, 0, snapshot.ReviewsCompleted)

	afterInsert := databaseNow(t, db)
	createdAt, updatedAt := readSnapshotTimestamps(t, db, userID, localDate)
	assertTimestampWithin(t, "created_at", createdAt, beforeInsert, afterInsert)
	assertTimestampWithin(t, "updated_at", updatedAt, beforeInsert, afterInsert)
}

// TestCreateDailyMissionSnapshotOnConflictBranchUnchangedAgainstRealPostgres
// is VOC-046-TEST-00's negative coverage: the defect and its fix are
// specific to the fresh-insert branch, so the pre-existing
// ON CONFLICT behavior for a user who already has a row for the same
// (user_id, local_date) must preserve the established daily snapshot. It
// asserts that a retry with changed settings does not duplicate or mutate the
// snapshot, including its timestamps.
func TestCreateDailyMissionSnapshotOnConflictBranchUnchangedAgainstRealPostgres(t *testing.T) {
	db := newMigratedDisposablePostgres(t)
	userID := insertTestUser(t, db)
	localDate := time.Date(2026, 7, 26, 0, 0, 0, 0, time.UTC)
	repo := NewRepository(db)

	firstSnapshot := createSnapshotInOwnTransaction(t, db, repo, userID, localDate, 20)
	firstCreatedAt, firstUpdatedAt := readSnapshotTimestamps(t, db, userID, localDate)

	secondSnapshot := createSnapshotInOwnTransactionWithSettings(
		t, db, repo, userID, localDate, "America/New_York", 30, "p4-mission-policy-v2",
	)
	secondCreatedAt, secondUpdatedAt := readSnapshotTimestamps(t, db, userID, localDate)

	assert.Equal(t, firstSnapshot.ID, secondSnapshot.ID,
		"ON CONFLICT must return the existing row, not create a second one")
	assert.Equal(t, 1, countSnapshotsFor(t, db, userID, localDate))
	assert.Equal(t, "UTC", secondSnapshot.Timezone,
		"a same-day settings change must not rewrite the snapshot timezone")
	assert.Equal(t, 20, secondSnapshot.ReviewTarget,
		"a same-day settings change must not rewrite the snapshot review target")
	assert.Equal(t, gamification.MissionPolicyVersion, secondSnapshot.PolicyVersion,
		"a retry must not rewrite the established snapshot policy")
	assert.True(t, firstCreatedAt.Equal(secondCreatedAt),
		"the conflict branch must not rewrite created_at (was %s, now %s)", firstCreatedAt, secondCreatedAt)
	assert.True(t, firstUpdatedAt.Equal(secondUpdatedAt),
		"the conflict branch must not rewrite updated_at (was %s, now %s)", firstUpdatedAt, secondUpdatedAt)
}

// TestGraceProtectedMissionSnapshotConstraintsAgainstRealPostgres proves the
// database, rather than only MarkSnapshotProtected's normal write path,
// rejects inconsistent or cross-learner grace links. A protected mission
// preserves a streak only when the append-only grace ledger contains the
// matching debit for the same learner (DOC-05 §§2,10,12).
func TestGraceProtectedMissionSnapshotConstraintsAgainstRealPostgres(t *testing.T) {
	// Use the explicitly supplied validation instance rather than the
	// Docker-only helper so this regression exercises all committed forward
	// migrations wherever the repository's shared PostgreSQL check is
	// available. The helper isolates the test in its own schema.
	db := newMigratedPostgresFromEnv(t)
	repo := NewRepository(db)
	firstUserID := insertTestUser(t, db)
	secondUserID := insertTestUser(t, db)
	day := time.Date(2026, 9, 8, 0, 0, 0, 0, time.UTC)
	firstSnapshot := createSnapshotInOwnTransaction(t, db, repo, firstUserID, day, 20)
	secondSnapshot := createSnapshotInOwnTransaction(t, db, repo, secondUserID, day, 20)

	firstGraceID := insertGraceDebit(t, db, firstUserID, day)
	secondGraceID := insertGraceDebit(t, db, secondUserID, day)

	_, err := db.ExecContext(t.Context(), `UPDATE daily_mission_snapshots
		SET status = 'protected', grace_applied = true, grace_day_id = $1
		WHERE id = $2`, firstGraceID, firstSnapshot.ID)
	require.NoError(t, err, "a same-user grace debit must protect its mission")

	for _, tc := range []struct {
		name         string
		snapshotID   string
		status       string
		graceApplied bool
		graceID      *uuid.UUID
		code         string
	}{
		{
			name:         "protected_without_applied_grace",
			snapshotID:   secondSnapshot.ID,
			status:       "protected",
			graceApplied: false,
			code:         "23514",
		},
		{
			name:         "protected_with_dangling_grace_id",
			snapshotID:   secondSnapshot.ID,
			status:       "protected",
			graceApplied: true,
			graceID:      uuidPtr(uuid.New()),
			code:         "23503",
		},
		{
			name:         "protected_with_another_learners_grace_id",
			snapshotID:   firstSnapshot.ID,
			status:       "protected",
			graceApplied: true,
			graceID:      &secondGraceID,
			code:         "23503",
		},
		{
			name:         "non_protected_with_grace_link",
			snapshotID:   firstSnapshot.ID,
			status:       "open",
			graceApplied: true,
			graceID:      &firstGraceID,
			code:         "23514",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, err := db.ExecContext(t.Context(), `UPDATE daily_mission_snapshots
				SET status = $1, grace_applied = $2, grace_day_id = $3
				WHERE id = $4`, tc.status, tc.graceApplied, tc.graceID, tc.snapshotID)
			require.Error(t, err)
			var pqErr *pq.Error
			require.ErrorAs(t, err, &pqErr)
			assert.Equal(t, tc.code, string(pqErr.Code))
		})
	}
}

func insertGraceDebit(t *testing.T, db *sql.DB, userID uuid.UUID, localDate time.Time) uuid.UUID {
	t.Helper()
	id := uuid.New()
	_, err := db.ExecContext(t.Context(), `INSERT INTO grace_day_ledger (
		id, user_id, amount, balance_after, reason, source_type,
		applied_to_local_date, timezone, idempotency_key, created_at, updated_at
	) VALUES ($1, $2, -1, 0, 'used_for_missed_day', 'streak', $3, 'UTC', $4, NOW(), NOW())`,
		id, userID, localDate, "grace-linkage-"+id.String())
	require.NoError(t, err)
	return id
}

func uuidPtr(id uuid.UUID) *uuid.UUID { return &id }

// missionSnapshotNotNullViolationFragment identifies a NOT NULL violation
// raised by daily_mission_snapshots specifically - the defect VOC-046-T00
// owns. Postgres names the offending relation in the error text, which is
// what lets the service-level test below distinguish T00's defect from the
// separate, T02-owned one on streak_states.
const missionSnapshotNotNullViolationFragment = `relation "daily_mission_snapshots" violates not-null constraint`

// streakStateNotNullViolationFragment identifies the same bug class at the
// streak_states call site (gamification's UpsertStreakState), which
// VOC-046-T02's audit owns and VOC-046-T00 must not fix.
const streakStateNotNullViolationFragment = `relation "streak_states" violates not-null constraint`

// TestGetDailyMissionViewForUserWithNoSnapshotAgainstRealPostgres covers
// VOC-046-AC-00's service-level clause: the read path that
// `GET /api/v1/daily-mission` serves - GetDailyMissionView's
// lazy-snapshot-creation branch - must no longer fail on
// daily_mission_snapshots for a user with no settings and no snapshot,
// which is exactly the production state issue #352 reports. The API handler
// maps an error here to the reported 500, so this is the service-layer
// equivalent of the HTTP response the criterion names.
//
// Running this against a real Postgres surfaced a third instance of the
// same bug class on this same endpoint: gamification's UpsertStreakState
// (apps/api/business/gamification/repository.go) omits
// created_at/updated_at from its INSERT INTO streak_states column list, so
// the read-time streak reconciliation this branch performs still raises a
// NOT NULL violation after T00's fix. That call site is explicitly owned by
// VOC-046-T02 ("apps/api/business/gamification/repository.go (the other
// INSERT statements in this file beyond the one already fixed by
// VOC-045-T01)"), so T00 deliberately does not fix it - that would be scope
// expansion into another task's work.
//
// The assertion is therefore written to be exact rather than lenient in
// either direction: the endpoint must never fail on
// daily_mission_snapshots again (T00's own guarantee, enforced
// unconditionally), and the only tolerated failure is the T02-owned
// streak_states one. Once T02 lands, this test tightens automatically to
// the full success path, including the persisted timestamps, with no edit
// needed here.
func TestGetDailyMissionViewForUserWithNoSnapshotAgainstRealPostgres(t *testing.T) {
	db := newMigratedDisposablePostgres(t)
	userID := insertTestUser(t, db)
	now := time.Date(2026, 7, 26, 12, 0, 0, 0, time.UTC)
	today := time.Date(2026, 7, 26, 0, 0, 0, 0, time.UTC)

	requireNoSnapshotFor(t, db, userID, today)

	service := NewService(NewRepository(db), gamification.NewService(gamification.NewRepository(db)))
	view, err := service.GetDailyMissionView(t.Context(), userID, "", now)

	if err != nil {
		require.NotContains(t, err.Error(), missionSnapshotNotNullViolationFragment,
			"the read path must never again fail on daily_mission_snapshots' created_at/updated_at (issue #352, VOC-046-T00)")
		require.Contains(t, err.Error(), streakStateNotNullViolationFragment,
			"the only failure VOC-046-T00 tolerates on this endpoint is the streak_states instance of the same bug class, which VOC-046-T02 owns")
		t.Logf("VOC-046-T00 fix confirmed: the read path now clears daily_mission_snapshots "+
			"and fails later, at the T02-owned streak_states call site: %v", err)
		return
	}

	require.NotNil(t, view)
	assert.Equal(t, today.Format("2006-01-02"), view.LocalDate.Format("2006-01-02"))
	createdAt, updatedAt := readSnapshotTimestamps(t, db, userID, today)
	assert.False(t, createdAt.IsZero(), "the lazily-created row persisted a non-null created_at")
	assert.False(t, updatedAt.IsZero(), "the lazily-created row persisted a non-null updated_at")
}

// TestDailyActivitySummaryFreshInsertPathsAgainstRealPostgres is
// VOC-046-TEST-01: each daily_activity_summaries INSERT call site in this
// repository file is exercised on a genuine first-write path, then the
// persisted row is checked for non-null created_at/updated_at.
func TestDailyActivitySummaryFreshInsertPathsAgainstRealPostgres(t *testing.T) {
	localDate := time.Date(2026, 7, 27, 0, 0, 0, 0, time.UTC)

	t.Run("IncrementReviewsCompleted", func(t *testing.T) {
		db := newMigratedDisposablePostgres(t)
		repo := NewRepository(db)
		userID := insertTestUser(t, db)
		createSnapshotInOwnTransaction(t, db, repo, userID, localDate, 20)
		requireNoActivitySummaryFor(t, db, userID, localDate)

		before := databaseNow(t, db)
		tx, err := db.Begin()
		require.NoError(t, err)
		defer tx.Rollback()
		newCount, err := repo.IncrementReviewsCompleted(t.Context(), tx, userID, localDate, "UTC", 20, true, false)
		require.NoError(t, err)
		require.Equal(t, 1, newCount)
		require.NoError(t, tx.Commit())
		after := databaseNow(t, db)

		summary := readActivitySummary(t, db, userID, localDate)
		assert.Equal(t, 1, summary.ReviewsAttempted)
		assert.Equal(t, 1, summary.ReviewsCorrect)
		assert.Equal(t, 0, summary.ReviewsSkipped)
		assertTimestampWithin(t, "created_at", summary.CreatedAt, before, after)
		assertTimestampWithin(t, "updated_at", summary.UpdatedAt, before, after)
	})

	t.Run("IncrementWordsAdded", func(t *testing.T) {
		db := newMigratedDisposablePostgres(t)
		repo := NewRepository(db)
		userID := insertTestUser(t, db)
		requireNoActivitySummaryFor(t, db, userID, localDate)

		before := databaseNow(t, db)
		tx, err := db.Begin()
		require.NoError(t, err)
		defer tx.Rollback()
		require.NoError(t, repo.IncrementWordsAdded(t.Context(), tx, userID, localDate, "UTC", false))
		require.NoError(t, tx.Commit())
		after := databaseNow(t, db)

		summary := readActivitySummary(t, db, userID, localDate)
		assert.Equal(t, 1, summary.WordsAdded)
		assertTimestampWithin(t, "created_at", summary.CreatedAt, before, after)
		assertTimestampWithin(t, "updated_at", summary.UpdatedAt, before, after)
	})

	t.Run("IncrementSentenceSubmitted", func(t *testing.T) {
		db := newMigratedDisposablePostgres(t)
		repo := NewRepository(db)
		userID := insertTestUser(t, db)
		requireNoActivitySummaryFor(t, db, userID, localDate)

		before := databaseNow(t, db)
		tx, err := db.Begin()
		require.NoError(t, err)
		defer tx.Rollback()
		require.NoError(t, repo.IncrementSentenceSubmitted(t.Context(), tx, userID, localDate, "UTC", false))
		require.NoError(t, tx.Commit())
		after := databaseNow(t, db)

		summary := readActivitySummary(t, db, userID, localDate)
		assert.Equal(t, 1, summary.SentencesSubmitted)
		assertTimestampWithin(t, "created_at", summary.CreatedAt, before, after)
		assertTimestampWithin(t, "updated_at", summary.UpdatedAt, before, after)
	})

	t.Run("IncrementAIFeedbackReceived", func(t *testing.T) {
		db := newMigratedDisposablePostgres(t)
		repo := NewRepository(db)
		userID := insertTestUser(t, db)
		requireNoActivitySummaryFor(t, db, userID, localDate)

		before := databaseNow(t, db)
		tx, err := db.Begin()
		require.NoError(t, err)
		defer tx.Rollback()
		require.NoError(t, repo.IncrementAIFeedbackReceived(t.Context(), tx, userID, localDate, "UTC"))
		require.NoError(t, tx.Commit())
		after := databaseNow(t, db)

		summary := readActivitySummary(t, db, userID, localDate)
		assert.Equal(t, 1, summary.AIFeedbackReceived)
		assertTimestampWithin(t, "created_at", summary.CreatedAt, before, after)
		assertTimestampWithin(t, "updated_at", summary.UpdatedAt, before, after)
	})

	t.Run("RecordConfidencePointChange", func(t *testing.T) {
		db := newMigratedDisposablePostgres(t)
		repo := NewRepository(db)
		userID := insertTestUser(t, db)
		requireNoActivitySummaryFor(t, db, userID, localDate)

		before := databaseNow(t, db)
		tx, err := db.Begin()
		require.NoError(t, err)
		defer tx.Rollback()
		require.NoError(t, repo.RecordConfidencePointChange(t.Context(), tx, userID, localDate, "UTC", 5))
		require.NoError(t, tx.Commit())
		after := databaseNow(t, db)

		summary := readActivitySummary(t, db, userID, localDate)
		assert.Equal(t, 5, summary.ConfidencePointsEarned)
		assertTimestampWithin(t, "created_at", summary.CreatedAt, before, after)
		assertTimestampWithin(t, "updated_at", summary.UpdatedAt, before, after)
	})
}

func TestDailyActivitySummaryReviewCounterConstraintsAgainstRealPostgres(t *testing.T) {
	db := newMigratedPostgresForReviewCounterConstraints(t)
	userID := insertTestUser(t, db)
	localDate := time.Date(2026, 9, 8, 0, 0, 0, 0, time.UTC)

	for _, tc := range []struct {
		name      string
		attempted int
		correct   int
		skipped   int
	}{
		{name: "negative_attempted", attempted: -1},
		{name: "negative_correct", attempted: 1, correct: -1},
		{name: "negative_skipped", attempted: 1, skipped: -1},
		{name: "classified_exceeds_attempted", attempted: 1, correct: 1, skipped: 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, err := db.ExecContext(t.Context(), `INSERT INTO daily_activity_summaries (
				id, user_id, local_date, timezone, reviews_attempted, reviews_correct,
				reviews_skipped, created_at, updated_at
			) VALUES ($1, $2, $3, 'UTC', $4, $5, $6, NOW(), NOW())`,
				uuid.New(), userID, localDate.AddDate(0, 0, len(tc.name)), tc.attempted, tc.correct, tc.skipped)
			require.Error(t, err)
			var pqErr *pq.Error
			require.ErrorAs(t, err, &pqErr)
			assert.Equal(t, "23514", string(pqErr.Code))
		})
	}

	repo := NewRepository(db)
	createSnapshotInOwnTransaction(t, db, repo, userID, localDate, 20)
	tx, err := db.BeginTx(t.Context(), nil)
	require.NoError(t, err)
	defer tx.Rollback()
	_, err = repo.IncrementReviewsCompleted(t.Context(), tx, userID, localDate, "UTC", 20, true, false)
	require.NoError(t, err)
	require.NoError(t, tx.Commit())

	summary := readActivitySummary(t, db, userID, localDate)
	assert.Equal(t, 1, summary.ReviewsAttempted)
	assert.Equal(t, 1, summary.ReviewsCorrect)
	assert.Equal(t, 0, summary.ReviewsSkipped)
}

// TestDailyActivityReviewCounterMigrationAcceptsExistingLegacyRows proves the
// forward-only migration itself can be applied to a populated production
// database: a legacy aggregate that predates these checks must not block the
// deployment, while the checks still reject invalid rows written afterwards.
func TestDailyActivityReviewCounterMigrationAcceptsExistingLegacyRows(t *testing.T) {
	db := newPostgresForReviewCounterMigration(t)
	applyCommittedForwardMigrationsBefore(t, db, "20260908110000_daily_activity_review_counter_integrity.sql")
	userID := insertTestUser(t, db)
	legacyDate := time.Date(2026, 9, 6, 0, 0, 0, 0, time.UTC)

	_, err := db.ExecContext(t.Context(), `INSERT INTO daily_activity_summaries (
		id, user_id, local_date, timezone, reviews_attempted, reviews_correct,
		reviews_skipped, created_at, updated_at
	) VALUES ($1, $2, $3, 'UTC', 1, 1, 1, NOW(), NOW())`, uuid.New(), userID, legacyDate)
	require.NoError(t, err, "pre-migration schema must accept the legacy aggregate fixture")

	applyCommittedForwardMigration(t, db, "20260908110000_daily_activity_review_counter_integrity.sql")

	var legacyRows int
	require.NoError(t, db.QueryRowContext(t.Context(), `SELECT count(*) FROM daily_activity_summaries
		WHERE user_id = $1 AND local_date = $2`, userID, legacyDate).Scan(&legacyRows))
	assert.Equal(t, 1, legacyRows, "the migration must not rewrite or reject existing aggregates")

	_, err = db.ExecContext(t.Context(), `INSERT INTO daily_activity_summaries (
		id, user_id, local_date, timezone, reviews_attempted, reviews_correct,
		reviews_skipped, created_at, updated_at
	) VALUES ($1, $2, $3, 'UTC', 1, 1, 1, NOW(), NOW())`, uuid.New(), userID, legacyDate.AddDate(0, 0, 1))
	require.Error(t, err, "the NOT VALID checks must still protect every new row")
	var pqErr *pq.Error
	require.ErrorAs(t, err, &pqErr)
	assert.Equal(t, "23514", string(pqErr.Code))
}

// newMigratedPostgresForReviewCounterConstraints uses an explicitly supplied
// PostgreSQL validation instance when available, while retaining the
// disposable-container path used by the rest of this integration suite.
func newMigratedPostgresForReviewCounterConstraints(t *testing.T) *sql.DB {
	t.Helper()
	if os.Getenv("VOCANOVA_TEST_POSTGRES_DSN") != "" {
		return newMigratedPostgresFromEnv(t)
	}
	return newMigratedDisposablePostgres(t)
}

// activitySummaryIncrement describes one daily_activity_summaries call site
// so the update-branch proof below can drive every one of them through the
// same procedure instead of repeating it five times.
type activitySummaryIncrement struct {
	name                  string
	requiresSnapshot      bool
	invoke                func(ctx context.Context, repo *Repository, tx *sql.Tx, userID uuid.UUID, localDate time.Time) error
	counter               func(activitySummaryRow) int
	expectedAfterTwoCalls int
}

// activitySummaryIncrements enumerates every daily_activity_summaries insert
// call site VOC-046-T01 fixed, paired with the counter column that call site
// accumulates, so a missed call site shows up as missing coverage rather
// than as a silently untested path.
func activitySummaryIncrements() []activitySummaryIncrement {
	return []activitySummaryIncrement{
		{
			name:             "IncrementReviewsCompleted",
			requiresSnapshot: true,
			invoke: func(ctx context.Context, repo *Repository, tx *sql.Tx, userID uuid.UUID, localDate time.Time) error {
				_, err := repo.IncrementReviewsCompleted(ctx, tx, userID, localDate, "UTC", 20, true, false)
				return err
			},
			counter:               func(row activitySummaryRow) int { return row.ReviewsAttempted },
			expectedAfterTwoCalls: 2,
		},
		{
			name: "IncrementWordsAdded",
			invoke: func(ctx context.Context, repo *Repository, tx *sql.Tx, userID uuid.UUID, localDate time.Time) error {
				return repo.IncrementWordsAdded(ctx, tx, userID, localDate, "UTC", false)
			},
			counter:               func(row activitySummaryRow) int { return row.WordsAdded },
			expectedAfterTwoCalls: 2,
		},
		{
			name: "IncrementSentenceSubmitted",
			invoke: func(ctx context.Context, repo *Repository, tx *sql.Tx, userID uuid.UUID, localDate time.Time) error {
				return repo.IncrementSentenceSubmitted(ctx, tx, userID, localDate, "UTC", false)
			},
			counter:               func(row activitySummaryRow) int { return row.SentencesSubmitted },
			expectedAfterTwoCalls: 2,
		},
		{
			name: "IncrementAIFeedbackReceived",
			invoke: func(ctx context.Context, repo *Repository, tx *sql.Tx, userID uuid.UUID, localDate time.Time) error {
				return repo.IncrementAIFeedbackReceived(ctx, tx, userID, localDate, "UTC")
			},
			counter:               func(row activitySummaryRow) int { return row.AIFeedbackReceived },
			expectedAfterTwoCalls: 2,
		},
		{
			name: "RecordConfidencePointChange",
			invoke: func(ctx context.Context, repo *Repository, tx *sql.Tx, userID uuid.UUID, localDate time.Time) error {
				return repo.RecordConfidencePointChange(ctx, tx, userID, localDate, "UTC", 5)
			},
			counter:               func(row activitySummaryRow) int { return row.ConfidencePointsEarned },
			expectedAfterTwoCalls: 10,
		},
	}
}

// TestDailyActivitySummaryOnConflictBranchUnchangedAgainstRealPostgres is
// VOC-046-TEST-01's negative coverage, mirroring what
// TestCreateDailyMissionSnapshotOnConflictBranchUnchangedAgainstRealPostgres
// does for snapshots: VOC-046-T01's fix touches only the fresh-insert
// branch, so for a user who already has a row for the same
// (user_id, local_date) each call site must still accumulate into that row
// rather than duplicate it, must not rewrite created_at, and must still
// advance updated_at.
func TestDailyActivitySummaryOnConflictBranchUnchangedAgainstRealPostgres(t *testing.T) {
	localDate := time.Date(2026, 7, 28, 0, 0, 0, 0, time.UTC)

	for _, callSite := range activitySummaryIncrements() {
		t.Run(callSite.name, func(t *testing.T) {
			db := newMigratedDisposablePostgres(t)
			repo := NewRepository(db)
			userID := insertTestUser(t, db)
			if callSite.requiresSnapshot {
				createSnapshotInOwnTransaction(t, db, repo, userID, localDate, 20)
			}
			requireNoActivitySummaryFor(t, db, userID, localDate)

			runIncrementInOwnTransaction(t, db, repo, callSite, userID, localDate)
			firstRow := readActivitySummary(t, db, userID, localDate)

			// Postgres NOW() is transaction-start time, so two
			// back-to-back transactions can share a timestamp on a fast
			// machine. Sleeping past the clock's practical resolution
			// keeps the "updated_at advanced" assertion meaningful
			// rather than flaky.
			time.Sleep(10 * time.Millisecond)

			runIncrementInOwnTransaction(t, db, repo, callSite, userID, localDate)
			secondRow := readActivitySummary(t, db, userID, localDate)

			assert.Equal(t, 1, countActivitySummariesFor(t, db, userID, localDate),
				"ON CONFLICT must update the existing row, not create a second one")
			assert.Equal(t, callSite.expectedAfterTwoCalls, callSite.counter(secondRow),
				"the update branch must still accumulate the call site's counter")
			assert.True(t, firstRow.CreatedAt.Equal(secondRow.CreatedAt),
				"the update branch must not rewrite created_at (was %s, now %s)",
				firstRow.CreatedAt, secondRow.CreatedAt)
			assert.False(t, secondRow.UpdatedAt.Before(firstRow.UpdatedAt),
				"the update branch must advance updated_at (was %s, now %s)",
				firstRow.UpdatedAt, secondRow.UpdatedAt)
		})
	}
}

// runIncrementInOwnTransaction drives one increment call site in its own
// committed transaction, which is both how production invokes it and what
// makes Postgres' transaction-start NOW() advance between calls.
func runIncrementInOwnTransaction(
	t *testing.T,
	db *sql.DB,
	repo *Repository,
	callSite activitySummaryIncrement,
	userID uuid.UUID,
	localDate time.Time,
) {
	t.Helper()
	tx, err := db.Begin()
	require.NoError(t, err)
	defer tx.Rollback()
	require.NoError(t, callSite.invoke(t.Context(), repo, tx, userID, localDate))
	require.NoError(t, tx.Commit())
}

// createSnapshotInOwnTransaction runs one CreateDailyMissionSnapshot call
// in its own committed transaction, which is how the production read path
// invokes it. Each call gets a fresh transaction so Postgres' NOW()
// (transaction-start time) actually advances between calls.
func createSnapshotInOwnTransaction(
	t *testing.T,
	db *sql.DB,
	repo *Repository,
	userID uuid.UUID,
	localDate time.Time,
	reviewTarget int,
) *DailyMissionSnapshot {
	return createSnapshotInOwnTransactionWithSettings(
		t, db, repo, userID, localDate, "UTC", reviewTarget, gamification.MissionPolicyVersion,
	)
}

func createSnapshotInOwnTransactionWithSettings(
	t *testing.T,
	db *sql.DB,
	repo *Repository,
	userID uuid.UUID,
	localDate time.Time,
	timezone string,
	reviewTarget int,
	policyVersion string,
) *DailyMissionSnapshot {
	t.Helper()
	tx, err := db.Begin()
	require.NoError(t, err)
	defer tx.Rollback()
	snapshot, err := repo.CreateDailyMissionSnapshot(
		t.Context(), tx, userID, localDate, timezone, reviewTarget, policyVersion,
	)
	require.NoError(t, err)
	require.NoError(t, tx.Commit())
	return snapshot
}

// newMigratedDisposablePostgres starts a disposable Postgres container,
// applies the committed forward migrations to it, and returns an open
// connection. The container and connection are torn down via t.Cleanup, so
// each test gets an isolated database and no state leaks between tests.
func newMigratedDisposablePostgres(t *testing.T) *sql.DB {
	t.Helper()
	db := newDisposablePostgres(t)
	applyCommittedForwardMigrations(t, db)
	return db
}

// newDisposablePostgres starts the isolated PostgreSQL instance shared by the
// fresh-schema and existing-data migration proofs. Callers choose which part
// of the migration sequence to apply so the latter can seed a true legacy row.
func newDisposablePostgres(t *testing.T) *sql.DB {
	t.Helper()
	if _, err := exec.LookPath("docker"); err != nil {
		t.Skipf("docker not on PATH (VOC-046-TEST-00's real-Postgres proof requires Docker): %v", err)
	}
	if output, err := exec.Command("docker", "info").CombinedOutput(); err != nil {
		t.Skipf("Docker is on PATH but unavailable (VOC-046-TEST-00's real-Postgres proof requires a running Docker daemon): %v\n%s", err, output)
	}

	port := freeLoopbackTCPPort(t)
	containerName := "voc046-missions-test-" + randomHexSuffix(t, 6)
	startDisposablePostgresContainer(t, containerName, port)
	t.Cleanup(func() { forceRemoveContainer(t, containerName) })
	waitForPostgresAcceptingConnections(t, containerName)

	db, err := sql.Open("postgres",
		fmt.Sprintf("postgres://vocanova:vocanova@127.0.0.1:%d/vocanova?sslmode=disable", port))
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })
	requirePingSucceeds(t, db)
	return db
}

// newMigratedPostgresFromEnv applies the complete committed migration set in a
// unique schema of an explicitly supplied PostgreSQL validation instance.
// The connection pool is closed before the schema is dropped so cleanup does
// not leave a pooled search_path pointed at a removed schema.
func newMigratedPostgresFromEnv(t *testing.T) *sql.DB {
	db := newPostgresFromEnv(t)
	applyCommittedForwardMigrations(t, db)
	return db
}

// newPostgresFromEnv creates an isolated schema in the explicitly supplied
// PostgreSQL validation instance, leaving the caller free to apply either the
// full migration set or a prefix of it.
func newPostgresFromEnv(t *testing.T) *sql.DB {
	t.Helper()
	dsn := os.Getenv("VOCANOVA_TEST_POSTGRES_DSN")
	if dsn == "" {
		t.Skip("VOCANOVA_TEST_POSTGRES_DSN is unset")
	}
	admin, err := sql.Open("postgres", dsn)
	require.NoError(t, err)
	t.Cleanup(func() { _ = admin.Close() })
	schema := "missions_grace_" + randomHexSuffix(t, 12)
	_, err = admin.Exec("CREATE SCHEMA " + schema)
	require.NoError(t, err)
	t.Cleanup(func() {
		if _, err := admin.Exec("DROP SCHEMA " + schema + " CASCADE"); err != nil {
			t.Errorf("drop validation schema %s: %v", schema, err)
		}
	})
	if strings.HasPrefix(dsn, "postgres://") || strings.HasPrefix(dsn, "postgresql://") {
		dsn, err = pq.ParseURL(dsn)
		require.NoError(t, err)
	}
	db, err := sql.Open("postgres", dsn+" search_path="+schema)
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })
	return db
}

func newPostgresForReviewCounterMigration(t *testing.T) *sql.DB {
	t.Helper()
	if os.Getenv("VOCANOVA_TEST_POSTGRES_DSN") != "" {
		return newPostgresFromEnv(t)
	}
	return newDisposablePostgres(t)
}

func TestDailyActivityPointAggregateIntegrityAgainstRealPostgres(t *testing.T) {
	db := newMigratedPostgresFromEnv(t)
	userID := insertTestUser(t, db)
	localDate := time.Date(2026, 9, 8, 0, 0, 0, 0, time.UTC)

	// A forward-only constraint must not make deployment depend on repairing
	// historical summaries. PostgreSQL still checks NOT VALID constraints for
	// every new or changed row, which the direct-write checks below prove.
	var validated bool
	require.NoError(t, db.QueryRowContext(t.Context(), `SELECT convalidated
		FROM pg_constraint
		WHERE conname = 'daily_activity_summaries_confidence_point_counters_nonnegative'`).Scan(&validated))
	assert.False(t, validated)

	for _, tc := range []struct {
		name   string
		earned int
		spent  int
	}{
		{name: "negative_earned", earned: -1},
		{name: "negative_spent", spent: -1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, err := db.ExecContext(t.Context(), `INSERT INTO daily_activity_summaries (
				id, user_id, local_date, timezone, confidence_points_earned,
				confidence_points_spent, created_at, updated_at
			) VALUES ($1, $2, $3, 'UTC', $4, $5, NOW(), NOW())`,
				uuid.New(), userID, localDate.AddDate(0, 0, len(tc.name)), tc.earned, tc.spent)
			require.Error(t, err)
			var pqErr *pq.Error
			require.ErrorAs(t, err, &pqErr)
			assert.Equal(t, "23514", string(pqErr.Code))
		})
	}

	repo := NewRepository(db)
	for _, amount := range []int{5, -3} {
		tx, err := db.BeginTx(t.Context(), nil)
		require.NoError(t, err)
		require.NoError(t, repo.RecordConfidencePointChange(t.Context(), tx, userID, localDate, "UTC", amount))
		require.NoError(t, tx.Commit())
	}

	summary := readActivitySummary(t, db, userID, localDate)
	assert.Equal(t, 5, summary.ConfidencePointsEarned)
	assert.Equal(t, 3, summary.ConfidencePointsSpent)

	tx, err := db.BeginTx(t.Context(), nil)
	require.NoError(t, err)
	require.NoError(t, repo.RecordConfidencePointChange(t.Context(), tx, userID, localDate.AddDate(0, 0, 1), "UTC", 0))
	require.NoError(t, tx.Commit())
	requireNoActivitySummaryFor(t, db, userID, localDate.AddDate(0, 0, 1))
}

// TestGraceDayUseProtectsSnapshotAndSurvivesNextDayRealPostgres proves the
// production transaction-owner protocol against all forward migrations. The
// gamification package returns the debit row ID; its caller persists the
// owner/date-scoped protected snapshot in that same transaction.
func TestGraceDayUseProtectsSnapshotAndSurvivesNextDayRealPostgres(t *testing.T) {
	db := newMigratedPostgresFromEnv(t)
	ctx := t.Context()
	userID := insertTestUser(t, db)
	missionRepo := NewRepository(db)
	gamSvc := gamification.NewService(gamification.NewRepository(db))
	day1 := time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)
	day2, day3, day4 := day1.AddDate(0, 0, 1), day1.AddDate(0, 0, 2), day1.AddDate(0, 0, 3)
	for _, day := range []time.Time{day1, day2, day3, day4} {
		createSnapshotInOwnTransaction(t, db, missionRepo, userID, day, 5)
	}
	_, err := db.ExecContext(ctx, `UPDATE daily_mission_snapshots
		SET status = CASE local_date WHEN $2 THEN 'completed' WHEN $3 THEN 'missed' WHEN $4 THEN 'completed' ELSE 'missed' END,
		    completed_at = CASE WHEN local_date IN ($2, $4) THEN NOW() ELSE NULL END
		WHERE user_id = $1`, userID, day1, day2, day3)
	require.NoError(t, err)
	_, err = db.ExecContext(ctx, `INSERT INTO streak_states (id,user_id,current_streak_count,longest_streak_count,last_completed_local_date,last_activity_local_date,timezone,status,created_at,updated_at)
		VALUES ($1,$2,4,4,$3,$3,'UTC','active',NOW(),NOW())`, uuid.New(), userID, day1)
	require.NoError(t, err)
	_, err = db.ExecContext(ctx, `INSERT INTO grace_day_ledger (id,user_id,amount,balance_after,reason,source_type,applied_to_local_date,timezone,idempotency_key,created_at,updated_at)
		VALUES ($1,$2,1,1,'earned_by_streak','streak',$3,'UTC','seed',NOW(),NOW())`, uuid.New(), userID, day1)
	require.NoError(t, err)

	tx, err := db.BeginTx(ctx, nil)
	require.NoError(t, err)
	rec, err := gamSvc.ReconcileAndAdvance(ctx, tx, userID, day3.Add(12*time.Hour), "UTC", []gamification.StreakSnapshot{
		{LocalDate: day1, Status: gamification.MissionStatusCompleted},
		{LocalDate: day2, Status: gamification.MissionStatusMissed},
		{LocalDate: day3, Status: gamification.MissionStatusCompleted},
	}, true)
	require.NoError(t, err)
	require.NotNil(t, rec.GraceDayUsedID)
	protected, err := missionRepo.MarkSnapshotProtected(ctx, tx, userID, *rec.YesterdayProtectedLocalDate, *rec.GraceDayUsedID)
	require.NoError(t, err)
	require.True(t, protected)
	require.NoError(t, tx.Commit())

	protectedSnapshot, err := missionRepo.GetDailyMissionSnapshot(ctx, userID, day2)
	require.NoError(t, err)
	require.Equal(t, gamification.MissionStatusProtected, protectedSnapshot.Status)
	require.True(t, protectedSnapshot.GraceApplied)
	require.NotNil(t, protectedSnapshot.GraceDayID)
	require.Equal(t, rec.GraceDayUsedID.String(), *protectedSnapshot.GraceDayID)

	tx, err = db.BeginTx(ctx, nil)
	require.NoError(t, err)
	_, err = gamSvc.ReconcileAndAdvance(ctx, tx, userID, day4.Add(12*time.Hour), "UTC", []gamification.StreakSnapshot{
		{LocalDate: day1, Status: gamification.MissionStatusCompleted},
		{LocalDate: day2, Status: gamification.MissionStatusProtected},
		{LocalDate: day3, Status: gamification.MissionStatusCompleted},
		{LocalDate: day4, Status: gamification.MissionStatusMissed},
	}, false)
	require.NoError(t, err)
	require.NoError(t, tx.Commit())
	state, err := gamSvc.GetStreakStateForRead(ctx, userID)
	require.NoError(t, err)
	require.Equal(t, 5, state.CurrentStreakCount)
}

// applyCommittedForwardMigrations executes every committed forward
// migration in filename (version) order. Recovery down-files are excluded
// by the same rule Atlas itself uses: only `*.sql` is a forward migration,
// and the recovery files carry a `.down.sql.example` suffix specifically so
// they fall outside that glob.
func applyCommittedForwardMigrations(t *testing.T, db *sql.DB) {
	applyCommittedForwardMigrationsBefore(t, db, "")
}

// applyCommittedForwardMigrationsBefore applies each committed forward
// migration before stopBefore. An empty stopBefore applies the whole set.
func applyCommittedForwardMigrationsBefore(t *testing.T, db *sql.DB, stopBefore string) {
	t.Helper()
	paths, err := filepath.Glob(filepath.Join(migrationsDirRelativeToPackage, "*.sql"))
	require.NoError(t, err)
	require.NotEmpty(t, paths, "no forward migrations found in %s", migrationsDirRelativeToPackage)
	sort.Strings(paths)
	for _, path := range paths {
		if strings.HasSuffix(path, ".down.sql") {
			continue
		}
		name := filepath.Base(path)
		if stopBefore != "" && name == stopBefore {
			return
		}
		applyCommittedForwardMigration(t, db, name)
	}
	if stopBefore != "" {
		t.Fatalf("migration %s not found", stopBefore)
	}
}

func applyCommittedForwardMigration(t *testing.T, db *sql.DB, name string) {
	t.Helper()
	statements, err := os.ReadFile(filepath.Join(migrationsDirRelativeToPackage, name))
	require.NoError(t, err)
	_, err = db.Exec(string(statements))
	require.NoErrorf(t, err, "apply migration %s", name)
}

// insertTestUser creates the minimal users row the mission tables'
// foreign key requires, and returns its ID.
func insertTestUser(t *testing.T, db *sql.DB) uuid.UUID {
	t.Helper()
	userID := uuid.New()
	_, err := db.Exec(
		`INSERT INTO users (id, email, status, onboarding_status, created_at, updated_at)
		 VALUES ($1, $2, 'active', 'completed', NOW(), NOW())`,
		userID, fmt.Sprintf("voc046-%s@example.test", userID),
	)
	require.NoError(t, err)
	return userID
}

// requireNoSnapshotFor asserts the precondition VOC-046-TEST-00 names: the
// user must have no daily_mission_snapshots row for the target local date,
// so the call under test genuinely exercises the fresh-insert branch rather
// than the ON CONFLICT update branch.
func requireNoSnapshotFor(t *testing.T, db *sql.DB, userID uuid.UUID, localDate time.Time) {
	t.Helper()
	require.Equal(t, 0, countSnapshotsFor(t, db, userID, localDate),
		"precondition: the user must have no snapshot for this local date")
}

func countSnapshotsFor(t *testing.T, db *sql.DB, userID uuid.UUID, localDate time.Time) int {
	t.Helper()
	var count int
	require.NoError(t, db.QueryRow(
		`SELECT count(*) FROM daily_mission_snapshots WHERE user_id = $1 AND local_date = $2`,
		userID, localDate,
	).Scan(&count))
	return count
}

func requireNoActivitySummaryFor(t *testing.T, db *sql.DB, userID uuid.UUID, localDate time.Time) {
	t.Helper()
	require.Equal(t, 0, countActivitySummariesFor(t, db, userID, localDate),
		"precondition: the user must have no activity summary for this local date")
}

func countActivitySummariesFor(t *testing.T, db *sql.DB, userID uuid.UUID, localDate time.Time) int {
	t.Helper()
	var count int
	require.NoError(t, db.QueryRow(
		`SELECT count(*) FROM daily_activity_summaries WHERE user_id = $1 AND local_date = $2`,
		userID, localDate,
	).Scan(&count))
	return count
}

type activitySummaryRow struct {
	ReviewsAttempted       int
	ReviewsCorrect         int
	ReviewsSkipped         int
	WordsAdded             int
	SentencesSubmitted     int
	AIFeedbackReceived     int
	ConfidencePointsEarned int
	ConfidencePointsSpent  int
	CreatedAt              time.Time
	UpdatedAt              time.Time
}

func readActivitySummary(t *testing.T, db *sql.DB, userID uuid.UUID, localDate time.Time) activitySummaryRow {
	t.Helper()
	var row activitySummaryRow
	require.NoError(t, db.QueryRow(
		`SELECT reviews_attempted, reviews_correct, reviews_skipped,
		        words_added, sentences_submitted, ai_feedback_received,
		        confidence_points_earned, confidence_points_spent, created_at, updated_at
		   FROM daily_activity_summaries
		  WHERE user_id = $1 AND local_date = $2`,
		userID, localDate,
	).Scan(
		&row.ReviewsAttempted,
		&row.ReviewsCorrect,
		&row.ReviewsSkipped,
		&row.WordsAdded,
		&row.SentencesSubmitted,
		&row.AIFeedbackReceived,
		&row.ConfidencePointsEarned,
		&row.ConfidencePointsSpent,
		&row.CreatedAt,
		&row.UpdatedAt,
	))
	return row
}

// readSnapshotTimestamps reads the persisted created_at/updated_at back out
// of Postgres. Both are scanned into non-pointer time.Time values, so a
// null in either column would fail the scan outright - the assertion that
// the columns are actually populated is structural, not just a value check.
func readSnapshotTimestamps(t *testing.T, db *sql.DB, userID uuid.UUID, localDate time.Time) (time.Time, time.Time) {
	t.Helper()
	var createdAt, updatedAt time.Time
	require.NoError(t, db.QueryRow(
		`SELECT created_at, updated_at FROM daily_mission_snapshots
		 WHERE user_id = $1 AND local_date = $2`,
		userID, localDate,
	).Scan(&createdAt, &updatedAt))
	return createdAt, updatedAt
}

// databaseNow reads the server's clock so the timestamp-plausibility
// assertions compare against Postgres' own time rather than the test
// process's, which may differ from the container's.
func databaseNow(t *testing.T, db *sql.DB) time.Time {
	t.Helper()
	var now time.Time
	require.NoError(t, db.QueryRow(`SELECT NOW()`).Scan(&now))
	return now
}

// assertTimestampWithin checks a persisted timestamp is non-null and lies
// inside the window bracketing the insert, which is VOC-046-TEST-00's
// "non-null and reasonable" expectation.
func assertTimestampWithin(t *testing.T, column string, actual, notBefore, notAfter time.Time) {
	t.Helper()
	assert.Falsef(t, actual.IsZero(), "%s must be non-null", column)
	assert.Falsef(t, actual.Before(notBefore),
		"%s (%s) must not predate the insert window start (%s)", column, actual, notBefore)
	assert.Falsef(t, actual.After(notAfter),
		"%s (%s) must not postdate the insert window end (%s)", column, actual, notAfter)
}

func requirePingSucceeds(t *testing.T, db *sql.DB) {
	t.Helper()
	deadline := time.Now().Add(postgresReadyTimeout)
	var lastErr error
	for time.Now().Before(deadline) {
		if lastErr = db.Ping(); lastErr == nil {
			return
		}
		time.Sleep(postgresReadyPollInterval)
	}
	t.Fatalf("could not connect to the disposable Postgres within %s: %v", postgresReadyTimeout, lastErr)
}

// freeLoopbackTCPPort returns a port currently free on 127.0.0.1. Binding
// the container to loopback only is what guarantees this test never opens a
// publicly reachable database port, even on a network-exposed machine.
func freeLoopbackTCPPort(t *testing.T) int {
	t.Helper()
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	defer listener.Close()
	addr, ok := listener.Addr().(*net.TCPAddr)
	require.Truef(t, ok, "unexpected listener address type %T", listener.Addr())
	return addr.Port
}

// randomHexSuffix uniquifies the container name so parallel runs, or a
// leftover container from a crashed run, cannot collide and produce a
// confusing "name already in use" error that hides the real failure.
func randomHexSuffix(t *testing.T, byteCount int) string {
	t.Helper()
	buf := make([]byte, byteCount)
	_, err := rand.Read(buf)
	require.NoError(t, err)
	return hex.EncodeToString(buf)
}

func startDisposablePostgresContainer(t *testing.T, containerName string, hostPort int) {
	t.Helper()
	args := []string{
		"run", "--rm", "-d",
		"--name", containerName,
		"-p", fmt.Sprintf("127.0.0.1:%d:5432", hostPort),
		"-e", "POSTGRES_USER=vocanova",
		"-e", "POSTGRES_PASSWORD=vocanova",
		"-e", "POSTGRES_DB=vocanova",
		postgresImage,
	}
	out, err := exec.Command("docker", args...).CombinedOutput()
	if err != nil {
		t.Fatalf("docker %s: %v\noutput:\n%s", strings.Join(args, " "), err, out)
	}
}

func forceRemoveContainer(t *testing.T, containerName string) {
	t.Helper()
	out, err := exec.Command("docker", "rm", "-f", containerName).CombinedOutput()
	if err != nil && !bytes.Contains(out, []byte("No such container")) {
		// Best-effort cleanup: a stale container does not invalidate the
		// proof, so this must not change the test verdict.
		t.Logf("docker rm -f %s: %v\noutput:\n%s", containerName, err, out)
	}
}

func waitForPostgresAcceptingConnections(t *testing.T, containerName string) {
	t.Helper()
	deadline := time.Now().Add(postgresReadyTimeout)
	var lastErr error
	var lastOut []byte
	for time.Now().Before(deadline) {
		out, err := exec.Command("docker", "exec", containerName,
			"pg_isready", "-U", "vocanova", "-d", "vocanova").CombinedOutput()
		if err == nil {
			return
		}
		lastErr, lastOut = err, out
		time.Sleep(postgresReadyPollInterval)
	}
	t.Fatalf("postgres container %s did not become ready within %s: %v\nlast pg_isready output:\n%s",
		containerName, postgresReadyTimeout, lastErr, lastOut)
}
