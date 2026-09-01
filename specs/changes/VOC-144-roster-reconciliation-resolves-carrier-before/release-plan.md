# VOC-144 — Release Plan

## Release and deployment authorization

This package does **not** authorize a new production deploy path, any
App-token permission change, reduced review rigor, or production
credential-value changes.

It lands the adoption-recovery repair so native adopt and documented
`reconcile` wait boundedly for the existing same-ref OPEN roster PR to
expose the locally pushed exact SHA, then reuse that carrier, instead of
failing at `Open roster PR` with `MISMATCHED_OPEN_CARRIER` on the first
lagged REST snapshot. Interrupted VOC-141 adoption (roster PR #1112,
reconcile runs `33437239322` and `33437514152`) is incident evidence. This
package does not merge #1112 itself and must not create a duplicate VOC-141
task, roster PR, promotion PR, or release audit.

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
| Coordinated infra merge | implementer + independent verifier | New `KARSIFT/karsift-ai-infra` PR; bounded SHA-lag wait; unchanged snapshot identity predicate; timeout-still-stale fail-closed; durable mismatch without wait; unchanged VOC-142 wait/guard | Exact infra merge SHA recorded in caller `PINNED_SHA.txt` |
| T00 caller merge | implementer + independent verifier | Adoption + task authorization; new VOC-144 branch from current `develop`; pin matches the infra merge | `VOC-144-EV-00` — implementation PR base; infra merge; named timeout/poll constants; wait change; App-authored review/check binds the live PR head |
| Post-repair VOC-141 resume | existing `pipeline.yml` `action=reconcile` | T00 caller changes live on `develop`; plan PR #1110 still merged; #1112 still the matching OPEN carrier or already merged exactly | Ordinary `reconcile` with `plan_pr_number=1110` reuses #1112 when it matches, including across post-push REST lag, waits for complete required checks, merges, cleans up, and dispatches the first not-yet-dispatched task exactly once |
| Post-merge promotion | existing `release.yml@main` via ordinary `reconcile-release` | T00 caller changes live on `develop`; valid completion marker | Live same-repository promotion merges; `develop` is advanced to that exact merge SHA before audit close; refs end 0 ahead / 0 behind with identical tree; allowlisted adopt/reconcile/release run metadata is recorded |
| Staging | VOC-111 path selection | Real tree change vs tree-equivalent develop sync | Staging only for the real tree change; tree-equivalent sync must not keep staging scheduled |
| Production deploy | existing `deploy-production.yml` on every `main` push | Promotion merge produced a `main` push | Automatic production deployment runs; this package does not add a new deploy path; verify its exact-SHA result |
| Audit reconciliation | implementer evidence + maintainers | Incident run/PR/SHA IDs preserved | Release/task/requirement records close with audit comments naming the exact promotion merge and the independently reviewed head; root issue #1122 closes only after allowlisted metadata from a successful adopt/reconcile run exists. Runs `33437239322` and `33437514152` remain audit context. |

Monitoring inventory remains unchanged (`monitoring_impact.state: none`).

No OpenAI credential or execution route is needed or authorized.

Incident runs document that documented reconcile pushed a new roster head
and then failed `MISMATCHED_OPEN_CARRIER` while GitHub still listed the
prior head, and that the same resolver returned `reuse_open` after
metadata caught up. They are not this package's implementation diff and
must not be restaged as a substitute for ordinary post-correction
adopt/release evaluation. Do not manually merge #1112; after T00,
documented reconcile of #1110 is the recovery.

Allowlisted live-run metadata only (no logs, secrets, or tokens): workflow
identity, event, branch, HEAD SHA, run ID, job ID(s), conclusion,
timestamps. That metadata is closure evidence for #1122, not a VOC-097
evidence-carrier task.

## Rollback

| Item | Value |
|------|-------|
| Trigger | post-push lag class recurs; a still-stale head is reused; a duplicate roster PR is created; in-progress parents become attestable; fabricated statuses; snapshot-gap commit; `roles.yml` changed; two-token guard changed; #1112 manually merged or duplicated; VOC-142 wait-completeness regresses |
| Mechanism | Revert the T00 caller fixture/test/doc changes and revert the coordinated infrastructure PR |
| Owner | Implementer PR + independent verification |
| Validation | Re-run caller governance/fixture suites against the restored `develop` merge and pin `8993e867…` (or the recorded pre-T00 pin) |
| Last-known-good | Caller `develop` before this package's merge (pin `8993e867…` unless superseded before T00). That last-known-good still has the #1122 SHA-lag defect; rollback is to reviewed state, not to a passing adoption recovery |

## Independent verification, human approvals, and closure

Under **active A-004**, no founder `approved` comment is required on
engineering-workflow gates. Required evidence:

1. Exact-SHA independent verification for the infrastructure PR and the
   caller implementation PR. The implementer must not approve or merge its
   own work. The independent-review comment must bind the live PR head
   exactly and must explicitly evaluate bounded post-push SHA-lag wait,
   stale-then-converge reuse, timeout-still-stale fail-closed, durable
   mismatch without wait, unchanged exact-SHA identity, no duplicate PR,
   current-state docs, and pin advance. Merge-gate must reject any mismatch.
2. Deterministic proof that stale listed head then exact pushed head reuses
   the OPEN carrier and does not call `gh pr create`.
3. Deterministic proof that timeout with a still-stale head remains
   `MISMATCHED_OPEN_CARRIER` and does not create a second PR.
4. Deterministic proof that ambiguous, wrong-repo, and wrong-base carriers,
   and API failure, fail closed without SHA-lag wait.
5. Deterministic proof that VOC-142 first-snapshot reuse, already-merged
   reuse, create-when-zero, existing-task reuse, and complete-required-set
   wait including `ci / ci` still hold.
6. Deterministic proof that wait still avoids `statusCheckRollup` /
   `gh pr checks` and that merge-gate/release still reject in-progress
   parents as attestable completed `ci / ci`.
7. Deterministic proof that tests exercise the live adapter path with
   injected snapshots or a fake clock.
8. Deterministic proof that `PINNED_SHA.txt` equals the new infra merge and
   exhaustive old-pin/hash assertions were reconciled.
9. Confirmation that current-state docs match the convergence-wait contract,
   fixture `roles.yml` is unchanged, and no OpenAI route was added.
10. Confirmation that `t00-evidence.md` names the implementation PR base,
    new infra merge, and named timeout/poll constants, states that the live
    head is bound by the App-authored independent-review comment/check, does
    not require a commit to contain its own SHA, and records validation after
    the repair was tracked and committed.
11. Explicit record that no snapshot-gap task was added; that #1112 was not
    manually merged, closed, recreated, or bypassed; that incident runs
    `33437239322` / `33437514152` and heads `958e0fed…` / `0206cb70…` are
    audit context; and that #1122 closes only after valid completion
    markers, successful promotion, and allowlisted adopt/reconcile metadata,
    not from closed state alone.
12. Confirmation that no bootstrap exception was used and no duplicate
    VOC-141 task, roster PR, promotion PR, or release audit was created.
13. After promotion: `develop` and `main` resolve to the same SHA for this
    package's promotion merge (0 ahead / 0 behind, identical tree); staging
    ran only for the real tree change; tree-equivalent convergence did not
    trigger an unnecessary staging deployment; automatic production
    deployment from the `main` push is verified. Post-merge audit may record
    the independently reviewed head and the promotion merge SHA.

Closure: T00 merges with passing deterministic checks and exact-SHA
independent verification, then ordinary `reconcile` of #1110 may resume
#1112, and ordinary `reconcile-release` promotes outstanding completed
packages including this correction. Do not conflate package merge with
runtime release; this package has no direct product deployment effect. An
issue-close event may wake evaluation, but closed state alone is not
completion proof.
