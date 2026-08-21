# VOC-106 — Acceptance Criteria

## VOC-106-AC-00 — Ownership resolved from exact-head contract before remediation retry

- Requirement source: `VOC-106-D00`
- Tasks: `VOC-106-T00`
- Tests: `VOC-106-TEST-00`, `VOC-106-TEST-01`
- Evidence: `VOC-106-EV-00`
- Result: pending

Before remediation selects any implementer retry, it resolves task ownership from
governed package data at the exact reviewed PR head. When
`<package_path>/.karsift/live-evidence/<task_id>.yaml` exists at that head, that
contract is the authoritative machine-readable ownership source. Package-context
checks remain fail-closed and consistent with VOC-102 ownership classification.
Narrative prose is never interpreted.

## VOC-106-AC-01 — No implementer dispatch for operator-owned failures

- Requirement source: `VOC-106-D01`, `VOC-106-D02`
- Tasks: `VOC-106-T00`, `VOC-106-T01`
- Tests: `VOC-106-TEST-02`, `VOC-106-TEST-03`, `VOC-106-TEST-08`
- Evidence: `VOC-106-EV-00`, `VOC-106-EV-01`
- Result: pending

When ownership is `operator` or `live-actions`, remediation does **not** start
`implement.yml` for review `FAIL`, CI failure, stale results, missing evidence
lifecycle states, or malformed/mismatched contracts. The condition routes to a
sanitized, bounded operator/reconcile escalation. Merge remains fail-closed.

## VOC-106-AC-02 — Ordinary implementer-owned remediation retained

- Requirement source: `VOC-106-D03`
- Tasks: `VOC-106-T00`, `VOC-106-T01`
- Tests: `VOC-106-TEST-04`, `VOC-106-TEST-09`
- Evidence: `VOC-106-EV-00`, `VOC-106-EV-01`
- Result: pending

When the task is ordinary (no live-evidence contract and no contradictory
operator-owned declaration), remediation still retries once on CI failure or
genuine review `FAIL` (attempt capped at 2), then stops. `WAITING`, `STALE`, and
`REVIEW_INFRA_FAILURE` remain non-retrying.

## VOC-106-AC-03 — Stale, missing, and malformed ownership states do not consume attempts

- Requirement source: `VOC-106-D04`
- Tasks: `VOC-106-T00`
- Tests: `VOC-106-TEST-05`, `VOC-106-TEST-06`, `VOC-106-TEST-07`
- Evidence: `VOC-106-EV-00`
- Result: pending

Stale caller runs, missing qualifying evidence, and
malformed/unreadable/contradictory ownership metadata do **not** dispatch the
implementer and do **not** consume an implementation attempt. Fail-closed /
escalate-operator outcomes are used instead of guessing ordinary ownership.

## VOC-106-AC-04 — Operator WAITING remains suppressed without retry

- Requirement source: `VOC-106-D03`; VOC-097 waiting semantics
- Tasks: `VOC-106-T00`
- Tests: `VOC-106-TEST-02`
- Evidence: `VOC-106-EV-00`
- Result: pending

`VERDICT: WAITING FOR OPERATOR LIVE EVIDENCE` on the current exact head continues
to suppress remediation retry and does not consume an implementation attempt.

## VOC-106-AC-05 — Deterministic test matrix landed

- Requirement source: `VOC-106-D05`
- Tasks: `VOC-106-T00`
- Tests: `VOC-106-TEST-00` through `VOC-106-TEST-07`, `VOC-106-TEST-10`,
  `VOC-106-TEST-11`
- Evidence: `VOC-106-EV-00`
- Result: pending

Operator WAITING, FAIL, CI failure, stale, malformed/mismatched, ordinary FAIL,
sanitization, and permission-boundary coverage exist and pass in CI or infra
self-ci.

## VOC-106-AC-06 — Controlled sanitized workflow proof

- Requirement source: `VOC-106-D06`
- Tasks: `VOC-106-T00`, `VOC-106-T01`
- Tests: `VOC-106-TEST-08`, `VOC-106-TEST-09`, `VOC-106-TEST-11`
- Evidence: `VOC-106-EV-00`, `VOC-106-EV-01`
- Result: pending

A controlled operator-owned FAIL or CI-failure scenario after T00 is live proves
zero executed `implement.yml` job from remediation, with sanitized escalation
metadata recorded. A later read-only verifier on the exact carrier head validates
that source run/job metadata. Ordinary implementer-owned FAIL still retries as
intended. Evidence is metadata-only; no logs, artifacts, secrets, or manufactured
unrelated-package live evidence.

## VOC-106-AC-07 — Docs match ownership-gated remediation when touched

- Requirement source: AGENTS.md doc-consistency rule; `VOC-106-D01`, `VOC-106-D03`
- Tasks: `VOC-106-T00`
- Tests: `VOC-106-TEST-10`
- Evidence: `VOC-106-EV-00`
- Result: pending

Infra README and any in-diff calling-repo operator/governance docs accurately
state that remediation skips general implementer dispatch for operator-owned /
live-evidence-only tasks and escalates instead — or those docs are left unchanged
only when they already do not claim universal implementer retry on FAIL/CI
failure.

## VOC-106-AC-08 — Escalation and decide path remain least-privileged and sanitized

- Requirement source: `VOC-106-D02`
- Tasks: `VOC-106-T00`
- Tests: `VOC-106-TEST-07`, `VOC-106-TEST-11`
- Evidence: `VOC-106-EV-00`
- Result: pending

Operator/fail-closed escalation emits only allowlisted metadata. The decide path
does not gain model credentials or a `secrets: inherit` path into `implement.yml`
for operator-owned outcomes. The proof verifier is independently read-only and
receives no App write token, model/deploy/application secret, or Actions-write
permission.
