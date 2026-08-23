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
2. Create immutable `.agents/skills/provenance.schema.yaml`. Each later skill owns a
   `.agents/skills/<name>/PROVENANCE.yaml`; there is no shared registry for parallel tasks
   to edit.
3. Define the Claude adapter contract in `.claude/skills/README.md` and implement a
   shared adapter template pattern: discovery metadata matches the canonical skill and the
   entire body is the exact one-line loader from `VOC-112-D00`, using
   `${CLAUDE_PROJECT_DIR}` to load `.agents/skills/<name>/SKILL.md`. No Git symlinks.
4. Add `scripts/foundation/voc112-agent-skills.test.mjs` enforcing:
   - canonical/adapters name parity;
   - allowed frontmatter;
   - adapter loader contract;
   - `${CLAUDE_PROJECT_DIR}` target resolution from root and nested cwd fixtures;
   - dynamically discovered per-skill provenance completeness;
   - forbidden-pattern denylist (`VOC-112-D03`);
   - referenced files exist;
   - description and SKILL.md size budgets (`VOC-112-D04`).
5. Wire the test into existing `pnpm test` foundation coverage (no new workflow required
   unless classifier demands it for unrelated reasons).
   The validator must be data-driven: T01/T02/T03 add conforming skill directories and do
   not edit the validator or a central registry simply to register themselves.
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
3. Add `.agents/skills/vocanova-repo-navigator/PROVENANCE.yaml` with
   `source: repository-native` and authoritative source paths.
4. Do not edit the T00 validator or shared framework files; navigator references must pass
   the generic validator as implemented. Put navigator-specific coverage in
   `scripts/foundation/voc112-navigator.test.mjs`.
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

1. For each required skill domain, choose an upstream source and create
   `.agents/skills/<skill-name>/SKILL.md` plus task-local `PROVENANCE.yaml`:
   - `context-mapping`
   - `systematic-debugging`
   - `verification-before-completion`
   - `github-actions-efficiency`
   - `react-next-performance`
   - `playwright-browser-testing`
   - `security-threat-modeling`
   Skill directory names may use kebab-case equivalents; validation must enforce a stable
   one-to-one mapping documented in provenance. Adapted records include exact upstream
   repository/commit/path, compatible license, `upstream_sha256`, `local_sha256`, retained
   license/NOTICE paths, adaptation notes, and a manifest covering every committed file.
2. Adapt generic upstream instructions to this repository:
   - cite `docs/development.md` validation tiers instead of inventing commands;
   - forbid secret/credential/log exfiltration;
   - reference real paths (`apps/web`, `apps/api`, `scripts/foundation`, `.github/workflows`);
   - preserve fail-closed governance behavior.
3. Add Claude adapters for each skill.
4. Do not edit T00's validator/schema or any central registry. Store optional upstream
   snapshot excerpts under `.agents/skills/<skill-name>/upstream/`
   only when needed for hash verification; do not duplicate entire upstream repos.
5. Review every committed instruction, script, reference, asset, license, and notice—not
   only SKILL.md or denylist matches. Record the security review and hash table in
   `t02-evidence.md` (no secrets/logs).
6. Record that `vercel-labs/agent-skills` commit
   `dd089a8c752c966dee8bf0f27cb625ba193ffd9e` was rejected as a React skill copy/adaptation
   source because it has no detected or committed license. Author React/Next guidance
   independently from official current React/Next documentation or use an explicitly
   licensed source.
7. Explicitly remove unsafe upstream behavior: no environment/credential greps; no raw CI
   log ingestion; no unpinned/global Playwright installs; no generic human-review pause;
   use repository-pinned Playwright and repository validation commands.
8. Re-run `voc112-agent-skills` and governance validation as applicable.
   Put domain/security assertions in task-local
   `scripts/foundation/voc112-shared-skills.test.mjs`; do not edit T00's validator.

### Explicitly out of scope for this task

- Navigator changes.
- Graphify pilot (T03).
- Installing unpinned global tools or fetching `@latest` dependencies during agent runs.

## VOC-112-T03 — Graphify pilot (code-only, opt-in)

- Requirement source: issue #933 Graphify pilot; `VOC-112-D08`
- Acceptance criteria: `VOC-112-AC-03`
- Tests: `VOC-112-TEST-10`, `VOC-112-TEST-11`
- Evidence: `VOC-112-EV-03` (`t03-evidence.md`)
- Status: pending — depends on `VOC-112-T00`

### Required work

1. Pin Graphify upstream repository/tag/commit, runtime package version, complete transitive
   dependency lock with hashes, and local adapted skill/runner hashes in
   `.agents/skills/graphify-pilot/PROVENANCE.yaml`; retain Apache-2.0 license/NOTICE files.
2. Add repository-owned runner/check under `scripts/graphify/` (exact filenames at
   implementer discretion) that:
   - provides a separate explicit operator setup into a repository-local isolated
     environment using only the reviewed lock/hash set;
   - normal use refuses a missing/mismatched environment and performs no download,
     upgrade, global fallback, profile mutation, installer call, hooks, or always-on
     instruction injection;
   - runs code-only extraction with query logging disabled and every provider credential
     unavailable to the Graphify process so backend auto-detection cannot occur;
   - applies `.graphifyignore` excluding `.env*`, secrets, `node_modules`, build output,
     vendor/generated trees, runtime data, and large binary artifacts;
   - exits non-zero with clear message when runtime missing — **no global auto-install**.
3. Add opt-in skill `.agents/skills/graphify-pilot/SKILL.md` (+ Claude adapter) stating
   graph output is hint-only, verification must use current source, and automatic invocation
   remains disabled throughout VOC-112. Any later enablement requires a separate governed
   change with Graphify-specific measurements and acceptance gates.
4. Gitignore default graph output directories unless T03/T04 evidence explicitly justifies
   committing them (expected: remain ignored).
5. Add deterministic coverage in task-local
   `scripts/foundation/voc112-graphify.test.mjs` for ignore rules, opt-in markers,
   lock/hash identity, and forbidden auto-enable/install/provider/hook patterns without
   editing the T00-owned generic validator.
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
   with fixed representative questions and expected authoritative paths/commands. Run
   controlled baseline and navigator-assisted agent sessions using the same exact revision,
   model/runtime version, permissions, and rubric. Derive files opened, search operations,
   elapsed time, correctness, and available context/usage metrics from structured traces;
   the deterministic test validates captured evidence and calculations and must not simulate
   or hardcode successful measurements.
2. Demonstrate actual hosted Cursor skill discovery/invocation from repository root and a
   nested directory (recommended `apps/web/`). Also exercise Claude Code from both locations
   when an already-authorized non-interactive runtime exists. Static filesystem assertions
   are not runtime proof; if Claude needs new interactive credentials, record
   `not-executed-external-credential-required` without claiming pass. Evidence remains
   metadata-only, with no prompts, raw logs, secrets, or personal data.
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
- Enabling Graphify automatic invocation. Graduation from the opt-in pilot is outside
  VOC-112 even if the navigator benchmark passes.

## Task ordering notes

- T00 blocks all other tasks.
- T01, T02, and T03 may proceed in parallel after T00 merges.
- Their write surfaces are disjoint: every skill owns its `PROVENANCE.yaml`; T00 owns the
  immutable provenance schema and data-driven validator. A task requiring a shared-file
  change must be serialized instead of creating predictable merge conflicts.
- T04 requires T01–T03 so measurement covers the full skill set and Graphify pilot.
- No task may be dispatched before this package is adopted and implementation-authorized.

Tasks preserve scope, separation of duties, and rollback safety.
