# VOC-037-EV-03 — Kill switches and rollback verification against production (T03)

## Scope and why this is founder-gate-delegate-authored

The implementer role correctly refused this task (recorded on issue #263):
`VOC-037-TEST-03` requires live toggling of production kill switches and a
real redeploy rehearsal, which no implementer sandbox has access to perform.
This evidence was produced directly by the founder-gate delegate against the
real production target (`130.185.123.152`), the same pattern used for
`VOC-037-T06`'s evidence and VOC-032-T09/T10 in R1.

## Standing of `VOC-037-AC-03`

**Still NOT satisfied, but materially advanced (2026-08-02)**, recorded
honestly rather than as a partial pass either direction. What's now
resolved since the original 2026-08-01 revision of this document:

- **Both previously-disclosed OAuth-blocking bugs are fixed and verified
  live**: the OAuth state cookie's cross-subdomain scoping (PR #276), and
  a much larger finding - the API had no CORS support at all, blocking
  every credentialed cross-origin browser request project-wide, not just
  OAuth (PR #283, independently reviewed). `GOOGLE_OAUTH_ENABLED` is now
  verified with real Google credentials producing a genuine authorization
  URL from a real browser-originated preflight.
- **A genuine two-different-artifact rollback is now proven**, not just
  the recreate/health-check mechanism in isolation - a real artifact swap,
  confirmed by an observable behavior difference (CORS present vs absent),
  rolled back and forward again, both health-checked.

What's still open:

- `NEW_USER_SIGNUP_ENABLED` was never toggled `true` or verified at all -
  completing this requires an actual human clicking through Google's
  consent screen (or real magic-link delivery once an email provider is
  configured), neither of which this founder-gate delegate can do alone.
- `AI_FEATURES_ENABLED` is still only verified at the startup-log level,
  not the documented HTTP surface (`/api/v1/sentence-feedback`), which
  needs an authenticated session - same blocker as above.
- The redeploy mechanism's `pull`/registry-auth half is still only
  confirmed via the real `deploy-production` workflow's own success, not
  independently re-demonstrated in a manual rehearsal.

Closing `VOC-037-AC-03` requires completing the two remaining
human-in-the-loop gaps or an explicit founder-accepted waiver of them;
this record does neither on its own.

## Kill switch verification (2026-08-01, live against production)

### `EMAIL_MAGIC_LINK_ENABLED`

| State | Request | Result |
| --- | --- | --- |
| `false` (baseline) | `POST /api/v1/auth/magic-links {"email":"..."}` | `503` "Magic link sign-in is disabled" |
| `true` | same request | `204` (request accepted) |

Toggle verified to produce the documented behavior change. Full delivery
(does the learner actually receive the email) is separately blocked on real
outbound email-provider credentials (production-tier equivalent of
`VOC-032-DEP-07`, still open) - `email.Fake` is wired in either state until
those exist, so this verifies the kill-switch gate, not delivery.

### `GOOGLE_OAUTH_ENABLED`

| State | Request | Result |
| --- | --- | --- |
| `false` (baseline) | `POST /api/v1/auth/oauth/google/start` | `503` "Google OAuth sign-in is disabled" |
| `true` | same request | `200`, real `state` token issued, `Set-Cookie: vocanova_oauth_state=...` |
| `false` | `GET /api/v1/auth/oauth/google/callback?code=x&state=y` | `503` |

Toggle verified to produce the documented behavior change on both the start
and callback endpoints.

**Bug found while attempting a full round-trip (not a kill-switch defect),
fixed the same day (PR #276):** `app/api/production.go` originally wired
the OAuth CSRF-state cookie's `Domain` to `cfg.SessionDomain`
(`SESSION_COOKIE_DOMAIN`, e.g. `production.vocanova.site` - the **web**
app's hostname). But `OAuthStart`/`OAuthCallback` are served from the
**API** host (`api-production.vocanova.site`), a sibling subdomain, not a
child of the web hostname, so the cookie was never sent back by a real
browser. Fixed via a dedicated `OAuthStateDomain` field (host-only,
independent of `SessionDomain`) - confirmed fixed live, later the same
session, via the `Set-Cookie` header genuinely carrying no `Domain`
attribute at all.

**Second, larger bug found 2026-08-02 while re-verifying `GOOGLE_OAUTH_ENABLED`
with real Google credentials for the first time:** clicking "Sign in with
Google" on the real deployed production web app failed client-side with
"Unable to start Google sign-in", even though the cookie-domain fix above
was already live. Root cause: the API had **no CORS support at all**,
despite `docs/engineering/06-backend-design.md` §16 explicitly requiring a
"CORS allowlist" as baseline security posture. Since the API and web app
are sibling subdomains, every credentialed cross-origin browser request
(not just OAuth - magic-link, sentence submission, review submission,
anything with a JSON body) has always failed its CORS preflight with no
`Access-Control-Allow-Origin` header, in both staging and production, for
the whole lifetime of the project. Verified directly: the browser's
`OPTIONS` preflight for `/api/v1/auth/oauth/google/start` received a bare
`405` with no CORS headers.

**Fixed same day** (PR #283, independently reviewed - PASS WITH
NON-BLOCKING FINDINGS, including an executable spoofing-attempt harness
against the real merged code) by adding a CORS allowlist middleware
(`apps/api/app/api/cors.go`), deployed to production and confirmed live:

```
$ curl -i -X OPTIONS https://api-production.vocanova.site:8443/api/v1/auth/oauth/google/start \
    -H "Origin: https://production.vocanova.site:8443" \
    -H "Access-Control-Request-Method: POST"
HTTP/2 204
access-control-allow-credentials: true
access-control-allow-headers: Content-Type, Authorization
access-control-allow-methods: GET, POST, PUT, PATCH, DELETE, OPTIONS
access-control-allow-origin: https://production.vocanova.site:8443
```

With both bugs fixed, `POST /api/v1/auth/oauth/google/start` from a real
browser now returns a genuine Google authorization URL built from the real
`GOOGLE_OAUTH_CLIENT_ID`/`GOOGLE_OAUTH_CLIENT_SECRET` (founder-provisioned
2026-08-02), with a correctly host-scoped state cookie. The remaining gap
to a fully-verified end-to-end sign-in is completing the actual Google
consent screen in a real browser, which requires a human - see
`NEW_USER_SIGNUP_ENABLED` below for what that would additionally verify.

### `NEW_USER_SIGNUP_ENABLED`

Gated inside `ConsumeMagicLink`/OAuth-callback for a never-before-seen
identity only (`auth/killswitches.go`). Completing a full sign-in to reach
this code path requires either real magic-link delivery or a working OAuth
round-trip - both are currently blocked (email: real provider credentials
don't exist yet; OAuth: the cookie-domain bug above). Verified instead by:

- Code citation confirming the gate exists and is wired
  (`auth/killswitches.go`'s `NewSignupsEnabled` / `ErrSignupsDisabled`).
- The production startup log line reflecting the live env value, toggled
  and re-observed: `signups=off` at every point in this rehearsal (the
  founder-decided safe default; never toggled `true` in production, since
  doing so with no way to complete a full sign-in would not have produced
  additional evidence).

**Not independently HTTP-verified end-to-end** - recorded honestly as a
partial result, not a fabricated pass, per this session's evidence
standard.

### `AI_FEATURES_ENABLED`

| State | Startup log |
| --- | --- |
| `true` (baseline) | `api: listening on :8080 (env=production, ai=on, magic=off, oauth=off, signups=off)` |
| `false` | `api: listening on :8080 (env=production, ai=off, magic=off, oauth=off, signups=off)` |
| `true` (restored) | `ai=on` confirmed again |

Toggle verified via the documented startup-log signal
(`cmd/api/main.go`'s `boolFlag` line) and code citation
(`aifeedback.NewDisabledGate()` vs `AlwaysEnabledGate()` in
`app/api/production.go`). HTTP-level verification via
`/api/v1/sentence-feedback` requires an authenticated learner session,
which is blocked by the same email/OAuth limitations above - not
independently HTTP-verified end-to-end, recorded honestly.

## Rollback-by-redeploy rehearsal (2026-08-01)

This is the **first** successful production deploy, so there is no
distinct prior-version artifact to roll back to yet. The redeploy
*mechanism* itself (DOC-11 §3: "redeploy previous known-good artifact") was
exercised in full instead:

1. Recorded the running image digest:
   `ghcr.io/karsift/vocanova-api@sha256:cc88b4e4403d897372777a3c9c75be5cc3dc0fb4be67b9949f144f238aa04aba`
2. Ran a full `docker compose pull` + `up -d --force-recreate` for all four
   services (`postgres`, `api`, `web`, `nginx`) against the same
   `sha-fd5d1f7` tag - the only artifact that exists at this point in the
   project's life.
3. `pull` itself reported `denied` in this manual SSH session because I was
   not logged into `ghcr.io` interactively (the real `deploy-production`
   workflow authenticates before pulling and does not have this problem -
   confirmed by the real deploy run, PR #273's evidence). Compose fell back
   to the already-present local images and proceeded.
4. All four containers recreated and reported healthy.
5. Both public endpoints re-verified immediately after:
   `https://production.vocanova.site:8443/` → `200`;
   `https://api-production.vocanova.site:8443/healthz` → `200`,
   `{"status":"ok","database":"ok"}`.

**Disclosed limitation (resolved below):** this proved the redeploy/
recreate/health-check mechanism works end-to-end, and that a fresh
`docker compose pull` failing does not corrupt the running deployment -
but it did not prove a genuine two-different-versions rollback, since only
one version had ever been deployed to production at that point.

## Genuine two-artifact rollback rehearsal (2026-08-02)

A second real production deploy (the CORS fix, PR #283/#284,
`sha-36e91b3`) made this test possible for the first time - two distinct,
genuinely different artifacts now exist (`sha-2560664`, the pre-CORS-fix
version, and `sha-36e91b3`, current).

1. Recorded the running artifact: `sha-36e91b3`.
2. Rolled back: `docker compose up -d --pull never --force-recreate api web`
   with `PRODUCTION_IMAGE_TAG=sha-2560664`. Both containers recreated,
   reported healthy.
3. **Confirmed the rollback genuinely changed running code, not just a
   label:** re-ran the CORS preflight check against the rolled-back
   version - `405` with no `Access-Control-Allow-Origin` header, exactly
   the pre-fix behavior, proving `sha-2560664` was really running (not a
   cosmetic swap). `/healthz` still `200` throughout.
4. Rolled forward again: `PRODUCTION_IMAGE_TAG=sha-36e91b3`, same
   recreate/health-check cycle.
5. Confirmed restored: `/healthz` → `200`; CORS preflight → `204` with the
   correct `Access-Control-Allow-Origin`/`Allow-Credentials`/`Allow-Methods`
   headers again.

This closes the previously-disclosed limitation: a real, observably
different artifact was deployed, verified to actually be running (not
assumed), and rolled forward again, with health checks passing at every
step and zero downtime beyond the container recreate window.

## Production state restored after this evidence was compiled

`AI_FEATURES_ENABLED=true`, `EMAIL_MAGIC_LINK_ENABLED=false`,
`GOOGLE_OAUTH_ENABLED=true` (real Google credentials, wired 2026-08-02 -
no longer the `false` safe default, since the switch itself and a real
provider round-trip through `OAuthStart` are now both genuinely verified),
`NEW_USER_SIGNUP_ENABLED=false` (still the founder-decided safe default -
never toggled `true` in production, since no signup path has been
completed end-to-end yet to observe it against). Verified via the startup
log and a final health check (`{"status":"ok","database":"ok"}`) after
this evidence was compiled, running artifact `sha-36e91b3`.

## Follow-ups this task surfaced

1. ~~**OAuth state cookie domain bug**~~ - **fixed** (PR #276, 2026-08-01),
   confirmed live via a `Set-Cookie` header with no `Domain` attribute.
2. ~~**No CORS support in the API**~~ - **fixed** (PR #283, 2026-08-02),
   independently reviewed, deployed to production, confirmed live via a
   genuine browser-shaped preflight request returning the correct
   `Access-Control-*` headers.
3. `NEW_USER_SIGNUP_ENABLED` and the HTTP-level half of
   `AI_FEATURES_ENABLED` remain unverified end-to-end. Both bugs above that
   were blocking this are now fixed, so the remaining blocker is purely
   "a human needs to click through a real Google consent screen" (or a
   real email provider needs to be configured for the magic-link path) -
   not a code defect.
4. **Why the CORS gap went uncaught for the project's whole lifetime,
   despite E2E testing:** `apps/web/tests/e2e/mock-api-server.mjs` (lines
   ~585-621) shows the *identical* bug was already diagnosed once before,
   while building the E2E harness - its own comment describes the exact
   same symptom (client-side mutations silently failing with
   `net::ERR_FAILED`, no HTTP response, because the preflight had no
   `Access-Control-Allow-Origin` header) and the exact same fix (echo the
   specific request `Origin`, never a wildcard, since credentialed
   requests forbid combining the two). The fix was applied **only to the
   mock**, to unblock E2E test development, and never propagated to the
   real `apps/api` server that the mock stands in for - so every E2E run
   exercised realistic CORS behavior against a server that only existed in
   tests, while the real deployed API had none. This is a process gap
   worth closing generally, not just for this instance: a mock server's
   security-relevant behavior (auth, CORS, CSRF) drifting from the real
   server it substitutes for is exactly the failure mode this repo's
   existing `scripts/foundation/mock-inventory` governance check already
   polices for other kinds of mock/real divergence - CORS parity between
   `mock-api-server.mjs` and the real API would be a natural, low-effort
   extension of that same check, flagged here as a genuine follow-up but
   not implemented in this task (out of scope for T03).
