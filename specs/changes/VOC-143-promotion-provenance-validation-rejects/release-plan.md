# VOC-143 — Release Plan

## Release and deployment authorization

This package does **not** authorize a new production deploy path, any
App-token permission change, reduced review rigor, or production
credential-value changes.

It lands the provenance-assertion repair so promotion-path required
`validate` (`squash-safe-push`) and `ci / ci` (promotion `pr-validation`)
accept a legitimate current `AGENTS.md` documentation update while the
historical VOC-112 fixture remains unmodified. Interrupted VOC-142 promotion
(PR #1119, head `376e00dd…`, audit #1118) is incident evidence. This package
does not merge #1119 itself and must not create a duplicate promotion PR or
release audit.

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
| T00 caller merge | implementer + independent verifier | Adoption + task authorization; new VOC-143 branch from current `develop`; ancestor-bind for `squash-safe-push` and promotion `pr-validation` `AGENTS.md` | `VOC-143-EV-00` — implementation PR base; assertion change; App-authored review/check binds the live PR head |
| Post-repair VOC-142 promotion resume | existing `pipeline.yml` `action=reconcile-release` | T00 caller changes live on `develop`; release audit #1118 still open; #1119 still the matching OPEN carrier or already merged exactly | Ordinary `reconcile-release` with `release_issue_number=1118` re-evaluates #1119 when it matches, waits for complete required checks, merges, and closes the audit |
| Post-merge promotion | existing `release.yml@main` via ordinary `reconcile-release` | T00 caller changes live on `develop`; valid completion marker | Live same-repository promotion merges; `develop` is advanced to that exact merge SHA before audit close; refs end 0 ahead / 0 behind with identical tree; allowlisted recovery/release run metadata is recorded |
| Staging | VOC-111 path selection | Real tree change vs tree-equivalent develop sync | Staging only for the real tree change; tree-equivalent sync must not keep staging scheduled |
| Production deploy | existing `deploy-production.yml` on every `main` push | Promotion merge produced a `main` push | Automatic production deployment runs; this package does not add a new deploy path; verify its exact-SHA result |
| Audit reconciliation | implementer evidence + maintainers | Incident PR/SHA/issue IDs preserved | Release/task/requirement records close with audit comments naming the exact promotion merge and the independently reviewed head; root issue #1120 closes only after allowlisted metadata from a successful recovery/release run exists. PR #1119, head `376e00dd…`, and audit #1118 remain audit context. |

Monitoring inventory remains unchanged (`monitoring_impact.state: none`).

No OpenAI credential or execution route is needed or authorized.

Incident PR #1119 documents that promotion `validate` and `ci / ci` failed
because historical fixture `agents_sha256` was compared to current
`AGENTS.md`. It is not this package's implementation diff and must not be
restaged as a substitute for ordinary post-correction release evaluation. Do
not manually merge #1119; after T00, documented `reconcile-release` of #1118
is the recovery.

Allowlisted live-run metadata only (no logs, secrets, or tokens): workflow
identity, event, branch, HEAD SHA, run ID, job ID(s), conclusion,
timestamps. That metadata is closure evidence for #1120, not a VOC-097
evidence-carrier task.

## Rollback

| Item | Value |
|------|-------|
| Trigger | working-tree-equality class recurs under `squash-safe-push`; promotion `pr-validation` again requires HEAD `AGENTS.md` equality; tampered hashes pass; `local` / `pr-ancestry` / ordinary `pr-validation` weakened; navigator HEAD-binding dropped; promotion check identity switched; fixtures recaptured; snapshot-gap commit; `roles.yml` changed; #1119 manually merged or duplicated |
| Mechanism | Revert the T00 caller test/doc changes |
| Owner | Implementer PR + independent verification |
| Validation | Re-run VOC-112 and VOC-114 Node suites against the restored `develop` merge |
| Last-known-good | Caller `develop` before this package's merge (issue-creation promotion head `376e00dd…`, pin `8993e867…`). That last-known-good still has the #1120 provenance defect; rollback is to reviewed state, not to a passing promotion |

## Independent verification, human approvals, and closure

Under **active A-004**, no founder `approved` comment is required on
engineering-workflow gates. Required evidence:

1. Exact-SHA independent verification for the caller implementation PR. The
   implementer must not approve or merge its own work. The independent-review
   comment must bind the live PR head exactly and must explicitly evaluate
   `squash-safe-push` historical-ancestor `AGENTS.md` bind, promotion
   `pr-validation` historical-ancestor `AGENTS.md` bind, retained `local` /
   `pr-ancestry` / ordinary `pr-validation`, retained navigator binding,
   unchanged promotion check identity, unre-captured fixtures, and
   current-state docs. Merge-gate must reject any mismatch.
2. Deterministic proof that `squash-safe-push` accepts historical fixture
   `agents_sha256` when working-tree `AGENTS.md` differs.
3. Deterministic proof that `squash-safe-push` fails closed on an
   unfound/tampered `agents_sha256`.
4. Deterministic proof that promotion `pr-validation` accepts the same
   historical fixture when HEAD `AGENTS.md` differs and navigator hashes
   remain HEAD-bound.
5. Deterministic proof that `local`, `pr-ancestry`, and ordinary
   `pr-validation` are not weakened.
6. Deterministic proof that promotion check identity and VOC-112 JSON
   fixtures are unchanged.
7. Deterministic proof that tests exercise `assertCapturedRevision`.
8. Confirmation that current-state docs match the ancestor-bind contract,
   fixture `roles.yml` is unchanged, and no OpenAI route was added.
9. Confirmation that `t00-evidence.md` names the implementation PR base,
   states that the live head is bound by the App-authored independent-review
   comment/check, does not require a commit to contain its own SHA, and
   records validation after the repair was tracked and committed.
10. Explicit record that no snapshot-gap task was added; that #1119 was not
    manually merged, closed, recreated, or bypassed; that PR #1119 / head
    `376e00dd…` / audit #1118 are audit context; and that #1120 closes only
    after valid completion markers, successful promotion, and allowlisted
    recovery/release metadata, not from closed state alone.
11. Confirmation that no pin-advance or duplicate promotion PR / release
    audit was created.
12. After promotion: `develop` and `main` resolve to the same SHA for this
    package's promotion merge (0 ahead / 0 behind, identical tree); staging
    ran only for the real tree change; tree-equivalent convergence did not
    trigger an unnecessary staging deployment; automatic production
    deployment from the `main` push is verified. Post-merge audit may record
    the independently reviewed head and the promotion merge SHA.

Closure: T00 merges with passing deterministic checks and exact-SHA
independent verification, then ordinary `reconcile-release` of #1118 may
resume #1119, and ordinary later promotion includes this correction. Do not
conflate package merge with runtime release; this package has no direct
product deployment effect. An issue-close event may wake evaluation, but
closed state alone is not completion proof.
