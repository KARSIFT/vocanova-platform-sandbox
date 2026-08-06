# VOC-043 — Implementation Plan

## Preconditions and protected areas

Do not begin until this package and each task are approved and implementation is
authorized, per this repository's `AGENTS.md` ("a chat prompt or issue alone is
not implementation authority"). `apps/web/src/lib/api-server.ts` is a protected
R3 area (the server-side cookie-forwarding path every authenticated
server-rendered page in `apps/web` depends on) — see `specification.md`'s risk
section.

## File reconciliation and implementation sequence

Existing target: `apps/web/src/lib/api-server.ts` (read in full at drafting
time — 32 lines). `createServerApiClient()` (lines 8-22) is the sole in-scope
function; `requireAuthRedirect()` (lines 24-31) is read-only reference at
drafting time, not touched. No conflicting in-flight work against this file is
known at drafting time.

Reference (read-only, not modified): `apps/web/src/middleware.ts`'s raw-header
forward, currently `const cookieHeader = request.headers.get("cookie") ?? "";`
(around line 62), and its subsequent use as the literal `Cookie` header value on
the `/api/v1/me` fetch. This is the behavior `api-server.ts` is being aligned to.

Ordered steps:

1. `VOC-043-T00`: in `apps/web/src/lib/api-server.ts`, replace the `cookies()`
   import and its `(await cookies()).toString()` reconstruction with a raw read
   of the incoming request's `Cookie` header via `next/headers`'s `headers()`
   API (`(await headers()).get("cookie") ?? ""`), matching
   `middleware.ts`'s exact forwarding semantics (forward the header value
   unmodified, including when absent — an absent cookie must still result in no
   `Cookie` header being set, not a literal `"null"`/`"undefined"` string).
   Leave `requireAuthRedirect()` and every other line in this file unchanged.
   Verify by reading the post-fix file directly and by
   `VOC-043-TEST-00`/`VOC-043-TEST-01`'s procedures.
2. `VOC-043-T01`: add a regression test (new file, e.g.
   `apps/web/tests/lib/api-server-cookie-forwarding.test.ts`, mirroring
   `apps/web/tests/middleware/auth-check-logging.test.ts`'s existing
   `node:test`-based convention and its use of a clearly-fake session-cookie
   value) that exercises `createServerApiClient()`'s custom `fetch` against a
   representative multi-cookie raw `Cookie` header (e.g. a session cookie plus a
   CSRF double-submit cookie, matching the two cookie kinds
   `apps/api/business/auth/tokens.go`'s `SessionCookie`/`CSRFToken` actually set)
   and asserts the header value passed to the injected `fetch` is byte-for-byte
   identical to the raw input. Must fail against the pre-fix
   (`cookies()`-reconstructed) implementation and pass against the post-fix one —
   implementer should verify this by temporarily reverting `VOC-043-T00`'s change
   in a throwaway local check, confirming the new test fails, then restoring the
   fix and confirming it passes. Runs via `pnpm validate` or a narrower
   documented script, per this repo's existing CI convention.

## Validation and independent verification

Deterministic commands (per `AGENTS.md`'s "Current validation" section):

```bash
bash scripts/governance/validate-governance.sh
bash scripts/governance/classify-change-risk.sh
git diff --check
pnpm validate   # or the narrower relevant script, once VOC-043-T01's test exists
```

Plus this package's own `VOC-043-TEST-00`/`01`/`02` procedures.

Independent verification: per `CLAUDE.md`, an independent reviewer (not the
implementer) must re-review the exact final revision against this specification,
confirm `VOC-043-AC-00`/`01`/`02` are each satisfied with real evidence (not
asserted), and confirm no self-approval occurred. The reviewer should also
confirm this package's stated non-goals (no change to API-side session
validation, `middleware.ts` itself, or any page component) were actually
preserved in the diff, and should specifically check that no temporary
diagnostic logging of cookie material was introduced, per
`impact-analysis.md`'s security note.

## Deployment and rollback

Authorization boundary: no deployment is authorized by this package. The fix
takes effect the next time `apps/web` is rebuilt and redeployed (via this repo's
existing deploy workflow); it does not itself perform that deploy.

Rollout sequence (once authorized): merge to `develop`; the fix ships with the
next redeploy of the `web` service.

Rollback trigger: a future edit reintroduces a `cookies()`-based reconstruction
(`VOC-043-T01`'s regression test should catch this before merge, but a rollback
trigger also applies if it somehow reaches production), or the fix is found not
to resolve the reported 401 per `specification.md`'s open question 1. Rollback
mechanism: revert this package's diff (a small, self-contained change to one
file plus one new test file) — no migration is involved, so rollback is expected
to be low-risk and fast. Owner: named explicitly in the implementation PR at
deploy time, not left implicit.
