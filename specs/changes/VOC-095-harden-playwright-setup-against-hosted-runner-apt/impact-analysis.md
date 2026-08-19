# VOC-095 — Impact Analysis

## Security and privacy

- **Secrets:** No new secrets. Install script and workflow changes must not log session
  cookies, minted tokens, SSH output, or OAuth state. CI output limited to attempt counts,
  exit codes, and cache hit/miss indicators.
- **Personal data:** No change to data handling. Evidence remains scrubbed.
- **Fail-closed:** Install failure must fail the job; no bypass that runs browser tests
  without a verified Chromium binary.

## Application and operational surface

- **Application code:** No intentional change.
- **Workflow:** Four CI/deploy workflows gain browser cache restore and a bounded install
  script. Test suites, timeouts, deploy ordering, health checks, and OAuth guards unchanged.
- **Operational risk reduction:** Deploy-staging should stop timing out spuriously after
  successful convergence when apt stalls; bounded install fails faster with clearer logs.
- **Residual risk:** Persistent apt outage still fails jobs (correct fail-closed behavior)
  until mirror recovery or apt cache follow-up (open question 3).

## Data and migrations

- No application schema migration.
- No database mutation.
- Rollback reverts script and workflow commits; workflows return to inline unbounded
  `--with-deps` behavior.

## Analytics and accessibility

- No analytics change — evidence-backed non-applicability.
- No product UI change — evidence-backed non-applicability.
- Accessibility **CI execution** remains required; install hardening reduces false
  cancellations without changing scan coverage or thresholds.

## Risks, dependencies, and evidence

- `VOC-095-R00`: **Bounded timeout too aggressive** causes flaky failures on slow but
  healthy mirrors. Mitigation: retry loop with documented constants; T02 live evidence;
  human review of timeout values at adoption (open question 2).
- `VOC-095-R01`: **Cache key drift** serves wrong browser revision after Playwright bump.
  Mitigation: key includes `hashFiles('pnpm-lock.yaml')`; script verifies binary exists.
- `VOC-095-R02`: **Validator drift** if scheduled-synthetics.mjs not updated with script
  name. Mitigation: T01 explicitly updates validator + VOC-086 regression tests.
- `VOC-095-R03`: **Lighthouse regression** if Playwright cache layout changes. Mitigation:
  `VOC-095-D06`; TEST-09 locks shell-expanded path resolution.
- `VOC-095-DEP-00`: Resolved — issue #792 and run 32299315180 identify apt stall during
  `--with-deps`.
- `VOC-095-DEP-01`: VOC-031 Playwright harness; VOC-086 browser cache precedent in
  scheduled-synthetics.yml.
- `VOC-095-DEP-02`: scheduled-synthetics.mjs cache-before-install validator.
- `VOC-095-EV-00`: T00 evidence — root cause, script constants, foundation test output.
- `VOC-095-EV-01`: T01 evidence — workflow diff summary, foundation test output.
- `VOC-095-EV-02`: T02 evidence — green deploy-staging run with core-loop success.
