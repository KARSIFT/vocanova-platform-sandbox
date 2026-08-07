# VOC-046 — Second Occurrence of the Missing created_at/updated_at Bug Class: daily_mission_snapshots INSERT Crashes GET /api/v1/daily-mission for Every New User, Plus a Repository-Wide Audit

**Status: proposed, not adopted.** Nothing in this package is
implementation-authorized. It is a draft response to
[issue #352](https://github.com/KARSIFT/vocanova-platform-sandbox/issues/352),
prepared for founder/steward review at adoption time.

## Why this exists

`daily_mission_snapshots`'s schema
(`apps/api/migrations/20260725130001_voc030_p4_mission_tables.sql`) declares
`created_at` and `updated_at` as `NOT NULL` with no `DEFAULT` — the same
convention as `user_settings`, whose own occurrence of this exact bug class
was fixed by [VOC-045](../VOC-045-user-settings-upsert-omits-created-at-updated-at/)
(issue #341, PRs #342/#347/#348/#349). `apps/api/business/missions/repository.go`'s
`CreateDailyMissionSnapshot` — called by `apps/api/business/missions/service.go`'s
`GetDailyMissionView` on first read of a new local day, i.e. every brand-new
user — omits both columns from its `INSERT` column list, setting `updated_at`
only inside the `ON CONFLICT (user_id, local_date) DO UPDATE ... SET` clause,
which never runs on a genuine first insert.

The issue reports this reproduced live in production during VOC-038-T03's
core-loop validation, immediately after VOC-045's fix landed: a real user who
just completed onboarding lands on `/home` and gets a generic
"Something went wrong / We couldn't load your home data" error, because
`GET /api/v1/daily-mission` returns a `500`. Unlike VOC-045's `user_settings`
occurrence — whose underlying Postgres error text happened to surface via a
`400` path — this one is masked by the API's generic `500` handling, so the
real cause needs verification against a real Postgres `NOT NULL` check at
implementation time, not just the issue's textual diagnosis.

**Every real user who reaches `/home` after completing onboarding is
currently blocked** — this is a `P0` that blocks the rest of VOC-038-T03's
core-loop validation entirely (discover/save, review, missions/progress all
sit behind `/home`).

## Why this is not just a one-table patch

The issue's own investigation confirms this `NOT NULL`-with-no-`DEFAULT`
convention repeats identically across all 13 files in
`apps/api/migrations/`, and that it is a deliberate codebase convention
(explicit application-level timestamp control via the injected
`clock.Clock` abstraction, rather than DB-level `now()` defaults — presumably
for deterministic testability). A DB-level `DEFAULT now()` fix would work but
would deviate from that established convention and is explicitly **not**
proposed here as the primary fix. Instead, following the issue's own
suggested fix, this package scopes both:

1. The immediate, confirmed crash (`CreateDailyMissionSnapshot`'s `INSERT`),
   fixed the same way VOC-045-T00/T01 fixed `user_settings`.
2. A repository-wide audit of every file the issue names as containing
   `INSERT INTO` statements against application tables, to find and fix any
   other call site sharing this same omission, plus regression coverage and
   an optional schema-scanning detection check so this bug class cannot
   recur a third time undetected.

## What this package deliberately does NOT do

- It does not propose a schema migration (e.g. adding `DEFAULT now()` to any
  table) as its fix, for the same reason VOC-045 didn't: it would deviate
  from this codebase's deliberate application-level timestamp convention.
  Flagged as an open question for the reviewing human, not assumed.
- It does not enumerate every individual audit fix as its own task up front.
  The exact set of additional call sites needing a fix is not knowable until
  `VOC-046-T02`'s audit actually runs against the thirteen named files (see
  `specification.md`'s open question 1) — this package structures the audit
  and its fixes as one bounded, evidence-producing task rather than guessing
  a fix list in advance.
- It does not change `apps/api`'s generic `500` error handling that masked
  this issue's underlying Postgres error (unlike `user_settings`'s `400`
  path). That is a separate, narrower observability concern the issue notes
  in passing, not its reported root cause or requested fix — flagged as an
  open question rather than silently folded into scope.
- It does not adopt itself. `change.yaml` leaves every adoption/authorization
  field at its template default. No task in `tasks.md` may be dispatched
  until a real adoption decision is recorded.

## Open questions flagged for the reviewing human

`specification.md`'s "Open questions" section flags: (1) the exact boundary
of `VOC-046-T02`'s audit — how to treat a file from the issue's list that
turns out to have no actual defect, or a defect the audit discovers outside
the named list; (2) whether the issue's suggested schema-level/lint-style
detection check (walking each migration's `NOT NULL`-no-`DEFAULT` columns
against discovered `INSERT` statements) is in scope for this package or a
separate follow-up; and (3) whether `apps/api`'s generic `500` masking of the
underlying Postgres error is in scope for this package or a separate,
narrower observability follow-up.

## Structure

Mirrors recent packages' convention (e.g. VOC-045, VOC-044, VOC-043):
`specification.md`, `acceptance-criteria.md`, `impact-analysis.md`,
`implementation-plan.md`, `tasks.md`, `test-plan.md`, `release-plan.md`.

## Recommended next action for the reviewing human

1. Confirm the proposed `R3` risk classification in `change.yaml` (a
   confirmed hard outage of the core post-onboarding path in production,
   plus a systemic audit of unknown-until-run size across many files, none
   of which are known at drafting time to require a schema migration).
2. Decide open questions 2 and 3 (schema-scanning check in scope; generic
   `500` masking in scope) before or during adoption.
3. Adopt (or request changes to) this package, then dispatch `VOC-046-T00`
   through `VOC-046-T03` individually, as prior packages' tasks were
   dispatched one at a time — `VOC-046-T00` (the confirmed crash) should be
   prioritized first given its `P0` production impact.
