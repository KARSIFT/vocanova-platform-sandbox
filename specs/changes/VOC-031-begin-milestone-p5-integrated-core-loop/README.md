# VOC-031 — Begin Milestone P5: Integrated Core Loop

**Draft package — not adopted, not approved, and not implementation authority.** Human adoption,
resolution of the stated open decisions, and separate implementation authorization are required
before work begins. No authorization, approval, activation, deployment, or closure field is set by
this draft.

## Identity and lifecycle

- Package ID: `VOC-031`; canonical path:
  `specs/changes/VOC-031-begin-milestone-p5-integrated-core-loop/`.
- Lifecycle: `draft`; every authorization field in `change.yaml` remains at its unadopted default
  (`approval_status: not-approved`, `implementation_authorized: false`,
  `automatic_merge_allowed: false`, `production_impact: unknown`,
  `repository_adoption_status: not-adopted`).
- Proposed risk: **R3** (proposal only — not a determination). This milestone adds **no** new
  database migration, Ent schema, business module, or API route — direct inspection at draft time
  confirms P1–P4's own mock-decommission inventory is already empty
  (`scripts/foundation/mock-inventory.mjs`'s `expectedMocks` array has no entries) and every
  A1/P1/P2/P3/P4 screen and route already renders real, non-mock data. What raises this proposal
  to R3 is `VOC-031-T01`'s touch of `apps/web/src/middleware.ts`, which
  `docs/governance/protected-areas.md` lists explicitly under "Authentication and authorization" —
  a protected path even though the change narrows a failure-mode conflation (treating a backend
  outage the same as "not signed in") rather than altering session or authorization logic. The
  path-based classifier (`scripts/governance/classify-change-risk.sh`) floors this path at R3; the
  implementation-time classifier, builder, verifier, and applicable human authority govern the
  actual class. Four open decisions below (`D01`–`D04`) are **open** and become R4 once decided;
  this draft does not decide them.
- Decision owner: founder; target branch: `develop`; request source: free text (the DOC-12 §5 P5
  paragraph plus the supplied request, grounding DOC-00, DOC-01 §§2–3, DOC-03, DOC-04 §16, DOC-07,
  DOC-08, and DOC-11 in full).
- A-003 is active: routine R3 requires strengthened controls and exact-SHA independent
  verification but not standing steward/founder approval solely because it is R3. An R4
  product/scope or material user-trust decision remains founder-controlled. Every P5 PR requires
  Claude Code review; authorization, accessibility, and data-integrity findings block release. EHR
  is not presumed.

## Objective and requirement source

Begin DOC-12 §5 P5: combine everything built by A1 (VOC-025), P1 (VOC-026), P2 (VOC-027), P3
(VOC-028), and P4 (VOC-030) into one coherent, reliable, mobile-first learner journey —
cross-feature integration, reliability/recovery, accessibility/performance, and final UX
consistency — so the full loop works coherently across supported layouts with no critical
product/security/data/accessibility/reliability defect. Ground the design in DOC-00 (product
vision/gamification), DOC-01 §§2–3 (MVP screens/completion criteria), DOC-03 (UI/UX — one-clear-
action, mobile-first, backend-authoritative, empty/loading/error-state, and accessibility rules),
DOC-04 §16 (security/reliability baseline), DOC-07 (API contract — standard error shape), DOC-08
(web app design — quality standards, accessibility/performance targets), and DOC-11 (DevOps/CI-CD
— environments, rollback philosophy). All five prior product milestones (A1/P1/P2/P3/P4) are
already merged to `develop` and functioning as real capability; this package integrates and
hardens what they built rather than extending the backend domain.

## Scope, non-goals, risk, and protected areas

Scope is a fixed ordered task sequence: (`T00`) loading/empty/error states for the four routes
that lack them (`discover`, `discover/[situation]`, `discover/[situation]/[word]`, `reviews`) plus
one root-level fallback error boundary; (`T01`) auth-check failure-mode reliability in
`middleware.ts` (fail-closed: never grants access on ambiguity, only changes what an ambiguous
case *shows*) and mid-flow interruption safety for the review session and the shared
sentence-practice component; (`T02`) a cross-feature UX consistency audit (design tokens, pattern
parity, sentence-practice parity across its three entry points, mobile-first/touch-target
compliance); (`T03`) a WCAG 2.2 AA accessibility pass, with an open decision on automated tooling;
(`T04`) a performance pass against DOC-08's named Lighthouse targets, with an open decision on
automated tooling; (`T05`) P5 gate evidence, a no-new-backend-scope mock-inventory extension,
staging-evidence documentation, and an honest gate-readiness report.

Excluded: any new backend business module, Ent schema, migration, or API route; the onboarding
flow (DOC-03 §3) and a Settings/account screen or API (DOC-08 routes) unless `D01` resolves to
require them; the DOC-11 §3 production kill switches and other R1/R2 release-operations scope;
leaderboards/badges/social/rewards-store (DOC-12 §10); live F3 staging validation (blocked,
`VOC-031-DEP-02`); production deployment; real secrets.

Protected: `apps/web/src/middleware.ts` (authentication/authorization path floor, R3), and —
conditionally, only if `D02`/`D03` activate new CI steps — `.github/workflows/` (deployment/
rollback protected area, a distinct specialist concern from the auth-path one). Rollback must
restore the pre-P5 A1–P4 loop cleanly; this package touches no backend write path, migration, or
schema, so no learner data state is at risk from reverting it.

## Verification, approvals, release, and closure

Every P5 PR requires Claude Code review bound to the exact final SHA; authorization
(fail-closed) findings on `T01`, duplicate-reward/lost-input findings on the retry-safety changes,
accessibility findings, and no-new-backend-scope findings all block release. Run installed
commands (`pnpm validate`, `go test`/`go vet` for `apps/api` to confirm no backend regression, the
`scripts/governance/*` checks as applicable) and the deterministic UI-state/reliability/
consistency/accessibility/performance/no-new-scope tests this package adds. Staging validation and
a rollback rehearsal are required before the DOC-12 P5 gate can be evaluated; live staging
evidence is blocked until the F3 staging environment exists (`VOC-031-DEP-02`), exactly as it was
for every prior product milestone this one now compounds. This draft grants no approval, merge,
activation, credentials, deployment, or closure authority, and the package is not adopted.

## Open decisions requiring founder input

- `VOC-031-D01` — whether onboarding (DOC-03 §3) and a Settings/account screen (DOC-08) are still
  in MVP scope given DOC-01 §3's canonical completion-criteria list omits both, or whether DOC-08's
  inclusion of them is a documentation drift to correct instead (DOC-12 §11 change-control rule).
- `VOC-031-D02` — whether to install automated accessibility tooling (axe-core/Playwright) now, or
  continue the A1–P4 precedent of a documented manual pass.
- `VOC-031-D03` — whether to wire automated Lighthouse CI now, or continue a documented manual
  spot-check.
- `VOC-031-D04` — how to define "supported layouts" for the DOC-12 §5 P5 gate wording, since no
  canonical document names an exact breakpoint matrix beyond DOC-03/DOC-08's mobile-first range.

See `specification.md`'s Decisions section for the full reasoning behind each; this draft proposes
a default for `D04` only, subject to founder confirmation, and does not decide any of the four.
