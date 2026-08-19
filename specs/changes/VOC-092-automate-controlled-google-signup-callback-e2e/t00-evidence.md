---
evidence_id: VOC-092-EV-00
task_id: VOC-092-T00
acceptance_criteria:
  - VOC-092-AC-00
  - VOC-092-AC-01
  - VOC-092-AC-02
  - VOC-092-AC-03
  - VOC-092-AC-04
  - VOC-092-AC-05
  - VOC-092-AC-06
  - VOC-092-AC-11
tests:
  - VOC-092-TEST-00
  - VOC-092-TEST-01
  - VOC-092-TEST-02
  - VOC-092-TEST-03
  - VOC-092-TEST-04
  - VOC-092-TEST-05
  - VOC-092-TEST-06
  - VOC-092-TEST-07
  - VOC-092-TEST-08
  - VOC-092-TEST-09
  - VOC-092-TEST-15
date: 2026-08-19
related_change: VOC-092
accountable_owner: unassigned
---

# VOC-092-T00 — Ephemeral OAuth controlled-signup callback E2E harness

## Scope and outcome

This task adds a repository-managed Go integration harness that exercises the
real OAuth start and callback HTTP handlers, the real `GoogleOAuthProvider`
token/userinfo HTTP boundary against a local fake Google server, and
database-backed user/identity/session persistence in disposable loopback
Postgres. Both allowlisted-success and unlisted-503 denial cases run with
`OAuthEnabled=true`, `NewSignupsEnabled=false`, and a controlled
`SignupAllowlist`. No production wiring, staging deploy configuration, CI
workflow, or public test-auth route is changed in this task.

## Repository deliverables

| Artifact | Path |
| -------- | ---- |
| OAuth callback E2E harness | `apps/api/app/api/controlled_signup_oauth_e2e_test.go` |
| Disposable Postgres helper | `apps/api/app/api/controlled_signup_oauth_postgres_test.go` |
| Optional local operator wrapper | `infra/scripts/run-controlled-signup-oauth-e2e.sh` |

## Acceptance mapping

| Acceptance criterion | Repository result |
| -------------------- | ----------------- |
| AC-00 | Harness POSTs `/api/v1/auth/oauth/google/start`, captures OAuth state cookie, and GETs `/api/v1/auth/oauth/google/callback` through real `RegisterAuth` handlers. |
| AC-01 | Real `GoogleOAuthProvider` with injectable `TokenURL` / `UserInfoURL` against `httptest` fake Google server; token and userinfo call counts asserted. |
| AC-02 | Disposable `postgres:16-alpine` on `127.0.0.1` with forward migrations applied; user, external identity, and session rows asserted on success. Harness never reads `DATABASE_URL`. |
| AC-03 | Allowlisted `@synthetic.vocanova.invalid` identity receives HTTP 302 redirect to configured onboarding return URL with persisted auth rows. |
| AC-04 | Unlisted synthetic identity receives HTTP 503 with stable `new sign-ups are disabled` body; no user row created. |
| AC-05 | Fixtures use `allowlisted-callback-e2e@synthetic.vocanova.invalid` and `unlisted-callback-e2e@synthetic.vocanova.invalid` only; reserved smoke-test bot email is configured but not used as a signup identity. Verbose test output omits OAuth codes, state, tokens, and cookie values (see log redaction section). |
| AC-06 | Fake Google server is `httptest.NewServer` inside the test process only; no production provider switch or public test-auth route added. |
| AC-11 | Shell wrapper refuses staging/production host arguments and accepts no arguments; Postgres container binds loopback only. |

## Deterministic validation

```bash
cd apps/api && go test ./app/api/... -run ControlledSignupOAuth -count=1 -v
```

Scrubbed result (2026-08-19):

```
=== RUN   TestControlledSignupOAuth_AllowlistedCallbackSucceeds
    controlled_signup_oauth_e2e_test.go:310: controlled-signup OAuth allowlisted callback succeeded with redirect to onboarding and persisted auth rows
--- PASS: TestControlledSignupOAuth_AllowlistedCallbackSucceeds (7.25s)
=== RUN   TestControlledSignupOAuth_UnlistedCallbackDenied
    controlled_signup_oauth_e2e_test.go:346: controlled-signup OAuth unlisted callback denied with HTTP 503 and no persisted user
--- PASS: TestControlledSignupOAuth_UnlistedCallbackDenied (2.12s)
PASS
ok  	github.com/KARSIFT/vocanova-platform/apps/api/app/api	9.370s
```

Optional local operator wrapper (same harness, scrubbed excerpt):

```bash
bash infra/scripts/run-controlled-signup-oauth-e2e.sh
```

```
--- PASS: TestControlledSignupOAuth_AllowlistedCallbackSucceeds (2.14s)
--- PASS: TestControlledSignupOAuth_UnlistedCallbackDenied (2.13s)
PASS
ok  	github.com/KARSIFT/vocanova-platform/apps/api/app/api	4.281s
```

Result: 2 tests passed, 0 failed.

## Log redaction (VOC-092-TEST-08)

Verbose `-v` output contains only outcome summary lines from `t.Log`:

- `controlled-signup OAuth allowlisted callback succeeded with redirect to onboarding and persisted auth rows`
- `controlled-signup OAuth unlisted callback denied with HTTP 503 and no persisted user`

The harness does not print authorization URLs, OAuth `code` or `state` query
values, access tokens, session cookie values, or Set-Cookie headers on success
or failure paths. Assertions validate outcomes without logging secret-bearing
request/response fields.

## Isolation notes

- Postgres container name prefix: `voc092-controlled-signup-oauth-*`, bound to
  `127.0.0.1` only.
- Migrations applied from `apps/api/migrations/*.sql` in version order via direct
  `db.Exec`; no Atlas CLI or environment `DATABASE_URL` consumption.
- Harness skips cleanly when Docker is absent (`t.Skipf`); permanent CI wiring
  is VOC-092-T01 scope.

## Out of scope for this task

CI workflow wiring, foundation tests, operations documentation, VOC-088 evidence
remediation, and live staging synthetic verification belong to VOC-092-T01
through T03.
