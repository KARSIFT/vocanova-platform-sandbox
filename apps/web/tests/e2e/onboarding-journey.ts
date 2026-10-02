import { expect, type Page } from "@playwright/test";

export async function completeOnboardingIfRedirected(page: Page): Promise<void> {
  if (!page.url().includes("/onboarding")) {
    return;
  }

  await expect(
    page.getByRole("heading", { name: "Welcome to Vocanova", level: 1 }),
  ).toBeVisible();

  await page.getByRole("radio", { name: /A2/ }).check();
  await page.getByRole("button", { name: "Continue" }).click();

  await expect(
    page.getByRole("heading", { name: /What's your native language\?/ }),
  ).toBeVisible();
  await page.getByRole("textbox", { name: "Native language" }).fill("es");
  await page.getByRole("button", { name: "Continue" }).click();

  await expect(
    page.getByRole("radiogroup", {
      name: /What's your main reason for learning\?/,
    }),
  ).toBeVisible();
  await page.getByRole("radio", { name: "General growth" }).check();
  await page.getByRole("button", { name: "Continue" }).click();

  await expect(
    page.getByRole("radiogroup", { name: /Where will you use English most\?/ }),
  ).toBeVisible();
  await page.getByRole("radio", { name: "Daily life" }).check();
  await page.getByRole("button", { name: "Continue" }).click();

  await expect(
    page.getByRole("heading", { name: "Daily review target" }),
  ).toBeVisible();
  await page
    .getByRole("radiogroup", { name: "Daily review target" })
    .getByText("15", { exact: true })
    .click();
  // Register the response wait before submitting. A click alone does not mean
  // the save committed, and leaving the page can interrupt the pending request.
  const [response] = await Promise.all([
    page.waitForResponse(
      (candidate) =>
        new URL(candidate.url()).pathname === "/api/v1/onboarding" &&
        candidate.request().method() === "POST",
    ),
    page.getByRole("button", { name: "Finish setup" }).click(),
  ]);
  expect(response.status(), "onboarding completion must succeed").toBe(200);

  // Let the app handle navigation only after success. On failure, leave the
  // current form and its feedback intact for the journey's diagnostics.
  await expect(page).toHaveURL(/\/home(\?|$)/);
}
