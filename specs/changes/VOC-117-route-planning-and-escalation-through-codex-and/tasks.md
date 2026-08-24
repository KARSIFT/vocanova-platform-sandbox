# VOC-117 — Tasks

None of the tasks below is implementation-authorized by this package.
Adoption and each task's own implementation authorization are separate.

This package intentionally defaults to **one task** because issue #978 (as
superseded) requires one coherent model-routing outcome covering configuration,
workflow compatibility, tests, documentation/comments, caller fixture/pin, and
validation together.

## VOC-117-T00 — Apply the Cursor role lineup, parameterized-model routing, tests, docs, and caller pin

- Requirement source: issue #978 superseding comment; `VOC-117-D00` through `VOC-117-D06`
- Acceptance criteria: `VOC-117-AC-00` through `VOC-117-AC-05`
- Tests: `VOC-117-TEST-00` through `VOC-117-TEST-05`
- Evidence: `VOC-117-EV-00` (`t00-evidence.md` in this package directory)
- Status: pending

### Required work

1. Update authoritative `KARSIFT/karsift-ai-infra/config/roles.yml` to exactly:
   - `implementer: cursor/composer-2.5`
   - `implementer_escalation: cursor/composer-2.5`
   - `planner: cursor/grok-4.6[fast=false]`
   - `reviewer: cursor/grok-4.6[fast=false]`
   - `reviewer_fast_retry: cursor/grok-4.6[fast=false]`
   - `plan_reviewer: cursor/grok-4.6[effort=high,fast=false]`
2. Make `plan.yml`, `implement.yml`, `review.yml`, and `plan-review.yml` compatible
   with parameterized Cursor model strings so the requested Standard/non-fast and
   plan-reviewer `effort=high` semantics are preserved after vendor-prefix handling.
3. Keep authentication fail-closed for Cursor paths (`CURSOR_API_KEY`); never print
   credentials; do not require or re-enable an OpenAI/Codex execution path.
4. Update current-state comments/docs so dormant historical OpenAI/Codex or obsolete
   Cursor bindings are not described as active.
5. Add/extend deterministic tests asserting the six exact mappings, parameterized
   routing for planner/review/plan-review paths, and negative fail-closed cases for
   missing credentials or unsupported prefixes.
6. Land the primary infra change through one reviewed infra PR, then sync and pin
   the caller `tooling/governance/fixtures/karsift-ai-infra/` mirror to that exact
   merge; update caller governance tests in the same task.
7. Run infra and caller validation suites and record results in `t00-evidence.md`,
   including the exact reviewed infra SHA used for the caller pin.
8. Preserve independent exact-SHA review, risk classification, protected checks, and
   one-retry limits.

### Explicitly out of scope for this task

- Product runtime, permission, production-data, deploy-credential, or monitor-inventory
  changes.
- Restoring OpenAI/Codex planner or escalation routing from the superseded issue body.
- Inventing a different escalation or review model than `VOC-117-D00`.
- Weakening fail-closed credential checks, exact-SHA review, risk floors, or retry caps.
- Splitting config, workflows, tests, docs, or fixture pin into separate tasks.

## Task ordering notes

- This package intentionally has one task because no concrete split boundary is
  required: coordinated source and caller PRs remain one outcome.
- If implementation discovers that parameterized CLI support cannot honor the stored
  bindings without changing the requested model IDs, stop and record the blocker in
  evidence rather than silently substituting another vendor/model.

Tasks preserve scope, separation of duties, and rollback safety.
