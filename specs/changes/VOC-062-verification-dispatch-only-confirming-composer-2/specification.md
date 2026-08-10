# VOC-062 — Verification Dispatch Only: Specification

## Objective and requirement source

**Objective:** Verify that the planner role can run to completion when bound to
Cursor's in-house `composer-2.5` model, confirming that model is not blocked by
Other-Models quota exhaustion.

**Requirement source:** Free-text planner dispatch (2026-08-10). Not a GitHub
issue, not an approved product requirement, and not governance authority. The
dispatch text explicitly states this is verification-only and instructs
operators to close or ignore any resulting package.

## Scope and non-goals

**In scope:**

1. This planner run completing successfully with `composer-2.5` as the bound
   planner model.
2. This draft package directory (`specs/changes/VOC-062-verification-dispatch-only-confirming-composer-2/`)
   containing the template's required files, filled with content that honestly
   records the verification purpose and the "do not adopt" instruction.

**Out of scope (explicit non-goals):**

- Any change to application code (`apps/web`, `apps/api`, `packages/`).
- Any change to workflows, governance documents, or repository settings.
- Adoption, implementation authorization, task dispatch, independent
  verification of implementation, release, or production deployment.
- Creating `VOC-062-T##` task entries (see `tasks.md` — zero tasks, deliberately,
  so a mistaken adoption cannot open implementable tracking issues via
  `adopt.yml`).

## Risk and protected areas

- **Proposed risk:** `R0`. No production effect; only this package directory is
  written.
- **Protected areas:** None outside this package directory. Planner scope
  discipline was followed.
- **Authority model:** A-003 active; no R4 founder decision is required because
  no real change is proposed. No EHR trigger.

## Decisions, contradictions, security, and privacy

**VOC-062-D00 — Verification dispatch is not implementation authority**

- The originating request is self-contradictory only in appearance: it asks the
  planner to run (which normally drafts a package) while also saying "if this
  drafts a package, close/ignore it." The resolution is: draft the package as
  evidence that the planner/model path works, but mark it permanently inert and
  include zero implementable tasks. Operators must not treat this directory as
  authority for any product or governance change.

**Security and privacy:** None. No secrets, credentials, personal data, or
production access are involved.

## Data, migrations, analytics, and accessibility

None. This package proposes no data, schema, analytics, or accessibility
changes.

## Open questions

None. The verification dispatch is fully specified by its own text; the only
outcome is "planner completed with composer-2.5" or "it did not."
