# VOC-090 — Acceptance Criteria

## VOC-090-AC-00 — Root cause recorded and bounded to CI job budget

- Requirement source: `VOC-090-D00`, `VOC-090-D03`
- Tasks: `VOC-090-T00`
- Tests: `VOC-090-TEST-00`
- Evidence: `VOC-090-EV-00`
- Result: pending

Task evidence identifies run 32271016931 and job
`synthetic.staging.authenticated-core-journey` as the cancellation source, with
either (a) install/setup overhead as the primary contributor or (b) documented
evidence that the Playwright journey itself exceeded its configured timeout
(triggering `VOC-090-D03` follow-up rather than silent scope expansion).

## VOC-090-AC-01 — Scheduled staging core-journey job uses dependency caching

- Requirement source: `VOC-090-D02`
- Tasks: `VOC-090-T00`
- Tests: `VOC-090-TEST-01`, `VOC-090-TEST-02`
- Evidence: `VOC-090-EV-00`
- Result: pending

The `staging-authenticated-core-journey` job in `scheduled-synthetics.yml`
caches pnpm install artifacts and Playwright Chromium browser binaries across
hourly runs using committed, reviewable workflow structure (not ad-hoc runner
state). Deterministic tests assert the caching wiring is present.

## VOC-090-AC-02 — Workflow job timeout aligns with registry synthetic budget

- Requirement source: `VOC-090-D02`, `VOC-090-D04`
- Tasks: `VOC-090-T00`
- Tests: `VOC-090-TEST-03`, `VOC-090-TEST-04`
- Evidence: `VOC-090-EV-00`
- Result: pending

`staging-authenticated-core-journey`'s `timeout-minutes` is greater than or
equal to `synthetic.staging.authenticated-core-journey.timeout_seconds / 60`
(after any rounding policy documented in the task PR). The registry
`timeout_seconds` value is updated to match the corrected job wall clock. Job
timeout remains sufficient for `playwright.staging.config.ts`'s journey timeout
plus the documented minimum setup reserve.

## VOC-090-AC-03 — Core-loop synthetic contract unchanged

- Requirement source: `VOC-090-D01`
- Tasks: `VOC-090-T00`
- Tests: `VOC-090-TEST-05`, existing `VOC-086-TEST-11` / `VOC-086-TEST-12`
- Evidence: `VOC-090-EV-00`
- Result: pending

The scheduled job still SSH-seeds the reserved synthetic account, mints a
session, and runs `core-loop.staging.spec.ts` with `retries: 0`. No reduction
of `MAX_REVIEW_CARDS`, no new Playwright retries, and no weakened step-7
assertions. All pre-existing VOC-086 scheduled-synthetics deterministic tests
pass.

## VOC-090-AC-04 — Production synthetics and observers unchanged in behavior

- Requirement source: `VOC-090-D04`
- Tasks: `VOC-090-T00`
- Tests: `VOC-086-TEST-11`, `VOC-086-TEST-12`, `VOC-088-TEST-08` (regression)
- Evidence: `VOC-090-EV-00`
- Result: pending

Production jobs in `scheduled-synthetics.yml` retain OAuth kill-switch
expectations, mint-secret usage, and non-mutating route-sweep profile.
`operational-failure-monitoring.yml` and `error-monitoring.yml` are not modified.

## VOC-090-AC-05 — Live scheduled-synthetics suite completes green

- Requirement source: issue #759 remediation outcome
- Tasks: `VOC-090-T01`
- Tests: `VOC-090-TEST-06`
- Evidence: `VOC-090-EV-01`
- Result: pending

After T00 merges to `develop`, a operator-triggered `workflow_dispatch` of
`scheduled-synthetics.yml` completes with conclusion `success` and job
`synthetic.staging.authenticated-core-journey` success within the declared job
timeout. Evidence records the run URL and duration without secrets, session
values, or personal data.

## VOC-090-AC-06 — No spurious duplicate operational issue for the same fingerprint

- Requirement source: `VOC-088-D06` (observer contract preserved)
- Tasks: `VOC-090-T01`
- Tests: `VOC-088-TEST-09` (regression)
- Evidence: `VOC-090-EV-01`
- Result: pending

Successful remediation does not leave an open issue #759 duplicate for
`scheduled-synthetics:cancelled` when the fingerprint is already owned. Closure
of issue #759 follows normal roster/package closure; this AC confirms no new
duplicate open issue is created by a green verification run.

Acceptance criteria must be observable, stable, security-aware, and bidirectionally
traceable to requirements, tasks, tests, and evidence.
