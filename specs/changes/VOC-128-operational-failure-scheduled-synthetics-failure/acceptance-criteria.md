# VOC-128 — Acceptance Criteria

## VOC-128-AC-00 — Root cause recorded and bounded to the production environment gate

- Requirement source: `VOC-128-D00`, `VOC-128-D01`
- Tasks: `VOC-128-T00`
- Tests: `VOC-128-TEST-00`
- Evidence: `VOC-128-EV-00`
- Result: pending

Task evidence identifies run 33039500346, names the three production jobs that
failed with empty steps and `runner_id` 0, names the two staging jobs that
succeeded with assigned runners, and explains why default-branch `develop`
plus `environment: production` produces that class. No secrets, session
values, OAuth state, or personal data are copied.

## VOC-128-AC-01 — Production jobs skip off main and run only on main

- Requirement source: `VOC-128-D02`, `VOC-128-D05`
- Tasks: `VOC-128-T00`
- Tests: `VOC-128-TEST-01`, `VOC-128-TEST-02`
- Evidence: `VOC-128-EV-00`
- Result: pending

Each production synthetic job still declares `environment: production` and
executes only when `github.ref` is `refs/heads/main` (combined with the
existing `synthetic_id` filter). On any other ref those jobs are skipped, not
failed. Staging jobs do not declare `environment: production`.

## VOC-128-AC-02 — Develop default-branch schedule dispatches production checks onto main

- Requirement source: `VOC-128-D03`, `VOC-128-D04`
- Tasks: `VOC-128-T00`
- Tests: `VOC-128-TEST-03`, `VOC-128-TEST-04`
- Evidence: `VOC-128-EV-00`
- Result: pending

A `schedule` event on a non-`main` ref includes job
`dispatch-production-synthetics-on-main`, which uses job-level `GITHUB_TOKEN`
(`actions: write` on that job only) to `workflow_dispatch`
`scheduled-synthetics.yml` at `main` for production checks only. The dispatch
must not start `synthetic.staging.authenticated-core-journey`. The observer
App token is not used.

## VOC-128-AC-03 — Deterministic tests and docs cover the ref gate

- Requirement source: `VOC-128-D06`
- Tasks: `VOC-128-T00`
- Tests: `VOC-128-TEST-00` through `VOC-128-TEST-05`, regression on
  `VOC-086-TEST-11`, `VOC-086-TEST-12`
- Evidence: `VOC-128-EV-00`
- Result: pending

`node --test scripts/foundation/voc086-scheduled-synthetics.test.mjs` passes.
New or extended tests lock the `main`-only production guard, the dispatcher
job, and the production-only dispatch filter. `docs/operations/monitoring.md`
describes the develop-schedule versus `main` production split.

## VOC-128-AC-04 — Live develop schedule completes success after T00

- Requirement source: issue #1037 remediation outcome; `VOC-128-D03`,
  `VOC-128-D07`
- Tasks: `VOC-128-T01`
- Tests: `VOC-128-TEST-06`
- Evidence: `VOC-128-EV-01`
- Result: pending

After T00 merges to `develop`, a repository-controlled `schedule` or
`workflow_dispatch` of `scheduled-synthetics.yml` on `develop` at a SHA that
contains T00 completes with conclusion `success`. Required successful jobs
are `synthetic.staging.oauth-expected-state`,
`synthetic.staging.authenticated-core-journey`, and
`dispatch-production-synthetics-on-main`. Production jobs on that run are
skipped, not failed. Evidence records allowlisted run metadata only.

## VOC-128-AC-05 — Live production jobs succeed on main without a duplicate observer issue

- Requirement source: `VOC-128-D03`; `VOC-088-D06` observer contract preserved
- Tasks: `VOC-128-T01`
- Tests: `VOC-128-TEST-07`
- Evidence: `VOC-128-EV-01`
- Result: pending

The `main` run created by the dispatcher (or an equivalent
`workflow_dispatch --ref main` of the production checks) completes with
conclusion `success` and success for
`synthetic.production.oauth-expected-state`,
`synthetic.production.journey-content`, and
`synthetic.production.authenticated-route-content-sweep`. No new open issue
exists with marker `<!-- operational-failure:scheduled-synthetics:failure -->`
beyond issue #1037. Closure of #1037 follows normal roster/package closure.

Acceptance criteria must be observable, stable, security-aware, and bidirectionally
traceable to requirements, tasks, tests, and evidence.
