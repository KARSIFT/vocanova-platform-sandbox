//go:build integration

package migrations_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"testing"
)

var stagingRecoveryMigrations = map[string]struct{}{
	"20260909142059_review_attempt_user_word_integrity.sql":     {},
	"20260909142060_mission_optional_goal_pair_integrity.sql":   {},
	"20260909142061_mission_completed_at_integrity.sql":         {},
	"20260909142062_review_attempt_result_rating_integrity.sql": {},
}

// TestAtlasMigrateApplyAfterStagingHistoryRemainsLinear reproduces the
// deployment shape that caused #1430. It first applies the migration history
// that staging could have reached before the four recovery migrations were
// rebased, then applies the complete current directory. Atlas must accept the
// latter as an append-only continuation; accepting it only with non-linear or
// linear-skip execution would leave the database's migration history unsafe.
func TestAtlasMigrateApplyAfterStagingHistoryRemainsLinear(t *testing.T) {
	if _, err := exec.LookPath("docker"); err != nil {
		t.Skipf("docker not on PATH (linear-history apply proof requires Docker): %v", err)
	}
	if _, err := exec.LookPath("atlas"); err != nil {
		t.Skipf("atlas not on PATH (linear-history apply proof requires Atlas): %v", err)
	}

	port := freeLocalhostTCPPort(t)
	containerName := "voc1430-atlas-linear-history-" + randomHex(t, 6)
	dbURL := "postgres://vocanova:vocanova@127.0.0.1:" + strconv.Itoa(port) + "/vocanova?sslmode=disable"

	startDisposablePostgres(t, containerName, port)
	t.Cleanup(func() { removeContainer(t, containerName) })
	waitForPostgresReady(t, containerName, applyProofTestTimeout)

	stagingHistoryDir := copyMigrationHistoryBeforeRecovery(t)
	applyAtlasMigrateFromDir(t, dbURL, stagingHistoryDir, "staging history before recovered migrations")
	applyAtlasMigrate(t, dbURL, "current history after recovered migrations")
}

func copyMigrationHistoryBeforeRecovery(t *testing.T) string {
	t.Helper()
	sourceDir := migrationDirectory(t)
	destinationDir := t.TempDir()
	entries, err := os.ReadDir(sourceDir)
	if err != nil {
		t.Fatalf("read committed migrations: %v", err)
	}
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".sql" {
			continue
		}
		if _, excluded := stagingRecoveryMigrations[entry.Name()]; excluded {
			continue
		}
		contents, err := os.ReadFile(filepath.Join(sourceDir, entry.Name()))
		if err != nil {
			t.Fatalf("read migration %s: %v", entry.Name(), err)
		}
		if err := os.WriteFile(filepath.Join(destinationDir, entry.Name()), contents, 0o644); err != nil {
			t.Fatalf("copy migration %s: %v", entry.Name(), err)
		}
	}

	cmd := exec.Command("atlas", "migrate", "hash", "--dir", "file://"+destinationDir)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("hash staging-history migration directory: %v\noutput:\n%s", err, out)
	}
	return destinationDir
}
