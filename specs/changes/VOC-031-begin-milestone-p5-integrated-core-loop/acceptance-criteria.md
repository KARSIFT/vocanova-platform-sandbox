# VOC-031 — Acceptance Criteria

Acceptance criteria are observable, stable, security-aware, and bidirectionally
traceable to requirements (`D00`–`D09`), tasks (`T00`–`T11`), tests
(`VOC-031-TEST-*`), and evidence. `D02`, `D03`, `D05`–`D07`, and `D09` are
open or this-draft-proposed decisions; the criteria below are written against
this draft's proposed defaults and must be re-verified against whatever the
founder actually confirms at adoption.

## VOC-031-AC-00 — `user_onboarding_profiles` and the onboarding seed rule

- Requirement source: `VOC-031-D00`, `VOC-031-D04`, DOC-05 §6
- Tasks: `VOC-031-T00`
- Tests: `VOC-031-TEST-00`..`VOC-031-TEST-03`
- Evidence: `VOC-031-EV-00`..`VOC-031-EV-03`
- Result: pending

Ent/Atlas create `user_onboarding_profiles` exactly per DOC-05 §6 (fields,
5–100 `daily_review_target` check, one row per user). No existing A1–P4
table, column, or constraint is altered. A pure, unit-tested function
implements the exact `VOC-031-D04` seed rule: no `user_settings` row → seed
freely; row present with `daily_review_target == 20` → overwrite; row
present with any other value → never overwrite.

## VOC-031-AC-01 — Onboarding API, frontend flow, and gating

- Requirement source: `VOC-031-D00`, `VOC-031-D02`, `VOC-031-D03`, DOC-03 §3
- Tasks: `VOC-031-T01`
- Tests: `VOC-031-TEST-04`..`VOC-031-TEST-07`
- Evidence: `VOC-031-EV-04`..`VOC-031-EV-07`
- Result: pending — `D03` open

`POST /api/v1/onboarding` persists the five answers in one submission (no
partial-save), calls the `T00` seed function, and sets
`onboarding_status='completed'`; `GET /api/v1/onboarding` returns the current
profile/status. `GET /api/v1/me`'s response gains an additive
`onboardingStatus` field; every other existing field is unchanged. An
authenticated learner whose `onboardingStatus` is not `completed` is
redirected to `/onboarding` from any other `(app)` route; a learner whose
`onboardingStatus` is `completed` is redirected away from `/onboarding` to
`/home`. Unauthenticated access to `/onboarding` or its API redirects/401s
the same way every other `(app)` route already does.

## VOC-031-AC-02 — Settings read/write

- Requirement source: `VOC-031-D00`, `VOC-031-D01`, `VOC-031-D06`, DOC-00 §3, DOC-05 §6
- Tasks: `VOC-031-T02`
- Tests: `VOC-031-TEST-08`..`VOC-031-TEST-11`
- Evidence: `VOC-031-EV-08`..`VOC-031-EV-11`
- Result: pending

`GET /api/v1/settings` returns `dailyReviewTarget`, `reviewIntervalPreset`,
`appLanguage`, `notificationsEnabled`, `marketingEmailsEnabled`,
`displayName`. `PATCH /api/v1/settings` accepts a partial update of the same
fields with strict validation (unknown fields rejected; `dailyReviewTarget`
5–100; `reviewIntervalPreset` a recognized preset; `appLanguage` restricted
to `en`; `displayName` length-limited). A first-ever write upserts the
`user_settings` row without a unique-constraint race against
`gamification`'s existing lazy-creation path. A `dailyReviewTarget` change
never rewrites the current local day's already-created
`daily_mission_snapshots.review_target` (DOC-00 §3). Unauthenticated → 401;
no cross-user parameter exists to enumerate another learner's settings.

## VOC-031-AC-03 — Email-change verification

- Requirement source: `VOC-031-D01`, `VOC-031-D05`, DOC-06 §6
- Tasks: `VOC-031-T03`
- Tests: `VOC-031-TEST-12`..`VOC-031-TEST-17`
- Evidence: `VOC-031-EV-12`..`VOC-031-EV-17`
- Result: pending

`POST /api/v1/settings/email-change-links` requires an authenticated
session, rate-limits by IP and session, and returns a generic success
response regardless of whether the requested new email is already
registered (no enumeration signal). The issued token matches the magic-link
pattern exactly: 32 random bytes, only its SHA-256 hash persisted, 15-minute
expiry, single-use via `consumed_at`, environment-scoped.
`POST /api/v1/settings/email-change-links/consume` validates the token,
re-checks new-email uniqueness atomically at confirm time (a collision
returns a stable, non-500 conflict, not a crash), updates `users.email`, and
sends a best-effort, non-blocking notification to the *old* email address.
Neither step invalidates the requester's current session. An existing
Google-OAuth-linked login (matched by `provider_subject`, not `email`) is
unaffected by an email change.

## VOC-031-AC-04 — Account deletion

- Requirement source: `VOC-031-D01`, `VOC-031-D07`, `VOC-031-D09`, DOC-05 §16, DOC-06 §§9,14,15, DOC-07
- Tasks: `VOC-031-T04`
- Tests: `VOC-031-TEST-18`..`VOC-031-TEST-23`
- Evidence: `VOC-031-EV-18`..`VOC-031-EV-23`
- Result: pending — production enablement additionally gated by `VOC-031-DEP-03`

`POST /api/v1/account-deletion-requests` requires `RequireAuth` and an
`Idempotency-Key` (DOC-07); a replayed key produces no duplicate effect. On
success, inside one transaction: `users.status` becomes `deleted`,
`deleted_at` is set, every active session and every unconsumed
magic-link/email-change-link for the account is revoked, and one
`account_deletion_requests` row is inserted (`status='deactivated'`,
`purge_after = requested_at + 30 days`). The idempotent anonymization sweep,
once a row passes `purge_after`, performs exactly the DOC-05 §16 per-table
disposition (soft-delete-pending-purge, irreversible de-identification, or
delete/de-identify-if-retained as documented per table) and transitions the
row to `status='completed'`; no `ON DELETE CASCADE` is used. `account_deletion`
is a valid `idempotency_keys.scope` value (`VOC-031-D09`). No production
enablement is authorized by this criterion or this package.

## VOC-031-AC-05 — Settings/account frontend

- Requirement source: `VOC-031-D00`, `VOC-031-D01`, `VOC-031-D05`, `VOC-031-D06`, DOC-08
- Tasks: `VOC-031-T05`
- Tests: `VOC-031-TEST-24`..`VOC-031-TEST-27`
- Evidence: `VOC-031-EV-24`..`VOC-031-EV-27`
- Result: pending

`/settings` renders and writes all six editable fields via `T02`, presenting
`appLanguage` as a single confirmed option, not a functioning language
picker. `/settings/account` renders the email-change request→pending→confirm
flow (input preserved on error, no duplicate submission) and the
account-deletion flow behind an explicit multi-step confirmation (not a
single click), ending in logout and a clear post-deletion message. Both
routes are covered by the auth-gate matcher. No screen shows a
client-fabricated fallback value on error.

## VOC-031-AC-06 — Cross-feature reliability and recovery

- Requirement source: `VOC-031-D00`, DOC-12 §5 P5, DOC-03 §9
- Tasks: `VOC-031-T06`
- Tests: `VOC-031-TEST-28`..`VOC-031-TEST-31`
- Evidence: `VOC-031-EV-28`..`VOC-031-EV-31`
- Result: pending

Every core-loop screen has a safe retry path on network failure (no lost
learner input, no falsely-implied completion) and a defined behavior when a
session expires mid-flow. Every already-shipped A1–P4 screen's pre-existing
behavior is unchanged (regression-tested, not merely asserted, mirroring
`VOC-030-R01`'s standard). No client-fabricated fallback value exists
anywhere in the core loop.

## VOC-031-AC-07 — Accessibility automation

- Requirement source: `VOC-031-D00`, `VOC-031-D01`, DOC-03 §10, DOC-08 quality standards
- Tasks: `VOC-031-T07`
- Tests: `VOC-031-TEST-32`..`VOC-031-TEST-35`
- Evidence: `VOC-031-EV-32`..`VOC-031-EV-35`
- Result: pending

Playwright + an axe-core integration are installed (net-new). An automated
accessibility scan runs across every core-loop screen at 360px, 430px, and
one representative desktop width ≥1024px, asserting zero critical/serious
axe violations, keyboard reachability, and non-color-only feedback,
explicitly, and is wired into CI as a required job.

## VOC-031-AC-08 — Full core-loop end-to-end suite

- Requirement source: `VOC-031-D00`, DOC-10 §7
- Tasks: `VOC-031-T08`
- Tests: `VOC-031-TEST-36`
- Evidence: `VOC-031-EV-36`
- Result: pending

A Playwright suite exercises auth → onboarding → discover → save → review
session → sentence submission → deterministic (mock-provider) AI feedback →
progress update → settings change → logout → unauthenticated-access
rejection, end to end, with no paid/nondeterministic provider call in CI.

## VOC-031-AC-09 — Performance automation

- Requirement source: `VOC-031-D00`, `VOC-031-D01`, DOC-08 quality standards
- Tasks: `VOC-031-T09`
- Tests: `VOC-031-TEST-37`..`VOC-031-TEST-39`
- Evidence: `VOC-031-EV-37`..`VOC-031-EV-39`
- Result: pending

Lighthouse CI (net-new) runs against a production build at the three
supported layouts on Home, Discover, Reviews, and Progress, asserting
Performance 85+, Accessibility 95+, Best Practices 90+ (DOC-08); any
threshold not met is reported as an explicit limitation, never silently
lowered or skipped.

## VOC-031-AC-10 — Final UX-consistency pass

- Requirement source: `VOC-031-D00`, DOC-03 §1, §11
- Tasks: `VOC-031-T10`
- Tests: `VOC-031-TEST-40`
- Evidence: `VOC-031-EV-40`
- Result: pending

A recorded design-principle audit (screen × principle × gap × fix) exists
for one-clear-action, encouraging non-gamified tone, and calm visual
framing across every core-loop screen, old and new — not a bare
"reviewed" assertion.

## VOC-031-AC-11 — Evidence, mock-inventory, staging, and P5 gate readiness

- Requirement source: `VOC-031-D00`, DOC-12 §5 P5
- Tasks: `VOC-031-T00`..`VOC-031-T11`
- Tests: `VOC-031-TEST-41`..`VOC-031-TEST-45`
- Evidence: `VOC-031-EV-41`..`VOC-031-EV-45`
- Result: pending — in-repository evidence only; live staging blocked until F3 exists

Applicable checks, the deterministic tests this package adds, exact-SHA
reviews, and the extended mock-inventory check pass; the mock-inventory
confirms no pre-existing mock remains (none did before this package) and no
P5-invented route/table/behavior beyond this package's own documented scope
exists. Staging exercises (the full core-loop flow, cross-user denial, the
three-new-tables rollback rehearsal) are documented and ready to run once F3
staging exists (`VOC-031-DEP-02`). This enables — but does not itself
declare — the DOC-12 P5 gate evaluation.
