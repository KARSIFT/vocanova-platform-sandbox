// VOC-073-T02 accessibility scan for /auth/magic.
//
// Scan states (VOC-073-DEP-01): the success path redirects to /home
// before a post-success scan is practical, so this spec exercises
// only the stable incomplete-link error surface — navigating to
// /auth/magic without `token` and `email` query params renders
// "This sign-in link is incomplete. Please request a new one."
// plus a "Back to sign in" link. No auth cookie required.

import { expect, test } from "@playwright/test";

import {
  assertKeyboardReachable,
  assertNonColorOnlyFeedback,
  formatViolations,
  scanForAxeViolations,
} from "./axe-helper.js";

test.describe("Auth magic-link accessibility (VOC-073-T02)", () => {
  test("/auth/magic incomplete link renders with zero critical/serious axe violations, is keyboard reachable, and uses text-based error feedback", async ({
    page,
  }, testInfo) => {
    await page.goto("/auth/magic");

    await expect(
      page.getByText(
        "This sign-in link is incomplete. Please request a new one.",
      ),
    ).toBeVisible();

    const { criticalOrSerious } = await scanForAxeViolations(page);
    expect(
      criticalOrSerious,
      `Expected zero critical or serious axe-core violations on /auth/magic (incomplete link); found:\n${formatViolations(
        criticalOrSerious,
      ).join("\n")}`,
    ).toEqual([]);

    // Error state exposes a single "Back to sign in" link.
    await assertKeyboardReachable(page, { minFocusable: 1, minTabStops: 1 });

    await assertNonColorOnlyFeedback(page, {
      contextLabel: "/auth/magic (incomplete link)",
      requireText: [
        "text=Sign in link",
        "text=This sign-in link is incomplete. Please request a new one.",
        "text=Back to sign in",
      ],
    });

    expect(testInfo.project.name).toMatch(/^(home-desktop-1280|mobile-360|mobile-430)$/);
  });
});
