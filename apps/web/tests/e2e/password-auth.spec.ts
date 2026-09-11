import { expect, test } from "@playwright/test";

const PASSWORD = "a password long enough";

function desktopOnly(testInfo: { project: { name: string } }) {
  test.skip(
    testInfo.project.name !== "home-desktop-1280",
    "Functional password regressions run once at the representative desktop viewport.",
  );
}

test.describe("Password authentication", () => {
  test("offers password and Google sign-in when both capabilities are enabled", async ({
    page,
    context,
    baseURL,
  }, testInfo) => {
    desktopOnly(testInfo);
    if (!baseURL) throw new Error("Expected a configured base URL.");
    await context.addCookies([
      { name: "e2e_oauth_enabled", value: "true", url: baseURL },
    ]);

    await page.goto("/signin");
    await expect(page.getByLabel("Email address")).toBeVisible();
    await expect(page.getByRole("textbox", { name: "Password" })).toBeVisible();
    await expect(
      page.getByRole("button", { name: "Continue with Google" }),
    ).toBeVisible();
    await expect(
      page.getByRole("button", { name: "Send sign-in link" }),
    ).toBeVisible();
  });

  test("offers Google signup without advertising a disabled password form", async ({
    page,
    context,
    baseURL,
  }, testInfo) => {
    desktopOnly(testInfo);
    if (!baseURL) throw new Error("Expected a configured base URL.");
    await context.addCookies([
      { name: "e2e_oauth_enabled", value: "true", url: baseURL },
      { name: "e2e_password_enabled", value: "false", url: baseURL },
    ]);

    await page.goto("/signup?returnTo=%2Freviews%3Fmode%3Ddue");
    await expect(
      page.getByRole("button", { name: "Continue with Google" }),
    ).toBeVisible();
    await expect(page.getByRole("textbox", { name: "Password" })).toHaveCount(0);
    await expect(page.getByText("Continue with Google to sign in or create your account.")).toBeVisible();
  });

  test("validates signup locally, then gives a generic check-email result without storing the password", async ({
    page,
  }, testInfo) => {
    desktopOnly(testInfo);
    await page.goto("/signup");
    await page.getByLabel("Email address").fill("learner@example.test");
    await page.getByRole("textbox", { name: "Password" }).fill("short");
    await page.getByRole("button", { name: "Create account" }).click();
    await expect(
      page
        .getByRole("alert")
        .filter({ hasText: "Use a password between 15 and 128 characters." }),
    ).toBeVisible();

    await page.getByRole("textbox", { name: "Password" }).fill(PASSWORD);
    await page.getByRole("button", { name: "Create account" }).click();
    await expect(
      page.getByRole("status").filter({ hasText: /If this address can create an account/ }),
    ).toBeVisible();
    expect(page.url()).not.toContain(PASSWORD);
    const browserStorage = await page.evaluate(() =>
      JSON.stringify({ localStorage, sessionStorage }),
    );
    expect(browserStorage).not.toContain(PASSWORD);
  });

  test("verifies a signup link once and explains invalid or replayed links without signing in", async ({
    page,
    context,
  }, testInfo) => {
    desktopOnly(testInfo);
    let verificationRequests = 0;
    page.on("request", (request) => {
      if (request.url().endsWith("/api/v1/auth/password/signups/verify")) {
        verificationRequests += 1;
      }
    });
    await page.goto("/auth/password/verify?token=valid-signup-token");
    await expect(
      page.locator('meta[name="referrer"][content="no-referrer"]'),
    ).toHaveCount(1);
    await expect(page.getByRole("button", { name: "Verify email" })).toBeVisible();
    expect(verificationRequests).toBe(0);
    await page.getByRole("button", { name: "Verify email" }).click();
    await expect(
      page.getByRole("status").filter({ hasText: /password is ready to use/ }),
    ).toBeVisible();
    expect((await context.cookies()).some((cookie) => cookie.name === "vocanova_session")).toBe(false);
    expect(page.url()).toMatch(/\/auth\/password\/verify$/);

    await page.goto("/auth/password/verify?token=valid-signup-token");
    await page.getByRole("button", { name: "Verify email" }).click();
    await expect(
      page
        .getByRole("alert")
        .filter({ hasText: /invalid, expired, or has already been used/ }),
    ).toBeVisible();

    await page.goto("/auth/password/verify?token=not-a-real-token");
    await page.getByRole("button", { name: "Verify email" }).click();
    await expect(
      page
        .getByRole("alert")
        .filter({ hasText: /invalid, expired, or has already been used/ }),
    ).toBeVisible();
  });

  test("handles wrong and unavailable password login, then establishes a cookie session on success", async ({
    page,
    context,
  }, testInfo) => {
    desktopOnly(testInfo);
    await page.goto("/signin?returnTo=%2Freviews%3Fmode%3Ddue");
    await page.getByLabel("Email address").fill("wrong-password@example.test");
    await page.getByRole("textbox", { name: "Password" }).fill(PASSWORD);
    const wrongResponse = page.waitForResponse(
      (response) =>
        response.url().endsWith("/api/v1/auth/password/login") &&
        response.status() === 401,
    );
    await page.getByRole("button", { name: "Sign in" }).click();
    await wrongResponse;
    await expect(
      page.getByRole("alert").filter({ hasText: /couldn't sign you in/ }),
    ).toBeVisible();

    await page.getByLabel("Email address").fill("unavailable@example.test");
    const unavailableResponse = page.waitForResponse(
      (response) =>
        response.url().endsWith("/api/v1/auth/password/login") &&
        response.status() === 503,
    );
    await page.getByRole("button", { name: "Sign in" }).click();
    await unavailableResponse;
    await expect(
      page.getByRole("alert").filter({ hasText: /temporarily unavailable/ }),
    ).toBeVisible();

    await page.getByLabel("Email address").fill("rate-limited@example.test");
    const rateLimitedResponse = page.waitForResponse(
      (response) =>
        response.url().endsWith("/api/v1/auth/password/login") &&
        response.status() === 429,
    );
    await page.getByRole("button", { name: "Sign in" }).click();
    await rateLimitedResponse;
    await expect(
      page
        .getByRole("alert")
        .filter({ hasText: /Too many password sign-in attempts/ }),
    ).toBeVisible();

    await page.getByLabel("Email address").fill("learner@example.test");
    await page.getByRole("button", { name: "Sign in" }).click();
    await expect(page).toHaveURL(/\/reviews\?mode=due$/);
    expect((await context.cookies()).some((cookie) => cookie.name === "vocanova_session")).toBe(true);
  });

  test("requests a generic reset email and resets once without auto-signing in", async ({
    page,
    context,
  }, testInfo) => {
    desktopOnly(testInfo);
    await page.goto("/auth/password/reset");
    await page.getByLabel("Email address").fill("learner@example.test");
    await page.getByRole("button", { name: "Send password reset link" }).click();
    await expect(page.getByRole("status").filter({ hasText: /sent instructions/ })).toBeVisible();

    await page.goto("/auth/password/reset?token=valid-reset-token");
    await expect(
      page.locator('meta[name="referrer"][content="no-referrer"]'),
    ).toHaveCount(1);
    await page.getByRole("textbox", { name: "Password" }).fill(PASSWORD);
    await page.getByRole("button", { name: "Save new password" }).click();
    await expect(page.getByRole("status").filter({ hasText: /password is ready/ })).toBeVisible();
    expect((await context.cookies()).some((cookie) => cookie.name === "vocanova_session")).toBe(false);
    expect(page.url()).toMatch(/\/auth\/password\/reset$/);

    await page.goto("/auth/password/reset?token=valid-reset-token");
    await page.getByRole("textbox", { name: "Password" }).fill(PASSWORD);
    await page.getByRole("button", { name: "Save new password" }).click();
    await expect(
      page
        .getByRole("alert")
        .filter({ hasText: /invalid, expired, or has already been used/ }),
    ).toBeVisible();
    await page.goto("/auth/password/reset?token=invalid-reset-token");
    await page.getByRole("textbox", { name: "Password" }).fill(PASSWORD);
    await page.getByRole("button", { name: "Save new password" }).click();
    await expect(
      page
        .getByRole("alert")
        .filter({ hasText: /invalid, expired, or has already been used/ }),
    ).toBeVisible();
  });
});
