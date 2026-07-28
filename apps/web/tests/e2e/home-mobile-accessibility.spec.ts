// VOC-031-T07b Home accessibility scan at the two mobile viewports
// (360px, 430px). T07a already covers /home at the 1280x720
// representative desktop width in home-accessibility.spec.ts; this
// file extends the coverage to the mobile breakpoints DOC-03 §10
// requires. The T07b acceptance criterion calls out that this
// coverage must add explicit keyboard-reachability and
// non-color-only-feedback assertions on top of the axe scan, not
// only infer them from a clean axe run.

import { expect, test } from "@playwright/test";

import {
  assertKeyboardReachable,
  assertNonColorOnlyFeedback,
  formatViolations,
  scanForAxeViolations,
} from "./axe-helper.js";

test.describe("Home accessibility (VOC-031-T07b mobile)", () => {
  test("Home renders with zero critical/serious axe violations, is keyboard reachable, and uses text-based state at 360 / 430", async ({
    page,
  }, testInfo) => {
    // T07b's home-mobile scan is intentionally scoped to the
    // mobile projects. The 1280x720 desktop scan is T07a's
    // home-accessibility.spec.ts.
    test.skip(
      testInfo.project.name === "home-desktop-1280",
      "T07b mobile scan is scoped to mobile-360 / mobile-430; desktop is T07a's home-accessibility.spec.ts.",
    );

    await page.goto("/home");

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

    // The Home page renders at least two links ("Go to Journey",
    // "Start review") and the sentence-feedback form's submit
    // button. Use a conservative floor.
    await assertKeyboardReachable(page, { minFocusable: 2 });

    // Non-color-only feedback: the mission progress text, the
    // streak text, and the "words due today" line all carry
    // their state in text, not just color. The empty-saved-words
    // message is also text.
    await assertNonColorOnlyFeedback(page, {
      contextLabel: "/home",
      requireText: [
        "text=Review target",
        "text=words reviewed today",
        "text=-day streak",
        "text=words due today",
      ],
    });
  });
});
