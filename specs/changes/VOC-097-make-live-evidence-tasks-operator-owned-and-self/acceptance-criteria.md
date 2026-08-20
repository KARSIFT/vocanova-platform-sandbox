# VOC-097 — Acceptance Criteria

## VOC-097-AC-00 — Live-evidence ownership is declarable by task authors

- Requirement source: `VOC-097-D03`; issue #823 outcomes 11
- Tasks: `VOC-097-T00`
- Tests: `VOC-097-TEST-00`, `VOC-097-TEST-01`
- Evidence: `VOC-097-EV-00`
- Result: pending

Governance/operator documentation and the change-package template describe how to
declare operator-owned live evidence and the allowlisted evidence contract fields.
Authors can locate the declaration shape without reading karsift-ai-infra source.

## VOC-097-AC-01 — Waiting is an explicit lifecycle state distinct from FAIL

- Requirement source: `VOC-097-D00`; issue #823 outcomes 1, 7
- Tasks: `VOC-097-T01`
- Tests: `VOC-097-TEST-02`, `VOC-097-TEST-03`
- Evidence: `VOC-097-EV-01`
- Result: pending

When a task PR's only unmet acceptance condition is operator-owned live evidence
per its declared contract, automation records an explicit waiting state (label,
PR/issue comment marker, and/or workflow output — implementer chooses one primary
machine-readable signal documented in evidence). That state is distinct from a
posted `VERDICT: FAIL` that means implementation/docs are wrong.

## VOC-097-AC-02 — Waiting does not invoke remediation retries

- Requirement source: `VOC-097-D00`; issue #823 outcomes 7
- Tasks: `VOC-097-T01`
- Tests: `VOC-097-TEST-03`, `VOC-097-TEST-04`
- Evidence: `VOC-097-EV-01`
- Result: pending

A controlled fixture shows: task enters waiting for live evidence →
`remediate.yml` (or equivalent caller path) does **not** set `should_retry=true`
and does **not** re-dispatch `implement.yml`. Genuine `VERDICT: FAIL` / CI failure
paths still remediate as today.

## VOC-097-AC-03 — Implementer stays least-privileged for Actions

- Requirement source: `VOC-097-D01`; issue #823 outcomes 2
- Tasks: `VOC-097-T01`, `VOC-097-T02`
- Tests: `VOC-097-TEST-05`
- Evidence: `VOC-097-EV-01`, `VOC-097-EV-02`
- Result: pending

`implement.yml` (and the implementer prompt) do not gain general GitHub Actions
dispatch/inspect credentials. Only the dedicated repository-controlled reconcile
path holds the narrowly scoped permissions needed to observe or dispatch
**declared** workflows/jobs.

## VOC-097-AC-04 — Observe/dispatch is limited to the declared contract

- Requirement source: `VOC-097-D01`, `VOC-097-D03`; issue #823 outcomes 3, 6
- Tasks: `VOC-097-T02`
- Tests: `VOC-097-TEST-06`, `VOC-097-TEST-07`, `VOC-097-TEST-08`
- Evidence: `VOC-097-EV-02`
- Result: pending

Automation refuses to treat a run as qualifying evidence when workflow identity,
job set, event, branch/SHA lineage, conclusion, or contract parse is wrong or
ambiguous. Malformed contracts fail closed (no wake, no remediation-as-fix).

## VOC-097-AC-05 — Evidence metadata is allowlisted and sanitized

- Requirement source: `VOC-097-D02`; issue #823 outcomes 4
- Tasks: `VOC-097-T02`
- Tests: `VOC-097-TEST-09`, `VOC-097-TEST-10`
- Evidence: `VOC-097-EV-02`
- Result: pending

Reconcile outputs, issue/PR comments, and evidence files may include only
allowlisted fields (workflow identity, event, branch, exact SHA, run/job IDs,
conclusion, timestamps, bounded duration). Fixtures prove logs, artifacts,
secrets, OAuth/session/cookie/token material, user identifiers, and arbitrary job
output are stripped or rejected.

## VOC-097-AC-06 — Qualifying evidence wakes the task and requires fresh exact-SHA review

- Requirement source: `VOC-097-D04`; issue #823 outcomes 5
- Tasks: `VOC-097-T02`
- Tests: `VOC-097-TEST-11`, `VOC-097-TEST-12`
- Evidence: `VOC-097-EV-02`
- Result: pending

A controlled qualifying workflow completion produces exactly one reconciliation
wake for the waiting task, records sanitized evidence on the governed path, and
leaves merge gated on a **new** independent verification bound to the exact
post-reconcile head SHA. Prior review comments on earlier SHAs do not satisfy
merge-gate alone.

## VOC-097-AC-07 — Timeout, escalation, and deduplication bound waiting

- Requirement source: `VOC-097-D06`; issue #823 outcomes 8
- Tasks: `VOC-097-T02`
- Tests: `VOC-097-TEST-13`, `VOC-097-TEST-14`
- Evidence: `VOC-097-EV-02`
- Result: pending

Missing live runs cannot loop indefinitely: after the configured bound, waiting
escalates once with a sanitized summary and stops automatic reconcile loops.
Duplicate events for the same run identity do not double-wake or double-comment.

## VOC-097-AC-08 — Stranded #779 and #785 are reconciled

- Requirement source: `VOC-097-D08`; issue #823 outcomes 10
- Tasks: `VOC-097-T04`
- Tests: `VOC-097-TEST-15`
- Evidence: `VOC-097-EV-04`
- Result: pending

Task issues #779 (`VOC-093-T01`) and #785 (`VOC-094-T01`) are either (a) placed on
the governed waiting/reconcile path and closed after qualifying evidence + fresh
exact-SHA review + merge, or (b) closed via an explicitly documented safe
migration path recorded in `t04-evidence.md`. No implementer Actions credential
grant is used.

## VOC-097-AC-09 — Deterministic fixture matrix passes

- Requirement source: issue #823 outcomes 9
- Tasks: `VOC-097-T03`
- Tests: `VOC-097-TEST-02` through `VOC-097-TEST-14`
- Evidence: `VOC-097-EV-03`
- Result: pending

Foundation/infra deterministic tests cover waiting, successful reconciliation,
timeout, wrong workflow/job/branch/SHA, duplicates, sanitization, and
no-credential/no-log invariants. Full repository validation and governance risk
classification succeed for in-scope diffs.

## VOC-097-AC-10 — Operational-failure observer and Sentry remain separate and healthy

- Requirement source: `VOC-097-D07`; issue #823 monitoring + verification
- Tasks: `VOC-097-T05`
- Tests: `VOC-097-TEST-16`
- Evidence: `VOC-097-EV-05`
- Result: pending

Live verification shows existing deploy/synthetic failure observation remains
healthy and is not coupled to live-evidence waiting signals. No sensitive data
appears in live-evidence lifecycle comments.

Acceptance criteria must be observable, stable, security-aware, and bidirectionally
traceable to requirements, tasks, tests, and evidence.
