# VOC-074-T01 — Write-path fix (no-op) and regression coverage mapping

**Task:** VOC-074-T01  
**Package:** VOC-074  
**Evidence ID:** VOC-074-EV-01  
**Root cause (from T00):** synthetic-account same-day queue exhaustion
(`reviewedCards = 0`) masked by a vacuous E2E step 7 pass — not a residual
post-VOC-065-T01 increment write-path defect on the cited runs  
**Investigation date:** 2026-08-12

## Fix applied

**No product code change.** T00 (`t00-evidence.md`) confirmed the cited staging
symptom (runs 31575459316 / 31583230574) is explained by zero review submissions
during the E2E journey, not by a defect in `SubmitReview` → `applyP4ReviewWiring`
→ `IncrementReviewsCompleted` after VOC-065-T01's composition-root wiring fix.

The `updated_at`-without-increment pattern (`07:50:56` with `reviews_completed = 0`)
matches lazy snapshot creation on the daily-mission read path, not a partial write
failure. A committed review with P4 wiring present must advance `reviews_completed`
(T00 §2–3).

Per `tasks.md`, when T00 routes the product-facing fix to T02 (E2E hardening +
optional queue reset), T01 is explicitly a no-op. **T02** carries the fix for
false-confidence staging gate behavior; **T03** proves increments when
`reviewedCards >= 1` on real staging.

VOC-065-T01's wiring fix in `apps/api/app/api/production.go` is preserved
unchanged.

## Regression coverage (existing — no new test added)

T00 scoped out a new API regression test. The confirmed gap for the cited runs is
the E2E vacuous pass (`reviewedCards = 0` still reaches a green step 7); that
regression belongs in **T02** (`core-loop.staging.spec.ts`).

Existing deterministic tests already guard the write path T00 traced and rule out
re-introducing the pre-VOC-065-T01 nil-wiring defect:

| Test | Location | What it guards |
|------|----------|----------------|
| `TestProductionReviewsRepositoryWiresP4Dependencies` | `apps/api/app/api/production_test.go` | `newProductionReviewsRepository` wires both P4 deps (`HasP4Wiring() == true`). |
| `TestProductionGo_NewProductionAPIConstructsP4WiredReviewsRepository` | `apps/api/app/api/production_test.go` | Live composition root passes `gamSvc` and `missionsSvc` into the reviews repository. |
| `TestPostgreSQLRepositorySubmitReviewP4NilDependenciesNoP4Wiring` | `apps/api/business/reviews/postgres_p4_test.go` | Nil gamification/missions skips P4 SQL entirely (characterization of the pre-T01 failure mode). |
| `TestPostgreSQLRepositorySubmitReviewP4RatingGoodWiring` | `apps/api/business/reviews/postgres_p4_test.go` | Wired `SubmitReview` issues `EnsureTodaySnapshot` then `IncrementReviewsCompleted` (`reviews_completed` 0 → 1) in the same transaction. |
| `TestPostgreSQLRepositorySubmitReviewP4MissionCompletion` | `apps/api/business/reviews/postgres_p4_test.go` | Increment path at review target boundary (completion grant, no double-award). |
| `TestPostgreSQLRepositoryIncrementReviewsCompleted` | `apps/api/business/missions/repository_test.go` | Repository-layer increment SQL and cap at `review_target`. |
| `TestPostgreSQLRepositoryIncrementReviewsCompletedSnapshotMissing` | `apps/api/business/missions/repository_test.go` | Increment against a missing snapshot row fails (no silent no-op). |

A new API test would not fail under the confirmed T00 cause (queue exhaustion with
no submissions); it would only duplicate coverage already present from VOC-065-T01
and the P4 characterization suite. Any second defect surfaced when T02/T03
guarantee `reviewedCards >= 1` should be addressed in a follow-up task with a
targeted regression test.

## How to run

From the repository root (Go toolchain required):

```bash
cd apps/api
go test ./app/api -run 'TestProductionReviewsRepositoryWiresP4Dependencies|TestProductionGo_NewProductionAPIConstructsP4WiredReviewsRepository' -count=1
go test ./business/reviews/... -run 'P4|SubmitReview' -count=1
go test ./business/missions/... -run 'IncrementReviewsCompleted' -count=1
```

Governance validation (no product diff expected for this task):

```bash
bash scripts/governance/validate-governance.sh
bash scripts/governance/classify-change-risk.sh
git diff --check
```

## Commands run during T01

```bash
cd apps/api
go test ./app/api -run 'TestProductionReviewsRepositoryWiresP4Dependencies|TestProductionGo_NewProductionAPIConstructsP4WiredReviewsRepository' -count=1
# ok github.com/KARSIFT/vocanova-platform/apps/api/app/api 0.008s

go test ./business/reviews/... -run 'P4|SubmitReview' -count=1
# ok github.com/KARSIFT/vocanova-platform/apps/api/business/reviews 0.007s

go test ./business/missions/... -run 'IncrementReviewsCompleted' -count=1
# ok github.com/KARSIFT/vocanova-platform/apps/api/business/missions 0.003s

bash scripts/governance/validate-governance.sh
# Repository foundation validation passed.
# Governance structure validation passed.
```

## Acceptance criterion mapping

- **VOC-074-AC-01:** satisfied by existing VOC-065-T01 wiring plus P4
  characterization tests — T00 found no residual write-path defect to fix on
  the cited runs. Runtime proof with `reviewedCards >= 1` is deferred to T03.
- **VOC-074-AC-02:** satisfied by mapping existing regression tests above; no
  new test added because the confirmed cause is E2E-side (T02 scope). Existing
  tests fail if nil-wiring or skipped P4 increment is re-introduced.
- **VOC-074-AC-05:** satisfied — VOC-065-T01 wiring untouched; VOC-063 retry
  bounds untouched; no migration; VOC-065-T02 not cited as closure.
- **VOC-074-TEST-01 / TEST-02 / TEST-03:** covered by commands and test table
  above on this revision.

## Out of scope (unchanged)

- No migration / historical backfill (`VOC-074-DEP-01`).
- No staging E2E change — **T02**.
- No staging deploy verification — **T03**.
- No revert of VOC-065-T01.

## Downstream implications

| Task | Direction |
|------|-----------|
| **T02** | Required — fail when `reviewedCards < 1`; always log counter integers; consider queue reset (`VOC-074-DEP-03`). |
| **T03** | Prove on real staging that `reviewedCards >= 1` yields DB-consistent increments; closes whether any second defect remains. |
