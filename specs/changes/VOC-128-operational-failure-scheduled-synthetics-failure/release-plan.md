# VOC-128 — Release Plan

## Release and deployment authorization

A merged package does not itself authorize production application deployment.
T00 changes a scheduled workflow on `develop`; production synthetic YAML on
`main` updates only through the normal develop-to-main promotion after roster
completion. Under **active A-004**, engineering-workflow gates require no
founder `approved` comment.

## Preconditions, monitoring, and outcome

- Exact reviewed T00 SHA merged to `develop`.
- Deterministic voc086/voc090 tests, governance validation, and
  `git diff --check` green on that SHA.
- Independent verification PASS or PASS WITH NON-BLOCKING FINDINGS bound to
  that exact SHA.
- Monitoring: existing five canonical synthetics. After T00, the develop
  hourly schedule must conclude `success` and must create a production-only
  `main` run. Outcome owner: unassigned at draft time.

## Rollback

- **Trigger:** develop schedule still fails at the production environment gate;
  dispatcher fails; production checks never run on `main`; or staging
  core-journey runs twice per hour.
- **Mechanism:** revert the T00 workflow, validator, test, and monitoring-doc
  commits on `develop` and re-promote if already on `main`.
- **Accountable owner:** unassigned at draft time.
- **Validation:** next `scheduled-synthetics` run metadata (conclusions and job
  names only) plus voc086 tests.
- **Last-known-good:** pre-T00 `scheduled-synthetics.yml` on `develop`, which
  currently fails every hourly production job. Rollback restores that failure
  mode; it does not restore a previously green develop-default schedule,
  because none existed for production jobs.

## Independent verification, human approvals, and closure

- Independent verifier binds to each task PR's exact head SHA.
- Strengthened R3 evidence: workflow ref-gate tests, dispatcher least-privilege
  check, monitoring-doc update in the same PR, and T01 live-evidence metadata.
- Remaining hosted control: GitHub Environment `production` stays restricted
  away from `develop`.
- Closure evidence: T00 merged; T01 live-evidence qualified; issue #1037
  closed only through the roster path. Do not conflate repository merge,
  release, activation, or closure.
- EHR is not triggered.
