# VOC-041 — Tasks

None of the tasks below is implementation-authorized by this package. Adoption and
each task's own implementation-authorization are separate, mirroring
VOC-039/VOC-040's convention. `T00` is the actual fix and should land first;
`T01`'s regression check depends on `T00`'s fixed values existing to assert against
(or may be written failing-first against the pre-fix values, at the implementer's
discretion).

## VOC-041-T00 — Port-qualify BASE_URL/OAUTH_REDIRECT_URI/OAUTH_REDIRECT_ALLOWLIST and correct the step's comment

- Requirement source: issue #312's confirmed-by-reproduction root cause and
  suggested fix
- Acceptance criteria: `VOC-041-AC-00`, `VOC-041-AC-01`
- Status: pending
- Summary: in `.github/workflows/deploy-production.yml`'s "Write production
  application configuration" step, append `:8443` to the host in the `BASE_URL`,
  `OAUTH_REDIRECT_URI`, and `OAUTH_REDIRECT_ALLOWLIST` lines exactly as specified in
  `implementation-plan.md` step 1. Leave `SESSION_COOKIE_DOMAIN` unchanged. Replace
  the step's comment block (currently asserting the disproven Cloudflare
  port-forwarding claim) with one recording the issue's live disproof, consistent
  with the sibling health-check step's existing comment. Verify by rendering the
  step's script locally against representative host values and confirming the
  three written lines match `VOC-041-AC-00`'s exact expected strings.

## VOC-041-T01 — Deterministic regression check for the port-qualification defect class

- Requirement source: issue #312's implicit ask (this defect class should be
  catchable without a live production reproduction)
- Acceptance criteria: `VOC-041-AC-02`
- Status: pending
- Depends on: `VOC-041-T00` (or may be written first as a failing-first check
  against pre-fix values, at the implementer's discretion)
- Summary: add a deterministic check (automated test or documented script, per
  `implementation-plan.md` step 2) that fails if any of the three values in this
  step is written without `:8443`, and passes against the post-fix step. Runs via
  `pnpm validate` or a narrower documented script.
