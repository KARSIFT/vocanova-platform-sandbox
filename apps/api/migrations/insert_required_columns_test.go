package migrations_test

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"
)

var createTableHeaderRE = regexp.MustCompile(`(?i)^\s*CREATE\s+TABLE\s+(\w+)\s*\(`)
var insertWithColumnsRE = regexp.MustCompile(`(?is)INSERT\s+INTO\s+(\w+)\s*\(([^)]*)\)`)
var insertWithoutColumnsRE = regexp.MustCompile(`(?is)INSERT\s+INTO\s+(\w+)\s+VALUES\s*\(`)

type insertLocation struct {
	file string
	line int
}

func TestInsertStatementsProvideMigrationRequiredColumns(t *testing.T) {
	requiredColumnsByTable, err := loadNotNullNoDefaultColumns(".")
	if err != nil {
		t.Fatalf("load migration constraints: %v", err)
	}

	apiRoot, err := filepath.Abs("..")
	if err != nil {
		t.Fatalf("resolve apps/api root: %v", err)
	}

	violations, scannedInsertCount, err := scanInsertCoverage(apiRoot, requiredColumnsByTable)
	if err != nil {
		t.Fatalf("scan insert coverage: %v", err)
	}
	if scannedInsertCount == 0 {
		t.Fatalf("no INSERT INTO statements were discovered under %q; scanner misconfiguration", apiRoot)
	}
	if len(violations) > 0 {
		sort.Strings(violations)
		for _, v := range violations {
			t.Error(v)
		}
	}
}

func TestInsertCoverageFixtureDetectsMissingRequiredColumn(t *testing.T) {
	tempRoot := t.TempDir()
	migrationsDir := filepath.Join(tempRoot, "migrations")
	sqlPath := filepath.Join(migrationsDir, "20260101000000_fixture.sql")
	goFilePath := filepath.Join(tempRoot, "business", "fixture.go")

	if err := os.MkdirAll(filepath.Dir(goFilePath), 0o755); err != nil {
		t.Fatalf("mkdir fixture source dir: %v", err)
	}
	if err := os.MkdirAll(migrationsDir, 0o755); err != nil {
		t.Fatalf("mkdir fixture migrations dir: %v", err)
	}

	const fixtureMigration = `
CREATE TABLE fixture_rows (
  id uuid PRIMARY KEY,
  created_at timestamptz NOT NULL,
  updated_at timestamptz NOT NULL
);
`
	if err := os.WriteFile(sqlPath, []byte(fixtureMigration), 0o644); err != nil {
		t.Fatalf("write fixture migration: %v", err)
	}

	const fixtureSource = `package business

func brokenInsert() string {
	return "INSERT INTO fixture_rows (id, updated_at) VALUES ($1, $2)"
}
`
	if err := os.WriteFile(goFilePath, []byte(fixtureSource), 0o644); err != nil {
		t.Fatalf("write fixture source: %v", err)
	}

	requiredColumnsByTable, err := loadNotNullNoDefaultColumns(migrationsDir)
	if err != nil {
		t.Fatalf("load fixture migration constraints: %v", err)
	}
	violations, _, err := scanInsertCoverage(tempRoot, requiredColumnsByTable)
	if err != nil {
		t.Fatalf("scan fixture insert coverage: %v", err)
	}
	if len(violations) == 0 {
		t.Fatalf("expected fixture to produce at least one violation, but none were reported")
	}

	const expectedFragment = `is missing migration-required column "created_at"`
	for _, violation := range violations {
		if strings.Contains(violation, expectedFragment) {
			return
		}
	}
	t.Fatalf("fixture violations did not mention missing created_at: %v", violations)
}

func loadNotNullNoDefaultColumns(migrationsDir string) (map[string]map[string]bool, error) {
	files, err := filepath.Glob(filepath.Join(migrationsDir, "*.sql"))
	if err != nil {
		return nil, err
	}
	requiredColumnsByTable := map[string]map[string]bool{}
	for _, name := range files {
		sqlBytes, err := os.ReadFile(name)
		if err != nil {
			return nil, err
		}
		parseCreateTableRequiredColumns(string(sqlBytes), requiredColumnsByTable)
	}
	return requiredColumnsByTable, nil
}

func parseCreateTableRequiredColumns(sql string, requiredColumnsByTable map[string]map[string]bool) {
	lines := strings.Split(sql, "\n")
	var currentTable string
	var depth int
	for _, line := range lines {
		if currentTable == "" {
			matches := createTableHeaderRE.FindStringSubmatch(line)
			if len(matches) > 1 {
				currentTable = strings.ToLower(matches[1])
				depth = strings.Count(line, "(") - strings.Count(line, ")")
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
		columnName, isRequired := parseRequiredColumnFromLine(trimmed)
		if !isRequired {
			continue
		}
		requiredSet := requiredColumnsByTable[currentTable]
		if requiredSet == nil {
			requiredSet = map[string]bool{}
			requiredColumnsByTable[currentTable] = requiredSet
		}
		requiredSet[strings.ToLower(columnName)] = true
	}
}

func parseRequiredColumnFromLine(line string) (string, bool) {
	trimmed := strings.TrimSuffix(strings.TrimSpace(line), ",")
	if trimmed == "" {
		return "", false
	}

	fields := strings.Fields(trimmed)
	if len(fields) < 2 {
		return "", false
	}

	firstWord := strings.ToUpper(fields[0])
	switch firstWord {
	case "CONSTRAINT", "PRIMARY", "UNIQUE", "CHECK", "FOREIGN":
		return "", false
	}

	upper := strings.ToUpper(trimmed)
	if !strings.Contains(upper, "NOT NULL") {
		return "", false
	}
	if strings.Contains(upper, "DEFAULT") {
		return "", false
	}
	return fields[0], true
}

func scanInsertCoverage(scanRoot string, requiredColumnsByTable map[string]map[string]bool) ([]string, int, error) {
	var violations []string
	var scannedInsertCount int
	err := filepath.WalkDir(scanRoot, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() {
			name := d.Name()
			if name == ".git" || name == "node_modules" || name == "tmp" {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}

		source, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		text := string(source)

		matches := insertWithColumnsRE.FindAllStringSubmatchIndex(text, -1)
		scannedInsertCount += len(matches)
		for _, match := range matches {
			table := strings.ToLower(text[match[2]:match[3]])
			requiredSet := requiredColumnsByTable[table]
			if len(requiredSet) == 0 {
				continue
			}
			columnsRaw := text[match[4]:match[5]]
			columns := splitAndNormalizeColumns(columnsRaw)
			loc := insertLocation{file: path, line: lineNumberAt(text, match[0])}
			for requiredColumn := range requiredSet {
				if !columns[requiredColumn] {
					violations = append(violations, fmt.Sprintf("%s:%d INSERT INTO %s is missing migration-required column %q", loc.file, loc.line, table, requiredColumn))
				}
			}
		}

		noColumnMatches := insertWithoutColumnsRE.FindAllStringSubmatchIndex(text, -1)
		for _, match := range noColumnMatches {
			table := strings.ToLower(text[match[2]:match[3]])
			requiredSet := requiredColumnsByTable[table]
			if len(requiredSet) == 0 {
				continue
			}
			loc := insertLocation{file: path, line: lineNumberAt(text, match[0])}
			violations = append(violations, fmt.Sprintf("%s:%d INSERT INTO %s omits an explicit column list; cannot verify required columns %v", loc.file, loc.line, table, sortedKeys(requiredSet)))
		}
		return nil
	})
	return violations, scannedInsertCount, err
}

func splitAndNormalizeColumns(raw string) map[string]bool {
	columns := map[string]bool{}
	for _, part := range strings.Split(raw, ",") {
		trimmed := strings.TrimSpace(strings.Trim(part, `"`))
		if trimmed == "" {
			continue
		}
		columns[strings.ToLower(trimmed)] = true
	}
	return columns
}

func lineNumberAt(text string, byteIndex int) int {
	return strings.Count(text[:byteIndex], "\n") + 1
}

func sortedKeys(m map[string]bool) []string {
	keys := make([]string, 0, len(m))
	for key := range m {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}
