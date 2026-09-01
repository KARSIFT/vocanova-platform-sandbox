# VOC-145 — Direct role-binding and regression-expectation drift lacks governed reconciliation

| Field | Value |
|-------|-------|
| Package | `VOC-145` |
| Title | Direct role-binding and regression-expectation drift lacks governed reconciliation |
| Path | `specs/changes/VOC-145-direct-role-binding-and-regression-expectation` |
| Status | `draft` |
| Risk | `R4` (draft proposal; active reviewer/plan-reviewer model, effort, and speed bindings plus regression-expectation authority) |
| Authority model | A-004 active |
| Requirement source | GitHub issue [#1124](https://github.com/KARSIFT/vocanova-platform-sandbox/issues/1124) |
| Target branch | `develop` |
| Approval | `not-approved` |
| Implementation authorized | `false` |
| `automatic_merge_allowed` | `true` (per AGENTS.md A-004 drafting rule) |

## Problem

`KARSIFT/karsift-ai-infra` `main` directly changed the active reviewer
bindings and then changed `tests/test_voc117_role_bindings.py` to accept
them, without a matching adopted caller change package or reconciliation of
every canonical document that fixes the role mapping.

| Item | Value |
|------|-------|
| Last governed VOC-142 infrastructure base | `8993e867640dfb604dec0466c4e0787e68d8e258` |
| Current unauthorized infra head | `d8720829b176cf1287e633f9382989fc8f258105` |
| Infra self-CI run claimed green | `33443684483` |
| Caller pin at issue creation | `8993e867640dfb604dec0466c4e0787e68d8e258` |

Observed drift from `8993e867…` to `origin/main`:

| Role | Last adopted (VOC-142 DEP-06 / AC-09) | Ungoverned infra `main` |
|------|----------------------------------------|-------------------------|
| `reviewer` | `cursor/grok-4.6[effort=high,fast=false]` | `cursor/grok-4.6[effort=xhigh,fast=false]` |
| `reviewer_fast_retry` | `cursor/grok-4.6[effort=high,fast=false]` | `cursor/grok-4.6[effort=xhigh,fast=true]` |
| `plan_reviewer` | `cursor/grok-4.6[effort=high,fast=false]` | `cursor/grok-4.6[effort=xhigh,fast=false]` |

Unchanged and still last-adopted: `implementer` / `implementer_escalation`
`cursor/composer-2.5`; `planner` `cursor/grok-4.6[effort=high,fast=false]`.

The VOC-117 regression expectations were rewritten to bless the new values.
`README.md`, `CHANGELOG.md`, caller fixtures/pin, and adopted VOC-142
DEP-06 / AC-09 current-state evidence were not reconciled. A green test that
was changed in the same direct sequence does not establish governed
authority for the role transition.

## Root cause

Role, model, speed, and effort changes are live agent-authority controls.
The last adopted current-state contract is VOC-142 DEP-06 / AC-09 at pin
`8993e867…`: planner, reviewer, `reviewer_fast_retry`, and `plan_reviewer`
are all `cursor/grok-4.6[effort=high,fast=false]`. Infra `main` mutated
those three review bindings and the VOC-117 current-expectation tests
outside an adopted caller package. Historical assertions were weakened to
match the new values instead of remaining historical. Current-state
documents that still describe the VOC-117 / VOC-142 lineup were left
stale.

## Required outcome (summary)

Use one largest-safe coherent task and one caller implementation PR,
coordinated with one infrastructure PR:

1. Reconcile live `config/roles.yml`, VOC-117 current-state tests, caller
   fixtures/pin, README, CHANGELOG, and every other current-state contract
   to one authorized binding set.
2. **Default (Path A):** restore the last adopted exact bindings from
   `8993e867…` / VOC-142 DEP-06. Restore VOC-117 regression expectations
   that were rewritten to bless the ungoverned values.
3. **Adoption-only alternative (Path B):** formally adopt the drifted
   `xhigh` review lineup as the new current-state contract, without
   rewriting historical VOC-117 assertions to pretend that lineup was
   always required.
4. Do not pin unauthorized head `d8720829…` as-is. Do not weaken retry
   caps, exact-SHA review, provider isolation, or fail-closed model
   resolution. Do not use this package as a carrier for the VOC-112
   provenance EHR escalation in issue [#1120](https://github.com/KARSIFT/vocanova-platform-sandbox/issues/1120).
5. Pin the caller fixture to the new independently reviewed infrastructure
   merge. Do not snapshot the develop/main gap (`karsift-ai-infra#15`).

This is a KARSIFT automation authority reconciliation, not product
behavior. Preserve A-004 risk classification, protected checks, review
independence, and release gates.

## Tasks

| Task | Summary | Depends on |
|------|---------|------------|
| T00 | Governed reconciliation of the unauthorized role-binding and VOC-117 expectation drift | — |

See `tasks.md` for the full task definition.

## Verification, approvals, release, and closure

See `test-plan.md`, `implementation-plan.md`, and `release-plan.md`.
Under **active A-004**, engineering-workflow gates do not wait on a founder
`approved` comment, but this draft itself carries no adoption or
implementation authority.

## Risk note

This package **proposes R4** because durable `tooling/governance/` fixture
and pin updates belong under the R4 path floor, and because the change
mutates live agent-authority model, effort, and speed bindings used by
independent review. The path-based classifier and independent verifier
remain authoritative; this draft proposal is not a determination.
