# VOC-130 — Test Plan

## VOC-130-TEST-00 — Pin equals exact infrastructure #165 merge

- Covers: `VOC-130-AC-00`
- Preconditions: live `tooling/governance/fixtures/karsift-ai-infra/PINNED_SHA.txt`
  and matching caller pin assertions; no secrets or production data
- Procedure: Assert `PINNED_SHA.txt` strips to
  `8ce2b77a09a729e458a9f4cbea1ca26eb114d398`. Assert it does not equal
  `863fc1f35b1d35e4981a59166b0e939be1a2b681`. Assert
  `scripts/foundation/voc097-fixture-matrix.test.mjs`,
  `voc104-ready-for-review-reuse.test.mjs`, and
  `voc108-authoritative-lifecycle.test.mjs` pin literals match the same SHA.
  Assert live governance tests that previously hard-coded `863fc1f…` as the
  authoritative pin now equal `8ce2b77…`.
- Expected result: the #164 pin class (`863fc1f…`) fails the new assertion.
- Evidence: `VOC-130-EV-00`

## VOC-130-TEST-01 — Fixture mirrors in-scope #165 restore files

- Covers: `VOC-130-AC-00`
- Preconditions: fixture tree after sync from merge
  `8ce2b77a09a729e458a9f4cbea1ca26eb114d398`
- Procedure: Assert the fixture contains `.github/workflows/release.yml` and
  `tests/test_release_policy.py`. Assert fixture `release.yml` contains
  exactly two occurrences of `Restore shared lifecycle policy after caller
  checkout`. Assert the mirrored policy tests include
  `test_caller_checkout_rehydrates_shared_policy_before_lifecycle_helpers` (or
  the exact test name present at that merge).
- Expected result: the caller fixture is the #165 restore contract, not the
  #164 pin.
- Evidence: `VOC-130-EV-00`

## VOC-130-TEST-02 — Identify restores shared policy before validate-task

- Covers: `VOC-130-AC-01`
- Preconditions: mirrored fixture `release.yml`
- Procedure: Split fixture `release.yml` at the `identify` job. Assert
  `Checkout shared lifecycle policy` precedes `release-checkout-ref-runner.py`,
  which precedes `Checkout caller release state`, which precedes `Restore
  shared lifecycle policy after caller checkout`, which precedes
  `task-completion-runner.py validate-task`.
- Expected result: run `33066533397`'s identify missing-helper class cannot
  recur in the pinned fixture without failing this suite.
- Evidence: `VOC-130-EV-00`

## VOC-130-TEST-03 — Converge restores shared policy before validate-roster

- Covers: `VOC-130-AC-02`
- Preconditions: mirrored fixture `release.yml`
- Procedure: Split fixture `release.yml` at the `converge` job. Assert
  `Checkout caller release state` precedes `Restore shared lifecycle policy
  after caller checkout`, which precedes
  `task-completion-runner.py validate-roster`.
- Expected result: the latent converge checkout-lifetime defect cannot recur
  in the pinned fixture without failing this suite.
- Evidence: `VOC-130-EV-00`

## VOC-130-TEST-04 — Restore uses the immutable reusable-workflow revision

- Covers: `VOC-130-AC-01`, `VOC-130-AC-02`
- Preconditions: mirrored fixture `release.yml`
- Procedure: For the restore step in both `identify` and `converge`, assert
  the restored block contains `repository: ${{ job.workflow_repository }}`,
  `ref: ${{ job.workflow_sha }}`, and `path: karsift-ai-infra`. Assert it does
  not restore from `inputs.integration_branch` or an operator SHA input.
- Expected result: later helpers are the same immutable revision used to
  resolve the caller ref.
- Evidence: `VOC-130-EV-00`

## VOC-130-TEST-05 — #164 missing-develop, exact-SHA sync, and unique-develop contracts remain

- Covers: `VOC-130-AC-03`
- Preconditions: mirrored checkout-ref, branch-sync, and release-policy tests
  at the #165 pin
- Procedure: Reuse or extend existing fixture tests to assert existing
  `develop` is selected without reading `main`; absent `develop` falls back to
  live `main`; ambiguous or malformed refs fail closed; fixture `release.yml`
  still contains `Synchronize integration to the exact promotion merge` /
  `branch-sync-runner.py`; unique develop commits fail closed; and
  `-f sha="$CHECKED_HEAD_SHA"` is not treated as successful post-merge
  restoration. Assert live `pipeline.yml` still exposes
  `reconcile-production-change` under the 25-input limit without operator SHA
  inputs.
- Expected result: pinning #165 does not drop #164 fail-closed contracts.
- Evidence: `VOC-130-EV-00`

## VOC-130-TEST-06 — Existing recovery, review, retry, and credential contracts remain

- Covers: `VOC-130-AC-04`
- Preconditions: final live `pipeline.yml`, fixture `implement.yml` /
  `release.yml`, `config/roles.yml`
- Procedure: Reuse or extend existing policy tests to assert roster-marker
  validation, exact-head `--match-head-commit`, required-check recovery,
  App-token on merge/sync mutation and job-token on recovery reads,
  two-attempt implementer bound, Cursor Composer implementer, Cursor Grok
  reviewer, no OpenAI path, sanitized raw-error controls, and unchanged
  `config/roles.yml`. Confirm nested `karsift-ai-infra/.git` is not staged as
  a caller gitlink.
- Expected result: #165 caller consumption lands without weakening prior
  safety gates.
- Evidence: `VOC-130-EV-00`

## VOC-130-TEST-07 — Caller suites, docs, and governance validation

- Covers: `VOC-130-AC-04`
- Preconditions: final changed file set
- Procedure: Run the commands listed in `implementation-plan.md`. Assert
  current-state fixture README/comments name pin `8ce2b77…` and the
  post-caller-checkout restore, and that historical A-003/VOC-075/VOC-127/
  VOC-129 records were not rewritten. Confirm `t00-evidence.md` records that
  no snapshot-gap task was added and no bootstrap exception was used. Confirm
  `git diff --check` is clean.
- Expected result: suites pass; docs match the live contract; pin updates are
  exact-SHA.
- Evidence: `VOC-130-EV-00`

## VOC-130-TEST-08 — Replacement carrier is VOC-130; no snapshot-gap; VOC-129 is not re-implemented

- Covers: `VOC-130-AC-05`
- Preconditions: this package's implementation PR and `t00-evidence.md`
- Procedure: Assert the implementation PR targets `develop` from a new
  VOC-130 branch, `Closes` only its own VOC-130 task issue, and does not
  `Closes` VOC-129 issues. Assert evidence records that VOC-129 PR #1046 was
  not re-implemented and that no develop/main gap snapshot was committed.
  After promotion, evidence records ordinary release/sync/closure for VOC-129
  and this package, with closed state not treated as completion proof.
- Expected result: repair is a new VOC-130 carrier, not a VOC-129 retry and
  not a snapshot-then-check package.
- Evidence: `VOC-130-EV-00`

Include positive, negative, authorization, failure, migration, accessibility, and
rollback coverage as applicable. Tests must not use secrets or production data.
