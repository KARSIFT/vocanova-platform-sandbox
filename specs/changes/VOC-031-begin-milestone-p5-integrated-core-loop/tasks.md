# VOC-031 — Tasks

Ordered PR sequence: `T00 → T01 → T02 → T03 → T04 → T05 → T06 → T07 → T08 →
T09`. Each PR is independently reviewable, remains R3-proposed (path floor
R3 for the new migration, the new CI workflow job, and the new
sensitive-data write surface), and requires Claude Code exact-SHA review.
**`D06`–`D08` are open decisions; no task in this roster may proceed past
the decision(s) it depends on by guessing.** `T00` is blocked on `D07`
(onboarding→`user_settings` seeding); `T02` is additionally blocked on `D06`
(settings/account write-field boundary); `T08` depends on `D08`'s
"do not rename" default holding (a founder reversal there would change
`T08`'s scope).

## VOC-031-T00 — `user_onboarding_profiles` persistence and `user_settings` chain extension

- Requirement source: `VOC-031-D00`, `VOC-031-D01`, `VOC-031-D07`, DOC-05
  §§6,18, DOC-06 §§3,13
- Acceptance criteria: `VOC-031-AC-00`, `VOC-031-AC-01`
- Tests: `VOC-031-TEST-00`..`VOC-031-TEST-05`
- Evidence: `VOC-031-EV-00`..`VOC-031-EV-05`
- Status: pending — blocked on `D07`

Add the `onboarding` Ent schema + reviewed versioned Atlas SQL for
`user_onboarding_profiles` (fields per specification.md item 1), added after
`grace_day_ledger` (DOC-05 §18 order), with FK, uniqueness (one row per
user), and the `daily_review_target` 5–100 check constraint. Implement a new
`apps/api/business/onboarding` module, transaction-scoped per DOC-06 §3
(exported functions accept an existing `*sql.Tx`, never open their own
transaction) — mirrors the `missions`/`gamification` pattern from P4. Per
the adopted `D07` resolution, wire the onboarding-completion transaction to
call the existing `gamification.EnsureUserSettings`-equivalent seeding path
so `daily_review_target` is set from the onboarding answer only when no
customized (non-default) `user_settings` value already exists; never
overwrite an existing customization. Unit-test the seeding logic in
isolation. Rehearse disposable forward migration and recovery; production
migration never runs at API startup. No API routes, no frontend in this PR.

## VOC-031-T01 — Onboarding API and `/onboarding` frontend flow

- Requirement source: `VOC-031-D00`, `VOC-031-D01`, DOC-03 §3, DOC-07
- Acceptance criteria: `VOC-031-AC-02`
- Tests: `VOC-031-TEST-06`..`VOC-031-TEST-11`
- Evidence: `VOC-031-EV-06`..`VOC-031-EV-11`
- Status: pending

Add `GET /api/v1/onboarding` and a submit endpoint (`PUT
/api/v1/onboarding` or equivalent upsert semantics; operation IDs
`GetOnboarding`/`SubmitOnboarding`) to the API surface: `RequireAuth`,
`CSRFMiddleware`/`Idempotency-Key` on the submit endpoint, explicit DTOs,
committed OpenAPI, matched `@vocanova/api-client` methods. Submission
validates the DOC-03 §3 enum fields, upserts `user_onboarding_profiles`,
sets `users.onboarding_status = 'completed'`, and triggers the `T00`
`user_settings` seeding — all in one transaction. Add the `/onboarding`
route (new `(onboarding)` route group) presenting the five-question
sequence with React Hook Form + Zod validation, then routing to `/home` on
success. Add the redirect rule: any `(app)` route redirects an
onboarding-incomplete learner to `/onboarding`; `/onboarding` itself
redirects an onboarding-complete learner to `/home`. No cross-user
exposure — the endpoints are implicitly self-scoped (no ID parameter).

## VOC-031-T02 — Settings and account write endpoints

- Requirement source: `VOC-031-D00`, `VOC-031-D02`, `VOC-031-D06`, DOC-07,
  DOC-08
- Acceptance criteria: `VOC-031-AC-03`
- Tests: `VOC-031-TEST-12`..`VOC-031-TEST-19`
- Evidence: `VOC-031-EV-12`..`VOC-031-EV-19`
- Status: pending — blocked on `D06`

Add `GET /api/v1/settings` / `PATCH /api/v1/settings` (owned by
`gamification`, extending its existing `user_settings` ownership from P4)
and `GET /api/v1/account` / `PATCH /api/v1/account` (owned by `auth`,
extending its existing `users` ownership from A1), following the exact
`RequireAuth()` + `CSRFMiddleware(authSvc)` + required `Idempotency-Key`
pattern used by every existing P1/P2 write (`SaveUserWord`, `SubmitReview`).
Per the adopted `D06` resolution: `/settings` reads/updates
`dailyReviewTarget`/`reviewIntervalPreset`/`appLanguage`/
`notificationsEnabled`/`marketingEmailsEnabled` only (`timezone` stays
internally resolved, not publicly editable); `/account` reads/updates
`displayName` only (email address and account deletion are out of scope —
no endpoint for either is added). Both are requester-scoped reads/writes
with no ID parameter to enumerate; unauthenticated → 401; a malformed or
out-of-range value (e.g. `dailyReviewTarget` outside 5–100) → 400. Add a
test proving no existing A1/P4 behavior (session handling, the internal
`gamification` timezone-resolution chain) is changed by adding this public
surface on top of it. No frontend wiring in this PR.

## VOC-031-T03 — `/settings` and `/settings/account` frontend screens

- Requirement source: `VOC-031-D00`, `VOC-031-D02`, DOC-08
- Acceptance criteria: `VOC-031-AC-04`
- Tests: `VOC-031-TEST-20`..`VOC-031-TEST-23`
- Evidence: `VOC-031-EV-20`..`VOC-031-EV-23`
- Status: pending

Add `/settings` (learning preferences: review target, interval preset,
language, notification/marketing toggles) and `/settings/account` (display
name) screens, wired to `T02`'s endpoints via `@vocanova/api-client` under
the A1 session, using React Hook Form + Zod and the shared loading/empty/
error components established in `T04`. No client DB access or duplicated
authorization. Save actions are optimistic-safe: a failed save leaves the
form showing the learner's attempted value with a clear error, never a
silently reverted or silently accepted value.

## VOC-031-T04 — Cross-feature integration pass

- Requirement source: `VOC-031-D00`, DOC-03 §4, DOC-08, VOC-026, VOC-027,
  VOC-028, VOC-030
- Acceptance criteria: `VOC-031-AC-05`
- Tests: `VOC-031-TEST-24`..`VOC-031-TEST-28`
- Evidence: `VOC-031-EV-24`..`VOC-031-EV-28`
- Status: pending

Introduce one shared loading/empty/error component set (extracted from the
best existing per-route pattern established in P1–P4) and adopt it across
every `(app)`/`(onboarding)` route, replacing each route's ad hoc version.
Audit and fix cross-screen navigation entry points (Home → Discover/Review/
Sentence practice/Progress/Settings and back) so every documented DOC-08
core-loop transition (DOC-03 §4) is reachable without a dead end or a
full-page reload where client navigation is expected. Add a consistency
test proving Home's `dueReviewWords` count and the Review screen's own
due-queue read never disagree for the same requester at the same instant
(extends the VOC-030 Home/Progress streak-consistency property, `VOC-030-EV-28`,
to this additional screen pair). No backend business-logic change in this
PR — frontend/shared-component and navigation wiring only.

## VOC-031-T05 — Reliability and recovery pass

- Requirement source: `VOC-031-D00`, DOC-08, DOC-12 §5 P3 (AI-optionality
  gate language)
- Acceptance criteria: `VOC-031-AC-06`
- Tests: `VOC-031-TEST-29`..`VOC-031-TEST-34`
- Evidence: `VOC-031-EV-29`..`VOC-031-EV-34`
- Status: pending

Harden the full loop against interrupted requests: a failed unsafe write
(network error, 5xx) must regenerate or safely reuse its `Idempotency-Key`
on retry — never silently resubmit with a stale key in a way that could
mask a genuine duplicate, and never leave the UI in a state where retry is
impossible. An expired session mid-flow redirects to sign-in and, on
successful re-authentication, returns the learner to their prior route
(e.g. `/review`, not unconditionally `/home`) where reasonably possible. A
slow or failed AI-feedback call (P3) degrades gracefully — the review/home
loop remains fully usable while sentence feedback is unavailable, matching
the existing "AI can be disabled without disabling non-AI learning" gate
language. Add tests simulating a mid-flight network failure on each of the
P1/P2/P3/T02 write endpoints and asserting no duplicate side effect and a
usable retry path. No backend business-logic change beyond retry-safety
wiring in this PR.

## VOC-031-T06 — Accessibility automation and pass

- Requirement source: `VOC-031-D00`, `VOC-031-D03`, `VOC-031-D05`, DOC-08
- Acceptance criteria: `VOC-031-AC-07`
- Tests: `VOC-031-TEST-35`..`VOC-031-TEST-38`
- Evidence: `VOC-031-EV-35`..`VOC-031-EV-38`
- Status: pending

Install `playwright` and `@axe-core/playwright` (neither currently
installed) in `apps/web`, add `tests/e2e` with an automated axe-core sweep
over every `(public)`/`(onboarding)`/`(app)` route at the three supported
layouts (`D05`: 360px, 430px, ≥1024px), and wire it into a new CI job. Fix
every violation the sweep finds against the DOC-08 WCAG 2.2 AA target
(labelled controls, visible focus, keyboard reachability, non-color-only
state, 44px minimum touch targets) across every route it covers, including
this package's own new `/onboarding`/`/settings`/`/settings/account`
screens. No manual-only fallback; any route the automation genuinely cannot
reach is recorded as an explicit limitation, never reported as passing.

## VOC-031-T07 — Performance automation and pass

- Requirement source: `VOC-031-D00`, `VOC-031-D04`, `VOC-031-D05`, DOC-08
- Acceptance criteria: `VOC-031-AC-08`
- Tests: `VOC-031-TEST-39`..`VOC-031-TEST-41`
- Evidence: `VOC-031-EV-39`..`VOC-031-EV-41`
- Status: pending

Install and configure Lighthouse CI (`@lhci/cli`, not currently installed)
against a production build, targeting `/home`, `/discover`, `/reviews`,
`/progress`, `/settings`, and `/onboarding`, asserting the DOC-08 thresholds
(Performance 85+, Accessibility 95+, Best Practices 90+) as a new CI job
that fails the build on regression. Remediate any threshold failure found
(route-level code splitting, dependency weight, image/asset handling per
DOC-08's existing performance guidance). No manual spot-check fallback.

## VOC-031-T08 — Final UX-consistency pass

- Requirement source: `VOC-031-D00`, `VOC-031-D05`, `VOC-031-D08`, DOC-08
- Acceptance criteria: `VOC-031-AC-09`
- Tests: `VOC-031-TEST-42`..`VOC-031-TEST-44`
- Evidence: `VOC-031-EV-42`..`VOC-031-EV-44`
- Status: pending — scope depends on `D08`'s "do not rename" default holding

Audit every P1–P5 screen at the three supported layouts (`D05`) for
consistent spacing/typography/color-token usage (no ad hoc values), 44px
minimum touch targets, consistent focus-visible styling, and consistent use
of the `T04` shared loading/empty/error components; fix drift found. Per the
adopted `D08` default, do **not** rename or restructure any already-shipped
route (`/reviews`, the inline saved-words presentation) as part of this
pass — record the DOC-08-vs-implementation drift `D08` found, unresolved,
for a future dedicated decision. No backend change in this PR.

## VOC-031-T09 — Evaluation, mock-inventory, staging evidence, and P5 gate readiness

- Requirement source: `VOC-031-D00`, DOC-12 §5 P5
- Acceptance criteria: `VOC-031-AC-10`
- Tests: `VOC-031-TEST-45`..`VOC-031-TEST-48`
- Evidence: `VOC-031-EV-45`..`VOC-031-EV-48`
- Status: pending

Run the full installed check suite (`T00`–`T08`'s additions included).
Update the deterministic mock-inventory check (`scripts/foundation/mock-inventory.mjs`)
to admit the new `onboarding` module/routes/schema/migration and the new
`settings`/`account` routes, and assert no P4 mock regressed and no R1/R2
behavior was invented. Collect staging-exercise procedures (full loop:
onboard → discover → save → review → complete mission → sentence practice →
progress → settings, at all three supported layouts), rollback rehearsal
for the one new table, and P5 gate readiness. Do not declare the DOC-12 P5
gate complete.

### Deliverables

- `mock-inventory.md`: records the new real P5 modules/tables/routes and
  confirms no P4-and-earlier mock regressed.
- `staging-evidence.md`: collects in-repository evidence and documents the
  staged exercises and rollback rehearsal that can only run once F3 exists.
- updated `scripts/foundation/mock-inventory.mjs` (+`.test.mjs`):
  deterministic check enforcing the new P5 boundaries.

### Blocker

`VOC-031-DEP-02` remains open: F3 staging does not exist, so the live
staging exercises — which the DOC-12 §5 P5 gate's own wording centers on
("works coherently in staging across supported layouts") — cannot be
executed. This task provides the procedures and the in-repository evidence
only; it does not declare the DOC-12 P5 gate complete.
