# VOC-136 — Release Plan

## Release and deployment authorization

This package does **not** authorize a new production deploy path, broader
workflow authority, reduced review rigor, or production credential changes.

It lands the caller-side pin/fixture/test/doc contract for already-merged
infrastructure #167 (`b263c0c110591cc798b89277dfc35542abb1597b`) on a new
VOC-136 carrier, so application checks receive an immutable PR base/head pair
without fetching evidence, both release jobs restore shared lifecycle policy
after caller checkout, implementer publication preserves helpers before the
unrestricted model, a complete eight-path no-change boundary and exhaustive
executable bypass scan are kept, and exact-revision binding is fail-closed
without a self-referential Git-tree SHA. VOC-129 caller PR #1046 remains the
VOC-129 carrier. Exhausted VOC-130-T00 (#1049 / PR #1051), exhausted
VOC-131-T00 (PR #1056), unpublished VOC-132-T00 (#1059), exhausted
VOC-133-T00 (#1063 / PR #1065), exhausted VOC-134-T00 (#1068 / PR #1070),
and exhausted VOC-135-T00 (#1073 / PR #1075) are not retried and are not
publishable sources.

Adoption and task implementation authorization remain separate gates under
active A-004. This draft itself is not adopted and does not authorize
implementation.

This package is not a request to promote the current integration tip to
production as its work. The pin proceeds through normal `develop` merge,
staging only for the real tree change, `develop` → `main` promotion, exact
post-promotion `develop` convergence without a staging redeploy for
tree-equivalent sync, and applicable deployment. Closed state alone is not
completion proof. Do not snapshot the current develop/main gap
(`karsift-ai-infra#15`).

## Preconditions, monitoring, and outcome

| Phase | Owner | Preconditions | Outcome evidence |
|-------|-------|---------------|------------------|
| Plan merge | plan reviewer + merge-gate | Draft package; `automatic_merge_allowed: true`; valid `monitoring_impact` | Adopted package on `develop`; no duplicate task roster |
| T00 caller merge | implementer + independent verifier | Adoption + task authorization; infra #167 already merged as `b263c0c…`; new VOC-136 branch from current `develop`; PR #1051 / #1056 / #1065 / #1070 / #1075 not reused; VOC-132-T00 through VOC-135-T00 not redispatched; VOC-129 #1046 not rewritten; eight no-change paths identical to the protected comparison anchor; exhaustive scan includes `tooling/governance/tests/**` and is scan-clean after commit; feasible exact-head evidence | `VOC-136-EV-00` — pin SHA; identify/converge restore coverage; nested-checkout helper-lifetime coverage; PR-context coverage; recorded #167 hashes; complete eight-path boundary vs protected comparison anchor; implementation PR base; exhaustive scan; negative cases; `package.json` identity; preserved #164 contracts; App-authored review/check binds the live PR head; final infra merge |
| VOC-129 skipped promotion | existing `release.yml@main` (#165/#166/#167) | Valid App-authored VOC-129 completion marker; repaired reusable workflow | Promotion merge exists or `reconcile-release` completes it; closed state alone is not proof; no manufactured VOC-129 marker |
| Exhausted VOC-127 / VOC-130 close | maintainers + audit comments | VOC-130-T00 exhausted; PR #1051 never merged; VOC-127 carriers superseded | Closed as superseded after this replacement is promoted; no manufactured completion markers |
| Exhausted VOC-131 close | maintainers + audit comments | VOC-131-T00 exhausted; PR #1056 never merged | #1056 closed as superseded (never merged) after this replacement is promoted; no manufactured VOC-131 marker |
| Superseded VOC-132 close | maintainers + audit comments | VOC-132-T00 (#1059) unpublished; not redispatched | #1059 closed as superseded after this replacement is promoted; no manufactured VOC-132 marker; VOC-132's adopted #165 pin is not silently rewritten |
| Exhausted VOC-133 close | maintainers + audit comments | VOC-133-T00 (#1063 / PR #1065) exhausted; not redispatched | #1065 closed as superseded (never merged) after this replacement is promoted; no manufactured VOC-133 marker |
| Exhausted VOC-134 close | maintainers + audit comments | VOC-134-T00 (#1068 / PR #1070) exhausted; not redispatched | #1070 closed as superseded (never merged) after this replacement is promoted; no manufactured VOC-134 marker |
| Exhausted VOC-135 close | maintainers + audit comments | VOC-135-T00 (#1073 / PR #1075) exhausted at heads `7c3e821…` and `dcd2da9…`; closed unmerged; not redispatched | #1075 remains closed as superseded (never merged); #1073 / #1071 remain closed not-planned with audit links naming the VOC-136 merge; no manufactured VOC-135 marker; VOC-135's adopted pin is not silently rewritten in place |
| Post-merge promotion of this package | existing `release.yml@main` (#167) | T00 caller changes live on `develop` with a valid completion marker | Promotion merge exists; `develop` is advanced to that exact merge SHA before audit close; refs end 0 ahead / 0 behind with identical tree |
| Staging | VOC-111 path selection | Real tree change vs tree-equivalent develop sync | Staging only for the real tree change; tree-equivalent sync must not keep staging scheduled |
| Production deploy | existing `deploy-production.yml` on `main` push | Promotion merge produced a `main` push; path selection if any | Production deploys if selected by existing path rules; this package does not add a new deploy path; verify deployment if selected |
| Audit reconciliation | implementer evidence + maintainers | VOC-127 through VOC-136 records reconcile | Release/task/requirement records close with audit comments naming the exact promotion merges and the independently reviewed head; root issue #1076 closes only after that evidence exists. Clean only obsolete remote automation branches/PRs/issues. Preserve unrelated VOC-128 and all user worktrees. |

Monitoring inventory remains unchanged (`monitoring_impact.state: none`).

No OpenAI credential or execution route is needed or authorized.

Hosted `implement.yml@main` / `release.yml@main` / `ci.yml@main` may already
contain the #167 PR-context repair, the #166 helper-lifetime repair, and the
#165 restore before this caller pin merges. That does not remove this
package's pin/test/doc obligation and must not become a second VOC-136 task.

## Rollback

| Item | Value |
|------|-------|
| Trigger | Pin not equal to `b263c0c…`; restore missing in identify or converge; helper copy missing before the unrestricted model; PR-context pair missing or evidence fetch introduced; unique develop commits erased; operator SHA paste; `roles.yml` changed; any of the eight no-change paths rewritten; capture-commit fetch helper, hydrate helper, or provenance-mode wrapper added under any filename; `tooling/governance/tests/` excluded from the complete-diff scan; evidence mutated at test time; self-referential exact-head SHA required; PR #1051 / #1056 / #1065 / #1070 / #1075 reused; VOC-132-T00 through VOC-135-T00 redispatched; `/tmp`-only byte comparison; false-revert evidence; snapshot-gap commit |
| Mechanism | Revert the T00 caller fixture/test/doc changes |
| Owner | Implementer PR + independent verification |
| Validation | Re-run caller governance/fixture suites against the restored `develop` pin `863fc1f35b1d35e4981a59166b0e939be1a2b681`, VOC-112 `subject_revision` `f9d11e23…`, published provenance-test fail-closed `local` mode, and unmodified eight no-change paths relative to protected comparison anchor `b9e74fc2…` |
| Last-known-good | Caller `develop` before this package's merge (pin `863fc1f…`, eight no-change paths unchanged vs `b9e74fc2…`). That last-known-good still lacks the #165 restore, #166 helper-lifetime contract, and #167 PR-context contract in the fixture; rollback is to reviewed state, not to a working #167 pin. Do not revert infrastructure #167. Do not restore #1051's VOC-112 JSON retargets, #1056's provenance-test weakening, #1065's `package.json` / evidence-stamp changes, #1070's runner fetch / hydrate helpers, or #1075's `SCAN_EXCLUDE_PREFIXES` weakening |

## Independent verification, human approvals, and closure

Under **active A-004**, no founder `approved` comment is required on
engineering-workflow gates. Required evidence:

1. Exact-SHA independent verification for the caller implementation PR. The
   implementer must not approve or merge its own work. The independent-review
   comment must bind the live PR head exactly. Merge-gate must reject any
   mismatch. Infrastructure #167 is already independently reviewed at merge
   `b263c0c110591cc798b89277dfc35542abb1597b` (PASS head
   `eb11c4fc6841ec73816e2e064dcd449d98c1e933`); this package does not reopen
   that review except to confirm the caller consumed that exact SHA.
2. Deterministic proof that `PINNED_SHA.txt` and live pin assertions equal
   `b263c0c…` and do not equal `863fc1f…`, `8ce2b77…`, or `f3d791…`.
3. Deterministic proof that the mirrored authoritative files match recorded
   SHA-256 hashes of infra merge `b263c0c…` without a machine-specific `/tmp`
   checkout.
4. Deterministic proof that both `identify` and `converge` restore shared
   policy after caller checkout and before task-completion helpers, using
   `job.workflow_repository`, `job.workflow_sha`, and
   `persist-credentials: false`.
5. Deterministic proof that `implement.yml` copies helpers before the
   unrestricted model and that nested-checkout classification fails closed on
   symlink / non-directory / parent-Git inheritance while `absent` continues
   caller publication.
6. Deterministic proof that `run-app-checks.sh` binds exact PR context
   without fetching evidence; that unchanged capture fixtures select
   `pr-validation`; that add/modify/delete select `pr-ancestry`; that
   comparison errors fail closed; that `ci.yml` uses `fetch-depth: 0` and
   event SHAs; and that implementer pre-push uses the integration anchor plus
   live HEAD including after self-correction.
7. Deterministic proof that the #164 checkout-ref ordering / missing-`develop`
   path, `mergeCommit.oid` sync, unique-develop fail-closed behavior, and
   `reconcile-production-change` remain.
8. Confirmation that roster markers, required-check recovery, App-token
   isolation, match-head-commit, two-attempt bound, Cursor Composer/Grok
   roles, unchanged `roles.yml`, sanitized raw-error controls, and
   non-closing source PR remain.
9. Deterministic proof that all eight no-change paths are byte-identical to
   the protected comparison anchor, JSON `subject_revision` remains
   `f9d11e23…`, the provenance test still fail-closes `local` mode on a
   missing capture commit in a full checkout, none of those paths appear in
   the implementation diff against that anchor, and the regression fails—not
   skips—if the exact commit or a named path cannot resolve.
10. Deterministic proof that an exhaustive caller-diff scan of changed
    executable paths, including `tooling/governance/tests/**` and this
    regression's own module, finds no capture fetch, hydrate/materialize
    helper, provenance-mode override, import side effect, environment
    override, evidence stamp, or local fail-closed bypass under any
    filename; that `SCAN_EXCLUDE_PREFIXES` does not list
    `tooling/governance/tests/`; that required negative cases fail; that
    benign mentions do not false-positive; and that the tracked committed
    regression is scan-clean.
11. Confirmation that `t00-evidence.md` names the protected comparison
    anchor, the implementation PR base, and the final infra merge, states
    that the live head is bound by the App-authored independent-review
    comment/check, does not require a commit to contain its own SHA, and
    does not claim a protected-path revert unless that path is absent from
    the diff against the protected comparison anchor.
12. Confirmation that current-state fixture README/comments name the #167 pin,
    restore, helper-lifetime, and PR-context contracts, and that historical
    caller CHANGELOG / A-003 / VOC-075 / VOC-127 / VOC-129 / VOC-130 /
    VOC-131 / VOC-132 / VOC-133 / VOC-134 / VOC-135 records, `AGENTS.md`, the
    navigator skill, the provenance test, the runner, `validate-workspace.mjs`,
    and `package.json` were not rewritten.
13. Explicit record that PR #1051, PR #1056, PR #1065, PR #1070, and PR #1075
    are not reused, merged, cherry-picked, or modified; that VOC-132-T00
    (#1059), VOC-133-T00 (#1063), VOC-134-T00 (#1068), and VOC-135-T00
    (#1073) are not redispatched; that VOC-129 PR #1046 is not
    re-implemented; that no VOC-127 through VOC-135 completion marker was
    manufactured; that no snapshot-gap task was added; that PR #1075
    attempt-2 PASS WITH NON-BLOCKING FINDINGS was not treated as sufficient;
    and that VOC-127 through VOC-136 close only after valid completion
    markers and successful promotion, not from closed state alone. Root issue
    #1076 closes only after that audit evidence exists.
14. Confirmation that no bootstrap exception was used.
15. After promotion: `develop` and `main` resolve to the same SHA for this
    package's promotion merge (0 ahead / 0 behind, identical tree); staging
    ran only for the real tree change; tree-equivalent convergence did not
    trigger an unnecessary staging deployment; production deployment is
    verified if selected; exhausted / superseded VOC-127 / VOC-130 /
    VOC-131 / VOC-132 / VOC-133 / VOC-134 / VOC-135 are closed with audit
    links naming this replacement. Cleanup preserves unrelated VOC-128 and
    all user worktrees. Post-merge audit may record the independently
    reviewed head and the promotion merge SHA.

Closure: T00 merges with passing deterministic checks and exact-SHA independent
verification, then promotes through the existing #167 `release.yml@main`
path. Do not conflate package merge with runtime release; this package has
no direct product deployment effect. An issue-close event may wake
evaluation, but closed state alone is not completion proof.
