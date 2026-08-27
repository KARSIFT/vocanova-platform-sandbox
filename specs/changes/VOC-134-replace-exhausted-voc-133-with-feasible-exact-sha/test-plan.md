# VOC-134 — Test Plan

## VOC-134-TEST-00 — Pin equals exact infrastructure #166 merge

- Covers: `VOC-134-AC-00`
- Preconditions: live `tooling/governance/fixtures/karsift-ai-infra/PINNED_SHA.txt`
  and matching caller pin assertions; no secrets or production data
- Procedure: Assert `PINNED_SHA.txt` strips to
  `f3d79177bf8a9abe0dae550f39502165d494c576`. Assert it does not equal
  `863fc1f35b1d35e4981a59166b0e939be1a2b681`. Assert it does not equal
  `8ce2b77a09a729e458a9f4cbea1ca26eb114d398`. Assert
  `scripts/foundation/voc097-fixture-matrix.test.mjs`,
  `voc104-ready-for-review-reuse.test.mjs`, and
  `voc108-authoritative-lifecycle.test.mjs` pin literals match the same SHA.
  Assert live governance tests that previously hard-coded `863fc1f…` as the
  authoritative pin now equal `f3d791…`.
- Expected result: the #164 pin class (`863fc1f…`) and the #165 pin class
  (`8ce2b77…`) fail the new assertion.
- Evidence: `VOC-134-EV-00`

## VOC-134-TEST-01 — Fixture mirrors the necessary #166 files by recorded SHA-256

- Covers: `VOC-134-AC-00`
- Preconditions: fixture tree after sync from merge
  `f3d79177bf8a9abe0dae550f39502165d494c576`; recorded expected hashes
  committed in the caller regression
- Procedure: Assert SHA-256 of fixture `.github/workflows/implement.yml`,
  `.github/workflows/release.yml`, `config/implementer_nested_checkout.py`,
  `tests/test_release_policy.py`, `tests/test_voc121_implement_policy.py`,
  `tests/test_voc123_source_bundle.py`, and `CHANGELOG.md` equal the hashes
  recorded from infra merge `f3d791…` (`VOC-134-D11`, reconfirmed at
  implementation). Assert the test does not open, clone, or require a
  machine-specific `/tmp` karsift-ai-infra checkout. Assert fixture
  `release.yml` contains exactly two occurrences of `Restore shared lifecycle
  policy after caller checkout`. Assert the mirrored policy tests include
  `test_caller_checkout_rehydrates_shared_policy_before_lifecycle_helpers`
  and the #166 nested-checkout classifier tests.
- Expected result: the caller fixture is the #166 contract, not the #164 pin,
  and not a rewritten subset. CI can prove byte identity without a `/tmp`
  tree.
- Evidence: `VOC-134-EV-00`

## VOC-134-TEST-02 — Identify restores shared policy before validate-task

- Covers: `VOC-134-AC-01`
- Preconditions: mirrored fixture `release.yml`
- Procedure: Split fixture `release.yml` at the `identify` job. Assert
  `Checkout shared lifecycle policy` precedes `release-checkout-ref-runner.py`,
  which precedes `Checkout caller release state`, which precedes `Restore
  shared lifecycle policy after caller checkout`, which precedes
  `task-completion-runner.py validate-task`.
- Expected result: run `33066533397`'s identify missing-helper class cannot
  recur in the pinned fixture without failing this suite.
- Evidence: `VOC-134-EV-00`

## VOC-134-TEST-03 — Converge restores shared policy before validate-roster

- Covers: `VOC-134-AC-02`
- Preconditions: mirrored fixture `release.yml`
- Procedure: Split fixture `release.yml` at the `converge` job. Assert
  `Checkout caller release state` precedes `Restore shared lifecycle policy
  after caller checkout`, which precedes
  `task-completion-runner.py validate-roster`.
- Expected result: the latent converge checkout-lifetime defect cannot recur
  in the pinned fixture without failing this suite.
- Evidence: `VOC-134-EV-00`

## VOC-134-TEST-04 — Restore uses the immutable reusable-workflow revision without persisting credentials

- Covers: `VOC-134-AC-01`, `VOC-134-AC-02`
- Preconditions: mirrored fixture `release.yml`
- Procedure: For the restore step in both `identify` and `converge`, assert
  the restored block contains `repository: ${{ job.workflow_repository }}`,
  `ref: ${{ job.workflow_sha }}`, `path: karsift-ai-infra`, and
  `persist-credentials: false`. Assert it does not restore from
  `inputs.integration_branch` or an operator SHA input.
- Expected result: later helpers are the same immutable revision used to
  resolve the caller ref, and credentials are not persisted.
- Evidence: `VOC-134-EV-00`

## VOC-134-TEST-05 — #164 missing-develop, exact-SHA sync, and unique-develop contracts remain

- Covers: `VOC-134-AC-04`
- Preconditions: mirrored checkout-ref, branch-sync, and release-policy tests
  at the #166 pin
- Procedure: Reuse or extend existing fixture tests to assert existing
  `develop` is selected without reading `main`; absent `develop` falls back to
  live `main`; ambiguous or malformed refs fail closed; fixture `release.yml`
  still contains `Synchronize integration to the exact promotion merge` /
  `branch-sync-runner.py`; unique develop commits fail closed; and
  `-f sha="$CHECKED_HEAD_SHA"` is not treated as successful post-merge
  restoration. Assert live `pipeline.yml` still exposes
  `reconcile-production-change` under the 25-input limit without operator SHA
  inputs.
- Expected result: pinning #166 does not drop #164 fail-closed contracts.
- Evidence: `VOC-134-EV-00`

## VOC-134-TEST-06 — #166 helper-lifetime and nested-checkout classifier contracts

- Covers: `VOC-134-AC-03`
- Preconditions: mirrored fixture `implement.yml` and
  `config/implementer_nested_checkout.py`; no secrets or production data
- Procedure: Assert `Preserve post-implementer lifecycle helpers` precedes
  `Run implementer (cursor-agent)`, which precedes `Commit implementer's
  work`. Assert the preserve step copies `run-app-checks.sh`,
  `prepare_cursor_model.py`, `implementer_source_carrier.py`,
  `cross_repo_reference.py`, and `implementer_nested_checkout.py` to immutable
  `/tmp` paths. Assert commit-time classification uses
  `python3 "$HELPER_DIR/implementer_nested_checkout.py"` and does not copy
  helpers from the possibly deleted nested tree. Assert `absent` continues
  caller publication with `has_source_changes=false`. Assert the classifier
  rejects parent-Git inheritance, symlink, and non-directory paths. Assert a
  distinct nested Git checkout still creates the exact-head bundle and uses
  force-with-lease / remote-head binding. If the mirrored
  `tests/test_voc121_implement_policy.py` omits a dedicated non-directory
  case, cover that class in caller `tooling/governance/tests/` without
  editing the mirrored bytes.
- Expected result: run `33079499176`'s post-removal `cp: cannot stat`
  class cannot recur in the pinned fixture without failing this suite.
- Evidence: `VOC-134-EV-00`

## VOC-134-TEST-07 — Existing recovery, review, retry, and credential contracts remain

- Covers: `VOC-134-AC-05`
- Preconditions: final live `pipeline.yml`, fixture `implement.yml` /
  `release.yml`, `config/roles.yml`
- Procedure: Reuse or extend existing policy tests to assert roster-marker
  validation, exact-head `--match-head-commit`, required-check recovery,
  App-token on merge/sync mutation and job-token on recovery reads,
  two-attempt implementer bound, Cursor Composer implementer, Cursor Grok
  reviewer, no OpenAI path, sanitized raw-error controls, and unchanged
  `config/roles.yml`. Confirm nested `karsift-ai-infra/.git` is not staged as
  a caller gitlink.
- Expected result: #166 caller consumption lands without weakening prior
  safety gates.
- Evidence: `VOC-134-EV-00`

## VOC-134-TEST-08 — Caller suites, docs, and governance validation

- Covers: `VOC-134-AC-05`
- Preconditions: final changed file set
- Procedure: Run the commands listed in `implementation-plan.md`. Assert
  current-state fixture README/comments name pin `f3d791…`, the
  post-caller-checkout restore, and the post-implementer helper-lifetime
  contract, and that historical A-003/VOC-075/VOC-127/VOC-129/VOC-130/
  VOC-131/VOC-132/VOC-133 records were not rewritten. Confirm the
  caller-owned fixture README was not replaced by the infrastructure
  repository README. Confirm `AGENTS.md` and
  `.agents/skills/vocanova-repo-navigator/SKILL.md` are not in the
  implementation diff. Confirm `t00-evidence.md` records the immutable
  carrier base, the final infra merge, and the feasible exact-head binding
  contract, that no snapshot-gap task was added, no bootstrap exception was
  used, PR #1051 / PR #1056 / PR #1065 were not reused, and VOC-132-T00
  (#1059) and VOC-133-T00 (#1063) were not redispatched. Confirm
  `git diff --check` is clean.
- Expected result: suites pass; docs match the live contract; pin updates are
  exact-SHA; hashed VOC-112 sources and `package.json` are untouched.
- Evidence: `VOC-134-EV-00`

## VOC-134-TEST-09 — Complete VOC-112 no-change boundary remains byte-identical to the immutable carrier-base SHA

- Covers: `VOC-134-AC-06`
- Preconditions: implementation PR against the immutable carrier-base SHA
  selected before any in-scope edit (expected
  `95a779f9e62090f856ed03f389e7ac1d901aaa14`); no secrets or production data
- Procedure: Read the committed immutable carrier-base SHA constant. Assert
  it is a 40-character lowercase hex SHA. Resolve that exact commit object
  (`git cat-file -e <sha>^{commit}` or equivalent). For each of the following
  paths, resolve `git show <sha>:<path>` (or equivalent) and assert the
  working tree bytes are identical and that `git diff <sha>` does not name
  the path:
  - `scripts/foundation/fixtures/voc112-navigation-benchmark-traces.json`
  - `scripts/foundation/fixtures/voc112-skill-discovery-evidence.json`
  - `scripts/foundation/voc112-navigation-benchmark.test.mjs`
  - `AGENTS.md`
  - `.agents/skills/vocanova-repo-navigator/SKILL.md`
  If the commit object or any named path cannot be resolved, fail—do not
  skip. Do not fall back to `develop` or `origin/develop` after the constant
  is selected. Assert both JSON files still contain `subject_revision`
  `f9d11e232a07c7d7a9c433d02c9267912543ba10`. Assert the provenance test still
  contains the fail-closed `local` mode requirement that a full checkout must
  already contain the captured commit, and does not treat a missing subject
  in a full checkout as implicit `squash-safe-push`. The regression must fail
  if any named path is retargeted, recaptured, weakened, or otherwise
  rewritten, including the JSON class on #1051 heads `a04a41a…` /
  `e846cc2…` and the provenance-test class on #1056 head `c11454e…`.
- Expected result: the VOC-130 JSON class, the VOC-131 provenance-test class,
  and a moving-`develop`-ref comparison cannot recur unnoticed; CI does not
  enter `pr-ancestry` merely because those fixtures changed.
- Evidence: `VOC-134-EV-00`

## VOC-134-TEST-10 — package.json is carrier-base-identical and provenance is not bypassed

- Covers: `VOC-134-AC-07`
- Preconditions: implementation PR against the immutable carrier-base SHA;
  no secrets or production data
- Procedure: Resolve `package.json` at the immutable carrier-base SHA and
  assert working-tree bytes are identical and that `git diff <sha>` does not
  name `package.json`. If the path cannot be resolved, fail—do not skip.
  Assert `package.json` does not set `VOC112_CAPTURE_PROVENANCE_MODE`, does
  not point `test` at a capture-commit fetch helper, and still invokes
  `node --test scripts/foundation/*.test.mjs` without a provenance-mode
  prefix. Assert `scripts/foundation/ensure-voc112-capture-commits.mjs` is
  absent. Assert no evidence-stamping helper mutates `t00-evidence.md` at
  test or runtime. Assert no test rewrites evidence in memory so that a
  prior failed head can pass.
- Expected result: the VOC-133 attempt-1 fetch-before-test class and
  attempt-2 `pr-validation`/`HEAD` wrapper plus evidence-stamp class cannot
  recur unnoticed.
- Evidence: `VOC-134-EV-00`

## VOC-134-TEST-11 — Exact-revision evidence is fail-closed and Git-feasible

- Covers: `VOC-134-AC-08`
- Preconditions: `t00-evidence.md` and the implementation PR diff against the
  immutable carrier-base SHA
- Procedure: Assert `t00-evidence.md` records the immutable 40-character
  carrier-base SHA and the final infra merge
  `f3d79177bf8a9abe0dae550f39502165d494c576`. Assert it records the
  reconfirmed SHA-256 hashes and the validation commands. Assert it states
  that the live implementation head is bound by the App-authored
  independent-review comment/check, that merge-gate must reject any mismatch,
  and that post-merge/root-issue audit may record the reviewed head and merge
  SHA. Assert it does **not** require a tracked file in the same commit to
  equal `HEAD` or to contain that commit's own SHA. Assert no test or helper
  rewrites `t00-evidence.md` on disk or in memory at test time. If evidence
  claims that a protected VOC-112 no-change path or `package.json` was
  reverted, unchanged, or restored, assert that path is absent from
  `git diff` against the carrier-base SHA. The VOC-131 class — evidence
  claiming `voc112-navigation-benchmark.test.mjs` was reverted while the tree
  still weakened `local` mode — and the VOC-133 class — tests passing while
  committed evidence named a prior failed head — must fail this test. Assert
  evidence does not contain raw provider responses, credentials, or a full
  CI log.
- Expected result: evidence truthfully describes the carrier base and the
  consumed infra merge, and exact-head binding remains fail-closed without a
  Git-impossible self-SHA.
- Evidence: `VOC-134-EV-00`

## VOC-134-TEST-12 — Replacement carrier is VOC-134; no snapshot-gap; VOC-129 through VOC-133 are not re-implemented

- Covers: `VOC-134-AC-09`
- Preconditions: this package's implementation PR and `t00-evidence.md`
- Procedure: Assert the implementation PR targets `develop` from a new
  VOC-134 branch, `Closes` only its own VOC-134 task issue, and does not
  `Closes` VOC-129, VOC-130, VOC-131, VOC-132, VOC-133, or root #1066. Assert
  evidence records that PR #1051, PR #1056, and PR #1065 were not reused,
  merged, cherry-picked, or modified; that VOC-132-T00 (#1059) and
  VOC-133-T00 (#1063) were not redispatched; that VOC-129 PR #1046 was not
  re-implemented; that no VOC-129, VOC-130, VOC-131, VOC-132, or VOC-133
  completion marker was manufactured; and that no develop/main gap snapshot
  was committed. After promotion, evidence records ordinary
  release/sync/closure for VOC-129, VOC-130, VOC-131, VOC-132, VOC-133, and
  this package, with closed state not treated as completion proof.
- Expected result: repair is a new VOC-134 carrier, not a VOC-130 / VOC-131 /
  VOC-132 / VOC-133 retry, not a VOC-129 rewrite, and not a snapshot-then-check
  package.
- Evidence: `VOC-134-EV-00`

Include positive, negative, authorization, failure, migration, accessibility, and
rollback coverage as applicable. Tests must not use secrets or production data.
