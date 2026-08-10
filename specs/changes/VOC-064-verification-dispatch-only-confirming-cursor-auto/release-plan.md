# VOC-064 — Release Plan

## Release and deployment authorization

Not authorized by this package, and not required. Merging or adopting this
package is not the intended outcome. The free-text request says to
close/ignore it. A merged package would not authorize production deployment
in any case; under this package's own fields, `release.required: false` and
`deployment: not-authorized-by-this-package`.

## Preconditions, monitoring, and outcome

- Exact revision: the planner draft commit SHA produced by the calling
  workflow for this package directory (verification evidence that
  cursor/auto drafted successfully).
- Checks: optional path-scoped inspection that only this package directory
  changed; governance scripts only if the plan workflow already runs them.
- Approvals: none requested. Do not adopt.
- Staged evidence: `VOC-064-EV-00` (draft presence + unadopted fields);
  operator-side note on Other-Models quota outcome if useful for the
  verification that triggered this dispatch.
- Monitoring: none — no production behavior changes.
- Outcome owner: the human who dispatched the verification run.

## Rollback

- Trigger: accidental adoption or accidental merge of this package as if it
  were real work.
- Mechanism: close the plan PR without adopting; if already merged as draft
  content only, leave it as historical inert record (comparable in spirit to
  VOC-048's permanently-inert draft) and do not create or run implementer
  tasks; if adoption fields were flipped in error, revert those fields to
  unadopted defaults before any `adopt.yml` success path creates tasks.
- Accountable owner: the human who dispatches or reviews the plan PR.
- Validation: confirm no `VOC-064-T*` implementer issues were opened, or
  close them immediately if they were.
- Last-known-good reference: repository tip before any accidental adoption
  merge.

## Independent verification, human approvals, and closure

- Verifier result: not required for close/ignore; if invoked, confirm scope
  discipline and unadopted defaults only.
- R3/R4 approvals: none. Proposed risk is R0; no R4 consequence; EHR not
  triggered. Under active A-003 this package still must not self-adopt.
- Remaining hosted controls: none.
- Closure evidence: human closes/ignores the plan PR or otherwise leaves
  the package unadopted; no production release, activation, or closure of
  product work is applicable. Do not conflate "planner drafted successfully"
  with repository merge, release, activation, or product closure.
