# VOC-039 — middleware.ts's auth check silently fails in production because Next.js Middleware runs on the Edge runtime, not Node

**Status: proposed, not adopted.** Nothing in this package is implementation-authorized.
It is a draft response to [issue #297](https://github.com/KARSIFT/vocanova-platform-sandbox/issues/297),
prepared for founder/steward review at adoption time.

## Why this exists

Issue #297 reports that VOC-038-T03 (validate the non-AI core loop with the real
allowlisted cohort) is blocked in production: Google OAuth sign-in completes and a
valid, correctly-sent session cookie exists, but `apps/web/src/middleware.ts`'s own
`fetch("${apiBaseURL}/api/v1/me", ...)` auth check still treats the request as
unauthenticated and redirects back to `/signin`. Two earlier, correctly-diagnosed
fixes in the same investigation (the `SameSite=Strict` cookie fix and the
`API_BASE_URL` internal-Docker-address fix, both already merged — see PR #296 and
this repo's `git log`) did not resolve it. The issue's own direct reproduction (SSH
into the production web container, calling the same URL with the same cookie from
plain `node -e "..."` vs. from `middleware.ts` itself) isolates the remaining cause
to Next.js Middleware's restricted Edge runtime sandbox, whose `fetch()` behaves
differently from Node's `fetch()` even in this self-hosted (non-Vercel) deployment.

## What this package deliberately does NOT do

- It does not declare the root cause fixed by assertion. `VOC-039-T00`'s acceptance
  criterion requires the same direct-reproduction technique the issue itself used
  (real session cookie, real production API, both Edge-runtime and Node-runtime
  fetch paths compared) to be re-run against the fix before it is accepted — a
  passing regression test alone is not sufficient evidence, because the entire
  reason this took three deploy cycles to diagnose is that the failure only
  reproduced in the real Edge runtime, not in ordinary Jest/Vitest unit tests that
  run under Node.
- It does not adopt itself. `change.yaml` leaves every adoption/authorization field
  at its template default. No task in `tasks.md` may be dispatched until a real
  adoption decision is recorded.
- It does not weaken the `/api/v1/me` auth check itself, only the runtime it
  executes under. The 401/non-ok/redirect logic in `apps/web/src/middleware.ts`
  is preserved as-is; this package's structured-logging task (`VOC-039-T02`) only
  adds observability around the existing failure branches, it does not change
  what those branches decide.

## Open question flagged for the reviewing human

`specification.md`'s "Open questions" section flags one thing this package cannot
resolve on its own: whether `runtime = "nodejs"` middleware has any interaction with
this repo's specific self-hosted deployment topology (e.g. the nginx/Docker network
path between the `web` and `api` containers) that differs from the plain-Node
reproduction in the issue, given that the reproduction was run *inside* the container
directly rather than through the actual middleware execution path with the fix
applied. `VOC-039-T00`'s acceptance criterion requires this to be checked with a real
staging (or production, if staging cannot exercise OAuth) deploy before the fix is
considered proven, not merely deployed.

## Structure

Mirrors recent packages' convention (e.g. VOC-036, VOC-038):
`specification.md`, `acceptance-criteria.md`, `impact-analysis.md`,
`implementation-plan.md`, `tasks.md`, `test-plan.md`, `release-plan.md`.

## Recommended next action for the reviewing human

1. Read `specification.md`'s open question and either accept the proposed
   verification approach in `VOC-039-T00` or adjust it.
2. Confirm the proposed `R3` risk classification in `change.yaml` (this touches the
   production authentication gate for every protected route).
3. Adopt (or request changes to) this package, then dispatch `VOC-039-T00` through
   `VOC-039-T02` individually, as prior packages' tasks were dispatched one at a time.
