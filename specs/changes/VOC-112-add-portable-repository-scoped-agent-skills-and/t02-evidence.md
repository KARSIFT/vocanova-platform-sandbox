# VOC-112-T02 evidence — pinned shared engineering skills

Draft carrier for implementation evidence. Record hashes and licenses only — no upstream
secret examples, credentials, or raw CI logs.

## gate_status

pending — populate at T02 implementation time

## Shared skill provenance table

| Skill directory | Upstream repo | Upstream commit | Path | License | content_sha256 |
|-----------------|---------------|-----------------|------|---------|----------------|
| context-mapping | pending | pending | pending | pending | pending |
| systematic-debugging | pending | pending | pending | pending | pending |
| verification-before-completion | pending | pending | pending | pending | pending |
| github-actions-efficiency | pending | pending | pending | pending | pending |
| react-next-performance | pending | pending | pending | pending | pending |
| playwright-browser-testing | pending | pending | pending | pending | pending |
| security-threat-modeling | pending | pending | pending | pending | pending |

## Security review summary

Record reviewer identity (or agent role), date, and confirmation that each skill:

- forbids secret/credential/log exfiltration;
- references repository validation tiers;
- contains no unpinned install or hidden network instructions;
- remains subordinate to `AGENTS.md`/canonical docs.

## Validation commands

| Command | Result |
|---------|--------|
| `node --test scripts/foundation/voc112-agent-skills.test.mjs` | pending |
| `bash scripts/governance/validate-governance.sh` | pending if applicable |
| `git diff --check` | pending |

## Acceptance mapping

- `VOC-112-AC-02` / `VOC-112-EV-02` — complete when seven skills pass validation and provenance table is filled.
