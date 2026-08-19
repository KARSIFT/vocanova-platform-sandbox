# VOC-094 — Impact Analysis

## Security and privacy

- **Secrets:** No new secrets. Classifier API calls use the existing automation App
  installation token already required by VOC-088-T02. Classifier must not persist or
  echo secret values, SSH output, session cookies, or OAuth state in issues or logs.
- **Personal data:** No change to data handling. Evidence remains scrubbed.
- **Observer:** Narrows which `cancelled` deploy conclusions open issues. Must not
  weaken sanitization, deduplication, or App-token identity for real failures.

## Application and operational surface

- **Application code:** No intentional change.
- **Workflow:** `deploy-staging.yml` and `deploy-production.yml` gain `queue: max`
  only. Deploy steps, health checks, OAuth guards, and core-loop ordering stay the same.
- **Queue depth risk:** Up to GitHub's documented pending limit, multiple deploys may
  queue during rapid merges. Each still runs serially (`cancel-in-progress: false`);
  staging/production hosts continue to receive one deploy at a time.

## Data and migrations

- No application schema migration.
- No database mutation.
- Rollback reverts workflow/script commits; rapid merges may again cancel pending
  deploys and reopen operational-failure noise until remerged.

## Analytics and accessibility

- No analytics change — evidence-backed non-applicability.
- No product UI change — evidence-backed non-applicability.
- No accessibility change — evidence-backed non-applicability.

## Risks, dependencies, and evidence

- `VOC-094-R00`: **Classifier false negative** skips a real deploy cancellation.
  Mitigation: fail-closed on ambiguous metadata (`VOC-094-D03`); require zero jobs
  and/or stable annotation substring; extend tests for failure and in-progress cancel
  fixtures.
- `VOC-094-R01`: **Deep deploy queue** during incident merges delays latest staging.
  Mitigation: serial deploy semantics unchanged; queue only replaces silent cancel;
  operators can still `workflow_dispatch` retry the latest revision.
- `VOC-094-R02`: **Production queue parity** increases pending production deploy count
  during rapid promotion bursts. Mitigation: same serial execution; production
  concurrency group remains separate from staging.
- `VOC-094-DEP-00`: Resolved — run 32290409156 public metadata identifies
  concurrency supersession with zero jobs.
- `VOC-094-DEP-01`: VOC-032 deploy-staging concurrency design + VOC-088 observer.
- `VOC-094-EV-00`: T00 evidence — root-cause metadata, queue/classifier diff,
  deterministic test output.
- `VOC-094-EV-01`: T01 evidence — green latest deploy run URL and supersession hygiene proof.
