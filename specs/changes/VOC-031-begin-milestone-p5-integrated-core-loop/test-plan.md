# VOC-031 — Test Plan

No test, fixture, seed file, OpenAPI example, or evidence may contain a real
secret, production URL/data, another learner's personal content, or a raw
session/CSRF/email-change/magic-link token. Discover installed commands at
the adopted base; a missing integration, staging credential, open decision
(`VOC-031-DEP-01` through `DEP-05`), or browser tool is never reported as a
pass — it is a recorded limitation or blocker.

## VOC-031-TEST-00 — `user_onboarding_profiles` migration invariants
- Covers: `VOC-031-AC-00`; Preconditions: T00, disposable PostgreSQL.
- Procedure: apply the T00 migration and assert the unique `user_id` FK, the
  `daily_review_target` 5–100 check, and that no existing A1–P4 table/column/
  constraint changed.
- Expected result: migration passes; no regression. Evidence: `VOC-031-EV-00`.

## VOC-031-TEST-01 — Seed rule: no existing row
- Covers: `VOC-031-AC-00`, `VOC-031-D04`; Preconditions: T00.
- Procedure: call the seed function for a user with no `user_settings` row.
- Expected result: a row is created with `daily_review_target` set from the
  onboarding answer. Evidence: `VOC-031-EV-01`.

## VOC-031-TEST-02 — Seed rule: default-value row
- Covers: `VOC-031-AC-00`, `VOC-031-D04`; Preconditions: T00.
- Procedure: call the seed function for a user with a `user_settings` row at
  `daily_review_target=20` (schema default).
- Expected result: the row is overwritten with the onboarding answer.
  Evidence: `VOC-031-EV-02`.

## VOC-031-TEST-03 — Seed rule: customized row is never overwritten
- Covers: `VOC-031-AC-00`, `VOC-031-D04`; Preconditions: T00.
- Procedure: call the seed function for a user with a `user_settings` row at
  a non-default `daily_review_target` (e.g. 35).
- Expected result: the existing value is preserved untouched. Evidence:
  `VOC-031-EV-03`.

## VOC-031-TEST-04 — Onboarding submission happy path
- Covers: `VOC-031-AC-01`; Preconditions: T01.
- Procedure: submit all five answers via `POST /api/v1/onboarding`.
- Expected result: `user_onboarding_profiles` row created,
  `onboarding_status='completed'`, `T00`'s seed function invoked exactly
  once. Evidence: `VOC-031-EV-04`.

## VOC-031-TEST-05 — Gate redirects an incomplete learner
- Covers: `VOC-031-AC-01`; Preconditions: T01.
- Procedure: authenticate as a learner with `onboardingStatus != 'completed'`
  and request `/home`, `/discover`, `/progress`, `/reviews`, `/settings`.
- Expected result: every route redirects to `/onboarding`. Evidence:
  `VOC-031-EV-05`.

## VOC-031-TEST-06 — Gate skips a completed learner
- Covers: `VOC-031-AC-01`; Preconditions: T01.
- Procedure: authenticate as a learner with `onboardingStatus == 'completed'`
  and request `/onboarding`.
- Expected result: redirected to `/home`; other `(app)` routes render
  normally. Evidence: `VOC-031-EV-06`.

## VOC-031-TEST-07 — `GET /api/v1/me` additive-field contract
- Covers: `VOC-031-AC-01`; Preconditions: T01.
- Procedure: call `GET /api/v1/me` before and after this package and diff the
  response shape.
- Expected result: only `onboardingStatus` is added; every previously
  existing field is unchanged. Evidence: `VOC-031-EV-07`.

## VOC-031-TEST-08 — Settings read/write happy path
- Covers: `VOC-031-AC-02`; Preconditions: T02.
- Procedure: `PATCH /api/v1/settings` with each of the six fields, then
  `GET /api/v1/settings`.
- Expected result: the read reflects exactly the written values. Evidence:
  `VOC-031-EV-08`.

## VOC-031-TEST-09 — Settings validation rejects invalid input
- Covers: `VOC-031-AC-02`; Preconditions: T02.
- Procedure: submit `dailyReviewTarget=200`, an unrecognized
  `reviewIntervalPreset`, an `appLanguage` other than `en`, an unknown field,
  and an over-length `displayName`.
- Expected result: each is rejected with a stable validation error; no
  partial write occurs. Evidence: `VOC-031-EV-09`.

## VOC-031-TEST-10 — First-write upsert race safety
- Covers: `VOC-031-AC-02`, `VOC-031-R05`; Preconditions: T02.
- Procedure: concurrently trigger a `PATCH /api/v1/settings` write and a
  `gamification` lazy-creation read for a user with no existing
  `user_settings` row.
- Expected result: exactly one row exists afterward; neither path errors on
  a unique-constraint violation. Evidence: `VOC-031-EV-10`.

## VOC-031-TEST-11 — `dailyReviewTarget` change is not retroactive
- Covers: `VOC-031-AC-02`; Preconditions: T02, an existing today's
  `daily_mission_snapshots` row.
- Procedure: change `dailyReviewTarget` after today's snapshot already
  exists, then read `GET /api/v1/daily-mission`.
- Expected result: today's `reviewTarget` is unchanged; the new value only
  applies to the next local day's snapshot. Evidence: `VOC-031-EV-11`.

## VOC-031-TEST-12 — Email-change request: generic response, no enumeration
- Covers: `VOC-031-AC-03`; Preconditions: T03.
- Procedure: request a change to an already-registered email and to an
  unregistered email.
- Expected result: identical generic success response in both cases.
  Evidence: `VOC-031-EV-12`.

## VOC-031-TEST-13 — Email-change token mechanics match magic-link pattern
- Covers: `VOC-031-AC-03`, `VOC-031-D05`; Preconditions: T03.
- Procedure: inspect the persisted `email_change_links` row after a request.
- Expected result: only a SHA-256 hash is stored, `expires_at` is 15 minutes
  out, `environment` is scoped. Evidence: `VOC-031-EV-13`.

## VOC-031-TEST-14 — Invalid/expired/consumed token rejected
- Covers: `VOC-031-AC-03`; Preconditions: T03.
- Procedure: attempt to consume an expired token, a tampered token, and a
  previously-consumed token.
- Expected result: each attempt is rejected (401/stable error), never
  distinguishing why. Evidence: `VOC-031-EV-14`.

## VOC-031-TEST-15 — Duplicate-email confirmation race
- Covers: `VOC-031-AC-03`, `VOC-031-R02`; Preconditions: T03.
- Procedure: two users request a change to the same new email and both
  attempt to confirm.
- Expected result: exactly one confirm succeeds; the other receives a
  stable conflict response, not a 500. Evidence: `VOC-031-EV-15`.

## VOC-031-TEST-16 — Session survives email change; old-email notification sent
- Covers: `VOC-031-AC-03`, `VOC-031-R01`; Preconditions: T03.
- Procedure: complete an email change and inspect the requester's session
  and the old-email notification dispatch.
- Expected result: the current session remains valid; a notification is
  dispatched to the old address (best-effort, non-blocking). Evidence:
  `VOC-031-EV-16`.

## VOC-031-TEST-17 — Google-OAuth login unaffected by email change
- Covers: `VOC-031-AC-03`; Preconditions: T03, a Google-linked test account.
- Procedure: change the account's email, then sign in via the existing
  Google OAuth flow.
- Expected result: login succeeds unchanged (matched by `provider_subject`,
  not `email`). Evidence: `VOC-031-EV-17`.

## VOC-031-TEST-18 — Account-deletion request: deactivation and session revocation
- Covers: `VOC-031-AC-04`; Preconditions: T04.
- Procedure: call `POST /api/v1/account-deletion-requests` with a valid
  `Idempotency-Key` for an account with multiple active sessions.
- Expected result: `users.status='deleted'`, every session/magic-link/
  email-change-link revoked, one `account_deletion_requests` row created.
  Evidence: `VOC-031-EV-18`.

## VOC-031-TEST-19 — Idempotent replay creates no duplicate effect
- Covers: `VOC-031-AC-04`; Preconditions: T04.
- Procedure: replay the same `Idempotency-Key` for an already-processed
  deletion request.
- Expected result: no second `account_deletion_requests` row, no error
  beyond the standard idempotent-replay response. Evidence: `VOC-031-EV-19`.

## VOC-031-TEST-20 — Anonymization sweep applies the exact DOC-05 §16 disposition
- Covers: `VOC-031-AC-04`; Preconditions: T04, a row past `purge_after`.
- Procedure: run the sweep against a deactivated test account with rows in
  every DOC-05 §16-listed table.
- Expected result: soft-delete-pending-purge tables are deleted/marked;
  immutable-history tables are de-identified, not deleted; deletion-dependent
  tables are deleted or de-identified; the request transitions to
  `completed`. Evidence: `VOC-031-EV-20`.

## VOC-031-TEST-21 — Sweep is idempotent and resumable
- Covers: `VOC-031-AC-04`, `VOC-031-D07`; Preconditions: T04.
- Procedure: interrupt the sweep mid-run (simulated failure) and re-run it.
- Expected result: the row reaches `completed` with no duplicate or
  corrupted anonymization on the second run. Evidence: `VOC-031-EV-21`.

## VOC-031-TEST-22 — Unauthorized/cross-user deletion rejected
- Covers: `VOC-031-AC-04`; Preconditions: T04.
- Procedure: call the endpoint unauthenticated and with another user's
  identifier.
- Expected result: 401 unauthenticated; no cross-user deletion path exists
  (the endpoint is self-scoped, no target-user parameter). Evidence:
  `VOC-031-EV-22`.

## VOC-031-TEST-23 — `idempotency_keys.scope` accepts `account_deletion`
- Covers: `VOC-031-AC-04`, `VOC-031-D09`; Preconditions: T04.
- Procedure: insert an `idempotency_keys` row with
  `scope='account_deletion'`.
- Expected result: accepted by the check constraint. Evidence:
  `VOC-031-EV-23`.

## VOC-031-TEST-24 — Settings frontend edit flow
- Covers: `VOC-031-AC-05`; Preconditions: T05.
- Procedure: render `/settings`, edit each field, and submit.
- Expected result: real `PATCH /api/v1/settings` calls are made; `appLanguage`
  renders as a single confirmed option, not a picker with inert alternatives.
  Evidence: `VOC-031-EV-24`.

## VOC-031-TEST-25 — Email-change frontend flow states
- Covers: `VOC-031-AC-05`; Preconditions: T05.
- Procedure: request a change, observe the pending state, then simulate a
  confirm error.
- Expected result: input preserved on error, no duplicate submission
  possible, honest pending/error states. Evidence: `VOC-031-EV-25`.

## VOC-031-TEST-26 — Account-deletion frontend confirmation flow
- Covers: `VOC-031-AC-05`; Preconditions: T05.
- Procedure: attempt deletion without completing the multi-step
  confirmation, then complete it.
- Expected result: deletion is blocked until confirmation is complete; on
  success the learner is logged out with a clear message. Evidence:
  `VOC-031-EV-26`.

## VOC-031-TEST-27 — Settings/account loading, empty, and error states
- Covers: `VOC-031-AC-05`; Preconditions: T05.
- Procedure: render both routes with a slow/failing backend.
- Expected result: no fabricated fallback value; accessible loading/error
  states. Evidence: `VOC-031-EV-27`.

## VOC-031-TEST-28 — Retry safety on network failure across the core loop
- Covers: `VOC-031-AC-06`; Preconditions: T06.
- Procedure: simulate a network failure during a review submission, a
  sentence submission, and a settings write, then retry.
- Expected result: no lost input, no falsely-implied completion, no
  duplicate side effect on retry. Evidence: `VOC-031-EV-28`.

## VOC-031-TEST-29 — Session-expiry mid-flow handling
- Covers: `VOC-031-AC-06`; Preconditions: T06.
- Procedure: expire the session mid-review-session, mid-sentence-submission,
  and mid-onboarding.
- Expected result: a clear re-authentication path in each case; no progress
  is claimed as saved when it was not. Evidence: `VOC-031-EV-29`.

## VOC-031-TEST-30 — No client-fabricated fallback values
- Covers: `VOC-031-AC-06`; Preconditions: T06.
- Procedure: sweep every core-loop screen's error/loading state for a
  hardcoded or fabricated data value.
- Expected result: none found. Evidence: `VOC-031-EV-30`.

## VOC-031-TEST-31 — A1–P4 regression check
- Covers: `VOC-031-AC-06`, `VOC-031-R06`; Preconditions: T06.
- Procedure: re-run the VOC-025/026/027/028/030 test suites against the T06
  code.
- Expected result: no regression. Evidence: `VOC-031-EV-31`.

## VOC-031-TEST-32 — Accessibility automation installed and runnable
- Covers: `VOC-031-AC-07`; Preconditions: T07.
- Procedure: run the new Playwright + axe-core suite locally and in CI.
- Expected result: the suite executes and reports pass/fail per screen per
  layout. Evidence: `VOC-031-EV-32`.

## VOC-031-TEST-33 — Zero critical/serious violations across core-loop screens
- Covers: `VOC-031-AC-07`; Preconditions: T07.
- Procedure: scan every core-loop screen at 360px, 430px, and the desktop
  width.
- Expected result: zero critical/serious axe violations; any found is fixed
  before this task is accepted. Evidence: `VOC-031-EV-33`.

## VOC-031-TEST-34 — Keyboard reachability and non-color-only feedback
- Covers: `VOC-031-AC-07`; Preconditions: T07.
- Procedure: tab through each screen's interactive elements; verify
  correct/incorrect review feedback and mission/streak state use an icon or
  text label, not color alone.
- Expected result: fully keyboard-operable; no color-only signal. Evidence:
  `VOC-031-EV-34`.

## VOC-031-TEST-35 — Accessibility suite wired into CI
- Covers: `VOC-031-AC-07`; Preconditions: T07.
- Procedure: inspect the CI workflow configuration.
- Expected result: the suite runs as a required job on relevant PRs.
  Evidence: `VOC-031-EV-35`.

## VOC-031-TEST-36 — Full core-loop end-to-end suite
- Covers: `VOC-031-AC-08`; Preconditions: T08.
- Procedure: run the documented DOC-10 §7 flow end to end against the mock
  AI provider.
- Expected result: every step completes; no paid/nondeterministic provider
  call occurs. Evidence: `VOC-031-EV-36`.

## VOC-031-TEST-37 — Lighthouse CI installed and runnable against a production build
- Covers: `VOC-031-AC-09`; Preconditions: T09.
- Procedure: build the web app for production and run Lighthouse CI against
  it locally and in CI.
- Expected result: the suite executes and reports scores per screen per
  layout. Evidence: `VOC-031-EV-37`.

## VOC-031-TEST-38 — DOC-08 thresholds met on core screens
- Covers: `VOC-031-AC-09`; Preconditions: T09.
- Procedure: assert Performance 85+, Accessibility 95+, Best Practices 90+ on
  Home, Discover, Reviews, Progress.
- Expected result: thresholds met, or any shortfall recorded as an explicit
  limitation with a follow-up. Evidence: `VOC-031-EV-38`.

## VOC-031-TEST-39 — Performance suite wired into CI against a stable target
- Covers: `VOC-031-AC-09`, `VOC-031-R04`; Preconditions: T09.
- Procedure: inspect the CI workflow and confirm the Lighthouse target is a
  fixed local production build, not a live/dev-server target.
- Expected result: stable, reproducible target; suite runs as a CI job.
  Evidence: `VOC-031-EV-39`.

## VOC-031-TEST-40 — UX-consistency audit recorded
- Covers: `VOC-031-AC-10`; Preconditions: T10.
- Procedure: inspect the audit record for every core-loop screen.
- Expected result: each screen has a recorded finding (or "no gap found")
  against DOC-03 §1/§11, with any fix applied. Evidence: `VOC-031-EV-40`.

## VOC-031-TEST-41 — Installed deterministic and security suite
- Covers: `VOC-031-AC-11`; Preconditions: each PR complete.
- Procedure: run relevant `pnpm validate`/`pnpm test`/`pnpm build`, Go
  format/vet/test/build, web lint/typecheck/build/format,
  `scripts/governance/*` checks, and the extended mock-inventory check.
- Expected result: available checks pass; absent checks reported honestly.
  Evidence: `VOC-031-EV-41`.

## VOC-031-TEST-42 — Mock-inventory check confirms zero legacy mocks and no invented behavior
- Covers: `VOC-031-AC-11`; Preconditions: T11.
- Procedure: run the extended `scripts/foundation/mock-inventory.mjs`.
- Expected result: confirms zero `MOCK_*` constants existed before this
  package and none are newly introduced without disposition; no route/table
  beyond this package's documented scope is found. Evidence: `VOC-031-EV-42`.

## VOC-031-TEST-43 — Staging full-loop and cross-user-denial procedures documented
- Covers: `VOC-031-AC-11`; Preconditions: F3 staging exists
  (`VOC-031-DEP-02`), open decisions resolved.
- Procedure: with non-production identities, exercise onboarding → settings
  → email change → account deletion → the full core loop, and confirm
  cross-user denial throughout.
- Expected result: P5 flow evidence recorded without production data.
  Evidence: `VOC-031-EV-43`.

## VOC-031-TEST-44 — New-tables rollback rehearsal
- Covers: `VOC-031-AC-11`; Preconditions: staged candidate, approved
  procedure.
- Procedure: rehearse non-production migration rollback for the three new
  tables; validate no learner state is left inconsistent.
- Expected result: controlled recovery; no partial-deletion or lost-settings
  state. Evidence: `VOC-031-EV-44`.

## VOC-031-TEST-45 — Exact-SHA independent verification
- Covers: `VOC-031-AC-11`; Preconditions: each PR at its final SHA.
- Procedure: Claude Code binds to the exact final SHA per PR and verifies
  scope, the classifier floor, migration safety, no A1–P4 regression, the
  account-deletion/email-change security guarantees with concrete evidence,
  the open-decision resolutions as actually implemented, contract/OpenAPI/
  client drift, the accessibility/performance automation's real operation
  (not a stubbed-out or always-passing check), staging/rollback evidence,
  and implementer separation; reports remaining R3/R4/adoption/activation
  gates.
- Expected result: `PASS` / `PASS WITH NON-BLOCKING FINDINGS` / `FAIL` with
  exact evidence; the implementer did not approve or merge its own work.
  Evidence: `VOC-031-EV-45`.
