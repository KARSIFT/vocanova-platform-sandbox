# VOC-117 — Acceptance Criteria

## VOC-117-AC-00 — Authoritative roles.yml stores the six exact bindings

- Requirement source: `VOC-117-D00`, `VOC-117-D01`
- Tasks: `VOC-117-T00`
- Tests: `VOC-117-TEST-00`
- Evidence: `VOC-117-EV-00`
- Result: pending

`KARSIFT/karsift-ai-infra/config/roles.yml` (and the caller mirrored fixture after
pin) contains exactly:

- `implementer: cursor/composer-2.5`
- `implementer_escalation: cursor/composer-2.5`
- `planner: cursor/grok-4.6[fast=false]`
- `reviewer: cursor/grok-4.6[fast=false]`
- `reviewer_fast_retry: cursor/grok-4.6[fast=false]`
- `plan_reviewer: cursor/grok-4.6[effort=high,fast=false]`

No OpenAI/Codex planner or escalation binding is restored as the active mapping.

## VOC-117-AC-01 — Parameterized Cursor model strings are workflow-compatible

- Requirement source: `VOC-117-D02`, `VOC-117-D05`
- Tasks: `VOC-117-T00`
- Tests: `VOC-117-TEST-01`, `VOC-117-TEST-02`
- Evidence: `VOC-117-EV-00`
- Result: pending

Plan, implement, review, and plan-review workflows resolve the stored parameterized
bindings and invoke Cursor without stripping or ignoring the requested `fast=false`
/ `effort=high` semantics, and without silent fallback to another vendor or model.

## VOC-117-AC-02 — Review and plan-review paths use Grok 4.6 Standard through Cursor

- Requirement source: `VOC-117-D00`, `VOC-117-D02`
- Tasks: `VOC-117-T00`
- Tests: `VOC-117-TEST-02`
- Evidence: `VOC-117-EV-00`
- Result: pending

Deterministic workflow tests prove `reviewer`, `reviewer_fast_retry`, and
`plan_reviewer` select the Grok 4.6 Standard (non-fast) Cursor form derived from
their stored bindings, with `plan_reviewer` retaining `effort=high`.

## VOC-117-AC-03 — Missing credentials or unsupported prefixes fail closed

- Requirement source: `VOC-117-D03`
- Tasks: `VOC-117-T00`
- Tests: `VOC-117-TEST-03`
- Evidence: `VOC-117-EV-00`
- Result: pending

Negative tests prove missing `CURSOR_API_KEY` on Cursor-backed paths, and
unsupported provider prefixes, fail closed without fallback to another vendor or
model. Credentials are never printed.

## VOC-117-AC-04 — Current-state docs/comments no longer describe dormant routes as active

- Requirement source: `VOC-117-D06`
- Tasks: `VOC-117-T00`
- Tests: `VOC-117-TEST-04`
- Evidence: `VOC-117-EV-00`
- Result: pending

Updated `roles.yml` / workflow current-state comments (and any caller docs that
assert an active lineup) do not claim dormant OpenAI/Codex routes or obsolete
Cursor bindings are the live configuration after this change.

## VOC-117-AC-05 — Safety gates and caller pin remain intact

- Requirement source: `VOC-117-D04`, `VOC-117-D05`
- Tasks: `VOC-117-T00`
- Tests: `VOC-117-TEST-05`
- Evidence: `VOC-117-EV-00`
- Result: pending

Source self-CI and caller governance/fixture suites pass. Exact-SHA independent
review, risk classification, protected checks, and one-retry limits remain
unchanged. The caller fixture pin equals the exact reviewed shared-infra merge.

Acceptance criteria must be observable, stable, security-aware, and bidirectionally
traceable to requirements, tasks, tests, and evidence.
