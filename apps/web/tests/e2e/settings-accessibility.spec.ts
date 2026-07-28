// VOC-031-T07b accessibility scans for /settings and
// /settings/account.
//
// /settings is the screen that renders every editable Settings
// field (daily review target, review rhythm, app language,
// notifications, marketing emails, display name) and the
// "Manage account" link to the deeper account sub-screen.
// /settings/account renders the email-change form, the current
// email, and the account-deletion form. Both are the highest-
// stakes a11y surfaces T07b covers: a learner who cannot
// operate the deletion flow because colour-only feedback
// signals the typed-phrase confirmation, or a learner whose
// screen reader cannot reach the email-change input, would be
// locked out of their own account. The keyboard-reachability
// and non-color-only assertions are therefore stricter here
// than on the read-only screens.

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

  test("/settings/account renders with zero critical/serious axe violations, is keyboard reachable, and uses text-based state", async ({
    page,
  }) => {
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
  });
});
