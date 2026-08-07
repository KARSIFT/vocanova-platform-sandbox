# VOC-046 — Second Occurrence of the Missing created_at/updated_at Bug Class: Specification

## Objective and requirement source

Restore the ability of every user with no mission yet today to reach
`/home` in production, by fixing `daily_mission_snapshots`'s `INSERT`
call site so it supplies non-null `created_at`/`updated_at` values, and by
auditing every other `INSERT INTO` call site named by the issue for the same
omission so this bug class does not recur a third time undetected. Grounded
in [issue #352](https://github.com/KARSIFT/vocanova-platform-sandbox/issues/352)
in full, including its live production reproduction (nginx and web-container
log evidence for `GET /api/v1/daily-mission -> 500`, repeated on every
subsequent `/home` load), its cited schema evidence
(`apps/api/migrations/20260725130001_voc030_p4_mission_tables.sql`), its
exact identification of the confirmed root cause
(`apps/api/business/missions/repository.go`'s `CreateDailyMissionSnapshot`),
its documentation that this is a systemic, deliberate codebase convention
across all 13 migration files, its list of thirteen files to audit, and its
suggested fix (add both columns at every affected call site, plus regression
coverage and an optional schema-scanning detection check). Not yet approved
by a founder or technical steward — see `change.yaml`'s
`requirement_approval_status`.

## Scope and non-goals

In scope:

- `apps/api/business/missions/repository.go`'s `CreateDailyMissionSnapshot`:
  add `created_at` and `updated_at` to its `daily_mission_snapshots` `INSERT`
  column list and `VALUES`, sourced from whatever `time.Time`/clock value is
  already in scope at this call site (or threaded in if not currently
  available), matching the existing `ON CONFLICT DO UPDATE ... ` pattern's
  intent for the update branch, consistent with VOC-045-T00's approach for
  `user_settings`.
- Confirming the crash against a real Postgres `NOT NULL` check at
  implementation time (not just the issue's textual diagnosis), since
  `apps/api`'s generic `500` handling masked the underlying Postgres error
  this time, unlike VOC-045's `user_settings` occurrence.
- `apps/api/business/missions/repository.go`'s other `INSERT INTO
  daily_activity_summaries` statements in the same file/module (confirmed at
  drafting time to exist at four additional call sites in this file, per
  `implementation-plan.md`), since `daily_activity_summaries` shares the
  identical `NOT NULL`-no-`DEFAULT` schema on `created_at`/`updated_at`.
- A repository-wide audit of every other file the issue names as containing
  `INSERT INTO` statements against application tables (excluding
  `_test.go` files, as the issue's own grep already excluded):
  `apps/api/business/accounts/auth.go`,
  `apps/api/business/accounts/postgres.go`,
  `apps/api/business/accounts/service.go`,
  `apps/api/business/aifeedback/postgres.go`,
  `apps/api/business/auth/postgres.go`,
  `apps/api/business/gamification/repository.go` (the *other* `INSERT`
  statements in this file beyond the one already fixed by VOC-045-T01),
  `apps/api/business/gamification/rewards.go`,
  `apps/api/business/gamification/streak.go`,
  `apps/api/business/learning/postgres.go`,
  `apps/api/business/learning/postgres_idempotency.go`,
  `apps/api/business/reviews/postgres.go`,
  `apps/api/business/users/postgres.go` (any `INSERT` beyond the two already
  fixed by VOC-045-T00),
  `apps/api/business/users/seed.go`.
  For each file, cross-reference every `INSERT INTO <table>` statement found
  against that table's migration-declared `NOT NULL`-no-`DEFAULT` columns,
  and fix every one found missing `created_at`/`updated_at` (or any other
  such column) the same way.
- Deterministic regression coverage for every call site actually fixed by
  this package (the confirmed `daily_mission_snapshots` crash, any
  `daily_activity_summaries` call site found broken, and any call site found
  broken by the audit), following VOC-045-T02's failing-first-then-passing
  convention.
- A recorded, complete audit trail: which of the thirteen named files had a
  genuine defect, which did not (and why — e.g. already supplies both
  columns, doesn't write to a `NOT NULL`-no-`DEFAULT` table, or is dead
  code), and which fixes were applied.

Non-goals:

- Any change to any table's schema (e.g. adding `DEFAULT now()`), for the
  same reason VOC-045 excluded it: it would deviate from this codebase's
  deliberate application-level timestamp convention via the injected
  `clock.Clock` abstraction. Not assumed here as the fix; flagged as an open
  question below only insofar as the issue's own suggested schema-scanning
  *detection* check (not a schema *change*) may be in scope.
- Any change to `apps/api`'s generic `500` error handling that masked this
  issue's underlying Postgres error — see open question 3.
- Any change to a table or call site not among the thirteen files the issue
  names, unless the audit itself surfaces a concrete defect there (in which
  case it is documented as evidence, per `VOC-046-AC-03`, and fixed only if
  cleanly within this package's bounded scope — otherwise flagged as a
  follow-up, per this repository's `AGENTS.md` "stay within the approved
  scope; record unrelated improvements separately").

## Risk and protected areas

Builder assessment: this package's confirmed fix (`daily_mission_snapshots`)
is a small, self-contained code change with no schema migration, matching
VOC-045's precedent. The systemic-audit task's scope is bounded to the
thirteen named files but its exact size (how many call sites actually need a
fix) is not knowable until the audit runs — see open question 1. No task in
this package is known at drafting time to touch a schema migration,
credentials, or a governance document. Protected areas: none of the thirteen
files nor `apps/api/migrations/` are flagged elsewhere in this repository's
governance documents as requiring authority beyond routine R3 application
code review, but the reviewing human should confirm this against the current
`docs/governance/a003-transition-state.yaml` at adoption time, since this
package was not able to independently verify that document's current
content.

## Decisions, contradictions, security, and privacy

No `VOC-046-D00`-numbered decision is defined here; per this template's own
convention, decisions are only defined after approval. No contradiction was
found between the issue's request and canonical repository policy.
Authorization impact: none — this fix does not change any authorization
check, only the data written on insert. Secrets impact: none. Personal-data
impact: none — `created_at`/`updated_at` are operational timestamps, not
personal data, and this fix does not add, remove, or change any personal
data field. Abuse impact: none identified; a correctly-populated
`created_at`/`updated_at` reduces (does not increase) any risk of
inconsistent audit trails compared to today's crash-on-insert behavior.

## Data, migrations, analytics, and accessibility

Data: this fix corrects previously-impossible-to-create rows in
`daily_mission_snapshots` (and possibly `daily_activity_summaries`, per the
in-scope audit) to actually persist with valid, non-null
`created_at`/`updated_at`, matching the table's existing constraint intent.
No existing row is modified or migrated by this fix, since the defect
prevented row creation entirely rather than creating rows with bad data.
Migrations: none proposed by this package's primary fix, consistent with the
non-goal above — no `apps/api/migrations/` file is expected to change.
Analytics: none identified — this fix does not add, remove, or change any
analytics event. Accessibility: not applicable — this is a backend-only data
persistence fix with no user-interface surface of its own; the user-visible
symptom (a generic error page) is a downstream effect of the `500`, not a
change this package makes to any UI component.

## Open questions

1. `VOC-046-T02`'s audit scope boundary: how should the implementer document
   a named file that turns out to have no defect (e.g. already supplies both
   columns, or doesn't write to a `NOT NULL`-no-`DEFAULT` table)? This
   package's position: record it explicitly as "checked, no defect found,
   because ..." in the audit evidence (`VOC-046-EV-03`), rather than silently
   omitting it — the issue's own scope note treats "no defect found" as a
   completion state, not a gap, and the reviewing human should confirm this
   at adoption.
2. Is the issue's suggested schema-level or lint-style detection check (a
   test that walks each migration's `NOT NULL`-no-`DEFAULT` columns and
   confirms every discovered `INSERT` statement supplies them) in scope for
   this package, or a separate follow-up package? This package does not
   assume it is in scope; `tasks.md`'s `VOC-046-T03` scopes it as a
   candidate task pending the reviewing human's decision at adoption, since
   it is meaningfully larger and more novel (a static-analysis-style check,
   not a per-call-site fix) than this package's other tasks.
3. Is `apps/api`'s generic `500` handling — which masked this issue's
   underlying Postgres error, unlike VOC-045's `user_settings` occurrence
   that happened to surface via a `400` path — in scope for this package, or
   a separate, narrower observability follow-up (e.g. surfacing a
   correlation ID or structured error class in logs without changing the
   external `500` response contract)? This package does not assume it is in
   scope; flagged here for the reviewing human to decide.
