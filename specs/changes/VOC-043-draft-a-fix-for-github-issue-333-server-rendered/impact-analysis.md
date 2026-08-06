# VOC-043 — Impact Analysis

## Security and privacy

This fix changes *how* `apps/web/src/lib/api-server.ts` reads and forwards the
incoming request's session cookie to the API's `/api/v1/me` endpoint — from a
`cookies()`-parsed-and-reconstructed value to a raw pass-through of the incoming
`Cookie` header, matching `apps/web/src/middleware.ts`'s own already-shipping
behavior. It does not change *who* is authenticated, *what* the API does with the
cookie once received, or the auth-check decision logic in either file. Today, the
divergence is fail-safe from an authorization standpoint (it denies a real,
already-authenticated user access to `/onboarding`/`/settings/account`, rather
than admitting anyone who shouldn't be), but it fully blocks real onboarding for
every real user who reaches this code path. The fix must not become an accidental
widening of what gets forwarded — `VOC-043-TEST-00`'s scope explicitly asserts the
forwarded header is byte-for-byte identical to the raw incoming header, not a
superset or a transformed value.

No secret, credential, or personal-data field is introduced, removed, or
newly logged. This package deliberately does not add temporary diagnostic
logging of cookie values to production (see `specification.md`'s "Non-goals"),
consistent with `apps/web/src/middleware.ts`'s own existing
`logAuthCheckFailure()` convention of never including cookie/credential material
in a log line (enforced today by
`apps/web/tests/middleware/auth-check-logging.test.ts`'s
`assertNoCredentialMaterial()` check). If a future task does add temporary
instrumentation per `specification.md`'s open question 1, it must follow that
same convention (length/prefix only, never the raw cookie value) and must be
removed once its diagnostic purpose is served — this package's own scope does not
add any such logging.

## Data and migrations

None. No schema, table, or migration is touched.

## Analytics and accessibility

None. No analytics event is added, and no user-facing UI or accessibility surface
is touched — this is a server-side request-forwarding fix only. (The fix's real
effect is user-facing in that it restores a broken onboarding flow, but no UI
code is changed to achieve it.)

## Risks, dependencies, and evidence

- `VOC-043-R00`: The fix is offered by the issue as the "likely" correction, not
  an empirically-confirmed one — see `specification.md`'s open question 1.
  Mitigated by explicitly flagging that if the implementer's own verification
  (a real end-to-end reproduction, or the closest available lower-environment
  equivalent) still shows the failure after this fix, the task should stop and
  add temporary redacted instrumentation rather than guessing at a second fix
  without evidence.
- `VOC-043-R01`: Any other call site in `apps/web` that independently
  reconstructs a `Cookie` header from `cookies()` would carry the same class of
  risk this issue surfaced, and is not audited by this package beyond a
  drafting-time targeted search that found none. See `specification.md`'s open
  question 2.
- `VOC-043-R02`: Real users may currently be stuck mid-onboarding in production
  because of this defect, independent of whether the code fix lands quickly.
  This package has no production access to determine or remediate that. See
  `specification.md`'s open question 3.
- `VOC-043-DEP-00`, `VOC-043-DEP-01`, `VOC-043-DEP-02`: see `change.yaml`'s
  dependency entries — each depends on a reviewing-human decision or on the
  issue's reported reproduction remaining accurate, not yet resolved at
  drafting time.
- `VOC-043-EV-00`, `VOC-043-EV-01`, `VOC-043-EV-02`: to be produced by
  `VOC-043-T00`/`T00`/`T01` respectively (post-fix file content and passing
  `VOC-043-TEST-00`/`01` runs, and the new regression check's passing/failing
  run). None exist yet.
