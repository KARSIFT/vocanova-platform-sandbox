# VOC-031 — Tasks

Ordered PR sequence: `T00 → T01 → T02 → T03 → T04 → T05`. Each PR is independently reviewable and
requires Claude Code exact-SHA review; `T01` is additionally floored at R3 by its touch of the
protected `apps/web/src/middleware.ts` auth-check path. **`D01`–`D04` are open founder decisions;
no task may proceed past the decision(s) it depends on by guessing.** `T03` and `T04` each have an
open tooling-scope decision (`D02`, `D03` respectively) that determines whether they add automated
test infrastructure or continue the documented-limitation precedent; `T05`'s gate-readiness
interpretation depends on `D01`.

## VOC-031-T00 — Loading/empty/error states and a root-level fallback boundary

- Requirement source: `VOC-031-D00`, DOC-03 §9, DOC-03 §11
- Acceptance criteria: `VOC-031-AC-00`, `VOC-031-AC-01`
- Tests: `VOC-031-TEST-00`..`VOC-031-TEST-06`
- Evidence: `VOC-031-EV-00`..`VOC-031-EV-06`
- Status: pending

Add `loading.tsx` and `error.tsx` to `apps/web/src/app/(app)/discover`,
`discover/[situation]`, `discover/[situation]/[word]`, and `apps/web/src/app/(app)/reviews`,
mirroring the existing `home`/`progress` pattern (calm `aria-busy` skeleton; an on-brand error
state with a `reset()`-driven retry control and a navigate-away fallback; no data loss on retry).
Preserve or add each route's honest first-time/no-data empty state. Add
`apps/web/src/app/global-error.tsx` as the one root-level fallback boundary for an error a
route-level boundary itself cannot catch. No backend, middleware, or business-logic change in this
PR.

## VOC-031-T01 — Auth-check reliability and mid-flow interruption safety

- Requirement source: `VOC-031-D00`, DOC-03 §1, DOC-03 §9, DOC-04 §16
- Acceptance criteria: `VOC-031-AC-02`, `VOC-031-AC-03`
- Tests: `VOC-031-TEST-07`..`VOC-031-TEST-14`
- Evidence: `VOC-031-EV-07`..`VOC-031-EV-14`
- Status: pending

In `apps/web/src/middleware.ts`, distinguish a genuine `401` from `GET /api/v1/me` (redirect to
`/signin`, unchanged) from a network exception or any other non-`401`/non-`ok` response (allow the
request through so the destination route's `T00` error boundary can render a recoverable
"can't reach Vocanova right now" state instead of a silent sign-out). The change must fail closed:
an ambiguous/errored check never grants access to data it would not otherwise grant — it only
changes what the learner is shown, never what is authorized. In
`reviews/_components/review-session.tsx` and `(app)/_components/sentence-feedback.tsx`, add or
verify: input preservation across a failed submission, no duplicate submission on retry (relying
on the existing P2/P3 idempotency guarantees, never inventing a new one), and no
completed/rewarded UI state rendered ahead of backend confirmation. Add tests proving the
pre-existing P2/P3 happy-path and idempotency behavior is unchanged. No new backend route,
migration, or business-logic change in this PR.

## VOC-031-T02 — Cross-feature UX consistency audit and fixes

- Requirement source: `VOC-031-D00`, `VOC-031-D04`, DOC-08 Quality standards, DOC-03 §7
- Acceptance criteria: `VOC-031-AC-04`
- Tests: `VOC-031-TEST-15`..`VOC-031-TEST-20`
- Evidence: `VOC-031-EV-15`..`VOC-031-EV-20`
- Status: pending

Audit every real screen for: design-token compliance (extend
`scripts/foundation/check-tailwind-token-usage.mjs` coverage if a gap is found rather than
silently leaving one uncaught); structural parity of `T00`'s loading/error/retry pattern across
all six screens it touches; sentence-practice component parity across its three entry points
(Home, Word Detail, Review Completion); and a 360–430px mobile-first / 44px-minimum-touch-target
pass, plus a check at the `VOC-031-D04`-proposed desktop width. Fix any confirmed inconsistency
found. No new component library, design-token scale, or route is introduced — this task audits and
aligns existing, already-approved UI, it does not redesign it.

## VOC-031-T03 — Accessibility pass (WCAG 2.2 AA)

- Requirement source: `VOC-031-D00`, `VOC-031-D02`, DOC-03 §10, DOC-08 Quality standards
- Acceptance criteria: `VOC-031-AC-05`
- Tests: `VOC-031-TEST-21`..`VOC-031-TEST-27`
- Evidence: `VOC-031-EV-21`..`VOC-031-EV-27`
- Status: pending — automated-tooling scope depends on `D02`

Audit and fix, across every real screen: labels on every interactive element including icon-only
controls, visible/logical keyboard focus order with no trap, non-color-only state indication
(review correctness, mission/streak status, AI-feedback result), sufficient contrast, and
screen-reader-announced errors for every `T00` state. If the adopted `D02` activates automated
tooling, add axe-core (or equivalent) scans covering every real screen and wire them into the
applicable check suite; otherwise perform and document a manual/visual pass and record the absent
automation honestly as a limitation, not a pass.

## VOC-031-T04 — Performance pass

- Requirement source: `VOC-031-D00`, `VOC-031-D03`, DOC-08 Quality standards
- Acceptance criteria: `VOC-031-AC-06`
- Tests: `VOC-031-TEST-28`..`VOC-031-TEST-32`
- Evidence: `VOC-031-EV-28`..`VOC-031-EV-32`
- Status: pending — automated-tooling scope depends on `D03`

Spot-check bundle size and route-level code splitting across the real screens against DOC-08's
Lighthouse targets (Performance 85+ / Accessibility 95+ / Best Practices 90+); remove or defer any
unnecessary dependency found. If the adopted `D03` activates automated Lighthouse CI, wire it in
against the real screens and record its results; otherwise document a manual spot-check and record
the gap honestly as a limitation, not a pass.

## VOC-031-T05 — P5 gate evidence, no-new-scope check, staging evidence, and gate readiness

- Requirement source: `VOC-031-D00`, `VOC-031-D01`, DOC-12 §5 P5
- Acceptance criteria: `VOC-031-AC-07`, `VOC-031-AC-08`
- Tests: `VOC-031-TEST-33`..`VOC-031-TEST-39`
- Evidence: `VOC-031-EV-33`..`VOC-031-EV-39`
- Status: pending

Extend `scripts/foundation/mock-inventory.mjs` with an explicit P1–P4 backend allow-list sweep
asserting no new `apps/api/business` module, `apps/api/ent/schema` file, `apps/api/migrations`
file, or `apps/api/app/api` route was introduced by `T00`–`T04`. Record the actually-adopted
`VOC-031-D01` onboarding/Settings-scope disposition (not this draft's proposal presented as
decided). Collect in-repository evidence for `T00`–`T04`, document the staged full-loop exercise
(sign in → discover → save → review → sentence-practice → mission-completes → progress-reads,
across the `D04`-proposed supported layouts) and rollback rehearsal that can only run once F3
exists, and report P5 gate readiness — explicitly including every reason the DOC-12 §5 P5 gate
cannot be declared complete by this work alone. Do not declare the DOC-12 P5 gate complete.

### Deliverables

- `mock-inventory.md`: records the P1–P4 backend allow-list sweep and confirms no P5-invented
  backend capability.
- `staging-evidence.md`: collects in-repository evidence and documents the staged full-loop
  exercise and rollback rehearsal that can only run once F3 exists.
- updated `scripts/foundation/mock-inventory.mjs` (+ its test file): deterministic check enforcing
  the P5 no-new-backend-scope boundary.

### Blocker

`VOC-031-DEP-02` remains open: F3 staging does not exist, so the live staging exercise cannot be
executed. This task provides the procedure and the in-repository evidence only; it does not
declare the DOC-12 P5 gate complete.
