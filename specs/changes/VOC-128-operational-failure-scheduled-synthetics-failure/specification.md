# VOC-128 — Fix scheduled-synthetics production jobs failing on the develop default-branch schedule: Specification

## Objective and requirement source

Remediate the operational failure recorded in
[GitHub issue #1037](https://github.com/KARSIFT/vocanova-platform-sandbox/issues/1037):
the hourly `scheduled-synthetics` workflow ended `failure` because every job
that declares GitHub Environment `production` was rejected before a runner
started, while both staging jobs in the same run succeeded.

**This draft package does not adopt or authorize itself**; adoption remains a
separate A-004 plan-review / adopt path.

Primary evidence (issue #1037 + public run and jobs API for run 33039500346;
no job logs copied):

| Item | Value |
|------|-------|
| Workflow | `scheduled-synthetics` |
| Run | [#200](https://github.com/KARSIFT/vocanova-platform-sandbox/actions/runs/33039500346), schedule-triggered 2026-08-27 04:28 UTC |
| Head branch / SHA | `develop` @ `0d0b0cdf0692d0349f380e9cae3285b4c7916b05` |
| Default branch | `develop` |
| Conclusion | `failure` (workflow wall clock ~54s) |
| Staging jobs | both `success` with assigned runners and real steps |
| Production jobs | all three `failure` in ~2s, empty `steps`, `runner_id` 0 |
| Issue origin | VOC-088-T02 `operational-failure-monitoring.yml` sanitized issue body |

Drafting-time repo read:

- `.github/workflows/scheduled-synthetics.yml` production jobs
  (`production-oauth-expected-state`, `production-journey-content`,
  `production-authenticated-route-content-sweep`) set
  `environment: production`. Staging jobs do not set a GitHub Environment.
- GitHub `on.schedule` runs only on the repository default branch. This
  sandbox default is `develop`, so hourly production jobs always request the
  `production` environment from a non-`main` ref.
- `docs/operations/monitoring.md` already tells operators to dispatch this
  workflow with `--ref main`. The schedule path does not.

## Scope and non-goals

In scope:

1. Record run 33039500346 public metadata in `t00-evidence.md` without copying
   secrets, session values, OAuth state, or personal data.
2. Change `scheduled-synthetics.yml` so production jobs execute only when
   `github.ref` is `refs/heads/main`. Off-`main` they must be skipped, not
   failed.
3. From a `schedule` (and equivalent develop default-branch) run, dispatch the
   same workflow onto `main` for the three production stable IDs only, using
   job-level `GITHUB_TOKEN` with `actions: write` on that dispatcher job.
   Workflow-level permissions stay `contents: read` for every other job.
4. Do not start a second `synthetic.staging.authenticated-core-journey` as a
   side effect of that dispatch.
5. Extend `infra/monitoring/scheduled-synthetics.mjs` and
   `scripts/foundation/voc086-scheduled-synthetics.test.mjs` so the ref gate
   and dispatcher contract cannot regress silently.
6. Update `docs/operations/monitoring.md` so it describes the develop-schedule
   vs `main` production split. AGENTS.md requires every workflow-behavior
   change to update docs that describe that behavior in the same PR.
7. Live verification after T00 merges to `develop` (T01).

Non-goals / explicitly excluded:

- Widening GitHub Environment `production` deployment branches to include
  `develop`.
- Putting `actions: write` on the VOC-088/VOC-098 observer App token.
- Changing Kuma inventory, synthetic stable IDs, mint-token names, OAuth
  expected-state assertions, or route-sweep coverage.
- Application schema, product UI, signup policy, or deploy-production.yml.
- Re-opening VOC-093 harness fixes or VOC-090 staging timeout work unless T00
  proves those checks are the failing class (run 33039500346 is not).
- Self-adoption / self-authorization of this package.

## Risk and protected areas

- **Draft package proposal:** **R3** (`.github/workflows/`, `infra/`).
- **Measured path floor at drafting:** **R3**. Not R4 unless a task touches
  `scripts/governance/*` or `tooling/governance/*`.
- Protected areas: `.github/workflows/scheduled-synthetics.yml`, production
  environment secret bindings, synthetic session mint secrets (read-only use
  only).
- EHR: not triggered.
- Active authority model: **A-004**. No founder `approved` comment is a
  merge/adopt/release gate.

This risk class is a draft proposal for the reviewing human at adoption time,
never a determination.

## Decisions

`VOC-128-D00`: Run 33039500346 failed because the three production synthetic
jobs never started on `develop`. Empty steps and `runner_id` 0 bound this to a
GitHub Environment gate, not a smoke-test exit code.

`VOC-128-D01`: Staging success in the same run bounds the defect to production
job `environment: production` plus default-branch `develop`. Session mint,
Playwright install, and staging OAuth checks were healthy.

`VOC-128-D02`: Production jobs must keep `environment: production` and must
run only on `refs/heads/main`. Skipping on other refs is required. Expanding
production environment branch policy to `develop` is forbidden because it
would expose production secrets to the default-branch schedule.

`VOC-128-D03`: Hourly production coverage remains required (VOC-086). Because
cron cannot target `main` while the default branch is `develop`, the
develop-schedule run must dispatch `scheduled-synthetics.yml` onto `main` for
production checks only. Dispatcher identity is job `name`
`dispatch-production-synthetics-on-main`. Same-repo dispatch uses job-level
`GITHUB_TOKEN`; it must not use the observer App token.

`VOC-128-D04`: That dispatch must not start
`synthetic.staging.authenticated-core-journey` a second time. Preferred
mechanism is a `workflow_dispatch` boolean input defaulting to false, set true
by the dispatcher, that skips staging jobs. An equivalent production-only
filter is acceptable if tests and docs use that exact identity.

`VOC-128-D05`: Manual `workflow_dispatch` of a production `synthetic_id` must
also be `--ref main` (already documented). Off-`main` production jobs skip
rather than fail. Do not convert an operator dispatch on `develop` into a
workflow-level failure solely because production jobs are skipped.

`VOC-128-D06`: Deterministic tests must fail if any production job lacks a
`main`-only ref guard, if staging jobs gain `environment: production`, if the
dispatcher is missing on non-`main` `schedule` events, or if the dispatcher
requests the full five-check suite including staging core-journey.

`VOC-128-D07`: T01 live-evidence must not wait on
`sha_lineage.mode: integration_contains_pr_head` against an unmerged T01
evidence PR (VOC-110 lesson). After T00 merges, T01's first commit sets
`sha_lineage.mode: exact_sha` to the T00 merge SHA (or the develop SHA
actually dispatched that contains T00) before entering waiting.

## Open questions for the reviewing human

1. Accept proposed **R3**, or raise in writing if dispatching `scheduled-synthetics`
   onto `main` from develop is treated as R4 operational risk.
2. Accept `production_jobs_only` as the preferred dispatch input name, or record
   an equivalent production-only filter identity during adoption.
3. Confirm production environment deployment branches remain `main`-only and
   must not be widened. This package does not change repository environment
   settings.

## Data, migrations, analytics, and accessibility

- No application schema migration.
- No intentional database mutation from this package. Staging core-journey
  remains the only mutating synthetic and is unchanged in behavior; it simply
  must not run twice per hour because of the dispatcher.
- No analytics change. Evidence-backed non-applicability.
- No accessibility impact. Evidence-backed non-applicability.
