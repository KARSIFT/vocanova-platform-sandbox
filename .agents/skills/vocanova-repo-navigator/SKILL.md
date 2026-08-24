---
name: vocanova-repo-navigator
description: Route repository work to the smallest authoritative context—web, API, database, auth, seed, deploy, monitoring, governance, validation, and change workflow—without duplicating canonical docs.
---

# Vocanova repository navigator

Use this skill when you need to find **where to start** in the checkout. It is a
**router only**: open the paths below, then follow the linked canonical documentation
and source. Do not treat this file as a substitute for `AGENTS.md`, `CLAUDE.md`,
approved change packages, tests, or application code.

## Governance precedence

When this skill conflicts with `AGENTS.md`, `CLAUDE.md`, approved change packages,
tests, or source code, the repository sources win.

## How to use

1. Match the task intent to one row in the routing table.
2. Open the listed paths (smallest set that answers the question).
3. Use repository validation commands from `docs/development.md`—do not invent
   substitutes.
4. For governed implementation work, confirm an adopted change package before
   editing product behavior.

## Routing table

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

## Safety

Never expose secrets, credentials, session tokens, OAuth material, cookies, or
personal data. Do not grep or read `.env*` files. Do not paste raw CI logs. Use
only repository-documented, pinned validation commands.

## Shared-edge reminder

Environment isolation and single-edge invariants are enforced in `infra/` and
`scripts/foundation/voc079-single-edge-invariants.test.mjs`. Read those before
changing deploy bundles, hostnames, or cross-environment routing.
