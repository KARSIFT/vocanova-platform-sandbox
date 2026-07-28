// VOC-031-T07a Playwright config.
//
// This config deliberately stays minimal. It exists to prove the
// accessibility harness (Playwright + axe-core + CI wiring) is
// operational end-to-end on ONE core-loop screen at ONE
// representative desktop width. Extending it to the full
// 360px/430px/desktop matrix and the remaining screens is T07b's
// scope - see specs/changes/VOC-031-begin-milestone-p5-integrated-core-loop/tasks.md.
//
// webServer runs two processes: this harness's mock API server
// (so the Home server component can render without the real
// Go/Postgres backend) and the Next.js production build (so the
// test is running the same bundle CI will ship, not the dev
// server's hot-reload variant). The mock API server is started
// first because the Next.js server's SSR path makes
// /api/v1/me + /api/v1/user-words + /api/v1/reviews/due +
// /api/v1/daily-mission requests during /home rendering.

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
      name: "home-desktop-1280",
      // T07a explicitly covers ONE representative desktop width
      // >=1024px. 360px and 430px are T07b's scope.
      use: {
        ...devices["Desktop Chrome"],
        viewport: { width: 1280, height: 720 },
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
