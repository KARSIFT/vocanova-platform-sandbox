# VOC-053 — Test Plan

## VOC-053-TEST-00 — Root-cause investigation produces direct, reproducible evidence

- Covers: `VOC-053-AC-00`
- Preconditions: Access to inspect real staging HTTP responses (headers,
  cache-status) for `/home` and its underlying data fetch, and read access to
  `apps/api/app/api/missions.go` (or wherever the live daily-mission `GET`
  handler is actually found).
- Procedure: Request `/home` against real staging twice, seconds apart, with
  no intervening review completed for the account under test; capture and
  compare response headers and bodies. Separately, trace the backend
  request-handling path from the HTTP handler down through
  `gamification.LocalDate` and `missions.Repository`'s read methods, noting
  exactly where/how `now` and the per-user timezone are resolved per request.
- Expected result: Either a clear caching signal (a cache-hit header, or the
  second request's body matching a stale first-request snapshot despite a
  known intervening state change) or a clear absence of one (both requests
  show `cache-control: no-store`/equivalent and reflect current backend
  state) — and either a clear timezone/local-date resolution defect in the
  traced backend path, or confirmation the traced path is correct as read.
  The investigation records whichever of these outcomes it actually found,
  not a predetermined conclusion.
- Evidence: `VOC-053-EV-00`

## VOC-053-TEST-01 — Fix targets the confirmed cause without touching the step 7 assertion

- Covers: `VOC-053-AC-01`
- Preconditions: `VOC-053-TEST-00` has confirmed a specific root cause.
- Procedure: Review the fix's diff against `tests/staging-e2e/
  core-loop.staging.spec.ts` (confirm no change to line ~375's assertion or
  its inputs' computation) and against the confirmed cause (confirm the fix
  addresses that specific mechanism, e.g. an explicit `cache: "no-store"` on
  the specific fetch identified, or a corrected timezone/local-date
  resolution in the specific handler traced).
- Expected result: The diff is scoped to the confirmed cause; the assertion
  and its surrounding step logic in the spec file are byte-identical to
  before the fix (unless the confirmed cause is itself in a spec-file
  helper, per candidate (c), in which case only that helper changes).
- Evidence: `VOC-053-EV-01`

## VOC-053-TEST-02 — Two same-day reads of the counter, seconds apart, do not decrease

- Covers: `VOC-053-AC-02`
- Preconditions: The fix from `VOC-053-T01` is deployed to a real or
  staging-equivalent environment; a test account with a known
  `reviewsCompleted` value for the current local day.
- Procedure: Read the daily-mission value (via the real `/home` page render
  and/or the underlying API response directly) twice, at least five seconds
  apart, performing no review action in between.
- Expected result: Both reads return the identical `reviewsCompleted` value
  — no decrease, matching the objective in `specification.md`.
- Evidence: `VOC-053-EV-01`

## VOC-053-TEST-03 — Real staging core-loop E2E step 7 passes under the original failure condition

- Covers: `VOC-053-AC-03`, `VOC-053-AC-04`
- Preconditions: `VOC-053-T01`'s fix is live in a real `deploy-staging.yml`
  run; the persistent synthetic account already has `reviewedBefore >= 1`
  residue from a prior run (the same condition run 31332238452 failed under)
  — not a freshly-reset account.
- Procedure: Run `tests/staging-e2e/core-loop.staging.spec.ts` against real
  staging (as `deploy-staging.yml` already does, post-deploy) and observe
  step 7's specific `reviewedBefore`, `reviewedCards`, and `reviewedAfter`
  values in the run log.
- Expected result: Step 7 passes
  (`reviewedAfter >= reviewedBefore + reviewedCards`) with the actually
  observed values recorded, not merely a green checkmark without recording
  what values were involved — confirming the fix holds under the exact
  condition that previously failed, and unblocking VOC-052-T01's evidence
  requirement.
- Evidence: `VOC-053-EV-02`

## Rollback coverage

Rolling back this package means reverting the affected task's code diff (see
`implementation-plan.md`'s "Deployment and rollback" section). No test in
this plan requires a corresponding data rollback rehearsal, since no
candidate fix under consideration is expected to be destructive; if
`VOC-053-R02`'s conditional migration path is taken, that task's own test
plan addition (recorded in its evidence file, per this repository's
migration-recovery conventions — see `apps/api/migrations/README.md`) must
include a rollback rehearsal before being claimed as passing.

## Constraints

No test in this plan uses secrets or production data. `VOC-053-TEST-00`'s
real staging HTTP inspection and `VOC-053-TEST-03`'s real staging E2E run use
only the existing, non-personal synthetic smoke-test account already
provisioned for this purpose by VOC-050 — never a real user's account or
production data.
