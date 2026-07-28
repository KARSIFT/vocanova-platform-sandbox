// VOC-031-T07b accessibility scan for /onboarding.
//
// The onboarding page is gated by /api/v1/me's `onboardingStatus`
// field. The mock fixture returns `onboardingStatus: "completed"`
// for the default request, which causes the Next.js middleware
// to redirect /onboarding -> /home. To exercise the onboarding
// form itself (rather than the redirect), this test uses
// Playwright's route interception to override the /api/v1/me
// response for this test only and return
// `onboardingStatus: "not_started"`, which lets the form render.
//
// The onboarding form's accessibility is the most important
// part of the screen to check: every field, every option, and
// the multi-step "Continue" / "Back" flow must be reachable
// and well-labelled. This test goes through the first step
// (English level) so the multi-step flow's accessibility is
// exercised, not just the form's static layout.

import { expect, test } from "@playwright/test";

import {
  assertKeyboardReachable,
  assertNonColorOnlyFeedback,
  formatViolations,
  scanForAxeViolations,
} from "./axe-helper.js";

test.describe("Onboarding accessibility (VOC-031-T07b)", () => {
  test("/onboarding renders with zero critical/serious axe violations, is keyboard reachable, and uses text-based state on step 1", async ({
    page,
  }, testInfo) => {
    // Override the /api/v1/me response for this test only so
    // the onboarding form actually renders. See the file-level
    // comment above for why.
    await page.route("**/api/v1/me", (route) => {
      route.fulfill({
        status: 200,
        contentType: "application/json",
        body: JSON.stringify({
          email: "accessibility-fixture@example.test",
          displayName: "Accessibility Fixture",
          emailVerifiedAt: "2026-01-01T00:00:00Z",
          onboardingStatus: "not_started",
        }),
      });
    });

    await page.goto("/onboarding");

    await expect(
      page.getByRole("heading", { name: "Welcome to Vocanova", level: 1 }),
    ).toBeVisible();

    const { criticalOrSerious } = await scanForAxeViolations(page);
    expect(
      criticalOrSerious,
      `Expected zero critical or serious axe-core violations on /onboarding; found:\n${formatViolations(
        criticalOrSerious,
      ).join("\n")}`,
    ).toEqual([]);

    // Step 0 (English level) renders 5 radio inputs (one per
    // English level option) + 1 "Continue" button. The "Back"
    // button is disabled on step 0, so it is not focusable.
    await assertKeyboardReachable(page, { minFocusable: 6 });

    await assertNonColorOnlyFeedback(page, {
      contextLabel: "/onboarding",
      requireText: [
        "text=Welcome to Vocanova",
        "text=Continue",
        // The English-level step's actual <legend>:
        "text=How would you describe your English?",
        "text=A1 — Beginner",
      ],
    });

    // Sanity: the test ran on a supported project, otherwise
    // the project name is asserted in the report.
    expect(testInfo.project.name).toMatch(/^(home-desktop-1280|mobile-360|mobile-430)$/);
  });
});
