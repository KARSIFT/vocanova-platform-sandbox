# VOC-043 — Test Plan

## VOC-043-TEST-00 — Real-request comparison identifies a concrete divergence

- Covers: `VOC-043-AC-00`
- Preconditions: `VOC-043-T00`'s diagnostic logging deployed to a real (or
  realistically constructed) test path.
- Procedure:
  1. Perform one sign-in attempt (real or realistically simulated) that
     reaches `/onboarding` via the middleware redirect chain the issue
     describes.
  2. Capture the redacted diagnostic output from both `middleware.ts`'s and
     `api-server.ts`'s cookie-forwarding construction for this single request.
  3. Compare the two outputs (length, presence/absence of the session cookie,
     any other recorded attribute) and record the concrete point of
     divergence, or record that no divergence was found in the cookie header
     itself and identify the actual cause elsewhere.
- Expected result: a specific, evidenced divergence (or an evidenced
  alternative cause) is recorded — not an unconfirmed guess.
- Evidence: `VOC-043-EV-00`

## VOC-043-TEST-01 — Fixed createServerApiClient() succeeds for a session middleware.ts already validated

- Covers: `VOC-043-AC-01`
- Preconditions: `VOC-043-T01`'s fix applied.
- Procedure:
  1. Construct (or reuse) a request carrying a valid session cookie of the
     same shape `middleware.ts`'s existing check already handles successfully.
  2. Call `createServerApiClient().getCurrentUser()` (as
     `apps/web/src/app/onboarding/page.tsx` does) with this request's cookies
     available.
  3. Confirm the call receives `200` from `/api/v1/me`, not `401`.
  4. Confirm the forwarded `Cookie` header contains exactly the same relevant
     cookie value `middleware.ts` would have forwarded for the same input (no
     unintended narrowing or widening).
- Expected result: `200`, with a cookie header equivalent to `middleware.ts`'s.
- Evidence: `VOC-043-EV-01`

## VOC-043-TEST-02 — Regression test fails pre-fix, passes post-fix, runs in CI

- Covers: `VOC-043-AC-02`
- Preconditions: `VOC-043-T02`'s test implemented.
- Procedure:
  1. Run the new test against the pre-fix `api-server.ts` (or a reverted copy)
     and confirm it fails.
  2. Run the new test against the post-fix `api-server.ts` and confirm it
     passes.
  3. Run the test as part of this repo's normal CI invocation (`pnpm validate`
     or the narrower relevant script) and confirm it is picked up.
- Expected result: fails pre-fix, passes post-fix, runs in normal CI.
- Evidence: `VOC-043-EV-02`

## VOC-043-TEST-03 — Every createServerApiClient() call site is inventoried and covered

- Covers: `VOC-043-AC-03`
- Preconditions: `VOC-043-T03`'s search performed.
- Procedure:
  1. Search the repository for every call site of `createServerApiClient`.
  2. Record the full list.
  3. For each listed call site, confirm it relies on the same
     `createServerApiClient()` implementation `VOC-043-T01` fixed (i.e. no
     call site has its own separate, unfixed cookie-forwarding logic).
- Expected result: a complete, recorded list, with every entry confirmed
  covered by the shared fix.
- Evidence: `VOC-043-EV-03`

This package introduces no migration, so no migration-rollback test is
applicable. No accessibility surface is affected, so no accessibility test is
applicable. Negative/authorization coverage: `VOC-043-TEST-01` step 4's
confirmation that the forwarded cookie header is exactly equivalent (not
wider) to `middleware.ts`'s is this package's authorization-boundary
coverage — the fix must not cause `createServerApiClient()` to forward more
than the request's own real cookies.
