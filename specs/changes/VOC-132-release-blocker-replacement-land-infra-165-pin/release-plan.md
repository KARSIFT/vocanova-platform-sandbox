# VOC-132 — Release Plan

## Release and deployment authorization

This package does **not** authorize a new production deploy path, broader
workflow authority, reduced review rigor, or production credential changes.

It lands the caller-side pin/fixture/test/doc contract for already-merged
infrastructure #165 (`8ce2b77a09a729e458a9f4cbea1ca26eb114d398`) on a new
VOC-132 carrier, so both release jobs restore shared lifecycle policy after
caller checkout, with a complete VOC-112 no-change boundary. VOC-129 caller
PR #1046 remains the VOC-129 carrier. Exhausted VOC-130-T00 (#1049 / PR
#1051) and exhausted VOC-131-T00 (PR #1056) are not retried and are not
publishable sources.

Adoption and task implementation authorization remain separate gates under
active A-004. This draft itself is not adopted and does not authorize
implementation.

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
| T00 caller merge | implementer + independent verifier | Adoption + task authorization; infra #165 already merged as `8ce2b77…`; new VOC-132 branch from current `develop`; PR #1051 and PR #1056 not reused; VOC-129 #1046 not rewritten; five VOC-112 no-change paths identical to develop base | `VOC-132-EV-00` — pin SHA; identify/converge restore coverage; recorded #165 hashes; complete VOC-112 boundary; preserved #164 contracts; exact reviewed caller SHA |
| VOC-129 skipped promotion | existing `release.yml@main` (#165) | Valid App-authored VOC-129 completion marker; repaired reusable workflow | Promotion merge exists or `reconcile-release` completes it; closed state alone is not proof; no manufactured VOC-129 marker |
| Exhausted VOC-130 close | maintainers + audit comments | VOC-130-T00 exhausted; PR #1051 never merged | #1051 closed as superseded (never merged) after this replacement is promoted; #1049 closed with audit links naming the VOC-132 merge; no manufactured VOC-130 marker |
| Exhausted VOC-131 close | maintainers + audit comments | VOC-131-T00 exhausted; PR #1056 never merged | #1056 closed as superseded (never merged) after this replacement is promoted; VOC-131 task/root records closed with audit links naming the VOC-132 merge; no manufactured VOC-131 marker |
| Post-merge promotion of this package | existing `release.yml@main` (#165) | T00 caller changes live on `develop` with a valid completion marker | Promotion merge exists; `develop` is advanced to that exact merge SHA before audit close; refs end at the same SHA |
| Staging | VOC-111 path selection | Tree-equivalent develop sync or specs-only paths | No unnecessary staging deployment; allowlisted runtime/deploy paths still deploy |
| Production deploy | existing `deploy-production.yml` on `main` push | Promotion merge produced a `main` push; path selection if any | Production deploys if selected by existing path rules; this package does not add a new deploy path |
| Audit reconciliation | implementer evidence + maintainers | VOC-129, VOC-130, VOC-131, and VOC-132 records reconcile | Release/task/requirement records close with audit comments naming the exact promotion merges; root issue #1057 closes only after that evidence exists |

Monitoring inventory remains unchanged (`monitoring_impact.state: none`).

No OpenAI credential or execution route is needed or authorized.

Hosted `release.yml@main` may already contain the #165 restore before this
caller pin merges. `reconcile-release` for VOC-129 may therefore succeed
independently of this package. That does not remove this package's
pin/test/doc obligation and must not become a second VOC-132 task.

## Rollback

| Item | Value |
|------|-------|
| Trigger | Pin not equal to `8ce2b77…`; restore missing in identify or converge; unique develop commits erased; operator SHA paste; `roles.yml` changed; any of the five VOC-112 no-change paths rewritten; PR #1051 or PR #1056 reused; `/tmp`-only byte comparison; false-revert evidence; snapshot-gap commit |
| Mechanism | Revert the T00 caller fixture/test/doc changes |
| Owner | Implementer PR + independent verification |
| Validation | Re-run caller governance/fixture suites against the restored `develop` pin `863fc1f35b1d35e4981a59166b0e939be1a2b681`, VOC-112 `subject_revision` `f9d11e23…`, and the published provenance-test fail-closed `local` mode |
| Last-known-good | Caller `develop` before this package's merge (pin `863fc1f…`, five VOC-112 no-change paths unchanged). That last-known-good still lacks the #165 restore in the fixture; rollback is to reviewed state, not to a working #165 pin. Do not revert infrastructure #165. Do not restore #1051's VOC-112 JSON retargets or #1056's provenance-test weakening |

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
3. Deterministic proof that fixture `release.yml` and
   `tests/test_release_policy.py` match recorded SHA-256 hashes of infra merge
   `8ce2b77…` without a machine-specific `/tmp` checkout.
4. Deterministic proof that both `identify` and `converge` restore shared
   policy after caller checkout and before task-completion helpers, using
   `job.workflow_repository`, `job.workflow_sha`, and
   `persist-credentials: false`.
5. Deterministic proof that the #164 checkout-ref ordering / missing-`develop`
   path, `mergeCommit.oid` sync, unique-develop fail-closed behavior, and
   `reconcile-production-change` remain.
6. Confirmation that roster markers, required-check recovery, App-token
   isolation, match-head-commit, two-attempt bound, Cursor Composer/Grok
   roles, unchanged `roles.yml`, sanitized raw-error controls, and
   non-closing source PR remain.
7. Deterministic proof that all five VOC-112 no-change paths are
   byte-identical to the carrier `develop` base, JSON `subject_revision`
   remains `f9d11e23…`, the provenance test still fail-closes `local` mode on
   a missing capture commit in a full checkout, and none of those paths
   appear in the implementation diff.
8. Confirmation that `t00-evidence.md` names the exact reviewed head and does
   not claim a protected-path revert unless that path is absent from the
   diff.
9. Confirmation that current-state fixture README/comments name the #165 pin
   and restore, and that historical CHANGELOG / A-003 / VOC-075 / VOC-127 /
   VOC-129 / VOC-130 / VOC-131 records, `AGENTS.md`, the navigator skill, and
   the provenance test were not rewritten.
10. Explicit record that PR #1051 and PR #1056 are not reused, merged, or
    modified; that VOC-129 PR #1046 is not re-implemented; that no VOC-129,
    VOC-130, or VOC-131 completion marker was manufactured; that no
    snapshot-gap task was added; and that VOC-129 / VOC-130 / VOC-131 /
    VOC-132 close only after valid completion markers and successful
    promotion, not from closed state alone. Root issue #1057 closes only
    after that audit evidence exists.
11. Confirmation that no bootstrap exception was used.
12. After promotion: `develop` and `main` resolve to the same SHA for this
    package's promotion merge; tree-equivalent convergence did not trigger an
    unnecessary staging deployment; VOC-129's skipped promotion is completed
    or explicitly recorded as already completed through the repaired path;
    exhausted VOC-130 and VOC-131 are closed with audit links naming this
    replacement.

Closure: T00 merges with passing deterministic checks and exact-SHA independent
verification, then promotes through the existing #165 `release.yml@main`
path. Do not conflate package merge with runtime release; this package has
no direct product deployment effect. An issue-close event may wake
evaluation, but closed state alone is not completion proof.
