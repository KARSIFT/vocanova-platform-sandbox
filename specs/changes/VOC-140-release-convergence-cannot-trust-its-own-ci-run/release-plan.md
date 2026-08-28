# VOC-140 — Release Plan

## Release and deployment authorization

This package does **not** authorize a new production deploy path, broader
workflow authority, reduced review rigor, or production credential changes.

It lands the recovery-identity and production-merge-guard token/API repair so
`reconcile-release` can attest recovered `ci / ci` without selecting its own
in-progress carrier and can prove the live non-bypassable production ruleset
with the App identity that merges. Promotion PR #1090 remains the live
promotion carrier after this repair merges; this package does not merge
#1090 itself and must not create a duplicate promotion PR or release audit.

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
| Coordinated infra merge | implementer + independent verifier | New `KARSIFT/karsift-ai-infra` PR; circular-CI identity repair; least-privilege mint that returns `bypass_actors`; distinct omitted-field failure | Exact infra merge SHA recorded in caller `PINNED_SHA.txt` |
| T00 caller merge | implementer + independent verifier | Adoption + task authorization; new VOC-140 branch from current `develop`; pin matches the infra merge | `VOC-140-EV-00` — implementation PR base; infra merge; identity and token/API change; App-authored review/check binds the live PR head |
| Post-merge promotion | existing `release.yml@main` via dedicated recovery if needed, then `reconcile-release` for #1089 | T00 caller changes live on `develop`; valid completion marker; recovered `ci / ci` is a completed non-carrier run; merge App identity can prove ruleset `20575146` | Promotion PR #1090 (or successor at the then-current `develop` head) merges; `develop` is advanced to that exact merge SHA before audit close; refs end 0 ahead / 0 behind with identical tree; allowlisted recovery/release run metadata is recorded |
| Staging | VOC-111 path selection | Real tree change vs tree-equivalent develop sync | Staging only for the real tree change; tree-equivalent sync must not keep staging scheduled |
| Production deploy | existing `deploy-production.yml` on every `main` push | Promotion merge produced a `main` push | Automatic production deployment runs; this package does not add a new deploy path; verify its exact-SHA result |
| Audit reconciliation | implementer evidence + maintainers | Incident run/job IDs preserved | Release/task/requirement records close with audit comments naming the exact promotion merge and the independently reviewed head; root issue #1102 closes only after allowlisted metadata from the successful recovery/release run exists. Runs `33136633666`, `33136865709`, `33136984634`, `33137091931` and jobs `98738317266`, `98739074178`, `98739420310` remain audit context. |

Monitoring inventory remains unchanged (`monitoring_impact.state: none`).

No OpenAI credential or execution route is needed or authorized.

Incident runs document that promotion was blocked before `main` changed.
They are not this package's implementation diff and must not be restaged as
a substitute for ordinary post-correction release evaluation.

Allowlisted live-run metadata only (no logs, secrets, or tokens): workflow
identity, event, branch, HEAD SHA, run ID, job ID(s), conclusion,
timestamps. That metadata is closure evidence for #1102, not a VOC-097
evidence-carrier task.

If D07's operator action is required (App installation lacks the diagnosed
permission), that settings change is the 2026-08-08 repository-settings
exception and must be recorded as allowlisted metadata only. The workflow
mint change still lands in the governed implementation PR.

## Rollback

| Item | Value |
|------|-------|
| Trigger | Circular-CI class recurs; omitted `bypass_actors` again reports `production_merge_guard_missing`; guard accepts non-empty or omitted bypass; dedicated promotion-pr-validation is no longer required to be completed; fabricated statuses; snapshot-gap commit; `roles.yml` changed |
| Mechanism | Revert the T00 caller fixture/test/doc changes and revert the coordinated infrastructure PR |
| Owner | Implementer PR + independent verification |
| Validation | Re-run caller governance/fixture suites against the restored `develop` merge `21eef755…` and pin `59943683…` |
| Last-known-good | Caller `develop` before this package's merge (issue-creation `21eef755…`, pin `59943683…`, `main` `0d0b0cdf…`). That last-known-good still has the #1089/#1090 deadlock; rollback is to reviewed state, not to a passing promotion PR |

## Independent verification, human approvals, and closure

Under **active A-004**, no founder `approved` comment is required on
engineering-workflow gates. Required evidence:

1. Exact-SHA independent verification for the infrastructure PR and the
   caller implementation PR. The implementer must not approve or merge its
   own work. The independent-review comment must bind the live PR head
   exactly and must explicitly evaluate the circular-CI identity case,
   dedicated promotion-pr-validation requirement, omitted-`bypass_actors`
   token shape, empty-bypass acceptance, non-empty-bypass rejection,
   no-fabricated-status constraint, and pin advance. Merge-gate must reject
   any mismatch.
2. Deterministic proof that an in-progress or failed release carrier is
   never trusted recovered `ci / ci`.
3. Deterministic proof that dedicated `promotion-pr-validation PR #<n>` is
   dispatched or selected and must be completed/successful before
   attestation.
4. Deterministic proof that VOC-138/VOC-139 exact-identity recovery
   semantics and doomed-job rerun refusal remain.
5. Deterministic proof that the production merge guard still requires an
   effective active repository-owned ruleset, pull-request rule, strict
   non-empty required checks, and `bypass_actors: []`.
6. Deterministic proof that omitted `bypass_actors` fails distinctly from
   `production_merge_guard_missing` and names the operator App-setting
   action.
7. Deterministic proof that tests exercise the real token-visible payload
   shape and the circular-CI parent-run fixture.
8. Deterministic proof that `PINNED_SHA.txt` equals the new infra merge.
9. Confirmation that current-state docs match the live recovery and
   App-token contract, that fixture `roles.yml` is unchanged, and that no
   OpenAI route was added.
10. Confirmation that `t00-evidence.md` names the implementation PR base and
    new infra merge, states that the live head is bound by the App-authored
    independent-review comment/check, does not require a commit to contain
    its own SHA, and records validation after the repair was tracked and
    committed.
11. Explicit record that no snapshot-gap task was added; that incident runs
    `33136633666` / `33136865709` / `33136984634` / `33137091931` and jobs
    `98738317266` / `98739074178` / `98739420310` are audit context; and that
    #1089 / #1090 / #1102 close only after valid completion markers,
    successful promotion, and allowlisted recovery/release metadata, not
    from closed state alone.
12. Confirmation that no bootstrap exception was used and no duplicate
    promotion PR or release audit was created.
13. After promotion: `develop` and `main` resolve to the same SHA for this
    package's promotion merge (0 ahead / 0 behind, identical tree); staging
    ran only for the real tree change; tree-equivalent convergence did not
    trigger an unnecessary staging deployment; automatic production
    deployment from the `main` push is verified. Post-merge audit may record
    the independently reviewed head and the promotion merge SHA.

Closure: T00 merges with passing deterministic checks and exact-SHA
independent verification, then ordinary `reconcile-release` for #1089
promotes outstanding completed packages including this correction. Do not
conflate package merge with runtime release; this package has no direct
product deployment effect. An issue-close event may wake evaluation, but
closed state alone is not completion proof.
