# VOC-031 — Tasks

Ordered PR sequence: `T00 → T01 → T02 → T03 → T04 → T05 → T06 → T07 → T08 → T09
→ T10 → T11`. Each PR is independently reviewable, remains R3-proposed (path
floor R3 for migrations/schemas/auth-adjacent modules), and requires Claude
Code exact-SHA review. `T01` is additionally blocked on `VOC-031-D03`'s
resolution (`VOC-031-DEP-05`) — it may not proceed on a guessed answer to
whether the onboarding gate applies retroactively to pre-existing accounts.

## VOC-031-T00 — `user_onboarding_profiles` schema, migration, and onboarding domain logic

- Requirement source: `VOC-031-D00`, `VOC-031-D01`, `VOC-031-D02`, `VOC-031-D04`, DOC-05 §6, DOC-03 §3
- Acceptance criteria: `VOC-031-AC-00`
- Tests: `VOC-031-TEST-00`..`VOC-031-TEST-03`
- Evidence: `VOC-031-EV-00`..`VOC-031-EV-03`
- Status: pending

Add the `user_onboarding_profiles` Ent schema + reviewed versioned Atlas SQL
(DOC-05 §6: `user_id` unique FK, `english_level`, `native_language`,
`learning_goal`, `main_use_case`, `daily_review_target` 5–100,
`completed_at`), in a new `apps/api/business/users` module. Implement the
`VOC-031-D04` seed-eligibility rule as a pure, unit-tested domain function (no
Ent/Huma dependency): no existing `user_settings` row → seed freely; existing
row with `daily_review_target == 20` (schema default) → overwrite; existing
row with any other value → never overwrite. No API route, no frontend, no
gating logic in this PR.

## VOC-031-T01 — Onboarding API, frontend flow, and gating

- Requirement source: `VOC-031-D00`, `VOC-031-D02`, `VOC-031-D03`, DOC-03 §3
- Acceptance criteria: `VOC-031-AC-01`
- Tests: `VOC-031-TEST-04`..`VOC-031-TEST-07`
- Evidence: `VOC-031-EV-04`..`VOC-031-EV-07`
- Status: pending — blocked on `VOC-031-D03` resolution (`VOC-031-DEP-05`)

Add `GET`/`POST /api/v1/onboarding` (`GetOnboarding`/`CompleteOnboarding`) to
the new `users` module: `POST` validates and persists the five answers,
calls `T00`'s seed function, and sets `users.onboarding_status='completed'`.
Additively extend `GET /api/v1/me`'s response with an `onboardingStatus`
field (no other field changes; existing consumers of `/me` are unaffected).
Add a new `apps/web/src/app/onboarding` route: English level, native
language, learning goal, main use case, and daily review target, collected
client-side across screens and submitted once (`VOC-031-D02` — no
resumability). Extend `apps/web/src/middleware.ts`'s matcher and redirect
logic so an authenticated learner whose `onboardingStatus` is not
`completed` is redirected to `/onboarding` from any other `(app)` route, and
away from `/onboarding` to `/home` once `completed`. Implement the gate per
the `VOC-031-D03` default (applies to every account, including pre-existing
ones) unless the founder has overridden it at adoption.

## VOC-031-T02 — Settings backend: read/write `user_settings` and `displayName`

- Requirement source: `VOC-031-D00`, `VOC-031-D01`, `VOC-031-D06`, DOC-00 §3, DOC-05 §6
- Acceptance criteria: `VOC-031-AC-02`
- Tests: `VOC-031-TEST-08`..`VOC-031-TEST-11`
- Evidence: `VOC-031-EV-08`..`VOC-031-EV-11`
- Status: pending

Add `GET`/`PATCH /api/v1/settings` (`GetSettings`/`UpdateSettings`) to the
`users` module, reading/writing exactly `dailyReviewTarget`,
`reviewIntervalPreset`, `appLanguage`, `notificationsEnabled`,
`marketingEmailsEnabled` (on the existing `user_settings` row VOC-030
created) and `displayName` (on `users`). Partial-update (`PATCH`) semantics;
strict validation (reject unknown fields, `dailyReviewTarget` 5–100,
`reviewIntervalPreset` enum, `appLanguage` restricted to `en` only per
`VOC-031-D06`, `displayName` length limit). Upsert the `user_settings` row on
first-ever write (handles a concurrent first read/write race with
`gamification`'s existing lazy-creation path — `VOC-031-R05`). No change to
`gamification`'s existing direct `user_settings` reads (accepted, documented
asymmetry — not refactored by this package). Confirm structurally that a
`dailyReviewTarget` change never rewrites an already-created
`daily_mission_snapshots.review_target` for the current local day (DOC-00
§3's "applies from the next local day" rule).

## VOC-031-T03 — Email-change verification backend

- Requirement source: `VOC-031-D01`, `VOC-031-D05`, `VOC-031-D09`, DOC-06 §6
- Acceptance criteria: `VOC-031-AC-03`
- Tests: `VOC-031-TEST-12`..`VOC-031-TEST-17`
- Evidence: `VOC-031-EV-12`..`VOC-031-EV-17`
- Status: pending

Add the `email_change_links` Ent schema + migration (mirrors `magic_links`:
`user_id` not-nullable, `new_email`, `token_hash` unique, `environment`,
`created_at`/`expires_at`/`consumed_at`/`revoked_at`), in a new
`apps/api/business/accounts` module. Add
`POST /api/v1/settings/email-change-links` (`RequestEmailChangeLink`,
requires an authenticated session, generic success response regardless of
new-email registration status, rate-limited by IP and session) and
`POST /api/v1/settings/email-change-links/consume` (`ConsumeEmailChangeLink`,
validates token hash/expiry/single-use/environment, re-checks new-email
uniqueness atomically at confirm time, updates `users.email`, sends a
best-effort old-email notification via the existing `email.Sender`, never
invalidates the current session). Reuse
`apps/api/business/auth`'s existing `RateLimiter`/`KeyForIP`/`KeyForSession`
and token-hash primitives; do not modify `auth`'s existing magic-link/session
code.

## VOC-031-T04 — Account-deletion backend

- Requirement source: `VOC-031-D01`, `VOC-031-D07`, `VOC-031-D09`, DOC-05 §16, DOC-06 §§9,14,15, DOC-07
- Acceptance criteria: `VOC-031-AC-04`
- Tests: `VOC-031-TEST-18`..`VOC-031-TEST-23`
- Evidence: `VOC-031-EV-18`..`VOC-031-EV-23`
- Status: pending

Add the `account_deletion_requests` Ent schema + migration (`user_id` unique,
`status` `deactivated`/`anonymizing`/`completed`, `requested_at`,
`purge_after`, `completed_at`) in `accounts`. Add
`account_deletion` to `idempotency_keys.scope`'s check constraint
(`VOC-031-D09`). Add `POST /api/v1/account-deletion-requests`
(`CreateAccountDeletionRequest`, `RequireAuth`, `Idempotency-Key` required per
DOC-07): inside one transaction, set `users.status='deleted'`/`deleted_at`,
revoke every active session and every unconsumed magic-link/email-change-link
for the account, and insert the `account_deletion_requests` row
(`status='deactivated'`, `purge_after = requested_at + 30 days`). Implement
the idempotent, resumable anonymization sweep (`VOC-031-D07`, mirrors
`apps/api/business/auth`'s existing `Cleanup()` pattern — no new queue) that
processes any row past `purge_after`: soft-delete-pending-purge for
`external_identities`/`user_words`/`learner_sentences`; irreversible
de-identification for `review_attempts`/`ai_feedback_attempts`/
`confidence_point_ledger`/`grace_day_ledger`/`feature_audit_logs`;
delete-or-de-identify for `user_onboarding_profiles`/`user_settings`/
`daily_mission_snapshots`/`daily_activity_summaries`/`streak_states`
(DOC-05 §16); transitions the row to `status='completed'`. No production
enablement is authorized by this task (`VOC-031-DEP-03`).

## VOC-031-T05 — Settings/account frontend

- Requirement source: `VOC-031-D00`, `VOC-031-D01`, `VOC-031-D05`, `VOC-031-D06`, DOC-08
- Acceptance criteria: `VOC-031-AC-05`
- Tests: `VOC-031-TEST-24`..`VOC-031-TEST-27`
- Evidence: `VOC-031-EV-24`..`VOC-031-EV-27`
- Status: pending

Add `apps/web/src/app/(app)/settings` and
`apps/web/src/app/(app)/settings/account`, wiring `T02`–`T04`: editable
Settings form (all six fields, `appLanguage` presented honestly as a single
confirmed option per `VOC-031-D06`, not a working language picker);
email-change flow (request form → pending-verification state → confirm
handling, matching DOC-03 §9's loading/empty/error conventions, input
preserved on error, no duplicate submission); account-deletion flow (explicit
confirmation step — e.g. a typed confirmation phrase, not a single click —
followed by immediate logout and a clear post-deletion message, never a
silent success with no feedback). Add `/settings`/`/settings/account` to
`apps/web/src/middleware.ts`'s auth-gate matcher. No client-fabricated
fallback value on any of these screens.

## VOC-031-T06 — Cross-feature reliability and recovery pass

- Requirement source: `VOC-031-D00`, DOC-12 §5 P5, DOC-03 §9
- Acceptance criteria: `VOC-031-AC-06`
- Tests: `VOC-031-TEST-28`..`VOC-031-TEST-31`
- Evidence: `VOC-031-EV-28`..`VOC-031-EV-31`
- Status: pending

Audit every core-loop screen (Home, Discover, Word Detail, Reviews, sentence
feedback, Progress, Onboarding, Settings, Settings/account) for: a safe retry
path on network failure that never loses learner input and never falsely
implies completion; correct behavior when a session expires mid-flow (e.g.
mid-review-session, mid-sentence-submission, mid-onboarding) — a clear
re-authentication path, no silently-lost progress claimed as saved when it
was not; and no client-fabricated fallback value anywhere. Add regression
tests proving every already-shipped A1–P4 screen's existing behavior is
unchanged (`VOC-031-R06`, mirrors `VOC-030-R01`'s "byte-for-byte unchanged"
requirement). Fix any gap found; record what was found and fixed, not merely
asserted as fine.

## VOC-031-T07 — Accessibility automation (axe-core + Playwright)

- Requirement source: `VOC-031-D00`, `VOC-031-D01`, DOC-03 §10, DOC-08 quality standards
- Acceptance criteria: `VOC-031-AC-07`
- Tests: `VOC-031-TEST-32`..`VOC-031-TEST-35`
- Evidence: `VOC-031-EV-32`..`VOC-031-EV-35`
- Status: pending

Install `@playwright/test` and an axe-core integration (net-new — confirmed
absent repository-wide, `VOC-031-D00`). Add `apps/web/tests/e2e/` (the path
`docs/design/08-web-app-design.md`'s architecture section already documents
but that has never been created). Add an automated accessibility scan
covering every core-loop screen at the three supported layouts (360px,
430px, and one representative desktop width ≥1024px): zero critical/serious
axe violations is the pass bar; non-color-only feedback and keyboard
reachability are asserted explicitly, not only inferred from a clean axe
run. Wire the suite into CI as a new job (DOC-10 §7 Level 2 — "selected
Playwright"). Document CI runner/browser-dependency reconciliation
(`VOC-031-DEP-04`).

## VOC-031-T08 — Full core-loop end-to-end Playwright suite

- Requirement source: `VOC-031-D00`, DOC-10 §7
- Acceptance criteria: `VOC-031-AC-08`
- Tests: `VOC-031-TEST-36`
- Evidence: `VOC-031-EV-36`
- Status: pending

On the `T07` Playwright install, build the DOC-10 §7-documented end-to-end
flow for the first time: auth → onboarding → discover → save → review
session → sentence submission → deterministic AI feedback (mock provider,
never a paid/nondeterministic call in CI, per DOC-10 §7) → progress update →
settings change → logout → unauthenticated-access rejection. This is a
functional-correctness suite, distinct from `T07`'s accessibility scans.

## VOC-031-T09 — Performance automation (Lighthouse CI)

- Requirement source: `VOC-031-D00`, `VOC-031-D01`, DOC-08 quality standards
- Acceptance criteria: `VOC-031-AC-09`
- Tests: `VOC-031-TEST-37`..`VOC-031-TEST-39`
- Evidence: `VOC-031-EV-37`..`VOC-031-EV-39`
- Status: pending

Install Lighthouse CI (net-new — confirmed absent repository-wide,
`VOC-031-D00`), configured against a production build (`next build` +
serve, not the dev server, to avoid hot-reload variance —
`VOC-031-R04`) at the three supported layouts, asserting the DOC-08
thresholds (Performance 85+, Accessibility 95+, Best Practices 90+) on Home,
Discover, Reviews, and Progress. Wire into CI; document the flakiness
mitigation (fixed local production build, not a live network target) and any
threshold not yet met as an explicit, honestly-reported limitation, never a
pass.

## VOC-031-T10 — Final UX-consistency pass

- Requirement source: `VOC-031-D00`, DOC-03 §1, §11
- Acceptance criteria: `VOC-031-AC-10`
- Tests: `VOC-031-TEST-40`
- Evidence: `VOC-031-EV-40`
- Status: pending

Design-principle audit against DOC-03 §1 (one clear action per screen,
practical-over-academic framing, encouraging non-gamified tone, mobile-first)
and §11 (calm visual tone, no "grading" patterns) across every core-loop
screen, old and new. This is a manual/design review that automation (`T07`,
`T09`) cannot perform — record findings (screen, principle, gap) and the fix
applied for each, not a bare "reviewed, looks fine" assertion.

## VOC-031-T11 — Evidence, mock-inventory, staging evidence, and P5 gate readiness

- Requirement source: `VOC-031-D00`, DOC-12 §5 P5
- Acceptance criteria: `VOC-031-AC-11`
- Tests: `VOC-031-TEST-41`..`VOC-031-TEST-45`
- Evidence: `VOC-031-EV-41`..`VOC-031-EV-45`
- Status: pending

Run every installed relevant check; update the deterministic mock-inventory
check (`scripts/foundation/mock-inventory.mjs`) to admit the new
`users`/`accounts` modules/routes/schemas/migrations and confirm no P5-invented
route/table/behavior beyond this package's own scope. Collect the
mock-decommission inventory (expected to be near-empty — see
`mock-inventory.md`), staging evidence (following the exact
`VOC-025-DEP-01`/`VOC-030-DEP-02` "blocked-with-documented-procedure"
pattern for `VOC-031-DEP-02`), rollback rehearsal for the three new tables,
and P5 gate readiness. Do not declare the DOC-12 P5 gate complete.

### Deliverables

- `mock-inventory.md`: confirms zero pre-existing mocks remain to retire and
  records any new mocks this package's own tasks temporarily introduce (if
  any) plus their disposition.
- `staging-evidence.md`: collects in-repository evidence and documents the
  staged exercises and rollback rehearsal that can only run once F3 exists.
- updated `scripts/foundation/mock-inventory.mjs` (+`.test.mjs`).

### Blocker

`VOC-031-DEP-02` remains open: F3 staging does not exist, so the live
staging exercises cannot be executed. This task provides the procedures and
the in-repository evidence only; it does not declare the DOC-12 P5 gate
complete.
