# VOC-141 — Release Plan

## Release and deployment authorization

This package does **not** authorize a new production deploy path, any
App-token permission change, reduced review rigor, or production
credential-value changes.

It lands the residual recovery-dispatch repair so `reconcile-release` can
immediately dispatch dedicated `promotion-pr-validation` when GitHub-required
`ci / ci` is SUCCESS and composed evidence is unattestable, instead of
polling for 1,800 seconds. Workaround runs `33341923799` and `33342062118`
already proved dedicated dispatch unblocks this class; they are incident
evidence. This package does not merge a promotion PR itself and must not
create a duplicate promotion PR or release audit.

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
| Coordinated infra merge | implementer + independent verifier | New `KARSIFT/karsift-ai-infra` PR; SUCCESS-plus-unattestable-parent dedicated dispatch; timeout diagnostic token; unchanged carrier/guard fail-closed | Exact infra merge SHA recorded in caller `PINNED_SHA.txt` |
| T00 caller merge | implementer + independent verifier | Adoption + task authorization; new VOC-141 branch from current `develop`; pin matches the infra merge | `VOC-141-EV-00` — implementation PR base; infra merge; dispatch and diagnostic change; App-authored review/check binds the live PR head |
| Post-merge promotion | existing `release.yml@main` via ordinary `reconcile-release` | T00 caller changes live on `develop`; valid completion marker; unattestable SUCCESS `ci / ci` causes immediate dedicated dispatch rather than a 1,800-second hang | Live same-repository promotion merges; `develop` is advanced to that exact merge SHA before audit close; refs end 0 ahead / 0 behind with identical tree; allowlisted recovery/release run metadata is recorded |
| Staging | VOC-111 path selection | Real tree change vs tree-equivalent develop sync | Staging only for the real tree change; tree-equivalent sync must not keep staging scheduled |
| Production deploy | existing `deploy-production.yml` on every `main` push | Promotion merge produced a `main` push | Automatic production deployment runs; this package does not add a new deploy path; verify its exact-SHA result |
| Audit reconciliation | implementer evidence + maintainers | Incident run/job IDs preserved | Release/task/requirement records close with audit comments naming the exact promotion merge and the independently reviewed head; root issue #1109 closes only after allowlisted metadata from a successful recovery/release run exists. Runs `33340381776`, `33340516672`, `33341923799`, `33342062118` and job `99334840338` remain audit context. |

Monitoring inventory remains unchanged (`monitoring_impact.state: none`).

No OpenAI credential or execution route is needed or authorized.

Incident runs document that promotion recovery hung for 1,800 seconds before
the documented workaround. They are not this package's implementation diff
and must not be restaged as a substitute for ordinary post-correction release
evaluation. If #1090 already merged via that workaround, do not recreate it;
this package's own later promotion is the recurrence proof.

Allowlisted live-run metadata only (no logs, secrets, or tokens): workflow
identity, event, branch, HEAD SHA, run ID, job ID(s), conclusion,
timestamps. That metadata is closure evidence for #1109, not a VOC-097
evidence-carrier task.

## Rollback

| Item | Value |
|------|-------|
| Trigger | SUCCESS-plus-unattestable-parent class recurs; dedicated promotion-pr-validation is no longer dispatched; timeout diagnostics again omit the unattestable-CI reason; release carriers become attestable; doomed `ci / ci` jobs are rerun; fabricated statuses; snapshot-gap commit; `roles.yml` changed; two-token guard changed |
| Mechanism | Revert the T00 caller fixture/test/doc changes and revert the coordinated infrastructure PR |
| Owner | Implementer PR + independent verification |
| Validation | Re-run caller governance/fixture suites against the restored `develop` merge and pin `67bdfd13…` |
| Last-known-good | Caller `develop` before this package's merge (issue-creation promotion head `c3a53bab…`, pin `67bdfd13…`). That last-known-good still has the #1109 hang; rollback is to reviewed state, not to a passing recovery dispatch |

## Independent verification, human approvals, and closure

Under **active A-004**, no founder `approved` comment is required on
engineering-workflow gates. Required evidence:

1. Exact-SHA independent verification for the infrastructure PR and the
   caller implementation PR. The implementer must not approve or merge its
   own work. The independent-review comment must bind the live PR head
   exactly and must explicitly evaluate SUCCESS-plus-unattestable-parent
   dedicated dispatch, completed dedicated parent without redispatch,
   active/successful exact dedicated suppression, timeout unattestable-CI
   diagnostics, unchanged production/release-carrier fail-closed boundaries,
   no doomed `ci / ci` rerun, current-state docs, and pin advance.
   Merge-gate must reject any mismatch.
2. Deterministic proof that required SUCCESS `ci / ci` plus unattestable
   parent immediately dispatches exactly one dedicated
   `recover-promotion-pr-checks`.
3. Deterministic proof that a valid completed dedicated parent completes
   without redispatch.
4. Deterministic proof that active/successful exact dedicated recovery
   suppresses duplicates and that release carriers do not.
5. Deterministic proof that timeout diagnostics identify unattestable CI
   evidence rather than `missing_checks: none` alone.
6. Deterministic proof that VOC-138/VOC-139/VOC-140 exact-identity recovery,
   doomed-job rerun refusal, release-carrier rejection, and the two-token
   production merge guard remain.
7. Deterministic proof that tests exercise the live planner path with
   GitHub-required SUCCESS rows.
8. Deterministic proof that `PINNED_SHA.txt` equals the new infra merge and
   exhaustive old-pin/hash assertions were reconciled.
9. Confirmation that current-state docs match the unattestable-SUCCESS
   dispatch contract, fixture `roles.yml` is unchanged, and no OpenAI route
   was added.
10. Confirmation that `t00-evidence.md` names the implementation PR base and
    new infra merge, states that the live head is bound by the App-authored
    independent-review comment/check, does not require a commit to contain
    its own SHA, and records validation after the repair was tracked and
    committed.
11. Explicit record that no snapshot-gap task was added; that incident runs
    `33340381776` / `33340516672` / `33341923799` / `33342062118` and job
    `99334840338` are audit context; and that #1109 closes only after valid
    completion markers, successful promotion, and allowlisted
    recovery/release metadata, not from closed state alone.
12. Confirmation that no bootstrap exception was used and no duplicate
    promotion PR or release audit was created.
13. After promotion: `develop` and `main` resolve to the same SHA for this
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
