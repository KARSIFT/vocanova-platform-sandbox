# VOC-029 — Add a One-Line Comment to the Top of README.md

**Draft package — not adopted, not approved, and not implementation authority.**
This is a trivial diagnostic drafting test, explicitly requested as "not a real
change." Human adoption and separate implementation authorization are required
before any implementation begins. No authorization, approval, activation,
deployment, or closure field is set by this draft.

## Identity and lifecycle

- Package ID: `VOC-029`; canonical path:
  `specs/changes/VOC-029-add-a-one-line-comment-to-the-top-of-readme-md/`.
- Lifecycle: `draft`; every authorization field in `change.yaml` remains at its
  unadopted default (`approval_status: not-approved`,
  `implementation_authorized: false`, `automatic_merge_allowed: false`,
  `production_impact: unknown`).
- Proposed risk: **R0** (proposal only — not a determination). The only touched
  path is `README.md` at the repository root, a documentation-only file. The
  path-based classifier (`scripts/governance/classify-change-risk.sh`) floors
  `*.md` paths at `R0`; the implementation-time classifier and applicable human
  judgment govern the actual class. Nothing in this change touches a path the
  classifier lists as `R1`–`R4` (no `docs/governance/*`, no `AGENTS.md`/
  `CLAUDE.md`, no `specs/README.md`/`specs/templates/*`, no migrations, no
  auth/payments/billing/infra paths).
- Decision owner: whichever human reviews/adopts this package; target branch:
  `develop`; request source: free text (an explicit diagnostic drafting test,
  stated as "not a real change" and "do not adopt this package").
- A-003 is active: routine R3 no longer requires standing steward/founder
  approval merely for being R3; this proposed R0 change requires strengthened
  applicable controls and independent verification proportionate to its risk,
  which for a one-line documentation comment is minimal. No R4 founder decision
  or EHR trigger applies to this change.

## Objective and requirement source

Requirement source is the supplied free-text request: add a one-line comment to
the top of `README.md` noting the date it was added, purely to exercise the
planner→package drafting pipeline end to end. The request itself states this is
"a trivial diagnostic drafting test — not a real change" and directs that this
package must not be adopted.

## Scope, non-goals, risk, and protected areas

Scope: a single one-line addition at the very top of `README.md`, above the
existing `# Vocanova` heading, recording the date the line was added. Nothing
else in `README.md` changes; no other file changes.

Excluded: any change to `README.md`'s existing content, headings, or links; any
change to any other file, package, workflow, or governance document; adoption,
approval, or implementation of this package; any production or user-facing
effect (this repository's README has no runtime/build/deploy role).

Protected areas: none are touched. `README.md` is not listed among this
repository's protected paths (`docs/governance/protected-areas.md` scope);
this package touches no migration, schema, auth, payments, billing, secrets,
workflow, or governance-document path.

## Verification, approvals, release, and closure

Given the change is a single documentation line with no code, schema, or
behavior effect, the only applicable installed check is
`git diff --check` (whitespace/conflict-marker hygiene) and, if run,
`scripts/governance/classify-change-risk.sh` to confirm the `R0` path floor.
`pnpm validate`/`pnpm test`/`pnpm build` and the Go/web suites are inspected for
applicability but have no relevant target here (no application code changes).
This draft grants no approval, merge, activation, deployment, or closure
authority, and — per the explicit request — this package must not be adopted.
