# Account email and release operations

Google OAuth and password authentication are independent. Staging Google login
already works. Adding passwords does not require changing its client credentials
or callback URLs.

## Activate verification and recovery email

The existing HTTP email adapter matches [Resend's send-email API](https://resend.com/docs/api-reference/emails/send-email).
Use a verified sending domain and a sending-only API key. Follow
[Resend's domain verification instructions](https://resend.com/docs/dashboard/domains/introduction)
for the exact DNS records supplied for your domain; do not invent DNS values.
A provider account and verified domain are external prerequisites, not created by
this code change. Other providers require a compatible adapter.

Configure these values in the API's existing protected environment file on the
target host, using the established secret-management process. Do not put the key
in a PR, chat, browser variable, image build argument, or shell history:

| Variable                 | Value                                 |
| ------------------------ | ------------------------------------- |
| `EMAIL_PROVIDER_URL`     | `https://api.resend.com/emails`       |
| `EMAIL_PROVIDER_API_KEY` | Sending-only key from the provider    |
| `EMAIL_FROM`             | Sender address on the verified domain |
| `EMAIL_PASSWORD_ENABLED` | `true` after the above are configured |

Keep the environment file mode `0600`. Restart/recreate the API using the normal
deployment workflow. Existing host configuration persists across deploys. Google
login remains available while password authentication is disabled.

`/healthz` exposes `kill_switches.password_enabled` only when the password switch
and all three email settings are present. This is configuration readiness, not a
live deliverability guarantee. Missing email configuration keeps password endpoints
disabled and the password option hidden. Do not enable a fake sender for real users.

Staging's controlled-signup allowlist still applies to new password accounts.
Existing verified Google users can choose **Settings → Account → Add a password**;
the emailed proof is required before a password is attached.

After configuration, use a controlled test mailbox to verify delivery, create an
account through the verification link, sign in, request a reset, and confirm the
old password and previous sessions no longer work. Test expired/reused links and
Google login as well. Use the provider delivery dashboard to diagnose rejected
mail; never copy proof URLs or API keys into logs or issues. Production activation
is a separate manual deploy and must repeat the delivery check.

## Identify a deployed release

`VERSION` is the human-readable release version, beginning with `0.2.0` for this
account/settings release. Bump it in a PR when releasing a new feature set or fix:
minor for additive features, patch for compatible fixes, major for incompatible
changes. Individual deployments are distinguished by their full Git commit even
when the human-readable version is unchanged.

Staging and production workflows inject the same version, full commit, explicit
environment and UTC build timestamp into their API and web builds. These are
public build metadata, never runtime configuration. Settings → About shows the
version, environment and abbreviated commit. Both services expose `/version` with
`version`, `commit`, `environment`, and `builtAt`; compare both services after a
deploy. An `unknown` field means the build metadata was absent or invalid.

- [Staging web version](https://staging.vocanova.site/version)
- [Staging API version](https://api-staging.vocanova.site/version)
- [Production web version](https://production.vocanova.site/version)
- [Production API version](https://api-production.vocanova.site/version)

Merging to `main` deploys staging. Production is still manually dispatched through
`deploy-production.yml`. A failed workflow is not proof that the old containers
remain active: inspect both `/version` responses and the deploy failure stage.
