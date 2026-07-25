# VOC-025 — A1 Staging and Rollback Evidence Collection

## Purpose

This document records the evidence required by `VOC-025-AC-07` (via
`VOC-025-TEST-21`..`VOC-025-TEST-23`). It describes the planned staging
exercises and rollback rehearsal for the A1 authentication and user foundation.

## Current status

Execution of live staging evidence is **blocked** by `VOC-025-DEP-01`: the F3
staging gate and supporting staging environment do not yet exist. T00–T04 have
implemented the code; T05 collects every piece of evidence that can be produced
in-repository and documents the remaining staging/rollback work for F3.

No A1-complete declaration is made here.

## Evidence collected in repository

| Evidence ID | Test | Status | Location |
|-------------|------|--------|----------|
| `VOC-025-EV-18` | `VOC-025-TEST-18` — Mock inventory and disposition | Collected | `specs/changes/VOC-025-begin-milestone-a1-authentication-and-user-foundat/mock-inventory.md` and `scripts/foundation/mock-inventory.mjs` |
| `VOC-025-EV-19` | `VOC-025-TEST-19` — Real-source authorization regression | Collected (N/A) | Mock inventory confirms no real learning-domain sources were wired; only the A1 auth shell was added. |
| `VOC-025-EV-20` | `VOC-025-TEST-20` — Installed deterministic and security suite | Collected | Run `pnpm run validate:workspace`, `pnpm run format:check`, `pnpm run lint`, `pnpm run typecheck`, `pnpm run test`, `pnpm run build` and `cd apps/api && go vet ./... && go test ./...` |
| `VOC-025-EV-21` | `VOC-025-TEST-21` — Staging auth and abuse validation | **Blocked** | Requires F3 staging environment with fake/test email and OAuth providers. See procedure below. |
| `VOC-025-EV-22` | `VOC-025-TEST-22` — Staging authorization and CSRF validation | **Blocked** | Requires F3 staging environment with two isolated test users and session revocation tooling. See procedure below. |
| `VOC-025-EV-23` | `VOC-025-TEST-23` — Rollback and recovery rehearsal | **Blocked** | Requires F3 staging candidate and approved rollback procedure. See procedure below. |

## Staging auth and abuse validation procedure (EV-21)

When F3 staging is available, perform the following with non-production
identities only:

1. **Magic-link happy path**
   - Request a link for a test email address.
   - Extract the link from the fake email provider or configured test inbox.
   - Consume the link and verify the response sets `vocanova_session` and
     `vocanova_csrf` cookies.
   - Call `GET /api/v1/me` and confirm the verified email is returned.

2. **Magic-link negative cases**
   - Attempt to consume an expired link: expect `401 Unauthorized`.
   - Attempt to replay a consumed link: expect `401 Unauthorized`.
   - Attempt to consume a tampered token or wrong-environment token: expect
     `401 Unauthorized`.
   - Confirm responses do not differ between registered and unregistered
     emails at the request stage.

3. **Google OAuth happy path**
   - Start OAuth with an allowed staging `redirectUri`.
   - Verify the OAuth state cookie is set and the returned URL is the real
     provider authorization endpoint.
   - Complete the provider flow with a test Google account.
   - Confirm callback creates/links exactly one internal identity and sets
     session/CSRF cookies.

4. **OAuth attack paths**
   - Replay a consumed state: expect `401 Unauthorized`.
   - Callback with a mismatched state cookie: expect `401 Unauthorized`.
   - Callback with a disabled user: expect `401 Unauthorized`.
   - Attempt to start OAuth with a disallowed `redirectUri`: expect
     `401 Unauthorized`.

5. **Rate-limit and abuse**
   - Exceed the configured request/consume/start/callback thresholds from a
     single IP.
   - Confirm provider/email work is blocked and a `429 Too Many Requests`
     response is returned.
   - Advance the clock past the window and confirm the service resumes.

6. **Navigation**
   - Authenticate and visit `/home`, `/discover`, `/discover/<situation>`,
     `/discover/<situation>/<word>`, and `/progress`.
   - Confirm all render without an auth redirect.

## Staging authorization and CSRF validation procedure (EV-22)

When F3 staging is available:

1. **Unauthenticated access**
   - With no session cookie, request `GET /api/v1/me` and each `(app)` route.
   - Confirm API returns `401 Unauthorized` and routes redirect to `/signin`.

2. **Cross-user isolation**
   - Create two isolated users (A and B).
   - Authenticate as A and request `GET /api/v1/me` with B's session cookie:
     expect `401 Unauthorized`.
   - Confirm no requester-scoped private resource leaks data across users.
   - Confirm inaccessible private resources return `404 Not Found`.

3. **Revoked/expired session**
   - Log out as A, then reuse the old session cookie: expect `401 Unauthorized`.
   - Advance the clock past 30 days and reuse the cookie: expect
     `401 Unauthorized`.

4. **CSRF enforcement**
   - Call `POST /api/v1/auth/logout` without `X-CSRF-Token`: expect `403 Forbidden`.
   - Call with a mismatched token: expect `403 Forbidden`.
   - Call with a valid cookie/header pair: expect `204 No Content`.

5. **Sensitive-data redaction**
   - Inspect response bodies, logs, database rows, and the OpenAPI document.
   - Confirm no raw session tokens, magic-link bearers, OAuth codes, or secrets
     are exposed.

## Rollback and recovery rehearsal procedure (EV-23)

When a staged candidate is ready and the rollback procedure is approved:

1. **Pre-rollback**
   - Capture a backup/snapshot of the staging database after migration.
   - Record the last-known-good code revision.
   - Identify all active sessions in the database.

2. **Rollback**
   - Stop the API service.
   - Revert the code to the last-known-good revision.
   - Restore the database to the pre-migration state using the approved
     procedure.

3. **Post-rollback safety**
   - Confirm all sessions created by the candidate are invalidated (do not
     trust restored revoked/expiry state alone; explicitly revoke or delete
     candidate-era sessions).
   - Confirm no raw bearer tokens remain in logs or test artifacts.
   - Confirm users and external-identity rows are consistent with the restored
     state.

4. **Service validation**
   - Start the API service.
   - Run health checks.
   - Attempt authentication with a pre-candidate session: expect rejection.
   - Attempt a fresh magic-link flow: expect success only if the rollback
     deliberately retained the A1 schema; otherwise expect the A1 routes to be
     unavailable until the next forward migration.

## Dependency blocker

`VOC-025-DEP-01`: F3 staging gate must pass before A1 acceptance.
`VOC-025-T05`'s live staging evidence (EV-21, EV-22, EV-23) cannot be executed
until F3 exists with disposable PostgreSQL, fake/test email and OAuth
providers, and an approved rollback procedure.
