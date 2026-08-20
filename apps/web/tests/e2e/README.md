# apps/web/tests/e2e/

End-to-end Playwright suites for `@vocanova/web`. Net-new as of
**VOC-031-T07a** (this directory was documented in
`docs/design/08-web-app-design.md` §"Architecture" but never
created). Subsequent tasks extend it:

| Task      | Status   | Adds                                                                                                                                          |
| --------- | -------- | --------------------------------------------------------------------------------------------------------------------------------------------- |
| T07a      | shipped  | This directory + `playwright.config.ts` + a single Home scan at 1280x720 + the mock API server. **Scaffolding only.**                        |
| T07b      | shipped  | Every remaining core-loop screen (Discover, Discover/[situation], Discover/[situation]/[word], Reviews, Progress, Onboarding, Settings) plus Home at the 360px and 430px viewports. Explicit keyboard-reachability and non-color-only-feedback assertions on top of the axe scan (axe alone is not sufficient for the T07b acceptance criterion's full wording). |
| T08       | shipped  | The DOC-10 §7 full core-loop functional flow (auth → onboarding → discover → save → review session → sentence submission → deterministic AI feedback → progress update → settings change → logout → unauthenticated-access rejection). One Playwright test, one representative desktop width (mirrors T07a's "ONE representative desktop width" scope; mobile projects self-skip). |
| T09       | shipped  | Lighthouse CI budgets in a separate directory (`apps/web/tests/lighthouse/`). 4 screens × 3 layouts = 12 audits, asserting the DOC-08 quality-standards thresholds (Performance 85+, Accessibility 95+, Best Practices 90+) against the same fixed local production build this directory's Playwright config serves. Wired into CI as `.github/workflows/lighthouse.yml`, mirroring this directory's `accessibility.yml` separation pattern. |
| VOC-073   | shipped  | Dedicated accessibility specs for four entry surfaces omitted from T07b: `/signin`, `/` (landing), `/auth/magic`, and `/settings/account` (extracted from `settings-accessibility.spec.ts` into its own file — no duplicate CI run for that screen). |

## T07b screen × viewport coverage matrix

| Screen                              | 360px | 430px | 1280x720 |
| ----------------------------------- | :---: | :---: | :------: |
| `/home`                             |  T07b |  T07b |  T07a    |
| `/discover`                         |  T07b |  T07b |  T07b    |
| `/discover/[situation]`             |  T07b |  T07b |  T07b    |
| `/discover/[situation]/[word]`      |  T07b |  T07b |  T07b    |
| `/reviews`                          |  T07b |  T07b |  T07b    |
| `/progress`                         |  T07b |  T07b |  T07b    |
| `/onboarding`                       |  T07b |  T07b |  T07b    |
| `/settings`                         |  T07b |  T07b |  T07b    |
| `/settings/account`                 | VOC-073 | VOC-073 | VOC-073 |
| `/signin`                           | VOC-073 | VOC-073 | VOC-073 |
| `/` (landing)                       | VOC-073 | VOC-073 | VOC-073 |
| `/auth/magic`                       | VOC-073 | VOC-073 | VOC-073 |

Every T07b cell above runs the same three checks:

1. `scanForAxeViolations` — WCAG 2.2 AA, zero critical/serious
   findings (the T07a acceptance bar, inherited unchanged).
2. `assertKeyboardReachable` — every focusable element on the
   screen is reachable by Tab; the page is operable from a
   keyboard alone.
3. `assertNonColorOnlyFeedback` — every screen-specific
   status indicator (e.g. "Done" / "Rest" on Progress,
   "Saved" on Discover, "You're all caught up" on Reviews)
   carries its state in text, not color alone.

## Running locally

```bash
# One-time: install the Playwright browser. Chromium only.
# CI uses infra/scripts/install-playwright-chromium.sh (browser cache +
# bounded deps install); local operators may run that script or the
# equivalent commands below.
bash infra/scripts/install-playwright-chromium.sh
# Or manually:
# pnpm --filter @vocanova/web exec playwright install chromium
# pnpm --filter @vocanova/web exec playwright install-deps chromium

# Build the web app once (Playwright's webServer uses `pnpm start`
# so the test runs against the production bundle, not the dev
# server's hot-reload variant).
pnpm --filter @vocanova/web build

# Run the suite. The Playwright config starts the mock API server
# and the Next.js production server itself; you do not need to
# start them by hand. Set CI=1 to mirror the CI experience.
pnpm --filter @vocanova/web test:e2e
```

## Layout

```
tests/e2e/
  README.md                           <- this file
  axe-helper.ts                       <- shared WCAG 2.2 AA scan +
                                         violation formatter +
                                         keyboard-reachability +
                                         non-color-only-feedback
                                         helpers (T07a + T07b)
  home-accessibility.spec.ts          <- the single T07a scan
                                         (Home @ 1280x720; self-
                                         skips on mobile projects
                                         to preserve T07a's "ONE
                                         representative desktop
                                         width" scope)
  home-mobile-accessibility.spec.ts   <- T07b's Home @ 360/430 scan
  discover-accessibility.spec.ts      <- T07b's Discover, Discover/
                                         [situation], Discover/
                                         [situation]/[word] scans
  reviews-accessibility.spec.ts       <- T07b's /reviews scan
  progress-accessibility.spec.ts      <- T07b's /progress scan
  onboarding-accessibility.spec.ts    <- T07b's /onboarding scan
  settings-accessibility.spec.ts      <- T07b's /settings scan
  settings-account-accessibility.spec.ts
                                      <- VOC-073-T03's /settings/account
                                         scan (extracted from
                                         settings-accessibility.spec.ts)
  signin-accessibility.spec.ts        <- VOC-073-T00's /signin scan
  landing-accessibility.spec.ts       <- VOC-073-T01's / (landing) scan
  auth-magic-accessibility.spec.ts    <- VOC-073-T02's /auth/magic scan
  core-loop.spec.ts                   <- T08's full core-loop
                                         functional flow (one
                                         representative desktop
                                         width; mobile projects
                                         self-skip)
  middleware-runtime-regression.spec.ts
                                      <- VOC-039-T01's Edge-vs-Node
                                         middleware runtime regression
                                         (issue #297); no browser or
                                         app server involved
  middleware-runtime-harness.ts       <- runs the real
                                         src/middleware.ts source in
                                         Next's own Edge sandbox and in
                                         Node, against a stub API on an
                                         ephemeral port
  mock-api-server.mjs                 <- dependency-free node:http
                                         server that stands in for
                                         the Go /api/v1 backend in
                                         the harness (T07a + T07b
                                         read-only fixtures + T08's
                                         full mutation surface and
                                         deterministic AI feedback
                                         rule table)
```

The root-level `playwright.config.ts` lives at `apps/web/` (not
under `tests/`) so the npm script `pnpm --filter @vocanova/web
exec playwright test` resolves it without extra flags; the
config's `testDir` points here.

## CI runner / browser-dependency reconciliation (VOC-031-DEP-04)

T07a / T07b / T08 introduce the first browser-driven test
harness this repository has ever had (`VOC-031-D00` confirmed
the absence of Playwright, axe-core, and Lighthouse CI in the
adopted base). The CI integration is a new workflow file
(`.github/workflows/accessibility.yml`) that runs separately
from `pipeline.yml`'s `ci` job - the `ci` job uses
`KARSIFT/karsift-ai-infra/.github/workflows/ci.yml@main` which
runs `pnpm run format:check lint typecheck test build`, none of
which install or invoke Playwright. Wiring the e2e suite into
that generic job would couple the entire pnpm validation
chain to a multi-minute Playwright browser install and
production build, regressing PR feedback time for changes
that have nothing to do with e2e coverage. The dedicated
workflow keeps the cost local to PRs that touch
`apps/web/**` or `apps/web/tests/**` (the `paths` filter) and
keeps the rest of the validation chain's contract intact.

The dedicated workflow's job time budget is **30 minutes**
(timeout-minutes), which covers:

- pnpm install against the frozen lockfile: ~2 min
- `next build` for the production bundle: ~3-5 min
- Playwright Chromium install via `install-playwright-chromium.sh`
  (restores `~/.cache/ms-playwright` from `actions/cache` first;
  bounded deps install with 120 s per-attempt timeout × 3 attempts):
  ~1-2 min on warm cache, longer on cold cache or slow apt mirrors
- `playwright test` (three projects, thirteen screens, one browser):
  ~1-3 min
- teardown + report upload: ~30 s

The first-time browser install is the largest unknown. A CI run with a
restored browser cache may skip the Chromium download, but system deps
are still installed with bounded retries. A clean-runner path pays the
full download cost. This is the same trade-off DOC-10 §7 "Level 2"
already accepts for "selected Playwright" coverage.

Browser version pinning: the shared install script invokes
`playwright install chromium` without specifying a version, so the
browser that ships with the locked `@playwright/test` version is used.
The lockfile pins the package version, so the browser version is
effectively pinned by transitivity - no separate `playwright install
--browser=...` config drift to manage.

System dependencies (`playwright install-deps chromium`) install the
missing shared libraries Chromium needs on `ubuntu-latest`. The script
uses explicit per-attempt timeouts and limited retries instead of an
unbounded apt wait. The workflow only runs on PRs that satisfy the
`paths` filter, so the cost is paid only when the e2e harness is
actually exercised.

## T07b-specific timing notes

The 4 prior implementer attempts on the unsplit T07 across
minimax-m3, kimi-k2.7-code, and claude-sonnet-5 all silently
hung for 35-55 minutes (T07a note in `tasks.md`). The split
into T07a (scaffolding) + T07b (full coverage) is the
founder's directive mitigation. With the harness proven by
T07a, T07b adds three projects to the same Playwright
config, two new mock endpoints, and seven new spec files;
none of those changes interact with the first-time browser
install (one-time, already paid by T07a's CI run on the same
runner image) or the production build (T07a's run cached
the `.next/` directory on the same PR branch). The
T07b-only CI run on a clean runner takes ~8 minutes, well
inside the workflow's 30-minute budget.

## Out of scope (recorded for the next implementer)

- Lighthouse CI budgets → T09.
- Replacing `mock-api-server.mjs` with a real
  Postgres-backed API + auth fixture → deferred until F3
  staging exists (`VOC-031-DEP-02`); the mock is a
  scaffolding artifact and is explicitly not a substitute
  for full staging evidence (the T11 staging evidence and
  T43 staged cross-user/CSRF/idempotency validations).
- The T08 suite runs on one representative desktop width
  (mirroring T07a's scope). The mobile functional flow is
  *not* covered separately; the mobile accessibility scans
  in T07b already exercise the same mobile-first screens.
  Promoting the T08 flow to the mobile projects would triple
  test time without a coverage gain beyond T07b's mobile
  accessibility scans, and is a T09-or-later decision once
  the staging evidence is in place.

## T08 specifics

The T08 test (`core-loop.spec.ts`) exercises every step of
the DOC-10 §7 documented core loop. Approach:

- **Auth (step 1).** The test sets the `vocanova_session`
  and `vocanova_csrf` cookies directly via
  `context.addCookies()`. The "auth" step in the spec is
  "be authenticated" - the magic-link / OAuth / session
  cookie issuance itself is a separate A1 surface covered by
  the auth module's own unit + integration tests. The T08
  test does, however, exercise the `POST /api/v1/auth/logout`
  mutation against the mock to verify the logout -> rejection
  handoff (steps 9-10).
- **CSRF.** Every mutation (POST/PATCH/DELETE except the
  auth routes) requires the `X-CSRF-Token` header to match
  the `vocanova_csrf` cookie value, matching the real
  backend's `CSRFMiddleware`
  (`apps/api/app/api/middleware.go`). The mock enforces this
  identically so a missing/mismatched CSRF in the e2e test
  fails in CI just as it would in production.
- **State isolation.** The test uses a per-run unique
  session ID (a `randomUUID()`-suffixed value) so its
  in-memory mock state never collides with the T07a/T07b
  accessibility scans, which use the mock's default session
  (no `vocanova_session` cookie).
- **Deterministic AI.** The mock's
  `evaluateSentenceFeedback` is a fixed rule table (length +
  target-word presence) - the same input always produces the
  same output, satisfying DOC-10 §7's "CI always uses a
  deterministic AI adapter, not a paid/nondeterministic
  provider call" rule. A real AI provider is never called.
- **Unauthenticated rejection (step 10).** The test sets an
  `e2e_unauthenticated=1` cookie (the same test-only cookie
  pattern `e2e_onboarding_status` uses) so the mock returns
  401 for `/api/v1/me`, which the Next.js auth-gate
  middleware uses to redirect `/home` -> `/signin`. The
  cookie is set only for this final step; clearing it (or
  the absence of a new test) leaves the T07a/T07b scans
  unaffected.

The T08 suite runs in the same
`.github/workflows/accessibility.yml` job as the T07a/T07b
scans (no separate CI wiring). T08 self-skips on the
mobile-360 / mobile-430 projects so it runs exactly once per
PR.

## VOC-039-T01 specifics (middleware runtime regression)

Issue #297: `src/middleware.ts` runs on the Edge runtime
unless it exports `runtime = "nodejs"`, and an Edge bundle
only sees environment values inlined at build time. This
deployment passes `API_BASE_URL` to the running container, so
under Edge `getApiBaseURL()` silently falls back to
`http://localhost:8080`, the `/api/v1/me` auth check never
reaches the deployed API, and authenticated learners are
bounced back to `/signin`.

`middleware-runtime-regression.spec.ts` reproduces that
runtime gap rather than asserting on source text:

- `middleware-runtime-harness.ts` loads the real
  `src/middleware.ts` and `src/lib/env.ts` sources (TypeScript
  annotations stripped, no statement rewritten) and evaluates
  them in Next's own vendored Edge sandbox
  (`next/dist/compiled/edge-runtime`) and in a Node `vm`
  context. Only the runtime differs between the two runs.
- The stub API listens on an **ephemeral** port, so its
  address exists only at run time and is discoverable only
  through `process.env.API_BASE_URL`. "Did the stub receive
  `/api/v1/me`" is therefore an unfakeable signal that the
  auth check resolved the runtime environment — and it stays
  deterministic even though the mock API server occupies
  `getApiBaseURL()`'s `localhost:8080` fallback during this
  suite's own run.
- The spec runs the auth check under whichever runtime
  `middleware.ts` actually declares, so it fails against a
  pre-VOC-039-T00 middleware and passes once the Node.js
  runtime is declared.

Like T08, it self-skips on the mobile projects (runtime
behavior is viewport-independent).
