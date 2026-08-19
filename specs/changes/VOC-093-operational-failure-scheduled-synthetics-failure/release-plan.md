# VOC-093 — Release Plan

## Release and deployment authorization

This package does not authorize production deployment by itself. Application fixes
merge through the normal develop → main promotion path. Harness and workflow changes
take effect on the next Actions run after merge. No founder `approved` comment is a
merge/adopt/release gate under active A-004.

## Preconditions, monitoring, and outcome

- **Preconditions:** Package adopted and implementation-authorized; T00 then T01
  merge in order; CI, governance validation, and independent verification pass
  on each task PR.
- **Exact revision:** recorded at task completion, not at drafting time.
- **Monitoring:** `synthetic.production.authenticated-route-content-sweep` must
  complete green on live `workflow_dispatch` (T01) and on subsequent hourly schedule
  ticks. Other canonical synthetics must remain green.
- **Outcome owner:** unassigned (set at adoption).
- **Issue #771:** closes with package roster completion after green verification;
  the operational-failure fingerprint `scheduled-synthetics:failure` must not spawn
  duplicate open issues during successful remediation.

## Rollback

- **Trigger:** Route sweep still fails after T00; fix weakens coverage; production
  route regression persists or worsens.
- **Mechanism:** Revert T00 commits on `develop`; promote revert through normal
  release path if already on `main`.
- **Owner:** unassigned (set at adoption).
- **Validation:** After rollback, confirm whether hourly failure reproduces; do not
  leave issue #771 open without a new governed plan if failure persists.
- **Last-known-good:** `develop` HEAD before T00 merge.

## Independent verification, human approvals, and closure

- Each task PR receives exact-SHA independent verification.
- Under active A-004, no founder `approved` comment is an engineering-workflow
  merge gate. R3 evidence obligations: deterministic tests pass, scope confirmed,
  live green route-sweep run recorded, VOC-085/VOC-086 regression tests green.
- Closure: all AC results with evidence in `t00-evidence.md` and `t01-evidence.md`.
  Package closure follows roster completion and normal develop → main promotion.
- EHR: not triggered.
- Do not conflate repository merge, release, activation, or closure.
