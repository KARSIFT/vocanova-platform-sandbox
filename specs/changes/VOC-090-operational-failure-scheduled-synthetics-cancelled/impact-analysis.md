# VOC-090 — Impact Analysis

## Security and privacy

- **Secrets:** No new secrets. Session mint tokens and SSH keys remain in existing
  GitHub secrets; caching must not persist secret values in cache artifacts.
  Cache keys must be derived from lockfile / tool version paths only.
- **Personal data:** No change to data handling. Evidence must remain scrubbed of
  email addresses, session cookies, and OAuth state.
- **Observer:** `operational-failure-monitoring.yml` is unchanged. Successful
  remediation should allow issue #759 to close without weakening sanitization.

## Application and operational surface

- **Application code:** No intentional change. If `VOC-090-D03` applies, a follow-up
  package owns application/test performance work.
- **Workflow:** Only `scheduled-synthetics.yml` staging core-journey job gains
  caching and timeout alignment. Production jobs in the same workflow must behave
  identically unless a shared setup refactor is explicitly documented and tested.
- **Deploy gate parity:** `deploy-staging.yml` post-deploy core-loop gate remains
  the deploy-time authority; scheduled synthetics must not diverge in journey
  coverage.

## Data and migrations

- No application schema migration.
- No database mutation.
- Rollback reverts workflow/registry commits; the next hourly run may timeout again
  until remerged.

## Analytics and accessibility

- No analytics change — evidence-backed non-applicability.
- No product UI change — evidence-backed non-applicability.
- No accessibility change — evidence-backed non-applicability.

## Risks, dependencies, and evidence

- `VOC-090-R00`: **Stale cache serves wrong dependencies** after lockfile bump.
  Mitigation: cache keys must include lockfile hash / pnpm version; cold miss on
  lockfile change is acceptable.
- `VOC-090-R01`: **Timeout increase without caching** masks recurring cold-start
  cost every hour. Mitigation: `VOC-090-D02` requires caching as primary fix;
  timeout alignment is secondary.
- `VOC-090-R02`: **Journey regression hidden by longer timeout** if staging is
  genuinely broken and slow. Mitigation: keep Playwright `timeout: 240_000` and
  `retries: 0`; only the GitHub Actions job wall clock changes.
- `VOC-090-DEP-00`: Resolved — run 32271016931 public metadata identifies failing
  job and 30m cancellation.
- `VOC-090-DEP-01`: VOC-086 scheduled synthetics + VOC-050 staging core-loop spec.
- `VOC-090-EV-00`: T00 evidence — root-cause phase breakdown, caching/timeout diff,
  deterministic test output.
- `VOC-090-EV-01`: T01 evidence — green live `workflow_dispatch` run URL and duration.
