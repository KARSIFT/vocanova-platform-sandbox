# VOC-045 — Acceptance Criteria

## VOC-045-AC-00 — A genuinely new user completes onboarding without a NOT NULL violation

- Requirement source: issue #341's live production reproduction
- Tasks: `VOC-045-T00`
- Tests: `VOC-045-TEST-00`
- Evidence: `VOC-045-EV-00`
- Result: pending
- Observable outcome: for a user with no prior `user_settings` row,
  `apps/api/business/users/postgres.go`'s `CompleteOnboarding` INSERT into
  `user_settings` succeeds (no `pq: null value in column "created_at" ...`
  or `"updated_at"` error), and the onboarding request returns success
  (not the `400` `mapOnboardingError`'s fallback currently produces).

## VOC-045-AC-01 — Gamification's own lazy-creation upsert also succeeds on a genuine first insert

- Requirement source: issue #341's identification of
  `apps/api/business/gamification/repository.go`'s `UpsertUserSettings` as
  sharing the same defect
- Tasks: `VOC-045-T01`
- Tests: `VOC-045-TEST-01`
- Evidence: `VOC-045-EV-01`
- Result: pending
- Observable outcome: for a user with no prior `user_settings` row,
  `UpsertUserSettings`'s INSERT succeeds (no `NOT NULL` violation on
  `created_at`/`updated_at`) and returns the expected `UserSettingsRow`.

## VOC-045-AC-02 — Deterministic regression coverage for both call sites' fresh-insert path

- Requirement source: issue #341's suggested fix ("confirm via a new
  regression test that a genuinely fresh user_settings INSERT ... succeeds
  for both call sites")
- Tasks: `VOC-045-T02`
- Tests: `VOC-045-TEST-02`
- Evidence: `VOC-045-EV-02`
- Result: pending
- Observable outcome: a deterministic test exists that exercises each call
  site's INSERT branch specifically (not the `ON CONFLICT` update branch)
  against a database state with no prior `user_settings` row for the test
  user, and fails against the pre-fix code (or is confirmed failing-first)
  and passes against the post-fix code.

## VOC-045-AC-03 — No other user_settings INSERT call site shares this defect

- Requirement source: this package's own in-scope repository-wide grep (see
  `specification.md`'s "Scope and non-goals")
- Tasks: `VOC-045-T02`
- Tests: n/a (inventory, not a runtime test)
- Evidence: `VOC-045-EV-03`
- Result: pending
- Observable outcome: a recorded, complete list of every
  `INSERT INTO user_settings` call site in the repository, with each one
  confirmed to either already supply `created_at`/`updated_at` or be covered
  by this package's fix.
