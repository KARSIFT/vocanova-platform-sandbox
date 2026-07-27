# VOC-031 — Begin Milestone P5: Integrated Core Loop: Specification

## Objective and requirement source

Deliver the DOC-12 §5 P5 gate: combine everything built by A1 (VOC-025), P1
(VOC-026), P2 (VOC-027), P3 (VOC-028), and P4 (VOC-030) into one coherent,
reliable, mobile-first journey, so that **the full loop works coherently in
staging across supported layouts with no critical product/security/data/
accessibility/reliability defect.** This package (a) builds the two remaining
DOC-03/DOC-08-scoped screens that no prior milestone built — onboarding
(DOC-03 §3) and Settings/account (DOC-08 routing table, `user_settings`
read/write on top of the table VOC-030 already created) — as real,
authenticated, tested frontend+backend work, per the supplied request's
founder decisions; (b) closes cross-feature reliability/recovery gaps across
the whole loop; (c) installs automated accessibility testing (axe-core +
Playwright) and automated Lighthouse CI performance budgeting, both net-new to
this repository; and (d) performs a final UX-consistency pass against DOC-03's
principles. Authority: DOC-12 §5 (P5 paragraph), DOC-03 (onboarding flow §3,
UX principles §1, accessibility §10), DOC-08 (routing table, Settings/account
scope, quality standards/Lighthouse thresholds), DOC-05 §§6,13,16 (`user_settings`,
`user_onboarding_profiles`, idempotency, account deletion/anonymization),
DOC-06 §§6,7,9,14,15 (auth/magic-link conventions, authorization, idempotency,
account deletion, no-queue constraint), DOC-07 (API/DTO rules, idempotency-required
endpoints including the already-named `POST /api/v1/account-deletion-requests`),
DOC-10 §7 (end-to-end testing strategy, which already names "settings change" and
"logout" as part of the documented core-loop E2E test), DOC-11 §3 (kill switches),
and the supplied request (which records five founder decisions, listed in
`VOC-031-D01`, and a fixed definition of "supported layouts": 360px, 430px, and
one representative desktop width ≥1024px). A1/P1/P2/P3/P4 carry forward
unchanged except where this package adds new reliability/consistency work on
top of already-shipped screens, and except for one small, explicitly additive
change to the already-shipped `GET /api/v1/me` response (`VOC-031-AC-02`).

## Scope and non-goals

In scope, as a fixed ordered task sequence (`T00`–`T11`, detailed in
`tasks.md`):

1. **Onboarding backend** (`T00`): `user_onboarding_profiles` (DOC-05 §6)
   schema + migration, pure domain logic for the seed-eligibility rule
   (`VOC-031-D04`), and `GET`/`POST /api/v1/onboarding` in a new
   `apps/api/business/users` module.
2. **Onboarding frontend** (`T01`): a new `/onboarding` route (English level,
   native language, learning goal, main use case, daily review target —
   DOC-03 §3's "handful of onboarding questions"), single-submit (not
   resumable — `VOC-031-D02`), gated so an authenticated learner whose
   `onboarding_status` is not `completed` is routed there before any other
   `(app)` route, and routed away from `/onboarding` once completed. Extends
   the already-shipped `GET /api/v1/me` response with an additive
   `onboardingStatus` field for this gating (`VOC-031-AC-02`).
3. **Settings backend** (`T02`): `GET`/`PATCH /api/v1/settings` in the new
   `users` module, exposing exactly the founder-directed field set —
   `dailyReviewTarget`, `reviewIntervalPreset`, `appLanguage`,
   `notificationsEnabled`, `marketingEmailsEnabled`, `displayName` — reading
   and writing the `user_settings` row VOC-030 already created (schema-complete
   but only `timezone`/`daily_review_target` previously wired) plus
   `users.display_name`. Per DOC-00 §3, a `dailyReviewTarget` change takes
   effect from the next local day, never retroactively — satisfied structurally
   because `daily_mission_snapshots.review_target` is copied at
   snapshot-creation time and is never rewritten by a later settings change.
4. **Email-change verification backend** (`T03`): a new `email_change_links`
   table (mirrors the existing `magic_links` token pattern exactly — see
   `VOC-031-D05`), `POST /api/v1/settings/email-change-links` (request) and
   `POST /api/v1/settings/email-change-links/consume` (confirm) in the new
   `apps/api/business/accounts` module.
5. **Account-deletion backend** (`T04`): a new `account_deletion_requests`
   table, `POST /api/v1/account-deletion-requests` (already named by DOC-07)
   in `accounts`, implementing DOC-05 §16's real deactivate-then-staged-purge
   workflow (`VOC-031-D07`) — not a soft no-op.
6. **Settings/account frontend** (`T05`): `/settings` and `/settings/account`
   routes wiring `T02`–`T04`.
7. **Cross-feature reliability and recovery pass** (`T06`): audit and close
   gaps in loading/empty/error/recovery states across the full loop (Home,
   Discover, Word Detail, Reviews, sentence feedback, Progress, the two new
   Settings routes, Onboarding) for network failure, session expiry mid-flow,
   and safe, non-duplicating retry — without regressing any already-accepted
   A1–P4 screen behavior (`VOC-031-R06`).
8. **Accessibility automation** (`T07`): install Playwright + an axe-core
   integration (net-new to this repository — confirmed absent, see
   `VOC-031-D00`) and add an automated accessibility test suite across the
   core-loop screens at the three supported layouts, wired into CI.
9. **Full core-loop end-to-end suite** (`T08`): the DOC-10 §7-documented
   Playwright flow — auth → onboarding → discover → save → review session →
   sentence submission → deterministic AI feedback → progress update →
   settings change → logout → unauthenticated-access rejection — built for
   the first time on the `T07` Playwright install.
10. **Performance automation** (`T09`): install Lighthouse CI (net-new,
    confirmed absent) wired to the DOC-08 thresholds (Performance 85+,
    Accessibility 95+, Best Practices 90+) against a production build at the
    three supported layouts.
11. **Final UX-consistency pass** (`T10`): a design-principle audit against
    DOC-03 §1 (one clear action per screen, encouraging non-gamified tone, no
    color-only feedback, 44px touch targets) across every core-loop screen,
    old and new, recorded with findings and fixes, not merely asserted.
12. **Evidence, mock-inventory, staging evidence, and P5 gate readiness**
    (`T11`).

Out of scope (do not invent): renaming any already-shipped route to match
DOC-08's documented table, or building the documented-but-never-shipped
`/words`/`/words/[userWordId]` routes (`VOC-031-D00`/`D08` — informational
carry-forward only, per the supplied request's explicit instruction not to act
on this here); a general Settings API surface beyond the founder-directed
field list; `appLanguage` actually changing rendered UI language (no i18n
infrastructure exists anywhere in this repository today — see
`VOC-031-D06`); production deployment or real secrets; onboarding resumability/
partial-save; leaderboards, badges, social challenges, rewards store (DOC-12
§10); R1/R2/L1 staging/production/launch-readiness work itself (this package
only produces the P5 gate's own evidence).

## Risk and protected areas

Proposed **R3** (proposal only — not a determination), matching the
established precedent for milestones that touch authentication, migrations,
and sensitive per-user data without deciding a new pricing/legal/public-launch
matter (VOC-025's A1 foundation and VOC-030's ledger/gamification work were
both R3 for the same reason). Protected paths: `/apps/api/migrations` and
`/apps/api/ent/schema` (three new tables — R3 path floor); the new
`apps/api/business/users` and `apps/api/business/accounts` modules
(first-party correctness for a real, irreversible account-deletion procedure
and an email-change flow that touches the primary auth-adjacent identifier);
`apps/api/foundation/auth`'s token/session conventions, which `accounts`
reuses but does not modify; and — as with VOC-030 — the already-shipped
`GET /api/v1/me` response this package additively extends. Under A-003,
routine R3 needs strengthened controls and exact-SHA independent verification,
not standing steward/founder approval solely for being R3; R4 founder
authority is unchanged. One specific consequence inside this otherwise-R3
package is flagged for the future reviewer/founder rather than silently
absorbed into "routine R3": **account deletion is, by its nature, a
difficult-to-reverse action against real learner data** (`change-risk-classification.md`'s
R4 test 1). This package builds and tests the mechanism against non-production
data only; *production enablement* of account deletion specifically — as
distinct from this package's own merge/staging work — warrants an explicit
founder go/no-go at the future production-activation decision, on top of the
ordinary R2/L1 production-readiness gate and the DOC-05 §16 "subject to legal
review before production" qualifier that document already carries. This draft
does not decide that; it records the need so it is not missed later
(`VOC-031-DEP-04`).

## Decisions, contradictions, security, and privacy

`VOC-031-D00` — **Carry-forward confirmation (confirmed at draft time,
2026-07-27, by direct repository inspection).** `user_settings`
(`apps/api/ent/schema/usersettings.go`, migration
`20260725130000_voc030_p4_user_settings.sql`) exists with the full DOC-05 §6
shape (`timezone`, `daily_review_target`, `review_interval_preset`,
`notifications_enabled`, `marketing_emails_enabled`, `app_language`), but only
`timezone`/`daily_review_target` are read/written today (by `gamification`,
for mission/streak resolution) — the other four columns hold only their
schema defaults, unread and unwritten by any code, exactly as VOC-030's own
`staging-evidence.md` "Follow-up work" section anticipated ("A future
Settings package: expose `user_settings` through a real API/UI"). `users.email`,
`users.display_name`, and `users.onboarding_status`
(`not_started`/`in_progress`/`completed`) already exist in
`apps/api/ent/schema/user.go` and are reused, not re-created.
`user_onboarding_profiles` (DOC-05 §6) does **not** exist — no schema, no
migration. No `settings` or `accounts` business module, and no
`GET/PATCH /api/v1/settings`, `/api/v1/onboarding`, or
`/api/v1/account-deletion-requests` route, exists anywhere in
`apps/api/app/api` or the committed OpenAPI document today — this is
genuinely greenfield backend work, not a re-wiring of an existing stub.
Account deletion has **zero implementation** anywhere in the codebase; every
reference to it is documentation-only (DOC-05 §16, DOC-06 §14, DOC-07,
DOC-08). Frontend: the actually-shipped route tree (`apps/web/src/app/`) is
`/`, `/signin`, `/auth/magic`, and, inside the authenticated `(app)` group,
`/home`, `/discover` (+ `[situation]`, `[situation]/[word]`), `/progress`,
`/reviews` (review session is an inline component, not a `/reviews/session`
sub-route) — there is no `/onboarding`, `/settings`, `/settings/account`, or
`/words` route today. This is a real, confirmed drift against DOC-08's
documented table (`/login`, `/magic-link`, `/review`, `/review/session`,
`/words`, `/words/[userWordId]`), noted informationally, as instructed, and
**not acted on** — no existing route is renamed and no `/words` route is built
by this package (`VOC-031-D08`). `apps/web/src/middleware.ts`'s auth-gate
matcher covers only `/home`, `/discover(/:path*)`, `/progress`,
`/reviews(/:path*)` — it does not yet cover any route this package adds, and
will need to. Zero `MOCK_*` constants remain anywhere in `apps/web/src` — P1–P4's
mock retirement is complete; this package has no legacy mock to retire, only
new-screen mock discipline to enforce going forward (`mock-inventory.md`).
Finally: no Playwright, no axe-core (or any axe integration), and no
Lighthouse CI configuration exists anywhere in this repository — the only
existing test tooling is Node's built-in `node:test` runner over
`scripts/foundation/*.test.mjs` plus `go test`. Installing browser-driven
accessibility and performance automation is genuinely new CI tooling, not
incremental wiring onto an existing harness.

`VOC-031-D01` — **Founder-directed scope (supplied as already decided in the
request; recorded here as given, not proposed by this draft).** Onboarding
(DOC-03 §3) and a Settings/account screen (DOC-08 routing table) are in MVP
scope now, built as real new tasks — frontend and backend, including new
write endpoints — following this repository's existing auth/session/security
conventions. The editable Settings/account field scope is exactly:
`dailyReviewTarget`, `reviewIntervalPreset`, `appLanguage`,
`notificationsEnabled`, `marketingEmailsEnabled`, `displayName`, email-address
change (with its own magic-link-style verification flow, matching this
repository's existing conventions, that must not weaken session/auth
guarantees), and account deletion (a real, safe procedure — confirmation
step, cascading/anonymizing removal of the learner's own data consistent with
DOC-05, session invalidation — not a fabricated soft no-op; if a genuinely
safe procedure could not be scoped, this draft was instructed to record that
as an explicit blocker/DEP instead of shipping something unsafe or fake).
Onboarding-seeds-`user_settings.daily_review_target` only when no
already-customized row exists (`VOC-031-D04`). "Supported layouts" is fixed as
360px, 430px, and one representative desktop width ≥1024px — not an open
decision. This draft records these as supplied; the human adopting this
package is the one who confirms the request accurately reflects actual
founder authority, the same as for any other requirement source.

`VOC-031-D02` — **This draft's own minimal-scope proposal (needs adoption-time
confirmation, distinct from `D01`'s founder-supplied items).** Onboarding is
single-submit and non-resumable in MVP: the frontend collects English level,
native language, learning goal, main use case, and daily review target across
its screens client-side and submits once via `POST /api/v1/onboarding`, which
sets `onboarding_status='completed'` directly. `onboarding_status='in_progress'`
is not actively used by this package (the enum value already exists in
`users.status` from before this package and is left alone, not deleted).
Chosen over a partial-save/resume design because DOC-03 §3 describes "a
handful of ... questions" as a low-friction, short sequence, and building
resumability (a partial-profile row, a resume-from-last-answered-question
flow) adds meaningfully more surface for a five-question form with no stated
requirement for it.

`VOC-031-D03` — **Genuine new open decision, not resolved by this draft
(flagged per the request's explicit instruction to surface new
ambiguities).** Every account that exists in this repository today has
`onboarding_status='not_started'`, because onboarding has never existed until
this package — that includes every non-production test/staging identity
created during the VOC-025/026/027/028/030 evidence-collection work. A naive
"redirect to `/onboarding` whenever `onboarding_status != 'completed'`" gate
(as `T01` implements by default) would therefore force **every existing
account**, including those staging/test identities, through onboarding before
they can reach Home/Discover/Reviews/Progress again — potentially
invalidating the repeatability of those milestones' own documented staging
procedures (e.g. VOC-025's `staging-evidence.md` procedures assume direct
navigation to `/home` after auth). Two resolutions are possible and this
draft does not choose between them: **(a)** apply the gate uniformly to every
account, including pre-existing ones (matches DOC-03 §3's flow literally, but
means every prior milestone's staging procedure must additionally run
onboarding first once F3 exists); or **(b)** grandfather accounts that
existed before this package's activation (e.g. accounts with `created_at`
before the migration lands, or a one-time backfill marking pre-existing
accounts `onboarding_status='completed'`) so only genuinely new sign-ups see
the gate. `T01`'s tasks/acceptance-criteria are written against **(a)** as the
literal reading of DOC-03 §3 and DOC-12's "no retroactive credit" precedent
(`VOC-030-D05`) — new gates apply going forward, not retroactively rewritten
history — but this is this draft's own default, not a founder decision, and
must be confirmed or overridden at adoption.

`VOC-031-D04` — **Seed-eligibility rule, restated precisely from `D01`'s
founder decision.** On a successful `POST /api/v1/onboarding`: if no
`user_settings` row exists yet for the user, create one with
`daily_review_target` set from the onboarding answer (no prior customization
is possible when no row exists — freely seed); if a row exists and its
`daily_review_target` still equals the schema default (`20`) exactly,
overwrite it with the onboarding answer; if a row exists with any other
value, never overwrite it. This is an exact restatement of the founder's rule
("seed only when no `user_settings` row with an already-customized (non-default)
`daily_review_target` exists"), not a new proposal.

`VOC-031-D05` — **Email-change verification design (mirrors the existing
magic-link conventions in `apps/api/business/auth`, per `D01`'s founder
instruction).** A new `email_change_links` table follows the exact
`magic_links` pattern: a 32-random-byte token, only its SHA-256 hash
persisted, single-use enforced via `consumed_at`, `environment`-scoped,
15-minute expiry (identical to the magic-link TTL), and rate-limited by both
IP and the requesting user's session (reusing the existing
`RateLimiter`/`KeyForIP`/`KeyForSession` primitives in
`apps/api/business/auth/rate.go`). It differs from a magic link in three
ways, all necessary because this is not a login mechanism: **(1)** requesting
a change requires an already-authenticated session (`user_id` is not
nullable, unlike `magic_links.email`); **(2)** a generic success response is
returned regardless of whether the requested new email is already registered
to another active account (same anti-enumeration posture as
`RequestMagicLink`), and the actual uniqueness check is re-verified
atomically at confirm time against the same `lower(email) where deleted_at is
null` partial unique index `users` already enforces — a confirm that would
violate it fails with a stable, non-500 conflict response, not a crash
(`VOC-031-R02`); **(3)** on successful confirmation, a best-effort,
non-blocking notification is sent to the *old* email address (using the
existing `email.Sender` interface) informing the account owner their sign-in
email changed — this is a required security control, not optional, because
the request step alone (which only needs a valid session, not
re-authentication) would otherwise let a hijacked session silently redirect
future magic-link logins to an attacker's address with no signal to the
legitimate owner. The learner's current session is **not** invalidated by
either step (matches "must not weaken session/auth guarantees" — changing the
login *address* is not the same as revoking the *session* already in use).
Google-OAuth-linked accounts are unaffected: OAuth login matches
`external_identities.provider_subject`, not `users.email`, so an email change
never breaks an existing Google sign-in.

`VOC-031-D06` — **`appLanguage` is a persisted-but-currently-inert preference,
not a functioning UI-language switch.** Confirmed by inspection: no i18n
library, translation catalog, or locale-routing logic exists anywhere in
`apps/web` today. `D01`'s founder decision requires `appLanguage` in the
editable field set, so `T02`/`T05` build a real, persisted `PATCH` write for
it — but this draft restricts the accepted value set to `en` only at launch
(matching the DOC-05 §6 default) and the frontend must present it honestly
(e.g. as a single confirmed option, not a working multi-language picker) so
the product never claims a capability (rendering the app in another
language) that does not exist. Expanding the accepted value set is future
scope once real i18n infrastructure exists, not invented here.

`VOC-031-D07` — **Account-deletion architecture under the "no queue system in
MVP" constraint (DOC-06 §15).** `POST /api/v1/account-deletion-requests`
performs, synchronously and inside one transaction: set `users.status='deleted'`,
`users.deleted_at=now()`, revoke every active session
(`apps/api/business/auth`'s existing `RevokeSession`, applied to all of the
user's sessions, not just the current one) and every unconsumed magic link
and email-change link for the account, and insert one
`account_deletion_requests` row (`status='deactivated'`, `requested_at`,
`purge_after` = `requested_at` + the DOC-05 §16 default 30-day target). The
actual per-table anonymization/deletion (DOC-05 §16's disposition list:
soft-delete-pending-purge for `users`/`external_identities`/`user_words`/
`learner_sentences`; irreversible de-identification for `review_attempts`/
`ai_feedback_attempts`/`confidence_point_ledger`/`grace_day_ledger`/
`feature_audit_logs`; delete-or-de-identify for
`user_onboarding_profiles`/`user_settings`/`daily_mission_snapshots`/
`daily_activity_summaries`/`streak_states`) runs from a new, idempotent,
resumable sweep function that mirrors the existing `Cleanup()` pattern already
used for expired sessions/magic links (`apps/api/business/auth/service.go`'s
`Cleanup()`) rather than introducing a job queue or worker service. The sweep
processes any `account_deletion_requests` row past its `purge_after` date (or,
for immediate non-production testing, any row explicitly forced past it),
performs the anonymization, and transitions it to `status='completed'`. No new
queue/worker infrastructure is introduced, consistent with DOC-06 §15 and the
DOC-12 §10 "unproven queues" exclusion. This mechanism is built and tested
against non-production data by this package in full; enabling it against real
production learner data additionally requires the DOC-05 §16 "subject to
legal review before production" sign-off and the founder go/no-go noted in
"Risk and protected areas" above — this package does not itself authorize
that (`VOC-031-DEP-04`).

`VOC-031-D08` — **Routing-drift carry-forward, informational only.** See
`D00`'s route-tree finding. Per the supplied request's explicit instruction,
this is recorded again here for completeness and **not acted on**: no
existing route is renamed, and `/words`/`/words/[userWordId]` are not built.
Only the genuinely new routes this package introduces (`/onboarding`,
`/settings`, `/settings/account`) are added, using DOC-08's exact documented
names, since there is no pre-existing shipped route at those paths to
conflict with.

`VOC-031-D09` — **`idempotency_keys.scope` enum gap (a genuine contradiction
between two approved documents, recorded per DOC-12 §11's change-control
rule, not silently resolved — mirrors the `VOC-030-D02` pattern).** DOC-05
§13's `idempotency_keys.scope` check-constraint enum
(`review_submission`/`daily_mission_completion`/`word_addition`/
`sentence_submission`/`ai_feedback_request`/`point_award`/
`grace_day_application`) has no value for account deletion, even though
DOC-06 §9 ("Idempotency ... [required for] ... account deletion") and DOC-07
("Idempotency ... required for: ... `POST /api/v1/account-deletion-requests`")
both require it. Proposed minimal reconciliation, matching how `VOC-030-D02`
resolved an equivalent gap: add `account_deletion` to the `scope` enum. Without
it, `T04`'s idempotency handling for account-deletion requests has no valid
`scope` value to write.

### Security and privacy

`user_onboarding_profiles`, the Settings fields, `email_change_links`, and
`account_deletion_requests` are all requester-owned personal state:
minimize, requester-scoped, never expose another learner's data, and return
404 (not 403) for any owner mismatch, consistent with A1–P4. Onboarding
answers and Settings fields are never learner free text beyond
`displayName` (validated for length and rejected-unknown-fields per DOC-07,
escaped on render — no AI moderation needed, since this is not sentence
content). Email-change tokens follow the same secret-handling discipline as
magic links: only a hash is ever persisted, tokens never appear in logs, and
a consumed or expired token is rejected the same way an expired/replayed
magic link is (`401`, not a distinguishing error). Account-deletion evidence,
fixtures, and tests must never use a real learner identity or real secrets;
non-production identities only, consistent with `test-plan.md`'s repository-wide
rule. No new provider, credential, or production-infrastructure surface is
introduced by this package.

## Data, migrations, analytics, and accessibility

Migrations: three new, reviewed, versioned Atlas SQL migrations
(`user_onboarding_profiles` in `T00`; `email_change_links` in `T03`;
`account_deletion_requests` in `T04`), each with FKs, uniqueness
(`user_id` unique on `user_onboarding_profiles` and `account_deletion_requests`;
`token_hash` unique on `email_change_links`), and check constraints
(`daily_review_target` 5–100 on `user_onboarding_profiles`, matching
`user_settings`'s existing constraint). No existing A1–P4 table, column, or
constraint is altered; migration tests assert this explicitly, the same way
`VOC-030-TEST-01` did for its own new tables. `idempotency_keys.scope`'s
check constraint is widened (additive) per `VOC-031-D09`. No `ON DELETE
CASCADE` is introduced anywhere by this package (DOC-05 §16); the account-deletion
sweep performs its per-table disposition explicitly in code, not via a
database cascade. Analytics: nothing new beyond aggregate/structured
counters already in place; onboarding answers and Settings values are never
duplicated into `daily_activity_summaries` or any ledger. Accessibility is
material to almost every task in this milestone (it is P5's own named scope,
not incidental): `T07` installs the first automated accessibility test
harness this repository has ever had, `T10` performs a manual/design audit
against DOC-03 §1's principles that automation cannot check (tone, one-clear-
action, non-gamified framing), and every new screen (`/onboarding`,
`/settings`, `/settings/account`) is built to the same 44px-touch-target,
keyboard-operable, non-color-only-feedback, WCAG 2.2 AA standard already
required of every other core-loop screen (DOC-03 §10, DOC-08 quality
standards).
