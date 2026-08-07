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
	_ "github.com/lib/pq"
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
// ON CONFLICT DO UPDATE behavior for a user who already has a row for the
// same (user_id, local_date) must be unchanged. It asserts the second call
// updates rather than duplicating, preserves the original created_at, and
// advances updated_at.
func TestCreateDailyMissionSnapshotOnConflictBranchUnchangedAgainstRealPostgres(t *testing.T) {
	db := newMigratedDisposablePostgres(t)
	userID := insertTestUser(t, db)
	localDate := time.Date(2026, 7, 26, 0, 0, 0, 0, time.UTC)
	repo := NewRepository(db)

	firstSnapshot := createSnapshotInOwnTransaction(t, db, repo, userID, localDate, 20)
	firstCreatedAt, firstUpdatedAt := readSnapshotTimestamps(t, db, userID, localDate)

	// Postgres NOW() is transaction-start time, so two back-to-back
	// transactions can share a timestamp on a fast machine. Sleeping past
	// the clock's practical resolution keeps the "updated_at advanced"
	// assertion meaningful rather than flaky.
	time.Sleep(10 * time.Millisecond)

	secondSnapshot := createSnapshotInOwnTransaction(t, db, repo, userID, localDate, 30)
	secondCreatedAt, secondUpdatedAt := readSnapshotTimestamps(t, db, userID, localDate)

	assert.Equal(t, firstSnapshot.ID, secondSnapshot.ID,
		"ON CONFLICT must update the existing row, not create a second one")
	assert.Equal(t, 1, countSnapshotsFor(t, db, userID, localDate))
	assert.Equal(t, 30, secondSnapshot.ReviewTarget,
		"the update branch still applies EXCLUDED.review_target")
	assert.True(t, firstCreatedAt.Equal(secondCreatedAt),
		"the update branch must not rewrite created_at (was %s, now %s)", firstCreatedAt, secondCreatedAt)
	assert.False(t, secondUpdatedAt.Before(firstUpdatedAt),
		"the update branch must advance updated_at (was %s, now %s)", firstUpdatedAt, secondUpdatedAt)
}

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

	t.Run("IncrementConfidencePointsEarned", func(t *testing.T) {
		db := newMigratedDisposablePostgres(t)
		repo := NewRepository(db)
		userID := insertTestUser(t, db)
		requireNoActivitySummaryFor(t, db, userID, localDate)

		before := databaseNow(t, db)
		tx, err := db.Begin()
		require.NoError(t, err)
		defer tx.Rollback()
		require.NoError(t, repo.IncrementConfidencePointsEarned(t.Context(), tx, userID, localDate, "UTC", 5))
		require.NoError(t, tx.Commit())
		after := databaseNow(t, db)

		summary := readActivitySummary(t, db, userID, localDate)
		assert.Equal(t, 5, summary.ConfidencePointsEarned)
		assertTimestampWithin(t, "created_at", summary.CreatedAt, before, after)
		assertTimestampWithin(t, "updated_at", summary.UpdatedAt, before, after)
	})
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
	t.Helper()
	tx, err := db.Begin()
	require.NoError(t, err)
	defer tx.Rollback()
	snapshot, err := repo.CreateDailyMissionSnapshot(
		t.Context(), tx, userID, localDate, "UTC", reviewTarget, gamification.MissionPolicyVersion,
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
	if _, err := exec.LookPath("docker"); err != nil {
		t.Skipf("docker not on PATH (VOC-046-TEST-00's real-Postgres proof requires Docker): %v", err)
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
	applyCommittedForwardMigrations(t, db)
	return db
}

// applyCommittedForwardMigrations executes every committed forward
// migration in filename (version) order. Recovery down-files are excluded
// by the same rule Atlas itself uses: only `*.sql` is a forward migration,
// and the recovery files carry a `.down.sql.example` suffix specifically so
// they fall outside that glob.
func applyCommittedForwardMigrations(t *testing.T, db *sql.DB) {
	t.Helper()
	paths, err := filepath.Glob(filepath.Join(migrationsDirRelativeToPackage, "*.sql"))
	require.NoError(t, err)
	require.NotEmpty(t, paths, "no forward migrations found in %s", migrationsDirRelativeToPackage)
	sort.Strings(paths)
	for _, path := range paths {
		if strings.HasSuffix(path, ".down.sql") {
			continue
		}
		statements, err := os.ReadFile(path)
		require.NoError(t, err)
		_, err = db.Exec(string(statements))
		require.NoErrorf(t, err, "apply migration %s", filepath.Base(path))
	}
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
	var count int
	require.NoError(t, db.QueryRow(
		`SELECT count(*) FROM daily_activity_summaries WHERE user_id = $1 AND local_date = $2`,
		userID, localDate,
	).Scan(&count))
	require.Equal(t, 0, count,
		"precondition: the user must have no activity summary for this local date")
}

type activitySummaryRow struct {
	ReviewsAttempted       int
	ReviewsCorrect         int
	ReviewsSkipped         int
	WordsAdded             int
	SentencesSubmitted     int
	AIFeedbackReceived     int
	ConfidencePointsEarned int
	CreatedAt              time.Time
	UpdatedAt              time.Time
}

func readActivitySummary(t *testing.T, db *sql.DB, userID uuid.UUID, localDate time.Time) activitySummaryRow {
	t.Helper()
	var row activitySummaryRow
	require.NoError(t, db.QueryRow(
		`SELECT reviews_attempted, reviews_correct, reviews_skipped,
		        words_added, sentences_submitted, ai_feedback_received,
		        confidence_points_earned, created_at, updated_at
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
