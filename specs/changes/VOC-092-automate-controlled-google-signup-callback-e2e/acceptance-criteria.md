# VOC-092 — Acceptance Criteria

## VOC-092-AC-00 — Ephemeral harness exercises real OAuth start and callback HTTP handlers

- Requirement source: `VOC-092-D00`, issue #769 requirement 1
- Tasks: `VOC-092-T00`
- Tests: `VOC-092-TEST-00`, `VOC-092-TEST-01`
- Evidence: `VOC-092-EV-00`
- Result: pending

A repository-managed integration harness drives the real
`POST /api/v1/auth/oauth/google/start` and
`GET /api/v1/auth/oauth/google/callback` handlers over HTTP, including OAuth
state cookie issuance, callback query parameters, and redirect response semantics.

## VOC-092-AC-01 — Real GoogleOAuthProvider HTTP boundary via local fake provider

- Requirement source: `VOC-092-D00`, issue #769 requirement 1
- Tasks: `VOC-092-T00`
- Tests: `VOC-092-TEST-02`, `VOC-092-TEST-03`
- Evidence: `VOC-092-EV-00`
- Result: pending

The harness configures `GoogleOAuthProvider` with injectable token and userinfo
URLs pointing at a local fake Google OAuth server. Token exchange and userinfo
fetch run through the real provider implementation, not `FakeOAuthProvider`
substitution alone.

## VOC-092-AC-02 — Database-backed user, identity, and session behavior

- Requirement source: issue #769 requirement 1; `VOC-092-D03`
- Tasks: `VOC-092-T00`
- Tests: `VOC-092-TEST-04`
- Evidence: `VOC-092-EV-00`
- Result: pending

On allowlisted success, the harness observes persisted user, external identity,
and session records in a disposable Postgres database. The database is ephemeral
and isolated from staging and production.

## VOC-092-AC-03 — Allowlisted synthetic identity reaches authenticated onboarding/home path

- Requirement source: `VOC-092-D01`, `VOC-092-D02`, issue #769 requirement 2
- Tasks: `VOC-092-T00`
- Tests: `VOC-092-TEST-05`
- Evidence: `VOC-092-EV-00`, `VOC-092-EV-03`
- Result: pending

With `NEW_USER_SIGNUP_ENABLED=false` and a never-before-seen allowlisted
`@synthetic.vocanova.invalid` identity, the callback completes successfully and
the HTTP response redirects to the configured onboarding/home return URL without
HTTP 503.

## VOC-092-AC-04 — Unlisted synthetic identity receives stable HTTP 503 denial

- Requirement source: `VOC-092-D02`, issue #769 requirement 3
- Tasks: `VOC-092-T00`
- Tests: `VOC-092-TEST-06`
- Evidence: `VOC-092-EV-00`, `VOC-092-EV-03`
- Result: pending

With the same global signup policy and a never-before-seen unlisted synthetic
identity, the callback returns HTTP 503 with the stable new-signups-disabled
response associated with `auth.ErrSignupsDisabled`.

## VOC-092-AC-05 — Synthetic identities and ephemeral data only

- Requirement source: `VOC-092-D01`, issue #769 requirement 4
- Tasks: `VOC-092-T00`, `VOC-092-T01`, `VOC-092-T02`, `VOC-092-T03`
- Tests: `VOC-092-TEST-07`, `VOC-092-TEST-08`
- Evidence: `VOC-092-EV-00`, `VOC-092-EV-01`, `VOC-092-EV-02`, `VOC-092-EV-03`
- Result: pending

Harness fixtures, fake provider payloads, logs, artifacts, issues, PRs, and
evidence use only `@synthetic.vocanova.invalid` identities and ephemeral test
data. No real addresses, Google passwords, cookies, codes, state values, refresh
tokens, access tokens, or session values appear anywhere.

## VOC-092-AC-06 — No public test-auth route or runtime authentication bypass

- Requirement source: `VOC-092-D04`, issue #769 requirement 5
- Tasks: `VOC-092-T00`, `VOC-092-T01`
- Tests: `VOC-092-TEST-09`
- Evidence: `VOC-092-EV-00`, `VOC-092-EV-01`
- Result: pending

No staging/production test-auth route, provider switch, backdoor, or runtime
authentication bypass is added. The fake Google server is localhost/ephemeral
test infrastructure only.

## VOC-092-AC-07 — Staging OAuth/readiness synthetic retained unchanged

- Requirement source: `VOC-092-D06`, issue #769 requirement 6
- Tasks: `VOC-092-T03`
- Tests: `VOC-092-TEST-10`, regression on `VOC-088-TEST-05`
- Evidence: `VOC-092-EV-03`
- Result: pending

`synthetic.staging.oauth-expected-state`, `verify-staging-oauth-start.sh`, and
their scheduled-synthetics wiring remain functionally unchanged. Post-deployment
evidence shows the live staging synthetic still passes.

## VOC-092-AC-08 — CI and deterministic tests wired with operational-failure eligibility

- Requirement source: `VOC-092-D05`, issue #769 requirement 7
- Tasks: `VOC-092-T01`
- Tests: `VOC-092-TEST-11`, `VOC-092-TEST-12`
- Evidence: `VOC-092-EV-01`, `VOC-092-EV-03`
- Result: pending

Deterministic foundation tests and CI execute both harness cases. A failing CI
run fails closed. When the harness runs inside an observed workflow, failures are
eligible for the existing operational-failure observer without weakening
deduplication or sanitization.

## VOC-092-AC-09 — Harness boundary documented with human real-provider audit

- Requirement source: `VOC-092-D07`, issue #769 requirement 8
- Tasks: `VOC-092-T02`
- Tests: `VOC-092-TEST-13`
- Evidence: `VOC-092-EV-02`
- Result: pending

Operations documentation states what the harness proves, what Google's interactive
UI still requires, and the periodic human real-provider audit procedure.

## VOC-092-AC-10 — VOC-088-T03 non-blocking review findings remediated

- Requirement source: `VOC-092-D08`, issue #769 requirement 9
- Tasks: `VOC-092-T02`
- Tests: `VOC-092-TEST-14`
- Evidence: `VOC-092-EV-02`
- Result: pending

VOC-088 `t03-evidence.md` no longer contains the self-referential
`reviewed_sha` placeholder, and the scrubbed allowlisted sign-in outcome explicitly
records reaching onboarding/home without HTTP 503.

## VOC-092-AC-11 — Staging/production isolation preserved

- Requirement source: issue #769 requirement 10; `VOC-092-D03`, `VOC-092-D04`
- Tasks: `VOC-092-T00`, `VOC-092-T01`, `VOC-092-T03`
- Tests: `VOC-092-TEST-15`
- Evidence: `VOC-092-EV-00`, `VOC-092-EV-01`, `VOC-092-EV-03`
- Result: pending

The harness and CI jobs do not connect to staging/production databases, secrets,
directories, Docker networks, deploy users, or shared-edge hosts. Full repository
validation passes without requiring production credentials.
