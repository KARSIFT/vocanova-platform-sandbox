# VOC-045 — Test Plan

## VOC-045-TEST-00 — CompleteOnboarding succeeds for a genuinely new user

- Covers: `VOC-045-AC-00`
- Preconditions: a test database with a `users` row for a test user and no
  corresponding `user_settings` row.
- Procedure: call `CompleteOnboarding` for that user with representative
  `OnboardingAnswers` and a fixed `now`. Before the fix, this must fail with
  the `pq: null value in column "created_at" ...` (or `"updated_at"`) error
  the issue reports; after the fix, it must succeed.
- Expected result: no error; the resulting `user_settings` row has
  `created_at` and `updated_at` both equal to the supplied `now` (not null,
  not a database-default fallback that diverges from the transaction's own
  timestamp).
- Evidence: `VOC-045-EV-00`

## VOC-045-TEST-01 — UpsertUserSettings succeeds for a genuinely new user

- Covers: `VOC-045-AC-01`
- Preconditions: a test database with a `users` row for a test user and no
  corresponding `user_settings` row.
- Procedure: call `UpsertUserSettings` with representative `timezone`/
  `dailyReviewTarget` values inside a test transaction. Before the fix, this
  must fail with the same class of `NOT NULL` violation; after the fix, it
  must succeed.
- Expected result: no error; the resulting `UserSettingsRow` reflects the
  supplied values, and the underlying row has non-null `created_at`/
  `updated_at`.
- Evidence: `VOC-045-EV-01`

## VOC-045-TEST-02 — Fresh-insert regression coverage and call-site inventory

- Covers: `VOC-045-AC-02`, `VOC-045-AC-03`
- Preconditions: same as `VOC-045-TEST-00`/`01` (no prior `user_settings`
  row for the test user in each case); a checkout of the repository at the
  revision under test for the grep step.
- Procedure: run `VOC-045-TEST-00` and `VOC-045-TEST-01` (or equivalent
  fresh-insert-path tests co-located with each fix) as part of the normal
  test suite so they run on every future change, not just this one; run a
  repository-wide search for `INSERT INTO user_settings` and record every
  match.
- Expected result: both fresh-insert tests pass after the fix and are
  confirmed to fail (or are documented as having been verified failing
  first) against the pre-fix code; the call-site inventory lists exactly the
  two known sites (or more, if the grep finds an additional one, in which
  case it must also be fixed or explicitly deferred with the reviewing
  human's sign-off before this task is considered complete).
- Evidence: `VOC-045-EV-02`, `VOC-045-EV-03`

## Negative and failure coverage

The `ON CONFLICT DO UPDATE` branch's existing behavior (updating a row that
already exists, preserving a customized `daily_review_target` via its
existing `CASE` guard) is not touched by this fix and is expected to
continue passing any existing test coverage for that branch unchanged — this
package does not add new tests for that branch, only for the previously
untested/failing fresh-insert branch.

## Rollback and migration coverage

Not applicable: this package's primary fix introduces no migration. If the
reviewing human adopts the schema-level `DEFAULT now()` alternative
(`specification.md`'s open question 2), that decision must add its own
migration-specific test coverage (e.g. via
`apps/api/migrations/migration_test.go`'s existing invariant-checking
pattern) before being considered complete — not assumed here.

No secret or production data is used by any test in this plan.
