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
| `.agents/skills/provenance.yaml` | create | Registry schema + native entries |
| `.claude/skills/README.md` | create | Adapter loader contract |
| `scripts/foundation/voc112-agent-skills.test.mjs` | create | Drift/safety validation |
| `docs/development/agent-skills.md` | create | Skeleton (completed T04) |
| `specs/changes/VOC-112-.../t00-evidence.md` | update | Validation results |

Ordered steps:

1. Land directory conventions and adapter template without skill bodies beyond README stubs
   if needed for validation bootstrapping.
2. Implement validation with fixtures for forbidden patterns and budget violations.
3. Wire into `pnpm test` via existing `node --test scripts/foundation/*.test.mjs` glob.
4. Record deterministic command results in `t00-evidence.md`.

### T01 — `vocanova-repo-navigator`

| File | Action | Notes |
|------|--------|-------|
| `.agents/skills/vocanova-repo-navigator/SKILL.md` | create | Router per `VOC-112-D05` |
| `.claude/skills/vocanova-repo-navigator/SKILL.md` | create | Loader adapter |
| `.agents/skills/provenance.yaml` | update | Native entry |
| `t01-evidence.md` | update | Review + validation |

### T02 — Shared engineering skills

| File | Action | Notes |
|------|--------|-------|
| `.agents/skills/<seven-skills>/SKILL.md` | create | Adapted content |
| `.claude/skills/<seven-skills>/SKILL.md` | create | Adapters |
| `.agents/skills/provenance.yaml` | update | Upstream pins/hashes |
| `t02-evidence.md` | update | Security review + hashes |

### T03 — Graphify pilot

| File | Action | Notes |
|------|--------|-------|
| `scripts/graphify/*` | create | Pinned runner/check |
| `.graphifyignore` | create | Secret/generated exclusions |
| `.agents/skills/graphify-pilot/SKILL.md` | create | Opt-in skill |
| `.claude/skills/graphify-pilot/SKILL.md` | create | Adapter |
| `.gitignore` | modify | Ignore default graph output (if not already) |
| `t03-evidence.md` | update | Pins + limitations |

### T04 — Benchmark, discovery, docs, AGENTS pointer

| File | Action | Notes |
|------|--------|-------|
| `scripts/foundation/voc112-navigation-benchmark.test.mjs` | create | Measurement harness |
| `docs/development/agent-skills.md` | update | Complete operator doc |
| `AGENTS.md` | modify | Short pointer; R3 review |
| `t04-evidence.md` | update | Metrics + discovery proof |

Ordered steps:

1. Run benchmark with T01–T03 artifacts present.
2. Record discovery evidence root + nested cwd.
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
- **Rollout:** Tasks merge to `develop` in order T00 → parallel T01–T03 → T04.
- **Rollback trigger:** Validation fails in CI after merge; discovered unsafe skill content;
  benchmark shows navigator regression beyond declared threshold.
- **Mechanism:** Revert the offending task PR(s); remove or disable skills and adapters together
  to avoid adapter/canonical drift.
- **Owner:** unassigned (set at adoption).
- **Last-known-good:** `develop` HEAD before T00 merge (exact SHA recorded at task time).
