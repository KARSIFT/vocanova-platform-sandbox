# VOC-130 — Acceptance Criteria

## VOC-130-AC-00 — Caller fixture and pin equal exact infrastructure #165 merge

- Requirement source: `VOC-130-D02`
- Tasks: `VOC-130-T00`
- Tests: `VOC-130-TEST-00`, `VOC-130-TEST-01`
- Evidence: `VOC-130-EV-00`
- Result: pending

`tooling/governance/fixtures/karsift-ai-infra/PINNED_SHA.txt` equals
`8ce2b77a09a729e458a9f4cbea1ca26eb114d398`. Every live caller pin assertion
equals that same SHA. The pin does not equal
`863fc1f35b1d35e4981a59166b0e939be1a2b681`. The mirrored fixture includes the
in-scope #165 `release.yml` restore and its policy tests.

## VOC-130-AC-01 — Identify restores shared policy after caller checkout before validate-task

- Requirement source: `VOC-130-D03`
- Tasks: `VOC-130-T00`
- Tests: `VOC-130-TEST-02`, `VOC-130-TEST-04`
- Evidence: `VOC-130-EV-00`
- Result: pending

Fixture `identify` contains `Restore shared lifecycle policy after caller
checkout` after `Checkout caller release state` and before
`task-completion-runner.py validate-task`. The restore uses
`repository: ${{ job.workflow_repository }}`,
`ref: ${{ job.workflow_sha }}`, and `path: karsift-ai-infra`. The #164
checkout-ref resolver still runs after the first shared-policy checkout and
before the caller checkout.

## VOC-130-AC-02 — Converge restores shared policy after caller checkout before validate-roster

- Requirement source: `VOC-130-D03`
- Tasks: `VOC-130-T00`
- Tests: `VOC-130-TEST-03`, `VOC-130-TEST-04`
- Evidence: `VOC-130-EV-00`
- Result: pending

Fixture `converge` contains `Restore shared lifecycle policy after caller
checkout` after `Checkout caller release state` and before
`task-completion-runner.py validate-roster`. The restore uses the same
immutable reusable-workflow revision as `identify`. The missing-validator
safe-no-op class from run `33066533397` cannot recur unnoticed in the pinned
fixture.

## VOC-130-AC-03 — #164 contracts remain after the #165 pin

- Requirement source: `VOC-130-D04`
- Tasks: `VOC-130-T00`
- Tests: `VOC-130-TEST-05`
- Evidence: `VOC-130-EV-00`
- Result: pending

The pinned fixture still selects existing `develop` without reading `main`,
falls back to live `main` when `develop` is absent, fails closed on
ambiguous or malformed refs, binds and advances `develop` to
`mergeCommit.oid` before audit close, and does not restore a missing
integration ref with `CHECKED_HEAD_SHA`. Unique develop commits remain
fail-closed. Live `reconcile-production-change` remains the exceptional
main-only identity under the 25-input limit.

## VOC-130-AC-04 — Existing controls, roles, and docs remain

- Requirement source: `VOC-130-D06`, `VOC-130-D07`
- Tasks: `VOC-130-T00`
- Tests: `VOC-130-TEST-06`, `VOC-130-TEST-07`
- Evidence: `VOC-130-EV-00`
- Result: pending

Roster markers, required-check recovery, independent review, retry caps, risk
floors, secret redaction, App-token isolation, sanitized raw-error controls,
and rollback controls remain. `config/roles.yml` is unchanged. No OpenAI
route is added. Current-state fixture README/comments name pin `8ce2b77…` and
the post-caller-checkout restore. Historical A-003 / VOC-075 / VOC-127 /
VOC-129 records are not rewritten.

## VOC-130-AC-05 — VOC-129 and this package promote without a snapshot-gap task

- Requirement source: `VOC-130-D08`
- Tasks: `VOC-130-T00`
- Tests: `VOC-130-TEST-08`
- Evidence: `VOC-130-EV-00`
- Result: pending

This package's implementation PR is a new VOC-130 carrier from current
`develop`. It `Closes` only its own VOC-130 task issue. VOC-129 PR #1046 is
not re-implemented. No snapshot of the current develop/main gap is committed.
After exact-SHA review and merge, ordinary release evaluation (or
`reconcile-release`) completes develop-to-main promotion, exact develop
synchronization, production deployment where selected, and audit
reconciliation for VOC-129 and this blocker. Closed state alone is not
completion proof.

Acceptance criteria must be observable, stable, security-aware, and bidirectionally
traceable to requirements, tasks, tests, and evidence.
