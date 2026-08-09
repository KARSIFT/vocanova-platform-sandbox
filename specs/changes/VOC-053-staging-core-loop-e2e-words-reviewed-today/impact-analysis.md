# VOC-053 — Impact Analysis

## Security and privacy

No new secret, credential, or personal-data handling is introduced by any
candidate fix under consideration. The synthetic account whose counter this
issue concerns is a non-personal, test-only account (per VOC-050's own
package), not a real user's data. If the confirmed root cause is a real
backend day-boundary/timezone bug, the fix would also correct behavior for
real users' own `reviews_completed` counts, which is exactly the intended
effect (an accurate, non-decreasing daily counter) rather than a new privacy
concern — no additional field, log, or exposure is added.

## Data and migrations

Unknown until `VOC-053-T00` confirms the root cause (see
`specification.md`'s open questions):

- If candidate (a) caching or (c) test-data interaction: no migration
  expected. The fix is code-only (a fetch-caching option, or a concurrency
  fix), touching no schema and no existing row.
- If candidate (b) a real backend computation bug: whether a corrective
  migration is needed depends on whether the bug already wrote incorrect
  `local_date` values into real (non-synthetic) users' `daily_mission_snapshots`
  or `daily_activity_summaries` rows, versus only affecting the read path
  without having corrupted any stored row. `apps/api/migrations/` is a more
  sensitive protected path than the application code fix itself per
  `docs/governance/change-risk-classification.md`; this package does not
  assume a migration is needed and requires the implementer to record an
  explicit yes/no determination with evidence, not silently add or silently
  omit one (`VOC-053-R02` below).

## Analytics and accessibility

No new analytics event, metric, or user-facing behavior is introduced by any
candidate fix — at most, an existing computation or caching bug is corrected
so the already-designed "words reviewed today" counter behaves as originally
intended. No accessibility change is anticipated; if the eventual fix happens
to change `apps/web/src/app/(app)/home/page.tsx`'s rendered markup for an
unrelated structural reason, the existing accessibility suite for that page
must be re-run, but this is not expected scope for this package.

## Risks, dependencies, and evidence

- `VOC-053-R00`: If the confirmed root cause turns out to be Next.js caching
  and the fix is an overly broad `cache: "no-store"` applied beyond the
  `getDailyMission()` read path, that could unnecessarily increase backend
  load on every `/home` render for unrelated fetches on the same page
  (`listSavedWords`, `listDueWords`). The fix must be scoped to the confirmed
  caching layer specifically, not applied blanket across `createServerApiClient`,
  unless the implementer confirms the same caching bug actually affects those
  other calls too.
- `VOC-053-R01`: If the confirmed root cause is a backend timezone/local-date
  resolution bug, a narrow fix to the wrong layer (e.g. patching
  `gamification.LocalDate`'s already-correct-as-read pure function instead of
  the actual per-request resolution caller this drafting pass could not
  locate — see `specification.md` finding 5 / open question 2) would leave
  the real bug in place while appearing to fix the symptom. `VOC-053-T00`'s
  evidence must show the traced call path, not just the symptom, before
  `VOC-053-T01` begins.
- `VOC-053-R02`: A real backend bug that already wrote incorrect
  `local_date`-keyed rows for real users would need an explicit,
  human-reviewed decision on whether a corrective migration is in scope for
  this package or should be a narrower, separately-adopted follow-up, given
  migrations are a more sensitive protected path — see
  `specification.md` open question 3 and `tasks.md`'s `VOC-053-T01` framing.
- `VOC-053-DEP-00`: Unresolved at drafting time — which of issue #450's
  candidates (or a fourth cause) is the real root cause. See
  `specification.md` open question 1.
- `VOC-053-DEP-01`: Unresolved at drafting time — where and how `now`/timezone
  is resolved per request for the daily-mission read path; this drafting
  pass located the pure `gamification.LocalDate` function and the
  `missions.Repository` read/write SQL but did not locate the specific
  `apps/api/app/api/missions.go` (or equivalent) HTTP handler that calls
  them for a live request. See `specification.md` open question 2.
- `VOC-053-EV-00`: Required evidence — the real staging HTTP response
  inspection (headers/cache-status) and/or the traced backend request-handling
  code path, with the confirmed root cause and the specific evidence that
  confirmed it (and why the other candidates were ruled out, or why they
  remain inconclusive if evidence does not fully rule them out).
- `VOC-053-EV-01`: Required evidence — the fix's diff, scoped to the
  confirmed cause, plus direct verification that two reads of
  `reviewsCompleted` for the same user/day seconds apart with no intervening
  review return the same value (not merely that the E2E spec happens to pass
  once).
- `VOC-053-EV-02`: Required evidence — a passing real staging
  `tests/staging-e2e/core-loop.staging.spec.ts` run specifically under the
  `reviewedBefore >= 1`-from-prior-run condition, with the workflow run log
  and the observed `reviewedBefore`/`reviewedCards`/`reviewedAfter` values
  recorded.
