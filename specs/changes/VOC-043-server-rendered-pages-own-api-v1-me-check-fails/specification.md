# VOC-043 — Server-Rendered Pages' Own /api/v1/me Check Fails with 401 Immediately After middleware.ts's Identical Check Succeeds, Blocking Real Onboarding After Google Sign-In: Specification

## Objective and requirement source

Restore real Google sign-in past the onboarding page's own server-side auth check
by resolving why `apps/web/src/app/onboarding/page.tsx`'s call to
`createServerApiClient().getCurrentUser()` (via `apps/web/src/lib/api-server.ts`)
receives `401` from `GET /api/v1/me` moments after `apps/web/src/middleware.ts`'s
own, textually near-identical call to the same endpoint, in the same request
lifecycle, already succeeded with `200`. Grounded in
[issue #333](https://github.com/KARSIFT/vocanova-platform-sandbox/issues/333) in
full, including its three-times-reproduced nginx access-log evidence (identical
`200, 200, 401` pattern even after fully clearing browser cookies), its explicit
ruling-out of stale/duplicate cookies and IP-based rate limiting with cited code
references, and its named (not asserted) candidate divergence between
`middleware.ts`'s raw `request.headers.get("cookie")` forward and
`api-server.ts`'s `next/headers` `cookies()`-based reconstruction. Not yet
approved by a founder or technical steward — see `change.yaml`'s
`requirement_approval_status`.

## Scope and non-goals

In scope:
- Empirically identifying the exact divergence between the two cookie-forwarding
  implementations for a single real request, per the issue's own suggested
  approach: instrument `apps/web/src/middleware.ts`'s `cookieHeader` and
  `apps/web/src/lib/api-server.ts`'s `cookieHeader` (or, at minimum, each one's
  length and the specific session cookie's value length, redacting the value
  itself) for one real sign-in attempt, and compare.
- Making the fix the instrumentation identifies — the issue's own suggested
  direction is to make `apps/web/src/lib/api-server.ts` forward cookies the same
  way `apps/web/src/middleware.ts` does (or vice versa), so exactly one
  cookie-forwarding implementation exists for this logical operation, not two that
  can silently disagree. This package does not commit to which direction in
  advance of the instrumentation finding — see "Open questions" below.
- A deterministic regression test that exercises `createServerApiClient()`'s
  cookie-forwarding behavior against a request carrying the same cookie shape
  `middleware.ts` already handles correctly, so this specific divergence (once
  identified and fixed) cannot silently regress.
- Confirming whether any other server component beyond
  `apps/web/src/app/onboarding/page.tsx` also calls `createServerApiClient()` and
  is therefore exposed to the same defect — a repository-wide grep for
  `createServerApiClient` call sites, not a live-production check.

Non-goals:
- Any change to `apps/api/business/auth/service.go`'s `ValidateSession`,
  `apps/api/app/api/middleware.go`'s `AuthMiddleware`, or the session-cookie
  issuance logic in `apps/api/business/auth/tokens.go`. The issue's own evidence
  (three identical reproductions, IP-rate-limiting and stale-cookie causes both
  ruled out with cited code) points at the two *client-side* (Next.js) cookie-
  forwarding implementations, not at the API's session-validation logic itself —
  unless the instrumentation step (in scope above) finds otherwise, in which case
  this non-goal must be revisited before implementation, not silently overridden.
- Any change to the OAuth callback flow itself (`apps/api/business/auth/service.go`'s
  `OAuthCallback` or the redirect chain up to and including the successful `302`).
  The issue's evidence shows the callback and both `middleware.ts` checks already
  succeed; only the third, page-level check fails.
- A broader audit of every other place this repository forwards cookies
  server-side beyond `middleware.ts` and `api-server.ts`, beyond the
  `createServerApiClient` call-site grep named above as in-scope. If the grep
  finds no other call site, no further audit is performed by this package.
- Removing or weakening `apps/web/src/app/onboarding/page.tsx`'s
  `requireAuthRedirect` 401-handling behavior itself — the page's redirect-on-401
  behavior is correct; the defect is that it receives a spurious 401 for an
  actually-valid session, not that it reacts to a genuine 401 incorrectly.

## Risk and protected areas

The most likely fix touches `apps/web/src/lib/api-server.ts`, which underlies every
server-rendered page's own auth check (today, confirmed via a repository read at
drafting time, at least `apps/web/src/app/onboarding/page.tsx`; the in-scope grep
above will confirm the full call-site list). This is a server-side authentication
boundary distinct from, but with the same real-world consequence as,
`apps/web/src/middleware.ts` — a defect here fully blocks sign-in for real users,
symmetrically with VOC-039/VOC-041/VOC-042's middleware-side defects. This package
proposes `R3` (see `change.yaml`); it does not touch schema migrations, billing, or
a new secret, so no higher class is proposed, but the reviewing human's own
judgment governs this, not this proposal.
`scripts/governance/classify-change-risk.sh` has not been run against a real,
task-scoped file list at drafting time — consistent with how
VOC-039/VOC-040/VOC-041/VOC-042 handled this field; that computation belongs to
each task's own implementation PR.

## Decisions, contradictions, security, and privacy

No `VOC-043-D00`-style founder/product decision is defined by this draft. Which
direction the fix takes (unifying on `middleware.ts`'s raw-header forward,
unifying on `api-server.ts`'s `cookies()`-based approach, or a third shared
helper) is deliberately left open pending the in-scope instrumentation finding —
see "Open questions" below — rather than this package guessing an answer the
issue itself explicitly declines to assert.

No contradiction between sibling documents is recorded for this package.

Security/privacy: the issue's own suggested instrumentation explicitly calls for
redacting the actual cookie/session-token value (logging only length, or a
redacted form) rather than the raw value — this package adopts that same
constraint for any diagnostic logging added during implementation, consistent
with `apps/web/src/middleware.ts`'s own existing `logAuthCheckFailure` comment
("The payload deliberately carries no cookie or credential material"). This fix
does not change *who* can authenticate or *what* the API's `ValidateSession`
considers a valid session — it only changes how a request's existing, real
cookies are forwarded from one internal Next.js call to another server-side
fetch within the same request lifecycle. No new attacker-controlled surface or
personal-data field is introduced.

## Data, migrations, analytics, and accessibility

None. This package touches only server-side cookie-forwarding code in
`apps/web` and its test coverage; no schema, migration, analytics event, or
accessibility surface is affected.

## Open questions

1. **The exact divergence between the two cookie-forwarding implementations is
   not yet confirmed empirically.** The issue names two candidates
   (`middleware.ts`'s raw header forward vs. `api-server.ts`'s `cookies()`-based
   reconstruction) and a suggested diagnostic approach, but explicitly frames the
   suggested fix direction as "for the implementer to confirm empirically, not
   asserted here." This package's `VOC-043-T00` therefore starts with
   instrumentation, not a code change, and the actual fix (`VOC-043-T01`) is
   scoped to whatever that instrumentation finds — flagged for the reviewing
   human that the task list below cannot fully specify the fix's exact diff in
   advance of that finding.
2. **Whether diagnostic logging added during instrumentation should be removed
   before merge or may remain as permanent low-cardinality diagnostics** (e.g. a
   redacted cookie-length comparison, useful for catching a future regression of
   this same class). Flagged for the reviewing human to decide at adoption or
   review time — this package does not decide it in advance.
3. **Whether any server component beyond `apps/web/src/app/onboarding/page.tsx`
   also calls `createServerApiClient()` and is therefore exposed to the same
   defect** is not confirmed at drafting time beyond a single targeted read of
   `onboarding/page.tsx`. The in-scope repository-wide grep (see "Scope and
   non-goals" above) will resolve this during implementation; if it finds
   additional call sites, they are automatically covered by the same fix (since
   the fix corrects the shared `createServerApiClient()` implementation), but
   this package does not assert their existence or absence in advance.
