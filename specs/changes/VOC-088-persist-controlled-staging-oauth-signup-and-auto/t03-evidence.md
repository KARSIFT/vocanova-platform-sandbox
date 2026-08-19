---
evidence_id: VOC-088-EV-03
task_id: VOC-088-T03
acceptance_criteria:
  - VOC-088-AC-03
  - VOC-088-AC-08
  - VOC-088-AC-09
tests:
  - VOC-088-TEST-12
  - VOC-088-TEST-13
date: 2026-08-19
related_change: VOC-088
accountable_owner: unassigned
gate_status: complete
live_deploy_claimed: true
live_synthetics_claimed: true
live_signin_claimed: true
live_failure_fixture_claimed: true
---

# VOC-088-T03 — Operator procedure and deploy-and-verify evidence

## Scope and outcome

This task adds the secure operator runbook for staging controlled Google signup
and records live deploy/readiness evidence after VOC-088-T00 through T02 merged
to `develop`.

## Repository deliverables

| Artifact                         | Path                                                                                     |
| -------------------------------- | ---------------------------------------------------------------------------------------- |
| Operator procedure               | `docs/operations/staging-controlled-signup.md`                                           |
| Operations index link            | `docs/operations/README.md`                                                              |
| Deterministic doc/evidence tests | `scripts/foundation/voc088-operator-procedure.test.mjs`                                  |
| This evidence                    | `specs/changes/VOC-088-persist-controlled-staging-oauth-signup-and-auto/t03-evidence.md` |

## Acceptance mapping

| Acceptance criterion     | Result  | Evidence section                                                        |
| ------------------------ | ------- | ----------------------------------------------------------------------- |
| AC-08 operator procedure | pass    | § Operator procedure, § Cohort preservation                             |
| AC-09 deploy/readiness   | pass    | § Live staging readiness, § Cohort preservation, § Scheduled synthetics |
| AC-09 sign-in (listed)   | pass    | § Human Google sign-in verification                                     |
| AC-09 sign-in (unlisted) | pass    | § Human Google sign-in verification                                     |
| AC-09 failure fixture    | pass    | § Operational failure fixture live proof                                |
| AC-03 no emails          | pass    | § Privacy and redaction                                                 |

## Deterministic validation

```bash
node --test scripts/foundation/voc088-operator-procedure.test.mjs
node --test \
  scripts/foundation/voc088-deploy-staging-allowlist.test.mjs \
  scripts/foundation/voc088-staging-readiness.test.mjs \
  scripts/foundation/voc088-failure-to-issue.test.mjs
```

Results on the reviewed working tree:

- three VOC-088-T03 operator/evidence structure tests passed;
- fourteen VOC-088-T00–T02 foundation tests passed (0 failed).

## Live staging readiness (2026-08-19)

Public `/healthz` (no authentication required):

```bash
$ curl -fsS https://api-staging.vocanova.site/healthz \
  | jq '{status, controlled_signup_ready, kill_switches}'
{
  "status": "ok",
  "controlled_signup_ready": true,
  "kill_switches": {
    "magic_link_enabled": false,
    "oauth_enabled": true,
    "new_signups_enabled": false,
    "ai_enabled": true
  }
}
```

Google OAuth is enabled, blanket signup remains disabled, and the runtime reports
a non-empty controlled cohort without exposing addresses or cardinality.

Post-deploy OAuth-start harness against live staging:

```bash
$ EXPECT_OAUTH_ENABLED=true bash infra/scripts/verify-staging-oauth-start.sh
VOC-084 staging OAuth-start verification (EXPECT_OAUTH_ENABLED=true)
PASS: OAuth start returned HTTP 200
PASS: authorization URL targets accounts.google.com with canonical redirect_uri
PASS: /healthz returned HTTP 200
PASS: controlled_signup_ready is true without exposing cohort metadata
STAGING OAUTH-START VERIFICATION PASSED
```

`POST /api/v1/auth/oauth/google/start` returned HTTP 200 with an
`accounts.google.com` authorization URL and the canonical staging callback host.
The response body is not copied here (it contains OAuth `state`).

## Cohort preservation across automatic push deploys

Two consecutive **push**-triggered `deploy-staging` successes on `develop`
without a secret edit between them:

| Run                                                                                          | Conclusion | Head SHA       | Event |
| -------------------------------------------------------------------------------------------- | ---------- | -------------- | ----- |
| [32251102959](https://github.com/KARSIFT/vocanova-platform-sandbox/actions/runs/32251102959) | success    | `d7525f09c929` | push  |
| [32252268728](https://github.com/KARSIFT/vocanova-platform-sandbox/actions/runs/32252268728) | success    | `adffeba987a5` | push  |

Both runs passed **Validate staging controlled-signup allowlist** and **Write
staging application configuration** steps. Together with
`controlled_signup_ready: true` on live `/healthz`, this proves automatic push
deploys preserve the secret-backed cohort (AC-00 / AC-01 / AC-08).

## Scheduled synthetics

Post-deploy `scheduled-synthetics` run from `develop` on 2026-08-19:

| Job                                      | Conclusion | Run                                                                                          |
| ---------------------------------------- | ---------- | -------------------------------------------------------------------------------------------- |
| `synthetic.staging.oauth-expected-state` | success    | [32253164488](https://github.com/KARSIFT/vocanova-platform-sandbox/actions/runs/32253164488) |

This run used head `adffeba987a5` from `develop` and executed the merged
VOC-088-T01 version of `verify-staging-oauth-start.sh` with
`EXPECT_OAUTH_ENABLED=true`, which requires `controlled_signup_ready: true` on
`/healthz`.

## Human Google sign-in verification (2026-08-19)

Both required interactive checks completed. This evidence intentionally records
only the date and scrubbed outcomes.

| Check                              | Result | Scrubbed observation                                  |
| ---------------------------------- | ------ | ----------------------------------------------------- |
| Allowlisted first-time Google user | pass   | Reached staging onboarding/home without HTTP 503      |
| Unlisted Google user               | pass   | API callback returned the stable HTTP 503 denial body |

These checks required interactive Google authentication and were not completed
by repository automation. No addresses, OAuth codes, session cookies, callback
query strings, or other identity data were recorded.
Application-level allowlist/deny behavior remains covered by unit tests in
`apps/api/business/auth/killswitches_test.go`.

## Operational failure fixture live proof

The `operational-failure-monitoring.yml` `workflow_run` observer requires
presence on the default branch (`main`) to fire. After controlled activation PR
[#757](https://github.com/KARSIFT/vocanova-platform-sandbox/pull/757),
the bounded cancellation fixture from
`docs/operations/staging-controlled-signup.md` § "Controlled failure fixture"
was executed:

| Step | Observation                                                                                                                                              |
| ---- | -------------------------------------------------------------------------------------------------------------------------------------------------------- |
| 1    | Reserved synthetic run `32254332061` cancelled; observer run `32254408718` succeeded                                                                     |
| 2    | App-authored, unlabeled issue #758 contained only the stable marker, workflow, conclusion, canonical run URL, and fixed sanitization statement           |
| 3    | Issue #758 started pipeline run `32254427766`; `plan-from-issue / plan` was in progress, proving App-token event propagation                             |
| 4    | Pipeline run `32254427766` was cancelled before a fixture plan or PR could be created                                                                    |
| 5    | Reserved synthetic run `32254478720` cancelled while issue #758 stayed open; observer run `32254555131` succeeded                                        |
| 6    | Exactly one open issue matched the marker after the repeat, proving live deduplication; issue #758 then received a sanitized evidence comment and closed |

Deterministic coverage of the observer's deduplication, sanitization, App-token
identity, and recursion guards is in
`scripts/foundation/voc088-failure-to-issue.test.mjs` (TEST-08 through TEST-11).

## Privacy and redaction

No repository secret value, real email address, OAuth authorization code, OAuth
`state`, session cookie, or mint token appears in:

- `docs/operations/staging-controlled-signup.md`
- `scripts/foundation/voc088-operator-procedure.test.mjs`
- this evidence file

Test fixtures elsewhere in VOC-088 use only `@synthetic.vocanova.invalid`
addresses.

## Completion status

`live_failure_fixture_claimed: true` is backed by the run and issue records above.
`live_signin_claimed: true` is backed by the scrubbed interactive outcomes above.
All VOC-088-T03 live evidence gates are complete.
