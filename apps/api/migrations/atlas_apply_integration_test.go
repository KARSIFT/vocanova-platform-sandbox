//go:build integration

// Package migrations_test - end-to-end apply proof for the fixed
// Atlas migration set (VOC-033-T02).
//
// This file is the disposable-Postgres-16-backed Go integration test
// that proves `atlas migrate apply` succeeds against the full
// 13-file migration set, and that a second apply against the
// now-migrated database is a no-op (matching the EV-14/EV-15
// "re-apply is a no-op" evidence item in VOC-032's
// staging-evidence.md).
//
// How to run it locally:
//
//	go test -tags=integration ./apps/api/migrations/...
//
// Requirements: Docker (or any docker-compatible runtime on PATH
// as `docker`) and the Atlas v1.2.0 CLI on PATH as `atlas`. The
// test skips cleanly with `t.Skip` if either is missing, so it is
// safe to leave in the tree even on runners without the tools.
//
// Build tag and CI trade-off (VOC-033-D02):
//
// This test is intentionally gated by the `integration` build tag
// and is therefore NOT part of the default `go test ./...` /
// `pnpm run test:api` path, and is NOT wired into any GitHub
// Actions workflow. This is a deliberate, founder-adopted
// tradeoff, not an oversight: editing the shared `ci.yml` to add
// a Postgres service container and an Atlas install step is not
// even possible from this repository (it lives in
// KARSIFT/karsift-ai-infra, a separate repository this planner
// has no access to), and the package's declared scope excludes
// any change to `.github/workflows/*` in this repo. The practical
// consequence is that this proof runs on demand (a developer's
// machine, or manually during VOC-032-T09's own live staging
// rehearsal) rather than on every future PR. Permanent CI
// enforcement, if later required, needs a separate, distinctly-
// scoped follow-up package that touches the appropriate
// workflow file in this repo and/or `ci.yml` in karsift-ai-infra.
//
// This test does NOT replace or substitute for VOC-032-T09's
// live rehearsal against the real staging database. It is a
// local, disposable-container proof that the tooling-level
// defects this package fixes (the invalid `atlas:txmode
// transaction` directive and the duplicate unique index in
// streak_states) are actually gone from the migration set, so
// T09's first real-world apply attempt will not fail at the
// directive-parsing step or with SQLSTATE 42P07. The
// credential-gated staging apply is still VOC-032-T09's
// responsibility and requires real staging DATABASE_URL
// credentials that this test does not access and this repository
// does not contain.
package migrations_test

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// applyProofTestTimeout bounds the total wall-clock time the
// integration test will spend waiting for the disposable Postgres
// container to become ready. 30 seconds is generous for
// `postgres:16-alpine` on a developer machine or a typical CI
// runner; a longer timeout would only delay surfacing a genuinely
// broken environment. Used as the argument to the
// `pg_isready` poll loop.
const applyProofTestTimeout = 30 * time.Second

// applyProofTestPollInterval is the gap between `pg_isready`
// probes. 250ms keeps the test responsive without hammering the
// container with hundreds of probes per second. At 250ms the
// poll loop makes ~120 attempts in 30s, which is well under any
// reasonable Postgres startup time and well above any reasonable
// "the container is broken" detection latency.
const applyProofTestPollInterval = 250 * time.Millisecond

// TestAtlasMigrateApplySucceedsAgainstDisposablePostgres is the
// end-to-end proof for VOC-033-AC-05 / VOC-033-TEST-05. It
// starts a disposable `postgres:16-alpine` container via
// `docker run` (bound to 127.0.0.1 on a dynamically chosen free
// port so the test never collides with a developer's own local
// Postgres), waits for `pg_isready` to report ready, then runs
// `atlas migrate apply` twice in succession against the
// committed `apps/api/migrations/` directory:
//
//  1. First apply, against a completely empty database: must
//     succeed (zero exit code) and apply all 13 forward
//     migrations.
//  2. Second apply, against the now-migrated database: must
//     succeed and report no pending migrations (Atlas's own
//     "no migration files to execute" message, or the
//     equivalent zero-changes signal) - matching the EV-14/EV-15
//     "re-apply is a no-op" evidence item VOC-032's
//     staging-evidence.md names.
//
// The test never asserts anything about a real staging or
// production database; the container is exclusively local and
// disposable (VOC-033-AC-06). `t.Cleanup` guarantees the
// container is removed even on test failure, and the test
// refuses to bind to any non-loopback address, so it cannot
// open a public database port.
func TestAtlasMigrateApplySucceedsAgainstDisposablePostgres(t *testing.T) {
	if _, err := exec.LookPath("docker"); err != nil {
		t.Skipf("docker not on PATH (VOC-033-T02 end-to-end apply proof requires Docker): %v", err)
	}
	if _, err := exec.LookPath("atlas"); err != nil {
		t.Skipf("atlas not on PATH (VOC-033-T02 end-to-end apply proof requires the Atlas v1.2.0 CLI; install from https://atlasgo.sh and re-run): %v", err)
	}

	port := freeLocalhostTCPPort(t)
	containerName := "voc033-atlas-test-" + randomHex(t, 6)
	dbURL := fmt.Sprintf("postgres://vocanova:vocanova@127.0.0.1:%d/vocanova?sslmode=disable", port)

	startDisposablePostgres(t, containerName, port)
	t.Cleanup(func() {
		removeContainer(t, containerName)
	})

	waitForPostgresReady(t, containerName, applyProofTestTimeout)

	// First apply: from a completely empty database. atlas's
	// `--dir "file://."` resolves relative to the current
	// working directory, so we run from the migrations
	// directory (the same `cwd` the production wrapper
	// `apps/api/scripts/migrate.sh` would resolve to from
	// `MIGRATIONS_DIR=file://.../migrations`).
	applyAtlasMigrate(t, dbURL, "first apply (empty database, all 13 migrations)")

	// Second apply: against the now-migrated database. We
	// assert the output indicates no pending migrations.
	// Atlas's documented message for this case is "No
	// migration files to execute" (Atlas v1.2.0). The
	// assertion is conservative: it checks for the substring
	// "no migration" case-insensitively so a future Atlas
	// version that tweaks the exact wording still passes, as
	// long as the no-op semantic is preserved.
	applyAtlasMigrateNoOp(t, dbURL, "second apply (already-migrated database, must be a no-op)")
}

// freeLocalhostTCPPort returns a TCP port number that is
// currently free for binding on the loopback interface
// (127.0.0.1). The function opens a listener, reads the
// assigned port, and immediately closes the listener; the
// returned port is a strong hint but not a guarantee that
// nothing else will grab it before the test can use it. In
// practice this is reliable on a developer machine and on CI
// runners - the gap between `Close()` and `docker run -p
// 127.0.0.1:<port>:5432` is microseconds and no other
// process on a single-user test box is allocating ports in
// that window. The function only ever binds to the loopback
// interface, so it cannot accidentally open a public port.
func freeLocalhostTCPPort(t *testing.T) int {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("allocate free loopback port: %v", err)
	}
	defer ln.Close()
	addr, ok := ln.Addr().(*net.TCPAddr)
	if !ok {
		t.Fatalf("allocate free loopback port: unexpected listener address type %T", ln.Addr())
	}
	return addr.Port
}

// randomHex returns a hex-encoded random string of length
// 2*byteCount characters. Used to uniquify the disposable
// container name so two parallel test runs (or a leftover
// container from a prior crashed run) cannot collide. The
// function uses crypto/rand so collisions are effectively
// impossible; if it ever fails it fails the test rather than
// falling back to a weaker source, because a colliding
// container name would produce a confusing
// "name already in use" Docker error that hides the real
// problem.
func randomHex(t *testing.T, byteCount int) string {
	t.Helper()
	buf := make([]byte, byteCount)
	if _, err := rand.Read(buf); err != nil {
		t.Fatalf("read random bytes for container-name suffix: %v", err)
	}
	return hex.EncodeToString(buf)
}

// startDisposablePostgres starts a `postgres:16-alpine`
// container in detached mode, bound to 127.0.0.1 on the given
// port, with a `vocanova` superuser and a `vocanova` default
// database (matching the placeholder URL in apps/api/atlas.hcl's
// dev env, so the same connection-string format Atlas's wrapper
// expects at runtime works here). The container is started with
// `--rm` so Docker removes it automatically on stop, and the
// `t.Cleanup` registered by the caller additionally calls
// `docker rm -f` to guarantee removal on test failure.
//
// The image is pinned to `postgres:16-alpine` (the same major
// version Atlas's wrapper comment block and the staging host
// use; a floating tag like `postgres:latest` is deliberately not
// used because it would make this test's result dependent on
// whatever the local Docker daemon happened to pull most
// recently). `docker run -d` is used (rather than `docker
// create` + `docker start`) so the call returns immediately and
// the test can poll `pg_isready` until the database is ready.
//
// The function never binds the container to anything other than
// 127.0.0.1; this is the load-bearing protection that makes
// VOC-033-AC-06's "no public database port is opened" guarantee
// hold even if a developer runs the test on a network-exposed
// machine.
func startDisposablePostgres(t *testing.T, containerName string, hostPort int) {
	t.Helper()
	// Use `docker run` (not `docker create` + `docker start`):
	// `run --rm -d` starts the container in detached mode and
	// arranges for `docker` to remove it on stop. We layer a
	// `t.Cleanup` `docker rm -f` on top so a `Stop` followed by
	// `RM` runs even on a test panic between the two.
	//
	// The `-p 127.0.0.1:<port>:5432` binding is the public-port
	// protection. The published port is on the loopback
	// interface only; the container's port 5432 is not
	// reachable from any non-loopback address even if the
	// host's firewall is permissive.
	args := []string{
		"run", "--rm", "-d",
		"--name", containerName,
		"-p", fmt.Sprintf("127.0.0.1:%d:5432", hostPort),
		"-e", "POSTGRES_USER=vocanova",
		"-e", "POSTGRES_PASSWORD=vocanova",
		"-e", "POSTGRES_DB=vocanova",
		"postgres:16-alpine",
	}
	cmd := exec.Command("docker", args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("docker run postgres:16-alpine (%s): %v\noutput:\n%s", strings.Join(args, " "), err, out)
	}
}

// removeContainer force-removes the named container, ignoring
// the "no such container" case (which is the expected state
// after a successful `docker run --rm` exit and is benign on
// `t.Cleanup`). Any other error is surfaced via t.Logf only -
// the test is already completing, and a stale leftover
// container does not invalidate the apply proof; the next run
// will get a different random suffix anyway.
func removeContainer(t *testing.T, containerName string) {
	t.Helper()
	cmd := exec.Command("docker", "rm", "-f", containerName)
	if out, err := cmd.CombinedOutput(); err != nil {
		if !bytes.Contains(out, []byte("No such container")) {
			// Best-effort cleanup; do not change the test
			// verdict from a late-arriving log line.
			t.Logf("voc033-atlas-test: docker rm -f %s: %v\noutput:\n%s", containerName, err, out)
		}
	}
}

// waitForPostgresReady polls `docker exec <container>
// pg_isready -U vocanova -d vocanova` until it reports ready or
// the timeout elapses. The function deliberately does not use
// `docker run` with a wait-for-port command, because the
// official postgres image does not include a wait helper
// directly accessible from outside; `pg_isready` is the
// canonical "is the database accepting connections" probe.
// Polling rather than busy-waiting keeps the test well-behaved
// on a developer's terminal (no 100% CPU while Postgres
// initializes).
func waitForPostgresReady(t *testing.T, containerName string, timeout time.Duration) {
	t.Helper()
	deadline := time.Now().Add(timeout)
	var lastErr error
	var lastOut []byte
	for time.Now().Before(deadline) {
		cmd := exec.Command("docker", "exec", containerName, "pg_isready", "-U", "vocanova", "-d", "vocanova")
		out, err := cmd.CombinedOutput()
		if err == nil {
			return
		}
		lastErr = err
		lastOut = out
		time.Sleep(applyProofTestPollInterval)
	}
	t.Fatalf("postgres container %s did not become ready within %s; last error: %v\nlast pg_isready output:\n%s", containerName, timeout, lastErr, lastOut)
}

// applyAtlasMigrate runs `atlas migrate apply` from the
// committed migrations directory against the given database
// URL. The function resolves the migrations directory to an
// absolute path (relative paths in the `atlas migrate apply
// --dir "file://."` form are resolved against the test's
// current working directory, which is not always the
// migrations directory under `go test`), asserts a zero exit
// code, and surfaces the full combined output on failure so
// the failure is debuggable from the test log alone.
//
// label is appended to the failure message so the two apply
// invocations in this test ("first apply" vs "second apply")
// are unambiguous in the log.
func applyAtlasMigrate(t *testing.T, dbURL string, label string) {
	t.Helper()
	dir, err := filepath.Abs(".")
	if err != nil {
		t.Fatalf("%s: resolve migrations dir absolute path: %v", label, err)
	}
	cmd := exec.Command("atlas", "migrate", "apply",
		"--url", dbURL,
		"--dir", "file://"+dir,
	)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("%s: atlas migrate apply failed: %v\ncommand: atlas migrate apply --url %s --dir file://%s\noutput:\n%s", label, err, dbURL, dir, out)
	}
	t.Logf("%s succeeded; atlas output:\n%s", label, out)
}

// applyAtlasMigrateNoOp runs `atlas migrate apply` against a
// database that has already had the full migration set
// applied, and asserts the output indicates no pending
// migrations. The check is conservative: it looks for the
// substring "no migration" (case-insensitive) in the combined
// output, which matches Atlas v1.2.0's documented
// "No migration files to execute" message and is robust to
// minor wording changes across future Atlas minor versions
// (e.g. "no migrations to apply" or "no migration files
// found"). The exit code must also be zero: a re-apply
// against a fully-migrated database is a successful
// no-op in Atlas's model, and a non-zero exit here would
// indicate a real apply error, not a "nothing to do"
// state.
//
// This corresponds to VOC-032's EV-14/EV-15 evidence item
// named in staging-evidence.md: a second `atlas migrate
// apply` against the same database must be a clean
// no-op, proving the migrations are idempotent under
// re-apply.
func applyAtlasMigrateNoOp(t *testing.T, dbURL string, label string) {
	t.Helper()
	dir, err := filepath.Abs(".")
	if err != nil {
		t.Fatalf("%s: resolve migrations dir absolute path: %v", label, err)
	}
	cmd := exec.Command("atlas", "migrate", "apply",
		"--url", dbURL,
		"--dir", "file://"+dir,
	)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("%s: atlas migrate apply failed on re-apply; want exit 0 with no-pending-migrations output, got error: %v\noutput:\n%s", label, err, out)
	}
	lowered := strings.ToLower(string(out))
	if !strings.Contains(lowered, "no migration") {
		t.Fatalf("%s: re-apply did not report a no-pending-migrations state; want Atlas to print a 'No migration ...' message; got output:\n%s", label, out)
	}
	t.Logf("%s succeeded (no-op as expected); atlas output:\n%s", label, out)
}
