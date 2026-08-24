---
id: VOC-###
title: Replace with change title
status: draft
type: feature
risk: R1
owner: replace-with-owner
approved_requirement: replace-with-document-or-decision-id
dependencies: []
---

# VOC-### — Change title

## Business or product objective

State the outcome and why it matters. Link the approved Vocanova requirement or
decision; a chat transcript alone is not approval evidence.

## Problem and desired outcome

Describe the current problem and the observable end state.

## Users affected

Identify affected users, operators, systems, or stakeholders.

## Scope

### In scope

- Required outcome.

### Out of scope

- Explicit exclusion.

## Requirements

Use **must**, **must not**, **should**, and **may** precisely.

### Functional and user-experience requirements

- REQ-01: Requirement.

### Business rules

- Rule.

### Data and API requirements

- Data lifecycle, contracts, compatibility, retention, and migration expectations.

### Security and privacy requirements

- Authentication, authorization, minimization, logging, secrets, and threat controls.

### Accessibility and performance requirements

- Observable standard or threshold.

### Error and edge-case behavior

- Failure behavior and recovery.

## Acceptance criteria

Use stable identifiers and the format in
[acceptance-criteria.md](acceptance-criteria.md).

### AC-01 — Observable outcome

Given a defined starting state
When an action or event occurs
Then an observable result occurs
And required side effects or protections hold

## Impact analysis

Mark each `Affected`, `Not affected`, or `Unknown — blocks readiness`, and explain
every affected or unknown entry.

| Area | Status | Evidence or required work |
|---|---|---|
| Product scope and UX |  |  |
| Living documents and decisions |  |  |
| Frontend and accessibility |  |  |
| Backend and API contracts |  |  |
| Database and migrations |  |  |
| Authentication and authorization |  |  |
| Privacy, personal data, audio, or voice |  |  |
| Security and secrets |  |  |
| Analytics |  |  |
| AI behavior/providers |  |  |
| Infrastructure and deployment |  |  |
| Testing |  |  |
| Support and operations |  |  |

## Implementation plan and tasks

Default to one end-to-end implementation task per package. Task IDs are
minimum-sufficient outcome traceability groupings; they are not component,
file-type, layer, or review-convenience buckets. Every task after the first must
record an explicit allowed split reason when a genuine boundary requires multiple
tasks.

| Task ID | Description | Acceptance criteria | Dependencies | Owner |
|---|---|---|---|---|
| VOC-###-T00 | One end-to-end outcome | AC-01 | None | Codex |

Record the technical approach, components, sequence, compatibility, and known risks.

## Test plan

| Acceptance criterion/risk | Test level or review | Command/evidence |
|---|---|---|
| AC-01 | Unit/integration/journey/manual | To be completed |

Include relevant formatting, lint, types, unit, integration, build, security,
accessibility, migration, preview, staging, and rollback validation. Do not list a
tool as required unless it exists or the implementation task includes installing it.

## Release and rollback plan

- Release class and rationale:
- Preview/staging validation:
- Migration order:
- Rollout and monitoring:
- Rollback trigger:
- Rollback mechanism and owner:
- Last known-good reference:
- User/support communication:
- Production outcome and observation window:

## Risk and approvals

- Declared risk: R#
- Path-detected floor: pending CI
- Protected areas:
- Active governance model: active-A-003 / separately governed rollback reference
- Independent verifier required: Yes
- Standing technical-steward approval required: Yes/No and governing reason
- EHR triggered: Yes/No and evidence
- Founder decision required: Yes/No
- Approval evidence:

Under active A-003, routine R3 does not require standing steward or founder approval
merely because it is R3. Record strengthened technical gates. R4 founder authority
and any triggered EHR remain independently applicable.

## Assumptions and open questions

Material open questions must be resolved before `implementation-ready`.

## Traceability

- Objective:
- Approved requirement/decision:
- Issue/specification:
- Branch/tasks:
- Pull request/commit:
- Tests and verification:
- Preview/release:
- Observed production outcome:
