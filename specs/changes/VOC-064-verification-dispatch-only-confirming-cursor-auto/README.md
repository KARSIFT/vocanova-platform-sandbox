# VOC-064 — Verification Dispatch Only (Cursor/Auto Planner Path)

**Status: proposed, not adopted. Close/ignore — not a real change request.**

Nothing in this package is implementation-authorized. The free-text planner
dispatch that created it says explicitly: confirmation that the cursor/auto
planner path works (and whether it avoids Other-Models quota exhaustion), and
that if a package is drafted, close or ignore it.

## Identity and lifecycle

- Package ID: VOC-064
- Title: Verification Dispatch Only — Confirming Cursor/Auto Planner Path
  (Not a Real Change Request)
- Canonical path:
  `specs/changes/VOC-064-verification-dispatch-only-confirming-cursor-auto`
- Lifecycle state: `draft` (not adopted, not authorized for implementation)
- Proposed risk: `R0` (draft proposal only — see `change.yaml`'s
  `planned_implementation_risk_floor`, not a determination)
- Owner: unassigned (see `change.yaml`'s `owners` block)
- Approval evidence: none — and none should be recorded; disposition is
  close/ignore (`approval_status: not-approved`,
  `implementation_authorized: false`)
- Target branch: `develop` (only if this package were ever adopted, which it
  should not be)
- Linked GitHub issue: none — free-text verification dispatch, not an issue
  thread

## Objective and requirement source

Exercise the repository's planner dispatch with the model bound to
`planner` in karsift-ai-infra's `config/roles.yml` (cursor/auto for this run)
far enough that a complete draft change package is produced under VOC-064.
Successful drafting *is* the verification evidence. There is no product bug
to fix, no feature to build, and no workflow or governance change to ship.
Full grounding is in `specification.md` and `change.yaml`'s
`requirement_source`.

## Scope, non-goals, risk, and protected areas

In scope: this package directory only, as a disposable verification artifact.
Out of scope: every application, CI, governance, secret, migration, and
production path. See `specification.md`. Proposed risk is R0 with no
protected-area touch intended. Active authority model remains A-003; no EHR
trigger; no R4 consequence.

## Verification, approvals, release, and closure

See `test-plan.md`, `release-plan.md`, and `implementation-plan.md`. This
package carries no standing approval and must not be adopted. Intended human
disposition: close or ignore the plan PR / package without flipping adoption
fields and without dispatching implementer tasks. No production release or
deployment is authorized or required.
