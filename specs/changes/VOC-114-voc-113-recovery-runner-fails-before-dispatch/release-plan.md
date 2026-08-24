# VOC-114 — Release Plan

## Release and deployment authorization

This package does **not** itself authorize a new production deploy policy.
It repairs VOC-113 recovery metadata reads so missing exact-SHA checks can be
detected and genuine workflow dispatches can proceed.

Unblocking promotion PR #947 after genuine checks may promote already-reviewed
`develop` history to `main`, which under the existing 2026-08-08 path can
automatically trigger `deploy-production`. That outcome is recovery of the
VOC-112 release handoff blocked by issue #956, not authorization of a new
deployment mechanism.

Adoption and task implementation authorization remain separate gates under
active A-004.

## Preconditions, monitoring, and outcome

| Phase | Owner | Preconditions | Outcome evidence |
|-------|-------|---------------|------------------|
| Plan merge | plan reviewer + merge-gate | Draft package; `automatic_merge_allowed: true`; valid `monitoring_impact` | Adopted package on `develop` |
| T00 merge | implementer + independent verifier | Adoption + task authorization; infra `Relates to` PRs | `VOC-114-EV-00`, green deterministic read-contract tests |
| T01 | operator + independent verifier | T00 live; any documented installation grants applied | `VOC-114-EV-01` — both recovery modes live; #947 genuine exact-head checks |
| VOC-113-T01/T02 | operator + VOC-113 roster | T01 unblocks #947 merge path | Existing VOC-113 evidence contracts |

Monitoring inventory unchanged (`monitoring_impact.state: none`).

## Rollback

| Item | Value |
|------|-------|
| Trigger | Recovery dispatches without readable metadata; over-broad App scopes; localized errors leak secrets; false merge eligibility |
| Mechanism | Revert T00 infra changes; `reconcile-release` remains available as wake path while revert lands |
| Owner | Implementer PR + independent verification |
| Validation | Deterministic absent-permission tests fail on unsafe behavior; release converge remains fail-closed without exact-head success |
| Last-known-good | Pre-T00 recovery runner and App mint configuration on consumed karsift-ai-infra `@main` |

## Independent verification, human approvals, and closure

Under **active A-004**, no founder `approved` comment is required on
engineering-workflow gates. Required evidence:

1. Exact-SHA independent verification for R4 T00 infra PR(s).
2. Deterministic read-contract fixtures (`VOC-114-TEST-01`–`05`).
3. Operator-owned live evidence for T01 (`VOC-114-TEST-06`, `VOC-114-TEST-07`)
   per `docs/operations/live-evidence.md`.
4. Preserved VOC-113 no-fabrication posture and VOC-108 authoritative selection.

Closure: all roster tasks carry valid App-authored completion markers bound to
their exact reviewed caller-PR merges, and T01 live evidence shows both recovery
modes succeed and PR #947 is unblocked. Do not conflate repository merge,
promotion, production deploy, or issue closure with package closure evidence.

Successful T01 satisfies the blocked portion of VOC-113-T01 acceptance; remaining
VOC-113-T02 post-promotion verification stays on the VOC-113 roster.
