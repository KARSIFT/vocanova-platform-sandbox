# VOC-074-T00 — Residual write-path root-cause evidence

**Task:** VOC-074-T00  
**Package:** VOC-074  
**Evidence ID:** VOC-074-EV-00  
**Investigation date:** 2026-08-12  
**Reviewed revision:** working tree on branch implementing VOC-074-T00 (develop tip)

## Confirmed root cause (cited runs 31575459316 / 31583230574)

**Synthetic-account same-day queue exhaustion (`reviewedCards = 0`) masked by a
vacuous E2E step 7 pass** — not a residual post-VOC-065-T01 increment write-path
defect on these runs.

For runs
[31575459316](https://github.com/KARSIFT/vocanova-platform-sandbox/actions/runs/31575459316)
(~07:47 UTC) and
[31583230574](https://github.com/KARSIFT/vocanova-platform-sandbox/actions/runs/31583230574)
(~09:31 UTC), the authoritative diagnostic dump showing `reviews_completed = 0`
with `updated_at = 2026-08-12 07:50:56` frozen across ~1h43m is consistent with
**no successful review submissions** during either E2E journey. Step 7's invariant
`reviewedAfter >= reviewedBefore + reviewedCards` passes vacuously when
`reviewedCards = 0`, and Playwright does not log `reviewedBefore` /
`reviewedCards` / `reviewedAfter` on success — so a green suite hides the case.

The `updated_at`-without-increment clue from issue #539 is explained by **lazy
snapshot row creation on the read path**, not by a partial write-path failure.

**T01 product-code scope:** no-op unless T03 surfaces a new defect when
`reviewedCards >= 1` is guaranteed. **T02** (fail when `reviewedCards < 1`,
always log counter integers, optional queue reset per `VOC-074-DEP-03`) carries
the product-facing fix for the false-confidence gate.

## Evidence chain

### 1. VOC-065-T01 wiring is present and deployed before the cited runs

VOC-065-T01 (PR #523, commit `8962c86`) merged at `2026-08-12 01:33:40 +0330`
(~`2026-08-11 22:03:40` UTC). Both cited runs are later the same UTC calendar
day. `deploy-staging.yml` deploys on every push to `develop`, so the post-T01
composition root should be live for both runs.

Current production wiring (VOC-065-T01 fix preserved):

```663:669:apps/api/app/api/production.go
	gamRepo := gamification.NewRepository(db)
	gamSvc := gamification.NewService(gamRepo)
	missionsRepo := missions.NewRepository(db)
	missionsSvc := missions.NewService(missionsRepo, gamSvc)

	reviewsRepo := newProductionReviewsRepository(db, clk, gamSvc, missionsSvc)
	reviewsSvc := reviews.NewService(reviewsRepo, learningIdem, clk)
```

```756:764:apps/api/app/api/production.go
func newProductionReviewsRepository(db *sql.DB, clk clock.Clock, gamSvc *gamification.Service, missionsSvc *missions.Service) *reviews.PostgreSQLRepository {
	return reviews.NewPostgreSQLRepository(db, clk,
		reviews.WithGamificationService(gamSvc),
		reviews.WithMissionsService(missionsSvc),
	)
}
```

Static guards: `TestProductionReviewsRepositoryWiresP4Dependencies` and
`TestProductionGo_NewProductionAPIConstructsP4WiredReviewsRepository` in
`apps/api/app/api/production_test.go`.

**Runtime P4 guard at deploy time:** not directly traced on live staging from
this environment (no staging credentials / GitHub Actions log access). Ruled out
as the explanation for the **cited runs' DB pattern** because that pattern
matches zero submissions, not nil-wiring submissions (see §4 and §5).

### 2. `updated_at` advanced once while `reviews_completed` stayed `0`

Two SQL paths touch `daily_mission_snapshots.updated_at` **without** incrementing
`reviews_completed`:

**A. Lazy read-path creation** (`GetDailyMissionView` when today's row is
missing):

```130:147:apps/api/business/missions/service.go
	// Lazy snapshot creation: if today's row does not exist yet, create it
	snap, err := s.missions.GetDailyMissionSnapshot(ctx, userID, today)
	// ...
	if snap == nil {
		snap, err = s.missions.CreateDailyMissionSnapshot(
			ctx, tx, userID, today, resolved.Timezone,
			resolved.DailyReviewTarget, gamification.MissionPolicyVersion,
		)
```

**B. `CreateDailyMissionSnapshot` ON CONFLICT branch** (also invoked by
`EnsureTodaySnapshot` on every review, but paired with increment in the same
transaction — see §3):

```154:157:apps/api/business/missions/repository.go
		ON CONFLICT (user_id, local_date) DO UPDATE
		  SET timezone = EXCLUDED.timezone,
		      review_target = EXCLUDED.review_target,
		      updated_at = NOW()
```

Neither path changes `reviews_completed`. A fresh INSERT also starts at
`reviews_completed = 0` with `updated_at = NOW()` (lines 148–152).

**Interpretation for run 31575459316:** E2E step 2 (`readReviewedTodayCount` via
`GET /api/v1/daily-mission` → home SSR) likely created today's row around
`07:50:56` UTC — the first access on the new local calendar day for the synthetic
account. That sets `updated_at` once while leaving `reviews_completed = 0`.

**Interpretation for run 31583230574:** `updated_at` unchanged across ~1h43m
means **no code path touched the row again** — consistent with no review
submissions and no second lazy-create (row already existed).

If `applyP4ReviewWiring` had committed even one review after T01, the same
transaction would run `IncrementReviewsCompleted`, which **must** raise
`reviews_completed` to at least `1`:

```188:193:apps/api/business/missions/repository.go
		`UPDATE daily_mission_snapshots
		 SET reviews_completed = LEAST(reviews_completed + 1, review_target),
		     updated_at = NOW()
		 WHERE user_id = $1 AND local_date = $2
		 RETURNING reviews_completed`,
```

A committed review with wiring present cannot leave `reviews_completed = 0`.

### 3. Runtime increment path (code trace)

When both dependencies are wired, `SubmitReview` always enters P4 wiring:

```306:309:apps/api/business/reviews/postgres.go
	if r.gamification != nil && r.missions != nil {
		if err := r.applyP4ReviewWiring(ctx, tx, req, attemptID, now); err != nil {
			return nil, err
		}
	}
```

`applyP4ReviewWiring` sequence:

1. `gamification.GetSettings` — timezone / review target (empty client TZ → D01
   UTC default for synthetic account with no `user_settings` row)
2. `missions.EnsureTodaySnapshot` — idempotent `CreateDailyMissionSnapshot`
3. `missions.IncrementReviewsCompleted` — uses **`snap.LocalDate`** from step 2

```363:386:apps/api/business/reviews/postgres.go
	snap, err := r.missions.EnsureTodaySnapshot(ctx, tx, req.UserID, resolved, now)
	// ...
	newReviewsCompleted, err := r.missions.IncrementReviewsCompleted(
		ctx, tx, req.UserID, snap.LocalDate, resolved.Timezone,
		resolved.DailyReviewTarget, correct, skipped,
	)
```

All writes share one transaction with P2 `review_attempts` / `user_words`; any
increment failure rolls back the review submission. The UI does not advance
without a successful API response:

```146:153:apps/web/src/app/(app)/reviews/_components/review-session.tsx
      const { data } = await client.submitReview(body, clientAttemptId, {
        headers: { "X-CSRF-Token": csrfToken },
      });
      setLastReviewedCard(currentCard);
      setLastReviewAttemptId(data.attemptId);
      setRemainingCount((count) => Math.max(0, count - 1));
      advance();
```

**Local deterministic check:** `go test ./business/reviews/... -run 'P4|SubmitReview'`
passed on this revision (sqlmock characterization tests including wired and
nil-deps paths).

**Live staging trace of one `SubmitReview`:** not executed (no staging DB/API
access). Not required to explain the cited runs' DB state.

### 4. Write-key mismatch — ruled out for cited runs

Diagnostic dump filters by synthetic email and orders by `local_date DESC`
(`deploy-staging.yml` lines 1325–1331). Write and read paths both resolve
`local_date` via `gamification.LocalDate(now, resolved.Timezone)` with the same
D01 settings chain.

When P4 wiring runs, increment uses the **`snap.LocalDate` returned by
`EnsureTodaySnapshot` in the same transaction** — there is no separate date
computation on the write path. A key mismatch would require P4 wiring to update
a different row than the diagnostic/read path sees; here **P4 increment did not
run** (zero submissions), so mismatch is not implicated.

### 5. Queue exhaustion vs masked write failure — `reviewedCards` disposition

**Conclusion for cited runs: queue exhaustion (or equivalent empty queue at step
5), not a masked increment failure.**

Reasoning:

| Signal | Observed | Interpretation |
|--------|----------|----------------|
| Diagnostic `reviews_completed` (today) | `0` both runs | No committed increments |
| Diagnostic `updated_at` (today) | `07:50:56` then frozen | One lazy-create touch; no review-path writes |
| E2E result | `1 passed` both runs | Step 7 can pass with `reviewedCards = 0` |
| Playwright stdout on pass | No `reviewedBefore` / `reviewedCards` logged | Vacuous pass is invisible in logs |
| VOC-065-T01 merge time | ~6h before first cited run | Nil-wiring defect not the default hypothesis |

**Step 5 allows zero reviews without failing:**

```332:374:apps/web/tests/staging-e2e/core-loop.staging.spec.ts
    const reviewedCards = await test.step("5. work the review queue", async () => {
      // ...
      let reviewed = 0;
      while (reviewed < MAX_REVIEW_CARDS) {
        await expect(caughtUpHeading.or(cardCounter).first()).toBeVisible();
        if (await caughtUpHeading.isVisible()) {
          break;
        }
        const didReview = await reviewOneCard(page);
        if (!didReview) {
          break;
        }
        reviewed++;
        await expect(caughtUpHeading.or(cardCounter).first()).toBeVisible();
      }
      return reviewed;
    });
```

When the synthetic account is caught up (`"You're all caught up"`), the loop exits
with `reviewed = 0`. This is **expected rerun behavior**: the spec documents
rerun safety on a stateful account (`core-loop.staging.spec.ts` lines 32–35);
`seed-synthetic-smoke-user.sql` refreshes account identity but does **not** reset
due-review backlog or mission counters.

**Step 7 passes vacuously:**

```417:425:apps/web/tests/staging-e2e/core-loop.staging.spec.ts
    await test.step("7. progress reflects the completed reviews", async () => {
      const reviewedAfter = await readReviewedTodayCountAfterReviews(
        page, testInfo, reviewedBefore, reviewedCards,
      );
      expect(reviewedAfter).toBeGreaterThanOrEqual(reviewedBefore + reviewedCards);
    });
```

With `reviewedBefore = 0`, `reviewedCards = 0`, any `reviewedAfter >= 0`
passes — including `reviewedAfter = 0` matching the diagnostic dump.

**Contrast with pre-T01 run 31429774964 (VOC-065-T00):** that run had
`reviewedCards = 2` with persistent `reviews_completed = 0` — the signature of
P2 commits with skipped P4 wiring. The cited VOC-074 runs lack any evidence of
reviews performed (`updated_at` frozen after initial lazy create; no counter
movement).

**Direct log proof of `reviewedCards = 0` on runs 31575459316 / 31583230574:**
not available in this environment (GitHub Actions API returned 401; workflow logs
not fetched). The queue-exhaustion conclusion follows from DB timing evidence +
E2E control-flow analysis above. T02's required logging will make this observable
on future runs.

## Candidate summary

| Candidate | Status for cited runs | Notes |
|-----------|----------------------|-------|
| Queue exhaustion / `reviewedCards = 0` vacuous pass | **Confirmed** | Explains dump + frozen `updated_at` + green E2E |
| `updated_at` without increment (lazy create / ON CONFLICT) | **Confirmed mechanism** | Explains single `07:50:56` touch, not a defect |
| Residual nil P4 wiring post-T01 | **Ruled out** as cause of cited symptom | Wiring fixed; DB pattern is zero submissions |
| `(user_id, local_date)` write-key mismatch | **Ruled out** | No increment path executed |
| `IncrementReviewsCompleted` SQL defect | **Ruled out for cited runs** | Would require committed reviews; none observed |
| Increment works when `reviewedCards >= 1` post-T01 | **Inconclusive** | Requires T03 with guaranteed-non-empty queue |

## Implications for downstream tasks

| Task | Direction |
|------|-----------|
| **T01** | **No product fix** for the confirmed cause. Optional: document that existing P4 wiring + unit tests already cover the increment path; defer any new regression test to a T03-discovered defect. |
| **T02** | **Required:** fail (default) when `reviewedCards < 1`; always annotate/log `reviewedBefore`, `reviewedCards`, `reviewedAfter`. Consider minimal queue reset/seeding (`VOC-074-DEP-03`) so step 5 reliably reaches `reviewedCards >= 1` on repeat same-day deploys. |
| **T03** | Prove on real staging: `reviewedCards >= 1`, step 7 passes with real increments, diagnostic dump moves consistently. This closes whether any second defect remains after T01. |

## Commands run during T00

Read-only code inspection plus one local test invocation. No staging DB query, no
workflow rerun, no product code changes.

```bash
# Static reads (ripgrep + file reads) across:
# apps/api/app/api/production.go
# apps/api/business/reviews/postgres.go
# apps/api/business/missions/repository.go
# apps/api/business/missions/service.go
# apps/web/tests/staging-e2e/core-loop.staging.spec.ts
# apps/web/src/app/(app)/reviews/_components/review-session.tsx
# .github/workflows/deploy-staging.yml
# apps/api/scripts/seed-synthetic-smoke-user.sql

git show -s --format='%ci %s' 8962c86   # VOC-065-T01 merge time

cd apps/api && go test ./business/reviews/... -run 'P4|SubmitReview' -count=1
# ok github.com/KARSIFT/vocanova-platform/apps/api/business/reviews 0.007s
```

## Acceptance criterion mapping

- **VOC-074-AC-00:** satisfied — confirmed cause named with file/line evidence;
  `reviewedCards = 0` disposition explained as queue exhaustion masked by vacuous
  step 7; all drafting-time candidates confirmed, ruled out, or marked
  inconclusive (increment-with-reviews post-T01 → T03).
- **VOC-074-TEST-00:** satisfied by this document (`VOC-074-EV-00`).
