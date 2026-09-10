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
    await expect(page.getByRole("alert")).toContainText(
      "confirmation link is incomplete",
    );
  });
});
