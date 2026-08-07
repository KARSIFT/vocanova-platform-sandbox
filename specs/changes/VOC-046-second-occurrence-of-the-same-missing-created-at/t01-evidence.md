# VOC-046-EV-01 — T01 daily_activity_summaries audit and fix evidence

Evidence for `VOC-046-AC-01` / `VOC-046-TEST-01`.

## Audit scope and outcome

Audited file: `apps/api/business/missions/repository.go`.

Migration cross-reference:
`apps/api/migrations/20260725130001_voc030_p4_mission_tables.sql` declares
`daily_activity_summaries.created_at` and `daily_activity_summaries.updated_at`
as `timestamptz NOT NULL` with no `DEFAULT`.

Every `INSERT INTO daily_activity_summaries` statement in this file previously
omitted both columns on the fresh-insert branch. All five statements are now
fixed to include `created_at, updated_at` and `NOW(), NOW()` in the insert
column/value list, while keeping each existing `ON CONFLICT ... DO UPDATE`
branch unchanged.

Call-site outcomes:

1. `IncrementReviewsCompleted` — fixed (added `created_at`, `updated_at`).
2. `IncrementWordsAdded` — fixed (added `created_at`, `updated_at`).
3. `IncrementSentenceSubmitted` — fixed (added `created_at`, `updated_at`).
4. `IncrementAIFeedbackReceived` — fixed (added `created_at`, `updated_at`).
5. `IncrementConfidencePointsEarned` — fixed (added `created_at`, `updated_at`).

No "already-correct" daily_activity_summaries insert call site remained in this
module at implementation time.

## Regression coverage added

`apps/api/business/missions/repository_postgres_integration_test.go` now
includes `TestDailyActivitySummaryFreshInsertPathsAgainstRealPostgres`, which
uses a real disposable Postgres and exercises each fixed call site on a true
fresh-insert path (precondition: no prior `daily_activity_summaries` row for
`(user_id, local_date)`).

For each subtest, it asserts:

- the call succeeds on first insert,
- the expected activity counter is persisted, and
- persisted `created_at`/`updated_at` are non-null and within the measured
  insert window.

This provides deterministic real-schema proof for every call site fixed by
`VOC-046-T01`.
