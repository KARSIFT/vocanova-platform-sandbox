---
name: systematic-debugging
description: Use when encountering any bug, test failure, or unexpected behavior, before proposing fixes
---

# Systematic debugging

**Core principle:** Find root cause before attempting fixes. Symptom patches are failure.

## Governance precedence

When this skill conflicts with `AGENTS.md`, `CLAUDE.md`, approved change packages, tests, or source code, the repository sources win.

## Iron law

```
NO FIXES WITHOUT ROOT CAUSE INVESTIGATION FIRST
```

Complete Phase 1 before proposing fixes.

## When to use

- Failing tests (`pnpm test`, `go test ./...`, Playwright under `apps/web`)
- Build, lint, or type errors from `pnpm validate`
- Unexpected runtime or CI behavior
- Performance regressions

Use especially under time pressure, after failed fix attempts, or when the issue looks "simple."

## Phase 1 — Root cause investigation

1. **Read errors completely** — stack traces, file paths, exit codes, foundation-test names.
2. **Reproduce reliably** — exact commands from `docs/development.md`; note environment (local vs CI).
3. **Check recent changes** — `git diff`, recent commits, dependency or workflow edits.
4. **Gather boundary evidence** in multi-component flows (web → API → database → CI):
   - Log inputs/outputs at each boundary with non-secret diagnostics only.
   - Identify which layer first diverges from expected behavior.
5. **Trace data flow backward** from the bad value to its origin; fix at the source.

## Phase 2 — Pattern analysis

1. Find a **working example** in the same codebase (`apps/web`, `apps/api`, `scripts/foundation/`).
2. Read the reference implementation fully — do not skim.
3. List every difference between working and broken paths.
4. Note required config, environment variables (names only — never values), and assumptions.

## Phase 3 — Hypothesis and testing

1. State one hypothesis: "X is the root cause because Y."
2. Make the **smallest** change that tests it — one variable at a time.
3. If the hypothesis fails, form a new one; do not stack speculative fixes.
4. Say "I don't understand X" when true; research or narrow scope instead of guessing.

## Phase 4 — Implementation

1. **Add a failing test** when feasible — foundation test, Go test, or Playwright spec under `apps/web/tests/`.
2. **Implement one fix** for the confirmed root cause — no bundled refactors.
3. **Verify** with the full relevant command (`pnpm validate`, targeted tests, or API `go test`).
4. Use the `verification-before-completion` skill before claiming success.
5. After **three failed fix attempts**, stop and document an architectural concern in the PR or issue instead of another symptom patch.

## Repository validation

| Area | Start here |
|------|------------|
| Workspace | `pnpm validate`, `docs/development.md` |
| Web / E2E | `pnpm --filter @vocanova/web test:e2e` |
| API | `cd apps/api && go test ./...` |
| Foundation | `node --test scripts/foundation/*.test.mjs` |

## Red flags — return to Phase 1

- "Quick fix for now"
- Changing multiple things before re-running tests
- Skipping reproduction
- Proposing fixes before boundary evidence
- A fourth fix attempt after three failures

## Safety

Never read or export `.env*` files, secrets, credentials, session tokens, or personal data. Do not paste raw CI logs. Use structured `gh run view --json` summaries when CI inspection is needed — not full log dumps.
