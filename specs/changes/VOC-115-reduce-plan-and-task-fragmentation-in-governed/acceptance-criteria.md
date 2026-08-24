# VOC-115 — Acceptance Criteria

## VOC-115-AC-00 — Ordinary coherent work drafts one package with one task by default

- Requirement source: `VOC-115-D00`, `VOC-115-D02`, `VOC-115-D10`, `VOC-115-D11`
- Tasks: `VOC-115-T00`
- Tests: `VOC-115-TEST-00`, `VOC-115-TEST-01`
- Evidence: `VOC-115-EV-00`
- Result: pending

Planner prompt and mirrored fixture coverage prove that a normal coherent feature or
bug request drafts one change package whose `tasks.md` contains exactly one end-to-end
implementation task by default and targets one outcome-sized implementation PR per
repository unless a recorded boundary genuinely requires otherwise.

## VOC-115-AC-01 — Code, tests, docs, and same-carrier evidence stay together by default

- Requirement source: `VOC-115-D02`, `VOC-115-D07`, `VOC-115-D10`, `VOC-115-D11`
- Tasks: `VOC-115-T00`
- Tests: `VOC-115-TEST-01`, `VOC-115-TEST-02`
- Evidence: `VOC-115-EV-00`
- Result: pending

Drafted packages and validation rules do not split code vs tests, docs vs code, or
evidence-only follow-ups into separate tasks unless a concrete boundary from
`VOC-115-D03` is recorded.

## VOC-115-AC-02 — Additional tasks require explicit split reasons

- Requirement source: `VOC-115-D03`, `VOC-115-D04`, `VOC-115-D05`, `VOC-115-D10`
- Tasks: `VOC-115-T00`
- Tests: `VOC-115-TEST-03`, `VOC-115-TEST-04`
- Evidence: `VOC-115-EV-00`
- Result: pending

Deterministic validation and/or fixture tests fail closed when a package defines a
second or later task without an explicit allowed split reason. Packages with more
than three tasks are flagged as exceptional and require package-level justification
for why consolidation is unsafe. A size label or changed-line threshold alone is not
an allowed split reason.

## VOC-115-AC-03 — Justified multi-task packages still preserve sequential advancement

- Requirement source: `VOC-115-D03`, `VOC-115-D08`
- Tasks: `VOC-115-T00`
- Tests: `VOC-115-TEST-05`
- Evidence: `VOC-115-EV-00`
- Result: pending

A package that includes a valid split reason and real multi-task boundary still
produces ordered task issues / dependency edges correctly and preserves existing
sequential auto-advance behavior.

## VOC-115-AC-04 — In-scope causal failures may remain under the active package

- Requirement source: `VOC-115-D00`, `VOC-115-D01`, `VOC-115-D06`
- Tasks: `VOC-115-T00`
- Tests: `VOC-115-TEST-06`, `VOC-115-TEST-07`
- Evidence: `VOC-115-EV-00`
- Result: pending

Canonical governance/docs and regression coverage explicitly distinguish:

1. in-scope causal remediation that remains under the active package/carrier; and
2. unrelated or authority-expanding work that still requires a new issue/plan.

The allowed in-scope path must remain bounded by original objective, acceptance
criteria, risk ceiling, and protected-area scope.

## VOC-115-AC-05 — Existing safety gates remain unchanged

- Requirement source: `VOC-115-D06`, `VOC-115-D09`
- Tasks: `VOC-115-T00`
- Tests: `VOC-115-TEST-08`
- Evidence: `VOC-115-EV-00`
- Result: pending

Updated docs, prompts, validation, and mirrored fixtures preserve exact-SHA
independent verification, deterministic risk floors, protected-branch checks, and
fail-closed behavior. No acceptance criterion claims reduced review rigor.

## VOC-115-AC-06 — Related multi-skill work remains one traceability grouping

- Requirement source: `VOC-115-D02`, `VOC-115-D11`
- Tasks: `VOC-115-T00`
- Tests: `VOC-115-TEST-09`
- Evidence: `VOC-115-EV-00`
- Result: pending

A request to add several related agent skills, including their configuration,
documentation, adapters, and tests, produces one plan and one end-to-end task. The
number of skills, components, directories, or files does not create task boundaries.

Acceptance criteria must be observable, stable, security-aware, and bidirectionally
traceable to requirements, tasks, tests, and evidence.
