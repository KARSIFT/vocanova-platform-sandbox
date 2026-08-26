# VOC-122 — Acceptance Criteria

## VOC-122-AC-00 — Newly appearing required rows are recovered in the same invocation

- Requirement source: `VOC-122-D01`, `VOC-122-D02`
- Tasks: `VOC-122-T00`
- Tests: `VOC-122-TEST-00`, `VOC-122-TEST-01`
- Evidence: `VOC-122-EV-00`
- Result: pending

On the exact reviewed infrastructure revision, a promotion-recovery invocation
whose first required-check snapshot lacks a context, and whose later snapshot
shows that context as a failed or cancelled exact pull-request Actions row,
reruns that selected run in the same invocation. Polling until the 1800-second
timeout while that recoverable row remains selected is a failing result. A
later independent recovery job succeeding, as in release run `32912851044` for
PR #1000, does not satisfy this criterion for the first invocation.

## VOC-122-AC-01 — Every rerun is identity-checked against the selected required row

- Requirement source: `VOC-122-D02`, `VOC-122-D03`
- Tasks: `VOC-122-T00`
- Tests: `VOC-122-TEST-03`, `VOC-122-TEST-05`
- Evidence: `VOC-122-EV-00`
- Result: pending

Before every rerun, including replans during polling, recovery validates the
open promotion PR, exact head SHA, branch, workflow path, `pull_request`
event, first attempt, and required context. SHA, branch, PR, path, event, or
attempt mismatch fails closed immediately and does not keep polling.

## VOC-122-AC-02 — Reruns and absent-context dispatches stay once-per-invocation

- Requirement source: `VOC-122-D02`
- Tasks: `VOC-122-T00`
- Tests: `VOC-122-TEST-00`, `VOC-122-TEST-01`, `VOC-122-TEST-02`
- Evidence: `VOC-122-EV-00`
- Result: pending

A selected failed/cancelled exact pull-request run ID is rerun at most once
per recovery invocation. A genuinely absent required context is dispatched at
most once. Repeated snapshots of the same unsatisfied view do not duplicate
those mutations. An earlier absent-context dispatch does not prevent a later
rerun of a different selected pull-request run ID for that same context.

## VOC-122-AC-03 — Alternate runs and same-named statuses cannot override the selected row

- Requirement source: `VOC-122-D02`
- Tasks: `VOC-122-T00`
- Tests: `VOC-122-TEST-04`
- Evidence: `VOC-122-EV-00`
- Result: pending

A successful workflow-dispatch run or same-named commit status does not mark a
cancelled or failed ruleset-selected pull-request row satisfied. Recovery still
reruns or continues waiting on GitHub's selected required row. This preserves
VOC-121; it does not replace it.

## VOC-122-AC-04 — Ambiguous, foreign, and mismatched rows fail closed on replan

- Requirement source: `VOC-122-D03`
- Tasks: `VOC-122-T00`
- Tests: `VOC-122-TEST-05`
- Evidence: `VOC-122-EV-00`
- Result: pending

If a later snapshot is ambiguous (more than one candidate run ID), names a
foreign workflow or non-`pull_request` event, or fails selected-run identity
checks, recovery exits non-zero without dispatching an override or continuing
to poll as if the row were merely pending.

## VOC-122-AC-05 — Timeout, credentials, merge decision, and retry limits remain

- Requirement source: `VOC-122-D04`, `VOC-122-D07`
- Tasks: `VOC-122-T00`
- Tests: `VOC-122-TEST-06`
- Evidence: `VOC-122-EV-00`
- Result: pending

The change does not weaken the 1800-second timeout, 30-second poll interval,
job-token versus App-token separation, exact-head promotion merge decision,
first-attempt rerun identity, implementer two-attempt bound, or
`integration_push` recovery semantics. It does not fabricate statuses or
bypass rulesets.

## VOC-122-AC-06 — Deterministic time-evolving tests cover the live failure class

- Requirement source: `VOC-122-D05`
- Tasks: `VOC-122-T00`
- Tests: `VOC-122-TEST-00` through `VOC-122-TEST-05`
- Evidence: `VOC-122-EV-00`
- Result: pending

Infrastructure tests include a time-evolving fixture where a required context
is absent in snapshot 1, appears cancelled in snapshot 2, is rerun once, and
then succeeds, plus idempotent repeated snapshots and fail-closed
ambiguous/foreign/mismatched replan cases. Tests do not use secrets or
production data.

## VOC-122-AC-07 — Current-state docs and caller pin follow the reviewed infra merge

- Requirement source: `VOC-122-D06`, `VOC-122-D08`
- Tasks: `VOC-122-T00`
- Tests: `VOC-122-TEST-07`
- Evidence: `VOC-122-EV-00`
- Result: pending

Current-state recovery comments and infra README no longer claim that
promotion recovery plans mutations only from the initial snapshot. Both
repositories are validated on their exact reviewed revisions. If the caller
fixture consumes the infrastructure change, `PINNED_SHA.txt` equals that exact
infra merge SHA and matching caller pin assertions are advanced. If not, the
pin is unchanged and non-consumption is recorded in `t00-evidence.md`.

Acceptance criteria must be observable, stable, security-aware, and bidirectionally
traceable to requirements, tasks, tests, and evidence.
