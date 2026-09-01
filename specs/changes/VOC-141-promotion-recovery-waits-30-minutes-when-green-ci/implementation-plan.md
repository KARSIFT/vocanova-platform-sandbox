# VOC-141 — Implementation Plan

## Preconditions and protected areas

- Do not begin until the package is adopted and implementation-authorized.
- Protected areas: caller `tooling/governance/` fixtures, pin, and tests;
  reusable required-check recovery and status attestation; named
  current-state docs.
- Prerequisites: confirm live `PINNED_SHA.txt` still equals
  `67bdfd13ef875dead23ce4be01d7d0e8b976e289`. Confirm
  `recovery_complete` still returns false when
  `promotion_ci_context_is_attestable` cannot bind SUCCESS `ci / ci` to an
  attestable completed non-carrier parent. Confirm
  `apply_promotion_pr_recovery_plan` still derives `dispatch_contexts` only
  from `plan_required_check_recovery` and therefore drops `pipeline.yml` when
  GitHub-required `ci / ci` is SUCCESS. Confirm
  `format_timeout_diagnostics` / `collect_missing` still omit an
  unattestable-CI reason. Confirm
  `suppress_active_or_successful_dispatches` still drops `pipeline.yml` when
  a supplied `gate_summary` context state is SUCCESS or PENDING. Confirm
  PR #1090, runs `33340381776`, `33340516672`, `33341923799`,
  `33342062118`, and job `99334840338` remain the incident record. Confirm
  fixture `config/roles.yml` still binds implementer/escalation
  `cursor/composer-2.5` and planner/reviewer/retry/plan_reviewer
  `cursor/grok-4.6[effort=high,fast=false]`.
- Resolve current `develop` to a 40-character SHA **before any in-scope
  edit**. Record that SHA as the implementation PR base. Fail closed on
  unrelated/material movement of `develop`. This package's own
  plan/adoption/roster commits after `c3a53bab…` do not count as
  protected-file drift.
- No bootstrap exception. T00's first run is attempt `1` on a new VOC-141
  carrier from current `develop`. Do not reuse PR #1090 as this package's
  implementation PR. Do not treat the untracked local `karsift-ai-infra/`
  checkout as this repository's tracked tree.
- Preserve two-attempt implementer bounds, exact-SHA independent review,
  fail-closed credentials, and runner/App-token isolation.
- Do not print credential values. Do not rotate `KARSIFT_BOT_*` secrets or
  edit `config/roles.yml`.
- Do not snapshot the current develop/main gap. Do not add OpenAI execution.
- Do not weaken the production merge guard, add bypass actors, fabricate
  statuses, change App-token mints, or manually merge a promotion PR.

## File reconciliation and implementation sequence

### T00 — Dispatch dedicated promotion-pr-validation immediately when green CI is unattestable

| Target | Action | Notes |
|--------|--------|-------|
| `KARSIFT/karsift-ai-infra` `config/actions-check-recovery-runner.py` | modify | `apply_promotion_pr_recovery_plan` must consult composed attestable CI (`gate_summary` + `workflow_runs`) and immediately dispatch dedicated `recover-promotion-pr-checks` when GitHub-required SUCCESS is unattestable |
| `KARSIFT/karsift-ai-infra` `config/actions_check_recovery.py` | modify as needed | Timeout/missing collection must name unattestable CI evidence; SUCCESS/PENDING suppress must not drop dedicated dispatch for unattestable CI |
| `KARSIFT/karsift-ai-infra` tests | extend | Live-planner SUCCESS-plus-unattestable-parent dispatch; completed dedicated no-redispatch; active/successful exact dedicated suppression; timeout diagnostic token; unchanged carrier fail-closed |
| Caller `tooling/governance/fixtures/karsift-ai-infra/**` | replace from new infra merge | Pin `PINNED_SHA.txt` to that exact merge; mirror every changed authoritative file |
| Caller `tooling/governance/tests/` | extend/reconcile | Advance every current-pin and mirrored-hash assertion while preserving historical authoritative-pin evidence |
| `docs/operations/11-devops-and-ci-cd.md` | modify | State that GitHub-required SUCCESS is not sufficient when composed CI is unattestable and that recovery dispatches dedicated validation immediately |
| Fixture `README.md` | modify | Record the new pin and the unattestable-SUCCESS dispatch contract |
| `specs/changes/VOC-140-…/`, `VOC-139-…/`, `VOC-138-…/` | **do not modify** | Audit evidence |
| `specs/changes/VOC-141-.../t00-evidence.md` | update | Record implementation PR base, new infra merge, dispatch and diagnostic change, validation after commit, feasible exact-head binding contract. Do not write the live implementation-head SHA into this file as a self-referential required value |

Ordered steps:

1. Resolve current `develop` to a 40-character SHA before any in-scope edit.
   Record that SHA as the implementation PR base at PR creation. Fail closed
   on unrelated/material movement.
2. Exhaustively search tracked source/docs for SUCCESS-completes-recovery
   claims, "dispatch when no completed non-carrier run exists" claims, and
   old pin/hash assertions; record each match as update, historical, or
   irrelevant.
3. Open the coordinated `KARSIFT/karsift-ai-infra` PR from current infra
   `main`. Implement D01–D07 there with tests that call the live planner
   path. Pass composed evidence into `apply_promotion_pr_recovery_plan`. Do
   not treat an untracked nested checkout as already-merged work.
4. Obtain independent exact-revision review of that infra PR and merge it.
   Record the exact merge SHA.
5. From current caller `develop`, create a new VOC-141 implementation branch.
   Set `PINNED_SHA.txt` and mirror every changed authoritative fixture file
   from that exact merge. Update the named current-state docs. Reconcile all
   current-pin and mirrored-hash assertions in the governance suite; preserve
   historical authoritative-pin constants and package records.
6. Confirm no VOC-140/VOC-139/VOC-138 package file is staged. Confirm
   `roles.yml` is untouched. Confirm no fabricated-status helper, bypass-actor
   addition, or App-token mint change was added.
7. Track and commit the caller repair. Re-run suites against the committed
   tree. A pass obtained only while untracked is not acceptance.
8. Record evidence in `t00-evidence.md`, including source-search disposition
   and the timeout-diagnostic token. This package's caller PR `Closes` only
   its own VOC-141 task issue.
9. After the exact reviewed caller PR merges, ordinary `reconcile-release`
   completes promotion of the live same-repository promotion at the
   then-current `develop` head. Do not add a snapshot-gap task.

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

- required SUCCESS `ci / ci` plus unattestable parent dispatches exactly one
  dedicated `recover-promotion-pr-checks` immediately;
- a valid completed dedicated parent completes without redispatch;
- active/successful exact dedicated recovery suppresses duplicates and
  release carriers do not;
- timeout diagnostics name unattestable CI evidence rather than
  `missing_checks: none` alone;
- production/release-carrier fail-closed boundaries and the two-token guard
  remain unchanged.

Do not treat a missing suite as a pass. Do not treat an untracked-only pass
as acceptance.

Independent verifier (exact reviewed caller SHA, and the infra PR SHA) should
confirm:

- SUCCESS-plus-unattestable-parent immediately dispatches dedicated
  `promotion-pr-validation`;
- completed dedicated parent completes without redispatch;
- active/successful exact dedicated recovery suppresses duplicates;
- timeout diagnostics identify unattestable CI evidence;
- VOC-138/VOC-139/VOC-140 exact-identity recovery, doomed-job rerun refusal,
  release-carrier rejection, and two-token production merge guard remain;
- tests exercise `apply_promotion_pr_recovery_plan` or the live loop with
  GitHub-required SUCCESS rows;
- `PINNED_SHA.txt` equals the new infra merge, not stale `67bdfd13…` if that
  merge still fails #1109's class;
- current-state source search is exhaustive; current docs state that
  GitHub-required SUCCESS is not sufficient when composed CI is unattestable;
- VOC-140, VOC-139, and VOC-138 package records are unchanged;
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
- **Operational effect:** After this repair is live, `reconcile-release`
  immediately dispatches dedicated `promotion-pr-validation` when
  GitHub-required `ci / ci` is SUCCESS and composed evidence is unattestable,
  instead of polling for 1,800 seconds.
- **Rollback trigger:** SUCCESS-plus-unattestable-parent class recurs;
  dedicated promotion-pr-validation is no longer dispatched; timeout
  diagnostics again omit the unattestable-CI reason; release carriers become
  attestable; doomed `ci / ci` jobs are rerun; fabricated statuses;
  snapshot-gap commit; `roles.yml` / OpenAI route changed; two-token guard
  changed.
- **Rollback mechanism:** Revert the caller fixture/test/doc changes to the
  last reviewed `develop` merge and revert the coordinated infrastructure
  PR. That last-known-good still has the #1109 hang; rollback restores a
  known reviewed state, not a passing promotion recovery.
- **Last-known-good reference:** caller `develop` before this package's merge
  (issue-creation promotion head `c3a53bab3035b7f08c0fb959bdf1b56bf330d291`)
  and pin `67bdfd13ef875dead23ce4be01d7d0e8b976e289`.
