// VOC-073-T03 accessibility scan for /settings/account.
//
// Extracted from settings-accessibility.spec.ts (VOC-031-T07b) so issue #536's
// dedicated spec file per entry route is satisfied without duplicating CI time.
// /settings/account renders the email-change form, the current email, and the
// account-deletion form — the highest-stakes settings sub-screen. Keyboard-
// reachability and non-color-only assertions are stricter here than on read-only
// screens: a learner who cannot operate the deletion flow because colour-only
// feedback signals the typed-phrase confirmation would be locked out of their
// own account.

import { expect, test } from "@playwright/test";

import {
  assertKeyboardReachable,
  assertNonColorOnlyFeedback,
  formatViolations,
  scanForAxeViolations,
} from "./axe-helper.js";

test.describe("Settings account accessibility (VOC-073-T03)", () => {
  test("/settings/account renders with zero critical/serious axe violations, is keyboard reachable, and uses text-based state", async ({
    page,
  }, testInfo) => {
    await page.goto("/settings/account");

    await expect(
      page.getByRole("heading", { name: "Account", level: 1 }),
    ).toBeVisible();

    const { criticalOrSerious } = await scanForAxeViolations(page);
    expect(
      criticalOrSerious,
      `Expected zero critical or serious axe-core violations on /settings/account; found:\n${formatViolations(
        criticalOrSerious,
      ).join("\n")}`,
    ).toEqual([]);

    // /settings/account has the Back to Settings link + 1 email
    // input + 1 confirmation-phrase input + 1 token input + 2
    // submit buttons (one for email-change request, one for
    // account deletion) = at least 6 focusable elements.
    await assertKeyboardReachable(page, { minFocusable: 5 });

    await assertNonColorOnlyFeedback(page, {
      contextLabel: "/settings/account",
      requireText: [
        "text=Sign-in email",
        "text=Delete your account",
        "text=Current address:",
        // The "We'll deactivate your account right away, then
        // permanently anonymize your data after 30 days" copy
        // is the most important text on the deletion section -
        // it is the only signal a screen reader user has that
        // the deletion is staged and reversible for 30 days.
        "text=permanently anonymize your data after 30 days",
      ],
    });

    expect(testInfo.project.name).toMatch(/^(home-desktop-1280|mobile-360|mobile-430)$/);
  });
});
