# VOC-110 — Acceptance Criteria

## VOC-110-AC-00 — Root cause and failing step recorded from run 32566405628

- Requirement source: `VOC-110-D00`, `VOC-110-D01`
- Tasks: `VOC-110-T00`
- Tests: `VOC-110-TEST-00`
- Evidence: `VOC-110-EV-00`
- Result: pending

Task evidence identifies run 32566405628 as an **actionable deploy-staging failure**
and records the confirmed chain: API health passed, `Poll staging.vocanova.site/`
failed, public web returned 502, and the Next.js 16.3.1 standalone container restarted
because its artifact omitted an `@swc/helpers` ESM module on Node 24. Evidence contains
no secrets, full logs, or personal data.

## VOC-110-AC-01 — Smallest correct fix applied for the identified failure mode

- Requirement source: `VOC-110-D01`, `VOC-110-D03`, `VOC-110-D05`
- Tasks: `VOC-110-T00`
- Tests: `VOC-110-TEST-01`, `VOC-110-TEST-02`, `VOC-110-TEST-03`
- Evidence: `VOC-110-EV-00`
- Result: pending

The task PR moves `next` and `@next/eslint-plugin-next` together to stable 16.3.2
while preserving the other PR #859 updates. The resulting production Docker image
boots, remains running, and serves HTTP 2xx. Next.js 16.3.0 remains the documented
rollback only if 16.3.2 cannot prove healthy.
Deploy fail-closed semantics remain: no `continue-on-error`, no removed health
checks, no skipped staging core-loop gate, no weakened OAuth-start check.

## VOC-110-AC-02 — Deterministic regression coverage for the fix

- Requirement source: `VOC-110-D03`, `VOC-110-D06`
- Tasks: `VOC-110-T00`
- Tests: `VOC-110-TEST-04`, `VOC-110-TEST-05`, regression `VOC-084-TEST-*`,
  `VOC-088-TEST-*`, `VOC-095-TEST-*` deploy-staging wiring as applicable
- Evidence: `VOC-110-EV-00`
- Result: pending

The repository pipeline contains a merge-gating container-runtime job that, for
relevant dependency/web changes, builds `apps/web/Dockerfile`, starts the image,
asserts it remains running, and requires HTTP 2xx. It safely avoids the expensive
build for irrelevant package-plan/docs-only diffs. Deterministic workflow tests lock
path selection, cleanup, and merge-gate dependency. Existing deploy-staging suites
remain green.

## VOC-110-AC-03 — Live verification: post-fix deploy-staging succeeds on develop

- Requirement source: issue #911 remediation outcome; `VOC-110-D00`
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
