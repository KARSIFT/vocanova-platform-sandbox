# VOC-130 — Release Plan

## Release and deployment authorization

This package does **not** authorize a new production deploy path, broader
workflow authority, reduced review rigor, or production credential changes.

It lands the caller-side pin/fixture/test/doc contract for already-merged
infrastructure #165 (`8ce2b77a09a729e458a9f4cbea1ca26eb114d398`) on a new
VOC-130 carrier, so both release jobs restore shared lifecycle policy after
caller checkout. VOC-129 caller PR #1046 remains the VOC-129 carrier.

Adoption and task implementation authorization remain separate gates under
active A-004.

This package is not a request to promote the current integration tip to
production as its work. The pin proceeds through normal `develop` merge,
`develop` → `main` promotion, exact post-promotion `develop` convergence, and
applicable deployment. Tree-equivalent convergence must not trigger an
unnecessary staging deployment. Closed state alone is not completion proof.
Do not snapshot the current develop/main gap (`karsift-ai-infra#15`).

## Preconditions, monitoring, and outcome

| Phase | Owner | Preconditions | Outcome evidence |
|-------|-------|---------------|------------------|
| Plan merge | plan reviewer + merge-gate | Draft package; `automatic_merge_allowed: true`; valid `monitoring_impact` | Adopted package on `develop`; no duplicate task roster |
| T00 caller merge | implementer + independent verifier | Adoption + task authorization; infra #165 already merged as `8ce2b77…`; new VOC-130 branch from current `develop`; VOC-129 #1046 not rewritten | `VOC-130-EV-00` — pin SHA; identify/converge restore coverage; preserved #164 contracts; exact reviewed caller SHA |
| VOC-129 skipped promotion | existing `release.yml@main` (#165) | Valid App-authored VOC-129 completion marker; repaired reusable workflow | Promotion merge exists or `reconcile-release` completes it; closed state alone is not proof |
| Post-merge promotion of this package | existing `release.yml@main` (#165) | T00 caller changes live on `develop` with a valid completion marker | Promotion merge exists; `develop` is advanced to that exact merge SHA before audit close; refs end at the same SHA |
| Staging | VOC-111 path selection | Tree-equivalent develop sync or specs-only paths | No unnecessary staging deployment; allowlisted runtime/deploy paths still deploy |
| Audit reconciliation | implementer evidence + maintainers | VOC-129 and VOC-130 promotions succeed | Release/task/requirement records close with audit comments naming both exact promotion merges |

Monitoring inventory remains unchanged (`monitoring_impact.state: none`).

No OpenAI credential or execution route is needed or authorized.

Hosted `release.yml@main` may already contain the #165 restore before this
caller pin merges. `reconcile-release` for VOC-129 may therefore succeed
independently of this package. That does not remove this package's
pin/test/doc obligation and must not become a second VOC-130 task.

## Rollback

| Item | Value |
|------|-------|
| Trigger | Pin not equal to `8ce2b77…`; restore missing in identify or converge; unique develop commits erased; operator SHA paste; `roles.yml` changed; snapshot-gap commit |
| Mechanism | Revert the T00 caller fixture/test/doc changes |
| Owner | Implementer PR + independent verification |
| Validation | Re-run caller governance/fixture suites against the restored `develop` pin `863fc1f35b1d35e4981a59166b0e939be1a2b681` |
| Last-known-good | Caller `develop` before this package's merge (pin `863fc1f…`). That last-known-good still lacks the #165 restore in the fixture; rollback is to reviewed state, not to a working #165 pin. Do not revert infrastructure #165 |

## Independent verification, human approvals, and closure

Under **active A-004**, no founder `approved` comment is required on
engineering-workflow gates. Required evidence:

1. Exact-SHA independent verification for the caller implementation PR. The
   implementer must not approve or merge its own work. Infrastructure #165 is
   already independently reviewed at merge
   `8ce2b77a09a729e458a9f4cbea1ca26eb114d398` (reviewed head
   `e33931d02f7bdbb094ae8177fd88324cd19ac5ce`); this package does not reopen
   that review except to confirm the caller consumed that exact SHA.
2. Deterministic proof that `PINNED_SHA.txt` and live pin assertions equal
   `8ce2b77…` and do not equal `863fc1f…`.
3. Deterministic proof that both `identify` and `converge` restore shared
   policy after caller checkout and before task-completion helpers, using
   `job.workflow_repository` and `job.workflow_sha`.
4. Deterministic proof that the #164 checkout-ref ordering / missing-`develop`
   path, `mergeCommit.oid` sync, unique-develop fail-closed behavior, and
   `reconcile-production-change` remain.
5. Confirmation that roster markers, required-check recovery, App-token
   isolation, match-head-commit, two-attempt bound, Cursor Composer/Grok
   roles, unchanged `roles.yml`, sanitized raw-error controls, and
   non-closing source PR remain.
6. Confirmation that current-state fixture README/comments name the #165 pin
   and restore, and that historical CHANGELOG / A-003 / VOC-075 / VOC-127 /
   VOC-129 records were not rewritten.
7. Explicit record that VOC-129 PR #1046 is not re-implemented; that no
   snapshot-gap task was added; and that VOC-129 / VOC-130 close only after
   valid completion markers and successful promotion, not from closed state
   alone.
8. Confirmation that no bootstrap exception was used.
9. After promotion: `develop` and `main` resolve to the same SHA for this
   package's promotion merge; tree-equivalent convergence did not trigger an
   unnecessary staging deployment; VOC-129's skipped promotion is completed
   or explicitly recorded as already completed through the repaired path.

Closure: T00 merges with passing deterministic checks and exact-SHA independent
verification, then promotes through the existing #165 `release.yml@main`
path. Do not conflate package merge with runtime release; this package has
no direct product deployment effect. An issue-close event may wake
evaluation, but closed state alone is not completion proof.
