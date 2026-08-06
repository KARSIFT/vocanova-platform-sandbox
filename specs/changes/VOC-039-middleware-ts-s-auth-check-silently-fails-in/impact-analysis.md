# VOC-039 — Impact Analysis

## Security and privacy

This fix corrects the runtime `apps/web/src/middleware.ts`'s existing auth check
executes under; it does not change the check's authorization decision. Today, the
check silently fails closed for every real user in production (per issue #297's
confirmed reproduction) — i.e. the current, unfixed defect is fail-safe from a
pure-security standpoint (it over-redirects to `/signin` rather than admitting an
unauthenticated user), but it fully blocks the product. The fix must not become
fail-open: `VOC-039-TEST-00`'s negative case (invalid/missing cookie still
redirects) exists specifically to confirm this.

The new structured logging (`VOC-039-T02`) is the one part of this package with a
real, if small, privacy/secrets surface: log lines added in the auth-check failure
paths must never include the raw `Cookie` header, the raw session cookie value, or
any other credential material — only the failure category, route path, and status
code. This constraint is stated in `specification.md` and enforced by
`VOC-039-TEST-02`'s step 4. No other secret, credential, or personal-data field is
introduced or touched by this package.

## Data and migrations

None. No schema, table, or migration is touched.

## Analytics and accessibility

None. No analytics event is added, and no user-facing UI or accessibility surface
is touched — this is a server-side runtime directive and logging change only.

## Risks, dependencies, and evidence

- `VOC-039-R00`: The open question in `specification.md` (whether
  `runtime = "nodejs"` alone closes the gap given this repo's specific
  nginx/Docker deployment topology, versus the issue's reproduction which ran
  outside the actual middleware execution path) is not yet resolved. Mitigated by
  requiring `VOC-039-T00`'s acceptance criterion to include a real
  deployed-environment re-run of the reproduction technique before the fix is
  accepted as proven.
- `VOC-039-R01`: Next.js's own documentation describes Node-runtime middleware as
  heavier (cold start, resource usage) than Edge-runtime middleware. Given this
  middleware already runs on every request to eight route patterns
  (`config.matcher`), a real but currently unquantified risk is a latency or
  resource-usage regression. Mitigated by `specification.md`'s open question 2
  (a monitoring note, not a blocker) and by `release-plan.md`'s monitoring section.
- `VOC-039-DEP-00`: depends on issue #297's reported reproduction remaining
  accurate — see `change.yaml`'s dependency entry. If a later comment on the issue
  reports something different, this package (specifically `VOC-039-T00`'s scope)
  needs to be revisited before implementation.
- `VOC-039-EV-00`, `VOC-039-EV-01`, `VOC-039-EV-02`: to be produced by
  `VOC-039-T00`/`T01`/`T02` respectively (real-environment reproduction result,
  CI-passing regression test run, and log-line inspection output). None exists yet;
  this package is a draft proposal, not evidence of any of these tasks having been
  performed.
