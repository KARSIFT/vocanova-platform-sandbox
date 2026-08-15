// VOC-073-T00 accessibility scan for /signin.
//
// The sign-in page is a public route (no auth cookie required).
// It renders an OAuth button, a magic-link email form, and
// divider copy. This test exercises the stable initial render
// before any submit/OAuth interaction.

import { expect, test } from "@playwright/test";

import {
  assertKeyboardReachable,
  assertNonColorOnlyFeedback,
  formatViolations,
  scanForAxeViolations,
} from "./axe-helper.js";

test.describe("Sign-in accessibility (VOC-073-T00)", () => {
  test("/signin renders with zero critical/serious axe violations, is keyboard reachable, and uses text-based labels", async ({
    page,
  }, testInfo) => {
    await page.goto("/signin");

    await expect(
      page.getByRole("heading", { name: "Sign in to Vocanova", level: 1 }),
    ).toBeVisible();

    const { criticalOrSerious } = await scanForAxeViolations(page);
    expect(
      criticalOrSerious,
      `Expected zero critical or serious axe-core violations on /signin; found:\n${formatViolations(
        criticalOrSerious,
      ).join("\n")}`,
    ).toEqual([]);

    // Email input and submit button are the two Tab stops when OAuth is
    // disabled (mock /healthz reports oauth_enabled=false by default).
    await assertKeyboardReachable(page, { minFocusable: 2, minTabStops: 2 });

    await assertNonColorOnlyFeedback(page, {
      contextLabel: "/signin",
      requireText: [
        "text=Choose a sign-in method",
        "text=Email address",
        "text=Send sign-in link",
      ],
    });

    expect(testInfo.project.name).toMatch(/^(home-desktop-1280|mobile-360|mobile-430)$/);
  });
});
