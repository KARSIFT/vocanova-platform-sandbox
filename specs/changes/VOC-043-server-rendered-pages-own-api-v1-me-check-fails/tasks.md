# VOC-043 — Tasks

None of the tasks below is implementation-authorized by this package. Adoption
and each task's own implementation-authorization are separate, mirroring
VOC-039/VOC-040/VOC-041/VOC-042's convention. `T00` (instrumentation) must land
and produce a real finding before `T01` (the actual fix) can be scoped
precisely, per `specification.md`'s open question 1 — the exact diff for `T01`
below is therefore described in terms of its required outcome, not a
predetermined line-level change.

## VOC-043-T00 — Instrument both cookie-forwarding paths and capture a real-request comparison

- Requirement source: issue #333's own suggested diagnostic approach
- Acceptance criteria: `VOC-043-AC-00`
- Status: pending
- Summary: add temporary, redacted diagnostic logging to
  `apps/web/src/middleware.ts`'s existing `cookieHeader` construction and to
  `apps/web/src/lib/api-server.ts`'s `createServerApiClient()`'s `cookieHeader`
  construction — logging at minimum each one's length and the specific session
  cookie's value length (never the raw value), consistent with
  `middleware.ts`'s existing `logAuthCheckFailure` no-credential-material
  convention. Exercise both paths for one real (or, if a real production
  request cannot be safely used for this step, a realistically constructed)
  request carrying a valid session cookie, and record the comparison as this
  task's evidence. Per `specification.md`'s open question 2, whether this
  logging is removed or kept as permanent low-cardinality diagnostics is decided
  at review time, not by this task.

## VOC-043-T01 — Fix the identified cookie-forwarding divergence

- Requirement source: issue #333's core reported defect, scoped by `VOC-043-T00`'s
  finding
- Acceptance criteria: `VOC-043-AC-01`
- Status: pending
- Depends on: `VOC-043-T00`'s real-request finding
- Summary: make `apps/web/src/lib/api-server.ts`'s `createServerApiClient()`
  forward cookies to `/api/v1/me` (and any other API call it makes) using the
  same effective cookie-header value `apps/web/src/middleware.ts` already
  forwards successfully for the same request — per `specification.md`'s named
  fix direction, either by changing `api-server.ts` to match `middleware.ts`'s
  raw-header forward, changing `middleware.ts` to match a corrected
  `cookies()`-based approach, or extracting one shared cookie-forwarding helper
  used by both — whichever `VOC-043-T00`'s finding indicates actually resolves
  the divergence. Verify by confirming a request with a valid session cookie
  now receives `200` (not `401`) from both `middleware.ts`'s check and
  `createServerApiClient().getCurrentUser()` in the same request lifecycle.

## VOC-043-T02 — Deterministic regression test for this cookie-forwarding divergence

- Requirement source: issue #333's implicit ask (this defect took three live
  reproductions to isolate; a regression should not require a fourth)
- Acceptance criteria: `VOC-043-AC-02`
- Status: pending
- Depends on: `VOC-043-T01` (or may be written failing-first against the
  pre-fix behavior, at the implementer's discretion)
- Summary: add a deterministic test in `apps/web`'s existing test stack that
  constructs a request carrying a representative session cookie, exercises
  `createServerApiClient()`'s cookie-forwarding behavior, and asserts the
  resulting `Cookie` header matches what `middleware.ts`'s forwarding would
  produce for the same input (or, more directly, asserts a mocked `/api/v1/me`
  call receives the expected cookie value). Runs via `pnpm validate` or a
  narrower documented script.

## VOC-043-T03 — Confirm the full set of createServerApiClient() call sites is covered

- Requirement source: `specification.md`'s open question 3
- Acceptance criteria: `VOC-043-AC-03`
- Status: pending
- Depends on: none (can run independently of T00-T02, though should be recorded
  alongside them for this package's closure)
- Summary: search the repository for every call site of `createServerApiClient`
  (e.g. `apps/web/src/app/onboarding/page.tsx` and any other server component),
  record the full list as this task's evidence, and confirm each one is covered
  by `VOC-043-T01`'s fix to the shared implementation (no per-call-site
  workaround should be needed, since the fix corrects
  `createServerApiClient()` itself).
