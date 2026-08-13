# VOC-076-T01 — Disabled-button fix and regression coverage

**Task:** VOC-076-T01  
**Package:** VOC-076  
**Evidence ID:** VOC-076-EV-01  
**Root cause (from T00):** sustained `isLoading === true` (`isSubmitting || isRefetching`)
kept MC options disabled while the fieldset stayed visible; secondary stale
`phase === "feedback"` across card changes via `useEffect` reset lag  
**Implementation date:** 2026-08-13

## Fix applied

### Product (`review-session.tsx` + `review-session-prompt.ts`)

1. **Narrow MC disabled rule (primary).** Multiple-choice meaning buttons now use
   `isSubmitting || phase === "feedback"` instead of `isLoading || phase === "feedback"`.
   Background batch-end `listDueWords` refetch (`isRefetching`) no longer disables
   prompt-ready MC options — the overlap case T00 identified when
   `reviewOneCard` returns while refetch is still in flight.

2. **Same submit-only lock for other learner actions.** "Show answer", rating,
   and "Continue" buttons use `isReviewActionDisabled(isSubmitting)` so a slow
   refetch cannot block rating submission on the current card or interaction on
   the next card after refetch data lands.

3. **Synchronous phase reset.** Card-change reset moved from `useEffect` to
   `useLayoutEffect` so a new card never paints one frame with stale
   `phase === "feedback"` (T00 §5 secondary defect).

4. **Stable MC option order.** `buildMultipleChoiceOptions` result is memoized
   per card (`useMemo` on `currentIndex` / `dueWords`) so `shuffleArray` does not
   reorder options every render (reduces Playwright detach retries; T00 §5 note).

5. **Accessibility preserved.** `aria-pressed`, fieldset/legend, focus styles,
   and `max-w-[28rem]` workaround unchanged. Card shell exposes
   `aria-busy={isRefetching || isSubmitting}` during async work without blocking
   clicks.

Pure readiness helpers live in `review-session-prompt.ts` for deterministic
regression tests.

### E2E (`core-loop.staging.spec.ts`)

`reviewOneCard` now calls `await expect(firstMcOption).toBeEnabled()` before
clicking the first MC meaning button (defense-in-depth per T00 §3 and
`VOC-076-AC-02`). VOC-074 `reviewedCards >= 1` assertion and VOC-050 journey
structure are unchanged.

## Regression coverage

| Test | Location | What it guards |
|------|----------|----------------|
| `review-session prompt readiness (VOC-076-T01)` | `apps/web/tests/lib/review-session-prompt-readiness.test.ts` | MC options enabled in prompt phase; disabled during feedback/submit; refetch excluded from action lock |
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
- MC options, ratings, show-answer, and continue remain disabled while
  `isSubmitting` is true (prevents double-submit).
- `isRefetching` alone no longer disables interactive controls.

## Out of scope (per package)

- No `deploy-staging.yml` edits (`VOC-076-DEP-01`).
- No API client timeout / `apps/api/` changes (T00 log-inconclusive; separate
  package if T02 still fails).
- VOC-074 increment path untouched.

## Commands run during T01

```bash
pnpm --filter @vocanova/api-client build
pnpm --filter @vocanova/web test:middleware
# 9 tests pass (includes 4 new VOC-076 prompt-readiness cases)

pnpm --filter @vocanova/web typecheck
pnpm --filter @vocanova/web typecheck:e2e
pnpm --filter @vocanova/web lint
bash scripts/governance/validate-governance.sh
bash scripts/governance/classify-change-risk.sh
# Detected path floor: R1
git diff --check
```
