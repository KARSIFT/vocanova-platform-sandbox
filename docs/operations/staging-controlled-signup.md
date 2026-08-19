# Staging controlled Google signup (VOC-088)

Operator guide for the persistent staging controlled-signup cohort backed by
repository secret `STAGING_NEW_USER_SIGNUP_ALLOWLIST`. Only
`.github/workflows/deploy-staging.yml` may read this secret;
`deploy-production.yml` never references it.

## Policy

| Setting | Staging value | Meaning |
| --- | --- | --- |
| `GOOGLE_OAUTH_ENABLED` | `true` when both `GOOGLE_OAUTH_CLIENT_ID` and `GOOGLE_OAUTH_CLIENT_SECRET` repository secrets are present | Google sign-in is available |
| `NEW_USER_SIGNUP_ENABLED` | always `false` | No blanket first-time signup |
| `NEW_USER_SIGNUP_ALLOWLIST` | synced from `STAGING_NEW_USER_SIGNUP_ALLOWLIST` on every deploy | Only listed Google identities may create a staging account |

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
3. Use a comma-separated list of Google account emails (no spaces required;
   the API normalizes case and surrounding whitespace). Example shape only —
   use real operator identities, never commit them:

   `operator-a@example.com,operator-b@example.com`

4. Save the secret. Do not record the value in chat, tickets, or evidence.

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

If you must temporarily disable controlled signup while keeping OAuth
credentials, remove or empty the Google OAuth repository secrets first (or
disable OAuth through the supported credential path documented in VOC-084), then
adjust the allowlist. Do not deploy OAuth-enabled staging with an empty cohort.

## Human sign-in verification (no credentials in evidence)

Repository automation proves OAuth **start** and readiness; completing Google
login still requires a human operator. Use this checklist after the secret change
has deployed:

| Check | How to verify | Expected result |
| --- | --- | --- |
| Allowlisted first-time Google user | Sign in at `https://staging.vocanova.site/signin` with an identity present in the secret | Reaches staging home/onboarding without HTTP 503 |
| Unlisted Google user | Attempt first-time sign-in with an identity **not** in the secret | Google may authenticate, but the API callback returns HTTP 503 with the stable "new sign-ups are disabled" body (`auth.ErrSignupsDisabled`) |
| Scheduled readiness | Confirm latest `scheduled-synthetics` run | Job `synthetic.staging.oauth-expected-state` succeeded |
| Cohort persistence | Two consecutive push deploy successes after secret edit | Both validate allowlist and log `controlled signup ready: true` |

Record only pass/fail and run URLs in evidence. Never paste email addresses,
OAuth codes, session cookies, or callback query strings.

## Operational failure → governed issue

Terminal `failure`, `cancelled`, or `timed_out` conclusions on
`scheduled-synthetics.yml`, `deploy-staging.yml`, and `deploy-production.yml`
are observed by `.github/workflows/operational-failure-monitoring.yml`. That
observer mints a GitHub App installation token (`KARSIFT_BOT_APP_ID` /
`KARSIFT_BOT_PRIVATE_KEY`) and calls `infra/scripts/open-failure-issue.sh`.
Issues are plain (no labels) so `plan-from-issue` can draft a change package.

Responsibility split:

| Layer | Workflow | Scope |
| --- | --- | --- |
| Sentry | `error-monitoring.yml` | Unexpected application errors |
| Kuma | `sync-monitoring.yml` | Availability/TLS/basic health |
| Synthetics | `scheduled-synthetics.yml` | OAuth expected state, authenticated journeys |
| Failure-to-issue | `operational-failure-monitoring.yml` | Failed operational workflows only |

Expected signup denials (HTTP 503 for unlisted identities) are correct
behavior and must not be sent to Sentry.

### Controlled failure fixture (one-time live proof)

Use only after T00–T02 are on the default branch and App credentials exist.
Goal: one sanitized App-created issue, `plan-from-issue` eligibility, and no
duplicate on repeat.

1. Choose a **non-production** failure that cannot recurse into the observer
   itself (never break `operational-failure-monitoring.yml` on purpose).
2. Safe pattern: on a disposable branch, dispatch `scheduled-synthetics.yml`
   with `synthetic_id=synthetic.staging.oauth-expected-state` after temporarily
   pointing `EXPECT_OAUTH_ENABLED` at `false` in that branch only, or after
   introducing a deliberate, reviewed test-only harness failure. Merge the
   disposable branch only through the normal governed loop; do not leave the
   fixture on `develop`.
3. Wait for the synthetic run to finish `failure`.
4. Confirm `operational-failure-monitoring` ran for that `workflow_run` and
   created **one** open issue titled
   `Operational failure: scheduled-synthetics (failure)` whose body contains
   only the workflow name, conclusion, run URL, fixed summary, and the marker
   `<!-- operational-failure:scheduled-synthetics:failure -->`.
5. Re-dispatch the same failure class while the issue remains open. Confirm the
   observer exits without a second issue (deduplication).
6. Close the proof issue through the normal planning/remediation path.

Repository deterministic coverage lives in
`scripts/foundation/voc088-failure-to-issue.test.mjs` (VOC-088-TEST-08 through
TEST-11).

## Related documentation

- `docs/operations/monitoring.md` — scheduled synthetics and Kuma operations
- `docs/operations/11-devops-and-ci-cd.md` — staging deploy automation
- `specs/changes/VOC-088-persist-controlled-staging-oauth-signup-and-auto/` —
  package specification, tasks, and evidence
