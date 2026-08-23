# VOC-112-T04 evidence — navigation benchmark, discovery, and documentation

Draft carrier for implementation evidence. Metrics and metadata only — no secrets or
personal data.

## gate_status

pending — populate at T04 implementation time

## Navigation benchmark results

Representative questions and keyed expected authoritative targets are defined in
`scripts/foundation/voc112-navigation-benchmark.test.mjs` at implementation time. Every
row below is derived from sanitized structured traces from controlled same-revision,
same-model/runtime sessions; hand-entered or hardcoded success values are invalid.

| Question ID | Baseline files | Navigator files | Baseline searches | Navigator searches | Baseline time | Navigator time | Baseline correct | Navigator correct | Skill metadata chars |
|-------------|----------------|-----------------|-------------------|--------------------|---------------|----------------|------------------|-------------------|----------------------|
| pending | pending | pending | pending | pending | pending | pending | pending | pending | pending |

**Acceptance threshold:** navigator-assisted path must improve or not regress correctness
and cost metrics versus baseline per benchmark script declarations.

## Discovery evidence

| Runtime | Context | Version/method | Result | Notes |
|---------|---------|----------------|--------|-------|
| Hosted Cursor | Repository root | pending | pending | actual list/invocation |
| Hosted Cursor | Nested cwd (`apps/web/` recommended) | pending | pending | actual list/invocation |
| Claude Code | Repository root | pending | pending | pass or `not-executed-external-credential-required` |
| Claude Code | Nested cwd (`apps/web/` recommended) | pending | pending | pass or same truthful limitation |

Static filesystem assertions do not satisfy runtime discovery. Record runtime/revision and
outcome metadata only—no prompts, raw logs, secrets, or personal data.

## Documentation and governance pointer

| Artifact | Result | Notes |
|----------|--------|-------|
| `docs/development/agent-skills.md` complete | pending | installation, updates, Graphify limits |
| `AGENTS.md` pointer added | pending | precedence preserved |
| `bash scripts/governance/validate-governance.sh` | pending | |
| `bash scripts/governance/classify-change-risk.sh` | pending | expect R3 when AGENTS.md touched |

## Validation commands

| Command | Result |
|---------|--------|
| `node --test scripts/foundation/voc112-agent-skills.test.mjs` | pending |
| `node --test scripts/foundation/voc112-navigation-benchmark.test.mjs` | pending |
| `pnpm test` | pending |
| `git diff --check` | pending |

## Acceptance principle sign-off

Record explicit statement that:

1. one authoritative copy of each skill exists in `.agents/skills/`;
2. supported agents discover skills without personal installation;
3. deterministic checks prevent drift and unsafe instructions;
4. measured evidence shows navigator improves or does not regress navigation cost/correctness.

## Acceptance mapping

- `VOC-112-AC-04` / `VOC-112-EV-04` — complete when benchmark, discovery, docs, and governance checks pass.
