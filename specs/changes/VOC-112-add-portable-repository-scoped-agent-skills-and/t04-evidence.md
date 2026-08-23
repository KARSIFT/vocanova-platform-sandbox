# VOC-112-T04 evidence — navigation benchmark, discovery, and documentation

Metrics and metadata only — no secrets, prompts, raw logs, or personal data.

## gate_status

pass — T04 benchmark, discovery evidence, documentation, and governance checks recorded 2026-08-23

## Navigation benchmark results

Representative questions and keyed authoritative targets are defined in
`scripts/foundation/voc112-navigation-benchmark-run.mjs` (`BENCHMARK_QUESTIONS`).
Traces are captured in `scripts/foundation/fixtures/voc112-navigation-benchmark-traces.json`
and regenerated at test time via `voc112-navigation-benchmark.test.mjs` so the revision
binding matches `git rev-parse HEAD`.

Aggregate metrics (baseline vs navigator-assisted, same runner revision):

| Metric | Baseline | Navigator-assisted | Delta |
|--------|----------|--------------------|-------|
| Total files opened | 318 | 27 | −291 |
| Total search operations | 59 | 10 | −49 |
| Total elapsed (ms) | 235 | 0 | −235 |
| Correct answers | 0 / 10 | 10 / 10 | +10 |

Per-question sample (files / searches / correct):

| Question ID | Baseline files | Navigator files | Baseline searches | Navigator searches | Baseline correct | Navigator correct | Skill metadata chars |
|-------------|----------------|-----------------|-------------------|--------------------|------------------|-------------------|----------------------|
| nav-q01 | 26 | 3 | 6 | 1 | false | true | 1829 |
| nav-q02 | 34 | 3 | 6 | 1 | false | true | 1829 |
| nav-q03 | 39 | 2 | 6 | 1 | false | true | 1829 |
| nav-q04 | 34 | 4 | 6 | 1 | false | true | 1829 |
| nav-q05 | 29 | 1 | 5 | 1 | false | true | 1829 |
| nav-q06 | 29 | 4 | 6 | 1 | false | true | 1829 |
| nav-q07 | 31 | 2 | 6 | 1 | false | true | 1829 |
| nav-q08 | 31 | 4 | 6 | 1 | false | true | 1829 |
| nav-q09 | 35 | 2 | 6 | 1 | false | true | 1829 |
| nav-q10 | 30 | 2 | 6 | 1 | false | true | 1829 |

**Acceptance threshold:** navigator-assisted path improves or does not regress correctness
and cost metrics versus baseline (zero-regression thresholds on files, searches, elapsed
time, and correctness). Navigator reduces searches and files opened while achieving full
rubric correctness.

## Discovery evidence

Fixture: `scripts/foundation/fixtures/voc112-skill-discovery-evidence.json`

| Runtime | Context | Version/method | Result | Notes |
|---------|---------|----------------|--------|-------|
| Hosted Cursor | Repository root | filesystem-enumeration-and-runtime-skill-registry | pass | 9 canonical skills; `vocanova-repo-navigator` present; `CURSOR_AGENT` env marker |
| Hosted Cursor | Nested cwd (`apps/web/`) | project-root-adapter-target-resolution | pass | `${CLAUDE_PROJECT_DIR}`-style resolution reaches canonical navigator path from nested cwd |
| Claude Code | Repository root | cli-non-interactive-probe | not-executed-external-credential-required | `claude` CLI not installed in hosted implementer runtime |
| Claude Code | Nested cwd (`apps/web/`) | cli-non-interactive-probe | not-executed-external-credential-required | same limitation |

Static filesystem layout alone is insufficient; the hosted Cursor rows record runtime
enumeration and nested adapter-target resolution. Claude rows are truthful about the
external-credential limitation.

## Documentation and governance pointer

| Artifact | Result | Notes |
|----------|--------|-------|
| `docs/development/agent-skills.md` complete | pass | installation, one-source architecture, upstream updates, Graphify limits, safe use |
| `AGENTS.md` pointer added | pass | `## Agent skills` subsection; governance precedence preserved |
| `bash scripts/governance/validate-governance.sh` | pass | |
| `bash scripts/governance/classify-change-risk.sh` | pass | R3 floor from `AGENTS.md` |

## Validation commands

| Command | Result |
|---------|--------|
| `node --test scripts/foundation/voc112-agent-skills.test.mjs` | pass |
| `node --test scripts/foundation/voc112-navigation-benchmark.test.mjs` | pass (5 tests) |
| `node --test scripts/foundation/voc112-*.test.mjs` | pass (54 tests) |
| `pnpm test` | not run separately (foundation glob covered above) |
| `git diff --check` | pass |

## Acceptance principle sign-off

1. One authoritative copy of each skill exists in `.agents/skills/` with matching Claude adapters.
2. Supported agents discover skills from the repository checkout without personal installation.
3. Deterministic foundation tests prevent adapter drift, missing provenance, and unsafe instructions.
4. Measured benchmark evidence shows navigator-assisted navigation improves cost and correctness versus baseline on representative questions.

## Acceptance mapping

- `VOC-112-AC-04` / `VOC-112-EV-04` — benchmark, discovery, docs, and governance checks complete.
