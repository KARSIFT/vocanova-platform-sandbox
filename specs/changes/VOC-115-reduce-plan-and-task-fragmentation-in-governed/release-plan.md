# VOC-115 — Release Plan

## Release and deployment authorization

This package does **not** authorize a new production deploy path, broader workflow
authority, or reduced review rigor.

It changes planning/adoption/task-structure policy so future governed work defaults
to fewer carriers for the same coherent objective while preserving existing exact-SHA
review, deterministic checks, and release gates.

Adoption and task implementation authorization remain separate gates under active
A-004.

## Preconditions, monitoring, and outcome

| Phase | Owner | Preconditions | Outcome evidence |
|-------|-------|---------------|------------------|
| Plan merge | plan reviewer + merge-gate | Draft package; `automatic_merge_allowed: true`; valid `monitoring_impact` | Adopted package on `develop` |
| T00 coordinated source + carrier merge | implementer + independent verifier | Adoption + task authorization; reviewed shared-infra PR precedes caller fixture pin | `VOC-115-EV-00` — primary infra SHA and caller docs/prompts/templates/validation/tests updated together under one task |
| Post-merge effect | repository maintainers + future planner/implementer/reviewer runs | T00 merged to integration branch | Future coherent work drafts one task by default; justified multi-task cases still sequence correctly |

Monitoring inventory remains unchanged (`monitoring_impact.state: none`).

## Rollback

| Item | Value |
|------|-------|
| Trigger | Consolidation policy allows scope creep, breaks justified multi-task packages, or creates documentation/prompt/validator contradictions |
| Mechanism | Revert the T00 governance/template/fixture/test changes |
| Owner | Implementer PR + independent verification |
| Validation | Re-run governance validation and fixture tests to confirm the reverted policy is internally consistent |
| Last-known-good | Current planning/adoption/task-fragmentation behavior before VOC-115 implementation |

## Independent verification, human approvals, and closure

Under **active A-004**, no founder `approved` comment is required on
engineering-workflow gates. Required evidence:

1. Exact-SHA independent verification for the governance/template/fixture/test PR.
2. Deterministic governance validation and updated regression tests proving one-task
   default plus justified multi-task preservation.
3. Explicit confirmation that in-scope causal remediation remains bounded by
   original objective, acceptance criteria, risk ceiling, and protected-area scope.
4. Confirmation that exact-SHA review, risk floors, protected-branch rules, and
   fail-closed behavior remain unchanged.

Closure: T00 merges with passing deterministic checks and exact-SHA independent
verification, and the recorded evidence shows policy/docs/prompts/templates/tests are
reconciled. Do not conflate package merge with runtime release; this package has no
direct product deployment effect.
