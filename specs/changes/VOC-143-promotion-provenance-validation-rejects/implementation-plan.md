# VOC-143 — Implementation Plan

## Preconditions and protected areas

- Do not begin until the package is adopted and implementation-authorized.
- Protected areas: VOC-112 capture provenance used by required `validate` and
  `ci / ci` before develop-to-main promotion.
- Prerequisites: confirm live
  `scripts/foundation/voc112-navigation-benchmark.test.mjs` still requires
  working-tree `AGENTS.md` equality for every mode except `pr-validation`.
  Confirm `repository-governance.yml` still selects `squash-safe-push` for
  same-repository `main` ← `develop`. Confirm reusable `ci.yml` still uses
  `--promotion-pr` for that pair. Confirm `PINNED_SHA.txt` still equals
  `8993e867640dfb604dec0466c4e0787e68d8e258`. Confirm PR #1119, head
  `376e00dd769253d7a255660f5391fb208781e2f3`, and audit #1118 remain the
  incident record. Confirm fixture `config/roles.yml` still binds
  implementer/escalation `cursor/composer-2.5` and planner/reviewer/retry/
  plan_reviewer `cursor/grok-4.6[effort=high,fast=false]`.
- Resolve current `develop` to a 40-character SHA **before any in-scope
  edit**. Record that SHA as the implementation PR base. Fail closed on
  unrelated/material movement of `develop`. This package's own
  plan/adoption/roster commits after `376e00dd…` do not count as
  protected-file drift. Do not edit `specs/changes/VOC-142-…/` or earlier
  package directories.
- T00's first run is attempt `1` on a new VOC-143 carrier from current
  `develop`. Do not reuse PR #1119, audit #1118, or VOC-142 artifacts as this
  package's implementation artifacts. Do not treat the untracked local
  `karsift-ai-infra/` checkout as this repository's tracked tree.
- Preserve two-attempt implementer bounds, exact-SHA independent review,
  fail-closed credentials, and runner/App-token isolation.
- Do not print credential values. Do not rotate `KARSIFT_BOT_*` secrets or
  edit `config/roles.yml`.
- Do not snapshot the current develop/main gap. Do not add OpenAI execution.
- Do not weaken the production merge guard, add bypass actors, fabricate
  statuses, recapture VOC-112 fixtures, or manually merge a promotion PR.

## File reconciliation and implementation sequence

### T00 — Bind promotion-path VOC-112 `AGENTS.md` provenance to an immutable historical ancestor

| Target | Action | Notes |
|--------|--------|-------|
| `scripts/foundation/voc112-navigation-benchmark.test.mjs` | modify | `squash-safe-push` ancestor-bind for `agents_sha256`; promotion `pr-validation` ancestor-bind for `agents_sha256`; keep other modes and navigator HEAD-binding |
| `docs/operations/11-devops-and-ci-cd.md` | modify if current-state claims match | Record that promotion `validate` uses `squash-safe-push` with historical-ancestor `AGENTS.md` bind; do not claim promotion `ci` uses `squash-safe-push` |
| `scripts/foundation/voc114-actions-check-recovery.test.mjs` | **do not modify mode selection** | Promotion `validate` must still select `squash-safe-push` |
| `scripts/foundation/fixtures/voc112-*.json` | **do not modify** | Historical capture; do not recapture |
| `tooling/governance/fixtures/karsift-ai-infra/PINNED_SHA.txt` | **do not modify** | Live pin `8993e867…` is not this defect |
| `.github/workflows/repository-governance.yml` | **prefer do not modify** | SHAs already exported; editing this file is an R4 path floor |
| `specs/changes/VOC-142-…/` through `VOC-112-…/` | **do not modify** | Audit evidence |
| `specs/changes/VOC-143-.../t00-evidence.md` | update | Record implementation PR base, assertion change, validation after commit, feasible exact-head binding contract. Do not write the live implementation-head SHA into this file as a self-referential required value |

Ordered steps:

1. Resolve current `develop` to a 40-character SHA before any in-scope edit.
   Record that SHA as the implementation PR base at PR creation. Fail closed
   on unrelated/material movement.
2. Exhaustively search tracked source/docs for working-tree-equality claims,
   promotion HEAD-equality claims, and mode-selection claims; record each
   match as update, historical, or irrelevant.
3. From current caller `develop`, create a new VOC-143 implementation branch.
   Implement D01–D04 in `voc112-navigation-benchmark.test.mjs` with tests
   that call `assertCapturedRevision`. Update VOC-139 expected error strings
   only if the navigator assertion fires first after D03.
4. Update named current-state docs. Do not recapture fixtures. Do not change
   the pin. Do not switch promotion check identity.
5. Confirm no VOC-142/VOC-139/VOC-114 package file is staged. Confirm
   `roles.yml` is untouched. Confirm no fetch/hydrate helper, bypass-actor
   addition, App-token mint change, or #1119 merge/close/recreate helper was
   added.
6. Track and commit the caller repair. Re-run suites against the committed
   tree. A pass obtained only while untracked is not acceptance.
7. Record evidence in `t00-evidence.md`, including source-search disposition
   and the ancestor-bind contract. This package's caller PR `Closes` only its
   own VOC-143 task issue.
8. After the exact reviewed caller PR merges, ordinary `reconcile-release`
   for #1118 may re-evaluate #1119 when that PR still matches. Do not add a
   snapshot-gap task.

## Validation and independent verification

Deterministic commands, as applicable to the final changed file set, after
the repair is tracked and committed:

```bash
bash scripts/governance/validate-governance.sh --base <implementation-pr-base> --head <implementation-head>
bash scripts/governance/classify-change-risk.sh --base <implementation-pr-base> --head <implementation-head>
VOC112_CAPTURE_PROVENANCE_MODE=squash-safe-push node --test scripts/foundation/voc112-navigation-benchmark.test.mjs
node --test scripts/foundation/voc112-navigation-benchmark.test.mjs
node --test scripts/foundation/voc114-actions-check-recovery.test.mjs
git diff --check
```

Also record the exact targeted commands that prove:

- `squash-safe-push` with current working-tree `AGENTS.md` different from the
  historical fixture hash succeeds;
- `squash-safe-push` with a tampered/unfound `agents_sha256` fails closed;
- `local` still requires working-tree equality;
- ordinary `pr-validation` still requires merge-base hashes;
- promotion `pr-validation` accepts historical fixture `agents_sha256` when
  HEAD `AGENTS.md` differs and navigator hashes remain HEAD-bound;
- VOC-114 still requires promotion `validate` to select `squash-safe-push`.

Do not treat a missing suite as a pass. Do not treat an untracked-only pass
as acceptance.

Independent verifier (exact reviewed caller SHA) should confirm:

- `squash-safe-push` historical-ancestor `AGENTS.md` bind;
- promotion `pr-validation` historical-ancestor `AGENTS.md` bind;
- retained `local` / `pr-ancestry` / ordinary `pr-validation`;
- retained navigator HEAD/working-tree binding;
- tests exercise `assertCapturedRevision`, not comment-only skips;
- VOC-112 JSON fixtures are byte-identical to the implementation PR base;
- `PINNED_SHA.txt` is unchanged;
- promotion check identity is unchanged;
- current-state source search is exhaustive;
- VOC-142 through VOC-112 package records are unchanged;
- `roles.yml` is unchanged and no OpenAI route was added;
- `t00-evidence.md` names the implementation PR base, states that the live
  head is bound by the App-authored independent-review comment/check, and
  does not require a commit to contain its own SHA;
- the independent-review comment binds this exact live PR head; merge-gate
  would reject a mismatch;
- no snapshot-gap task and no manual merge of #1119 were used;
- the implementer did not approve or merge its own work.

## Deployment and rollback

- **Staging/production effect:** None intentional for application runtime on
  the ordinary path (tree-equivalent develop-ref update). Staging remains
  path-selected and must run only for a real tree change, not for
  tree-equivalent post-promotion sync.
- **Operational effect:** After this repair is live, promotion-path
  `validate` and `ci / ci` accept a legitimate `AGENTS.md` documentation
  update with an unmodified historical VOC-112 fixture.
- **Rollback trigger:** working-tree-equality class recurs under
  `squash-safe-push`; promotion `pr-validation` again requires HEAD
  `AGENTS.md` equality; tampered fixture hashes pass; `local` /
  `pr-ancestry` / ordinary `pr-validation` weakened; navigator HEAD-binding
  dropped; promotion check identity switched; fixtures recaptured;
  snapshot-gap commit; `roles.yml` / OpenAI route changed; #1119 manually
  merged or duplicated.
- **Rollback mechanism:** Revert the caller test/doc changes to the last
  reviewed `develop` merge. That last-known-good still has the #1120
  provenance defect; rollback restores a known reviewed state, not a
  passing promotion.
- **Last-known-good reference:** caller `develop` before this package's merge
  (issue-creation promotion head `376e00dd769253d7a255660f5391fb208781e2f3`)
  and pin `8993e867640dfb604dec0466c4e0787e68d8e258`.
