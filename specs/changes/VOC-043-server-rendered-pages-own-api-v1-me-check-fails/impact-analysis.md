# VOC-043 — Impact Analysis

## Security and privacy

This fix corrects a server-side cookie-forwarding implementation that today
spuriously denies access to real, validly-authenticated users (a fail-safe but
fully-blocking defect, not an over-permissive one) — the same posture as
VOC-042's finding for its own defect. The fix must not widen what
`createServerApiClient()` forwards beyond the cookies the incoming request
actually carries — `VOC-043-TEST-01`/`TEST-02`'s scope should confirm the
forwarded header reflects only the real incoming cookies, not an expanded or
substituted set. Any diagnostic logging added during `VOC-043-T00`'s
instrumentation step must redact the actual session-token value, logging at
most length/presence, consistent with `apps/web/src/middleware.ts`'s existing
`logAuthCheckFailure` convention and the issue's own explicit instruction. No
secret, credential, or personal-data field is introduced beyond what these two
code paths already handle today.

## Data and migrations

None. No schema, table, or migration is touched.

## Analytics and accessibility

None. No analytics event is added, and no user-facing UI or accessibility
surface is touched — this is a server-side cookie-forwarding fix only. (The
fix's real effect is user-facing in that it restores a broken onboarding flow,
but no UI/rendering code is changed to achieve it.)

## Risks, dependencies, and evidence

- `VOC-043-R00`: The exact root cause is not confirmed at drafting time — only
  two candidate implementations and a suggested diagnostic approach are named.
  If `VOC-043-T00`'s instrumentation finds the divergence is not in the cookie
  header at all (e.g. it is instead in how the two calls reach the API, or in
  something specific to the `onboarding` route), `VOC-043-T01`'s scope as
  currently described (unifying the two cookie-forwarding implementations) may
  need to be revised before implementation continues — see
  `specification.md`'s open question 1. Mitigated by explicitly sequencing
  instrumentation (`T00`) before the fix (`T01`) rather than committing to an
  unconfirmed diff in advance.
- `VOC-043-R01`: Real production sign-in remains broken (in the same way issue
  #333 describes) until this defect is actually fixed and deployed — no interim
  mitigation is proposed by this package, since the issue's evidence shows no
  safe workaround exists (both middleware and page-level checks are legitimate,
  necessary auth boundaries; disabling either is not an option this package
  considers).
- `VOC-043-R02`: If diagnostic logging added during `T00` is not removed before
  merge and is later found to be higher-cardinality or more sensitive than
  intended, it could become an unintended standing logging surface. Mitigated
  by `specification.md`'s open question 2 explicitly flagging this decision for
  the reviewing human rather than defaulting silently either way.
- `VOC-043-DEP-00`, `VOC-043-DEP-01`, `VOC-043-DEP-02`: see `change.yaml`'s
  dependency entries — this package's implementation depends on the issue's
  reported reproduction remaining accurate, on `T00`'s real-request finding
  actually confirming (or redirecting) the suggested fix direction, and on the
  reviewing human's decision on whether diagnostic logging persists.
- `VOC-043-EV-00` through `VOC-043-EV-03`: to be produced by `VOC-043-T00`
  through `T03` respectively (the real-request comparison, the fixed
  behavior's passing verification, the new regression test's passing run, and
  the call-site inventory). None exist yet.
