# VOC-112-T04 evidence — navigation benchmark, discovery, and documentation

Metadata only: no secrets, prompts, raw logs, response bodies, or personal data.

## gate_status

`hosted-cursor-discovery-pending`

The benchmark and Claude Code discovery are complete. T04 must remain draft until an
authorized hosted Cursor runtime captures both root and nested discovery from structured
runtime events. Filesystem enumeration or static path resolution is not accepted as proof.

## Real navigation benchmark

Capture kind: `real-agent-structured-trace`. Subject revision:
`d29b179db2eb32a7ef8cc21156bb797be5d3fdd9`. Runtime: Codex CLI `0.149.0`, model
`gpt-5.6-sol`, read-only sandbox, ephemeral sessions, personal configuration ignored.
Three fixed questions were each run once as baseline and once with the canonical
`vocanova-repo-navigator` explicitly invoked. Pair order alternated. Raw traces existed
only in process memory; the committed fixture contains sanitized derived metrics.

| Metric                  | Baseline  | Navigator-assisted | Delta      |
| ----------------------- | --------- | ------------------ | ---------- |
| Keyed correct answers   | 1 / 3     | 3 / 3              | +2         |
| Repository files opened | 13        | 3                  | -10        |
| Search operations       | 6         | 3                  | -3         |
| Tool calls              | 7         | 6                  | -1         |
| Elapsed time            | 65,182 ms | 48,757 ms          | -16,425 ms |
| Input tokens            | 215,002   | 180,897            | -34,105    |
| Output tokens           | 1,420     | 775                | -645       |

The keyed threshold requires full navigator correctness, correctness non-regression,
no increase in repository files/searches, and at most one additional skill-read tool call
per question. The capture passes all thresholds and also improves every recorded aggregate.

Fixture: `scripts/foundation/fixtures/voc112-navigation-benchmark-traces.json`.

## Runtime discovery

| Runtime                                   | Context               | Evidence                                                                          | Result  |
| ----------------------------------------- | --------------------- | --------------------------------------------------------------------------------- | ------- |
| Claude Code `2.1.229` / `claude-sonnet-5` | repository root       | structured `Read` event for canonical navigator plus success marker               | pass    |
| Claude Code `2.1.229` / `claude-sonnet-5` | nested `apps/web` cwd | structured `Read` event resolves the same canonical navigator plus success marker | pass    |
| Hosted Cursor                             | repository root       | real structured capture required                                                  | pending |
| Hosted Cursor                             | nested `apps/web` cwd | real structured capture required                                                  | pending |

Fixture: `scripts/foundation/fixtures/voc112-skill-discovery-evidence.json`.

## One-source architecture and documentation

- Canonical instructions: `.agents/skills/<name>/SKILL.md`.
- Claude Code adapters: `.claude/skills/<name>/SKILL.md`; frontmatter is mirrored and
  the sole body loads the canonical file through `${CLAUDE_PROJECT_DIR}`.
- No per-device skill installation is required after checkout.
- `docs/development/agent-skills.md` documents scope, updates, explicit evidence capture,
  Graphify limitations, and safe use.
- `AGENTS.md` points to the documentation without weakening A-004 or safety rules.

## Validation

| Command                                                               | Result                                   |
| --------------------------------------------------------------------- | ---------------------------------------- |
| `node --test scripts/foundation/voc112-agent-skills.test.mjs`         | pending final head                       |
| `node --test scripts/foundation/voc112-navigation-benchmark.test.mjs` | intentionally pending hosted Cursor rows |
| `node --test scripts/foundation/voc112-*.test.mjs`                    | pending final head                       |
| `bash scripts/governance/validate-governance.sh`                      | pending final head                       |
| `bash scripts/governance/classify-change-risk.sh`                     | pending final head                       |
| `git diff --check`                                                    | pending final head                       |

## Acceptance mapping

- `VOC-112-AC-04` / `VOC-112-EV-04` — pending only the two real hosted Cursor discovery rows.
