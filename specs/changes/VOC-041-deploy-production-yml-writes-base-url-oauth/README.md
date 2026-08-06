# VOC-041 — deploy-production.yml Writes BASE_URL/OAUTH_REDIRECT_URI/OAUTH_REDIRECT_ALLOWLIST Without the Required :8443 Port, Breaking Real Google Sign-In via CORS

**Status: proposed, not adopted.** Nothing in this package is implementation-authorized.
It is a draft response to [issue #312](https://github.com/KARSIFT/vocanova-platform-sandbox/issues/312),
prepared for founder/steward review at adoption time.

## Why this exists

Issue #312 reports that VOC-038-T03 (real Google sign-in against production) still
fails after VOC-039's Edge-runtime middleware fix was deployed: clicking "Sign in
with Google" shows the generic "Unable to start Google sign-in. Please try again."
The issue traces this to `.github/workflows/deploy-production.yml`'s "Write
production application configuration" step (lines 256-322), which writes
`BASE_URL`, `OAUTH_REDIRECT_URI`, and `OAUTH_REDIRECT_ALLOWLIST` without the `:8443`
port that production actually needs, because production and staging share one host
and staging already owns the host's default port 443 (VOC-037/T06's D00
supersession note). The step's own comment (lines 289-298) justifies the missing
port by asserting Cloudflare now forwards the plain `:443` hostname to the origin's
`:8443` — but the issue directly reproduces, today, against the live host, that this
is false: a real browser hitting `https://production.vocanova.site/signin` (no
port) gets nginx's `444` catch-all, not the app, and the same workflow's own "Poll
production API health endpoint" step (lines 476-499) already carries a comment
recording the same finding from the first real deploy ("the first real deploy's
poll steps (unqualified :443) silently hit Cloudflare's 520 fallback instead of
production's actual nginx").

Because `apps/api/app/api/production.go`'s `corsAllowedOrigins` derives the API's
CORS allow-list directly from `OAUTH_REDIRECT_ALLOWLIST` (scheme+host, no port), a
real browser loaded from `https://production.vocanova.site:8443` sends an `Origin`
header that includes `:8443`, which does not string-match the unqualified allow-list
entry the workflow writes today. The issue confirms via nginx access logs that the
CORS preflight (`OPTIONS .../oauth/google/start`) gets a `204`, but the real `POST`
that should follow it never does — the browser silently blocks it client-side, which
surfaces to the user as the generic non-`ApiResponseError` fallback in
`apps/web/src/app/signin/_components/auth-forms.tsx`.

## What this package deliberately does NOT do

- It does not touch `SESSION_COOKIE_DOMAIN` — the issue explicitly states cookie
  domain-matching ignores port, and this package's own reading of the same step
  (line 312) confirms that line derives from `PRODUCTION_WEB_HOST` stripped to its
  parent domain, not the full host, so it is unaffected by this bug.
- It does not touch `production_api_base_url`'s workflow-input default (line 20,
  currently `https://api-production.vocanova.site`, no port) even though the same
  CORS/routing failure mode plausibly applies to it too (it is baked into the
  browser bundle as `NEXT_PUBLIC_API_BASE_URL` and used for real browser fetches).
  The issue's suggested fix does not name this input, and its own description
  (line 17) asserts it was "already fixed for the browser-bundle case earlier
  today" — a claim this package's drafting-time read of the actual file
  contradicts (see `specification.md`'s open question 1). This package flags the
  discrepancy rather than silently expanding scope to fix an input the issue
  did not ask about.
- It does not touch the Cloudflare-side configuration itself (this package has no
  Cloudflare access) — it only stops relying on the unverified Cloudflare rule the
  issue disproves, consistent with the issue's own suggested fix.
- It does not adopt itself. `change.yaml` leaves every adoption/authorization field
  at its template default. No task in `tasks.md` may be dispatched until a real
  adoption decision is recorded.

## Open question flagged for the reviewing human

`specification.md`'s "Open questions" section flags that `production_api_base_url`'s
workflow-input default (line 20) is not in this package's stated scope but is very
likely affected by the identical bug class, and that the issue's own claim that this
input was "already fixed" does not match the file's current content at drafting
time. This is flagged rather than silently fixed or silently ignored.

## Structure

Mirrors recent packages' convention (e.g. VOC-039, VOC-040):
`specification.md`, `acceptance-criteria.md`, `impact-analysis.md`,
`implementation-plan.md`, `tasks.md`, `test-plan.md`, `release-plan.md`.

## Recommended next action for the reviewing human

1. Read `specification.md`'s open question 1 and decide whether
   `production_api_base_url`'s default should be fixed in the same package, a
   follow-up package, or left as-is with a documented rationale.
2. Confirm the proposed `R3` risk classification in `change.yaml` (this touches
   CI/CD deploy configuration and the production authentication/CORS boundary).
3. Confirm whether the currently-live production `api.env` (already written with
   the unqualified values, per the issue's direct SSH confirmation) needs a
   one-time manual correction independent of the next scheduled deploy, since this
   fix only changes what a *future* dispatch of `deploy-production.yml` writes, not
   the file already on the host.
4. Adopt (or request changes to) this package, then dispatch `VOC-041-T00` through
   `VOC-041-T02` individually, as prior packages' tasks were dispatched one at a
   time.
