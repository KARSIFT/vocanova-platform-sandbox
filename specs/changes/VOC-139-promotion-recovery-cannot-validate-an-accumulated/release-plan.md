# VOC-139 — Release Plan

## Release and deployment authorization

This package does **not** authorize a new production deploy path, broader
workflow authority, reduced review rigor, or production credential changes.

It lands the promotion source-hash and recovery-metadata repair so a
same-repository `main` ← `develop` promotion whose hashed sources
legitimately changed on `develop` can pass required `ci / ci`, and so
`reconcile-release` can resolve immutable PR metadata without a checkout.
Promotion PR #1090 remains the live promotion carrier after this repair
merges; this package does not merge #1090 itself.

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
| Coordinated infra merge | implementer + independent verifier | New `KARSIFT/karsift-ai-infra` PR; promotion hash-anchor export; repository-explicit no-checkout metadata; ordinary merge-base retention | Exact infra merge SHA recorded in caller `PINNED_SHA.txt` |
| T00 caller merge | implementer + independent verifier | Adoption + task authorization; new VOC-139 branch from current `develop`; pin matches the infra merge; seven VOC-112 paths unchanged vs `b9e74fc2…`; provenance test implements head-binding | `VOC-139-EV-00` — implementation PR base; infra merge; hash-rule and metadata change; App-authored review/check binds the live PR head |
| Post-merge promotion | existing `release.yml@main` via `reconcile-release` for #1089 | T00 caller changes live on `develop`; valid completion marker; required `ci / ci` can pass on the live promotion PR; recovery metadata can resolve without a checkout | Promotion PR #1090 (or successor at the then-current `develop` head) merges; `develop` is advanced to that exact merge SHA before audit close; refs end 0 ahead / 0 behind with identical tree; allowlisted recovery/release run metadata is recorded |
| Staging | VOC-111 path selection | Real tree change vs tree-equivalent develop sync | Staging only for the real tree change; tree-equivalent sync must not keep staging scheduled |
| Production deploy | existing `deploy-production.yml` on every `main` push | Promotion merge produced a `main` push | Automatic production deployment runs; this package does not add a new deploy path; verify its exact-SHA result |
| Audit reconciliation | implementer evidence + maintainers | Incident run/job IDs preserved | Release/task/requirement records close with audit comments naming the exact promotion merge and the independently reviewed head; root issue #1096 closes only after allowlisted metadata from the successful recovery/release run exists. Runs `33130426061`, `33130473438`, `33130527834` and jobs `98718413924`, `98718739912` remain audit context. |

Monitoring inventory remains unchanged (`monitoring_impact.state: none`).

No OpenAI credential or execution route is needed or authorized.

Incident runs document that promotion was blocked before `main` changed.
They are not this package's implementation diff and must not be restaged as
a substitute for ordinary post-correction release evaluation.

Allowlisted live-run metadata only (no logs, secrets, or tokens): workflow
identity, event, branch, HEAD SHA, run ID, job ID(s), conclusion,
timestamps. That metadata is closure evidence for #1096, not a VOC-097
evidence-carrier task.

## Rollback

| Item | Value |
|------|-------|
| Trigger | Accumulated promotion still requires merge-base hashes; ordinary merge-base `pr-validation` no longer fails closed; promotion PR switched to `--squash-safe-push`; fetch/hydrate/recapture helper added; seven VOC-112 paths rewritten; metadata still lacks `-R`; evidence mutated at test time; self-referential exact-head SHA required; snapshot-gap commit; `roles.yml` changed |
| Mechanism | Revert the T00 caller fixture/test/doc/workflow/provenance-test changes and revert the coordinated infrastructure PR |
| Owner | Implementer PR + independent verification |
| Validation | Re-run caller governance/fixture suites against the restored `develop` merge `4812fb91…`, pin `123735c80…`, and unmodified seven no-change paths relative to `b9e74fc2…` |
| Last-known-good | Caller `develop` before this package's merge (issue-creation `4812fb91…`, pin `123735c80…`, seven no-change paths unchanged vs `b9e74fc2…`, `main` `0d0b0cdf…`). That last-known-good still has the #1090 deadlock; rollback is to reviewed state, not to a passing promotion PR. Do not recapture VOC-112 fixtures as rollback |

## Independent verification, human approvals, and closure

Under **active A-004**, no founder `approved` comment is required on
engineering-workflow gates. Required evidence:

1. Exact-SHA independent verification for the infrastructure PR and the
   caller implementation PR. The implementer must not approve or merge its
   own work. The independent-review comment must bind the live PR head
   exactly and must explicitly evaluate the accumulated-promotion head-hash
   case, ordinary merge-base `pr-validation` retention, no-checkout metadata,
   identity negatives, no-fetch/no-recapture constraint, and seven-path
   freeze. Merge-gate must reject any mismatch.
2. Deterministic proof that an authenticated same-repository `main` ←
   `develop` promotion keeps `pr-validation` and binds hashes to
   `PR_HEAD_SHA` / working tree.
3. Deterministic proof that `VOC-112-TEST-12` and `VOC-112-TEST-13` pass
   for a checkout whose `AGENTS.md` differs between base and head.
4. Deterministic proof that ordinary `pr-validation` remains merge-base
   anchored and that ordinary fixture-changing PRs remain `pr-ancestry`.
5. Deterministic proof that malformed SHAs, unrelated commits,
   unrelated repository/PR, wrong refs, and tampered current hashes still
   fail closed.
6. Deterministic proof that `promotion-pr-metadata` is repository-explicit,
   has no checkout, and is exercised in a non-git directory.
7. Deterministic proof that recovery still rejects weaker same-head
   evidence and does not rerun doomed `pull_request` jobs.
8. Deterministic proof that `PINNED_SHA.txt` equals the new infra merge and
   that all seven remaining VOC-112 no-change paths remain byte-identical to
   `b9e74fc2…`.
9. Confirmation that current-state docs no longer claim accumulated
   promotions are merge-base hash-bound, that fixture `roles.yml` is
   unchanged, and that no OpenAI route was added.
10. Confirmation that `t00-evidence.md` names the implementation PR base and
    new infra merge, states that the live head is bound by the App-authored
    independent-review comment/check, does not require a commit to contain
    its own SHA, and records validation after the repair was tracked and
    committed.
11. Explicit record that no snapshot-gap task was added; that incident runs
    `33130426061` / `33130473438` / `33130527834` and jobs `98718413924` /
    `98718739912` are audit context; and that #1089 / #1090 / #1096 close
    only after valid completion markers, successful promotion, and
    allowlisted recovery/release metadata, not from closed state alone.
12. Confirmation that no bootstrap exception was used.
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
