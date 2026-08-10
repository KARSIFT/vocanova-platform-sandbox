# VOC-053 — Tasks

## VOC-053-T00 — Investigate and confirm the real root cause with direct evidence

- Requirement source: issue #450's "candidate root causes ... not
  prescriptive" framing; `specification.md` open questions 1 and 2
- Acceptance criteria: `VOC-053-AC-00`
- Tests: `VOC-053-TEST-00`
- Evidence: `VOC-053-EV-00`
- Status: **complete** — investigation objective satisfied per three independent
  investigation passes (VOC-053-T00 attempts 1 and 2 plus issue #473's third
  pass with live staging/production access). Third-pass evidence:
  https://github.com/KARSIFT/vocanova-platform-sandbox/issues/450#issuecomment-5238054774.
  Closure rationale and supersession of the fix path: issue
  [#473](https://github.com/KARSIFT/vocanova-platform-sandbox/issues/473).

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
- Status: **cancelled** — superseded by
  [VOC-063](https://github.com/KARSIFT/vocanova-platform-sandbox/tree/develop/specs/changes/VOC-063-voc-053-investigation-exhausted-3-independent).
  All named root-cause candidates from issue #450 were exhausted by three
  independent investigation passes with direct live evidence; no
  evidence-backed production fix was identified. See issue
  [#473](https://github.com/KARSIFT/vocanova-platform-sandbox/issues/473).

This task's original fix scope (production code changes scoped to a confirmed
root cause) is preserved in git history but will not be implemented under
VOC-053. Forward path: `VOC-063-T01` (staging E2E step-7 retry hardening).

## VOC-053-T02 — Verify the fix on real staging, including under the exact
failure condition

- Requirement source: issue #450's reported failure; `specification.md`
  scope item 3
- Acceptance criteria: `VOC-053-AC-03`, `VOC-053-AC-04`
- Tests: `VOC-053-TEST-03`
- Evidence: `VOC-053-EV-02`
- Status: **cancelled** — superseded by
  [VOC-063](https://github.com/KARSIFT/vocanova-platform-sandbox/tree/develop/specs/changes/VOC-063-voc-053-investigation-exhausted-3-independent).
  Depends on `VOC-053-T01`, which was cancelled when the investigation found no
  evidence-backed production defect to verify. See issue
  [#473](https://github.com/KARSIFT/vocanova-platform-sandbox/issues/473).

This task's original verification scope is preserved in git history but will
not be executed under VOC-053. Forward path: `VOC-063-T02` (real staging
verification of hardened step 7).
