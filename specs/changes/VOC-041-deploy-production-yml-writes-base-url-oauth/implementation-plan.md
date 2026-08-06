# VOC-041 — Implementation Plan

## Preconditions and protected areas

Do not begin until this package and each task are approved and implementation is
authorized, per this repository's `AGENTS.md` ("a chat prompt or issue alone is not
implementation authority"). `.github/workflows/deploy-production.yml` is a
protected R3 area (CI/CD deploy tooling that writes the production
authentication/CORS boundary) — see `specification.md`'s risk section.

## File reconciliation and implementation sequence

Existing target: `.github/workflows/deploy-production.yml` (read in full at
drafting time — 539 lines). The "Write production application configuration" step
(lines 256-322) is the sole in-scope step. No conflicting in-flight work against
this file is known at drafting time. This package's diff is narrow and additive to
existing lines: the three `echo` lines at 299-301 gain `:8443`, and the comment
block at 289-298 is corrected — no other line in the step, and no other step in the
file, is touched.

Ordered steps:

1. `VOC-041-T00`: in the "Write production application configuration" step,
   change:
   - `echo "BASE_URL=https://${PRODUCTION_API_HOST}"` to
     `echo "BASE_URL=https://${PRODUCTION_API_HOST}:8443"`
   - `echo "OAUTH_REDIRECT_URI=https://${PRODUCTION_API_HOST}/api/v1/auth/oauth/google/callback"`
     to
     `echo "OAUTH_REDIRECT_URI=https://${PRODUCTION_API_HOST}:8443/api/v1/auth/oauth/google/callback"`
   - `echo "OAUTH_REDIRECT_ALLOWLIST=https://${PRODUCTION_WEB_HOST}/onboarding,https://${PRODUCTION_WEB_HOST}/home"`
     to
     `echo "OAUTH_REDIRECT_ALLOWLIST=https://${PRODUCTION_WEB_HOST}:8443/onboarding,https://${PRODUCTION_WEB_HOST}:8443/home"`

   Leave `SESSION_COOKIE_DOMAIN` (line 312) unchanged. Replace the comment at lines
   289-298 (currently asserting the disproven Cloudflare port-forwarding claim)
   with one that records the issue's live disproof and points to the sibling
   "Poll production API health endpoint" step's comment as the earlier instance of
   the same finding — mirroring that step's own comment style (a short "why", not
   a repeat of the full investigation). Verify by rendering the step's script
   locally (e.g. `bash -c` with `PRODUCTION_API_HOST`/`PRODUCTION_WEB_HOST` set to
   representative values, running just the `sed`/`echo` block in isolation against
   a scratch file) that the three written lines match `VOC-041-AC-00`'s exact
   expected strings.
2. `VOC-041-T01`: add a deterministic regression check for this defect class.
   Concretely: a small script or test (implementer to identify this repo's best-fit
   mechanism — e.g. a shell test under an existing `scripts`/CI-test convention, or
   a lightweight YAML-parsing test if `apps/web`'s or a shared package's test stack
   already parses YAML for another purpose) that extracts the three `echo` lines
   from the step (or a copy of the relevant script fragment) and asserts each
   contains `:8443` immediately after the host variable and before the path (for
   `OAUTH_REDIRECT_URI`/`OAUTH_REDIRECT_ALLOWLIST`) or at the end of the value (for
   `BASE_URL`). Must fail against the pre-fix (unqualified) lines and pass against
   the post-fix ones — implementer should verify this by temporarily reverting
   `VOC-041-T00`'s change in a throwaway local check, confirming the new check
   fails, then restoring the fix and confirming it passes. Runs via `pnpm validate`
   or a narrower documented script, per this repo's existing CI convention.

## Validation and independent verification

Deterministic commands (per `AGENTS.md`'s "Current validation" section):

```bash
bash scripts/governance/validate-governance.sh
bash scripts/governance/classify-change-risk.sh
git diff --check
pnpm validate   # or the narrower relevant script, once VOC-041-T01's check exists
```

Plus this package's own `VOC-041-TEST-00`/`01`/`02` procedures.

Independent verification: per `CLAUDE.md`, an independent reviewer (not the
implementer) must re-review the exact final revision against this specification,
confirm `VOC-041-AC-00`/`01`/`02` are each satisfied with real evidence (not
asserted), and confirm no self-approval occurred. The reviewer should also confirm
this package's stated non-goal (leaving `production_api_base_url`'s default
untouched) was actually preserved in the diff, not silently expanded.

## Deployment and rollback

Authorization boundary: no deployment is authorized by this package. This package's
change only takes effect on the *next* real dispatch of `deploy-production.yml`;
it does not itself deploy or touch the currently-running production host. Whether a
separate, immediate manual correction of the live `/opt/vocanova/production/secrets/api.env`
file is needed before the next scheduled dispatch is an operational decision for
the reviewing human (see `README.md`'s recommended next action 3) — this package
does not perform that correction and has no production access to do so.

Rollout sequence (once authorized): merge to `develop`; the fix takes effect the
next time `deploy-production.yml` is dispatched to production. No staging
equivalent step exists to rehearse this specific change against (staging's own
config-writing step, if one exists with the same host-sharing constraint, is out of
this package's scope — implementer should check whether an equivalent
`deploy-staging.yml` step has the identical defect and flag it as a follow-up if
so, without fixing it here).

Rollback trigger: a future dispatch writes a value that breaks sign-in in a new way
(e.g. a typo in the appended port, or a change that widens the CORS allow-list
beyond production's own host). Rollback mechanism: revert the three-line change (or
redeploy the prior known-good artifact) — the same mechanism already documented for
VOC-039's single-line revert. Owner: named explicitly in the implementation PR at
deploy time, not left implicit.
