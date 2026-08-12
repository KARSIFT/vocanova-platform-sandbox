# VOC-074-EV-02 — T02 evidence (staging E2E hardening)

Evidence for `VOC-074-T02` (`VOC-074-AC-03`, test `VOC-074-TEST-04`).

**Task:** VOC-074-T02  
**Package:** VOC-074  
**Investigation basis:** `t00-evidence.md` confirmed queue exhaustion /
`reviewedCards = 0` vacuous step-7 pass as the cause of false confidence on runs
31575459316 / 31583230574 — not a residual increment write-path defect.

## Changes

### 1. `apps/web/tests/staging-e2e/core-loop.staging.spec.ts`

- **Step 5 gate (VOC-074-D00):** after the review loop, `expect(reviewed).toBeGreaterThanOrEqual(1)` fails before step 6/7 when the queue was empty or caught up. Step 7's `reviewedAfter >= reviewedBefore + reviewedCards` invariant is unchanged; VOC-063 step-7 retry constants (`STEP_7_REVIEWED_COUNT_MAX_ATTEMPTS`, `STEP_7_REVIEWED_COUNT_RETRY_DELAY_MS`) are untouched.
- **Step 5 reporting:** pushes a `step-5-reviewed-cards` Playwright annotation and logs `[staging core-loop] step 5 reviewedCards=N` to stdout.
- **Step 7 reporting:** on every successful read (pass or fail path that returns), pushes a `step-7-reviewed-counts` annotation with `reviewedBefore`, `reviewedCards`, `reviewedAfter`, and `minimumExpected`, and logs the same to stdout. Retry attempts still add a separate `step-7-retry` annotation when needed.

### 2. `apps/api/scripts/seed-synthetic-smoke-user.sql` (VOC-074-DEP-03)

T00 recommended a minimal queue reset so repeat same-day deploys can reach
`reviewedCards >= 1` without weakening the increment assertion. The seed script
already runs on every staging/production deploy before the core-loop gate.

Added an idempotent `UPDATE` that sets `next_review_at = NOW()` on the most
recently updated `user_words` row for the synthetic account when any saved words
exist. Accounts with no saved words yet still rely on step 4 of the journey to
create a due card.

**Not edited:** `.github/workflows/deploy-staging.yml` — seed path is sufficient
and matches the diagnostic dump's existing `docker compose exec postgres psql`
assumptions.

## VOC-074-DEP-03 resolution

**Chosen:** extend `seed-synthetic-smoke-user.sql` (not a new deploy-staging step).

Reason: the seed already runs immediately after migrations on every deploy; one
due word is enough for step 5 to review at least one card; no new workflow
surface or SSH-only step ordering.

## How to verify locally

```bash
# Confirm the staging spec is still isolated from PR-time e2e
cd apps/web
pnpm exec playwright test --config playwright.staging.config.ts --list

# Lint/typecheck for the touched web package (per docs/development.md)
pnpm --filter @vocanova/web lint
pnpm --filter @vocanova/web typecheck
```

Code review confirms: with `reviewedCards = 0`, step 5 throws and step 7 is never
reached — vacuous pass eliminated.

## Acceptance mapping

- **VOC-074-AC-03:** satisfied by step-5 failure + always-on step-7 count logging.
- **VOC-074-AC-05:** VOC-063 retry preserved; no VOC-065-T01 revert; VOC-065-T02 not
  cited as closure (T03 still required).
