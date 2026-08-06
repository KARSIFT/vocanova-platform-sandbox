# VOC-039 — Tasks

None of the tasks below is implementation-authorized by this package. Adoption and
each task's own implementation-authorization are separate, mirroring VOC-037/VOC-038's
convention. Ordering reflects dependency, not priority — `T00` is the actual fix and
should land first so `T01`/`T02` have real fixed behavior to test/log against, though
a reviewer may choose to dispatch `T01`'s test as a failing-first regression test
before `T00` lands, if that ordering is preferred.

## VOC-039-T00 — Force middleware onto the Node.js runtime

- Requirement source: issue #297's confirmed-by-reproduction root cause and
  suggested fix
- Acceptance criteria: `VOC-039-AC-00`
- Status: pending
- Summary: Add `export const runtime = "nodejs";` to `apps/web/src/middleware.ts`
  (Next.js 15.2+/16.2 supports this; this repo is on 16.2.10). Must not change any
  other exported behavior (`config.matcher`, the `middleware` function's auth/
  onboarding-status logic). After deploying to a real environment that can exercise
  the actual OAuth flow (staging if it can, else a controlled production check),
  re-run the issue's own reproduction technique (a real, valid session cookie
  against the real deployed API, this time through the actual middleware execution
  path with the fix applied, not the plain-Node-outside-middleware comparison the
  issue used to diagnose the cause) and confirm the auth check now succeeds. A
  passing regression test (`VOC-039-T01`) alone does not satisfy this task —
  see `specification.md`'s open question 1.

## VOC-039-T01 — Regression test exercising the real Edge-vs-Node fetch behavior difference

- Requirement source: issue #297's first follow-up ask
- Acceptance criteria: `VOC-039-AC-01`
- Status: pending
- Depends on: `VOC-039-T00` (or may be written first as a failing-first test against
  pre-fix `middleware.ts`, at the implementer's discretion)
- Summary: Add a test that would have caught this specific bug class — i.e. one that
  actually runs (or faithfully emulates) Next.js's Edge runtime constraints for
  `middleware.ts`'s `fetch()` call, not a test that only exercises route-matching or
  calls the auth-check logic as a plain Node function. Concretely, this likely means
  using Next.js's own middleware test tooling/`next dev`'s Edge runtime simulation
  (or an equivalent that this repo's existing test stack, per `apps/web/tests`,
  already supports) rather than a bare Jest/Vitest unit test — the implementer
  should confirm which of this repo's already-installed tools can actually exercise
  the Edge runtime distinction before assuming a specific mechanism, and record that
  investigation in the implementation PR if the obvious approach doesn't work.
  Must fail against a version of `middleware.ts` without `VOC-039-T00`'s fix (i.e.
  demonstrably exercises the real bug) and pass with it.

## VOC-039-T02 — Structured logging in the auth-check failure paths

- Requirement source: issue #297's second follow-up ask
- Acceptance criteria: `VOC-039-AC-02`
- Status: pending
- Summary: Add structured log lines in `apps/web/src/middleware.ts`'s three distinct
  auth-check failure branches (the `try/catch` around the `/api/v1/me` fetch — "fetch
  threw"; the `meResponse.status === 401` branch — "got 401"; the
  `!meResponse.ok` branch for any other non-2xx status — "got some other non-ok
  status"), using whatever structured-logging/observability convention this repo's
  `apps/web` already has for server-side code (implementer must identify and reuse
  it rather than introducing a new one). Log lines must include the failure category,
  the route path, and (for the two response-based branches) the actual HTTP status
  code, and must NOT include the raw `Cookie` header, the raw session cookie value,
  or any other credential material — see `specification.md`'s security/privacy
  section. Confirm the resulting log lines are actually visible in this repo's real
  production logging/monitoring destination (not just printed to stdout locally).
