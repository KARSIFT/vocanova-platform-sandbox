# VOC-107 — Acceptance Criteria

## VOC-107-AC-00 — Bundle carries complete integration-anchored task lineage

- Requirement source: `VOC-107-D00`, `VOC-107-D01`
- Tasks: `VOC-107-T00`
- Tests: `VOC-107-TEST-00`, `VOC-107-TEST-01`, `VOC-107-TEST-04`
- Evidence: `VOC-107-EV-00`
- Result: pending

When the implementer produces committed changes, the recovery/publish bundle
contains every task-branch-only commit from the recorded integration-anchor SHA
through `HEAD`. Soft-reset for committing continues to use the pre-model tip and
does not squash prior task commits into the current attempt.

## VOC-107-AC-01 — Clean publisher verifies and imports remediation bundles

- Requirement source: `VOC-107-D00`, `VOC-107-D02`, `VOC-107-D04`
- Tasks: `VOC-107-T00`
- Tests: `VOC-107-TEST-01`, `VOC-107-TEST-02`, `VOC-107-TEST-04`
- Evidence: `VOC-107-EV-00`
- Result: pending

In a clean bare repository that has fetched only the integration branch, an
attempt-2 (including rebase-derived) integration-anchored bundle verifies and
imports successfully. Imported head equals the declared publish head. Publication
does not require fetching the prior task-PR head.

## VOC-107-AC-02 — Publisher security and exactness controls preserved

- Requirement source: `VOC-107-D02`
- Tasks: `VOC-107-T00`
- Tests: `VOC-107-TEST-03`, `VOC-107-TEST-05`
- Evidence: `VOC-107-EV-00`
- Result: pending

The isolated clean publisher, exact published SHA check, integration ancestry
check (against the integration anchor), workflow-path deny rule over the full
published lineage relative to that anchor, and SHA-valued force-with-lease all
remain enforced. The App token stays off the model-controlled runner.

## VOC-107-AC-03 — Two-attempt cap preserved; no model rerun for prerequisite miss

- Requirement source: `VOC-107-D03`
- Tasks: `VOC-107-T00`
- Tests: `VOC-107-TEST-06`
- Evidence: `VOC-107-EV-00`
- Result: pending

The implementer/remediation attempt cap remains two. The package does not add a
third model attempt and does not define “thin bundle missing publisher
prerequisite” as a reason to rerun the model. Integration-anchored bundles avoid
that failure class by construction for the covered remediation lineage.

## VOC-107-AC-04 — Deterministic Git fixture matrix landed

- Requirement source: `VOC-107-D05`
- Tasks: `VOC-107-T00`
- Tests: `VOC-107-TEST-00` through `VOC-107-TEST-05`
- Evidence: `VOC-107-EV-00`
- Result: pending

Positive attempt-2 / rebase-derived import, negative malformed/stale incomplete
lineage, and attempt-1 regression fixtures exist and pass in CI or infra self-ci.

## VOC-107-AC-05 — Docs describe integration-anchored implementer bundles

- Requirement source: AGENTS.md doc-consistency rule; `VOC-107-D00`, `VOC-107-D02`
- Tasks: `VOC-107-T00`
- Tests: `VOC-107-TEST-07`
- Evidence: `VOC-107-EV-00`
- Result: pending

Infra README (and calling-repo docs only if otherwise false) accurately state that
implementer publish bundles are integration-anchored complete task lineages for
the clean publisher. No doc claims a remediating thin `pre_model_tip..HEAD`
bundle is always publisher-sufficient.

## VOC-107-AC-06 — Evidence stays metadata-only

- Requirement source: `VOC-107-D06`
- Tasks: `VOC-107-T00`
- Tests: `VOC-107-TEST-08`
- Evidence: `VOC-107-EV-00`
- Result: pending

Task evidence and related issue/review records use allowlisted metadata only.
No logs, artifacts, secrets, or user identifiers are copied into the package or
issue thread as acceptance proof.
