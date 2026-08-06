# VOC-042 — infra/docker-compose.production.yml Hardcodes web's API_BASE_URL Without the Required :8443 Port, Still Breaking Real Google Sign-In After VOC-041

**Status: proposed, not adopted.** Nothing in this package is implementation-authorized.
It is a draft response to [issue #319](https://github.com/KARSIFT/vocanova-platform-sandbox/issues/319),
prepared for founder/steward review at adoption time.

## Why this exists

Issue #319 reports that real Google sign-in still fails after VOC-041 deployed —
one step further downstream than before. VOC-041 fixed
`deploy-production.yml`'s `BASE_URL`/`OAUTH_REDIRECT_URI`/`OAUTH_REDIRECT_ALLOWLIST`
(the API's own env, written at deploy time), and the issue confirms via nginx
logs that the Google OAuth callback itself now completes for real (`GET
.../oauth/google/callback?...code=...` -> `302`). But the browser is then bounced
`/home` -> `/onboarding` -> `/signin?returnTo=%2Fonboarding` — the same "back to the
login page" symptom as before, just moved past the callback.

The issue traces this to a *different* file than VOC-041 touched:
`infra/docker-compose.production.yml`'s `web` service (currently lines 82-118)
hardcodes `API_BASE_URL: https://api-production.vocanova.site`, with no port,
directly in the compose file — not templated from any `deploy-production.yml`
workflow input, so VOC-041's fix (and any future dispatch-time override) cannot
reach it. Confirmed live via direct SSH: `docker exec vocanova-production-web env
| grep API_BASE_URL` returns exactly this unqualified value, unchanged by VOC-041.

This env var exists specifically because `apps/web/src/lib/env.ts`'s
`getApiBaseURL()` prefers the server-side `API_BASE_URL` over the client-facing
`NEXT_PUBLIC_API_BASE_URL` when running server-side — and that is exactly the code
path `apps/web/src/middleware.ts` uses for its `/api/v1/me` auth check on every
protected route. Since production's API is only reachable on port `8443`
externally (the same host-sharing constraint already named in VOC-037/T06, VOC-038,
VOC-041, and the health-check-poll `:8443` requirement already documented in
`deploy-production.yml`), middleware's server-side fetch to
`https://api-production.vocanova.site/api/v1/me` (no port) never reaches the real
API, so every protected route's auth check fails and redirects back to `/signin` —
even though the OAuth flow and session cookie are both now working correctly, per
VOC-041.

The file's own existing comment block (lines 93-106, quoted in the issue) already
documents *why* `API_BASE_URL` is set this way — it just never got the port
appended, unlike the sibling API service's now-corrected env values.

## What this package deliberately does NOT do

- It does not touch `deploy-production.yml` or any other workflow input. The issue
  is explicit that this value is hardcoded directly in the compose file, not
  templated from a workflow input, so VOC-041's mechanism (dispatch-time
  overrides) cannot reach it — the fix has to be a direct edit to this file.
- It does not touch the API service's own environment block in this same file
  (`infra/docker-compose.production.yml`'s `api` service), which this package's own
  drafting-time read of the file confirms does not set `API_BASE_URL`/
  `OAUTH_REDIRECT_URI`/`OAUTH_REDIRECT_ALLOWLIST` at all (those come from
  `deploy-production.yml`'s generated `api.env`, per VOC-041) — out of scope, not
  implicated by this issue.
- It does not touch `infra/docker-compose.yml` (the non-production/staging compose
  file) — see "Open questions" in `specification.md` for why this is flagged rather
  than silently checked or silently ignored.
- It does not adopt itself. `change.yaml` leaves every adoption/authorization field
  at its template default. No task in `tasks.md` may be dispatched until a real
  adoption decision is recorded.

## Open question flagged for the reviewing human

`specification.md`'s "Open questions" section flags that this exact bug class
(a hardcoded same-host URL missing the `:8443` port production actually serves on)
has now recurred three times across three different files (VOC-041's
`deploy-production.yml`, this package's `docker-compose.production.yml`, and
VOC-038-T03's original CORS bug) — the reviewing human should decide whether a
broader audit of every same-host URL construction site is warranted as a follow-up,
beyond this package's narrowly-scoped single-line fix.

## Structure

Mirrors recent packages' convention (e.g. VOC-041, VOC-039, VOC-040):
`specification.md`, `acceptance-criteria.md`, `impact-analysis.md`,
`implementation-plan.md`, `tasks.md`, `test-plan.md`, `release-plan.md`.

## Recommended next action for the reviewing human

1. Confirm the proposed `R3` risk classification in `change.yaml` (this touches
   production infrastructure configuration and the server-side auth-check boundary
   `apps/web/src/middleware.ts` relies on).
2. Read `specification.md`'s open questions and decide whether a broader same-host
   URL audit is warranted as a follow-up package.
3. Confirm whether the currently-running production `web` container (already
   started with the unqualified `API_BASE_URL`, per the issue's direct SSH
   confirmation) needs a one-time manual restart/recreate independent of the next
   scheduled deploy, since a compose-file change only takes effect the next time the
   `web` service is recreated (e.g. via `docker compose up -d web` or a full
   redeploy), not merely by editing this file.
4. Adopt (or request changes to) this package, then dispatch `VOC-042-T00` through
   `VOC-042-T01` individually, as prior packages' tasks were dispatched one at a
   time.
