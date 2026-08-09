# VOC-053-EV-00 — T00 root-cause investigation evidence

Evidence for `VOC-053-T00` (`VOC-053-AC-00`, `VOC-053-TEST-00`). This is
attempt 2 of T00. The previous attempt (commit 74fd157, on the
`agent/voc-053-voc-053-t00` branch's prior history) reached the same
overall conclusion; this re-investigation independently re-read every
file the conclusion depends on, confirmed the dependency remains
correct on `develop` as of this attempt's base (1811026), and recorded
any new observations below. Where the static-analysis findings here
match the previous attempt's, that is independent re-confirmation, not
a copy.

## Scope of this evidence file

`VOC-053-T00` is the investigation task only: determine, with direct
evidence, which of issue #450's three non-prescriptive candidates
(Next.js fetch/route caching, a real backend day-boundary/timezone bug
in `reviews_completed`/`local_date` computation, a synthetic-account
test-data interaction) — or a fourth cause this drafting pass did not
anticipate — actually explains the 2026-08-09 failure of
`tests/staging-e2e/core-loop.staging.spec.ts` step 7
(`reviewedAfter >= reviewedBefore + reviewedCards`, run 31332238452,
19:42 UTC, `Expected: >= 1, Received: 0`).

The package's own drafting pass already established, by static reading
of the same files I re-read here, that no obvious code-level bug was
locatable in the four files issue #450 names
(`apps/web/src/app/(app)/home/page.tsx`,
`apps/web/src/lib/api-server.ts`,
`apps/api/business/missions/repository.go`,
`apps/api/business/gamification/timezone.go`). The drafting pass also
flagged that the live staging HTTP response / cache-status inspection
the issue's "candidate root causes ... not prescriptive" framing
requires **cannot be done from this implementation environment** —
issue #450 says so itself, and this investigation confirms the same
constraint.

This file records, in order:

1. What the implementer environment could and could not verify directly.
2. The specific code paths I re-read and the findings from each
   (independently of the drafting pass and the previous attempt).
3. The Next.js 16.3.0 caching model as it actually applies to this
   repository's configuration, since candidate (a) lives or dies there.
4. The specific handler, service, and repository code paths that resolve
   `now` and `timezone` per request (resolves
   `VOC-053-DEP-01`).
5. The candidate-by-candidate ruling, honest about what is and is not
   fully ruled out by static reading alone.
6. What must be verified live in staging to move from "inconclusive"
   to a confirmed root cause, so `VOC-053-T01` does not have to
   re-derive it.

## VOC-053-DEP-00 status (issue #450's three candidates)

**Inconclusive from this environment.** No single candidate is fully
confirmed. The analysis below rules candidate (a) out by static
reading of this repository's Next.js 16.3.0 configuration, rules
candidate (b) out by full trace of the handler → service →
repository call path (no code path can produce a different
`local_date` or a decremented `reviews_completed` between two
requests seconds apart at 19:42 UTC for a synthetic account with no
`user_settings` row), and leaves candidate (c) and a "fourth cause"
(out-of-band caching) as the most plausible remaining explanations
pending live staging verification. The exact live checks needed to
resolve this are listed under
`findings/VOC-053-required-live-verification` below.

## VOC-053-DEP-01 status (live request handler's per-request `now`/timezone resolution)

**Resolved by static reading.** The drafting pass could not locate
the HTTP handler that calls `gamification.LocalDate` and
`missions.Repository.GetDailyMissionSnapshot` for a live `GET
/api/v1/daily-mission` request, and left this as an open question
for the implementer. I located it and traced it; the trace is
recorded under `findings/VOC-053-DEP-01` below. Nothing in that
traced path can produce a plausible per-request drift in
`local_date` or `reviewsCompleted` between two requests seconds
apart at 19:42 UTC under the documented conditions (no midnight
boundary, synthetic account has no stored `user_settings` row, so
D01 falls through to UTC default).

## What the implementer environment could verify

- Static code reading of every Go file that participates in the
  daily-mission read path: `apps/api/app/api/missions.go` (handler
  + DTOs), `apps/api/business/missions/{service.go, repository.go,
  types.go}`, and `apps/api/business/gamification/{service.go,
  timezone.go, repository.go}`.
- Static code reading of `apps/web/src/app/(app)/home/page.tsx`,
  `apps/web/src/lib/api-server.ts`, `apps/web/src/middleware.ts`,
  `apps/web/next.config.ts`, and the relevant
  `packages/api-client/src/index.ts` methods.
- Next.js 16.3.0 caching model as documented at
  <https://nextjs.org/docs/app/api-reference/functions/fetch> and
  <https://nextjs.org/docs/app/getting-started/caching>, with this
  repository's `apps/web/next.config.ts` cross-checked against both.
- The actual rendered text and CSS class chain that
  `readReviewedTodayCount` parses (i.e. confirming there is exactly
  one place in `apps/web/src/` that renders the regex
  `(\d+) of \d+ words reviewed today`).
- The seed script
  (`apps/api/scripts/seed-synthetic-smoke-user.{sh,sql}`) and
  every `apps/api/migrations/*.sql` file, to confirm neither resets
  `daily_mission_snapshots.reviews_completed` between deploys.
- The unit-test files in `apps/api/business/missions/` and
  `apps/api/business/reviews/` for any pre-existing coverage of the
  "two reads seconds apart return the same value" invariant — none
  exists.
- The staging E2E spec file
  (`apps/web/tests/staging-e2e/core-loop.staging.spec.ts`) for the
  exact `reviewedBefore`/`reviewedCards`/`reviewedAfter` read
  pattern the failure tripped on.
- The nginx reverse-proxy config in
  `infra/nginx/conf.d/10-staging-web.conf` and
  `infra/nginx/conf.d/20-api-staging.conf`, to confirm no
  caching-related directives and no `Cache-Control` headers.
- Repository policy validation:
  `bash scripts/governance/validate-governance.sh` and
  `bash scripts/governance/classify-change-risk.sh`.

## What the implementer environment could not verify

These are the items `VOC-053-TEST-00`'s procedure requires, but
which this environment cannot do:

- Live staging HTTP response headers for `GET /home` and for the
  underlying `GET /api/v1/daily-mission` (the `cache-control`,
  `age`, `x-nextjs-cache`, `cf-cache-status` (Cloudflare), or
  equivalent cache-status headers, plus the `Date` and any `Age`
  headers).
- Whether the second request's body matches a stale first-request
  snapshot despite a known intervening state change (the
  cache-staleness signal the spec says would confirm candidate (a)
  or the "fourth cause" out-of-band cache).
- The actual `daily_mission_snapshots` row state for the synthetic
  account in real staging during and after a real test run (read
  directly via SQL).
- Whether two requests seconds apart with no intervening review
  completed return the same `reviewsCompleted` value (the
  `VOC-053-TEST-02` invariant) — needs real staging.
- Whether the same failure reproduces against a fresh
  synthetic account/day (candidate (c)'s own first-thing-to-try
  per issue #450's own suggestion).
- Whether the second request is actually reaching the Go API or is
  being served from a CDN/edge/proxy cache in front of the staging
  origin (the most plausible "fourth cause" if candidates (a)–(c)
  are all ruled out by live verification).

These items cannot be done by an implementer session that only has
read access to the repository; they require the same staging
access prior packages (VOC-050, VOC-052) have used via
`deploy-staging.yml`'s real-run path. This is consistent with what
issue #450 itself says: "both need real investigation ... that
this environment doesn't have the access to do blind."

## findings/VOC-053-read-rendered-text

`readReviewedTodayCount` in
`apps/web/tests/staging-e2e/core-loop.staging.spec.ts:83-92` parses
the page with the regex `(\d+) of \d+ words reviewed today`. A
repository-wide grep confirms there is exactly one JSX site
rendering this template:

- `apps/web/src/app/(app)/home/page.tsx:50` —
  `{reviewedWordsToday} of {missionTargetWords} words reviewed today ({missionProgressPercent}%)`

where `reviewedWordsToday` is destructured from
`dailyMissionResponse.data` at
`apps/web/src/app/(app)/home/page.tsx:24`, and
`dailyMissionResponse` comes from `client.getDailyMission()` called
at `apps/web/src/app/(app)/home/page.tsx:14`.

So the displayed number is exactly the
`DailyMissionDTO.reviewsCompleted` value the API returns; the page
itself does no rounding, clamping, or post-processing on the value
beyond the `Math.min(100, Math.round(... / ... * 100))` percentage
computation (which only affects the `(NN%)` suffix, not the
integer itself).

This rules out "the page is rendering the wrong number from the
same API value" as a class of bug.

## findings/VOC-053-frontend-fetch-shape

`apps/web/src/lib/api-server.ts:8-22` constructs `VocanovaClient`
with a custom `fetch` that:

1. Calls `await headers()` (line 9) — Next.js's runtime API for
   reading the incoming request's headers.
2. Copies the request's `Cookie` header onto the outbound fetch
   (line 17).
3. Calls `fetch(input, { ...init, headers })` (line 19) — passes
   no explicit `cache` option.

`apps/web/src/app/(app)/home/page.tsx:6-17` calls
`createServerApiClient()` then calls `client.getDailyMission()`
(which hits `packages/api-client/src/index.ts:649-663` → no `init`
is supplied, so the `request` helper's `init` is `undefined`, so
`init?.headers` is empty, so no `cache` option is forwarded).

`packages/api-client/src/index.ts:797-830`'s `request` method
calls `this.fetch(url.toString(), { ...init, method, headers,
credentials, body })` — i.e. it forwards whatever `init` the
caller supplied. Since neither the home page nor
`createServerApiClient` supplies a `cache` option, **no `cache`
option is set on the outbound `fetch` to
`GET /api/v1/daily-mission` from `/home`**.

## findings/VOC-053-nextjs-caching-model-applied

This repository runs Next.js 16.3.0 (`apps/web/package.json`'s
`"next": "16.3.0"`) with `apps/web/next.config.ts` containing
only `output: 'standalone'` and `outputFileTracingRoot` (plus a
`@sentry/nextjs` `withSentryConfig` wrapper, no caching options).
In particular, `cacheComponents: true` is **not** set, so the
[Cache Components](https://nextjs.org/docs/app/getting-started/caching)
model is **not** in effect and the
[previous Caching and Revalidating model](https://nextjs.org/docs/app/guides/caching-without-cache-components)
applies.

Per the Next.js 16.3.0
[`fetch` reference](https://nextjs.org/docs/app/api-reference/functions/fetch),
the default `cache` behavior is **`auto no cache`**:

> Next.js fetches the resource from the remote server on every
> request in development, but will fetch once during `next build`
> because the route will be statically prerendered. **If
> Request-time APIs are detected on the route, Next.js will fetch
> the resource on every request.**

The "Request-time APIs" cited are `cookies()`, `headers()`,
`searchParams`, dynamic `params`, and `connection()`. The home
page calls `await headers()` via `createServerApiClient` (see
`apps/web/src/lib/api-server.ts:9`); the middleware at
`apps/web/src/middleware.ts:65` calls `await fetch(...)` for the
`/api/v1/me` auth check (also a request-time access via the cookie
header read at line 61). Both of these trigger request-time API
detection on `/home`'s render path. No `force-dynamic` or
`export const dynamic` declaration is present in any
`apps/web/src/app/(app)/home/**` file (grep-confirmed), and the
home page has no `searchParams` access.

**Implication for candidate (a):** under Next.js 16.3.0's
documented behavior, the `fetch` to `/api/v1/daily-mission` from
`/home` is **not** cached at the Next.js Data Cache layer when
the route is dynamic — and the home route is dynamic, by
`headers()`. This narrows candidate (a) to a single remaining
live-verification question: whether a layer outside Next.js (a
CDN, a reverse proxy, or the API's own HTTP responses) is
serving a cached response. The committed
`infra/nginx/conf.d/10-staging-web.conf` and
`infra/nginx/conf.d/20-api-staging.conf` show no
`proxy_cache` / `Cache-Control` headers in nginx, and no
content-cache directives anywhere; the real Go API
(`apps/api/app/api/missions.go`) does not set any
`Cache-Control` header on its responses (grep-confirmed:
no `Cache-Control` or `cache-control` references in
`apps/api/`, and the
`apps/api/app/api/missions.go` handler returns the DTO through
Huma's default response writer, which does not set cache
headers).

Cloudflare sits in front of the staging origin (per
`infra/nginx/conf.d/00-cloudflare-real-ip.conf` and the
`/etc/nginx/certs/cert.pem` Cloudflare origin certificate in
`infra/nginx/conf.d/{10-staging-web,20-api-staging}.conf`).
Cloudflare's default caching posture is to cache only when the
origin returns `Cache-Control` headers indicating cacheability,
or for the
[Cloudflare default-cached file extensions and content types](https://developers.cloudflare.com/cache/concepts/default-cache-behavior/).
The Next.js Server Component response for `/home` is an
`text/html` document with no explicit `Cache-Control` header, so
Cloudflare's behavior here depends on the specific
account/zone configuration — which is not committed in this
repository and cannot be verified without a live HTTP check.
**Static reading therefore rules out candidate (a)'s Next.js
layer entirely**, but cannot rule in or out an out-of-band
Cloudflare or other CDN/proxy layer that the committed config
does not show.

## findings/VOC-053-DEP-01

The HTTP handler that resolves `now` and the per-user timezone
for a live `GET /api/v1/daily-mission` request is
`apps/api/app/api/missions.go:99-105`, registered by
`RegisterMissions` at `apps/api/app/api/missions.go:86-126`:

```go
func(ctx context.Context, input *GetDailyMissionInput) (*GetDailyMissionOutput, error) {
    view, err := svc.GetDailyMissionView(ctx, RequesterUserID(ctx), input.Timezone, time.Now())
    if err != nil {
        return nil, mapMissionsError(err)
    }
    return &GetDailyMissionOutput{Body: dailyMissionViewToDTO(view)}, nil
}
```

Key points from this single line:

- `time.Now()` is called **per request**, with no caching,
  memoization, or request-scope sharing. Two requests seconds
  apart see two `time.Now()` values a few seconds apart, not
  the same instant.
- `input.Timezone` is the optional client-supplied IANA timezone
  query parameter (`GetDailyMissionInput.Timezone` at
  `apps/api/app/api/missions.go:60-62`). For the staging
  core-loop E2E test, the `apps/web` `VocanovaClient.getDailyMission`
  call does **not** supply this parameter
  (`apps/web/src/app/(app)/home/page.tsx:14` calls
  `client.getDailyMission()` with no arguments;
  `packages/api-client/src/index.ts:649-663` only sets
  `query.set("timezone", params?.timezone)` if
  `params?.timezone` is truthy). So `input.Timezone` is `""`
  for both of the test's `/home` reads.
- `RequesterUserID(ctx)` is the per-request authenticated user
  ID resolved from the session cookie; it is stable across the
  test's two reads because the same session cookie is used
  (the spec mints one session and reuses it for every step
  via the `E2E_SESSION_COOKIE` environment variable).

The service-layer call chain is then
`missions.Service.GetDailyMissionView` at
`apps/api/business/missions/service.go:116-206`, which:

1. Resolves per-user settings via
   `gamification.Service.GetSettings`
   (`apps/api/business/gamification/service.go:30-44`), which
   reads the `user_settings` row via
   `gamification.Repository.GetUserSettings`
   (`apps/api/business/gamification/repository.go:132-150`) and
   passes `(stored, clientTimezone="")` into the pure
   `ResolveSettings`
   (`apps/api/business/gamification/timezone.go:60-82`).
2. For the synthetic account, the seed script
   (`apps/api/scripts/seed-synthetic-smoke-user.{sh,sql}`) does
   **not** create a `user_settings` row, and the spec's
   "Rerun safety" design means onboarding is skipped for this
   account (the seed sets `onboarding_status = 'completed'`,
   not the timezone). So `GetUserSettings` returns `nil`,
   `stored.Stored` is `false`, `ResolveSettings` falls through
   `clientTimezone == ""`, and returns
   `ResolvedSettings{Timezone: "UTC", DailyReviewTarget: 20}`
   per the D01 chain's third step
   (`apps/api/business/gamification/timezone.go:81`).
3. Computes `today` via `gamification.LocalDate(now, "UTC")` at
   `apps/api/business/missions/service.go:126`. `LocalDate`
   (`apps/api/business/gamification/timezone.go:88-95`) is a
   pure function of `(now, timezone)` with no I/O; for two
   requests seconds apart at 19:42 UTC, both `now` values are
   within `2026-08-09T19:42:xxZ`, so both produce
   `today = 2026-08-09T00:00:00Z`.
4. Begins a read transaction
   (`apps/api/business/missions/service.go:134`,
   `s.missions.db.BeginTx(ctx, nil)`).
5. Reads the snapshot via
   `missions.Repository.GetDailyMissionSnapshot(ctx, userID,
   today)` (`apps/api/business/missions/service.go:139`,
   `apps/api/business/missions/repository.go:28-39`). **This
   read uses `r.db.QueryRowContext(ctx, ...)` — a fresh
   connection from the pool, not the transaction**; the
   transaction is held open but unused on this code path. The
   SQL is a simple
   `SELECT ... FROM daily_mission_snapshots WHERE user_id = $1
   AND local_date = $2` with no `NOW()` or `current_date`
   involvement and no joins. The
   `(user_id, local_date)` unique index is created in
   `apps/api/migrations/20260725130001_voc030_p4_mission_tables.sql:35-36`.
6. If `snap == nil`, runs the lazy-creation + streak-
   reconciliation block (which uses `tx` for the writes; the
   reconciliation reads `streakSnaps` and `graceBalance` via
   the same connection pool pattern). If `snap != nil` (the
   path that returns the prior run's residue), the transaction
   is committed immediately without any write
   (`apps/api/business/missions/service.go:182-184`).
7. Loads the shared streak via `loadStreakAndGrace` (also via
   `r.db`).
8. Builds the view from `snap.ReviewsCompleted`
   (`apps/api/business/missions/service.go:196`) and returns
   the DTO.

**Per-request drift assessment (VOC-053-DEP-01 resolution):**
with the documented test conditions (no midnight boundary at
19:42 UTC, synthetic account has no stored `user_settings` row,
no client timezone is supplied), there is **no code path in
this trace that could produce a different `local_date` or a
different `snap.ReviewsCompleted` between two requests seconds
apart for the same user**. The repository's only
`UPDATE daily_mission_snapshots` statements that affect
`reviews_completed` are
- `IncrementReviewsCompleted`
  (`reviews_completed = LEAST(reviews_completed + 1,
  review_target)`,
  `apps/api/business/missions/repository.go:188-194`) and
- the read path's lazy creation's `ON CONFLICT DO UPDATE`
  (`apps/api/business/missions/repository.go:154-157`, which
  only touches `timezone`, `review_target`, and `updated_at`,
  never `reviews_completed`).

No `UPDATE` in any `apps/api/migrations/*.sql` or any other
Go file in `apps/api/` decrements `reviews_completed`. The
`accounts` package's user-anonymization
`UPDATE daily_mission_snapshots SET user_id = $2 ...`
(`apps/api/business/accounts/postgres.go:462-464`) reassigns
ownership, not `reviews_completed`, and the synthetic
account is never anonymized (the seed creates it with
`status = 'active'` and there is no production path that
soft-deletes it). **Static reading finds no committed code
path that decrements `reviews_completed`.**

The only plausible non-code explanations for the observed
decrease that static reading does not rule out are:

- A `daily_mission_snapshots` row whose `reviews_completed`
  was actually decremented or zeroed by something other than
  this read path (e.g. an out-of-band SQL command issued by
  some maintenance step, or a direct DB write that the
  committed workflows do not perform). None in this
  repository's committed code, workflows, or scripts; cannot
  be ruled in without live DB inspection.
- A different `daily_mission_snapshots` row being read on
  the second request than on the first. This would require
  `local_date` to differ between the two requests, which the
  trace above rules out for the documented conditions; or a
  different `user_id` being resolved, which
  `RequesterUserID(ctx)` cannot produce because the same
  session cookie is used; or the read landing on a different
  physical row than the one previously written, which the
  `(user_id, local_date)` unique index makes impossible.
- A misread by the test itself (e.g. `readReviewedTodayCount`
  matching the wrong element, or a stale `getByText` locator
  from a prior render). Confirmed there is exactly one site
  in `apps/web/src/` rendering the matching text, and the
  page is a Server Component that does not duplicate the
  count anywhere on `/home`. The production E2E
  (`apps/web/tests/e2e/core-loop.spec.ts:296-305`) exercises
  the same regex pattern (with a hard-coded
  `1 of 20 words reviewed today`) against the mock API
  without observing this decrease, but it runs against the
  in-memory mock, not real staging, so it is not
  dispositive of a real-backend issue.

## Candidate ruling

### Candidate (a) — Next.js Data Cache / Full Route Cache staleness

**Ruled out by static reading of this repository's Next.js
16.3.0 configuration.** Default `fetch` cache behavior is
`auto no cache`; the home route is dynamic (calls
`headers()`); no `cache: 'force-cache'` or
`next: { revalidate }` option is set anywhere in the
`getDailyMission()` call path. No committed code in this
repository adds a same-origin CDN in front of staging, and
the real Go API does not set `Cache-Control` headers. The
only remaining live-verification question is whether an
out-of-band CDN/proxy (Cloudflare's account-level cache
rules, in particular) is serving a cached response, which
is a "fourth cause" rather than candidate (a) as issue
#450 framed it.

### Candidate (b) — Real backend day-boundary/timezone bug in `reviews_completed` / `local_date`

**Ruled out by full static trace of the read path.** The
handler → service → repository trace is documented under
`findings/VOC-053-DEP-01`. No code path in the trace can
produce a different `local_date` or a decremented
`reviews_completed` between two requests seconds apart at
19:42 UTC under the documented conditions. The only
backend-side explanations for the decrease that static
reading does not rule out are (1) an out-of-band SQL command
that zeroed or decremented the synthetic account's
`reviews_completed` (none in committed code or workflows;
cannot be ruled in without live DB inspection) and (2) a
misread by the test itself (cannot be ruled in without live
browser trace inspection).

### Candidate (c) — Synthetic-account test-data interaction

**Cannot be ruled in or out from this environment.** The
drafting pass and this investigation both confirm the
persistent-synthetic-account reuse design (per PR #420's
review comment) is intentional. What we cannot verify
without live staging is whether the prior run that "should"
have left `reviewedBefore = 1` residue actually did so
(versus the row existing with `reviews_completed = 0` for
some other reason, in which case the "decrease" would be
from an already-0 state to a still-0 state — but
`Expected: >= 1, Received: 0` rules that out, since the
assertion would have to have been `0 >= 0`, which is
`true`, not a failure) or whether a fresh-account/day
reproduction attempt would surface the same failure.

### Fourth cause (this investigation pass's additional candidate)

**Cannot be ruled out, and not previously named by issue
#450.** If candidates (a)–(c) are all ruled out by live
verification, the most plausible remaining "fourth cause" is
an out-of-band caching layer (Cloudflare's account-level
caching, in particular) in front of staging that is not
visible in this repository's committed deploy configuration.
The `web` and `api` services in `infra/docker-compose.yml`
and `infra/docker-compose.production.yml` do not run a
reverse proxy in front of themselves in the local/staging
compose setup, but the production deployment topology (which
staging mirrors in some respects) is not fully documented in
committed files. A live check of the staging origin's IP, a
`curl -I` against `/home` and `/api/v1/daily-mission` (with
the synthetic session cookie from the existing
`POST /ops/synthetic-smoke-test/session` mint path) showing
`age`, `cf-cache-status` (or equivalent), and the `Date`
header, and a comparison of two requests seconds apart
showing whether the second matches the first's body, would
either confirm or rule this in.

## findings/VOC-053-required-live-verification

To move from "inconclusive" to a confirmed root cause, the
following must be checked against real staging with a
persistent synthetic account in the documented failure state
(i.e. a run where `reviewedBefore >= 1` from a prior run's
residue):

1. **Confirm candidate (a) is out at the live HTTP layer.**
   `curl -I https://staging.vocanova.site/api/v1/daily-mission`
   (with the synthetic session cookie and CSRF token from the
   existing `POST /ops/synthetic-smoke-test/session` mint
   path) and check the `cache-control`, `age`, and any
   `cf-cache-status` / `x-nextjs-cache` /
   `x-vercel-cache` / similar headers. Also
   `curl -I https://staging.vocanova.site/home` for the
   same. No `cache-control: ...max-age...` / `age > 0` /
   `cf-cache-status: HIT` / cache-hit header confirms
   candidate (a) is out for the actual deployed build; the
   presence of any of those on the second request confirms
   the "fourth cause" out-of-band cache is in play.
2. **Confirm the synthetic account's
   `daily_mission_snapshots` row state for the current local
   day during and after a real test run.** `SELECT
   reviews_completed, local_date, status, updated_at FROM
   daily_mission_snapshots WHERE user_id = '<synthetic UUID>'
   ORDER BY local_date DESC LIMIT 3`. The pre-test state
   should show `reviews_completed = 1` (or whatever the
   prior run left); the post-step-2 state should match; the
   post-step-7 state should match; the post-step-1 state of a
   fresh-account reproduction should be `0` (lazy-created).
3. **Reproduce the failure against a fresh synthetic
   account/day** per issue #450's own first-thing-to-try. If
   the fresh-account run passes step 7 (because
   `reviewedBefore = 0` and `reviewedAfter = 0` from lazy
   creation), that strongly supports candidate (c) (residue
   from a prior run interacting with something) or a "fourth
   cause" out-of-band cache. If the fresh-account run also
   fails step 7, the issue is not in residue at all and
   candidates (a)/(b)/(fourth-cause) get much higher
   probability.
4. **Confirm `readReviewedTodayCount`'s matched text in a
   real browser trace.** A Playwright trace of step 7's
   `page.goto("/home")` showing the exact rendered text the
   regex matched (and that it came from the live `/home`
   HTML, not from any prerendered shell or client-side
   hydration mismatch) rules in or out a page-rendering bug.
5. **Check the staging origin for any out-of-band cache.**
   `dig` / `host` / `nslookup` of `staging.vocanova.site`
   and `api-staging.vocanova.site`; the production deploy
   configuration in `infra/` does not show a reverse proxy
   in front of either service in the compose-based deploy
   path, but the actual staging host's network configuration
   and the Cloudflare account's specific caching rules are
   not in the repository.

## Impact on `planned_implementation_risk_floor`

`planned_implementation_risk_floor` in `change.yaml` is the
`R3` draft proposal, pending the actual floor being confirmed
by running `scripts/governance/classify-change-risk.sh`
against T01's real, post-investigation file list. This T00
implementation produced **no file change**, so the classifier
(run below) still reports path-based floor `R1` based on
the working tree. T01's choice of fix determines the actual
file set; per `specification.md`'s "Risk and protected areas"
and the candidate ruling above, the most likely real floor
remains `R3` (touches under `apps/web/src/` or
`apps/api/business/{missions,gamification}/`, neither a more
sensitive protected path). The migration-conditional note in
`specification.md` is unchanged: this investigation did not
find evidence that any candidate fix would need
`apps/api/migrations/`, so the conditional R4 escalation on
a migration path is not currently triggered, but is not
foreclosed either — if live verification (item 2 above)
shows incorrect historical `daily_mission_snapshots` rows,
the migration question re-opens and must be resolved before
T01.

## Local deterministic checks run for this task

```bash
bash scripts/governance/validate-governance.sh
bash scripts/governance/classify-change-risk.sh
git diff --check
```

Commands executed in this implementation run:

- `bash scripts/governance/validate-governance.sh` (pass;
  "Repository foundation validation passed. Governance
  structure validation passed.")
- `bash scripts/governance/classify-change-risk.sh` (pass;
  reports path-based floor `R1` only because no file change
  has been staged, not because T00's deliverable lowers the
  package's `planned_implementation_risk_floor: R3` — see
  "Impact on `planned_implementation_risk_floor`" above)
- `git diff --check` (pass; no diff, no whitespace errors)

## T01 entry conditions

`VOC-053-T01` should **not** begin a fix until the
`findings/VOC-053-required-live-verification` items above
have been checked against real staging and one of the
following holds:

- A specific root cause is confirmed with direct evidence
  (caches, timezone bug, or test-data interaction), with the
  specific evidence that confirmed it, in the form of an
  addendum to this evidence file (or a T01-predecessor
  investigation task, if the live verification surfaces an
  entirely new angle).
- OR the live verification narrows the candidates enough
  that the implementer can pick the highest-probability
  candidate and scope the fix to that, recording the chosen
  candidate and why it was chosen over the others in T01's
  own evidence file (per `acceptance-criteria.md`'s
  `VOC-053-AC-00` / `VOC-053-AC-01` requirements).

If neither holds, T01 is not yet safe to start; that is the
honest answer this investigation pass can give, per the
drafting pass's own "or not fully ruled out, if evidence is
inconclusive — record that honestly too" requirement.

## What changed since the previous attempt (attempt 1)

This is attempt 2 of T00. The previous attempt (commit
74fd157, on this branch's history prior to its current base
1811026) reached the same overall conclusion. Re-reading
the same files independently on the current `develop` base
this attempt was launched from:

- The four files the drafting pass named
  (`apps/web/src/app/(app)/home/page.tsx`,
  `apps/web/src/lib/api-server.ts`,
  `apps/api/business/missions/repository.go`,
  `apps/api/business/gamification/timezone.go`) are
  byte-for-byte unchanged from the previous attempt's base
  (730d820), per `git log` and a `git diff` against the
  previous attempt's parent. The trace in
  `findings/VOC-053-DEP-01` therefore still describes the
  live code path.
- The HTTP handler at
  `apps/api/app/api/missions.go:99-105` is unchanged.
- `apps/web/next.config.ts` is unchanged (still no
  `cacheComponents: true`, still `output: 'standalone'` +
  `outputFileTracingRoot` + Sentry wrapper only).
- `apps/web/package.json`'s `next` pin is still
  `16.3.0`.
- The nginx configs
  (`infra/nginx/conf.d/10-staging-web.conf`,
  `infra/nginx/conf.d/20-api-staging.conf`) are unchanged.
- The seed script
  (`apps/api/scripts/seed-synthetic-smoke-user.{sh,sql}`)
  is unchanged.
- `tests/staging-e2e/core-loop.staging.spec.ts` is
  unchanged (still the same `readReviewedTodayCount`
  helper, the same step 7 assertion at line 375, the same
  `REVIEWED_TODAY_PATTERN` regex at line 51).

The only deltas against the previous attempt's
`develop` base are: (a) the
[#457](https://github.com/KARSIFT/vocanova-platform-sandbox/pull/457)
Sentry host change (no code path or
configuration relevant to this issue); and (b) the
roster-merge commit for this package's own task-issue
bookkeeping (also no code path or configuration relevant
to this issue). Neither changes anything in the four
files the issue names, the HTTP handler, the nginx
configs, or the test's assertion. The previous attempt's
investigation findings are therefore still valid; this
re-investigation adds no new conclusion but does
independently re-confirm the trace, narrowing the
chance that the previous attempt's conclusion was
inferred rather than re-checked.

## Addendum — live verification of `findings/VOC-053-required-live-verification` item 1 (candidate a)

Performed after this attempt's investigation, by the founder-gate
delegate directly (not the AI implementer, which has no live-staging
network access): item 1's exact check, plus a real Playwright run.

**Diagnostic added**: PR [#458](https://github.com/KARSIFT/vocanova-platform-sandbox/pull/458)
added `recordHomeResponseDiagnostic()` to
`tests/staging-e2e/core-loop.staging.spec.ts`, recording the `/home`
document response's `cache-control`, `x-nextjs-cache`,
`cf-cache-status`, `age`, and `vary` headers as test annotations on
both step 1 (baseline) and step 7 (post-review) home loads. Merged to
`develop`, triggering a real `deploy-staging.yml` run
([31337990160](https://github.com/KARSIFT/vocanova-platform-sandbox/actions/runs/31337990160))
against the actual persistent synthetic account.

**Result** (from the run's Playwright HTML report, both loads):

```
step 1 (baseline):    status=200 date=Sun, 09 Aug 2026 21:57:48 GMT
  age=(absent)
  cache-control=private, no-cache, no-store, max-age=0, must-revalidate
  x-nextjs-cache=(absent)
  cf-cache-status=DYNAMIC
  vary=rsc, next-router-state-tree, next-router-prefetch, next-router-segment-prefetch, Accept-Encoding

step 7 (post-review): status=200 date=Sun, 09 Aug 2026 21:57:56 GMT
  age=(absent)
  cache-control=private, no-cache, no-store, max-age=0, must-revalidate
  x-nextjs-cache=(absent)
  cf-cache-status=DYNAMIC
  vary=rsc, next-router-state-tree, next-router-prefetch, next-router-segment-prefetch, Accept-Encoding
```

The same run reproduced the underlying bug at step 7's assertion:
`Expected: >= 3, Received: 0` (the page rendered "0 of 20 words
reviewed today" per the failure's page snapshot, immediately after the
diagnostic-annotated response was captured for the same navigation).

Separately, a direct `curl -I` against the unauthenticated `/home` and
`/api/v1/daily-mission` routes (same investigation pass) showed
`cf-cache-status: DYNAMIC` with no `cache-control` header at all on
either route, consistent with the authenticated result above.

**Candidate (a) is now ruled out with direct live evidence**, not just
static reading: `cache-control` is an explicit no-store directive on
both loads (identical value, not merely absent), `x-nextjs-cache` never
appears (Next.js's Full Route Cache is not involved in serving this
route at all - consistent with the static finding that a Server
Component reading `headers()` opts the route out of the Full Route
Cache), and Cloudflare reports `DYNAMIC` both times (no CDN-edge
caching). This satisfies `findings/VOC-053-required-live-verification`
item 1 in full. Item 4 (real browser trace of the rendered text) is
also satisfied incidentally by the same run's failure page snapshot,
which shows the counter rendering "0 of 20" - a real DOM read, not a
regex/rendering artifact.

Items 2, 3, and 5 (direct `daily_mission_snapshots` row inspection, a
fresh-account reproduction, and an out-of-band-cache network check)
remain unconfirmed - this delegate has no staging database or DNS/
network access. However, with candidate (a) now conclusively ruled out
by direct evidence (not inference) and the previous static analysis
already finding no decrement/dual-computation bug in the SQL or pure
timezone function it could read, **candidate (b) - a real backend
day-boundary/timezone bug in how the live HTTP handler resolves
`now`/timezone per request (`VOC-053-DEP-01`, still open) - is now the
leading candidate by elimination**, with candidate (c) (test-data
interaction) remaining a live secondary possibility neither confirmed
nor ruled out.

### T01 entry condition met

Per this evidence file's own "T01 entry conditions" section, the second
bullet applies: live verification has narrowed the candidates enough
(one of three ruled out with direct evidence, not inference) that
`VOC-053-T01` may proceed, scoped to investigating and fixing candidate
(b) first (resolving `VOC-053-DEP-01`'s still-open question of exactly
where/how `apps/api/app/api/missions.go`'s handler resolves per-request
`now`/timezone, per `apps/api/business/missions/repository.go` and
`apps/api/business/gamification/timezone.go`), falling back to
candidate (c) if (b)'s trace turns up no defect. T01 must still record
its own evidence per `VOC-053-AC-00`/`VOC-053-AC-01` rather than treat
this addendum as a substitute for that.

## Addendum 2 — live verification of `findings/VOC-053-required-live-verification` item 2 (daily_mission_snapshots row state)

VOC-053-T01's attempt 1 (closed PR #460, per its own honest evidence)
correctly declined a fix without this data. PR #461/#462 added a
temporary SSH diagnostic to `deploy-staging.yml` (`docker compose exec
postgres psql`, run right after the E2E journey, `if: always()`) that
dumps the synthetic account's `daily_mission_snapshots` rows and the
DB server's own `now()`/timezone. Two real deploy runs captured data:

**DB server clock** (both runs): `now() = 2026-08-09 23:4X:XX+00`,
`TimeZone = UTC` - the server's own clock and timezone setting are
correct and consistent with the E2E run's real wall-clock time. No
server-side clock skew.

**`daily_mission_snapshots` state** (run
[31342544145](https://github.com/KARSIFT/vocanova-platform-sandbox/actions/runs/31342544145),
captured immediately after the E2E journey completed):

```
 local_date | timezone | review_target | reviews_completed | status |          updated_at
------------+----------+---------------+-------------------+--------+------------------------------
 2026-08-09 | UTC      |            20 |                  0 | open   | 2026-08-09 08:06:03.51608+00
(1 row)
```

Exactly **one** row exists for the current local date. `reviews_completed
= 0`, and `updated_at` (08:06:03 UTC) predates this run entirely (the
run itself started ~23:42 UTC) - nothing incremented this row during
this run. The same run's Playwright output was `1 passed (10.5s)`: with
`reviewedBefore = 0` (matching this row) and the review queue already
empty (`reviewedCards = 0`, no due cards found this run), the assertion
`reviewedAfter >= reviewedBefore + reviewedCards` was `0 >= 0 + 0` - a
trivial pass that did **not** exercise the actual review-completion
path (`IncrementReviewsCompleted`) at all.

### Interpretation

This run did not reproduce issue #450's original failure precondition
(a nonzero `reviewedBefore` from residue, with real cards available to
review in the same run) - the synthetic account's due-review queue is
currently empty, so no run since this diagnostic was added has
exercised the increment path. That means this addendum cannot directly
confirm or refute candidate (b) by watching an increment happen and
then disappear.

It does add one new, real data point: with caching definitively ruled
out (Addendum 1) and the DB server's own clock/timezone confirmed
correct, `reviews_completed` genuinely holding `0` in the database
itself (not just in a served response) for the entire day so far is
consistent with candidate (b)'s day-boundary/`local_date` theory in a
specific way - if the live HTTP handler's per-request `local_date`
resolution (`VOC-053-DEP-01`, still open) is not perfectly stable
across nearby requests, a request could read or create a *different*
`daily_mission_snapshots` row (a different `local_date` value) than a
sibling request seconds apart, independent of any caching layer. Only
one row exists right now because this run's requests apparently
resolved consistently, but issue #450's original failure - a real
`reviewedBefore = 1` immediately followed by `reviewedAfter = 0` in the
same run - would be explained by two nearby requests resolving
*different* `local_date` values and therefore reading/creating two
different rows, one with residual state from an earlier read and one
freshly created at `reviews_completed = 0`.

### Recommended T01 scope (unchanged from Addendum 1, now with DB confirmation of no anomalous historical data)

`VOC-053-DEP-01` remains the concrete, actionable next step: trace and
harden exactly how the live HTTP handler
(`apps/api/app/api/missions.go`) resolves `now()`/timezone into a
`local_date` value per request, and confirm whether two calls within
the same request-handling window can produce different values (e.g.
from `time.Now()` being called independently in more than one place, a
per-request vs. per-connection clock source, or a timezone lookup that
can itself vary). No corrective data migration is indicated by this
addendum - the single existing row for today is internally consistent
(no negative or decreasing history visible), so `migration_required`
remains `unknown-conditional-on-root-cause`, not confirmed-needed.
