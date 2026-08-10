# VOC-064 — Test Plan

## VOC-064-TEST-00 — Draft package completeness and unadopted defaults

- Covers: `VOC-064-AC-00`
- Preconditions: working tree or plan-branch tip containing this package
  directory
- Procedure: list
  `specs/changes/VOC-064-verification-dispatch-only-confirming-cursor-auto/`
  and confirm the nine template artifacts are present; read `change.yaml`
  and confirm `status: draft`, `approval_status: not-approved`,
  `implementation_authorized: false`, and `implementation.authorized: false`
- Expected result: all artifacts present; all adoption/authorization fields
  remain at unadopted defaults
- Evidence: `VOC-064-EV-00`

## VOC-064-TEST-01 — Close/ignore disposition and empty product scope

- Covers: `VOC-064-AC-01`
- Preconditions: same as TEST-00
- Procedure: read `README.md`, `specification.md`, and `change.yaml`
  (`blocking_reasons`, `affected_areas`, `implementation.required`,
  `release.required`); confirm explicit close/ignore language and that
  affected areas are limited to this package directory with
  implementation/release not required
- Expected result: disposition and non-goals match the free-text request;
  no product/process change is proposed for implementation
- Evidence: `VOC-064-EV-00`

No positive/negative application, authorization, migration, accessibility,
or rollback tests apply — there is no product change under test. Do not use
secrets or production data.
