# VOC-046-EV-00 — T00 real-Postgres confirmation

Evidence for `VOC-046-AC-00` / `VOC-046-TEST-00`. This document records the
real-Postgres verification `tasks.md`'s `VOC-046-T00` requires ("Verify
against a real Postgres instance ... inspection alone is not sufficient
verification here"), which the sqlmock unit test alone cannot provide.

## The fix under test

`apps/api/business/missions/repository.go`'s `CreateDailyMissionSnapshot`
now supplies `created_at` and `updated_at` in its `daily_mission_snapshots`
INSERT column list and `VALUES` (`NOW(), NOW()`), matching this file's
existing `updated_at = NOW()` convention and VOC-045-T01's gamification
precedent. `Repository` has no injected clock to thread, so SQL `NOW()` is
the in-scope timestamp source, as `specification.md` permits ("sourced from
whatever `time.Time`/clock value is already in scope at this call site").

The pre-existing `ON CONFLICT (user_id, local_date) DO UPDATE` branch is
untouched.

## Regression coverage added

`apps/api/business/missions/repository_postgres_integration_test.go` (new,
build tag `integration`) starts a disposable `postgres:16-alpine` container
bound to loopback only, applies the committed forward migrations from
`apps/api/migrations/` in version order, and then:

- `TestCreateDailyMissionSnapshotFreshInsertAgainstRealPostgres` — the
  `VOC-046-TEST-00` precondition is asserted first (the user has no
  `daily_mission_snapshots` row for the target local date), so the call
  genuinely exercises the fresh-insert branch. It then drives
  `CreateDailyMissionSnapshot` and reads the persisted `created_at`/
  `updated_at` back out, into non-pointer `time.Time` values (a null in
  either column fails the scan outright) and inside the window bracketing
  the insert, measured against Postgres' own clock.
- `TestCreateDailyMissionSnapshotOnConflictBranchUnchangedAgainstRealPostgres`
  — `test-plan.md`'s negative coverage: a second call for the same
  `(user_id, local_date)` still updates rather than duplicating, preserves
  the original `created_at`, and advances `updated_at`.
- `TestGetDailyMissionViewForUserWithNoSnapshotAgainstRealPostgres` —
  `VOC-046-AC-00`'s service-level clause, exercising the same
  lazy-snapshot-creation read path `GET /api/v1/daily-mission` serves. See
  "Third occurrence discovered" below for why this test's assertion is
  written the way it is.

`apps/api/business/missions/repository_test.go`'s sqlmock expectation was
also made whitespace-tolerant (`dailyMissionSnapshotInsertColumnsPattern`),
matching VOC-045's `userSettingsInsertColumnsPattern` convention, so it
pins the required column list rather than the SQL literal's indentation.

### Build-tag trade-off

The integration test is gated behind `-tags=integration` and is therefore
not part of the default `go test ./...` path. This is the same deliberate
trade-off VOC-033-D02 recorded for
`apps/api/migrations/atlas_apply_integration_test.go`: adding a Postgres
service container to the shared CI workflow is not possible from this
repository (the workflow lives in `KARSIFT/karsift-ai-infra`) and is out of
`VOC-046-T00`'s declared scope. The test skips cleanly when Docker is
absent, so it is safe to leave in the tree. Wiring it into CI permanently
would need its own separately-scoped package.

## Failing-first verification (pre-fix code)

Verified directly against real Postgres, not asserted. With `T00`'s
`created_at, updated_at` columns temporarily removed from the production
statement (and then restored byte-for-byte — confirmed by an empty
`git diff` of `repository.go` against the restored state):

```bash
cd apps/api && go test -tags=integration -count=1 -v ./business/missions/... -run 'AgainstRealPostgres'
--- FAIL: TestCreateDailyMissionSnapshotFreshInsertAgainstRealPostgres (1.87s)
    Error: Received unexpected error:
           scan snapshot: pq: null value in column "created_at" of relation "daily_mission_snapshots" violates not-null constraint
    Messages: fresh insert must not raise a NOT NULL violation on created_at/updated_at (issue #352)
--- FAIL: TestCreateDailyMissionSnapshotOnConflictBranchUnchangedAgainstRealPostgres (2.14s)
    Error: scan snapshot: pq: null value in column "created_at" of relation "daily_mission_snapshots" violates not-null constraint
--- FAIL: TestGetDailyMissionViewForUserWithNoSnapshotAgainstRealPostgres (2.09s)
    Error: "scan snapshot: pq: null value in column \"created_at\" of relation \"daily_mission_snapshots\" violates not-null constraint" should not contain "relation \"daily_mission_snapshots\" violates not-null constraint"
    Messages: the read path must never again fail on daily_mission_snapshots' created_at/updated_at (issue #352, VOC-046-T00)
FAIL	github.com/KARSIFT/vocanova-platform/apps/api/business/missions	6.103s
```

This is the actual Postgres error issue #352 diagnosed textually and that
the API's generic `500` handler masked in production — reproduced here
against the real schema, which is what `VOC-046-AC-00` requires beyond
inspection.

## Passing verification (post-fix code)

```bash
cd apps/api && go test -tags=integration -count=1 -v ./business/missions/... -run 'AgainstRealPostgres'
--- PASS: TestCreateDailyMissionSnapshotFreshInsertAgainstRealPostgres (2.11s)
--- PASS: TestCreateDailyMissionSnapshotOnConflictBranchUnchangedAgainstRealPostgres (2.15s)
--- PASS: TestGetDailyMissionViewForUserWithNoSnapshotAgainstRealPostgres (2.11s)
ok  	github.com/KARSIFT/vocanova-platform/apps/api/business/missions	6.378s
```

The full default API suite is also green with no other package affected:

```bash
cd apps/api && go test -count=1 ./...
ok  	github.com/KARSIFT/vocanova-platform/apps/api/app/api	0.256s
ok  	github.com/KARSIFT/vocanova-platform/apps/api/business/gamification	0.029s
ok  	github.com/KARSIFT/vocanova-platform/apps/api/business/missions	0.009s
ok  	github.com/KARSIFT/vocanova-platform/apps/api/business/users	0.007s
# ...all other packages ok / no test files; no FAIL lines
```

## Third occurrence discovered: `streak_states` (VOC-046-T02's, not T00's)

Running `GET /api/v1/daily-mission`'s read path against real Postgres —
which the previous, mock-only revision of this task could not do — surfaced
a third instance of this same bug class on the same endpoint:

```
reconcile streak: upsert streak state: pq: null value in column "created_at"
of relation "streak_states" violates not-null constraint
```

Root cause: `apps/api/business/gamification/repository.go`'s
`UpsertStreakState` (line 307 at implementation time) omits
`created_at`/`updated_at` from its `INSERT INTO streak_states` column list,
while
`apps/api/migrations/20260725130002_voc030_p4_gamification_tables.sql`
declares both `timestamptz NOT NULL` with no `DEFAULT`. `GetDailyMissionView`
runs read-time streak reconciliation inside its lazy-snapshot-creation
branch, so this fires for exactly the same user population issue #352
describes.

This call site is **not** `VOC-046-T00`'s. It is explicitly named in
`tasks.md`'s `VOC-046-T02` scope ("`apps/api/business/gamification/
repository.go` (statements beyond the one already fixed by VOC-045-T01)"),
so fixing it here would be scope expansion into another task's work, which
`AGENTS.md` forbids. It is recorded here as evidence and in the
implementation PR so `T02` picks it up.

Consequence for `VOC-046-AC-00`: T00's own observable outcome — the
`daily_mission_snapshots` INSERT no longer raises a `NOT NULL` violation on
a fresh insert, confirmed against real Postgres — is met. The criterion's
further clause that `GET /api/v1/daily-mission` returns success end-to-end
cannot be fully satisfied by `T00` alone, because a distinct, separately-
owned defect on the same endpoint still blocks it. The service-level test
above is therefore written to assert exactly that boundary: the read path
must never again fail on `daily_mission_snapshots` (enforced
unconditionally, T00's own guarantee), and the only tolerated failure is
the T02-owned `streak_states` one. When `T02` lands, the test tightens
automatically to the full success path with no edit needed. The reviewing
human should treat `AC-00`'s endpoint clause as satisfied at the package
level once `T02` merges, not at `T00`'s.
