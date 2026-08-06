# VOC-043 — Acceptance Criteria

## VOC-043-AC-00 — api-server.ts forwards the raw incoming Cookie header, not a reconstructed one

- Requirement source: issue #333's "likely area to investigate" and "suggested fix
  direction"
- Tasks: `VOC-043-T00`
- Tests: `VOC-043-TEST-00`
- Evidence: `VOC-043-EV-00`
- Result: pending
- Observable outcome: `apps/web/src/lib/api-server.ts`'s `createServerApiClient()`
  no longer imports or calls `cookies()` from `next/headers` to build the `Cookie`
  header it forwards. Instead, it forwards the exact raw incoming `Cookie` header
  value (via `next/headers`'s `headers()` API), matching
  `apps/web/src/middleware.ts`'s own `request.headers.get("cookie")` forward.
  Verified by reading the post-fix file directly and by
  `VOC-043-TEST-00`'s assertion that the header value forwarded to the injected
  test `fetch` is byte-for-byte identical to a representative multi-cookie raw
  input header.

## VOC-043-AC-01 — Every existing caller of createServerApiClient() still authenticates correctly

- Requirement source: issue #333's implicit ask (the fix must not break the
  already-working callers while fixing the failing case)
- Tasks: `VOC-043-T00`
- Tests: `VOC-043-TEST-01`
- Evidence: `VOC-043-EV-01`
- Result: pending
- Observable outcome: `apps/web/src/app/onboarding/page.tsx` and
  `apps/web/src/app/(app)/settings/account/page.tsx` (the two confirmed callers of
  `createServerApiClient()` at drafting time) are unchanged by this package's diff
  and continue to compile and behave identically for any request that already
  carries a valid `Cookie` header, absence of a cookie, or an API-returned `401`
  (redirect to `/signin` via `requireAuthRedirect`, unchanged).

## VOC-043-AC-02 — A deterministic regression check catches a reintroduced cookie-reconstruction path

- Requirement source: issue #333's implicit ask (a defect this subtle — two
  implementations that look "identical" but silently disagree — should be
  catchable without a live production reproduction)
- Tasks: `VOC-043-T01`
- Tests: `VOC-043-TEST-02`
- Evidence: `VOC-043-EV-02`
- Result: pending
- Observable outcome: an automated test exists (per `apps/web`'s existing
  `node:test`-based convention, e.g. alongside
  `apps/web/tests/middleware/auth-check-logging.test.ts`) that fails if
  `createServerApiClient()`'s forwarded `Cookie` header ever again diverges from
  the raw incoming header — for example if a future edit reintroduces a
  `cookies()`-based reconstruction, drops a cookie, or re-encodes a value. This
  check runs via `pnpm validate` or a narrower documented script, per this
  repository's existing CI convention, not only manually.
