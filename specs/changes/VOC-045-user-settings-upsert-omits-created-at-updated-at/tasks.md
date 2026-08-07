# VOC-045 — Tasks

## VOC-045-T00 — Fix CompleteOnboarding's user_settings INSERT to supply created_at/updated_at

- Requirement source: `specification.md`'s scope; issue #341's identification
  of `apps/api/business/users/postgres.go`'s `CompleteOnboarding`
- Acceptance criteria: `VOC-045-AC-00`
- Tests: `VOC-045-TEST-00`
- Evidence: `VOC-045-EV-00`
- Status: pending

Add `created_at` and `updated_at` to the `user_settings` INSERT's column list
and VALUES in `CompleteOnboarding`, using the `now time.Time` parameter
already in scope for this function (the same `now` already used for the
`users` UPDATE and the `user_onboarding_profiles` INSERT in the same
transaction). Do not change the existing `ON CONFLICT DO UPDATE` branch's
`CASE`-guarded `daily_review_target` logic or its `updated_at = NOW()` clause.

## VOC-045-T01 — Fix gamification's UpsertUserSettings INSERT to supply created_at/updated_at

- Requirement source: `specification.md`'s scope and open question 1;
  issue #341's identification of
  `apps/api/business/gamification/repository.go`'s `UpsertUserSettings`
- Acceptance criteria: `VOC-045-AC-01`
- Tests: `VOC-045-TEST-01`
- Evidence: `VOC-045-EV-01`
- Status: pending

Add `created_at` and `updated_at` to the `user_settings` INSERT's column list
and VALUES in `UpsertUserSettings`. Resolve `specification.md`'s open
question 1 first: either use the SQL `NOW()` function directly in the
INSERT's VALUES (consistent with this function's own existing
`ON CONFLICT DO UPDATE ... SET updated_at = NOW()` clause, no signature
change), or thread a `now time.Time` parameter through `UpsertUserSettings`
and its caller (`apps/api/business/gamification/service.go`'s call at
service.go line 58) for consistency with `CompleteOnboarding`'s pattern.
Record which approach was taken and why as this task's evidence. Do not
change the existing `ON CONFLICT DO UPDATE` branch's `timezone`/
`daily_review_target` preservation logic.

## VOC-045-T02 — Add regression coverage and confirm no other user_settings INSERT site shares this defect

- Requirement source: `specification.md`'s scope; issue #341's suggested
  fix ("confirm via ... a new regression test that a genuinely fresh
  user_settings INSERT ... succeeds for both call sites")
- Acceptance criteria: `VOC-045-AC-02`, `VOC-045-AC-03`
- Tests: `VOC-045-TEST-02`
- Evidence: `VOC-045-EV-02`, `VOC-045-EV-03`
- Status: pending

Add a deterministic test (implementer to identify the best-fit location per
each package's existing test conventions, e.g. alongside
`apps/api/business/users/postgres.go`'s and
`apps/api/business/gamification/repository.go`'s existing repository-level
tests, or `apps/api/migrations/migration_test.go` if this repository already
runs integration-style tests there) that exercises each call site's INSERT
branch specifically (a user with no prior `user_settings` row, not the
`ON CONFLICT` path) and confirms it fails against the pre-`T00`/`T01` code
(or is written failing-first) and passes after the fix. Also run a
repository-wide search for `INSERT INTO user_settings` and record the full
call-site list as evidence, confirming each is either already correct or
covered by `T00`/`T01`.
