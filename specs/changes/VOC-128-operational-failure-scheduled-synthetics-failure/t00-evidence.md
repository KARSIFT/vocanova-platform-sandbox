# VOC-128-T00 — Evidence

Task: `VOC-128-T00` — Gate production synthetics to main and dispatch them
from the develop schedule.

Do not record secrets, credentials, session values, OAuth material, personal
data, complete CI logs, or App token values.

## Discovery recorded at planning time (issue #1037)

| Item | Value |
|------|-------|
| Requirement source | GitHub issue #1037 |
| Workflow | `scheduled-synthetics` |
| Actions run | `33039500346` (run number 200) |
| Event | `schedule` at `2026-08-27T04:28:07Z` |
| Head branch / SHA | `develop` @ `0d0b0cdf0692d0349f380e9cae3285b4c7916b05` |
| Default branch | `develop` |
| Workflow conclusion | `failure` (updated `2026-08-27T04:29:01Z`) |
| `synthetic.staging.oauth-expected-state` | success; job `98409640160`; runner assigned |
| `synthetic.staging.authenticated-core-journey` | success; job `98409640102`; runner assigned; Playwright steps present |
| `synthetic.production.oauth-expected-state` | failure; job `98409640172`; empty steps; `runner_id` 0 |
| `synthetic.production.journey-content` | failure; job `98409640167`; empty steps; `runner_id` 0 |
| `synthetic.production.authenticated-route-content-sweep` | failure; job `98409640174`; empty steps; `runner_id` 0 |
| Failure class | GitHub Environment `production` gate on a non-`main` default-branch schedule, not a smoke-test exit 1 |
| Distinct from | VOC-093 (route-sweep steps ran on `main`); VOC-090 (staging timeout/cancel) |
| Deferred by | VOC-108 (scheduled-synthetics branch-selection footgun listed as an exclusion) |

## Chosen delivery path

| Item | Constraint |
|------|------------|
| Production jobs | keep `environment: production`; add `main`-only ref guard; skip off-`main` |
| Hourly production coverage | dispatcher job `dispatch-production-synthetics-on-main` uses job-level `GITHUB_TOKEN` to `workflow_dispatch` this workflow at `main` |
| Staging core-journey | must not run a second time because of that dispatch |
| Observer App | no `actions: write` (VOC-098) |
| Production environment settings | do not widen deployment branches to `develop` |
| Docs | update `docs/operations/monitoring.md` in the T00 PR |

## Changed surfaces (to be filled at implementation)

- `.github/workflows/scheduled-synthetics.yml`
- `infra/monitoring/scheduled-synthetics.mjs`
- `scripts/foundation/voc086-scheduled-synthetics.test.mjs`
- `docs/operations/monitoring.md`
- this evidence file

## Validation commands

```bash
node --test scripts/foundation/voc086-scheduled-synthetics.test.mjs
node --test scripts/foundation/voc090-scheduled-synthetics-budget.test.mjs
bash scripts/governance/validate-governance.sh
bash scripts/governance/classify-change-risk.sh
git diff --check
```

| Command | Result |
|---------|--------|
| `node --test scripts/foundation/voc086-scheduled-synthetics.test.mjs` | pending implementation |
| `node --test scripts/foundation/voc090-scheduled-synthetics-budget.test.mjs` | pending implementation |
| `bash scripts/governance/validate-governance.sh` | pending implementation |
| `bash scripts/governance/classify-change-risk.sh` | pending implementation |
| `git diff --check` | pending implementation |

## Acceptance mapping

- `VOC-128-AC-00` / `VOC-128-EV-00` — planning-time diagnosis recorded; implementation still pending
- `VOC-128-AC-01` / `VOC-128-EV-00` — pending implementation
- `VOC-128-AC-02` / `VOC-128-EV-00` — pending implementation
- `VOC-128-AC-03` / `VOC-128-EV-00` — pending implementation
