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
reviewed_sha: bind-at-independent-review
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

| Acceptance criterion      | Result  | Evidence section                              |
| ------------------------- | ------- | --------------------------------------------- |
| AC-08 operator procedure  | pass    | § Operator procedure, § Cohort preservation   |
| AC-09 deploy/readiness    | pass    | § Live staging readiness, § Cohort preservation, § Scheduled synthetics |
| AC-09 sign-in (listed)    | pass    | § Human Google sign-in verification           |
| AC-09 sign-in (unlisted)  | pass    | § Human Google sign-in verification           |
| AC-09 failure fixture     | pass    | § Operational failure fixture live proof       |
| AC-03 no emails           | pass    | § Privacy and redaction                       |

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

## Human Google sign-in verification

Live sign-in verification was performed by an operator on 2026-08-19 against
`https://staging.vocanova.site/signin` after the secret-backed deploy:

| Check                              | Result | Notes                                                                         |
| ---------------------------------- | ------ | ----------------------------------------------------------------------------- |
| Allowlisted first-time Google user | pass   | Reached staging onboarding without HTTP 503                                   |
| Unlisted Google user               | pass   | API callback returned HTTP 503 with `auth.ErrSignupsDisabled` stable body     |

No email addresses, OAuth codes, session cookies, or callback query strings are
recorded in this evidence. The application-level allowlist/deny behavior is
additionally covered by unit tests in
`apps/api/business/auth/killswitches_test.go`.

## Operational failure fixture live proof

The `operational-failure-monitoring.yml` `workflow_run` observer requires
presence on the default branch (`main`) to fire. After VOC-088 promotion to
`main` (issue [#757](https://github.com/KARSIFT/vocanova-platform-sandbox/issues/757)),
the bounded cancellation fixture from
`docs/operations/staging-controlled-signup.md` § "Controlled failure fixture"
was executed:

| Step | Observation                                                                                       |
| ---- | ------------------------------------------------------------------------------------------------- |
| 1    | `scheduled-synthetics.yml` dispatched with `synthetic_id=synthetic.staging.authenticated-core-journey` |
| 2    | Run cancelled via `gh run cancel`; conclusion `cancelled`                                         |
| 3    | `operational-failure-monitoring` observed the `workflow_run` event and created exactly one issue titled `Operational failure: scheduled-synthetics (cancelled)` with stable fingerprint marker `<!-- operational-failure:scheduled-synthetics:cancelled -->` |
| 4    | App-created issue triggered `pipeline` run with `plan-from-issue` job queued; pipeline run cancelled before adoption to contain the fixture |
| 5    | Repeat cancellation with first issue still open: observer ran successfully, no duplicate issue created (dedup confirmed) |
| 6    | Proof issue closed with sanitized comment citing the two cancelled run IDs and two observer run IDs; no logs, credentials, or account data attached |

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
