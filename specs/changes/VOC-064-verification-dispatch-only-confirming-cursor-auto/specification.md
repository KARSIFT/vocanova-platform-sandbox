# VOC-064 — Verification Dispatch Only: Specification

## Objective and requirement source

Confirm that a free-text planner dispatch bound to the cursor/auto planner
role completes a full draft change package under package ID VOC-064, and that
doing so is usable as a live check of whether that binding avoids
Other-Models quota exhaustion. The originating free-text request is:

> VERIFICATION dispatch only - confirming cursor/auto works and checking
> whether it avoids the Other-Models quota exhaustion. If this drafts a
> package, close/ignore it - not a real change request.

There is no separate approved product requirement, GitHub issue, or DOC-*
decision authorizing application or process change. This package is the
verification artifact the dispatch asked for, and the request's own
disposition instruction is to close or ignore it after drafting succeeds.

## Scope and non-goals

In scope:

- Produce the complete change-package file set under
  `specs/changes/VOC-064-verification-dispatch-only-confirming-cursor-auto`,
  matching the calling project's template shape, with every adoption /
  authorization field left at its unadopted default.
- Record, in this package, that the intended human action is close/ignore —
  not adoption, not implementation, not release.

Non-goals / explicitly excluded:

- No change to `apps/web`, `apps/api`, `packages/`, `.github/workflows/`,
  governance documents, secrets, migrations, staging, or production.
- No investigation, bugfix, feature, documentation amendment, or release
  promotion.
- No adoption, implementation authorization, or task-issue creation.
- Not a standing template for future verification packages to reuse as
  authority for real work.

## Risk and protected areas

Builder assessment (draft proposal only): `R0`. No protected technical or
governance path is intended to change. The only files this package creates
are under its own `specs/changes/VOC-064-...` directory. Path-based risk
classification via `scripts/governance/classify-change-risk.sh` against a
real task file list is not applicable unless someone incorrectly expands
scope later; if that happens, reclassify rather than inheriting this R0
proposal. Active authority model: A-003. EHR: not triggered. No R4
consequence identified for a close/ignore verification artifact.

## Decisions, contradictions, security, and privacy

No `VOC-064-D00`-numbered decision is recorded here because this package is
not adopted and must not be adopted per the originating request. Open
questions for the reviewing human (operational, not product):

1. **Quota-check outcome is outside this package's contents.** Whether the
   cursor/auto run actually avoided Other-Models quota exhaustion is an
   observation for the operator who dispatched the workflow (Actions logs /
   provider quota UI), not something this drafted package can assert from
   inside the working tree. This package only proves the planner completed a
   draft; the quota observation belongs to the human who triggered the run.
2. **Disposition.** The request says close/ignore. If a plan PR is opened
   for this package by the calling workflow, the human should close it
   without adoption rather than merge-as-adopted.

No new secret, credential, personal-data, authorization, or abuse surface is
introduced.

## Data, migrations, analytics, and accessibility

- **Data / migrations:** None. No schema or data change.
- **Analytics:** None. No new events or metrics.
- **Accessibility:** None. No UI change.
