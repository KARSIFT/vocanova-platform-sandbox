# VOC-117 — Route planning and review through Cursor Grok 4.6 Standard and align implementer escalation: Specification

## Objective and requirement source

Update the governed AI role lineup as one coherent model-routing change. Use Cursor
Composer 2.5 for implementer and implementer escalation, and use Cursor Grok 4.6
Standard (`fast=false`) for planner and review roles, with explicit `effort=high`
on every Grok 4.6 binding because live Cursor CLI discovery proved the
effort-omitted identifier unavailable.

**Requirement source:** [GitHub issue #978](https://github.com/KARSIFT/vocanova-platform-sandbox/issues/978),
as superseded by the comment from `m-e-h-r-d-a-a-d` at `2026-08-24T22:04:19Z`. That
comment supersedes the original OpenAI/Codex mappings and removes any
`OPENAI_API_KEY` / OpenAI execution-path requirement.

Plan PR #979 adopted this package under A-004; task issue #980 separately carries
implementation authority for the single task.

## Scope and non-goals

### In scope

1. Update the authoritative `KARSIFT/karsift-ai-infra/config/roles.yml` to the six
   exact stored values listed under Decisions.
2. Make workflow model resolution/invocation compatible with parameterized Cursor
   model strings (`model[param=value,...]`) for planner, implementer,
   implementer_escalation, reviewer, reviewer_fast_retry, and plan_reviewer paths.
3. Preserve Cursor/OpenCode (or other still-supported prefix) routing behavior for
   prefixes that remain in use; do not silently fall back to another vendor/model
   when the configured binding cannot be honored.
4. Keep authentication fail-closed: Cursor-backed paths require `CURSOR_API_KEY`;
   never print credentials. Do not add, re-enable, or require OpenAI/Codex execution
   for this package.
5. Update current-state comments and docs so dormant historical OpenAI/Codex or
   obsolete Cursor bindings are not described as the active lineup.
6. Update deterministic role-resolution and workflow routing tests in the primary
   infra repository, then sync and pin the caller mirrored fixture to the exact
   reviewed infra merge.
7. Preserve independent exact-SHA review, risk classification, protected checks, and
   one-retry limits.

### Non-goals / explicitly excluded

- Changing product behavior, permissions, review authority, production data access,
  or deployment credentials.
- Reintroducing OpenAI/Codex planner or escalation routing from the superseded
  issue body.
- Inventing alternate model IDs when the configured Cursor binding cannot be
  honored.
- Weakening fail-closed credential checks, exact-SHA review, risk floors, or
  one-retry caps.
- Self-adoption or self-authorization of this package.
- Monitor/synthetic inventory changes.

## Risk and protected areas

- **Adopted risk:** **R4**.
- Protected areas: AI provider/model routing; CI/CD workflows; repository
  governance fixtures under `tooling/governance/`; shared infra role config.
- Protected technical effect: which model occupies each pipeline role and how
  parameterized model strings are passed to Cursor; no application runtime effect
  is intended.
- EHR: not triggered.
- Active authority model: **A-004**. No founder `approved` comment is a
  merge/adopt/release gate, but governance/workflow changes still require exact-SHA
  independent verification and fail-closed controls.

## Decisions

`VOC-117-D00`: The authoritative stored role bindings are exactly:

| Role | Stored value |
|------|----------------|
| `implementer` | `cursor/composer-2.5` |
| `implementer_escalation` | `cursor/composer-2.5` |
| `planner` | `cursor/grok-4.6[effort=high,fast=false]` |
| `reviewer` | `cursor/grok-4.6[effort=high,fast=false]` |
| `reviewer_fast_retry` | `cursor/grok-4.6[effort=high,fast=false]` |
| `plan_reviewer` | `cursor/grok-4.6[effort=high,fast=false]` |

`composer2.5` spelling from informal requests normalizes to Cursor model ID
`composer-2.5`. `grok-4.6` is the model ID; Standard is the non-Fast speed tier.
Live Cursor CLI model discovery and direct invocation established that the
effort-qualified high/non-Fast identifier works while
`grok-4.6[fast=false]` is unavailable, so `high` is explicit for every Grok role.

`VOC-117-D01`: The superseding issue comment replaces the original OpenAI/Codex
objective. This package must not require `OPENAI_API_KEY`, must not route planner
or escalation through `openai/codex-action`, and must not keep OpenAI-routing
acceptance criteria from the superseded issue body.

`VOC-117-D02`: Workflows must accept the parameterized stored strings. After any
vendor-prefix handling, the invoked Cursor model must preserve the requested
Standard/non-fast behavior and explicit `effort=high`. If the pinned
CLI requires an equivalent supported form, convert deterministically; do not
silently substitute a different vendor, model family, or speed/effort tier.

`VOC-117-D03`: Missing required provider credentials or unsupported model prefixes
fail closed without fallback to another vendor or model.

`VOC-117-D04`: Independent exact-SHA review, deterministic risk classification,
protected checks, and one-retry limits remain unchanged. Same-model implementer
escalation and same-family planner/plan_reviewer bindings are accepted tradeoffs of
`VOC-117-D00`, not license to weaken review authority.

`VOC-117-D05`: Primary source of truth is `KARSIFT/karsift-ai-infra`. The caller
fixture under `tooling/governance/fixtures/karsift-ai-infra/` must be pinned to the
exact reviewed infra merge in the same task. Coordinated source and caller PRs may
remain one task.

`VOC-117-D06`: Current-state comments in `roles.yml`, workflow headers, and any
caller docs that assert an active model lineup must be updated so dormant historical
routes are not described as active after this change lands.

## Data, migrations, analytics, and accessibility

None for application/runtime behavior. This is a model-routing configuration and
workflow change only.

## Security, privacy, and authorization

No new secrets are introduced. Cursor paths continue to require `CURSOR_API_KEY`.
Credentials must never be printed. The change must not broaden production-data,
deploy, or merge authority.

Abuse/process risks:

1. Silent fallback to another model would falsify the requested lineup — mitigated
   by fail-closed routing tests.
2. Stale comments claiming OpenAI/Codex or obsolete Cursor bindings are active —
   mitigated by same-task current-state doc/comment updates.
3. Treating same-model escalation as a review-authority change — out of scope;
   review roles remain distinct from implementer (`composer-2.5` vs `grok-4.6`).

## Contradictions and open questions

1. **Parameterized CLI mechanics (`VOC-117-DEP-02`, resolved):** Live Cursor CLI
   discovery and direct invocation proved
   `grok-4.6[effort=high,fast=false]` succeeds while
   `grok-4.6[fast=false]` is unavailable. The authoritative stored bindings now
   use the working effort-qualified form without changing model family or speed.
2. **Same-model escalation:** `implementer` and `implementer_escalation` are
   identical by request. Historical escalation philosophy preferred a different
   lab; this package deliberately does not restore that.
3. **Planner vs plan_reviewer independence:** both use the same explicit-high
   non-Fast `grok-4.6` form. Cross-family independence from implementer is
   preserved; full cross-model independence between planner and plan_reviewer is
   not.
4. **Package path slug:** the directory retains the workflow-assigned
   `...-through-codex-and` slug even though OpenAI/Codex is out of scope. Do not
   rename the adopted package path; treat the title/specification wording as the
   current-state description.
