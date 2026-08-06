# VOC-043 — Implementation Plan

## Preconditions and protected areas

Do not begin until this package and each task are approved and implementation
is authorized, per this repository's `AGENTS.md` ("a chat prompt or issue alone
is not implementation authority"). `apps/web/src/lib/api-server.ts` and
`apps/web/src/middleware.ts` are both server-side authentication boundaries —
see `specification.md`'s risk section. Neither has any known in-flight
conflicting work at drafting time.

## File reconciliation and implementation sequence

Existing targets (both read in full at drafting time):
- `apps/web/src/middleware.ts` (128 lines) — the working reference
  implementation; not expected to change unless `VOC-043-T00`'s finding points
  the fix in the other direction (see `specification.md`'s open question 1).
- `apps/web/src/lib/api-server.ts` (32 lines) — the most likely target of the
  actual fix.
- `apps/web/src/app/onboarding/page.tsx` (51 lines) — the only confirmed caller
  of `createServerApiClient()` at drafting time; `VOC-043-T03`'s grep confirms
  whether others exist.

No other file is expected to require changes, pending `VOC-043-T00`'s finding.

Ordered steps:

1. `VOC-043-T00`: add temporary, redacted diagnostic logging to both
   `middleware.ts`'s `cookieHeader` construction (line 61,
   `request.headers.get("cookie") ?? ""`) and `api-server.ts`'s `cookieHeader`
   construction (line 10, `cookieStore.toString()`). Log at minimum each
   value's length and the specific session cookie's value length (obtain the
   session cookie's name via a server-side-safe read, e.g. an existing
   auth-config export, or hardcode the known cookie name if already visible in
   non-secret config — never log the raw token value). Exercise both paths for
   one real (or realistically constructed) request and record the exact
   comparison as this task's evidence.
2. `VOC-043-T01`: apply the fix `VOC-043-T00`'s finding indicates. If the
   finding confirms `specification.md`'s named hypothesis (the two
   implementations diverge in how they reconstruct the `Cookie` header),
   change `api-server.ts` to forward the raw incoming cookie header exactly as
   `middleware.ts` does (or extract one shared helper both use), rather than
   relying on `next/headers`'s `cookies()` API's own reconstruction. Verify by
   confirming `createServerApiClient().getCurrentUser()` succeeds with `200`
   for a session `middleware.ts` already validated as `200` in the same
   request.
3. `VOC-043-T02`: add a deterministic regression test in `apps/web`'s existing
   test stack (implementer to identify the best-fit location, e.g. alongside
   any existing `api-server.ts` or `middleware.ts` unit tests) asserting the
   fixed cookie-forwarding behavior for a representative session-cookie input.
   Must fail against the pre-fix behavior (or be written failing-first) and
   pass against the post-fix behavior.
4. `VOC-043-T03`: run a repository-wide search for `createServerApiClient` call
   sites and record the full list as evidence, confirming each is covered by
   `T01`'s fix.

## Validation and independent verification

Deterministic commands (per `AGENTS.md`'s "Current validation" section):

```bash
bash scripts/governance/validate-governance.sh
bash scripts/governance/classify-change-risk.sh
git diff --check
pnpm validate   # or the narrower relevant apps/web script
```

Plus this package's own `VOC-043-TEST-00`/`01`/`02`/`03` procedures.

Independent verification: per `CLAUDE.md`, an independent reviewer (not the
implementer) must re-review the exact final revision against this
specification, confirm `VOC-043-AC-00` through `AC-03` are each satisfied with
real evidence (not asserted), confirm no raw cookie/session-token value was
logged or committed anywhere (including in the implementation PR's own
description or diagnostic output), and confirm no self-approval occurred.

## Deployment and rollback

Authorization boundary: no deployment is authorized by this package. The fix
takes effect for the `apps/web` (Next.js) service the next time it is
built/deployed after merge — implementer/reviewer to confirm the exact
deploy/redeploy mechanism against this repo's existing `apps/web` release
process (no compose-file or infrastructure change is anticipated here, unlike
VOC-041/VOC-042, since this is an application-code fix, not a configuration
value).

Rollback trigger: the fix does not actually resolve the `401`, or resolves it
but introduces an unintended over-forwarding of unrelated cookie data, or any
diagnostic logging added during `T00` is found post-merge to log more than
intended.

Rollback mechanism: revert the code change (a small, self-contained diff
confined to `apps/web/src/lib/api-server.ts` and/or
`apps/web/src/middleware.ts`, plus the new test) — no migration or
irreversible state change is introduced, so rollback is expected to be
low-risk and fast, consistent with VOC-039/VOC-041/VOC-042's small-diff
rollback precedent.

Owner: named explicitly in the implementation PR at deploy time, not left
implicit.
