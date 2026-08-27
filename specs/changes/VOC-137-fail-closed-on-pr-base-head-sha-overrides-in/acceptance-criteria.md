# VOC-137 — Acceptance Criteria

## VOC-137-AC-00 — PR base/head SHA override detection is filename-independent

- Requirement source: `VOC-137-D01`, `VOC-137-D06`
- Tasks: `VOC-137-T00`
- Tests: `VOC-137-TEST-00`, `VOC-137-TEST-02`, `VOC-137-TEST-03`, `VOC-137-TEST-04`, `VOC-137-TEST-05`
- Evidence: `VOC-137-EV-00`
- Result: pending

`PR_SHA_SET_PATTERN` (or its successor in the same scanner) is applied to every
added or modified scannable caller executable outside
`tooling/governance/fixtures/karsift-ai-infra/**`. Detection does not require
the relative filename to contain `validate-workspace` or to end in
`.test.mjs`. An added `scripts/arbitrary-wrapper.sh` that assigns
`PR_BASE_SHA` before `pnpm test` fails closed. Equivalent Node and Python
assignment around validation/test execution fails closed under arbitrary
names, including an added or modified `*.py` path outside the fixture mirror.

## VOC-137-AC-01 — Source-safe scanner construction prevents self-rejection

- Requirement source: `VOC-137-D02`, `VOC-137-D06`
- Tasks: `VOC-137-T00`
- Tests: `VOC-137-TEST-01`, `VOC-137-TEST-06`
- Evidence: `VOC-137-EV-00`
- Result: pending

The scanner and its tests represent pattern definitions and assertion literals
as non-contiguous source-safe values so self-scanning a tracked committed
module does not false-positive. Existing contiguous literals that would match
the widened scan, including
`tooling/governance/tests/test_voc136_caller_replacement.py`'s
`export PR_BASE_SHA=` fixture assertion, are made source-safe in the same
task. Actual semantic setters reconstructed in memory as contiguous payloads
still fail. The regression passes after it is tracked and committed, not
merely while untracked.

## VOC-137-AC-02 — Arbitrary-filename negative cases cover shell, Node, and Python

- Requirement source: `VOC-137-D03`
- Tasks: `VOC-137-T00`
- Tests: `VOC-137-TEST-02`, `VOC-137-TEST-03`, `VOC-137-TEST-04`, `VOC-137-TEST-05`
- Evidence: `VOC-137-EV-00`
- Result: pending

Deterministic negative unit cases reconstruct contiguous forbidden payloads
only in memory (or in non-executable, non-scanned test data) and prove
rejection of at least:

- a shell wrapper assigning `PR_BASE_SHA` before `pnpm test` under an
  arbitrary name that does not contain `validate-workspace` and does not end
  in `.test.mjs`, including the issue example path
  `scripts/arbitrary-wrapper.sh`;
- a Node wrapper assigning `process.env.PR_HEAD_SHA` before validation/test
  execution under an arbitrary name that does not contain
  `validate-workspace` and does not end in `.test.mjs`;
- a Python wrapper assigning `os.environ["PR_BASE_SHA"]` before
  validation/test execution under an arbitrary name;
- an added or modified `*.py` file outside the fixture mirror.

Rejection is independent of the payload filename matching a previously banned
helper or the VOC-136 `validate-workspace-wrapper.mjs` case.

## VOC-137-AC-03 — Positive controls do not restore wholesale exclusions

- Requirement source: `VOC-137-D04`, `VOC-137-D06`
- Tasks: `VOC-137-T00`
- Tests: `VOC-137-TEST-06`, `VOC-137-TEST-00`
- Evidence: `VOC-137-EV-00`
- Result: pending

Benign textual discussion, source-safe pattern construction, and the exact
excluded infra fixture mirror do not trigger false positives. Mirrored
`tooling/governance/fixtures/karsift-ai-infra/config/run-app-checks.sh` that
legitimately exports an immutable PR pair remains excluded as the #167
contract, not this failure class. `SCAN_EXCLUDE_PREFIXES` (or equivalent)
does not list `tooling/governance/tests/`, this regression's own module,
`scripts/`, or another executable directory wholesale. The only allowed
executable-tree exclusion remains
`tooling/governance/fixtures/karsift-ai-infra/**`.

## VOC-137-AC-04 — #167 pin and mirrored fixture bytes are unchanged

- Requirement source: `VOC-137-D05`, `VOC-137-D07`
- Tasks: `VOC-137-T00`
- Tests: `VOC-137-TEST-07`
- Evidence: `VOC-137-EV-00`
- Result: pending

`tooling/governance/fixtures/karsift-ai-infra/PINNED_SHA.txt` still equals
`b263c0c110591cc798b89277dfc35542abb1597b`. No file under
`tooling/governance/fixtures/karsift-ai-infra/` appears in the implementation
diff. Issue #1083's "eight caller workflow/fixture paths" is bound to the
complete VOC-136-D02/D11 mirrored set plus `PINNED_SHA.txt`, including
`.github/workflows/ci.yml`, `.github/workflows/implement.yml`,
`.github/workflows/release.yml`, `config/run-app-checks.sh`,
`config/implementer_nested_checkout.py`, `tests/test_app_check_context.py`,
`tests/test_release_policy.py`, `tests/test_voc121_implement_policy.py`,
`tests/test_voc123_source_bundle.py`, and fixture `CHANGELOG.md`. Recorded
SHA-256 hashes from VOC-136-D11 remain the committed expected values.

## VOC-137-AC-05 — VOC-112 no-change paths, roles, and credentials remain

- Requirement source: `VOC-137-D07`, `VOC-137-D08`
- Tasks: `VOC-137-T00`
- Tests: `VOC-137-TEST-08`
- Evidence: `VOC-137-EV-00`
- Result: pending

All eight named VOC-112 no-change paths remain byte-identical to protected
comparison anchor `b9e74fc2db4691c48c637639b265d527de9f4505` and are absent
from the implementation diff against that SHA. `config/roles.yml` (caller
fixture copy) is unchanged: implementer and escalation `cursor/composer-2.5`;
planner, reviewer, reviewer_fast_retry, and plan_reviewer
`cursor/grok-4.6[effort=high,fast=false]`. No OpenAI route is added. No
OpenAI credential is requested.

## VOC-137-AC-06 — Deterministic suites and the arbitrary-wrapper reproduction pass

- Requirement source: `VOC-137-D09`
- Tasks: `VOC-137-T00`
- Tests: `VOC-137-TEST-09`
- Evidence: `VOC-137-EV-00`
- Result: pending

After the regression is tracked and committed, the complete caller governance
suite, mirrored infra suite, risk classification, governance validation, and
`git diff --check` run against the exact implementation PR base and head. A
direct reproduction using the issue's `scripts/arbitrary-wrapper.sh` payload
fails closed. `classify-change-risk.sh` reports R4 for the implementation
range. A pass obtained only while untracked is not acceptance.

## VOC-137-AC-07 — Exact-revision review evaluates the arbitrary-filename cases

- Requirement source: `VOC-137-D10`
- Tasks: `VOC-137-T00`
- Tests: `VOC-137-TEST-10`
- Evidence: `VOC-137-EV-00`
- Result: pending

Independent exact-revision review binds the live implementation PR head and
explicitly evaluates the arbitrary-filename shell/Node/Python negative cases
and benign controls before merge. Merge-gate rejects any mismatch between
that bound SHA and the head it would merge. The implementer does not approve
or merge its own work. Committed evidence does not require a commit to
contain its own SHA.

## VOC-137-AC-08 — VOC-136 merge, completion, and review records are preserved

- Requirement source: `VOC-137-D11`
- Tasks: `VOC-137-T00`
- Tests: `VOC-137-TEST-11`
- Evidence: `VOC-137-EV-00`
- Result: pending

Existing VOC-136 merge/completion/review records remain audit evidence.
`specs/changes/VOC-136-complete-infra-167-caller-pin-with-exhaustive/` is not
rewritten. PR #1080, reviewed head
`5d0c2350ab9a20ace586eaadd1169203140ffad0`, develop merge
`0cee20c87e0411a95f368d2b7d39ac2bb118dfb8`, and the three named #1080
comments are not manufactured, replaced, or silently restated as this
package's review. This package's implementation PR `Closes` only its own
VOC-137 task issue.

## VOC-137-AC-09 — Ordinary release proceeds after this correction merges; no snapshot-gap

- Requirement source: `VOC-137-D12`
- Tasks: `VOC-137-T00`
- Tests: `VOC-137-TEST-12`
- Evidence: `VOC-137-EV-00`
- Result: pending

No snapshot of the current develop/main gap is committed. After exact-SHA
review and merge into `develop`, ordinary release evaluation (or
`reconcile-release`) reconciles outstanding completed packages, including
already-merged VOC-136 and this package, promotes `develop` to `main`,
deploys where applicable, and converges `develop` to the exact resulting main
merge SHA (0 ahead / 0 behind, identical tree) without a tree-equivalent
staging loop. Closed state alone is not completion proof. Canceled release
runs `33113425829` and `33113547909` and draft release PR #1082 are not
this package's implementation diff.

Acceptance criteria must be observable, stable, security-aware, and bidirectionally
traceable to requirements, tasks, tests, and evidence.
