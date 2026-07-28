// VOC-031-T07a + VOC-031-T07b Playwright config.
//
// T07a wired one project ("home-desktop-1280") that scans a single
// screen (Home) at one representative desktop width as the
// end-to-end proof the accessibility harness worked. T07b extends
// the matrix to every remaining core-loop screen at the three
// supported layouts: 360px, 430px, and the desktop width T07a
// already covered.
//
// webServer runs two processes: this harness's mock API server
// (so each server component can render without the real
// Go/Postgres backend) and the Next.js production build (so the
// test is running the same bundle CI will ship, not the dev
// server's hot-reload variant). The mock API server is started
// first because the Next.js server's SSR path makes
// /api/v1/me + the per-screen data reads during page rendering.
//
// T07a scope discipline: the T07a home-accessibility.spec.ts test
// self-skips on the two mobile projects (T07a's own scope is
// "ONE representative desktop width >=1024px") so the addition
// of mobile projects does not silently expand T07a. T07b's own
// home-mobile-accessibility.spec.ts covers the 360 and 430
// viewports for the same screen.

import { defineConfig, devices } from "@playwright/test";

const PORT_WEB = Number(process.env.PORT_WEB ?? 3000);
const PORT_MOCK_API = Number(process.env.MOCK_API_PORT ?? 8080);
const MOCK_API_URL = `http://127.0.0.1:${PORT_MOCK_API}/healthz`;
const WEB_URL = `http://127.0.0.1:${PORT_WEB}`;

// `reuseExistingServer` lets a developer run `pnpm start` and
// `node tests/e2e/mock-api-server.mjs` by hand before invoking
// `playwright test`; in CI neither is running ahead of time, so
// Playwright starts and tears them down itself.
const isCI = Boolean(process.env.CI);

const sharedLaunchOptions = {
  headless: true,
};

export default defineConfig({
  testDir: "./tests/e2e",
  fullyParallel: false,
  forbidOnly: isCI,
  retries: isCI ? 1 : 0,
  workers: 1,
  reporter: isCI
    ? [["list"], ["github"]]
    : [["list"]],
  outputDir: "./test-results",
  timeout: 30_000,
  expect: { timeout: 5_000 },
  use: {
    baseURL: WEB_URL,
    trace: "retain-on-failure",
    video: "retain-on-failure",
  },
  projects: [
    {
      // T07a: keep the original desktop-1280 project name. The
      // T07a test in home-accessibility.spec.ts self-skips on the
      // mobile projects below, so this project remains the one
      // T07a actually runs in.
      name: "home-desktop-1280",
      use: {
        ...devices["Desktop Chrome"],
        viewport: { width: 1280, height: 720 },
        ...sharedLaunchOptions,
      },
    },
    {
      // T07b: 360px mobile (DOC-03 §10 mobile-first minimum).
      // No test runs in only this project - T07a skips itself
      // here, and every T07b spec runs across all three projects.
      name: "mobile-360",
      use: {
        ...devices["Pixel 5"],
        viewport: { width: 360, height: 640 },
        ...sharedLaunchOptions,
      },
    },
    {
      // T07b: 430px mobile (DOC-03 §10 larger mobile breakpoint).
      name: "mobile-430",
      use: {
        ...devices["Pixel 5"],
        viewport: { width: 430, height: 720 },
        ...sharedLaunchOptions,
      },
    },
  ],
  webServer: [
    {
      command: `node tests/e2e/mock-api-server.mjs`,
      url: MOCK_API_URL,
      timeout: 30_000,
      reuseExistingServer: !isCI,
      stdout: "pipe",
      stderr: "pipe",
    },
    {
      command: `pnpm start --port ${PORT_WEB}`,
      url: WEB_URL,
      timeout: 120_000,
      reuseExistingServer: !isCI,
      stdout: "pipe",
      stderr: "pipe",
      env: {
        API_BASE_URL: `http://127.0.0.1:${PORT_MOCK_API}`,
        NEXT_PUBLIC_API_BASE_URL: `http://127.0.0.1:${PORT_MOCK_API}`,
        PORT: String(PORT_WEB),
      },
    },
  ],
});
