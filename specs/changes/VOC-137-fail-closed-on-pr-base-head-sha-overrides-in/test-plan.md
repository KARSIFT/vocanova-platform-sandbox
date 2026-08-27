# VOC-137 — Test Plan

## VOC-137-TEST-00 — PR SHA detection no longer depends on filename

- Covers: `VOC-137-AC-00`, `VOC-137-AC-03`
- Preconditions: scanner function callable from the caller governance test
  suite; no secrets or production data
- Procedure: Assert `scan_changed_path_for_bypasses` invokes PR base/head SHA
  detection for scannable paths whose names do not contain
  `validate-workspace` and do not end in `.test.mjs`, including
  `scripts/arbitrary-wrapper.sh`, an arbitrary `*.mjs` path, an arbitrary
  `*.py` path, and `package.json` if that file assigns those variables.
  Assert the implementation source no longer gates `PR_SHA_SET_PATTERN` on
  `"validate-workspace" in relative or relative.endswith(".test.mjs")`.
  Assert `SCAN_EXCLUDE_PREFIXES` equals the fixture-mirror prefix only.
- Expected result: renaming a wrapper cannot hide a PR SHA assignment.
- Evidence: `VOC-137-EV-00`

## VOC-137-TEST-01 — Scanner and tests are source-safe and scan-clean when tracked

- Covers: `VOC-137-AC-01`
- Preconditions: scanner and test modules tracked and committed
- Procedure: Assert `PR_SHA_SET_PATTERN` is constructed from non-contiguous
  parts (`_PR_BASE_ENV` / `_PR_HEAD_ENV` or equivalent). Assert
  `test_voc136_caller_replacement.py` no longer contains a contiguous
  `export PR_BASE_SHA=` literal. Assert the committed scanner module, the
  VOC-136 replacement test module, and this package's new test module (if
  any) are scan-clean. Assert `git ls-files` names those modules and that
  scanning the implementation diff does not false-positive on them. Assert
  the suite is not skipped or xfailed when the files are tracked.
- Expected result: VOC-135 attempt-1 self-scan class cannot recur; a pass
  obtained only while untracked is not accepted.
- Evidence: `VOC-137-EV-00`

## VOC-137-TEST-02 — Shell arbitrary-wrapper negative case

- Covers: `VOC-137-AC-00`, `VOC-137-AC-02`
- Preconditions: contiguous payload reconstructed only in memory from
  source-safe parts
- Procedure: Reconstruct
  `export PR_BASE_SHA=deadbeef\npnpm test\n` (the issue example) and assert
  `scan_changed_path_for_bypasses("scripts/arbitrary-wrapper.sh", payload)`
  raises. Assert the relative path does not contain `validate-workspace` and
  does not end in `.test.mjs`.
- Expected result: the issue counterexample now fails closed.
- Evidence: `VOC-137-EV-00`

## VOC-137-TEST-03 — Node arbitrary-wrapper negative case

- Covers: `VOC-137-AC-00`, `VOC-137-AC-02`
- Preconditions: contiguous payload reconstructed only in memory from
  source-safe parts
- Procedure: Reconstruct a Node payload that assigns `process.env.PR_HEAD_SHA`
  before validation/test execution. Assert the scanner raises for an
  arbitrary relative path such as `scripts/arbitrary-head-wrapper.mjs` that
  does not contain `validate-workspace` and does not end in `.test.mjs`.
- Expected result: Node wrappers cannot hide PR head SHA assignment by
  renaming away from `validate-workspace`.
- Evidence: `VOC-137-EV-00`

## VOC-137-TEST-04 — Python arbitrary-wrapper negative case

- Covers: `VOC-137-AC-00`, `VOC-137-AC-02`
- Preconditions: contiguous payload reconstructed only in memory from
  source-safe parts
- Procedure: Reconstruct a Python payload that assigns
  `os.environ["PR_BASE_SHA"]` before validation/test execution. Assert the
  scanner raises for an arbitrary relative path such as
  `scripts/arbitrary_wrapper.py`.
- Expected result: Python wrappers cannot hide PR base SHA assignment by
  filename.
- Evidence: `VOC-137-EV-00`

## VOC-137-TEST-05 — Added or modified `*.py` outside the fixture mirror is scanned

- Covers: `VOC-137-AC-00`, `VOC-137-AC-02`
- Preconditions: `should_scan_path` and the live complete-diff scan
- Procedure: Assert `should_scan_path` is true for an added `*.py` path
  outside `tooling/governance/fixtures/karsift-ai-infra/`, including
  `tooling/governance/tests/` modules and `scripts/*.py`. Assert
  `should_scan_path` is false for
  `tooling/governance/fixtures/karsift-ai-infra/tests/test_app_check_context.py`.
  Assert the live complete-diff scan includes every added/changed `*.py`
  outside the fixture mirror in this implementation range. Assert an
  in-memory PR SHA assignment payload at such a `*.py` path raises.
- Expected result: Python executables outside the fixture mirror cannot opt
  out of PR SHA detection.
- Evidence: `VOC-137-EV-00`

## VOC-137-TEST-06 — Benign mentions, pattern construction, and fixture mirror do not false-positive

- Covers: `VOC-137-AC-01`, `VOC-137-AC-03`
- Preconditions: scanner callable; fixture bytes unchanged
- Procedure: Assert a benign payload whose only matches are discussion or
  test-data mentions, reconstructed from source-safe parts, does not fail the
  scanner. Assert scanning the scanner module itself (tracked) does not fail.
  Assert `should_scan_path("tooling/governance/fixtures/karsift-ai-infra/config/run-app-checks.sh")`
  is false. Assert the live complete-diff scan does not raise on that fixture
  file even though it contains legitimate `export PR_BASE_SHA=`.
- Expected result: source-safe construction and the exact fixture-mirror
  exclusion remain; no wholesale tests/ exclusion is added.
- Evidence: `VOC-137-EV-00`

## VOC-137-TEST-07 — Pin and mirrored fixture bytes are unchanged

- Covers: `VOC-137-AC-04`
- Preconditions: live fixture tree after VOC-136; no secrets or production data
- Procedure: Assert `PINNED_SHA.txt` strips to
  `b263c0c110591cc798b89277dfc35542abb1597b`. Assert SHA-256 of each
  VOC-136-D11 mirrored file still equals the recorded hash. Assert
  `git diff` against the implementation PR base does not name any path under
  `tooling/governance/fixtures/karsift-ai-infra/`.
- Expected result: this correction does not retarget or rewrite the #167 pin.
- Evidence: `VOC-137-EV-00`

## VOC-137-TEST-08 — Eight no-change paths, roles, and no OpenAI route

- Covers: `VOC-137-AC-05`
- Preconditions: implementation PR against protected comparison anchor
  `b9e74fc2db4691c48c637639b265d527de9f4505`
- Procedure: For each of the eight VOC-112 no-change paths, assert working-tree
  bytes equal `git show b9e74fc2…:path` and that `git diff b9e74fc2…` does
  not name the path. If the commit or path cannot resolve, fail—do not skip.
  Assert JSON `subject_revision` remains
  `f9d11e232a07c7d7a9c433d02c9267912543ba10`. Assert fixture `config/roles.yml`
  still binds implementer/escalation `cursor/composer-2.5` and
  planner/reviewer/retry/plan_reviewer
  `cursor/grok-4.6[effort=high,fast=false]`. Assert no OpenAI route is added.
- Expected result: VOC-112 identity and current KARSIFT roles remain.
- Evidence: `VOC-137-EV-00`

## VOC-137-TEST-09 — Full suites, classification, and direct reproduction

- Covers: `VOC-137-AC-06`
- Preconditions: regression tracked and committed; exact implementation PR
  base and head
- Procedure: Run the commands in `implementation-plan.md` with exact
  base/head. Assert caller governance discovery, mirrored infra discovery,
  `validate-governance.sh`, `classify-change-risk.sh` (R4), and
  `git diff --check` pass. Assert the direct reproduction in TEST-02 is
  included in the recorded suite. Confirm `t00-evidence.md` records those
  commands after commit.
- Expected result: hosted-equivalent deterministic checks pass; classification
  is R4; the arbitrary-wrapper example fails closed.
- Evidence: `VOC-137-EV-00`

## VOC-137-TEST-10 — Exact-revision evidence is fail-closed and Git-feasible

- Covers: `VOC-137-AC-07`
- Preconditions: `t00-evidence.md` and the implementation PR
- Procedure: Assert `t00-evidence.md` records a 40-character implementation
  PR base distinct from a self-referential live-HEAD requirement, the pin
  freeze, scan-scope change, negative-case results, and the contract that the
  live head is bound by the App-authored independent-review comment/check.
  Assert it states that independent review must explicitly evaluate the
  arbitrary-filename shell/Node/Python cases and benign controls. Assert it
  does not require a tracked file in the same commit to equal `HEAD`. Assert
  no test rewrites this file at runtime.
- Expected result: exact-head binding remains fail-closed without a
  Git-impossible self-SHA.
- Evidence: `VOC-137-EV-00`

## VOC-137-TEST-11 — VOC-136 package records are not rewritten

- Covers: `VOC-137-AC-08`
- Preconditions: implementation PR diff
- Procedure: Assert `git diff` against the implementation PR base does not
  name any file under
  `specs/changes/VOC-136-complete-infra-167-caller-pin-with-exhaustive/`.
  Assert this package's PR `Closes` only its own VOC-137 task issue.
  Assert evidence names PR #1080, reviewed head
  `5d0c2350ab9a20ace586eaadd1169203140ffad0`, and develop merge
  `0cee20c87e0411a95f368d2b7d39ac2bb118dfb8` as preserved audit references
  and does not claim those comments as this package's review.
- Expected result: VOC-136 records remain historical evidence.
- Evidence: `VOC-137-EV-00`

## VOC-137-TEST-12 — No snapshot-gap; ordinary release handoff

- Covers: `VOC-137-AC-09`
- Preconditions: this package's implementation PR and `t00-evidence.md`
- Procedure: Assert no develop/main gap snapshot was committed. Assert
  evidence states that after this correction merges, ordinary release
  evaluation (or `reconcile-release`) promotes outstanding completed
  packages including VOC-136 and this package, converges `develop` to the
  exact main merge SHA, and does not keep staging scheduled for
  tree-equivalent sync. Assert canceled release runs `33113425829` and
  `33113547909` and draft PR #1082 are not treated as this package's
  implementation work. Assert closed state is not treated as completion
  proof.
- Expected result: repair is a scanner correction plus ordinary release
  handoff, not a snapshot-then-check package.
- Evidence: `VOC-137-EV-00`

Include positive, negative, authorization, failure, migration, accessibility, and
rollback coverage as applicable. Tests must not use secrets or production data.
