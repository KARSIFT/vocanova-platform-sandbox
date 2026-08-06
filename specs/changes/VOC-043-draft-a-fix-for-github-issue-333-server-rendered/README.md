# VOC-043 — Server-Rendered Pages' Own /api/v1/me Check Fails with 401 Immediately After middleware.ts's Identical Check Succeeds

**Status: proposed, not adopted.** Nothing in this package is
implementation-authorized. It is a draft response to
[issue #333](https://github.com/KARSIFT/vocanova-platform-sandbox/issues/333),
prepared for founder/steward review at adoption time.

## Why this exists

Issue #333 reports that after VOC-039/VOC-041/VOC-042 (all merged, all
port-qualification fixes for the shared-host `:8443` bug class), real Google
sign-in against production now gets meaningfully further: the OAuth callback
succeeds for real, and `apps/web/src/middleware.ts`'s own `/api/v1/me` auth check
succeeds twice (once for `/home`, once for `/onboarding`, both `200`). But the
`/onboarding` page's *own* server-side auth check — a third, separate call to the
same endpoint, moments later in the same request lifecycle — gets `401`, and the
page redirects back to `/signin?returnTo=%2Fonboarding`. Net result: real sign-in
still loops back to the login page, just one step further downstream than
before, reproduced identically three times including after fully clearing
browser cookies.

The issue traces this to two independent implementations of "forward the
incoming request's cookies to the API" for what should be the same logical
operation: `apps/web/src/middleware.ts` reads the raw incoming `Cookie` header
directly (`request.headers.get("cookie")`) and forwards it unmodified — this
succeeds, twice. `apps/web/src/lib/api-server.ts`'s `createServerApiClient()`
(used by `apps/web/src/app/onboarding/page.tsx`'s own `client.getCurrentUser()`
call) instead uses Next.js's `cookies()` API from `next/headers`, then
reconstructs the header via `cookieStore.toString()` — this fails.

The issue itself names the likely fix: make `api-server.ts` forward cookies the
same way `middleware.ts` already does, so there is one cookie-forwarding
implementation instead of two that can silently disagree — while explicitly
noting this is offered as a likely direction "for the implementer to confirm
empirically, not asserted here."

## What this package deliberately does NOT do

- It does not touch `apps/api/business/auth/service.go`'s `ValidateSession` or
  any other API-side session-validation code. The issue's own evidence (three
  real HTTP responses, not connection failures; no rate limiting on this path)
  points at a web-side cookie-forwarding divergence, not an API-side defect.
- It does not add temporary production logging of cookie values to empirically
  confirm the exact byte-level divergence before fixing it. The issue names the
  direct fix as independently likely regardless of the exact encoding/parsing
  step at fault; this package takes that direct fix and flags what happens if it
  turns out insufficient (see `specification.md`'s open question 1) rather than
  adding cookie-adjacent diagnostic logging as a first step.
- It does not touch `apps/web/src/middleware.ts` itself — its raw-header forward
  is the reference behavior this fix aligns to, not a file this package edits.
- It does not touch any page component (`onboarding/page.tsx`,
  `settings/account/page.tsx`, or any other caller of
  `createServerApiClient()`). They all benefit from the fix without their own
  call sites changing.
- It does not adopt itself. `change.yaml` leaves every adoption/authorization
  field at its template default. No task in `tasks.md` may be dispatched until a
  real adoption decision is recorded.

## Open questions flagged for the reviewing human

`specification.md`'s "Open questions" section flags: (1) whether the direct
raw-header-forward fix alone actually resolves the reported `401`, since the
issue offers it as a likely but not empirically pre-confirmed correction; (2)
whether any other file in `apps/web` besides `api-server.ts` independently
reconstructs a `Cookie` header from `cookies()`, carrying the same class of risk;
and (3) whether any real user is currently stuck mid-onboarding in production
because of this defect, needing separate operational remediation this package
has no access to perform.

## Structure

Mirrors recent packages' convention (e.g. VOC-042, VOC-041, VOC-040, VOC-039):
`specification.md`, `acceptance-criteria.md`, `impact-analysis.md`,
`implementation-plan.md`, `tasks.md`, `test-plan.md`, `release-plan.md`.

## Recommended next action for the reviewing human

1. Confirm the proposed `R3` risk classification in `change.yaml` (this touches
   authentication/authorization: the server-side cookie-forwarding path every
   authenticated server-rendered page in `apps/web` depends on).
2. Read `specification.md`'s open questions and decide whether the direct fix
   should proceed as scoped, or whether temporary empirical instrumentation
   should be added first.
3. Decide whether a broader audit of other possible `cookies()`-reconstruction
   call sites, and/or operational remediation for any currently-stuck real
   user, should be commissioned as separate follow-up work.
4. Adopt (or request changes to) this package, then dispatch `VOC-043-T00`
   through `VOC-043-T01` individually, as prior packages' tasks were dispatched
   one at a time.
