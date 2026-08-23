# VOC-112-T01 evidence — `vocanova-repo-navigator` router skill

Draft carrier for implementation evidence. No secrets, credentials, or personal data.

## gate_status

complete — T01 implementation recorded 2026-08-23

## Navigator review checklist

| Check                                  | Result | Notes                                                                                 |
| -------------------------------------- | ------ | ------------------------------------------------------------------------------------- |
| All `VOC-112-D05` domains routed       | pass   | Web, API, database, auth, seed, deploy, monitoring, governance, validation, lifecycle |
| Governance precedence stated           | pass   | `## Governance precedence` section; repository sources win                            |
| No large `AGENTS.md`/governance paste  | pass   | 2501 body bytes; no distinctive AGENTS.md markers                                     |
| Referenced paths exist                 | pass   | Glob tokens (`deploy-*.yml`, `scripts/foundation/*.test.mjs`) resolved in test        |
| Intent-to-path mapping is row-scoped   | pass   | Negative fixture moves an auth path into the web row and must fail                    |
| Wildcard targets are exact             | pass   | Basename glob matcher rejects unrelated directory entries                             |
| Claude adapter loader-only             | pass   | Exact one-line `${CLAUDE_PROJECT_DIR}` loader                                         |
| `voc112-agent-skills` validation green | pass   | See commands below                                                                    |

## Size budget

| Metric                | Value | Limit                       |
| --------------------- | ----- | --------------------------- |
| `description` chars   | 191   | 512                         |
| `SKILL.md` body bytes | 2501  | 4096 router / 32768 generic |
| `SKILL.md` body lines | 49    | 120 router / 400 generic    |

## Validation commands

```bash
node --test scripts/foundation/voc112-agent-skills.test.mjs
node --test scripts/foundation/voc112-navigator.test.mjs
```

Both exited 0 on the T01 working tree.

## Acceptance mapping

- `VOC-112-AC-01` / `VOC-112-EV-01` — checklist and tests above satisfy T01 scope.
