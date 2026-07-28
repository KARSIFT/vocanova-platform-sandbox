// VOC-031-T07a Home accessibility scan.
//
// This is the SINGLE accessibility scan T07a adds. It proves the
// Playwright + axe-core harness is wired end-to-end (Playwright
// spins up, axe-core injects into the page, the WCAG 2.2 AA rule
// set runs, the impact filter narrows to critical/serious, the
// assertion fails the run if any are present, and the test
// summary in CI shows the violation rule + element + impact).
//
// Adding the remaining core-loop screens (Discover, Word Detail,
// Reviews, sentence feedback, Progress, Onboarding, Settings,
// Settings/account) at 360px, 430px, and the desktop width
// covered here is explicitly T07b's scope - see the package's
// tasks.md. Adding a second project/viewport here would silently
// expand T07a's scope and is forbidden by the T07a acceptance
// criterion. The T07b follow-up also added two mobile projects
// (mobile-360, mobile-430) to the Playwright config; this test
// self-skips on those projects so T07a keeps running on exactly
// ONE representative desktop width >=1024px, as its own scope
// required. T07b's home-mobile-accessibility.spec.ts is the
// follow-up that scans /home at 360 and 430.

import { expect, test } from "@playwright/test";

import {
  formatViolations,
  scanForAxeViolations,
} from "./axe-helper.js";

test.describe("Home accessibility (VOC-031-T07a)", () => {
  test("Home renders with zero critical or serious axe-core violations at 1280x720", async ({
    page,
  }, testInfo) => {
    // T07a explicitly covers ONE representative desktop width
    // >=1024px. The mobile projects added by T07b (mobile-360
    // and mobile-430) are out of T07a's scope; the equivalent
    // coverage for those viewports lives in T07b's
    // home-mobile-accessibility.spec.ts.
    test.skip(
      testInfo.project.name !== "home-desktop-1280",
      "T07a scope is one representative desktop width >=1024px; mobile viewports are T07b's scope.",
    );

    // The mock API server (started by playwright.config.ts's
    // webServer entries) handles /api/v1/me + the home page's data
    // reads. The auth-gate middleware reads the same /api/v1/me
    // response, so a 200 there lets the request through to
    // /home instead of redirecting to /signin.
    await page.goto("/home");

    // The "Today's Mission" heading is the most specific
    // signal that the Home page server component has rendered
    // with the mocked data; waiting on it (rather than on a
    // generic network-idle) keeps the scan from running
    // against a still-streaming SSR response and from
    // reporting false-positive violations for half-rendered
    // markup.
    await expect(
      page.getByRole("heading", { name: "Today's Mission", level: 1 }),
    ).toBeVisible();

    const { criticalOrSerious } = await scanForAxeViolations(page);

    expect(
      criticalOrSerious,
      `Expected zero critical or serious axe-core violations on /home; found:\n${formatViolations(
        criticalOrSerious,
      ).join("\n")}`,
    ).toEqual([]);
  });
});
