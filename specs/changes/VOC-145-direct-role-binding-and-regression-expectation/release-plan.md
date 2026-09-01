# VOC-145 — Release Plan

## Release and deployment authorization

This package does **not** authorize a new production deploy path, any
App-token permission change, reduced review rigor, or production
credential-value changes.

It lands the governed role-binding reconciliation so live
`config/roles.yml`, VOC-117 current-state tests, caller fixtures/pin,
README, and CHANGELOG describe the same authorized binding set. Unauthorized
infra head `d8720829…` and self-CI run `33443684483` are incident evidence.
This package does not pin that head and must not create a duplicate
promotion PR or release audit.

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
| Plan merge | plan reviewer + merge-gate | Draft package; `automatic_merge_allowed: true`; valid `monitoring_impact` | Adopted package on `develop`; Path A default unless adoption records Path B; no duplicate task roster |
| Coordinated infra merge | implementer + independent verifier | New `KARSIFT/karsift-ai-infra` PR; authorized current bindings; historical VOC-117 assertions preserved; README/CHANGELOG reconciled; unchanged retry/exact-SHA/fail-closed controls | Exact infra merge SHA recorded in caller `PINNED_SHA.txt`; SHA is not `d8720829…` |
| T00 caller merge | implementer + independent verifier | Adoption + task authorization; new VOC-145 branch from current `develop`; pin matches the infra merge | `VOC-145-EV-00` — implementation PR base; infra merge; authorized path; binding table; App-authored review/check binds the live PR head |
| Post-merge promotion | existing `release.yml@main` via ordinary `reconcile-release` | T00 caller changes live on `develop`; valid completion marker | Live same-repository promotion merges; `develop` is advanced to that exact merge SHA before audit close; refs end 0 ahead / 0 behind with identical tree; allowlisted implement/release run metadata is recorded |
| Staging | VOC-111 path selection | Real tree change vs tree-equivalent develop sync | Staging only for the real tree change; tree-equivalent sync must not keep staging scheduled |
| Production deploy | existing `deploy-production.yml` on every `main` push | Promotion merge produced a `main` push | Automatic production deployment runs; this package does not add a new deploy path; verify its exact-SHA result |
| Audit reconciliation | implementer evidence + maintainers | Incident SHA/run IDs preserved | Release/task/requirement records close with audit comments naming the exact promotion merge and the independently reviewed head; root issue #1124 closes only after allowlisted metadata from a successful implement/release path exists. Unauthorized head `d8720829…`, governed base `8993e867…`, and run `33443684483` remain audit context. Issue #1120 is not resumed. |

Monitoring inventory remains unchanged (`monitoring_impact.state: none`).

No OpenAI credential or execution route is needed or authorized.

Incident SHAs and run `33443684483` document that infra `main` drifted and
that a rewritten green test was treated as sufficient. They are not this
package's implementation diff and must not be restaged as a substitute for
ordinary post-correction implement/release evaluation.

Allowlisted live-run metadata only (no logs, secrets, or tokens): workflow
identity, event, branch, HEAD SHA, run ID, job ID(s), conclusion,
timestamps. That metadata is closure evidence for #1124, not a VOC-097
evidence-carrier task.

## Rollback

| Item | Value |
|------|-------|
| Trigger | live bindings again disagree with the authorized set; VOC-117 tests again rewritten to bless an ungoverned lineup; pin equals `d8720829…`; retry/exact-SHA/fail-closed controls weakened; OpenAI route added; snapshot-gap commit; #1120 carrier edits |
| Mechanism | Revert the T00 caller fixture/test/doc changes and revert the coordinated infrastructure PR |
| Owner | Implementer PR + independent verification |
| Validation | Re-run caller governance/fixture suites against the restored `develop` merge and pin `8993e867…` (or the recorded pre-T00 pin) |
| Last-known-good | Caller `develop` before this package's merge (issue-creation pin `8993e867…` unless superseded before T00). That last-known-good still has the #1124 unreconciliation against drifted infra `main`; rollback is to reviewed caller state, not to a claim that infra `main` was already governed |

## Independent verification, human approvals, and closure

Under **active A-004**, no founder `approved` comment is required on
engineering-workflow gates. Required evidence:

1. Exact-SHA independent verification for the infrastructure PR and the
   caller implementation PR. The implementer must not approve or merge its
   own work. The independent-review comment must bind the live PR head
   exactly and must explicitly evaluate the authorized path, six exact
   current bindings, historical-versus-current VOC-117 split,
   README/CHANGELOG/fixture reconciliation, pin advance not equal to
   `d8720829…`, unchanged retry/exact-SHA/fail-closed controls, and
   exclusion of #1120. Merge-gate must reject any mismatch.
2. Deterministic proof that live `roles.yml` matches the authorized six
   bindings.
3. Deterministic proof that VOC-117 historical assertions were not
   rewritten to bless a later lineup.
4. Deterministic proof that effort-omitted Grok 4.6 and missing
   `CURSOR_API_KEY` still fail closed.
5. Deterministic proof that README/CHANGELOG/fixture README describe the
   authorized current set.
6. Deterministic proof that `PINNED_SHA.txt` equals the new infra merge and
   is not `d8720829…`.
7. Deterministic proof that retry caps, exact-SHA review wiring, and
   no-OpenAI constraints remain.
8. Deterministic proof that exhaustive current-state matches were
   reconciled and historical package directories were not rewritten.
9. Confirmation that `t00-evidence.md` names the implementation PR base,
   new infra merge, and authorized path, states that the live head is bound
   by the App-authored independent-review comment/check, does not require a
   commit to contain its own SHA, and records validation after the repair
   was tracked and committed.
10. Explicit record that no snapshot-gap task was added; that issue #1120
    was not resumed; that `d8720829…`, `8993e867…`, and run `33443684483`
    are audit context; and that #1124 closes only after valid completion
    markers, successful promotion, and allowlisted implement/release
    metadata, not from closed state alone.
11. Confirmation that no bootstrap exception was used and no duplicate
    promotion PR or release audit was created.
12. After promotion: `develop` and `main` resolve to the same SHA for this
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
