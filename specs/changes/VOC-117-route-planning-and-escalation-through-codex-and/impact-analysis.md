# VOC-117 — Impact Analysis

## Security and privacy

This package changes which Cursor models occupy governed pipeline roles and how
parameterized model strings are passed to the Cursor CLI. It does not introduce new
secrets, OAuth/session material, production-data access, or user-facing data flows.

Security controls that must remain:

- Cursor-backed paths require `CURSOR_API_KEY` and fail closed when it is absent.
- Credentials are never printed in logs, evidence, or PR comments.
- No OpenAI/Codex path is required or re-enabled by this package.
- Unsupported prefixes fail closed without silent vendor/model substitution.

## Data and migrations

None. No database, schema, seed, or migration changes are proposed.

## Analytics and accessibility

None for product/runtime behavior. No direct analytics instrumentation or
user-interface accessibility effect.

## Risks, dependencies, and evidence

- `VOC-117-R00`: **High workflow risk** if parameterized model strings are stripped,
  ignored, or silently remapped so Standard/non-fast or `effort=high` is not what
  actually runs. Mitigation: deterministic workflow tests for the six bindings and
  fail-closed unsupported-prefix cases (`VOC-117-AC-01`–`AC-03`).
- `VOC-117-R01`: **High documentation risk** if historical OpenAI/Codex or obsolete
  Cursor comments remain written as current state. Mitigation: same-task
  current-state comment/doc updates (`VOC-117-D06`, `VOC-117-AC-04`).
- `VOC-117-R02`: **Medium independence tradeoff** — implementer escalation matches
  implementer, and planner/plan_reviewer share `grok-4.6`. Accepted by
  `VOC-117-D00` / `VOC-117-DEP-03` / `VOC-117-DEP-04`. Implementer vs reviewer
  remain distinct (`composer-2.5` vs `grok-4.6`).
- `VOC-117-R03`: **Medium operational risk** if the pinned Cursor CLI does not honor
  bracket parameters identically to IDE/docs. Mitigation: implementer verifies
  against the pinned CLI and records the compatible invocation form without
  changing stored `roles.yml` bindings (`VOC-117-DEP-02`).
- `VOC-117-R04`: **Low release risk** because no application runtime deployment
  change is intended; rollback is config/workflow/fixture reversion.
- Protected surfaces: `KARSIFT/karsift-ai-infra/config/roles.yml`, plan/implement/
  review/plan-review workflows, infra tests, caller `tooling/governance/` fixtures
  and tests, and any caller docs asserting the active model lineup.
- `VOC-117-DEP-00` through `VOC-117-DEP-04`: see `change.yaml`.
- `VOC-117-EV-00`: T00 evidence — changed files, exact role bindings, validation
  commands/results, infra merge SHA, and caller fixture pin.

## Monitoring

`monitoring_impact.state: none` — no monitor or synthetic inventory changes.

## Authority and EHR

Active authority model: **A-004**. No founder `approved` comment is required for
engineering-workflow adopt/merge/release gates. EHR is not triggered.

This draft proposes **R4** because it changes protected AI model routing and
governance fixtures/workflows, but the path classifier and independent verifier
remain authoritative.
