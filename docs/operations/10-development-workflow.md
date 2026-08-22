---
id: DOC-10
title: VocaNova Development Workflow
version: 1.0
document_type: engineering-workflow
status: approved
owner: founder
canonical_path: docs/operations/10-development-workflow.md
approved_at: 2026-07-21
last_reviewed_at: 2026-07-21
review_cycle: quarterly
supersedes: null
related_documents:
  - DOC-11
  - DOC-15
  - DOC-16
  - DOC-19
related_decisions: []
adoption_change: VOC-008
source_files:
  - path: 10-development-workflow.md
    sha256: 7fdd38cb7f877051907cc68e0930ece507fe3466dab3e008795c2827eeb21aaf
---
# 10 — VocaNova Development Workflow

## 1. Principles

Approved product documents are the source of truth for product behavior; feature coding begins only
after requirements and an implementation plan are approved; ChatGPT defines requirements/architecture
handoffs; Codex is the primary implementation agent; Claude Code is the independent reviewer; GitHub
is the operational source of truth; GitHub Actions is the CI/CD orchestrator; deterministic tests
enforce objective rules while Claude evaluates contextual concerns; AI agents never hold unrestricted
production secrets or merge their own work; database migrations are validated through automated
tests; product and release authority follows the live R0–R4 and RL1–RL3 governance model. See
[DOC-19](19-governance-reconciliation-notes.md) for orientation and the linked canonical sources.

## 2. Repository

One monorepo, `vocanova-platform` (private during MVP — see
[the migration notes](../archive/README-migration-notes.md#5-repository-name-conflict)). Recommended
structure:

```text
vocanova-platform/
├── apps/web/                      # Next.js
├── apps/api/                      # Go backend module
│   ├── ent/schema/
│   ├── migrations/
│   └── openapi/vocanova.openapi.json
├── packages/{api-client,design-tokens,eslint-config,typescript-config}/
├── docs/{product,research,design,engineering,operations,architecture,planning,governance}/
├── scripts/
├── .github/{workflows,ISSUE_TEMPLATE,ai}/
├── AGENTS.md
├── CLAUDE.md
├── REVIEW.md
```

The canonical document corpus is split by category and indexed in [docs/README.md](../README.md).
The Go backend remains a normal module at `apps/api`, not part of the pnpm workspace.

## 3. Branch strategy

The intended topology has two permanent branches: `develop` (default integration branch and future
staging source) and `main` (production source, accepting governed release changes plus emergency
hotfixes). No `release/*` branch is planned for MVP; a `develop → main` PR represents a release
candidate, subject to the live authority and release-class rules. Automatic staging deployment,
automatic merge, and production deployment are not technically active as of 2026-07-21; consult
[the A-003 transition state](../governance/a003-transition-state.yaml) rather than inferring
activation from this topology.

```text
feature/* ──PR──► develop ──release PR──► main
                    │                       │
              Future staging         Production source
```

Task branches: `<type>/<issue-number>-<short-description>` (`feature/`, `fix/`, `hotfix/`,
`refactor/`, `chore/`, `docs/`, `test/`, `security/`, `revert/`). Squash merge only into `develop`;
merge commit for `develop → main`. Rebase before final review; `git push --force-with-lease`, never
plain `--force`. These mechanics do not grant merge or release authority; that authority comes from
canonical governance.

## 4. Work hierarchy

```text
Milestone → Epic → GitHub Issue → Pull Request
```

Sources of truth: product vision → `docs/product/`; architecture → `docs/architecture/` + ADRs;
implementation work → GitHub Issues; status → GitHub Projects; release scope → Milestones;
production history → Releases/deployments.

Priorities P0 (critical: outage, data loss, active security incident) through P3 (post-MVP
polish); most planned work is P2. Sizes XS/S/M/L — Codex should never receive an `L` issue directly;
split it first.

## 5. Definition of Ready / Definition of Done

**Ready**: clear objective, stated value, linked requirement sources, testable acceptance criteria,
defined scope/exclusions, identified technical/security/privacy impact, defined testing
expectations, resolved dependencies, sized XS/S/M, no blocking product or architecture decision
remaining.

**Done**: acceptance criteria satisfied, scope respected, required tests pass (unit/integration/
contract/migration/e2e as applicable), security/authorization correct, migrations tested,
OpenAPI/generated types synchronized, documentation updated, no secrets exposed, required review
resolved, merged through a PR, staging deployment succeeds and is verified (production verification
additionally required for release work).

## 6. Pull-request standards

One coherent outcome per PR; unrelated work becomes a separate issue. Preferred size 100–500
meaningful changed lines (under 200 for fixes; over 800 normally split). Conventional Commits
(`feat`, `fix`, `refactor`, `test`, `docs`, `chore`, `build`, `ci`, `perf`, `security`, `revert`); no
AI agent names in commit messages. PR body must cover: summary, linked issue, requirement sources,
scope, implementation, acceptance criteria, testing performed, security/database/API/environment
impact, documentation impact, known risks, and review status.

## 7. Testing strategy

Layers: end-to-end → integration/contract → component/service → unit, with most tests below the
end-to-end layer. Go: unit, service, repository, handler/API, PostgreSQL integration (never SQLite
for anything DB-behavior-relevant), auth, transaction, AI-response-parsing tests. Frontend: utility,
component, form, feature-level, error/empty/loading/success states, accessibility, responsive,
API-integration tests. End-to-end (Playwright) covers the full core loop: auth → discover → save →
review session → sentence submission → deterministic AI feedback → progress update → settings change
→ logout → unauthenticated-access rejection. Ordinary CI always uses a deterministic AI adapter, not
a paid/nondeterministic provider call.

Coverage direction (guides quality, not a target to game): backend ~70%, frontend unit/component
~60%, critical domain/security logic (auth, sessions, spaced repetition, daily progress, ownership
boundaries, AI structured-output parsing, migrations) 90%+.

CI levels: **Level 1** (fast PR checks — format, lint, typecheck, unit tests, generated/OpenAPI
drift, build, basic security); **Level 2** (full PR checks — PostgreSQL integration, migration
tests, contract tests, component tests, selected Playwright, Claude review, and the pipeline
`web container runtime` job for pull requests that touch any repository-root file (including
`.dockerignore` and workspace manifests), `apps/web/**`, any `packages/**` path, or the gate's
own workflow/test definitions: it builds `apps/web/Dockerfile`, boots the local smoke image
without registry push or repository secrets, requires HTTP 2xx from `/`, verifies the container
remains running, and always cleans up); **Level 3**
(post-merge staging checks); **Level 4** (production release checks).

## 8. Database migrations

`Ent schema → Atlas migration → PostgreSQL`. Standard flow: update Ent schema → generate Ent code →
generate Atlas migration → categorize risk (low/medium/high) → migration tests → integration tests →
commit together → Claude migration-risk review. High-risk migrations (drop table/column, populated type change, large-table rewrite,
primary-key change, user-data deletion, or irreversible transformation) follow the live R0–R4
classification, protected-area controls, approval matrix, and EHR rules. Required evidence includes
migration lint, from-zero and upgrade-path tests, destructive-operation detection, recovery proof,
and independent migration-risk review as applicable. R4 consequences require founder approval;
routine R3 does not require standing founder or steward approval merely for being R3. Use
expand-and-contract so `develop` does not carry an unrecoverable migration between merge and release.
See the [canonical governance index](../governance/README.md).

## 9. Security workflow

Mandatory Claude security review for changes touching authentication, sessions, cookies, OAuth,
magic links, identity, user-owned data, AI-provider integration, logging, secrets, environment
variables, GitHub Actions, Cloudflare config, dependencies, or migrations. Severity: Critical (blocks
merge+deploy), High (normally blocks merge), Medium (fixed or explicitly tracked), Low (fixed or
scheduled).
