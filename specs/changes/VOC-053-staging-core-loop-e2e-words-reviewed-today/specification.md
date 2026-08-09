# VOC-053 — "Words Reviewed Today" Reads 1, Then 0, Moments Later: Specification

## Objective and requirement source

Close the gap reported in
[GitHub issue #450](https://github.com/KARSIFT/vocanova-platform-sandbox/issues/450):
on 2026-08-09, the first staging core-loop E2E run to get past the
review-queue race fixes (PRs #446/#448/#449) failed at
`tests/staging-e2e/core-loop.staging.spec.ts:375`, step 7
("progress reflects the completed reviews"):

```
Error: expect(received).toBeGreaterThanOrEqual(expected)
Expected: >= 1
Received:    0

  373 |       await page.goto("/home");
  374 |       const reviewedAfter = await readReviewedTodayCount(page);
> 375 |       expect(reviewedAfter).toBeGreaterThanOrEqual(reviewedBefore + reviewedCards);
```

(run [31332238452](https://github.com/KARSIFT/vocanova-platform-sandbox/actions/runs/31332238452))

Per the issue's own timing analysis: the whole test ran in 9.8s, consistent
with `reviewedCards` (step 5's return value) being `0` — the synthetic
account's review queue was already empty at the start of this run (expected;
anticipated by PR #420's own review comment). `Expected: >= 1` therefore means
`reviewedBefore = 1` (recorded earlier the same day, by an earlier deploy run
against the same persistent synthetic account per this spec's own documented
"Rerun safety" design). `reviewedAfter`, read via a fresh `page.goto("/home")`
roughly ten seconds later in the *same* test run, returned `0` — a real
decrease on what should be a same-day, monotonically non-decreasing counter.
No midnight/day boundary was crossed (the run happened at ~19:42 UTC).

The objective is: after this package's implementation, this step passes
reliably against real staging because the counter genuinely does not decrease
within the same day — not because the assertion, the synthetic-account reuse
design, or the caching behavior was weakened or routed around.

## Confirmed findings (independently re-checked during drafting)

- `apps/web/src/app/(app)/home/page.tsx` is a Server Component. It calls
  `client.getDailyMission()` on every render (no explicit caching option in
  the component itself) and renders `dailyMissionResponse.data.reviewsCompleted`
  as the "X of Y words reviewed today" text `readReviewedTodayCount` parses
  (`REVIEWED_TODAY_PATTERN` in the spec file) — confirmed by reading the file
  directly.
- `apps/web/src/lib/api-server.ts`'s `createServerApiClient` calls
  `await headers()` (Next.js's dynamic API for reading the incoming request's
  headers) on every invocation, before constructing the API client. Per
  Next.js's App Router caching model, calling a dynamic API inside a route's
  render path opts that route out of the Full Route Cache and forces dynamic
  (per-request) rendering — this narrows, but does not eliminate, issue #450's
  first candidate (stale cached `/home` HTML). It does **not** by itself
  control whether the *underlying* `fetch(input, { ...init, headers })` call
  inside `createServerApiClient`'s custom `fetch` implementation is subject to
  Next.js's separate Data Cache — that call passes no explicit `cache` option,
  so its caching behavior follows Next.js 16.3.0's actual default for `fetch`
  calls made from a dynamic route, which this drafting pass did not verify
  against a live response (see Open Question 1 — the issue explicitly says
  this needs real HTTP header / cache-status inspection this environment does
  not have).
- `apps/api/business/missions/repository.go`'s `IncrementReviewsCompleted`
  updates `reviews_completed` via a single `UPDATE ... SET reviews_completed =
  LEAST(reviews_completed + 1, review_target) ... WHERE user_id = $1 AND
  local_date = $2` inside the caller's transaction, and
  `GetDailyMissionSnapshot` / the read path select the same column keyed by
  `(user_id, local_date)`. Read directly: there is no code path in this file
  that decrements `reviews_completed`, and no code path that reads a
  differently-computed `local_date` between the write and a subsequent read —
  both use whatever `localDate` value the caller passes in. This drafting pass
  found no decrement or dual-computation bug in this file by static reading
  alone.
- `apps/api/business/gamification/timezone.go`'s `LocalDate` computes the
  calendar date by loading the resolved IANA timezone, converting `now` into
  it, and truncating to a UTC-midnight-anchored `time.Time` for that local
  calendar day — a pure, deterministic function of `(now, timezone)` with no
  I/O. A same-run flip in "which day counts" would require either (a) the
  resolved timezone itself changing between the two requests (this drafting
  pass did not find where `missions`'s HTTP-handling caller resolves and
  passes `now`/timezone per request — that caller was not located during this
  reading pass and is an open question for the implementer, not this package)
  or (b) a request landing within a few hundred milliseconds of an actual
  midnight boundary in the resolved timezone — which issue #450 already rules
  out for this specific failure (~19:42 UTC, no synthetic-account timezone
  documented as being anywhere near a midnight boundary at that instant).
- This package's drafting pass did **not** locate or inspect the HTTP
  handler/caller in `apps/api/app/api/missions.go` that resolves `now` and the
  per-user timezone for a live request and calls `GetDailyMissionSnapshot` —
  only the repository and pure-timezone-math layers above. Whether that
  caller re-resolves the timezone (and therefore the local date) fresh on
  every single request in a way that could plausibly differ between two
  requests ~10s apart is unconfirmed and is Open Question 2 below.

## Scope and non-goals

In scope:

- Determine, with real evidence (live staging HTTP response inspection and/or
  targeted backend investigation), which of issue #450's three non-prescriptive
  candidates — or a fourth cause this drafting pass did not anticipate — is
  the actual root cause of the observed same-run decrease (T00).
- Fix the confirmed root cause, scoped narrowly to that cause, without
  weakening `tests/staging-e2e/core-loop.staging.spec.ts`'s existing step 7
  assertion (T01).
- Produce and record real, reproduced-on-staging evidence that step 7 passes
  reliably afterward, including at least one run where `reviewedBefore >= 1`
  from a prior run's residue, matching the exact condition the 2026-08-09
  failure occurred under (T02).

Non-goals / explicitly excluded:

- Not re-litigating the three already-fixed review-queue timing races (PRs
  #446, #448, #449) — those are closed, different failures. This is a new,
  different failure only reachable after those were fixed.
- Not a claim that any single one of issue #450's three candidates is
  definitely the cause; this package explicitly does not resolve that in
  advance of real investigation (see "Confirmed findings" and Open Questions
  below).
- No weakening, retrying, polling-until-pass, or removal of the step 7
  assertion in `tests/staging-e2e/core-loop.staging.spec.ts`. A fix that makes
  the test pass without the counter actually being correct is not this
  package's objective.
- No change to the synthetic-account reuse design (`tests/staging-e2e/
  core-loop.staging.spec.ts`'s documented "Rerun safety" convention) unless
  the confirmed root cause is specifically in that interaction (candidate 3) —
  and even then, only the interaction, not the reuse design itself, which is
  intentional per PR #420.
- Not a general audit of Next.js caching behavior across every route in
  `apps/web` — scoped to `/home`'s `getDailyMission()` read path specifically,
  unless the implementer's investigation finds the same class of bug affects
  another route the daily-mission read shares code with.

## Risk and protected areas

The confirmed root cause is unknown at drafting time, so the affected-area
list is a draft superset, not a determination — see `impact-analysis.md` for
the full reasoning per candidate. Candidate 1 (Next.js caching) would most
likely touch `apps/web/src/lib/api-server.ts` and/or
`apps/web/src/app/(app)/home/page.tsx`, both under `apps/web/src/`, which
`docs/governance/change-risk-classification.md`'s path-based floor classifies
per that document's `apps/web`/`apps/api` application-code class. Candidate 2
(a real backend bug) would touch `apps/api/business/missions/` and/or
`apps/api/business/gamification/`, and — if it requires a data-correcting
migration for already-written incorrect rows — would additionally touch
`apps/api/migrations/`, a more sensitive protected path per that same
document. This package proposes `R3` for the change as a whole (see
`change.yaml`'s `planned_implementation_risk_floor`) as a draft ceiling
covering both candidates' most likely footprint, pending the actual floor
being confirmed by running `scripts/governance/classify-change-risk.sh`
against T01's real, post-investigation file list — which may float higher if
a migration turns out to be needed for candidate 2, a case this drafting pass
flags explicitly rather than assumes away.

No protected governance, workflow, or secret-handling area is implicated by
any of the three candidates as currently understood.

## Decisions, contradictions, security, and privacy

No `VOC-053-D00`-numbered decision is recorded here because this package is
not yet adopted; per the template, decisions are only defined after approval.
The following are recorded as open questions for the reviewing human and/or
the eventual implementer, not resolved by this drafting pass:

1. **Which candidate is the real root cause (`VOC-053-DEP-00`).** This
   package cannot determine, from static code reading alone, whether the
   observed decrease is caused by (a) Next.js Data Cache / Full Route Cache
   staleness on `/home`'s `getDailyMission()` fetch or route render (narrowed
   but not eliminated by the "Confirmed findings" section's `headers()`
   observation — the underlying `fetch()` call's own Data Cache behavior is
   still unverified against a live response), (b) a real backend
   day-boundary/timezone bug in how `reviews_completed`/`local_date` is
   computed (this drafting pass found no bug in the pure `LocalDate` function
   or the repository's increment/read SQL, but did not locate or inspect the
   HTTP-handler caller that resolves `now`/timezone per request — see finding
   5 above), or (c) a test-data artifact specific to the persistent synthetic
   account being reused across many deploy runs the same night, interacting
   with (a) or (b). The implementer must actually inspect real staging HTTP
   response headers (`cache-control`, `x-nextjs-cache`, or equivalent) and/or
   the backend's request-handling code path for `daily_mission`, and record
   which candidate is confirmed, with the specific evidence that confirmed it,
   before implementing a fix targeted at it. Per issue #450, this package
   explicitly does not guess past this.
2. **Where and how `now`/timezone is resolved per request for the daily
   mission read (`VOC-053-DEP-01`).** This drafting pass located the pure
   `LocalDate` function and the repository's parameterized read/write SQL,
   but did not locate the specific HTTP handler in `apps/api/app/api/
   missions.go` (or elsewhere) that calls them for a live `GET` request, so
   it could not confirm whether that caller re-resolves the timezone/`now`
   fresh on every request in a way that could plausibly differ between two
   requests seconds apart, or caches/memoizes it in a way that could itself
   be a bug. The implementer must trace this path as part of resolving Open
   Question 1's candidate (b).
3. **Whether a data-correcting migration is needed.** If the confirmed root
   cause is candidate (b) and it already wrote an incorrect `local_date` for
   real historical rows (as opposed to only affecting the read path), the
   implementer must flag whether a corrective migration is in scope for this
   package or should be deferred to a narrower follow-up, given migrations
   are a more sensitive protected path than the code fix itself (see
   `impact-analysis.md`'s "Data and migrations" section).

No new secret, credential, or personal-data handling is introduced by any
candidate fix under consideration. The synthetic account's own review counts
are not personal data (see VOC-050's own package for the synthetic account's
non-personal, test-only status).

## Data, migrations, analytics, and accessibility

- **Data / migrations:** Unknown until Open Question 1 resolves. If the root
  cause is caching (candidate a) or a test-data interaction (candidate c), no
  migration is expected — the fix is code-only. If the root cause is a real
  backend computation bug (candidate b), a migration may be needed to correct
  already-written `daily_mission_snapshots`/`daily_activity_summaries` rows
  for real (non-synthetic) users if any were affected by the same bug in
  production-equivalent code paths — this package does not assume either way
  and flags it as Open Question 3 above.
- **Analytics:** None expected to be applicable — no new user-facing
  behavior, event, or metric is introduced by any candidate fix under
  consideration; at most, an existing computation or caching bug is
  corrected.
- **Accessibility:** None applicable. No UI structure change is anticipated
  by any candidate fix; if the fix path does end up changing
  `apps/web/src/app/(app)/home/page.tsx`'s rendered markup for an unrelated
  reason, the implementer must re-run the existing accessibility suite
  against that page per standard practice, but this is not anticipated scope.
