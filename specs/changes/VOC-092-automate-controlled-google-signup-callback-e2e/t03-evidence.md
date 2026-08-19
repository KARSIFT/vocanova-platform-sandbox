---
evidence_id: VOC-092-EV-03
task_id: VOC-092-T03
acceptance_criteria:
  - VOC-092-AC-03
  - VOC-092-AC-04
  - VOC-092-AC-07
  - VOC-092-AC-08
  - VOC-092-AC-11
tests:
  - VOC-092-TEST-05
  - VOC-092-TEST-06
  - VOC-092-TEST-10
date: 2026-08-19
related_change: VOC-092
accountable_owner: unassigned
gate_status: repository-complete-exact-sha-review-pending
live_ci_claimed: true
live_staging_synthetic_claimed: true
---

# VOC-092-T03 — Live CI and staging synthetic verification

## Scope and outcome

After VOC-092-T00 through T02 merged to `develop`, this task records live
repository CI proof that both controlled-signup OAuth callback E2E cases execute
and pass, confirms the unchanged staging OAuth-readiness synthetic remains green,
and captures full repository validation on the task working tree.

No real email address, OAuth authorization code, OAuth `state`, access token,
refresh token, session cookie value, or authorization URL body appears in this
evidence.

## Repository deliverables

| Artifact | Path |
| --- | --- |
| T00 harness (dependency) | `apps/api/app/api/controlled_signup_oauth_e2e_test.go` |
| T01 CI workflow (dependency) | `.github/workflows/controlled-signup-oauth-e2e.yml` |
| Staging OAuth synthetic (unchanged) | `infra/scripts/verify-staging-oauth-start.sh`, `infra/monitoring/synthetics.yaml` |
| This evidence | `specs/changes/VOC-092-automate-controlled-google-signup-callback-e2e/t03-evidence.md` |

## Acceptance mapping

| Acceptance criterion | Result | Evidence section |
| --- | --- | --- |
| AC-03 allowlisted synthetic reaches onboarding/home | pass | § Live `develop` CI, § Scrubbed exact-job harness outcome |
| AC-04 unlisted synthetic receives HTTP 503 | pass | § Live `develop` CI, § Scrubbed exact-job harness outcome |
| AC-07 staging OAuth synthetic retained | pass | § Staging synthetic retention, § Live staging readiness |
| AC-08 CI executes both cases fail-closed | pass | § Live `develop` CI, § Fail-closed confirmation |
| AC-11 staging/production isolation preserved | pass | § Deterministic validation, § Privacy and redaction |

## Live `develop` CI (post T00–T02 merge)

Push to `develop` after VOC-092-T02 merge (`c78a537bf093a9986d49db695e0ecba1813883ea`):

| Run | Conclusion | Head SHA | Trigger | URL |
| --- | --- | --- | --- | --- |
| [32312272570](https://github.com/KARSIFT/vocanova-platform-sandbox/actions/runs/32312272570) | success | `c78a537bf093` | push (`VOC-092-T02`) | [job 96257511360](https://github.com/KARSIFT/vocanova-platform-sandbox/actions/runs/32312272570/job/96257511360) |

Earlier push after VOC-092-T01 merge for continuity:

| Run | Conclusion | Head SHA | Trigger | URL |
| --- | --- | --- | --- | --- |
| [32299315200](https://github.com/KARSIFT/vocanova-platform-sandbox/actions/runs/32299315200) | success | `fdde12da` | push (`VOC-092-T01`) | workflow run |

The dedicated workflow requires Docker, runs
`go test ./app/api/... -run ControlledSignupOAuth -count=1 -v`, and fails closed
unless both named cases report `PASS` and no `TestControlledSignupOAuth_` case
`SKIP`s.

## Scrubbed exact-job harness outcome (2026-08-19)

The following four outcome lines are from exact post-T02 Actions job
[96257511360](https://github.com/KARSIFT/vocanova-platform-sandbox/actions/runs/32312272570/job/96257511360).
Workflow/step prefixes and timestamps were removed; no application log lines,
OAuth codes, state, tokens, cookies, authorization URLs, or identity values were
copied.

```
=== RUN   TestControlledSignupOAuth_AllowlistedCallbackSucceeds
--- PASS: TestControlledSignupOAuth_AllowlistedCallbackSucceeds (7.69s)
=== RUN   TestControlledSignupOAuth_UnlistedCallbackDenied
--- PASS: TestControlledSignupOAuth_UnlistedCallbackDenied (2.13s)
```

Those named tests assert, respectively, redirect to the configured
onboarding/home path without HTTP 503 and the stable new-signups-disabled HTTP
503 response with no persisted user row. The workflow's post-test guards require
both exact `PASS` lines and reject any controlled-signup `SKIP`, so this excerpt
and the successful job conclusion are bound to the same CI execution.

## Fail-closed confirmation

T01 evidence (`VOC-092-EV-01`) already records deliberate fail-closed rehearsal
for harness failure, missing named `PASS` lines, silent `SKIP`, and missing Docker.
The live `develop` workflow run above succeeded with the same guards; no
`continue-on-error` or skip path is present in
`.github/workflows/controlled-signup-oauth-e2e.yml`.

## Staging synthetic retention (AC-07)

VOC-092 commits did not modify `infra/scripts/verify-staging-oauth-start.sh`,
`synthetic.staging.oauth-expected-state` in `infra/monitoring/synthetics.yaml`,
or `.github/workflows/scheduled-synthetics.yml`. The complementary repository
harness adds CI coverage only (`VOC-092-D06`).

Deterministic regression:

```bash
node --test scripts/foundation/voc088-staging-readiness.test.mjs
```

Result: three VOC-088 staging-readiness tests passed on the task working tree.

## Live staging readiness

Post-T02 `deploy-staging` run
[32312272542](https://github.com/KARSIFT/vocanova-platform-sandbox/actions/runs/32312272542)
completed successfully on exact `develop` merge SHA
`c78a537bf093a9986d49db695e0ecba1813883ea`. The repository workflow passed
controlled-signup allowlist validation, public API/web health, the canonical
OAuth-start readiness gate, hardened Chromium installation, and the staging
browser core-loop journey.

Public `/healthz` (2026-08-19):

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

The OAuth start response body is not copied here (it contains OAuth `state`).

## Scheduled synthetics

Post-T02 successful `scheduled-synthetics` run from exact `develop` merge SHA
`c78a537bf093a9986d49db695e0ecba1813883ea` on 2026-08-19:

| Job | Conclusion | Run |
| --- | --- | --- |
| `synthetic.staging.oauth-expected-state` | success | [32312588151](https://github.com/KARSIFT/vocanova-platform-sandbox/actions/runs/32312588151) ([job 96258426406](https://github.com/KARSIFT/vocanova-platform-sandbox/actions/runs/32312588151/job/96258426406)) |

The job ran after the successful post-T02 deployment and executed the retained
merged VOC-088-T01 version of
`verify-staging-oauth-start.sh` with `EXPECT_OAUTH_ENABLED=true`, requiring
`controlled_signup_ready: true` on `/healthz`.

## Deterministic validation

Commands (repository root):

```bash
node --test scripts/foundation/voc092-controlled-signup-oauth-e2e.test.mjs
node --test scripts/foundation/voc092-operator-docs.test.mjs
node --test scripts/foundation/voc088-staging-readiness.test.mjs
bash scripts/governance/validate-governance.sh
git diff --check
pnpm validate
cd apps/api && go test ./app/api/... -run ControlledSignupOAuth -count=1 -v
```

Results on the task working tree (2026-08-19):

- twelve VOC-092 / VOC-088 foundation tests passed (0 failed);
- governance structure validation passed;
- `git diff --check` passed;
- `pnpm validate` passed (Node engine warning only: wanted `24.18.0`, ran `24.5.0`);
- both controlled-signup OAuth callback E2E cases passed locally with disposable
  Postgres.

## Privacy and redaction

- Harness fixtures use only `@synthetic.vocanova.invalid` identities distinct from
  the reserved deploy smoke-test bot address.
- CI and local logs assert outcomes without printing authorization URLs, callback
  query strings, OAuth codes, tokens, or `Set-Cookie` values.
- Live staging checks record only public `/healthz` kill-switch posture and
  scrubbed PASS/FAIL lines from `verify-staging-oauth-start.sh`.
- No staging or production database URL, secret value, or session mint token was
  used by the repository-managed callback E2E harness.

## Remaining gates

- exact-SHA CI after this evidence commit;
- independent exact-SHA review binding to the merged task revision.
