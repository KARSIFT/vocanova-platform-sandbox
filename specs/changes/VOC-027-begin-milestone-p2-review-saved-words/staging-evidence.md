# VOC-027 — P2 Review Staging Evidence

## Scope and authority

This document collects the in-repository staging evidence required by
`VOC-027-AC-06` and `VOC-027-T04`. It is produced under the adopted resolutions
`VOC-027-D01` through `VOC-027-D06`.

Live staging exercises (EV-21, EV-22, EV-23) are blocked by the open
environmental dependency `VOC-027-DEP-02`: the F3 staging environment does not
exist yet. The procedures below are documented and ready to run once F3 is
available. This document does **not** declare the DOC-12 P2 milestone gate
complete.

## In-repository evidence (ready now)

| Evidence | Requirement | Status | Location |
| -------- | ----------- | ------ | -------- |
| `EV-19` | Home `dueReviewWords` wired to real due-queue | Ready | `apps/web/src/app/(app)/home/page.tsx` fetches `GET /api/v1/reviews/due` and uses `totalCount`; `MOCK_HOME_STATE` no longer contains `dueReviewWords`. |
| `EV-20` | Mock inventory enforces P2 dispositions | Ready | `scripts/foundation/mock-inventory.mjs` + `mock-inventory.test.mjs` verify the decommissioned `dueReviewWords` field, the retained P4 mocks, and the authorized API/business/schema/migration boundary. |
| `EV-13` | P2 boundary enforced (API routes, business modules, schemas, migrations) | Ready | `scripts/foundation/mock-inventory.mjs` asserts no unexpected P3/P4 routes, tables, modules, or migrations; `apps/api/business/reviews/` is the only new P2 module; `apps/api/ent/schema/reviewattempt.go` is the only new P2 schema; `apps/api/migrations/20260725110000_voc027_p2_review_attempts.sql` is the only new P2 migration. |
| `EV-24` | No P4 tables or behavior invented | Ready | No `daily_mission_snapshots`, `confidence_point_ledger`, `streak_states`, or `daily_activity_summaries` schema, route, or business logic exists. Confirmed by mock inventory and the implementation scope. |

## Staging exercise plan (blocked by F3)

Once `VOC-027-DEP-02` is resolved and a non-production F3 environment is
available, the following exercises must be executed and their results recorded
here or in the PR evidence.

### EV-21 — Due-queue → submit → schedule-update-exactly-once → completion

1. With a non-production learner identity, save at least one word so it enters
   the due queue (`next_review_at` null, `status` in `new`/`learning`/`reviewing`).
2. Call `GET /api/v1/reviews/due` and verify the word appears with accurate
   `wordText`, `shortDefinition`, `status`, and `reviewStep`.
3. Submit a `POST /api/v1/reviews/submissions` response with a valid
   `X-CSRF-Token` and a fresh `Idempotency-Key` / `client_attempt_id`.
4. Verify the response contains the inserted `review_attempts` row with the
   matching `client_attempt_id`, `review_step_before`, and `review_step_after`.
5. Re-query `GET /api/v1/reviews/due` and confirm the word no longer appears
   (or the `totalCount` decreased exactly by one) and the `user_words` row shows
   the updated `review_step`, `next_review_at`, `last_result`, `last_rating`,
   `last_reviewed_at`, and incremented counters.
6. Drain the queue by reviewing all due words, then confirm the session shows
   the "You're all caught up" empty state and the Home page reports `0` words
   due.

### EV-22 — Cross-user denial, CSRF, and idempotency

1. **Cross-user denial:** As learner A, create a saved word. Attempt to submit
   or query the due queue as learner B. Expect `401` for unauthenticated
   requests and `404` for owner mismatches (no cross-learner enumeration).
2. **CSRF:** Submit a review without `X-CSRF-Token`. Expect `403` and no
   `review_attempts` row or `user_words` update.
3. **Idempotency:** Replay the same `client_attempt_id` with the same request
   fingerprint. Expect `200` with the original attempt, no duplicate
   `review_attempts` row, and no double update to `user_words`. Reuse the same
   key with a changed fingerprint. Expect `409` and no schedule change.

### EV-23 — Rollback rehearsal

1. Record the current `review_attempts` and `user_words` state for a test
   learner.
2. Apply the VOC-027 P2 build and migration in the staging environment.
3. Run a few review submissions, then perform a rollback to the previously
   known-good revision.
4. Verify that:
   - All committed `review_attempts` rows created before the rollback remain
     immutable (no accidental `ON DELETE CASCADE` or migration deletion).
   - `user_words` schedule state is preserved for rows written before the
     rollback window.
   - No daily-mission, point-ledger, streak-state, or P4 tables were created.
   - The application health checks and the due-queue API remain functional.

## Rollback triggers

Per `VOC-027` implementation-plan §Deployment and rollback, initiate rollback on:

- Cross-learner access to due words or attempts.
- Idempotency failure (duplicate schedule update or duplicate attempt row).
- Scheduling-rule violation (step outside 0–7, floor/cap error, or two
  consecutive incorrect/Again not resetting to 0).
- Migration integrity fault (missing FK, partial-unique index, or check
  constraint; accidental cascade on `review_attempts` or `user_words`).
- Leaked secret, token, or PII in logs/responses.
- Failed health checks or inability to complete the due-queue → submit →
  completion flow.

## Rollback procedure

1. Preserve immutable `review_attempts` history: never drop committed attempt
   rows.
2. Preserve `user_words` schedule state written before the rollback window.
3. Revert the deployment to the last-known-good revision.
4. Validate with non-production identities.
5. Record the last-known-good revision and the rollback reason.

## Limitations / open dependencies

- `VOC-027-DEP-02`: F3 staging does not exist, so EV-21/EV-22/EV-23 cannot be run
  live. The procedures and in-repository evidence are complete; live execution
  is recorded as blocked.
- Accessibility automation for the review card controls is not yet implemented;
  this is recorded as a limitation, not a pass.

## Follow-up work

- Replace the retained P4-only mock fields in `HomePage` and `ProgressPage`
  (`missionTargetWords`, `reviewedWordsToday`, `currentStreakDays`,
  `confidencePointsTotal`, `longestStreakDays`, `completionHistory`) when P4
  gamification, streak, and daily-mission systems are implemented.
- Implement `typing` and `sentence_usage` prompt types as a later superset
  (P3/P4, per DOC-05 §9).
- Add the daily-mission, point-ledger, streak-state, and daily-activity-summary
  steps to the submission transaction once the P4 tables exist.
