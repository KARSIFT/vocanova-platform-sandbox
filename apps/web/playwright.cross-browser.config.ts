import { defineConfig, devices } from "@playwright/test";
import baseConfig from "./playwright.config";

// Run each engine independently so existing exact project-name guards retain
// their desktop/phone meaning. Reports and artifacts remain engine-specific.
// VOCANOVA_TEST_BROWSER=webkit pnpm exec playwright test --config playwright.cross-browser.config.ts
const engine = process.env.VOCANOVA_TEST_BROWSER;
if (engine !== "chromium" && engine !== "firefox" && engine !== "webkit") {
  throw new Error("Set VOCANOVA_TEST_BROWSER to chromium, firefox or webkit.");
}

const browserName: "chromium" | "firefox" | "webkit" = engine;
const desktop =
  devices[
    engine === "webkit"
      ? "Desktop Safari"
      : engine === "firefox"
        ? "Desktop Firefox"
        : "Desktop Chrome"
  ];
const phone =
  engine === "webkit"
    ? devices["iPhone 13"]
    : engine === "chromium"
      ? devices["Pixel 5"]
      : desktop;

export default defineConfig(baseConfig, {
  outputDir: `./test-results/cross-browser-${engine}`,
  metadata: { browserEngine: engine },
  projects: [
    {
      name: "home-desktop-1280",
      use: { ...desktop, browserName, viewport: { width: 1280, height: 720 } },
    },
    ...([360, 430] as const).map((width) => ({
      name: `mobile-${width}`,
      use: {
        ...phone,
        browserName,
        viewport: { width, height: width === 360 ? 640 : 720 },
        // Firefox supports touch and narrow viewports, but not isMobile emulation.
        isMobile: engine !== "firefox",
        hasTouch: true,
      },
    })),
  ],
});
