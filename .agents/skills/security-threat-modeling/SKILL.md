---
name: security-threat-modeling
description: Use when the user explicitly asks to threat-model a codebase path, enumerate abuse paths, or produce an AppSec threat model—not for general architecture summaries or routine code review.
---

# Security threat modeling

Deliver a **repository-grounded** threat model for vocanova-platform-sandbox or an in-scope path. Anchor claims to evidence in source, `infra/`, and operations docs.

## Governance precedence

When this skill conflicts with `AGENTS.md`, `CLAUDE.md`, approved change packages, tests, or source code, the repository sources win.

## When to use

- Explicit request to threat model, enumerate threats, or map abuse paths
- Pre-release security review of auth, deploy, or data flows

Do **not** use for generic code review, lint fixes, or non-security design summaries.

## Workflow

### 1. Scope and system model

- Identify components: `apps/web`, `apps/api`, `infra/`, `.github/workflows/`, `packages/`.
- Separate **runtime** behavior from CI, tests, and local dev tooling.
- Read `docs/operations/`, `docs/engineering/`, `docs/development.md`, and auth docs before inferring exposure.
- Use `.agents/skills/vocanova-repo-navigator/SKILL.md` to locate authoritative paths.

### 2. Boundaries, assets, entry points

- Trust boundaries (browser ↔ edge ↔ API ↔ database ↔ third parties).
- Assets: user accounts, session tokens, OAuth state, audit logs, deploy credentials (names only).
- Entry points: HTTP routes, auth callbacks, webhooks, admin tooling, CI triggers.

### 3. Threats as abuse paths

- Map attacker goals to assets: exfiltration, privilege escalation, integrity loss, DoS.
- Keep the list small and evidence-backed.

### 4. Prioritize

- Qualitative likelihood and impact with short justification.
- Note existing controls with file references (`apps/web/src/app/auth/`, middleware, API handlers).

### 5. Mitigations

- Distinguish **existing** vs **recommended** controls.
- Tie recommendations to concrete locations and control types (authZ, validation, rate limits, secrets isolation).

### 6. Output

Follow [references/prompt-template.md](./references/prompt-template.md). Write Markdown to `<scope>-threat-model.md` in the working task directory unless the user specifies another path. Keep assumptions explicit.

## Risk guidance (illustrative)

- **High:** auth bypass, cross-tenant access, secret theft, pre-auth RCE
- **Medium:** partial data exposure, targeted DoS, rate-limit bypass with impact
- **Low:** low-sensitivity leaks with easy mitigation

## Safety

Never read `.env*` or credential stores. Do not paste production data, session tokens, or raw CI logs into the model. Threat models describe controls — they do not grant deploy, `pnpm validate`, or merge authority.

## References

- [references/prompt-template.md](./references/prompt-template.md) — output contract
