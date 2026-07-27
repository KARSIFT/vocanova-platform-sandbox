# VOC-031 — Begin Milestone P5: Integrated Core Loop: Specification

## Objective and requirement source

Deliver the DOC-12 §5 P5 gate: combine everything built by A1 (VOC-025), P1 (VOC-026), P2
(VOC-027), P3 (VOC-028), and P4 (VOC-030) into one coherent, reliable, mobile-first learner
journey — cross-feature integration, reliability/recovery, accessibility/performance, and final
UX consistency — so that the full loop (sign in → discover a word → save it → review it with
spaced repetition → practice a sentence → receive AI feedback → complete the daily mission → see
progress) **works coherently across supported layouts with no critical product/security/data/
accessibility/reliability defect.** Authority: DOC-12 §5 (P5 paragraph), DOC-00 (product vision
and gamification model), DOC-01 §§2–3 (MVP screens and completion criteria), DOC-03 (UI/UX
design — one-clear-action, mobile-first, backend-authoritative, empty/loading/error-state, and
accessibility rules), DOC-04 (technical architecture — modular-monolith and reliability
principles), DOC-07 (API contract — standard error/response shape), DOC-08 (web app design —
quality standards, accessibility/performance targets, Claude-review checklist), DOC-09 (AI
feedback UX contract, referenced here, not re-specified), DOC-11 (DevOps/CI-CD — environments,
rollback philosophy), and the supplied free-text request. This package activates directly on top
of A1's auth/session foundation (VOC-025), P1's `content`/`learning` modules (VOC-026), P2's
`reviews` module (VOC-027), P3's `aifeedback` module (VOC-028), and P4's `missions`/`gamification`
modules (VOC-030) — all five already merged to `develop` and functioning as real, non-mock
capability per direct repository inspection (`VOC-031-D00`).

## Scope and non-goals

In scope — five ordered tasks (`T00`–`T04`) plus a P5 gate-evidence task (`T05`), all confined to
the existing `apps/web` frontend, its middleware, and its CI-facing tooling. **No new backend
business module, Ent schema, migration, or API route is introduced** — per DOC-12 §5's own
wording, P5 integrates and hardens what P1–P4 already built; it does not add a new domain
capability:

1. **Loading/empty/error states (`T00`).** DOC-03 §9 requires an empty/loading/"error and retry"
   state "for every screen with dynamic content." Direct inspection (`VOC-031-D00`) confirms
   `home` and `progress` (built under VOC-019/VOC-020, wired to real data under VOC-030) each have
   a `loading.tsx`/`error.tsx` pair, but `discover`, `discover/[situation]`,
   `discover/[situation]/[word]`, and `reviews` (built under VOC-021/VOC-022/VOC-024/VOC-027) do
   not, and no `(app)`-group or root-level fallback error boundary exists either — an uncaught
   error on any of those four routes today falls through to Next.js's unstyled default error
   page, which matches neither DOC-03 §11's supportive-visual-tone rule nor DOC-08's accessibility
   standards. This task extends the already-established home/progress pattern (calm skeleton
   loading, a retry-without-data-loss error state, an honest first-time empty state) to the
   remaining four routes and adds one root-level fallback boundary
   (`apps/web/src/app/global-error.tsx`, the only segment Next.js lets catch an error a
   route-level boundary itself throws).
2. **Reliability/recovery (`T01`).** Two concrete, already-identified gaps. First,
   `apps/web/src/middleware.ts`'s auth check (added under VOC-025) treats every non-`401`,
   non-`ok` response from `GET /api/v1/me` — including a transient backend `5xx` or a network
   exception — identically to "not signed in" and redirects to `/signin`, silently discarding an
   active, valid session on a backend hiccup instead of showing a recoverable "can't reach
   Vocanova right now" state. Second, DOC-03 §9's "safe retry path that doesn't lose learner
   input" rule needs an explicit, tested guarantee across the two multi-step flows this package
   combines end to end for the first time: the review session
   (`reviews/_components/review-session.tsx`, VOC-027) and the sentence-practice component it
   embeds (`_components/sentence-feedback.tsx`, reused from VOC-028/VOC-030 at Home, Word Detail,
   and Review Completion) — a session expiring or a network failure mid-review or
   mid-sentence-submission must never lose the learner's typed sentence, never silently
   double-submit, and never report a mission/streak/point change that did not actually happen on
   the backend (DOC-03 §1's "backend is authoritative" rule).
3. **Cross-feature UX consistency (`T02`).** An audit-and-fix pass confirming the now-fully-real
   Discover → Word Detail → Save → Review → Sentence Practice → Mission/Progress loop reads as one
   product: consistent use of `@vocanova/design-tokens` (extending the existing
   `check-tailwind-token-usage.mjs` lint check's coverage if a gap is found), one visual/
   interaction pattern for loading/error/retry across all six real screens (built by `T00`),
   sentence-practice parity across its three entry points, and a 360–430px mobile-first /
   44px-minimum-touch-target audit across every real screen (DOC-08 Quality standards, DOC-03
   §7).
4. **Accessibility (`T03`).** A WCAG 2.2 AA pass (DOC-03 §10, DOC-08 Quality standards) across
   every real screen — sign-in, magic-link, home, discover, situation list, word detail, review
   session, sentence practice, progress — covering labelled controls (including icon-only
   controls), visible focus, keyboard operability, non-color-only state (review correctness,
   mission/streak status), and screen-reader-friendly forms/errors. `VOC-031-D02` (open) covers
   whether this package also installs automated accessibility tooling (axe-core / Playwright,
   named in DOC-08's own frontend stack but never yet installed — confirmed absent at
   `VOC-031-D00`) or continues the A1–P4 precedent of a manual/visual pass with absent automation
   recorded honestly as a limitation.
5. **Performance (`T04`).** A Core Web Vitals / bundle-size spot-check across the real screens
   against DOC-08's named Lighthouse targets (Performance 85+ / Accessibility 95+ / Best
   Practices 90+ — never yet automated in CI, confirmed absent at `VOC-031-D00`). `VOC-031-D03`
   (open) covers whether this package wires an automated Lighthouse CI check now or continues
   recording the gap as a limitation.
6. **P5 gate evidence (`T05`).** Extend the deterministic mock-inventory check to assert this
   package introduces no new backend route/table/module (guarding against P5 silently absorbing
   P1–P4-deferred scope), collect in-repository evidence for `T00`–`T04`, document the staged
   cross-feature exercise and rollback rehearsal that can only run once F3 exists, record the
   `VOC-031-D01` (onboarding/Settings scope) disposition, and report P5 gate readiness honestly —
   including every reason the DOC-12 §5 P5 gate cannot be declared complete by this work alone.

Out of scope (do not invent): any new backend business module, Ent schema, migration, or API
route (P1–P4 already built the domain; P5 integrates and hardens it, it does not extend it); the
onboarding flow (DOC-03 §3) and a Settings/account screen or API (DOC-08 routes) unless
`VOC-031-D01` is resolved to require them — this draft does not build them; production kill
switches (`AI_FEATURES_ENABLED`, `EMAIL_MAGIC_LINK_ENABLED`, `GOOGLE_OAUTH_ENABLED`,
`NEW_USER_SIGNUP_ENABLED` — DOC-11 §3), production infrastructure, and release-operations
readiness generally, all of which are R1/R2 scope, not P5's; leaderboards, badges, social
challenges, rewards store (DOC-12 §10); live F3 staging validation (blocked, `VOC-031-DEP-02`);
production deployment; real secrets.

## Risk and protected areas

Proposed **R3** — not a determination. Most of this package's paths
(`apps/web/src/app/(app)/**`, its `_components`, CSS/token usage) carry no protected-area floor
on their own, but `VOC-031-T01` modifies `apps/web/src/middleware.ts`, which
`docs/governance/protected-areas.md` lists explicitly under "Authentication and authorization"
(middleware) — an R3 path floor requiring security/authorization review even though the change
narrows a failure-mode conflation rather than altering session or authorization logic itself. If
`VOC-031-D02`/`D03` activate new CI steps, `.github/workflows/` is itself a listed protected area
(deployment/rollback — supply chain, permissions, environment gates), which would raise a second,
independent protected-area concern for whichever task adds them. Under A-003, routine R3 needs
strengthened controls and exact-SHA independent verification, not standing steward/founder
approval solely for being R3; R4 founder authority is unchanged. `VOC-031-D01`–`D04` below are
**open** product/scope/tooling decisions that become R4 once decided. This draft does not decide
them and does not modify any P1–P4 backend write path.

## Decisions, contradictions, security, and privacy

`VOC-031-D00` — **Carry-forward confirmation (confirmed at draft time, 2026-07-27).** Direct
inspection confirms: `apps/api/business/{auth,content,learning,reviews,aifeedback,missions,
gamification}` all exist; `apps/api/migrations` runs through `20260725130002_voc030_p4_
gamification_tables.sql` with no later migration; `apps/web/src/app/(app)/{home,discover,
discover/[situation],discover/[situation]/[word],reviews}` and `apps/web/src/app/{signin,
auth/magic}` all exist and render real (non-mock) data — `scripts/foundation/mock-inventory.mjs`'s
`expectedMocks` array is empty, confirming no P4-pending mock remains anywhere in the retained-mock
inventory. `apps/web/src/app/(app)/home` and `.../progress` each have a `loading.tsx` and
`error.tsx`; `discover`, `discover/[situation]`, `discover/[situation]/[word]`, and `reviews` have
neither, and no `(app)/error.tsx` or root `global-error.tsx` exists. `apps/web/src/middleware.ts`
redirects to `/signin` on any response from `GET /api/v1/me` that is not `2xx`, including a caught
network exception or a `5xx`, with no distinction from a genuine `401`. `apps/web/package.json`
declares no `playwright`, `@axe-core/*`, or `vitest`/`@testing-library/*` dependency and no
`apps/web` file matches `*.test.*`/`*.spec.*`; no `.github/workflows/*` file references
"lighthouse". No `onboarding` or `settings` route exists under `apps/web/src/app`. `infra/README.md`
still reads "This directory is a non-deploying structural boundary... authorizes no Cloudflare,
staging, production, release, or autonomous-development infrastructure" — the F3 staging
environment does not exist. No file under `apps/api` references `AI_FEATURES_ENABLED`,
`EMAIL_MAGIC_LINK_ENABLED`, `GOOGLE_OAUTH_ENABLED`, or `NEW_USER_SIGNUP_ENABLED` (DOC-11 §3's
named kill switches). No A1/P1/P2/P3/P4 mechanic is re-litigated by this draft; every task below
is additive UX/reliability/tooling work over the existing, already-accepted backend contract.

`VOC-031-D01` — **OPEN.** DOC-01 §3's canonical eight-item MVP completion-criteria list does not
mention onboarding or a Settings/account screen. DOC-03 §3 nonetheless describes an onboarding
question flow ("English level, native language, learning goal, main use case, daily review
target"), and DOC-08's routing table lists `/onboarding`, `/settings`, and `/settings/account`
and its own "MVP completion criteria" section adds "onboarding" and "settings/account management"
to a list it claims (line 120) merely "restates" DOC-01 §3 "not a separate decision" — but DOC-01
§3 has no such items, so DOC-08 has silently drifted from the document it claims to restate. A1
(`VOC-025` scope) and P4 (`VOC-030-D01`) each independently deferred onboarding/profile/settings
as future-package scope, consistent with DOC-12 §3's roadmap having no dedicated Settings/
onboarding milestone. P5 is the last product-feature milestone before R1/R2/L1, so this is the
last natural point to resolve whether onboarding and a Settings/account screen are still MVP
scope (and, if so, whose milestone builds them) or whether DOC-01 §3 governs as the sole canonical
MVP gate and DOC-08's two extra completion-criteria bullets are a documentation defect to correct
via DOC-12 §11's change-control rule, not a live requirement. This draft proposes, subject to
founder confirmation, treating DOC-01 §3 as canonical (per DOC-03 §13's own subordination clause,
"nothing in this document introduces a screen or flow beyond what DOC-01 §§2–3 already scope") and
therefore **not** building onboarding or Settings in P5 — but does not build on that proposal
itself; `T05` records whichever disposition is actually adopted.

`VOC-031-D02` — **OPEN.** Whether to install automated accessibility testing (axe-core and/or
Playwright, both named in DOC-08's own frontend stack, neither installed) as part of `T03`, versus
continuing the A1–P4 precedent — repeated in every prior milestone's own impact-analysis — of a
manual/visual accessibility pass with absent automation recorded honestly as a limitation. P5 is
the first milestone whose own gate wording names "accessibility" directly and it is the last
product milestone before R1's "all required tests pass" gate, which weighs toward automating now;
against that, adding a new test harness and its CI wiring is itself additive scope and (per the
protected-areas note above) may touch the protected `.github/workflows/` path. This draft does not
decide it.

`VOC-031-D03` — **OPEN.** Whether to wire an automated Lighthouse CI check into `T04` against
DOC-08's named Performance 85+ / Accessibility 95+ / Best Practices 90+ thresholds, versus
recording a manual/spot-check pass with the gap noted as a limitation, for the same reasons as
`VOC-031-D02`. This draft does not decide it.

`VOC-031-D04` — **OPEN.** No canonical document defines an exact "supported layouts" matrix for
the DOC-12 §5 P5 gate's own wording ("works coherently... across supported layouts"); DOC-03 §1
and DOC-08 Quality standards define only a mobile-first target range (360–430px) and note desktop
as "a wider layout of the same experience," with no explicit tablet breakpoint. This draft proposes
testing at 360px, 430px, and one representative desktop width (≥1024px) as "supported layouts,"
subject to founder confirmation; it does not invent an unapproved breakpoint system.

### Security and privacy

No new personal-data field, write endpoint, or authorization boundary is introduced. `T01`'s
middleware change must **fail closed**: distinguishing "backend unreachable" from "unauthenticated"
may only change what the learner *sees* on denial (a recoverable error instead of a silent
sign-out), never what is *granted* — an ambiguous or errored auth check must never be treated as
"still authenticated," and ambiguity must still deny access to the protected route, exactly as
today. Any accessibility/performance tooling added under `VOC-031-D02`/`D03` must not transmit
learner data, sentences, or session material to a third-party service; local/CI-only tools (axe-
core, Lighthouse CI, Playwright) meet this without further review. No secret, credential, or
production endpoint is introduced.

## Data, migrations, analytics, and accessibility

No migration, no Ent schema change, no new table. Analytics is unaffected — no new event or
identifier is introduced. Accessibility is the direct subject of `T03`, evaluated per DOC-03 §10 and
DOC-08's WCAG 2.2 AA target: labelled controls (including icon-only ones), visible focus order,
keyboard operability for every interactive element (including the review-rating buttons and the
sentence-practice submit/report-feedback actions), non-color-only state indication (review
correctness, mission/streak status), and screen-reader-friendly error messaging for every state
`T00` adds. Absent test automation, unless `VOC-031-D02`/`D03` activate it, is recorded honestly as
a limitation, never reported as a pass.
