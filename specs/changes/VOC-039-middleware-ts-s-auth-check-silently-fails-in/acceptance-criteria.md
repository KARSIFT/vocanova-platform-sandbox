# VOC-039 — Acceptance Criteria

## VOC-039-AC-00 — Middleware auth check succeeds in a real deployed environment

- Requirement source: issue #297's confirmed root cause
- Tasks: `VOC-039-T00`
- Tests: `VOC-039-TEST-00`
- Evidence: `VOC-039-EV-00`
- Result: pending
- Observable outcome: `apps/web/src/middleware.ts` exports
  `runtime = "nodejs"`. A real allowlisted user (or an equivalent controlled check
  using a real, valid session cookie against the real deployed API through the
  actual middleware execution path) reaches a protected route
  (e.g. `/onboarding`) after Google OAuth sign-in without being redirected back to
  `/signin`, in a real deployed environment — not merely in a local dev server.

## VOC-039-AC-01 — A regression test exercises the real Edge-vs-Node fetch behavior gap

- Requirement source: issue #297's first follow-up ask
- Tasks: `VOC-039-T01`
- Tests: `VOC-039-TEST-01`
- Evidence: `VOC-039-EV-01`
- Result: pending
- Observable outcome: A test exists in `apps/web/tests` (or wherever this repo's
  middleware tests already live, if that differs) that fails against a
  pre-`VOC-039-T00` version of `middleware.ts` and passes against the
  post-`VOC-039-T00` version, using a mechanism that actually exercises Next.js's
  Edge runtime constraints on `fetch()` rather than a plain-Node unit test that
  would pass regardless of the `runtime` export. This test runs in this repo's
  normal CI (`pnpm validate` or the relevant narrower script), not only manually.

## VOC-039-AC-02 — Auth-check failures are diagnosable from normal logs without SSH access

- Requirement source: issue #297's second follow-up ask
- Tasks: `VOC-039-T02`
- Tests: `VOC-039-TEST-02`
- Evidence: `VOC-039-EV-02`
- Result: pending
- Observable outcome: Each of the three auth-check failure branches ("fetch threw",
  "got 401", "got some other non-ok status") in `apps/web/src/middleware.ts` emits
  a distinguishable structured log line, visible in this repo's real production
  logging/monitoring destination, containing the failure category, route path, and
  (where applicable) HTTP status code, and containing no session cookie value, raw
  `Cookie` header, or other credential material.
