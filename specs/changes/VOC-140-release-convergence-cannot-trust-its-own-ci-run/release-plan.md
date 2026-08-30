# VOC-140 — Release Plan

## Release and deployment authorization

This package does **not** authorize a new production deploy path, any
Administration grant on the mutation token, reduced review rigor, or
production credential-value changes. Its separately scoped guard-token
permission and external App-installation prerequisite are bounded by D06/D07.

It lands the recovery-identity and production-merge-guard token/API repair so
`reconcile-release` can attest recovered `ci / ci` without selecting its own
in-progress carrier and can prove the live non-bypassable production ruleset
with the isolated guard token before the unchanged mutation token
merges. Promotion PR #1090 remains the live
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
| Coordinated infra merge | implementer + independent verifier | New `KARSIFT/karsift-ai-infra` PR; circular-CI repair; mutation token unchanged at exactly Contents/Issues/Pull requests write; separate current-repository Administration-write-only guard token in release and merge-gate production paths | Exact infra merge SHA recorded in caller `PINNED_SHA.txt`; tests prove permission sets, scope, order, and use isolation |
| T00 caller merge | implementer + independent verifier | Adoption + task authorization; new VOC-140 branch from current `develop`; pin matches the infra merge | `VOC-140-EV-00` — implementation PR base; infra merge; identity and token/API change; App-authored review/check binds the live PR head |
| External App activation | KARSIFT organization installation owner, after T00 merge | `karsift-ai-infra-bot` requests Administration: Read and write; installation `148001476` currently has `repository_selection: all` and no Administration permission; no secret rotation; this operator phase does not consume implementer retries | Permission approved; workflow guard mint remains explicitly scoped to `KARSIFT/vocanova-platform-sandbox`; hosted guard-only-token proof returns explicit `bypass_actors: []`; omission/nonempty remains fail closed |
| Post-merge promotion | existing `release.yml@main` via dedicated recovery if needed, then `reconcile-release` for #1089 | T00 caller changes live on `develop`; external App activation proof passed; valid completion marker; recovered `ci / ci` is a completed non-carrier run; isolated guard token proves ruleset `20575146` before the mutation-token merge | Promotion PR #1090 (or successor at the then-current `develop` head) merges; `develop` is advanced to that exact merge SHA before audit close; refs end 0 ahead / 0 behind with identical tree; allowlisted recovery/release run metadata is recorded |
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

The D07 App registration/installation approval is a known activation
prerequisite and uses the repository-settings exception. The precise action is
to set `karsift-ai-infra-bot` Repository permissions to Administration: Read
and write, have the owner approve it on KARSIFT organization installation
`148001476`, retain the workflow's explicit single-repository guard-token scope,
then rerun the failed guard / `reconcile-release`. No App ID/private-key secret
is rotated. The workflow mint change still lands in the governed implementation
PR. Both tokens share the same App/private key and the installation currently
uses `repository_selection: all`, leaving a documented organization-installation
permission-ceiling residual risk; a dedicated single-repository guard App is
optional future hardening, not part of T00.

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
   exact token permissions/scope/use isolation and guard-before-merge order in
   both workflows, external activation/hosted proof, current-state docs,
   same-App residual risk, no-fabricated-status constraint, and pin advance. Merge-gate must reject
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
6. Deterministic proof that the mutation token has exactly
   Contents/Issues/Pull requests write and is sole for merge/mutations, while
   the current-repository-scoped guard token has only Administration write,
   is used only for guard verification immediately before merge, and never
   reaches mutation/status/issues/PR/content execution in either workflow.
7. Deterministic proof that omitted `bypass_actors` fails distinctly from
   `production_merge_guard_missing` and names the operator App-setting
   action.
8. Post-T00 hosted proof after installation-owner approval that the guard-only
   token exposes explicit `bypass_actors: []`; no secret rotation and no
   implementer retry consumed.
9. Deterministic proof that tests exercise the real token-visible payload
   shape and the circular-CI parent-run fixture.
10. Deterministic proof that `PINNED_SHA.txt` equals the new infra merge and
    exhaustive old-pin/hash assertions were reconciled.
11. Confirmation that repository-settings and all other current-state docs
    match active A-004/current automatic release plus the two-token contract,
    fixture `roles.yml` is unchanged, and no OpenAI route was added.
12. Confirmation that `t00-evidence.md` names the implementation PR base and
    new infra merge, states that the live head is bound by the App-authored
    independent-review comment/check, does not require a commit to contain
    its own SHA, and records validation after the repair was tracked and
    committed.
13. Explicit record that no snapshot-gap task was added; that incident runs
    `33136633666` / `33136865709` / `33136984634` / `33137091931` and jobs
    `98738317266` / `98739074178` / `98739420310` are audit context; and that
    #1089 / #1090 / #1102 close only after valid completion markers,
    successful promotion, and allowlisted recovery/release metadata, not
    from closed state alone.
14. Confirmation that no bootstrap exception was used and no duplicate
    promotion PR or release audit was created.
15. After promotion: `develop` and `main` resolve to the same SHA for this
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
