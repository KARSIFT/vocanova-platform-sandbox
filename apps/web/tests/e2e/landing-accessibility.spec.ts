// VOC-073-T01 accessibility scan for / (landing).
//
// The root page is a public placeholder with no auth cookie
// required. Remediation stays within accessibility fixes on
// the current copy (VOC-073-DEP-02) — not a marketing rewrite.

import { expect, test } from "@playwright/test";

import {
  assertKeyboardReachable,
  assertNonColorOnlyFeedback,
  formatViolations,
  scanForAxeViolations,
} from "./axe-helper.js";

test.describe("Landing accessibility (VOC-073-T01)", () => {
  test("/ renders with zero critical/serious axe violations, is keyboard reachable, and uses text-based content", async ({
    page,
  }, testInfo) => {
    await page.goto("/");

    await expect(
      page.getByText("Vocanova web foundation is running."),
    ).toBeVisible();

    const { criticalOrSerious } = await scanForAxeViolations(page);
    expect(
      criticalOrSerious,
      `Expected zero critical or serious axe-core violations on /; found:\n${formatViolations(
        criticalOrSerious,
      ).join("\n")}`,
    ).toEqual([]);

    // The landing placeholder is read-only with no interactive
    // controls — the correct posture for a static status page.
    await assertKeyboardReachable(page, { minFocusable: 0 });

    await assertNonColorOnlyFeedback(page, {
      contextLabel: "/",
      requireText: ["text=Vocanova web foundation is running."],
    });

    expect(testInfo.project.name).toMatch(/^(home-desktop-1280|mobile-360|mobile-430)$/);
  });
});
