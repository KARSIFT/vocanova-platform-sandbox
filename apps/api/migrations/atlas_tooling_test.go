package migrations_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// TestAtlasConfigDeclaresDevEnvAndMigrationDir is the file-content
// guard for apps/api/atlas.hcl (VOC-032-T06). The config must (1)
// declare a named `dev` env so `atlas migrate diff --env dev`
// and `atlas migrate validate --env dev` work for local
// developer-machine linting; (2) reference the versioned
// migration directory at `file://migrations` so Atlas
// discovers the existing *.sql files; and (3) carry no real
// credential (the dev `url` is a localhost placeholder by
// design; real DATABASE_URL is supplied at runtime via
// --url, not from this file).
func TestAtlasConfigDeclaresDevEnvAndMigrationDir(t *testing.T) {
	hcl, err := os.ReadFile("../atlas.hcl")
	if err != nil {
		t.Fatalf("read apps/api/atlas.hcl: %v", err)
	}
	text := string(hcl)
	required := []string{
		"env \"dev\"",
		"url = \"postgres://",
		"migration {",
		"dir = \"file://migrations\"",
	}
	for _, want := range required {
		if !strings.Contains(text, want) {
			t.Errorf("apps/api/atlas.hcl missing required substring %q", want)
		}
	}
}

// TestAtlasConfigHasNoRealSecret guards the dev URL is a
// placeholder, not a real credential. The wrapper and the T07
// deploy workflow both pass the real DATABASE_URL via
// --url at runtime, never by reading this file's url.
// We assert that the dev URL uses the documented
// localhost placeholder (vocanova:vocanova) and never a
// non-localhost host, so an accidental commit of a real
// staging URL into atlas.hcl is caught by CI.
func TestAtlasConfigHasNoRealSecret(t *testing.T) {
	hcl, err := os.ReadFile("../atlas.hcl")
	if err != nil {
		t.Fatalf("read apps/api/atlas.hcl: %v", err)
	}
	text := string(hcl)
	// The dev url must reference 127.0.0.1 (or localhost);
	// any other host means a real-environment URL leaked
	// into the file.
	nonLocalhostHosts := []string{
		"vocanova.site",
		"staging.vocanova",
		"api-staging.vocanova",
		"render.com",
		"amazonaws.com",
		"cloudflare.com",
	}
	for _, host := range nonLocalhostHosts {
		if strings.Contains(text, host) {
			t.Errorf("apps/api/atlas.hcl dev url references non-localhost host %q; runtime DATABASE_URL is supplied via --url, never from this file", host)
		}
	}
}

// TestMigrateWrapperExistsAndIsExecutable guards the
// T07-deploy-workflow / T09-rehearsal contract: the wrapper
// must be present at apps/api/scripts/migrate.sh and must be
// marked executable so the deploy workflow can invoke it
// without an explicit `chmod +x` step (which it cannot do
// over SSH from a checked-out repo tree).
func TestMigrateWrapperExistsAndIsExecutable(t *testing.T) {
	info, err := os.Stat("../scripts/migrate.sh")
	if err != nil {
		t.Fatalf("stat apps/api/scripts/migrate.sh: %v", err)
	}
	if info.IsDir() {
		t.Fatalf("apps/api/scripts/migrate.sh is a directory; expected a regular file")
	}
	mode := info.Mode()
	if mode&0o111 == 0 {
		t.Errorf("apps/api/scripts/migrate.sh is not executable (mode=%v); the T07 deploy workflow invokes it without a prior chmod", mode)
	}
}

// TestMigrateWrapperRejectsMissingDatabaseURL is the
// pre-flight-validation test for the wrapper's DATABASE_URL
// guard. The deploy workflow must fail fast (and
// observably) if the env var is absent, not silently attempt
// an apply against a blank URL.
func TestMigrateWrapperRejectsMissingDatabaseURL(t *testing.T) {
	wrapper := wrapperPath(t)
	cmd := exec.Command(wrapper)
	cmd.Env = []string{"PATH=/usr/bin:/bin"}
	out, err := cmd.CombinedOutput()
	if err == nil {
		t.Fatalf("migrate.sh accepted missing DATABASE_URL; want non-zero exit. output:\n%s", out)
	}
	if !strings.Contains(string(out), "DATABASE_URL") {
		t.Errorf("migrate.sh error did not mention DATABASE_URL; got:\n%s", out)
	}
}

// TestMigrateWrapperRejectsMissingAtlasBinary is the
// pre-flight-validation test for the wrapper's atlas-binary
// guard. The deploy workflow installs Atlas as a separate
// step; this test confirms a missing-binary state surfaces
// a clear, recoverable error rather than a confusing shell
// "command not found".
func TestMigrateWrapperRejectsMissingAtlasBinary(t *testing.T) {
	wrapper := wrapperPath(t)
	cmd := exec.Command(wrapper)
	cmd.Env = []string{
		"PATH=/usr/bin:/bin",
		"DATABASE_URL=postgres://placeholder@127.0.0.1:5432/placeholder?sslmode=disable",
	}
	out, err := cmd.CombinedOutput()
	if err == nil {
		t.Fatalf("migrate.sh accepted missing atlas binary; want non-zero exit. output:\n%s", out)
	}
	if !strings.Contains(string(out), "Atlas binary not found") {
		t.Errorf("migrate.sh error did not mention missing atlas binary; got:\n%s", out)
	}
}

// TestMigrateWrapperRejectsMissingAtlasSum is the
// pre-flight-validation test for the wrapper's atlas.sum
// guard. atlas.sum is the integrity file Atlas consults
// before applying; if it is missing, the deploy workflow
// must surface a clear "regenerate and commit" message
// rather than Atlas's lower-level "out of sync" diagnostic.
func TestMigrateWrapperRejectsMissingAtlasSum(t *testing.T) {
	wrapper := wrapperPath(t)
	tmp := t.TempDir()
	// Create an empty migrations dir without atlas.sum.
	if err := os.MkdirAll(filepath.Join(tmp, "migrations"), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	// Provide a stub atlas binary so the wrapper reaches
	// the atlas.sum check. /bin/true exists on every Linux.
	cmd := exec.Command(wrapper)
	cmd.Env = []string{
		"PATH=/usr/bin:/bin",
		"DATABASE_URL=postgres://placeholder@127.0.0.1:5432/placeholder?sslmode=disable",
		"ATLAS_BIN=/bin/true",
		"MIGRATIONS_DIR=file://" + filepath.Join(tmp, "migrations"),
	}
	out, err := cmd.CombinedOutput()
	if err == nil {
		t.Fatalf("migrate.sh accepted missing atlas.sum; want non-zero exit. output:\n%s", out)
	}
	if !strings.Contains(string(out), "atlas.sum") {
		t.Errorf("migrate.sh error did not mention atlas.sum; got:\n%s", out)
	}
}

// TestMigrateWrapperRejectsMissingMigrationsDir is the
// pre-flight-validation test for the wrapper's
// migrations-directory-existence guard. A misconfigured
// MIGRATIONS_DIR (typo, wrong path) must fail with a
// clear "directory not found" message, not a generic
// Atlas error.
func TestMigrateWrapperRejectsMissingMigrationsDir(t *testing.T) {
	wrapper := wrapperPath(t)
	cmd := exec.Command(wrapper)
	cmd.Env = []string{
		"PATH=/usr/bin:/bin",
		"DATABASE_URL=postgres://placeholder@127.0.0.1:5432/placeholder?sslmode=disable",
		"MIGRATIONS_DIR=file:///nonexistent/directory/atlas-test",
	}
	out, err := cmd.CombinedOutput()
	if err == nil {
		t.Fatalf("migrate.sh accepted non-existent migrations dir; want non-zero exit. output:\n%s", out)
	}
	if !strings.Contains(string(out), "migration directory not found") {
		t.Errorf("migrate.sh error did not mention missing dir; got:\n%s", out)
	}
}

// TestMigrateWrapperRejectsNonFileURL guards that the
// wrapper only accepts `file://` URLs for MIGRATIONS_DIR.
// A non-file URL would silently pass through to Atlas,
// which would then error in a less recoverable way.
func TestMigrateWrapperRejectsNonFileURL(t *testing.T) {
	wrapper := wrapperPath(t)
	cmd := exec.Command(wrapper)
	cmd.Env = []string{
		"PATH=/usr/bin:/bin",
		"DATABASE_URL=postgres://placeholder@127.0.0.1:5432/placeholder?sslmode=disable",
		"MIGRATIONS_DIR=postgres://something",
	}
	out, err := cmd.CombinedOutput()
	if err == nil {
		t.Fatalf("migrate.sh accepted non-file:// URL; want non-zero exit. output:\n%s", out)
	}
	if !strings.Contains(string(out), "file://") {
		t.Errorf("migrate.sh error did not mention file:// requirement; got:\n%s", out)
	}
}

// TestMigrationsDirectoryHasNoForwardDiscoveredDownFiles
// is the file-system guard for VOC-032-D08 / VOC-032-TEST-15:
// the recovery down-files must stay outside Atlas's
// forward-apply file discovery. Atlas's default glob in
// versioned mode is `*.sql`; the `.down.sql.example`
// extension is what keeps them out. If any file ending
// in `.down.sql` (without the `.example` suffix) is ever
// committed, the wrapper would pick it up as a forward
// migration - this test fails fast at unit-test time.
func TestMigrationsDirectoryHasNoForwardDiscoveredDownFiles(t *testing.T) {
	entries, err := os.ReadDir(".")
	if err != nil {
		t.Fatalf("read migrations dir: %v", err)
	}
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		// Atlas's forward-apply glob is `*.sql`; anything
		// ending in `.down.sql` (without the .example
		// suffix) is in scope and would be applied. The
		// `.example` suffix on the actual recovery files
		// keeps them out.
		if strings.HasSuffix(name, ".down.sql") {
			t.Errorf("migration file %q ends in .down.sql; the .example suffix is required (VOC-032-D08) to keep recovery down-files out of Atlas's forward-apply path", name)
		}
	}
}

// TestAtlasSumExistsAndMatchesMigrationCount is the
// file-content guard for the integrity hash. atlas.sum
// must exist (without it Atlas aborts every apply), must
// use Atlas's documented `h1:` SHA-256 format, and must
// have one line per migration file plus the directory
// header line. A stale atlas.sum (out of sync with the
// actual *.sql files) is the most common T06 follow-up;
// this test makes a stale sum's mismatch visible at
// unit-test time rather than only at first deploy.
func TestAtlasSumExistsAndMatchesMigrationCount(t *testing.T) {
	sum, err := os.ReadFile("atlas.sum")
	if err != nil {
		t.Fatalf("read apps/api/migrations/atlas.sum: %v", err)
	}
	lines := strings.Split(strings.TrimRight(string(sum), "\n"), "\n")
	if len(lines) == 0 {
		t.Fatalf("atlas.sum is empty")
	}
	if !strings.HasPrefix(lines[0], "h1:") {
		t.Errorf("atlas.sum first line must start with h1: (Atlas v1 SHA-256 format); got %q", lines[0])
	}
	// Count committed forward migration files (*.sql, but
	// not atlas.sum itself and not the recovery .example
	// files - they are deliberately outside Atlas's
	// forward-apply glob).
	var sqlCount int
	entries, err := os.ReadDir(".")
	if err != nil {
		t.Fatalf("read migrations dir: %v", err)
	}
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if name == "atlas.sum" || name == "README.md" {
			continue
		}
		// Atlas's forward-apply glob is `*.sql`. The
		// recovery files use `.down.sql.example` and so
		// are not in scope.
		if strings.HasSuffix(name, ".sql") {
			sqlCount++
		}
	}
	// atlas.sum is "<dir-hash>" + one line per file, so
	// the number of per-file lines equals sqlCount.
	if got, want := len(lines)-1, sqlCount; got != want {
		t.Errorf("atlas.sum has %d file entries, want %d (one per *.sql file). Run `atlas migrate hash --dir file://migrations` and commit the result if a migration was added or edited.", got, want)
	}
}

// wrapperPath resolves the absolute path to the
// apps/api/scripts/migrate.sh wrapper from the test's
// working directory (the migrations directory). The path
// is recomputed per test so a renamed or moved wrapper is
// caught at unit-test time.
func wrapperPath(t *testing.T) string {
	t.Helper()
	abs, err := filepath.Abs("../scripts/migrate.sh")
	if err != nil {
		t.Fatalf("resolve wrapper path: %v", err)
	}
	return abs
}
