# VOC-128 — Tasks

None of the tasks below is implementation-authorized by this package.
Adoption and each task's own implementation authorization are separate.
Order is mandatory: **T00 → T01**.

This package uses two tasks because live GitHub Actions proof cannot exist in
the T00 carrier. File count, workflow-versus-tests-versus-docs, and
implementation convenience are not split reasons.

## VOC-128-T00 — Gate production synthetics to main and dispatch them from the develop schedule

- Requirement source: issue #1037; `VOC-128-D00`–`D06`
- Acceptance criteria: `VOC-128-AC-00`, `VOC-128-AC-01`, `VOC-128-AC-02`,
  `VOC-128-AC-03`
- Tests: `VOC-128-TEST-00`–`VOC-128-TEST-05`, regression on `VOC-086-TEST-11`,
  `VOC-086-TEST-12`, `VOC-090-TEST-*` budget tests
- Evidence: `VOC-128-EV-00` (`t00-evidence.md` in this package directory)
- Status: pending

### Required work

1. Record in `t00-evidence.md` the public metadata for run 33039500346 (three
   production jobs failed with empty steps and `runner_id` 0; both staging
   jobs succeeded). Do not copy secrets, session values, OAuth state, logs, or
   personal data.
2. Update `.github/workflows/scheduled-synthetics.yml`:
   - Production jobs keep `environment: production` and run only when
     `github.ref == 'refs/heads/main'` (plus the existing `synthetic_id`
     filter). Off-`main` they skip, not fail.
   - Add a `workflow_dispatch` boolean input (preferred name
     `production_jobs_only`, default false) that skips staging jobs when true.
     An equivalent production-only filter is acceptable if tests and docs use
     that exact identity.
   - Add job `name: dispatch-production-synthetics-on-main` that runs on
     `schedule` when the workflow ref is not `main`, with job-level
     `permissions: { contents: read, actions: write }`, and dispatches this
     workflow at `main` with the production-only filter. Use `GITHUB_TOKEN`.
     Do not use the observer App token. Do not grant the dispatcher production
     environment secrets.
3. Extend `infra/monitoring/scheduled-synthetics.mjs` and
   `scripts/foundation/voc086-scheduled-synthetics.test.mjs` so a missing
   `main` guard, a staging job with `environment: production`, a missing
   dispatcher, or a full-suite dispatch that would start staging core-journey
   fails the suite.
4. Update `docs/operations/monitoring.md` to describe the develop-schedule
   skip-and-dispatch behavior and that production `synthetic_id` dispatch
   remains `--ref main`.
5. Run:
   - `node --test scripts/foundation/voc086-scheduled-synthetics.test.mjs`
   - `node --test scripts/foundation/voc090-scheduled-synthetics-budget.test.mjs`
   - `bash scripts/governance/validate-governance.sh`
   - `bash scripts/governance/classify-change-risk.sh`
   - `git diff --check`

### Explicitly out of scope for this task

- Live Actions proof (T01).
- Changing GitHub Environment `production` deployment-branch settings.
- Changing `operational-failure-monitoring.yml`, Kuma inventory, mint-token
  names, OAuth assertions, or route-sweep coverage.
- VOC-093 harness edits or VOC-090 timeout edits unless this run class is
  disproven (it is not).

## VOC-128-T01 — Record live verification that develop schedule succeeds and production jobs run on main

- Requirement source: issue #1037; `VOC-128-D03`, `VOC-128-D07`
- Acceptance criteria: `VOC-128-AC-04`, `VOC-128-AC-05`
- Tests: `VOC-128-TEST-06`, `VOC-128-TEST-07`
- Evidence: `VOC-128-EV-01` (`t01-evidence.md`)
- Live-evidence contract: `.karsift/live-evidence/VOC-128-T01.yaml`
  (operator-owned; governed reconcile per `docs/operations/live-evidence.md`)
- Status: pending — depends on `VOC-128-T00`
- Split reason: post-merge-evidence-not-in-carrier — a qualifying scheduled-synthetics run on develop can exist only after T00 merges, and the implementer has no general Actions dispatch credentials
- Automation ownership: operator

### Required work

1. After T00 merges to `develop`, retarget
   `.karsift/live-evidence/VOC-128-T01.yaml` `sha_lineage` to `exact_sha` of
   that merge (or the develop SHA actually dispatched that contains T00). Do
   not wait on plan-time `integration_contains_pr_head` (VOC-110 lesson).
2. Enter the operator-owned waiting state. Acceptance requires
   **operator-owned live evidence**, not implementer Actions access.
   Repository-controlled reconcile may dispatch `scheduled-synthetics.yml` on
   `develop` using the contract `dispatch` block.
3. Confirm develop-run conclusion `success` and success for
   `synthetic.staging.oauth-expected-state`,
   `synthetic.staging.authenticated-core-journey`, and
   `dispatch-production-synthetics-on-main`. Production jobs on that run must
   be skipped, not failed.
4. Confirm the follow-on `main` run's three production jobs succeeded. Record
   run IDs, branches, SHAs, conclusions, and durations only.
5. Confirm no new open issue exists with marker
   `<!-- operational-failure:scheduled-synthetics:failure -->` beyond
   issue #1037.

### Explicitly out of scope for this task

- Code changes (T00 owns all fixes).
- Implementer-owned Actions dispatch or log inspection.
- Closing issue #1037 manually outside the normal package roster closure path.

## Task ordering notes

- T00 blocks T01: live proof requires the skip-and-dispatch workflow on
  `develop`.
- No task may be dispatched before this package is adopted and
  implementation-authorized.

Tasks preserve scope, separation of duties, and rollback safety.
