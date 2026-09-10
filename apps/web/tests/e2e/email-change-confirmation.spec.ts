import { expect, test } from "@playwright/test";

test.describe("email-change confirmation", () => {
  test("a signed-in learner can complete the emailed confirmation link", async ({
    page,
  }) => {
    await page.goto("/");
    await page.evaluate(() => {
      document.cookie = "vocanova_csrf=email-change-test; Path=/; SameSite=Lax";
    });

    await page.goto("/auth/email-change?token=test-token");

    await expect(page.getByRole("status")).toContainText(
      "Your sign-in email is now",
    );
    await expect(
      page.getByRole("link", { name: "Return to account settings" }),
    ).toHaveAttribute("href", "/settings/account");
  });

  test("an incomplete confirmation link fails safely", async ({ page }) => {
    await page.goto("/auth/email-change");
    await expect(page.locator("main").getByRole("alert")).toContainText(
      "confirmation link is incomplete",
    );
  });

  test("an invalid token does not send an authenticated learner into a sign-in loop", async ({
    page,
  }) => {
    await page.goto("/");
    await page.evaluate(() => {
      document.cookie = "vocanova_csrf=email-change-test; Path=/; SameSite=Lax";
    });

    await page.goto("/auth/email-change?token=invalid-token");

    await expect(page.locator("main").getByRole("alert")).toContainText(
      "invalid or has expired",
    );
    await expect(page.getByRole("link", { name: "Sign in" })).toHaveCount(0);
  });

  test("a signed-out confirmation flow preserves the token and requires magic-link sign-in", async ({
    page,
  }) => {
    await page.goto("/");
    await page.evaluate(() => {
      document.cookie = "e2e_unauthenticated=1; Path=/; SameSite=Lax";
      document.cookie = "vocanova_csrf=email-change-test; Path=/; SameSite=Lax";
    });

    await page.goto("/auth/email-change?token=test-token");

    const signInLink = page.getByRole("link", { name: "Sign in" });
    await expect(signInLink).toHaveAttribute(
      "href",
      /\/login\?.*returnTo=.*email-change.*magicOnly=1/,
    );
  });
});
