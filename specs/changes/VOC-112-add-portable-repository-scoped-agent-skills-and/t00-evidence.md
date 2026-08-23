# VOC-112-T00 evidence — skill framework, adapters, and validation

Do not record secrets, credentials, session values, OAuth material, personal data, or
complete CI logs.

## gate_status

pass — T00 framework, adapters contract, provenance schema, and validation landed

## Scope implemented

| File | Action |
|------|--------|
| `.agents/skills/README.md` | created — one-source architecture, frontmatter, budgets, governance precedence |
| `.agents/skills/provenance.schema.yaml` | created — immutable per-skill record schema (`repository-native`, `adapted`) |
| `.claude/skills/README.md` | created — adapter loader contract (`${CLAUDE_PROJECT_DIR}` one-line loader) |
| `scripts/foundation/voc112-agent-skills.test.mjs` | created — data-driven validator with positive/negative fixtures |
| `docs/development/agent-skills.md` | created — skeleton sections (completed in T04) |

No canonical skill bodies were added in T00; later tasks add skill directories and
`PROVENANCE.yaml` records without editing the T00 validator or schema.

## Validation commands

| Command | Result | Notes |
|---------|--------|-------|
| `node --test scripts/foundation/voc112-agent-skills.test.mjs` | pass | 9 tests, 0 failures; exact adapter target, opt-in metadata parity, denylist attribution, and strict provenance fixtures included |
| `bash scripts/governance/validate-governance.sh` | pass | AGENTS.md unchanged in T00 |
| `bash scripts/governance/classify-change-risk.sh` | pass | path floor R1 reported |
| `git diff --check` | pass | no whitespace errors |

## Adapter/canonical parity sample

T00 introduces framework files only; no skill subdirectories yet.

| Skill | Canonical path | Claude adapter | Symlink |
|-------|----------------|----------------|---------|
| _(none — T01/T02/T03 add skills)_ | — | — | none |

The validator dynamically discovers `.agents/skills/*/SKILL.md` and requires matching
`.claude/skills/*/SKILL.md` adapters, per-skill `PROVENANCE.yaml`, and manifest
coverage. Parallel tasks add conforming skill directories without a shared registry or
T00 validator edit.

## Acceptance mapping

- `VOC-112-AC-00` / `VOC-112-EV-00` — framework validation passes; parity table recorded;
  T01–T03 can add skills in parallel.
