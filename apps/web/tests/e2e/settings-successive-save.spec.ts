import { randomUUID } from "node:crypto";

import { expect, test } from "@playwright/test";
import type { TestInfo } from "@playwright/test";

const INITIAL_DISPLAY_NAME = "Core Loop Fixture";

async function authenticateForSettings(testInfo: TestInfo) {
  const baseURL = testInfo.project.use.baseURL;
  if (!baseURL) {
    throw new Error("Expected Playwright to configure a base URL.");
  }
  return {
    baseURL,
    cookies: [
      {
        name: "vocanova_session",
        value: `settings-successive-save-${randomUUID()}`,
        url: baseURL,
      },
      {
        name: "vocanova_csrf",
        value: `settings-successive-save-csrf-${randomUUID()}`,
        url: baseURL,
      },
    ],
  };
}

test.describe("Settings successive saves", () => {
  test("persists a reversion after a successful save", async ({
    page,
    context,
  }, testInfo) => {
    test.skip(
      testInfo.project.name !== "home-desktop-1280",
      "This functional regression runs once at the representative desktop viewport.",
    );

    const { cookies } = await authenticateForSettings(testInfo);
    await context.addCookies(cookies);
    await page.goto("/settings");

    const displayName = page.getByRole("textbox", { name: "Display name" });
    await displayName.fill("Saved name");
    await page.getByRole("button", { name: "Save settings" }).click();
    await expect(
      page.getByText("Your settings have been saved."),
    ).toBeVisible();

    await displayName.fill(INITIAL_DISPLAY_NAME);
    await page.getByRole("button", { name: "Save settings" }).click();
    await expect(
      page.getByText("Your settings have been saved."),
    ).toBeVisible();

    await page.reload();
    await expect(displayName).toHaveValue(INITIAL_DISPLAY_NAME);
  });

  test("does not advance the baseline when a save fails", async ({
    page,
    context,
  }, testInfo) => {
    test.skip(
      testInfo.project.name !== "home-desktop-1280",
      "This functional regression runs once at the representative desktop viewport.",
    );

    const { cookies } = await authenticateForSettings(testInfo);
    await context.addCookies(cookies);

    let failNextPatch = true;
    await page.route("**/api/v1/settings", async (route) => {
      if (route.request().method() === "PATCH" && failNextPatch) {
        failNextPatch = false;
        await route.fulfill({
          status: 500,
          contentType: "application/json",
          body: JSON.stringify({ error: "temporary_failure" }),
        });
        return;
      }
      await route.continue();
    });

    await page.goto("/settings");
    const displayName = page.getByRole("textbox", { name: "Display name" });
    await displayName.fill("Retry after failure");
    await page.getByRole("button", { name: "Save settings" }).click();
    await expect(page.getByText("HTTP 500", { exact: true })).toBeVisible();

    await page.getByRole("button", { name: "Save settings" }).click();
    await expect(
      page.getByText("Your settings have been saved."),
    ).toBeVisible();

    await page.reload();
    await expect(displayName).toHaveValue("Retry after failure");
  });
});
