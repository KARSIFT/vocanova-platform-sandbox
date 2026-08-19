---
evidence_id: VOC-088-EV-01
task_id: VOC-088-T01
acceptance_criteria:
  - VOC-088-AC-03
  - VOC-088-AC-04
  - VOC-088-AC-05
tests:
  - VOC-088-TEST-05
  - VOC-088-TEST-06
  - VOC-088-TEST-07
date: 2026-08-19
related_change: VOC-088
accountable_owner: unassigned
gate_status: repository-complete-live-deploy-pending
live_deployment_claimed: false
---

# VOC-088-T01 — Controlled-signup readiness and synthetic evidence

## Scope and outcome

The public `/healthz` contract now exposes only the boolean
`controlled_signup_ready`. It is true when Google OAuth is enabled, blanket new
signup remains disabled, and at least one non-reserved allowlisted identity can
sign up. It exposes neither email values nor cohort cardinality and does not
change the endpoint's HTTP health semantics.

The existing `synthetic.staging.oauth-expected-state` harness now requires that
boolean to be true whenever staging OAuth is expected enabled, in addition to
its existing HTTP 200, exact Google hostname, and canonical callback checks.
The hourly scheduled workflow uses the enabled expectation and therefore fails
closed if the controlled cohort disappears.

## Repository deliverables

| Artifact                               | Path                                                   |
| -------------------------------------- | ------------------------------------------------------ |
| Readiness calculation and response     | `apps/api/app/api/production.go`                       |
| API unit tests                         | `apps/api/app/api/production_test.go`                  |
| Live/scheduled OAuth readiness harness | `infra/scripts/verify-staging-oauth-start.sh`          |
| Disposable harness self-test           | `infra/scripts/verify-staging-oauth-start.selftest.sh` |
| Synthetic registry coverage            | `infra/monitoring/synthetics.yaml`                     |
| Scheduled workflow wiring              | `.github/workflows/scheduled-synthetics.yml`           |
| Deterministic repository tests         | `scripts/foundation/voc088-staging-readiness.test.mjs` |

## Acceptance mapping

| Acceptance criterion | Repository result                                                                                                    |
| -------------------- | -------------------------------------------------------------------------------------------------------------------- |
| AC-03                | Health and synthetic output contain no address or cohort-size field; fixtures use only the invalid synthetic domain. |
| AC-04                | Enabled staging OAuth verification fails when `controlled_signup_ready` is absent/false and passes only when true.   |
| AC-05                | `/healthz` exposes one accurately derived boolean; HTTP status continues to represent API/database health.           |

## Deterministic validation

```bash
bash infra/scripts/verify-staging-oauth-start.selftest.sh
node --test scripts/foundation/voc088-staging-readiness.test.mjs
(cd apps/api && go test ./app/api)
```

Results on the reviewed working tree:

- seven OAuth/readiness shell scenarios passed;
- three VOC-088 foundation tests passed;
- the API package test suite passed.

No secret, real email address, OAuth state, authorization code, or session value
was used or recorded.

## Live deployment

Pending merge-triggered staging deployment. After merge, the deployment and
scheduled synthetic must prove HTTP 200 OAuth start, the exact Google callback,
and `controlled_signup_ready=true` without completing a real OAuth login.
