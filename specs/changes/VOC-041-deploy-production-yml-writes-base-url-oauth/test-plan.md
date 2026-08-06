# VOC-041 — Test Plan

## VOC-041-TEST-00 — Rendered step output matches the expected port-qualified values

- Covers: `VOC-041-AC-00`
- Preconditions: `VOC-041-T00`'s fix applied to
  `.github/workflows/deploy-production.yml`.
- Procedure:
  1. Extract the step's `sed`/`echo` script block (or run it directly with
     `PRODUCTION_API_HOST=api-production.vocanova.site` and
     `PRODUCTION_WEB_HOST=production.vocanova.site` exported, redirecting output
     to a scratch file instead of the real `/opt/vocanova/production/secrets/api.env`
     path).
  2. Confirm the rendered `BASE_URL`, `OAUTH_REDIRECT_URI`, and
     `OAUTH_REDIRECT_ALLOWLIST` lines exactly match `VOC-041-AC-00`'s expected
     strings (all three including `:8443` immediately after the host).
  3. Confirm the rendered `SESSION_COOKIE_DOMAIN` line is unchanged
     (`.vocanova.site`, no port).
  4. Confirm the CORS allow-list derived from the rendered `OAUTH_REDIRECT_ALLOWLIST`
     (per `apps/api/app/api/production.go`'s `corsAllowedOrigins` derivation logic)
     would now string-match a real browser's `Origin` header when loaded from
     `https://production.vocanova.site:8443` — i.e. confirm the scheme+host+port
     the derivation produces is `https://production.vocanova.site:8443`, not
     `https://production.vocanova.site`.
- Expected result: all three values include `:8443`; `SESSION_COOKIE_DOMAIN` is
  unaffected; the derived CORS origin matches a real browser's actual origin.
- Evidence: `VOC-041-EV-00`

## VOC-041-TEST-01 — The step's comment no longer asserts the disproven Cloudflare claim

- Covers: `VOC-041-AC-01`
- Preconditions: `VOC-041-T00`'s comment correction applied.
- Procedure:
  1. Read the step's comment block in the post-fix file.
  2. Confirm it does not state or imply that Cloudflare forwards the plain `:443`
     hostname to the origin's `:8443` for these three values.
  3. Confirm it references the issue's live disproof (or an equivalent factual
     basis) as the reason `:8443` is required, consistent with the sibling
     "Poll production API health endpoint" step's existing comment.
- Expected result: the comment accurately reflects the confirmed behavior, with no
  disproven claim remaining.
- Evidence: `VOC-041-EV-01`

## VOC-041-TEST-02 — Regression check fails pre-fix, passes post-fix, runs in CI

- Covers: `VOC-041-AC-02`
- Preconditions: `VOC-041-T01`'s check implemented.
- Procedure:
  1. Run the new check against a copy of the step with `VOC-041-T00`'s fix
     temporarily reverted (unqualified host values) and confirm it fails.
  2. Run the new check against the post-fix step and confirm it passes.
  3. Run the check as part of this repo's normal CI invocation (`pnpm validate` or
     the narrower relevant script) and confirm it is picked up.
- Expected result: fails pre-fix, passes post-fix, runs in normal CI.
- Evidence: `VOC-041-EV-02`

This package introduces no migration, so no migration-rollback test is applicable.
No accessibility surface is affected, so no accessibility test is applicable. This
package's fix narrows the practical CORS gap (correcting a value that currently
causes over-blocking) rather than widening authorization, so no separate
negative/authorization-failure test beyond `VOC-041-TEST-00` step 4 is applicable —
that step's confirmation that the allow-list still names only production's own host
(no wildcard, no extra origin) is this package's authorization-boundary coverage.
