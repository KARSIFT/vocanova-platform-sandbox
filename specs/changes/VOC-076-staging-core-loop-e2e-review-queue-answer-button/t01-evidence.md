# VOC-076-T01 — Disabled-button fix and regression coverage

**Task:** VOC-076-T01  
**Package:** VOC-076  
**Evidence ID:** VOC-076-EV-01  
**Root cause (from T00):** sustained `isLoading === true` (`isSubmitting || isRefetching`)
kept MC options disabled while the fieldset stayed visible; secondary stale
`phase === "feedback"` across card changes via `useEffect` reset lag  
**Implementation date:** 2026-08-13
**Remediation:** addresses independent-review High finding on commit
`e790e6bc887212d810402ce26676d42969fe1239` (duplicate-submit window during
batch-end refetch)

## Fix applied

### Product (`review-session.tsx` + `review-session-prompt.ts`)

1. **Narrow MC disabled rule (primary).** Multiple-choice meaning buttons use
   `isSubmitting || phase === "feedback"` via `isMultipleChoiceOptionDisabled`.
   Background batch-end `listDueWords` refetch (`isRefetching`) does **not**
   disable prompt-ready MC options — the overlap case T00 identified when a
   new card's `dueWords` lands before `isRefetching` clears.

2. **Post-submit actions stay locked during refetch (remediation).** "Show
   answer", rating, and "Continue" use
   `isReviewActionDisabled(isSubmitting, isRefetching)`. After a successful
   `submitReview`, `advance()` may set `isRefetching` while the same card is
   still on screen; `finally` clears `isSubmitting` first. Keeping rate /
   continue / show-answer disabled for the whole refetch closes the
   duplicate-`submitAttempt` window the prior T01 revision opened by treating
   those controls as submit-only.

3. **Synchronous phase reset.** Card-change reset uses `useLayoutEffect` so a
   new card never paints one frame with stale `phase === "feedback"` (T00 §5
   secondary defect).

4. **Stable MC option order.** `buildMultipleChoiceOptions` result is memoized
   per card (`useMemo` on `currentIndex` / `dueWords`) so `shuffleArray` does
   not reorder options every render (reduces Playwright detach retries; T00 §5
   note).

5. **Accessibility preserved.** `aria-pressed`, fieldset/legend, focus styles,
   and `max-w-[28rem]` workaround unchanged. Card shell exposes
   `aria-busy={isRefetching || isSubmitting}` during async work.

Pure readiness helpers live in `review-session-prompt.ts` for deterministic
regression tests.

### E2E (`core-loop.staging.spec.ts`)

`reviewOneCard` calls `await expect(firstMcOption).toBeEnabled()` before
clicking the first MC meaning button (defense-in-depth per T00 §3 and
`VOC-076-AC-02`). VOC-074 `reviewedCards >= 1` assertion and VOC-050 journey
structure are unchanged.

## Regression coverage

| Test | Location | What it guards |
|------|----------|----------------|
| `review-session prompt readiness (VOC-076-T01)` | `apps/web/tests/lib/review-session-prompt-readiness.test.ts` | MC enabled in prompt / disabled in feedback+submit; MC independent of refetch; rate/continue/show-answer locked for submit **and** refetch |
| Staging core-loop step 5 | `apps/web/tests/staging-e2e/core-loop.staging.spec.ts` | Explicit enabled wait before MC click; `reviewedCards >= 1` (VOC-074) |

## How to run

From the repository root:

```bash
pnpm --filter @vocanova/web test:middleware
pnpm --filter @vocanova/web typecheck
pnpm --filter @vocanova/web typecheck:e2e
pnpm --filter @vocanova/web lint
```

Governance:

```bash
bash scripts/governance/validate-governance.sh
bash scripts/governance/classify-change-risk.sh
git diff --check
```

Real staging proof is **T02** (`t02-evidence.md` after merge to `develop`).

## Intentional disabled states preserved (VOC-076-TEST-02)

- MC options remain disabled during `phase === "feedback"` after selection.
- MC options remain disabled while `isSubmitting` is true.
- Rate / continue / show-answer remain disabled while `isSubmitting` **or**
  `isRefetching` is true (prevents double-submit on the still-visible card
  during batch-end refetch).
- `isRefetching` alone does **not** disable MC meaning options on a
  prompt-ready card.

## Residual risk (Medium from independent review; out of T01 closure)

A hung `listDueWords` that never settles can leave the prior card in
`phase === "feedback"` with MC options disabled for the full E2E timeout.
That matches a stalled API path T00 marked log-inconclusive and left to T02 /
a separate package if product+E2E changes do not close the gate. This
remediation does not add an API client timeout.

## Out of scope (per package)

- No `deploy-staging.yml` edits (`VOC-076-DEP-01`).
- No API client timeout / `apps/api/` changes (T00 log-inconclusive; separate
  package if T02 still fails).
- VOC-074 increment path untouched.

## Commands run during T01

```bash
pnpm --filter @vocanova/api-client build
pnpm --filter @vocanova/web test:middleware
# includes VOC-076 prompt-readiness cases (MC + post-submit refetch lock)

pnpm --filter @vocanova/web typecheck
pnpm --filter @vocanova/web typecheck:e2e
pnpm --filter @vocanova/web lint
bash scripts/governance/validate-governance.sh
bash scripts/governance/classify-change-risk.sh
# Detected path floor: R1
git diff --check
```
