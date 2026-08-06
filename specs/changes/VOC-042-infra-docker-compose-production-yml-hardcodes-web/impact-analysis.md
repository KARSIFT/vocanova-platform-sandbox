# VOC-042 — Impact Analysis

## Security and privacy

This fix corrects the *value* of a server-side base-URL configuration used by
`apps/web/src/middleware.ts`'s auth check; it does not change the auth check's
decision logic, session-cookie handling, or OAuth flow itself. Today, the
unqualified value causes every server-side `/api/v1/me` call to fail to reach the
real API, which is fail-safe from an authorization standpoint (it denies access to
users who should be allowed in, rather than admitting anyone) but fully blocks
sign-in for every real user. The fix must not become an accidental widening of what
`getApiBaseURL()` resolves to beyond production's own API host — `VOC-042-TEST-00`'s
scope explicitly confirms the resolved value is exactly
`https://api-production.vocanova.site:8443`, nothing else.

No secret, credential, or personal-data field is introduced, removed, or touched.
`API_BASE_URL` is a plain, non-sensitive hostname configuration value set via
`environment:` (not `secrets/`), consistent with this file's own existing treatment
of it.

## Data and migrations

None. No schema, table, or migration is touched.

## Analytics and accessibility

None. No analytics event is added, and no user-facing UI or accessibility surface
is touched — this is a production infrastructure configuration change only. (The
fix's real effect is user-facing in that it restores a broken sign-in flow, but no
UI code is changed to achieve it.)

## Risks, dependencies, and evidence

- `VOC-042-R00`: This is the third confirmed recurrence of the same
  same-host-missing-port bug class (VOC-038-T03, VOC-041, this package), all
  traceable to the interim production/staging host-sharing arrangement
  (VOC-037/T06's D00 supersession note). If a broader audit is not eventually
  commissioned, further call sites likely still carry the same defect and will
  keep surfacing one live failure at a time. Mitigated by flagging this explicitly
  in `specification.md`'s open question 1 rather than silently expanding this
  package's scope to audit every call site without approval to do so.
- `VOC-042-R01`: This package's fix only changes what a *future* recreation of the
  `web` service uses. The currently-running production `web` container already has
  the broken (unqualified) value per the issue's own SSH-confirmed evidence. Real
  Google sign-in stays broken past the OAuth callback until either a fresh
  recreate/redeploy runs or an operator manually corrects the running container.
  Flagged in `README.md`'s recommended next action 3 and `specification.md`'s open
  question 2.
- `VOC-042-R02`: `infra/docker-compose.yml` (the non-production/staging compose
  file) is not audited by this package for an equivalent defect. If it has one,
  sign-in against that environment would independently fail for the same reason.
  Flagged in `specification.md`'s open question 3.
- `VOC-042-DEP-00`: depends on issue #319's reported reproduction remaining
  accurate — see `change.yaml`'s dependency entry. If a later comment on the issue
  reports something different, this package (specifically `VOC-042-T00`'s scope)
  needs to be revisited before implementation.
- `VOC-042-DEP-01`, `VOC-042-DEP-02`: depend on the reviewing human's decisions on
  the broader-audit follow-up and the interim manual container recreate,
  respectively — see `change.yaml`'s dependency entries.
- `VOC-042-EV-00`, `VOC-042-EV-01`, `VOC-042-EV-02`: to be produced by
  `VOC-042-T00`/`T00`/`T01` respectively (post-fix file content, a server-side
  `getApiBaseURL()` resolution check, and the new deterministic check's
  passing/failing run). None exist yet.
