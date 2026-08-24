# VOC-112 — Implementation Plan

## Preconditions and protected areas

- Do not begin until the package is adopted and implementation-authorized.
- Protected areas: agent instructions (`AGENTS.md`, `CLAUDE.md`), repository governance
  semantics; skills must not override canonical policy or grant authority.
- Prerequisites: issue #933 requirement thread; existing foundation test harness in
  `scripts/foundation/`; `docs/development.md` validation documentation.

## File reconciliation and implementation sequence

### T00 — Framework, adapters, provenance, validation

| File | Action | Notes |
|------|--------|-------|
| `.agents/skills/README.md` | create | One-source architecture |
| `.agents/skills/provenance.schema.yaml` | create | Immutable per-skill record schema |
| `.claude/skills/README.md` | create | Adapter loader contract |
| `scripts/foundation/voc112-agent-skills.test.mjs` | create | Drift/safety validation |
| `docs/development/agent-skills.md` | create | Skeleton (completed T04) |
| `specs/changes/VOC-112-.../t00-evidence.md` | update | Validation results |

Ordered steps:

1. Land directory conventions, immutable provenance schema, and adapter template without
   skill bodies beyond README stubs if needed for validation bootstrapping.
2. Implement a data-driven validator with internal fixtures for forbidden patterns and
   budget violations. Later tasks must not edit this validator or a shared registry merely
   to register a skill; it discovers canonical skill directories and their local
   `PROVENANCE.yaml` records.
3. Wire into `pnpm test` via existing `node --test scripts/foundation/*.test.mjs` glob.
4. Record deterministic command results in `t00-evidence.md`.

### T01 — `vocanova-repo-navigator`

| File | Action | Notes |
|------|--------|-------|
| `.agents/skills/vocanova-repo-navigator/SKILL.md` | create | Router per `VOC-112-D05` |
| `.claude/skills/vocanova-repo-navigator/SKILL.md` | create | Loader adapter |
| `.agents/skills/vocanova-repo-navigator/PROVENANCE.yaml` | create | Native entry |
| `scripts/foundation/voc112-navigator.test.mjs` | create | Task-local navigator assertions |
| `t01-evidence.md` | update | Review + validation |

### T02 — Shared engineering skills

| File | Action | Notes |
|------|--------|-------|
| `.agents/skills/<seven-skills>/SKILL.md` | create | Adapted content |
| `.agents/skills/<seven-skills>/PROVENANCE.yaml` | create | Task-isolated pins, hashes, licenses |
| `.claude/skills/<seven-skills>/SKILL.md` | create | Adapters |
| `scripts/foundation/voc112-shared-skills.test.mjs` | create | Task-local domain/security assertions |
| `t02-evidence.md` | update | Security review + hashes |

T02 reviews every committed instruction, script, reference, asset, license, and notice.
For adapted material, provenance stores distinct upstream and local hashes and adaptation
notes. Do not copy/adapt `vercel-labs/agent-skills` React guidance at reviewed commit
`dd089a8c752c966dee8bf0f27cb625ba193ffd9e`: it has no detected or committed license.
Author that skill independently from current official React/Next documentation or choose an
explicitly licensed source, and record the rejection.

### T03 — Graphify pilot

| File | Action | Notes |
|------|--------|-------|
| `scripts/graphify/*` | create | Pinned runner/check |
| `.graphifyignore` | create | Secret/generated exclusions |
| `.agents/skills/graphify-pilot/SKILL.md` | create | Opt-in skill |
| `.agents/skills/graphify-pilot/PROVENANCE.yaml` | create | Package/upstream/local identities |
| `.claude/skills/graphify-pilot/SKILL.md` | create | Adapter |
| `.gitignore` | modify | Ignore default graph output (if not already) |
| `scripts/foundation/voc112-graphify.test.mjs` | create | Task-local runner/config assertions |
| `t03-evidence.md` | update | Pins + limitations |

T03 also commits a complete transitive lock with hashes and uses an isolated repository-local
environment. Environment setup is a separate explicit operator action. The normal skill and
runner do not download/upgrade packages, invoke Graphify's installer, use global executables,
auto-detect LLM providers, register hooks, inject always-on AGENTS/CLAUDE content, or mutate a
user profile. They fail closed when the local locked runtime identity is absent or mismatched.

### T04 — Benchmark, discovery, docs, AGENTS pointer

| File | Action | Notes |
|------|--------|-------|
| `scripts/foundation/voc112-navigation-benchmark.test.mjs` | create | Measurement harness |
| `docs/development/agent-skills.md` | update | Complete operator doc |
| `AGENTS.md` | modify | Short pointer; R3 review |
| `t04-evidence.md` | update | Metrics + discovery proof |

Ordered steps:

1. Run controlled baseline and navigator-assisted sessions with fixed prompts, keyed rubric,
   identical model/runtime/permissions/revision, and machine-derived structured metrics.
2. Record actual current hosted Cursor discovery/invocation evidence from root + nested cwd.
   Exercise Claude Code likewise only when already authorized non-interactively; otherwise
   record the external-credential limitation without claiming success.
3. Complete documentation and single AGENTS.md pointer edit.
4. Run governance validation and classify-change-risk on final task diff.

## Validation and independent verification

Deterministic (each task as applicable):

```bash
node --test scripts/foundation/voc112-agent-skills.test.mjs
node --test scripts/foundation/voc112-navigation-benchmark.test.mjs   # T04
bash scripts/governance/validate-governance.sh                          # when AGENTS.md/docs touched
bash scripts/governance/classify-change-risk.sh
git diff --check
pnpm test                                                               # ensure foundation glob includes new tests
```

Independent verifier binds to exact task PR SHA; confirms scope, tests, and evidence under
active A-004. Semantic review must confirm skills do not weaken governance or introduce
secret-exfiltration instructions.

## Deployment and rollback

- **Authorization boundary:** No staging/production deploy behavior change. Merging to
  `develop` only adds repository-local agent tooling and docs.
- **Rollout:** Tasks merge to `develop` in order T00 → parallel T01–T03 → T04. Parallel
  tasks have disjoint write surfaces: each skill owns its `PROVENANCE.yaml`, and T00 owns
  the immutable schema/validator.
- **Rollback trigger:** Validation fails in CI after merge; discovered unsafe skill content;
  benchmark shows navigator regression beyond declared threshold.
- **Mechanism:** Revert the offending task PR(s); remove or disable skills and adapters together
  to avoid adapter/canonical drift.
- **Owner:** unassigned (set at adoption).
- **Last-known-good:** `develop` HEAD before T00 merge (exact SHA recorded at task time).
