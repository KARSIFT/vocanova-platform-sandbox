# VOC-146 — Release Plan

## Release and deployment authorization

This package does **not** authorize a new production deploy path, any
App-token permission change, reduced review rigor, or production
credential-value changes.

It lands the fail-closed range-loading repair so
`scripts/governance/validate-governance.sh` and
`scripts/governance/classify-change-risk.sh` return nonzero on an unresolved
`--base`/`--head` commit or invalid diff range instead of treating Git's
fatal error as an empty valid change set. Emergency PR #1126 is discovery
context only. This package must not recapture VOC-112 fixtures or create a
duplicate promotion PR or release audit.

Adoption and task implementation authorization remain separate gates under
active A-004. This draft itself is not adopted and does not authorize
implementation.

This package is not a request to snapshot the current integration tip onto
`main` as its work (`karsift-ai-infra#15`). The repair proceeds through
normal `develop` merge, staging only for the real tree change, `develop` →
`main` promotion of outstanding completed packages including this one,
exact post-promotion `develop` convergence without a staging redeploy for
tree-equivalent sync, and applicable deployment. Closed state alone is not
completion proof.

## Preconditions, monitoring, and outcome

| Phase | Owner | Preconditions | Outcome evidence |
|-------|-------|---------------|------------------|
| Plan merge | plan reviewer + merge-gate | Draft package; `automatic_merge_allowed: true`; valid `monitoring_impact` | Adopted package on `develop`; no duplicate task roster |
| T00 caller merge | implementer + independent verifier | Adoption + task authorization; new VOC-146 branch from current `develop` | `VOC-146-EV-00` — implementation PR base; range-loading change; App-authored review/check binds the live PR head |
| Post-merge promotion | existing `release.yml@main` via ordinary `reconcile-release` | T00 caller changes live on `develop`; valid completion marker | Live same-repository promotion merges; `develop` is advanced to that exact merge SHA before audit close; refs end 0 ahead / 0 behind with identical tree; allowlisted implementation/promotion run metadata is recorded |
| Staging | VOC-111 path selection | Real tree change vs tree-equivalent develop sync | Staging only for the real tree change; tree-equivalent sync must not keep staging scheduled |
| Production deploy | existing `deploy-production.yml` on every `main` push | Promotion merge produced a `main` push | Automatic production deployment runs; this package does not add a new deploy path; verify its exact-SHA result |
| Audit reconciliation | implementer evidence + maintainers | Issue #1127 reproduction SHAs preserved | Release/task/requirement records close with audit comments naming the exact promotion merge and the independently reviewed head; root issue #1127 closes only after allowlisted metadata from a successful implementation/promotion path exists. Reproduction head `79b2b3f1…` and nonexistent `--base` `376e00dd…` remain audit context. |

Monitoring inventory remains unchanged (`monitoring_impact.state: none`).

No OpenAI credential or execution route is needed or authorized.

The issue #1127 reproduction documents that governance validation claimed
success after Git reported an invalid symmetric difference. It is not this
package's implementation diff and must not be restaged as a substitute for
ordinary post-correction adopt/release evaluation.

Allowlisted live-run metadata only (no logs, secrets, or tokens): workflow
identity, event, branch, HEAD SHA, run ID, job ID(s), conclusion,
timestamps. That metadata is closure evidence for #1127, not a VOC-097
evidence-carrier task.

## Rollback

| Item | Value |
|------|-------|
| Trigger | issue #1127 class recurs; classifier accepts an invalid range as empty; valid PR range or `--files-from` regresses; VOC-086 missing-range fail-closed regresses; snapshot-gap commit; `roles.yml` changed |
| Mechanism | Revert the T00 caller script/test/doc changes |
| Owner | Implementer PR + independent verification |
| Validation | Re-run governance validation, classification, VOC-146 and VOC-086 foundation tests against the restored `develop` merge |
| Last-known-good | Caller `develop` before this package's merge (issue-creation reproduction commit `79b2b3f1…` unless superseded before T00). That last-known-good still has the #1127 fail-open range defect; rollback is to reviewed state, not to a passing invalid-range gate |

## Independent verification, human approvals, and closure

Under **active A-004**, no founder `approved` comment is required on
engineering-workflow gates. Required evidence:

1. Exact-SHA independent verification for the caller implementation PR. The
   implementer must not approve or merge its own work. The independent-review
   comment must bind the live PR head exactly and must explicitly evaluate
   nonexistent base, nonexistent head, no-merge-base, status-preserving
   diff, preserved valid range and `--files-from`, classifier parity, and
   current-state docs. Merge-gate must reject any mismatch.
2. Deterministic proof that the issue #1127 class exits nonzero and does not
   print `Governance structure validation passed.`
3. Deterministic proof that nonexistent `--head` and no-merge-base ranges
   fail closed.
4. Deterministic proof that `mapfile < <(git diff "$base...$head")` is not
   the requested-range load path and that Git nonzero status is preserved.
5. Deterministic proof that partial `--base`/`--head` does not fall through
   to working-tree discovery.
6. Deterministic proof that valid ranges, `--files-from`, and VOC-086
   missing-range fail-closed remain.
7. Deterministic proof that `classify-change-risk.sh` shares the fail-closed
   range contract.
8. Confirmation that current-state docs match the fail-closed range
   contract, fixture `roles.yml` is unchanged, and no OpenAI route was
   added.
9. Confirmation that `t00-evidence.md` names the implementation PR base,
   states that the live head is bound by the App-authored independent-review
   comment/check, does not require a commit to contain its own SHA, and
   records validation after the repair was tracked and committed.
10. Explicit record that no snapshot-gap task, pin advance, or VOC-112
    recapture was added; that emergency PR #1126 remains discovery context;
    and that #1127 closes only after valid completion markers and successful
    promotion, not from closed state alone.
11. After promotion: `develop` and `main` resolve to the same SHA for this
    package's promotion merge (0 ahead / 0 behind, identical tree); staging
    ran only for the real tree change; tree-equivalent convergence did not
    trigger an unnecessary staging deployment; automatic production
    deployment from the `main` push is verified. Post-merge audit may record
    the independently reviewed head and the promotion merge SHA.

Closure: T00 merges with passing deterministic checks and exact-SHA
independent verification, then ordinary `reconcile-release` promotes
outstanding completed packages including this correction. Do not conflate
package merge with runtime release; this package has no direct product
deployment effect. An issue-close event may wake evaluation, but closed
state alone is not completion proof.
