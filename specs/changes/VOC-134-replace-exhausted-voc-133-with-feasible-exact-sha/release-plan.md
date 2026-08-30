# VOC-134 — Release Plan

## Release and deployment authorization

This package does **not** authorize a new production deploy path, broader
workflow authority, reduced review rigor, or production credential changes.

It lands the caller-side pin/fixture/test/doc contract for already-merged
infrastructure #166 (`f3d79177bf8a9abe0dae550f39502165d494c576`) on a new
VOC-134 carrier, so both release jobs restore shared lifecycle policy after
caller checkout, implementer publication preserves helpers before the
unrestricted model, a complete VOC-112 no-change boundary and unchanged
`package.json` are kept, and exact-revision binding is fail-closed without a
self-referential Git-tree SHA. VOC-129 caller PR #1046 remains the VOC-129
carrier. Exhausted VOC-130-T00 (#1049 / PR #1051), exhausted VOC-131-T00
(PR #1056), unpublished VOC-132-T00 (#1059), and exhausted VOC-133-T00
(#1063 / PR #1065) are not retried and are not publishable sources.

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
| T00 caller merge | implementer + independent verifier | Adoption + task authorization; infra #166 already merged as `f3d791…`; new VOC-134 branch from current `develop`; PR #1051 / #1056 / #1065 not reused; VOC-132-T00 and VOC-133-T00 not redispatched; VOC-129 #1046 not rewritten; five VOC-112 no-change paths and `package.json` identical to the immutable carrier-base SHA; feasible exact-head evidence | `VOC-134-EV-00` — pin SHA; identify/converge restore coverage; nested-checkout helper-lifetime coverage; recorded #166 hashes; complete VOC-112 boundary vs immutable carrier-base SHA; `package.json` identity; preserved #164 contracts; App-authored review/check binds the live PR head; final infra merge |
| VOC-129 skipped promotion | existing `release.yml@main` (#165/#166) | Valid App-authored VOC-129 completion marker; repaired reusable workflow | Promotion merge exists or `reconcile-release` completes it; closed state alone is not proof; no manufactured VOC-129 marker |
| Exhausted VOC-130 close | maintainers + audit comments | VOC-130-T00 exhausted; PR #1051 never merged | #1051 closed as superseded (never merged) after this replacement is promoted; #1049 closed with audit links naming the VOC-134 merge; no manufactured VOC-130 marker |
| Exhausted VOC-131 close | maintainers + audit comments | VOC-131-T00 exhausted; PR #1056 never merged | #1056 closed as superseded (never merged) after this replacement is promoted; VOC-131 task/root records closed with audit links naming the VOC-134 merge; no manufactured VOC-131 marker |
| Superseded VOC-132 close | maintainers + audit comments | VOC-132-T00 (#1059) unpublished after run `33079499176`; not redispatched | #1059 closed as superseded after this replacement is promoted; no manufactured VOC-132 marker; VOC-132's adopted #165 pin is not silently rewritten |
| Exhausted VOC-133 close | maintainers + audit comments | VOC-133-T00 (#1063 / PR #1065) exhausted at heads `e88fbda…` and `70930cf…`; not redispatched | #1065 closed as superseded (never merged) after this replacement is promoted; #1063 closed with audit links naming the VOC-134 merge; no manufactured VOC-133 marker; VOC-133's adopted pin is not silently rewritten in place |
| Post-merge promotion of this package | existing `release.yml@main` (#166) | T00 caller changes live on `develop` with a valid completion marker | Promotion merge exists; `develop` is advanced to that exact merge SHA before audit close; refs end at the same SHA |
| Staging | VOC-111 path selection | Tree-equivalent develop sync or specs-only paths | No unnecessary staging deployment; allowlisted runtime/deploy paths still deploy |
| Production deploy | existing `deploy-production.yml` on `main` push | Promotion merge produced a `main` push; path selection if any | Production deploys if selected by existing path rules; this package does not add a new deploy path |
| Audit reconciliation | implementer evidence + maintainers | VOC-129 through VOC-134 records reconcile | Release/task/requirement records close with audit comments naming the exact promotion merges and the independently reviewed head; root issue #1066 closes only after that evidence exists |

Monitoring inventory remains unchanged (`monitoring_impact.state: none`).

No OpenAI credential or execution route is needed or authorized.

Hosted `implement.yml@main` / `release.yml@main` may already contain the #166
helper-lifetime repair and the #165 restore before this caller pin merges.
`reconcile-release` for VOC-129 may therefore succeed independently of this
package. That does not remove this package's pin/test/doc obligation and must
not become a second VOC-134 task.

## Rollback

| Item | Value |
|------|-------|
| Trigger | Pin not equal to `f3d791…`; restore missing in identify or converge; helper copy missing before the unrestricted model; unique develop commits erased; operator SHA paste; `roles.yml` changed; any of the five VOC-112 no-change paths or `package.json` rewritten; capture-commit fetch helper or provenance-mode wrapper added; evidence mutated at test time; self-referential exact-head SHA required; PR #1051 / #1056 / #1065 reused; VOC-132-T00 or VOC-133-T00 redispatched; `/tmp`-only byte comparison; false-revert evidence; snapshot-gap commit |
| Mechanism | Revert the T00 caller fixture/test/doc changes |
| Owner | Implementer PR + independent verification |
| Validation | Re-run caller governance/fixture suites against the restored `develop` pin `863fc1f35b1d35e4981a59166b0e939be1a2b681`, VOC-112 `subject_revision` `f9d11e23…`, published provenance-test fail-closed `local` mode, and unmodified `package.json` |
| Last-known-good | Caller `develop` before this package's merge (pin `863fc1f…`, five VOC-112 no-change paths and `package.json` unchanged; expected carrier-base SHA `95a779f9…`). That last-known-good still lacks the #165 restore and #166 helper-lifetime contract in the fixture; rollback is to reviewed state, not to a working #166 pin. Do not revert infrastructure #166. Do not restore #1051's VOC-112 JSON retargets, #1056's provenance-test weakening, or #1065's `package.json` / evidence-stamp changes |

## Independent verification, human approvals, and closure

Under **active A-004**, no founder `approved` comment is required on
engineering-workflow gates. Required evidence:

1. Exact-SHA independent verification for the caller implementation PR. The
   implementer must not approve or merge its own work. The independent-review
   comment must bind the live PR head exactly. Merge-gate must reject any
   mismatch. Infrastructure #166 is already independently reviewed at merge
   `f3d79177bf8a9abe0dae550f39502165d494c576` (PASS head
   `1488619d0d37aaa179d8e739bfe931881d6c51aa`; initial failed head
   `ce86f5d77c733e0e9f30397c167fbd1dfc7c5a8f`); this package does not reopen
   that review except to confirm the caller consumed that exact SHA.
2. Deterministic proof that `PINNED_SHA.txt` and live pin assertions equal
   `f3d791…` and do not equal `863fc1f…` or `8ce2b77…`.
3. Deterministic proof that the mirrored authoritative files match recorded
   SHA-256 hashes of infra merge `f3d791…` without a machine-specific `/tmp`
   checkout.
4. Deterministic proof that both `identify` and `converge` restore shared
   policy after caller checkout and before task-completion helpers, using
   `job.workflow_repository`, `job.workflow_sha`, and
   `persist-credentials: false`.
5. Deterministic proof that `implement.yml` copies helpers before the
   unrestricted model and that nested-checkout classification fails closed on
   symlink / non-directory / parent-Git inheritance while `absent` continues
   caller publication.
6. Deterministic proof that the #164 checkout-ref ordering / missing-`develop`
   path, `mergeCommit.oid` sync, unique-develop fail-closed behavior, and
   `reconcile-production-change` remain.
7. Confirmation that roster markers, required-check recovery, App-token
   isolation, match-head-commit, two-attempt bound, Cursor Composer/Grok
   roles, unchanged `roles.yml`, sanitized raw-error controls, and
   non-closing source PR remain.
8. Deterministic proof that all five VOC-112 no-change paths and
   `package.json` are byte-identical to the immutable carrier-base SHA, JSON
   `subject_revision` remains `f9d11e23…`, the provenance test still
   fail-closes `local` mode on a missing capture commit in a full checkout,
   none of those paths appear in the implementation diff, and the regression
   fails—not skips—if the exact commit or a named path cannot resolve.
9. Confirmation that `t00-evidence.md` names the immutable carrier base and
   the final infra merge, states that the live head is bound by the
   App-authored independent-review comment/check, does not require a commit
   to contain its own SHA, and does not claim a protected-path revert unless
   that path is absent from the diff.
10. Confirmation that current-state fixture README/comments name the #166 pin,
    restore, and helper-lifetime contract, and that historical caller
    CHANGELOG / A-003 / VOC-075 / VOC-127 / VOC-129 / VOC-130 / VOC-131 /
    VOC-132 / VOC-133 records, `AGENTS.md`, the navigator skill, the
    provenance test, and `package.json` were not rewritten.
11. Explicit record that PR #1051, PR #1056, and PR #1065 are not reused,
    merged, cherry-picked, or modified; that VOC-132-T00 (#1059) and
    VOC-133-T00 (#1063) are not redispatched; that VOC-129 PR #1046 is not
    re-implemented; that no VOC-129 through VOC-133 completion marker was
    manufactured; that no snapshot-gap task was added; and that VOC-129
    through VOC-134 close only after valid completion markers and successful
    promotion, not from closed state alone. Root issue #1066 closes only after
    that audit evidence exists.
12. Confirmation that no bootstrap exception was used.
13. After promotion: `develop` and `main` resolve to the same SHA for this
    package's promotion merge; tree-equivalent convergence did not trigger an
    unnecessary staging deployment; VOC-129's skipped promotion is completed
    or explicitly recorded as already completed through the repaired path;
    exhausted VOC-130 / VOC-131 / VOC-133 and unpublished VOC-132 are closed
    with audit links naming this replacement. Post-merge audit may record the
    independently reviewed head and the promotion merge SHA.

Closure: T00 merges with passing deterministic checks and exact-SHA independent
verification, then promotes through the existing #166 `release.yml@main`
path. Do not conflate package merge with runtime release; this package has
no direct product deployment effect. An issue-close event may wake
evaluation, but closed state alone is not completion proof.
