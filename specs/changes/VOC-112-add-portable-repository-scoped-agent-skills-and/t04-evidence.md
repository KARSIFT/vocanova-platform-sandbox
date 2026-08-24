# VOC-112-T04 evidence — navigation benchmark, discovery, and documentation

Metadata only: no secrets, prompts, raw logs, response bodies, or personal data.

## gate_status

`complete`

The benchmark plus Claude Code and hosted Cursor root/nested discovery are complete.
Filesystem enumeration, static path resolution, model prose, and path mentions are not
accepted as proof; the captured Cursor rows require completed read-tool events for the
canonical skill.

## Real navigation benchmark

Capture kind: `real-agent-structured-trace`. Subject revision:
`d29b179db2eb32a7ef8cc21156bb797be5d3fdd9`. Runtime: Codex CLI `0.149.0`, model
`gpt-5.6-sol`, read-only sandbox, ephemeral sessions, personal configuration ignored.
Three fixed questions were each run once as baseline and once with the canonical
`vocanova-repo-navigator` explicitly invoked. Pair order alternated. Raw traces existed
only in process memory; the committed fixture contains sanitized derived metrics.

| Metric                                                           | Baseline  | Navigator-assisted | Delta      |
| ---------------------------------------------------------------- | --------- | ------------------ | ---------- |
| Keyed correct answers                                            | 1 / 3     | 3 / 3              | +2         |
| Readable repository files referenced by completed command events | 13        | 3                  | -10        |
| Search operations                                                | 6         | 3                  | -3         |
| Tool calls                                                       | 7         | 6                  | -1         |
| Elapsed time                                                     | 65,182 ms | 48,757 ms          | -16,425 ms |
| Input tokens                                                     | 215,002   | 180,897            | -34,105    |
| Output tokens                                                    | 1,420     | 775                | -645       |

The keyed threshold requires full navigator correctness, correctness non-regression,
no increase in repository files/searches, and at most one additional skill-read tool call
per question. The capture passes all thresholds and also improves every recorded aggregate.
Because this Codex CLI version emits repository reads inside structured completed-command
events rather than as a separate first-class read event, the file metric counts only
existing readable repository paths referenced by those completed events. It is an explicit
trace-derived proxy, not a claim that every runtime read is observable.

Revision validation is fail closed and squash-aware in Repository Governance. Pull-request
runs require every captured commit, prove ancestry, and compare captured/current hashes.
Branch-push runs after the repository's squash merge re-validate the current hashes without
requiring discarded intermediate PR commits to remain ancestors. Generic shallow
application CI is offline and non-mutating.

Fixture: `scripts/foundation/fixtures/voc112-navigation-benchmark-traces.json`.

## Runtime discovery

| Runtime                                   | Context               | Evidence                                                                          | Result |
| ----------------------------------------- | --------------------- | --------------------------------------------------------------------------------- | ------ |
| Claude Code `2.1.229` / `claude-sonnet-5` | repository root       | structured `Read` event for canonical navigator plus success marker               | pass   |
| Claude Code `2.1.229` / `claude-sonnet-5` | nested `apps/web` cwd | structured `Read` event resolves the same canonical navigator plus success marker | pass   |
| Hosted Cursor `2026.08.11-e8db854` / Auto | repository root       | 26 structured events; 2 completed read calls including the canonical navigator    | pass   |
| Hosted Cursor `2026.08.11-e8db854` / Auto | nested `apps/web` cwd | 34 structured events; completed read call for the same canonical navigator        | pass   |

Fixture: `scripts/foundation/fixtures/voc112-skill-discovery-evidence.json`.
Each runtime row carries its own capture timestamp, subject revision, and source hashes;
later runtime captures preserve earlier pass, fail, or external-credential-limitation rows.
The sanitized hosted capture came from workflow run `32673904614` against subject
revision `00f220bbbd4bd5db51cb94912453227bae54fa3d`; the evidence-only workflow carrier
was removed before the final PR head.

## One-source architecture and documentation

- Canonical instructions: `.agents/skills/<name>/SKILL.md`.
- Claude Code adapters: `.claude/skills/<name>/SKILL.md`; frontmatter is mirrored and
  the sole body loads the canonical file through `${CLAUDE_PROJECT_DIR}`.
- No per-device skill installation is required after checkout.
- `docs/development/agent-skills.md` documents scope, updates, explicit evidence capture,
  Graphify limitations, and safe use.
- `AGENTS.md` points to the documentation without weakening A-004 or safety rules.

## Validation

| Command                                                               | Result                                                                                                                                                                                |
| --------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `node --test scripts/foundation/voc112-agent-skills.test.mjs`         | pass                                                                                                                                                                                  |
| `node --test scripts/foundation/voc112-navigation-benchmark.test.mjs` | 8/8 pass                                                                                                                                                                              |
| synthetic squash-merge reproduction                                   | squash-safe branch-push mode 8/8 pass; PR-ancestry mode correctly rejects the non-ancestor squash                                                                                     |
| `node --test scripts/foundation/voc112-*.test.mjs`                    | pass                                                                                                                                                                                  |
| `bash scripts/governance/validate-governance.sh`                      | pass                                                                                                                                                                                  |
| `bash scripts/governance/classify-change-risk.sh`                     | R4 after adding strict full-history validation to Repository Governance                                                                                                               |
| `git diff --check`                                                    | pass                                                                                                                                                                                  |
| `pnpm validate`                                                       | executed locally; stopped only at the two Docker-backed OAuth API tests because Docker is unavailable in this WSL checkout; all preceding phases passed; exact-SHA hosted CI required |
| `pnpm build`                                                          | pass (packages, web, API)                                                                                                                                                             |

## Acceptance mapping

- `VOC-112-AC-04` / `VOC-112-EV-04` — complete.

T04's final semantic/path risk is R4 because the evidence hardening updates
`.github/workflows/repository-governance.yml`; application and deployment behavior remain
unchanged. The workflow uses its existing full-history checkout and a pinned Node action,
so strict provenance validation is offline and does not mutate the checkout.
