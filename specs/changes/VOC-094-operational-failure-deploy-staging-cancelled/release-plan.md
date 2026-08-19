# VOC-094 — Release Plan

## Release and deployment authorization

This package does not authorize production deployment. Workflow and observer changes
take effect when task PRs merge to `develop` and on the repository-controlled
promotion path to `main`. No founder `approved` comment is a merge/adopt/release
gate under active A-004.

## Preconditions, monitoring, and outcome

- **Preconditions:** Package adopted and implementation-authorized; T00 then T01
  merge in order; CI, governance validation, and independent verification pass
  on each task PR.
- **Exact revision:** recorded at task completion, not at drafting time.
- **Monitoring:** No new Kuma monitors or synthetics. Outcome signal is operational:
  `deploy-staging` runs for latest `develop` commits reach `success`; benign
  superseded pending cancels no longer spawn governed issues.
- **Outcome owner:** unassigned (set at adoption).
- **Issue #781:** closes with package roster completion after verification; fingerprint
  `deploy-staging:cancelled` must not accumulate duplicate open issues during
  successful remediation.

## Rollback

- **Trigger:** Classifier suppresses a real deploy cancellation; queued deploy depth
  causes unacceptable delay before latest staging revision lands.
- **Mechanism:** Revert T00 workflow and script commits on `develop`; promote
  revert through normal release path if already on `main`.
- **Owner:** unassigned (set at adoption).
- **Validation:** After rollback, confirm whether pending-run cancellations and
  issue #781-class noise reproduce (expected if revert removes fix); do not leave
  issue #781 open without a new governed plan if failure persists.
- **Last-known-good:** `develop` HEAD before T00 merge.

## Independent verification, human approvals, and closure

- Each task PR receives exact-SHA independent verification.
- Under active A-004, no founder `approved` comment is an engineering-workflow
  merge gate. R3 evidence obligations: deterministic tests pass, scope confirmed,
  live deploy success and supersession hygiene recorded, VOC-088 regression tests green.
- Closure: all AC results with evidence in `t00-evidence.md` and `t01-evidence.md`.
  Package closure follows roster completion and normal develop → main promotion.
- EHR: not triggered.
- Do not conflate repository merge, release, activation, or closure.
