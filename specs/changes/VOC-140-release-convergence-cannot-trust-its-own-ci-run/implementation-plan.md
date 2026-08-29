# VOC-140 — Implementation Plan

## Preconditions and protected areas

- Do not begin until the package is adopted and implementation-authorized.
- Protected areas: caller `tooling/governance/` fixtures, pin, and tests;
  reusable required-check recovery, status attestation, and
  production-merge-guard authorization; GitHub App token mint on release
  and the production-branch merge-gate path; named current-state docs.
- Prerequisites: confirm live `PINNED_SHA.txt` still equals
  `599436835371f27fac52ec6b47a18b36257366ac`. Confirm recovery completion
  still treats required-check SUCCESS as sufficient without requiring the
  parent workflow run to be completed or to be dedicated
  `promotion-pr-validation`. Confirm `select_authoritative` still picks the
  newest `ci / ci` check without parent-run completion. Confirm
  `verify_promotion_required_run_semantics` still raises
  `untrusted_ci_recovery_identity` when `status != completed`. Confirm
  `release.yml` still mints the mutation App token with exactly
  `permission-contents: write`, `permission-issues: write`, and
  `permission-pull-requests: write`, then calls
  `verify-production-merge-guard.sh` with that token before `gh pr merge`.
  Confirm current production-merge-guard tests construct full fixtures with
  `bypass_actors: []` and do not exercise an omitted-field payload. Confirm
  PR #1090, runs `33136633666`, `33136865709`, `33136984634`,
  `33137091931`, and jobs `98738317266`, `98739074178`, `98739420310`
  remain the incident record. Confirm fixture `config/roles.yml` still
  binds implementer/escalation `cursor/composer-2.5` and
  planner/reviewer/retry/plan_reviewer
  `cursor/grok-4.6[effort=high,fast=false]`.
- Resolve current `develop` to a 40-character SHA **before any in-scope
  edit**. Record that SHA as the implementation PR base. Fail closed on
  unrelated/material movement of `develop`. This package's own
  plan/adoption/roster commits after `21eef755…` do not count as
  protected-file drift.
- No bootstrap exception. T00's first run is attempt `1` on a new VOC-140
  carrier from current `develop`. Do not reuse PR #1090 as this package's
  implementation PR. Do not treat the untracked local `karsift-ai-infra/`
  checkout as this repository's tracked tree.
- Preserve two-attempt implementer bounds, exact-SHA independent review,
  fail-closed credentials, and runner/App-token isolation.
- Do not print credential values. Do not rotate `KARSIFT_BOT_*` secrets or
  edit `config/roles.yml`.
- External activation prerequisite: `karsift-ai-infra-bot` Repository
  permissions must request Administration: Read and write and the installation
  owner must approve that change on KARSIFT organization installation
  `148001476`. That installation currently selects all repositories; the guard
  mint must still explicitly restrict each token to the caller repository. No
  secret rotation.
- Do not snapshot the current develop/main gap. Do not add OpenAI execution.
- Do not weaken the production merge guard, add bypass actors, fabricate
  statuses, or manually merge #1090.

## File reconciliation and implementation sequence

### T00 — Unblock release-convergence CI identity and production-merge-guard App-token visibility

| Target | Action | Notes |
|--------|--------|-------|
| `KARSIFT/karsift-ai-infra` recovery/selection/attestation | modify | Never treat in-progress/failed release carriers as attestable `ci / ci`; require dedicated completed `promotion-pr-validation` when no completed non-carrier run exists |
| `KARSIFT/karsift-ai-infra` `release.yml` token mints | modify | Keep mutation mint exactly Contents/Issues/Pull requests write and sole for merge/mutations; add a distinct ephemeral current-caller-repository-scoped Administration-write-only guard mint immediately before guard verification; guard token never reaches merge/mutations/status/issues/PR |
| `KARSIFT/karsift-ai-infra` `merge-gate.yml` production-branch token mints | modify | Apply the same two-token separation to the production-branch guard path |
| `KARSIFT/karsift-ai-infra` `production_merge_guard.py` / `verify-production-merge-guard.sh` | modify | Distinct fail-closed for omitted/non-array `bypass_actors`; keep empty-array required |
| `KARSIFT/karsift-ai-infra` tests | extend | Circular-CI fixture; dedicated promotion-pr-validation; omitted-field subprocess; empty/non-empty bypass; exact mint permissions, repository scope, step ordering, and token-use isolation |
| Caller `tooling/governance/fixtures/karsift-ai-infra/**` | replace from new infra merge | Pin `PINNED_SHA.txt` to that exact merge; mirror every changed authoritative file |
| Caller `tooling/governance/tests/` | extend/reconcile | Advance every current-pin and mirrored-hash assertion while preserving historical authoritative-pin evidence |
| `docs/operations/11-devops-and-ci-cd.md` | modify | Replace App-token mutation-only / contents-issues-PR-only claims with the live recovery-identity and guard-visibility contract |
| `docs/governance/repository-settings.md` | modify | Reconcile stale active-A-003 and release/production-disabled claims to active A-004 and current enabled repository-controlled release/deploy path; retain RL1/RL2 disabled |
| `docs/governance/post-merge-activation-checklist.md`, `docs/operations/19-governance-reconciliation-notes.md` | modify if current-state search confirms stale claims | Reconcile current sections; preserve explicitly historical snapshots |
| `scripts/governance/validate-governance.sh` and relevant tests | modify if they enforce stale current wording | Reconcile assertions with active A-004/current release state without weakening other invariants |
| Fixture `README.md` | modify | Record the new pin and the recovery/guard contract |
| `specs/changes/VOC-139-…/` and `VOC-138-…/` | **do not modify** | Audit evidence |
| `specs/changes/VOC-140-.../t00-evidence.md` | update | Record implementation PR base, new infra merge, identity and token/API change, validation after commit, feasible exact-head binding contract. Do not write the live implementation-head SHA into this file as a self-referential required value |

Ordered steps:

1. Resolve current `develop` to a 40-character SHA before any in-scope edit.
   Record that SHA as the implementation PR base at PR creation. Fail closed
   on unrelated/material movement.
2. Exhaustively search tracked source/docs for old pin and hash assertions,
   mutation-only token claims, active-A-003 claims, and disabled automatic
   release/production-deploy claims; record each match as update, historical,
   or irrelevant. Confirm the known external prerequisite:
   `karsift-ai-infra-bot` Administration: Read and write with installation-owner
   approval on KARSIFT organization installation `148001476`; record its current
   `repository_selection: all` ceiling and retain explicit single-repository
   runtime token scope. Do not rotate secrets.
3. Open the coordinated `KARSIFT/karsift-ai-infra` PR from current infra
   `main`. Implement D01–D08 there with tests. Preserve the mutation mint
   exactly; add the guard-only mint and isolated verification step immediately
   before exact-head merge in both workflow paths. Do not treat an untracked
   nested checkout as already-merged work.
4. Obtain independent exact-revision review of that infra PR and merge it.
   Record the exact merge SHA.
5. After installation-owner approval, run hosted guard verification and record
   sanitized proof that the guard-only token exposes explicit
   `bypass_actors: []`. Omission, non-array, non-empty bypass, or pending App
   permission approval fails closed with the precise approve-and-rerun action.
6. From current caller `develop`, create a new VOC-140 implementation branch.
   Set `PINNED_SHA.txt` and mirror every changed authoritative fixture file
   from that exact merge. Update the named current-state docs. Reconcile all
   current-pin and mirrored-hash assertions in the governance suite; preserve
   historical authoritative-pin constants and package records.
7. Confirm no VOC-139/VOC-138 package file is staged. Confirm `roles.yml` is
   untouched. Confirm no fabricated-status helper or bypass-actor addition
   was added.
8. Track and commit the caller repair. Re-run suites against the committed
   tree. A pass obtained only while untracked is not acceptance.
9. Record evidence in `t00-evidence.md`, including source-search disposition,
   external activation proof, exact token sets/scope/isolation, and the
   same-App/private-key residual risk. This package's caller PR `Closes`
   only its own VOC-140 task issue.
10. After the exact reviewed caller PR merges, rerun dedicated promotion
   recovery if necessary, then `reconcile-release` for #1089 completes
   promotion of #1090 (or the live promotion at the then-current `develop`
   head). Do not add a snapshot-gap task.

## Validation and independent verification

Deterministic commands, as applicable to the final changed file set, after
the repair is tracked and committed:

```bash
bash scripts/governance/validate-governance.sh --base <implementation-pr-base> --head <implementation-head>
bash scripts/governance/classify-change-risk.sh --base <implementation-pr-base> --head <implementation-head>
python3 -m unittest discover -s tooling/governance/tests -p 'test_*.py'
python3 -m unittest discover -s tooling/governance/fixtures/karsift-ai-infra/tests -p 'test_*.py'
git diff --check
```

Also record the exact targeted infra and caller commands that prove:

- newest SUCCESS `ci / ci` on a still-running release carrier is not
  attestable and does not make recovery complete without dispatch;
- dedicated `promotion-pr-validation PR #<n>` is dispatched or selected and
  is required to be completed/successful;
- omitted `bypass_actors` fails distinctly from `production_merge_guard_missing`;
- `bypass_actors: []` still prints `production-merge-guard: ok`;
- non-empty bypass still fails closed;
- release and production-branch merge-gate mints request the diagnosed
  exact two-token permission sets and repository scope; guard verification is
  immediately before merge; the guard token never reaches `gh pr merge` or any
  mutation/status/issue/PR/content step.

Do not treat a missing suite as a pass. Do not treat an untracked-only pass
as acceptance.

Independent verifier (exact reviewed caller SHA, and the infra PR SHA) should
confirm:

- in-progress/failed release carriers are never trusted recovered `ci / ci`;
- dedicated `promotion-pr-validation` is dispatched or selected and must be
  completed/successful;
- VOC-138/VOC-139 exact-identity recovery semantics and doomed-job rerun
  refusal remain;
- the production merge guard still requires an effective active
  repository-owned ruleset, pull-request rule, strict non-empty required
  checks, and `bypass_actors: []`;
- omitted `bypass_actors` fails with a distinct operator-action class;
- tests exercise the real token-visible payload shape and the circular-CI
  parent-run fixture;
- `PINNED_SHA.txt` equals the new infra merge, not stale `59943683…` if that
  merge still fails #1102's class;
- current-state source search is exhaustive; repository-settings and other
  current docs reflect active A-004/current release activation and distinguish
  the unchanged mutation token from the guard-only token;
- external App permission approval and hosted explicit `bypass_actors: []`
  evidence exist; no secret was rotated;
- the same-App/private-key residual risk and optional dedicated guard App are
  documented;
- VOC-139 and VOC-138 package records are unchanged;
- `roles.yml` is unchanged and no OpenAI route was added;
- `t00-evidence.md` names the implementation PR base and new infra merge,
  states that the live head is bound by the App-authored independent-review
  comment/check, and does not require a commit to contain its own SHA;
- the independent-review comment binds this exact live PR head; merge-gate
  would reject a mismatch;
- no snapshot-gap task and no bootstrap exception were used;
- the implementer did not approve or merge its own work.

## Deployment and rollback

- **Staging/production effect:** None intentional for application runtime on
  the ordinary path (tree-equivalent develop-ref update). Staging remains
  path-selected and must run only for a real tree change, not for
  tree-equivalent post-promotion sync.
- **Operational effect:** After this repair is live, `reconcile-release` can
  attest recovered `ci / ci` without selecting its own in-progress carrier
  and can prove the live production merge guard with the isolated guard token
  before the mutation token performs the merge.
  Ordinary release can then merge #1090 and trigger automatic deployment.
- **Rollback trigger:** circular-CI class recurs; omitted `bypass_actors`
  again reports `production_merge_guard_missing`; guard accepts non-empty or
  omitted bypass; dedicated promotion-pr-validation is no longer required to
  be completed; fabricated statuses; snapshot-gap commit; `roles.yml` /
  OpenAI route changed.
- **Rollback mechanism:** Revert the caller fixture/test/doc changes to the
  last reviewed `develop` merge and revert the coordinated infrastructure
  PR. That last-known-good still has the #1089/#1090 deadlock; rollback
  restores a known reviewed state, not a passing promotion PR.
- **Last-known-good reference:** caller `develop` before this package's merge
  (issue-creation `21eef75549226766fc4f78f62f232ee5fbdb8d6d`), pin
  `599436835371f27fac52ec6b47a18b36257366ac`, and `main`
  `0d0b0cdf0692d0349f380e9cae3285b4c7916b05`.
