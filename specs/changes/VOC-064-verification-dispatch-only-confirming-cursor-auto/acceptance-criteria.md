# VOC-064 — Acceptance Criteria

## VOC-064-AC-00 — A complete draft package exists under the VOC-064 canonical path with all adoption/authorization fields at unadopted defaults

- Requirement source: free-text verification dispatch; `specification.md`
  scope item 1
- Tasks: `VOC-064-T00`
- Tests: `VOC-064-TEST-00`
- Evidence: `VOC-064-EV-00`
- Result: pending

Observable outcome: this directory contains the template-required artifacts
(`README.md`, `change.yaml`, `specification.md`, `acceptance-criteria.md`,
`impact-analysis.md`, `implementation-plan.md`, `tasks.md`, `test-plan.md`,
`release-plan.md`), and `change.yaml` still has `status: draft`,
`approval_status: not-approved`, `implementation_authorized: false`, and
`implementation.authorized: false`.

## VOC-064-AC-01 — The package explicitly records close/ignore disposition and no product/process change scope

- Requirement source: free-text verification dispatch ("close/ignore it -
  not a real change request"); `specification.md` scope item 2
- Tasks: `VOC-064-T00`
- Tests: `VOC-064-TEST-01`
- Evidence: `VOC-064-EV-00`
- Result: pending

Observable outcome: `README.md`, `specification.md`, and `change.yaml`'s
`blocking_reasons` state that this is verification-only and must not be
adopted; `affected_areas` lists only this package directory; `release.required`
and `implementation.required` are false.

Acceptance criteria must remain unsatisfied as "implementer work" — they are
satisfied by the draft itself existing, then by the human closing/ignoring
the package without adoption.
