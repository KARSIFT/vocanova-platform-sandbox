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
    await page.evaluate(() => {
      sessionStorage.setItem(
        "vocanova:sentence-feedback-draft:learner:word_detail:failed-logout",
        JSON.stringify({
          attemptId: "failed-logout",
          savedAt: Date.now(),
          sentence: "I practise every day.",
          source: "word_detail",
        }),
      );
    });

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

    page.once("dialog", (dialog) => dialog.accept());
    await page.getByRole("button", { name: "Log out" }).click();

    const alert = page.getByText(
      "We couldn't sign you out. Please try again.",
      {
        exact: true,
      },
    );
    const header = page.getByRole("banner");
    await expect(alert).toHaveText(
      "We couldn't sign you out. Please try again.",
    );
    expect(logoutRequestCount).toBe(1);
    await expect(page.getByRole("button", { name: "Log out" })).toBeEnabled();
    await expect
      .poll(() =>
        page.evaluate(() =>
          sessionStorage.getItem(
            "vocanova:sentence-feedback-draft:learner:word_detail:failed-logout",
          ),
        ),
      )
      .not.toBeNull();

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

  test("keeps an unsent draft when the learner cancels logout", async ({
    page,
  }) => {
    await page.goto("/home");
    await page.evaluate(() => {
      sessionStorage.setItem(
        "vocanova:sentence-feedback-draft:learner:word_detail:attempt",
        JSON.stringify({
          attemptId: "attempt",
          savedAt: Date.now(),
          sentence: "I practise every day.",
          source: "word_detail",
        }),
      );
    });
    let logoutRequestCount = 0;
    await page.route("**/api/v1/auth/logout", async (route) => {
      logoutRequestCount += 1;
      await route.continue();
    });
    page.once("dialog", (dialog) => dialog.dismiss());

    await page.getByRole("button", { name: "Log out" }).click();

    expect(logoutRequestCount).toBe(0);
    await expect
      .poll(() =>
        page.evaluate(() =>
          sessionStorage.getItem(
            "vocanova:sentence-feedback-draft:learner:word_detail:attempt",
          ),
        ),
      )
      .not.toBeNull();
  });

  test("clears an unsent draft after confirmed server-side logout", async ({
    page,
  }) => {
    await page.goto("/home");
    await page
      .context()
      .addCookies([
        { name: "vocanova_csrf", value: "logout-draft-csrf", url: page.url() },
      ]);
    await page.evaluate(() => {
      sessionStorage.setItem(
        "vocanova:sentence-feedback-draft:learner:word_detail:attempt",
        JSON.stringify({
          attemptId: "attempt",
          savedAt: Date.now(),
          sentence: "I practise every day.",
          source: "word_detail",
        }),
      );
    });
    page.once("dialog", (dialog) => dialog.accept());

    await page.getByRole("button", { name: "Log out" }).click();
    await expect(page).toHaveURL(/\/login\?signedOut=1$/);
    await expect
      .poll(() =>
        page.evaluate(() =>
          sessionStorage.getItem(
            "vocanova:sentence-feedback-draft:learner:word_detail:attempt",
          ),
        ),
      )
      .toBeNull();
  });

  test("restores a missing CSRF cookie before logging out", async ({
    page,
    context,
  }, testInfo) => {
    const baseURL = testInfo.project.use.baseURL;
    if (!baseURL) {
      throw new Error("Expected a configured base URL.");
    }
    // Let the shell observe an existing token, so this test isolates recovery
    // initiated by the later logout click.
    await context.addCookies([
      { name: "vocanova_csrf", value: "initial-csrf", url: baseURL },
    ]);
    await page.goto("/home");
    await expect(page.getByRole("button", { name: "Log out" })).toBeVisible();

    let recoveryRequests = 0;
    await page.route("**/api/v1/me", async (route) => {
      recoveryRequests += 1;
      await route.continue();
    });
    await page.evaluate(() => {
      document.cookie = "vocanova_csrf=; Max-Age=0; path=/";
    });

    let logoutCSRFHeader: string | undefined;
    await page.route("**/api/v1/auth/logout", async (route) => {
      logoutCSRFHeader = route.request().headers()["x-csrf-token"];
      await route.continue();
    });

    await page.getByRole("button", { name: "Log out" }).click();
    await expect(page).toHaveURL(/\/login\?signedOut=1$/);
    expect(recoveryRequests).toBe(1);
    expect(logoutCSRFHeader).toBeTruthy();
  });

  test("returns to sign-in when CSRF recovery confirms an expired session", async ({
    page,
    context,
  }, testInfo) => {
    const baseURL = testInfo.project.use.baseURL;
    if (!baseURL) {
      throw new Error("Expected a configured base URL.");
    }
    await context.addCookies([
      { name: "vocanova_csrf", value: "initial-csrf", url: baseURL },
    ]);
    await page.goto("/home");
    await expect(page.getByRole("button", { name: "Log out" })).toBeVisible();

    await page.route("**/api/v1/me", async (route) => {
      await route.fulfill({
        status: 401,
        contentType: "application/problem+json",
        body: JSON.stringify({ detail: "authentication required" }),
      });
    });
    await page.evaluate(() => {
      document.cookie = "vocanova_csrf=; Max-Age=0; path=/";
    });

    await page.getByRole("button", { name: "Log out" }).click();
    await expect(page).toHaveURL(
      /\/login\?returnTo=%2Fhome&reason=session-expired$/,
    );
    await expect(
      page.getByText("Your session expired. Sign in again to continue."),
    ).toBeVisible();
  });

  test("lets keyboard users skip the persistent app shell on every authenticated route", async ({
    page,
  }) => {
    const authenticatedRoutes = [
      "/home",
      "/discover",
      "/discover/ordering-at-a-cafe",
      "/discover/ordering-at-a-cafe/pour",
      "/words",
      "/review",
      "/review/session",
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
      const documentWidth = await page.evaluate(
        () => document.documentElement.scrollWidth,
      );
      expect(documentWidth).toBeLessThanOrEqual(page.viewportSize()!.width);

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
    // Desktop includes the three primary destinations between the brand and
    // account controls; mobile keeps those destinations in the bottom bar.
    for (let step = 0; step < 6; step += 1) {
      if (
        await settingsLink.evaluate(
          (element) => element === document.activeElement,
        )
      )
        break;
      await page.keyboard.press("Tab");
    }
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

    await assertKeyboardReachable(page, { minFocusable: 2 });

    // Non-color-only feedback: the mission progress text, the
    // streak text, and the "words due today" line all carry
    // their state in text, not just color. The empty-saved-words
    // message is also text.
    await assertNonColorOnlyFeedback(page, {
      contextLabel: "/home",
      requireText: ["text=reviews complete", "text=STREAK", "text=In progress"],
    });
    await expect(
      page.getByRole("progressbar", { name: "Today’s mission progress" }),
    ).toHaveAttribute("aria-valuenow", "0");
  });
});
