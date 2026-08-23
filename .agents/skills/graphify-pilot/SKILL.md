---
name: graphify-pilot
description: Opt-in code-only Graphify index pilot—build a local navigation hint graph via the repository-locked runner; verify consequential claims in current source.
disable-model-invocation: true
---

# Graphify pilot (opt-in, code-only)

**Automatic invocation is disabled** for this skill throughout VOC-112. Use it only when an operator explicitly opts in. Graduation to always-on use requires a separate governed change with Graphify-specific measurement and acceptance gates.

## Governance precedence

When this skill conflicts with `AGENTS.md`, `CLAUDE.md`, approved change packages, tests, or source code, the repository sources win.

## What this pilot does

Graphify builds a **local code-relationship index** (tree-sitter AST extraction) as a **navigation hint**. It does **not** replace direct source inspection, the `vocanova-repo-navigator` skill, or canonical documentation.

- **Code-only:** extraction uses `--code-only` (no document/image semantic pass, no LLM provider).
- **Query logging off:** the runner sets `GRAPHIFY_QUERY_LOG_DISABLE=1`.
- **Hint-only:** treat `graphify-out/` output as suggestions; verify consequential claims in current source before acting.
- **No committed graph output by default:** `graphify-out/` is gitignored unless a separate governed change justifies committing artifacts.

## Prerequisites (explicit operator setup)

The locked runtime is **not** installed during ordinary agent sessions.

1. Review pins in `scripts/graphify/runtime-identity.yaml` and `scripts/graphify/requirements.lock`.
2. Run **`bash scripts/graphify/setup.sh`** once to create `scripts/graphify/.venv/` from the hash-locked lockfile.
3. Confirm identity with **`bash scripts/graphify/check`**.

If setup was skipped or the lock identity mismatches, the check exits non-zero with remediation guidance. The check and run scripts perform **no download, upgrade, global fallback, hook registration, or user-profile mutation**.

## Running extraction

From the repository root after a successful check:

```bash
bash scripts/graphify/run.sh
```

Optional target directory (defaults to repository root):

```bash
bash scripts/graphify/run.sh apps/api
```

Output directory defaults to `graphify-out/` (override with `GRAPHIFY_OUTPUT_DIR` for disposable runs).

## What agents must not do

- Do not run `graphify install`, hook installers, or provider auto-detection flows.
- Do not use global `pip install`, `uv tool install`, `pipx install`, or unpinned package installs for Graphify.
- Do not treat graph edges or summaries as authoritative over current source or tests.
- Do not export environment secrets, credentials, or raw CI logs while exploring graph output.

## Validation

```bash
bash scripts/graphify/check
node --test scripts/foundation/voc112-graphify.test.mjs
```

See `docs/development/agent-skills.md` (completed in T04) for operator documentation.
