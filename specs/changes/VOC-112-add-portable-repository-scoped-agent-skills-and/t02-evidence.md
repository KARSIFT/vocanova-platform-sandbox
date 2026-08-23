# VOC-112-T02 evidence — pinned shared engineering skills

Record hashes and licenses only — no upstream secret examples, credentials, or raw CI logs.

## gate_status

complete — seven shared skills, adapters, provenance records, and `voc112-shared-skills.test.mjs` pass

## Shared skill provenance table

| Skill directory | Source/commit/path | License + retained notice | Upstream SHA-256 | Local SKILL.md SHA-256 | Adaptation/rejection note |
|-----------------|--------------------|---------------------------|------------------|------------------------|---------------------------|
| context-mapping | `addyosmani/agent-skills` `5a5ea45e806f82273549fd85e60adb95d55f510d` `skills/context-engineering/SKILL.md` | MIT / `LICENSE` | `ff9d4e5706bdd2eb7de1bfed569f1f42d28e478979ce6fcc32e617e7861b491d` | `7b3d72d5128e6295735faa7d44578d143ac5bc000192a43bc539a48f2d2361ba` | Renamed to context-mapping; vocanova-repo-navigator routing; `docs/development.md` validation tiers |
| systematic-debugging | `obra/superpowers` `b36e0829c6d0140e93cfef2ca599b1b07d4a7797` `skills/systematic-debugging/SKILL.md` | MIT / `LICENSE` | `808fc5717aa88ad65efff312b11c186294d3e6ee301afb584e2f86599b137787` | `091fd15690865ba07d2cc30e139bd5fa7ac25eff41f6135d492a402843e3a4fa` | Repository commands; removed secret/credential examples and human-review pauses |
| verification-before-completion | `obra/superpowers` `b36e0829c6d0140e93cfef2ca599b1b07d4a7797` `skills/verification-before-completion/SKILL.md` | MIT / `LICENSE` | `2befe7fc55bcadaa3d97dd9e8efeb633d2561c0ebe74c5a8b17c4d9e7e4520b3` | `995868ab8cef678c58f34126044ada02260f56b2792c5941150502f2a8cb9ede` | Mapped claims to `pnpm validate`, `pnpm test`, `go test`, foundation/governance commands |
| github-actions-efficiency | `github/awesome-copilot` `83561bd7d8a46fcda0581aedabdf8eac7cb196b6` `skills/github-actions-efficiency/SKILL.md` | MIT / `LICENSE` | `9c41e860468e5c88d83ab6eec70c585b0f122facec2632ea65497b3389139e43` | `7c72123a2a544b02996819a68dcde17fbfdad49466167d17f8852168edfb309b` | JSON-only `gh` metadata; condensed `references/actions.md`; no raw log export |
| react-next-performance | repository-native (React/Next official docs + repo paths) | N/A | N/A | `20a945567551b22d521d381f4e252a6c32890ce793e3fb36ebc6aec6a21871cd` | **Rejected** `vercel-labs/agent-skills` `dd089a8c752c966dee8bf0f27cb625ba193ffd9e` — no compatible license; authored independently |
| playwright-browser-testing | `testdino-hq/playwright-skill` `d3be9ca4d7303e2aee3eba4842963abf573117b0` `SKILL.md` | MIT / `LICENSE` | `4929cff14de61f93d72cbbef192edc880cf2fc244572effcc3bf8724610284ba` | `b9829d1d18ab48ca522c0c6bab9f200a524040b5810f481b1595b2f607bf7209` | Scoped to `@playwright/test` `1.62.1` / `apps/web`; no global or unpinned installs |
| security-threat-modeling | `openai/skills` `49f948faa9258a0c61caceaf225e179651397431` `skills/.curated/security-threat-model/SKILL.md` | Apache-2.0 / `LICENSE.txt` | `1283c0dd62a8104d9edda4583569b5d8510b4ddaa45120687c999250fd96bad2` | `c97376298949c19da9aa800834d4728888ba8b07fb23d62334745d7133a13e9c` | Vocanova paths; condensed prompt template; removed interactive pause gates |

## Security review summary

- **Reviewer:** implementer agent (VOC-112-T02)
- **Date:** 2026-08-23
- Every committed instruction file, retained license, and reference under the seven skill directories was reviewed.
- Each skill states governance precedence and forbids `.env*` access, secret/credential exposure, and raw CI log pasting.
- Skills reference repository validation (`docs/development.md`, `pnpm validate`, targeted package commands) instead of inventing installs.
- No unpinned global installs, hidden network fetches, or profile-mutation instructions remain in adapted content.
- `github-actions-efficiency` uses `gh run view --json` only — not `--log-failed` or log export.
- `playwright-browser-testing` uses workspace-pinned Playwright via `pnpm --filter @vocanova/web test:e2e`.
- **Vercel rejection:** `vercel-labs/agent-skills` at `dd089a8c752c966dee8bf0f27cb625ba193ffd9e` has no detected or committed license for the React skill; its text was not copied or adapted. `react-next-performance` is repository-native.

## Validation commands

| Command | Result |
|---------|--------|
| `node --test scripts/foundation/voc112-agent-skills.test.mjs` | pass |
| `node --test scripts/foundation/voc112-shared-skills.test.mjs` | pass |
| `bash scripts/governance/validate-governance.sh` | not required (no `AGENTS.md`/governance doc edits in T02) |
| `git diff --check` | pass |

## Acceptance mapping

- `VOC-112-AC-02` / `VOC-112-EV-02` — seven shared skills with adapters, provenance, security review, and passing validation.
