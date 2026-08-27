# VOC-133 — Implementation Plan

## Preconditions and protected areas

- Do not begin until the package is adopted and implementation-authorized.
- Protected areas: caller `tooling/governance/` fixtures and tests; live
  `.github/workflows/*` only if implementation proves a caller dispatch file
  must change to consume #166; current-state fixture README / pin comments.
  The five VOC-112 no-change paths are protected against change.
- Prerequisites: confirm live `PINNED_SHA.txt` still equals
  `863fc1f35b1d35e4981a59166b0e939be1a2b681`. Confirm fixture `release.yml`
  still lacks `Restore shared lifecycle policy after caller checkout`. Confirm
  fixture `implement.yml` still lacks `Preserve post-implementer lifecycle
  helpers` and that `config/implementer_nested_checkout.py` is still absent
  from the caller fixture. Confirm KARSIFT/karsift-ai-infra#166 remains merged
  as `f3d79177bf8a9abe0dae550f39502165d494c576`. Confirm VOC-129 caller PR
  #1046 remains merged as `429d8c6d49303148ca1cc14dba5f6768a7863346`. Confirm
  VOC-130-T00 (#1049) remains exhausted and PR #1051 remains unmerged. Confirm
  VOC-131-T00 / PR #1056 remains exhausted at
  `c11454e717a6d778143de1f2023acc4480305845` and unmerged. Confirm VOC-132-T00
  (#1059) still has no published caller branch, PR, or completion marker from
  run `33079499176`. Confirm the two named VOC-112 fixtures still contain
  `subject_revision` `f9d11e232a07c7d7a9c433d02c9267912543ba10`. Confirm live
  `voc112-navigation-benchmark.test.mjs` still fail-closes `local` mode when a
  full checkout is missing the captured commit. Confirm live `pipeline.yml`
  still dispatches `implement.yml@main` and `release.yml@main`, still exposes
  `reconcile-production-change`, and still has at most 25 `workflow_dispatch`
  inputs.
- No bootstrap exception. VOC-124 already published
  `permission-workflows: write` on `publish-source`. T00's first run is
  attempt `1` on a new VOC-133 carrier from current `develop`. Do not treat
  an untracked local `karsift-ai-infra/` checkout as this repository's
  tracked tree. Do not reuse PR #1051 or PR #1056. Do not redispatch
  VOC-132-T00 (#1059).
- Preserve two-attempt implementer bounds, exact-SHA independent review,
  fail-closed credentials, and runner/App-token isolation.
- Do not print credential values. Do not change App installation permissions,
  rotate `KARSIFT_BOT_*` secrets, or edit `config/roles.yml`.
- Do not re-implement VOC-129, VOC-130, VOC-131, or VOC-132. Do not
  manufacture completion markers for those packages. Do not snapshot the
  current develop/main gap. Do not add operator SHA inputs. Do not add OpenAI
  execution. Do not retarget VOC-112 evidence. Do not edit `AGENTS.md`, the
  navigator skill, or the VOC-112 provenance test.

## File reconciliation and implementation sequence

### T00 — Pin exact infra #166 with a complete VOC-112 no-change boundary

| Target | Action | Notes |
|--------|--------|-------|
| `tooling/governance/fixtures/karsift-ai-infra/.github/workflows/implement.yml` | replace from #166 | Byte-identical to infra merge `f3d79177bf8a9abe0dae550f39502165d494c576`; drafting-time SHA-256 `5e44f6a82cdb127f9716faea56cd226965ab3cf86566bde009af375c205ff03c` must be reconfirmed |
| `tooling/governance/fixtures/karsift-ai-infra/.github/workflows/release.yml` | replace from #166 | Byte-identical to the same merge; drafting-time SHA-256 `fd11e45f999d26c9e009eb0d40c67c7a644ed2c8dd721a29b98c1fea4e790f08` must be reconfirmed. These bytes match the #165 restore contract and still must be mirrored because the live fixture remains #164 |
| `tooling/governance/fixtures/karsift-ai-infra/config/implementer_nested_checkout.py` | create from #166 | New caller fixture file; drafting-time SHA-256 `e9190e7d5b1d48e76b0da63409005d27c12a36cbaf713033c3f2d9fa887216a9` must be reconfirmed |
| `tooling/governance/fixtures/karsift-ai-infra/tests/test_release_policy.py` | replace from #166 | Drafting-time SHA-256 `082c67fb26f221cf6e44e07364915f77bb4aee10b46e5b03be9c2d57c33a1e07` must be reconfirmed |
| `tooling/governance/fixtures/karsift-ai-infra/tests/test_voc121_implement_policy.py` | replace from #166 | Drafting-time SHA-256 `78bf3a05829ae76c9571ec5acc6099c49b405f9e5007c34001a50500e6044975` must be reconfirmed |
| `tooling/governance/fixtures/karsift-ai-infra/tests/test_voc123_source_bundle.py` | replace from #166 | Drafting-time SHA-256 `d0f28a862eb04e8cf5ff5ffa13f58749f95e26401c470d8e68f8f9b80f1b7936` must be reconfirmed |
| `tooling/governance/fixtures/karsift-ai-infra/CHANGELOG.md` | replace from #166 | Fixture copy of infrastructure changelog; drafting-time SHA-256 `7cdb3d6c863ccaab15012ef3944aac223d5a4fcc044c4f990955dfd02f70e4ea` must be reconfirmed |
| `tooling/governance/fixtures/karsift-ai-infra/PINNED_SHA.txt` | replace | Must equal `f3d79177bf8a9abe0dae550f39502165d494c576`; must not equal `863fc1f…` or `8ce2b77…` |
| `tooling/governance/fixtures/karsift-ai-infra/README.md` | modify current-state pin paragraph | Record VOC-133-T00 advance to #166; do not rewrite historical pin sentences as if they never happened; do not replace this caller-owned README with the infrastructure repository README |
| `tooling/governance/tests/` | create/extend | Exact pin equality; inequality with #164 and #165 pins; identify and converge restore-before-helper coverage; helper-copy-before-model and nested-checkout classifier coverage including non-directory; recorded SHA-256 of mirrored #166 files; complete five-path VOC-112 no-change boundary vs develop base; update VOC-129 and other live pin assertions that currently hard-code `863fc1f…` |
| `scripts/foundation/voc097-fixture-matrix.test.mjs`, `voc104-ready-for-review-reuse.test.mjs`, `voc108-authoritative-lifecycle.test.mjs` | modify | Advance pin literals from `863fc1f…` to `f3d791…` |
| `scripts/foundation/fixtures/voc112-*.json` | **do not modify** | Must remain byte-identical to the carrier `develop` base |
| `scripts/foundation/voc112-navigation-benchmark.test.mjs` | **do not modify** | Must remain byte-identical to the carrier `develop` base; keep `local` fail-closed |
| `AGENTS.md`, `.agents/skills/vocanova-repo-navigator/SKILL.md` | **do not modify** | Hashed by the VOC-112 fixtures that must stay unchanged |
| `.github/workflows/pipeline.yml` | modify only if `VOC-133-DEP-07` proves required | Expected no-op; caller already uses `implement.yml@main` and `release.yml@main` |
| `specs/changes/VOC-133-.../t00-evidence.md` | update | Record pin SHA, restore coverage, nested-checkout coverage, complete VOC-112 boundary, hashes, exact reviewed head, final infra merge, validation commands, and VOC-129 / VOC-130 / VOC-131 / VOC-132 / VOC-133 promotion handoff |

Ordered steps:

1. From current `develop`, create a new VOC-133 implementation branch. Do not
   reuse PR #1051 or PR #1056. Do not redispatch VOC-132-T00 (#1059). Do not
   re-check out or rewrite VOC-129 PR #1046.
2. Fetch KARSIFT/karsift-ai-infra at exact
   `f3d79177bf8a9abe0dae550f39502165d494c576` into an untracked
   workspace-relative nested checkout. Copy the necessary changed
   authoritative files listed above into the caller fixture, byte-for-byte.
   Write `PINNED_SHA.txt` with that SHA only. Compute SHA-256 of the copied
   files; reconfirm against `VOC-133-D11` drafting-time hashes; commit the
   confirmed hashes into the caller regression. Remove the nested checkout
   before commit so it cannot stage as a gitlink. If exact comparison proves
   another existing caller-fixture file differs from #166, record that in
   `t00-evidence.md` and mirror it; do not copy unrelated infra-only tests.
   Do not leave a `/tmp` checkout as a test-time dependency.
3. Advance every live caller pin assertion that currently equals `863fc1f…`
   to `f3d791…`. Add caller regressions that fixture `identify` and
   `converge` restore shared policy after caller checkout and before
   task-completion helpers, and that `implement.yml` copies helpers before
   the unrestricted model and classifies nested checkout from the preserved
   helper.
4. Add the complete VOC-112 no-change-boundary regression against the carrier
   `develop` base for all five named paths. Confirm none of those paths is
   staged. Confirm `local` mode fail-closed text remains in the provenance
   test. Confirm `t00-evidence.md` names the exact head and the final infra
   merge and does not claim a revert for any path that appears in the diff.
5. Update current-state fixture README/comments. Record evidence in
   `t00-evidence.md`. This package's caller PR `Closes` only its own VOC-133
   task issue.
6. After the exact reviewed caller PR merges, ordinary release evaluation (or
   `reconcile-release`) completes VOC-129's skipped promotion and this
   package's promotion. Do not add a snapshot-gap task. Do not manufacture a
   VOC-129, VOC-130, VOC-131, or VOC-132 completion marker.

## Validation and independent verification

Deterministic commands, as applicable to the final changed file set:

```bash
bash scripts/governance/validate-governance.sh
bash scripts/governance/classify-change-risk.sh
python3 -m unittest discover -s tooling/governance/tests -p 'test_*.py'
python3 -m unittest discover -s tooling/governance/fixtures/karsift-ai-infra/tests -p 'test_*.py'
node scripts/foundation/voc097-fixture-matrix.test.mjs
node scripts/foundation/voc104-ready-for-review-reuse.test.mjs
node scripts/foundation/voc108-authoritative-lifecycle.test.mjs
git diff --check
```

If implementation adds narrower targeted commands (for example
`python3 -m unittest tooling.governance.tests.test_voc133_complete_voc112_boundary`
or a recorded-hash module), record the exact commands in `t00-evidence.md`
and run them in addition to the suite above. Do not treat a missing suite as
a pass. Do not run a command that clones karsift-ai-infra into `/tmp` as a
required check.

Independent verifier (exact reviewed caller SHA) should confirm:

- `PINNED_SHA.txt` equals `f3d79177bf8a9abe0dae550f39502165d494c576` and does
  not equal `863fc1f…` or `8ce2b77…`;
- the mirrored authoritative files match the recorded SHA-256 hashes of infra
  merge `f3d791…` without a `/tmp` checkout;
- fixture `release.yml` restores shared policy in both `identify` and
  `converge` after caller checkout and before task-completion helpers;
- restore uses `job.workflow_repository`, `job.workflow_sha`, and
  `persist-credentials: false`;
- fixture `implement.yml` copies helpers before the unrestricted model and
  classifies nested checkout from the preserved helper; absent continues
  caller publication; symlink / non-directory / parent-Git inheritance fail
  closed; a distinct nested checkout keeps exact-head bundle/lease
  publication;
- the #164 checkout-ref ordering / missing-`develop` path and
  `mergeCommit.oid` sync remain;
- live `pipeline.yml` still exposes `reconcile-production-change`, stays at
  most 25 inputs, and does not expose operator SHAs unless a proven #166
  caller-template delta requires a live edit;
- all five VOC-112 no-change paths are absent from the diff; JSON fixtures
  still carry `subject_revision` `f9d11e232a07c7d7a9c433d02c9267912543ba10`;
  the provenance test still fail-closes `local` mode on a missing capture
  commit in a full checkout;
- `t00-evidence.md` names this exact SHA and the final infra merge and does
  not claim a protected-path revert that the tree still contains;
- VOC-113 through VOC-132 isolation, recovery, App-token split, lease,
  retry limits, Cursor Composer/Grok roles, unchanged `roles.yml`, and
  non-closing source PR remain;
- PR #1051 and PR #1056 were not reused; VOC-132-T00 (#1059) was not
  redispatched; no snapshot-gap task and no bootstrap exception were used; no
  VOC-129 / VOC-130 / VOC-131 / VOC-132 completion marker was manufactured;
- the implementer did not approve or merge its own work.

## Deployment and rollback

- **Staging/production effect:** None intentional for application runtime on
  the ordinary path (tree-equivalent develop-ref update). Staging remains
  path-selected.
- **Operational effect:** After this pin is live, `release.yml@main` restores
  nested shared policy after caller checkout, and `implement.yml@main`
  preserves post-model helpers before the unrestricted model, so
  task-completion validation and implementer publication can run instead of
  no-op'ing or failing on a deleted nested checkout. After each governed
  promotion merge, `develop` equals that merge SHA before the audit closes.
- **Rollback trigger:** pin not equal to `f3d791…`; restore missing in
  `identify` or `converge`; helper copy missing before the unrestricted
  model; unique develop commits erased; operator SHA paste accepted;
  `roles.yml` / OpenAI route changed; any of the five VOC-112 no-change paths
  rewritten; PR #1051 or PR #1056 reused; VOC-132-T00 redispatched;
  `/tmp`-only byte comparison; false-revert evidence; or a snapshot-gap
  commit.
- **Rollback mechanism:** Revert the caller fixture/test/doc changes to the
  last reviewed `develop` pin
  `863fc1f35b1d35e4981a59166b0e939be1a2b681`. That last-known-good still has
  the checkout-lifetime defect in the fixture and the post-implementer
  helper-lifetime defect; rollback restores a known reviewed state, not a
  fully pinned #166 caller. Do not roll back by reverting infrastructure
  #166. Do not restore the VOC-112 JSON retargets from #1051 or the
  provenance-test weakening from #1056.
- **Last-known-good reference:** caller `develop` before this package's merge,
  with pin `863fc1f35b1d35e4981a59166b0e939be1a2b681`, VOC-112
  `subject_revision` `f9d11e232a07c7d7a9c433d02c9267912543ba10`, and the
  published provenance-test fail-closed `local` mode. Infrastructure last
  known good for the consumed contract is #166 merge
  `f3d79177bf8a9abe0dae550f39502165d494c576`, which this package does not
  revert.
