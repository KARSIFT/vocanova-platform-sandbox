package api

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

	_ "github.com/lib/pq"
	"github.com/stretchr/testify/require"
)

const (
	controlledSignupPostgresReadyTimeout      = 60 * time.Second
	controlledSignupPostgresReadyPollInterval = 250 * time.Millisecond
	controlledSignupPostgresImage             = "postgres:16-alpine"
	controlledSignupMigrationsDirRelative     = "../../migrations"
)

// newControlledSignupDisposablePostgres starts a loopback-bound disposable
// Postgres container, applies the committed forward migrations, and returns
// an open *sql.DB. The harness never reads DATABASE_URL from the environment.
func newControlledSignupDisposablePostgres(t *testing.T) *sql.DB {
	t.Helper()
	if _, err := exec.LookPath("docker"); err != nil {
		t.Skipf("docker not on PATH (controlled-signup OAuth E2E requires Docker): %v", err)
	}

	port := controlledSignupFreeLoopbackTCPPort(t)
	containerName := "voc092-controlled-signup-oauth-" + controlledSignupRandomHexSuffix(t, 6)
	controlledSignupStartPostgresContainer(t, containerName, port)
	t.Cleanup(func() { controlledSignupForceRemoveContainer(t, containerName) })
	controlledSignupWaitForPostgres(t, containerName)

	db, err := sql.Open("postgres",
		fmt.Sprintf("postgres://vocanova:vocanova@127.0.0.1:%d/vocanova?sslmode=disable", port))
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })
	controlledSignupRequirePingSucceeds(t, db)
	controlledSignupApplyForwardMigrations(t, db)
	return db
}

func controlledSignupApplyForwardMigrations(t *testing.T, db *sql.DB) {
	t.Helper()
	paths, err := filepath.Glob(filepath.Join(controlledSignupMigrationsDirRelative, "*.sql"))
	require.NoError(t, err)
	require.NotEmpty(t, paths, "no forward migrations found in %s", controlledSignupMigrationsDirRelative)
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

func controlledSignupRequirePingSucceeds(t *testing.T, db *sql.DB) {
	t.Helper()
	deadline := time.Now().Add(controlledSignupPostgresReadyTimeout)
	var lastErr error
	for time.Now().Before(deadline) {
		if lastErr = db.Ping(); lastErr == nil {
			return
		}
		time.Sleep(controlledSignupPostgresReadyPollInterval)
	}
	t.Fatalf("could not connect to disposable Postgres within %s: %v", controlledSignupPostgresReadyTimeout, lastErr)
}

func controlledSignupFreeLoopbackTCPPort(t *testing.T) int {
	t.Helper()
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	defer listener.Close()
	addr, ok := listener.Addr().(*net.TCPAddr)
	require.Truef(t, ok, "unexpected listener address type %T", listener.Addr())
	return addr.Port
}

func controlledSignupRandomHexSuffix(t *testing.T, byteCount int) string {
	t.Helper()
	buf := make([]byte, byteCount)
	_, err := rand.Read(buf)
	require.NoError(t, err)
	return hex.EncodeToString(buf)
}

func controlledSignupStartPostgresContainer(t *testing.T, containerName string, hostPort int) {
	t.Helper()
	args := []string{
		"run", "--rm", "-d",
		"--name", containerName,
		"-p", fmt.Sprintf("127.0.0.1:%d:5432", hostPort),
		"-e", "POSTGRES_USER=vocanova",
		"-e", "POSTGRES_PASSWORD=vocanova",
		"-e", "POSTGRES_DB=vocanova",
		controlledSignupPostgresImage,
	}
	out, err := exec.Command("docker", args...).CombinedOutput()
	if err != nil {
		t.Fatalf("docker %s: %v\noutput:\n%s", strings.Join(args, " "), err, out)
	}
}

func controlledSignupForceRemoveContainer(t *testing.T, containerName string) {
	t.Helper()
	out, err := exec.Command("docker", "rm", "-f", containerName).CombinedOutput()
	if err != nil && !bytes.Contains(out, []byte("No such container")) {
		t.Logf("docker rm -f %s: %v", containerName, err)
	}
}

func controlledSignupWaitForPostgres(t *testing.T, containerName string) {
	t.Helper()
	deadline := time.Now().Add(controlledSignupPostgresReadyTimeout)
	var lastErr error
	var lastOut []byte
	for time.Now().Before(deadline) {
		out, err := exec.Command("docker", "exec", containerName,
			"pg_isready", "-U", "vocanova", "-d", "vocanova").CombinedOutput()
		if err == nil {
			return
		}
		lastErr, lastOut = err, out
		time.Sleep(controlledSignupPostgresReadyPollInterval)
	}
	t.Fatalf("postgres container %s did not become ready within %s: %v\nlast pg_isready output:\n%s",
		containerName, controlledSignupPostgresReadyTimeout, lastErr, lastOut)
}
