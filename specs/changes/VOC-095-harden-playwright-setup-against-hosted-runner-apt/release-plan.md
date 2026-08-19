# VOC-095 — Release Plan

## Release and deployment authorization

This package does not authorize production deployment. Workflow and install-script changes
take effect when task PRs merge to `develop` and on the repository-controlled promotion
path to `main`. No founder `approved` comment is a merge/adopt/release gate under active
A-004.

## Preconditions, monitoring, and outcome

- **Preconditions:** Package adopted and implementation-authorized; T00 → T01 → T02
  merge in order; CI, governance validation, and independent verification pass on each
  task PR.
- **Exact revision:** recorded at task completion, not at drafting time.
- **Monitoring:** No new Kuma monitors or synthetics (`monitoring_impact.state: none`).
  Outcome signal is operational: `deploy-staging` runs reach `success` with core-loop
  completing; accessibility/lighthouse PR workflows no longer hang unbounded on apt during
  install.
- **Outcome owner:** unassigned (set at adoption).
- **Issue #792:** closes with package roster completion after T02 verification.

## Rollback

- **Trigger:** Bounded install causes systematic false failures; cache serves incompatible
  browser binary; deploy-staging blocked despite healthy staging host.
- **Mechanism:** Revert T00/T01 commits on `develop`; promote revert through normal release
  path if already on `main`.
- **Owner:** unassigned (set at adoption).
- **Validation:** After rollback, workflows return to inline `--with-deps` behavior;
  confirm whether mirror-stall timeouts recur (expected if revert removes fix); do not
  leave issue #792 open without a new governed plan if failure persists.
- **Last-known-good:** `develop` HEAD before T00 merge.

## Independent verification, human approvals, and closure

- Each task PR receives exact-SHA independent verification.
- Under active A-004, no founder `approved` comment is an engineering-workflow merge gate.
  R3 evidence obligations: deterministic tests pass, scope confirmed, live deploy-staging
  core-loop success recorded in T02, VOC-086 regression tests green.
- Closure: all AC results with evidence in `t00-evidence.md`, `t01-evidence.md`, and
  `t02-evidence.md`. Package closure follows roster completion and normal develop → main
  promotion.
- EHR: not triggered.
- Do not conflate repository merge, release, activation, or closure.
