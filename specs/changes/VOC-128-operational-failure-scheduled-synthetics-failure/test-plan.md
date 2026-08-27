# VOC-128 — Test Plan

## VOC-128-TEST-00 — Public failure metadata matches the environment-gate class

- Covers: `VOC-128-AC-00`
- Preconditions: public Actions jobs API for run 33039500346 is readable;
  no job logs required
- Procedure: compare staging vs production job conclusions, step lists, and
  `runner_id` values recorded in `t00-evidence.md`
- Expected result: three production jobs failed with empty steps and
  `runner_id` 0; both staging jobs succeeded with assigned runners
- Evidence: `VOC-128-EV-00`

## VOC-128-TEST-01 — Production jobs are main-only and keep environment production

- Covers: `VOC-128-AC-01`
- Preconditions: T00 workflow file is in the task diff
- Procedure: parse `.github/workflows/scheduled-synthetics.yml` via
  `extractTopLevelJobBlock` in `infra/monitoring/scheduled-synthetics.mjs`;
  assert each production job block contains `environment: production` and a
  `github.ref` / `refs/heads/main` conjunct in its `if:`
- Expected result: off-`main` production jobs are skipped; they still request
  the production environment when they run
- Evidence: `VOC-128-EV-00`

## VOC-128-TEST-02 — Staging jobs do not declare environment production

- Covers: `VOC-128-AC-01`
- Preconditions: same as TEST-01
- Procedure: assert staging job blocks omit `environment: production` and still
  match VOC-086 stable IDs, mint-token reuse, and OAuth expected-state
  `EXPECT_OAUTH_ENABLED: "true"`
- Expected result: staging schedule on `develop` can start a runner without
  the production environment gate
- Evidence: `VOC-128-EV-00`

## VOC-128-TEST-03 — Dispatcher job exists on non-main schedule events

- Covers: `VOC-128-AC-02`
- Preconditions: same as TEST-01
- Procedure: assert a job whose `name:` is
  `dispatch-production-synthetics-on-main` runs when `github.event_name` is
  `schedule` and `github.ref` is not `refs/heads/main`; assert job-level
  `actions: write`; assert it invokes `workflow_dispatch` of
  `scheduled-synthetics.yml` at `main`
- Expected result: develop-default cron can create a `main` production run
  without the observer App token
- Evidence: `VOC-128-EV-00`

## VOC-128-TEST-04 — Dispatcher does not start staging core-journey

- Covers: `VOC-128-AC-02`
- Preconditions: same as TEST-01
- Procedure: assert the dispatcher sets the production-only filter (preferred
  input `production_jobs_only: true`, or the adopted equivalent). Assert
  staging jobs skip when that filter is true, including
  `synthetic.staging.authenticated-core-journey`
- Expected result: hourly CI does not run two 40-minute staging core-journeys
- Evidence: `VOC-128-EV-00`

## VOC-128-TEST-05 — VOC-086 wiring and VOC-090 budget regressions stay green

- Covers: `VOC-128-AC-03`
- Preconditions: Node test runner available
- Procedure:

  ```bash
  node --test scripts/foundation/voc086-scheduled-synthetics.test.mjs
  node --test scripts/foundation/voc090-scheduled-synthetics-budget.test.mjs
  bash scripts/governance/validate-governance.sh
  bash scripts/governance/classify-change-risk.sh
  git diff --check
  ```

- Expected result: all listed commands exit 0. New VOC-128 assertions are
  part of the voc086 suite or an adjacent voc128 foundation file invoked from
  the same command. Tests must not use secrets or production data.
- Evidence: `VOC-128-EV-00`

## VOC-128-TEST-06 — Live develop schedule succeeds after T00

- Covers: `VOC-128-AC-04`
- Preconditions: T00 merged to `develop`; operator-owned live-evidence
  contract retargeted per `VOC-128-D07`
- Procedure: repository-controlled dispatch or next hourly `schedule` of
  `scheduled-synthetics.yml` on `develop`; inspect conclusion and job names
  only
- Expected result: workflow conclusion `success`; staging oauth, staging
  core-journey, and dispatcher jobs success; production jobs skipped
- Evidence: `VOC-128-EV-01`

## VOC-128-TEST-07 — Live production jobs succeed on main

- Covers: `VOC-128-AC-05`
- Preconditions: dispatcher from TEST-06 created a `main` run, or an
  equivalent `workflow_dispatch --ref main` of production checks exists
- Procedure: inspect the `main` run conclusion and the three production job
  names only; search open issues for the stable operational-failure marker
- Expected result: production jobs success; issue #1037 remains the sole open
  `scheduled-synthetics:failure` fingerprint owner
- Evidence: `VOC-128-EV-01`

Include positive, negative, authorization, failure, and rollback coverage as
applicable. Tests must not use secrets or production data.
