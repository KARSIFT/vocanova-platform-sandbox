# VOC-112-T03 evidence — Graphify pilot (code-only, opt-in)

Draft carrier for implementation evidence. No graph dumps, secrets, or personal data.

## gate_status

pending — populate at T03 implementation time

## Pin table

| Component | Pin | Notes |
|-----------|-----|-------|
| Graphify upstream repository | pending | e.g. Graphify-Labs/graphify |
| Graphify upstream commit/tag | pending | exact reviewed revision |
| Runtime package (`graphifyy`) | pending | exact version |
| Provenance registry entry | pending | `.agents/skills/provenance.yaml` |

## Pilot configuration checklist

| Check | Result | Notes |
|-------|--------|-------|
| Code-only / `--code-only` enforced | pending | |
| Query logging disabled | pending | |
| `.graphifyignore` excludes secrets/generated/vendor | pending | |
| Skill marked opt-in / auto-invocation disabled | pending | |
| Generated graph output gitignored by default | pending | |
| Runner fails safely without global install | pending | exit code + message |
| Hint-only language in skill | pending | verify in source |

## Validation commands

| Command | Result |
|---------|--------|
| `node --test scripts/foundation/voc112-agent-skills.test.mjs` | pending |
| `scripts/graphify/<check>` (exact name at implementation) | pending |
| `git diff --check` | pending |

## Acceptance mapping

- `VOC-112-AC-03` / `VOC-112-EV-03` — complete when checklist passes and pin table filled.
