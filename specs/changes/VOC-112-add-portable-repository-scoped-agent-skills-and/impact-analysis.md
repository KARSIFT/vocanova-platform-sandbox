# VOC-112 — Impact Analysis

## Security and privacy

Agent skills are instruction surfaces adjacent to repository governance. Incorrect or
hostile skill content could steer agents toward credential exposure, personal-data search,
or authority overreach. Mitigations:

- deterministic forbidden-pattern validation (`VOC-112-D03`);
- task-isolated provenance with distinct upstream/local hashes, retained licenses/notices,
  and complete committed-file review (`VOC-112-D07`);
- explicit governance precedence (`VOC-112-D02`);
- Graphify ignore rules excluding secrets and generated/vendor trees;
- evidence files bounded to metadata — no logs, tokens, cookies, OAuth material, or
  personal data.

`AGENTS.md` changes are limited to a pointer in T04 and require governance validation.

## Protected and operational surfaces

- **Primary write surfaces:** `.agents/skills/**`, `.claude/skills/**`,
  `scripts/foundation/voc112-*.test.mjs`, `scripts/graphify/**`, `.graphifyignore`,
  `.gitignore`, `docs/development/agent-skills.md`, short `AGENTS.md` subsection.
- **Preserve unchanged:** application runtime, deploy workflows semantics, database schemas,
  auth/OAuth behavior, monitoring inventory, secret values, branch protection, and merge/release
  automation except incidental doc/skills presence in the monorepo.
- **Agent authority:** skills must not grant merge, deploy, secret, or production-data access;
  they route agents to canonical sources only.

## Data and migrations

None. Optional local Graphify indexes remain gitignored by default.

## Analytics and accessibility

No product analytics or UI accessibility changes.

## Risks, dependencies, and evidence

- `VOC-112-R00`: **Skill/governance conflict** — skill prose could contradict `AGENTS.md` or
  protected-area rules. Mitigation: precedence statements, compact router design, independent
  semantic review, governance validation on `AGENTS.md` edits.
- `VOC-112-R01`: **Secret or PII exfiltration via instructions** — adapted upstream skills may
  contain unsafe patterns. Mitigation: security review in T02, denylist validation, repository-specific
  adaptations forbidding env/credential greps and raw CI log export.
- `VOC-112-R02`: **Adapter/canonical drift** — duplicate procedures or missing adapters.
  Mitigation: T00 validation; fail-closed CI; no symlinks.
- `VOC-112-R02A`: **Parallel-branch conflicts** — T01/T02/T03 were originally drafted to
  update one registry/test file. Mitigation: T00 owns immutable schema/validator files;
  later parallel tasks add per-skill provenance and fixtures without shared writes.
- `VOC-112-R03`: **Context bloat** — oversized skills degrade agent performance. Mitigation:
  progressive disclosure budgets (`VOC-112-D04`) and T04 measurement.
- `VOC-112-R04`: **Unpinned upstream content** — stale or tampered third-party skills.
  Mitigation: provenance hashes/commits; no `@latest` installs.
- `VOC-112-R05`: **Graphify false authority** — agents treat graph hints as truth.
  Mitigation: opt-in pilot throughout VOC-112, hint-only language, source verification
  requirement, and no automatic invocation in this package. Any later enablement requires
  a separate governed change with Graphify-specific measurement and acceptance gates.
- `VOC-112-R06`: **Graphify supply chain / local runtime** — the published package uses
  broad transitive ranges and upstream install can mutate agent configuration. Mitigation:
  exact upstream/package pin plus reviewed transitive lock/hashes, repository-local isolated
  environment, explicit setup only, offline/fail-closed ordinary runner, code-only mode, and
  no global fallback, provider auto-detection, hooks, or always-on instruction injection.
- `VOC-112-R07`: **Unsupported redistribution** — public skill text may lack a compatible
  license. Mitigation: reject unlicensed sources (including the reviewed Vercel React skill),
  retain required license/NOTICE files, and author repository-native guidance where needed.
- `VOC-112-DEP-00`: Issue #933 requirement thread (resolved at drafting).
- `VOC-112-DEP-01`: Protected agent instructions (`AGENTS.md`, `CLAUDE.md`) (resolved).
- `VOC-112-DEP-02`: Existing development workflow docs (resolved).
- `VOC-112-DEP-03`: T02 upstream pin selection (open until implementation).
- `VOC-112-DEP-04`: T03 Graphify exact pin (open until implementation).
- `VOC-112-EV-00`: `t00-evidence.md` — framework validation.
- `VOC-112-EV-01`: `t01-evidence.md` — navigator review.
- `VOC-112-EV-02`: `t02-evidence.md` — shared skill provenance/security review.
- `VOC-112-EV-03`: `t03-evidence.md` — Graphify pilot pins/limitations.
- `VOC-112-EV-04`: `t04-evidence.md` — benchmark, discovery, documentation completion.
