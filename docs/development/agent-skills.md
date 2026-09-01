# Agent skills

Repository-scoped agent skills give every supported agent one authoritative copy of
each procedure from a fresh checkout, with deterministic safety validation.

## Installation scope

Skills and agents ship in-repo under `.agents/skills/` and `.agents/agents/` — the
one canonical source. No personal or global installation is required for discovery
from the repository root. Cursor, Codex, and compatible agents read the canonical
tree directly; `.claude/skills`, `.claude/agents`, and `.opencode/agents` are plain
symlinks to it, so there's nothing separate to keep in sync. `.codex/agents/` and
`.cursor/rules/` hold real, tool-specific translations where a tool's own format
can't just be a symlink.

## One-source architecture

- **Canonical tree:** `.agents/skills/<skill-name>/SKILL.md`, `.agents/agents/`
- **Architecture reference:** `.agents/skills/README.md`
- **Repository navigator:** `.agents/skills/vocanova-repo-navigator/SKILL.md`

Governance precedence: when skill prose conflicts with `AGENTS.md`, `CLAUDE.md`,
approved change packages, tests, or source code, the **repository sources win**.

## Validation

```bash
node --test scripts/foundation/voc112-agent-skills.test.mjs
pnpm test   # includes foundation tests
```

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

For implementation work, follow `AGENTS.md` (open a PR against `main`, checks run,
tag `@claude` for review). Use `vocanova-repo-navigator` to find authoritative paths;
do not treat skill prose as a substitute for canonical docs.
