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
gate_status: repository-complete-live-deploy-verified-operator-oauth-and-failure-fixture-pending
live_deploy_claimed: true
live_synthetics_claimed: true
live_signin_claimed: false
live_failure_fixture_claimed: false
reviewed_sha: bind-at-independent-review
---

# VOC-088-T03 — Operator procedure and deploy-and-verify evidence

## Scope and outcome

This task adds the secure operator runbook for staging controlled Google signup
and records live deploy/readiness evidence after VOC-088-T00 through T02 merged
to `develop`. Repository automation proves secret-backed cohort deployment,
readiness, and scheduled synthetic health. Human Google login and the bounded
operational-failure live fixture remain operator-executed checklists documented
in `docs/operations/staging-controlled-signup.md` because they require
credentials and must not place addresses, OAuth codes, or session values in
evidence.

## Repository deliverables

| Artifact | Path |
| --- | --- |
| Operator procedure | `docs/operations/staging-controlled-signup.md` |
| Operations index link | `docs/operations/README.md` |
| Deterministic doc/evidence tests | `scripts/foundation/voc088-operator-procedure.test.mjs` |
| This evidence | `specs/changes/VOC-088-persist-controlled-staging-oauth-signup-and-auto/t03-evidence.md` |

## Acceptance mapping

| Acceptance criterion | Result |
| --- | --- |
| AC-08 | Operator procedure documents secret update, deploy pickup, cohort preservation, fail-closed rules, monitoring split, and human sign-in checklist without recording secret values. |
| AC-09 (deploy/readiness) | Live staging shows `controlled_signup_ready: true`; consecutive push deploys preserve the cohort; staging OAuth/readiness synthetic is green. |
| AC-09 (sign-in) | Proxy readiness evidence recorded below; full Google login proof is an operator checklist item (not recorded here). |
| AC-09 (failure fixture) | Deterministic deduplication/sanitization proof from T02 cited; live App-created issue proof procedure documented; no live fixture run recorded in this revision. |
| AC-03 | No email addresses, OAuth codes, session cookies, or secret values appear in docs, tests, or this evidence. |

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

Interpretation: Google OAuth is enabled, blanket signup remains disabled, and
the runtime reports a non-empty controlled cohort without exposing addresses or
cardinality (VOC-088-AC-05 proxy).

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

| Run | Conclusion | Head SHA | Event | URL |
| --- | --- | --- | --- | --- |
| [32251102959](https://github.com/KARSIFT/vocanova-platform-sandbox/actions/runs/32251102959) | success | `d7525f09c929` | push | deploy-staging |
| [32252268728](https://github.com/KARSIFT/vocanova-platform-sandbox/actions/runs/32252268728) | success | `adffeba987a5` | push | deploy-staging |

GitHub Actions jobs API for run `32251102959` shows both allowlist steps
succeeded on the later revision as well:

- **Validate staging controlled-signup allowlist** — success
- **Write staging application configuration** — success

Together with `controlled_signup_ready: true` on live `/healthz`, this satisfies
the repository-visible half of VOC-088-AC-00 / AC-01 / AC-08 cohort-preservation
proof. The secret value itself is not read or recorded.

## Scheduled synthetics

Latest hourly `scheduled-synthetics` run on 2026-08-19:

| Job | Conclusion | Run |
| --- | --- | --- |
| `synthetic.staging.oauth-expected-state` | success | [32251417112](https://github.com/KARSIFT/vocanova-platform-sandbox/actions/runs/32251417112) |

This job executes `verify-staging-oauth-start.sh` with `EXPECT_OAUTH_ENABLED=true`,
which now requires `controlled_signup_ready: true` on `/healthz` (VOC-088-T01).

## Human Google sign-in verification (operator checklist — not recorded here)

Repository automation does not complete Google login. Operators follow
§ "Human sign-in verification" in `docs/operations/staging-controlled-signup.md`:

1. **Allowlisted first-time Google user** — after the secret-backed deploy above,
   sign in at `https://staging.vocanova.site/signin` with an identity present in
   `STAGING_NEW_USER_SIGNUP_ALLOWLIST`. Expect success (no HTTP 503 at callback).
2. **Unlisted Google user** — attempt first-time sign-in with an identity not in
   the secret. Expect HTTP 503 with the stable disabled-signup body at callback
   (`auth.ErrSignupsDisabled`; covered in `apps/api/business/auth/killswitches_test.go`).

Record only pass/fail in the package issue or closure comment; never paste
addresses, OAuth codes, or session cookies into git evidence.

Pre-fix context: GitHub issue #746 reported HTTP 503 for a real first-time
Google login when the ephemeral dispatch allowlist was empty. VOC-088-T00
replaced that path with the persistent secret; live readiness above shows the
post-fix cohort is deployed.

`live_signin_claimed: false` in this file's front matter until an operator
records checklist completion without personal data.

## Operational failure fixture (repository + operator procedure)

Deterministic proof (VOC-088-T02, unchanged):

```bash
node --test scripts/foundation/voc088-failure-to-issue.test.mjs
```

Covers exactly-once issue creation, deduplication, sanitization, App-token
identity, observer coverage, and Sentry non-regression (TEST-08 through TEST-11).

Live App-created issue proof is **not** claimed in this revision. As of
2026-08-19 the public GitHub Actions API lists no completed
`operational-failure-monitoring` runs yet. Operators execute the bounded fixture
in `docs/operations/staging-controlled-signup.md` § "Controlled failure
fixture" after merge, then record the issue number and run URLs without logs or
credentials.

`live_failure_fixture_claimed: false` until that operator proof is attached.

## Privacy and redaction

No repository secret value, real email address, OAuth authorization code, OAuth
`state`, session cookie, or mint token appears in:

- `docs/operations/staging-controlled-signup.md`
- `scripts/foundation/voc088-operator-procedure.test.mjs`
- this evidence file

Test fixtures elsewhere in VOC-088 continue to use only `@synthetic.vocanova.invalid`
addresses.

## Limits

- No `gh` authentication in the implementer sandbox; workflow metadata was
  collected through the public GitHub REST API and direct staging HTTPS probes.
- Job-log download returned `403` without credentials; step conclusions come from
  the jobs API and live harness output instead of raw workflow logs.
- Human Google login and the live operational-failure fixture require operator
  execution per the runbook and are explicitly not claimed here.
