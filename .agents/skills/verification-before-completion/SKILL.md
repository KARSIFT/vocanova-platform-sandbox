---
name: verification-before-completion
description: Use when about to claim work is complete, fixed, or passing, before committing or creating PRs - requires running verification commands and confirming output before making any success claims; evidence before assertions always
---

# Verification before completion

**Core principle:** Evidence before claims, always.

## Governance precedence

When this skill conflicts with `AGENTS.md`, `CLAUDE.md`, approved change packages, tests, or source code, the repository sources win.

## Iron law

```
NO COMPLETION CLAIMS WITHOUT FRESH VERIFICATION EVIDENCE
```

If you have not run the verification command in this message, you cannot claim it passes.

## Gate function

```
BEFORE claiming any status:

1. IDENTIFY — What command proves this claim?
2. RUN — Execute the FULL command (fresh, complete)
3. READ — Exit code, failure count, relevant output lines
4. VERIFY — Does output support the claim?
5. ONLY THEN — State the claim with evidence
```

Skipping a step is not verification.

## Vocanova command map

| Claim | Required command (from `docs/development.md`) |
|-------|--------------------------------------------------|
| Workspace healthy | `pnpm validate` |
| Tests pass | `pnpm test` and/or targeted package tests |
| Web E2E pass | `pnpm --filter @vocanova/web test:e2e` |
| API tests pass | `cd apps/api && go test ./...` |
| Agent skills valid | `node --test scripts/foundation/voc112-agent-skills.test.mjs` |
| Governance docs OK | `bash scripts/governance/validate-governance.sh` (when applicable) |

Use the narrowest command that still proves the claim; use `pnpm validate` when the change spans multiple tiers.

## Common failures

| Claim | Requires | Not sufficient |
|-------|----------|----------------|
| Tests pass | Latest test output: 0 failures | Earlier run, assumptions |
| Lint clean | Latest lint output: 0 errors | Partial file check |
| Build succeeds | Build exit 0 | Lint alone |
| Bug fixed | Reproduction test passes | Code changed only |
| Package complete | Acceptance criteria checklist | Tests alone |

## Red flags — stop

- "Should", "probably", "seems to"
- Satisfaction before verification ("Done!", "Fixed!")
- Trusting agent reports without diff and command output
- Partial checks when the claim is global

## Examples

```
✅ [pnpm test → 0 failures] "Foundation tests pass"
❌ "Should pass now"

✅ [pnpm validate → exit 0] "Workspace validation passes"
❌ "Lint looked fine"
```

## Safety

Do not read `.env*` files. Do not paste raw CI logs. Do not expose secrets or personal data as verification evidence — summarize exit codes, failure counts, and test names only.
