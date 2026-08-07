# VOC-045-EV-02 / VOC-045-EV-03 — T02 regression coverage and call-site inventory

## Regression coverage added (VOC-045-EV-02)

- `apps/api/business/users/postgres_test.go`:
  `TestPostgreSQLRepositoryCompleteOnboardingFreshUserSettingsInsertSuppliesTimestamps`
  exercises `CompleteOnboarding` with a fresh `user_settings` insert path and
  requires the insert SQL to include `created_at` and `updated_at`.
- `apps/api/business/gamification/repository_test.go`:
  `TestRepositoryUpsertUserSettingsFreshInsertSuppliesTimestamps` exercises
  `UpsertUserSettings` with a fresh `user_settings` insert path and requires the
  insert SQL to include `created_at` and `updated_at`.

Failing-first verification against pre-fix SQL:

```bash
go test ./business/users ./business/gamification
```

Both new tests fail before T00/T01 SQL changes are applied, because the current
queries still omit `created_at`/`updated_at` from the `INSERT INTO user_settings`
column list.

## Repository-wide inventory (VOC-045-EV-03)

Search command:

```bash
rg "INSERT INTO user_settings" apps/api
```

Runtime call sites found:

1. `apps/api/business/users/postgres.go` — `CompleteOnboarding` upsert (`INSERT INTO user_settings`).
2. `apps/api/business/users/postgres.go` — `UpdateSettings` upsert (`INSERT INTO user_settings`).
3. `apps/api/business/gamification/repository.go` — `UpsertUserSettings` upsert (`INSERT INTO user_settings`).

T00/T01 scope explicitly covers (1) and (3). Call site (2) is an additional
fresh-insert path discovered by the required inventory step and needs explicit
scope/deferral handling before AC-03 can be considered fully satisfied.
