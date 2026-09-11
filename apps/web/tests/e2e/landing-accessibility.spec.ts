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
  test("keeps the learning action visible on arrival without horizontal scrolling", async ({
    page,
  }) => {
    await page.goto("/");
    const brand = await page
      .getByRole("link", { name: "VocaNova home" })
      .boundingBox();
    expect(brand).not.toBeNull();
    expect(brand!.height).toBeGreaterThanOrEqual(44);
    const action = page.getByRole("link", {
      name: "Start learning",
      exact: true,
    });
    await expect(action).toBeVisible();
    const box = await action.boundingBox();
    expect(box).not.toBeNull();
    expect(box!.height).toBeGreaterThanOrEqual(44);
    expect(box!.y).toBeGreaterThanOrEqual(0);
    expect(box!.y + box!.height).toBeLessThanOrEqual(
      page.viewportSize()!.height,
    );
    expect(
      await page.evaluate(() => document.documentElement.scrollWidth),
    ).toBeLessThanOrEqual(page.viewportSize()!.width);
  });

  test("/ renders the product entry point with zero critical/serious axe violations and keyboard-reachable actions", async ({
    page,
  }, testInfo) => {
    await page.goto("/");

    await expect(
      page.getByRole("heading", {
        name: "Learn the words you will actually use.",
      }),
    ).toBeVisible();
    await expect(
      page.getByRole("link", { name: "Start learning" }),
    ).toHaveAttribute("href", "/login");

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
      requireText: ["text=Discover", "text=Remember", "text=Use"],
    });

    expect(testInfo.project.name).toMatch(
      /^(home-desktop-1280|mobile-360|mobile-430)$/,
    );
  });
});
