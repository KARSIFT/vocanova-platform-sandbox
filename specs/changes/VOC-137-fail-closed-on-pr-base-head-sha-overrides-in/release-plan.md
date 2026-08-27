# VOC-137 — Release Plan

## Release and deployment authorization

This package does **not** authorize a new production deploy path, broader
workflow authority, reduced review rigor, or production credential changes.

It lands the caller-side scanner correction so PR base/head SHA overrides
fail closed on every scannable caller executable outside the exact infra
fixture mirror, without changing the already-pinned infrastructure #167
fixture, VOC-112 provenance bytes, roles, or VOC-136 audit records. VOC-136
caller PR #1080 remains the VOC-136 carrier and remains historical evidence.

Adoption and task implementation authorization remain separate gates under
active A-004. This draft itself is not adopted and does not authorize
implementation.

This package is not a request to promote the current integration tip to
production as its work. The scanner correction proceeds through normal
`develop` merge, staging only for the real tree change, `develop` → `main`
promotion of outstanding completed packages, exact post-promotion `develop`
convergence without a staging redeploy for tree-equivalent sync, and
applicable deployment. Closed state alone is not completion proof. Do not
snapshot the current develop/main gap (`karsift-ai-infra#15`).

## Preconditions, monitoring, and outcome

| Phase | Owner | Preconditions | Outcome evidence |
|-------|-------|---------------|------------------|
| Plan merge | plan reviewer + merge-gate | Draft package; `automatic_merge_allowed: true`; valid `monitoring_impact` | Adopted package on `develop`; no duplicate task roster |
| T00 caller merge | implementer + independent verifier | Adoption + task authorization; new VOC-137 branch from current `develop`; PR #1080 not reused as this review; filename gate removed; arbitrary-filename shell/Node/Python cases fail closed; fixture pin frozen at `b263c0c…`; VOC-136 package records untouched | `VOC-137-EV-00` — implementation PR base; scan-scope change; negative/positive results; pin freeze; App-authored review/check binds the live PR head and explicitly evaluates the arbitrary-filename cases |
| Already-merged VOC-136 promotion | existing `release.yml@main` | Valid App-authored VOC-136 completion marker; this scanner gap closed on `develop` | Ordinary release evaluation (or `reconcile-release`) may promote VOC-136 together with this package; do not manufacture a VOC-136 marker; do not merge draft release PR #1082 as a substitute |
| Post-merge promotion of this package | existing `release.yml@main` | T00 caller changes live on `develop` with a valid completion marker | Promotion merge exists; `develop` is advanced to that exact merge SHA before audit close; refs end 0 ahead / 0 behind with identical tree |
| Staging | VOC-111 path selection | Real tree change vs tree-equivalent develop sync | Staging only for the real tree change; tree-equivalent sync must not keep staging scheduled |
| Production deploy | existing `deploy-production.yml` on `main` push | Promotion merge produced a `main` push; path selection if any | Production deploys if selected by existing path rules; this package does not add a new deploy path; verify deployment if selected |
| Audit reconciliation | implementer evidence + maintainers | VOC-136 records preserved; VOC-137 records close | Release/task/requirement records close with audit comments naming the exact promotion merges and the independently reviewed head; root issue #1083 closes only after that evidence exists. Canceled runs `33113425829` / `33113547909` remain audit context. |

Monitoring inventory remains unchanged (`monitoring_impact.state: none`).

No OpenAI credential or execution route is needed or authorized.

Canceled release runs `33113425829` and `33113547909` and draft release PR
#1082 document that automatic release was stopped before `main` changed.
They are not this package's implementation diff and must not be restaged as
a substitute for ordinary post-correction release evaluation.

## Rollback

| Item | Value |
|------|-------|
| Trigger | Filename gate still present; arbitrary-wrapper example accepted; `tooling/governance/tests/` excluded from the scan; fixture pin retargeted; eight VOC-112 paths rewritten; VOC-136 records rewritten; evidence mutated at test time; self-referential exact-head SHA required; PR #1080 review reused as this verdict; snapshot-gap commit; `roles.yml` changed |
| Mechanism | Revert the T00 caller scanner/test/doc changes |
| Owner | Implementer PR + independent verification |
| Validation | Re-run caller governance/fixture suites against the restored `develop` merge `0cee20c87e0411a95f368d2b7d39ac2bb118dfb8`, pin `b263c0c…`, and unmodified eight no-change paths relative to `b9e74fc2…` |
| Last-known-good | Caller `develop` before this package's merge (issue-creation `0cee20c…`, pin `b263c0c…`, eight no-change paths unchanged vs `b9e74fc2…`). That last-known-good still has the filename heuristic; rollback is to reviewed state, not to a complete fail-closed PR SHA scan. Do not revert infrastructure #167. Do not rewrite VOC-136 records as part of rollback |

## Independent verification, human approvals, and closure

Under **active A-004**, no founder `approved` comment is required on
engineering-workflow gates. Required evidence:

1. Exact-SHA independent verification for the caller implementation PR. The
   implementer must not approve or merge its own work. The independent-review
   comment must bind the live PR head exactly and must explicitly evaluate
   the arbitrary-filename shell/Node/Python negative cases and benign
   controls. Merge-gate must reject any mismatch.
2. Deterministic proof that `PR_SHA_SET_PATTERN` is no longer gated on
   filename, and that `scripts/arbitrary-wrapper.sh` with the issue payload
   fails closed.
3. Deterministic proof of the Node `process.env.PR_HEAD_SHA` and Python
   `os.environ["PR_BASE_SHA"]` arbitrary-filename cases, and that added or
   modified `*.py` outside the fixture mirror is scanned.
4. Deterministic proof that benign discussion, source-safe pattern
   construction, and the excluded fixture `run-app-checks.sh` do not
   false-positive, and that `SCAN_EXCLUDE_PREFIXES` does not list
   `tooling/governance/tests/`.
5. Deterministic proof that `PINNED_SHA.txt` still equals `b263c0c…` and that
   no fixture-mirror path appears in the implementation diff.
6. Deterministic proof that all eight VOC-112 no-change paths remain
   byte-identical to `b9e74fc2…`.
7. Confirmation that fixture `roles.yml` is unchanged and no OpenAI route
   was added.
8. Confirmation that `t00-evidence.md` names the implementation PR base,
   states that the live head is bound by the App-authored
   independent-review comment/check, does not require a commit to contain
   its own SHA, and records validation after the regression was tracked and
   committed.
9. Confirmation that VOC-136 package records were not rewritten and that PR
   #1080's review is preserved as audit evidence rather than reused as this
   verdict.
10. Explicit record that no snapshot-gap task was added; that canceled
    release runs `33113425829` / `33113547909` and draft PR #1082 are not
    this package's implementation work; and that VOC-136 and VOC-137 close
    only after valid completion markers and successful promotion, not from
    closed state alone. Root issue #1083 closes only after that audit
    evidence exists.
11. Confirmation that no bootstrap exception was used.
12. After promotion: `develop` and `main` resolve to the same SHA for this
    package's promotion merge (0 ahead / 0 behind, identical tree); staging
    ran only for the real tree change; tree-equivalent convergence did not
    trigger an unnecessary staging deployment; production deployment is
    verified if selected. Post-merge audit may record the independently
    reviewed head and the promotion merge SHA.

Closure: T00 merges with passing deterministic checks and exact-SHA
independent verification, then ordinary release promotes outstanding
completed packages including already-merged VOC-136 and this correction. Do
not conflate package merge with runtime release; this package has no direct
product deployment effect. An issue-close event may wake evaluation, but
closed state alone is not completion proof.
