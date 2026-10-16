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

`VERSION` is a committed semantic-version base, starting at `0.4.1` for the
automatic-version policy introduced on 16 October 2026. The shared release script
adds the number of first-parent commits since the most recent change to `VERSION`
to its patch component. With squash merging, each merged PR therefore advances
one patch: `0.4.1`, `0.4.2`, `0.4.3`, including documentation-only changes. A
multi-commit topic branch merged normally also counts as one mainline release.
The build uses checked-out Git history, never a workflow run number or a mutable
"latest" tag. Rerunning a deployment or promoting the same commit to production
keeps the version unchanged. Rolling back an older commit restores its version.

Every push to `main` now deploys staging, without a path filter. No version-bump
commit is pushed by automation, so there is no recursive workflow or write access
to the protected branch. Both deployment checkouts require full Git history;
shallow histories and uncommitted `VERSION` changes fail instead of guessing.
To begin a deliberate minor or major release, change `VERSION` in that release's
PR (for example `0.5.0`); subsequent merges increment from that new base. Choose a
base higher than the currently deployed version. Ordinary patches need no edit.
The workspace package versions are package/tooling metadata, not this product
release number.

Staging and production workflows inject the same version, full commit, explicit
environment and UTC build timestamp into their API and web builds. These are
public build metadata, never runtime configuration. Commit image tags are scoped
by environment (`staging-sha-…` and `production-sha-…`), because the web bundle
and image identity differ between environments. The existing `dev` and `prod`
convenience tags remain available. Settings → About shows the
version, environment and abbreviated commit. Both services expose `/version` with
`version`, `commit`, `environment`, and `builtAt`. Both deployment workflows
verify these public responses against the intended build after health checks,
so a stale or mixed release fails deployment verification. An `unknown` field means the build metadata was absent or invalid.

- [Staging web version](https://staging.vocanova.site/version)
- [Staging API version](https://api-staging.vocanova.site/version)
- [Production web version](https://production.vocanova.site/version)
- [Production API version](https://api-production.vocanova.site/version)

Merging to `main` deploys staging. Production is still manually dispatched through
`deploy-production.yml`. A failed workflow is not proof that the old containers
remain active: inspect both `/version` responses and the deploy failure stage.
