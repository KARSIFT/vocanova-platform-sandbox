# VOC-090 — Release Plan

## Release and deployment authorization

This package does not authorize production deployment. Workflow changes take
effect when task PRs merge to `develop` and on the repository-controlled
promotion path to `main`. No founder `approved` comment is a merge/adopt/release
gate under active A-004.

## Preconditions, monitoring, and outcome

- **Preconditions:** Package adopted and implementation-authorized; T00 then T01
  merge in order; CI, governance validation, and independent verification pass
  on each task PR.
- **Exact revision:** recorded at task completion, not at drafting time.
- **Monitoring:** `synthetic.staging.authenticated-core-journey` must complete
  within its declared job budget on live `workflow_dispatch` (T01) and on subsequent
  hourly schedule ticks. Other canonical synthetics must remain green.
- **Outcome owner:** unassigned (set at adoption).
- **Issue #759:** closes with package roster completion after green verification;
  the operational-failure fingerprint `scheduled-synthetics:cancelled` must not
  spawn duplicate open issues during successful remediation.

## Rollback

- **Trigger:** Hourly scheduled run still cancels or fails after T00; cache
  causes incorrect dependency resolution; staging core-journey false green.
- **Mechanism:** Revert T00 workflow and registry commits on `develop`; promote
  revert through normal release path if already on `main`.
- **Owner:** unassigned (set at adoption).
- **Validation:** After rollback, confirm whether hourly timeout reproduces (expected
  if revert removes fix) or prior green state returns; do not leave issue #759
  open without a new governed plan if failure persists.
- **Last-known-good:** `develop` HEAD before T00 merge.

## Independent verification, human approvals, and closure

- Each task PR receives exact-SHA independent verification.
- Under active A-004, no founder `approved` comment is an engineering-workflow
  merge gate. R3 evidence obligations: deterministic tests pass, scope confirmed,
  live green run recorded, VOC-086 regression tests green.
- Closure: all AC results with evidence in `t00-evidence.md` and `t01-evidence.md`.
  Package closure follows roster completion and normal develop → main promotion.
- EHR: not triggered.
- Do not conflate repository merge, release, activation, or closure.
