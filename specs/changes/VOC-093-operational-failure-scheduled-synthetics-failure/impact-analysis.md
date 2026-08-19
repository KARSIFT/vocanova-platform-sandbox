# VOC-093 — Impact Analysis

## Security and privacy

- **Secrets:** No new secrets. Production session mint tokens remain in existing
  GitHub secrets; evidence must remain scrubbed of session cookies and OAuth state.
- **Personal data:** Route-sweep checks use the reserved synthetic account only.
  Job log excerpts in evidence must not include email addresses or user identifiers.
- **Observer:** `operational-failure-monitoring.yml` is unchanged. Successful
  remediation should allow issue #771 to close without weakening sanitization.

## Application and operational surface

- **Application code:** Change only when T00 evidence proves a production route
  regression. Harness-only fixes stay in `infra/scripts/`.
- **Workflow:** Default expectation is no `scheduled-synthetics.yml` change because
  journey-content succeeded with the same mint and kill-switch env. Touch the
  workflow only when misconfiguration is proven.
- **Production behavior:** A route fix restores user-visible correctness; the
  synthetic remains read-only.

## Data and migrations

- No application schema migration in the default path.
- No database mutation from the synthetic suite.
- Rollback reverts T00 commits; the next hourly run may fail again until remerged.

## Analytics and accessibility

- No analytics change unless a route fix incidentally touches analytics code
  (record in task PR if so).
- Accessibility impact follows any route fix; no standalone a11y scope unless
  required by the specific route change.

## Risks, dependencies, and evidence

- `VOC-093-R00`: **False green route sweep** if coverage is weakened to mask a
  regression. Mitigation: `VOC-093-D02` forbids skipping routes or accepting
  sign-in redirects on protected paths.
- `VOC-093-R01`: **Misdiagnosis** because journey-content and route-sweep run
  concurrently — transient production blip could differ on re-run. Mitigation:
  T01 live verification and `VOC-093-D04`.
- `VOC-093-R02`: **Harness over-tight redirect rules** causing false failures on
  acceptable in-app redirects. Mitigation: extend selftests with the proven case;
  do not remove sign-in redirect rejection.
- `VOC-093-DEP-00`: Resolved — run 32288703894 public metadata identifies failing
  job and exit code 1 on route-sweep smoke step.
- `VOC-093-DEP-01`: VOC-085 route sweep + VOC-086 scheduled synthetics wiring.
- `VOC-093-EV-00`: T00 evidence — failing check identification, fix diff,
  deterministic test output.
- `VOC-093-EV-01`: T01 evidence — green live `workflow_dispatch` run URL and duration.
