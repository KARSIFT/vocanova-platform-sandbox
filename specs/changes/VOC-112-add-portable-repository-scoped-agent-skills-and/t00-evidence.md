# VOC-112-T00 evidence — skill framework, adapters, and validation

Draft carrier for implementation evidence. Do not record secrets, credentials, session
values, OAuth material, personal data, or complete CI logs.

## gate_status

pending — populate at T00 implementation time

## Scope implemented

Record the exact files added for:

- `.agents/skills/README.md`
- `.agents/skills/provenance.schema.yaml` (immutable per-skill schema)
- `.claude/skills/README.md`
- `scripts/foundation/voc112-agent-skills.test.mjs`
- `docs/development/agent-skills.md` skeleton

## Validation commands

| Command | Result | Notes |
|---------|--------|-------|
| `node --test scripts/foundation/voc112-agent-skills.test.mjs` | pending | |
| `bash scripts/governance/validate-governance.sh` | pending | if AGENTS.md unchanged in T00, note N/A |
| `bash scripts/governance/classify-change-risk.sh` | pending | |
| `git diff --check` | pending | |

## Adapter/canonical parity sample

Record a table of skill names validated in T00 (may be README-only bootstrap until later tasks add bodies).

| Skill | Canonical path | Claude adapter | Symlink |
|-------|----------------|----------------|---------|
| pending | pending | pending | none expected |

Also record that the validator dynamically discovers per-skill `PROVENANCE.yaml` files and
that T01/T02/T03 require no shared registry or validator edit.

## Acceptance mapping

- `VOC-112-AC-00` / `VOC-112-EV-00` — complete when validation passes and parity table recorded.
