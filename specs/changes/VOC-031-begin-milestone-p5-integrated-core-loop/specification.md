# VOC-031 — Begin Milestone P5: Integrated Core Loop: Specification

## Objective and requirement source

Deliver the DOC-12 §5 P5 gate: combine A1 (auth), P1 (discover/save), P2
(review), P3 (sentence practice + AI feedback), and P4 (daily habit/progress)
into one coherent, reliable, mobile-first journey — cross-feature
integration, reliability/recovery, accessibility/performance, and final UX
consistency — such that **the full loop works coherently in staging across
supported layouts with no critical product/security/data/accessibility/
reliability defect.** This package also delivers two pieces of genuinely new
product scope the founder has explicitly pulled into MVP for this milestone:
the DOC-03 §3 onboarding question flow and a real Settings/account surface
(DOC-08 routing table `/onboarding`, `/settings`, `/settings/account`),
built as real frontend and backend work, not deferred and not merely
recorded as a disposition. Authority: DOC-12 §5 (P5 paragraph), DOC-00 §3,
DOC-03 §§3–4, DOC-05 §6 and §§16–18, DOC-06, DOC-07, DOC-08, DOC-09, and the
supplied request (whose three founder decisions and supported-layouts
default are recorded verbatim below and are not re-opened). The A1
auth/requester context (VOC-025), P1 `content`/`learning` foundation
(VOC-026), P2 `reviews` module (VOC-027), P3 `aifeedback` module (VOC-028),
and P4 `missions`/`gamification` modules plus Home/Progress wiring (VOC-030)
carry forward; this package touches their frontend/reliability surface but
not their backend business logic, except where explicitly scoped below
(item 3, the `user_settings` resolution-chain extension).

## Scope and non-goals

In scope:

1. **`user_onboarding_profiles` persistence** (DOC-05 §6): `user_id`
   (unique, one row per user), `english_level` (`a1`/`a2`/`b1`/`b2`/
   `unknown`), `native_language`, `learning_goal`
   (`general`/`work`/`travel`/`study`/`conversation`/`exam`),
   `main_use_case` (`daily_life`/`work`/`travel`/`study`/`social`),
   `daily_review_target integer` (check 5–100), `completed_at`, added after
   `grace_day_ledger` (the last table VOC-030 created), following DOC-05
   §18's migration ordering. Owned by a new `apps/api/business/onboarding`
   module, transaction-scoped per DOC-06 §3 (no own-transaction opens),
   consistent with `missions`/`gamification`'s established pattern.
2. **Onboarding API and frontend flow** (`VOC-031-D01`): `GET
   /api/v1/onboarding` (current profile/completion state) and a submit
   endpoint that upserts the five DOC-03 §3 answers and sets
   `users.onboarding_status = 'completed'` (the `users` table already has
   this column, unused by any prior milestone) in one transaction. The
   `/onboarding` frontend route (new `(onboarding)` route group per DOC-08's
   documented grouping) presents the short question sequence — English
   level, native language, learning goal, main use case, daily review target
   — then routes straight into Home/first mission, matching DOC-03 §3's
   "straight into the first Today's Mission" framing. A learner who has not
   completed onboarding is redirected there from any `(app)` route; a
   learner who has completed it is redirected away from `/onboarding` if
   revisited. No password, no long survey — a handful of low-friction
   questions only.
3. **Extend the VOC-030-D01 timezone/target resolution chain now that its
   own missing data source exists.** VOC-030-D01 resolved
   `user_settings.timezone`/`daily_review_target` → a request-time client
   IANA timezone → UTC/default-20, explicitly noting "the onboarding answer
   step has no data source until `user_onboarding_profiles` exists." It now
   does. `VOC-031-D07` (below) proposes the minimal, single-direction
   extension: at onboarding completion, seed `user_settings.daily_review_target`
   from the onboarding answer if no `user_settings` row with a non-default
   value already exists (never overwrite an already-customized setting).
   Timezone is **not** captured by the DOC-03 §3 question list, so timezone
   resolution is unchanged from VOC-030-D01 (client-supplied IANA timezone,
   validated, else UTC).
4. **Real `user_settings` and account write endpoints**, following the
   existing A1/P1–P4 authorization/session/CSRF/idempotency conventions
   (`RequireAuth()` + `CSRFMiddleware(authSvc)` + a required
   `Idempotency-Key` header on every unsafe method, explicit DTOs, never Ent
   models, matched `@vocanova/api-client` methods, committed OpenAPI):
   - `GET /api/v1/settings` / `PATCH /api/v1/settings` (owned by
     `gamification`, which already owns `user_settings`): read/update
     `dailyReviewTarget`, `reviewIntervalPreset`, `appLanguage`,
     `notificationsEnabled`, `marketingEmailsEnabled` — the "email
     preferences" the founder decision names map to the latter two existing
     boolean columns (`VOC-031-D06`). `timezone` remains internally resolved
     (item 3), not a public-editable field in this package.
   - `GET /api/v1/account` / `PATCH /api/v1/account` (owned by `auth`, which
     already owns `users`): read/update `displayName` only.
     Email-address change and account deletion are explicitly **out of
     scope** (`VOC-031-D06`) — email is the A1 identity signal itself and a
     change there is an auth-sensitive operation with no existing precedent
     in this codebase; deletion is a DOC-05 §16 deletion-dependent,
     difficult-to-reverse action. Both are materially larger than the
     founder decision's own named example list (display name, email
     preferences, daily review target, language) and are not guessed into
     scope here.
   - Every write is requester-scoped (no ID parameter — nothing to
     enumerate), returns 401 unauthenticated, and writes only the
     authenticated requester's own row.
5. **`/settings` and `/settings/account` frontend screens** wired to item
   4's endpoints via `@vocanova/api-client`, under the A1 session, following
   DOC-08's "learning preferences, account basics" split and this repo's
   loading/empty/error conventions (item 6).
6. **Cross-feature integration pass.** The P1–P4 screens (Home, Discover,
   Review, Sentence practice, Progress) plus the two new screens (item 5)
   must present one coherent product: consistent navigation entry points
   (Home → Discover/Review/Progress/Settings and back), a single shared
   loading/empty/error UI pattern reused across every route instead of each
   route inventing its own, and no screen presenting data that contradicts
   another screen reading the same backend state (extends VOC-030's Home/
   Progress streak-consistency property to every remaining screen pair that
   shares state, e.g. Home's `dueReviewWords` count vs. the Review screen's
   own due-queue read).
7. **Reliability and recovery pass.** Across the full loop: a request that
   fails mid-flow (network error, expired session, backend 5xx) must leave
   the learner able to retry without data loss or a duplicate side effect —
   an interrupted write must regenerate or safely reuse its
   `Idempotency-Key` on retry, never silently resubmit with a stale key that
   could mask a real duplicate; an expired session mid-flow must redirect to
   sign-in and, on return, resume (not discard) the learner's place in the
   loop where reasonably possible (e.g. return to `/review` rather than
   `/home` after a session refresh triggered from the review screen); a
   slow or failed AI-feedback call must never block the rest of the loop
   (matches the existing P3 "AI can be disabled without disabling non-AI
   learning" gate language, DOC-12 §5 P3).
8. **Accessibility automation and pass** (`VOC-031-D03`, founder-decided):
   install `@axe-core/playwright` and `playwright` (DOC-08 already names
   Playwright as the intended stack; no test framework of any kind is
   currently installed for `apps/web`), add an automated a11y sweep over
   every `(app)`, `(onboarding)`, and `(public)` route at the three
   supported layouts (`VOC-031-D05`), wire it into CI, and remediate every
   violation found against the DOC-08 WCAG 2.2 AA target (labelled
   controls, visible focus, keyboard reachability, non-color-only state,
   44px minimum touch targets). No manual-only fallback.
9. **Performance automation and pass** (`VOC-031-D04`, founder-decided):
   install and wire Lighthouse CI (`@lhci/cli`, no such tooling currently
   installed) against the DOC-08 thresholds — Performance 85+, Accessibility
   95+, Best Practices 90+ — for the key routes (`/home`, `/discover`,
   `/reviews`, `/progress`, `/settings`, `/onboarding`), running in CI
   against a production build, and remediate any threshold failure. No
   manual spot-check fallback.
10. **Final UX-consistency pass** across every P1–P5 screen: shared spacing/
    typography/color tokens (no ad hoc values reintroduced), consistent
    44px-minimum touch targets, consistent focus-visible styling, and
    consistent empty/loading/error component usage (the same components
    item 6 establishes), verified at the three supported layouts.
11. Extend the deterministic mock-inventory check, collect staging evidence,
    rollback rehearsal, and P5 gate readiness.

Out of scope (do not invent): R1 staging-readiness activities beyond what
P5 itself requires; production deployment; any new gamification/mission
mechanic (P4's domain logic is frozen; only its UI/reliability surface is
touched); leaderboards, badges, social challenges, rewards store (DOC-12
§10); email-address change; account deletion; renaming or restructuring any
already-shipped route (`VOC-031-D08` records a found DOC-08-vs-implementation
routing drift without resolving it); real secrets; production credentials.

## Risk and protected areas

Proposed **R3** — not a determination. Protected paths: `/apps/api/migrations`
and `/apps/api/ent/schema` (one new table — R3 path floor),
`.github/workflows/*` (new CI job — R3 path floor), the new public write
surface over `user_settings`/`users.display_name` (first public write onto
previously internal-only or untouched personal-data columns — "sensitive-data
handling" R3), and — distinctly — every already-shipped P1–P4 frontend route
this package's reliability/cross-feature/UX-consistency work touches: a
mistake in shared loading/error/retry logic can degrade an already-working,
already-relied-upon flow across every feature at once, not just one. Under
A-003, routine R3 needs strengthened controls and exact-SHA independent
verification, not standing steward/founder approval solely for being R3;
R4 founder authority is unchanged. `VOC-031-D06` and `VOC-031-D08` below are
**open** product/scope decisions that become R4 once decided. This draft
does not decide them and does not modify any P1–P4 backend business-logic
path (only frontend/reliability wrapping and the new item-3/item-4 surface).

## Decisions, contradictions, security, and privacy

`VOC-031-D00` — **Carry-forward confirmation (confirmed at draft time,
2026-07-27).** Direct inspection of `apps/api/ent/schema/`,
`apps/api/migrations/`, `apps/api/business/`, `apps/api/app/api/`,
`apps/web/src/app/`, `package.json`, `apps/web/package.json`, and
`.github/workflows/` confirms: A1/P1/P2/P3/P4 exist and match their
respective packages' descriptions; P4's `MOCK_HOME_STATE`/
`MOCK_PROGRESS_STATE` mocks are already retired and both `home`/`progress`
pages read real `getDailyMission()`/`getProgress()` data (no P5 mock
retirement work remains from P4); `user_settings` exists with an
internal-only lazy-upsert (`gamification.EnsureUserSettings`, timezone/
target only) and no public API; `user_onboarding_profiles` does **not**
exist; no `/onboarding`, `/settings`, or `/settings/account` route exists
under `apps/web/src/app`; no accessibility or performance test tooling
(axe-core, `@axe-core/playwright`, `playwright`, Lighthouse, `@lhci/cli`) is
installed anywhere in the repository, and no `tests/e2e` directory exists
despite DOC-08 naming Vitest/RTL/Playwright as the intended web stack; the
shipped frontend routes are `/reviews` (a single route; **not** the DOC-08-
documented `/review` + `/review/session` pair) and there is no dedicated
`/words`/`/words/[userWordId]` route (saved words render inline on Home/
Progress instead) — `VOC-031-D08` records this drift. No A1/P1/P2/P3/P4
mechanic is re-litigated by this draft.

`VOC-031-D01` — **Founder-decided (per the supplied request; not re-opened
by this draft).** Onboarding (DOC-03 §3 question flow) is in MVP scope and
is built as real frontend and backend work in this package (specification
items 1–3), not deferred and not merely recorded as a disposition.

`VOC-031-D02` — **Founder-decided (per the supplied request; not re-opened
by this draft).** A Settings/account screen (DOC-08 routing:
`/settings`, `/settings/account`) is in MVP scope, including a real backend
write endpoint where a real settings/account update needs one, following
this repository's existing auth/session/security conventions (specification
items 4–5). Treated as real new scope with its own tasks, acceptance
criteria, and security/privacy review — any new write endpoint follows the
same authorization/session rules as existing routes.

`VOC-031-D03` — **Founder-decided (per the supplied request; not re-opened
by this draft).** Automated accessibility testing (axe-core and/or
Playwright, both DOC-08-named) is installed as part of the accessibility-pass
task (specification item 8); no manual-only fallback.

`VOC-031-D04` — **Founder-decided (per the supplied request; not re-opened
by this draft).** Automated Lighthouse CI against the DOC-08 documented
thresholds (Performance 85+, Accessibility 95+, Best Practices 90+) is wired
as part of the performance-pass task (specification item 9); no manual
spot-check-only fallback.

`VOC-031-D05` — **Adopted default (per the supplied request; not flagged as
an open decision).** Supported layouts for the P5 gate's "supported layouts"
wording, and for the accessibility/performance/UX-consistency passes: 360px,
430px, and one representative desktop width ≥1024px. No canonical breakpoint
matrix exists beyond DOC-03/DOC-08's mobile-first 360–430px range; this is a
reasonable default per the supplied request, not itself reopened here.

`VOC-031-D06` — **OPEN — not decided by this draft.** The exact
settings/account write-field boundary. Proposed default (specification item
4): `user_settings.dailyReviewTarget`/`reviewIntervalPreset`/`appLanguage`/
`notificationsEnabled`/`marketingEmailsEnabled` and `users.displayName` are
writable; `timezone` stays internally resolved (not publicly editable);
email-address change and account deletion are excluded from this package
entirely (DOC-08 names both under "Settings," but the founder decision's own
example list does not, and both are materially larger, separately-governed
scope — see specification item 4 and the non-goals list). This proposed
boundary needs founder confirmation at adoption, the same way VOC-030-D01
needed confirmation for the equivalent `user_settings`-scope question.

`VOC-031-D07` — **OPEN — not decided by this draft, though a
single-direction, low-risk extension of an already-approved chain.**
Whether onboarding completion should seed `user_settings.daily_review_target`
from the onboarding answer (specification item 3), extending VOC-030-D01's
resolution chain now that `user_onboarding_profiles` exists as the data
source that chain's own text said was missing. Proposed default: yes, seed
only when no `user_settings` row with an already-customized (non-default)
`daily_review_target` exists, never overwrite a learner's existing
customization. This is a genuine product-behavior choice (what happens to a
learner who customizes review target before completing onboarding, or
revisits onboarding) and is recorded as open rather than guessed, even
though the direction is low-risk.

`VOC-031-D08` — **OPEN — not decided by this draft, informational.** A
DOC-08-vs-implementation routing drift, found during `VOC-031-D00`'s
inspection and predating this package (built during P1/P2): DOC-08
documents `/review` and `/review/session` as two routes; the shipped app has
one route, `/reviews`. DOC-08 documents `/words` and
`/words/[userWordId]`; the shipped app has neither — saved words render
inline on Home and Progress instead. Per DOC-12 §11's change-control rule,
this is recorded rather than silently resolved. This package's T08 final
UX-consistency pass explicitly does **not** rename any already-shipped
route to close this drift — a rename is its own breaking, redirect-needing
change with its own blast radius, outside DOC-12 §5 P5's gate wording. The
founder should either amend DOC-08 to match the shipped routing, or
schedule a dedicated rename package; this draft does neither.

`VOC-031-D09` — composite record: `D01` (onboarding in scope, real
work), `D02` (Settings/account in scope, real work), `D03` (axe-core/
Playwright accessibility automation), `D04` (Lighthouse CI performance
automation), `D05` (360px/430px/≥1024px supported layouts) are all
founder-decided or founder-supplied-default and are **not** re-opened by
this draft. `D06` (settings/account write-field boundary), `D07`
(onboarding-seeds-`user_settings` extension), and `D08` (routing drift) are
**open** and must be resolved (into a future `D10` composite record) before
`T00`/`T02`/`T08` respectively proceed past their proposed defaults.

### Security and privacy

`user_onboarding_profiles` and the newly-public `user_settings`/`users`
write surface are requester-owned personal state: minimize, requester-scoped,
never expose another learner's onboarding/settings/account data, and return
404 (not 403) for any owner mismatch, consistent with A1/P1–P4.
`english_level`/`native_language`/`learning_goal`/`main_use_case` are
low-sensitivity self-reported learning preferences, not health/financial/
special-category data, but are still personal data under the same
minimization/retention posture as every other requester-owned table. The two
new write endpoints (`PATCH /api/v1/settings`, `PATCH /api/v1/account`) and
the onboarding submit endpoint are the first public writes onto
`user_settings`/`users` columns; they follow the exact same
`RequireAuth()` + `CSRFMiddleware(authSvc)` + required `Idempotency-Key`
pattern as every existing P1/P2 write, with explicit DTOs (never Ent models)
and no new cross-user write surface. `displayName` is free text the learner
controls entirely — no format assumed beyond reasonable length/character
validation; it is rendered back to the same learner only in this package's
scope (no public profile page exists). No secret, credential, or provider
detail is introduced by this package. The new accessibility/performance CI
tooling (axe-core/Playwright, Lighthouse CI) runs against non-production
builds/fixtures only; no production URL, credential, or real learner data is
used in any fixture, config, or CI run this package adds.

## Data, migrations, analytics, and accessibility

Migrations: reviewed versioned Atlas SQL for `user_onboarding_profiles`,
added after `grace_day_ledger` (the last table VOC-030 created), following
DOC-05 §18's dependency ordering logically even though it cannot be
physically reordered against already-applied migrations. `user_onboarding_profiles`
is deletion-dependent per DOC-05 §16 (retain only if de-identified and
unlinkable, otherwise delete — owned by future account-deletion work, not
built here, consistent with `VOC-031-D06` excluding deletion from this
package). Migration tests assert the new FK, uniqueness (one row per user),
check constraint (`daily_review_target` 5–100), and that no existing
A1/P1/P2/P3/P4 table, column, or constraint is altered. Analytics: no new
free-text analytics identifier is introduced; onboarding answers and
settings values are structured, low-cardinality enums/booleans/integers
suitable for aggregate reporting only, never joined to individual-identifying
exports by this package. Accessibility is the explicit, dedicated subject of
specification item 8 across the entire application (not just this
package's own new screens) — labelled controls, visible focus, keyboard
reachability, non-color-only state, 44px minimum touch targets, and WCAG 2.2
AA — verified by installed automation (`VOC-031-D03`), not a manual pass;
any residual gap the automation cannot cover is recorded honestly as a
limitation, never presented as a pass.
