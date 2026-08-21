# VOC-102 — Acceptance Criteria

## VOC-102-AC-00 — Ownership detected from governed package data before dispatch

- Requirement source: `VOC-102-D00`
- Tasks: `VOC-102-T00`
- Tests: `VOC-102-TEST-00`, `VOC-102-TEST-01`
- Evidence: `VOC-102-EV-00`
- Result: pending

Before auto-advance sets `should_dispatch=true`, it reads next-task ownership from
governed package data. When
`<package_path>/.karsift/live-evidence/<next_task_id>.yaml` exists, that contract
is the authoritative machine-readable ownership source.

## VOC-102-AC-01 — No implementer dispatch for operator-owned or live-evidence-only next tasks

- Requirement source: `VOC-102-D01`, `VOC-102-D02`
- Tasks: `VOC-102-T00`, `VOC-102-T01`
- Tests: `VOC-102-TEST-02`, `VOC-102-TEST-08`
- Evidence: `VOC-102-EV-00`, `VOC-102-EV-01`
- Result: pending

When next-task ownership is `operator` or `live-actions`, auto-advance does **not**
start `implement.yml`. The next task issue remains open and is clearly marked as
waiting for the dedicated reconcile/evidence path (sanitized signal per
`VOC-102-D02`).

## VOC-102-AC-02 — Ordinary implementation next tasks still dispatch

- Requirement source: `VOC-102-D03`
- Tasks: `VOC-102-T00`, `VOC-102-T01`
- Tests: `VOC-102-TEST-03`, `VOC-102-TEST-09`
- Evidence: `VOC-102-EV-00`, `VOC-102-EV-01`
- Result: pending

When the next task is ordinary implementation-owned (no live-evidence contract and
no contradictory operator-owned declaration), auto-advance still sets
`should_dispatch=true` and invokes `implement.yml` attempt 1 after existing
open-issue and existing-PR guards.

## VOC-102-AC-03 — Fail closed on missing, malformed, or contradictory ownership metadata

- Requirement source: `VOC-102-D04`
- Tasks: `VOC-102-T00`
- Tests: `VOC-102-TEST-04`, `VOC-102-TEST-05`, `VOC-102-TEST-06`
- Evidence: `VOC-102-EV-00`
- Result: pending

Malformed YAML, missing/unrecognized `ownership`, `task_id` mismatch, unreadable
contracts, or contradictory operator-owned declarations without a valid contract
do **not** dispatch the implementer. A sanitized failure/escalation signal is
emitted instead of guessing.

## VOC-102-AC-04 — Final-roster release behavior preserved

- Requirement source: `VOC-102-D05`
- Tasks: `VOC-102-T00`
- Tests: `VOC-102-TEST-07`
- Evidence: `VOC-102-EV-00`
- Result: pending

Skipping implementer for an operator-owned next task does not open or advance
release. No-next-task / last-task behavior still defers to existing release
check-completion. Release still waits until the operator task actually closes.

## VOC-102-AC-05 — Deterministic test matrix landed

- Requirement source: `VOC-102-D06`
- Tasks: `VOC-102-T00`
- Tests: `VOC-102-TEST-00` through `VOC-102-TEST-07`
- Evidence: `VOC-102-EV-00`
- Result: pending

Positive dispatch, negative skip, malformed/contradictory fail-closed, and
last-task/no-next regression coverage exist and pass in CI or infra self-ci.

## VOC-102-AC-06 — Controlled sanitized workflow proof

- Requirement source: `VOC-102-D07`
- Tasks: `VOC-102-T01`
- Tests: `VOC-102-TEST-08`, `VOC-102-TEST-09`
- Evidence: `VOC-102-EV-01`
- Result: pending

A controlled, sanitized workflow event proves zero implementer dispatch for an
operator-owned next task, and that an ordinary implementation next task still
dispatches as intended. Evidence is metadata-only; no secrets; no manufactured
unrelated-package live evidence.

## VOC-102-AC-07 — Docs match skip-vs-dispatch behavior when touched

- Requirement source: AGENTS.md doc-consistency rule; `VOC-102-D01`, `VOC-102-D03`
- Tasks: `VOC-102-T00`
- Tests: `VOC-102-TEST-10`
- Evidence: `VOC-102-EV-00`
- Result: pending

Infra README and any in-diff calling-repo operator/governance docs accurately
state that auto-advance skips general implementer dispatch for operator-owned /
live-evidence-only next tasks and leaves them on the reconcile path — or those
docs are left unchanged only when they already do not claim universal implement
dispatch.
