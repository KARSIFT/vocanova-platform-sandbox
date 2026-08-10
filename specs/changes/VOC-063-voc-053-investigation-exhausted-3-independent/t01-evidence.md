# VOC-063-EV-01 — T01 step-7 retry hardening and diagnostic removal evidence

Evidence for `VOC-063-T01` (`VOC-063-AC-01`, `VOC-063-AC-02`, `VOC-063-TEST-01`,
`VOC-063-TEST-02`, `VOC-063-TEST-03`).

## Code changes

File updated: `apps/web/tests/staging-e2e/core-loop.staging.spec.ts`

- Removed `recordHomeResponseDiagnostic` and both call sites (steps 1 and 7).
- Added `readReviewedTodayCountAfterReviews` helper used by step 7.

## Step-7 retry parameters (`VOC-063-DEP-02`)

| Parameter | Value | Rationale |
|-----------|-------|-----------|
| Max attempts | 4 (initial read + up to 3 retries) | Within the 2–5 guardrail; enough tolerance for a transient stale `/home` load without masking a persistent regression. |
| Wait between attempts | 1,500 ms | Within the 500 ms–3 s guardrail; gives the server-rendered counter time to reflect completed reviews without adding large delay to the deploy gate. |
| Total deliberate wait budget | up to ~4.5 s (three inter-attempt waits) | Remains well inside `playwright.staging.config.ts`'s 240 s journey timeout. |
| Invariant | `reviewedAfter >= reviewedBefore + reviewedCards` | Unchanged; helper throws with all observed values when the bound is exhausted. |
| Annotations | `step-7-retry` when `attempt > 1` | Records attempt count, inputs, and every observed counter value. |

## Local validation (`VOC-063-TEST-03`)

```bash
pnpm --filter @vocanova/web lint          # exit 0
pnpm --filter @vocanova/api-client build  # prerequisite for workspace typecheck
pnpm --filter @vocanova/web typecheck     # exit 0
```

Staging E2E execution is deferred to `VOC-063-T02` (requires a real deploy).
