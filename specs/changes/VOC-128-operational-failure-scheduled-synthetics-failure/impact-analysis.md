# VOC-128 — Impact Analysis

## Security and privacy

- **Secrets:** No new secrets. Production mint tokens stay in GitHub
  Environment `production`. The dispatcher job receives only `GITHUB_TOKEN`
  with job-level `actions: write`; it must not read production mint tokens,
  SSH keys, or session cookies.
- **Environment boundary:** Production jobs keep `environment: production` and
  run only on `main`. Widening that environment's deployment branches to
  `develop` is out of scope and forbidden.
- **App tokens:** The VOC-098 observer App remains without Actions write.
  Same-repo `workflow_dispatch` uses job-level `GITHUB_TOKEN` only.
- **Personal data:** Evidence stays scrubbed. No job logs, session values,
  OAuth state, cookies, or user identifiers.
- **Observer:** `operational-failure-monitoring.yml` is unchanged.

## Data and migrations

None. No database, schema, seed, or migration changes are proposed. Staging
core-journey mutation of the reserved synthetic account is unchanged and must
not run twice per hourly tick because of the dispatcher.

## Analytics and accessibility

None for product/runtime behavior. There is no analytics instrumentation or
user-interface accessibility effect.

## Risks, dependencies, and evidence

- `VOC-128-R00`: **High false-red risk** if production jobs still fail the
  develop-default schedule at the environment gate. Mitigation: `VOC-128-D02`,
  `VOC-128-AC-01`.
- `VOC-128-R01`: **High coverage-loss risk** if production jobs only skip on
  develop and are never dispatched onto `main`. Mitigation: `VOC-128-D03`,
  `VOC-128-AC-02`, `VOC-128-AC-05`.
- `VOC-128-R02`: **High privilege-expansion risk** if Actions write is granted
  to the observer App, or if the dispatcher inherits production secrets.
  Mitigation: `VOC-128-D03`, `VOC-128-AC-02`.
- `VOC-128-R03`: **High secret-exposure risk** if production environment branch
  policy is widened to `develop` instead of skipping-and-dispatching.
  Mitigation: `VOC-128-D02` and the explicit non-goal.
- `VOC-128-R04`: **Medium CI-budget risk** if the dispatcher starts a second
  staging core-journey (VOC-090 40-minute job) on `main` every hour.
  Mitigation: `VOC-128-D04`.
- `VOC-128-R05`: **Medium live-evidence deadlock risk** if T01 waits on
  `integration_contains_pr_head` against an unmerged evidence PR (VOC-110).
  Mitigation: `VOC-128-D07`.
- `VOC-128-DEP-00` through `VOC-128-DEP-02`, plus VOC-088, VOC-086, VOC-098,
  and VOC-110: see `change.yaml`.
- `VOC-128-EV-00`: T00 evidence — public run metadata, workflow diff, tests.
- `VOC-128-EV-01`: T01 evidence — allowlisted develop-success and main
  production-success run metadata.

## Monitoring

`monitoring_impact.state: existing` — the five canonical synthetic IDs already
cover this change. No new monitor or synthetic IDs. The package remediates
how those existing production synthetics are scheduled.

## Authority and EHR

Active authority model: **A-004**. No founder `approved` comment is required for
engineering-workflow adopt/merge/release gates. EHR is not triggered. This
draft does not adopt or authorize itself.
