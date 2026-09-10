import { expect, test } from "@playwright/test";

test.describe("documented route compatibility", () => {
  test("/login renders the passwordless sign-in experience", async ({ page }) => {
    await page.goto("/login");
    await expect(
      page.getByRole("heading", { name: "Sign in to Vocanova" }),
    ).toBeVisible();
  });

  test("/magic-link renders the magic-link result experience", async ({ page }) => {
    await page.goto("/magic-link");
    await expect(page.getByRole("heading", { name: "Sign in link" })).toBeVisible();
  });

  test("documented review routes render the active review screen", async ({
    page,
  }) => {
    await page.goto("/review");
    await expect(page).toHaveURL(/\/review$/);
    await expect(page.getByRole("heading", { name: "Review" })).toBeVisible();

    await page.goto("/review/session");
    await expect(page).toHaveURL(/\/review\/session$/);
    await expect(page.getByRole("heading", { name: "Review" })).toBeVisible();
  });
});
