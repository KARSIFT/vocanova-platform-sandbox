# VOC-046 — Test Plan

## VOC-046-TEST-00 — CreateDailyMissionSnapshot's fresh-insert path succeeds

- Covers: `VOC-046-AC-00`
- Preconditions: a test user with no existing `daily_mission_snapshots` row
  for their current local date, against a real (test/CI) Postgres instance
  running the current migration set, not a mock.
- Procedure: call `CreateDailyMissionSnapshot` (or the higher-level
  `GetDailyMissionView` path that exercises it) for that user; confirm the
  test fails against pre-fix code with a Postgres `NOT NULL` violation on
  `created_at` or `updated_at` (failing-first), then confirm it passes
  after `VOC-046-T00`'s fix, with the resulting row's `created_at` and
  `updated_at` both non-null and reasonable (e.g. equal to the clock value
  used at insert time).
- Expected result: insert succeeds; `GET /api/v1/daily-mission` (or the
  service-layer equivalent) returns success, not a `500`.
- Evidence: `VOC-046-EV-00`

## VOC-046-TEST-01 — daily_activity_summaries fresh-insert paths in the same file succeed

- Covers: `VOC-046-AC-01`
- Preconditions: a test user with no existing `daily_activity_summaries`
  row for their current local date, against a real Postgres instance.
- Procedure: for each `INSERT INTO daily_activity_summaries` call site in
  `apps/api/business/missions/repository.go` found to need a fix by
  `VOC-046-T01`, exercise its fresh-insert path and confirm failing-first
  against pre-fix code, then passing after the fix. For any call site found
  to already be correct, confirm (and record) that its existing test
  coverage, if any, already exercises the fresh-insert path — add coverage
  if it does not.
- Expected result: every `daily_activity_summaries` insert in this file
  succeeds on a genuine first write, with non-null `created_at`/`updated_at`.
- Evidence: `VOC-046-EV-01`

## VOC-046-TEST-02 — Every audited file's fixed call sites succeed on fresh insert

- Covers: `VOC-046-AC-02`, `VOC-046-AC-03`
- Preconditions: for each call site `VOC-046-T02` finds and fixes, a test
  state with no prior row for the relevant unique key, against a real
  Postgres instance.
- Procedure: for each fixed call site, exercise its fresh-insert path,
  confirming failing-first against pre-fix code and passing after the fix,
  following `VOC-045-T02`'s convention. Cross-check the audit trail
  (`VOC-046-EV-03`) against the actual diff to confirm every recorded
  "fixed" outcome has corresponding test coverage and every recorded
  "already correct" outcome has a stated, verifiable reason.
- Expected result: every call site this package's audit identifies as
  needing a fix has both the fix and a corresponding failing-first-then-
  passing regression test; the audit trail is complete and matches the
  actual code changes.
- Evidence: `VOC-046-EV-02`, `VOC-046-EV-03`, `VOC-046-EV-04`

## VOC-046-TEST-03 — Schema-scanning detection check catches a reintroduced instance (conditional)

- Covers: `VOC-046-AC-04`
- Preconditions: `VOC-046-T03` is dispatched (see its conditional status in
  `tasks.md`).
- Procedure: deliberately reintroduce a scratch `INSERT INTO` statement
  against a `NOT NULL`-no-`DEFAULT` table that omits `created_at`, and
  confirm the new check fails; then remove the scratch statement and confirm
  the check passes against the repository's actual, fully-fixed state.
- Expected result: the check reliably fails on a reintroduced instance of
  this exact bug class and passes otherwise, providing a check equivalent in
  spirit to what would have caught both the `user_settings` and
  `daily_mission_snapshots` occurrences before merge.
- Evidence: `VOC-046-EV-05`

## Negative, authorization, and rollback coverage

Negative: for each fixed call site, existing `ON CONFLICT DO UPDATE` (or
equivalent) behavior for a user who already has a row for the relevant key
must be confirmed unchanged by this package's diff — the defect and its fix
are specific to the fresh-insert branch, and no test in this plan should
require or accept a change to the existing update-branch behavior.
Authorization: not applicable — this fix does not change any
authentication or authorization check.
Rollback: confirm, after a hypothetical revert of this package's diff, that
each fixed call site returns to its exact pre-fix state (the same defect
issue #352 and its audit report today), not some new intermediate state,
matching `implementation-plan.md`'s rollback section.
No test in this plan uses secrets or production data; all tests run against
disposable test/CI Postgres state.
