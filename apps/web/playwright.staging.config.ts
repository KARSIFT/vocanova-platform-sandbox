// VOC-050-T02 - Playwright config for the post-deploy core-loop
// journey against a real deployed environment (staging).
//
// This config lives beside playwright.config.ts, at the apps/web
// package root, because Playwright resolves testDir/outputDir
// relative to the CONFIG FILE's own directory. A config nested
// under tests/ with package-root-relative paths silently resolves
// to directories that do not exist, which is how the previous
// revision of this task ended up pointing at
// apps/web/tests/e2e/tests/e2e.
//
// Differences from playwright.config.ts (the PR-time, mock-backed
// accessibility/E2E suite), all deliberate:
//
//   * No `webServer`. The target is an already-deployed
//     environment reached over HTTPS; nothing is started or torn
//     down locally, and mock-api-server.mjs is never involved.
//   * A separate `testDir` (tests/staging-e2e). Keeping the
//     staging journey out of tests/e2e is what stops `pnpm
//     test:e2e` - which runs the default config over that whole
//     directory - from picking up a spec that requires a minted
//     real-backend session and would fail against the mock server.
//   * `retries: 0`. This run is a deploy gate (VOC-050-AC-03):
//     a retry could let a genuinely broken staging deploy pass on
//     a second attempt, which is exactly the fail-open behavior
//     the acceptance criterion forbids.
//   * Longer timeouts. Every step crosses the public internet,
//     Cloudflare, nginx, the Go API, and Postgres rather than a
//     loopback mock.

import { defineConfig, devices } from "@playwright/test";

const DEFAULT_STAGING_WEB_BASE_URL = "https://staging.vocanova.site";
const JOURNEY_TIMEOUT_MS = 240_000;
const ASSERTION_TIMEOUT_MS = 20_000;

const baseURL =
  process.env.STAGING_WEB_BASE_URL ?? DEFAULT_STAGING_WEB_BASE_URL;

export default defineConfig({
  testDir: "./tests/staging-e2e",
  fullyParallel: false,
  forbidOnly: true,
  retries: 0,
  workers: 1,
  reporter: [
    ["list"],
    ["github"],
    ["html", { outputFolder: "playwright-report-staging", open: "never" }],
  ],
  outputDir: "./test-results-staging",
  timeout: JOURNEY_TIMEOUT_MS,
  expect: { timeout: ASSERTION_TIMEOUT_MS },
  use: {
    baseURL,
    trace: "retain-on-failure",
    video: "retain-on-failure",
    screenshot: "only-on-failure",
  },
  projects: [
    {
      // One representative desktop width, mirroring
      // core-loop.spec.ts's own project scope: the mobile
      // viewports are the accessibility suite's concern
      // (VOC-031-T07b), not this functional deploy gate's.
      name: "staging-desktop-1280",
      use: {
        ...devices["Desktop Chrome"],
        viewport: { width: 1280, height: 720 },
        headless: true,
      },
    },
  ],
});
