# VOC-112 — Acceptance Criteria

## VOC-112-AC-00 — Canonical skill framework with Claude adapters and validation

- Requirement source: issue #933; `VOC-112-D00`–`D04`, `D07`
- Tasks: `VOC-112-T00`
- Tests: `VOC-112-TEST-00` through `VOC-112-TEST-05`
- Evidence: `VOC-112-EV-00`
- Result: pending

A fresh checkout contains `.agents/skills/` and matching `.claude/skills/` adapters for
every skill introduced by merged tasks. Adapters load canonical skills only; no Git
symlinks. `scripts/foundation/voc112-agent-skills.test.mjs` passes in `pnpm test` and
fails closed on adapter drift, missing provenance, forbidden patterns, broken references,
or budget violations.

## VOC-112-AC-01 — `vocanova-repo-navigator` routes without duplicating governance

- Requirement source: issue #933; `VOC-112-D05`
- Tasks: `VOC-112-T01`
- Tests: `VOC-112-TEST-06`, `VOC-112-TEST-07`
- Evidence: `VOC-112-EV-01`
- Result: pending

The navigator skill covers all required domains, points to authoritative paths/commands,
states governance precedence, stays within documented size budgets, and does not embed
full copies of `AGENTS.md` or canonical governance documents.

## VOC-112-AC-02 — Seven pinned shared engineering skills adapted for this repository

- Requirement source: issue #933; `VOC-112-D06`, `D07`
- Tasks: `VOC-112-T02`
- Tests: `VOC-112-TEST-08`, `VOC-112-TEST-09`
- Evidence: `VOC-112-EV-02`
- Result: pending

Exactly seven shared skills exist with complete provenance records, repository-specific
safety adaptations (validation tiers, forbidden secret/log behavior, stack references),
matching Claude adapters, and passing validation. No unpinned upstream content.

## VOC-112-AC-03 — Graphify pilot is opt-in, code-only, and safely bounded

- Requirement source: issue #933; `VOC-112-D08`
- Tasks: `VOC-112-T03`
- Tests: `VOC-112-TEST-10`, `VOC-112-TEST-11`
- Evidence: `VOC-112-EV-03`
- Result: pending

Graphify is pinned with recorded provenance, exposed only through an explicit opt-in
skill and repository-owned runner/check, runs in code-only mode with query logging disabled
and repository ignore rules, does not commit generated graph output by default, and fails
safely when prerequisites are missing without global auto-install.

## VOC-112-AC-04 — Measurement, discovery, and documentation complete the acceptance principle

- Requirement source: issue #933 acceptance principle; `VOC-112-D09`, `D10`
- Tasks: `VOC-112-T04`
- Tests: `VOC-112-TEST-12`, `VOC-112-TEST-13`, `VOC-112-TEST-14`
- Evidence: `VOC-112-EV-04`
- Result: pending

Deterministic benchmark evidence shows navigator-assisted navigation improves or does not
regress cost and correctness versus baseline on representative questions. Discovery is
demonstrated from repository root and a nested directory. `docs/development/agent-skills.md`
and the `AGENTS.md` pointer document one-source architecture, updating pinned upstream
material, Graphify limitations, and safe use. No product/runtime/deployment behavior change.

Acceptance criteria must be observable, stable, security-aware, and bidirectionally
traceable to requirements, tasks, tests, and evidence.
