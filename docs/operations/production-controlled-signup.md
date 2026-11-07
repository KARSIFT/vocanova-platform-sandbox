# Production Google signup

Production supports public Google signup and a restricted cohort. The restricted
cohort is backed by repository secret `PRODUCTION_NEW_USER_SIGNUP_ALLOWLIST`. Only
`.github/workflows/deploy-production.yml` may read this secret;
`deploy-staging.yml` never references it.

See also: [Staging controlled Google signup](staging-controlled-signup.md) for
the parallel staging procedure and the repository-managed OAuth callback E2E
harness (VOC-092).

## Policy

| Setting                     | Production value                                                                                          | Meaning                                                       |
| --------------------------- | --------------------------------------------------------------------------------------------------------- | ------------------------------------------------------------- |
| `GOOGLE_OAUTH_ENABLED`      | `true` when both `GOOGLE_OAUTH_CLIENT_ID` and `GOOGLE_OAUTH_CLIENT_SECRET` repository secrets are present | Google sign-in is available                                   |
| `NEW_USER_SIGNUP_ENABLED`   | Repository Actions variable `PRODUCTION_NEW_USER_SIGNUP_ENABLED`, default `false`                         | `true` opens registration to verified Google accounts         |
| `NEW_USER_SIGNUP_ALLOWLIST` | Synced from `PRODUCTION_NEW_USER_SIGNUP_ALLOWLIST` in restricted mode                                     | Only listed Google identities can register in restricted mode |

The former `workflow_dispatch` input `new_user_signup_allowlist` was removed.
Manual production deploys read the repository Actions policy and restricted-mode secret. Merging to `main`
deploys staging, not production; see [AGENTS.md](../../AGENTS.md).

Personal data rules:

- Store cohort member addresses only in the GitHub secret.
- Never commit addresses, never paste them into issues/PRs/evidence, and never
  expect deploy logs to print them. Successful validation logs only
  the signup mode and readiness.
- `/healthz` exposes only `controlled_signup_ready` (boolean). It never exposes
  email values or cohort cardinality.

## Open or restrict public signup

Set the repository Actions variable `PRODUCTION_NEW_USER_SIGNUP_ENABLED` to
`true` to open signup and dispatch the reviewed production workflow. Public
mode requires configured Google OAuth. In public mode, expect
`kill_switches.new_signups_enabled: true` and `controlled_signup_ready: false`
on `/healthz`. Confirm the real provider flow with a verified Google account
that has never signed in to production.

To return to the restricted cohort, set the variable to `false` and deploy
again. The restricted cohort secret is required for this mode.

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

Manually dispatch `deploy-production.yml` on the reviewed `main` revision. It:

1. Validates signup policy and the restricted-mode secret on the runner.
2. Writes signup mode and, in restricted mode, the allowlist into production `api.env`.
3. Redeploys the API container and runs the production smoke suite.

The same manual dispatch applies a secret-only change without a code merge
(Workflows → Deploy production → Run workflow).

### Verify the deploy

In the successful run log, confirm:

- Step **Validate production signup policy** succeeded.
- Step **Write production application configuration** logged
  `public signup enabled: true` or `controlled signup ready: true` for the selected mode.
- Step **Run production smoke-test suite** passed.

Then confirm live readiness without exposing cohort metadata:

```bash
curl -fsS https://api-production.vocanova.site/healthz \
  | jq '{status, controlled_signup_ready, kill_switches}'
```

Expect `controlled_signup_ready: true` when OAuth is enabled and at least one
non-reserved allowlisted identity exists.

Run the harness with the expected signup mode. This example checks public signup.

```bash
EXPECT_OAUTH_ENABLED=true EXPECT_NEW_SIGNUPS_ENABLED=true \
  bash infra/scripts/verify-production-oauth-start.sh
```

## Repository-managed OAuth callback E2E harness (VOC-092)

CI exercises the full application OAuth **callback** contract without
interactive Google login. See
[Staging controlled Google signup — Repository-managed OAuth callback E2E harness](staging-controlled-signup.md#repository-managed-oauth-callback-e2e-harness-voc-092)
for scope, non-goals, and local run commands. The harness mirrors controlled
signup policy and does not touch production hosts, secrets, or databases.

The live production synthetic `synthetic.production.oauth-expected-state` and
`verify-production-oauth-start.sh` assert OAuth start, canonical callback, and
the configured signup mode and matching `controlled_signup_ready` on `/healthz`.

## Prove cohort preservation across deploys

After changing the secret, record two consecutive manually dispatched
`deploy-production` successes on `main` without editing the secret again. Each
run must pass **Validate production signup policy** and log
`controlled signup ready: true`. That pair proves repeated deploys preserve
the secret-backed cohort (VOC-096-AC-00 / AC-01).

## Fail-closed behavior

In restricted mode, when Google OAuth credentials are present and the allowlist secret is missing,
empty, multiline, or malformed, the deploy fails before writing `api.env`. This
is intentional: OAuth-enabled production with an empty controlled cohort blocks
every first-time Google signup.

Public mode requires working Google OAuth credentials and bypasses cohort
validation. Apply signup policy through the deployment workflow.

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

| Check                              | How to verify                                                                               | Expected result                                                                       |
| ---------------------------------- | ------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------- |
| Allowlisted first-time Google user | Sign in at `https://production.vocanova.site/signin` with an identity present in the secret | Reaches production home/onboarding without HTTP 503                                   |
| Unlisted Google user               | Attempt first-time sign-in with an identity outside the cohort                              | Public mode reaches onboarding. Restricted mode returns the signup-disabled response. |
| Scheduled readiness                | Confirm latest `scheduled-synthetics` run                                                   | Job `synthetic.production.oauth-expected-state` succeeded                             |
| Cohort persistence                 | Two consecutive manual deploy successes after secret edit                                   | Both validate allowlist and log `controlled signup ready: true`                       |

Record only pass/fail and run URLs in evidence. Never paste email addresses,
OAuth codes, session cookies, or callback query strings.

## Related documentation

- [Staging controlled Google signup](staging-controlled-signup.md) — staging
  procedure and VOC-092 callback harness details
- `docs/operations/monitoring.md` — scheduled synthetics and Kuma operations
- `docs/operations/11-devops-and-ci-cd.md` — production deploy automation
- `specs/changes/VOC-096-persist-controlled-production-google-signup-cohort/` —
  package specification, tasks, and evidence
