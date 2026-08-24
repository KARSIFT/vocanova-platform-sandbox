# VOC-113 — Release Plan

## Release and deployment authorization

This package does **not** itself authorize a new production deploy policy.
It restores governance-automation recovery so stranded App-driven merges and
promotion PRs obtain genuine exact-SHA checks.

Completing the live fixture (promotion PR #947) after genuine checks may promote
already-reviewed `develop` history to `main`, which under the existing
2026-08-08 path can automatically trigger `deploy-production`. That outcome is
recovery of the VOC-112 release handoff, not authorization of a new deployment
mechanism.

Adoption and task implementation authorization remain separate gates under
active A-004.

## Preconditions, monitoring, and outcome

| Phase | Owner | Preconditions | Outcome evidence |
|-------|-------|---------------|------------------|
| Plan merge | plan reviewer + merge-gate | Draft package; `automatic_merge_allowed: true`; valid `monitoring_impact` | Adopted package on `develop` |
| T00 merge | implementer + independent verifier | Adoption + task authorization; infra `Relates to` PRs | `VOC-113-EV-00`, green deterministic recovery tests |
| T01 | operator + independent verifier | T00 live on consumed `@main` / caller pipeline | `VOC-113-EV-01` — PR #947 genuine exact-head checks + merge |
| T02 | operator + independent verifier | #947 merged to `main` | `VOC-113-EV-02` — post-promotion workflow metadata |
| Remediation close | roster completion markers | T01 + T02 evidence complete | Issue #948 may close; closure alone is not proof |

Monitoring inventory unchanged (`monitoring_impact.state: none`).

## Rollback

| Item | Value |
|------|-------|
| Trigger | Fabricated checks; duplicate promotion PRs/releases; unbounded recursion; false-negative blocks on legitimate green SHAs |
| Mechanism | Revert T00 infra/caller changes; keep pre-change `reconcile-release` wake path available |
| Owner | Implementer PR + independent verification |
| Validation | Deterministic negative fixtures fail on reverted unsafe behavior; release converge remains fail-closed without exact-head success |
| Last-known-good | Pre-T00 merge-gate/release on consumed karsift-ai-infra `@main` |

## Independent verification, human approvals, and closure

Under **active A-004**, no founder `approved` comment is required on
engineering-workflow gates. Required evidence:

1. Exact-SHA independent verification for T00 (and any caller contract PR).
2. Deterministic recovery fixtures (`VOC-113-TEST-01`–`07`).
3. Operator-owned live evidence for T01 (`VOC-113-TEST-08`) and T02
   (`VOC-113-TEST-09`) per `docs/operations/live-evidence.md`.
4. Preserved App-token mutation posture, ruleset contexts, and VOC-108
   authoritative selection.

Closure: all roster tasks carry valid App-authored completion markers bound to
their exact reviewed caller-PR merges, and T02 live evidence is recorded. Do not
conflate repository merge, promotion, production deploy, or issue closure with
package closure evidence.
