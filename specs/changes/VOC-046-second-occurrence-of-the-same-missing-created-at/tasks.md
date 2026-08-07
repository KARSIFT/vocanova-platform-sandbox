# VOC-046 — Tasks

## VOC-046-T00 — Fix CreateDailyMissionSnapshot's INSERT to supply created_at/updated_at

- Requirement source: issue #352's confirmed root cause and immediate
  suggested fix
- Acceptance criteria: `VOC-046-AC-00`
- Tests: `VOC-046-TEST-00`
- Evidence: `VOC-046-EV-00`
- Status: pending

Add `created_at` and `updated_at` to
`apps/api/business/missions/repository.go`'s `CreateDailyMissionSnapshot`
INSERT column list and VALUES for `daily_mission_snapshots`, sourced from
whatever `time.Time`/clock value is already in scope at this call site (or
threaded in if not currently available), matching this repository's
existing `s.clock.Now()`-based convention and VOC-045-T00's approach.
Verify against a real Postgres instance (this issue's own note that the
API's generic 500 handling masks the real error means inspection alone is
not sufficient verification here). Priority: `P0`, given the confirmed live
production outage — this task should be dispatched first among this
package's tasks.

## VOC-046-T01 — Audit and fix daily_activity_summaries INSERTs in the same file

- Requirement source: issue #352's note that `daily_activity_summaries`
  shares the identical NOT-NULL-no-DEFAULT schema and should be checked in
  the same file/module as the confirmed defect
- Acceptance criteria: `VOC-046-AC-01`
- Tests: `VOC-046-TEST-01`
- Evidence: `VOC-046-EV-01`
- Status: pending

Check every `INSERT INTO daily_activity_summaries` statement in
`apps/api/business/missions/repository.go` (multiple exist at drafting
time, per `implementation-plan.md`) against the `NOT NULL`-no-`DEFAULT`
`created_at`/`updated_at` columns declared in
`apps/api/migrations/20260725130001_voc030_p4_mission_tables.sql`. Fix any
found missing either column, following the same pattern as `VOC-046-T00`.
Record each statement's outcome explicitly (fixed, or already correct and
why), not silently.

## VOC-046-T02 — Repository-wide audit of the remaining named files for the same omission

- Requirement source: issue #352's "Scope: this is systemic, not isolated
  to two tables" section and its full thirteen-file list
- Acceptance criteria: `VOC-046-AC-02`, `VOC-046-AC-03`
- Tests: `VOC-046-TEST-02`
- Evidence: `VOC-046-EV-02`, `VOC-046-EV-03`, `VOC-046-EV-04`
- Status: pending

For each of `apps/api/business/accounts/auth.go`,
`apps/api/business/accounts/postgres.go`,
`apps/api/business/accounts/service.go`,
`apps/api/business/aifeedback/postgres.go`,
`apps/api/business/auth/postgres.go`,
`apps/api/business/gamification/repository.go` (statements beyond the one
already fixed by VOC-045-T01),
`apps/api/business/gamification/rewards.go`,
`apps/api/business/gamification/streak.go`,
`apps/api/business/learning/postgres.go`,
`apps/api/business/learning/postgres_idempotency.go`,
`apps/api/business/reviews/postgres.go`,
`apps/api/business/users/postgres.go` (statements beyond the two already
fixed by VOC-045-T00), and `apps/api/business/users/seed.go`: locate every
`INSERT INTO` statement, identify its target table, cross-reference that
table's migration-declared `NOT NULL`-no-`DEFAULT` columns (across all 13
files in `apps/api/migrations/`), and fix any statement found missing such
a column, adding deterministic failing-first-then-passing regression
coverage for each fix (mirroring `VOC-045-T02`'s convention). Record a
complete audit trail listing every file, every `INSERT INTO` statement
found in it, and its outcome (fixed / already correct and why / not
applicable and why), per `specification.md`'s open question 1's resolution
that "no defect found" is a valid, explicitly-recorded completion state,
not a gap to be silently skipped. If a defect is found outside the named
thirteen files during this audit, record it as evidence and flag it as a
follow-up per `specification.md`'s scope boundary, rather than silently
expanding this task's fix set.

## VOC-046-T03 — Schema-scanning detection check (conditional on adoption decision)

- Requirement source: issue #352's suggested "(new) a schema-level or
  lint-style check if practical"
- Acceptance criteria: `VOC-046-AC-04`
- Tests: `VOC-046-TEST-03`
- Evidence: `VOC-046-EV-05`
- Status: pending, dispatch conditional on the reviewing human resolving
  `specification.md`'s open question 2 in favor of including it in this
  package's scope

Only if adopted in scope: add a deterministic test or check that walks each
migration file's declared `NOT NULL`-no-`DEFAULT` columns per table, scans
the Go source tree for `INSERT INTO` statements against those tables, and
fails if any discovered statement's column list omits a `NOT NULL`-no-
`DEFAULT` column. Verify it fails against a deliberately-reintroduced
instance of this bug class (e.g. a scratch statement missing
`created_at`) before confirming it passes against the fully-fixed
repository state. If the reviewing human instead defers this to a separate
follow-up package, this task is dropped from this package's scope entirely
rather than partially implemented.
