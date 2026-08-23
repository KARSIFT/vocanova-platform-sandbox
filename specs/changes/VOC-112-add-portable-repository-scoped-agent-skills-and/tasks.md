# VOC-112 — Tasks

None of the tasks below is implementation-authorized by this package.
Adoption and each task's own implementation authorization are separate.
Order is mandatory: **T00 → (T01 ∥ T02 ∥ T03) → T04**.

## VOC-112-T00 — Canonical skill framework, adapters, provenance registry, and validation

- Requirement source: issue #933; `VOC-112-D00`–`D04`, `D07`
- Acceptance criteria: `VOC-112-AC-00`
- Tests: `VOC-112-TEST-00` through `VOC-112-TEST-05`
- Evidence: `VOC-112-EV-00` (`t00-evidence.md`)
- Status: pending

### Required work

1. Create `.agents/skills/README.md` documenting the one-source architecture, directory
   layout, frontmatter rules, reference-file conventions, and governance precedence
   (`VOC-112-D02`).
2. Create `.agents/skills/provenance.yaml` schema and initial entries for repository-native
   skills; adapted entries are added by later tasks but validation must exist in T00.
3. Define the Claude adapter contract in `.claude/skills/README.md` and implement a
   shared adapter template pattern: each adapter's entire procedural content is limited to
   loading `.agents/skills/<name>/SKILL.md`. No Git symlinks.
4. Add `scripts/foundation/voc112-agent-skills.test.mjs` enforcing:
   - canonical/adapters name parity;
   - allowed frontmatter;
   - adapter loader contract;
   - provenance completeness for skills marked adapted;
   - forbidden-pattern denylist (`VOC-112-D03`);
   - referenced files exist;
   - description and SKILL.md size budgets (`VOC-112-D04`).
5. Wire the test into existing `pnpm test` foundation coverage (no new workflow required
   unless classifier demands it for unrelated reasons).
6. Add `docs/development/agent-skills.md` skeleton sections (completed in T04) so T00 can
   link validation commands.
7. Record commands and results in `t00-evidence.md`:
   - `node --test scripts/foundation/voc112-agent-skills.test.mjs`;
   - `bash scripts/governance/validate-governance.sh` when AGENTS.md changes in this task;
   - `bash scripts/governance/classify-change-risk.sh`;
   - `git diff --check`.

### Explicitly out of scope for this task

- Full navigator content (T01), shared skill bodies (T02), Graphify runner (T03),
  benchmark harness (T04).
- `AGENTS.md` pointer (T04 unless T00 needs no AGENTS touch — prefer T04 single edit).
- Product/runtime/deployment changes.

## VOC-112-T01 — Add `vocanova-repo-navigator` router skill

- Requirement source: issue #933 first capability; `VOC-112-D05`
- Acceptance criteria: `VOC-112-AC-01`
- Tests: `VOC-112-TEST-06`, `VOC-112-TEST-07`
- Evidence: `VOC-112-EV-01` (`t01-evidence.md`)
- Status: pending — depends on `VOC-112-T00`

### Required work

1. Add `.agents/skills/vocanova-repo-navigator/SKILL.md` implementing the router in
   `VOC-112-D05` with progressive disclosure; optional `reference.md` for extended routes
   only if within budgets.
2. Add matching `.claude/skills/vocanova-repo-navigator/SKILL.md` adapter per T00 contract.
3. Register `source: repository-native` provenance entry.
4. Extend validation fixtures if needed for navigator-specific references (paths must exist
   at implementation time).
5. Record review notes in `t01-evidence.md`: confirm no large `AGENTS.md`/governance paste,
   budgets pass, validation green.

### Explicitly out of scope for this task

- Shared upstream skills (T02), Graphify (T03), benchmark (T04).

## VOC-112-T02 — Add seven pinned shared engineering skills

- Requirement source: issue #933 shared engineering skills; `VOC-112-D06`, `D07`
- Acceptance criteria: `VOC-112-AC-02`
- Tests: `VOC-112-TEST-08`, `VOC-112-TEST-09`
- Evidence: `VOC-112-EV-02` (`t02-evidence.md`)
- Status: pending — depends on `VOC-112-T00`

### Required work

1. For each required skill domain, choose an upstream source, pin exact commit/path, record
   license and `content_sha256` in `.agents/skills/provenance.yaml`, and adapt content into
   `.agents/skills/<skill-name>/SKILL.md`:
   - `context-mapping`
   - `systematic-debugging`
   - `verification-before-completion`
   - `github-actions-efficiency`
   - `react-next-performance`
   - `playwright-browser-testing`
   - `security-threat-modeling`
   Skill directory names may use kebab-case equivalents; validation must enforce a stable
   one-to-one mapping documented in provenance.
2. Adapt generic upstream instructions to this repository:
   - cite `docs/development.md` validation tiers instead of inventing commands;
   - forbid secret/credential/log exfiltration;
   - reference real paths (`apps/web`, `apps/api`, `scripts/foundation`, `.github/workflows`);
   - preserve fail-closed governance behavior.
3. Add Claude adapters for each skill.
4. Store optional upstream snapshot excerpts under `.agents/skills/<skill-name>/upstream/`
   only when needed for hash verification; do not duplicate entire upstream repos.
5. Record security-review summary and hash table in `t02-evidence.md` (no secrets/logs).
6. Re-run `voc112-agent-skills` and governance validation as applicable.

### Explicitly out of scope for this task

- Navigator changes except provenance/registry updates required for validation.
- Graphify pilot (T03).
- Installing unpinned global tools or fetching `@latest` dependencies during agent runs.

## VOC-112-T03 — Graphify pilot (code-only, opt-in)

- Requirement source: issue #933 Graphify pilot; `VOC-112-D08`
- Acceptance criteria: `VOC-112-AC-03`
- Tests: `VOC-112-TEST-10`, `VOC-112-TEST-11`
- Evidence: `VOC-112-EV-03` (`t03-evidence.md`)
- Status: pending — depends on `VOC-112-T00`

### Required work

1. Pin Graphify upstream (repository tag/commit) and runtime package version; record in
   `.agents/skills/provenance.yaml`.
2. Add repository-owned runner/check under `scripts/graphify/` (exact filenames at
   implementer discretion) that:
   - documents prerequisites explicitly;
   - runs code-only extraction with query logging disabled;
   - applies `.graphifyignore` excluding `.env*`, secrets, `node_modules`, build output,
     vendor/generated trees, runtime data, and large binary artifacts;
   - exits non-zero with clear message when runtime missing — **no global auto-install**.
3. Add opt-in skill `.agents/skills/graphify-pilot/SKILL.md` (+ Claude adapter) stating
   graph output is hint-only, verification must use current source, and automatic invocation
   is disabled until T04 evidence says otherwise.
4. Gitignore default graph output directories unless T03/T04 evidence explicitly justifies
   committing them (expected: remain ignored).
5. Add deterministic test coverage in `voc112-agent-skills.test.mjs` or sibling fixture for
   ignore rules, opt-in markers, and forbidden auto-enable patterns.
6. Record pilot limitations and pin table in `t03-evidence.md`.

### Explicitly out of scope for this task

- Document/image semantic extraction, LLM provider auto-detection, or committed graph
  artifacts without T04 justification.
- Product/runtime/deployment changes.

## VOC-112-T04 — Navigation benchmark, discovery evidence, and documentation completion

- Requirement source: issue #933 verification/documentation; `VOC-112-D09`, `D10`
- Acceptance criteria: `VOC-112-AC-04`
- Tests: `VOC-112-TEST-12`, `VOC-112-TEST-13`, `VOC-112-TEST-14`
- Evidence: `VOC-112-EV-04` (`t04-evidence.md`)
- Status: pending — depends on `VOC-112-T01`, `VOC-112-T02`, `VOC-112-T03`

### Required work

1. Add `scripts/foundation/voc112-navigation-benchmark.test.mjs` (or script + test wrapper)
   with a fixed set of representative navigation questions keyed to expected authoritative
   paths/commands. Record baseline vs navigator-assisted metrics in `t04-evidence.md`:
   files opened, search operations, elapsed time, correctness score, skill metadata
   character budget.
2. Demonstrate skill discovery from repository root and from a nested directory (recommended
   `apps/web/`). Record metadata-only evidence; if hosted runtime proof requires operator
   access, document what was observed without interactive secrets.
3. Complete `docs/development/agent-skills.md`: installation scope, one-source architecture,
   updating pinned upstream material, Graphify limitations, safe use, validation commands.
4. Add a short `AGENTS.md` subsection pointing to canonical skills and stating governance
   precedence — **do not weaken existing rules**.
5. Run full applicable validation:
   - `node --test scripts/foundation/voc112-*.test.mjs`;
   - `bash scripts/governance/validate-governance.sh`;
   - `bash scripts/governance/classify-change-risk.sh`;
   - `git diff --check`.
6. Confirm acceptance principle: one authoritative copy per skill, discovery without
   personal installation, deterministic drift/safety checks, measured navigation benefit
   or non-regression.

### Explicitly out of scope for this task

- Application code changes beyond documentation/agent tooling already introduced.
- Enabling Graphify automatic invocation without positive benchmark evidence.

## Task ordering notes

- T00 blocks all other tasks.
- T01, T02, and T03 may proceed in parallel after T00 merges.
- T04 requires T01–T03 so measurement covers the full skill set and Graphify pilot.
- No task may be dispatched before this package is adopted and implementation-authorized.

Tasks preserve scope, separation of duties, and rollback safety.
