# VOC-127 — Implementation Plan

## Preconditions and protected areas

- Do not begin until the package is adopted and implementation-authorized.
- Protected areas: `KARSIFT/karsift-ai-infra` `release.yml`, release helpers,
  and tests; caller `.github/workflows/pipeline.yml` exceptional reconcile
  dispatch; `deploy-staging.yml` path selection; caller
  `tooling/governance/` fixtures and tests; current-state release and branch
  documentation.
- Prerequisites: confirm `release.yml` still merges with
  `--match-head-commit "$CHECKED_HEAD_SHA"`, still restores a missing
  integration ref at `CHECKED_HEAD_SHA`, still no-ops when the promotion PR
  is already terminal, and still closes as "Already promoted" when
  `ahead_by == 0`. Confirm `test_release_policy.py` still requires
  `CHECKED_HEAD_SHA` restoration. Confirm live `develop` and `main` currently
  match after the 2026-08-27 repair, and do not snapshot that gap. Confirm
  VOC-111 `deploy-staging.yml` path allowlist remains the staging selector.
  Confirm VOC-126 `pipeline.yml` is at most 25 `workflow_dispatch` inputs
  with `existing_pr_number` implement-only and verifiers on
  `pipeline-verify.yml`.
- No bootstrap exception. VOC-124 already published
  `permission-workflows: write` on `publish-source`. T00's first run is
  attempt `1`. Do not treat an untracked local `karsift-ai-infra/` checkout
  as this repository's tracked tree.
- Preserve two-attempt implementer bounds, exact-SHA independent review,
  fail-closed credentials, and runner/App-token isolation.
- Do not print credential values. Do not change App installation permissions,
  rotate `KARSIFT_BOT_*` secrets, or edit `config/roles.yml`.
- Do not implement VOC-122. Do not fast-forward `main`. Do not add operator
  SHA inputs.

## File reconciliation and implementation sequence

### T00 — Exact-merge-SHA develop sync and exceptional main-only reconciliation

| Target | Action | Notes |
|--------|--------|-------|
| `KARSIFT/karsift-ai-infra/.github/workflows/release.yml` | modify | Bind-and-sync `develop` to `mergeCommit.oid` before audit close; already-merged recover; stop `ahead_by == 0` unequal-SHA short-circuit; recreate missing ref at merge SHA; keep one `gh pr merge` |
| `KARSIFT/karsift-ai-infra/config/` | create/extend | Testable bind-and-sync helper (`VOC-127-DEP-07` name) |
| `KARSIFT/karsift-ai-infra/tests/test_release_policy.py` and new `test_voc127_*.py` | modify/create | Replace CHECKED_HEAD_SHA restore assertion; cover `VOC-127-D08` matrix |
| `KARSIFT/karsift-ai-infra/README.md` | modify | Current-state: promotion merge then exact-SHA develop sync; `reconcile-release` retries sync; exceptional main-only path is not ordinary |
| `KARSIFT/karsift-ai-infra/templates/project-repo/.github/workflows/pipeline.yml` | modify | Exceptional mutating action; ≤25 inputs; no operator SHAs; do not overload `existing_pr_number` |
| `.github/workflows/pipeline.yml` | modify | Same exceptional action as the template |
| `.github/workflows/deploy-staging.yml` | modify only if needed | Job-level skip when `on.push.paths` cannot skip tree-equivalent merge-commit fast-forwards; do not broaden VOC-111 allowlist |
| `scripts/foundation/voc111-deploy-staging-paths.test.mjs` | extend | Empty diff and specs-only paths do not select deploy; allowlisted paths still do |
| `tooling/governance/fixtures/karsift-ai-infra/**` | sync/pin | Exact reviewed infra merge when consumed; do not pin `60afda3a…` |
| `tooling/governance/tests/` | modify/extend | Fixture regressions; pin literal; exceptional dispatch; input-count |
| `scripts/foundation/*` pin tests | modify if they still assert the previous pin | Same-task pin literals |
| `AGENTS.md` | modify current-state reconcile-release / release-authority sentences that would become false | Do not add a new internals runbook |
| `docs/operations/11-devops-and-ci-cd.md` | modify current-state | Post-promotion develop equals merge SHA; staging path selection retained |
| `docs/operations/10-development-workflow.md` | modify current-state branch strategy | After promotion merge, `develop` is advanced to that merge commit |
| `docs/operations/15-ai-native-product-and-engineering-operating-model.md` | modify current-state release paragraphs only | Historical correction notes stay; emergency-PR "reconcile back into develop" gains the governed path |
| `docs/governance/16-autonomous-development-operating-model.md` | modify current-state branch/merge bullets if they would become false | Historical records unchanged |
| `specs/changes/VOC-127-.../t00-evidence.md` | update | Record mechanism, validation commands, infra SHA, pin applicability |

Ordered steps:

1. In a clean isolated `KARSIFT/karsift-ai-infra` worktree based on current
   `main`, implement bind-and-sync after the single promotion merge, including
   already-`MERGED` recovery and fail-closed races. Recreate missing
   integration refs at the merge SHA. Keep exactly one `gh pr merge`.
2. Add helper unit tests and update `test_release_policy.py`. Add exceptional
   main-only dispatch to the project-repo `pipeline.yml` template. Update
   current-state README/comments. Count `workflow_dispatch` inputs and fail
   if any block exceeds 25.
3. Run the infra unit/policy suite. Open one reviewed infra PR that
   `Relates to KARSIFT/vocanova-platform-sandbox#<task>` and does not use a
   closing keyword. Merge it first so `release.yml@main` syncs `develop` to
   the merge SHA before the caller consumes it.
4. After a different actor merges that exact reviewed infra head, update the
   live caller `pipeline.yml`, staging selector if required, VOC-111 tests,
   current-state docs listed above, fixture pin, and caller tests. Record
   evidence in `t00-evidence.md`. Do not snapshot the already-repaired
   develop/main equality as a drift baseline.
5. This package's caller PR `Closes` only its own VOC-127 task issue.

## Validation and independent verification

Deterministic commands, as applicable to the final changed file set:

```bash
# In the checked-out primary KARSIFT/karsift-ai-infra source:
python3 -m unittest discover -s tests -p 'test_*.py'

# In this caller repository:
bash scripts/governance/validate-governance.sh
bash scripts/governance/classify-change-risk.sh
python3 -m unittest discover -s tooling/governance/tests -p 'test_*.py'
node scripts/foundation/voc111-deploy-staging-paths.test.mjs
git diff --check
```

If implementation adds narrower targeted commands (for example
`python3 -m unittest tests.test_voc127_merge_sha_sync`) or other foundation
pin tests, record the exact commands in `t00-evidence.md` and run them in
addition to the suite above.

Independent verifier (exact reviewed caller SHA, and exact reviewed infra SHA
when an infra PR is opened) should confirm:

- after `--merge`, `develop` is at `mergeCommit.oid` before audit close;
- already-merged `reconcile-release` syncs without a new PR or second merge;
- unique develop commits, moved `main`, malformed merges, and missing refs
  fail closed;
- auto-deleted `develop` is recreated at the merge SHA;
- equal tips do not open a new promotion PR;
- exceptional main-only action is adopted-package + merged main PR, not
  operator SHAs, and is not scheduled;
- tree-equivalent sync does not keep staging scheduled; VOC-111 allowlist
  is not broadened;
- VOC-113 through VOC-126 isolation, recovery, App-token split, lease,
  retry limits, Cursor Composer/Grok roles, unchanged `roles.yml`, and
  non-closing source PR remain;
- current-state docs describe exact-merge-SHA sync and exceptional
  main-only reconciliation;
- the caller fixture pin equals the exact reviewed infra merge when
  consumed, or evidence records why the pin was not applicable — and the
  pin is not the drafting-time `60afda3a…`;
- no snapshot-gap task was added; no bootstrap exception was used;
- the implementer did not approve or merge its own work on either carrier.

## Deployment and rollback

- **Staging/production effect:** None intentional for application runtime on
  the ordinary path (tree-equivalent develop-ref update). Staging remains
  path-selected. Exceptional main-only recon may introduce an already
  authorized main tree onto `develop`; allowlisted paths may then deploy
  staging as they do today.
- **Operational effect:** After each governed promotion merge, `develop`
  equals that merge SHA before the audit closes. `reconcile-release` retries
  that sync. Operators can reconcile an adopted main-only merged PR onto
  `develop` through the exceptional action.
- **Rollback trigger:** `develop` left behind a promotion merge SHA; unique
  develop commits erased; a second promotion PR or merge; operator SHA paste
  accepted; unadopted main-only sync; tree-equivalent sync performing a full
  staging deploy; or `roles.yml` / OpenAI route changed.
- **Rollback mechanism:** Revert the infra and caller workflow/helper/fixture/
  test/doc changes to the last reviewed `release.yml` that restored
  `CHECKED_HEAD_SHA`. That last-known-good still recreates the #1033 ancestor
  gap; rollback restores a known reviewed state, not a fully converged
  post-merge develop.
- **Last-known-good reference:** `KARSIFT/karsift-ai-infra` `release.yml` on
  `main` before this package's infra merge, and the caller pin
  `60afda3a44fd06b8c00b219771de7112f1aded6e` at drafting time. Do not roll
  back by repeating the 2026-08-27 repository-settings recreation of
  `develop`.
