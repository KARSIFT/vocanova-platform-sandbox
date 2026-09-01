# VOC-146 — Implementation Plan

## Preconditions and protected areas

- Do not begin until the package is adopted and implementation-authorized.
- Protected areas: `scripts/governance/`; named current-state docs including
  `AGENTS.md`; foundation tests added under `scripts/foundation/`.
- Prerequisites: confirm live
  `scripts/governance/validate-monitoring-impact.sh` still loads
  `$base...$head` via `mapfile < <(git diff …)` and that
  `validate-governance.sh` with the issue #1127 class (unresolved `--base`,
  resolvable `--head`) still prints `Governance structure validation
  passed.` with exit 0. Confirm `classify-change-risk.sh` still uses the
  same loader. Confirm VOC-086 `pull_request` missing-range fail-closed
  still exists. Confirm fixture `config/roles.yml` still binds
  implementer/escalation `cursor/composer-2.5` and
  planner/reviewer/retry/plan_reviewer
  `cursor/grok-4.6[effort=high,fast=false]`.
- Resolve current `develop` to a 40-character SHA **before any in-scope
  edit**. Record that SHA as the implementation PR base. Fail closed on
  unrelated/material movement of `develop`. This package's own
  plan/adoption/roster commits after `79b2b3f1…` do not count as
  protected-file drift. Do not edit `specs/changes/VOC-086-…/` or
  `specs/changes/VOC-112-…/`.
- No bootstrap exception. T00's first run is attempt `1` on a new VOC-146
  carrier from current `develop`. Do not treat the untracked local
  `karsift-ai-infra/` checkout as this repository's tracked tree.
- Preserve two-attempt implementer bounds, exact-SHA independent review,
  fail-closed credentials, and runner/App-token isolation.
- Do not print credential values. Do not rotate `KARSIFT_BOT_*` secrets or
  edit `config/roles.yml`.
- Do not snapshot the current develop/main gap. Do not add OpenAI execution.
- Do not weaken the production merge guard, add bypass actors, recapture
  VOC-112 fixtures, or open an infrastructure pin PR.

## File reconciliation and implementation sequence

### T00 — Fail closed on an unresolved or invalid `--base`/`--head` range

| Target | Action | Notes |
|--------|--------|-------|
| `scripts/governance/validate-monitoring-impact.sh` | modify | Resolve both revisions as commits; capture `git diff "$base...$head"` through a status-preserving path; fail closed on partial `--base`/`--head` |
| `scripts/governance/classify-change-risk.sh` | modify | Same fail-closed range contract as the monitoring-impact wrapper |
| `scripts/governance/validate-governance.sh` | modify only if needed | Nested nonzero exit already prevents the success line under `set -e`; do not add unrelated structure checks |
| Optional `scripts/governance/` helper sourced by both scripts | add only if it prevents drift | Stay inside `scripts/governance/`; do not expand into workflows unless a live false claim requires it |
| `scripts/foundation/voc146-*.test.mjs` (or live equivalent) | add | Invoke live scripts for nonexistent base/head, no-merge-base, partial range, valid range, `--files-from`, and classifier parity |
| `scripts/foundation/voc086-monitoring-impact.test.mjs` | do not rewrite VOC-086 records; extend only if a live assertion would otherwise be false | Preserve VOC-086-TEST-16 missing-range fail-closed |
| `AGENTS.md` | modify | Range fail-closed includes unresolved commits and invalid diff ranges, not only a missing range |
| Other current-state docs found by exhaustive search | modify if a live false claim is found | Do not rewrite historical VOC-086/VOC-112 package directories |
| `.github/workflows/*` | **do not modify by default** | Live pull_request paths already pass `--base`/`--head`; push-path classifier without range is existing behavior |
| `tooling/governance/fixtures/karsift-ai-infra/**`, `PINNED_SHA.txt` | **do not modify** | No pin advance |
| `specs/changes/VOC-086-…/`, `VOC-112-…/` | **do not modify** | Audit evidence |
| `specs/changes/VOC-146-.../t00-evidence.md` | update | Record implementation PR base, range-loading change, negative-case results, validation after commit, feasible exact-head binding contract. Do not write the live implementation-head SHA into this file as a self-referential required value |

Ordered steps:

1. Resolve current `develop` to a 40-character SHA before any in-scope edit.
   Record that SHA as the implementation PR base at PR creation. Fail closed
   on unrelated/material movement.
2. Exhaustively search tracked source/docs for missing-range-only fail-closed
   claims, `mapfile < <(git diff` range loaders, and "Governance structure
   validation passed" success contracts; record each match as update,
   historical, or irrelevant.
3. From current caller `develop`, create a new VOC-146 implementation branch.
   Replace the process-substitution loaders. Resolve both revisions as
   commits. Capture `git diff` status. Fail closed on partial range. Keep
   `--files-from`, `--declarations-only`, VOC-086 missing-range fail-closed,
   and working-tree fallback when no range was requested. A shared sourced
   helper under `scripts/governance/` is allowed if both scripts stay in
   lockstep; inlining the same contract in both files is also allowed.
4. Add deterministic tests that call the live scripts. Cover nonexistent
   `--base` (issue #1127 class), nonexistent `--head`, no-merge-base,
   partial range, valid range, `--files-from`, and classifier parity. Do not
   treat a grep-only assertion as covering #1127.
5. Update `AGENTS.md` and any other current-state docs found by the search.
   Do not rewrite VOC-086 or VOC-112 package records.
6. Confirm `roles.yml` is untouched. Confirm no VOC-112 fixture recapture,
   no pin advance, no snapshot-gap commit, and no workflow rewrite unless a
   live false claim required it.
7. Track and commit the repair. Re-run suites against the committed tree. A
   pass obtained only while untracked is not acceptance.
8. Record evidence in `t00-evidence.md`, including source-search
   disposition. This package's caller PR `Closes` only its own VOC-146 task
   issue.
9. After the exact reviewed caller PR merges, ordinary later promotion uses
   `release.yml`. Do not add a snapshot-gap task.

## Validation and independent verification

Deterministic commands, as applicable to the final changed file set, after
the repair is tracked and committed:

```bash
bash scripts/governance/validate-governance.sh --base <implementation-pr-base> --head <implementation-head>
bash scripts/governance/classify-change-risk.sh --base <implementation-pr-base> --head <implementation-head>
node --test scripts/foundation/voc146-*.test.mjs
node --test scripts/foundation/voc086-monitoring-impact.test.mjs
git diff --check
```

Also record the exact targeted commands that prove:

- nonexistent `--base` (issue #1127 class) exits nonzero and does not print
  `Governance structure validation passed.`;
- nonexistent `--head` exits nonzero;
- no-merge-base revisions exit nonzero;
- partial `--base` or `--head` exits nonzero;
- a valid implementation-PR range still succeeds;
- `--files-from` still succeeds;
- `classify-change-risk.sh` fails closed on the invalid-range class and
  still classifies a valid range (expect R4 for this change).

Do not treat a missing suite as a pass. Do not treat an untracked-only pass
as acceptance.

Independent verifier (exact reviewed caller SHA) should confirm:

- nonexistent base, nonexistent head, and no-merge-base fail closed;
- `mapfile < <(git diff "$base...$head")` is not the requested-range load
  path;
- Git nonzero status is preserved;
- partial range does not fall through to working-tree discovery;
- valid range, `--files-from`, `--declarations-only`, and VOC-086
  missing-range fail-closed remain;
- classifier parity;
- current-state source search is exhaustive; `AGENTS.md` states unresolved
  commits and invalid diff ranges are fail-closed;
- VOC-086 and VOC-112 package records are unchanged;
- `roles.yml` is unchanged and no OpenAI route was added;
- no pin advance, no VOC-112 recapture, and no snapshot-gap task;
- `t00-evidence.md` names the implementation PR base, states that the live
  head is bound by the App-authored independent-review comment/check, and
  does not require a commit to contain its own SHA;
- the independent-review comment binds this exact live PR head; merge-gate
  would reject a mismatch;
- the implementer did not approve or merge its own work.

## Deployment and rollback

- **Staging/production effect:** None intentional for application runtime on
  the ordinary path (tree-equivalent develop-ref update). Staging remains
  path-selected and must run only for a real tree change, not for
  tree-equivalent post-promotion sync.
- **Operational effect:** After this repair is live, governance validation
  and path-based risk classification fail closed on unresolved or invalid
  `--base`/`--head` ranges instead of accepting an empty file list.
- **Rollback trigger:** issue #1127 class recurs; classifier accepts an
  invalid range as empty; valid PR range or `--files-from` regresses;
  VOC-086 missing-range fail-closed regresses; snapshot-gap commit;
  `roles.yml` / OpenAI route changed.
- **Rollback mechanism:** Revert the caller script/test/doc changes to the
  last reviewed `develop` merge. That last-known-good still has the #1127
  fail-open range defect; rollback restores a known reviewed state, not a
  passing invalid-range gate.
- **Last-known-good reference:** caller `develop` before this package's merge
  (issue-creation reproduction commit
  `79b2b3f1f4224235bdda3f77ee887c3004978deb` unless a later reviewed merge
  superseded it before T00 started).
