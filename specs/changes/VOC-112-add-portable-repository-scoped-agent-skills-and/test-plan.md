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
- Procedure: Validation parses Claude adapters and rejects procedural steps beyond the
  documented loader pattern pointing at `.agents/skills/<name>/SKILL.md`.
- Expected result: Adapters with independent procedures fail closed
- Evidence: `VOC-112-EV-00`

## VOC-112-TEST-03 — Forbidden-pattern denylist

- Covers: `VOC-112-AC-00`, `VOC-112-D03`
- Preconditions: T00 denylist fixtures
- Procedure: Validation scans canonical skills and adapters for prohibited secret exfiltration,
  unpinned install, and hidden network patterns using positive and negative fixtures.
- Expected result: Negative fixtures fail; clean repository skills pass
- Evidence: `VOC-112-EV-00`

## VOC-112-TEST-04 — Provenance registry completeness

- Covers: `VOC-112-AC-00`, `VOC-112-D07`
- Preconditions: `.agents/skills/provenance.yaml` from T00; adapted entries added in T02/T03
- Procedure: Validation requires provenance fields for every non-native skill and matching skill
  directory name.
- Expected result: Missing or stale hash/commit records fail closed
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
- Procedure: Read `.agents/skills/vocanova-repo-navigator/SKILL.md`; assert routing table
  includes web, API, database, auth/OAuth, seed, deploy, monitoring, governance, testing,
  validation tiers, shared-edge invariants, and issue→plan→task lifecycle pointers.
- Expected result: All domains present with authoritative path targets that exist in repo
- Evidence: `VOC-112-EV-01`

## VOC-112-TEST-07 — Navigator does not duplicate governance corpora

- Covers: `VOC-112-AC-01`, `VOC-112-D02`
- Preconditions: T01 navigator skill
- Procedure: Assert navigator SKILL.md byte size and content heuristics stay below documented
  router budget; reject inline paste of full `AGENTS.md` or governance amendment text beyond
  short pointers.
- Expected result: Router remains compact; governance precedence statement present
- Evidence: `VOC-112-EV-01`

## VOC-112-TEST-08 — Exactly seven shared skills with adapters

- Covers: `VOC-112-AC-02`, `VOC-112-D06`
- Preconditions: T02 merged
- Procedure: Count adapted shared skills in provenance registry; assert count equals seven and
  each has canonical + Claude adapter entries.
- Expected result: No missing/extra shared skills
- Evidence: `VOC-112-EV-02`

## VOC-112-TEST-09 — Shared skills include repository adaptations and pins

- Covers: `VOC-112-AC-02`, `VOC-112-D07`
- Preconditions: T02 merged
- Procedure: For each shared skill, verify provenance hash/commit/path/license fields,
  repository-specific validation references, and absence of forbidden patterns.
- Expected result: Unpinned or unadapted upstream copy fails validation
- Evidence: `VOC-112-EV-02`

## VOC-112-TEST-10 — Graphify pilot opt-in and code-only configuration

- Covers: `VOC-112-AC-03`, `VOC-112-D08`
- Preconditions: T03 merged
- Procedure: Inspect graphify skill, runner script, and ignore file; assert code-only mode,
  logging disabled flags, opt-in/disabled-auto-invocation markers, and no committed graph output
  by default.
- Expected result: Semantic/doc extraction paths not enabled; graph dirs gitignored
- Evidence: `VOC-112-EV-03`

## VOC-112-TEST-11 — Graphify runner fails safely without global install

- Covers: `VOC-112-AC-03`
- Preconditions: T03 runner on CI runner without Graphify installed (expected default)
- Procedure: Execute repository check script; assert non-zero exit with prerequisite message
  and no attempted global installation.
- Expected result: Fail-closed, explicit prerequisite guidance
- Evidence: `VOC-112-EV-03`

## VOC-112-TEST-12 — Navigation benchmark baseline vs navigator-assisted

- Covers: `VOC-112-AC-04`, `VOC-112-D09`
- Preconditions: T04 benchmark script; T01 navigator present
- Procedure: Run `node --test scripts/foundation/voc112-navigation-benchmark.test.mjs`;
  read `t04-evidence.md` metric table for representative questions.
- Expected result: Navigator-assisted path improves or does not regress correctness and cost
  metrics versus baseline thresholds declared in the benchmark script
- Evidence: `VOC-112-EV-04`

## VOC-112-TEST-13 — Discovery from root and nested working directory

- Covers: `VOC-112-AC-04`, `VOC-112-D10`
- Preconditions: T04 evidence recorded
- Procedure: Verify evidence documents skill discovery/listing from repository root and nested
  directory (recommended `apps/web/`) for the active hosted agent runtime or documented equivalent
  deterministic check.
- Expected result: Supported agents can find canonical skills without personal installation
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
