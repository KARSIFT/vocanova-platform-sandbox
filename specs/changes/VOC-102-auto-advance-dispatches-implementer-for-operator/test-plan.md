# VOC-102 — Test Plan

## VOC-102-TEST-00 — Next-task ownership loaded from live-evidence contract path

- Covers: `VOC-102-AC-00`
- Preconditions: T00 auto-advance ownership gate present
- Procedure: Fixture package directory with
  `.karsift/live-evidence/<next_task_id>.yaml`; invoke ownership classifier /
  auto-advance decision helper; assert the contract path is read for that task id.
- Expected result: Contract is authoritative when present.
- Evidence: `VOC-102-EV-00`

## VOC-102-TEST-01 — Ownership values operator and live-actions recognized

- Covers: `VOC-102-AC-00`, `VOC-102-AC-01`
- Preconditions: Fixtures with `ownership: operator` and `ownership: live-actions`
- Procedure: Classify both fixtures.
- Expected result: Both classify as skip-implementer (operator-owned /
  live-evidence-only).
- Evidence: `VOC-102-EV-00`

## VOC-102-TEST-02 — Negative: operator-owned next task does not dispatch

- Covers: `VOC-102-AC-01`
- Preconditions: Roster fixture where closed task has an open operator-owned next
  task with valid contract; no existing PR on agent branch
- Procedure: Run auto-advance decision fixture (or workflow unit harness).
- Expected result: decision is `prepare-live-evidence`; implement job is not
  requested; next issue remains open; clean carrier publisher path is selected.
- Evidence: `VOC-102-EV-00`

## VOC-102-TEST-03 — Positive: ordinary implementation next task still dispatches

- Covers: `VOC-102-AC-02`
- Preconditions: Roster fixture with open next task and **no** live-evidence
  contract; no contradictory operator declaration; issue open; no existing PR
- Procedure: Run auto-advance decision fixture.
- Expected result: `should_dispatch=true` with correct change_id / package_path /
  task_id / issue_number outputs.
- Evidence: `VOC-102-EV-00`

## VOC-102-TEST-04 — Malformed contract YAML fail-closed

- Covers: `VOC-102-AC-03`
- Preconditions: Next-task contract file contains invalid YAML
- Procedure: Run decision fixture.
- Expected result: No implementer dispatch; sanitized fail-closed / escalation
  outcome (not silent treat-as-ordinary).
- Evidence: `VOC-102-EV-00`

## VOC-102-TEST-05 — Missing or unrecognized ownership fail-closed

- Covers: `VOC-102-AC-03`
- Preconditions: Contract present without `ownership`, or with unrecognized value
- Procedure: Run decision fixture for each case.
- Expected result: No implementer dispatch; fail-closed signal.
- Evidence: `VOC-102-EV-00`

## VOC-102-TEST-06 — Contradictory metadata fail-closed

- Covers: `VOC-102-AC-03`
- Preconditions: Fixture where tasks.md (or equivalent governed declaration used by
  the chosen checker) marks operator-owned live evidence but contract is absent or
  `task_id` mismatches filename / roster id
- Procedure: Run decision fixture.
- Expected result: No implementer dispatch; fail-closed signal per `VOC-102-D04`.
- Evidence: `VOC-102-EV-00`

## VOC-102-TEST-07 — Last-task / no-next-task does not invent implement or early release

- Covers: `VOC-102-AC-04`
- Preconditions: Closed task is last in roster (or next missing)
- Procedure: Run decision fixture.
- Expected result: `should_dispatch=false` / no-op as today; no release issue opened
  by auto-advance itself (release remains check-completion's job).
- Evidence: `VOC-102-EV-00`

## VOC-102-TEST-08 — Controlled workflow: zero implementer for operator-owned next

- Covers: `VOC-102-AC-01`, `VOC-102-AC-06`
- Preconditions: T00 live; operator-owned next task (preferred: this package T01);
  live-evidence contract present
- Procedure: Observe sanitized pipeline/auto-advance run after predecessor close;
  confirm no `implement.yml` job for the operator-owned next task.
- Expected result: Zero implementer dispatch; next issue still open with one
  waiting marker, one deterministic draft carrier PR, and the contract-declared
  pending evidence path ready for the existing reconciler.
- Evidence: `VOC-102-EV-01` (metadata only)

## VOC-102-TEST-09 — Controlled or fixture proof: ordinary next still dispatches

- Covers: `VOC-102-AC-02`, `VOC-102-AC-06`
- Preconditions: T00 live; ordinary implementation next-task fixture or sanitized
  observation
- Procedure: Confirm `should_dispatch=true` path still starts implement for
  non-operator next tasks.
- Expected result: Ordinary dispatch retained.
- Evidence: `VOC-102-EV-01` (and/or `VOC-102-EV-00` if fully covered by TEST-03
  fixture reused as live-adjacent proof — record which)

## VOC-102-TEST-10 — Doc consistency when docs touched

- Covers: `VOC-102-AC-07`
- Preconditions: T00 may or may not edit infra README / live-evidence.md / AGENTS.md
- Procedure: If in diff, assert docs describe skip for operator-owned /
  live-evidence-only next tasks and retained dispatch for ordinary tasks. If
  untouched, assert they do not claim universal implement dispatch for every next
  task.
- Expected result: No false doc claim about auto-advance always dispatching
  implement.
- Evidence: `VOC-102-EV-00`

## VOC-102-TEST-11 — Evidence carrier is deterministic and idempotent

- Covers: `VOC-102-AC-01`, `VOC-102-AC-05`, `VOC-102-AC-08`
- Preconditions: Valid operator/live-actions contract with allowlisted
  `evidence_path`; open task issue; no carrier, then an existing carrier
- Procedure: Exercise clean publisher fixture twice.
- Expected result: First call creates one deterministic task branch/draft PR and
  pending evidence file; repeat reuses it and retains exactly one waiting marker.
  No general implementer or LLM step is invoked.
- Evidence: `VOC-102-EV-00`

## VOC-102-TEST-12 — Permission and credential boundary

- Covers: `VOC-102-AC-08`
- Preconditions: Final reusable workflow/helper diff
- Procedure: Assert classifier job remains contents/issues/pull-requests read;
  assert only clean publisher references App credentials and only requests
  contents/issues/pull-requests write; reject Actions-write, model-key exposure,
  or `secrets: inherit` to a general implementer on operator/fail-closed paths.
- Expected result: Least-privilege boundary is explicit and deterministic.
- Evidence: `VOC-102-EV-00`

Include positive, negative, malformed-metadata, carrier idempotency,
authorization, and regression coverage as above. Tests must not use secrets or
production data beyond public Actions metadata and issue fields already governed
as sanitized.
