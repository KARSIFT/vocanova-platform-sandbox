# VOC-092 — Automate controlled Google signup callback E2E safely: Specification

## Objective and requirement source

Deliver repository-managed, privacy-preserving OAuth controlled-signup E2E coverage
that would catch the original empty/runtime allowlist regression and
callback-policy regression described in
[GitHub issue #769](https://github.com/KARSIFT/vocanova-platform-sandbox/issues/769).

**This draft package does not adopt or authorize itself**; adoption remains a
separate A-004 plan-review / adopt path.

Primary context (issue #769 + drafting-time repo read 2026-08-19):

| Item | Value |
|------|-------|
| Current automation | `verify-staging-oauth-start.sh` + `synthetic.staging.oauth-expected-state` assert OAuth start, accounts.google.com target, canonical callback, blanket signup disabled, and `controlled_signup_ready=true` |
| Gap | No automated exercise of OAuth callback handler, state cookie validation, real `GoogleOAuthProvider` HTTP exchange, DB user/session creation, or controlled-signup allowlist enforcement |
| VOC-088 human evidence | Allowlisted first-time Google user reached staging without HTTP 503; unlisted user received stable HTTP 503 denial (scrubbed in `t03-evidence.md`) |
| VOC-088-T03 review findings | Self-referential `reviewed_sha: bind-at-independent-review` placeholder; allowlisted outcome should explicitly record reaching onboarding/home without HTTP 503 |
| Provider wiring | `GoogleOAuthProvider` supports injectable `TokenURL` and `UserInfoURL`; production uses real Google endpoints only |
| Reserved synthetic | `smoke-test-bot@synthetic.vocanova.invalid` is deploy-seeded and must not be used as a harness signup identity |
| Policy | `NEW_USER_SIGNUP_ENABLED=false` with `SignupAllowlist` admits controlled first-time signups only |

## Scope and non-goals

In scope:

1. **Ephemeral E2E harness** (Go integration test and optional local runner script)
   that drives the real HTTP stack for:
   - `POST /api/v1/auth/oauth/google/start`
   - OAuth state cookie issuance and validation
   - Simulated provider redirect to `GET /api/v1/auth/oauth/google/callback`
   - Real `GoogleOAuthProvider.Verify` against a local fake Google OAuth server
     implementing token exchange and userinfo endpoints
   - Database-backed user, external identity, and session persistence (disposable
     Postgres; same isolation pattern as existing API integration tests)
   - Final redirect to the configured app return URL (onboarding/home path)
2. **Allowlisted success case:** a never-before-seen synthetic allowlisted email
   completes the callback and reaches the authenticated redirect target while global
   signup remains disabled.
3. **Unlisted denial case:** a never-before-seen synthetic unlisted email receives
   HTTP 503 with the stable new-signups-disabled response body mapped from
   `auth.ErrSignupsDisabled`.
4. **Synthetic data only:** harness identities use `@synthetic.vocanova.invalid`
   addresses distinct from the reserved deploy smoke-test identity.
5. **CI wiring:** deterministic execution on pull requests and integration branch
   CI; failures fail closed and are eligible for observation by
   `operational-failure-monitoring.yml` when the harness runs in an observed
   workflow.
6. **Foundation tests:** `scripts/foundation/voc092-*.test.mjs` lock harness
   presence, CI wiring, redaction rules, and that no test-auth route or runtime
   bypass was introduced.
7. **Documentation:** update `docs/operations/staging-controlled-signup.md` with:
   - what the harness proves vs what still requires human Google UI login
   - periodic human real-provider audit procedure
8. **VOC-088 evidence remediation:** fix `t03-evidence.md` non-blocking findings
   without altering VOC-088 scope or re-opening closed tasks.
9. **Retention proof:** post-merge evidence that
   `synthetic.staging.oauth-expected-state` remains green and unchanged.

Non-goals / explicitly excluded:

- Automating Google's interactive consent/login UI, CAPTCHA, or account chooser.
- Storing, logging, or replaying real Google credentials, cookies, authorization
  codes, OAuth state values, refresh tokens, access tokens, or session values.
- Adding a public staging/production test-auth route, `?test=1` bypass, provider
  switch env for production, or any runtime authentication backdoor.
- Changing production or staging signup policy, allowlist secrets, or deploy wiring.
- Modifying `synthetic.staging.oauth-expected-state` behavior or coverage (retain
  as-is).
- Connecting the harness to staging or production databases, secrets, Docker
  networks, or shared edge hosts.
- Self-adoption / self-authorization of this package.

## Risk and protected areas

- **Draft package proposal:** **R3** (auth paths in `apps/api/`, CI workflows).
- **Measured path floor at drafting:** **R3** for `apps/api/` and
  `.github/workflows/`. Not R4 unless a task touches `scripts/governance/*` or
  introduces a production runtime bypass (explicitly forbidden).
- Protected areas: `apps/api/business/auth/`, `apps/api/app/api/auth.go`,
  `.github/workflows/` CI entrypoints, `docs/operations/staging-controlled-signup.md`,
  VOC-088 `t03-evidence.md` (evidence-only edits).
- EHR: not triggered.
- Active authority model: **A-004**. No founder `approved` comment is a
  merge/adopt/release gate.

## Decisions

`VOC-092-D00`: The harness must exercise the **real** application HTTP handlers
and the **real** `GoogleOAuthProvider` implementation with injectable endpoint
URLs pointing at a **local fake Google OAuth HTTP server** started inside the test
process or ephemeral test fixture. Unit-level `FakeOAuthProvider` substitution alone
does not satisfy the requirement because it skips the provider HTTP boundary the
issue targets.

`VOC-092-D01`: Harness identities must use `@synthetic.vocanova.invalid` emails
that are not the reserved deploy smoke-test address
(`smoke-test-bot@synthetic.vocanova.invalid`). Example fixture names may use
deterministic suffixes such as `allowlisted-callback-e2e@synthetic.vocanova.invalid`
and `unlisted-callback-e2e@synthetic.vocanova.invalid` (implementer may choose
final names; must remain in the reserved synthetic domain).

`VOC-092-D02`: Kill-switch configuration in the harness must mirror staging
controlled signup: `OAuthEnabled=true`, `NewSignupsEnabled=false`, non-empty
`SignupAllowlist` containing only the allowlisted fixture identity for the success
case; the denial case uses the same policy with an allowlist that excludes the
unlisted fixture identity.

`VOC-092-D03`: The harness runs against **disposable Postgres only**, started via
the repository's existing integration-test pattern (`testcontainers` or equivalent
already used in `apps/api`). It must not read `DATABASE_URL` from the environment
when that could target staging/production.

`VOC-092-D04`: No new public HTTP route may be added for test authentication.
The fake Google server is bound to localhost/ephemeral ports inside the test
harness only and is not registered in production wiring or exposed through the
shared edge.

`VOC-092-D05`: CI executes the Go integration test package (or a dedicated
`go test` target) on every relevant PR and on `develop`. An optional
`infra/scripts/run-controlled-signup-oauth-e2e.sh` wrapper may exist for local
operator rehearsal but must not accept staging/production hosts or real credentials.

`VOC-092-D06`: `synthetic.staging.oauth-expected-state` and
`verify-staging-oauth-start.sh` remain unchanged. VOC-092 adds complementary CI
coverage; it does not replace the live staging synthetic.

`VOC-092-D07`: Documentation must state clearly:
- **Harness proves:** application OAuth start/callback contract, state cookie
  validation, provider token/userinfo HTTP handling, DB persistence, controlled
  signup policy, redirect behavior.
- **Harness does not prove:** Google account login UI, consent screen changes,
  Google-side redirect URI registration drift, or real-user identity edge cases.
- **Human audit:** periodic interactive sign-in using the checklist already in
  `docs/operations/staging-controlled-signup.md`, updated to reference the harness
  boundary.

`VOC-092-D08`: VOC-088 `t03-evidence.md` remediation is limited to:
- Replace `reviewed_sha: bind-at-independent-review` with the actual independent
  review SHA once known at remediation time, or remove the placeholder field until
  the verifier binds it (never self-referential placeholder text).
- Expand the scrubbed allowlisted sign-in row to explicitly state the user reached
  onboarding/home without HTTP 503.

## Open questions

None blocking package drafting. Implementer chooses exact file names for the
integration test and optional shell wrapper, provided foundation tests lock the
committed paths.

## Security and privacy

- No real email addresses, Google secrets, OAuth codes, state values, tokens, or
  session cookies in logs, test output, issues, PRs, or evidence.
- Fake provider responses use synthetic subject IDs and `@synthetic.vocanova.invalid`
  emails only.
- Harness must redact or omit authorization URLs, callback query strings, and
  Set-Cookie values from CI logs (assert outcomes without printing secrets).
- No staging/production database or secret access from CI harness jobs.

## Data, migrations, analytics, and accessibility

- **Data/migrations:** None. Harness uses ephemeral databases torn down after tests.
- **Analytics:** None — evidence-backed non-applicability.
- **Accessibility:** No product UI change. Test and documentation changes only.
