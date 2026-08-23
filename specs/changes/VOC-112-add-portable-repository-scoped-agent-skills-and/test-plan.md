# VOC-112 — Test Plan

## VOC-112-TEST-00 — Canonical/adapters directory contract

- Covers: `VOC-112-AC-00`, `VOC-112-D00`, `VOC-112-D01`
- Preconditions: T00 task branch
- Procedure: Run `node --test scripts/foundation/voc112-agent-skills.test.mjs`; assert every
  `.agents/skills/*/SKILL.md` has a matching `.claude/skills/*/SKILL.md` and no symlink entries
  exist under either tree.
- Expected result: Test exits 0; adapter/canonical names match one-to-one
- Evidence: `VOC-112-EV-00`

## VOC-112-TEST-01 — Frontmatter and reference integrity

- Covers: `VOC-112-AC-00`, `VOC-112-D04`
- Preconditions: T00 validation present
- Procedure: Validation asserts required YAML frontmatter keys and that every relative file
  referenced from a skill exists.
- Expected result: Missing/invalid frontmatter or broken references fail the test
- Evidence: `VOC-112-EV-00`

## VOC-112-TEST-02 — Adapter loader-only contract

- Covers: `VOC-112-AC-00`, `VOC-112-D00`
- Preconditions: T00 adapters present
- Procedure: Validation parses Claude adapters, requires exact canonical discovery-metadata
  parity and the exact one-line `${CLAUDE_PROJECT_DIR}` loader from `VOC-112-D00`, resolves
  the target under root and nested-cwd fixtures, and rejects any additional procedure.
- Expected result: Wrong targets, metadata drift, or independent adapter procedures fail closed
- Evidence: `VOC-112-EV-00`

## VOC-112-TEST-03 — Forbidden-pattern denylist

- Covers: `VOC-112-AC-00`, `VOC-112-D03`
- Preconditions: T00 denylist fixtures
- Procedure: Validation scans canonical skills and adapters for prohibited secret exfiltration,
  unpinned install, and hidden network patterns using positive and negative fixtures.
- Expected result: Negative fixtures fail; clean repository skills pass
- Evidence: `VOC-112-EV-00`

## VOC-112-TEST-04 — Task-isolated provenance completeness

- Covers: `VOC-112-AC-00`, `VOC-112-D07`
- Preconditions: `.agents/skills/provenance.schema.yaml` from T00; every skill carries its
  own `PROVENANCE.yaml`
- Procedure: Data-driven validation discovers every canonical skill without later tasks
  editing a shared registry or validator. It requires separate upstream/local hashes,
  compatible license/retained notices/adaptation notes for adapted skills and authoritative
  sources for native skills; the committed-file manifest must cover all skill artifacts.
- Expected result: Missing/stale records, uncovered files, ambiguous hashes, or a shared
  parallel-task registry fail closed
- Evidence: `VOC-112-EV-00`, `VOC-112-EV-02`, `VOC-112-EV-03`

## VOC-112-TEST-05 — Context-size budgets

- Covers: `VOC-112-AC-00`, `VOC-112-D04`
- Preconditions: T00 budgets documented in validation script
- Procedure: Assert each skill `description` and SKILL.md body remain under declared caps;
  optional reference files excluded from startup budget when not linked from frontmatter.
- Expected result: Oversized skills fail validation
- Evidence: `VOC-112-EV-00`

## VOC-112-TEST-06 — Navigator covers required domains

- Covers: `VOC-112-AC-01`, `VOC-112-D05`
- Preconditions: T01 navigator skill merged
- Procedure: Run `node --test scripts/foundation/voc112-navigator.test.mjs`; assert the
  `.agents/skills/vocanova-repo-navigator/SKILL.md` routing table
  includes web, API, database, auth/OAuth, seed, deploy, monitoring, governance, testing,
  validation tiers, shared-edge invariants, and issue→plan→task lifecycle pointers.
- Expected result: All domains present with authoritative path targets that exist in repo
- Evidence: `VOC-112-EV-01`

## VOC-112-TEST-07 — Navigator does not duplicate governance corpora

- Covers: `VOC-112-AC-01`, `VOC-112-D02`
- Preconditions: T01 navigator skill
- Procedure: The task-local navigator test asserts SKILL.md byte size and content
  heuristics stay below the documented router budget; reject inline paste of full
  `AGENTS.md` or governance amendment text beyond short pointers.
- Expected result: Router remains compact; governance precedence statement present
- Evidence: `VOC-112-EV-01`

## VOC-112-TEST-08 — Exactly seven shared skills with adapters

- Covers: `VOC-112-AC-02`, `VOC-112-D06`
- Preconditions: T02 merged
- Procedure: Run `node --test scripts/foundation/voc112-shared-skills.test.mjs`; count shared
  skills from their task-local provenance records; assert count equals seven and
  each has canonical + Claude adapter entries.
- Expected result: No missing/extra shared skills
- Evidence: `VOC-112-EV-02`

## VOC-112-TEST-09 — Shared skills include repository adaptations and pins

- Covers: `VOC-112-AC-02`, `VOC-112-D07`
- Preconditions: T02 merged
- Procedure: The task-local shared-skill test verifies for each skill its provenance
  hash/commit/path/license fields,
  separate upstream/local hashes, retained notices, full committed-file review,
  repository-specific validation references, and absence of forbidden patterns. Assert the
  unlicensed Vercel React source is recorded as rejected and its text is not vendored.
- Expected result: Unpinned, unlicensed, ambiguously hashed, incompletely reviewed, or
  unadapted upstream copy fails validation
- Evidence: `VOC-112-EV-02`

## VOC-112-TEST-10 — Graphify pilot opt-in and code-only configuration

- Covers: `VOC-112-AC-03`, `VOC-112-D08`
- Preconditions: T03 merged
- Procedure: Run `node --test scripts/foundation/voc112-graphify.test.mjs`; inspect the
  graphify skill, runner script, and ignore file; assert code-only mode,
  logging disabled flags, opt-in/disabled-auto-invocation markers, complete transitive
  lock/hashes, isolated local environment, no global fallback/provider auto-detection/hooks/
  always-on injection, and no committed graph output by default.
- Expected result: Semantic/doc extraction and ordinary-use downloads are not enabled;
  graph dirs are gitignored
- Evidence: `VOC-112-EV-03`

## VOC-112-TEST-11 — Graphify runner fails safely without global install

- Covers: `VOC-112-AC-03`
- Preconditions: T03 runner with the repository-local locked environment absent, mismatched,
  and valid in separate fixtures/runs
- Procedure: The task-local Graphify test asserts absent/mismatched cases exit non-zero
  with prerequisite guidance and no network/profile/global/provider/hook action; a valid
  locked fixture proves the identity check reaches code-only invocation without exposing
  provider credentials.
- Expected result: Fail-closed prerequisite behavior and exact locked identity enforcement
- Evidence: `VOC-112-EV-03`

## VOC-112-TEST-12 — Navigation benchmark baseline vs navigator-assisted

- Covers: `VOC-112-AC-04`, `VOC-112-D09`
- Preconditions: T04 benchmark script; T01 navigator present
- Procedure: Run controlled same-revision/model/runtime baseline and navigator-assisted
  sessions for fixed questions, then run
  `node --test scripts/foundation/voc112-navigation-benchmark.test.mjs` against their
  sanitized structured traces and keyed rubric.
- Expected result: Captured—not simulated or hardcoded—metrics prove navigator-assisted
  navigation improves or does not regress correctness and cost thresholds
- Evidence: `VOC-112-EV-04`

## VOC-112-TEST-13 — Discovery from root and nested working directory

- Covers: `VOC-112-AC-04`, `VOC-112-D10`
- Preconditions: T04 evidence recorded
- Procedure: Verify current hosted Cursor actually lists/invokes a repository skill from the
  root and nested directory (`apps/web/` recommended). Exercise an already-authorized
  non-interactive Claude Code runtime likewise; when that would require new interactive
  credentials, require the explicit limited result instead of a pass. Reject static layout
  assertions as runtime discovery evidence.
- Expected result: Cursor runtime discovery passes in both locations; Claude passes where
  executable without new credentials or is truthfully recorded as externally limited
- Evidence: `VOC-112-EV-04`

## VOC-112-TEST-14 — Documentation and AGENTS.md pointer preserve governance precedence

- Covers: `VOC-112-AC-04`, `VOC-112-D02`
- Preconditions: T04 doc updates
- Procedure: Read `docs/development/agent-skills.md` and `AGENTS.md` delta; assert pointer
  exists, installation scope documented, and no weakening of governance/security language.
  Run `bash scripts/governance/validate-governance.sh`.
- Expected result: Governance validation passes; precedence language intact
- Evidence: `VOC-112-EV-04`

Include positive, negative, authorization, failure, migration, accessibility, and
rollback coverage as applicable. Tests must not use secrets or production data.
