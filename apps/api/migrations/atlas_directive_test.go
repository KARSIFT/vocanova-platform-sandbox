package migrations_test

import (
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"
)

// validAtlasTxmodes is the set of per-file atlas:txmode values Atlas v1.x
// accepts. From Atlas's documented behavior, a per-file directive accepts
// exactly "none" and "file"; "all" is global-only and rejected inside a
// per-file directive. Any other value (e.g. the historic
// "-- atlas:txmode transaction" the pre-VOC-033 migration set used) aborts
// `atlas migrate apply` at directive-parsing time, before any SQL is run.
// See .karsift/lessons.md 2026-07-29 for the failing reproduction and
// VOC-033-D00 for the resolution choice (use "file", not delete the line).
var validAtlasTxmodes = map[string]bool{
	"none": true,
	"file": true,
}

// atlasTxmodeDirectiveRE matches a per-file "-- atlas:txmode <value>"
// directive on a line of its own and captures the value as group 1. (?m)
// enables multi-line mode so ^/$ match line boundaries; the value is a
// single non-whitespace token in this repo's migrations.
var atlasTxmodeDirectiveRE = regexp.MustCompile(`(?m)^\s*--\s*atlas:txmode\s+(\S+)\s*$`)

// createTableRE matches "CREATE TABLE <name> (" at the start of a line.
// Captures the table name as group 1. (?m) enables multi-line mode so ^
// matches line starts.
var createTableRE = regexp.MustCompile(`(?m)^\s*CREATE\s+TABLE\s+(\w+)\s*\(`)

// createUniqueIndexRE matches the full text of a CREATE UNIQUE INDEX
// statement, which may span multiple lines in this repo's style
// (e.g. `... ON <table> (<cols>)\n  WHERE ...;`). Captures the index
// name (group 1), table name (group 2), the column list including any
// whitespace and function expressions (group 3), and an optional WHERE
// clause up to the statement-terminating semicolon (group 4; empty
// when absent). (?s) makes `.` match newline; (?m) is not needed here
// because the whole statement is captured as one multi-line block.
var createUniqueIndexRE = regexp.MustCompile(
	`(?is)CREATE\s+UNIQUE\s+INDEX\s+(\S+)\s+ON\s+(\w+)\s*\(([^)]*)\)\s*(WHERE[^;]*)?\s*;`,
)

// inlineUniqueKeywordRE matches the SQL keyword UNIQUE with word
// boundaries, so it does NOT match identifiers that happen to contain
// the substring UNIQUE (e.g. a column named "user_unique_id" or a
// constraint named "user_unique_constraint"). The boundary check is
// what distinguishes the column-level UNIQUE constraint from those
// false-positive substrings.
var inlineUniqueKeywordRE = regexp.MustCompile(`\bUNIQUE\b`)

// tableLevelConstraintFirstWords are the first words on a line that
// indicate a table-level constraint, not a column definition. A line
// inside a CREATE TABLE block whose first word is one of these is
// skipped when looking for inline UNIQUE column constraints. Postgres
// table-level constraints are not preceded by a column name; they
// appear on their own line.
var tableLevelConstraintFirstWords = map[string]bool{
	"CONSTRAINT": true,
	"CHECK":      true,
	"UNIQUE":     true,
	"PRIMARY":    true,
	"FOREIGN":    true,
}

// inlineUnique represents a column with an inline UNIQUE column
// constraint inside a specific table, e.g. `user_id uuid NOT NULL
// UNIQUE` in `CREATE TABLE streak_states (...)`. Postgres
// auto-generates a unique index named "<table>_<col>_key" for this
// constraint, so a same-named explicit `CREATE UNIQUE INDEX` on the
// same column collides on first apply with SQLSTATE 42P07
// ("relation already exists").
type inlineUnique struct {
	table string
	col   string
}

// explicitUniqueIndex represents a non-partial, single-column
// explicit `CREATE UNIQUE INDEX`. Only this shape can collide with an
// inline UNIQUE column constraint: multi-column indexes and partial
// indexes (with a WHERE clause) do not produce the same
// "<table>_<col>_key" name that Postgres auto-generates for the
// inline UNIQUE, so they cannot collide on that auto-name.
type explicitUniqueIndex struct {
	table string
	col   string
}

// tableColumn is a (table, column) pair used as a map key when the
// caller does not care which side of the inline/explicit distinction
// produced the pair (e.g. when computing the intersection of the
// two sets in inlineUniqueCollisionsInDir).
type tableColumn struct {
	table string
	col   string
}

// extractInlineUniqueColumns parses a SQL migration file's text and
// returns every (table, column) pair that carries an inline UNIQUE
// column constraint. The parser walks CREATE TABLE blocks line by
// line, tracking parenthesis depth so the block's terminating `);`
// is recognized, and treats each non-comment, non-table-level-
// constraint line as a column definition whose first identifier is
// the column name. A column has an inline UNIQUE iff the line
// contains the SQL keyword UNIQUE with word boundaries. Multi-line
// column definitions are not currently used in this repo's
// migrations, so the per-line parsing is sufficient.
func extractInlineUniqueColumns(text string) []inlineUnique {
	var result []inlineUnique
	lines := strings.Split(text, "\n")
	var currentTable string
	var depth int
	for _, line := range lines {
		if currentTable == "" {
			if m := createTableRE.FindStringSubmatch(line); m != nil {
				currentTable = m[1]
				opens := strings.Count(line, "(")
				closes := strings.Count(line, ")")
				depth = opens - closes
				if depth <= 0 {
					currentTable = ""
				}
			}
			continue
		}
		depth += strings.Count(line, "(") - strings.Count(line, ")")
		if depth <= 0 {
			currentTable = ""
			continue
		}
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "--") {
			continue
		}
		first := firstWord(trimmed)
		if first == "" || tableLevelConstraintFirstWords[first] {
			continue
		}
		if inlineUniqueKeywordRE.MatchString(trimmed) {
			result = append(result, inlineUnique{table: currentTable, col: first})
		}
	}
	return result
}

// extractExplicitUniqueIndexes parses a SQL migration file's text and
// returns every (table, column) pair that is the target of a
// non-partial, single-column explicit `CREATE UNIQUE INDEX`. Multi-
// column and partial (`WHERE ...`) indexes are deliberately excluded
// because they cannot collide with an inline UNIQUE column
// constraint on the auto-generated "<table>_<col>_key" name; the
// test only needs to detect the actual collision class that broke
// the pre-VOC-033 migration set.
func extractExplicitUniqueIndexes(text string) []explicitUniqueIndex {
	var result []explicitUniqueIndex
	for _, m := range createUniqueIndexRE.FindAllStringSubmatch(text, -1) {
		whereClause := strings.TrimSpace(m[4])
		if whereClause != "" {
			continue
		}
		cols := splitColumnList(m[3])
		if len(cols) != 1 {
			continue
		}
		result = append(result, explicitUniqueIndex{table: m[2], col: cols[0]})
	}
	return result
}

// firstWord returns the first whitespace-delimited word on a line.
// Used to identify the column name at the start of a CREATE TABLE
// column definition. The function stops at the first space, tab, or
// open paren; open paren is treated as a separator because some
// types in this repo's migrations include a parenthesized argument
// (e.g. "varchar(64)") and the column name must end before it.
func firstWord(line string) string {
	line = strings.TrimLeft(line, " \t")
	for i, r := range line {
		if r == ' ' || r == '\t' || r == '(' {
			return line[:i]
		}
	}
	return line
}

// splitColumnList splits a SQL column list like "user_id,
// idempotency_key" into its trimmed parts. Parentheses inside
// expressions (e.g. "lower(email)") are not specially handled, but
// none of the column lists in this repo's migrations that this
// function is asked to split contain a comma inside a parenthesized
// expression on the right-hand side of a single-column unique
// index; functional indexes are filtered as multi-column at the
// call site.
func splitColumnList(raw string) []string {
	parts := strings.Split(raw, ",")
	for i, p := range parts {
		parts[i] = strings.TrimSpace(p)
	}
	return parts
}

// sortedValidTxmodeKeys returns the keys of validAtlasTxmodes in
// deterministic alphabetical order, for stable error messages.
func sortedValidTxmodeKeys() []string {
	keys := make([]string, 0, len(validAtlasTxmodes))
	for k := range validAtlasTxmodes {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// firstInvalidTxmodeInFile reads a SQL file and returns the first
// `-- atlas:txmode <value>` per-file directive value it finds that
// is not in validAtlasTxmodes. The empty string is returned when
// every directive in the file is valid (or the file has none). The
// function does not itself fail; the caller decides what to do with
// the result so the same helper is reusable by the real-file test
// and the fixture test.
func firstInvalidTxmodeInFile(path string) (string, error) {
	text, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	for _, match := range atlasTxmodeDirectiveRE.FindAllStringSubmatch(string(text), -1) {
		if !validAtlasTxmodes[match[1]] {
			return match[1], nil
		}
	}
	return "", nil
}

// inlineUniqueCollisionsInDir returns the set of (table, column)
// pairs that have BOTH an inline UNIQUE column constraint AND a
// non-partial, single-column explicit `CREATE UNIQUE INDEX` across
// all `*.sql` files in dir. Only the colliding pairs are returned;
// pairs that appear in only one set (which is the expected state
// for almost every table in this repo) are silently filtered. The
// function is reused by both the real-file test and the fixture
// test, so the fixture test proves the real-file test's detection
// path actually runs.
func inlineUniqueCollisionsInDir(dir string) (map[tableColumn]bool, error) {
	files, err := filepath.Glob(filepath.Join(dir, "*.sql"))
	if err != nil {
		return nil, err
	}
	inline := map[tableColumn]bool{}
	explicit := map[tableColumn]bool{}
	for _, name := range files {
		text, err := os.ReadFile(name)
		if err != nil {
			return nil, err
		}
		for _, p := range extractInlineUniqueColumns(string(text)) {
			inline[tableColumn{table: p.table, col: p.col}] = true
		}
		for _, p := range extractExplicitUniqueIndexes(string(text)) {
			explicit[tableColumn{table: p.table, col: p.col}] = true
		}
	}
	collisions := map[tableColumn]bool{}
	for k := range inline {
		if explicit[k] {
			collisions[k] = true
		}
	}
	return collisions, nil
}

// TestAtlasTxmodeDirectiveIsValidInEveryMigration is the real-file
// guard for VOC-033-AC-03 / VOC-033-TEST-03. Reads every forward
// migration file in this directory and asserts that every
// `-- atlas:txmode <value>` per-file directive uses a value Atlas
// v1.x accepts. The test globs `*.sql` rather than naming each of
// the 13 current files individually, so a migration added after
// this package merges is covered automatically (per VOC-033-D03's
// "general, not hardcoded" requirement). The test would have failed
// against the pre-VOC-033 file set, which declared
// `-- atlas:txmode transaction` (an invalid value) in every
// migration; VOC-033-T00 changed those to `-- atlas:txmode file`.
func TestAtlasTxmodeDirectiveIsValidInEveryMigration(t *testing.T) {
	files, err := filepath.Glob("*.sql")
	if err != nil {
		t.Fatalf("glob *.sql: %v", err)
	}
	if len(files) == 0 {
		t.Fatalf("no *.sql files in migrations directory; cannot validate txmode coverage")
	}
	for _, name := range files {
		invalid, err := firstInvalidTxmodeInFile(name)
		if err != nil {
			t.Fatalf("read %s: %v", name, err)
		}
		if invalid != "" {
			t.Errorf("file %q declares invalid atlas:txmode %q; Atlas v1.x accepts only %v in a per-file directive (see .karsift/lessons.md 2026-07-29 and VOC-033-D00)", name, invalid, sortedValidTxmodeKeys())
		}
	}
}

// TestAtlasTxmodeFixtureRejectsTransactionValue is the synthetic-
// fixture half of VOC-033-AC-03's "test is proven to actually
// catch the regression, not just observe today's files happen to
// pass" requirement. It writes a single SQL file to t.TempDir()
// declaring the exact invalid value this package fixes
// (`-- atlas:txmode transaction`) and asserts the same
// firstInvalidTxmodeInFile helper used by the real-file test
// detects it. Without this fixture, the real-file test could pass
// by happenstance (today's files all happen to use a valid value)
// without exercising the reject path, and a future regression that
// re-introduces an invalid value would silently get through. This
// mirrors the wrapper-fixture pattern that
// TestMigrateWrapperRejectsMissingAtlasSum in atlas_tooling_test.go
// already uses for the same reason.
func TestAtlasTxmodeFixtureRejectsTransactionValue(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "fixture.sql")
	if err := os.WriteFile(path, []byte("-- atlas:txmode transaction\nSELECT 1;\n"), 0o644); err != nil {
		t.Fatalf("write fixture: %v", err)
	}
	invalid, err := firstInvalidTxmodeInFile(path)
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}
	if invalid != "transaction" {
		t.Errorf("validator did not detect -- atlas:txmode transaction in the synthetic fixture (got invalid=%q, want %q); the real-file test would not actually catch a regression of this value", invalid, "transaction")
	}
}

// TestNoInlineUniqueCollidesWithExplicitUniqueIndex is the real-file
// guard for VOC-033-AC-04 / VOC-033-TEST-04. Parses every migration
// file for two patterns and fails if any (table, column) pair
// appears in both: (1) an inline UNIQUE column constraint inside a
// CREATE TABLE block, and (2) a non-partial, single-column explicit
// `CREATE UNIQUE INDEX` on the same table and column. The collision
// pattern is exactly what
// 20260725130002_voc030_p4_gamification_tables.sql had pre-fix:
// inline `user_id UNIQUE` on `streak_states` plus an explicit
// `CREATE UNIQUE INDEX streak_states_user_id_key ON streak_states
// (user_id)`. Postgres auto-creates a unique index named
// "streak_states_user_id_key" for the inline UNIQUE, and the
// explicit statement then collides with `relation
// "streak_states_user_id_key" already exists` (SQLSTATE 42P07) on
// the first apply against an empty database. Multi-column and
// partial (`WHERE ...`) unique indexes are deliberately NOT flagged:
// this migration set already uses them correctly elsewhere (e.g.
// `confidence_point_ledger_user_id_idempotency_key_key`,
// `user_words_user_id_meaning_id_active_key`), and a false
// positive on either pattern would itself be a bug in this test,
// not a migration defect.
func TestNoInlineUniqueCollidesWithExplicitUniqueIndex(t *testing.T) {
	collisions, err := inlineUniqueCollisionsInDir(".")
	if err != nil {
		t.Fatalf("find collisions: %v", err)
	}
	keys := make([]tableColumn, 0, len(collisions))
	for k := range collisions {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool {
		if keys[i].table != keys[j].table {
			return keys[i].table < keys[j].table
		}
		return keys[i].col < keys[j].col
	})
	for _, k := range keys {
		autoName := k.table + "_" + k.col + "_key"
		t.Errorf("(table=%q, column=%q) has both an inline UNIQUE column constraint and a non-partial single-column explicit CREATE UNIQUE INDEX; Postgres auto-creates a unique index named %q for the inline UNIQUE, and the explicit statement then collides with `relation %q already exists` (SQLSTATE 42P07) on the first apply against an empty database", k.table, k.col, autoName, autoName)
	}
}

// TestInlineUniqueCollisionFixtureIsDetected is the synthetic-fixture
// half of VOC-033-AC-04's "test is proven to actually catch the
// regression" requirement. It writes a single SQL file to
// t.TempDir() that reproduces the exact `streak_states` collision
// shape this package fixes (inline `user_id UNIQUE` on
// `streak_states` plus an explicit `CREATE UNIQUE INDEX
// streak_states_user_id_key ON streak_states (user_id)`) and
// asserts the same inlineUniqueCollisionsInDir helper used by the
// real-file test detects the collision. Without this fixture, the
// real-file test could pass by happenstance after VOC-033-T00's
// deletion (today's files happen to have no collision) without
// exercising the detect path, and a future migration that re-
// introduces the same shape would silently get through. Mirrors the
// fixture pattern TestAtlasTxmodeFixtureRejectsTransactionValue
// uses for the txmode defect class.
func TestInlineUniqueCollisionFixtureIsDetected(t *testing.T) {
	const fixture = `
CREATE TABLE streak_states (
  id uuid PRIMARY KEY,
  user_id uuid NOT NULL UNIQUE REFERENCES users(id) ON DELETE RESTRICT
);

CREATE UNIQUE INDEX streak_states_user_id_key
  ON streak_states (user_id);
`
	dir := t.TempDir()
	path := filepath.Join(dir, "fixture.sql")
	if err := os.WriteFile(path, []byte(fixture), 0o644); err != nil {
		t.Fatalf("write fixture: %v", err)
	}
	collisions, err := inlineUniqueCollisionsInDir(dir)
	if err != nil {
		t.Fatalf("find collisions in fixture: %v", err)
	}
	want := tableColumn{table: "streak_states", col: "user_id"}
	if !collisions[want] {
		t.Errorf("validator did not detect the synthetic (table=%q, column=%q) collision in the fixture; the real-file test would not actually catch a regression of this pattern", want.table, want.col)
	}
}
