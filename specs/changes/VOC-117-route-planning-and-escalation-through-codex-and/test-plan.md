# VOC-117 — Test Plan

## VOC-117-TEST-00 — Role-resolution asserts all six exact mappings

- Covers: `VOC-117-AC-00`
- Preconditions: updated `config/roles.yml` in primary infra (and mirrored fixture after pin)
- Procedure: Deterministic role-resolution / roles.yml parse tests assert the six
  active bindings equal exactly the values in `VOC-117-D00`.
- Expected result: all six match; no active OpenAI/Codex planner or escalation binding.
- Evidence: `VOC-117-EV-00`

## VOC-117-TEST-01 — Planner parameterized string is Cursor-compatible

- Covers: `VOC-117-AC-01`
- Preconditions: updated `plan.yml` (and any shared model-parse helper)
- Procedure: Workflow/fixture tests resolve `planner` and assert the Cursor invocation
  receives the Standard/non-fast form derived from
  `cursor/grok-4.6[effort=high,fast=false]`
  without silent vendor/model fallback.
- Expected result: planner routes through Cursor with exact CLI model
  `grok-4.6[effort=high,fast=false]`, not through OpenAI/Codex; the
  effort-omitted form fails closed.
- Evidence: `VOC-117-EV-00`

## VOC-117-TEST-02 — Reviewer, fast retry, and plan reviewer use Grok 4.6 Standard

- Covers: `VOC-117-AC-01`, `VOC-117-AC-02`
- Preconditions: updated `review.yml` and `plan-review.yml`
- Procedure: Workflow/fixture tests resolve `reviewer`, `reviewer_fast_retry`, and
  `plan_reviewer` and assert Cursor invocation uses the explicit-high
  Standard/non-fast Grok 4.6 form.
- Expected result: all three review-side roles use exact CLI model
  `grok-4.6[effort=high,fast=false]`; no silent remapping to a different model
  family, effort, or speed tier.
- Evidence: `VOC-117-EV-00`

## VOC-117-TEST-03 — Missing credentials and unsupported prefixes fail closed

- Covers: `VOC-117-AC-03`
- Preconditions: updated workflow routing / auth checks and negative fixtures
- Procedure: Negative tests simulate missing `CURSOR_API_KEY` on a Cursor-backed path
  and an unsupported provider prefix; assert fail-closed behavior with no fallback
  to another vendor/model and no credential printing.
- Expected result: jobs fail closed; no silent substitution; no secrets in output.
- Evidence: `VOC-117-EV-00`

## VOC-117-TEST-04 — Current-state comments do not claim dormant routes are active

- Covers: `VOC-117-AC-04`
- Preconditions: updated roles.yml / workflow comments (and any caller docs touched)
- Procedure: Deterministic string/policy checks or reviewed fixture assertions confirm
  active-state commentary describes the new Cursor lineup, while historical OpenAI/
  Codex narrative is clearly historical rather than current.
- Expected result: no current-state claim that OpenAI/Codex planner/escalation or
  obsolete Cursor bindings are the live configuration.
- Evidence: `VOC-117-EV-00`

## VOC-117-TEST-05 — Source self-CI, caller suites, and safety gates pass

- Covers: `VOC-117-AC-05`
- Preconditions: final implementation branch with infra merge SHA pinned in the caller
- Procedure: Run infra unit suite and caller governance/fixture suites listed in
  `implementation-plan.md`; inspect that exact-SHA review, risk classification,
  protected checks, and one-retry limits remain unchanged; confirm caller fixture pin
  equals the exact reviewed infra merge.
- Expected result: suites pass; safety gates unchanged; pin matches infra merge SHA.
- Evidence: `VOC-117-EV-00`

Include positive, negative, authorization, failure, migration, accessibility, and
rollback coverage as applicable. Tests must not use secrets or production data.
