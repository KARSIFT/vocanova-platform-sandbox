# VOC-030 — P4 Daily Habit and Progress Staging Evidence

## Scope and authority

This document is the `T06` in-repository evidence index for the
P4 daily-habit and progress gate (DOC-12 §5 P4). It is updated at
`T06` implementation time to record:

1. The in-repository evidence (`EV-00`..`EV-37`) actually produced by
   the merged `T00`–`T05` PRs plus `T06`.
2. The staged exercises and rollback rehearsal procedures that can
   only run once F3 staging exists (`VOC-030-DEP-02`).
3. The `D05` (no retroactive backfill) interpretation note for any
   staging exercise that reuses pre-existing test accounts.
4. The P4 gate readiness summary — what is ready, what remains
   blocked, and what is explicitly out of scope.

The previous draft text (which was written **before adoption**,
2026-07-26, and predated the `T00`–`T05` work) is preserved in the
spec file's git history; this version supersedes it.

## In-repository evidence (produced by `T00`–`T06`)

| Evidence | Requirement | Status / source |
| --- | --- | --- |
| `EV-00` | New-table migration invariants (FKs, uniqueness, checks) | **Produced** — `apps/api/migrations/20260725130001_voc030_p4_mission_tables.sql`, `20260725130002_voc030_p4_gamification_tables.sql`; covered by `apps/api/migrations/migration_test.go` |
| `EV-01` | Migration compatibility and no A1/P1/P2/P3 regression | **Produced** — same migration set + the existing migration test suite (no A1/P1/P2/P3 schema altered) |
| `EV-02` | `user_settings` lazy creation and resolution chain (`D01`) | **Produced** — `apps/api/business/gamification/timezone.go` + `timezone_test.go`; `TestResolveSettingsStoredNonDefault`, `TestResolveSettingsStoredDefaultFallsThroughToClientTimezone`, `TestResolveSettingsNoStoredFallsThroughToClientTimezone`, `TestResolveSettingsNoStoredNoClientFallsToDefaults` |
| `EV-03` | Invalid client-supplied timezone rejected (`R02`) | **Produced** — `TestResolveSettingsRejectsInvalidClientTimezone`, `TestResolveSettingsRejectsInvalidStoredTimezone`, `TestLocalDate` (rejects `Not/A/Real/Zone`); also `TestGetDailyMissionRejectsInvalidClientTimezone` in `apps/api/app/api/missions_test.go`; T06 extension: `TestCrossCuttingNewReadsRejectInvalidClientTimezone` in `apps/api/app/api/missions_cross_cutting_test.go` |
| `EV-04` | Reward configuration matches DOC-06 §11 exactly | **Produced** — `apps/api/business/gamification/constants.go` (`RewardAddWord=2`, `RewardReviewAgain=1`, `RewardReviewHard=2`, `RewardReviewGood=5`, `RewardReviewEasy=6`, `RewardDailyMissionDone=10`, `RewardSentenceSubmitted=3`, `RewardAIFeedbackGot=2`); covered by `rewards_test.go` and `TestReviewRewardKind` in `reviews/postgres_p4_test.go` |
| `EV-05` | Streak reconciliation: advance, protect, break | **Produced** — `apps/api/business/gamification/streak.go` + `streak_test.go` (advance / protect / break / no-op tests); T06 extension: `TestCrossCuttingMultiDayGapReconciliationOnRead` in `missions/cross_cutting_safety_test.go` |
| `EV-06` | Grace-day earn/cap rule | **Produced** — `apps/api/business/gamification/streak.go` + `streak_test.go` (earn-every-7, cap-2) |
| `EV-07` | Per-source-event idempotency-key derivation is deterministic | **Produced** — `apps/api/business/gamification/idempotency.go` + `idempotency_test.go`; T06 extension: `TestCrossCuttingPerSourceEventIdempotencyKeysAreDistinct` in `missions/cross_cutting_safety_test.go` |
| `EV-08` | `missions`/`gamification` are transaction-scoped (no own-transaction opens) | **Produced** — every `missions` and `gamification` exported function takes an existing `*sql.Tx` and is exercised through the caller's existing transaction in the per-task T01–T03 tests |
| `EV-09` | Word-add reward: happy path and exactly-once | **Produced** — `apps/api/business/learning/postgres.go` (T01 wiring); covered by `learning/postgres_test.go`'s P1 success tests |
| `EV-10` | Word-add: failed insert awards nothing | **Produced** — `TestPostgreSQLRepositorySaveUserWordFailedInsertNoReward` in `learning/postgres_test.go` |
| `EV-11` | Pre-existing P1 behavior unchanged | **Produced** — `TestPostgreSQLRepositorySaveUserWordWithGamificationNil` in `learning/postgres_test.go` |
| `EV-12` | Review reward: rating-tiered amounts and exactly-once | **Produced** — `TestPostgreSQLRepositorySubmitReviewP4RatingGoodWiring` in `reviews/postgres_p4_test.go` |
| `EV-13` | Mission-completion transition on target met | **Produced** — `TestPostgreSQLRepositorySubmitReviewP4MissionCompletion` in `reviews/postgres_p4_test.go` |
| `EV-14` | Streak advance/protect/break driven from real review submissions | **Produced** — `TestPostgreSQLRepositorySubmitReviewP4RatingGoodWiring` and the streak_rehearsal_test (`TestReconcileStreakMultiDayGapBreak_RehearsalR06`) |
| `EV-15` | Replayed submission never double-rewards or double-completes | **Produced** — `TestPostgreSQLRepositorySubmitReviewP4AlreadyCompletedSnapshotNoDoubleReward` and `TestPostgreSQLRepositorySubmitReviewP4IdempotentMatchNoP4Wiring` in `reviews/postgres_p4_test.go` |
| `EV-16` | Pre-existing P2 behavior unchanged | **Produced** — `TestPostgreSQLRepositorySubmitReviewP4NilDependenciesNoP4Wiring` |
| `EV-17` | Real `MissionUpdater` replaces the stub | **Produced** — `TestAIFeedbackP4RealMissionUpdaterSuccessWiring` in `aifeedback/aifeedback_p4_test.go` |
| `EV-18` | Sentence/AI-feedback rewards: exactly-once, only on success | **Produced** — `TestUpdateForSentenceP4SuccessWiring` and `TestUpdateForSentenceP4IdempotencyKeyDerivation` in `missions/service_p4_test.go` |
| `EV-19` | Blocked/self-harm outcomes trigger no reward or mission update | **Produced** — `TestAIFeedbackP4RealMissionUpdaterSafetyBlockedNoReward` and `TestAIFeedbackP4RealMissionUpdaterSelfHarmNoReward` |
| `EV-20` | Pre-existing P3 behavior unchanged | **Produced** — `TestUpdateForSentenceP4RollbackOnError` (SQL-error rollback) and the safety outcomes above |
| `EV-21` | `GetDailyMission` contract and lazy creation | **Produced** — `TestGetDailyMissionLazilyCreatesAndReturnsProjection` and `TestGetDailyMissionReturnsExistingSnapshot` in `app/api/missions_test.go` |
| `EV-22` | `GetProgress` contract and shared streak object | **Produced** — `TestGetProgressReturnsBalanceStreakAndHistory`, `TestGetProgressEmptyHistory`, and `TestGetProgressSharedStreakObjectAgreesWithGetDailyMission` |
| `EV-23` | Authentication and self-scoping on both reads | **Produced** — `TestGetDailyMissionRequiresAuth`, `TestGetProgressRequiresAuth`, `TestGetDailyMissionCrossUserIsolation` (no path parameter on either endpoint) |
| `EV-24` | Confidence-points balance correctness | **Produced** — `TestGetProgressReturnsBalanceStreakAndHistory` (balance=42 read from ledger) |
| `EV-25` | Contract and OpenAPI drift | **Produced** — `TestContractContainsMissionsEndpoints` in `app/api/contract_test.go` |
| `EV-26` | Home wiring: mission progress bar and streak | **Produced** — `apps/web/src/app/(app)/home/page.tsx` now reads `reviewTarget`, `reviewsCompleted`, and `streak.currentStreakCount` from `getDailyMission()` |
| `EV-27` | Progress wiring: points, streaks, completion history | **Produced** — `apps/web/src/app/(app)/progress/page.tsx` now reads `confidencePointsBalance`, `streak`, and `completionHistory` from `getProgress()`, with `formatWeekdayLabel` deriving day-of-week from `localDate` |
| `EV-28` | Cross-capability consistency: Home and Progress agree | **Produced** — `TestGetProgressSharedStreakObjectAgreesWithGetDailyMission` (the streak object is byte-identical between the two endpoints) |
| `EV-29` | Loading, day-one-empty, and error states | **Produced** — `TestGetDailyMissionLazilyCreatesAndReturnsProjection` (no row path), `TestGetProgressEmptyHistory` (no history path), and the `error.tsx`/`loading.tsx` files in both routes |
| `EV-30` | Cross-cutting: duplicate actions across all three transactions | **Produced** — `TestCrossCuttingOnConflictSecondLineOfDefenseAcrossAllThreeWiredRewards` in `apps/api/business/missions/cross_cutting_safety_test.go` (exercises the `(user_id, idempotency_key)` partial unique index for P1 word-add, P2 review, P3 sentence-submitted, P3 AI-feedback-received) |
| `EV-31` | Cross-cutting: failed/incomplete actions across all three transactions | **Produced** — `TestCrossCuttingFailedActionsAcrossAllThreeWiredTransactions` in `apps/api/business/missions/cross_cutting_safety_test.go` (forces a SQL error mid-grant and asserts the helper returns the error and issues no follow-up SQL) |
| `EV-32` | Cross-cutting: unauthorized/cross-user requests never reach a reward path | **Produced** — `TestCrossCuttingUnauthenticatedRequestsNeverReachRewardPath` (unauth → 401, no SQL) and `TestCrossCuttingNoPathParameterOnNewReads` (OpenAPI contract guards cross-user enumeration) in `apps/api/app/api/missions_cross_cutting_test.go` |
| `EV-33` | Multi-day-gap streak break surfaces correctly on read (`R06`) | **Produced** — `TestCrossCuttingMultiDayGapReconciliationOnRead` in `apps/api/business/missions/cross_cutting_safety_test.go` (4-day gap, pre-existing 12/12 active streak; read surfaces broken/0 with longest preserved) |
| `EV-34` | Installed deterministic and security suite | **Produced** — `go test ./...` passes for the full `apps/api` module; the extended `scripts/foundation/mock-inventory.mjs` passes (prints `VOC-030-T06 mock inventory validation passed.`); the `apps/web/src/middleware.ts` matcher covers all `(app)` routes |
| `EV-35` | Staging save→review→submit-sentence→mission-completes→progress-reads | **Blocked by `VOC-030-DEP-02`** — F3 staging does not exist. Procedure is documented below; live execution is recorded as blocked. |
| `EV-36` | New-tables rollback rehearsal | **Blocked by `VOC-030-DEP-02`** — F3 staging does not exist. Procedure is documented below; live execution is recorded as blocked. |
| `EV-37` | Exact-SHA independent verification (per PR) | **Performed by Claude Code (different model binding) at each PR's exact final SHA** — this is not the implementer's evidence to record; it is produced by the independent-verification role and reported per-PR |

## Staging exercise plan (blocked by F3 + `D01`–`D05`)

Once `VOC-030-DEP-02` is resolved and a non-production F3 environment
is available, the following exercises must be executed and their
results appended to this document or linked from PR evidence.

### `EV-35` — Save → review → submit-sentence → mission-completes → progress-reads

1. With a non-production learner identity that has **no pre-existing
   activity** (per `D05`, no retroactive backfill applies; see
   `D05` interpretation note below), save a word, submit reviews
   until the daily review target is met, and submit a sentence for
   AI feedback.
2. Verify the mission transitions to `completed`, the streak
   advances by one, and the appropriate point awards appear in the
   ledger exactly once each.
3. Call `GET /api/v1/daily-mission` and `GET /api/v1/progress` and
   verify both agree with each other and with what Home / Progress
   render.
4. Repeat the review-submission and sentence-submission steps with a
   replayed idempotency key (P2's `client_attempt_id`, P3's
   request-hash) and verify no duplicate reward or double
   completion.

### `EV-36` — New-tables rollback rehearsal

1. Record current `daily_mission_snapshots`,
   `confidence_point_ledger`, `streak_states`, `grace_day_ledger`,
   and `user_settings` state for a test learner.
2. Apply the VOC-030 build and migration in staging.
3. Run several mission-completing flows, then perform a rollback to
   the previously known-good revision.
4. Verify that:
   - All committed `confidence_point_ledger` / `grace_day_ledger`
     rows created before the rollback remain immutable.
   - `daily_mission_snapshots` / `streak_states` / `user_settings`
     state is preserved.
   - The P1/P2/P3 write paths remain functional after rollback (the
     P3 `StubMissionUpdater` fallback keeps the AI-feedback flow
     itself working even if `missions` is rolled back independently
     of the P3 module it plugs into).

## `D05` interpretation note (no retroactive backfill)

Per the adopted `VOC-030-D05`, **no retroactive point / mission /
streak backfill** was performed for pre-activation P1–P3 activity.
This applies in two places:

- **Code / data at activation**: the `confidence_point_ledger`,
  `daily_mission_snapshots`, `daily_activity_summaries`,
  `streak_states`, `grace_day_ledger`, and `user_settings` tables
  start clean at this package's activation. No historical entries
  are fabricated for pre-activation `user_words`, `review_attempts`,
  `learner_sentences`, or `ai_feedback_attempts` rows.
- **Staging exercise interpretation (`EV-35`)**: any staging
  exercise that reuses a pre-existing test account (e.g. a shared
  `user_words` row created during the VOC-026 P1 evidence
  collection) must explicitly note in its results that the
  pre-activation activity did **not** produce points, mission
  progress, or streak state. The exercise is valid as a
  post-activation only; pre-existing activity contributes nothing
  to the post-activation reward calculations.

## Cross-capability consistency check (F3-blocked)

1. Load Home and Progress for the same non-production identity in
   the same session.
2. Verify the streak figure shown on each screen is identical.
3. Verify Home's mission progress bar matches the
   `reviewsCompleted` / `reviewTarget` values `GetDailyMission`
   returns at that instant.

**In-repository evidence for the consistency property** is provided
by `TestGetProgressSharedStreakObjectAgreesWithGetDailyMission`
(EV-28) — the same `StreakView` struct is built by both endpoints
through the same `gamification.loadStreakAndGrace` call, so they
cannot disagree at runtime.

## Rollback triggers

Per `VOC-030` implementation-plan §Deployment and rollback /
release-plan §Rollback, initiate rollback on:

- False mission / streak / point state reaching a learner.
- Suspected cross-user exposure of progress data.
- A confirmed duplicate reward in production.
- Inconsistent Home-vs-Progress figures.
- A regression in the underlying P1/P2/P3 write paths this package
  extends.
- Migration or schema failure.

## Rollback procedure

1. Preserve immutable `confidence_point_ledger` / `grace_day_ledger`
   history: never drop committed reward rows.
2. Preserve `daily_mission_snapshots` / `streak_states` /
   `user_settings` state.
3. Restore the pre-P4 P1/P2/P3 transaction behavior cleanly (verify
   the `StubMissionUpdater` fallback keeps P3 functional if the
   real `missions.MissionUpdater` is rolled back independently).
4. Revert the deployment to the last-known-good revision.
5. Validate with non-production identities.
6. Record the last-known-good revision and the rollback reason.

## Limitations / open dependencies

- **`VOC-030-DEP-02`**: F3 staging does not exist, so `EV-35` /
  `EV-36` cannot be run live. Procedures are documented; live
  execution is recorded as blocked.
- **`D01`–`D05`**: open founder decisions were resolved at adoption
  (2026-07-26) into the `D06` composite record (see
  `specification.md` §Decisions).
- **Accessibility automation for the Home/Progress wiring** is not
  yet implemented; the per-component accessibility is exercised
  visually (labelled progress indicators, visible focus, keyboard
  reachability, mobile layout) and recorded as a limitation, not
  a pass.

## P4 gate readiness

Per the DOC-12 §5 P4 gate wording: **"missions accurately reflect
completed behavior, progress is understandable, and
duplicate/failed/unauthorized actions can't create false
progress."**

- **Accurate mission reflection** — produced (`EV-00..EV-08`,
  `EV-12..EV-14`, `EV-17`, `EV-21..EV-25`).
- **Understandable progress** — produced (`EV-26..EV-29`).
- **Duplicate / failed / unauthorized actions can't create false
  progress** — produced (`EV-30..EV-33`, plus the per-task
  `EV-09`, `EV-10`, `EV-15`, `EV-18`, `EV-19`).

The P4 gate **cannot be declared complete by this work alone**
because:

1. The cross-cutting test suite is in-repository only; live
   `EV-35` (save → review → submit-sentence → mission-completes
   → progress-reads) and `EV-36` (new-tables rollback rehearsal)
   are blocked by the missing F3 staging environment
   (`VOC-030-DEP-02`).
2. Exact-SHA independent verification of every PR (T00..T06) is
   Claude Code's responsibility (`EV-37`); this implementer does
   not self-approve.
3. Production deployment is separately governed (A-003 §11/12);
   R3 / R4 founder authority, RL1/RL2 technical activation, and
   autonomous production release are all still disabled.

## Follow-up work

- Execute `EV-35` / `EV-36` once F3 staging exists.
- P5: fold the daily-habit loop into the full cross-feature
  integration and reliability pass.
- A future Settings package: expose `user_settings` through a real
  API / UI on top of the `user_settings` table this package
  creates minimally (`D01`).
