# VOC-062 — Test Plan

## VOC-062-TEST-00 — Planner workflow completes with composer-2.5 bound

- Covers: `VOC-062-AC-00`
- Preconditions: `composer-2.5` is the active `planner` model binding in
  `karsift-ai-infra`'s `config/roles.yml`; a planner `workflow_dispatch` was
  triggered with this package's ID and path.
- Procedure:
  1. Inspect the planner workflow run metadata and logs.
  2. Confirm the run selected `composer-2.5` (not an Other-Models quota
     class that was exhausted).
  3. Confirm the run reached the package-drafting step and completed without
     model/quota errors.
  4. Confirm
     `specs/changes/VOC-062-verification-dispatch-only-confirming-composer-2/`
     exists with all required template files.
  5. Confirm `change.yaml` leaves `status: draft`, `approval_status:
     not-approved`, and `implementation_authorized: false`.
- Expected result: Run status success; package directory present; no adoption
  fields flipped; no files modified outside this package directory.
- Evidence: `VOC-062-EV-00` (workflow run URL and/or commit SHA recorded by
  the operator)

No application tests, governance scripts, or integration checks apply.
