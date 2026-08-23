# Claude skill adapters

Claude Code discovers skills under `.claude/skills/`. This repository uses a
**loader-only adapter** pattern: every adapter repeats canonical discovery metadata
and instructs Claude to load exactly one file from the canonical tree.

## Loader contract (VOC-112-D00 / VOC-112-D01)

For each skill named `<name>`:

1. Create `.claude/skills/<name>/SKILL.md`.
2. YAML frontmatter `name` and `description` must match
   `.agents/skills/<name>/SKILL.md` exactly.
3. The adapter body must be **exactly** this single line (replace `<name>`):

   ```
   Load and follow the sole canonical procedure at ${CLAUDE_PROJECT_DIR}/.agents/skills/<name>/SKILL.md completely.
   ```

4. No additional procedure, steps, or commentary in the adapter body.
5. No Git symlinks between adapter and canonical trees.

`${CLAUDE_PROJECT_DIR}` resolves to the repository root regardless of the agent's
current working directory inside the checkout.

## Governance precedence

When adapter or canonical skill prose conflicts with `AGENTS.md`, `CLAUDE.md`,
approved change packages, tests, or source code, the repository sources win.

## Validation

Adapter parity and loader shape are enforced by:

```bash
node --test scripts/foundation/voc112-agent-skills.test.mjs
```
