# VOC-112 — Portable repository-scoped agent skills and codebase navigation

| Field | Value |
|-------|-------|
| Package | `VOC-112` |
| Title | Add portable repository-scoped agent skills and codebase navigation |
| Path | `specs/changes/VOC-112-add-portable-repository-scoped-agent-skills-and` |
| Status | `draft` |
| Risk | `R3` (draft proposal; AGENTS.md path floor and agent-instruction semantics apply) |
| Authority model | A-004 active |
| Requirement source | GitHub issue [#933](https://github.com/KARSIFT/vocanova-platform-sandbox/issues/933) |
| Target branch | `develop` |
| Approval | `not-approved` |
| Implementation authorized | `false` |
| `automatic_merge_allowed` | `true` (per AGENTS.md A-004 drafting rule) |

## Problem

Agents repeatedly rediscover VocaNova repository structure and operational invariants.
Personal or global skills are not portable to another device or CI checkout, and
duplicate Claude/Codex/Cursor copies can drift. A broad third-party skill install can
also waste context, download unpinned tools, read sensitive files, or make stale
generated knowledge look authoritative.

## Required outcome (summary)

1. **One canonical skill tree** under `.agents/skills/<name>/SKILL.md`, discovered
   directly by Codex, Cursor, and compatible agents.
2. **Minimal Claude adapters** under `.claude/skills/<name>/SKILL.md` that load the
   canonical skill only — no independent procedure, no Git symlinks.
3. **Deterministic validation** for format, adapters, provenance, forbidden patterns,
   and context-size budgets; wired into `pnpm test` foundation coverage.
4. **`vocanova-repo-navigator`** — a router skill for web/API/database/auth/seed/deploy/
   monitoring/governance/testing, validation tiers, shared-edge invariants, and the
   issue→plan→task lifecycle without duplicating `AGENTS.md` or canonical docs.
5. **Shared engineering skills** — only security-reviewed, repository-adapted, pinned
   upstream material that materially helps this stack.
6. **Graphify pilot** — complementary, code-only, opt-in, pinned, with repository-owned
   ignore rules and no committed generated graph unless evidence proves benefit.
7. **Measurement and documentation** — baseline vs skill-assisted navigation cost and
   correctness; discovery proof from repository root and a nested working directory.

No product, runtime, or deployment behavior change.

## Tasks

| Task | Summary | Depends on |
|------|---------|------------|
| T00 | Canonical skill framework, Claude adapters, provenance registry, validation | — |
| T01 | `vocanova-repo-navigator` router skill | T00 |
| T02 | Pinned shared engineering skills (seven domains) | T00 |
| T03 | Graphify pilot (code-only, opt-in, repo-owned runner) | T00 |
| T04 | Navigation benchmark, discovery evidence, documentation completion | T01, T02, T03 |

See `tasks.md` for full task definitions.

## Verification, approvals, release, and closure

See `test-plan.md`, `release-plan.md`, and `implementation-plan.md`.
Under **active A-004**, engineering-workflow gates do not wait on a founder
`approved` comment. This draft carries no adoption or implementation authority.
