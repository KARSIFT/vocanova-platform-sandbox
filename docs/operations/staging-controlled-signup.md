# Staging controlled Google signup (VOC-088)

Operator guide for the persistent staging controlled-signup cohort backed by
repository secret `STAGING_NEW_USER_SIGNUP_ALLOWLIST`. Only
`.github/workflows/deploy-staging.yml` may read this secret;
`deploy-production.yml` never references it.

## Policy

| Setting                     | Staging value                                                                                             | Meaning                                                    |
| --------------------------- | --------------------------------------------------------------------------------------------------------- | ---------------------------------------------------------- |
| `GOOGLE_OAUTH_ENABLED`      | `true` when both `GOOGLE_OAUTH_CLIENT_ID` and `GOOGLE_OAUTH_CLIENT_SECRET` repository secrets are present | Google sign-in is available                                |
| `NEW_USER_SIGNUP_ENABLED`   | always `false`                                                                                            | No blanket first-time signup                               |
| `NEW_USER_SIGNUP_ALLOWLIST` | synced from `STAGING_NEW_USER_SIGNUP_ALLOWLIST` on every deploy                                           | Only listed Google identities may create a staging account |

The former `workflow_dispatch` input `new_user_signup_allowlist` was removed.
Automatic push deploys and manual dispatches both read the repository secret
only. A normal merge to `develop` cannot silently erase the cohort.

Personal data rules:

- Store cohort member addresses only in the GitHub secret.
- Never commit addresses, never paste them into issues/PRs/evidence, and never
  expect deploy logs to print them. Successful validation logs only
  `controlled signup ready: true` or `controlled signup ready: false`.
- `/healthz` exposes only `controlled_signup_ready` (boolean). It never exposes
  email values or cohort cardinality.

## Add or remove a cohort member

1. Open **GitHub → Settings → Secrets and variables → Actions → Repository
   secrets**.
2. Edit `STAGING_NEW_USER_SIGNUP_ALLOWLIST`.
3. Replace the value with the complete comma-separated cohort (no spaces
   required; the API normalizes case and surrounding whitespace). GitHub does
   not reveal the prior value, and saving replaces the whole secret, so obtain
   the current cohort through the approved secure credential-management path
   before adding or removing a member.
4. Save the secret. Do not record the value in chat, tickets, terminal output,
   or evidence.

Removal is the same path: delete the address from the comma-separated list (or
clear the entire secret only when staging Google OAuth is intentionally
disabled — see fail-closed behavior below).

## Pick up the change

Every merge to `develop` triggers `deploy-staging.yml`, which:

1. Validates the secret on the runner (`infra/scripts/validate-staging-signup-allowlist.sh`).
2. Writes `NEW_USER_SIGNUP_ALLOWLIST` into staging `api.env`.
3. Redeploys the API container.

To apply a secret-only change without a code merge, manually dispatch
`deploy-staging.yml` on `develop` (Workflows → Deploy staging → Run workflow).

### Verify the deploy

In the successful run log, confirm:

- Step **Validate staging controlled-signup allowlist** succeeded.
- Step **Write staging application configuration** logged
  `controlled signup ready: true` (when OAuth is enabled).
- Step **Verify staging OAuth start (post-deploy)** passed.
- Public polls for `/healthz` and the web root passed.

Then confirm live readiness without exposing cohort metadata:

```bash
curl -fsS https://api-staging.vocanova.site/healthz \
  | jq '{status, controlled_signup_ready, kill_switches}'
```

Expect `controlled_signup_ready: true` when OAuth is enabled and at least one
non-reserved allowlisted identity exists.

Optional harness (does not complete Google login):

```bash
EXPECT_OAUTH_ENABLED=true bash infra/scripts/verify-staging-oauth-start.sh
```

## Repository-managed OAuth callback E2E harness (VOC-092)

CI and local operators can exercise the full application OAuth **callback**
contract without interactive Google login. The harness drives the real
`POST /api/v1/auth/oauth/google/start` and
`GET /api/v1/auth/oauth/google/callback` handlers, validates the OAuth state
cookie, runs the real `GoogleOAuthProvider` token and userinfo HTTP boundary
against a localhost fake Google server inside the test process, and persists
auth rows in disposable loopback PostgreSQL only. It mirrors staging controlled
signup policy: `OAuthEnabled=true`, `NewSignupsEnabled=false`, and a non-empty
controlled allowlist for the success case.

### What the harness proves

| Area | Coverage |
| --- | --- |
| OAuth start/callback HTTP contract | State cookie issuance and validation, callback query handling, redirect semantics |
| Provider HTTP boundary | Real token exchange and userinfo fetch through `GoogleOAuthProvider` (not a runtime provider switch) |
| Database persistence | User, external identity, and session rows on allowlisted success |
| Controlled signup policy | Allowlisted first-time synthetic identity reaches the configured onboarding/home return URL without HTTP 503; unlisted synthetic identity receives the stable HTTP 503 new-signups-disabled body |
| Isolation | No staging/production hosts, secrets, databases, or public test-auth routes |

### What the harness does not prove

- Google account login UI, consent screens, CAPTCHA, or account chooser
- Google-side redirect URI registration drift or OAuth client configuration outside this repository
- Real-user identity edge cases that require a human Google account

The live staging synthetic `synthetic.staging.oauth-expected-state` and
`verify-staging-oauth-start.sh` remain unchanged. They still assert OAuth start
and readiness on staging; the callback harness adds complementary repository CI
coverage only.

### Run locally

Requires Docker on PATH for disposable Postgres.

```bash
bash infra/scripts/run-controlled-signup-oauth-e2e.sh
```

Equivalent direct command:

```bash
pnpm test:controlled-signup-oauth-e2e
```

CI executes the same cases in
`.github/workflows/controlled-signup-oauth-e2e.yml` on pull requests and on
`develop`.

## Prove cohort preservation across automatic deploys

After changing the secret, record two consecutive **push**-triggered
`deploy-staging` successes on `develop` without editing the secret again. Each
run must pass **Validate staging controlled-signup allowlist** and log
`controlled signup ready: true`. That pair proves automatic deploys preserve
the secret-backed cohort (VOC-088-AC-00 / AC-01).

## Fail-closed behavior

When Google OAuth credentials are present and the allowlist secret is missing,
empty, multiline, or malformed, the deploy fails before writing `api.env`. This
is intentional: OAuth-enabled staging with an empty controlled cohort blocks
every first-time Google signup.

If staging must intentionally have no controlled cohort, keep the current secret
until a separately reviewed repository change has disabled the OAuth capability.
Do not bypass the validator, delete credentials as an ad hoc switch, or edit the
host environment manually.

## Human sign-in verification (no credentials in evidence)

Repository automation proves OAuth **start**, readiness, and the application
callback contract (see **Repository-managed OAuth callback E2E harness** above).
Completing Google login against the real provider still requires a human operator.

**Periodic real-provider audit:** run the checklist below after any auth callback
handler, allowlist policy, or Google OAuth client/redirect change, and at least
quarterly even when no such change occurred. A green callback harness or staging
OAuth-start synthetic does not replace this audit.

Use the checklist after the secret change has deployed (or during the quarterly
audit):

| Check                              | How to verify                                                                            | Expected result                                                                                                                             |
| ---------------------------------- | ---------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------- |
| Allowlisted first-time Google user | Sign in at `https://staging.vocanova.site/signin` with an identity present in the secret | Reaches staging home/onboarding without HTTP 503                                                                                            |
| Unlisted Google user               | Attempt first-time sign-in with an identity **not** in the secret                        | Google may authenticate, but the API callback returns HTTP 503 with the stable "new sign-ups are disabled" body (`auth.ErrSignupsDisabled`) |
| Scheduled readiness                | Confirm latest `scheduled-synthetics` run                                                | Job `synthetic.staging.oauth-expected-state` succeeded                                                                                      |
| Cohort persistence                 | Two consecutive push deploy successes after secret edit                                  | Both validate allowlist and log `controlled signup ready: true`                                                                             |

Record only pass/fail and run URLs in evidence. Never paste email addresses,
OAuth codes, session cookies, or callback query strings.

## Operational failure → governed issue

Terminal `failure`, `cancelled`, or `timed_out` conclusions on
`scheduled-synthetics.yml`, `deploy-staging.yml`, and `deploy-production.yml`
are observed by `.github/workflows/operational-failure-monitoring.yml`. That
observer uses the job's default `GITHUB_TOKEN` (**issues write only**) and
passes it to `infra/scripts/open-failure-issue.sh` for create/dedupe. Deploy
benign-cancel classification (`classify-deploy-concurrency-cancel.sh`) uses
the same `GITHUB_TOKEN` with workflow `actions: read` for bounded jobs
metadata. Issues are plain (no labels); there is no downstream
`plan-from-issue` automation, so a human or agent picks each one up.

Responsibility split:

| Layer            | Workflow                             | Scope                                        |
| ---------------- | ------------------------------------ | -------------------------------------------- |
| Sentry           | `error-monitoring.yml`               | Unexpected application errors                |
| Kuma             | `sync-monitoring.yml`                | Availability/TLS/basic health                |
| Synthetics       | `scheduled-synthetics.yml`           | OAuth expected state, authenticated journeys |
| Failure-to-issue | `operational-failure-monitoring.yml` | Failed operational workflows only            |

Expected signup denials (HTTP 503 for unlisted identities) are correct
behavior and must not be sent to Sentry.

### Controlled failure fixture (one-time live proof)

Use only after T00–T02 are on the default branch and App credentials exist.
Goal: one sanitized App-created issue, `plan-from-issue` eligibility, and no
duplicate on repeat.

1. Never break an application, deployment, or observer file for this proof.
   Dispatch only the reserved staging browser synthetic from the default branch:

   ```bash
   gh workflow run scheduled-synthetics.yml \
     --repo KARSIFT/vocanova-platform-sandbox \
     --ref main \
     -f synthetic_id=synthetic.staging.authenticated-core-journey
   ```

2. Resolve the new run ID, wait until it is queued or in progress, and cancel it.
   Cancellation affects only the reserved synthetic account and creates no
   application or production fault:

   ```bash
   gh run list \
     --repo KARSIFT/vocanova-platform-sandbox \
     --workflow scheduled-synthetics.yml \
     --event workflow_dispatch \
     --limit 1
   gh run cancel RUN_ID --repo KARSIFT/vocanova-platform-sandbox
   ```

3. Wait for conclusion `cancelled`. Confirm
   `operational-failure-monitoring` ran for that `workflow_run` and
   created **one** open issue titled
   `Operational failure: scheduled-synthetics (cancelled)` whose body contains
   only the workflow name, conclusion, run URL, fixed summary, and the marker
   `<!-- operational-failure:scheduled-synthetics:cancelled -->`.
4. Confirm the App-created issue started a `pipeline` run and that its
   `plan-from-issue` job was queued or started. Cancel that pipeline run before it
   can draft or adopt a fixture package; `pipeline` is not observed, so this
   containment cannot recurse.
5. Repeat steps 1–3 while the first issue remains open. Confirm the second
   observer run exits successfully and the open-issue count for the marker stays
   exactly one.
6. Add a fixed, sanitized comment identifying the two cancelled run IDs and two
   observer run IDs, then close the proof issue. Do not attach logs or account
   data. A controlled cancellation needs no remediation package.

Repository deterministic coverage lives in
`scripts/foundation/voc088-failure-to-issue.test.mjs` (VOC-088-TEST-08 through
TEST-11).

## Related documentation

- `docs/operations/monitoring.md` — scheduled synthetics and Kuma operations
- `docs/operations/11-devops-and-ci-cd.md` — staging deploy automation
- `specs/changes/VOC-088-persist-controlled-staging-oauth-signup-and-auto/` —
  package specification, tasks, and evidence
- `specs/changes/VOC-092-automate-controlled-google-signup-callback-e2e/` —
  callback E2E harness specification and evidence
