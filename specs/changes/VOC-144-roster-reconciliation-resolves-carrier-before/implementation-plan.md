# VOC-144 — Implementation Plan

## Preconditions and protected areas

- Do not begin until the package is adopted and implementation-authorized.
- Protected areas: caller `tooling/governance/` fixtures, pin, and tests;
  reusable adoption carrier resolution; named current-state docs including
  `AGENTS.md`.
- Prerequisites: confirm live `PINNED_SHA.txt` still equals
  `8993e867640dfb604dec0466c4e0787e68d8e258`, or record a later pin if
  another package advanced it, and still confirm that pin's `Open roster PR`
  invokes `roster-carrier-runner.py` once immediately after `Push roster
  branch`, and that `resolve_roster_carrier` returns
  `MISMATCHED_OPEN_CARRIER` for a same-ref OPEN PR whose listed SHA differs.
  Confirm PR #1112 remains the VOC-141 carrier, and runs `33437239322` /
  `33437514152` remain the incident record. Confirm fixture
  `config/roles.yml` still binds implementer/escalation
  `cursor/composer-2.5` and planner/reviewer/retry/plan_reviewer
  `cursor/grok-4.6[effort=high,fast=false]`.
- Resolve current `develop` to a 40-character SHA **before any in-scope
  edit**. Record that SHA as the implementation PR base. Fail closed on
  unrelated/material movement of `develop`. This package's own
  plan/adoption/roster commits do not count as protected-file drift. Do not
  edit `specs/changes/VOC-142-…/`, `VOC-141-…/`, `VOC-140-…/`, or
  `VOC-139-…/`.
- No bootstrap exception. T00's first run is attempt `1` on a new VOC-144
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

### T00 — Boundedly wait for existing roster-PR head metadata to expose the pushed SHA, then reuse that carrier

| Target | Action | Notes |
|--------|--------|-------|
| `KARSIFT/karsift-ai-infra` `config/roster-carrier-runner.py` | modify | After listing pulls, boundedly re-fetch the unique same-repo/same-ref/same-base OPEN PR until listed SHA equals local HEAD or the named bound is exhausted; then call `resolve_roster_carrier` |
| `KARSIFT/karsift-ai-infra` `config/roster_carrier.py` | preserve identity predicate; optional helper | Keep single-snapshot fail-closed identity. A small "is this unique SHA-lag?" helper is allowed; do not treat stale SHA as `reuse_open` on one snapshot |
| `KARSIFT/karsift-ai-infra` `.github/workflows/adopt.yml` | comment/header only unless a second runner invocation is required | Prefer one adapter call. Header must state bounded wait for post-push PR-head convergence |
| `KARSIFT/karsift-ai-infra` tests including `tests/test_voc142_roster_wait_and_carrier.py` plus new VOC-144 cases | extend | Stale-then-converge; timeout-still-stale; durable mismatch without wait; API failure; preserved VOC-142 cases |
| Caller `tooling/governance/fixtures/karsift-ai-infra/**` | replace from new infra merge | Pin `PINNED_SHA.txt` to that exact merge; mirror every changed authoritative file |
| Caller `tooling/governance/tests/` | extend/reconcile | Advance every current-pin and mirrored-hash assertion while preserving historical authoritative-pin evidence |
| `AGENTS.md` | modify | Reconcile procedure must say carrier resolution waits boundedly for the existing same-ref OPEN PR to expose the exact pushed SHA |
| Fixture `README.md` | modify | Record the new pin and that post-push REST lag is not a durable mismatch |
| `docs/operations/15-ai-native-product-and-engineering-operating-model.md` | modify if current-state claims match | Only if exhaustive search finds a live false claim |
| `specs/changes/VOC-142-…/`, `VOC-141-…/`, `VOC-140-…/`, `VOC-139-…/` | **do not modify** | Audit evidence |
| `specs/changes/VOC-144-.../t00-evidence.md` | update | Record implementation PR base, new infra merge, named timeout/poll constants, wait change, validation after commit, feasible exact-head binding contract. Do not write the live implementation-head SHA into this file as a self-referential required value |

Ordered steps:

1. Resolve current `develop` to a 40-character SHA before any in-scope edit.
   Record that SHA as the implementation PR base at PR creation. Fail closed
   on unrelated/material movement.
2. Exhaustively search tracked source/docs for single-immediate-snapshot
   carrier-resolution claims, same-ref-SHA-difference-is-always-durable
   claims, and old pin/hash assertions; record each match as update,
   historical, or irrelevant.
3. Open the coordinated `KARSIFT/karsift-ai-infra` PR from current infra
   `main`. Implement D01–D07 and D17 there with tests that call the live
   adapter path using injected snapshots or a fake clock. Do not treat an
   untracked nested checkout as already-merged work. Do not weaken
   `resolve_roster_carrier` so a stale SHA on one snapshot becomes
   `reuse_open`.
4. Obtain independent exact-revision review of that infra PR and merge it.
   Record the exact merge SHA.
5. From current caller `develop`, create a new VOC-144 implementation branch.
   Set `PINNED_SHA.txt` and mirror every changed authoritative fixture file
   from that exact merge. Update the named current-state docs. Reconcile all
   current-pin and mirrored-hash assertions in the governance suite; preserve
   historical authoritative-pin constants and package records.
6. Confirm no VOC-142/VOC-141/VOC-140/VOC-139 package file is staged. Confirm
   `roles.yml` is untouched. Confirm no fabricated-status helper, bypass-actor
   addition, App-token mint change, or #1112 merge/close/recreate helper was
   added.
7. Track and commit the caller repair. Re-run suites against the committed
   tree. A pass obtained only while untracked is not acceptance.
8. Record evidence in `t00-evidence.md`, including source-search disposition,
   named timeout/poll constants, and the convergence-wait contract. This
   package's caller PR `Closes` only its own VOC-144 task issue.
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

- stale listed head then exact pushed head reuses the OPEN carrier and does
  not call `gh pr create`;
- timeout with still-stale head is `MISMATCHED_OPEN_CARRIER` and does not
  create;
- wrong-base / wrong-repo / ambiguous OPEN carriers fail without SHA-lag
  wait;
- API failure fails closed without create;
- exact first-snapshot match, already-merged reuse, and create-when-zero
  still hold;
- roster wait still requires the complete required set including `ci / ci`;
- merge-gate/release still reject in-progress parents as attestable
  completed `ci / ci`.

Do not treat a missing suite as a pass. Do not treat an untracked-only pass
as acceptance.

Independent verifier (exact reviewed caller SHA, and the infra PR SHA) should
confirm:

- bounded post-push SHA-lag wait in the GitHub adapter;
- `resolve_roster_carrier` still fail-closed on a single stale snapshot;
- stale-then-converge reuse and timeout-still-stale fail-closed;
- durable mismatch and API failure without wait;
- no duplicate PR and no reuse of a still-stale head;
- VOC-142 complete-required-set wait unchanged;
- wait still avoids `statusCheckRollup` / `gh pr checks`;
- merge-gate/release in-progress-parent rejection remains;
- tests exercise the live adapter path with injected snapshots or a fake
  clock, not wall-clock GitHub lag;
- `PINNED_SHA.txt` equals the new infra merge, not stale `8993e867…` if that
  merge still fails #1122's class;
- current-state source search is exhaustive; current docs state reconcile
  waits boundedly for the existing same-ref OPEN PR to expose the exact
  pushed SHA;
- VOC-142, VOC-141, VOC-140, and VOC-139 package records are unchanged;
- `roles.yml` is unchanged and no OpenAI route was added;
- `t00-evidence.md` names the implementation PR base, new infra merge, and
  named timeout/poll constants, states that the live head is bound by the
  App-authored independent-review comment/check, and does not require a
  commit to contain its own SHA;
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
  `reconcile` wait boundedly for GitHub to expose the just-pushed exact
  roster-PR head before treating a SHA difference as a durable mismatch,
  then reuse that carrier and wait for the complete required check set.
- **Rollback trigger:** post-push lag class recurs; a still-stale head is
  reused; a duplicate roster PR is created; in-progress parents become
  attestable; fabricated statuses; snapshot-gap commit; `roles.yml` /
  OpenAI route changed; two-token guard changed; #1112 manually merged or
  duplicated; VOC-142 wait-completeness regresses.
- **Rollback mechanism:** Revert the caller fixture/test/doc changes to the
  last reviewed `develop` merge and revert the coordinated infrastructure
  PR. That last-known-good still has the #1122 SHA-lag defect; rollback
  restores a known reviewed state, not a passing adoption recovery.
- **Last-known-good reference:** caller `develop` before this package's merge
  and pin `8993e867640dfb604dec0466c4e0787e68d8e258` unless a later reviewed
  pin superseded it before T00 started.
