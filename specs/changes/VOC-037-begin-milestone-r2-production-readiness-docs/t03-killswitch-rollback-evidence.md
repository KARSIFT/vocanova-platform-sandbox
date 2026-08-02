# VOC-037-EV-03 — Kill switches and rollback verification against production (T03)

## Scope and why this is founder-gate-delegate-authored

The implementer role correctly refused this task (recorded on issue #263):
`VOC-037-TEST-03` requires live toggling of production kill switches and a
real redeploy rehearsal, which no implementer sandbox has access to perform.
This evidence was produced directly by the founder-gate delegate against the
real production target (`130.185.123.152`), the same pattern used for
`VOC-037-T06`'s evidence and VOC-032-T09/T10 in R1.

## Standing of `VOC-037-AC-03`

**SATISFIED (2026-08-02).** All four documented kill switches and a
genuine rollback are now independently verified against the real
production target, at the actual documented HTTP/behavioral surface, not
proxies for it:

- `EMAIL_MAGIC_LINK_ENABLED` - `503` disabled / `204` accepted (2026-08-01).
- `GOOGLE_OAUTH_ENABLED` - `503` disabled / `200` real Google authorization
  URL with real credentials (2026-08-01, re-verified with real credentials
  2026-08-02 after two blocking bugs, below, were found and fixed).
- `NEW_USER_SIGNUP_ENABLED` - `503` disabled / `200` real new user +
  session created, via a genuine `ConsumeMagicLink` call (2026-08-02).
- `AI_FEATURES_ENABLED` - `200` documented disabled response /
  `404` past-the-gate eligibility check, via a genuine authenticated
  `SubmitSentenceFeedback` call (2026-08-02).
- Rollback: a real, observably-different second artifact deployed, rolled
  back, confirmed genuinely running (not a label swap), rolled forward
  again - all four containers healthy throughout, no data loss (the
  rollback never touched the database) (2026-08-02).

Two real, previously-undiscovered bugs were found and fixed along the way
rather than worked around or hidden - an OAuth state-cookie cross-subdomain
scoping defect (PR #276) and a much larger finding, the API having no CORS
support at all, which blocked every credentialed cross-origin browser
request project-wide, not just OAuth (PR #283, independently reviewed).
Both are confirmed fixed live in production.

`NEW_USER_SIGNUP_ENABLED` and `AI_FEATURES_ENABLED`'s HTTP-level
verification did not require the human-in-the-loop step (a real Google
consent screen, or real email delivery) originally expected to block them:
constructing a disposable session/magic-link row directly against the real
database, matching exactly what the real service code itself produces,
exercises the identical `ConsumeMagicLink`/`ValidateSession` logic a real
sign-in would - only the outbound email/OAuth-provider round-trip is
bypassed, and that round-trip is what `EMAIL_MAGIC_LINK_ENABLED`/
`GOOGLE_OAUTH_ENABLED` already verify separately. All test data (disposable
users, sessions, magic links) was deleted immediately after each check and
verified gone by re-query and by the test session no longer authenticating.

**One remaining, disclosed limitation, not blocking this AC's observable
outcome:** the redeploy mechanism's `pull`/registry-auth half is confirmed
only via the real `deploy-production` workflow's own repeated successes
(it authenticates to `ghcr.io` before pulling, and every real deploy this
milestone has completed that step correctly), not independently
re-demonstrated in a manual SSH rehearsal - manually authenticating to the
registry from an interactive session was out of scope for this
verification and isn't part of AC-03's documented observable outcome.

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

**HTTP-level verification completed (2026-08-02), fully end-to-end, both
directions.** Gated inside `ConsumeMagicLink`/OAuth-callback for a
never-before-seen identity only (`auth/killswitches.go`). Rather than
waiting on real email delivery, two disposable `magic_links` rows were
inserted directly against the real production database - `token_hash` =
`SHA-256` of a locally-generated raw token, `environment='production'`,
matching exactly what `RequestMagicLink`'s `generateToken()` +
`CreateMagicLink` would have produced. This tests the real
`ConsumeMagicLink` service logic genuinely (not a mock) - only the
outbound email *delivery* step is bypassed, which `EMAIL_MAGIC_LINK_ENABLED`
already covers separately above.

| State | Request (real magic-link token, genuinely new email) | Result |
| --- | --- | --- |
| `false` (baseline) | `POST /api/v1/auth/magic-links/consume` | `503` `"new sign-ups are disabled"` |
| `true` | same request, different never-before-seen email | `200`, real user created (`email_verified_at` set), real session + CSRF cookies issued with correct `Domain=vocanova.site` scoping |

Both real HTTP responses observed directly. All test artifacts (2 disposable
`magic_links` rows, the 1 real user genuinely created by the `true` case,
its session) deleted immediately after; deletion verified by re-query
(`0` rows matching the test email prefix) and by the created session no
longer authenticating (`GET /api/v1/me` → `401`). Both switches restored
to `false` afterward, confirmed via a final live probe
(`POST /api/v1/auth/magic-links` → `503` again).

### `AI_FEATURES_ENABLED`

| State | Startup log |
| --- | --- |
| `true` (baseline) | `api: listening on :8080 (env=production, ai=on, magic=off, oauth=off, signups=off)` |
| `false` | `api: listening on :8080 (env=production, ai=off, magic=off, oauth=off, signups=off)` |
| `true` (restored) | `ai=on` confirmed again |

Toggle verified via the documented startup-log signal
(`cmd/api/main.go`'s `boolFlag` line) and code citation
(`aifeedback.NewDisabledGate()` vs `AlwaysEnabledGate()` in
`app/api/production.go`).

**HTTP-level verification completed (2026-08-02).** The blocker recorded
above (no working authenticated session) is resolved the same way as
`NEW_USER_SIGNUP_ENABLED`: a disposable session was constructed directly
against the real production database, matching exactly what
`auth.Service.ValidateSession` expects (`sessions.token_hash` = `SHA-256`
of the raw session-token bytes, per `apps/api/business/auth/tokens.go`'s
`generateToken`) - deleted immediately after, with revocation confirmed
live (`GET /api/v1/me` with the same token → `401` post-cleanup).

`aifeedback.Service.SubmitSentenceFeedback` checks `Gate.Check()` *before*
loading the attempt target, so the two states are distinguishable via
HTTP status/body alone, without needing a real saved word:

| State | Real request result |
| --- | --- |
| `true` | `404` `"target not found"` - passed the AI gate, reached (and correctly failed) the attempt-eligibility check |
| `false` | `200` `{"errorCode":"AI_FEEDBACK_GENERATION_DISABLED", "canRetry":true}` - the documented stable, non-crash disabled response |

Both states verified live via `POST /api/v1/sentence-feedback` with a
real session cookie, a matching double-submit CSRF header/cookie pair,
and a real `Idempotency-Key`, against the real production host. Restored
to `true` afterward and confirmed via `/healthz`.

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

## Production state as of the final verification pass (2026-08-02)

`AI_FEATURES_ENABLED=true`, `EMAIL_MAGIC_LINK_ENABLED=false`,
`GOOGLE_OAUTH_ENABLED=false`, `NEW_USER_SIGNUP_ENABLED=false` - all four
back at the founder-decided safe launch defaults.

**Note on `GOOGLE_OAUTH_ENABLED` specifically:** an intermediate revision
of this document (superseded here) recorded it as left `true` after the
real-credential verification above. That was accurate at the moment it
was written, but `deploy-production.yml`'s "Write production application
configuration" step hardcodes all four switches back to their safe
defaults (`GOOGLE_OAUTH_ENABLED=false` included) on **every** deploy, by
design - and a production deploy (the CORS fix, PR #283/#284) ran after
that verification and before this final check, silently resetting it.
This is correct, intended behavior (kill switches should never survive a
deploy in an accidentally-enabled state), not a bug - but it means the
`GOOGLE_OAUTH_ENABLED=true` verification earlier in this document is a
point-in-time proof that the switch and the real credentials work, not a
claim about the current running state. The real `GOOGLE_OAUTH_CLIENT_ID`/
`GOOGLE_OAUTH_CLIENT_SECRET` remain configured and untouched by this reset
- only the boolean enable flag reverts - so re-enabling for a real launch
is a one-line config change plus redeploy, not a re-provisioning.

Verified via the startup log and a final health check
(`{"status":"ok","database":"ok"}`), running artifact `sha-36e91b3`.

## Follow-ups this task surfaced

1. ~~**OAuth state cookie domain bug**~~ - **fixed** (PR #276, 2026-08-01),
   confirmed live via a `Set-Cookie` header with no `Domain` attribute.
2. ~~**No CORS support in the API**~~ - **fixed** (PR #283, 2026-08-02),
   independently reviewed, deployed to production, confirmed live via a
   genuine browser-shaped preflight request returning the correct
   `Access-Control-*` headers.
3. ~~`NEW_USER_SIGNUP_ENABLED` and the HTTP-level half of
   `AI_FEATURES_ENABLED` remain unverified end-to-end~~ - **both verified**
   (2026-08-02), via disposable database rows constructed to match exactly
   what the real service code produces, rather than waiting on a human
   completing a real Google consent screen or real email delivery. See the
   per-switch sections above for the full evidence.
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
