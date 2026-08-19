# VOC-092 — Test Plan

## VOC-092-TEST-00 — Harness drives OAuth start over HTTP

- Covers: `VOC-092-AC-00`
- Preconditions: T00 integration test merged
- Procedure: Run the harness test; assert it POSTs to
  `/api/v1/auth/oauth/google/start` and receives HTTP 200 with a provider URL
  and Set-Cookie for OAuth state.
- Expected result: Start handler exercised without bypassing HTTP layer
- Evidence: `VOC-092-EV-00`

## VOC-092-TEST-01 — Harness drives OAuth callback over HTTP with state cookie

- Covers: `VOC-092-AC-00`
- Preconditions: T00 integration test
- Procedure: Continue harness flow; assert callback request includes state cookie
  matching issued state and receives redirect or error response via HTTP handler.
- Expected result: Callback handler and cookie validation exercised
- Evidence: `VOC-092-EV-00`

## VOC-092-TEST-02 — Fake Google token endpoint invoked by GoogleOAuthProvider

- Covers: `VOC-092-AC-01`
- Preconditions: T00 fake server records requests
- Procedure: Assert provider POSTs authorization code to configured token URL and
  parses access token response.
- Expected result: Real `GoogleOAuthProvider.Verify` hits fake token endpoint
- Evidence: `VOC-092-EV-00`

## VOC-092-TEST-03 — Fake Google userinfo endpoint invoked by GoogleOAuthProvider

- Covers: `VOC-092-AC-01`
- Preconditions: T00 fake server records requests
- Procedure: Assert provider GETs userinfo with bearer token and maps identity
  fields used by callback policy.
- Expected result: Real provider userinfo HTTP boundary exercised
- Evidence: `VOC-092-EV-00`

## VOC-092-TEST-04 — Allowlisted success persists user and session in disposable DB

- Covers: `VOC-092-AC-02`
- Preconditions: T00 success case
- Procedure: Query disposable Postgres after successful callback; assert user,
  external identity, and session rows exist for synthetic allowlisted identity.
- Expected result: Database-backed auth behavior verified
- Evidence: `VOC-092-EV-00`

## VOC-092-TEST-05 — Allowlisted never-before-seen synthetic identity redirects to onboarding/home

- Covers: `VOC-092-AC-03`
- Preconditions: `NEW_USER_SIGNUP_ENABLED=false`, allowlisted fixture, T00/T03
- Procedure: Assert HTTP 302/303 redirect Location matches configured app return
  URL (onboarding/home path) and status is not 503.
- Expected result: Controlled allowlisted first-time signup succeeds end-to-end
- Evidence: `VOC-092-EV-00`, `VOC-092-EV-03`

## VOC-092-TEST-06 — Unlisted never-before-seen synthetic identity returns HTTP 503

- Covers: `VOC-092-AC-04`
- Preconditions: Same policy with unlisted fixture, T00/T03
- Procedure: Assert callback HTTP 503 and stable new-signups-disabled body; assert
  no new user row.
- Expected result: Controlled signup denial matches `ErrSignupsDisabled` mapping
- Evidence: `VOC-092-EV-00`, `VOC-092-EV-03`

## VOC-092-TEST-07 — Fixtures use synthetic.vocanova.invalid only

- Covers: `VOC-092-AC-05`
- Preconditions: T00/T01 sources committed
- Procedure: Grep harness and foundation test fixtures for email patterns; assert
  only `@synthetic.vocanova.invalid` and no real-domain addresses.
- Expected result: Synthetic domain only
- Evidence: `VOC-092-EV-00`, `VOC-092-EV-01`

## VOC-092-TEST-08 — Logs and evidence omit OAuth secrets

- Covers: `VOC-092-AC-05`
- Preconditions: T00 test run with `-v` or CI log capture
- Procedure: Inspect test stdout and sample CI log; assert no authorization codes,
  access tokens, refresh tokens, session cookie values, or raw state tokens printed.
- Expected result: Redacted/scrubbed output
- Evidence: `VOC-092-EV-00`, `VOC-092-EV-03`

## VOC-092-TEST-09 — No test-auth route or runtime bypass introduced

- Covers: `VOC-092-AC-06`
- Preconditions: T00/T01 diff
- Procedure: Foundation test greps routes and production wiring for forbidden
  patterns (`test-auth`, `FakeOAuthProvider` production fallback toggles, public
  fake provider endpoints).
- Expected result: No bypass paths in production wiring
- Evidence: `VOC-092-EV-00`, `VOC-092-EV-01`

## VOC-092-TEST-10 — Staging OAuth synthetic script and registry unchanged

- Covers: `VOC-092-AC-07`
- Preconditions: T03 post-merge tree
- Procedure: Assert `infra/scripts/verify-staging-oauth-start.sh` and
  `synthetic.staging.oauth-expected-state` entry in `synthetics.yaml` have no
  functional regression diff from pre-VOC-092 baseline except documentation links;
  run `node --test scripts/foundation/voc088-staging-readiness.test.mjs`.
- Expected result: VOC-088 staging synthetic retained
- Evidence: `VOC-092-EV-03`

## VOC-092-TEST-11 — CI job executes harness on pull requests

- Covers: `VOC-092-AC-08`
- Preconditions: T01 workflow wiring
- Procedure: Parse committed workflow; assert PR trigger includes harness
  `go test` command.
- Expected result: CI wired
- Evidence: `VOC-092-EV-01`, `VOC-092-EV-03`

## VOC-092-TEST-12 — Foundation tests pass

- Covers: `VOC-092-AC-08`
- Preconditions: T01 foundation files
- Procedure:
  ```bash
  node --test scripts/foundation/voc092-controlled-signup-oauth-e2e.test.mjs
  ```
- Expected result: The T01 harness/CI foundation tests pass without depending on
  T02 documentation deliverables
- Evidence: `VOC-092-EV-01`

## VOC-092-TEST-13 — Operations doc states harness boundary and human audit

- Covers: `VOC-092-AC-09`
- Preconditions: T02 documentation
- Procedure: Read updated `docs/operations/staging-controlled-signup.md`; assert
  sections exist for harness scope, non-goals (Google UI), and human audit cadence.
- Expected result: Honest boundary documentation
- Evidence: `VOC-092-EV-02`

## VOC-092-TEST-14 — VOC-088 t03 evidence placeholder removed

- Covers: `VOC-092-AC-10`
- Preconditions: T02 evidence edit
- Procedure: Assert `t03-evidence.md` does not contain
  `bind-at-independent-review`; assert allowlisted row mentions onboarding/home
  without HTTP 503.
- Expected result: Non-blocking findings remediated
- Evidence: `VOC-092-EV-02`

## VOC-092-TEST-15 — Harness sources exclude staging/production hosts

- Covers: `VOC-092-AC-11`
- Preconditions: T00/T01 sources
- Procedure: Foundation test denylist for `api-staging.vocanova.site`,
  `api-production.vocanova.site`, and production/staging secret env names in harness
  test files.
- Expected result: Isolation preserved in committed harness
- Evidence: `VOC-092-EV-00`, `VOC-092-EV-01`

## VOC-092-TEST-16 — Operator documentation foundation tests pass

- Covers: `VOC-092-AC-09`, `VOC-092-AC-10`
- Preconditions: T02 documentation and evidence remediation committed
- Procedure:
  ```bash
  node --test scripts/foundation/voc092-operator-docs.test.mjs
  ```
- Expected result: The T02 documentation boundary and VOC-088 evidence checks pass
- Evidence: `VOC-092-EV-02`

Include positive, negative, authorization, failure, migration, accessibility, and
rollback coverage as applicable. Tests must not use secrets or production data.
