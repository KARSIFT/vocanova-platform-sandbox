# VOC-145 — Test Plan

## VOC-145-TEST-00 — Live `roles.yml` has the six exact authorized current bindings

- Covers: `VOC-145-AC-00`
- Preconditions: adopted path recorded (Path A default, or Path B only if
  adoption named `VOC-145-DEP-07`); fixture and infra `config/roles.yml`
  readable; no secrets or production data
- Procedure: Parse each of the six roles. Assert they equal the authorized
  current set from D01 (Path A) or D02 (Path B). Assert no additional
  active role lines exist. Assert every active value starts with `cursor/`.
  Assert header comments describe that authorized current set and do not
  describe the other path as current. Assert the caller mirrored fixture
  bytes match the independently reviewed infra merge for this file.
- Expected result: live current bindings are the authorized set, not an
  undeclared `d8720829…` leftover.
- Evidence: `VOC-145-EV-00`

## VOC-145-TEST-01 — Historical VOC-117 assertions are preserved

- Covers: `VOC-145-AC-01`
- Preconditions: same harness as TEST-00
- Procedure: Path A: assert `VOC117_BINDINGS` (or live equivalent) equals
  the `8993e867…` / VOC-142 DEP-06 lineup, including
  `effort=high,fast=false` for planner and all review roles, and that
  `prepare_cursor_model` still asserts those exact CLI strings rather than
  a prefix-only comparison that would bless any later effort/speed. Path B:
  assert a named historical constant still equals the VOC-117 /
  `8993e867…` bindings and a separate current-state constant equals the
  Path B lineup; assert tests do not claim VOC-117 originally required
  `xhigh` or `fast=true`. Assert `specs/changes/VOC-117-…/` is absent from
  the implementation diff.
- Expected result: the #1124 "rewrite VOC-117 expectations in the same
  sequence as the binding change" class cannot recur as acceptance proof.
- Evidence: `VOC-145-EV-00`

## VOC-145-TEST-02 — Parameterized Cursor identifiers survive without silent fallback

- Covers: `VOC-145-AC-02`
- Preconditions: `prepare_cursor_model` callable
- Procedure: For each authorized current binding, assert
  `prepare_cursor_model` returns the stored identifier with the `cursor/`
  prefix removed and effort/speed parameters intact. Assert
  `cursor/grok-4.6` and `cursor/grok-4.6[fast=false]` raise
  `CursorModelError`.
- Expected result: effort-qualified identifiers are required; effort-
  omitted forms fail closed.
- Evidence: `VOC-145-EV-00`

## VOC-145-TEST-03 — Missing credentials and unsupported prefixes fail closed

- Covers: `VOC-145-AC-02`
- Preconditions: subprocess harness; `CURSOR_API_KEY` removed from the
  child environment
- Procedure: Invoke `prepare_cursor_model.py --require-api-key` with an
  authorized current binding. Assert exit code 1, `missing_cursor_api_key`
  on stderr, and no unrelated sentinel or secret-like value printed. Assert
  unsupported prefixes such as `openai/codex-action` raise
  `CursorModelError`.
- Expected result: missing key and unsupported prefix cannot silently
  select another provider or model.
- Evidence: `VOC-145-EV-00`

## VOC-145-TEST-04 — README, CHANGELOG, and fixture README describe the authorized current set

- Covers: `VOC-145-AC-03`
- Preconditions: infra and caller fixture docs in the implementation diff
- Procedure: Assert infra `README.md`, infra `CHANGELOG.md`, and caller
  fixture README describe the authorized current six bindings. Path A:
  assert they do not present `xhigh` or `fast=true` retry as current, and
  that CHANGELOG records the governed restore of the ungoverned drift.
  Path B: assert they present the `xhigh` review lineup as current without
  claiming VOC-117 originally required it. Assert historical 2026-08-25
  high-effort CHANGELOG entries remain historical.
- Expected result: no current-state doc claims a lineup the live
  `roles.yml` does not implement.
- Evidence: `VOC-145-EV-00`

## VOC-145-TEST-05 — Pin and mirrored fixture bytes match the new infra merge

- Covers: `VOC-145-AC-04`
- Preconditions: new independently reviewed infra merge exists
- Procedure: Assert `PINNED_SHA.txt` strips to that 40-character merge SHA.
  Assert it is not `d8720829b176cf1287e633f9382989fc8f258105`. Assert
  SHA-256 of each changed mirrored file equals the same path at that merge.
  Discover every live governance test that asserts `CURRENT_PIN`, the live
  fixture pin, or mirrored-file hashes and assert it expects the new merge.
  Assert historical `AUTHORITATIVE_PIN` / issue-era constants and package
  evidence still retain their original merge, using separate current and
  historical constants where needed.
- Expected result: caller fixtures match the reconciled infra contract;
  unauthorized head is not the pin.
- Evidence: `VOC-145-EV-00`

## VOC-145-TEST-06 — Exhaustive source/pin search and current docs match the live contract

- Covers: `VOC-145-AC-03`
- Preconditions: caller docs and tests in the implementation diff
- Procedure: Exhaustively search tracked source/current docs for the six
  role literals, `xhigh`, `VOC117_BINDINGS`, `d8720829`, `8993e867`, and
  live reviewer-binding hard-codes; assert evidence disposes every match as
  updated, intentionally historical, or irrelevant. Assert
  `test_voc117_role_bindings.py`, `test_voc136_caller_replacement.py`, and
  `test_voc137_pr_sha_scan.py` (and any additional match) agree with the
  authorized current reviewer binding. Assert `git diff` against the
  implementation PR base does not name files under
  `specs/changes/VOC-142-…/`, `specs/changes/VOC-141-…/`,
  `specs/changes/VOC-140-…/`, or `specs/changes/VOC-117-…/`.
- Expected result: no current-state test or doc claims a contract the code
  does not implement; historical packages remain unmodified.
- Evidence: `VOC-145-EV-00`

## VOC-145-TEST-07 — Retry, exact-SHA review, and provider isolation remain

- Covers: `VOC-145-AC-05`
- Preconditions: implement, review, remediate, and merge-gate workflows
  readable in the mirrored fixture
- Procedure: Assert implementer still stops when `next_attempt > 2` / "no
  attempt 3". Assert review and merge-gate still bind `expected_head_sha`
  and reject mismatch. Assert no OpenAI/Codex active role binding is
  restored. Path B: assert `reviewer_fast_retry` `fast=true` does not add
  an extra implementer attempt or skip exact-SHA review. Assert no
  credential values are printed by the binding tests.
- Expected result: binding reconciliation does not weaken retry, review
  independence, or fail-closed provider isolation.
- Evidence: `VOC-145-EV-00`

## VOC-145-TEST-08 — Full suites, classification, and feasible exact-head evidence

- Covers: `VOC-145-AC-06`
- Preconditions: repair tracked and committed; exact implementation PR base
  and head
- Procedure: Run the commands in `implementation-plan.md` with exact
  base/head. Assert caller governance discovery, mirrored infra discovery,
  `validate-governance.sh`, `classify-change-risk.sh` (R4), and
  `git diff --check` pass. Assert `t00-evidence.md` names the
  implementation PR base, new infra merge, and authorized path, states that
  the live head is bound by the App-authored independent-review
  comment/check, and does not require a tracked file in the same commit to
  equal `HEAD`. Assert no App ID/private-key secret changed.
- Expected result: hosted-equivalent deterministic checks pass;
  classification is R4; exact-head binding remains Git-feasible.
- Evidence: `VOC-145-EV-00`

## VOC-145-TEST-09 — No snapshot-gap; #1120 is not a carrier; #1124 closes on metadata

- Covers: `VOC-145-AC-07`
- Preconditions: this package's implementation PR and `t00-evidence.md`
- Procedure: Assert no develop/main gap snapshot was committed. Assert no
  VOC-112 fixture recapture or #1120 resume helper was added. Assert
  evidence states that after this correction merges, ordinary later
  promotion uses existing release evaluation. Assert unauthorized head
  `d8720829…`, governed base `8993e867…`, and self-CI run `33443684483` are
  preserved as audit evidence, not restaged as this package's
  implementation work. Assert closed state is not treated as completion
  proof. Assert root issue #1124 is documented to close only after
  allowlisted metadata from a successful implement/release path exists.
  Assert no bootstrap exception was used.
- Expected result: repair is governed binding reconciliation plus ordinary
  release handoff, not a snapshot-then-check package and not an #1120
  bypass.
- Evidence: `VOC-145-EV-00`

Include positive, negative, authorization, failure, migration, accessibility, and
rollback coverage as applicable. Tests must not use secrets or production data.
