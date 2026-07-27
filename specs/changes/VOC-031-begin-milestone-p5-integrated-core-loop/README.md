# VOC-031 — Begin Milestone P5: Integrated Core Loop

**Draft package — not adopted, not approved, and not implementation authority.**
Human adoption, resolution of the stated open decisions, and separate
implementation authorization are required before work begins. No authorization,
approval, activation, deployment, or closure field is set by this draft.

## Identity and lifecycle

- Package ID: `VOC-031`; canonical path:
  `specs/changes/VOC-031-begin-milestone-p5-integrated-core-loop/`.
- Lifecycle: `draft`; every authorization field in `change.yaml` remains at its
  unadopted default (`approval_status: not-approved`,
  `implementation_authorized: false`, `automatic_merge_allowed: false`,
  `production_impact: unknown`, `repository_adoption_status: not-adopted`).
- Proposed risk: **R3** (proposal only — not a determination). This milestone
  adds a new owned table (`user_onboarding_profiles`) and its migration
  (`/apps/api/migrations`, `/apps/api/ent/schema` — R3 path floor); adds the
  first **public** read/write API surface over the existing `user_settings`
  table (previously internal-only, VOC-030) plus a `users.display_name`
  write, both touching requester-owned personal data (R3 "sensitive-data
  handling"); adds a CI workflow job for Lighthouse/accessibility automation
  (`.github/workflows/*` — R3 path floor); and **touches the UX/reliability
  surface of every already-shipped P1–P4 screen** (Home, Discover, Review,
  Sentence practice, Progress) for cross-feature integration, reliability,
  accessibility, and final UX consistency, without changing any of their
  backend business logic. The path-based classifier
  (`scripts/governance/classify-change-risk.sh`) floors these paths at R3; the
  implementation-time classifier, builder, verifier, and applicable human
  authority govern the actual class. This is the broadest milestone by design
  (DOC-12 §5 P5 explicitly combines everything already built), so its task
  count and surface area are larger than any prior single milestone package.
- Decision owner: founder; target branch: `develop`; request source: free
  text (the DOC-12 §5 P5 paragraph plus the supplied request, which resolves
  three founder decisions up front — onboarding and a Settings/account screen
  are in MVP scope with real backend work where a write is needed; automated
  accessibility testing via axe-core/Playwright; automated Lighthouse CI
  against the DOC-08 thresholds — and grounds the rest in DOC-00, DOC-03,
  DOC-04, DOC-05, DOC-06, DOC-07, DOC-08, and DOC-09).
- A-003 is active: routine R3 requires strengthened controls and exact-SHA
  independent verification but not standing steward/founder approval solely
  because it is R3. An R4 product/scope or material user-trust decision
  remains founder-controlled. Every P5 PR requires Claude Code review;
  false-progress, authorization, migration, accessibility-regression, and
  cross-feature-consistency findings block release. EHR is not presumed.

## Objective and requirement source

Begin DOC-12 §5 P5: combine everything built in A1/P1/P2/P3/P4 into one
coherent, reliable, mobile-first journey — cross-feature integration,
reliability/recovery, accessibility/performance, and final UX consistency.
Ground the design in DOC-00 §3 (habit-loop model), DOC-03 §3 (the onboarding
question flow: English level, native language, learning goal, main use case,
daily review target) and §4 (the core daily-session flow this milestone must
make coherent end to end), DOC-05 §6 (`user_onboarding_profiles`,
`user_settings`) and §§16–18 (deletion-dependence, migration order), DOC-06
(module boundaries and cross-module coordination), DOC-07 (API contract
conventions), DOC-08 (routing table including `/onboarding`, `/settings`,
`/settings/account`; mobile-first quality standards; Lighthouse thresholds),
and DOC-09 (AI feature behavior this milestone must not regress). The A1
auth/requester context (VOC-025), P1 content/learning foundation (VOC-026),
P2 `reviews` module (VOC-027), P3 `aifeedback` module (VOC-028), and P4
`missions`/`gamification` modules plus the Home/Progress wiring (VOC-030)
carry forward and are the direct dependency this package integrates and
hardens.

## Scope, non-goals, risk, and protected areas

Scope is a fixed ordered ten-PR sequence: (T00) `user_onboarding_profiles`
persistence and its integration into the existing `user_settings`
resolution chain; (T01) the onboarding API and `/onboarding` frontend flow;
(T02) real `user_settings`/account write endpoints, following A1/P1–P4
authorization/session/CSRF/idempotency conventions; (T03) `/settings` and
`/settings/account` frontend screens; (T04) cross-feature integration —
coherent navigation and shared state/loading/empty/error patterns across the
full P1→P4 loop; (T05) reliability and recovery — network-failure,
session-expiry, and interrupted-request recovery across that same loop;
(T06) accessibility automation (axe-core/Playwright, per the resolved
founder decision) and remediation; (T07) performance automation (Lighthouse
CI against the DOC-08 thresholds, per the resolved founder decision) and
remediation; (T08) a final UX-consistency pass across every P1–P5 screen;
(T09) evaluation, mock-inventory extension, staging evidence, and P5 gate
readiness.

Excluded: R1 staging-readiness hardening beyond what P5 itself requires;
production deployment; leaderboards, badges, social challenges, rewards
store (DOC-12 §10); any new gamification/mission mechanic (P4's scope is
frozen here, only its UI/reliability surface is touched); email-address
change and account deletion (DOC-08 mentions both under "Settings," but the
founder decision's own example list — display name, email preferences,
daily review target, language — does not name either, and both are
materially larger, separately-governed scope: email is the A1 identity
signal itself, and deletion is a difficult-to-reverse, DOC-05 §16
deletion-dependent-data action; `VOC-031-D06` records this as open, not
guessed); renaming or restructuring any already-shipped route (`VOC-031-D08`
records an existing DOC-08-vs-implementation routing drift found at draft
time, without resolving it); real secrets; production credentials.

Protected: database migrations, Ent schemas, `.github/workflows/*` (new
Lighthouse/accessibility CI job), the already-shipped A1/P1/P2/P3/P4
transactional write paths and requester-scoped authorization this package's
reliability/UX work must not weaken, the committed OpenAPI/client contract,
and — newly for this milestone — the first public write surface over
learner-owned `user_settings`/`users.display_name` data (CSRF, idempotency,
requester scoping, no cross-user write). Rollback must preserve learner
onboarding/settings data, existing P1–P4 behavior, and must never leave a
CI gate silently disabled instead of failing closed.

## Verification, approvals, release, and closure

Every P5 PR requires Claude Code review bound to the exact final SHA;
false-progress, authorization, migration, accessibility-regression,
performance-regression, and cross-feature-consistency findings block
release. Run installed commands (`pnpm validate`, `pnpm test`, `pnpm build`,
the `scripts/governance/*` checks as applicable, the Go format/vet/test/build
and web lint/typecheck/build suites discovered at the adopted base, plus the
new Playwright/axe-core and Lighthouse CI checks this package installs) and
the deterministic tests this package adds. Staging validation across the
three supported layouts (360px, 430px, one desktop width ≥1024px — this
draft's reasonable default, DOC-03/DOC-08 name no canonical breakpoint
matrix beyond "mobile-first") is required before the DOC-12 P5 gate can be
evaluated; live staging evidence is blocked until the F3 staging environment
exists (`VOC-031-DEP-02`), the same way VOC-025-DEP-01 first recorded this
gap and every milestone since has carried it forward. This draft grants no
approval, merge, activation, credentials, deployment, or closure authority,
and the package is not adopted.
