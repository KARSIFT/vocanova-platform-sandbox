# VOC-094-T00 evidence — root cause and remediation inputs

Recorded at implementation time from the public GitHub Actions REST API
(no authentication required for this repository's public run metadata).

## Run 32290409156 (deploy-staging #299)

| Field | Value |
|-------|-------|
| Run URL | https://github.com/KARSIFT/vocanova-platform-sandbox/actions/runs/32290409156 |
| Workflow | `deploy-staging` (run number 299) |
| Event | `push` |
| Head branch | `develop` |
| Head SHA | `411fb60157d437e495fc07f599e268669f139e5a` |
| Conclusion | `cancelled` |
| Status | `completed` |
| Created at | `2026-08-19T18:58:36Z` |
| Updated at | `2026-08-19T19:01:04Z` |
| Wall-clock duration | ~2m 28s (created → updated) |
| Jobs API `total_count` | `0` (no jobs started) |

## Classification

This run is a **pending-run concurrency supersession**, not a deploy-step
failure:

- Concurrency group at drafting time: `staging-deploy` with
  `cancel-in-progress: false` and no `queue: max` (GitHub's default
  single-pending-slot queue).
- Public workflow annotation (from issue #781 / Checks UI):
  `Canceling since a higher priority waiting request for staging-deploy exists`.
- Zero jobs started — the runner never scheduled the `deploy` job before
  GitHub cancelled the pending run when a newer push entered the group.

## Remediation applied in T00

1. Add `queue: max` to `deploy-staging.yml` and `deploy-production.yml` so
   superseded pending deploys queue instead of cancelling.
2. Add `classify-deploy-concurrency-cancel.sh` and wire it in
   `operational-failure-monitoring.yml` so `cancelled` deploy conclusions
   with zero jobs do not open governed issues (fail-closed on API errors or
   non-zero job counts).
3. Mint the App installation token with `permission-actions: read` alongside
   `permission-issues: write` so the classifier's bounded jobs API call is
   authorized in live runs (remediation for independent-review finding on
   commit `064b9ca1`).

No deploy script, health-check, or SSH semantics were changed.

## Deterministic validation (VOC-094-EV-00 / TEST-06 regression)

Recorded at remediation time (`2026-08-19`):

```text
$ node --test scripts/foundation/voc094-deploy-concurrency.test.mjs
✔ VOC-094-TEST-00 through TEST-05 (7/7 pass)

$ node --test scripts/foundation/voc088-failure-to-issue.test.mjs
✔ VOC-088-TEST-08 through TEST-11 (4/4 pass)

$ node --test scripts/foundation/voc084-deploy-staging-oauth.test.mjs
✔ VOC-084 deploy-staging OAuth regression (9/9 pass)

$ node --test scripts/foundation/voc088-deploy-staging-allowlist.test.mjs
✔ VOC-088 deploy-staging allowlist regression (7/7 pass)
```

All suites exited 0; no deploy step removal, `continue-on-error`, or
health/core-loop ordering changes observed in deploy workflow diffs.
