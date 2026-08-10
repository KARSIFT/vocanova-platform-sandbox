# VOC-064 — Implementation Plan

## Preconditions and protected areas

Do not begin implementation. This package must not be adopted
(`change.yaml` stays at `status: draft`, `implementation_authorized: false`).
The originating free-text request instructs the human to close/ignore the
package. There is no protected-area change to reconcile because no
application, workflow, governance, secret, or migration path is in scope.

## File reconciliation and implementation sequence

1. **Planner (this run):** write the draft package files under
   `specs/changes/VOC-064-verification-dispatch-only-confirming-cursor-auto`
   only. Leave all adoption/authorization fields at unadopted defaults.
2. **Human operator:** confirm the planner workflow run completed (and,
   separately, whether Other-Models quota was avoided). Close or ignore
   any plan PR opened for this package. Do not flip adoption fields. Do
   not dispatch implementer tasks.
3. **No implementer sequence.** `implementation.required: false`. There is
   no ordered product change to land.

No existing file outside this package directory is edited, deleted, or
replaced by this package.

## Validation and independent verification

- Deterministic (optional, only if someone inspects the draft PR): confirm
  with `git status` / path listing that only this package directory
  changed; optionally run
  `bash scripts/governance/validate-governance.sh` and
  `bash scripts/governance/classify-change-risk.sh` if the calling
  workflow already does so for plan PRs — do not invent a pass for a check
  that was not run.
- Exact-SHA independent verification: not required for a close/ignore
  verification artifact. If a verifier is still invoked by the workflow,
  the expected outcome is confirmation that (a) scope stayed inside this
  package directory, (b) adoption fields remain unadopted, and (c) no
  product/process change is proposed for implementation.

## Deployment and rollback

- Authorization: none. This package does not authorize deployment.
- Rollout: none.
- Rollback trigger: not applicable for an unadopted draft; if the draft
  files are unwanted on a branch, close the plan PR or delete the branch
  without merging.
- Owner: unassigned.
- Last-known-good: repository tip prior to any accidental merge of this
  package; preferred path is never merging it as adopted work.
