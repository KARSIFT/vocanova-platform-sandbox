# VOC-139 — Implementation Plan

## Preconditions and protected areas

- Do not begin until the package is adopted and implementation-authorized.
- Protected areas: caller `tooling/governance/` fixtures, pin, and tests;
  reusable CI provenance export and required-check recovery metadata; live
  `.github/workflows/pipeline.yml`; the provenance test (in scope to change);
  named current-state docs. The seven remaining VOC-112 no-change paths are
  protected against change relative to
  `b9e74fc2db4691c48c637639b265d527de9f4505`.
- Prerequisites: confirm live `PINNED_SHA.txt` still equals
  `123735c80fec813a5b46a004f3e1122bd425cde2`. Confirm fixture
  `run-app-checks.sh` still selects `pr-validation` for `--promotion-pr`
  without exporting a promotion hash-anchor env. Confirm
  `assertPrValidationMergeBase` still requires stored hashes to equal
  merge-base files. Confirm live and template `promotion-pr-metadata` still
  invoke `gh pr view "$PROMOTION_PR_NUMBER"` with no `-R`/`--repo` and no
  checkout, and confirm the queried `headRepository.nameWithOwner` projection
  is absent from the live response. Confirm PR #1090, run `33130426061`, job `98718413924`, recovery
  run `33130527834`, job `98718739912`, and release run `33130473438` remain
  the incident record. Confirm `subject_revision` remains
  `f9d11e232a07c7d7a9c433d02c9267912543ba10`. Confirm the seven remaining
  VOC-112 paths still match `b9e74fc2…`. Confirm fixture `config/roles.yml`
  still binds implementer/escalation `cursor/composer-2.5` and
  planner/reviewer/retry/plan_reviewer
  `cursor/grok-4.6[effort=high,fast=false]`.
- Resolve current `develop` to a 40-character SHA **before any in-scope
  edit**. Record that SHA as the implementation PR base. Fail closed on
  unrelated/material movement of `develop`. This package's own
  plan/adoption/roster commits after `4812fb91…` do not count as
  protected-file drift. If any of the seven no-change paths differs from the
  anchor, fail closed.
- No bootstrap exception. T00's first run is attempt `1` on a new VOC-139
  carrier from current `develop`. Do not reuse PR #1090 as this package's
  implementation PR. Do not treat the untracked local `karsift-ai-infra/`
  checkout as this repository's tracked tree.
- Preserve two-attempt implementer bounds, exact-SHA independent review,
  fail-closed credentials, and runner/App-token isolation.
- Do not print credential values. Do not change App installation permissions,
  rotate `KARSIFT_BOT_*` secrets, or edit `config/roles.yml`.
- Do not snapshot the current develop/main gap. Do not add OpenAI execution.
- Do not fetch, hydrate, or recapture VOC-112 JSON fixtures. Do not edit the
  seven remaining VOC-112 no-change paths.

## File reconciliation and implementation sequence

### T00 — Unblock accumulated promotion hash validation and no-checkout recovery metadata

| Target | Action | Notes |
|--------|--------|-------|
| `KARSIFT/karsift-ai-infra` `config/run-app-checks.sh` | modify | When `--promotion-pr` is set, export `VOC112_PROMOTION_PR=true` (or equivalent) and keep mode `pr-validation` with exact PR SHAs. Unset the env otherwise. Do not fetch |
| `KARSIFT/karsift-ai-infra` `templates/project-repo/.github/workflows/pipeline.yml` | modify | `promotion-pr-metadata` must address `$GITHUB_REPOSITORY` explicitly and validate supported owner/repository fields, refs, open state, and exact SHAs. No checkout step |
| `KARSIFT/karsift-ai-infra` tests | extend | Promotion hash-anchor export; no-checkout metadata subprocess using the live response shape without `headRepository.nameWithOwner`; fork/identity negatives; ordinary merge-base retention |
| Caller `scripts/foundation/voc112-navigation-benchmark.test.mjs` | modify | Promotion signal → hashes bind to `PR_HEAD_SHA` + working tree; absent signal → merge-base anchoring remains |
| Caller `.github/workflows/pipeline.yml` | modify | Same repository-explicit, supported-field metadata lookup as the template |
| Caller `tooling/governance/fixtures/karsift-ai-infra/**` | replace from new infra merge | Pin `PINNED_SHA.txt` to that exact merge; mirror every changed authoritative file |
| Caller `tooling/governance/tests/` | extend | Accumulated-hash, no-checkout metadata, identity negatives; narrow VOC-138 `NO_CHANGE_PATHS` so the provenance test is no longer frozen |
| Seven remaining VOC-112 no-change paths | **do not modify** | Must remain byte-identical to `b9e74fc2…` |
| `docs/operations/11-devops-and-ci-cd.md` | modify | Replace promotion-PR merge-base hash claim with head/source-revision binding |
| `docs/development/agent-skills.md` | modify | Same hash-anchor contract; keep ordinary merge-base / `pr-ancestry` language |
| Fixture `README.md` | modify | Record the new pin and the promotion hash/metadata contract |
| `specs/changes/VOC-138-…/` | **do not modify** | Audit evidence |
| `specs/changes/VOC-139-.../t00-evidence.md` | update | Record implementation PR base, new infra merge, hash-rule and metadata change, validation after commit, feasible exact-head binding contract. Do not write the live implementation-head SHA into this file as a self-referential required value |

Ordered steps:

1. Resolve current `develop` to a 40-character SHA before any in-scope edit.
   Record that SHA as the implementation PR base at PR creation. Fail closed
   on unrelated/material movement.
2. Open the coordinated `KARSIFT/karsift-ai-infra` PR from current infra
   `main`. Implement D02 and D07 there with tests. Do not treat an untracked
   nested checkout as already-merged work.
3. Obtain independent exact-revision review of that infra PR and merge it.
   Record the exact merge SHA.
4. From current caller `develop`, create a new VOC-139 implementation branch.
   Implement the provenance-test hash rule and live `pipeline.yml` metadata
   fix. Set `PINNED_SHA.txt` and mirror every changed authoritative fixture
   file from that exact merge. Update the named current-state docs. Narrow
   live VOC-138 `NO_CHANGE_PATHS`.
5. Confirm no seven-path VOC-112 file and no VOC-138 package file is staged.
   Confirm `roles.yml` is untouched. Confirm no fetch/hydrate/recapture
   helper was added.
6. Track and commit the caller repair. Re-run suites against the committed
   tree. A pass obtained only while untracked is not acceptance.
7. Record evidence in `t00-evidence.md`. This package's caller PR `Closes`
   only its own VOC-139 task issue.
8. After the exact reviewed caller PR merges, `reconcile-release` for #1089
   completes promotion of #1090 (or the live promotion at the then-current
   `develop` head). Do not add a snapshot-gap task.

## Validation and independent verification

Deterministic commands, as applicable to the final changed file set, after
the repair is tracked and committed:

```bash
bash scripts/governance/validate-governance.sh --base <implementation-pr-base> --head <implementation-head>
bash scripts/governance/classify-change-risk.sh --base <implementation-pr-base> --head <implementation-head>
python3 -m unittest discover -s tooling/governance/tests -p 'test_*.py'
python3 -m unittest discover -s tooling/governance/fixtures/karsift-ai-infra/tests -p 'test_*.py'
node --test scripts/foundation/voc112-navigation-benchmark.test.mjs
git diff --check
```

Also record the exact targeted infra and caller commands that prove:

- accumulated promotion with differing `AGENTS.md` → `VOC-112-TEST-12` /
  `VOC-112-TEST-13` pass under `pr-validation`;
- ordinary no-promotion differing hashes → merge-base fail-closed;
- malformed SHA / base-not-ancestor / wrong refs / unrelated repository/PR,
  including a same-name fork and closed PR, fail closed;
- metadata step succeeds in a non-git directory with explicit repository
  context and supported response fields, and fails without either;
- recovery still rejects weaker same-head dispatch and does not rerun doomed
  `pull_request` jobs.

Do not treat a missing suite as a pass. Do not treat an untracked-only pass
as acceptance.

Independent verifier (exact reviewed caller SHA, and the infra PR SHA) should
confirm:

- authenticated promotion PRs keep `pr-validation` and bind hashes to
  `PR_HEAD_SHA` / working tree, not historical `main`;
- `VOC-112-TEST-12` and `VOC-112-TEST-13` pass for the #1090 accumulated class;
- ordinary `pr-validation` remains merge-base anchored;
- ordinary fixture-changing PRs remain `pr-ancestry` fail-closed;
- hash/SHA/identity negatives still fail closed;
- no fetch/hydrate/recapture helper exists;
- `promotion-pr-metadata` is repository-explicit, validates supported
  owner/repository fields, and has no checkout;
- a no-checkout subprocess test exists and would have caught job
  `98718739912`;
- `PINNED_SHA.txt` equals the new infra merge, not stale `123735c80…` if that
  merge still fails #1096's class;
- all seven remaining VOC-112 no-change paths are absent from the diff
  against `b9e74fc2…`;
- current-state docs no longer claim accumulated promotions are merge-base
  hash-bound;
- VOC-138 package records are unchanged;
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
- **Operational effect:** After this repair is live, a same-repository
  `main` ← `develop` promotion whose hashed sources changed on `develop` can
  pass required `ci / ci` under head-bound `pr-validation`, and recovery
  metadata can resolve without a checkout. Ordinary PRs still fail closed
  when merge-base hashes do not match. Ordinary release can then merge #1090
  and trigger automatic deployment.
- **Rollback trigger:** accumulated promotion still requires merge-base
  hashes; ordinary merge-base `pr-validation` no longer fails closed;
  promotion PR switched to `--squash-safe-push`; fetch/hydrate/recapture
  helper added; seven VOC-112 paths rewritten; metadata still lacks `-R`;
  evidence mutated at test time; self-referential exact-head SHA required;
  snapshot-gap commit; `roles.yml` / OpenAI route changed.
- **Rollback mechanism:** Revert the caller fixture/test/doc/workflow/
  provenance-test changes to the last reviewed `develop` merge and revert
  the coordinated infrastructure PR. That last-known-good still has the
  #1090 deadlock; rollback restores a known reviewed state, not a passing
  promotion PR. Do not roll back by recapturing VOC-112 fixtures.
- **Last-known-good reference:** caller `develop` before this package's merge
  (issue-creation `4812fb91ab1b674f9a9ec03906f90c0edf50421d`), pin
  `123735c80fec813a5b46a004f3e1122bd425cde2`, VOC-112 `subject_revision`
  `f9d11e232a07c7d7a9c433d02c9267912543ba10`, unmodified seven no-change
  paths relative to `b9e74fc2db4691c48c637639b265d527de9f4505`, and `main`
  `0d0b0cdf0692d0349f380e9cae3285b4c7916b05`.
