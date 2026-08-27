# VOC-130 — Implementation Plan

## Preconditions and protected areas

- Do not begin until the package is adopted and implementation-authorized.
- Protected areas: caller `tooling/governance/` fixtures and tests; live
  `.github/workflows/*` only if implementation proves a caller dispatch file
  must change to consume #165; current-state fixture README / pin comments.
- Prerequisites: confirm live `PINNED_SHA.txt` still equals
  `863fc1f35b1d35e4981a59166b0e939be1a2b681`. Confirm fixture `release.yml`
  still lacks `Restore shared lifecycle policy after caller checkout`. Confirm
  KARSIFT/karsift-ai-infra#165 remains merged as
  `8ce2b77a09a729e458a9f4cbea1ca26eb114d398`. Confirm VOC-129 caller PR #1046
  remains merged as `429d8c6d49303148ca1cc14dba5f6768a7863346`. Confirm live
  `pipeline.yml` still dispatches `release.yml@main`, still exposes
  `reconcile-production-change`, and still has at most 25 `workflow_dispatch`
  inputs.
- No bootstrap exception. VOC-124 already published
  `permission-workflows: write` on `publish-source`. T00's first run is
  attempt `1` on a new VOC-130 carrier. Do not treat an untracked local
  `karsift-ai-infra/` checkout as this repository's tracked tree.
- Preserve two-attempt implementer bounds, exact-SHA independent review,
  fail-closed credentials, and runner/App-token isolation.
- Do not print credential values. Do not change App installation permissions,
  rotate `KARSIFT_BOT_*` secrets, or edit `config/roles.yml`.
- Do not re-implement VOC-129. Do not snapshot the current develop/main gap.
  Do not add operator SHA inputs. Do not add OpenAI execution.

## File reconciliation and implementation sequence

### T00 — Pin exact infra #165 and cover the shared-policy restore

| Target | Action | Notes |
|--------|--------|-------|
| `tooling/governance/fixtures/karsift-ai-infra/**` | sync/pin | Exact merge `8ce2b77a09a729e458a9f4cbea1ca26eb114d398`; include `release.yml` restore and #165 restore tests; preserve #164 checkout-ref and branch-sync files already present |
| `tooling/governance/fixtures/karsift-ai-infra/PINNED_SHA.txt` | replace | Must equal `8ce2b77a09a729e458a9f4cbea1ca26eb114d398`; must not equal `863fc1f…` |
| `tooling/governance/fixtures/karsift-ai-infra/README.md` | modify current-state pin paragraph | Record VOC-130-T00 advance to #165; do not rewrite historical pin sentences as if they never happened |
| `tooling/governance/tests/` | create/extend | Exact pin equality; inequality with #164 pin; identify and converge restore-before-helper coverage; update VOC-129 and other live pin assertions that currently hard-code `863fc1f…` |
| `scripts/foundation/voc097-fixture-matrix.test.mjs`, `voc104-ready-for-review-reuse.test.mjs`, `voc108-authoritative-lifecycle.test.mjs` | modify | Advance pin literals from `863fc1f…` to `8ce2b77…` |
| `.github/workflows/pipeline.yml` | modify only if `VOC-130-DEP-07` proves required | Expected no-op; caller already uses `release.yml@main` |
| `specs/changes/VOC-130-.../t00-evidence.md` | update | Record pin SHA, restore coverage, validation commands, and VOC-129 / VOC-130 promotion handoff |

Ordered steps:

1. From current `develop`, create a new VOC-130 implementation branch. Do not
   re-check out or rewrite VOC-129 PR #1046.
2. Fetch KARSIFT/karsift-ai-infra at exact
   `8ce2b77a09a729e458a9f4cbea1ca26eb114d398` into an untracked nested
   checkout. Mirror the in-scope fixture subset from that SHA, including
   `release.yml` and the restore tests. Write `PINNED_SHA.txt` with that SHA
   only. Remove the nested checkout before commit so it cannot stage as a
   gitlink.
3. Advance every live caller pin assertion that currently equals `863fc1f…`
   to `8ce2b77…`. Add caller regressions that fixture `identify` and
   `converge` restore shared policy after caller checkout and before
   task-completion helpers, using `job.workflow_repository` /
   `job.workflow_sha` / `path: karsift-ai-infra`.
4. Update current-state fixture README/comments. Record evidence in
   `t00-evidence.md`. This package's caller PR `Closes` only its own VOC-130
   task issue.
5. After the exact reviewed caller PR merges, ordinary release evaluation (or
   `reconcile-release`) completes VOC-129's skipped promotion and this
   package's promotion. Do not add a snapshot-gap task.

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
`python3 -m unittest tooling.governance.tests.test_voc130_shared_policy_restore`),
record the exact commands in `t00-evidence.md` and run them in addition to
the suite above. Do not treat a missing suite as a pass.

Independent verifier (exact reviewed caller SHA) should confirm:

- `PINNED_SHA.txt` equals `8ce2b77a09a729e458a9f4cbea1ca26eb114d398` and does
  not equal `863fc1f…`;
- fixture `release.yml` restores shared policy in both `identify` and
  `converge` after caller checkout and before task-completion helpers;
- restore uses `job.workflow_repository` and `job.workflow_sha`;
- the #164 checkout-ref ordering / missing-`develop` path and
  `mergeCommit.oid` sync remain;
- live `pipeline.yml` still exposes `reconcile-production-change`, stays at
  most 25 inputs, and does not expose operator SHAs unless a proven #165
  caller-template delta requires a live edit;
- VOC-113 through VOC-129 isolation, recovery, App-token split, lease,
  retry limits, Cursor Composer/Grok roles, unchanged `roles.yml`, and
  non-closing source PR remain;
- no snapshot-gap task and no bootstrap exception were used;
- the implementer did not approve or merge its own work.

## Deployment and rollback

- **Staging/production effect:** None intentional for application runtime on
  the ordinary path (tree-equivalent develop-ref update). Staging remains
  path-selected.
- **Operational effect:** After this pin is live, `release.yml@main` (#165)
  restores nested shared policy after caller checkout, so task-completion
  validation can run instead of no-op'ing. After each governed promotion
  merge, `develop` equals that merge SHA before the audit closes.
- **Rollback trigger:** pin not equal to `8ce2b77…`; restore missing in
  `identify` or `converge`; unique develop commits erased; operator SHA paste
  accepted; `roles.yml` / OpenAI route changed; or a snapshot-gap commit.
- **Rollback mechanism:** Revert the caller fixture/test/doc changes to the
  last reviewed `develop` pin
  `863fc1f35b1d35e4981a59166b0e939be1a2b681`. That last-known-good still has
  the checkout-lifetime defect in the fixture; rollback restores a known
  reviewed state, not a fully pinned #165 caller. Do not roll back by
  reverting infrastructure #165.
- **Last-known-good reference:** caller `develop` before this package's merge,
  with pin `863fc1f35b1d35e4981a59166b0e939be1a2b681`. Infrastructure last
  known good for the consumed contract is #165 merge
  `8ce2b77a09a729e458a9f4cbea1ca26eb114d398`, which this package does not
  revert.
