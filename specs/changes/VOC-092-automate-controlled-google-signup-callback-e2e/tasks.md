# VOC-092 — Tasks

None of the tasks below is implementation-authorized by this package.
Adoption and each task's own implementation authorization are separate.
Order is mandatory: **T00 → T01 → T02 → T03**.

## VOC-092-T00 — Ephemeral OAuth controlled-signup callback E2E harness

- Requirement source: `VOC-092-D00`–`D04`; issue #769 requirements 1–5 and 10
- Acceptance criteria: `VOC-092-AC-00` through `VOC-092-AC-06`, `VOC-092-AC-11`
- Tests: `VOC-092-TEST-00` through `VOC-092-TEST-09`, `VOC-092-TEST-15`
- Evidence: `VOC-092-EV-00` (`t00-evidence.md` in this package directory)
- Status: pending

### Required work

1. Add a Go integration test package under `apps/api/` (exact path is an
   implementer choice, e.g. `apps/api/app/api/controlled_signup_oauth_e2e_test.go`
   or a dedicated `apps/api/tests/controlled_signup_oauth/` package) that:
   - Starts a local fake Google OAuth HTTP server implementing token exchange
     and userinfo endpoints expected by `GoogleOAuthProvider`.
   - Starts disposable Postgres using the repository's existing integration-test
     pattern and applies migrations needed for auth tables.
   - Wires production-shaped auth HTTP handlers with kill switches
     `OAuthEnabled=true`, `NewSignupsEnabled=false`, and a controlled
     `SignupAllowlist`.
   - Drives `POST /api/v1/auth/oauth/google/start`, captures the OAuth state
     cookie, simulates the provider redirect, and calls
     `GET /api/v1/auth/oauth/google/callback` with synthetic code/state query
     parameters.
2. Implement the **allowlisted success** case:
   - Fixture identity is never-before-seen, allowlisted, and uses
     `@synthetic.vocanova.invalid` (not the reserved smoke-test bot).
   - Assert HTTP redirect to the configured onboarding/home return URL, persisted
     user/external identity/session records, and no HTTP 503.
3. Implement the **unlisted denial** case:
   - Fixture identity is never-before-seen, not allowlisted, synthetic domain only.
   - Assert HTTP 503 with stable new-signups-disabled body; no new user row.
4. Ensure test output and failure messages do not print OAuth codes, state,
   tokens, cookies, or authorization URLs with secrets.
5. Optionally add `infra/scripts/run-controlled-signup-oauth-e2e.sh` that runs
   the same `go test` target locally with explicit localhost-only guards and no
   staging/production host arguments.
6. Add Go unit-level helpers only if needed; do not modify production provider
   selection to expose a runtime fake provider.

### Explicitly out of scope for this task

- CI workflow wiring (T01).
- Documentation and VOC-088 evidence edits (T02).
- Live staging synthetic verification (T03).
- Changes to `verify-staging-oauth-start.sh` or staging deploy wiring.

## VOC-092-T01 — Wire CI and deterministic foundation tests

- Requirement source: `VOC-092-D05`; issue #769 requirement 7
- Acceptance criteria: `VOC-092-AC-05`, `VOC-092-AC-06`, `VOC-092-AC-08`,
  `VOC-092-AC-11`
- Tests: `VOC-092-TEST-07`, `VOC-092-TEST-08`, `VOC-092-TEST-09`,
  `VOC-092-TEST-11`, `VOC-092-TEST-12`, `VOC-092-TEST-15`
- Evidence: `VOC-092-EV-01` (`t01-evidence.md`)
- Status: pending — depends on `VOC-092-T00`

### Required work

1. Wire the T00 `go test` target into repository CI (`.github/workflows/` —
   prefer extending the existing API validation job or `repository-governance.yml`
   / karsift CI caller rather than inventing a parallel unchecked path). The job
   must run on pull requests and on `develop`.
2. Ensure the CI job has Docker available for disposable Postgres if required by
   the integration test pattern, or document and implement an equivalent
   isolated database strategy consistent with existing integration tests.
3. Add `scripts/foundation/voc092-controlled-signup-oauth-e2e.test.mjs` (or
   equivalent) covering:
   - Harness source files exist at committed paths.
   - CI workflow references the harness test command.
   - No forbidden test-auth route strings or production fake-provider switches.
   - Harness sources do not reference staging/production API hosts or real secret
     env var names beyond denylist checks.
   - Redaction expectations for OAuth-sensitive log patterns.
4. Register the foundation test in the repository's existing `pnpm test` or
   foundation test aggregation if applicable.
5. Confirm a failing harness test fails the CI job (fail-closed).

### Explicitly out of scope for this task

- Harness implementation changes except CI-discovered fixes strictly required for
  green CI (record in task PR if any).
- Documentation (T02).
- Live staging proof (T03).

## VOC-092-T02 — Document harness boundary and remediate VOC-088-T03 evidence

- Requirement source: `VOC-092-D07`, `VOC-092-D08`; issue #769 requirements 8–9
- Acceptance criteria: `VOC-092-AC-09`, `VOC-092-AC-10`
- Tests: `VOC-092-TEST-13`, `VOC-092-TEST-14`, `VOC-092-TEST-16`
- Evidence: `VOC-092-EV-02` (`t02-evidence.md`)
- Status: pending — depends on `VOC-092-T01`

### Required work

1. Update `docs/operations/staging-controlled-signup.md`:
   - Add a section describing the repository-managed OAuth callback E2E harness,
     what it proves, and what still requires human Google UI login.
   - Add or refine the periodic human real-provider audit procedure (cadence may
     be "after auth callback or allowlist policy changes and at least quarterly"
     unless product specifies otherwise — record choice in task PR).
   - Link to the local `go test` / optional shell wrapper command for operators.
2. Remediate `specs/changes/VOC-088-persist-controlled-staging-oauth-signup-and-auto/t03-evidence.md`:
   - Remove or replace `reviewed_sha: bind-at-independent-review` with the actual
     verifier-bound SHA (never a self-referential placeholder).
   - Update the allowlisted sign-in table row to explicitly state the scrubbed
     outcome reached onboarding/home without HTTP 503.
3. Add deterministic tests (`scripts/foundation/voc092-operator-docs.test.mjs` or
   extend the T01 foundation file) asserting the documentation contains required
   boundary language and the VOC-088 evidence no longer contains the placeholder
   `reviewed_sha` value.

### Explicitly out of scope for this task

- Harness or CI code changes except doc-linked command strings.
- Live staging synthetic run (T03).

## VOC-092-T03 — Record live CI and staging synthetic verification

- Requirement source: issue #769 evidence expectations; `VOC-092-D06`
- Acceptance criteria: `VOC-092-AC-03`, `VOC-092-AC-04`, `VOC-092-AC-07`,
  `VOC-092-AC-08`, `VOC-092-AC-11`
- Tests: `VOC-092-TEST-05`, `VOC-092-TEST-06`, `VOC-092-TEST-10`
- Evidence: `VOC-092-EV-03` (`t03-evidence.md`)
- Status: pending — depends on `VOC-092-T02`

### Required work

1. Record a CI run URL on `develop` after T00–T02 merge showing the harness job
   green with both allowlisted-success and unlisted-503 cases executed (scrubbed
   log excerpt only; no OAuth secrets).
2. Optionally record a deliberate failing run or test log excerpt proving fail-closed
   behavior if not already captured in T01 evidence.
3. Confirm live staging `synthetic.staging.oauth-expected-state` remains green after
   deployment (post-merge `scheduled-synthetics` or latest hourly run URL).
4. Run full repository validation (`pnpm validate` or documented equivalent) and
   record commands inspected in evidence.
5. All evidence scrubbed per `VOC-092-AC-05`.

### Explicitly out of scope for this task

- New application or workflow feature work.
- Modifying the staging synthetic harness behavior.

## Task ordering notes

- T00 blocks T01: CI wiring requires a committed harness target.
- T01 blocks T02: documentation should reference stable commands and CI path.
- T02 blocks T03: VOC-088 evidence remediation completes before closure evidence.
- No task may be dispatched before this package is adopted and
  implementation-authorized.

Tasks preserve scope, separation of duties, and rollback safety.
