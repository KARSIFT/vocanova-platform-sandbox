# VOC-052 — Seed the Canonical P1 Content on Staging (and, Conditionally, Production)

## Identity and lifecycle

- Package ID: VOC-052
- Title: Seed the Canonical P1 Content (`journey_situations`, `canonical_words`, and
  related tables) on Staging and Production Deploy
- Canonical path: `specs/changes/VOC-052-staging-has-never-had-its-canonical-p1-content`
- Lifecycle state: `draft` (not adopted, not authorized for implementation)
- Proposed risk: `R3` (draft proposal only — see `change.yaml`'s
  `planned_implementation_risk_floor`, not a determination)
- Owner: unassigned (see `change.yaml`'s `owners` block)
- Approval evidence: none yet — `approval_status: not-approved`,
  `implementation_authorized: false`
- Target branch: `develop`
- Linked GitHub issue: [#437](https://github.com/KARSIFT/vocanova-platform-sandbox/issues/437)

## Objective and requirement source

Make the real `/discover` page on staging (and, once relevant, production) render
actual journey-situation content, by running the existing, idempotent
`apps/api/cmd/seed` tool against the target database as part of the deploy pipeline,
after migrations apply — closing the gap that currently makes VOC-050's real-backend
staging core-loop E2E check fail at its "discover a situation and open a word" step
for any change, unconditionally. Full requirement grounding is in `specification.md`
and `change.yaml`'s `requirement_source`.

## Scope, non-goals, risk, and protected areas

See `specification.md` for full scope and non-goals, and `impact-analysis.md` for
risk, protected areas, and dependencies. In short: this package adds a deploy-time
seed step to `deploy-staging.yml` (T00), re-runs the real staging E2E check as
verification evidence (T01), and conditionally proposes the same for
`deploy-production.yml` (T02, gated on an explicit human decision this package
cannot make — see `VOC-052-DEP-02`). It does not modify `apps/api/cmd/seed` itself,
the migration files, or the seed content (`voc026-p1.json`).

## Verification, approvals, release, and closure

See `test-plan.md`, `release-plan.md`, and `implementation-plan.md`. This package
carries no standing approval; adoption, implementation authorization, independent
verification, and any required human approval remain to be recorded against the
exact implemented revision, per AGENTS.md and CLAUDE.md.
