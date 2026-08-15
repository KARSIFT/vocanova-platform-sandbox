# VOC-084 — Restore repository-managed Google sign-in in staging: Specification

## Objective and requirement source

Close the defect reported in
[GitHub issue #691](https://github.com/KARSIFT/vocanova-platform-sandbox/issues/691)
(full body plus clarifying comment by `m-e-h-r-d-a-a-d` at
2026-08-15T21:38:52Z): restore staging Google sign-in through
repository-managed deployment, using the existing repository Google OAuth
credentials safely, keeping staging signups controlled, and proving OAuth
initiation without completing a real login.

Authority context recorded in the issue: founder/user remediation directive
dated 2026-08-16 authorizes issue creation, package adoption/implementation,
PRs, workflow execution, staging deployment, live verification,
remediation, and closure without routine approval. **This draft package still
does not adopt or authorize itself**; adoption remains a separate A-004
plan-review / adopt path. No Google Cloud mutation is authorized unless access
already exists; otherwise record the exact one-time external callback
registration required.

Primary context (issue #691 + drafting-time repo read):

| Item | Value |
|------|--------|
| Failing request | `POST /api/v1/auth/oauth/google/start` on staging → HTTP 503 |
| Production control | Same request → HTTP 200, `accounts.google.com`, production `/api/v1` callback |
| Staging host posture | `GOOGLE_OAUTH_ENABLED=false`; client id/secret absent; `NEW_USER_SIGNUP_ENABLED=false` |
| Production deploy | Syncs pair, rejects partial pair, derives callback, enables from availability |
| Staging deploy | Does not sync pair / canonicalize OAuth config |
| UI defect | Sign-in advertises "Continue with Google" while OAuth is disabled |
| Topology | Shared-edge + public health healthy; Cloudflare not implicated |

**Objective:** after this package's implementation, staging deploy converges a
coherent fail-closed Google OAuth state from repository secrets; OAuth start
returns 200 with the exact staging callback embedded; signup remains closed
except for an explicit empty-by-default allowlist; the UI does not advertise
Google when OAuth capability is disabled; deterministic tests cover the
branches; and Google client callback authorization is either verified with
evidence or reported as the sole precise external action.

## Confirmed findings (issue #691 + drafting-time re-read)

- `deploy-production.yml` already implements the safe pattern this package
  must mirror for staging: skip cleanly when both Google secrets are unset;
  refuse a partial pair; write `GOOGLE_OAUTH_CLIENT_ID` /
  `GOOGLE_OAUTH_CLIENT_SECRET` into the tier secret file with mode `0600`;
  derive `OAUTH_REDIRECT_URI=https://<api-host>/api/v1/auth/oauth/google/callback`;
  set `GOOGLE_OAUTH_ENABLED` from actual pair availability; keep
  `NEW_USER_SIGNUP_ENABLED=false` and write `NEW_USER_SIGNUP_ALLOWLIST` from a
  workflow input.
- `deploy-staging.yml` syncs AI, Sentry, and smoke-mint secrets into
  `/opt/vocanova/infra/secrets/api.env` but has **no** Google OAuth sync /
  canonical OAuth config write analogous to production.
- `apps/api` already exposes `kill_switches.oauth_enabled` on unauthenticated
  `GET /healthz` (`KillSwitchStatus` in `apps/api/app/api/production.go`).
- `apps/web/src/app/signin/page.tsx` unconditionally renders `<OAuthButton />`
  ("Continue with Google") with no capability gate.
- Staging and production secret directories, deploy users, databases, and
  Docker networks remain isolated under the single shared-edge nginx
  architecture (VOC-067); this package must not regress that.

## Scope and non-goals

In scope:

1. **Staging Google OAuth credential synchronization** mirroring the safe
   production pattern in `deploy-staging.yml`.
2. **Reject partial credentials** before application convergence; never log
   credential values.
3. **Canonical staging callback**:
   `OAUTH_REDIRECT_URI=https://api-staging.vocanova.site/api/v1/auth/oauth/google/callback`
   (exact).
4. **`GOOGLE_OAUTH_ENABLED` from actual complete credential availability**,
   removing stale contradictory values on each deploy convergence.
5. **Keep `NEW_USER_SIGNUP_ENABLED=false`** and add a controlled staging
   `NEW_USER_SIGNUP_ALLOWLIST` mechanism suitable for explicitly admitted
   testers. Empty means no first-time signup; existing users remain able to
   sign in.
6. **Web sign-in UI capability signal** so it does not advertise Google when
   OAuth is disabled. Prefer the existing `/healthz`
   `kill_switches.oauth_enabled` signal. Avoid a staging-only hardcoded lie or
   enabling a button merely because a build succeeded.
7. **Deterministic workflow/config/UI tests** for both-present, both-absent,
   partial-pair rejection, canonical callback, controlled allowlist, and
   disabled-method rendering.
8. **Post-deploy OAuth-start check** in `deploy-staging.yml` requiring HTTP
   200, an `accounts.google.com` authorization URL, and
   `redirect_uri` exactly
   `https://api-staging.vocanova.site/api/v1/auth/oauth/google/callback`. Do
   not follow Google or complete OAuth.
9. **Google client callback authorization**: determine from available
   access/evidence whether the existing client authorizes the staging
   callback; if Console/API access is unavailable, finish repository work and
   record the exact external configuration requirement without claiming it
   complete.
10. **Preserve** staging/production secret, directory, deploy-user,
    database, and Docker-network isolation and the single shared-edge nginx
    architecture.

Non-goals / explicitly excluded:

- Manual server edits, database edits, Cloudflare rules.
- Real OAuth login automation / completing the OAuth callback in CI.
- Production OAuth behavior changes (production already works and must not
  regress).
- Unrelated auth redesign (FakeOAuth redesign, magic-link provider work,
  session cookie redesign, etc.).
- Snapshot-then-recheck-drift promotion tasks (not applicable).
- Adopting or authorizing this package from within the draft.

## Risk and protected areas

Builder assessment: expected paths include
`.github/workflows/deploy-staging.yml` (R3 floor), authentication UI, and
staging secret synchronization. Drafting-time
`scripts/governance/classify-change-risk.sh --files-from` against the
expected list reported **Detected path-based risk floor: R3**.

This package **proposes R3** for the change as a whole. This is a **draft
proposal for the reviewing human at adoption time, not a determination**.
The independent verifier may raise to R4 if semantic review of
auth/secret/deploy consequences warrants it.

Protected areas: `.github/workflows/deploy-staging.yml`, authentication
UI/config/tests, staging secret synchronization. Do not weaken staging vs
production isolation.

Under **active A-004**, engineering-workflow gates require no founder
`approved` comment. EHR is not triggered by this drafting pass.

## Decisions, contradictions, security, and privacy

`VOC-084-D00` (recorded for traceability; formal acceptance at adoption):
Staging deploy must converge Google OAuth the same fail-closed way
production does — both secrets present, or both absent (coherent disabled);
never a partial pair.

`VOC-084-D01`: Canonical staging OAuth callback URI is exactly
`https://api-staging.vocanova.site/api/v1/auth/oauth/google/callback`.

`VOC-084-D02`: `GOOGLE_OAUTH_ENABLED` on staging is derived from actual
complete credential availability on each deploy, not left as a stale
manual host value.

`VOC-084-D03`: `NEW_USER_SIGNUP_ENABLED` remains `false` on staging. First-time
signup is only via an explicit repository-workflow-controlled allowlist
that defaults empty. Existing users remain able to sign in.

`VOC-084-D04`: The sign-in UI must not display "Continue with Google" when
live/deployed OAuth capability is disabled. The signal must be
deploy-derived / capability-based (prefer `/healthz`
`kill_switches.oauth_enabled`), not a staging-only hardcoded lie.

`VOC-084-D05`: CI must not complete a real OAuth login and must never print
credential values into repository, issue, PR, or workflow logs.

`VOC-084-D06`: Production OAuth sync/behavior is out of scope and must not
regress. Staging/production isolation and shared-edge topology must be
preserved.

Open questions for the reviewing human:

1. **Google callback authorization (`VOC-084-DEP-01`).** If Google Cloud
   Console/API access is unavailable to the implementer, is it acceptable for
   package closure to finish repository/deployment work with the exact
   external action recorded as the sole remaining ops step (issue #691
   default), rather than blocking forever on Console access?
2. **Allowlist control surface (`VOC-084-DEP-02`).** Accept the recommended
   default (mirror production's `workflow_dispatch` input
   `new_user_signup_allowlist`, default empty), or name a different
   repository-workflow control?
3. **Risk.** Accept proposed R3 (path floor R3), or elevate to R4 for
   auth/secret deploy sensitivity.

Security / privacy:

- Credential confidentiality: write only to staging secret file; mode
  `0600`; never log client id/secret values.
- OAuth redirect integrity: exact staging callback URI; do not follow Google
  in CI.
- Controlled account creation: signup kill switch stays off; allowlist
  defaults empty.
- No real-user data mutation required for acceptance.
- No secrets in repository, issue, PR, or workflow output.

## Data, migrations, analytics, and accessibility

- **Data / migrations:** None. No Atlas/Postgres schema migration. Allowlist
  and OAuth flags are environment/config only.
- **Analytics:** None expected — evidence-backed non-applicability.
- **Accessibility:** Sign-in UI changes must preserve existing accessible
  control labeling and focus behavior. If the Google button is hidden when
  disabled, remaining methods (magic link) must remain usable. Preserve the
  `max-w-[28rem]` workaround on the sign-in page per `.karsift/lessons.md`
  (do not "simplify" back to `max-w-md`).
