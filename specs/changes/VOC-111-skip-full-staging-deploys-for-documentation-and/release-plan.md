# VOC-111 — Release Plan

## Release and deployment authorization

This package does **not** authorize production deployment or develop→main promotion.
It changes **when** push-triggered `deploy-staging` runs on `develop`.

Adoption and task implementation authorization remain separate human/automation gates
under active A-004.

## Preconditions, monitoring, and outcome

| Phase | Owner | Preconditions | Outcome evidence |
|-------|-------|---------------|------------------|
| Plan merge | plan reviewer + merge-gate | Draft package; `automatic_merge_allowed: true`; valid `monitoring_impact` | Adopted package on `develop` |
| T00 merge | implementer + independent verifier | Adoption + task authorization | `VOC-111-EV-00`, green selector/regression tests |
| T01 close | operator + independent verifier | T00 on `develop`; docs/evidence-only push | `VOC-111-EV-01` absence metadata |
| Release promotion | existing release.yml path | Full roster completion markers | Out of scope unless later packages promote runtime changes |

Monitoring inventory unchanged (`monitoring_impact.state: none`). Staging synthetics
reflect the last **selected** deploy, not skipped docs-only pushes.

## Rollback

| Item | Value |
|------|-------|
| Trigger | Runtime changes merge without staging deploy; or non-runtime pushes still run full deploy after T00 |
| Mechanism | Revert T00 (and T01 evidence PR if needed) on `develop` |
| Owner | Implementer PR + independent verification |
| Validation | Selector tests fail on reverted behavior; operator may use `workflow_dispatch` to redeploy staging while revert lands |
| Last-known-good | `deploy-staging.yml` immediately before T00 merge |

## Independent verification, human approvals, and closure

Under **active A-004**, no founder `approved` comment is required on engineering-workflow
gates. Required evidence:

1. Exact-SHA independent verification for T00 and T01 task PRs.
2. Deterministic positive/negative selector tests (`VOC-111-TEST-01`–`06`, `08`, `09`).
3. Operator-owned absence metadata for T01 (`VOC-111-TEST-07`) — not governed
   live-evidence reconcile because no success run exists.
4. Preserved deploy semantics on selected pushes and unchanged `workflow_dispatch`.

Closure: both roster tasks carry valid completion markers bound to their exact reviewed
caller-PR merges. Issue #920 may close under normal roster closure after verification.

Do not conflate repository merge, release promotion, or issue closure with package
closure evidence.
