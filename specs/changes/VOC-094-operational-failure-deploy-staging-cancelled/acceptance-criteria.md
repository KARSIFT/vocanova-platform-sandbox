# VOC-094 — Acceptance Criteria

## VOC-094-AC-00 — Root cause recorded as concurrency queue supersession

- Requirement source: `VOC-094-D00`
- Tasks: `VOC-094-T00`
- Tests: `VOC-094-TEST-00`
- Evidence: `VOC-094-EV-00`
- Result: pending

Task evidence identifies run 32290409156 as a **pending-run supersession** in group
`staging-deploy` (public annotation, ~2m 28s duration, zero jobs started), not a
deploy step failure or job timeout.

## VOC-094-AC-01 — Staging deploy concurrency allows multi-run queue

- Requirement source: `VOC-094-D01`
- Tasks: `VOC-094-T00`
- Tests: `VOC-094-TEST-01`
- Evidence: `VOC-094-EV-00`
- Result: pending

`.github/workflows/deploy-staging.yml` declares `concurrency.group: staging-deploy`,
`cancel-in-progress: false`, and **`queue: max`**. Deterministic tests assert the
three fields together; `cancel-in-progress` is not changed to `true`.

## VOC-094-AC-02 — Production deploy concurrency matches queue posture

- Requirement source: `VOC-094-D04`
- Tasks: `VOC-094-T00`
- Tests: `VOC-094-TEST-02`
- Evidence: `VOC-094-EV-00`
- Result: pending

`.github/workflows/deploy-production.yml` declares `concurrency.group: production-deploy`,
`cancel-in-progress: false`, and **`queue: max`**.

## VOC-094-AC-03 — Observer skips benign deploy cancellations only

- Requirement source: `VOC-094-D02`, `VOC-094-D03`, `VOC-094-D05`
- Tasks: `VOC-094-T00`
- Tests: `VOC-094-TEST-03`, `VOC-094-TEST-04`, `VOC-094-TEST-05`, regression
  `VOC-088-TEST-08`–`VOC-088-TEST-11`
- Evidence: `VOC-094-EV-00`
- Result: pending

When classification metadata matches concurrency supersession for
`deploy-staging` or `deploy-production`, the observer does **not** POST a new issue.
Real `failure`, `timed_out`, and unclassified `cancelled` conclusions still open
(or deduplicate) issues exactly as before. Classifier uses bounded metadata only.

## VOC-094-AC-04 — Deploy workflow contract unchanged aside from queue depth

- Requirement source: `VOC-094-D01`
- Tasks: `VOC-094-T00`
- Tests: `VOC-094-TEST-06`, existing deploy foundation tests (regression)
- Evidence: `VOC-094-EV-00`
- Result: pending

No deploy step removed, no `continue-on-error` added, no health-check or core-loop
ordering change. Existing VOC-084/VOC-088 deploy-staging deterministic tests remain
green.

## VOC-094-AC-05 — Live verification: latest develop deploy succeeds; no spurious issue

- Requirement source: issue #781 remediation outcome; `VOC-094-D03`
- Tasks: `VOC-094-T01`
- Tests: `VOC-094-TEST-07`
- Evidence: `VOC-094-EV-01`
- Result: pending

After T00 merges to `develop`, evidence shows (a) a `deploy-staging` run for the
latest integration commit reaches conclusion `success`, and (b) a controlled
supersession scenario or fixture-backed proof demonstrates that benign
`deploy-staging:cancelled` events no longer leave a new open operational issue
when the fingerprint is not already owned.

## VOC-094-AC-06 — Issue #781 fingerprint hygiene preserved

- Requirement source: `VOC-088-D04` (observer contract)
- Tasks: `VOC-094-T01`
- Tests: `VOC-088-TEST-09` (regression)
- Evidence: `VOC-094-EV-01`
- Result: pending

Successful remediation does not create duplicate open issues for
`deploy-staging:cancelled` beyond issue #781 while that fingerprint is owned.
Closure of issue #781 follows normal roster/package closure.

Acceptance criteria must be observable, stable, security-aware, and bidirectionally
traceable to requirements, tasks, tests, and evidence.
