# VOC-045 — user_settings Upsert Omits created_at/updated_at from the INSERT Branch: Specification

## Objective and requirement source

Restore the ability of a genuinely new user (no prior `user_settings` row) to
complete onboarding, by fixing both known `user_settings` upsert call sites so
their INSERT branch supplies non-null `created_at`/`updated_at` values,
consistent with the table's `NOT NULL` constraint. Grounded in
[issue #341](https://github.com/KARSIFT/vocanova-platform-sandbox/issues/341)
in full, including its live production reproduction (real user, real
onboarding submission, exact `pq: null value in column "created_at" ...`
error), its cited schema evidence
(`apps/api/migrations/20260725130000_voc030_p4_user_settings.sql`), its exact
identification of both call sites and their SQL, and its suggested fix
(add both columns to each INSERT column list and VALUES, using each
function's own already-available timestamp source). Not yet approved by a
founder or technical steward — see `change.yaml`'s
`requirement_approval_status`.

## Scope and non-goals

In scope:
- `apps/api/business/users/postgres.go`'s `CompleteOnboarding`: add
  `created_at` and `updated_at` to the `user_settings` INSERT's column list
  and VALUES, using the `now time.Time` parameter already in scope for this
  function (already used for the `user_onboarding_profiles` insert in the
  same transaction, and for the `users` UPDATE immediately above it).
- `apps/api/business/gamification/repository.go`'s `UpsertUserSettings`: add
  `created_at` and `updated_at` to the `user_settings` INSERT's column list
  and VALUES. This function does not currently receive a `time.Time`
  parameter (see "Open questions" below for how it should source the value
  consistently with its own `ON CONFLICT DO UPDATE ... SET updated_at =
  NOW()` clause).
- A deterministic regression test (in `apps/api/migrations/migration_test.go`
  or a new test alongside each package's existing test suite — implementer to
  identify the best-fit location per each package's existing test
  conventions) confirming a genuinely fresh `user_settings` INSERT (no prior
  row, no `ON CONFLICT` path taken) succeeds for both call sites, so this
  specific defect class cannot silently regress.
- Confirming no other `user_settings` INSERT call site in the repository
  shares this same omission — a repository-wide grep for
  `INSERT INTO user_settings`, not a live-production check.

Non-goals:
- Any change to `user_settings`'s schema
  (`apps/api/migrations/20260725130000_voc030_p4_user_settings.sql`) or any
  new migration, unless the reviewing human decides at adoption time that a
  schema-level `DEFAULT now()` fix is also in scope (see "Open questions"
  below) — not assumed here, and this package's primary fix does not require
  one.
- Any change to `apps/api/app/api/onboarding.go`'s `mapOnboardingError`
  fallback-message behavior. The issue notes this fallback misleadingly
  presents the underlying database error as a generic validation error, but
  that is a distinct, narrower error-mapping/UX concern from this issue's
  reported root cause and requested fix — flagged as an open question rather
  than silently folded into this package's scope.
- Any change to the `daily_review_target`-preservation `CASE` logic already
  present in both `ON CONFLICT DO UPDATE` branches, or to any other column in
  either upsert's SET clause. The defect and its fix are scoped to the
  INSERT branch's missing `created_at`/`updated_at` columns only.
- Any change to `user_onboarding_profiles`'s own insert inside
  `CompleteOnboarding` — that insert already supplies `created_at`/`updated_at`
  correctly (via its own `$8, $8, $8` parameter reuse of `now`) and is not
  part of the reported defect.

## Risk and protected areas

This is a confirmed hard outage of the core new-user onboarding path in
production — every genuinely new user is currently blocked from completing
onboarding, per the issue's live reproduction. The fix touches two call
sites against `user_settings`, a table with existing production data (real
rows already exist for users who onboarded before this defect was exercised,
or via gamification's own lazy-creation path against an already-existing
row) and a `NOT NULL`-constrained schema, but requires no schema migration
in its primary fix. This package proposes `R3` (see `change.yaml`); it does
not, on its own, introduce a new migration, secret, or billing-adjacent
change, so no higher class is proposed by this draft, but the reviewing
human's own judgment governs this, not this proposal.
`scripts/governance/classify-change-risk.sh` has not been run against a
real, task-scoped file list at drafting time — consistent with how
VOC-039/VOC-040/VOC-041/VOC-042/VOC-043 handled this field; that computation
belongs to each task's own implementation PR.

## Decisions, contradictions, security, and privacy

No `VOC-045-D00`-style founder/product decision is defined by this draft.
Whether a schema-level `DEFAULT now()` fix is also in scope, and exactly how
`gamification/repository.go`'s `UpsertUserSettings` should source its
INSERT-branch timestamp given it has no `time.Time` parameter today, are
deliberately left open pending the reviewing human's decision — see "Open
questions" below — rather than this package guessing an answer in a
protected data-access path.

No contradiction between sibling documents is recorded for this package.

Security/privacy: this fix touches only how two existing, already-authorized
upsert operations populate two audit-timestamp columns for a row the
authenticated caller already owns (`user_id` is always the authenticated
session's own user ID at both call sites, per the surrounding code read at
drafting time). No new attacker-controlled input, new column, new personal-data
field, or change to who may read or write `user_settings` is introduced. No
secret or credential is involved.

## Data, migrations, analytics, and accessibility

Data: this fix changes what value two existing `NOT NULL` timestamp columns
receive on a genuine first INSERT; it does not change the shape of the
`user_settings` table, add or remove a column, or touch any other table.
Existing rows (already correctly populated via whichever path first
succeeded in creating them) are unaffected — this fix only changes behavior
for the specific INSERT path that was previously failing outright with a
constraint violation and therefore never persisted a row at all.

Migrations: none required for this package's primary fix (see "Non-goals"
above regarding the open question on a possible schema-level `DEFAULT`).

Analytics: none. This fix does not add, remove, or change any analytics
event.

Accessibility: none. This fix does not touch any UI; the observable
user-facing effect is that a request that previously failed with a `400`
now succeeds, using the same onboarding form and response shape already in
place.

## Open questions

1. **How should `apps/api/business/gamification/repository.go`'s
   `UpsertUserSettings` source its INSERT-branch `created_at`/`updated_at`
   value, given it has no `time.Time` parameter in scope today** (unlike
   `apps/api/business/users/postgres.go`'s `CompleteOnboarding`, which already
   receives a `now time.Time` parameter usable for this)? Two options exist:
   (a) use the SQL `NOW()` function directly in the INSERT's VALUES list,
   consistent with this same function's own existing `ON CONFLICT DO UPDATE
   ... SET updated_at = NOW()` clause, requiring no signature change; or (b)
   thread a `now time.Time` parameter through `UpsertUserSettings` (and its
   caller, `apps/api/business/gamification/service.go`) for testability and
   consistency with `CompleteOnboarding`'s pattern, requiring a small
   signature change and caller update. This package does not decide between
   them in advance — flagged for the implementer and reviewing human to
   settle in `VOC-045-T01`, favoring whichever keeps the diff minimal and
   consistent with this repository's existing conventions in each file.
2. **Whether a schema-level fix (adding `DEFAULT now()` to `user_settings`'s
   `created_at` and `updated_at` columns) is also in scope for this package**,
   closing off this entire class of omission at any future
   `INSERT INTO user_settings` call site, not just the two named here. The
   issue's own suggested fix targets the two known call sites directly and
   does not request a schema change; this package follows that scope by
   default, but flags the schema-level alternative for the reviewing human to
   explicitly accept or reject at adoption time rather than silently omitting
   it from consideration.
3. **Whether `apps/api/app/api/onboarding.go`'s `mapOnboardingError` fallback
   message (which the issue notes misleadingly presents this specific
   database error as a generic validation error) should be corrected as part
   of this package**, or is a separate, narrower follow-up. The issue raises
   this in passing while describing the symptom, not as its primary requested
   fix; this package treats it as out of scope by default (see "Non-goals"
   above) but flags it here so the reviewing human can pull it in explicitly
   if desired.
