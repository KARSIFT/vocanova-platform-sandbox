# VOC-132 — Test Plan

## VOC-132-TEST-00 — Pin equals exact infrastructure #165 merge

- Covers: `VOC-132-AC-00`
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
- Evidence: `VOC-132-EV-00`

## VOC-132-TEST-01 — Fixture mirrors the necessary #165 files by recorded SHA-256

- Covers: `VOC-132-AC-00`
- Preconditions: fixture tree after sync from merge
  `8ce2b77a09a729e458a9f4cbea1ca26eb114d398`; recorded expected hashes
  committed in the caller regression
- Procedure: Assert SHA-256 of fixture `.github/workflows/release.yml` and
  `tests/test_release_policy.py` equal the hashes recorded from infra merge
  `8ce2b77…` (`VOC-132-D11`, reconfirmed at implementation). Assert the test
  does not open, clone, or require a machine-specific `/tmp` karsift-ai-infra
  checkout. Assert fixture `release.yml` contains exactly two occurrences of
  `Restore shared lifecycle policy after caller checkout`. Assert the
  mirrored policy tests include
  `test_caller_checkout_rehydrates_shared_policy_before_lifecycle_helpers`.
- Expected result: the caller fixture is the #165 restore contract, not the
  #164 pin, and not a rewritten subset. CI can prove byte identity without a
  `/tmp` tree.
- Evidence: `VOC-132-EV-00`

## VOC-132-TEST-02 — Identify restores shared policy before validate-task

- Covers: `VOC-132-AC-01`
- Preconditions: mirrored fixture `release.yml`
- Procedure: Split fixture `release.yml` at the `identify` job. Assert
  `Checkout shared lifecycle policy` precedes `release-checkout-ref-runner.py`,
  which precedes `Checkout caller release state`, which precedes `Restore
  shared lifecycle policy after caller checkout`, which precedes
  `task-completion-runner.py validate-task`.
- Expected result: run `33066533397`'s identify missing-helper class cannot
  recur in the pinned fixture without failing this suite.
- Evidence: `VOC-132-EV-00`

## VOC-132-TEST-03 — Converge restores shared policy before validate-roster

- Covers: `VOC-132-AC-02`
- Preconditions: mirrored fixture `release.yml`
- Procedure: Split fixture `release.yml` at the `converge` job. Assert
  `Checkout caller release state` precedes `Restore shared lifecycle policy
  after caller checkout`, which precedes
  `task-completion-runner.py validate-roster`.
- Expected result: the latent converge checkout-lifetime defect cannot recur
  in the pinned fixture without failing this suite.
- Evidence: `VOC-132-EV-00`

## VOC-132-TEST-04 — Restore uses the immutable reusable-workflow revision without persisting credentials

- Covers: `VOC-132-AC-01`, `VOC-132-AC-02`
- Preconditions: mirrored fixture `release.yml`
- Procedure: For the restore step in both `identify` and `converge`, assert
  the restored block contains `repository: ${{ job.workflow_repository }}`,
  `ref: ${{ job.workflow_sha }}`, `path: karsift-ai-infra`, and
  `persist-credentials: false`. Assert it does not restore from
  `inputs.integration_branch` or an operator SHA input.
- Expected result: later helpers are the same immutable revision used to
  resolve the caller ref, and credentials are not persisted.
- Evidence: `VOC-132-EV-00`

## VOC-132-TEST-05 — #164 missing-develop, exact-SHA sync, and unique-develop contracts remain

- Covers: `VOC-132-AC-03`
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
- Evidence: `VOC-132-EV-00`

## VOC-132-TEST-06 — Existing recovery, review, retry, and credential contracts remain

- Covers: `VOC-132-AC-04`
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
- Evidence: `VOC-132-EV-00`

## VOC-132-TEST-07 — Caller suites, docs, and governance validation

- Covers: `VOC-132-AC-04`
- Preconditions: final changed file set
- Procedure: Run the commands listed in `implementation-plan.md`. Assert
  current-state fixture README/comments name pin `8ce2b77…` and the
  post-caller-checkout restore, and that historical A-003/VOC-075/VOC-127/
  VOC-129/VOC-130/VOC-131 records were not rewritten. Confirm `AGENTS.md` and
  `.agents/skills/vocanova-repo-navigator/SKILL.md` are not in the
  implementation diff. Confirm `t00-evidence.md` records the exact reviewed
  head, that no snapshot-gap task was added, no bootstrap exception was used,
  and PR #1051 and PR #1056 were not reused. Confirm `git diff --check` is
  clean.
- Expected result: suites pass; docs match the live contract; pin updates are
  exact-SHA; hashed VOC-112 sources are untouched.
- Evidence: `VOC-132-EV-00`

## VOC-132-TEST-08 — Complete VOC-112 no-change boundary remains byte-identical to the carrier develop base

- Covers: `VOC-132-AC-05`
- Preconditions: implementation PR against current `develop`; no secrets or
  production data
- Procedure: Resolve the carrier `develop` base (pull-request merge-base with
  `develop`, or `origin/develop` when that SHA is available). Assert each of
  the following paths is byte-identical to that base and is not named by
  `git diff` against that base:
  - `scripts/foundation/fixtures/voc112-navigation-benchmark-traces.json`
  - `scripts/foundation/fixtures/voc112-skill-discovery-evidence.json`
  - `scripts/foundation/voc112-navigation-benchmark.test.mjs`
  - `AGENTS.md`
  - `.agents/skills/vocanova-repo-navigator/SKILL.md`
  Assert both JSON files still contain `subject_revision`
  `f9d11e232a07c7d7a9c433d02c9267912543ba10`. Assert the provenance test still
  contains the fail-closed `local` mode requirement that a full checkout must
  already contain the captured commit, and does not treat a missing subject
  in a full checkout as implicit `squash-safe-push`. The regression must fail
  if any named path is retargeted, recaptured, weakened, or otherwise
  rewritten, including the JSON class on #1051 heads `a04a41a…` /
  `e846cc2…` and the provenance-test class on #1056 head `c11454e…`.
- Expected result: the VOC-130 JSON class and the VOC-131 provenance-test
  class cannot recur unnoticed; CI does not enter `pr-ancestry` merely
  because those fixtures changed.
- Evidence: `VOC-132-EV-00`

## VOC-132-TEST-09 — Evidence names the exact reviewed head and does not claim a false revert

- Covers: `VOC-132-AC-05`
- Preconditions: `t00-evidence.md` and the implementation PR diff against the
  carrier `develop` base
- Procedure: Assert `t00-evidence.md` records the exact 40-character
  implementation head SHA that independent review binds. If evidence claims
  that a protected VOC-112 no-change path was reverted, unchanged, or
  restored, assert that path is absent from `git diff` against the carrier
  base. The VOC-131 class — evidence claiming
  `voc112-navigation-benchmark.test.mjs` was reverted while the tree still
  weakened `local` mode — must fail this test.
- Expected result: evidence truthfully describes the exact reviewed head.
- Evidence: `VOC-132-EV-00`

## VOC-132-TEST-10 — Replacement carrier is VOC-132; no snapshot-gap; VOC-129/VOC-130/VOC-131 are not re-implemented

- Covers: `VOC-132-AC-06`
- Preconditions: this package's implementation PR and `t00-evidence.md`
- Procedure: Assert the implementation PR targets `develop` from a new
  VOC-132 branch, `Closes` only its own VOC-132 task issue, and does not
  `Closes` VOC-129, VOC-130, VOC-131, or root #1057. Assert evidence records
  that PR #1051 and PR #1056 were not reused, merged, or modified; that
  VOC-129 PR #1046 was not re-implemented; that no VOC-129, VOC-130, or
  VOC-131 completion marker was manufactured; and that no develop/main gap
  snapshot was committed. After promotion, evidence records ordinary
  release/sync/closure for VOC-129, VOC-130, VOC-131, and this package, with
  closed state not treated as completion proof.
- Expected result: repair is a new VOC-132 carrier, not a VOC-130 or VOC-131
  retry, not a VOC-129 rewrite, and not a snapshot-then-check package.
- Evidence: `VOC-132-EV-00`

Include positive, negative, authorization, failure, migration, accessibility, and
rollback coverage as applicable. Tests must not use secrets or production data.
