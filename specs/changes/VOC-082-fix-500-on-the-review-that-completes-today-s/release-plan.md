# VOC-082 — Release Plan

## Release and deployment authorization

This package does **not** authorize production deployment by being
merged as a draft or even by being adopted alone. Adoption authorizes
implementation PRs only. Each task PR still requires independent
verification against the exact revision.

Proposed risk is **R2** (draft): path floor R1 with semantic elevation
for every-user daily-mission completion failure (HTTP 500 + rollback at
target). Under **active A-004**, engineering-workflow gates (plan
adoption, merge, release promotion, repository-controlled deploy)
require **no** founder `approved` comment. R2 still requires staged
evidence, monitoring, named rollback owner, and tested recovery.
`automatic_merge_allowed: true` is set per AGENTS.md (`VOC-080-DEP-02`);
setting true does not bypass path floors, CI, independent verification,
unparseable-risk fail-closed, or EHR.

After task merges, auto-promotion from `develop` to `main` and
`deploy-production.yml` on push to `main` apply per AGENTS.md.
Interrupted promotion retries via `reconcile-release`; failed gates
remain fail-closed.

## Preconditions, monitoring, and outcome

Preconditions:

- Package adopted with R2 proposal accepted or amended in writing
  (including any elevation to R3).
- Stance on stuck target-1 rows recorded (default: forward-fix only).
- T00 merged with independent-verification PASS (or PASS WITH
  NON-BLOCKING FINDINGS) on the exact SHA.
- T01 staging core-loop evidence recorded before claiming package
  closure against issue #675.

Monitoring after T00 / during T01:

- Staging core-loop pass/fail, especially the completing review.
- API 5xx rate on `POST /api/v1/reviews/submissions`.
- Daily-mission completion counts / unexpected open snapshots stuck at
  target-1 after successful UI reviews.
- Streak break / grace-day anomalies if reconciliation behavior changes
  incorrectly.

Outcome owner: named in `VOC-082-EV-01` (unassigned at drafting).
Success = `VOC-082-AC-00` through `VOC-082-AC-04` with linked evidence.

## Rollback

Trigger: completing reviews still 500 after T00; incorrect streak
resets/advances; double daily-mission completion rewards; staging
core-loop still failing for this defect after claimed fix.

Mechanism:

1. Revert `VOC-082-T00` commit(s) (primary).
2. Re-deploy via normal repository path.
3. Re-check review submissions below target still succeed; document that
   pre-fix trees reintroduce the known 500-at-target behavior.

Validation: staging core-loop and/or unit suite match the rolled-back
expectation; no unexplained ledger double-writes remain from a partial
fix.

Accountable owner: T00/T01 evidence authors. Last-known-good: tree
immediately preceding T00 merge (known-broken completing-review path
per issue #675 / run 31886780600).

## Independent verification, human approvals, and closure

Independent verifier (per `CLAUDE.md`) must:

- Bind each task report to the exact reviewed commit SHA.
- Confirm implementer did not approve/merge its own work.
- Identify active authority model **A-004** (`a004-active`).
- Confirm AC/test/evidence traceability and VOC-081 isolation.
- Report remaining R2 evidence obligations; EHR not expected for this
  package unless separately triggered.

Do not conflate repository merge, release promotion, production deploy,
and issue closure. Close issue #675 only when AC results with evidence
support it.
