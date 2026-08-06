# VOC-041 — Acceptance Criteria

## VOC-041-AC-00 — Deploy workflow writes port-qualified BASE_URL/OAUTH_REDIRECT_URI/OAUTH_REDIRECT_ALLOWLIST

- Requirement source: issue #312's confirmed root cause and suggested fix
- Tasks: `VOC-041-T00`
- Tests: `VOC-041-TEST-00`
- Evidence: `VOC-041-EV-00`
- Result: pending
- Observable outcome: `.github/workflows/deploy-production.yml`'s "Write
  production application configuration" step writes
  `BASE_URL=https://${PRODUCTION_API_HOST}:8443`,
  `OAUTH_REDIRECT_URI=https://${PRODUCTION_API_HOST}:8443/api/v1/auth/oauth/google/callback`,
  and `OAUTH_REDIRECT_ALLOWLIST=https://${PRODUCTION_WEB_HOST}:8443/onboarding,https://${PRODUCTION_WEB_HOST}:8443/home`.
  `SESSION_COOKIE_DOMAIN` remains unchanged (still the parent-domain form, no port).
  Verified by rendering the step's script with representative host values (e.g. via
  a shell harness that sources the same `sed`/`echo` logic, or an equivalent
  deterministic check) rather than only by visual inspection of the diff.

## VOC-041-AC-01 — The step's own comment no longer asserts the disproven Cloudflare port-forwarding claim

- Requirement source: issue #312's root-cause section (contradiction with the
  sibling health-check step's comment)
- Tasks: `VOC-041-T00`
- Tests: `VOC-041-TEST-01`
- Evidence: `VOC-041-EV-01`
- Result: pending
- Observable outcome: the step's comment (currently lines 289-298) no longer
  states or implies that Cloudflare forwards the plain `:443` hostname to the
  origin's `:8443` for these three values. The replacement comment records the
  issue's live disproof and is consistent with the sibling "Poll production API
  health endpoint" step's existing comment (currently lines 476-486), which already
  documents the same underlying finding for its own curl calls.

## VOC-041-AC-02 — A deterministic regression check catches an unqualified (portless) value in this step before the next real dispatch

- Requirement source: issue #312's implicit ask (the same defect class — a
  config-writing step silently drifting from the deploy target's real serving
  port — should not require a live production reproduction to catch again)
- Tasks: `VOC-041-T01`
- Tests: `VOC-041-TEST-02`
- Evidence: `VOC-041-EV-02`
- Result: pending
- Observable outcome: a deterministic check exists (as an automated test if this
  repository's existing test stack can parse/exercise workflow YAML content, or as
  a documented script otherwise) that fails if any of `BASE_URL`,
  `OAUTH_REDIRECT_URI`, or `OAUTH_REDIRECT_ALLOWLIST` in this step is written
  without `:8443`, and passes against the post-fix step. This check runs in this
  repo's normal CI (`pnpm validate` or a narrower documented script), not only
  manually.
