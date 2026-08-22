# VOC-110 — Acceptance Criteria

## VOC-110-AC-00 — Root cause and failing step recorded from run 32566405628

- Requirement source: `VOC-110-D00`, `VOC-110-D01`
- Tasks: `VOC-110-T00`
- Tests: `VOC-110-TEST-00`
- Evidence: `VOC-110-EV-00`
- Result: pending

Task evidence identifies run 32566405628 as an **actionable deploy-staging failure**
(job `deploy to staging`, conclusion `failure`, exit code 1, ~8m duration, head
`f25e4cc…` after Dependabot PR #859 merge) and names the **failing workflow step**
from job logs without copying secrets or personal data.

## VOC-110-AC-01 — Smallest correct fix applied for the identified failure mode

- Requirement source: `VOC-110-D01`, `VOC-110-D03`, `VOC-110-D05`
- Tasks: `VOC-110-T00`
- Tests: `VOC-110-TEST-01`, `VOC-110-TEST-02`, `VOC-110-TEST-03`
- Evidence: `VOC-110-EV-00`
- Result: pending

The task PR fixes the defect evidenced in AC-00 in the smallest correct surface.
Deploy fail-closed semantics remain: no `continue-on-error`, no removed health
checks, no skipped staging core-loop gate, no weakened OAuth-start check.

## VOC-110-AC-02 — Deterministic regression coverage for the fix

- Requirement source: `VOC-110-D03`
- Tasks: `VOC-110-T00`
- Tests: `VOC-110-TEST-04`, `VOC-110-TEST-05`, regression `VOC-084-TEST-*`,
  `VOC-088-TEST-*`, `VOC-095-TEST-*` deploy-staging wiring as applicable
- Evidence: `VOC-110-EV-00`
- Result: pending

New or extended foundation tests lock the corrected behavior or a fixture matching
the failure mode from AC-00. Existing deploy-staging deterministic suites remain
green.

## VOC-110-AC-03 — Live verification: post-fix deploy-staging succeeds on develop

- Requirement source: issue #911 remediation outcome; `VOC-110-D04`
- Tasks: `VOC-110-T01`
- Tests: `VOC-110-TEST-06`
- Evidence: `VOC-110-EV-01`
- Result: pending

After T00 merges to `develop`, operator-owned live evidence shows a `deploy-staging`
run whose HEAD SHA contains the fix reaches conclusion `success` with job
`deploy to staging` succeeding.

## VOC-110-AC-04 — Issue #911 fingerprint hygiene preserved

- Requirement source: `VOC-088-D04` (observer contract)
- Tasks: `VOC-110-T01`
- Tests: `VOC-088-TEST-09` (regression)
- Evidence: `VOC-110-EV-01`
- Result: pending

Successful remediation does not create duplicate open issues for
`deploy-staging:failure` beyond issue #911 while that fingerprint is owned.
Closure of issue #911 follows normal roster/package closure.

Acceptance criteria must be observable, stable, security-aware, and bidirectionally
traceable to requirements, tasks, tests, and evidence.
