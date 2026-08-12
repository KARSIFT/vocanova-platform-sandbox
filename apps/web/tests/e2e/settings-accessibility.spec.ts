// VOC-031-T07b accessibility scan for /settings.
//
// /settings is the screen that renders every editable Settings
// field (daily review target, review rhythm, app language,
// notifications, marketing emails, display name) and the
// "Manage account" link to the deeper account sub-screen.
// /settings/account coverage lives in settings-account-accessibility.spec.ts
// (VOC-073-T03).

import { expect, test } from "@playwright/test";

import {
  assertKeyboardReachable,
  assertNonColorOnlyFeedback,
  formatViolations,
  scanForAxeViolations,
} from "./axe-helper.js";

test.describe("Settings accessibility (VOC-031-T07b)", () => {
  test("/settings renders with zero critical/serious axe violations, is keyboard reachable, and uses text-based state", async ({
    page,
  }) => {
    await page.goto("/settings");

    await expect(
      page.getByRole("heading", { name: "Settings", level: 1 }),
    ).toBeVisible();

    const { criticalOrSerious } = await scanForAxeViolations(page);
    expect(
      criticalOrSerious,
      `Expected zero critical or serious axe-core violations on /settings; found:\n${formatViolations(
        criticalOrSerious,
      ).join("\n")}`,
    ).toEqual([]);

    // /settings has 8 daily-review-target radios + 3 review-rhythm
    // radios + 2 checkbox toggles + 1 display-name input + 1 save
    // button + 2 links (Back to Home, Manage account) = 17+
    // focusable elements. Use a conservative floor.
    await assertKeyboardReachable(page, { minFocusable: 10 });

    await assertNonColorOnlyFeedback(page, {
      contextLabel: "/settings",
      requireText: [
        "text=Daily review target",
        "text=Review rhythm",
        "text=Manage account",
        "text=Vocanova default",
        "text=Faster reminders",
      ],
    });
  });
});
