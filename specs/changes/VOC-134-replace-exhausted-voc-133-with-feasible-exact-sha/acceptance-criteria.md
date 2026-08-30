# VOC-134 — Acceptance Criteria

## VOC-134-AC-00 — Caller fixture and pin equal exact infrastructure #166 merge

- Requirement source: `VOC-134-D02`, `VOC-134-D11`
- Tasks: `VOC-134-T00`
- Tests: `VOC-134-TEST-00`, `VOC-134-TEST-01`
- Evidence: `VOC-134-EV-00`
- Result: pending

`tooling/governance/fixtures/karsift-ai-infra/PINNED_SHA.txt` equals
`f3d79177bf8a9abe0dae550f39502165d494c576`. Every live caller pin assertion
equals that same SHA. The pin does not equal
`863fc1f35b1d35e4981a59166b0e939be1a2b681` and does not equal
`8ce2b77a09a729e458a9f4cbea1ca26eb114d398`. Fixture
`.github/workflows/implement.yml`, `.github/workflows/release.yml`,
`config/implementer_nested_checkout.py`, `tests/test_release_policy.py`,
`tests/test_voc121_implement_policy.py`, `tests/test_voc123_source_bundle.py`,
and fixture `CHANGELOG.md` match the recorded SHA-256 hashes of those paths at
infra merge `f3d791…`. Test-time comparison does not depend on a
machine-specific `/tmp` checkout.

## VOC-134-AC-01 — Identify restores shared policy after caller checkout before validate-task

- Requirement source: `VOC-134-D03`
- Tasks: `VOC-134-T00`
- Tests: `VOC-134-TEST-02`, `VOC-134-TEST-04`
- Evidence: `VOC-134-EV-00`
- Result: pending

Fixture `identify` contains `Restore shared lifecycle policy after caller
checkout` after `Checkout caller release state` and before
`task-completion-runner.py validate-task`. The restore uses
`repository: ${{ job.workflow_repository }}`,
`ref: ${{ job.workflow_sha }}`, `path: karsift-ai-infra`, and
`persist-credentials: false`. The #164 checkout-ref resolver still runs after
the first shared-policy checkout and before the caller checkout.

## VOC-134-AC-02 — Converge restores shared policy after caller checkout before validate-roster

- Requirement source: `VOC-134-D03`
- Tasks: `VOC-134-T00`
- Tests: `VOC-134-TEST-03`, `VOC-134-TEST-04`
- Evidence: `VOC-134-EV-00`
- Result: pending

Fixture `converge` contains `Restore shared lifecycle policy after caller
checkout` after `Checkout caller release state` and before
`task-completion-runner.py validate-roster`. The restore uses the same
immutable reusable-workflow revision as `identify`, with credentials not
persisted. The missing-validator safe-no-op class from run `33066533397`
cannot recur unnoticed in the pinned fixture.

## VOC-134-AC-03 — #166 post-implementer nested-checkout contract is present

- Requirement source: `VOC-134-D05`
- Tasks: `VOC-134-T00`
- Tests: `VOC-134-TEST-06`
- Evidence: `VOC-134-EV-00`
- Result: pending

Fixture `implement.yml` copies lifecycle helpers to immutable `/tmp` paths
before the unrestricted model. After the model, nested-checkout classification
uses the preserved helper. An absent nested checkout means no infrastructure
source carrier while caller changes continue. A plain subdirectory inheriting
caller Git, a non-directory, or a symlink fails closed. A distinct nested Git
checkout preserves exact-head bundle, ancestry, remote, and lease
publication. The `cp: cannot stat karsift-ai-infra/config/run-app-checks.sh`
class from run `33079499176` cannot recur unnoticed in the pinned fixture.

## VOC-134-AC-04 — #164 contracts remain after the #166 pin

- Requirement source: `VOC-134-D04`
- Tasks: `VOC-134-T00`
- Tests: `VOC-134-TEST-05`
- Evidence: `VOC-134-EV-00`
- Result: pending

The pinned fixture still selects existing `develop` without reading `main`,
falls back to live `main` when `develop` is absent, fails closed on
ambiguous or malformed refs, binds and advances `develop` to
`mergeCommit.oid` before audit close, and does not restore a missing
integration ref with `CHECKED_HEAD_SHA`. Unique develop commits remain
fail-closed. Live `reconcile-production-change` remains the exceptional
main-only identity under the 25-input limit.

## VOC-134-AC-05 — Existing controls, roles, and docs remain

- Requirement source: `VOC-134-D07`, `VOC-134-D08`
- Tasks: `VOC-134-T00`
- Tests: `VOC-134-TEST-07`, `VOC-134-TEST-08`
- Evidence: `VOC-134-EV-00`
- Result: pending

Roster markers, required-check recovery, independent review, retry caps, risk
floors, secret redaction, App-token isolation, sanitized raw-error controls,
and rollback controls remain. `config/roles.yml` is unchanged. No OpenAI
route is added. Current-state fixture README/comments name pin `f3d791…`, the
post-caller-checkout restore, and the post-implementer helper-lifetime
contract. Historical A-003 / VOC-075 / VOC-127 / VOC-129 / VOC-130 / VOC-131 /
VOC-132 / VOC-133 records are not rewritten. `AGENTS.md` and the navigator
skill are not edited for this pin. The caller-owned fixture README is not
replaced by the infrastructure repository README.

## VOC-134-AC-06 — Complete VOC-112 no-change boundary remains the immutable carrier-base bytes

- Requirement source: `VOC-134-D10`, `VOC-134-D15`
- Tasks: `VOC-134-T00`
- Tests: `VOC-134-TEST-09`
- Evidence: `VOC-134-EV-00`
- Result: pending

All five named paths are byte-for-byte identical to the immutable
carrier-base SHA selected before implementation (expected
`95a779f9e62090f856ed03f389e7ac1d901aaa14`, revalidated at dispatch) and are
absent from the implementation PR diff against that SHA:

- `scripts/foundation/fixtures/voc112-navigation-benchmark-traces.json`
- `scripts/foundation/fixtures/voc112-skill-discovery-evidence.json`
- `scripts/foundation/voc112-navigation-benchmark.test.mjs`
- `AGENTS.md`
- `.agents/skills/vocanova-repo-navigator/SKILL.md`

JSON `subject_revision` values remain
`f9d11e232a07c7d7a9c433d02c9267912543ba10`. The provenance test still
fail-closes `local` mode when a full checkout is missing the captured commit.
The regression requires the exact commit object and all five paths to resolve
and fails—not skips—if proof cannot run. After the SHA is selected, a moving
branch ref is not the authority.

## VOC-134-AC-07 — package.json remains carrier-base-identical and provenance is not bypassed

- Requirement source: `VOC-134-D14`
- Tasks: `VOC-134-T00`
- Tests: `VOC-134-TEST-10`
- Evidence: `VOC-134-EV-00`
- Result: pending

`package.json` is byte-identical to the immutable carrier-base SHA and is
absent from the implementation diff. No capture-commit fetch helper, no
provenance-mode wrapper, no evidence-stamping helper, and no test-time or
runtime evidence mutation is added. Default `local` fail-closed behavior
remains reachable. The VOC-133 attempt-1 fetch-before-test class and
attempt-2 `pr-validation`/`HEAD` wrapper class cannot recur unnoticed.

## VOC-134-AC-08 — Exact-revision evidence is fail-closed and Git-feasible

- Requirement source: `VOC-134-D12`
- Tasks: `VOC-134-T00`
- Tests: `VOC-134-TEST-11`
- Evidence: `VOC-134-EV-00`
- Result: pending

Committed package evidence records the immutable carrier base, the exact
infra merge, confirmed hashes, validation, and the contract that the final
implementation head is bound by the App-authored independent-review
comment/check. The final review comment must bind the live PR head exactly.
Merge-gate must reject any mismatch. Post-merge/root-issue audit may record
the reviewed head and merge SHA. A commit is not required to contain its own
SHA. Evidence does not claim a protected-path revert unless that path is
absent from the diff. Evidence is not rewritten at test or runtime.

## VOC-134-AC-09 — Replacement carrier is VOC-134; no snapshot-gap; VOC-129 through VOC-133 are not re-implemented

- Requirement source: `VOC-134-D01`, `VOC-134-D09`, `VOC-134-D13`
- Tasks: `VOC-134-T00`
- Tests: `VOC-134-TEST-12`
- Evidence: `VOC-134-EV-00`
- Result: pending

This package's implementation PR is a new VOC-134 carrier from current
`develop`. It does not reuse, merge, cherry-pick, or modify PR #1051, PR #1056,
or PR #1065. It does not redispatch VOC-132-T00 (#1059) or VOC-133-T00
(#1063). It `Closes` only its own VOC-134 task issue. VOC-129 PR #1046 is not
re-implemented. No VOC-129, VOC-130, VOC-131, VOC-132, or VOC-133 completion
marker is manufactured. No snapshot of the current develop/main gap is
committed. After exact-SHA review and merge, ordinary release evaluation (or
`reconcile-release`) completes develop-to-main promotion, exact develop
synchronization, production deployment where selected, and audit
reconciliation for VOC-129, VOC-130, VOC-131, VOC-132, VOC-133, and this
replacement, then closes root issue #1066. Closed state alone is not
completion proof.

Acceptance criteria must be observable, stable, security-aware, and bidirectionally
traceable to requirements, tasks, tests, and evidence.
