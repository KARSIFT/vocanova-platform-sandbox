//go:build integration

// Package migrations_test - disposable-Postgres proof that VOC-050-T00's
// synthetic smoke-test seed is idempotent (VOC-050-AC-00 /
// VOC-050-TEST-00) and refuses to adopt an account it does not own
// (VOC-050-AC-01).
//
// How to run it locally:
//
//	go test -tags=integration -run SyntheticSmokeUserSeed ./apps/api/migrations/...
//
// Requirements: Docker (or any docker-compatible runtime on PATH as
// `docker`). Unlike the Atlas apply proof in this same package, this
// test does not need the Atlas CLI: it applies the committed migration
// files in filename order with psql inside the container, which is
// enough to produce the schema the seed runs against. The test skips
// cleanly when Docker is missing.
//
// The same `integration` build-tag trade-off documented at length in
// atlas_apply_integration_test.go applies here: this proof runs on
// demand rather than on every PR, because adding a Postgres service
// container to the shared CI workflow is not possible from this
// repository.
package migrations_test

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

// syntheticSeedTestEmail is the reserved identity this test seeds. It
// deliberately differs from the production default so a passing test
// can never be mistaken for evidence about a real environment's row.
const syntheticSeedTestEmail = "seed-proof-bot@synthetic.vocanova.invalid"

// syntheticSeedRotatedTestEmail exercises the address-rotation branch:
// seeding a second, different reserved identity must retire the first
// one rather than collide with users_single_synthetic_test_account_idx.
const syntheticSeedRotatedTestEmail = "seed-proof-bot-rotated@synthetic.vocanova.invalid"

// TestSyntheticSmokeUserSeedIsIdempotent runs the committed seed SQL
// twice in a row against a freshly-migrated disposable database and
// asserts that the second run neither errors nor duplicates the
// account - the exact procedure VOC-050-TEST-00 describes.
func TestSyntheticSmokeUserSeedIsIdempotent(t *testing.T) {
	container := startMigratedDisposablePostgres(t)

	runSyntheticSeed(t, container, syntheticSeedTestEmail, "first seed run (account absent)")
	firstID := querySingleValue(t, container,
		"SELECT id FROM users WHERE is_synthetic_test_account AND deleted_at IS NULL")

	runSyntheticSeed(t, container, syntheticSeedTestEmail, "second seed run (account already present)")

	if got := querySingleValue(t, container,
		"SELECT count(*) FROM users WHERE is_synthetic_test_account AND deleted_at IS NULL"); got != "1" {
		t.Fatalf("after a repeated seed run: want exactly 1 marked synthetic account, got %s", got)
	}
	secondID := querySingleValue(t, container,
		"SELECT id FROM users WHERE is_synthetic_test_account AND deleted_at IS NULL")
	if secondID != firstID {
		t.Fatalf("repeated seed run replaced the account: first id %s, second id %s", firstID, secondID)
	}

	assertSeededAccountShape(t, container, syntheticSeedTestEmail)
}

// TestSyntheticSmokeUserSeedRotatesTheReservedAddress proves the seed
// stays rerunnable when the reserved identity is reconfigured: the
// previously-seeded account is retired (soft-deleted, marker cleared)
// so the single-synthetic-account unique index admits the new one.
func TestSyntheticSmokeUserSeedRotatesTheReservedAddress(t *testing.T) {
	container := startMigratedDisposablePostgres(t)

	runSyntheticSeed(t, container, syntheticSeedTestEmail, "seed original reserved address")
	runSyntheticSeed(t, container, syntheticSeedRotatedTestEmail, "seed rotated reserved address")

	if got := querySingleValue(t, container,
		"SELECT count(*) FROM users WHERE is_synthetic_test_account AND deleted_at IS NULL"); got != "1" {
		t.Fatalf("after rotating the reserved address: want exactly 1 marked synthetic account, got %s", got)
	}
	if got := querySingleValue(t, container,
		fmt.Sprintf("SELECT status FROM users WHERE lower(email) = %s", sqlLiteral(syntheticSeedTestEmail))); got != "deleted" {
		t.Fatalf("previously-seeded account was not retired: want status deleted, got %q", got)
	}
	assertSeededAccountShape(t, container, syntheticSeedRotatedTestEmail)
}

// TestSyntheticSmokeUserSeedRefusesToAdoptARealAccount is the negative
// case behind VOC-050-AC-01: if the reserved address somehow already
// belongs to an unmarked account, the seed must fail loudly instead of
// promoting that account to the synthetic identity.
func TestSyntheticSmokeUserSeedRefusesToAdoptARealAccount(t *testing.T) {
	container := startMigratedDisposablePostgres(t)

	execSQL(t, container, fmt.Sprintf(`INSERT INTO users (id, email, display_name, status, onboarding_status, created_at, updated_at)
VALUES (gen_random_uuid(), %s, 'Pre-existing account', 'active', 'completed', now(), now())`,
		sqlLiteral(syntheticSeedTestEmail)), "insert a pre-existing unmarked account")

	out, err := syntheticSeedCommand(container, syntheticSeedTestEmail).CombinedOutput()
	if err == nil {
		t.Fatalf("seed succeeded against a pre-existing unmarked account; want a refusal\noutput:\n%s", out)
	}
	if !strings.Contains(string(out), "already registered to a non-synthetic account") {
		t.Fatalf("seed failed for the wrong reason; want the non-synthetic refusal, got output:\n%s", out)
	}
	if got := querySingleValue(t, container,
		"SELECT count(*) FROM users WHERE is_synthetic_test_account"); got != "0" {
		t.Fatalf("refused seed still marked an account as synthetic (count %s)", got)
	}
}

// startMigratedDisposablePostgres boots a throwaway postgres:16-alpine
// container and applies every committed forward migration to it in
// filename order, returning the container name. The container is
// removed by t.Cleanup even when the test fails.
func startMigratedDisposablePostgres(t *testing.T) string {
	t.Helper()
	if _, err := exec.LookPath("docker"); err != nil {
		t.Skipf("docker not on PATH (VOC-050-T00 seed idempotency proof requires Docker): %v", err)
	}

	containerName := "voc050-seed-test-" + randomHex(t, 6)
	startDisposablePostgres(t, containerName, freeLocalhostTCPPort(t))
	t.Cleanup(func() {
		removeContainer(t, containerName)
	})
	waitForPostgresReady(t, containerName, applyProofTestTimeout)

	for _, migration := range forwardMigrationFiles(t) {
		runPsqlFile(t, containerName, migration, "apply "+filepath.Base(migration))
	}
	return containerName
}

// forwardMigrationFiles lists the committed forward migrations in the
// order Atlas would apply them. The `.down.sql.example` recovery files
// are excluded for the same reason Atlas's own `*.sql` glob excludes
// them (see apps/api/migrations/README.md).
func forwardMigrationFiles(t *testing.T) []string {
	t.Helper()
	matches, err := filepath.Glob("*.sql")
	if err != nil {
		t.Fatalf("glob forward migrations: %v", err)
	}
	var forward []string
	for _, name := range matches {
		if strings.HasSuffix(name, ".down.sql") {
			continue
		}
		abs, err := filepath.Abs(name)
		if err != nil {
			t.Fatalf("resolve migration path %q: %v", name, err)
		}
		forward = append(forward, abs)
	}
	if len(forward) == 0 {
		t.Fatal("no forward migrations discovered; test misconfiguration")
	}
	sort.Strings(forward)
	return forward
}

// syntheticSeedCommand builds the psql invocation that runs the
// committed seed SQL, mirroring how
// apps/api/scripts/seed-synthetic-smoke-user.sh invokes it (same file,
// same psql variables) so the test exercises the shipped SQL rather
// than a copy of it.
func syntheticSeedCommand(containerName, email string) *exec.Cmd {
	cmd := exec.Command("docker", "exec", "-i", containerName,
		"psql",
		"--set=ON_ERROR_STOP=1",
		"--username", "vocanova",
		"--dbname", "vocanova",
		"--set=synthetic_email="+email,
		"--set=synthetic_display_name=VOC-050 Synthetic Smoke Test User",
		"--file", "-",
	)
	cmd.Stdin = mustOpenSeedSQL()
	return cmd
}

func mustOpenSeedSQL() *os.File {
	f, err := os.Open(filepath.Join("..", "scripts", "seed-synthetic-smoke-user.sql"))
	if err != nil {
		panic(fmt.Sprintf("open seed SQL: %v", err))
	}
	return f
}

func runSyntheticSeed(t *testing.T, containerName, email, label string) {
	t.Helper()
	out, err := syntheticSeedCommand(containerName, email).CombinedOutput()
	if err != nil {
		t.Fatalf("%s: seed failed: %v\noutput:\n%s", label, err, out)
	}
	t.Logf("%s succeeded; psql output:\n%s", label, out)
}

// assertSeededAccountShape checks the properties every consumer of the
// account depends on: it is marked synthetic, active, already past
// onboarding (so no core-loop check needs an onboarding override), and
// registered under the reserved address.
func assertSeededAccountShape(t *testing.T, containerName, email string) {
	t.Helper()
	got := querySingleValue(t, containerName,
		"SELECT email || '|' || status || '|' || onboarding_status FROM users WHERE is_synthetic_test_account AND deleted_at IS NULL")
	want := email + "|active|completed"
	if got != want {
		t.Fatalf("seeded account shape: want %q, got %q", want, got)
	}
}

func runPsqlFile(t *testing.T, containerName, path, label string) {
	t.Helper()
	f, err := os.Open(path)
	if err != nil {
		t.Fatalf("%s: open %s: %v", label, path, err)
	}
	defer f.Close()

	cmd := exec.Command("docker", "exec", "-i", containerName,
		"psql", "--set=ON_ERROR_STOP=1", "--username", "vocanova", "--dbname", "vocanova", "--file", "-")
	cmd.Stdin = f
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("%s: %v\noutput:\n%s", label, err, out)
	}
}

func execSQL(t *testing.T, containerName, statement, label string) {
	t.Helper()
	cmd := exec.Command("docker", "exec", "-i", containerName,
		"psql", "--set=ON_ERROR_STOP=1", "--username", "vocanova", "--dbname", "vocanova", "--command", statement)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("%s: %v\noutput:\n%s", label, err, out)
	}
}

func querySingleValue(t *testing.T, containerName, query string) string {
	t.Helper()
	cmd := exec.Command("docker", "exec", "-i", containerName,
		"psql", "--set=ON_ERROR_STOP=1", "--tuples-only", "--no-align",
		"--username", "vocanova", "--dbname", "vocanova", "--command", query)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("query %q: %v\noutput:\n%s", query, err, out)
	}
	return strings.TrimSpace(string(out))
}

// sqlLiteral renders a Go string as a single-quoted SQL literal for the
// small, test-only statements above. Only the test's own constants pass
// through it; no external input reaches it.
func sqlLiteral(value string) string {
	return "'" + strings.ReplaceAll(value, "'", "''") + "'"
}
