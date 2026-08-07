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

The same file also adds
`TestDailyActivitySummaryOnConflictBranchUnchangedAgainstRealPostgres`,
`test-plan.md`'s negative coverage for this task (the counterpart to
`VOC-046-TEST-00`'s snapshot on-conflict test): each of the five call sites
is driven twice, in separate committed transactions, and the pre-existing
`ON CONFLICT ... DO UPDATE` behavior is asserted unchanged — one row rather
than a duplicate, the call site's counter accumulated, `created_at`
preserved, and `updated_at` advanced.

This provides deterministic real-schema proof for every call site fixed by
`VOC-046-T01`.

## Failing-first verification (pre-fix code)

Verified directly against real Postgres, not asserted. With `T01`'s
`created_at, updated_at` columns and their `NOW(), NOW()` values temporarily
removed from all five production statements (and then restored
byte-for-byte — confirmed by an identical `md5sum` of `repository.go`
against the pre-revert copy):

```bash
cd apps/api && go test -tags=integration -count=1 -v ./business/missions/... -run 'TestDailyActivitySummary'
--- FAIL: TestDailyActivitySummaryFreshInsertPathsAgainstRealPostgres (14.94s)
    --- FAIL: TestDailyActivitySummaryFreshInsertPathsAgainstRealPostgres/IncrementReviewsCompleted (6.50s)
        Error: upsert activity summary reviews: pq: null value in column "created_at" of relation "daily_activity_summaries" violates not-null constraint
    --- FAIL: TestDailyActivitySummaryFreshInsertPathsAgainstRealPostgres/IncrementWordsAdded (2.11s)
        Error: upsert activity summary words_added: pq: null value in column "created_at" of relation "daily_activity_summaries" violates not-null constraint
    --- FAIL: TestDailyActivitySummaryFreshInsertPathsAgainstRealPostgres/IncrementSentenceSubmitted (2.12s)
        Error: upsert activity summary sentences_submitted: pq: null value in column "created_at" of relation "daily_activity_summaries" violates not-null constraint
    --- FAIL: TestDailyActivitySummaryFreshInsertPathsAgainstRealPostgres/IncrementAIFeedbackReceived (2.11s)
        Error: upsert activity summary ai_feedback_received: pq: null value in column "created_at" of relation "daily_activity_summaries" violates not-null constraint
    --- FAIL: TestDailyActivitySummaryFreshInsertPathsAgainstRealPostgres/IncrementConfidencePointsEarned (2.10s)
        Error: upsert activity summary points earned: pq: null value in column "created_at" of relation "daily_activity_summaries" violates not-null constraint
--- FAIL: TestDailyActivitySummaryOnConflictBranchUnchangedAgainstRealPostgres (10.59s)
    --- FAIL: TestDailyActivitySummaryOnConflictBranchUnchangedAgainstRealPostgres/IncrementReviewsCompleted (2.10s)
    --- FAIL: TestDailyActivitySummaryOnConflictBranchUnchangedAgainstRealPostgres/IncrementWordsAdded (2.11s)
    --- FAIL: TestDailyActivitySummaryOnConflictBranchUnchangedAgainstRealPostgres/IncrementSentenceSubmitted (2.11s)
    --- FAIL: TestDailyActivitySummaryOnConflictBranchUnchangedAgainstRealPostgres/IncrementAIFeedbackReceived (2.11s)
    --- FAIL: TestDailyActivitySummaryOnConflictBranchUnchangedAgainstRealPostgres/IncrementConfidencePointsEarned (2.16s)
FAIL	github.com/KARSIFT/vocanova-platform/apps/api/business/missions	25.534s
```

Each of the five call sites raises the same Postgres error (23502) issue
#352 diagnosed, one per call site, against the real committed schema — so
every fix in this task is individually failing-first, not covered
collectively by one test. The on-conflict test fails pre-fix for the same
reason: its first call is itself a fresh insert, so it cannot reach the
update branch until the fix is in place.

## Passing verification (post-fix code)

```bash
cd apps/api && go test -tags=integration -count=1 -v ./business/missions/... -run 'TestDailyActivitySummary'
--- PASS: TestDailyActivitySummaryFreshInsertPathsAgainstRealPostgres (10.59s)
    --- PASS: TestDailyActivitySummaryFreshInsertPathsAgainstRealPostgres/IncrementReviewsCompleted (2.11s)
    --- PASS: TestDailyActivitySummaryFreshInsertPathsAgainstRealPostgres/IncrementWordsAdded (2.11s)
    --- PASS: TestDailyActivitySummaryFreshInsertPathsAgainstRealPostgres/IncrementSentenceSubmitted (2.12s)
    --- PASS: TestDailyActivitySummaryFreshInsertPathsAgainstRealPostgres/IncrementAIFeedbackReceived (2.11s)
    --- PASS: TestDailyActivitySummaryFreshInsertPathsAgainstRealPostgres/IncrementConfidencePointsEarned (2.14s)
--- PASS: TestDailyActivitySummaryOnConflictBranchUnchangedAgainstRealPostgres (10.76s)
    --- PASS: TestDailyActivitySummaryOnConflictBranchUnchangedAgainstRealPostgres/IncrementReviewsCompleted (2.12s)
    --- PASS: TestDailyActivitySummaryOnConflictBranchUnchangedAgainstRealPostgres/IncrementWordsAdded (2.17s)
    --- PASS: TestDailyActivitySummaryOnConflictBranchUnchangedAgainstRealPostgres/IncrementSentenceSubmitted (2.16s)
    --- PASS: TestDailyActivitySummaryOnConflictBranchUnchangedAgainstRealPostgres/IncrementAIFeedbackReceived (2.14s)
    --- PASS: TestDailyActivitySummaryOnConflictBranchUnchangedAgainstRealPostgres/IncrementConfidencePointsEarned (2.17s)
ok  	github.com/KARSIFT/vocanova-platform/apps/api/business/missions	21.353s
```

The full default API suite is also green, with no other package affected:

```bash
cd apps/api && go test -count=1 ./...
ok  	github.com/KARSIFT/vocanova-platform/apps/api/app/api	0.277s
ok  	github.com/KARSIFT/vocanova-platform/apps/api/business/gamification	0.015s
ok  	github.com/KARSIFT/vocanova-platform/apps/api/business/missions	0.011s
ok  	github.com/KARSIFT/vocanova-platform/apps/api/business/users	0.007s
ok  	github.com/KARSIFT/vocanova-platform/apps/api/migrations	0.027s
# ...all other packages ok / no test files; no FAIL lines
```

`gofmt -l ./business/missions/` and `go vet -tags=integration
./business/missions/` are both clean.

Build-tag trade-off: identical to `t00-evidence.md`'s — this test lives
behind `-tags=integration` and skips cleanly without Docker, because adding
a Postgres service container to the shared CI workflow is not possible from
this repository and is out of `VOC-046-T01`'s scope.
