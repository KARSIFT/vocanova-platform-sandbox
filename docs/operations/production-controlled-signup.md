# Production controlled Google signup (VOC-096)

Operator guide for the persistent production controlled-signup cohort backed by
repository secret `PRODUCTION_NEW_USER_SIGNUP_ALLOWLIST`. Only
`.github/workflows/deploy-production.yml` may read this secret;
`deploy-staging.yml` never references it.

See also: [Staging controlled Google signup](staging-controlled-signup.md) for
the parallel staging procedure and the repository-managed OAuth callback E2E
harness (VOC-092).

## Policy

| Setting                     | Production value                                                                                          | Meaning                                                    |
| --------------------------- | --------------------------------------------------------------------------------------------------------- | ---------------------------------------------------------- |
| `GOOGLE_OAUTH_ENABLED`      | `true` when both `GOOGLE_OAUTH_CLIENT_ID` and `GOOGLE_OAUTH_CLIENT_SECRET` repository secrets are present | Google sign-in is available                                |
| `NEW_USER_SIGNUP_ENABLED`   | always `false`                                                                                            | No blanket first-time signup                               |
| `NEW_USER_SIGNUP_ALLOWLIST` | synced from `PRODUCTION_NEW_USER_SIGNUP_ALLOWLIST` on every deploy                                        | Only listed Google identities may create a production account |

The former `workflow_dispatch` input `new_user_signup_allowlist` was removed.
Automatic push deploys to `main` and manual dispatches both read the repository
secret only. A normal merge to `main` cannot silently erase the cohort.

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
2. Edit `PRODUCTION_NEW_USER_SIGNUP_ALLOWLIST`.
3. Replace the value with the complete comma-separated cohort (no spaces
   required; the API normalizes case and surrounding whitespace). GitHub does
   not reveal the prior value, and saving replaces the whole secret, so obtain
   the current cohort through the approved secure credential-management path
   before adding or removing a member.
4. Save the secret. Do not record the value in chat, tickets, terminal output,
   or evidence.

Removal is the same path: delete the address from the comma-separated list (or
clear the entire secret only when production Google OAuth is intentionally
disabled — see fail-closed behavior below).

## Pick up the change

Every push to `main` triggers `deploy-production.yml`, which:

1. Validates the secret on the runner (`infra/scripts/validate-production-signup-allowlist.sh`).
2. Writes `NEW_USER_SIGNUP_ALLOWLIST` into production `api.env`.
3. Redeploys the API container and runs the production smoke suite.

To apply a secret-only change without a code merge, manually dispatch
`deploy-production.yml` on `main` (Workflows → Deploy production → Run workflow).

### Verify the deploy

In the successful run log, confirm:

- Step **Validate production controlled-signup allowlist** succeeded.
- Step **Write production application configuration** logged
  `controlled signup ready: true` (when OAuth is enabled).
- Step **Run production smoke-test suite** passed.

Then confirm live readiness without exposing cohort metadata:

```bash
curl -fsS https://api-production.vocanova.site/healthz \
  | jq '{status, controlled_signup_ready, kill_switches}'
```

Expect `controlled_signup_ready: true` when OAuth is enabled and at least one
non-reserved allowlisted identity exists.

Optional harness (does not complete Google login):

```bash
EXPECT_OAUTH_ENABLED=true bash infra/scripts/verify-production-oauth-start.sh
```

## Repository-managed OAuth callback E2E harness (VOC-092)

CI exercises the full application OAuth **callback** contract without
interactive Google login. See
[Staging controlled Google signup — Repository-managed OAuth callback E2E harness](staging-controlled-signup.md#repository-managed-oauth-callback-e2e-harness-voc-092)
for scope, non-goals, and local run commands. The harness mirrors controlled
signup policy and does not touch production hosts, secrets, or databases.

The live production synthetic `synthetic.production.oauth-expected-state` and
`verify-production-oauth-start.sh` assert OAuth start, canonical callback, and
`controlled_signup_ready: true` on `/healthz` when OAuth is expected enabled.

## Prove cohort preservation across automatic deploys

After changing the secret, record two consecutive **push**-triggered
`deploy-production` successes on `main` without editing the secret again. Each
run must pass **Validate production controlled-signup allowlist** and log
`controlled signup ready: true`. That pair proves automatic deploys preserve
the secret-backed cohort (VOC-096-AC-00 / AC-01).

## Fail-closed behavior

When Google OAuth credentials are present and the allowlist secret is missing,
empty, multiline, or malformed, the deploy fails before writing `api.env`. This
is intentional: OAuth-enabled production with an empty controlled cohort blocks
every first-time Google signup.

If production must intentionally have no controlled cohort, keep the current secret
until a separately reviewed repository change has disabled the OAuth capability.
Do not bypass the validator, delete credentials as an ad hoc switch, or edit the
host environment manually.

## Human sign-in verification (no credentials in evidence)

Repository automation proves OAuth **start**, readiness, and the application
callback contract (see the VOC-092 harness section in the staging guide).
Completing Google login against the real provider still requires a human operator.

**Periodic real-provider audit:** run the checklist below after any auth callback
handler, allowlist policy, or Google OAuth client/redirect change, and at least
quarterly even when no such change occurred. A green callback harness or production
OAuth-start synthetic does not replace this audit.

Use the checklist after the secret change has deployed (or during the quarterly
audit):

| Check                              | How to verify                                                                            | Expected result                                                                                                                             |
| ---------------------------------- | ---------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------- |
| Allowlisted first-time Google user | Sign in at `https://production.vocanova.site/signin` with an identity present in the secret | Reaches production home/onboarding without HTTP 503                                                                                         |
| Unlisted Google user               | Attempt first-time sign-in with an identity **not** in the secret                        | Google may authenticate, but the API callback returns HTTP 503 with the stable "new sign-ups are disabled" body (`auth.ErrSignupsDisabled`) |
| Scheduled readiness                | Confirm latest `scheduled-synthetics` run                                                | Job `synthetic.production.oauth-expected-state` succeeded                                                                                   |
| Cohort persistence                 | Two consecutive push deploy successes after secret edit                                    | Both validate allowlist and log `controlled signup ready: true`                                                                             |

Record only pass/fail and run URLs in evidence. Never paste email addresses,
OAuth codes, session cookies, or callback query strings.

## Related documentation

- [Staging controlled Google signup](staging-controlled-signup.md) — staging
  procedure and VOC-092 callback harness details
- `docs/operations/monitoring.md` — scheduled synthetics and Kuma operations
- `docs/operations/11-devops-and-ci-cd.md` — production deploy automation
- `specs/changes/VOC-096-persist-controlled-production-google-signup-cohort/` —
  package specification, tasks, and evidence
