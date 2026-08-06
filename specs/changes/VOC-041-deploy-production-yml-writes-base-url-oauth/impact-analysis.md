# VOC-041 — Impact Analysis

## Security and privacy

This fix corrects the *value* fed into an already-deployed CORS allow-list and
OAuth redirect mechanism; it does not change the mechanism's authorization
decision or widen it beyond production's own web/API hosts. Today, the mismatch
between the CORS allow-list's stored value (no port) and a real browser's actual
origin (with port `:8443`) causes the browser to block the real `POST` after a
`204` preflight — i.e. the current, unfixed defect is fail-safe from a pure-CORS
standpoint (it over-blocks a legitimate same-product request rather than admitting
a cross-origin one), but it fully blocks Google sign-in. The fix must not become an
accidental broadening of the allow-list (e.g. adding a wildcard, an extra host, or
weakening `apps/api/app/api/production.go`'s existing allow-list matching logic) —
`VOC-041-TEST-00`'s scope explicitly confirms the allow-list still names only
production's own web host, with its correct port, nothing else.

No secret, credential, or personal-data field is introduced, removed, or touched.
`BASE_URL`, `OAUTH_REDIRECT_URI`, and `OAUTH_REDIRECT_ALLOWLIST` are all
non-sensitive configuration values derived from workflow inputs, per this step's
own existing comment (lines 241-255) — this package does not change that
sensitivity classification.

## Data and migrations

None. No schema, table, or migration is touched.

## Analytics and accessibility

None. No analytics event is added, and no user-facing UI or accessibility surface
is touched — this is a CI/CD deploy-configuration change only. (The fix's real
effect is user-facing in that it restores a broken sign-in flow, but no UI code is
changed to achieve it.)

## Risks, dependencies, and evidence

- `VOC-041-R00`: This package's own open question 1 (`specification.md`) —
  `production_api_base_url`'s workflow-input default is very likely affected by
  the same bug class but is out of scope here. If left unfixed, real Google
  sign-in could still fail (or a different real-browser API call could still fail)
  even after this package's fix lands, for a related but distinct reason. Mitigated
  by flagging it explicitly for the reviewing human rather than silently leaving it
  for a future live-reproduction cycle to rediscover.
- `VOC-041-R01`: The underlying host-sharing arrangement (production and staging on
  one host, production shifted to `:8443`) is documented as interim
  (VOC-037/T06's D00 supersession note). Any future call site that assumes the
  default port without checking which environment it targets can reintroduce this
  same defect class. Mitigated by `VOC-041-T01`'s regression check and by this
  step's corrected comment (`VOC-041-AC-01`) explicitly recording the disproven
  Cloudflare claim so a future editor does not reintroduce it on the same mistaken
  belief.
- `VOC-041-R02`: This package's fix only changes what a *future* dispatch of
  `deploy-production.yml` writes. The currently-live production `api.env` already
  has the broken (unqualified) values per the issue's own SSH-confirmed evidence.
  Real Google sign-in stays broken until either a fresh full dispatch runs or an
  operator manually corrects the three lines on the live host — this package
  cannot do either (no production access). Flagged in `README.md`'s recommended
  next action 3 and `specification.md`'s open question 2.
- `VOC-041-DEP-00`: depends on issue #312's reported reproduction remaining
  accurate — see `change.yaml`'s dependency entry. If a later comment on the issue
  reports something different, this package (specifically `VOC-041-T00`'s scope)
  needs to be revisited before implementation.
- `VOC-041-DEP-01`: depends on the reviewing human's decision on
  `production_api_base_url`'s default (open question 1) — see `change.yaml`'s
  second dependency entry.
- `VOC-041-EV-00`, `VOC-041-EV-01`, `VOC-041-EV-02`: to be produced by
  `VOC-041-T00`/`T00`/`T01` respectively (rendered-step verification output,
  before/after diff of the corrected comment, and the new deterministic check's
  passing/failing run). None exist yet.
