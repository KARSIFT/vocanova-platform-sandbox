# VOC-045-EV-02 / VOC-045-EV-03 — T02 regression coverage and call-site inventory

## Regression coverage added (VOC-045-EV-02)

- `apps/api/business/users/postgres_test.go`:
  `TestPostgreSQLRepositoryCompleteOnboardingFreshUserSettingsInsertSuppliesTimestamps`
  drives `CompleteOnboarding` through a fresh `user_settings` insert (the
  `sqlmock` expectation returns a row directly from the INSERT, so no
  `ON CONFLICT` update branch is exercised) and requires the insert SQL to
  include `created_at` and `updated_at`.
- `apps/api/business/gamification/repository_test.go`:
  `TestRepositoryUpsertUserSettingsFreshInsertSuppliesTimestamps` does the same
  for `UpsertUserSettings`, and additionally asserts the returned
  `UserSettingsRow` carries the expected user ID, timezone, and daily review
  target (`VOC-045-AC-01`'s "returns the expected `UserSettingsRow`").

Each test matches on the INSERT column list only. It does not pin the
placeholder numbering `T00`/`T01` chose, nor whether the timestamp is sourced
from SQL `NOW()` or a threaded `now` parameter (`specification.md`'s open
question 1), so either resolution of that open question satisfies both tests.
`sqlmock` collapses every whitespace run in the actual statement to a single
space before matching, so the shared `userSettingsInsertColumnsPattern` in each
package tolerates optional whitespace around the parentheses rather than
assuming how the SQL string literal happens to be wrapped — `CompleteOnboarding`
keeps its column list on one line while `UpsertUserSettings` wraps it onto its
own line, and both must match the same intent.

### Failing-first verification

Verified directly, not asserted. With `T00`'s and `T01`'s `created_at,
updated_at` columns temporarily removed from both production statements (and
then restored byte-for-byte):

```bash
cd apps/api && go test ./business/gamification/... ./business/users/... \
  -run 'FreshInsertSuppliesTimestamps|FreshUserSettingsInsertSuppliesTimestamps'
--- FAIL: TestRepositoryUpsertUserSettingsFreshInsertSuppliesTimestamps (0.00s)
FAIL	github.com/KARSIFT/vocanova-platform/apps/api/business/gamification	0.003s
--- FAIL: TestPostgreSQLRepositoryCompleteOnboardingFreshUserSettingsInsertSuppliesTimestamps (0.00s)
FAIL	github.com/KARSIFT/vocanova-platform/apps/api/business/users	0.003s
```

Against the post-fix code now in this branch's merge base (`T00` via
`develop`, `T01` via PR #348), the full API suite passes:

```bash
cd apps/api && go test ./...
ok  	github.com/KARSIFT/vocanova-platform/apps/api/business/gamification	0.004s
ok  	github.com/KARSIFT/vocanova-platform/apps/api/business/users	0.005s
# ...all other packages ok / no test files; no FAIL lines
```

The earlier revision of this document recorded a sequencing blocker: `T00` and
`T01` were not yet in this branch's merge base, so these tests necessarily
red-lighted CI. That blocker is resolved — both fixes are now present and this
branch is green with no change to either production statement by `T02`.

## Repository-wide inventory (VOC-045-EV-03)

Search command (repository-wide, not scoped to `apps/api`):

```bash
rg "INSERT INTO user_settings"
```

Every runtime call site, with its `created_at`/`updated_at` status:

| # | Call site | Supplies `created_at`/`updated_at`? |
|---|---|---|
| 1 | `apps/api/business/users/postgres.go` — `CompleteOnboarding` | Yes, fixed by `VOC-045-T00` (`$5, $5` from the `now` parameter) |
| 2 | `apps/api/business/gamification/repository.go` — `UpsertUserSettings` | Yes, fixed by `VOC-045-T01` (SQL `NOW(), NOW()`) |
| 3 | `apps/api/business/users/postgres.go` — `UpdateSettings` (line 322) | **No** — see follow-up below |

The remaining matches are non-runtime: the two test files above, and this
package's own specification, tasks, acceptance-criteria, test-plan, and
implementation-plan documents.

### Follow-up: call site 3 shares the same defect and is outside this package's scope

`UpdateSettings`'s upsert omits `created_at`/`updated_at` from its INSERT
column list in exactly the same way the two fixed call sites did, so a genuine
first insert through it would hit the same `NOT NULL` violation. It is not named
in issue #341, not in `specification.md`'s scope, and not owned by `T00`, `T01`,
or `T02`, so `T02` deliberately does not fix it inline — that would be scope
expansion into a protected data-access path. It is recorded here and in the
implementation PR as a follow-up needing its own change package (or an explicit
scope extension decided by the reviewing human).

This is the same class of defect `specification.md`'s open question 2 anticipates:
a schema-level `DEFAULT now()` on `user_settings.created_at`/`updated_at` would
close off call site 3 and any future one at once. That decision remains with the
reviewing human and is out of scope here.
