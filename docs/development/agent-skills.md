# Agent skills

Repository-scoped agent skills give every supported agent one authoritative copy of
each procedure from a fresh checkout, with deterministic safety validation.

## Installation scope

Skills ship in-repo under `.agents/skills/` and `.claude/skills/` (Claude loader
adapters). No personal or global installation is required for discovery from the
repository root. Cursor, Codex, and compatible agents read the canonical tree
directly; Claude adapters load canonical skills via `${CLAUDE_PROJECT_DIR}` so nested
working directories still resolve the same authoritative files.

## One-source architecture

- **Canonical tree:** `.agents/skills/<skill-name>/SKILL.md`
- **Claude adapters:** `.claude/skills/<skill-name>/SKILL.md` (loader-only; no Git
  symlinks)
- **Architecture reference:** `.agents/skills/README.md`
- **Adapter contract:** `.claude/skills/README.md`
- **Repository navigator:** `.agents/skills/vocanova-repo-navigator/SKILL.md`

Governance precedence: when skill prose conflicts with `AGENTS.md`, `CLAUDE.md`,
approved change packages, tests, or source code, the **repository sources win**.

## Validation

```bash
node --test scripts/foundation/voc112-agent-skills.test.mjs
node --test scripts/foundation/voc112-navigation-benchmark.test.mjs
pnpm test   # includes foundation tests
```

Capture fresh navigation/discovery evidence after rubric or navigator routing changes.
These are explicit, authenticated operator actions; the deterministic test validates the
committed sanitized capture and never starts an agent or makes a network request. The
required Repository Governance check uses `fetch-depth: 0`: authenticated
same-repository `main` ← `develop` promotion pull requests deterministically use
head/source-revision-bound `pr-validation` with exact PR base/head SHAs regardless of
capture-subject object availability; other pull requests that change the capture
fixture require each captured commit, prove ancestry, and bind captured/current
hashes; ordinary unchanged-fixture pull requests remain merge-base-anchored
`pr-validation`. Post-squash branch pushes re-validate the current hashes without requiring
discarded intermediate PR commits to be ancestors. A generic shallow
application-test checkout also remains non-mutating:

```bash
node scripts/foundation/voc112-navigation-benchmark-run.mjs --capture-codex
node scripts/foundation/voc112-navigation-benchmark-run.mjs --capture-claude-discovery
# Run only in the authorized hosted Cursor environment:
node scripts/foundation/voc112-navigation-benchmark-run.mjs --capture-cursor-discovery
```

Benchmark and discovery evidence fixtures live under
`scripts/foundation/fixtures/`. They bind to an exact ancestor revision plus hashes of
`AGENTS.md` and the canonical navigator skill, so a later evidence-only commit cannot
silently change the measured inputs. Raw runtime traces, prompts, and response bodies are
not written; only sanitized runtime identity, usage/count metrics, repository paths, and
rubric results are retained. Never pass or print a credential on the command line.

## Updating pinned upstream material

Adapted skills (shared engineering skills, Graphify pilot) record upstream identity,
separate upstream/local hashes, retained licenses, and adaptation notes in per-skill
`.agents/skills/<name>/PROVENANCE.yaml`. The schema is immutable at
`.agents/skills/provenance.schema.yaml` (T00-owned).

To update adapted material:

1. Choose a reviewed upstream commit with a compatible license; do not copy unlicensed
   sources (the Vercel `agent-skills` React guidance at commit
   `dd089a8c752c966dee8bf0f27cb625ba193ffd9e` is explicitly rejected).
2. Adapt content for this repository (`docs/development.md` validation tiers, real
   paths, governance fail-closed behavior).
3. Update `PROVENANCE.yaml` committed-file manifest hashes and adaptation notes.
4. Run `node --test scripts/foundation/voc112-agent-skills.test.mjs` and task-local
   shared-skill or Graphify tests as applicable.

Repository-native skills (for example `vocanova-repo-navigator`) declare
`source: repository-native` and list authoritative documentation sources in provenance.

## Graphify pilot limitations

The `graphify-pilot` skill is **opt-in only** (`disable-model-invocation: true`).
Ordinary agent sessions do not auto-invoke it. Graduation to always-on use requires a
separate governed change with Graphify-specific measurement and acceptance gates.

- **Explicit setup:** operators run `bash scripts/graphify/setup.sh` once; ordinary
  skill and runner use fail closed when the locked local runtime is missing.
- **Code-only:** extraction uses `--code-only`; query logging is disabled; provider
  credentials are unavailable to the Graphify process.
- **Hint-only output:** `graphify-out/` is gitignored by default; verify consequential
  claims in current source before acting.
- **No ordinary-use downloads:** `scripts/graphify/check` and `run.sh` do not install,
  upgrade, or mutate user profiles.

See `.agents/skills/graphify-pilot/SKILL.md` and `specs/changes/VOC-112-*/t03-evidence.md`.

## Safe use

Skills are instruction surfaces, not authority grants. They must not be used to:

- print, export, or grep environment secrets, `.env*` files, credentials, cookies,
  OAuth material, session tokens, or personal data;
- paste or export raw CI logs containing secrets;
- run unpinned global installs (`@latest`, `npm install -g`, etc.) or hidden network
  fetches outside repository tooling;
- mutate user-profile or global agent configuration outside this repository.

For governed implementation work, follow `AGENTS.md` (issue → plan → task lifecycle).
Use `vocanova-repo-navigator` to find authoritative paths; do not treat skill prose as
a substitute for canonical docs or approved change packages.
