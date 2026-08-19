# VOC-090 — Test Plan

## VOC-090-TEST-00 — Root-cause evidence references run 32271016931 and failing job

- Covers: `VOC-090-AC-00`
- Preconditions: T00 evidence file drafted
- Procedure: Read `t00-evidence.md`; assert it names run 32271016931, job
  `synthetic.staging.authenticated-core-journey`, and the timeout cancellation
  annotation; assert it identifies the dominant time consumer.
- Expected result: Evidence bounds remediation to CI budget unless `VOC-090-D03`
  escalation is explicitly documented
- Evidence: `VOC-090-EV-00`

## VOC-090-TEST-01 — Staging core-journey job declares pnpm caching

- Covers: `VOC-090-AC-01`
- Preconditions: T00 workflow changes merged in task branch
- Procedure: Parse `staging-authenticated-core-journey` job block from
  `scheduled-synthetics.yml`; assert a pnpm or Node dependency cache step exists
  before `pnpm install --frozen-lockfile`.
- Expected result: Committed workflow structure includes pnpm install caching
- Evidence: `VOC-090-EV-00`

## VOC-090-TEST-02 — Staging core-journey job declares Playwright browser caching

- Covers: `VOC-090-AC-01`
- Preconditions: T00 workflow changes merged in task branch
- Procedure: Parse the job block; assert Playwright browser cache restore/save
  wraps or replaces a cold `playwright install --with-deps chromium` on cache hit.
- Expected result: Committed workflow structure includes Playwright browser caching
- Evidence: `VOC-090-EV-00`

## VOC-090-TEST-03 — Registry timeout_seconds matches job timeout-minutes

- Covers: `VOC-090-AC-02`
- Preconditions: T00 registry and workflow changes
- Procedure: Load `synthetic.staging.authenticated-core-journey` from
  `synthetics.yaml`; parse job `timeout-minutes`; assert
  `timeout_seconds >= timeout_minutes * 60 - 59` and
  `timeout_seconds <= timeout_minutes * 60` (or exact equality if the package
  documents integer-minute alignment).
- Expected result: Registry and workflow budgets are consistent
- Evidence: `VOC-090-EV-00`

## VOC-090-TEST-04 — Job timeout covers Playwright journey timeout plus setup reserve

- Covers: `VOC-090-AC-02`
- Preconditions: T00 changes; `playwright.staging.config.ts` unchanged at 240s
- Procedure: Assert job `timeout-minutes * 60` is at least
  `240 + documented_setup_reserve_seconds` where setup reserve is at least
  SSH seed budget + mint + a conservative install ceiling documented in the task PR.
- Expected result: Job wall clock cannot expire before the journey's configured
  Playwright timeout under documented setup assumptions
- Evidence: `VOC-090-EV-00`

## VOC-090-TEST-05 — Core-loop synthetic wiring unchanged

- Covers: `VOC-090-AC-03`
- Preconditions: T00 task branch
- Procedure: Run `node --test scripts/foundation/voc086-scheduled-synthetics.test.mjs`
  and VOC-090 budget tests; assert staging job still runs SSH seed before mint and
  invokes `playwright.staging.config.ts` / `core-loop.staging.spec.ts`.
- Expected result: All tests pass; no retries added; MAX_REVIEW_CARDS unchanged
- Evidence: `VOC-090-EV-00`

## VOC-090-TEST-06 — Live workflow_dispatch completes success

- Covers: `VOC-090-AC-05`, `VOC-090-AC-06`
- Preconditions: T00 merged to `develop`
- Procedure: Operator dispatches `scheduled-synthetics.yml`; inspect run conclusion
  and job `synthetic.staging.authenticated-core-journey` status and duration.
- Expected result: Conclusion `success`; staging core-journey job success within
  declared timeout; no duplicate open cancelled fingerprint issue
- Evidence: `VOC-090-EV-01`

Include positive, negative, authorization, failure, migration, accessibility, and
rollback coverage as applicable. Tests must not use secrets or production data.

## Regression tests (existing, must remain green)

- `VOC-086-TEST-11`, `VOC-086-TEST-12` — scheduled synthetics registry/workflow wiring
- `VOC-088-TEST-08`–`VOC-088-TEST-11` — failure-to-issue agent (unchanged files)
