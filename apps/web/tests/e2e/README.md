# apps/web/tests/e2e/

End-to-end Playwright suites for `@vocanova/web`. Net-new as of
**VOC-031-T07a** (this directory was documented in
`docs/design/08-web-app-design.md` §"Architecture" but never
created). Subsequent tasks extend it:

| Task      | Adds                                                                                                                                          |
| --------- | --------------------------------------------------------------------------------------------------------------------------------------------- |
| T07a      | This directory + `playwright.config.ts` + a single Home scan at 1280x720 + the mock API server. **Scaffolding only.**                        |
| T07b      | Every remaining core-loop screen (Discover, Discover/[situation], Discover/[situation]/[word], Reviews, Progress, Onboarding, Settings, Settings/account) plus Home at the 360px and 430px viewports. Explicit keyboard-reachability and non-color-only-feedback assertions on top of the axe scan (axe alone is not sufficient for the T07b acceptance criterion's full wording). |
| T08       | The DOC-10 §7 full core-loop functional flow (auth → onboarding → discover → save → review session → sentence submission → deterministic AI feedback → progress update → settings change → logout → unauthenticated-access rejection). |
| T09       | Lighthouse CI budgets (separate harness; this directory's Playwright config does not host it).                                               |

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
| `/settings/account`                 |  T07b |  T07b |  T07b    |

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
pnpm --filter @vocanova/web exec playwright install --with-deps chromium

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
  settings-accessibility.spec.ts      <- T07b's /settings and
                                         /settings/account scans
  mock-api-server.mjs                 <- dependency-free node:http
                                         server that stands in for
                                         the Go /api/v1 backend in
                                         the harness (T07a + T07b
                                         fixture data)
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
- `playwright install --with-deps chromium` (first run only;
  the GitHub-hosted runner image has no cached browser
  binaries): ~1-2 min
- `playwright test` (three projects, ten screens, one browser):
  ~1-3 min
- teardown + report upload: ~30 s

The first-time browser install is the largest unknown. It is
**not** suppressed on cache miss: a CI run with a cached
browser skips the install, but a clean-runner path always
pays the full download cost. This is the same trade-off
DOC-10 §7 "Level 2" already accepts for "selected Playwright"
coverage.

Browser version pinning: the workflow invokes
`pnpm exec playwright install --with-deps chromium` without
specifying a version, so the browser that ships with the
locked `@playwright/test` version is used. The lockfile pins
the package version, so the browser version is effectively
pinned by transitivity - no separate `playwright install
--browser=...` config drift to manage.

System dependencies (`--with-deps`) install the missing
shared libraries Chromium needs on `ubuntu-latest`. This
requires `apt-get`; the workflow only runs on PRs that
satisfy the `paths` filter, so the cost is paid only when
the e2e harness is actually exercised.

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

- Full DOC-10 §7 core-loop functional flow → T08.
- Lighthouse CI budgets → T09.
- Replacing `mock-api-server.mjs` with a real
  Postgres-backed API + auth fixture → deferred until F3
  staging exists (`VOC-031-DEP-02`); the mock is a
  scaffolding artifact and is explicitly not a substitute for
  the T08 functional coverage.
