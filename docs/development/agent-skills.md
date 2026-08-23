# Agent skills

Repository-scoped agent skills give every supported agent one authoritative copy of
each procedure from a fresh checkout, with deterministic safety validation.

> **Skeleton (VOC-112-T00):** Operator sections marked *completed in T04* are
> placeholders until the documentation task lands.

## Installation scope

*Completed in T04.* Skills ship in-repo under `.agents/skills/`; no personal or
global installation is required for discovery from the repository root.

## One-source architecture

- **Canonical tree:** `.agents/skills/<skill-name>/SKILL.md`
- **Claude adapters:** `.claude/skills/<skill-name>/SKILL.md` (loader-only)
- **Architecture reference:** `.agents/skills/README.md`
- **Adapter contract:** `.claude/skills/README.md`

Governance precedence: `AGENTS.md`, `CLAUDE.md`, approved change packages, tests,
and source code override skill prose.

## Validation

```bash
node --test scripts/foundation/voc112-agent-skills.test.mjs
pnpm test   # includes foundation tests
```

## Updating pinned upstream material

*Completed in T04.* Adapted skills record upstream identity and hashes in per-skill
`PROVENANCE.yaml` (see `.agents/skills/provenance.schema.yaml`).

## Graphify pilot limitations

*Completed in T04.* Graphify remains explicit opt-in; ordinary agent sessions do not
auto-invoke the pilot skill.

## Safe use

*Completed in T04.* Skills must not be used to exfiltrate secrets, credentials, CI
logs, or personal data. Follow `AGENTS.md` for governed change workflow.
