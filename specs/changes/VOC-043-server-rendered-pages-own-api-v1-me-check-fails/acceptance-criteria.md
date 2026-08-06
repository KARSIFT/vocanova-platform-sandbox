# VOC-043 — Acceptance Criteria

## VOC-043-AC-00 — The exact divergence between the two cookie-forwarding paths is identified from a single real request

- Requirement source: issue #333's own suggested diagnostic approach
- Tasks: `VOC-043-T00`
- Tests: `VOC-043-TEST-00`
- Evidence: `VOC-043-EV-00`
- Result: pending
- Observable outcome: for one real (or realistically simulated, if a real
  production request cannot be safely exercised for this instrumentation step)
  sign-in attempt, a redacted comparison exists (e.g. cookie-header length, and
  the specific session cookie's value length) between what
  `apps/web/src/middleware.ts` forwarded to `/api/v1/me` and what
  `apps/web/src/lib/api-server.ts`'s `createServerApiClient()` forwarded for the
  same request, and the comparison identifies a concrete point of divergence
  (or, if no divergence is found in the cookie header itself, identifies the
  actual cause elsewhere, per specification.md's open question 1) — not asserted
  in advance of this evidence.

## VOC-043-AC-01 — apps/web/src/lib/api-server.ts's createServerApiClient() forwards cookies such that the onboarding page's own /api/v1/me check succeeds for the same session middleware.ts already validated

- Requirement source: issue #333's core reported defect
- Tasks: `VOC-043-T01`
- Tests: `VOC-043-TEST-01`
- Evidence: `VOC-043-EV-01`
- Result: pending
- Observable outcome: given a request carrying a valid session cookie (the same
  shape `middleware.ts`'s existing, working forward already handles), a call to
  `createServerApiClient().getCurrentUser()` (as `apps/web/src/app/onboarding/page.tsx`
  performs) receives a `200` from `/api/v1/me`, not a `401`, using exactly one
  cookie-forwarding implementation shared by (or made equivalent to) both
  `middleware.ts` and `api-server.ts`, per specification.md's in-scope fix
  direction.

## VOC-043-AC-02 — A deterministic regression test catches this specific cookie-forwarding divergence before the next live reproduction is needed

- Requirement source: issue #333's implicit ask (three live reproductions were
  needed to isolate this; a future regression of the same class should not
  require a fourth)
- Tasks: `VOC-043-T02`
- Tests: `VOC-043-TEST-02`
- Evidence: `VOC-043-EV-02`
- Result: pending
- Observable outcome: a deterministic, non-production test exists (in
  `apps/web`'s existing test stack) that constructs a request carrying a
  representative session cookie, exercises `createServerApiClient()`'s
  cookie-forwarding path, and fails if a future change reintroduces a
  cookie-header divergence from `middleware.ts`'s forwarding behavior for the
  same input. This test runs via `pnpm validate` or a narrower documented
  script, not only manually.

## VOC-043-AC-03 — Any other createServerApiClient() call site is confirmed covered by the same fix, or explicitly confirmed to not exist

- Requirement source: specification.md's open question 3 (in scope for this
  package, not deferred)
- Tasks: `VOC-043-T03`
- Tests: `VOC-043-TEST-03`
- Evidence: `VOC-043-EV-03`
- Result: pending
- Observable outcome: a repository-wide search for `createServerApiClient` call
  sites is performed and its result (the full list, or an explicit confirmation
  that `apps/web/src/app/onboarding/page.tsx` is the only one) is recorded in
  this package's evidence, with each listed call site confirmed to be covered by
  `VOC-043-T01`'s fix (since the fix corrects the shared implementation, not a
  per-call-site workaround).
