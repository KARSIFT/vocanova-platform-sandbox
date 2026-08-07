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

### Sequencing blocker — T02 cannot go green on its own

`VOC-045-AC-02` requires these tests to fail against the pre-fix code and pass
against the post-fix code. Neither `VOC-045-T00` nor `VOC-045-T01` is present on
`develop` or on this branch at the time of writing (`git log origin/develop`
tips at `c1a25ca`, the package's task-issue roster commit; both call sites still
carry the four-column INSERT). A test that satisfies AC-02's failing-first
requirement therefore necessarily red-lights `pnpm test` / `go test ./...` until
the T00 and T01 fixes land.

This is a task-ordering problem, not a defect in the tests. Resolving it inside
`VOC-045-T02` would require either editing the two production SQL statements
that `VOC-045-T00`/`T01` explicitly own (scope expansion) or neutering the
assertion so it no longer detects the defect (weakening a check to make the
change pass). Neither is available to the implementer role. Merge `VOC-045-T00`
and `VOC-045-T01` first; this branch turns green with no further edits once they
are in the merge base.

The insert-column matcher is deliberately the only constraint on the statement:
it does not pin the placeholder numbering T00/T01 choose for `created_at` /
`updated_at`, nor whether `UpsertUserSettings` sources its timestamp from SQL
`NOW()` or a threaded `now` parameter (`specification.md`'s open question 1), so
either resolution of that open question satisfies these tests.

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
