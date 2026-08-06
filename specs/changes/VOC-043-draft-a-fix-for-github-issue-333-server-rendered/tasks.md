# VOC-043 — Tasks

None of the tasks below is implementation-authorized by this package. Adoption
and each task's own implementation-authorization are separate, mirroring
VOC-039/VOC-040/VOC-041/VOC-042's convention. `T00` is the actual fix and should
land first; `T01`'s regression test depends on `T00`'s fixed behavior existing to
assert against (or may be written failing-first against the pre-fix behavior, at
the implementer's discretion).

## VOC-043-T00 — Forward the raw incoming Cookie header from api-server.ts, matching middleware.ts

- Requirement source: issue #333's "likely area to investigate" and "suggested
  fix direction"
- Acceptance criteria: `VOC-043-AC-00`, `VOC-043-AC-01`
- Status: pending
- Summary: in `apps/web/src/lib/api-server.ts`'s `createServerApiClient()`,
  replace the `cookies()`-based reconstruction with a raw read of the incoming
  `Cookie` header via `next/headers`'s `headers()` API, exactly as specified in
  `implementation-plan.md` step 1. Leave `requireAuthRedirect()` and every other
  line in this file unchanged. Verify by reading the post-fix file directly and
  by exercising the existing callers (`apps/web/src/app/onboarding/page.tsx`,
  `apps/web/src/app/(app)/settings/account/page.tsx`) against the post-fix
  behavior.

## VOC-043-T01 — Deterministic regression test for cookie-forwarding divergence

- Requirement source: issue #333's implicit ask (this defect class — two
  "identical-looking" implementations that silently disagree — should be
  catchable without a live production reproduction)
- Acceptance criteria: `VOC-043-AC-02`
- Status: pending
- Depends on: `VOC-043-T00` (or may be written first as a failing-first test
  against the pre-fix behavior, at the implementer's discretion)
- Summary: add a new test file (e.g.
  `apps/web/tests/lib/api-server-cookie-forwarding.test.ts`) that asserts
  `createServerApiClient()`'s forwarded `Cookie` header is byte-for-byte
  identical to a representative multi-cookie raw incoming header, exactly as
  specified in `implementation-plan.md` step 2. Must fail against the pre-fix
  implementation and pass against the post-fix one. Runs via `pnpm validate` or
  a narrower documented script.
