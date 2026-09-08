// VOC-031-T07b Home accessibility scan at the two mobile viewports
// (360px, 430px). T07a already covers /home at the 1280x720
// representative desktop width in home-accessibility.spec.ts. The
// Settings-navigation regression also runs at desktop, while the
// T07b-specific scan below remains mobile-only. The T07b acceptance
// criterion calls out that this coverage must add explicit
// keyboard-reachability and non-color-only-feedback assertions on
// top of the axe scan, not only infer them from a clean axe run.

import { expect, test } from "@playwright/test";

import {
  assertKeyboardReachable,
  assertNonColorOnlyFeedback,
  formatViolations,
  scanForAxeViolations,
} from "./axe-helper.js";

test.describe("Home accessibility (VOC-031-T07b mobile)", () => {
  test("keeps a logout failure readable and retryable below the header", async ({
    page,
  }) => {
    await page.goto("/home");
    // Mirror the authenticated mutation contract so clicking Log out reaches
    // the intercepted API failure instead of returning at the CSRF guard.
    await page.context().addCookies([
      {
        name: "vocanova_csrf",
        value: "logout-failure-csrf",
        url: page.url(),
      },
    ]);

    let logoutRequestCount = 0;
    await page.route("**/api/v1/auth/logout", async (route) => {
      logoutRequestCount += 1;
      await route.fulfill({
        status: 503,
        contentType: "application/problem+json",
        body: JSON.stringify({
          detail: "Unable to log out. Please try again.",
        }),
      });
    });

    await page.getByRole("button", { name: "Log out" }).click();

    const alert = page.getByText("Unable to log out. Please try again.", {
      exact: true,
    });
    const header = page.getByRole("banner");
    await expect(alert).toHaveText("Unable to log out. Please try again.");
    expect(logoutRequestCount).toBe(1);
    await expect(page.getByRole("button", { name: "Log out" })).toBeEnabled();

    const [alertBox, headerBox] = await Promise.all([
      alert.boundingBox(),
      header.boundingBox(),
    ]);
    expect(alertBox).not.toBeNull();
    expect(headerBox).not.toBeNull();
    expect(alertBox!.y).toBeGreaterThanOrEqual(
      headerBox!.y + headerBox!.height,
    );

    const documentWidth = await page.evaluate(
      () => document.documentElement.scrollWidth,
    );
    expect(documentWidth).toBeLessThanOrEqual(page.viewportSize()!.width);
  });

  test("lets keyboard users skip the persistent app shell on every authenticated route", async ({
    page,
  }) => {
    const authenticatedRoutes = [
      "/home",
      "/discover",
      "/discover/ordering-at-a-cafe",
      "/discover/ordering-at-a-cafe/pour",
      "/reviews",
      "/progress",
      "/settings",
      "/settings/account",
    ];

    for (const route of authenticatedRoutes) {
      await page.goto(route);

      const skipLink = page.getByRole("link", {
        name: "Skip to main content",
      });
      const main = page.getByRole("main");
      await expect(main).toHaveCount(1);
      await expect(main).toHaveAttribute("id", "main-content");
      await expect(main).toHaveAttribute("tabindex", "-1");
      await expect(skipLink).toHaveAttribute("href", "#main-content");

      await page.keyboard.press("Tab");

      await expect(skipLink).toBeFocused();
      await expect(skipLink).toBeVisible();
      await page.keyboard.press("Enter");

      await expect(main).toBeFocused();
      await expect(page).toHaveURL(new RegExp(`${route}#main-content$`));
    }
  });

  test("Settings is reachable from the authenticated header", async ({
    page,
  }) => {
    for (const route of ["/home", "/discover", "/progress"]) {
      await page.goto(route);
      await expect(page.getByRole("link", { name: "Settings" })).toBeVisible();
      await expect(
        page.getByRole("navigation", { name: "Primary" }).getByRole("link"),
      ).toHaveText(["Home", "Journey", "Progress"]);
    }
    const settingsLink = page.getByRole("link", { name: "Settings" });
    await expect(settingsLink).toBeVisible();
    await expect(settingsLink).toHaveCSS("min-height", "44px");

    const documentWidth = await page.evaluate(
      () => document.documentElement.scrollWidth,
    );
    expect(documentWidth).toBeLessThanOrEqual(page.viewportSize()!.width);

    await page.keyboard.press("Tab");
    await expect(
      page.getByRole("link", { name: "Skip to main content" }),
    ).toBeFocused();
    await page.keyboard.press("Tab");
    await expect(settingsLink).toBeFocused();
    await page.keyboard.press("Enter");
    await expect(page).toHaveURL(/\/settings$/);
    await expect(
      page.getByRole("heading", { name: "Settings", level: 1 }),
    ).toBeVisible();
  });

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
