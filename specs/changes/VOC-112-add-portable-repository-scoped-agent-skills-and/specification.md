# VOC-112 — Add portable repository-scoped agent skills and codebase navigation: Specification

## Objective and requirement source

Establish a portable, repository-scoped agent-skill framework so a fresh checkout
gives every supported agent one authoritative skill copy, deterministic safety checks,
a focused repository navigator, adapted shared engineering skills, and an optional
Graphify code-index pilot — without changing product runtime or deployment behavior.

**Requirement source:** [GitHub issue #933](https://github.com/KARSIFT/vocanova-platform-sandbox/issues/933).

**This draft package does not adopt or authorize itself**; adoption remains a separate
A-004 plan-review / adopt path.

## Scope and non-goals

### In scope

1. **Canonical skill tree** — every authoritative skill lives at
   `.agents/skills/<skill-name>/SKILL.md` with optional supporting files in the same
   directory only when referenced from the canonical skill.
2. **Cross-agent discovery** — Codex, Cursor, and compatible agents discover the
   canonical tree directly from the repository root without personal skill installation.
3. **Claude adapters** — for each canonical skill, a minimal adapter at
   `.claude/skills/<skill-name>/SKILL.md` that instructs the agent to load exactly one
   canonical file and contains **no independent procedure**. No Git symlinks.
4. **Deterministic validation** — foundation tests (and optional lightweight script
   invoked from them) enforce:
   - allowed YAML frontmatter (`name`, `description`; optional fields documented);
   - adapter/canonical name parity;
   - adapter loader contract (points at canonical path; no procedural steps beyond load);
   - task-isolated `PROVENANCE.yaml` completeness for every canonical skill;
   - forbidden instruction patterns (secret/credential search, raw CI log exfiltration,
     `latest`/unpinned downloads, hidden network actions, user-profile mutation);
   - referenced-file existence;
   - startup metadata size budgets (description length and SKILL.md byte/line caps).
5. **`vocanova-repo-navigator`** — first repository-specific skill; routes agents to
   the smallest relevant context for:
   - web (`apps/web/`), API (`apps/api/`), database/migrations, authentication/OAuth,
     content seed, deployment/infra, monitoring, governance, and testing;
   - repository-owned commands and validation tiers (`docs/development.md`, `pnpm validate`);
   - shared-edge and environment-isolation invariants (`infra/`, VOC-079 foundation tests);
   - issue → plan → task implementation lifecycle (`AGENTS.md`, `specs/changes/`).
   It must remain a **router**, not a duplicate of `AGENTS.md` or canonical docs.
6. **Shared engineering skills** — add only these seven, each security-reviewed and
   repository-adapted:
   - context mapping;
   - systematic debugging;
   - verification before completion;
   - GitHub Actions efficiency;
   - React/Next.js performance guidance;
   - Playwright/browser testing;
   - security threat modeling.
   Each skill carries its own `.agents/skills/<name>/PROVENANCE.yaml`, validated from
   the T00-owned schema without later tasks editing a shared registry. Adapted material
   records separate upstream and local hashes, adaptation notes, and retained license/
   NOTICE paths. Repository-native material records the authoritative documentation
   sources used to author it.
7. **Graphify pilot** — evaluate as complementary code-relationship index, not a
   replacement for the navigator or direct source inspection:
   - pin an exact reviewed Graphify commit, package version, and complete transitive
     environment lock with hashes;
   - code-only/local mode; no document/image semantic extraction; no LLM auto-detection;
   - force query logging off;
   - repository-owned ignore rules excluding secrets, generated/vendor content, runtime
     data, and large/noisy artifacts;
   - do not commit generated graph output unless deterministic evidence proves benefit;
   - treat graph output as navigation hint; verify consequential claims in current source;
   - remain explicit/opt-in until measurement demonstrates reliable automatic use;
   - keep setup an explicit operator action into a repository-local isolated environment;
     ordinary skill/runner use never downloads, upgrades, falls back to a global install,
     mutates a user profile, installs hooks, or injects always-on instructions;
   - provide a repository-owned runner/check with clear prerequisite and safe failure
     when the locked local runtime is missing or mismatched.
8. **Verification and documentation** — navigation benchmark harness; discovery proof
   from repository root and a nested directory; operator-facing doc at
   `docs/development/agent-skills.md`; short `AGENTS.md` pointer preserving governance
   precedence.

### Non-goals / explicitly excluded

- Application behavior, API contracts, database schema, auth policy, or deployment
  bundle changes.
- Production or staging runtime configuration changes.
- Weakening `AGENTS.md`, `CLAUDE.md`, governance docs, CI gates, or protected-area
  rules via skill prose.
- Committing unpinned third-party skill content without provenance.
- Replacing canonical repository documentation with generated or skill-local knowledge.
- Auto-enabling Graphify during normal agent sessions before T04 measurement evidence.
- Self-adoption / self-authorization of this package.

## Risk and protected areas

- **Draft package proposal:** **R3** (agent instructions + `AGENTS.md` pointer;
  repository-governance semantic protection).
- **Measured path floor at drafting:** **R3** because `AGENTS.md` is in scope.
  `scripts/foundation/` alone would be lower; the combined package floor is R3.
- Protected areas: agent instructions (`AGENTS.md`, `CLAUDE.md`), repository governance
  semantics; skills must not override canonical policy or expand authority.
- EHR: not triggered at drafting time.
- Active authority model: **A-004**. No founder `approved` comment is a merge/adopt/release
  gate.

## Decisions

`VOC-112-D00`: **One canonical source.** Every skill's authoritative procedure lives
only under `.agents/skills/<name>/`. Claude adapters and any future agent-specific
stubs may only load that file. The Claude adapter repeats only the canonical discovery
metadata required by Claude and uses this sole body instruction (with `<name>` replaced):
`Load and follow the sole canonical procedure at
${CLAUDE_PROJECT_DIR}/.agents/skills/<name>/SKILL.md completely.` The project-root
substitution makes the target independent of the starting/nested working directory.

`VOC-112-D01`: **No symlinks.** Adapter parity is enforced by deterministic tests, not
Git symlink indirection, to preserve Windows/device portability and avoid discovery
regressions noted in issue #933. Validation requires exact metadata parity and the exact
single-instruction body template; any additional adapter procedure fails closed.

`VOC-112-D02`: **Governance precedence.** When skill prose conflicts with `AGENTS.md`,
`CLAUDE.md`, approved change packages, tests, or source code, the repository sources
win. Skills must state this explicitly.

`VOC-112-D03`: **Safety denylist (minimum).** Validation must fail closed when any
skill or adapter contains prohibited patterns including, at minimum:
- instructions to print, export, or grep environment secrets, `.env*`, credentials,
  cookies, OAuth material, session tokens, or personal data;
- instructions to paste raw CI logs containing secrets;
- unpinned package/tool installs (`@latest`, `npm install -g`, `curl | bash`, etc.);
- hidden network fetches not already governed by repository tooling;
- user-profile or global agent config mutation outside the repository.

`VOC-112-D04`: **Progressive disclosure.** Keep YAML `description` and skill opening
sections small; move depth into referenced files inside the same skill directory.
Enforce documented byte/line budgets in validation.

`VOC-112-D05`: **`vocanova-repo-navigator` routing table (minimum).** The skill must
include a compact table mapping intent → authoritative paths/commands, including at
minimum:

| Intent | Start here |
|--------|------------|
| Web UI / Next.js | `apps/web/`, `docs/design/08-web-app-design.md`, `docs/development.md` |
| API / Go backend | `apps/api/`, `docs/engineering/06-backend-design.md`, `docs/engineering/07-api-contract-and-dto-design.md` |
| Database / migrations | `apps/api/migrations/`, `docs/engineering/05-database-design.md` |
| Auth / OAuth | `apps/web/src/app/auth/`, `docs/operations/staging-controlled-signup.md`, `docs/operations/production-controlled-signup.md`, relevant packages under `specs/changes/` |
| Content seed | `apps/api/cmd/seed/` |
| Deploy / infra / shared edge | `infra/`, `.github/workflows/deploy-*.yml`, `docs/operations/11-devops-and-ci-cd.md`, `scripts/foundation/voc079-single-edge-invariants.test.mjs` |
| Monitoring | `infra/monitoring/`, `docs/operations/monitoring.md` |
| Governance / change workflow | `AGENTS.md`, `docs/governance/`, `specs/changes/`, `specs/templates/change-package/` |
| Validation / tests | `docs/development.md`, `pnpm validate`, `scripts/foundation/*.test.mjs` |
| Issue → plan → task lifecycle | `AGENTS.md` (Reporting a bug / change workflow), `specs/changes/` |

Implementer may refine wording and add sub-routes but must not remove domains or
duplicate full governance text inline.

`VOC-112-D06`: **Shared skill set is fixed to seven** named in issue #933. Additional
skills require a separate governed package.

`VOC-112-D07`: **Task-isolated provenance.** T00 owns an immutable schema at
`.agents/skills/provenance.schema.yaml`; every canonical skill owns
`.agents/skills/<name>/PROVENANCE.yaml`. The validator discovers these records
dynamically, so parallel T01/T02/T03 branches add files without editing one registry or
the T00 validator. Adapted skills record `upstream_repo`, `upstream_commit`,
`upstream_path`, `upstream_sha256`, `local_sha256`, `license`, retained license/NOTICE
paths, and adaptation notes. Native skills declare `source: repository-native` and their
authoritative documentation sources. All committed instructions, scripts, references,
assets, licenses, and notices are included in the review and local-hash manifest.

`VOC-112-D08`: **Graphify opt-in.** Provide skill `graphify-pilot` (or equivalent name)
disabled by default for automatic invocation (`disable-model-invocation: true` or
documented explicit opt-in only). Runner lives under repository control
(`scripts/graphify/`). Default pilot uses `--code-only`, logging disabled, and
`.graphifyignore` aligned to repository secret/generated paths. The reviewed Graphify
commit/package and every transitive dependency are locked with hashes in an isolated
repository-local environment. Setup is a separate explicit operator command; the normal
runner is offline/fail-closed and must never use upstream `install`, provider
auto-detection, agent hooks, always-on AGENTS/CLAUDE injection, a global executable, or a
user profile.

`VOC-112-D09`: **Measured agent benchmark.** T04 runs controlled baseline and
navigator-assisted sessions against fixed questions, the same repository revision,
model/runtime version, permissions, and keyed correctness rubric. Metrics are derived from
the structured run/tool trace rather than hand-entered or hardcoded values: files opened,
search operations, elapsed time, correctness, and context/usage metadata exposed by the
runtime. A deterministic test validates the captured metadata, exact revisions, rubric,
and threshold calculations; it does not simulate the measurements.

`VOC-112-D10`: **Real discovery evidence.** T04 uses the current hosted Cursor runtime to
list or invoke a repository skill from the repository root and from at least one nested
working directory (recommended: `apps/web/`); static filesystem assertions are not a
substitute. It also exercises Claude Code from both locations when an already-authorized
non-interactive runtime is available. If Claude requires new interactive credentials, the
evidence must say `not-executed-external-credential-required` and must not claim a pass.
Evidence is metadata-only and contains no prompts, raw logs, secrets, or personal data.

## Data, migrations, analytics, and accessibility

None for product data. Optional local Graphify index artifacts remain gitignored unless
T03/T04 produce explicit commit justification evidence.

## Security, privacy, and authorization

Skills are agent-instruction surfaces. They must not grant merge, deploy, secret, or
production-data authority. Validation and skill prose must forbid credential exfiltration
and personal-data harvesting. Evidence files contain bounded metadata only.

## Open questions

1. **T02 upstream pins:** T02 independent verification must confirm each shared skill's
   chosen upstream repository/commit is license-compatible and security-acceptable before
   the T02 task PR merges. Package adoption does not resolve this implementation-time
   dependency, and the planner intentionally does not fix upstream SHAs in this draft.
   Exact review has already established that `vercel-labs/agent-skills` at
   `dd089a8c752c966dee8bf0f27cb625ba193ffd9e` has no detected or committed license;
   its React skill must not be copied or adapted. Author React/Next guidance independently
   from current official React/Next documentation or select an explicitly licensed source,
   and record the rejection decision.
2. **T03 Graphify pin:** Implementer selects the exact Graphify-Labs/graphify tag/commit
   and matching `graphifyy` PyPI version during T03, then produces and reviews the complete
   transitive lock/hash set before enabling the pilot skill.
3. **Cursor discovery path:** If a future Cursor release requires an additional project
   marker beyond `.agents/skills/`, T00/T04 may add the smallest compatible marker in the
   same task PR after verifying discovery — not a separate scope expansion.
