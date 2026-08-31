# VOC-142 — Release Plan

## Release and deployment authorization

This package does **not** authorize a new production deploy path, any
App-token permission change, reduced review rigor, or production
credential-value changes.

It lands the adoption-recovery repair so native adopt and documented
`reconcile` wait for the complete required roster-check set including
`ci / ci`, and reuse a matching open or already-merged roster PR instead of
failing at `Open roster PR`. Interrupted VOC-141 adoption (task #1111,
roster PR #1112, reconcile `33343250178`) is incident evidence. This package
does not merge #1112 itself and must not create a duplicate VOC-141 task,
roster PR, promotion PR, or release audit.

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
| Coordinated infra merge | implementer + independent verifier | New `KARSIFT/karsift-ai-infra` PR; complete-required-set wait; exact open-PR reuse; already-merged reuse; mismatch/ambiguous fail-closed; unchanged attestation/guard | Exact infra merge SHA recorded in caller `PINNED_SHA.txt` |
| T00 caller merge | implementer + independent verifier | Adoption + task authorization; new VOC-142 branch from current `develop`; pin matches the infra merge | `VOC-142-EV-00` — implementation PR base; infra merge; wait and reuse change; App-authored review/check binds the live PR head |
| Post-repair VOC-141 resume | existing `pipeline.yml` `action=reconcile` | T00 caller changes live on `develop`; plan PR #1110 still merged; #1112 still the matching OPEN carrier or already merged exactly | Ordinary `reconcile` with `plan_pr_number=1110` reuses #1112 when it matches, waits for complete required checks, merges, cleans up, and dispatches the first not-yet-dispatched task exactly once |
| Post-merge promotion | existing `release.yml@main` via ordinary `reconcile-release` | T00 caller changes live on `develop`; valid completion marker | Live same-repository promotion merges; `develop` is advanced to that exact merge SHA before audit close; refs end 0 ahead / 0 behind with identical tree; allowlisted adopt/reconcile/release run metadata is recorded |
| Staging | VOC-111 path selection | Real tree change vs tree-equivalent develop sync | Staging only for the real tree change; tree-equivalent sync must not keep staging scheduled |
| Production deploy | existing `deploy-production.yml` on every `main` push | Promotion merge produced a `main` push | Automatic production deployment runs; this package does not add a new deploy path; verify its exact-SHA result |
| Audit reconciliation | implementer evidence + maintainers | Incident run/job/PR/SHA IDs preserved | Release/task/requirement records close with audit comments naming the exact promotion merge and the independently reviewed head; root issue #1113 closes only after allowlisted metadata from a successful adopt/reconcile run exists. Runs `33343125733`, `33343147453`, `33343250178` and jobs `99342230038`, `99342299218`, `99342577393` remain audit context. |

Monitoring inventory remains unchanged (`monitoring_impact.state: none`).

No OpenAI credential or execution route is needed or authorized.

Incident runs document that native wait completed while `ci / ci` was still
IN_PROGRESS and that documented reconcile then failed at `Open roster PR`.
They are not this package's implementation diff and must not be restaged as
a substitute for ordinary post-correction adopt/release evaluation. Do not
manually merge #1112; after T00, documented reconcile of #1110 is the
recovery.

Allowlisted live-run metadata only (no logs, secrets, or tokens): workflow
identity, event, branch, HEAD SHA, run ID, job ID(s), conclusion,
timestamps. That metadata is closure evidence for #1113, not a VOC-097
evidence-carrier task.

## Rollback

| Item | Value |
|------|-------|
| Trigger | wait-without-`ci / ci` class recurs; reconcile again fails because the exact open roster PR exists; in-progress parents become attestable; fabricated statuses; snapshot-gap commit; `roles.yml` changed; two-token guard changed; #1112 manually merged or duplicated |
| Mechanism | Revert the T00 caller fixture/test/doc changes and revert the coordinated infrastructure PR |
| Owner | Implementer PR + independent verification |
| Validation | Re-run caller governance/fixture suites against the restored `develop` merge and pin `67bdfd13…` (or the recorded pre-T00 pin) |
| Last-known-good | Caller `develop` before this package's merge (issue-creation plan merge `bb4ffdf…`, pin `67bdfd13…` unless superseded before T00). That last-known-good still has the #1113 wait and reuse defects; rollback is to reviewed state, not to a passing adoption recovery |

## Independent verification, human approvals, and closure

Under **active A-004**, no founder `approved` comment is required on
engineering-workflow gates. Required evidence:

1. Exact-SHA independent verification for the infrastructure PR and the
   caller implementation PR. The implementer must not approve or merge its
   own work. The independent-review comment must bind the live PR head
   exactly and must explicitly evaluate complete-required-set wait including
   `ci / ci`, fail-closed partial green snapshots, exact open-PR reuse,
   already-merged reuse, mismatched/ambiguous rejection, duplicate
   task/dispatch suppression, current-state docs, and pin advance.
   Merge-gate must reject any mismatch.
2. Deterministic proof that IN_PROGRESS or unregistered required `ci / ci`
   keeps roster wait incomplete across two stable subset snapshots.
3. Deterministic proof that an exact matching OPEN roster PR is reused and
   `gh pr create` is not called.
4. Deterministic proof that an already-merged exact carrier is not
   duplicated and that mismatched/ambiguous carriers fail closed.
5. Deterministic proof that existing tasks are reused and root dispatch
   happens exactly once.
6. Deterministic proof that wait still avoids `statusCheckRollup` /
   `gh pr checks` and that merge-gate/release still reject in-progress
   parents as attestable completed `ci / ci`.
7. Deterministic proof that tests exercise the live wait and open-PR paths.
8. Deterministic proof that `PINNED_SHA.txt` equals the new infra merge and
   exhaustive old-pin/hash assertions were reconciled.
9. Confirmation that current-state docs match the wait and reuse contract,
   fixture `roles.yml` is unchanged, and no OpenAI route was added.
10. Confirmation that `t00-evidence.md` names the implementation PR base and
    new infra merge, states that the live head is bound by the App-authored
    independent-review comment/check, does not require a commit to contain
    its own SHA, and records validation after the repair was tracked and
    committed.
11. Explicit record that no snapshot-gap task was added; that #1112 was not
    manually merged, closed, recreated, or bypassed; that incident runs
    `33343125733` / `33343147453` / `33343250178` and jobs `99342230038` /
    `99342299218` / `99342577393` are audit context; and that #1113 closes
    only after valid completion markers, successful promotion, and
    allowlisted adopt/reconcile metadata, not from closed state alone.
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
