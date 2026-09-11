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
        "text=No password needed",
        "text=Email address",
        "text=Send sign-in link",
      ],
    });

    expect(testInfo.project.name).toMatch(
      /^(home-desktop-1280|mobile-360|mobile-430)$/,
    );
  });

  test("shows a check-email state with an address edit path and a bounded resend", async ({
    page,
  }) => {
    await page.goto("/signin");
    await page.getByLabel("Email address").fill("learner@example.test");
    await page.getByRole("button", { name: "Send sign-in link" }).click();

    await expect(
      page.getByRole("heading", { name: "Check your email", level: 2 }),
    ).toBeVisible();
    await expect(page.getByText("learner@example.test")).toBeVisible();
    await expect(
      page.getByRole("button", { name: /Resend available in \d+s/ }),
    ).toBeDisabled();

    await page.getByRole("button", { name: "Use a different email" }).click();
    await expect(page.getByLabel("Email address")).toHaveValue(
      "learner@example.test",
    );
  });

  test("explains when every sign-in method is unavailable", async ({
    page,
    context,
  }, testInfo) => {
    const baseURL = testInfo.project.use.baseURL;
    if (!baseURL) {
      throw new Error("Expected a configured base URL.");
    }
    await context.addCookies([
      { name: "e2e_magic_link_enabled", value: "false", url: baseURL },
    ]);

    await page.goto("/signin");
    const unavailable = page.getByText(
      "Sign-in is temporarily unavailable. Please try again later.",
    );
    await expect(unavailable).toBeVisible();
    await expect(unavailable).toHaveAttribute("role", "alert");
    await expect(page.getByLabel("Email address")).toHaveCount(0);
  });

  test("offers Google when an email-only recovery route cannot send email", async ({
    page,
    context,
  }, testInfo) => {
    const baseURL = testInfo.project.use.baseURL;
    if (!baseURL) {
      throw new Error("Expected a configured base URL.");
    }
    await context.addCookies([
      { name: "e2e_magic_link_enabled", value: "false", url: baseURL },
      { name: "e2e_oauth_enabled", value: "true", url: baseURL },
    ]);

    await page.goto("/login?magicOnly=1");
    await expect(
      page.getByText("Email sign-in is unavailable right now."),
    ).toBeVisible();
    await expect(
      page.getByRole("button", { name: "Continue with Google" }),
    ).toBeVisible();
    await expect(
      page.getByRole("link", { name: "use the standard sign-in page" }),
    ).toBeVisible();
  });

  test("continues a Google sign-in to its safe original destination", async ({
    page,
    context,
  }, testInfo) => {
    const baseURL = testInfo.project.use.baseURL;
    if (!baseURL) {
      throw new Error("Expected a configured base URL.");
    }
    await context.addCookies([
      { name: "e2e_oauth_enabled", value: "true", url: baseURL },
    ]);
    await page.route("**/api/v1/auth/oauth/google/start", async (route) => {
      await route.fulfill({
        status: 200,
        contentType: "application/json",
        body: JSON.stringify({ url: `${baseURL}/home` }),
      });
    });

    await page.goto("/signin?returnTo=%2Freviews%3Fmode%3Ddue");
    await page.getByRole("button", { name: "Continue with Google" }).click();
    await expect(page).toHaveURL(/\/reviews\?mode=due$/);
  });
});
