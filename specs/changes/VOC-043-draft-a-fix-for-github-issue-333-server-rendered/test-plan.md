# VOC-043 — Test Plan

## VOC-043-TEST-00 — Forwarded Cookie header matches the raw incoming header byte-for-byte

- Covers: `VOC-043-AC-00`
- Preconditions: `VOC-043-T00`'s fix applied to `apps/web/src/lib/api-server.ts`.
- Procedure:
  1. Construct a representative multi-cookie raw `Cookie` header value (e.g. a
     session cookie plus a CSRF double-submit cookie, matching
     `apps/api/business/auth/tokens.go`'s `SessionCookie`/`CSRFToken` naming),
     using clearly-fake, non-production values (mirroring
     `auth-check-logging.test.ts`'s existing convention).
  2. Set that value as the incoming request's raw `cookie` header in the test
     harness (via `next/headers`'s test-mocking convention, or an equivalent
     request-scoped stub, whichever this repo's existing test stack supports for
     `apps/web`'s Server Component request context).
  3. Call `createServerApiClient()`, then invoke its resulting client's
     `getCurrentUser()` (or inspect the custom `fetch` closure directly).
  4. Confirm the `Cookie` header value actually sent by the injected `fetch` is
     byte-for-byte identical to the value from step 1.
  5. Repeat with no incoming cookie at all; confirm no `Cookie` header is set
     (not a literal `"null"`/`"undefined"` string).
- Expected result: forwarded header matches raw input exactly in both cases.
- Evidence: `VOC-043-EV-00`

## VOC-043-TEST-01 — Existing callers behave identically post-fix

- Covers: `VOC-043-AC-01`
- Preconditions: `VOC-043-T00`'s fix applied; `apps/web/src/app/onboarding/page.tsx`
  and `apps/web/src/app/(app)/settings/account/page.tsx` unchanged by this
  package (only their shared dependency's internals change).
- Procedure:
  1. Confirm both files are byte-for-byte unchanged by this package's diff (a
     `git diff` check on these two paths, expected to show no lines).
  2. With a valid, non-expired session cookie forwarded, confirm
     `createServerApiClient().getCurrentUser()` still resolves successfully
     (200-path behavior unchanged).
  3. With no cookie, or an API-returned `401`, confirm `requireAuthRedirect()`
     still redirects to `/signin?returnTo=...` exactly as before (unchanged
     behavior, since `requireAuthRedirect()` itself is not touched by this
     package).
- Expected result: no caller-visible behavior change for either the success or
  failure path; only the internal cookie-forwarding mechanism differs.
- Evidence: `VOC-043-EV-01`

## VOC-043-TEST-02 — Regression test fails pre-fix, passes post-fix, runs in CI

- Covers: `VOC-043-AC-02`
- Preconditions: `VOC-043-T01`'s test implemented.
- Procedure:
  1. Run the new test against a copy of `api-server.ts` with `VOC-043-T00`'s fix
     temporarily reverted (the `cookies()`-based reconstruction restored) and
     confirm it fails.
  2. Run the new test against the post-fix file and confirm it passes.
  3. Run the test as part of this repo's normal CI invocation (`pnpm validate`
     or the narrower relevant script) and confirm it is picked up.
- Expected result: fails pre-fix, passes post-fix, runs in normal CI.
- Evidence: `VOC-043-EV-02`

This package introduces no migration, so no migration-rollback test is
applicable. No accessibility surface is affected, so no accessibility test is
applicable. This package's fix narrows a currently-fail-safe-but-fully-blocking
defect (denying access to an already-authenticated real user) rather than
widening authorization, so no separate negative/authorization-failure test
beyond `VOC-043-TEST-01` step 3's confirmation that the `401`/no-cookie redirect
path is unchanged is applicable — that step is this package's
authorization-boundary coverage.
