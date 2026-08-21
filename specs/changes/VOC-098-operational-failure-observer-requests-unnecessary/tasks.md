# VOC-098 — Tasks

None of the tasks below is implementation-authorized by this package.
Adoption and each task's own implementation authorization are separate.
Order is mandatory: **T00 → T01**.

## VOC-098-T00 — Least-privilege App token + classifier token split and deterministic tests

- Requirement source: issue #840; `VOC-098-D00`–`D03`, `D05`
- Acceptance criteria: `VOC-098-AC-00`, `VOC-098-AC-01`, `VOC-098-AC-02`,
  `VOC-098-AC-05`
- Tests: `VOC-098-TEST-00` through `VOC-098-TEST-04`, `VOC-098-TEST-07`,
  regression `VOC-088-TEST-11`, `VOC-094-TEST-05` and related VOC-088/VOC-094
  foundation suites
- Evidence: `VOC-098-EV-00` (`t00-evidence.md` in this package directory)
- Status: pending

### Required work

1. Update `.github/workflows/operational-failure-monitoring.yml` so the App mint
   step requests **only** `permission-issues: write` (remove
   `permission-actions: read` and any unused permission).
2. Wire `infra/scripts/classify-deploy-concurrency-cancel.sh` to the job
   `GITHUB_TOKEN` (or equivalent non-App job token) and ensure the workflow/job
   `permissions` floor includes `actions: read` (retain `contents: read`). Keep
   the open/dedupe step on `steps.app-token.outputs.token` only.
3. Update header/step comments so they no longer claim the App token carries
   Actions read.
4. Update `scripts/foundation/voc088-failure-to-issue.test.mjs` and
   `scripts/foundation/voc094-deploy-concurrency.test.mjs` (and add a focused
   `voc098-*.test.mjs` only if that is cleaner) so tests **fail** if the App mint
   still requests Actions, and **pass** when issue creation remains App-token-only
   and the classifier is not fed the App token for Actions API reads.
5. Update `docs/operations/staging-controlled-signup.md` only if its observer
   section would otherwise remain false about token roles.
6. Run the updated foundation tests and applicable governance validation for
   changed paths; record commands and results in `t00-evidence.md` (no secrets).

### Explicitly out of scope for this task

- Live controlled failure/cancel proof (T01).
- Changing App installation permissions in GitHub settings.
- Changing watched workflows, marker format, or Sentry/Kuma observers.
- Application or monitoring-inventory ID changes.

## VOC-098-T01 — Controlled live observer proof (operator-owned live evidence)

- Requirement source: issue #840 acceptance; `VOC-098-D04`
- Acceptance criteria: `VOC-098-AC-03`, `VOC-098-AC-04`
- Tests: `VOC-098-TEST-05`, `VOC-098-TEST-06`
- Evidence: `VOC-098-EV-01` (`t01-evidence.md`)
- Status: pending — depends on `VOC-098-T00`
- **Acceptance requires operator-owned live evidence** (not implementer Actions
  access). Contract:
  `.karsift/live-evidence/VOC-098-T01.yaml`.

### Required work

1. After T00 merges and is live on the branch the observer executes from
   (expected: default branch / `main` after normal develop→main promotion — see
   open question 4), perform a **controlled** watched non-success using the
   existing fixture in `docs/operations/staging-controlled-signup.md`
   (preferred: dispatch then cancel
   `synthetic.staging.authenticated-core-journey` on `scheduled-synthetics.yml`).
   Do not break application, deploy, or observer files to manufacture failure.
2. Confirm `operational-failure-monitoring` runs for that `workflow_run` and
   concludes `success` (token mint succeeds).
3. Confirm exactly one open sanitized App-authored unlabeled issue owns marker
   `<!-- operational-failure:scheduled-synthetics:cancelled -->` (create or
   dedupe). Re-run or re-cancel to the same fingerprint and confirm **no**
   duplicate open issue.
4. Record allowlisted metadata only in `t01-evidence.md` (run IDs/URLs,
   conclusions, issue number, marker, App authorship indicator). Never copy
   logs, secrets, sessions, OAuth data, cookies, tokens, or user identifiers.
5. Do not expand scope into unrelated workflow or pipeline edits to manufacture
   evidence; waiting/reconcile are handled by governed automation after VOC-097.

### Explicitly out of scope for this task

- Code changes (T00 owns all workflow/script/test/doc edits).
- Granting implementer Actions credentials.
- Manual closure of unrelated open operational-failure issues outside roster
  hygiene.

## Task ordering notes

- T00 blocks T01: live proof requires the least-privilege mint fix to be what the
  default-branch observer actually runs.
- No task may be dispatched before this package is adopted and
  implementation-authorized.

Tasks preserve scope, separation of duties, and rollback safety.
