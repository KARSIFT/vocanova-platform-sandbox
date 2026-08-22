# VOC-107 — Isolated implementer publisher cannot verify incremental remediation bundles: Specification

## Objective and requirement source

Make the implementer recovery/publish bundle carry the complete task-branch-only
lineage the clean isolated publisher needs, anchored to the reviewed integration
lineage, so remediation commits verify and import without fetching the prior
task-PR head — as recorded in
[GitHub issue #891](https://github.com/KARSIFT/vocanova-platform-sandbox/issues/891).

Today `implement.yml` creates recovery/publish bundles from the implementation
step’s immediate `base_sha..HEAD`. On remediation attempts, `base_sha` can be the
existing PR head or a locally rebased derivative, which is not necessarily
reachable from the integration branch fetched by the isolated publisher.
`git bundle verify` correctly refuses the thin bundle when that prerequisite
object is absent.

**This draft package does not adopt or authorize itself**; adoption remains a
separate A-004 plan-review / adopt path.

Drafting-time grounding:

| Item                 | Current state                                                                                                                                 |
| -------------------- | --------------------------------------------------------------------------------------------------------------------------------------------- |
| Bundle create        | `git bundle create … "${{ steps.branch.outputs.base_sha }}..HEAD"`                                                                            |
| Attempt-2 `base_sha` | Set to `HEAD` after checkout/rebase of existing `agent/…` branch — often not on the integration tip the publisher fetches                     |
| Soft-reset commit    | `git reset --soft` to `base_sha` so only this attempt’s model commits are collapsed — must keep that tip, not the integration tip             |
| Isolated publisher   | Fresh bare repo; fetches only `integration_branch`; `bundle verify`; exact head SHA; ancestry vs `PUBLISH_BASE_SHA`; workflow-path deny; FWL |
| Incident             | Run `32539352323`: valid committed remediation rejected before publication; manual recovery required                                          |
| Attempt policy       | Two-attempt cap; must not burn a model rerun solely because a valid committed bundle lacked a publisher prerequisite                         |

## Scope and non-goals

In scope:

1. Change implementer bundle creation so the artifact contains the complete
   task-branch-only lineage from the reviewed integration tip to `HEAD`, not
   only the last incremental delta from the remediating `base_sha`.
2. Align publisher validation (`PUBLISH_BASE_SHA` / integration ancestry /
   workflow-path deny scanning range) with that same integration anchor so the
   deny rule covers the full published lineage relative to integration.
3. Preserve the isolated clean publisher, exact published SHA check, integration
   ancestry check, workflow-path deny rule, and SHA-valued force-with-lease.
4. Preserve the two-attempt cap. A valid committed bundle that previously failed
   only for a missing publisher prerequisite must not force a model rerun by
   itself; the fix is lineage completeness, not another implementer attempt.
5. Keep soft-reset commit semantics on the pre-model tip so attempt-2 does not
   squash prior task commits into the new attempt.
6. Cover rebase-derived remediation lineage so a locally recreated base cannot
   strand publication.
7. Add a deterministic end-to-end Git fixture: integration, attempt-1, and
   attempt-2 commits; prove attempt-2 verifies/imports in a clean bare repository
   containing only integration; include a negative malformed/stale lineage case.
8. Update infra README (and calling-repo docs only if current wording would
   become false) so operators understand integration-anchored bundles.
9. Keep logs and artifacts out of issues and review evidence.

Non-goals / explicitly excluded:

- Node runtime deprecations, Go cache warnings, dependency updates,
  OAuth/application behavior, or deployment changes (issue #891 root focus).
- Changing `plan.yml` planner-bundle behavior (related thin-bundle pattern;
  explicit follow-up if observed there).
- Weakening publisher isolation, exact-SHA checks, workflow deny, or
  force-with-lease.
- Expanding the two-attempt cap or inventing a third model attempt.
- Application, migration, signup-policy, secrets, database, or
  `infra/monitoring/` inventory ID changes.
- Self-adoption / self-authorization of this package.

## Risk and protected areas

- **Draft package proposal:** **R3** (CI/CD / implementer publication lineage).
- **Measured path floor at drafting:** **R3** for `.github/workflows/` and
  related governance automation. Not proposed as R4; no authority-model or
  amendment docs. Calling-repo foundation-test or doc touches may raise the path
  floor at implementation time.
- Protected areas: isolated clean publisher; exact published SHA; integration
  ancestry; `.github/workflows/**` deny before push; SHA-valued
  force-with-lease; two-attempt implementer/remediation cap; least-privilege
  App token only on the clean publish runner.
- EHR: not triggered.
- Active authority model: **A-004**. No founder `approved` comment is a
  merge/adopt/release gate.

The R3 value is a **draft proposal for the reviewing human at adoption time,
never a determination**. The path-based classifier and independent verifier
govern each task PR.

## Decisions

`VOC-107-D00`: The implementer recovery/publish bundle MUST contain the complete
task-branch-only commit lineage from the reviewed integration tip present when
the implementation branch was prepared (or an equivalent explicitly recorded
integration-anchor SHA that is an ancestor of `HEAD` and present on the
publisher-fetched integration ref) through `HEAD`. It MUST NOT be limited to the
last incremental `pre_model_tip..HEAD` delta when that tip is not guaranteed to
exist in the publisher’s clean bare repository.

`VOC-107-D01`: The soft-reset commit step MUST continue to reset to the
pre-model tip (today’s `steps.branch.outputs.base_sha` after branch
create/rebase). That tip remains the correct boundary for collapsing only the
current attempt’s model commits. Bundle creation and publish-base ancestry MUST
use a separate integration-anchor output (for example `integration_sha` /
`bundle_base_sha`), not overload the soft-reset tip.

`VOC-107-D02`: The isolated clean publisher MUST remain a separate job that:
downloads only the committed Git bundle; initializes a fresh bare repository;
fetches only the integration branch before `bundle verify`; imports the exact
declared head; verifies exact published SHA; verifies integration ancestry of
the published head against the integration-anchor / `PUBLISH_BASE_SHA` used for
publication; rejects any `.github/workflows/**` path change in the published
lineage relative to that integration anchor; and pushes with SHA-valued
force-with-lease. No model-controlled runner may receive the App token.

`VOC-107-D03`: The two-attempt implementer/remediation cap is unchanged. This
package MUST NOT authorize a third model attempt, and MUST NOT treat “valid
committed work whose thin bundle lacked a publisher prerequisite” as a reason to
rerun the model. After the lineage fix, that class of failure MUST be avoided by
construction for integration-anchored remediation bundles.

`VOC-107-D04`: Rebase-derived remediation lineage is in scope. After a successful
`git rebase` onto integration (new commit SHAs for prior attempt commits), the
bundle MUST still verify and import in a clean bare repository that contains only
the integration tip. A locally recreated base MUST NOT strand publication.

`VOC-107-D05`: Deterministic end-to-end Git fixture coverage MUST include at
least:

1. **Positive:** create integration commit I, attempt-1 commit A1 on top of I,
   rebase/recreate lineage as needed, attempt-2 commit A2; build the
   integration-anchored bundle for A2; in a clean bare repo that has fetched only
   I’s integration ref, `bundle verify` and import succeed; imported head equals
   A2; ancestry and workflow-deny scan use the integration anchor.
2. **Negative:** a thin/stale/malformed bundle whose prerequisite is a task-PR
   head (or other object) absent from that clean integration-only bare repo MUST
   fail closed at verify/import (matching today’s correct refusal behavior for
   incomplete lineage).
3. **Regression:** attempt-1 fresh-from-integration bundles continue to publish
   under the same publisher guards.

`VOC-107-D06`: Issues, package evidence, and review records for this work MUST
use allowlisted metadata only (run IDs, SHAs, boolean outcomes, scrubbed reason
codes). Forbidden: logs, artifacts, secrets, OAuth/session/cookie/token material,
user identifiers.

`VOC-107-D07`: Keep root scope focused. Node runtime deprecations, Go cache
warnings, dependency updates, OAuth/application behavior, and deployment changes
are out of scope. Planner (`plan.yml`) thin-bundle behavior is out of scope
unless implementation discovers an identical live failure class that must be
recorded as a separate follow-up issue rather than silently expanded here.

## Open questions

- Exact helper placement (inline workflow shell vs extracted script under
  `karsift-ai-infra/config/`) is an implementer choice inside `VOC-107-D00`–
  `D02`, provided the fixture in `VOC-107-D05` can exercise the same lineage
  rules deterministically.
- Whether calling-repo DOC-15 / operator docs need wording updates depends on
  whether current text claims thin `base_sha..HEAD` bundles are always
  publisher-sufficient; update only if that claim would become false.

## Data, migrations, analytics, and accessibility

- Data / migrations: None — evidence-backed non-applicability.
- Analytics: None — evidence-backed non-applicability.
- Accessibility: None — evidence-backed non-applicability (no product UI).

## Security and privacy

- No new secrets. No broadening of implementer token scopes.
- App credentials remain only on the clean publish runner.
- Evidence is allowlisted metadata only (`VOC-107-D06`).
- Workflow-path deny and force-with-lease remain mandatory publisher gates.
