# VOC-046 — Acceptance Criteria

## VOC-046-AC-00 — A user with no mission yet today loads /home without a NOT NULL violation

- Requirement source: issue #352's live production reproduction
- Tasks: `VOC-046-T00`
- Tests: `VOC-046-TEST-00`
- Evidence: `VOC-046-EV-00`
- Result: pending
- Observable outcome: for a user with no existing `daily_mission_snapshots`
  row for their current local date,
  `apps/api/business/missions/repository.go`'s `CreateDailyMissionSnapshot`
  INSERT succeeds (no Postgres `NOT NULL` violation on `created_at` or
  `updated_at`), confirmed against a real Postgres instance (not just
  inspection), and `GET /api/v1/daily-mission` returns a successful response
  instead of the `500` the issue reports.

## VOC-046-AC-01 — daily_activity_summaries call sites in the same file are confirmed safe or fixed

- Requirement source: issue #352's note that `daily_activity_summaries`
  shares the identical NOT-NULL-no-DEFAULT schema and should be checked in
  the same file/module
- Tasks: `VOC-046-T01`
- Tests: `VOC-046-TEST-01`
- Evidence: `VOC-046-EV-01`
- Result: pending
- Observable outcome: every `INSERT INTO daily_activity_summaries` statement
  in `apps/api/business/missions/repository.go` is confirmed to either
  already supply `created_at`/`updated_at` or is fixed to do so, with each
  outcome recorded explicitly (not silently assumed).

## VOC-046-AC-02 — Every named file is audited and any other defect is fixed

- Requirement source: issue #352's systemic scope section and its list of
  thirteen files to audit
- Tasks: `VOC-046-T02`
- Tests: `VOC-046-TEST-02`
- Evidence: `VOC-046-EV-02`, `VOC-046-EV-03`
- Result: pending
- Observable outcome: a recorded, complete audit of every `INSERT INTO`
  statement in each of the thirteen files the issue names (excluding
  `_test.go` files), cross-referenced against each target table's
  migration-declared `NOT NULL`-no-`DEFAULT` columns, with every discovered
  omission fixed and every "no defect found" file explicitly justified (per
  `specification.md`'s open question 1) — not a partial or best-effort scan.

## VOC-046-AC-03 — Deterministic regression coverage exists for every call site fixed by this package

- Requirement source: issue #352's suggested fix ("Add regression coverage
  per call site the same way VOC-045-T02 did")
- Tasks: `VOC-046-T00`, `VOC-046-T01`, `VOC-046-T02`
- Tests: `VOC-046-TEST-00`, `VOC-046-TEST-01`, `VOC-046-TEST-02`
- Evidence: `VOC-046-EV-04`
- Result: pending
- Observable outcome: for every call site this package actually fixes, a
  deterministic test exists that exercises that call site's fresh-insert
  path (not an `ON CONFLICT` update branch, where one exists) against a
  database state with no prior row, and fails against the pre-fix code (or
  is confirmed failing-first) and passes against the post-fix code.

## VOC-046-AC-04 — A schema-scanning detection check, if adopted in scope

- Requirement source: issue #352's suggested "(new) a schema-level or
  lint-style check if practical"
- Tasks: `VOC-046-T03` (conditional — see `specification.md`'s open
  question 2)
- Tests: `VOC-046-TEST-03`
- Evidence: `VOC-046-EV-05`
- Result: pending, and only applicable if the reviewing human confirms
  `VOC-046-T03` is in scope at adoption
- Observable outcome: a deterministic check exists that walks each
  migration's `NOT NULL`-no-`DEFAULT` columns and confirms every `INSERT`
  statement discovered via source scanning supplies them, and that check
  fails against a deliberately-reintroduced instance of this exact bug
  class (proving it would have caught both the `user_settings` and
  `daily_mission_snapshots` occurrences before merge).
