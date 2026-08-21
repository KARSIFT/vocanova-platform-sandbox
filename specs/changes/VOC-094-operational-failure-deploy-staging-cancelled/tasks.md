# VOC-094 — Tasks

None of the tasks below is implementation-authorized by this package.
Adoption and each task's own implementation authorization are separate.
Order is mandatory: **T00 → T01**.

## VOC-094-T00 — Fix deploy concurrency queue and benign-cancel observer filter

- Requirement source: issue #781; `VOC-094-D00`–`D05`
- Acceptance criteria: `VOC-094-AC-00` through `VOC-094-AC-04`
- Tests: `VOC-094-TEST-00` through `VOC-094-TEST-06`, regression on
  `VOC-088-TEST-08`–`VOC-088-TEST-11`
- Evidence: `VOC-094-EV-00` (`t00-evidence.md` in this package directory)
- Status: pending

### Required work

1. Read run 32290409156 metadata at implementation time (public run page and/or
   `gh api` / Actions REST). Record in `t00-evidence.md`: conclusion, duration,
   job count, concurrency annotation, and head SHA. Do not copy secrets, SSH
   output, session values, or personal data.
2. Add `queue: max` under `concurrency` in `.github/workflows/deploy-staging.yml`.
   Keep `group: staging-deploy` and `cancel-in-progress: false`. Update the workflow
   header comment to document why `queue: max` coexists with
   `cancel-in-progress: false`.
3. Add the same `queue: max` under `concurrency` in `.github/workflows/deploy-production.yml`
   (`production-deploy` group, `cancel-in-progress: false` unchanged).
4. Implement benign concurrency-supersession classification for deploy workflows:
   - Prefer a small helper under `infra/scripts/` (new or extended
     `open-failure-issue.sh`) invoked from
     `.github/workflows/operational-failure-monitoring.yml` **before** issue creation.
   - Classifier input: workflow name, conclusion, run id — bounded API reads only.
   - Skip issue creation when metadata indicates pending-run supersession (minimum:
     `cancelled` + zero jobs for deploy workflows; strengthen with annotation substring
     when API exposes it).
   - Fail closed (`VOC-094-D03`): ambiguous or API-error paths keep opening issues.
5. Add `scripts/foundation/voc094-deploy-concurrency.test.mjs` asserting queue wiring
   on both deploy workflows and classifier skip/create behavior via fixtures.
6. Extend `scripts/foundation/voc088-failure-to-issue.test.mjs` only as needed for
   new skip-path coverage without weakening existing sanitization/deduplication tests.
7. Run new VOC-094 tests, VOC-088 failure-to-issue tests, and applicable deploy
   foundation tests; run governance validation if required for changed paths.

### Explicitly out of scope for this task

- Live supersession reproduction (T01).
- Changing `scheduled-synthetics.yml` or synthetic registry timeouts.
- Application, migration, or Kuma inventory changes.
- Setting `cancel-in-progress: true` on deploy workflows.

## VOC-094-T01 — Record live verification that superseded cancels stop opening issues

- Requirement source: issue #781; `VOC-094-D00`, `VOC-094-D03`
- Acceptance criteria: `VOC-094-AC-05`, `VOC-094-AC-06`
- Tests: `VOC-094-TEST-07`
- Evidence: `VOC-094-EV-01` (`t01-evidence.md`)
- Live-evidence contract: `.karsift/live-evidence/VOC-094-T01.yaml` (operator-owned;
  governed reconcile per `docs/operations/live-evidence.md`)
- Status: pending — depends on `VOC-094-T00`

### Required work

1. After T00 merges to `develop`, confirm a `deploy-staging` run for the latest
   integration commit completes with conclusion `success`. Record run URL and duration
   (no secrets).
2. Demonstrate benign supersession handling using one of:
   - **Preferred:** controlled triple-push or merge sequence on `develop` that
     produces a superseded pending cancel without blocking the latest deploy; record
     scrubbed run IDs and confirm no **new** open issue with marker
     `<!-- operational-failure:deploy-staging:cancelled -->` beyond issue #781; or
   - **Fallback (open question 3):** deterministic fixture proof from T00 plus green
     latest deploy, documented explicitly in `t01-evidence.md`.
3. Note whether issue #781 can close under normal roster closure after verification.

### Explicitly out of scope for this task

- Code changes (T00 owns all workflow/script edits).
- Manual issue closure outside the governed roster path.

## Task ordering notes

- T00 blocks T01: queue and classifier changes must be on `develop` before live proof.
- No task may be dispatched before this package is adopted and
  implementation-authorized.

Tasks preserve scope, separation of duties, and rollback safety.
