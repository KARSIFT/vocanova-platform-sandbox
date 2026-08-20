---
evidence_id: VOC-096-EV-01
task_id: VOC-096-T01
acceptance_criteria:
  - VOC-096-AC-04
  - VOC-096-AC-05
  - VOC-096-AC-06
  - VOC-096-AC-08
tests:
  - VOC-096-TEST-06
  - VOC-096-TEST-07
  - VOC-096-TEST-08
  - VOC-096-TEST-09
date: 2026-08-20
related_change: VOC-096
accountable_owner: unassigned
gate_status: repository-complete-live-production-pending
live_production_claimed: false
live_synthetic_claimed: false
---

# VOC-096-T01 — Production controlled-signup readiness synthetic

## Scope and outcome

The production OAuth expected-state harness now checks the public `/healthz`
response whenever OAuth is expected enabled. It requires
`controlled_signup_ready=true` and rejects email-like metadata, while retaining
the existing HTTP 200, exact `accounts.google.com` host, and canonical production
callback checks. The scheduled synthetic inventory records the additional
readiness coverage under the existing stable synthetic ID.

No real Google login is performed. The disposable self-test uses only a local
fake HTTP provider/API and does not read production secrets or user data.

## Repository deliverables

| Artifact | Path |
| --- | --- |
| Production OAuth/readiness harness | `infra/scripts/verify-production-oauth-start.sh` |
| Disposable positive/negative fixtures | `infra/scripts/verify-production-oauth-start.selftest.sh` |
| Synthetic inventory coverage | `infra/monitoring/synthetics.yaml` |
| Scheduled workflow contract | `.github/workflows/scheduled-synthetics.yml` |
| Deterministic tests | `scripts/foundation/voc096-production-readiness.test.mjs` |

## Acceptance mapping

| Acceptance criterion | Repository result |
| --- | --- |
| AC-04 | The production OAuth synthetic fails when controlled signup readiness is false and OAuth is expected enabled. |
| AC-05 | The harness requires the boolean readiness field and rejects email-like health metadata. Live production proof remains pending. |
| AC-06 | Existing Google authorization-host and exact canonical callback checks remain mandatory. |
| AC-08 | Foundation tests and the disposable self-test cover readiness pass/fail, inventory wiring, callback retention, and redacted output. |

## Deterministic validation

```bash
bash infra/scripts/verify-production-oauth-start.selftest.sh
node --test scripts/foundation/voc096-production-readiness.test.mjs
```

Full repository validation, governance checks, and independent review must pass
on the exact PR SHA before merge.

## Live production verification

Not claimed by T01. After this task merges, a separately reviewed controlled
activation promotion will deploy the secret-backed policy and readiness check to
production. VOC-096-T02 will record the production deploy, public boolean-only
health response, canonical OAuth start/callback result, scheduled synthetic, and
route/content smoke evidence without identities or OAuth artifacts.
