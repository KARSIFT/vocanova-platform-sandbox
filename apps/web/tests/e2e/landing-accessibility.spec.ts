// VOC-073-T01 accessibility scan for / (landing).
//
// The root page is the public product entry point and needs no auth cookie.

import { expect, test } from "@playwright/test";

import {
  assertKeyboardReachable,
  assertNonColorOnlyFeedback,
  formatViolations,
  scanForAxeViolations,
} from "./axe-helper.js";

test.describe("Landing accessibility (VOC-073-T01)", () => {
  test("/ renders the product entry point with zero critical/serious axe violations and keyboard-reachable actions", async ({
    page,
  }, testInfo) => {
    await page.goto("/");

    await expect(
      page.getByRole("heading", { name: "Words for the moments that matter." }),
    ).toBeVisible();
    await expect(page.getByRole("link", { name: "Start learning" })).toHaveAttribute(
      "href",
      "/login",
    );

    const { criticalOrSerious } = await scanForAxeViolations(page);
    expect(
      criticalOrSerious,
      `Expected zero critical or serious axe-core violations on /; found:\n${formatViolations(
        criticalOrSerious,
      ).join("\n")}`,
    ).toEqual([]);

    await assertKeyboardReachable(page, { minFocusable: 3 });

    await assertNonColorOnlyFeedback(page, {
      contextLabel: "/",
      requireText: [
        "text=Learn in context",
        "text=Remember for longer",
        "text=Put words into practice",
      ],
    });

    expect(testInfo.project.name).toMatch(/^(home-desktop-1280|mobile-360|mobile-430)$/);
  });
});
