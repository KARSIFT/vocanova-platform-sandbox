# VOC-031 — Acceptance Criteria

Acceptance criteria are observable, stable, security-aware, and
bidirectionally traceable to requirements (`D00`–`D09`), tasks
(`T00`–`T09`), tests (`VOC-031-TEST-*`), and evidence. `D06`–`D08` are
**open decisions**; the criteria below are written against this draft's
proposed defaults and must be re-verified against whatever the founder
actually resolves at adoption.

## VOC-031-AC-00 — `user_onboarding_profiles` table and migration integrity

- Requirement source: `VOC-031-D00`, `VOC-031-D01`, DOC-05 §§6,18
- Tasks: `VOC-031-T00`
- Tests: `VOC-031-TEST-00`, `VOC-031-TEST-01`
- Evidence: `VOC-031-EV-00`, `VOC-031-EV-01`
- Result: pending

Ent/Atlas creates `user_onboarding_profiles` exactly per DOC-05 §6 (fields,
checks, uniqueness as specified in specification.md item 1), added after
`grace_day_ledger`. No existing A1/P1/P2/P3/P4 table, column, or constraint
is altered. Empty-db migration and disposable recovery rehearsal preserve
integrity; production migration never runs at API startup.

## VOC-031-AC-01 — Onboarding-completion seeds `user_settings.daily_review_target` (`D07`)

- Requirement source: `VOC-031-D00`, `VOC-031-D07`, DOC-06 §13
- Tasks: `VOC-031-T00`
- Tests: `VOC-031-TEST-02`..`VOC-031-TEST-05`
- Evidence: `VOC-031-EV-02`..`VOC-031-EV-05`
- Result: pending — blocked on `D07` resolution at adoption

At onboarding completion, `user_settings.daily_review_target` is set from
the onboarding answer only when no `user_settings` row with an
already-customized (non-default) value exists; an existing customization is
never overwritten. `user_settings.timezone` resolution is unchanged from
`VOC-030-D01` (client-supplied IANA timezone, validated, else UTC) — the
DOC-03 §3 question list does not capture timezone. The `onboarding`/
`gamification` modules remain transaction-scoped (no own-transaction opens,
DOC-06 §3).

## VOC-031-AC-02 — Onboarding API and `/onboarding` frontend flow

- Requirement source: `VOC-031-D00`, `VOC-031-D01`, DOC-03 §3, DOC-07
- Tasks: `VOC-031-T01`
- Tests: `VOC-031-TEST-06`..`VOC-031-TEST-11`
- Evidence: `VOC-031-EV-06`..`VOC-031-EV-11`
- Result: pending

`GET /api/v1/onboarding` and the submit endpoint are requester-scoped with
explicit DTOs, committed OpenAPI, and matched `@vocanova/api-client`
methods. A successful submission upserts all five DOC-03 §3 answers, sets
`users.onboarding_status='completed'`, and triggers `AC-01`'s seeding — all
in one transaction. An onboarding-incomplete learner visiting any `(app)`
route is redirected to `/onboarding`; an onboarding-complete learner
visiting `/onboarding` is redirected to `/home`. Unauthenticated → 401; an
invalid enum value → 400. No cross-user exposure.

## VOC-031-AC-03 — Settings and account write endpoints (`D06`)

- Requirement source: `VOC-031-D00`, `VOC-031-D02`, `VOC-031-D06`, DOC-07,
  DOC-08
- Tasks: `VOC-031-T02`
- Tests: `VOC-031-TEST-12`..`VOC-031-TEST-19`
- Evidence: `VOC-031-EV-12`..`VOC-031-EV-19`
- Result: pending — `D06` open

`GET`/`PATCH /api/v1/settings` and `GET`/`PATCH /api/v1/account` exist,
require `RequireAuth()`, and every unsafe method requires `CSRFMiddleware`
and a valid `Idempotency-Key` header, matching the existing P1/P2 write
pattern exactly. Per the proposed `D06` boundary: `/settings` covers
`dailyReviewTarget`/`reviewIntervalPreset`/`appLanguage`/
`notificationsEnabled`/`marketingEmailsEnabled` only; `/account` covers
`displayName` only; no endpoint exists for email-address change or account
deletion. Both are requester-scoped (no ID parameter to enumerate);
unauthenticated → 401; an out-of-range or malformed value → 400; a
cross-user request is never reachable. No existing A1/P4 internal
settings-resolution behavior (`VOC-030-D01`'s chain) is changed by adding
this public surface on top of it.

## VOC-031-AC-04 — Settings and account frontend wiring

- Requirement source: `VOC-031-D00`, `VOC-031-D02`, DOC-08
- Tasks: `VOC-031-T03`
- Tests: `VOC-031-TEST-20`..`VOC-031-TEST-23`
- Evidence: `VOC-031-EV-20`..`VOC-031-EV-23`
- Result: pending

`/settings` and `/settings/account` render real data from `AC-03`'s
endpoints via `@vocanova/api-client` under the A1 session, using the shared
loading/empty/error components (`AC-05`). A failed save leaves the
learner's attempted value visible with a clear error, never a silently
reverted or silently accepted value. No client DB access or duplicated
authorization.

## VOC-031-AC-05 — Cross-feature integration and navigation consistency

- Requirement source: `VOC-031-D00`, DOC-03 §4, DOC-08
- Tasks: `VOC-031-T04`
- Tests: `VOC-031-TEST-24`..`VOC-031-TEST-28`
- Evidence: `VOC-031-EV-24`..`VOC-031-EV-28`
- Result: pending

Every `(app)`/`(onboarding)` route uses one shared loading/empty/error
component set. Every DOC-03 §4 core-loop transition (Home → Discover/
Review/Sentence practice/Progress/Settings and back) is reachable via
client-side navigation without a dead end. Home's `dueReviewWords` count and
the Review screen's own due-queue read never disagree for the same
requester at the same instant. No backend business-logic regression in
P1–P4.

## VOC-031-AC-06 — Reliability and recovery across the loop

- Requirement source: `VOC-031-D00`, DOC-08, DOC-12 §5 P3
- Tasks: `VOC-031-T05`
- Tests: `VOC-031-TEST-29`..`VOC-031-TEST-34`
- Evidence: `VOC-031-EV-29`..`VOC-031-EV-34`
- Result: pending

A simulated mid-flight network failure on any P1/P2/P3/`AC-03` unsafe write
leaves the learner able to retry with no duplicate side effect (a stale
`Idempotency-Key` is never silently resubmitted in a way that could mask a
real duplicate). A simulated expired session mid-flow redirects to sign-in
and returns the learner to their prior route on success where reasonably
possible. A simulated slow/failed AI-feedback call does not block the
review/mission loop. No backend business-logic change beyond retry-safety
wiring.

## VOC-031-AC-07 — Accessibility automation and WCAG 2.2 AA pass (`D03`)

- Requirement source: `VOC-031-D00`, `VOC-031-D03`, `VOC-031-D05`, DOC-08
- Tasks: `VOC-031-T06`
- Tests: `VOC-031-TEST-35`..`VOC-031-TEST-38`
- Evidence: `VOC-031-EV-35`..`VOC-031-EV-38`
- Result: pending

`playwright` and `@axe-core/playwright` are installed; an automated a11y
sweep runs in CI over every `(public)`/`(onboarding)`/`(app)` route at
360px, 430px, and ≥1024px (`D05`). The sweep reports zero unresolved
violations against WCAG 2.2 AA for every route it covers, including
`/onboarding`, `/settings`, and `/settings/account`. Any route the
automation cannot reach is recorded as an explicit limitation, never
reported as passing. No manual-only fallback exists for this criterion.

## VOC-031-AC-08 — Performance automation and Lighthouse thresholds (`D04`)

- Requirement source: `VOC-031-D00`, `VOC-031-D04`, `VOC-031-D05`, DOC-08
- Tasks: `VOC-031-T07`
- Tests: `VOC-031-TEST-39`..`VOC-031-TEST-41`
- Evidence: `VOC-031-EV-39`..`VOC-031-EV-41`
- Result: pending

Lighthouse CI runs in CI against a production build for `/home`,
`/discover`, `/reviews`, `/progress`, `/settings`, and `/onboarding`, and
fails the build if Performance < 85, Accessibility < 95, or Best Practices
< 90 on any of them. All six routes pass all three thresholds at the time
of `T07`'s evidence collection. No manual spot-check-only fallback exists
for this criterion.

## VOC-031-AC-09 — Final UX consistency across P1–P5

- Requirement source: `VOC-031-D00`, `VOC-031-D05`, `VOC-031-D08`, DOC-08
- Tasks: `VOC-031-T08`
- Tests: `VOC-031-TEST-42`..`VOC-031-TEST-44`
- Evidence: `VOC-031-EV-42`..`VOC-031-EV-44`
- Result: pending

Every P1–P5 screen uses consistent design tokens (no ad hoc spacing/
typography/color), consistent 44px-minimum touch targets, consistent
focus-visible styling, and the `AC-05` shared loading/empty/error
components, verified at 360px/430px/≥1024px. No already-shipped route is
renamed or restructured (`D08` default); the found DOC-08-vs-implementation
routing drift is recorded, not resolved, by this criterion.

## VOC-031-AC-10 — Evidence, mock-inventory, staging, rollback, and P5 gate readiness

- Requirement source: `VOC-031-D00`, DOC-12 §5 P5
- Tasks: `VOC-031-T00`..`VOC-031-T09`
- Tests: `VOC-031-TEST-45`..`VOC-031-TEST-48`
- Evidence: `VOC-031-EV-45`..`VOC-031-EV-48`
- Result: pending — in-repository evidence only; live staging blocked until
  F3 exists

Applicable checks, the deterministic tests this package adds (migration,
transaction, contract, accessibility, performance, reliability,
consistency), exact-SHA reviews, and the extended mock-inventory test pass;
mock-inventory verifies no P4-and-earlier mock regressed and no R1/R2
behavior was invented. Staging exercises for the full onboard → discover →
save → review → complete → sentence-practice → progress → settings loop at
all three supported layouts, cross-user denial, and the one-new-table
rollback rehearsal are documented and ready to run once F3 staging exists
(`VOC-031-DEP-02`). This enables — but does not itself declare — the DOC-12
P5 gate evaluation; the milestone gate is not satisfied by package merge or
staging deploy alone, and `D06`–`D08` must be resolved before the affected
tasks' evidence can be treated as final.
