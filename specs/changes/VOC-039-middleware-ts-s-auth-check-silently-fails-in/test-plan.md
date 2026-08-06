# VOC-039 — Test Plan

## VOC-039-TEST-00 — Real deployed-environment auth-check reproduction (positive and negative)

- Covers: `VOC-039-AC-00`
- Preconditions: `VOC-039-T00`'s fix deployed to a real environment capable of
  exercising the full OAuth flow (staging preferred; production only if staging
  cannot exercise OAuth, using an existing allowlisted/test account, never a newly
  created production-only credential created solely for this test).
- Procedure:
  1. Positive case: complete Google OAuth sign-in as a real (allowlisted or
     equivalent test) account; confirm the session cookie is set; navigate to a
     protected route (e.g. `/onboarding`); confirm no redirect back to `/signin`.
  2. Negative case: send a request to the same protected route with no cookie, or
     with a deliberately invalid/expired session cookie; confirm the middleware
     still correctly redirects to `/signin` (i.e. this fix does not weaken the
     check — it only fixes the runtime it runs under).
  3. Compare against the issue's own reproduction technique (real cookie, real API,
     via `node -e` inside the container) as a sanity cross-check that the two paths
     now agree.
- Expected result: positive case reaches the protected route; negative case still
  redirects to `/signin`.
- Evidence: `VOC-039-EV-00`

## VOC-039-TEST-01 — Automated regression test for the Edge-vs-Node fetch gap

- Covers: `VOC-039-AC-01`
- Preconditions: a mechanism in this repo's test stack capable of actually
  exercising Next.js's Edge runtime constraints for middleware (implementer to
  identify — see `VOC-039-T01`'s note on investigating before assuming a specific
  tool).
- Procedure:
  1. Run the new test against a version of `middleware.ts` without
     `runtime = "nodejs"` (e.g. by temporarily reverting just that line in a
     throwaway local check, or via the test's own before/after harness if it
     supports toggling this) and confirm it fails, demonstrating it actually
     exercises the real bug rather than passing unconditionally.
  2. Run the new test against the fixed `middleware.ts` and confirm it passes.
  3. Run the test as part of this repo's normal CI invocation
     (`pnpm validate` or the narrower relevant script) and confirm it is picked up.
- Expected result: fails pre-fix, passes post-fix, runs in normal CI.
- Evidence: `VOC-039-EV-01`

## VOC-039-TEST-02 — Structured log line presence and content for each failure branch

- Covers: `VOC-039-AC-02`
- Preconditions: `VOC-039-T02`'s logging implemented; access to a real or
  realistic non-production log sink to inspect output without needing SSH access
  to a live production box for this test itself.
- Procedure:
  1. Trigger the "fetch threw" branch (e.g. point at an unreachable API URL in a
     controlled non-production check) and confirm a distinguishable log line
     appears with the correct category.
  2. Trigger the "got 401" branch (a request with no/invalid session cookie against
     a real running API) and confirm a distinguishable log line appears with
     status `401`.
  3. Trigger the "got some other non-ok status" branch (e.g. simulate a `500` from
     the API in a controlled check) and confirm a distinguishable log line appears
     with the correct status code.
  4. Inspect all three log lines for the absence of any `Cookie` header value,
     session cookie value, or other credential material.
  5. Confirm at least one of the three log lines is visible in this repo's real
     production logging/monitoring destination after deployment (not only printed
     locally), satisfying the "diagnosable from normal logs" requirement.
- Expected result: three distinguishable, credential-free log lines, each visible
  in the real destination.
- Evidence: `VOC-039-EV-02`

This package introduces no migration, so no migration-rollback test is applicable.
`VOC-039-TEST-00`'s negative case is this package's authorization-failure coverage.
No accessibility surface is affected, so no accessibility test is applicable.
