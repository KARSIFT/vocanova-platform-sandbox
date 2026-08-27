# VOC-131 — Acceptance Criteria

## VOC-131-AC-00 — Caller fixture and pin equal exact infrastructure #165 merge

- Requirement source: `VOC-131-D02`
- Tasks: `VOC-131-T00`
- Tests: `VOC-131-TEST-00`, `VOC-131-TEST-01`
- Evidence: `VOC-131-EV-00`
- Result: pending

`tooling/governance/fixtures/karsift-ai-infra/PINNED_SHA.txt` equals
`8ce2b77a09a729e458a9f4cbea1ca26eb114d398`. Every live caller pin assertion
equals that same SHA. The pin does not equal
`863fc1f35b1d35e4981a59166b0e939be1a2b681`. Fixture
`.github/workflows/release.yml` and `tests/test_release_policy.py` are
byte-identical to those paths at infra merge `8ce2b77…`.

## VOC-131-AC-01 — Identify restores shared policy after caller checkout before validate-task

- Requirement source: `VOC-131-D03`
- Tasks: `VOC-131-T00`
- Tests: `VOC-131-TEST-02`, `VOC-131-TEST-04`
- Evidence: `VOC-131-EV-00`
- Result: pending

Fixture `identify` contains `Restore shared lifecycle policy after caller
checkout` after `Checkout caller release state` and before
`task-completion-runner.py validate-task`. The restore uses
`repository: ${{ job.workflow_repository }}`,
`ref: ${{ job.workflow_sha }}`, `path: karsift-ai-infra`, and
`persist-credentials: false`. The #164 checkout-ref resolver still runs after
the first shared-policy checkout and before the caller checkout.

## VOC-131-AC-02 — Converge restores shared policy after caller checkout before validate-roster

- Requirement source: `VOC-131-D03`
- Tasks: `VOC-131-T00`
- Tests: `VOC-131-TEST-03`, `VOC-131-TEST-04`
- Evidence: `VOC-131-EV-00`
- Result: pending

Fixture `converge` contains `Restore shared lifecycle policy after caller
checkout` after `Checkout caller release state` and before
`task-completion-runner.py validate-roster`. The restore uses the same
immutable reusable-workflow revision as `identify`, with credentials not
persisted. The missing-validator safe-no-op class from run `33066533397`
cannot recur unnoticed in the pinned fixture.

## VOC-131-AC-03 — #164 contracts remain after the #165 pin

- Requirement source: `VOC-131-D04`
- Tasks: `VOC-131-T00`
- Tests: `VOC-131-TEST-05`
- Evidence: `VOC-131-EV-00`
- Result: pending

The pinned fixture still selects existing `develop` without reading `main`,
falls back to live `main` when `develop` is absent, fails closed on
ambiguous or malformed refs, binds and advances `develop` to
`mergeCommit.oid` before audit close, and does not restore a missing
integration ref with `CHECKED_HEAD_SHA`. Unique develop commits remain
fail-closed. Live `reconcile-production-change` remains the exceptional
main-only identity under the 25-input limit.

## VOC-131-AC-04 — Existing controls, roles, and docs remain

- Requirement source: `VOC-131-D06`, `VOC-131-D07`
- Tasks: `VOC-131-T00`
- Tests: `VOC-131-TEST-06`, `VOC-131-TEST-07`
- Evidence: `VOC-131-EV-00`
- Result: pending

Roster markers, required-check recovery, independent review, retry caps, risk
floors, secret redaction, App-token isolation, sanitized raw-error controls,
and rollback controls remain. `config/roles.yml` is unchanged. No OpenAI
route is added. Current-state fixture README/comments name pin `8ce2b77…` and
the post-caller-checkout restore. Historical A-003 / VOC-075 / VOC-127 /
VOC-129 / VOC-130 records are not rewritten. `AGENTS.md` and the navigator
skill are not edited for this pin.

## VOC-131-AC-05 — VOC-112 fixtures remain the develop-base bytes

- Requirement source: `VOC-131-D10`
- Tasks: `VOC-131-T00`
- Tests: `VOC-131-TEST-08`
- Evidence: `VOC-131-EV-00`
- Result: pending

`scripts/foundation/fixtures/voc112-navigation-benchmark-traces.json` and
`scripts/foundation/fixtures/voc112-skill-discovery-evidence.json` are
byte-for-byte identical to the new carrier's `develop` base. Their
`subject_revision` values remain
`f9d11e232a07c7d7a9c433d02c9267912543ba10`. Neither path appears in the
implementation PR diff against that base. A deterministic regression fails if
either file differs from that base.

## VOC-131-AC-06 — Replacement carrier is VOC-131; no snapshot-gap; VOC-129 and VOC-130 are not re-implemented

- Requirement source: `VOC-131-D01`, `VOC-131-D08`, `VOC-131-D09`
- Tasks: `VOC-131-T00`
- Tests: `VOC-131-TEST-09`
- Evidence: `VOC-131-EV-00`
- Result: pending

This package's implementation PR is a new VOC-131 carrier from current
`develop`. It does not reuse, merge, or modify PR #1051. It `Closes` only its
own VOC-131 task issue. VOC-129 PR #1046 is not re-implemented. No VOC-129 or
VOC-130 completion marker is manufactured. No snapshot of the current
develop/main gap is committed. After exact-SHA review and merge, ordinary
release evaluation (or `reconcile-release`) completes develop-to-main
promotion, exact develop synchronization, production deployment where
selected, and audit reconciliation for VOC-129, VOC-130, and this
replacement. Closed state alone is not completion proof.

Acceptance criteria must be observable, stable, security-aware, and bidirectionally
traceable to requirements, tasks, tests, and evidence.
