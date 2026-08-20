---
evidence_id: VOC-096-EV-02
task_id: VOC-096-T02
acceptance_criteria:
  - VOC-096-AC-09
  - VOC-096-AC-10
  - VOC-096-AC-03
  - VOC-096-AC-05
tests:
  - VOC-096-TEST-12
  - VOC-096-TEST-13
date: 2026-08-20
related_change: VOC-096
accountable_owner: unassigned
gate_status: complete
live_production_claimed: true
live_synthetic_claimed: true
live_signin_claimed: false
---

# VOC-096-T02 — Operator procedure and production deploy-and-verify evidence

## Scope and outcome

This task adds the secure operator runbook for production controlled Google signup
and records the completed, scrubbed production deploy-and-verify result. VOC-096-T00
and T01 were promoted through the separately checked controlled activation PR #817;
the automatic production deploy and the dedicated production OAuth synthetic passed.

## Repository deliverables

| Artifact | Path |
| --- | --- |
| Operator procedure | `docs/operations/production-controlled-signup.md` |
| Operations index link | `docs/operations/README.md` |
| Deterministic doc/evidence tests | `scripts/foundation/voc096-operator-docs.test.mjs` |
| This evidence | `specs/changes/VOC-096-persist-controlled-production-google-signup-cohort/t02-evidence.md` |

## Acceptance mapping

| Acceptance criterion | Repository result | Evidence section |
| --- | --- | --- |
| AC-09 operator procedure | pass | § Operator procedure |
| AC-10 deploy-and-verify checklist | pass | § Pre-activation baseline, § Post-activation closure |
| AC-03 no emails in docs/evidence | pass | § Privacy and redaction |
| AC-05 boolean-only `/healthz` | pass | § Post-activation closure |

## Operator procedure

`docs/operations/production-controlled-signup.md` documents:

- `PRODUCTION_NEW_USER_SIGNUP_ALLOWLIST` as the sole production cohort source
- removal of the ephemeral `new_user_signup_allowlist` dispatch input
- secret edit workflow and push/manual `deploy-production.yml` pickup
- fail-closed validation when OAuth is enabled and the cohort is invalid
- boolean-only logging and `/healthz` readiness checks
- cohort preservation proof via two consecutive push deploys
- human sign-in verification checklist and quarterly real-provider audit cadence
- cross-links to the staging guide and VOC-092 callback harness documentation

## Deterministic validation

```bash
node --test scripts/foundation/voc096-operator-docs.test.mjs
node --test \
  scripts/foundation/voc096-deploy-production-allowlist.test.mjs \
  scripts/foundation/voc096-production-readiness.test.mjs
```

Results on the task working tree (2026-08-20):

- four VOC-096-T02 operator/documentation tests passed;
- VOC-096-T00/T01 foundation tests passed (regression).

## Pre-activation production baseline (2026-08-20)

Before VOC-096 reaches `main`, production still runs the pre-T00 workflow that
resolved an empty dispatch allowlist on push deploys. The public readiness probe
therefore reports `controlled_signup_ready: false` even though OAuth is enabled.

```bash
$ curl -fsS https://api-production.vocanova.site/healthz \
  | jq '{status, controlled_signup_ready, kill_switches}'
{
  "status": "ok",
  "controlled_signup_ready": false,
  "kill_switches": {
    "magic_link_enabled": false,
    "oauth_enabled": true,
    "new_signups_enabled": false,
    "ai_enabled": true
  }
}
```

Pre-activation OAuth-start/readiness harness (expected to fail until promotion):

```bash
$ EXPECT_OAUTH_ENABLED=true bash infra/scripts/verify-production-oauth-start.sh
VOC-086 production OAuth-start verification (EXPECT_OAUTH_ENABLED=true)
PASS: OAuth start returned HTTP 200
PASS: authorization URL targets accounts.google.com with canonical redirect_uri
PASS: /healthz returned HTTP 200
FAIL: /healthz did not report controlled_signup_ready=true without email metadata
PRODUCTION OAUTH-START VERIFICATION FAILED: 1 check(s) failed
```

This baseline is intentional: it documents the gap VOC-096 remediates and
confirms the T01 harness correctly fails closed when readiness is false.

Latest successful pre-activation `deploy-production` push run on `main`:

| Run | Conclusion | Head SHA | Event |
| --- | --- | --- | --- |
| [32313920181](https://github.com/KARSIFT/vocanova-platform-sandbox/actions/runs/32313920181) | success | `0515c2923a40` | push |

`develop` at implementation time includes merged VOC-096-T00/T01 (`4ac21fc`).

## Post-activation closure (2026-08-20)

Controlled activation PR #817 merged to `main` at
`0963e8db0502890581350b7134e19f72c3be3f0a`. The following evidence records only
run identifiers, boolean state, and scrubbed outcomes:

| Verification | Result | Evidence |
| --- | --- | --- |
| Production push deployment | pass | [`deploy-production` run 32316192584](https://github.com/KARSIFT/vocanova-platform-sandbox/actions/runs/32316192584) completed successfully on the activation merge SHA |
| Secret-backed validation and configuration | pass | The deploy's **Validate production controlled-signup allowlist** and **Write production application configuration** steps passed; no cohort data was inspected or copied |
| Public production health | pass | `status: ok`; `controlled_signup_ready: true`; OAuth enabled; blanket new signup disabled; only boolean policy fields were retained |
| Production OAuth start | pass | `EXPECT_OAUTH_ENABLED=true bash infra/scripts/verify-production-oauth-start.sh` returned HTTP 200, targeted `accounts.google.com`, retained the exact canonical production callback, and accepted boolean readiness without cohort metadata |
| Production OAuth scheduled synthetic | pass | [`scheduled-synthetics` run 32316391341](https://github.com/KARSIFT/vocanova-platform-sandbox/actions/runs/32316391341), job `synthetic.production.oauth-expected-state`, passed on the same `main` SHA |
| Production route/content smoke | pass | **Run production smoke-test suite** (`smoke-test-production.sh`) passed in deploy run 32316192584, including the non-mutating route/content checks |

No real-provider production sign-in is claimed. Interactive allowlisted/unlisted
Google checks remain an operator audit and were not required for this automated
deployment/readiness closure.

## Cohort preservation across automatic push deploys

The activation supplied the first successful **push**-triggered production deploy:
run 32316192584. The later normal package promotion is the second push-deploy
checkpoint; root-package closure remains pending until that run independently
passes validation and live readiness without a secret edit between the two runs.

## Privacy and redaction

No repository secret value, real email address, OAuth authorization code, OAuth
`state`, session cookie, or mint token appears in the operations documentation,
foundation tests, or this evidence file. Pre-activation harness output records only
pass/fail lines; OAuth start response bodies are not copied.
