# VOC-117 — Route planning and review through Cursor Grok 4.6 Standard and align implementer escalation

| Field | Value |
|-------|-------|
| Package | `VOC-117` |
| Title | Route planning and review through Cursor Grok 4.6 Standard and align implementer escalation |
| Path | `specs/changes/VOC-117-route-planning-and-escalation-through-codex-and` |
| Status | `draft` |
| Risk | `R4` (draft proposal; AI role bindings, workflow model routing, and mirrored governance fixtures) |
| Authority model | A-004 active |
| Requirement source | GitHub issue [#978](https://github.com/KARSIFT/vocanova-platform-sandbox/issues/978), superseded by the 2026-08-24 comment |
| Target branch | `develop` |
| Approval | `not-approved` |
| Implementation authorized | `false` |
| `automatic_merge_allowed` | `true` (per AGENTS.md A-004 drafting rule) |

## Problem

The governed AI role lineup still mixes older Cursor bindings (`cursor/auto`,
`cursor/cursor-grok-4.5-medium`) and historical comments that describe dormant
OpenAI/Codex routes as if they were current. Issue #978 originally proposed Codex
for planning/escalation; the superseding thread comment replaces that with an
all-Cursor mapping that stores explicit Standard/non-fast (and plan-reviewer high
effort) parameters on Grok 4.6.

## Required outcome (summary)

1. Persist the six authoritative role bindings from the superseding comment in
   `KARSIFT/karsift-ai-infra/config/roles.yml`.
2. Make plan/implement/review/plan-review workflows compatible with parameterized
   Cursor model strings such as `grok-4.6[fast=false]` without silent vendor/model
   fallback.
3. Keep authentication fail-closed: Cursor paths require `CURSOR_API_KEY`; never
   print credentials; do not require or introduce an OpenAI execution path.
4. Update current-state comments/docs so dormant historical routes are not
   described as active.
5. Update deterministic tests and the caller mirror/pin together; preserve
   independent exact-SHA review, risk classification, protected checks, and
   one-retry limits.

## Authoritative stored role mapping

| Role | Stored value |
|------|----------------|
| `implementer` | `cursor/composer-2.5` |
| `implementer_escalation` | `cursor/composer-2.5` |
| `planner` | `cursor/grok-4.6[fast=false]` |
| `reviewer` | `cursor/grok-4.6[fast=false]` |
| `reviewer_fast_retry` | `cursor/grok-4.6[fast=false]` |
| `plan_reviewer` | `cursor/grok-4.6[effort=high,fast=false]` |

## Tasks

| Task | Summary | Depends on |
|------|---------|------------|
| T00 | Apply the Cursor role lineup, parameterized-model routing, tests, docs, and caller pin | — |

See `tasks.md` for the full task definition.

## Verification, approvals, release, and closure

See `test-plan.md`, `implementation-plan.md`, and `release-plan.md`.
Under **active A-004**, engineering-workflow gates do not wait on a founder
`approved` comment, but this draft itself carries no adoption or implementation
authority.

## Risk note

This package **proposes R4** because it changes protected AI model routing and
`tooling/governance/` fixtures/workflows. The path-based classifier and
independent verifier remain authoritative; this draft proposal is not a
determination.
