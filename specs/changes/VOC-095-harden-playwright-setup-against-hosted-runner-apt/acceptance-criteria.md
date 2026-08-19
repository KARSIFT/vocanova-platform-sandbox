# VOC-095 — Acceptance Criteria

## VOC-095-AC-00 — Root cause recorded as unbounded apt stall during Playwright install

- Requirement source: `VOC-095-D00`
- Tasks: `VOC-095-T00`
- Tests: `VOC-095-TEST-00`
- Evidence: `VOC-095-EV-00`
- Result: pending

Task evidence identifies run 32299315180 as timing out during
`playwright install --with-deps chromium` after deploy convergence, not during SSH,
migration, health check, or core-loop test execution.

## VOC-095-AC-01 — Shared install script is bounded, retry-limited, and fail-closed

- Requirement source: `VOC-095-D01`, `VOC-095-D03`, `VOC-095-D07`
- Tasks: `VOC-095-T00`
- Tests: `VOC-095-TEST-01`, `VOC-095-TEST-02`, `VOC-095-TEST-03`
- Evidence: `VOC-095-EV-00`
- Result: pending

`infra/scripts/install-playwright-chromium.sh` exists, uses explicit per-attempt timeout
and fixed retry count for the system-deps path, verifies the Chromium binary under
`~/.cache/ms-playwright` after success, and exits non-zero when deps cannot be installed.
The script contains no skip flags, `continue-on-error`, or logic that bypasses browser
test prerequisites.

## VOC-095-AC-02 — All four Chromium workflows use cache-then-script contract

- Requirement source: `VOC-095-D02`, `VOC-095-D05`
- Tasks: `VOC-095-T01`
- Tests: `VOC-095-TEST-04`, `VOC-095-TEST-05`, `VOC-095-TEST-06`, `VOC-095-TEST-07`,
  regression `VOC-086-TEST-11`
- Evidence: `VOC-095-EV-01`
- Result: pending

`accessibility.yml`, `lighthouse.yml`, `deploy-staging.yml`, and
`scheduled-synthetics.yml` each restore `~/.cache/ms-playwright` via `actions/cache`
before invoking `infra/scripts/install-playwright-chromium.sh`. No workflow retains the
inline `playwright install --with-deps chromium` one-liner.

## VOC-095-AC-03 — Browser test gates remain mandatory and unchanged in semantics

- Requirement source: issue #792 required outcome; `VOC-095-D05`
- Tasks: `VOC-095-T01`
- Tests: `VOC-095-TEST-08`, `VOC-095-TEST-09`, existing deploy/accessibility foundation
  regression
- Evidence: `VOC-095-EV-01`
- Result: pending

Accessibility, Lighthouse, staging core-loop, and scheduled core-loop jobs still run
after install failure would fail the job. Workflow `timeout-minutes` values are not
reduced. Lighthouse `LIGHTHOUSE_CHROME_PATH` shell expansion remains in the `run:` step.

## VOC-095-AC-04 — Monitoring validator aligned with shared install contract

- Requirement source: `VOC-095-DEP-02`
- Tasks: `VOC-095-T01`
- Tests: `VOC-095-TEST-10`, `VOC-086-TEST-11` (regression)
- Evidence: `VOC-095-EV-01`
- Result: pending

`infra/monitoring/scheduled-synthetics.mjs` validates cache-before-install ordering and
references the shared install script for the staging authenticated core-journey job.

## VOC-095-AC-05 — Live staging core-loop succeeds after remediation merges

- Requirement source: issue #792 required outcome; `VOC-095-D00`
- Tasks: `VOC-095-T02`
- Tests: `VOC-095-TEST-11`
- Evidence: `VOC-095-EV-02`
- Result: pending

After T01 merges to `develop`, evidence shows a `deploy-staging` run for the latest
integration commit reaches conclusion `success` and the staging core-loop Playwright
step completes (not cancelled at job timeout during browser install).

## VOC-095-AC-06 — Operator documentation references shared install contract

- Requirement source: `VOC-095-D01`
- Tasks: `VOC-095-T01`
- Tests: `VOC-095-TEST-12`
- Evidence: `VOC-095-EV-01`
- Result: pending

`apps/web/tests/e2e/README.md` documents the repository-managed install script for CI
and local first-time setup, including bounded-deps behavior and cache expectations.

Acceptance criteria must be observable, stable, security-aware, and bidirectionally
traceable to requirements, tasks, tests, and evidence.
