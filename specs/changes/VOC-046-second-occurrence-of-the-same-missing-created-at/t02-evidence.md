# VOC-046-EV-02 / VOC-046-EV-03 / VOC-046-EV-04 — T02 audit and regression evidence

This document records the `VOC-046-T02` audit for the twelve files scoped by
the task (the thirteen-file package list minus `apps/api/business/missions/repository.go`,
which belongs to `T00/T01`), plus the fixes and deterministic regression
coverage added for every call site that was found defective.

## Audit results by file and `INSERT INTO` statement

### `apps/api/business/accounts/auth.go`

- `INSERT` statements found: none.
- Outcome: no defect in scope for this file because there are no insert call
  sites to audit.

### `apps/api/business/accounts/postgres.go`

- `CreateEmailChangeLink` -> `INSERT INTO email_change_links (...)`
  - Migration check: `email_change_links` requires `created_at`; no
    `updated_at` column exists.
  - Outcome: already correct (`created_at` supplied).
- `CreateAccountDeletionRequest` -> `INSERT INTO account_deletion_requests (...)`
  - Migration check: requires `created_at` and `updated_at`.
  - Outcome: already correct (both supplied).

### `apps/api/business/accounts/service.go`

- `INSERT` statements found: none.
- Outcome: no defect in scope for this file because there are no insert call
  sites to audit.

### `apps/api/business/aifeedback/postgres.go`

- `CreatePendingAttempt` -> `INSERT INTO learner_sentences (...)`
  - Migration check: requires `created_at` and `updated_at`.
  - Outcome: already correct (both supplied).
- `CreatePendingAttempt` -> `INSERT INTO ai_feedback_attempts (...)`
  - Migration check: requires `created_at` and `updated_at`.
  - Outcome: already correct (both supplied).

### `apps/api/business/auth/postgres.go`

- `CreateUser` -> `INSERT INTO users (...)`
  - Migration check: requires `created_at` and `updated_at`.
  - Outcome: already correct.
- `CreateSession` -> `INSERT INTO sessions (...)`
  - Migration check: requires `created_at`; no `updated_at` column exists.
  - Outcome: already correct.
- `CreateMagicLink` -> `INSERT INTO magic_links (...)`
  - Migration check: requires `created_at`; no `updated_at` column exists.
  - Outcome: already correct.
- `CreateOAuthState` -> `INSERT INTO oauth_states (...)`
  - Migration check: requires `created_at`; no `updated_at` column exists.
  - Outcome: already correct.
- `CreateExternalIdentity` -> `INSERT INTO external_identities (...)`
  - Migration check: requires `created_at` and `updated_at`.
  - Outcome: already correct.

### `apps/api/business/gamification/repository.go`

- `InsertPointLedger` -> `INSERT INTO confidence_point_ledger (...)`
  - Migration check: requires `created_at` and `updated_at`.
  - Outcome: **fixed**. Added both columns and values (`NOW(), NOW()`).
- `InsertGraceLedger` -> `INSERT INTO grace_day_ledger (...)`
  - Migration check: requires `created_at` and `updated_at`.
  - Outcome: **fixed**. Added both columns and values (`NOW(), NOW()`).
- `UpsertStreakState` -> `INSERT INTO streak_states (...)`
  - Migration check: requires `created_at` and `updated_at`.
  - Outcome: **fixed**. Added both columns and values (`NOW(), NOW()`).
- `UpsertUserSettings` -> `INSERT INTO user_settings (...)`
  - Migration check: requires `created_at` and `updated_at`.
  - Outcome: already correct (covered previously by VOC-045 work and remains
    compliant).

### `apps/api/business/gamification/rewards.go`

- `INSERT` statements found: none.
- Outcome: no defect in scope for this file because there are no insert call
  sites to audit.

### `apps/api/business/gamification/streak.go`

- `INSERT` statements found: none.
- Outcome: no defect in scope for this file because there are no insert call
  sites to audit.

### `apps/api/business/learning/postgres.go`

- `SaveUserWord` -> `INSERT INTO user_words (...)`
  - Migration check: requires `created_at` and `updated_at`.
  - Outcome: already correct.

### `apps/api/business/learning/postgres_idempotency.go`

- `Record` -> `INSERT INTO idempotency_keys (...)`
  - Migration check: requires `created_at`; no `updated_at` column exists.
  - Outcome: already correct.

### `apps/api/business/reviews/postgres.go`

- `SubmitReview` -> `INSERT INTO review_attempts (...)`
  - Migration check: requires `created_at` and `updated_at`.
  - Outcome: already correct.

### `apps/api/business/users/postgres.go`

- `CompleteOnboarding` -> `INSERT INTO user_onboarding_profiles (...)`
  - Migration check: requires `created_at` and `updated_at`.
  - Outcome: already correct.
- `CompleteOnboarding` -> `INSERT INTO user_settings (...)`
  - Migration check: requires `created_at` and `updated_at`.
  - Outcome: already correct (one of the previously-fixed VOC-045 call sites).
- `UpdateSettings` -> `INSERT INTO user_settings (...)`
  - Migration check: requires `created_at` and `updated_at`.
  - Outcome: **fixed**. Added both columns and bound `now` for fresh-insert
    path values.

### `apps/api/business/users/seed.go`

- `INSERT` statements found: none.
- Outcome: no defect in scope for this file because there are no insert call
  sites to audit.

## Regression coverage for fixed call sites (`VOC-046-EV-04`)

Added deterministic sqlmock tests that fail if the fixed call sites regress and
omit timestamp columns again:

- `apps/api/business/gamification/repository_test.go`
  - strengthened `InsertPointLedger` test to require `created_at` and
    `updated_at` in the insert column list.
  - strengthened `InsertGraceLedger` test to require `created_at` and
    `updated_at` in the insert column list.
  - added `TestRepositoryUpsertStreakStateFreshInsertSuppliesTimestamps`.
- `apps/api/business/users/postgres_test.go`
  - added `TestPostgreSQLRepositoryUpdateSettingsFreshInsertSuppliesTimestamps`
    to require `created_at` and `updated_at` in the `UpdateSettings` fresh
    insert SQL.

## Validation run

- `cd apps/api && go test ./business/gamification ./business/users` -> pass.
