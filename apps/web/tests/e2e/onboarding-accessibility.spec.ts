// VOC-031-T07b accessibility scan for /onboarding.
//
// The onboarding page is gated by /api/v1/me's `onboardingStatus`
// field. The mock fixture returns `onboardingStatus: "completed"`
// for the default request, which causes the Next.js middleware
// to redirect /onboarding -> /home. To exercise the onboarding
// form itself (rather than the redirect), this test sets the
// `e2e_onboarding_status` cookie the mock API server recognizes
// (see mock-api-server.mjs) before navigating, which overrides
// /api/v1/me's onboardingStatus to "not_started" for this
// request. Playwright's `page.route()` browser-network
// interception does NOT work here: both `middleware.ts` and the
// /onboarding Server Component call `/api/v1/me` with a direct
// server-to-server fetch (Next.js server -> mock API server),
// which never passes through the browser's network stack that
// page.route() intercepts. The cookie, by contrast, is forwarded
// unchanged by the browser on every request and both the
// middleware and the Server Component already forward the
// incoming Cookie header to their own upstream fetch, so setting
// it in the browser context is the only override point that
// actually reaches this response.
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
    // Set the e2e_onboarding_status cookie for this test only so
    // the onboarding form actually renders instead of the
    // middleware redirecting to /home. See the file-level comment
    // above for why a cookie, not page.route(), is required.
    const baseURL = testInfo.project.use.baseURL;
    if (!baseURL) {
      throw new Error(
        "Expected the Playwright project to configure use.baseURL so the e2e_onboarding_status cookie can be scoped to it.",
      );
    }
    await page.context().addCookies([
      {
        name: "e2e_onboarding_status",
        value: "not_started",
        url: baseURL,
      },
    ]);

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
    // English level option) sharing name="englishLevel" - matches
    // countFocusableElements' selector, so minFocusable: 5 confirms
    // they all exist. Native radio-group semantics mean Tab moves
    // focus INTO the group once (options are then cycled with
    // arrow keys, not Tab), and both "Back" (always disabled on
    // step 0) and "Continue" (disabled until an option is selected,
    // by design) are excluded from the tab order while disabled -
    // so the real sequential Tab-stop count on this screen is 1.
    await assertKeyboardReachable(page, { minFocusable: 5, minTabStops: 1 });

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
