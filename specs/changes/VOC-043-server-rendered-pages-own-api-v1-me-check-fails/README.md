# VOC-043 — Server-Rendered Pages' Own /api/v1/me Check Fails with 401 Immediately After middleware.ts's Identical Check Succeeds, Blocking Real Onboarding After Google Sign-In

**Status: proposed, not adopted.** Nothing in this package is
implementation-authorized. It is a draft response to
[issue #333](https://github.com/KARSIFT/vocanova-platform-sandbox/issues/333),
prepared for founder/steward review at adoption time.

## Why this exists

After VOC-039/VOC-041/VOC-042 (all merged, all port-qualification fixes for
the shared-host `:8443` bug class), real Google sign-in against production now
gets meaningfully further: the OAuth callback succeeds for real, and
`middleware.ts`'s own auth check against both `/home` and `/onboarding`
succeeds (`GET /api/v1/me` -> `200`, twice, per nginx access logs from three
separate real sign-in attempts, reproduced identically even after fully
clearing browser cookies). But `/onboarding`'s own server-side page-level auth
check — a separate call to the same endpoint, moments later — gets `401` and
redirects back to `/signin?returnTo=%2Fonboarding`. Net result: real sign-in
still loops back to the login page, just one step further downstream than
before.

The issue rules out stale/duplicate cookies (retested with browser cookies
fully cleared; identical pattern) and IP-based rate limiting
(`ValidateSession` does not call the rate limiter at all — confirmed by a
direct read of `apps/api/business/auth/service.go`) with cited evidence. It
names two candidate implementations for the same logical operation —
`apps/web/src/middleware.ts`'s raw `request.headers.get("cookie")` forward,
versus `apps/web/src/lib/api-server.ts`'s `next/headers` `cookies()`-based
reconstruction — as the likely area to investigate, but explicitly frames this
as a hypothesis for the implementer to confirm empirically via instrumentation,
not as an asserted root cause.

## What this package deliberately does NOT do

- It does not assert a confirmed root cause or commit to a specific code diff
  in advance of `VOC-043-T00`'s real-request instrumentation finding — the
  issue itself explicitly declines to assert this, and this package follows
  that same discipline rather than guessing past it.
- It does not touch `apps/api/business/auth/service.go`'s `ValidateSession` or
  any other API-side session-validation logic, since the issue's evidence
  points at the two Next.js-side cookie-forwarding implementations, not the
  API's own validation — unless `VOC-043-T00`'s finding indicates otherwise,
  in which case `specification.md`'s stated non-goal must be revisited before
  implementation, not silently overridden.
- It does not adopt itself. `change.yaml` leaves every adoption/authorization
  field at its template default. No task in `tasks.md` may be dispatched
  until a real adoption decision is recorded.

## Open questions flagged for the reviewing human

`specification.md`'s "Open questions" section flags: (1) the root cause is
not yet empirically confirmed, so `VOC-043-T01`'s exact diff is scoped by
`VOC-043-T00`'s finding rather than predetermined here; (2) whether diagnostic
logging added during instrumentation should be temporary or may persist as a
regression signal; and (3) whether any server component beyond
`apps/web/src/app/onboarding/page.tsx` also calls `createServerApiClient()`
and shares this exposure.

## Structure

Mirrors recent packages' convention (e.g. VOC-042, VOC-041, VOC-039, VOC-040):
`specification.md`, `acceptance-criteria.md`, `impact-analysis.md`,
`implementation-plan.md`, `tasks.md`, `test-plan.md`, `release-plan.md`.

## Recommended next action for the reviewing human

1. Confirm the proposed `R3` risk classification in `change.yaml` (this
   touches a server-side authentication boundary every protected
   server-rendered page depends on to actually render for a real user, not
   merely to pass `middleware.ts`).
2. Read `specification.md`'s open questions and confirm the sequencing
   (`VOC-043-T00`'s instrumentation before `VOC-043-T01`'s fix) is acceptable,
   rather than requesting the implementer commit to a specific diff upfront.
3. Decide whether temporary diagnostic logging added during instrumentation
   should be removed before merge or may persist (open question 2).
4. Adopt (or request changes to) this package, then dispatch `VOC-043-T00`
   through `VOC-043-T03` individually, as prior packages' tasks were
   dispatched one at a time.
