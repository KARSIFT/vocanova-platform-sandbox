# VOC-053-EV-01 — T01 root-cause fix: investigation, ruling, and decision

Evidence for `VOC-053-T01` (`VOC-053-AC-00`, `VOC-053-AC-01`, `VOC-053-TEST-01`,
`VOC-053-TEST-02`). This is attempt 1 of T01.

## Status of this attempt: scope cannot be completed as written

**`VOC-053-T01`'s premise is that a confirmed root cause from `VOC-053-T00`
exists and a narrowly-scoped fix can be implemented against it.** Per
`tasks.md`'s T01 status line and the `implementer` role prompt, T01 must
not begin a fix "until T00's evidence names a specific, evidence-backed
cause." T00's evidence does not name a specific, evidence-backed cause.
It rules out two candidates by direct evidence/static analysis and
leaves the remaining candidates explicitly unconfirmed pending live
verification the implementer environment cannot perform. As such, this
attempt is the honest "the fix is not safe to write yet" deliverable
T00's own "T01 entry conditions" section anticipates, not a partial or
scope-expanded code change.

## What this attempt re-investigated (and re-confirmed) from T00

T00's addendum (recorded at the bottom of `t00-evidence.md`) already
ruled out candidate (a) (Next.js Data Cache / Full Route Cache) with
direct live evidence from a real `deploy-staging.yml` run: the `/home`
response shows `cache-control: private, no-cache, no-store,
max-age=0, must-revalidate`, no `x-nextjs-cache` header, and
`cf-cache-status: DYNAMIC` on both the step 1 baseline and the step 7
post-review load, and the failure was reproduced on the same run. I
did not re-run that live check (this environment has no live-staging
network access) and treated the T00 addendum's ruling as settled.

T00's `findings/VOC-053-DEP-01` (also in `t00-evidence.md`) traced the
read path from `apps/api/app/api/missions.go` down through
`missions.Service.GetDailyMissionView` →
`gamification.Service.GetSettings` → `gamification.ResolveSettings` →
`gamification.LocalDate` → `missions.Repository.GetDailyMissionSnapshot`
/ `CreateDailyMissionSnapshot` and concluded no read-path bug could
produce a per-request drift in `local_date` or a decremented
`reviews_completed` between two requests seconds apart at 19:42 UTC
under the documented conditions. I independently re-read every file
this conclusion depends on, on the current `develop` base
(2da3505), and re-confirm the conclusion. The full re-read is
recorded below; the original conclusion stands unchanged.

I also re-read the previously-not-read write paths (P1 word-add,
P2 review submission, P3 sentence feedback, the `MarkSnapshotMissed`
helper, the accounts-anonymization `UPDATE`) for any place that could
decrement `reviews_completed` or delete the row out from under the
read. No such place exists in committed code. The detailed audit is
in `findings/VOC-053-write-path-audit` below.

## T00 entry-condition check (from `t00-evidence.md`)

Per T00's own "T01 entry conditions" section, T01 may proceed under
either of:

1. "A specific root cause is confirmed with direct evidence ...
   in the form of an addendum to this evidence file"; or
2. "the live verification narrows the candidates enough that the
   implementer can pick the highest-probability candidate and scope
   the fix to that, recording the chosen candidate and why it was
   chosen over the others in T01's own evidence file."

Condition 1 is not met: T00's addendum ruled out (a) only. (b) is
ruled out only by static analysis, not by direct evidence. (c) and
"fourth cause" remain candidates this environment cannot confirm.
There is no candidate with a direct, evidence-backed confirmation
matching the wording of condition 1.

Condition 2 is not met either. T00's addendum narrowed the field
to "candidate (b) is now the leading candidate by elimination," but
this T01 attempt's full re-investigation of (b)'s read path
(recorded below) found no defect to fix. There is no
"highest-probability candidate" left to scope a fix to that would
not be, per `tasks.md`'s T01 framing, "a fix that makes the test
pass without the counter actually being correct" — i.e. a retry,
poll-until-pass, or assertion-loosening workaround rather than a
real fix.

This is the exact case T00's "T01 entry conditions" section
forecasts: "If neither holds, T01 is not yet safe to start; that is
the honest answer this investigation pass can give, per the
drafting pass's own 'or not fully ruled out, if evidence is
inconclusive — record that honestly too' requirement."

## What this attempt's re-investigation of (b)'s read path looked at

Files re-read in full on the current `develop` base (2da3505) for
this T01 attempt, with the relevant lines and findings:

### apps/api/app/api/missions.go (HTTP handler)

The `GetDailyMission` handler at
`apps/api/app/api/missions.go:99-105` calls

```go
view, err := svc.GetDailyMissionView(ctx, RequesterUserID(ctx), input.Timezone, time.Now())
```

`time.Now()` is evaluated per request — no caching, memoization, or
request-scope sharing. Two requests seconds apart see two `now`
values a few seconds apart, not the same instant, but both fall in
the same UTC day for the documented failure conditions (~19:42 UTC,
no midnight boundary).

`input.Timezone` is the optional client-supplied IANA timezone
query parameter (`GetDailyMissionInput.Timezone` at
`apps/api/app/api/missions.go:60-62`). The `apps/web`
`VocanovaClient.getDailyMission` call does **not** supply it
(`apps/web/src/app/(app)/home/page.tsx:14` calls
`client.getDailyMission()` with no arguments;
`packages/api-client/src/index.ts:649-663` only sets
`query.set("timezone", params?.timezone)` if
`params?.timezone` is truthy), and the T00 evidence already
verified there is exactly one place in `apps/web/src/` rendering
the `(\d+) of \d+ words reviewed today` text. So `input.Timezone`
is `""` for both of the test's `/home` reads.

`RequesterUserID(ctx)` is the per-request authenticated user ID
resolved from the session cookie; it is stable across the test's
two reads because the same session cookie is used.

### apps/api/business/missions/service.go (service layer)

`GetDailyMissionView` at
`apps/api/business/missions/service.go:116-206`:

- Line 122: `s.gamification.GetSettings(ctx, userID, clientTimezone)`.
  For the synthetic account, `GetUserSettings` returns `nil` (the
  seed script does not create a `user_settings` row, and onboarding
  is skipped because `onboarding_status = 'completed'` per
  `apps/api/scripts/seed-synthetic-smoke-user.sql:74`). So
  `ResolveSettings` falls through to the third step of the D01
  chain and returns
  `ResolvedSettings{Timezone: "UTC", DailyReviewTarget: 20}` per
  `apps/api/business/gamification/timezone.go:81`. Both reads see
  the same `resolved.Timezone = "UTC"`.

- Line 126: `today, err := gamification.LocalDate(now, resolved.Timezone)`.
  `LocalDate` is a pure function of `(now, timezone)` with no I/O.
  For two requests seconds apart at 19:42 UTC, both `now` values
  are within `2026-08-09T19:42:xxZ`, so both produce
  `today = 2026-08-09T00:00:00Z` (midnight UTC of the local day).
  Both reads compute the same `today`.

- Line 134: `tx, err := s.missions.db.BeginTx(ctx, nil)`. The
  transaction is begun but **not used for the read** — see line
  139.

- Line 139: `snap, err := s.missions.GetDailyMissionSnapshot(ctx, userID, today)`.
  The repository's read at
  `apps/api/business/missions/repository.go:28-39` uses `r.db`
  (a fresh connection from the pool), **not** the transaction just
  begun. The SQL is a simple
  `SELECT ... FROM daily_mission_snapshots WHERE user_id = $1 AND
  local_date = $2` with no `NOW()` or `current_date` involvement
  and no joins. The `(user_id, local_date)` unique index is
  created in
  `apps/api/migrations/20260725130001_voc030_p4_mission_tables.sql:35-36`.

- Line 143-181: lazy-create + streak-reconciliation block uses `tx`
  (the transaction) for any writes. The reconciliation's only writes
  are to `streak_states` (via `s.gamification.UpsertStreakState`)
  and `grace_day_ledger` (via `s.gamification.GrantGraceDay`); it
  does not touch `daily_mission_snapshots` or `reviews_completed`.

- Line 182: `tx.Commit()`. If `snap != nil` (the path that returns
  the prior run's residue), the transaction held no writes and
  commits as a no-op. If `snap == nil` (lazy creation), the
  reconciliation writes are committed.

- Line 196: `view.ReviewsCompleted = snap.ReviewsCompleted`. The
  value passed back to the client is exactly the value the SQL
  read returned, with no rounding, clamping, or post-processing
  in this function (the only `Math.min`/`Math.round` in
  `apps/web/src/app/(app)/home/page.tsx:29-32` affects the `(NN%)`
  suffix, not the integer itself).

**Per-request drift assessment, re-confirmed on this attempt's
base:** with the documented test conditions (no midnight boundary
at 19:42 UTC, synthetic account has no stored `user_settings` row,
no client timezone is supplied, both `now` values are within
`2026-08-09T19:42:xxZ`), there is **no code path in this trace
that could produce a different `local_date` or a different
`snap.ReviewsCompleted` between two requests seconds apart for
the same user**.

### apps/api/business/missions/repository.go (data layer)

`GetDailyMissionSnapshot` at
`apps/api/business/missions/repository.go:28-39`: simple read by
`(user_id, local_date)`. Both parameters are stable across the
two reads (see above), so both reads hit the same physical row.

`CreateDailyMissionSnapshot` at
`apps/api/business/missions/repository.go:125-165`: the
`ON CONFLICT (user_id, local_date) DO UPDATE` clause (lines 154-157)
updates only `timezone`, `review_target`, and `updated_at` — never
`reviews_completed`. The RETURNING clause (lines 158-161) returns
the row's values after the update; for a column not in the SET
clause, that is the existing (pre-conflict) value, per Postgres's
`ON CONFLICT DO UPDATE` semantics
(<https://www.postgresql.org/docs/current/sql-insert.html#SQL-ON-CONFLICT>).
So if the row exists and the ON CONFLICT fires, `reviews_completed`
is preserved. If the row does not exist, the INSERT proceeds and
`reviews_completed = 0` from the VALUES clause, with the
DEFAULT-zero value being the only one that can land in a fresh row
(via this path or via the equivalent
`MarkSnapshotMissed`/P3-path lazy creation). The READ path cannot
re-insert a row that already exists with a non-zero
`reviews_completed`, because the read returns non-nil first.

`IncrementReviewsCompleted` at
`apps/api/business/missions/repository.go:173-233`: the only SQL
statement in the entire `apps/api/` Go code that writes
`reviews_completed`. It does
`reviews_completed = LEAST(reviews_completed + 1, review_target)`
— strictly an increment, capped at the target. It cannot
decrement. Its only caller is the P2 review-submission path
(`apps/api/business/reviews/postgres.go:374`), which itself is
only reached when the test actually submits a review attempt
through `POST /api/v1/reviews/submissions`. With
`reviewedCards = 0` (the documented condition in issue #450 and
the T00 evidence's addendum), this path is not entered.

`MarkSnapshotCompleted` (line 382-406): writes `status` and
`completed_at` only. Does not touch `reviews_completed`.

`MarkSnapshotProtected` (line 411-434): writes `status`,
`grace_applied`, `grace_day_id`. Does not touch
`reviews_completed`.

`MarkSnapshotMissed` (line 439-465): `INSERT ... ON CONFLICT DO
UPDATE` that sets `status` to `'missed'` when the row is in
`'open'` state, never `reviews_completed`. Has **no application
callers** in `apps/api/` Go code (grep-confirmed: the only hits
are the function definition itself and one test that uses it as
a unit-level test fixture). It is therefore not called during
either read, the P1/P2/P3 paths, the seed, or the deploy. Its
existence is a latent capability for future streak-reconciliation
work, not an active contributor to issue #450.

`apps/api/business/accounts/postgres.go:463`:
`UPDATE daily_mission_snapshots SET user_id = $2, ...` — this
reassigns ownership on account anonymization, not
`reviews_completed`. The synthetic account is never anonymized
(the seed creates it with `status = 'active'` and no
production-equivalent path soft-deletes it).

### apps/api/business/gamification/{timezone,service,repository}.go

`apps/api/business/gamification/timezone.go:60-82` —
`ResolveSettings`: pure function, no I/O. For the synthetic
account with no `user_settings` row, returns
`{Timezone: "UTC", DailyReviewTarget: 20}` consistently.

`apps/api/business/gamification/timezone.go:88-95` — `LocalDate`:
pure function of `(now, timezone)`. Returns midnight UTC of the
local day in the given IANA timezone, deterministically.

`apps/api/business/gamification/service.go:30-44` —
`Service.GetSettings`: reads `user_settings` (returns nil for
the synthetic account) and calls `ResolveSettings`. No
write side-effects.

`apps/api/business/gamification/repository.go:132-150` —
`Repository.GetUserSettings`: simple read; returns nil for a
non-existent row.

No part of the gamification layer mutates `daily_mission_snapshots`
or `reviews_completed` on the read path. The streak
reconciliation it runs as part of the lazy-create branch
(`ReconcileAndAdvance`, `service.go:143-217`) writes to
`streak_states` and `grace_day_ledger` only.

## findings/VOC-053-write-path-audit

I additionally re-read the write paths the test exercises
between its step 1 baseline read (where `reviewedBefore` is
recorded) and its step 7 post-review read (where `reviewedAfter`
is recorded), specifically to verify no committed write path can
decrement `reviews_completed` to `0` between those two reads. The
test's between-read steps are: P1 word-add (step 4), an empty
review queue (step 5, no submissions), and P3 sentence feedback
(step 6). With `reviewedCards = 0`, no P2 review submission
fires.

- **P1 word-add (`apps/api/business/learning/postgres.go:88-134`,
  specifically the P4 reward wiring at lines 107-128):** calls
  `gamification.GrantPoint` for the `+2` add-word reward only
  (writes to `confidence_point_ledger`). The comment at
  `apps/api/business/learning/postgres.go:109-110` explicitly
  states "D03 keeps the optional new-word mission goal
  disabled, so we don't call IncrementWordsAdded." So P1 does
  not call `IncrementWordsAdded`, and even if it did, that
  method only writes to `daily_activity_summaries.words_added`
  and `daily_mission_snapshots.new_words_completed` (not
  `reviews_completed`) per
  `apps/api/business/missions/repository.go:239-276`. **No P1
  write touches `daily_mission_snapshots.reviews_completed`.**

- **P2 review submission (`apps/api/business/reviews/postgres.go:168-331`,
  P4 wiring at lines 341-440+):** calls
  `missions.EnsureTodaySnapshot` (preserves existing
  `reviews_completed` via the same ON CONFLICT analysis above)
  and then `missions.IncrementReviewsCompleted` (strictly
  increments, capped at target). Cannot decrement. **With
  `reviewedCards = 0`, this path is not entered during the
  failing test run.**

- **P3 sentence feedback (`apps/api/business/aifeedback/mission.go`
  seam, real `missions.MissionUpdater.Update` at
  `apps/api/business/missions/service.go:341-348` delegating to
  `UpdateForSentence` at `service.go:365-468`):** calls
  `EnsureTodaySnapshot` (preserves existing
  `reviews_completed`), `gamification.GrantPoint` twice
  (sentence-submitted, AI-feedback-received, both writing
  `confidence_point_ledger`), `missions.IncrementSentenceSubmitted`
  (writes `daily_activity_summaries.sentences_submitted` and
  optionally `daily_mission_snapshots.sentence_practices_completed`,
  not `reviews_completed`),
  `missions.IncrementAIFeedbackReceived` (writes
  `daily_activity_summaries.ai_feedback_received`),
  `missions.IncrementConfidencePointsEarned` (writes
  `daily_activity_summaries.confidence_points_earned`), and
  `gamification.ReconcileAndAdvance` (writes `streak_states`
  and `grace_day_ledger`). The structural
  `missionCompleted := false` (line 463) is the honest
  P3-invariant return value the test/issue text and the
  surrounding comment both confirm: "P3 never increments
  reviews_completed and never transitions the snapshot"
  (line 459-461). **No P3 write touches
  `daily_mission_snapshots.reviews_completed`.**

- **Seed (`apps/api/scripts/seed-synthetic-smoke-user.sql`):**
  the seed updates only the `users` table — it does not touch
  `daily_mission_snapshots` or any mission/activity row.
  Confirmed by reading the full file: it sets `display_name`,
  `status = 'active'`, `onboarding_status = 'completed'`,
  `email_verified_at`, `is_synthetic_test_account = true`, and
  `updated_at` on `users`, plus a soft-delete of any
  previously-seeded synthetic account under a *different*
  email address. The same-email account is preserved with its
  current `user_id`, so any `daily_mission_snapshots` rows
  keyed to that `user_id` survive the seed intact.

- **Session-mint endpoint
  (`apps/api/app/api/production.go:839-884`, calling
  `auth.Service.MintSyntheticSmokeTestSession` at
  `apps/api/business/auth/service.go:492-508`):** looks up the
  synthetic user by reserved email and creates a session row.
  Does not touch `daily_mission_snapshots`.

- **No `DELETE FROM daily_mission_snapshots` statement exists
  anywhere in `apps/api/`** (Go code, SQL files, or otherwise —
  grep-confirmed across the entire `apps/api/` subtree, including
  tests). The only `DELETE` statements in `apps/api/` are for
  `user_onboarding_profiles`, `user_settings`, `sessions`,
  `magic_links`, and `oauth_states` — none of which affect
  `daily_mission_snapshots` rows. The `DROP TABLE IF EXISTS
  daily_mission_snapshots;` line in
  `apps/api/migrations/20260725130001_voc030_p4_mission_tables.down.sql.example`
  is in a `.down.sql.example` file that is intentionally outside
  Atlas's forward-apply glob (`*.sql` per
  `apps/api/atlas.hcl:90`), so Atlas cannot apply it during a
  normal deploy. Per
  `apps/api/migrations/atlas_tooling_test.go::TestMigrationsDirectoryHasNoForwardDiscoveredDownFiles`,
  a guard fails-fast if any file ending in `.down.sql` (without
  the `.example` suffix) is ever committed — see also the
  lessons file entry "2026-07-29: Atlas's default forward-apply
  file glob is `*.sql`; the `.example` suffix is the
  load-bearing protection for recovery down-files".

- **No `TRUNCATE` statement exists anywhere in `apps/api/`.**

**Static reading conclusion: there is no committed code path in
`apps/api/` that can decrement `daily_mission_snapshots.reviews_completed`
for the synthetic account between the test's step 1 baseline
read and its step 7 post-review read.** The only places that
write `reviews_completed` are `IncrementReviewsCompleted` (which
increments) and `CreateDailyMissionSnapshot` (which inserts a
fresh row with `reviews_completed = 0` only when no row exists).
The only way the value can be `0` for a `(user_id, today)` row
that was observed non-zero earlier in the same run is if the row
was deleted by something outside committed code, or if the row
was never there in the first place and the second read hit the
lazy-create branch — neither of which can happen via any Go code
or migration in this repository.

## What this attempt did not (and could not) verify

The T00 evidence already established that the items below
require access this implementation environment does not have.
This T01 attempt did not (and could not) re-attempt them:

1. Direct `daily_mission_snapshots` row inspection for the
   synthetic account in real staging (`SELECT reviews_completed,
   local_date, status, updated_at FROM daily_mission_snapshots
   WHERE user_id = '<synthetic UUID>' ORDER BY local_date DESC
   LIMIT 3`) during and after a real test run. Only this would
   distinguish "the row is preserved with its prior
   `reviews_completed` value, but the read path is somehow
   returning a stale `0`" from "the row's `reviews_completed`
   has actually been decremented to 0 in the database" from
   "the row has been deleted, and the lazy-create path re-inserts
   it with `reviews_completed = 0`." The T00 evidence's static
   analysis rules out the first option (read-path returns stale
   0); the second and third remain candidates.

2. A fresh-account reproduction (T00 checklist item 3). The
   T00 evidence already records that a fresh-synthetic-account
   reproduction would resolve the residue-vs-not question for
   candidate (c) and the "fourth cause" out-of-band cache
   hypothesis. This requires the same staging access prior
   packages (VOC-050, VOC-052) have used.

3. An out-of-band cache / network check (T00 checklist item
   5) — `dig`/`host` of the staging origin, account-level
   Cloudflare cache rules, and similar. The T00 addendum's
   `/home` direct `curl -I` (recorded at the bottom of
   `t00-evidence.md`) showed `cf-cache-status: DYNAMIC` on
   the same unauthenticated routes, but the Cloudflare
   account-level cache configuration is not in the repository
   and this environment cannot inspect it.

## Decision: not implementing a fix in this attempt

Per `tasks.md`'s T01 status ("do not begin this task's fix
until T00's evidence names a specific, evidence-backed cause")
and the `implementer` role prompt's instruction to "say so
plainly instead of producing a partial or scope-expanded change"
when the task is missing a dependency the package needs, this
attempt produces **no code change**. The only artifact is this
evidence file.

The realistic next step that would unblock T01 is **item 1
above** (a live-staging DB read of the synthetic account's
`daily_mission_snapshots` row state during and after a real
test run, plus ideally a fresh-account reproduction per item
2). With those results, T01 can be re-attempted with an actual
confirmed cause — most likely (c) "test-data interaction" or
"fourth cause" out-of-band cache, given (a) is ruled out by
live evidence and (b) is ruled out by the static read-path
analysis re-confirmed above — and a narrowly-scoped fix can
be written against the confirmed cause.

## Risk and protected areas

No files were modified in this attempt. `git diff` shows only
this evidence file as an untracked addition. The path-based
floor for "an evidence file inside
`specs/changes/VOC-053-staging-core-loop-e2e-words-reviewed-today/`"
is `R0` per
`docs/governance/change-risk-classification.md`'s
"Documentation, comments, formatting, tests, or metadata with
no behavioral, policy, authority, security, or release effect"
class. `planned_implementation_risk_floor` in `change.yaml`
remains the package's draft `R3` proposal; per
`change.yaml`'s own "blocking_reasons," the actual floor is
"not yet computed by running
`scripts/governance/classify-change-risk.sh` against a real
task-scoped file list," which this attempt's no-code-change
file list would not change (no application-code files were
touched, so the `R3` application-code class floor that would
apply under either candidate fix still applies if/when the
fix is written).

No protected governance, workflow, or secret-handling area is
touched by this evidence file.

## Local deterministic checks run for this task

```bash
bash scripts/governance/validate-governance.sh
bash scripts/governance/classify-change-risk.sh
git diff --check
```

Commands executed in this implementation run (each in the
working directory, with the untracked evidence file as the
only working-tree change):

- `bash scripts/governance/validate-governance.sh` — pass.
  Output: "Repository foundation validation passed.
  Governance structure validation passed." The same output
  T00 reported, since this attempt's only working-tree
  addition is an evidence file inside the change package's
  own canonical path (which the validator explicitly allows).
- `bash scripts/governance/classify-change-risk.sh` — pass.
  Path-based floor is reported as `R0` for this attempt's
  untracked-only working tree, consistent with "an evidence
  file inside the change package's own directory." The
  package's `planned_implementation_risk_floor: R3` ceiling
  in `change.yaml` is unchanged; it remains a draft
  proposal pending the real, post-investigation file list
  that a future T01 fix attempt would produce, and may
  float higher if a future attempt needs a corrective
  migration (per `change.yaml`'s
  `migration_required: unknown-conditional-on-root-cause`
  and `specification.md`'s open question 3).
- `git diff --check` — pass. No diff against tracked files;
  the only working-tree change is the untracked
  `t01-evidence.md` file, which `git diff --check` does not
  inspect.

For `apps/api` Go code, this attempt did not modify any Go
file, so the broader `pnpm validate` (lint / typecheck /
test / build) suite documented in `docs/development.md` is
not triggered by this attempt. The next T01 attempt that
does modify `apps/api/` Go code must run the relevant subset
of `pnpm validate` (`pnpm typecheck` and
`pnpm --filter @vocanova/api test` at minimum, per
`docs/development.md`'s `apps/api` section) before claiming
that attempt complete.

## T01 evidence handoff

This evidence file is the deliverable for this T01 attempt.
It does not satisfy `VOC-053-AC-01`, `VOC-053-AC-02`,
`VOC-053-TEST-01`, or `VOC-053-TEST-02`; those acceptance
criteria require a fix to exist and to demonstrably hold,
neither of which is the case here. The package is
deliberately in the state T00's "T01 entry conditions"
section calls "T01 is not yet safe to start" — the honest
answer T00's "or not fully ruled out, if evidence is
inconclusive — record that honestly too" requirement
anticipates.

A follow-up issue or task-prerequisite to perform the live
verification items above (most importantly the
`daily_mission_snapshots` row inspection) is the natural
unblock path. Once those results exist, T01 can be
re-attempted with an actual confirmed cause and a fix
narrowly scoped to that cause, per `tasks.md`'s T01
description and `test-plan.md`'s `VOC-053-TEST-01` /
`VOC-053-TEST-02` procedures.
