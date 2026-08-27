# VOC-136 — Implementation Plan

## Preconditions and protected areas

- Do not begin until the package is adopted and implementation-authorized.
- Protected areas: caller `tooling/governance/` fixtures and tests; live
  `.github/workflows/*` only if implementation proves a caller dispatch file
  must change to consume #167; current-state fixture README / pin comments.
  The eight no-change paths are protected against change relative to
  protected comparison anchor `b9e74fc2db4691c48c637639b265d527de9f4505`.
- Prerequisites: confirm live `PINNED_SHA.txt` still equals
  `863fc1f35b1d35e4981a59166b0e939be1a2b681`. Confirm fixture `release.yml`
  still lacks `Restore shared lifecycle policy after caller checkout`. Confirm
  fixture `implement.yml` still lacks `Preserve post-implementer lifecycle
  helpers` and that `config/implementer_nested_checkout.py` and
  `tests/test_app_check_context.py` are still absent from the caller fixture.
  Confirm fixture `ci.yml` still lacks `fetch-depth: 0` and exact PR SHA
  pair passing. Confirm fixture `run-app-checks.sh` still lacks
  `--pr-base-sha` / `--pr-head-sha`. Confirm KARSIFT/karsift-ai-infra#167
  remains merged as `b263c0c110591cc798b89277dfc35542abb1597b`. Confirm
  VOC-129 caller PR #1046 remains merged as
  `429d8c6d49303148ca1cc14dba5f6768a7863346`. Confirm VOC-130-T00 (#1049)
  remains exhausted and PR #1051 remains unmerged. Confirm VOC-131-T00 /
  PR #1056 remains exhausted and unmerged. Confirm VOC-132-T00 (#1059) still
  has no published caller branch, PR, or completion marker. Confirm
  VOC-133-T00 (#1063 / PR #1065) remains exhausted and unmerged. Confirm
  VOC-134-T00 (#1068 / PR #1070) remains exhausted and unmerged. Confirm
  VOC-135-T00 (#1073 / PR #1075) remains exhausted at attempt-1 head
  `7c3e821bb3228c0fd2c7279b6d577a388148cb18` and attempt-2 head
  `dcd2da9ccbdf471c5294db30f09c9bb99aacdcd1`, closed unmerged, with no
  completion marker. Confirm the two named VOC-112 fixtures still contain
  `subject_revision` `f9d11e232a07c7d7a9c433d02c9267912543ba10`. Confirm live
  `voc112-navigation-benchmark.test.mjs` still fail-closes `local` mode when a
  full checkout is missing the captured commit. Confirm live `package.json`,
  `validate-workspace.mjs`, and `voc112-navigation-benchmark-run.mjs` have no
  `VOC112_CAPTURE_PROVENANCE_MODE` wrapper, capture-commit fetch helper, or
  hydrate helper. Confirm live `pipeline.yml` still dispatches
  `implement.yml@main` and `release.yml@main`, still exposes
  `reconcile-production-change`, and still has at most 25
  `workflow_dispatch` inputs. Confirm live `config/roles.yml` still binds
  implementer/escalation `cursor/composer-2.5` and planner/reviewer/retry/plan
  reviewer `cursor/grok-4.6[effort=high,fast=false]`.
- Resolve current `develop` to a 40-character SHA **before any in-scope
  edit**. Record that SHA as the implementation PR base. Fail closed on
  unrelated/material movement of `develop`. Retain protected comparison
  anchor `b9e74fc2db4691c48c637639b265d527de9f4505` for the eight no-change
  paths. This package's own plan/adoption/roster commits after that anchor
  do not count as protected-file drift. If any of the eight no-change paths
  differs from the anchor, or if the live pin/fixture already changed, fail
  closed. Do not proceed against a moving branch name as the eight-path
  comparison authority.
- No bootstrap exception. VOC-124 already published
  `permission-workflows: write` on `publish-source`. T00's first run is
  attempt `1` on a new VOC-136 carrier from current `develop`. Do not treat
  an untracked local `karsift-ai-infra/` checkout as this repository's
  tracked tree. Do not reuse PR #1051, PR #1056, PR #1065, PR #1070, or
  PR #1075. Do not redispatch VOC-132-T00 (#1059), VOC-133-T00 (#1063),
  VOC-134-T00 (#1068), or VOC-135-T00 (#1073).
- Preserve two-attempt implementer bounds, exact-SHA independent review,
  fail-closed credentials, and runner/App-token isolation.
- Do not print credential values. Do not change App installation permissions,
  rotate `KARSIFT_BOT_*` secrets, or edit `config/roles.yml`.
- Do not re-implement VOC-127 through VOC-135. Do not manufacture completion
  markers for those packages. Do not snapshot the current develop/main gap.
  Do not add operator SHA inputs. Do not add OpenAI execution. Do not
  retarget VOC-112 evidence. Do not edit `AGENTS.md`, the navigator skill, the
  VOC-112 provenance test, the navigation-benchmark runner,
  `validate-workspace.mjs`, or `package.json`. Do not add a capture-commit
  fetch helper, hydrate/materialize helper, provenance-mode wrapper, import
  side effect, environment override, or evidence-stamping helper. Do not
  require a commit to contain its own SHA. Do not exclude
  `tooling/governance/tests/**` from the complete-diff scan.

## File reconciliation and implementation sequence

### T00 — Pin exact infra #167 with immutable PR-context validation and exhaustive executable bypass scanning

| Target | Action | Notes |
|--------|--------|-------|
| `tooling/governance/fixtures/karsift-ai-infra/.github/workflows/ci.yml` | replace from #167 | Byte-identical to infra merge `b263c0c…`; drafting-time SHA-256 `0e0d485359d31325bf8b4c41b2047752ac42c6a5139251bd46b03cf7d671a9bb` must be reconfirmed |
| `tooling/governance/fixtures/karsift-ai-infra/.github/workflows/implement.yml` | replace from #167 | Drafting-time SHA-256 `e0612aa46dff58d3c06ff338864af3fa32cc725f151235cbe8b6789a80995d2a` must be reconfirmed |
| `tooling/governance/fixtures/karsift-ai-infra/.github/workflows/release.yml` | replace from #167 | Byte-identical to the #166 restore contract already in merge `b263c0c…`; drafting-time SHA-256 `fd11e45f999d26c9e009eb0d40c67c7a644ed2c8dd721a29b98c1fea4e790f08` must be reconfirmed. Still must be mirrored because the live fixture remains #164 |
| `tooling/governance/fixtures/karsift-ai-infra/config/run-app-checks.sh` | replace from #167 | Drafting-time SHA-256 `2ee4eaa25788af72146eee1ef9adb8cc9f42f2c4077d24e6aed25b55deddd1b2` must be reconfirmed |
| `tooling/governance/fixtures/karsift-ai-infra/config/implementer_nested_checkout.py` | create from #167 | New caller fixture file; drafting-time SHA-256 `e9190e7d5b1d48e76b0da63409005d27c12a36cbaf713033c3f2d9fa887216a9` must be reconfirmed |
| `tooling/governance/fixtures/karsift-ai-infra/tests/test_app_check_context.py` | create from #167 | New caller fixture file; drafting-time SHA-256 `74b8a0c1bcc00a137801e28888b3e6b78371934c14a99ea8eb34f4e0793bb5e0` must be reconfirmed |
| `tooling/governance/fixtures/karsift-ai-infra/tests/test_release_policy.py` | replace from #167 | Drafting-time SHA-256 `082c67fb26f221cf6e44e07364915f77bb4aee10b46e5b03be9c2d57c33a1e07` must be reconfirmed |
| `tooling/governance/fixtures/karsift-ai-infra/tests/test_voc121_implement_policy.py` | replace from #167 | Drafting-time SHA-256 `78bf3a05829ae76c9571ec5acc6099c49b405f9e5007c34001a50500e6044975` must be reconfirmed |
| `tooling/governance/fixtures/karsift-ai-infra/tests/test_voc123_source_bundle.py` | replace from #167 | Drafting-time SHA-256 `d0f28a862eb04e8cf5ff5ffa13f58749f95e26401c470d8e68f8f9b80f1b7936` must be reconfirmed |
| `tooling/governance/fixtures/karsift-ai-infra/CHANGELOG.md` | replace from #167 | Fixture copy of infrastructure changelog; drafting-time SHA-256 `7cdb3d6c863ccaab15012ef3944aac223d5a4fcc044c4f990955dfd02f70e4ea` must be reconfirmed |
| `tooling/governance/fixtures/karsift-ai-infra/PINNED_SHA.txt` | replace | Must equal `b263c0c110591cc798b89277dfc35542abb1597b`; must not equal `863fc1f…`, `8ce2b77…`, or `f3d791…` |
| `tooling/governance/fixtures/karsift-ai-infra/README.md` | modify current-state pin paragraph | Record VOC-136-T00 advance to #167 and the immutable PR-context contract; do not rewrite historical pin sentences as if they never happened; do not replace this caller-owned README with the infrastructure repository README. Infra README documentation-source SHA-256 `06bcea7696ff80a59c9bab846061ce7a4c6d8bc175a1d02a2d7eefab0bcf46a3` |
| `tooling/governance/tests/` | create/extend | Exact pin equality; inequality with #164, #165, and #166 pins; identify and converge restore-before-helper coverage; helper-copy-before-model and nested-checkout classifier coverage including non-directory; #167 PR-context coverage; recorded SHA-256 of mirrored #167 files; complete eight-path no-change boundary vs protected comparison anchor (fail, do not skip, if the commit or path cannot resolve); exhaustive caller-diff hydration/bypass scan **including this directory**; source-safe literals; required negative unit cases; benign-mention non-false-positives; tracked-after-commit pass; `package.json` identity; feasible exact-revision evidence contract; update VOC-129 and other live pin assertions that currently hard-code `863fc1f…` |
| `scripts/foundation/voc097-fixture-matrix.test.mjs`, `voc104-ready-for-review-reuse.test.mjs`, `voc108-authoritative-lifecycle.test.mjs` | modify | Advance pin literals from `863fc1f…` to `b263c0c…` |
| `scripts/foundation/fixtures/voc112-*.json` | **do not modify** | Must remain byte-identical to the protected comparison anchor |
| `scripts/foundation/voc112-navigation-benchmark.test.mjs` | **do not modify** | Must remain byte-identical to the protected comparison anchor; keep `local` fail-closed |
| `scripts/foundation/voc112-navigation-benchmark-run.mjs` | **do not modify** | Must remain byte-identical; no import-time fetch |
| `scripts/foundation/validate-workspace.mjs` | **do not modify** | Must remain byte-identical; no hydrate invocation |
| `AGENTS.md`, `.agents/skills/vocanova-repo-navigator/SKILL.md` | **do not modify** | Hashed by the VOC-112 fixtures that must stay unchanged |
| `package.json` | **do not modify** | Must remain byte-identical to the protected comparison anchor; no provenance-mode wrapper or capture-commit fetch helper |
| `.github/workflows/pipeline.yml` | modify only if `VOC-136-DEP-07` proves required | Expected no-op; caller already uses `implement.yml@main` and `release.yml@main` |
| `specs/changes/VOC-136-.../t00-evidence.md` | update | Record pin SHA, restore coverage, nested-checkout coverage, PR-context coverage, complete eight-path boundary vs protected comparison anchor, implementation PR base, exhaustive scan including `tooling/governance/tests/**`, negative cases, `package.json` identity, hashes, feasible exact-head binding contract, final infra merge, validation commands after commit, and VOC-127 through VOC-136 promotion handoff. Do not write the live implementation-head SHA into this file as a self-referential required value |

Ordered steps:

1. Resolve current `develop` to a 40-character SHA before any in-scope edit.
   Record that SHA as the implementation PR base at PR creation. Fail closed
   on unrelated/material movement. Retain protected comparison anchor
   `b9e74fc2db4691c48c637639b265d527de9f4505`.
2. From current `develop`, create a new VOC-136 implementation branch. Do not
   reuse PR #1051, PR #1056, PR #1065, PR #1070, or PR #1075. Do not
   redispatch VOC-132-T00 (#1059), VOC-133-T00 (#1063), VOC-134-T00 (#1068),
   or VOC-135-T00 (#1073). Do not re-check out or rewrite VOC-129 PR #1046.
3. Fetch KARSIFT/karsift-ai-infra at exact
   `b263c0c110591cc798b89277dfc35542abb1597b` into an untracked
   workspace-relative nested checkout. Copy the necessary changed
   authoritative files listed above into the caller fixture, byte-for-byte.
   Write `PINNED_SHA.txt` with that SHA only. Compute SHA-256 of the copied
   files; reconfirm against `VOC-136-D11` drafting-time hashes; commit the
   confirmed hashes into the caller regression. Remove the nested checkout
   before commit so it cannot stage as a gitlink. If exact comparison proves
   another existing caller-fixture file differs from #167, record that in
   `t00-evidence.md` and mirror it; do not copy unrelated infra-only tests.
   Do not leave a `/tmp` checkout as a test-time dependency.
4. Advance every live caller pin assertion that currently equals `863fc1f…`
   to `b263c0c…`. Add caller regressions that fixture `identify` and
   `converge` restore shared policy after caller checkout and before
   task-completion helpers; that `implement.yml` copies helpers before the
   unrestricted model and classifies nested checkout from the preserved
   helper; and that `ci.yml` / `run-app-checks.sh` / implementer pre-push
   apply the immutable PR-context contract without fetching evidence.
5. Add the complete eight-path no-change-boundary regression against the
   protected comparison anchor. Add the exhaustive caller-diff
   hydration/bypass scan that includes `tooling/governance/tests/**` and this
   regression's own module. Use non-contiguous source-safe literals.
   Reconstruct contiguous forbidden payloads only in memory for negative
   cases. Confirm none of the eight paths is staged. Confirm `local` mode
   fail-closed text remains in the provenance test. Confirm no capture-commit
   fetch helper, hydrate helper, or provenance-mode wrapper exists under any
   filename. Confirm `t00-evidence.md` names the protected comparison
   anchor, the implementation PR base, and the final infra merge, states the
   App-authored review/check binding contract, and does not require a commit
   to contain its own SHA.
6. Track and commit the regression. Re-run the complete-diff scan and the
   caller governance suite against the committed tree. A pass obtained only
   while the module was untracked is not acceptance.
7. Update current-state fixture README/comments. Record evidence in
   `t00-evidence.md`. This package's caller PR `Closes` only its own VOC-136
   task issue.
8. After the exact reviewed caller PR merges, ordinary release evaluation (or
   `reconcile-release`) completes promotion. Do not add a snapshot-gap task.
   Do not manufacture a VOC-127 through VOC-135 completion marker.

## Validation and independent verification

Deterministic commands, as applicable to the final changed file set, after
the regression is tracked and committed:

```bash
bash scripts/governance/validate-governance.sh --base <implementation-pr-base> --head <implementation-head>
bash scripts/governance/classify-change-risk.sh --base <implementation-pr-base> --head <implementation-head>
python3 -m unittest discover -s tooling/governance/tests -p 'test_*.py'
python3 -m unittest discover -s tooling/governance/fixtures/karsift-ai-infra/tests -p 'test_*.py'
node scripts/foundation/voc097-fixture-matrix.test.mjs
node scripts/foundation/voc104-ready-for-review-reuse.test.mjs
node scripts/foundation/voc108-authoritative-lifecycle.test.mjs
git diff --check
```

If implementation adds narrower targeted commands (for example a
`test_voc136_complete_no_change_boundary` module, a recorded-hash module, an
exhaustive-diff hydration/bypass scanner, negative-case helpers, or a
`package.json` identity module), record the exact commands in
`t00-evidence.md` and run them in addition to the suite above. Do not treat
a missing suite as a pass. Do not run a command that clones karsift-ai-infra
into `/tmp` as a required check. Do not treat an untracked-only pass as
acceptance.

Independent verifier (exact reviewed caller SHA) should confirm:

- `PINNED_SHA.txt` equals `b263c0c110591cc798b89277dfc35542abb1597b` and does
  not equal `863fc1f…`, `8ce2b77…`, or `f3d791…`;
- the mirrored authoritative files match the recorded SHA-256 hashes of infra
  merge `b263c0c…` without a `/tmp` checkout;
- fixture `release.yml` restores shared policy in both `identify` and
  `converge` after caller checkout and before task-completion helpers;
- restore uses `job.workflow_repository`, `job.workflow_sha`, and
  `persist-credentials: false`;
- fixture `implement.yml` copies helpers before the unrestricted model and
  classifies nested checkout from the preserved helper; absent continues
  caller publication; symlink / non-directory / parent-Git inheritance fail
  closed; a distinct nested checkout keeps exact-head bundle/lease
  publication;
- fixture `run-app-checks.sh` binds exact PR context without `git fetch`;
  unchanged capture fixture selects `pr-validation`; add/modify/delete select
  `pr-ancestry`; comparison errors fail closed;
- fixture `ci.yml` uses `fetch-depth: 0` and event base/head SHAs;
  implementer pre-push and post-self-correction use the integration anchor
  plus live committed HEAD;
- the #164 checkout-ref ordering / missing-`develop` path and
  `mergeCommit.oid` sync remain;
- live `pipeline.yml` still exposes `reconcile-production-change`, stays at
  most 25 inputs, and does not expose operator SHAs unless a proven #167
  caller-template delta requires a live edit;
- all eight no-change paths are absent from the diff against the protected
  comparison anchor; JSON fixtures still carry `subject_revision`
  `f9d11e232a07c7d7a9c433d02c9267912543ba10`; the provenance test still
  fail-closes `local` mode on a missing capture commit in a full checkout; no
  capture-commit fetch helper, hydrate helper, or provenance-mode wrapper
  exists under any filename, proven by an exhaustive caller-diff scan of
  changed executable paths including `tooling/governance/tests/**`;
- `SCAN_EXCLUDE_PREFIXES` or equivalent does not list
  `tooling/governance/tests/` or this regression's own module;
- negative unit cases reject the required bypass classes; benign mentions do
  not false-positive; the tracked committed regression is scan-clean;
- `t00-evidence.md` names the protected comparison anchor, the
  implementation PR base, and the final infra merge, states that the live
  head is bound by the App-authored independent-review comment/check, and
  does not require a commit to contain its own SHA or claim a protected-path
  revert that the tree still contains;
- the independent-review comment binds this exact live PR head; merge-gate
  would reject a mismatch;
- VOC-113 through VOC-135 isolation, recovery, App-token split, lease,
  retry limits, Cursor Composer/Grok roles, unchanged `roles.yml`, and
  non-closing source PR remain;
- PR #1051, PR #1056, PR #1065, PR #1070, and PR #1075 were not reused;
  VOC-132-T00 (#1059), VOC-133-T00 (#1063), VOC-134-T00 (#1068), and
  VOC-135-T00 (#1073) were not redispatched; no snapshot-gap task and no
  bootstrap exception were used; no VOC-127 through VOC-135 completion
  marker was manufactured; PR #1075 attempt-2 PASS WITH NON-BLOCKING
  FINDINGS was not treated as sufficient;
- the implementer did not approve or merge its own work.

## Deployment and rollback

- **Staging/production effect:** None intentional for application runtime on
  the ordinary path (tree-equivalent develop-ref update). Staging remains
  path-selected and must run only for a real tree change, not for
  tree-equivalent post-promotion sync.
- **Operational effect:** After this pin is live, `ci.yml@main` and
  `implement.yml@main` pre-push supply an immutable PR base/head pair so
  unchanged capture fixtures use `pr-validation` instead of default `local`
  provenance that induced VOC-134 hydration bypasses. `release.yml@main`
  restores nested shared policy after caller checkout. After each governed
  promotion merge, `develop` equals that merge SHA before the audit closes.
- **Rollback trigger:** pin not equal to `b263c0c…`; restore missing in
  `identify` or `converge`; helper copy missing before the unrestricted
  model; PR-context pair missing or evidence fetch introduced; unique develop
  commits erased; operator SHA paste accepted; `roles.yml` / OpenAI route
  changed; any of the eight no-change paths rewritten; capture-commit fetch
  helper, hydrate helper, or provenance-mode wrapper added under any
  filename; `tooling/governance/tests/` excluded from the complete-diff scan;
  evidence mutated at test time; self-referential exact-head SHA required;
  PR #1051, PR #1056, PR #1065, PR #1070, or PR #1075 reused; VOC-132-T00,
  VOC-133-T00, VOC-134-T00, or VOC-135-T00 redispatched; `/tmp`-only byte
  comparison; false-revert evidence; or a snapshot-gap commit.
- **Rollback mechanism:** Revert the caller fixture/test/doc changes to the
  last reviewed `develop` pin
  `863fc1f35b1d35e4981a59166b0e939be1a2b681`. That last-known-good still has
  the checkout-lifetime defect, the post-implementer helper-lifetime defect,
  and the missing PR-context contract in the fixture; rollback restores a
  known reviewed state, not a fully pinned #167 caller. Do not roll back by
  reverting infrastructure #167. Do not restore the VOC-112 JSON retargets
  from #1051, the provenance-test weakening from #1056, the `package.json` /
  evidence-stamp changes from #1065, the runner fetch / hydrate helpers from
  #1070, or the `SCAN_EXCLUDE_PREFIXES` weakening from #1075.
- **Last-known-good reference:** caller `develop` before this package's merge,
  with pin `863fc1f35b1d35e4981a59166b0e939be1a2b681`, VOC-112
  `subject_revision` `f9d11e232a07c7d7a9c433d02c9267912543ba10`, published
  provenance-test fail-closed `local` mode, and unmodified eight no-change
  paths relative to protected comparison anchor
  `b9e74fc2db4691c48c637639b265d527de9f4505`. Infrastructure last known good
  for the consumed contract is #167 merge
  `b263c0c110591cc798b89277dfc35542abb1597b`, which this package does not
  revert.
