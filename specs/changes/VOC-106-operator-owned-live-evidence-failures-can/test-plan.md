# VOC-106 — Test Plan

## VOC-106-TEST-00 — Exact-head ownership loaded from live-evidence contract path

- Covers: `VOC-106-AC-00`
- Preconditions: T00 remediation ownership gate present
- Procedure: Fixture package directory with
  `.karsift/live-evidence/<task_id>.yaml` at the reviewed head; invoke
  remediation decision helper; assert the contract path is read for that task id.
- Expected result: Contract is authoritative when present at the exact head.
- Evidence: `VOC-106-EV-00`

## VOC-106-TEST-01 — Ownership values operator and live-actions recognized

- Covers: `VOC-106-AC-00`, `VOC-106-AC-01`
- Preconditions: Fixtures with `ownership: operator` and `ownership: live-actions`
- Procedure: Classify both fixtures under FAIL and under CI failure.
- Expected result: Both classify as no-implementer / escalate-operator (or
  equivalent), not RETRY.
- Evidence: `VOC-106-EV-00`

## VOC-106-TEST-02 — Operator WAITING remains suppressed

- Covers: `VOC-106-AC-04`
- Preconditions: Exact-head review state WAITING; operator-owned contract present
- Procedure: Run remediation decision fixture.
- Expected result: `WAITING` / `should_retry=false`; no implementer dispatch; no
  implementation attempt consumed.
- Evidence: `VOC-106-EV-00`

## VOC-106-TEST-03 — Negative: operator-owned FAIL does not dispatch implementer

- Covers: `VOC-106-AC-01`
- Preconditions: Exact-head review FAIL; valid operator-owned contract; attempt
  remaining
- Procedure: Run remediation decision fixture (or workflow unit harness).
- Expected result: No `implement.yml` request; sanitized escalate-operator
  outcome; merge remains fail-closed.
- Evidence: `VOC-106-EV-00`

## VOC-106-TEST-04 — Positive: ordinary implementer FAIL and CI failure still retry

- Covers: `VOC-106-AC-02`
- Preconditions: Exact-head review FAIL or `ci_failed=true`; **no** live-evidence
  contract; no contradictory operator declaration; attempt 1 of 2
- Procedure: Run separate remediation decision fixtures for review FAIL and CI
  failure.
- Expected result: `should_retry=true` with next_attempt=2 and correct task /
  package / issue outputs.
- Evidence: `VOC-106-EV-00`

## VOC-106-TEST-05 — Stale result does not retry or consume an attempt

- Covers: `VOC-106-AC-03`
- Preconditions: expected head SHA differs from live PR head (or equivalent
  stale base/head)
- Procedure: Run remediation decision fixture for operator-owned and ordinary
  fixtures.
- Expected result: `STALE` / `should_retry=false`; no implementer; no attempt
  consumed.
- Evidence: `VOC-106-EV-00`

## VOC-106-TEST-06 — Malformed or mismatched contract fail-closed

- Covers: `VOC-106-AC-03`
- Preconditions: Contract YAML invalid; missing/unrecognized `ownership`;
  `task_id` mismatch; exact automation-ownership marker without valid contract;
  unreadable package context
- Procedure: Run decision fixture for each case under FAIL and CI failure.
- Expected result: No implementer dispatch; fail-closed escalate / equivalent;
  no attempt consumed; not silently treated as ordinary RETRY.
- Evidence: `VOC-106-EV-00`

## VOC-106-TEST-07 — Operator CI failure escalates without log leakage

- Covers: `VOC-106-AC-01`, `VOC-106-AC-08`
- Preconditions: `ci_failed=true`; valid operator-owned contract
- Procedure: Run remediation decision / workflow harness; inspect escalation
  marker/summary content.
- Expected result: No implementer dispatch; escalation contains only allowlisted
  metadata; no logs, artifacts, secrets, identity data, or evidence payloads.
- Evidence: `VOC-106-EV-00`

## VOC-106-TEST-08 — Controlled workflow: zero implementer for operator-owned FAIL/CI failure

- Covers: `VOC-106-AC-01`, `VOC-106-AC-06`
- Preconditions: T00 live; operator-owned carrier (preferred: this package T01);
  controlled FAIL or CI-failure scenario
- Procedure: Observe sanitized pipeline/remediate metadata; confirm the reusable
  implement job did not execute. After source metadata is recorded, dispatch the
  read-only proof action on the exact carrier ref and reconcile that verifier run
  under the T01 contract.
- Expected result: No implementer execution; sanitized escalation present; later
  verifier succeeds at `exact_pr_head`.
- Evidence: `VOC-106-EV-01` (metadata only)

## VOC-106-TEST-09 — Controlled or fixture proof: ordinary FAIL still retries

- Covers: `VOC-106-AC-02`, `VOC-106-AC-06`
- Preconditions: T00 live; ordinary implementer-owned FAIL fixture or sanitized
  observation
- Procedure: Confirm `should_retry=true` path still starts implement for
  non-operator FAIL/CI failure within the attempt cap.
- Expected result: Ordinary bounded remediation retained.
- Evidence: `VOC-106-EV-01` (and/or `VOC-106-EV-00` if fully covered by TEST-04
  fixture reused as live-adjacent proof — record which)

## VOC-106-TEST-10 — Operator docs describe ownership-gated remediation

- Covers: `VOC-106-AC-07`
- Preconditions: T00 edits the infra README and calling-repo
  `docs/operations/live-evidence.md`; AGENTS.md may remain unchanged when policy
  text stays accurate
- Procedure: Assert operator docs describe skip/escalate for operator-owned
  FAIL/CI tasks and retained bounded retry for ordinary tasks.
- Expected result: Operators have an explicit, accurate state transition guide;
  no doc claims remediation always dispatches implement.
- Evidence: `VOC-106-EV-00`

## VOC-106-TEST-11 — Post-carrier verifier is exact-head and fail-closed

- Covers: `VOC-106-AC-05`, `VOC-106-AC-06`, `VOC-106-AC-08`
- Preconditions: Source-run and carrier fixtures, including wrong
  event/branch/task, executed implement job, missing escalation, stale PR, and
  exact-head cases
- Procedure: Exercise the verifier metadata adapter without logs or artifacts;
  assert the caller job is read-only and the T01 contract names the deterministic
  task branch, `workflow_dispatch`, and `exact_pr_head`.
- Expected result: Only the matching source run plus current carrier succeeds.
  Every mismatch fails closed with a sanitized reason; the verifier has no
  mutation, model, deploy, or application-secret capability.
- Evidence: `VOC-106-EV-00`

Include positive, negative, stale, malformed-metadata, sanitization,
authorization, verifier lineage, and regression coverage as above. Tests must not
use secrets or production data beyond public Actions metadata and issue fields
already governed as sanitized.
