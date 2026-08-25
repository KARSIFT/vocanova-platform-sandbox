# VOC-117 — Release Plan

## Release and deployment authorization

This package does **not** authorize a new production deploy path, broader workflow
authority, reduced review rigor, or production credential changes.

It changes AI role bindings and model-string routing so future governed plan,
implement, review, and plan-review runs use the requested Cursor lineup.

Adoption and task implementation authorization remain separate gates under active
A-004.

## Preconditions, monitoring, and outcome

| Phase | Owner | Preconditions | Outcome evidence |
|-------|-------|---------------|------------------|
| Plan merge | plan reviewer + merge-gate | Draft package; `automatic_merge_allowed: true`; valid `monitoring_impact` | Adopted package on `develop` |
| T00 coordinated source + caller merge | implementer + independent verifier | Adoption + task authorization; reviewed shared-infra PR precedes caller fixture pin | `VOC-117-EV-00` — six bindings live; parameterized routing tested; fixture pin equals infra merge SHA |
| Post-merge effect | repository maintainers + future role runs | T00 merged to integration branch | Planner/review/plan-review use Grok 4.6 Standard; implementer and escalation use composer-2.5 |

Monitoring inventory remains unchanged (`monitoring_impact.state: none`).

## Rollback

| Item | Value |
|------|-------|
| Trigger | Parameterized routing fails closed incorrectly, silent model substitution appears, implementer/reviewer distinctness regresses, or current-state docs claim dormant routes are active |
| Mechanism | Revert the T00 infra and caller fixture/test/doc changes |
| Owner | Implementer PR + independent verification |
| Validation | Re-run infra and caller governance/fixture suites against the restored bindings |
| Last-known-good | Role bindings and workflow routing before VOC-117 implementation |

## Independent verification, human approvals, and closure

Under **active A-004**, no founder `approved` comment is required on
engineering-workflow gates. Required evidence:

1. Exact-SHA independent verification for the coordinated infra + caller task PR(s).
2. Deterministic tests proving the six exact bindings, parameterized Cursor routing,
   and fail-closed credential/prefix behavior.
3. Confirmation that OpenAI/Codex is not required or re-enabled as the active path.
4. Confirmation that exact-SHA review, risk floors, protected checks, and one-retry
   limits remain unchanged.
5. Caller fixture pin bound to the exact reviewed shared-infra merge.

Closure: T00 merges with passing deterministic checks and exact-SHA independent
verification, and recorded evidence shows roles, workflows, tests, docs/comments, and
caller pin are reconciled. Do not conflate package merge with runtime release; this
package has no direct product deployment effect.
