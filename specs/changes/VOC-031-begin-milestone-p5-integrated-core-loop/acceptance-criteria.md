# VOC-031 — Acceptance Criteria

Acceptance criteria are observable, stable, security-aware, and bidirectionally traceable to
requirements (`D00`–`D04`), tasks (`T00`–`T05`), tests (`VOC-031-TEST-*`), and evidence.
`D01`–`D04` are **open founder decisions**; the criteria below are written against this draft's
proposed defaults where one is offered, and must be re-verified against whatever the founder
actually resolves at adoption.

## VOC-031-AC-00 — Loading/empty/error states on every real dynamic screen

- Requirement source: `VOC-031-D00`, DOC-03 §9
- Tasks: `VOC-031-T00`
- Tests: `VOC-031-TEST-00`..`VOC-031-TEST-04`
- Evidence: `VOC-031-EV-00`..`VOC-031-EV-04`
- Result: pending

`apps/web/src/app/(app)/discover/page.tsx`, `discover/[situation]/page.tsx`,
`discover/[situation]/[word]/page.tsx`, and `apps/web/src/app/(app)/reviews/page.tsx` each gain a
`loading.tsx` (calm, `aria-busy`/`aria-label`-carrying skeleton, matching the existing home/
progress pattern) and an `error.tsx` (on-brand message, a `reset()`-driven "Try again" control, and
a "Reload page" / navigate-away fallback, matching the existing home/progress pattern). Each
route's existing first-time/no-data empty state (e.g. "You're all caught up" on `reviews`) is
preserved or added where currently absent, per DOC-03 §9's "explain what will appear here and how
to get there" rule.

## VOC-031-AC-01 — Root-level fallback error boundary

- Requirement source: `VOC-031-D00`, DOC-03 §11
- Tasks: `VOC-031-T00`
- Tests: `VOC-031-TEST-05`, `VOC-031-TEST-06`
- Evidence: `VOC-031-EV-05`, `VOC-031-EV-06`
- Result: pending

`apps/web/src/app/global-error.tsx` exists and renders an on-brand, keyboard-accessible fallback
(not Next.js's default unstyled error page) for an error a route-level boundary itself cannot
catch (e.g. one thrown from `apps/web/src/app/layout.tsx`). No `(app)` route relies on this
boundary as its primary error handling — `T00`'s per-route boundaries remain the first line of
defense.

## VOC-031-AC-02 — Auth-check failure-mode reliability

- Requirement source: `VOC-031-D00`, DOC-03 §9, DOC-04 §16
- Tasks: `VOC-031-T01`
- Tests: `VOC-031-TEST-07`..`VOC-031-TEST-10`
- Evidence: `VOC-031-EV-07`..`VOC-031-EV-10`
- Result: pending

`apps/web/src/middleware.ts` distinguishes a genuine `401` (redirect to `/signin`, unchanged
behavior) from a backend-unreachable condition — a network exception, a `5xx`, or any other
non-`401`/non-`ok` response from `GET /api/v1/me` — which instead allows the request through to a
route that itself renders a recoverable "can't reach Vocanova right now, try again" state (via
`T00`'s error boundaries) rather than silently redirecting an authenticated learner to sign-in.
Ambiguous or errored auth checks never grant access to data; they only change what a genuinely
undecidable case *shows* the learner, never what is *authorized* — a request that legitimately
lacks a valid session (`401`) is denied exactly as before.

## VOC-031-AC-03 — Mid-flow interruption safety (review session and sentence practice)

- Requirement source: `VOC-031-D00`, DOC-03 §1, DOC-03 §9
- Tasks: `VOC-031-T01`
- Tests: `VOC-031-TEST-11`..`VOC-031-TEST-14`
- Evidence: `VOC-031-EV-11`..`VOC-031-EV-14`
- Result: pending

A network failure or session expiry while a review-session response or a sentence-practice
submission is in flight (`reviews/_components/review-session.tsx`,
`(app)/_components/sentence-feedback.tsx`) never discards the learner's already-typed sentence
text, never issues a second/duplicate submission on retry (relying on P2's/P3's existing
idempotency guarantees, `VOC-027`/`VOC-028`), and never renders a completed/rewarded state that the
backend did not actually confirm (DOC-03 §1's backend-authoritative rule). A safe, visible retry
action is always available from the resulting error state.

## VOC-031-AC-04 — Cross-feature UX consistency

- Requirement source: `VOC-031-D00`, DOC-08 Quality standards, DOC-03 §7
- Tasks: `VOC-031-T02`
- Tests: `VOC-031-TEST-15`..`VOC-031-TEST-20`
- Evidence: `VOC-031-EV-15`..`VOC-031-EV-20`
- Result: pending

Every real screen (sign-in, magic-link, home, discover, situation list, word detail, review
session, sentence practice, progress) uses only `@vocanova/design-tokens`-backed values (no
literal hex/px value the existing `check-tailwind-token-usage.mjs` check should catch and
currently misses is left uncaught by the time this task closes); the loading/error/retry visual
and interaction pattern introduced by `T00` is identical in structure across all six screens it
touches; the sentence-practice component behaves identically (pending state, input preservation,
disabled duplicate submission, disclaimer, report-feedback action) at all three of its entry
points (Home, Word Detail, Review Completion); every real screen meets the 360–430px mobile-first
target with 44px-minimum touch targets (DOC-08 Quality standards) and remains usable at the
`VOC-031-D04`-proposed desktop width.

## VOC-031-AC-05 — Accessibility (WCAG 2.2 AA)

- Requirement source: `VOC-031-D00`, `VOC-031-D02`, DOC-03 §10, DOC-08 Quality standards
- Tasks: `VOC-031-T03`
- Tests: `VOC-031-TEST-21`..`VOC-031-TEST-27`
- Evidence: `VOC-031-EV-21`..`VOC-031-EV-27`
- Result: pending — scope of automated coverage depends on `D02`

Every real screen has: a labelled control for every interactive element including icon-only
controls; a visible, logical keyboard focus order with no keyboard trap; no information conveyed
by color alone (review-rating correctness, mission/streak status, AI-feedback result all carry a
text or icon label in addition to color); sufficient contrast per WCAG 2.2 AA; and
screen-reader-announced errors for every state `T00` introduces. If `VOC-031-D02` activates
automated tooling, an axe-core (or equivalent) scan reports no critical/serious violation on any
real screen; if not, the manual pass is documented with the gap recorded as an explicit,
honest limitation rather than a pass.

## VOC-031-AC-06 — Performance

- Requirement source: `VOC-031-D00`, `VOC-031-D03`, DOC-08 Quality standards
- Tasks: `VOC-031-T04`
- Tests: `VOC-031-TEST-28`..`VOC-031-TEST-32`
- Evidence: `VOC-031-EV-28`..`VOC-031-EV-32`
- Result: pending — scope of automated coverage depends on `D03`

The real screens meet, or a documented gap is recorded against, DOC-08's Lighthouse targets
(Performance 85+ / Accessibility 95+ / Best Practices 90+). No unnecessary dependency, unbounded
client bundle growth, or missing route-level code splitting is introduced by `T00`–`T03`'s changes.
If `VOC-031-D03` activates automated Lighthouse CI, it runs against the real screens and its
results are recorded; if not, a manual spot-check is documented with the gap recorded as an
explicit, honest limitation.

## VOC-031-AC-07 — No new backend capability introduced

- Requirement source: `VOC-031-D00`, DOC-12 §5 P5
- Tasks: `VOC-031-T00`..`VOC-031-T05`
- Tests: `VOC-031-TEST-33`, `VOC-031-TEST-34`
- Evidence: `VOC-031-EV-33`, `VOC-031-EV-34`
- Result: pending

No `apps/api/business` module, `apps/api/ent/schema` file, `apps/api/migrations` file, or
`apps/api/app/api` route is added, renamed, or removed by this package. The extended
`scripts/foundation/mock-inventory.mjs` check asserts this via an explicit P1–P4 backend
allow-list sweep, so a future PR that silently expands P5 into new domain scope is caught at check
time, not discovered at review time.

## VOC-031-AC-08 — Evidence, open-decision disposition, staging evidence, and P5 gate readiness

- Requirement source: `VOC-031-D00`, `VOC-031-D01`, DOC-12 §5 P5
- Tasks: `VOC-031-T00`..`VOC-031-T05`
- Tests: `VOC-031-TEST-35`..`VOC-031-TEST-39`
- Evidence: `VOC-031-EV-35`..`VOC-031-EV-39`
- Result: pending — in-repository evidence only; live staging blocked until F3 exists; `D01`–`D04`
  must be resolved before the affected tasks' evidence is treated as final

Applicable checks, the deterministic tests this package adds, exact-SHA reviews, and the extended
mock-inventory check all pass. The `VOC-031-D01` onboarding/Settings-scope disposition actually
adopted is recorded (not this draft's proposal silently treated as decided). Staging tests for the
full sign-in → discover → save → review → sentence-practice → mission-completes →
progress-reads loop across the `VOC-031-D04`-proposed supported layouts, and a rollback
rehearsal for this package's changes, are documented and ready to run once F3 staging exists
(`VOC-031-DEP-02`). This enables — but does not itself declare — the DOC-12 P5 gate evaluation;
package merge or staging deploy alone never satisfies it.
