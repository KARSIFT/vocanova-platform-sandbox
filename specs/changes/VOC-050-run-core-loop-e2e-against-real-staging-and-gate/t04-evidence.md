# VOC-050-EV-04 — T04 production core-loop activation evidence

Evidence for `VOC-050-T04` (`VOC-050-AC-05`, test `VOC-050-TEST-05`).

This file is the durable in-repo record because the implementer in this workflow
does not author the generated pull-request body directly.

`VOC-050-TEST-05` requires the smoke suite's `# 5.` section to print `PASS:`
lines instead of `SKIP:` when run "exactly as `deploy-production.yml`'s step
will invoke it", against "a real or disposable production-shaped target with the
synthetic account seeded and `SMOKE_TEST_SESSION_COOKIE` set to a value minted
by `T01`'s mechanism". The rehearsal below is that disposable target: real
Postgres, real `apps/api/scripts/migrate.sh`, real
`apps/api/scripts/seed-synthetic-smoke-user.sh`, real `apps/api/cmd/api` server,
real `POST /ops/synthetic-smoke-test/session` mint, and the unmodified
`# 5.` checks in `infra/scripts/smoke-test-production.sh`. Nothing is stubbed
except the web tier, which section `# 5.` never touches.

## 0. Why the previous revision needed remediation

The previous revision of this task (commit `602f7800`) passed the mint
endpoint's `session_cookie` value straight into `SMOKE_TEST_SESSION_COOKIE`.
That value is the **opaque** session token
(`apps/api/app/api/production.go`'s `SessionCookie` field), whereas
`smoke-test-production.sh` sends the variable as an entire `Cookie:` header
(`-H "Cookie: $smoke_session_cookie"`), so the API would have seen no
`vocanova_session` cookie at all. Independent review flagged this as a High
finding; case C below reproduces the failure it predicted, and case D
demonstrates the fix (`vocanova_session=<opaque token>`).

## 1. Disposable production-shaped target

```
$ docker run -d --name voc050-t04-pg \
    -e POSTGRES_USER=vocanova -e POSTGRES_PASSWORD=vocanova -e POSTGRES_DB=vocanova \
    -p 55432:5432 postgres:16-alpine

$ DATABASE_URL='postgres://vocanova:vocanova@127.0.0.1:55432/vocanova?sslmode=disable' \
    sh apps/api/scripts/migrate.sh
  ...
  -- migrating version 20260808141000
    -> ALTER TABLE users
         ADD COLUMN is_synthetic_test_account boolean NOT NULL DEFAULT false;
    -> CREATE UNIQUE INDEX users_single_synthetic_test_account_idx
         ON users (is_synthetic_test_account)
         WHERE is_synthetic_test_account = true;
  -- ok (4.545861ms)
  -- 14 migrations
  -- 76 sql statements

$ DOCKER_COMPOSE_CMD=<shim onto the disposable container> \
    sh apps/api/scripts/seed-synthetic-smoke-user.sh
  NOTICE:  VOC-050 seed: created synthetic smoke-test account smoke-test-bot@synthetic.vocanova.invalid
  seeded synthetic smoke-test user: smoke-test-bot@synthetic.vocanova.invalid
```

Atlas was the same pinned `v1.2.0` build `deploy-production.yml` installs. The
`DOCKER_COMPOSE_CMD` shim only rewrites `<cmd> exec -T postgres ...` onto
`docker exec -i voc050-t04-pg ...`; the seed script and its SQL ran unmodified.

The API was then started from source with production-shaped configuration
(`ENVIRONMENT=production`, all three auth kill switches off, AI on) and
`SMOKE_TEST_SESSION_MINT_TOKEN` set, so `T01`'s mint route is registered:

```
api: listening on :18080 (env=production, ai=on, magic=off, oauth=off, signups=off)
```

A plain `python3 -m http.server` on `127.0.0.1:18081` stood in for the web tier,
which only the suite's section `# 3.` reachability check uses.

## 2. Live-observed SKIP → PASS

### A. Baseline (no `SMOKE_TEST_SESSION_COOKIE`) — the pre-T04 production behavior

```
PASS: oauth start correctly refused (503) while disabled
SKIP: core-loop happy path (SMOKE_TEST_SESSION_COOKIE not set - pass vocanova_session=<value> to run it)

SMOKE TEST PASSED
exit=0
```

The suite passes overall while the core-loop journey is never exercised. This is
exactly the gap `VOC-050-AC-05` exists to close.

### B. Mint through `T01`'s real endpoint (what the new workflow step does)

```
$ curl -fsS -X POST -H "Authorization: Bearer <token>" -H "Accept: application/json" \
    http://127.0.0.1:18080/ops/synthetic-smoke-test/session
status: issued
session_cookie length: 44
expires_at: 2026-09-08T07:43:59Z
```

### C. Previous revision's value shape (bare opaque token) — fails, as review predicted

```
FAIL: GET /api/v1/me returned 401 with smoke-test session (expected 200)
FAIL: GET /api/v1/journey-situations returned 401 with smoke-test session (expected 200)
SMOKE TEST FAILED: 2 check(s) failed
exit=1
```

### D. This revision's value shape (`vocanova_session=<opaque token>`) — passes

```
PASS: /healthz reports status=ok database=ok
PASS: kill switch magic_link_enabled=false (expected)
PASS: kill switch oauth_enabled=false (expected)
PASS: kill switch new_signups_enabled=false (expected)
PASS: kill switch ai_enabled=true (expected)
PASS: GET http://127.0.0.1:18081/ returned 200
PASS: magic-link request correctly refused (503) while disabled
PASS: oauth start correctly refused (503) while disabled
PASS: GET /api/v1/me returned 200 with smoke-test session
PASS: GET /api/v1/journey-situations returned 200 with smoke-test session

SMOKE TEST PASSED
exit=0
```

Both `# 5.` checks print `PASS:` against the real API authenticated as the real
seeded synthetic account, and the overall exit code is a genuine pass rather
than a masked skip. Case C additionally proves the section is not passing
vacuously: a wrong credential produces `FAIL` and a non-zero exit.

`GET /api/v1/journey-situations` returns 200 for this account against a database
holding only the migration set and the synthetic-user seed, so the
implementation-plan's "re-verify the script's expectations against the synthetic
account's actual seeded data" step found no mismatch and the script's checks
needed no change.

## 3. No state-mutating request added

Section `# 5.` is unchanged apart from comments: two `curl` calls, both `GET`
(`/api/v1/me`, `/api/v1/journey-situations`). The new workflow mint step's only
request is `POST /ops/synthetic-smoke-test/session`, which is not part of the
smoke suite and writes only the synthetic account's own session row — it never
sends a magic-link email and never completes an OAuth flow. The suite's
documented no-side-effects design is preserved.

## 4. Regression check on the smoke script's own harness

```
$ bash infra/scripts/smoke-test-production.selftest.sh
SELFTEST OK: healthy target passes with default expectations
SELFTEST OK: unhealthy database fails the suite
SELFTEST OK: kill-switch mismatch fails the suite
SELFTEST OK: matching non-default expectations pass
SELFTEST OK: core-loop check passes with a valid smoke-test cookie
ALL SELFTEST CASES PASSED
```

Case 5 of that harness already encoded the `vocanova_session=smoke-test-token`
contract; this revision's workflow now produces a value of that exact shape.

## 5. Fail-closed wiring, confirmed by parsing the workflow

```
$ python3 -c "import yaml; d=yaml.safe_load(open('.github/workflows/deploy-production.yml')); \
    print('\n'.join(s.get('name','?') + ('  [continue-on-error]' if s.get('continue-on-error') else '') \
    for s in d['jobs']['deploy']['steps']))"
...
Sync production smoke-test session mint token from GitHub secrets
...
Poll production API health endpoint
Poll production web endpoint
Mint synthetic smoke-test session for the production smoke suite
Run production smoke-test suite
```

No step carries `continue-on-error`. The mint step exits non-zero when its
secret is missing or the endpoint is unreachable, rather than degrading to the
old SKIP.

## 6. Operator prerequisite (not a waiver)

`PRODUCTION_SMOKE_TEST_SESSION_MINT_TOKEN` must exist as a GitHub secret for the
`production` environment before this revision deploys. The new sync step writes
it into the api container's `SMOKE_TEST_SESSION_MINT_TOKEN`, so the two sides
cannot drift; if the secret is empty, that step fails the deploy **before**
`docker compose up -d` runs, leaving the running containers untouched. This is
deliberate: an unset token means `T01` does not register the mint route at all,
which would silently restore the SKIP that `VOC-050-AC-05` removes.

## 7. Limits

- This is a disposable production-shaped rehearsal, not a run against real
  production infrastructure. An agent has no production credentials or deploy
  authority (`AGENTS.md`, "Safety"), so the first real `deploy-production.yml`
  execution of these steps necessarily happens after merge and promotion.
- The web tier in the rehearsal was a static file server. It affects only the
  suite's section `# 3.` reachability check, not the core-loop section this task
  activates.
- Cross-repo release gating remains the open item recorded in `t03-evidence.md`;
  nothing in this task changes it.
