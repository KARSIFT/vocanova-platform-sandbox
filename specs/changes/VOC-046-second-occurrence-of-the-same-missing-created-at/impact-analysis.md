# VOC-046 — Impact Analysis

## Security and privacy

No secrets are touched by this package. No personal-data field is added,
removed, or changed — `created_at`/`updated_at` are operational timestamps
already declared in each affected table's schema, not new data. No
authorization or authentication check is changed by this fix; it corrects
what values are written on insert, not who is allowed to write them. The
audit task (`VOC-046-T02`) reads and potentially edits application-code
`INSERT` statements across thirteen files but does not touch credential
handling, session/token logic, or access-control logic — if the audit
surfaces a defect that does implicate one of those areas, it should be
flagged as a follow-up rather than silently fixed under this package's
scope, per `specification.md`'s non-goals.

## Data and migrations

No migration is proposed by this package's primary fix, consistent with
this codebase's deliberate application-level timestamp convention (see
`specification.md`'s non-goals). The fix corrects previously-impossible
inserts (rows that could never be created due to the `NOT NULL` violation)
to now succeed with valid data — no existing row is modified, backfilled, or
migrated, since the defect blocked creation rather than corrupting existing
rows. Rollback is a pure code revert with no data cleanup required, since
rows created after the fix and before any rollback are correctly-formed, not
corrupted (same reasoning as VOC-045's `release-plan.md`). If the audit
(`VOC-046-T02`) surfaces a call site that, unlike the confirmed
`daily_mission_snapshots`/`daily_activity_summaries` cases, requires a
schema change to fix safely, that is explicitly out of this package's scope
and must be flagged as a follow-up rather than folded in, per
`specification.md`'s non-goals.

## Analytics and accessibility

Analytics: none identified. This fix does not add, remove, or change any
analytics event; it only allows previously-crashing inserts to succeed.
Accessibility: not applicable. This is a backend data-persistence fix with
no user-interface surface of its own. The user-visible symptom this fix
resolves (a generic "Something went wrong" error page after a `500`) is a
downstream effect of the underlying API failure, not a UI component this
package changes.

## Risks, dependencies, and evidence

- `VOC-046-R00`: the confirmed `daily_mission_snapshots` crash is a `P0`
  production outage blocking every user from reaching `/home` after
  onboarding — the same severity class as VOC-045's `user_settings` outage,
  and it directly blocks the rest of VOC-038-T03's core-loop validation.
  This is a draft risk proposal for the reviewing human, not a
  determination.
- `VOC-046-R01`: the systemic-audit task's scope (thirteen named files) is
  bounded, but its actual size in terms of call sites needing a fix is
  unknown until the audit runs — a materially larger-than-expected number of
  broken call sites could reasonably warrant re-classifying that task's risk
  before implementation, or splitting it into multiple smaller tasks/PRs at
  the reviewing human's or implementer's discretion, consistent with this
  repository's preference for small, focused, independently reviewable
  changes.
- `VOC-046-R02`: this is the second confirmed occurrence of the identical
  bug class in production within the same short window (VOC-045's
  `user_settings` occurrence and this issue's `daily_mission_snapshots`
  occurrence), which is itself evidence that the underlying convention
  (`NOT NULL`-no-`DEFAULT` timestamps requiring every call site to remember
  to supply them, with no automated check enforcing this) is a
  higher-than-routine correctness risk across this codebase — this is the
  motivating rationale for `specification.md`'s open question 2 (the
  schema-scanning detection check), not assumed to be resolved by this
  package alone.
- `VOC-046-DEP-00`: depends on issue #352's reported reproduction and root
  cause remaining accurate at implementation time.
- `VOC-046-DEP-01`: depends on `VOC-046-T02`'s audit actually enumerating
  the full scope before any task claims completeness.
- `VOC-046-DEP-02`: depends on the reviewing human's decision on
  `specification.md`'s open question 2 (schema-scanning check in scope or
  not) before `VOC-046-T03` can be dispatched, if adopted at all.
- `VOC-046-EV-00` through `VOC-046-EV-05`: the evidence artifacts named in
  `acceptance-criteria.md`, to be produced during implementation and
  independent verification, not asserted here.
