---
name: github-actions-efficiency
description: Audit GitHub Actions workflow efficiency and recommend fixes to reduce CI minutes and wasted runs in this repository.
---

# GitHub Actions efficiency

Lean entrypoint for CI cost and latency work in `.github/workflows/`. Align proposed checks with validation tiers in `docs/development.md`. Load [references/actions.md](./references/actions.md) when you need audit detail beyond this page.

## Governance precedence

When this skill conflicts with `AGENTS.md`, `CLAUDE.md`, approved change packages, tests, or source code, the repository sources win.

## When to use

- Reducing GitHub Actions runtime or duplicate runs
- Adding caching, concurrency, path filters, or matrix tuning
- Auditing workflows under `.github/workflows/` (including `pipeline.yml`, deploy, accessibility, lighthouse)

## Core workflow

### 1. Measure first (static + metadata)

```bash
rg -n "on:|concurrency:|paths:|paths-ignore:|strategy:|matrix:|cache:" .github/workflows
gh run list --limit 10 --json databaseId,conclusion,workflowName,createdAt
gh run view "$run_id" --json conclusion,jobs,status
```

Look for: missing pnpm/Go caches, missing `concurrency` cancellation, over-broad triggers, duplicate coverage, and jobs that ignore path scope.

Do not paste or export raw CI logs. Use JSON summaries and job conclusions only.

### 2. Apply guardrails

1. Do not hide required validation — release, governance, migration, foundation checks, or `pnpm validate` tiers stay.
2. Do not cut parallelism without justification.
3. Keep matrix legs that match documented version commitments.
4. Formatter or bot write-back jobs should be opt-in triggers.
5. Separate repo YAML changes from org-level settings.

### 3. Select up to three fixes

Rank by estimated minutes saved. Common candidates:

1. Lockfile-keyed dependency caches (`pnpm`, Go modules)
2. `concurrency` cancel-in-progress for PR branches
3. Remove duplicate workflow coverage
4. Narrow `paths` / job-level `if` gates
5. Reduce matrix breadth for non-release events
6. Parallelize independent jobs on the critical path

### 4. Verify

- Prefer `workflow_dispatch` or a test push on a non-protected branch when `gh` is available.
- State explicitly when only static YAML review was possible.

## Required output

1. **Waste sources** — top drivers from step 1
2. **Proposed fixes** — up to three with audit evidence
3. **Validation** — live vs static-only
4. **Impact** — expected vs measured runner time

## Safety

Do not read `.env*` files or exfiltrate secrets from workflows or logs. Do not weaken governance workflows (`repository-governance.yml`, merge gates) without an approved change package.

## References

- [references/actions.md](./references/actions.md) — audit order and safe-change rules
