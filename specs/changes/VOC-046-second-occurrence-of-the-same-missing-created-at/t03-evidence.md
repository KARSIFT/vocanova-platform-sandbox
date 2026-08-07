# VOC-046-EV-05 — T03 schema-scanning detection check evidence

`VOC-046-T03` adds `apps/api/migrations/insert_required_columns_test.go`, a
deterministic Go test that parses every committed migration's `CREATE TABLE`
bodies for `NOT NULL`-no-`DEFAULT` columns, scans every non-`_test.go` Go
source file under `apps/api` for `INSERT INTO <table> (...)` statements, and
fails when a discovered statement's column list omits one of those columns.
Statements that omit an explicit column list entirely are also reported,
because their required-column coverage cannot be verified by scanning.

The check runs in the default `go test ./...` path (no build tag, no
database, no external service), so it gates every future change the same way
CI already gates the rest of `apps/api`.

## VOC-046-TEST-03 — reintroduced instance is caught

The check's fixture test, `TestInsertCoverageFixtureDetectsMissingRequiredColumn`,
is the permanent, self-contained form of the test-plan procedure: it writes a
throwaway migration declaring `fixture_rows.created_at`/`updated_at` as
`NOT NULL` with no default plus a throwaway Go source file whose `INSERT INTO
fixture_rows` omits `created_at`, runs the same scanner over that tree, and
fails if no violation is reported. This keeps the "would it actually catch a
regression?" question answered on every CI run rather than only at
implementation time.

The manual reintroduction the test plan describes was also performed against
the real repository tree. Removing `created_at, updated_at` from
`MarkSnapshotMissed`'s `INSERT INTO daily_mission_snapshots` and re-running:

```
$ cd apps/api && go test ./migrations/ -run TestInsertStatementsProvideMigrationRequiredColumns
--- FAIL: TestInsertStatementsProvideMigrationRequiredColumns (0.05s)
    insert_required_columns_test.go:44: .../apps/api/business/missions/repository.go:450 INSERT INTO daily_mission_snapshots is missing migration-required column "created_at"
    insert_required_columns_test.go:44: .../apps/api/business/missions/repository.go:450 INSERT INTO daily_mission_snapshots is missing migration-required column "updated_at"
FAIL	github.com/KARSIFT/vocanova-platform/apps/api/migrations	0.048s
```

Restoring both columns and re-running the same command:

```
ok  	github.com/KARSIFT/vocanova-platform/apps/api/migrations	0.048s
```

This is the `VOC-046-AC-04` proof that the check would have caught both the
`user_settings` (VOC-045) and `daily_mission_snapshots` (issue #352)
occurrences before merge: both are exactly the shape of the failure above.

## A real defect the check found: MarkSnapshotMissed

Running the new check against the repository state left by `T00`, `T01`, and
`T02` did not pass — it reported a genuine, still-unfixed instance of this
bug class at `apps/api/business/missions/repository.go:450`,
`MarkSnapshotMissed`'s `INSERT INTO daily_mission_snapshots`. That call site
fell in the seam between this package's tasks: `T00` fixed
`CreateDailyMissionSnapshot`, `T01` was scoped to
`daily_activity_summaries`, and `T02` was scoped to the twelve files other
than `missions/repository.go`. `implementation-plan.md`'s file
reconciliation named this statement ("one further `INSERT INTO
daily_mission_snapshots` statement for the missed-mission path (lines
450-456 at drafting time)") as needing its own cross-check, so it is inside
this package's declared scope and inside a file the package already
discloses as touched.

Fix applied (same pattern as `T00`, same file, `NOW()` for both columns, no
change to the existing `ON CONFLICT DO UPDATE` branch, which already set
`updated_at` and never set `created_at`):

- `apps/api/business/missions/repository.go` — `MarkSnapshotMissed`'s column
  list and `VALUES` now supply `created_at, updated_at`.

Note that the pre-existing `ON CONFLICT DO UPDATE ... SET ... updated_at =
NOW()` clause is exactly why this defect survived review three times in this
bug class: it makes the statement *look* like it maintains its timestamps
while doing nothing for the first-insert branch that the `NOT NULL`
constraints actually apply to.

## Regression coverage for the fixed call site

`TestPostgreSQLRepositoryMarkSnapshotMissedSuppliesTimestamps` in
`apps/api/business/missions/repository_test.go` asserts the fresh-insert
column list includes `created_at, updated_at`, following the same sqlmock
column-pattern convention `T00` established for
`CreateDailyMissionSnapshot`. It fails against the pre-fix statement (the
expected `INSERT` pattern does not match, so sqlmock's expectation is unmet)
and passes after the fix. The scanning check itself is the second,
schema-derived layer of coverage for the same call site — the failing-first
run recorded above is that proof.

No real-Postgres integration test was added for this call site: the existing
`apps/api/business/missions/repository_postgres_integration_test.go` proofs
are gated behind the `integration` build tag and cannot run in this
repository's CI (see that file's header for the recorded trade-off), whereas
the schema-scanning check added by this task runs unconditionally and derives
its expectation from the committed migration rather than from a hand-written
assertion.

## Commands run

```
cd apps/api && go test ./migrations/ ./business/missions/
git diff --check
```

Both clean at the reviewed revision.

## Follow-ups noted, not fixed here

- The check reports only missing columns in explicit column lists and
  `INSERT INTO <table> VALUES (...)` statements without a column list. It
  does not parse dynamically-assembled SQL (string concatenation or query
  builders); no such `INSERT` exists under `apps/api` today, but a future one
  would be invisible to it.
- The scanner is textual, not a full SQL parser. A `CREATE TABLE` written in
  a style the line-oriented column parser does not recognise (e.g. an entire
  table definition on one line) would silently contribute no required
  columns. Every committed migration uses the one-column-per-line style the
  parser expects.
