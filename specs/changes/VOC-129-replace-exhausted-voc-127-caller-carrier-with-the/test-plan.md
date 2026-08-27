# VOC-129 — Test Plan

## VOC-129-TEST-00 — Pin equals exact infrastructure #164 merge

- Covers: `VOC-129-AC-00`
- Preconditions: live `tooling/governance/fixtures/karsift-ai-infra/PINNED_SHA.txt`
  and matching caller pin assertions; no secrets or production data
- Procedure: Assert `PINNED_SHA.txt` strips to
  `863fc1f35b1d35e4981a59166b0e939be1a2b681`. Assert it does not equal
  `a9df74a63976d5239b84151fd01310835c999e7c` or
  `60afda3a44fd06b8c00b219771de7112f1aded6e`. Assert
  `scripts/foundation/voc097-fixture-matrix.test.mjs`,
  `voc104-ready-for-review-reuse.test.mjs`, and
  `voc108-authoritative-lifecycle.test.mjs` pin literals match the same SHA.
- Expected result: the attempt-2 stale-pin class (`a9df74a6…`) fails the new
  assertion; current develop pin `60afda3a…` cannot remain after this task.
- Evidence: `VOC-129-EV-00`

## VOC-129-TEST-01 — Fixture mirrors in-scope #164 files

- Covers: `VOC-129-AC-00`
- Preconditions: fixture tree after sync from merge
  `863fc1f35b1d35e4981a59166b0e939be1a2b681`
- Procedure: Assert the fixture contains
  `.github/workflows/release.yml`,
  `.github/workflows/reconcile-production-change.yml`,
  `config/release-checkout-ref-runner.py`, `config/branch_sync.py`,
  `config/branch-sync-runner.py`,
  `templates/project-repo/.github/workflows/pipeline.yml`, and the #164 tests
  `tests/test_release_checkout_ref_runner.py`, `tests/test_branch_sync.py`,
  and `tests/test_branch_sync_runner.py` (or the exact filenames present at
  that merge). Assert fixture `release.yml` contains
  `Synchronize integration to the exact promotion merge` / `branch-sync-runner.py`
  and does not restore a missing integration ref with
  `-f sha="$CHECKED_HEAD_SHA"`.
- Expected result: the caller fixture is the #164 contract, not the #163 pin.
- Evidence: `VOC-129-EV-00`

## VOC-129-TEST-02 — Checkout-ref ordering and missing-develop fallback

- Covers: `VOC-129-AC-01`
- Preconditions: mirrored `config/release-checkout-ref-runner.py` and
  `tests/test_release_checkout_ref_runner.py` at the #164 pin
- Procedure: Run the mirrored checkout-ref tests (or an equivalent caller
  wrapper that executes them). Assert existing `develop` is selected without
  reading `main`. Assert absent `develop` falls back to live `main`. Assert
  ambiguous or malformed refs fail closed. Assert fixture `release.yml`
  resolves the checkout ref after shared-policy checkout and before caller
  release-state checkout, and no longer hard-codes
  `ref: ${{ inputs.integration_branch }}` as the caller checkout.
- Expected result: the #164 absent-`develop` preflight defect cannot recur in
  the pinned fixture without failing this suite.
- Evidence: `VOC-129-EV-00`

## VOC-129-TEST-03 — Live pipeline exposes reconcile-production-change under 25 inputs

- Covers: `VOC-129-AC-02`
- Preconditions: live `.github/workflows/pipeline.yml` and
  `pipeline-verify.yml`
- Procedure: Assert live `pipeline.yml` `workflow_dispatch.inputs.action.options`
  includes `reconcile-production-change`. Assert a `reconcile-production-change`
  job calls `reconcile-production-change.yml@main` and forwards
  `authority_issue_number` from `issue_number` or the closed issue. Assert
  `existing_pr_number` remains implement-only. Assert no operator-typed SHA
  inputs exist on any live caller `workflow_dispatch`. Assert every live
  workflow with `workflow_dispatch` has at most 25 input keys (reuse VOC-126
  TEST-00 counting). Update VOC-126 action-options tests so they expect the
  new option without dropping recover/reconcile/implement actions.
- Expected result: GitHub-valid caller dispatch; #164 exceptional path is
  reachable; VOC-126 25-input bound still holds.
- Evidence: `VOC-129-EV-00`

## VOC-129-TEST-04 — Release and auto-advance wait on production-change reconcile

- Covers: `VOC-129-AC-02`
- Preconditions: live `pipeline.yml`
- Procedure: Assert `release` `needs: [merge-gate, reconcile-production-change]`
  (or the exact #164 template equivalent). Assert `auto-advance` on
  issue-closed wakes needs a successful `reconcile-production-change` result
  and does not proceed on failure. Assert `reconcile-production-change` is
  not a `schedule` trigger against arbitrary `main` deltas. Assert
  `existing_pr_number` is not used as the production-change identity.
- Expected result: exceptional reconciliation precedes ordinary release
  evaluation and cannot be skipped by a failed job.
- Evidence: `VOC-129-EV-00`

## VOC-129-TEST-05 — Mirrored release contract binds develop to mergeCommit.oid

- Covers: `VOC-129-AC-03`
- Preconditions: fixture `release.yml`, `config/branch_sync.py`, and
  `tests/test_release_policy.py` at the #164 pin
- Procedure: Run the mirrored release/branch-sync policy tests, or assert
  the same strings/behaviors they encode: one `gh pr merge`; sync to
  `mergeCommit.oid` before audit close; missing integration ref created at
  the merge SHA; unique develop commits fail closed; `ahead_by == 0` with
  unequal SHAs is not fully promoted. Assert fixture tests no longer require
  `-f sha="$CHECKED_HEAD_SHA"` restoration as success.
- Expected result: the #1033 checked-head restore class fails in the pinned
  fixture.
- Evidence: `VOC-129-EV-00`

## VOC-129-TEST-06 — Tree-equivalent sync does not schedule staging; allowlisted paths still do

- Covers: `VOC-129-AC-04`
- Preconditions: live `.github/workflows/deploy-staging.yml` and
  `scripts/foundation/voc111-deploy-staging-paths.test.mjs`
- Procedure: Assert an empty changed-path set does not select a staging
  deploy. Assert `specs/**` evidence-only paths do not select deploy. Assert
  existing VOC-111 allowlisted paths still select deploy. If implementation
  adds a job-level skip, assert it exits successfully without SSH/deploy
  steps when no allowlisted path changed, and does not skip when an
  allowlisted path did change. Do not treat a broadened allowlist as a pass.
- Expected result: merge-SHA develop sync does not waste a staging deploy;
  real runtime/deploy tree changes still deploy.
- Evidence: `VOC-129-EV-00`

## VOC-129-TEST-07 — Replacement carrier is VOC-129; #1041 and VOC-127 attempt 3 are absent

- Covers: `VOC-129-AC-05`
- Preconditions: this package's implementation PR and `t00-evidence.md`
- Procedure: Assert the implementation PR targets `develop` from a new
  VOC-129 branch, `Closes` only its own VOC-129 task issue, and does not
  `Closes` #1041, #1039, or #1035. Assert evidence records that #1041 was not
  merged or published and that VOC-127-T00 was not dispatched as attempt `3`.
  After promotion, evidence records the superseded close of #1041 and the
  audit-comment closes of #1039 and #1035 naming the VOC-129 merge, with no
  VOC-127 completion marker bound to #1041.
- Expected result: replacement is a new VOC-129 carrier, not VOC-127 attempt
  `3` and not a #1041 publication.
- Evidence: `VOC-129-EV-00`

## VOC-129-TEST-08 — Existing recovery, review, retry, and credential contracts remain

- Covers: `VOC-129-AC-06`
- Preconditions: final live `pipeline.yml`, fixture `implement.yml` /
  `release.yml`, `config/roles.yml`
- Procedure: Reuse or extend existing policy tests to assert roster-marker
  validation, exact-head `--match-head-commit`, required-check recovery,
  App-token on merge/sync mutation and job-token on recovery reads,
  two-attempt implementer bound, Cursor Composer implementer, Cursor Grok
  reviewer, no OpenAI path, and unchanged `config/roles.yml`. Confirm nested
  `karsift-ai-infra/.git` is not staged as a caller gitlink.
- Expected result: #164 caller consumption lands without weakening prior
  safety gates.
- Evidence: `VOC-129-EV-00`

## VOC-129-TEST-09 — Caller suites, docs, and governance validation

- Covers: `VOC-129-AC-06`
- Preconditions: final changed file set
- Procedure: Run the commands listed in `implementation-plan.md`. Assert
  current-state README/`pipeline.yml` comments/`AGENTS.md` and the listed
  operations/governance docs describe exact-merge-SHA develop sync,
  `reconcile-release` retry, and `reconcile-production-change`, and that
  historical A-003/VOC-075/VOC-127 records were not rewritten. Confirm
  `t00-evidence.md` records that no snapshot-gap task was added and no
  bootstrap exception was used. Confirm `git diff --check` is clean.
- Expected result: suites pass; docs match the live contract; pin updates are
  exact-SHA.
- Evidence: `VOC-129-EV-00`

Include positive, negative, authorization, failure, migration, accessibility, and
rollback coverage as applicable. Tests must not use secrets or production data.
