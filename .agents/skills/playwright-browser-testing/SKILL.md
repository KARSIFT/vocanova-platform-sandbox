---
name: playwright-browser-testing
description: Use when writing or debugging Playwright tests in apps/web—E2E, accessibility, OAuth flows, and CI-safe browser automation with the repository-pinned @playwright/test version.
---

# Playwright browser testing (vocanova web)

Guidance for `apps/web` using **`@playwright/test` `1.62.1`** (declared in `apps/web/package.json`). Config: `apps/web/playwright.config.ts`, specs under `apps/web/tests/e2e/`. See `docs/development.md` for workspace validation tiers.

## Governance precedence

When this skill conflicts with `AGENTS.md`, `CLAUDE.md`, approved change packages, tests, or source code, the repository sources win.

## When to use

- Adding or fixing E2E or accessibility specs
- Debugging flaky Playwright runs locally or in CI
- OAuth or auth-flow tests under `apps/web/tests/e2e/`

## Golden rules

1. **`getByRole()` first** — resilient, user-centric locators
2. **No `page.waitForTimeout()`** — use web-first `expect(locator)` or `page.waitForURL()`
3. **Isolate tests** — no shared mutable state between specs
4. **`baseURL` from config** — no hardcoded hosts in tests
5. **Retries `1` in CI, `0` locally** — matches `playwright.config.ts`
6. **Traces `retain-on-failure`** — use Playwright trace viewer locally; do not paste raw trace archives into chat
7. **Mock external services only** — use `apps/web/tests/e2e/mock-api-server.mjs` for SSR/API boundaries

## Repository commands

```bash
# From repository root
pnpm --filter @vocanova/web test:e2e

# Typecheck e2e helpers
pnpm --filter @vocanova/web typecheck:e2e
```

`pnpm install --frozen-lockfile` installs the pinned package, not its Chromium
binary. Provision Chromium separately with the repository-owned prerequisite:

```bash
bash infra/scripts/install-playwright-chromium.sh
```

That script invokes the workspace-pinned Playwright version and is the path used by
CI. Follow `apps/web/tests/e2e/README.md`; **do not** run global Playwright installs
or unpinned `npx playwright` downloads.

## Layout

| Path                                     | Purpose                              |
| ---------------------------------------- | ------------------------------------ |
| `apps/web/playwright.config.ts`          | Projects, webServer, reporters       |
| `apps/web/tests/e2e/*.spec.ts`           | Core-loop and accessibility specs    |
| `apps/web/tests/e2e/axe-helper.ts`       | Shared axe assertions                |
| `apps/web/tests/e2e/mock-api-server.mjs` | Local API stub for SSR               |
| `apps/web/tests/staging-e2e/`            | Staging-only flows (separate config) |

## Writing tests

1. Colocate new specs in `apps/web/tests/e2e/` following existing naming (`*-accessibility.spec.ts`, `core-loop.spec.ts`).
2. Reuse `axe-helper.ts` for accessibility checks (`@axe-core/playwright`).
3. Respect project matrix (desktop `1280`, mobile `360` / `430`) — see config comments for T07a/T07b scope.
4. For auth flows, consult `signin-oauth-capability.spec.ts` and controlled-signup docs under `docs/operations/`.

## Debugging

1. Re-run a single file: `pnpm --filter @vocanova/web test:e2e -- path/to/spec.ts`
2. Open trace after failure: `pnpm exec playwright show-trace apps/web/test-results/.../trace.zip` from `apps/web`
3. Use `systematic-debugging` for root-cause work; use `verification-before-completion` and `pnpm validate` when the change spans packages.

## CI notes

- CI sets `CI=true`; config enables GitHub reporter and retries.
- Lighthouse workflow resolves Chromium via Playwright cache; watch for shell-expansion constraints in that resolution script.

## Safety

Test only applications this repository owns or staging environments documented for E2E. Do not read `.env*` files. Do not paste raw CI logs. Never automate third-party sites without permission. Do not expose session material, tokens, or screenshots containing personal data.
