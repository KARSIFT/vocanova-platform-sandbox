# VOC-053 — Tasks

## VOC-053-T00 — Investigate and confirm the real root cause with direct evidence

- Requirement source: issue #450's "candidate root causes ... not
  prescriptive" framing; `specification.md` open questions 1 and 2
- Acceptance criteria: `VOC-053-AC-00`
- Tests: `VOC-053-TEST-00`
- Evidence: `VOC-053-EV-00`
- Status: pending

No fix is written in this task. Its job is to determine, with direct
evidence, which of issue #450's three candidates (Next.js fetch/route caching;
a real backend day-boundary/timezone bug in `reviews_completed`/`local_date`
computation; a synthetic-account test-data interaction) — or a fourth cause
this drafting pass did not anticipate — actually explains the 2026-08-09
failure. Concretely:

- Inspect the real HTTP response for `GET /home` (and, if reachable directly,
  the underlying `GET /api/v1/daily-mission`-equivalent call) against staging,
  specifically `cache-control`, any Next.js cache-status header (e.g.
  `x-nextjs-cache`), and whether two requests seconds apart return different
  bodies for the same backend state.
- Trace the backend request-handling path in `apps/api/app/api/missions.go`
  (or wherever the daily-mission `GET` handler lives) end to end: confirm
  where and how it resolves the per-request `now` and per-user timezone
  before calling `gamification.LocalDate` and
  `missions.Repository.GetDailyMissionSnapshot`/`GetDailyActivitySummary`,
  and whether that resolution could plausibly differ between two requests
  seconds apart under real staging conditions.
- If neither (a) nor (b) reproduces or explains the decrease, investigate
  whether the persistent synthetic account's own state (e.g. a concurrent
  deploy run touching the same account, per the issue's note that "many
  deploy runs" happened the same night) is a contributing factor, per
  candidate (c) — first by trying to reproduce against a fresh
  synthetic account/day, as the issue itself suggests, before concluding
  candidate (c) alone explains it.
- Record the confirmed cause, the exact evidence that confirmed it, and why
  the other candidates were ruled out (or not fully ruled out, if evidence
  is inconclusive — record that honestly too) in this task's evidence file.

## VOC-053-T01 — Fix the confirmed root cause

- Requirement source: `specification.md` scope item 2
- Acceptance criteria: `VOC-053-AC-01`, `VOC-053-AC-02`
- Tests: `VOC-053-TEST-01`, `VOC-053-TEST-02`
- Evidence: `VOC-053-EV-01`
- Status: pending — depends on `VOC-053-T00` producing a confirmed root cause;
  do not begin this task's fix until T00's evidence names a specific,
  evidence-backed cause

Implement a fix scoped narrowly to `VOC-053-T00`'s confirmed cause:

- If caching is confirmed: make the relevant fetch(es) in
  `apps/web/src/lib/api-server.ts` and/or the `getDailyMission()` call site in
  `apps/web/src/app/(app)/home/page.tsx` explicitly bypass Next.js's Data
  Cache (e.g. `cache: "no-store"`, or the minimal explicit option that matches
  the confirmed caching layer), rather than a broader, unrequested caching
  policy change across `apps/web`.
- If a backend computation bug is confirmed: fix the specific
  timezone/local-date resolution defect in the traced request-handling path
  (see `VOC-053-T00`), without changing `gamification.LocalDate`'s or
  `missions.Repository`'s already-correct-as-read SQL/logic unless the bug is
  actually located there once fully traced. If a corrective migration for
  already-written incorrect rows is needed (per `specification.md` open
  question 3), record that as a distinct, explicitly flagged decision in this
  task's evidence rather than silently including or silently omitting it.
- If a test-data interaction is confirmed: fix the interaction (e.g. a
  concurrency-safety gap in how the persistent synthetic account's state is
  read/written across simultaneous or near-simultaneous deploy runs), not the
  reuse design itself.
- Do not modify `tests/staging-e2e/core-loop.staging.spec.ts`'s step 7
  assertion. If the fix requires any change to the spec file at all (e.g. a
  helper it also uses), record why and confirm the change does not weaken
  what step 7 checks.

## VOC-053-T02 — Verify the fix on real staging, including under the exact
failure condition

- Requirement source: issue #450's reported failure; `specification.md`
  scope item 3
- Acceptance criteria: `VOC-053-AC-03`, `VOC-053-AC-04`
- Tests: `VOC-053-TEST-03`
- Evidence: `VOC-053-EV-02`
- Status: pending — depends on `VOC-053-T01` landing and a real staging
  deploy running with it in place

No further source change is expected in this task (beyond fixing any gap this
verification itself surfaces, narrowly, per `VOC-052-T01`'s own precedent for
this kind of task). After `VOC-053-T01`'s fix is live in a real
`deploy-staging.yml` run, confirm
`tests/staging-e2e/core-loop.staging.spec.ts` passes step 7 reliably —
specifically including at least one run where the synthetic account already
has `reviewedBefore >= 1` from a prior run's residue (the same condition the
2026-08-09 failure occurred under), not only a freshly-reset state that would
not actually exercise the failure condition. Record the passing run's log and
the specific `reviewedBefore`/`reviewedCards`/`reviewedAfter` values observed
as evidence.
