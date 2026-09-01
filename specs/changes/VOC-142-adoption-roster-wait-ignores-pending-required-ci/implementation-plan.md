# VOC-142 — Implementation Plan

## Preconditions and protected areas

- Do not begin until the package is adopted and implementation-authorized.
- Protected areas: caller `tooling/governance/` fixtures, pin, and tests;
  reusable adoption wait/reuse; named current-state docs including
  `AGENTS.md`.
- Prerequisites: confirm live `PINNED_SHA.txt` still equals
  `67bdfd13ef875dead23ce4be01d7d0e8b976e289`, or record a later pin if
  VOC-141 (or another package) advanced it, and still confirm that pin's
  `adopt.yml` wait completes on two stable zero-pending snapshots and that
  `Open roster PR` always calls `gh pr create` when `changed == 'true'`.
  Confirm `_workflow_runs` without `--ordinary-pr-gate` still omits
  non-attestable in-progress parents. Confirm `_ordinary_pr_pipeline_parent`
  still rejects `karsift/roster-*` refs. Confirm PR #1112, task #1111, runs
  `33343125733`, `33343147453`, `33343250178`, and jobs `99342230038`,
  `99342299218`, `99342577393` remain the incident record. Confirm fixture
  `config/roles.yml` still binds implementer/escalation
  `cursor/composer-2.5` and planner/reviewer/retry/plan_reviewer
  `cursor/grok-4.6[effort=high,fast=false]`.
- Resolve current `develop` to a 40-character SHA **before any in-scope
  edit**. Record that SHA as the implementation PR base. Fail closed on
  unrelated/material movement of `develop`. This package's own
  plan/adoption/roster commits after `bb4ffdf…` do not count as
  protected-file drift. Do not edit `specs/changes/VOC-141-…/`.
- No bootstrap exception. T00's first run is attempt `1` on a new VOC-142
  carrier from current `develop`. Do not reuse PR #1110, #1112, or task
  #1111 as this package's implementation artifacts. Do not treat the
  untracked local `karsift-ai-infra/` checkout as this repository's tracked
  tree.
- Preserve two-attempt implementer bounds, exact-SHA independent review,
  fail-closed credentials, and runner/App-token isolation.
- Do not print credential values. Do not rotate `KARSIFT_BOT_*` secrets or
  edit `config/roles.yml`.
- Do not snapshot the current develop/main gap. Do not add OpenAI execution.
- Do not weaken the production merge guard, add bypass actors, fabricate
  statuses, change App-token mints, or manually merge a roster PR.

## File reconciliation and implementation sequence

### T00 — Wait for the complete required roster-check set and reuse the exact open roster PR on reconcile

| Target | Action | Notes |
|--------|--------|-------|
| `KARSIFT/karsift-ai-infra` `.github/workflows/adopt.yml` | modify | Wait must require the complete ruleset-required set including `ci / ci`; `Open roster PR` must reuse a matching OPEN carrier and create only when none exists |
| `KARSIFT/karsift-ai-infra` `config/authoritative-checks-runner.py` / `config/authoritative_checks.py` | modify as needed | Wait-completeness must observe IN_PROGRESS/unregistered required rows as not-ready without making in-progress parents attestable SUCCESS for merge-gate/release |
| `KARSIFT/karsift-ai-infra` tests including `tests/test_adoption_handoff.py` | extend | Live wait with IN_PROGRESS/late `ci / ci`; open-PR reuse; already-merged reuse; mismatch/ambiguous rejection; duplicate task/dispatch suppression |
| Caller `tooling/governance/fixtures/karsift-ai-infra/**` | replace from new infra merge | Pin `PINNED_SHA.txt` to that exact merge; mirror every changed authoritative file |
| Caller `tooling/governance/tests/` | extend/reconcile | Advance every current-pin and mirrored-hash assertion while preserving historical authoritative-pin evidence |
| `AGENTS.md` | modify | Reconcile procedure must say wait requires the complete required set and reconcile reuses a matching open or already-merged roster PR |
| Fixture `README.md` | modify | Record the new pin and that two stable subset snapshots are not complete |
| `docs/operations/15-ai-native-product-and-engineering-operating-model.md` | modify if current-state claims match | Only if exhaustive search finds a live false claim |
| `specs/changes/VOC-141-…/`, `VOC-140-…/`, `VOC-139-…/` | **do not modify** | Audit evidence |
| `specs/changes/VOC-142-.../t00-evidence.md` | update | Record implementation PR base, new infra merge, wait and reuse change, validation after commit, feasible exact-head binding contract. Do not write the live implementation-head SHA into this file as a self-referential required value |

Ordered steps:

1. Resolve current `develop` to a 40-character SHA before any in-scope edit.
   Record that SHA as the implementation PR base at PR creation. Fail closed
   on unrelated/material movement.
2. Exhaustively search tracked source/docs for two-stable-snapshot-completes-
   wait claims, reconcile-always-creates-PR claims, and old pin/hash
   assertions; record each match as update, historical, or irrelevant.
3. Open the coordinated `KARSIFT/karsift-ai-infra` PR from current infra
   `main`. Implement D01–D07 there with tests that call the live wait and
   open-PR paths. Do not treat an untracked nested checkout as already-merged
   work. Do not add `--ordinary-pr-gate` unchanged as the sole fix: that
   helper still rejects `karsift/roster-*`.
4. Obtain independent exact-revision review of that infra PR and merge it.
   Record the exact merge SHA.
5. From current caller `develop`, create a new VOC-142 implementation branch.
   Set `PINNED_SHA.txt` and mirror every changed authoritative fixture file
   from that exact merge. Update the named current-state docs. Reconcile all
   current-pin and mirrored-hash assertions in the governance suite; preserve
   historical authoritative-pin constants and package records.
6. Confirm no VOC-141/VOC-140/VOC-139 package file is staged. Confirm
   `roles.yml` is untouched. Confirm no fabricated-status helper, bypass-actor
   addition, App-token mint change, or #1112 merge/close/recreate helper was
   added.
7. Track and commit the caller repair. Re-run suites against the committed
   tree. A pass obtained only while untracked is not acceptance.
8. Record evidence in `t00-evidence.md`, including source-search disposition
   and the wait-completeness/reuse contract. This package's caller PR
   `Closes` only its own VOC-142 task issue.
9. After the exact reviewed caller PR merges, ordinary `reconcile` for plan
   PR #1110 may resume #1112 when that PR still matches. Do not add a
   snapshot-gap task.

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

- IN_PROGRESS required `ci / ci` plus SUCCESS `governance-policy`/`validate`
  does not complete roster wait across two stable snapshots;
- late-unregistered required `ci / ci` does not complete wait;
- an exact matching OPEN roster PR is reused and `gh pr create` is not
  called;
- already-merged exact carriers are not duplicated;
- mismatched/ambiguous carriers fail closed;
- existing tasks are reused and root dispatch happens exactly once;
- merge-gate/release still reject in-progress parents as attestable
  completed `ci / ci`.

Do not treat a missing suite as a pass. Do not treat an untracked-only pass
as acceptance.

Independent verifier (exact reviewed caller SHA, and the infra PR SHA) should
confirm:

- complete-required-set wait including `ci / ci`;
- fail-closed partial green snapshots;
- exact open-PR reuse and already-merged reuse;
- mismatched/ambiguous rejection;
- duplicate task/dispatch suppression;
- wait still avoids `statusCheckRollup` / `gh pr checks`;
- merge-gate/release in-progress-parent rejection remains;
- tests exercise the live wait and open-PR paths, not YAML-only
  `stable_green_count` assertions;
- `PINNED_SHA.txt` equals the new infra merge, not stale `67bdfd13…` if that
  merge still fails #1113's class;
- current-state source search is exhaustive; current docs state wait
  requires the complete required set and reconcile reuses matching carriers;
- VOC-141, VOC-140, and VOC-139 package records are unchanged;
- `roles.yml` is unchanged and no OpenAI route was added;
- `t00-evidence.md` names the implementation PR base and new infra merge,
  states that the live head is bound by the App-authored independent-review
  comment/check, and does not require a commit to contain its own SHA;
- the independent-review comment binds this exact live PR head; merge-gate
  would reject a mismatch;
- no snapshot-gap task, no bootstrap exception, and no manual merge of
  #1112 were used;
- the implementer did not approve or merge its own work.

## Deployment and rollback

- **Staging/production effect:** None intentional for application runtime on
  the ordinary path (tree-equivalent develop-ref update). Staging remains
  path-selected and must run only for a real tree change, not for
  tree-equivalent post-promotion sync.
- **Operational effect:** After this repair is live, native adopt and
  `reconcile` wait for the complete required roster-check set including
  `ci / ci`, and reuse a matching open or already-merged roster PR instead
  of failing on create.
- **Rollback trigger:** wait-without-`ci / ci` class recurs; reconcile
  again fails because the exact open roster PR exists; in-progress parents
  become attestable; fabricated statuses; snapshot-gap commit; `roles.yml` /
  OpenAI route changed; two-token guard changed; #1112 manually merged or
  duplicated.
- **Rollback mechanism:** Revert the caller fixture/test/doc changes to the
  last reviewed `develop` merge and revert the coordinated infrastructure
  PR. That last-known-good still has the #1113 wait and reuse defects;
  rollback restores a known reviewed state, not a passing adoption recovery.
- **Last-known-good reference:** caller `develop` before this package's merge
  (issue-creation plan merge `bb4ffdf5d53d27baf4c25c28caf3acfeda9e07a2`)
  and pin `67bdfd13ef875dead23ce4be01d7d0e8b976e289` unless a later reviewed
  pin superseded it before T00 started.
