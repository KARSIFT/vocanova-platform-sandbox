# Repository agent skills (canonical tree)

Every authoritative agent skill procedure for this repository lives under
`.agents/skills/<skill-name>/`. Cursor, Codex, and compatible agents discover this
tree from the repository root without personal skill installation.

## One-source architecture

- **Canonical:** `.agents/skills/<skill-name>/SKILL.md` is the sole authoritative
  procedure. Optional supporting files live in the same directory and are referenced
  from `SKILL.md` only.
- **Claude adapters:** `.claude/skills/<skill-name>/SKILL.md` repeats all allowed
  discovery metadata exactly and loads the canonical file via the loader contract in
  `.claude/skills/README.md`. Adapters contain no independent procedure. No Git
  symlinks.
- **Governance precedence:** When skill prose conflicts with `AGENTS.md`,
  `CLAUDE.md`, approved change packages, tests, or source code, the repository
  sources win.

## Directory layout

```
.agents/skills/
  README.md                 # this file
  provenance.schema.yaml    # immutable per-skill record schema (T00-owned)
  <skill-name>/
    SKILL.md                # required canonical procedure
    PROVENANCE.yaml         # required per-skill provenance record
    <optional-supporting-files>
```

## YAML frontmatter (`SKILL.md`)

Required keys:

| Key | Rules |
| --- | --- |
| `name` | Kebab-case skill directory name; must match the parent directory |
| `description` | Short discovery summary; max **512** characters |

Optional keys (documented; validation allows only these beyond required):

| Key | Purpose |
| --- | --- |
| `disable-model-invocation` | When `true`, the skill is opt-in only (e.g. Graphify pilot) |

No other frontmatter keys are permitted without a governed schema change.

## Reference files

- Reference sibling files with relative markdown links from `SKILL.md`, e.g.
  `[routes](reference.md)`.
- Referenced paths must exist under the same skill directory.
- Keep the opening `SKILL.md` section compact; move depth into referenced files.
- Supporting files outside `SKILL.md` are excluded from startup metadata budgets;
  linked content is loaded only when the canonical procedure directs the agent to it.

## Startup metadata budgets

Enforced by `scripts/foundation/voc112-agent-skills.test.mjs`:

| Surface | Limit |
| --- | --- |
| `description` | 512 characters |
| `SKILL.md` body (after frontmatter) | 32,768 bytes and 400 lines |

The forbidden-instruction denylist scans `SKILL.md`, its Claude adapter, and every
text-like supporting artifact in the canonical skill directory. Binary artifacts are
hash-validated through provenance but are not decoded as instructions.

## Provenance

Each skill owns `.agents/skills/<skill-name>/PROVENANCE.yaml` conforming to
`provenance.schema.yaml`. There is no shared skill registry; parallel tasks add
records without editing T00's validator or schema. `committed_files` hashes every
skill artifact except `PROVENANCE.yaml` itself, because a file cannot contain its own
stable digest. For adapted skills, `local_sha256` is the digest of the canonical
adapted `SKILL.md`; retained license and NOTICE files must also appear in the manifest.

## Validation

```bash
node --test scripts/foundation/voc112-agent-skills.test.mjs
```

Full workspace validation includes this test via `pnpm test`.

## Operator documentation

See `docs/development/agent-skills.md` for installation scope, updates, and safe use.
