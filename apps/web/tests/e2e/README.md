# apps/web/tests/e2e/

End-to-end Playwright suites for `@vocanova/web`. Net-new as of
**VOC-031-T07a** (this directory was documented in
`docs/design/08-web-app-design.md` §"Architecture" but never
created). Subsequent tasks will extend it:

| Task      | Adds                                                                                                                                          |
| --------- | --------------------------------------------------------------------------------------------------------------------------------------------- |
| T07a      | This directory + `playwright.config.ts` + a single Home scan at 1280x720 + the mock API server. **Scaffolding only.** (this PR)               |
| T07b      | Every remaining core-loop screen (Discover, Word Detail, Reviews, sentence feedback, Progress, Onboarding, Settings, Settings/account) at 360px, 430px, and the 1280x desktop width T07a already covers. Plus explicit keyboard-reachability and non-color-only-feedback assertions. |
| T08       | The DOC-10 §7 full core-loop functional flow (auth → onboarding → discover → save → review session → sentence submission → deterministic AI feedback → progress update → settings change → logout → unauthenticated-access rejection). |
| T09       | Lighthouse CI budgets (separate harness; this directory's Playwright config does not host it).                                               |

## Running locally

```bash
# One-time: install the Playwright browser. Chromium only - T07a
# runs exactly one project, the "home-desktop-1280" project.
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
  README.md                  <- this file
  home-accessibility.spec.ts <- the single T07a scan
  axe-helper.ts              <- shared WCAG 2.2 AA scan + violation formatter
  mock-api-server.mjs        <- dependency-free node:http server that stands
                                in for the Go /api/v1 backend in the harness
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
- `playwright test` (one project, one spec, one browser): ~10-30 s
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

## Out of scope (recorded for the next implementer)

- Multi-screen coverage → T07b.
- Multi-viewport coverage → T07b.
- Keyboard-reachability and non-color-only-feedback explicit
  assertions → T07b (axe alone is not enough for the
  acceptance criterion's full wording).
- Full DOC-10 §7 core-loop functional flow → T08.
- Lighthouse CI budgets → T09.
- Replacing `mock-api-server.mjs` with a real
  Postgres-backed API + auth fixture → deferred until F3
  staging exists (`VOC-031-DEP-02`); the mock is a
  scaffolding artifact and is explicitly not a substitute for
  the T08 functional coverage.
