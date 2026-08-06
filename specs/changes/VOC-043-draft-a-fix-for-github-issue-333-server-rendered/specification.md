# VOC-043 — Server-Rendered Pages' Own /api/v1/me Check Fails with 401 Immediately After middleware.ts's Identical Check Succeeds: Specification

## Objective and requirement source

Restore real Google sign-in past the `/onboarding` page's own server-side auth
check by unifying `apps/web/src/lib/api-server.ts`'s cookie-forwarding mechanism
with the one `apps/web/src/middleware.ts` already uses successfully, so both
server-side call sites forward the incoming request's session cookie to
`/api/v1/me` identically instead of via two independent implementations that can
silently disagree. Grounded in
[issue #333](https://github.com/KARSIFT/vocanova-platform-sandbox/issues/333) in
full (its only content at drafting time — no comments existed yet), including its
three-times-reproduced nginx evidence (`200`, `200`, `401` for the three `/api/v1/me`
calls in one real sign-in attempt), what it explicitly ruled out (stale/duplicate
cookies, IP rate limiting, port/CORS/reachability), and its own named "likely area
to investigate" and "suggested fix direction". Not yet approved by a founder or
technical steward — see `change.yaml`'s `requirement_approval_status`.

## Scope and non-goals

In scope:
- `apps/web/src/lib/api-server.ts`'s `createServerApiClient()`: replace the
  `cookies()`-based reconstruction (`(await cookies()).toString()`) with the same
  raw-incoming-header forward `apps/web/src/middleware.ts` already uses
  successfully (`request.headers.get("cookie")`), sourced via `next/headers`'s
  `headers()` function inside the Server Component request scope (the
  `next/headers` API available for reading, not mutating, the incoming request's
  raw headers from a Server Component/Route Handler — distinct from `cookies()`,
  which is the API this issue's evidence implicates). This removes the second,
  independently-implemented cookie-forwarding path the issue names as the "likely
  area to investigate", leaving exactly one implementation.
- A regression test exercising `createServerApiClient()`'s custom `fetch` against
  a representative multi-cookie `Cookie` header, asserting the forwarded header
  byte-for-byte matches the raw incoming header — so a future edit that
  reintroduces a reconstructing/re-parsing path is caught deterministically.
- Confirming (by reading, not modifying) that `apps/web/src/middleware.ts`'s own
  raw-header forward is unchanged and remains the reference behavior this fix
  aligns `api-server.ts` to.

Non-goals:
- Any change to `apps/api/business/auth/service.go`'s `ValidateSession`,
  `apps/api/app/api/middleware.go`'s `AuthMiddleware`, or any other API-side
  session-validation code. The issue's own evidence (three real HTTP responses,
  not connection failures; no rate limiting on this path per a direct read of
  `service.go`) points at a client-side (web) cookie-forwarding divergence, not an
  API-side defect — this package does not open that area.
- Empirically re-confirming the exact byte-level divergence between the two
  cookie-forwarding implementations before writing the fix. The issue's own
  "suggested fix direction" explicitly offers instrumentation as one path to
  *identify* the divergence, but also independently names the direct fix ("making
  `api-server.ts` forward cookies the same way `middleware.ts` does") as the likely
  correction regardless of which exact encoding/parsing step is at fault. This
  package takes that direct fix rather than adding temporary production logging of
  cookie-adjacent data, per this package's own risk analysis in
  `impact-analysis.md`. See "Open questions" below for what happens if the
  implementer's own investigation finds this insufficient.
- Any change to `apps/web/src/middleware.ts` itself. Its raw-header forward is the
  reference behavior this fix aligns to, not a file this package edits.
- Any change to `apps/web/src/app/onboarding/page.tsx`,
  `apps/web/src/app/(app)/settings/account/page.tsx`, or any other caller of
  `createServerApiClient()`. They all benefit from the fix without their own call
  sites changing, since the fix is inside `createServerApiClient()` itself.
- A broader audit of every other place `next/headers`'s `cookies()` is used
  elsewhere in `apps/web` (a drafting-time read of the codebase found no other
  call site importing `cookies()` besides `api-server.ts`, but this package does
  not assert that search was exhaustive). See "Open questions" below.

## Risk and protected areas

`apps/web/src/lib/api-server.ts` is the server-side auth-forwarding path every
authenticated server-rendered page in `apps/web` depends on (confirmed callers at
drafting time: `apps/web/src/app/onboarding/page.tsx` and
`apps/web/src/app/(app)/settings/account/page.tsx`). Per
`docs/governance/change-risk-classification.md`, "authentication/authorization" is
named explicitly as an R3-floor category. This package proposes `R3` (see
`change.yaml`); the fix does not itself touch schema migrations, billing, or a new
secret, so no higher class is proposed, but the reviewing human's own judgment
governs this, not this proposal. `scripts/governance/classify-change-risk.sh` has
not been run against a real, task-scoped file list at drafting time — consistent
with how VOC-039/VOC-040/VOC-041/VOC-042 handled this field, that computation
belongs to each task's own implementation PR.

## Decisions, contradictions, security, and privacy

No `VOC-043-D00`-style founder/product decision is defined by this draft — the
fix direction (unify the two cookie-forwarding implementations onto the one that
already works) is the issue's own stated suggestion, not a product judgment call.
If the reviewing human disagrees with taking the direct fix over first adding
temporary diagnostic instrumentation (see "Non-goals" above), they should record
that at adoption time rather than this package silently choosing for them.

No contradiction is recorded for this package — the issue does not contain an
opposing claim about the correct fix, only an open invitation for the implementer
to confirm the divergence empirically before or during the fix.

Security/privacy: this fix changes *how* the session cookie is read and forwarded
server-side (from a parsed-and-reconstructed value to a raw pass-through of the
incoming header), not *who* is authenticated or *what* the API does with the
cookie once received. `apps/web/src/middleware.ts`'s equivalent raw-header forward
already ships today and is the reference this fix aligns to, so this is a
reduction in implementation surface (one cookie-forwarding path instead of two),
not a new one. No new secret, credential, or personal-data field is introduced.
The regression test asserts on a synthetic, non-production cookie value only (see
`test-plan.md`), consistent with this repository's existing
`auth-check-logging.test.ts` convention of using a clearly-fake session value.

## Data, migrations, analytics, and accessibility

None. This package touches only `apps/web/src/lib/api-server.ts` and its test
coverage; no schema, migration, analytics event, or accessibility surface is
affected.

## Open questions

1. **Whether the raw-header-forward fix alone actually resolves the 401**, versus
   the true divergence being something more specific this package's drafting-time
   read did not surface (the issue is explicit this is offered as a "likely fix",
   "for the implementer to confirm empirically, not asserted here"). If the
   implementer applies this fix and the issue's exact reproduction (a real Google
   sign-in against production, or the closest available lower-environment
   equivalent) still shows the third `/api/v1/me` call failing, that is new
   information this package's own scope does not anticipate, and the task should
   stop and add temporary redacted-value instrumentation (length/prefix only, per
   `impact-analysis.md`'s security note) rather than guessing at a second fix
   without evidence.
2. **Whether any other file in `apps/web` besides `api-server.ts` independently
   reconstructs the `Cookie` header from `next/headers`'s `cookies()` API**, which
   would carry the same class of risk this issue surfaced. This package's
   drafting-time read found none, but that read was not exhaustive tooling (a
   targeted grep, not a full audit) — flagged for the reviewing human to decide
   whether a broader audit is warranted as a follow-up, similar to how
   VOC-042's open question 1 flagged its own recurring bug class rather than
   silently expanding scope to audit it here.
3. **Whether production or staging currently has any real authenticated user
   stuck mid-onboarding because of this defect**, and whether any operational
   remediation (e.g. contacting an affected user) is needed independent of the
   code fix. This package has no production access to determine this — flagged
   for the reviewing human.
