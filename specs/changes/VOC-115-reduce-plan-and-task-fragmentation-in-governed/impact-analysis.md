# VOC-115 — Impact Analysis

## Security and privacy

This package changes governance and workflow policy, not application runtime or
stored data. No new secrets, OAuth/session material, production data access, or
user-facing data flows are introduced.

The main security/process risk is accidental scope broadening: if the new
"stay with the active package" policy were worded too loosely, unrelated work or
higher-risk protected changes could bypass a required new plan. Mitigations:

- explicit bounded criteria (`objective`, `acceptance criteria`, `risk ceiling`,
  `protected-area scope`);
- deterministic tests for allowed vs disallowed continuation cases;
- preserved fail-closed behavior when classification is ambiguous.

## Data and migrations

None. No database, schema, seed, or migration changes are proposed.

## Analytics and accessibility

None for product/runtime behavior. There is no direct analytics instrumentation or
user-interface accessibility effect.

## Risks, dependencies, and evidence

- `VOC-115-R00`: **High governance risk** if consolidation wording accidentally
  weakens scope boundaries or allows authority expansion inside an active package.
  Mitigation: explicit new-plan triggers and negative regression tests.
- `VOC-115-R01`: **High workflow risk** if split-reason enforcement breaks justified
  multi-task packages or auto-advance sequencing. Mitigation: regression coverage
  for valid multi-task cases and dependency order preservation.
- `VOC-115-R02`: **Medium documentation risk** if docs, prompts, templates, and
  fixtures are not updated together, leaving contradictory instructions. Mitigation:
  same-package synchronization requirement (`VOC-115-D09`).
- `VOC-115-R03`: **Medium operational risk** if evidence-only or docs-only task
  splitting remains implicitly encouraged by old planner text or validator gaps.
  Mitigation: one-task default in both prose and deterministic fixture coverage.
- `VOC-115-R04`: **Low release risk** because no runtime deployment change is
  intended; rollback is documentation/prompt/validator reversion.
- Protected surfaces: `AGENTS.md`, `DOC-15`, `DOC-16`, package templates,
  `tooling/governance/`, mirrored `karsift-ai-infra` planner/plan-review/adopt
  fixtures, and task-sequencing tests.
- `VOC-115-DEP-00`: issue #962 provides the change objective and acceptance evidence.
- `VOC-115-DEP-01`: A-004 is active; founder comment gates stay removed, while
  exact-SHA independent verification and fail-closed enforcement remain mandatory.
- `VOC-115-DEP-02`: current planner/adopt behavior still encodes fragmentation.
- `VOC-115-DEP-03`: canonical docs still need reconciliation with the new default.
- `VOC-115-EV-00`: T00 evidence — changed policy text, file list, validation results,
  and regression tests demonstrating one-task default plus justified multi-task
  preservation.

## Monitoring

`monitoring_impact.state: none` — no monitor or synthetic inventory changes.

## Authority and EHR

Active authority model: **A-004**. No founder `approved` comment is required for
engineering-workflow adopt/merge/release gates. EHR is not triggered.

This draft proposes **R4** because it changes governance and workflow behavior, but
the path classifier and independent verifier remain authoritative.
