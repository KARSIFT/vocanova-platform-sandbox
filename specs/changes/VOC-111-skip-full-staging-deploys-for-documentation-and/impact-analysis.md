# VOC-111 — Impact Analysis

## Security and privacy

No new secrets, credentials, or personal-data processing. Push selection only changes
**when** the existing deploy-staging workflow is scheduled on `develop`; it does not
change which secrets are read, written, or validated during a selected deploy.

Evidence files and task comments must remain metadata-only (run IDs, SHAs, conclusions,
timestamps). No logs, OAuth material, session cookies, cohort values, or SSH output.

Risk if implemented incorrectly: a too-narrow allowlist could skip deploy for a
runtime-affecting change (fail-closed tests and independent verification are the
control). A too-broad allowlist would retain today's unnecessary deploy cost but would
not weaken runtime gates.

## Data and migrations

None. No database schema, seed, or migration changes. Staging database state is
unchanged by skipped pushes; the last successfully deployed runtime remains live until
a selected push deploys again.

## Analytics and accessibility

None. No product analytics instrumentation or user-facing UI changes.

## Risks, dependencies, and evidence

- `VOC-111-R00`: **Medium operational risk** if runtime paths are omitted from the
  allowlist — staging could serve stale runtime while `develop` advances on application
  changes merged alongside docs in separate commits. Mitigation: explicit allowlist,
  positive fixtures for every runtime class, and independent verification of the real
  diff file list.
- `VOC-111-R01`: **Low operational risk** — skipped non-runtime pushes no longer
  refresh staging immediately after docs-only merges. Acceptable per issue #920; manual
  `workflow_dispatch` remains for explicit redeploy.
- `VOC-111-DEP-00`: Issue #920 sanitized run metadata (`32568473144`, `32568622178`,
  `32572863842`).
- `VOC-111-DEP-01`: Current deploy-staging push trigger and stale near-no-op comment.
- `VOC-111-DEP-02`: VOC-032/VOC-094/VOC-050/VOC-084/VOC-088/VOC-110 predecessor
  packages own adjacent deploy semantics; this package must not weaken them.
- `VOC-111-DEP-03`: Production deploy filtering explicitly out of scope.
- `VOC-111-EV-00`: T00 evidence — issue table, allowlist, deterministic test results.
- `VOC-111-EV-01`: T01 evidence — operator absence metadata for docs/evidence-only push.

## Monitoring

`monitoring_impact.state: none` — no monitor or synthetic inventory changes. Existing
staging synthetics continue to observe the last deployed environment until a
runtime-selected push updates staging.

## Authority and EHR

Active authority model: **A-004**. No founder `approved` comment is required for
engineering-workflow merge/adopt/release gates. EHR not triggered.

This draft proposes **R3**; the path classifier and independent verifier remain
authoritative.
