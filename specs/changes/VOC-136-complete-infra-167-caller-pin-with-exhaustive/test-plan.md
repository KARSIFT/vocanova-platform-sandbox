# VOC-136 — Test Plan

## VOC-136-TEST-00 — Pin equals exact infrastructure #167 merge

- Covers: `VOC-136-AC-00`
- Preconditions: live `tooling/governance/fixtures/karsift-ai-infra/PINNED_SHA.txt`
  and matching caller pin assertions; no secrets or production data
- Procedure: Assert `PINNED_SHA.txt` strips to
  `b263c0c110591cc798b89277dfc35542abb1597b`. Assert it does not equal
  `863fc1f35b1d35e4981a59166b0e939be1a2b681`. Assert it does not equal
  `8ce2b77a09a729e458a9f4cbea1ca26eb114d398`. Assert it does not equal
  `f3d79177bf8a9abe0dae550f39502165d494c576`. Assert
  `scripts/foundation/voc097-fixture-matrix.test.mjs`,
  `voc104-ready-for-review-reuse.test.mjs`, and
  `voc108-authoritative-lifecycle.test.mjs` pin literals match the same SHA.
  Assert live governance tests that previously hard-coded `863fc1f…` as the
  authoritative pin now equal `b263c0c…`.
- Expected result: the #164 pin class (`863fc1f…`), the #165 pin class
  (`8ce2b77…`), and the unmerged #166 pin class (`f3d791…`) fail the new
  assertion.
- Evidence: `VOC-136-EV-00`

## VOC-136-TEST-01 — Fixture mirrors the necessary #167 files by recorded SHA-256

- Covers: `VOC-136-AC-00`
- Preconditions: fixture tree after sync from merge
  `b263c0c110591cc798b89277dfc35542abb1597b`; recorded expected hashes
  committed in the caller regression
- Procedure: Assert SHA-256 of fixture `.github/workflows/ci.yml`,
  `.github/workflows/implement.yml`, `.github/workflows/release.yml`,
  `config/run-app-checks.sh`, `config/implementer_nested_checkout.py`,
  `tests/test_app_check_context.py`, `tests/test_release_policy.py`,
  `tests/test_voc121_implement_policy.py`,
  `tests/test_voc123_source_bundle.py`, and `CHANGELOG.md` equal the hashes
  recorded from infra merge `b263c0c…` (`VOC-136-D11`, reconfirmed at
  implementation). Assert the test does not open, clone, or require a
  machine-specific `/tmp` karsift-ai-infra checkout. Assert fixture
  `release.yml` contains exactly two occurrences of `Restore shared lifecycle
  policy after caller checkout`. Assert the mirrored policy tests include
  `test_app_check_context.py` and the #166 nested-checkout classifier tests.
- Expected result: the caller fixture is the #167 contract, not the #164 pin,
  and not a rewritten subset. CI can prove byte identity without a `/tmp`
  tree.
- Evidence: `VOC-136-EV-00`

## VOC-136-TEST-02 — Identify restores shared policy before validate-task

- Covers: `VOC-136-AC-01`
- Preconditions: mirrored fixture `release.yml`
- Procedure: Split fixture `release.yml` at the `identify` job. Assert
  `Checkout shared lifecycle policy` precedes `release-checkout-ref-runner.py`,
  which precedes `Checkout caller release state`, which precedes `Restore
  shared lifecycle policy after caller checkout`, which precedes
  `task-completion-runner.py validate-task`.
- Expected result: the identify missing-helper class cannot recur in the
  pinned fixture without failing this suite.
- Evidence: `VOC-136-EV-00`

## VOC-136-TEST-03 — Converge restores shared policy before validate-roster

- Covers: `VOC-136-AC-02`
- Preconditions: mirrored fixture `release.yml`
- Procedure: Split fixture `release.yml` at the `converge` job. Assert
  `Checkout caller release state` precedes `Restore shared lifecycle policy
  after caller checkout`, which precedes
  `task-completion-runner.py validate-roster`.
- Expected result: the latent converge checkout-lifetime defect cannot recur
  in the pinned fixture without failing this suite.
- Evidence: `VOC-136-EV-00`

## VOC-136-TEST-04 — Restore uses the immutable reusable-workflow revision without persisting credentials

- Covers: `VOC-136-AC-01`, `VOC-136-AC-02`
- Preconditions: mirrored fixture `release.yml`
- Procedure: For the restore step in both `identify` and `converge`, assert
  the restored block contains `repository: ${{ job.workflow_repository }}`,
  `ref: ${{ job.workflow_sha }}`, `path: karsift-ai-infra`, and
  `persist-credentials: false`. Assert it does not restore from
  `inputs.integration_branch` or an operator SHA input.
- Expected result: later helpers are the same immutable revision used to
  resolve the caller ref, and credentials are not persisted.
- Evidence: `VOC-136-EV-00`

## VOC-136-TEST-05 — #164 missing-develop, exact-SHA sync, and unique-develop contracts remain

- Covers: `VOC-136-AC-04`
- Preconditions: mirrored checkout-ref, branch-sync, and release-policy tests
  at the #167 pin
- Procedure: Reuse or extend existing fixture tests to assert existing
  `develop` is selected without reading `main`; absent `develop` falls back to
  live `main`; ambiguous or malformed refs fail closed; fixture `release.yml`
  still contains `Synchronize integration to the exact promotion merge` /
  `branch-sync-runner.py`; unique develop commits fail closed; and
  `-f sha="$CHECKED_HEAD_SHA"` is not treated as successful post-merge
  restoration. Assert live `pipeline.yml` still exposes
  `reconcile-production-change` under the 25-input limit without operator SHA
  inputs.
- Expected result: pinning #167 does not drop #164 fail-closed contracts.
- Evidence: `VOC-136-EV-00`

## VOC-136-TEST-06 — #166 helper-lifetime and nested-checkout classifier contracts

- Covers: `VOC-136-AC-03`
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
- Evidence: `VOC-136-EV-00`

## VOC-136-TEST-07 — #167 immutable PR-context validation without evidence fetch

- Covers: `VOC-136-AC-05`
- Preconditions: mirrored fixture `ci.yml`, `implement.yml`, and
  `config/run-app-checks.sh`; mirrored `tests/test_app_check_context.py`; no
  secrets or production data
- Procedure: Assert `run-app-checks.sh` contains `--pr-base-sha SHA
  --pr-head-sha SHA`, `git cat-file -e "${validation_base_sha}^{commit}"`,
  `git merge-base`, `validation_mode="pr-validation"`,
  `validation_mode="pr-ancestry"`, and `export PR_BASE_SHA=`. Assert it does
  not contain `git fetch`. Assert unchanged capture-fixture comparison
  selects `pr-validation` and add/modify/delete select `pr-ancestry`. Assert
  a `git diff` comparison error fails closed with
  `capture fixture comparison failed`. Assert fixture `ci.yml` caller
  checkout uses `fetch-depth: 0` and passes
  `github.event.pull_request.base.sha` / `head.sha`. Assert fixture
  `implement.yml` passes `--pr-base-sha "${{ steps.branch.outputs.integration_sha }}"`
  and `--pr-head-sha "$(git rev-parse HEAD)"` at least twice (pre-push and
  post-self-correction). The mirrored `tests/test_app_check_context.py` must
  remain byte-identical to merge `b263c0c…`.
- Expected result: the VOC-134 class — full PR-style checks under default
  `local` provenance on a checkout missing `f9d11e23…` — cannot recur in the
  pinned fixture without failing this suite. Evidence commits are never
  fetched.
- Evidence: `VOC-136-EV-00`

## VOC-136-TEST-08 — Existing recovery, review, retry, and credential contracts remain

- Covers: `VOC-136-AC-06`
- Preconditions: final live `pipeline.yml`, fixture `implement.yml` /
  `release.yml` / `ci.yml`, `config/roles.yml`
- Procedure: Reuse or extend existing policy tests to assert roster-marker
  validation, exact-head `--match-head-commit`, required-check recovery,
  App-token on merge/sync mutation and job-token on recovery reads,
  two-attempt implementer bound, Cursor Composer implementer, Cursor Grok
  reviewer, no OpenAI path, sanitized raw-error controls, and unchanged
  `config/roles.yml` bindings (`cursor/composer-2.5` implementer/escalation;
  `cursor/grok-4.6[effort=high,fast=false]` planner/reviewer/retry/plan
  reviewer). Confirm nested `karsift-ai-infra/.git` is not staged as a caller
  gitlink.
- Expected result: #167 caller consumption lands without weakening prior
  safety gates or changing model bindings.
- Evidence: `VOC-136-EV-00`

## VOC-136-TEST-09 — Caller suites, docs, and governance validation

- Covers: `VOC-136-AC-06`
- Preconditions: final changed file set after the regression is tracked and
  committed
- Procedure: Run the commands listed in `implementation-plan.md` with exact
  base/head where those scripts require it. Assert current-state fixture
  README/comments name pin `b263c0c…`, the post-caller-checkout restore, the
  post-implementer helper-lifetime contract, and the immutable PR-context
  contract, and that historical
  A-003/VOC-075/VOC-127/VOC-129/VOC-130/VOC-131/VOC-132/VOC-133/VOC-134/VOC-135
  records were not rewritten. Confirm the caller-owned fixture README was
  not replaced by the infrastructure repository README. Confirm `AGENTS.md`
  and `.agents/skills/vocanova-repo-navigator/SKILL.md` are not in the
  implementation diff against the protected comparison anchor. Confirm
  `t00-evidence.md` records the protected comparison anchor, the actual
  implementation PR base, the final infra merge, and the feasible exact-head
  binding contract, that no snapshot-gap task was added, no bootstrap
  exception was used, PR #1051 / PR #1056 / PR #1065 / PR #1070 / PR #1075
  were not reused, and VOC-132-T00 (#1059), VOC-133-T00 (#1063), VOC-134-T00
  (#1068), and VOC-135-T00 (#1073) were not redispatched. Confirm
  `git diff --check` is clean. Confirm `classify-change-risk.sh` reports R4
  for the implementation range.
- Expected result: suites pass; docs match the live contract; pin updates are
  exact-SHA; hashed VOC-112 sources and `package.json` are untouched;
  classification is R4.
- Evidence: `VOC-136-EV-00`

## VOC-136-TEST-10 — Eight-path no-change boundary remains byte-identical to the protected comparison anchor

- Covers: `VOC-136-AC-07`
- Preconditions: implementation PR against protected comparison anchor
  `b9e74fc2db4691c48c637639b265d527de9f4505`; no secrets or production data
- Procedure: Read the committed protected comparison-anchor SHA constant.
  Assert it equals `b9e74fc2db4691c48c637639b265d527de9f4505`. Resolve that
  exact commit object (`git cat-file -e <sha>^{commit}` or equivalent). For
  each of the following paths, resolve `git show <sha>:<path>` (or
  equivalent) and assert the working tree bytes are identical and that
  `git diff <sha>` does not name the path:
  - `scripts/foundation/fixtures/voc112-navigation-benchmark-traces.json`
  - `scripts/foundation/fixtures/voc112-skill-discovery-evidence.json`
  - `scripts/foundation/voc112-navigation-benchmark.test.mjs`
  - `scripts/foundation/voc112-navigation-benchmark-run.mjs`
  - `scripts/foundation/validate-workspace.mjs`
  - `AGENTS.md`
  - `.agents/skills/vocanova-repo-navigator/SKILL.md`
  - `package.json`
  If the commit object or any named path cannot be resolved, fail—do not
  skip. Do not fall back to `develop` or `origin/develop` after the constant
  is selected. Assert both JSON files still contain `subject_revision`
  `f9d11e232a07c7d7a9c433d02c9267912543ba10`. Assert the provenance test still
  contains the fail-closed `local` mode requirement that a full checkout must
  already contain the captured commit, and does not treat a missing subject
  in a full checkout as implicit `squash-safe-push`. The regression must fail
  if any named path is retargeted, recaptured, weakened, or otherwise
  rewritten, including the JSON class on #1051, the provenance-test class on
  #1056, the runner fetch class on #1070 attempt 1, and the
  `validate-workspace.mjs` hydrate class on #1070 attempt 2.
- Expected result: the VOC-130 JSON class, the VOC-131 provenance-test class,
  the VOC-134 runner/hydrate classes, and a moving-`develop`-ref comparison
  cannot recur unnoticed; CI does not enter `pr-ancestry` merely because
  those fixtures changed.
- Evidence: `VOC-136-EV-00`

## VOC-136-TEST-11 — package.json is comparison-anchor-identical and named hydration helpers are absent

- Covers: `VOC-136-AC-08`
- Preconditions: implementation PR against the protected comparison anchor;
  no secrets or production data
- Procedure: Resolve `package.json` at the protected comparison anchor and
  assert working-tree bytes are identical and that `git diff <sha>` does not
  name `package.json`. If the path cannot be resolved, fail—do not skip.
  Assert `package.json` does not set `VOC112_CAPTURE_PROVENANCE_MODE`, does
  not point `test` at a capture-commit fetch helper, and still invokes
  `node --test scripts/foundation/*.test.mjs` without a provenance-mode
  prefix. Assert `scripts/foundation/ensure-voc112-capture-commits.mjs` is
  absent. Assert `scripts/foundation/hydrate-voc112-git-objects.mjs` is
  absent. Assert no evidence-stamping helper mutates `t00-evidence.md` at
  test or runtime. Assert no test rewrites evidence in memory so that a
  prior failed head can pass.
- Expected result: the VOC-133 package.json wrapper class and the VOC-134
  named hydrate-helper class cannot recur unnoticed. This test is necessary
  but not sufficient; `VOC-136-TEST-12` covers relocated equivalents.
- Evidence: `VOC-136-EV-00`

## VOC-136-TEST-12 — Exhaustive caller-diff scan includes caller tests and rejects relocated bypasses

- Covers: `VOC-136-AC-08`, `VOC-136-AC-11`
- Preconditions: implementation PR diff against the protected comparison
  anchor after the regression is tracked and committed; no secrets or
  production data
- Procedure: Enumerate every added or modified path in `git diff --name-only
  <protected-comparison-anchor>`. For every changed executable path
  (`scripts/**`, `package.json`, added/changed `*.mjs` / `*.js` / `*.sh` /
  `*.py` under the caller tree outside the mirrored infra fixture, and any
  other newly executable caller file), read the file and fail if it:
  1. invokes Git fetch for VOC-112 capture subjects or evidence objects;
  2. sets `VOC112_CAPTURE_PROVENANCE_MODE`, `PR_BASE_SHA`, or `PR_HEAD_SHA`
     around tests, `pnpm test`, or `validate-workspace`;
  3. hydrates, materializes, or otherwise plants evidence Git objects;
  4. adds an import-time or pre-test side effect that fetches or stamps
     evidence;
  5. skips or otherwise makes default `local` fail-closed unreachable.
  Assert the scanned set **includes** every added/changed path under
  `tooling/governance/tests/` and includes this regression's own module.
  Assert the implementation does **not** list `tooling/governance/tests/`
  (or this module, or another executable directory) in
  `SCAN_EXCLUDE_PREFIXES` or an equivalent wholesale exclusion. The only
  allowed executable-tree exclusion is
  `tooling/governance/fixtures/karsift-ai-infra/**`. A new helper under any
  name that performs the VOC-134 attempt-2 hydrate, or an import-time fetch
  relocated from `voc112-navigation-benchmark-run.mjs` into another caller
  script including a `tooling/governance/tests/*.py` module, must fail this
  test. Mirrored fixture copies of infra `run-app-checks.sh` that export
  provenance for an immutable PR pair are the #167 contract, not this
  failure class. If any of the eight no-change paths appears in the diff,
  fail. Test/assertion literals in committed executable files must be
  source-safe (non-contiguous) so this scan does not self-fail on the
  regression module after it is tracked.
- Expected result: relocating a hydration bypass to a new filename, wrapping
  tests from an unexpected script, adding an import side effect inside
  `tooling/governance/tests/*.py`, or excluding that directory from the scan
  cannot pass. The VOC-135 attempt-2 `SCAN_EXCLUDE_PREFIXES` class fails this
  test.
- Evidence: `VOC-136-EV-00`

## VOC-136-TEST-13 — Exact-revision evidence is fail-closed and Git-feasible

- Covers: `VOC-136-AC-09`
- Preconditions: `t00-evidence.md` and the implementation PR diff against the
  protected comparison anchor
- Procedure: Assert `t00-evidence.md` records the protected comparison
  anchor `b9e74fc2db4691c48c637639b265d527de9f4505`, the actual
  implementation PR base as a distinct 40-character SHA, and the final infra
  merge `b263c0c110591cc798b89277dfc35542abb1597b`. Assert it records the
  reconfirmed SHA-256 hashes, complete-diff scan scope (including that
  `tooling/governance/tests/**` was scanned), negative-case results, and the
  validation commands. Assert it states that the live implementation head is
  bound by the App-authored independent-review comment/check, that
  merge-gate must reject any mismatch, and that post-merge/root-issue audit
  may record the reviewed head and merge SHA. Assert it does **not** require
  a tracked file in the same commit to equal `HEAD` or to contain that
  commit's own SHA. Assert no test or helper rewrites `t00-evidence.md` on
  disk or in memory at test time. If evidence claims that a protected
  no-change path was reverted, unchanged, or restored, assert that path is
  absent from `git diff` against the protected comparison anchor. Assert
  evidence does not contain raw provider responses, credentials, or a full
  CI log.
- Expected result: evidence truthfully describes the protected comparison
  anchor, the implementation PR base, and the consumed infra merge, and
  exact-head binding remains fail-closed without a Git-impossible self-SHA.
- Evidence: `VOC-136-EV-00`

## VOC-136-TEST-14 — Replacement carrier is VOC-136; no snapshot-gap; VOC-127 through VOC-135 are not re-implemented

- Covers: `VOC-136-AC-10`
- Preconditions: this package's implementation PR and `t00-evidence.md`
- Procedure: Assert the implementation PR targets `develop` from a new
  VOC-136 branch, `Closes` only its own VOC-136 task issue, and does not
  `Closes` VOC-127, VOC-129, VOC-130, VOC-131, VOC-132, VOC-133, VOC-134,
  VOC-135, root #1071, or root #1076. Assert evidence records that PR #1051,
  PR #1056, PR #1065, PR #1070, and PR #1075 were not reused, merged,
  cherry-picked, or modified; that VOC-132-T00 (#1059), VOC-133-T00 (#1063),
  VOC-134-T00 (#1068), and VOC-135-T00 (#1073) were not redispatched; that
  VOC-129 PR #1046 was not re-implemented; that no VOC-127, VOC-129,
  VOC-130, VOC-131, VOC-132, VOC-133, VOC-134, or VOC-135 completion marker
  was manufactured; and that no develop/main gap snapshot was committed.
  After promotion, evidence records ordinary release/sync/closure for
  VOC-127, VOC-130, VOC-131, VOC-132, VOC-133, VOC-134, VOC-135, and this
  package, with closed state not treated as completion proof, and records
  that cleanup preserved unrelated VOC-128 and user worktrees.
- Expected result: repair is a new VOC-136 carrier, not a VOC-130 / VOC-131 /
  VOC-132 / VOC-133 / VOC-134 / VOC-135 retry, not a VOC-129 rewrite, and
  not a snapshot-then-check package.
- Evidence: `VOC-136-EV-00`

## VOC-136-TEST-15 — Negative unit cases prove the scanner rejects relocated bypasses

- Covers: `VOC-136-AC-08`, `VOC-136-AC-11`
- Preconditions: scanner function callable from the caller governance test
  suite; no secrets or production data; contiguous forbidden payloads
  reconstructed only in memory (or in non-executable, non-scanned test data)
- Procedure: Reconstruct contiguous forbidden payloads at test runtime from
  non-contiguous source-safe parts. Assert the scanner rejects at least:
  1. an import-time or pre-test `git fetch` for VOC-112/evidence;
  2. a hydrate/materialize helper under an arbitrary new filename;
  3. shell, Node, and Python forms that set `VOC112_CAPTURE_PROVENANCE_MODE`;
  4. relevant `PR_BASE_SHA` / `PR_HEAD_SHA` setters around tests or
     `validate-workspace`;
  5. an import-time side effect inside an added
     `tooling/governance/tests/*.py` module;
  6. a wrapper or skip that makes default local fail-closed unreachable.
  Assert each rejection is independent of the payload filename matching a
  previously banned helper. Assert the scanner still inspects a synthetic
  path under `tooling/governance/tests/` rather than skipping that prefix.
- Expected result: the VOC-134 relocated-hydrate class and the VOC-135
  import-time-in-caller-tests class cannot pass merely because the filename
  or directory changed.
- Evidence: `VOC-136-EV-00`

## VOC-136-TEST-16 — Benign mentions do not false-positive; tracked committed regression passes

- Covers: `VOC-136-AC-11`
- Preconditions: regression module tracked and committed; no secrets or
  production data
- Procedure: Assert a benign payload whose only matches are assertion or
  test-data mentions, reconstructed from source-safe parts, does not fail
  the scanner. Assert the committed regression module itself, once tracked,
  is scan-clean (VOC-135 attempt-1 class). Assert `git ls-files` names the
  regression module and that running the scanner over the implementation
  diff against the protected comparison anchor exits 0 on the real tree.
  Assert the suite is not skipped or xfailed when the file is tracked.
- Expected result: self-scanning the committed regression does not
  false-positive, and a pass obtained only while the module was untracked
  is not accepted.
- Evidence: `VOC-136-EV-00`

## VOC-136-TEST-17 — Implementation PR base is recorded separately; unrelated develop movement fails closed

- Covers: `VOC-136-AC-07`
- Preconditions: `t00-evidence.md` and the protected comparison-anchor
  constant; no secrets or production data
- Procedure: Assert `t00-evidence.md` records a 40-character implementation
  PR base SHA that is distinct from the live implementation HEAD (the
  implementation commit cannot contain its own SHA as a required value).
  Assert the protected comparison anchor remains
  `b9e74fc2db4691c48c637639b265d527de9f4505`. Assert the eight no-change
  paths are compared to that anchor, not to a moving `develop` ref. Assert
  evidence states that this package's own plan/adoption/roster commits after
  the anchor do not count as protected-file drift. Assert a helper or
  documented pre-edit check fails closed if live `develop` contains
  unrelated/material tree changes outside this package directory,
  already-merged VOC-135 package adoption/roster metadata, and other
  governance-only commits that cannot affect the eight protected paths or
  the live pin/fixture.
- Expected result: VOC-135-D15's "fail closed if develop moved at all" rule
  is not reimposed in a way that makes this package unimplementable after
  its plan merges, and unrelated product or pin drift cannot ride along.
- Evidence: `VOC-136-EV-00`

Include positive, negative, authorization, failure, migration, accessibility, and
rollback coverage as applicable. Tests must not use secrets or production data.
