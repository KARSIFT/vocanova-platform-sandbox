# VOC-097 — Test Plan

## VOC-097-TEST-00 — Operator docs describe live-evidence declaration

- Covers: `VOC-097-AC-00`
- Preconditions: T00 docs committed
- Procedure: Assert the new/updated `docs/operations/` guide exists and names
  ownership, allowlisted metadata fields, fail-closed rules, and the contract path
  chosen in `VOC-097-D03` / DEP-03 resolution. Assert AGENTS.md or template
  cross-links are present where required to avoid contradictory docs.
- Expected result: Authors can declare live evidence without reading infra source
- Evidence: `VOC-097-EV-00`

## VOC-097-TEST-01 — Change-package template mentions live-evidence ownership

- Covers: `VOC-097-AC-00`
- Preconditions: T00 template updates committed
- Procedure: Read `specs/templates/change-package/` guidance; assert live-evidence /
  operator-owned waiting is described and does not leave `TBD` placeholders for this
  topic.
- Expected result: Future planners/implementers see the convention in the template
- Evidence: `VOC-097-EV-00`

## VOC-097-TEST-02 — Waiting state is machine-detectable and distinct from FAIL

- Covers: `VOC-097-AC-01`
- Preconditions: T01 infra changes
- Procedure: Fixture posts or simulates waiting-for-operator-live-evidence with a
  declared contract and pending evidence; assert the documented marker/label/output
  is present and is not identical to a bare code-defect FAIL path.
- Expected result: Lifecycle state is explicit and parseable
- Evidence: `VOC-097-EV-01`

## VOC-097-TEST-03 — Waiting does not set remediate should_retry

- Covers: `VOC-097-AC-02`
- Preconditions: T01 remediate changes + fixture
- Procedure: Controlled repository fixture: waiting signal present, CI green,
  no genuine code-defect FAIL → parse remediate decide outputs →
  `should_retry=false` and no implement re-dispatch.
- Expected result: Waiting does not consume remediation attempts
- Evidence: `VOC-097-EV-01`, `VOC-097-EV-03`

## VOC-097-TEST-04 — Genuine FAIL still remediates once

- Covers: `VOC-097-AC-02`
- Preconditions: T01 remediate changes
- Procedure: Fixture with `VERDICT: FAIL` for a code defect (no waiting marker, or
  waiting marker absent) → remediate still schedules attempt 2 within the existing
  two-attempt cap.
- Expected result: Real implementation failures remain remediable
- Evidence: `VOC-097-EV-01`

## VOC-097-TEST-05 — Implementer workflow has no general Actions credential grant

- Covers: `VOC-097-AC-03`
- Preconditions: T01/T02 diffs
- Procedure: Static assert `implement.yml` permissions/secrets do not add general
  Actions dispatch/inspect credentials for the implementer agent; reconciler holds
  any new narrowly scoped Actions permissions separately.
- Expected result: Least-privilege invariant holds
- Evidence: `VOC-097-EV-01`, `VOC-097-EV-02`

## VOC-097-TEST-06 — Wrong workflow identity is rejected

- Covers: `VOC-097-AC-04`
- Preconditions: T02 reconciler
- Procedure: Feed a completed run whose workflow does not match the contract → no
  wake; sanitized rejection recorded.
- Expected result: Fail closed on wrong workflow
- Evidence: `VOC-097-EV-02`

## VOC-097-TEST-07 — Wrong job / missing job is rejected

- Covers: `VOC-097-AC-04`
- Preconditions: T02 reconciler
- Procedure: Matching workflow but missing/failed required job → no wake.
- Expected result: Fail closed on job set
- Evidence: `VOC-097-EV-02`

## VOC-097-TEST-08 — Wrong branch or SHA lineage is rejected

- Covers: `VOC-097-AC-04`
- Preconditions: T02 reconciler
- Procedure: Successful run on unrelated branch or SHA that fails the declared
  lineage rule → no wake.
- Expected result: Fail closed on lineage
- Evidence: `VOC-097-EV-02`

## VOC-097-TEST-09 — Allowlisted metadata only in reconcile outputs

- Covers: `VOC-097-AC-05`
- Preconditions: T02 sanitizer
- Procedure: Fixture input includes forbidden fields (log snippets, tokens,
  cookies, user ids); assert outputs/comments/evidence contain only allowlisted
  keys.
- Expected result: Sanitization holds
- Evidence: `VOC-097-EV-02`

## VOC-097-TEST-10 — No log or artifact copy invariant

- Covers: `VOC-097-AC-05`
- Preconditions: T02 reconciler
- Procedure: Assert reconciler code/tests never call log-download or artifact-download
  APIs for evidence collection; denylist patterns in tests.
- Expected result: No-log / no-artifact invariant locked
- Evidence: `VOC-097-EV-02`

## VOC-097-TEST-11 — Qualifying completion produces one wake and fresh review gate

- Covers: `VOC-097-AC-06`
- Preconditions: T02 reconciler
- Procedure: Waiting task + qualifying success run matching contract → exactly one
  wake; merge-gate still requires new exact-SHA independent verification after
  reconcile.
- Expected result: Single reconciliation; fresh review required
- Evidence: `VOC-097-EV-02`, `VOC-097-EV-05`

## VOC-097-TEST-12 — Stale or non-success conclusions are rejected

- Covers: `VOC-097-AC-04`, `VOC-097-AC-06`
- Preconditions: T02 reconciler
- Procedure: Runs with `failure` / `cancelled` / expired max_age → no wake.
- Expected result: Only fresh successful conclusions qualify (per contract)
- Evidence: `VOC-097-EV-02`

## VOC-097-TEST-13 — Timeout escalates once and stops looping

- Covers: `VOC-097-AC-07`
- Preconditions: T02 timeout logic
- Procedure: Advance waiting past configured bound without qualifying evidence →
  one escalation; further ticks do not spam or re-enter remediate.
- Expected result: Bounded waiting
- Evidence: `VOC-097-EV-02`

## VOC-097-TEST-14 — Duplicate events are idempotent

- Covers: `VOC-097-AC-07`
- Preconditions: T02 dedup
- Procedure: Deliver the same qualifying `workflow_run` / run id twice → still one
  wake / one evidence record.
- Expected result: Deduplicated reconciliation
- Evidence: `VOC-097-EV-02`

## VOC-097-TEST-15 — #779 and #785 reach governed closure or documented migration

- Covers: `VOC-097-AC-08`
- Preconditions: T04 complete
- Procedure: Read `t04-evidence.md`; assert each of #779 and #785 has either merged
  evidence-complete closure with scrubbed run references or an explicit safe
  migration record meeting `VOC-097-D08`.
- Expected result: Stranded tasks are no longer indefinitely pending
- Evidence: `VOC-097-EV-04`

## VOC-097-TEST-16 — Operational-failure observer remains separate and healthy

- Covers: `VOC-097-AC-10`
- Preconditions: T05 live window
- Procedure: Confirm waiting/reconcile comments do not use operational-failure
  fingerprint markers; spot-check that deploy/synthetic failure observation path
  still functions (scrubbed run or recent healthy observer activity). Assert no
  Sentry coupling for expected waiting outcomes.
- Expected result: Separate mechanisms; no sensitive leakage
- Evidence: `VOC-097-EV-05`

Include positive, negative, authorization, failure, and rollback coverage as
applicable. Tests must not use secrets or production data.
