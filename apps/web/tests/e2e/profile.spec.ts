import { randomUUID } from "node:crypto";

import { expect, test } from "@playwright/test";
import type { TestInfo } from "@playwright/test";

function desktopOnly(testInfo: TestInfo) {
  test.skip(
    testInfo.project.name !== "home-desktop-1280",
    "Functional profile regressions run once at the representative desktop viewport.",
  );
}

async function authenticatedCookies(testInfo: TestInfo) {
  const baseURL = testInfo.project.use.baseURL;
  if (!baseURL) throw new Error("Expected a configured base URL.");
  return {
    baseURL,
    cookies: [
      {
        name: "vocanova_session",
        value: `profile-${randomUUID()}`,
        url: baseURL,
      },
      {
        name: "vocanova_csrf",
        value: `profile-csrf-${randomUUID()}`,
        url: baseURL,
      },
    ],
  };
}

test.describe("Profile and password security", () => {
  test("does not describe an unverified server email as verified", async ({
    page,
    context,
  }, testInfo) => {
    desktopOnly(testInfo);
    const { baseURL, cookies } = await authenticatedCookies(testInfo);
    await context.addCookies([
      ...cookies,
      { name: "e2e_email_verified", value: "false", url: baseURL },
    ]);
    await page.goto("/settings/profile");

    await expect(page.getByRole("heading", { name: "Email", level: 2 })).toBeVisible();
    await expect(
      page.getByText("This email has not been verified yet."),
    ).toBeVisible();
  });

  test("saves the display name and preserves it after reload", async ({
    page,
    context,
  }, testInfo) => {
    desktopOnly(testInfo);
    const { cookies } = await authenticatedCookies(testInfo);
    await context.addCookies(cookies);
    await page.goto("/settings/profile");

    const displayName = page.getByRole("textbox", { name: "Display name" });
    await displayName.fill("Profile learner");
    await page.getByRole("button", { name: "Save profile" }).click();
    await expect(
      page.getByRole("status").filter({ hasText: "Your profile has been saved." }),
    ).toBeVisible();
    await page.reload();
    await expect(displayName).toHaveValue("Profile learner");
    await expect(page.getByText("core-loop-fixture@example.test")).toBeVisible();
  });

  test("routes an expired profile update back to sign-in without claiming success", async ({
    page,
    context,
  }, testInfo) => {
    desktopOnly(testInfo);
    const { baseURL, cookies } = await authenticatedCookies(testInfo);
    await context.addCookies(cookies);
    await page.goto("/settings/profile");
    await page.getByRole("textbox", { name: "Display name" }).fill("Needs auth");
    await context.addCookies([
      { name: "e2e_unauthenticated", value: "1", url: baseURL },
    ]);

    await page.getByRole("button", { name: "Save profile" }).click();
    await expect(page).toHaveURL(/\/login\?returnTo=%2Fsettings%2Fprofile/);
    await expect(page.getByText("Your session expired. Sign in again to continue.")).toBeVisible();
  });

  test("offers a Google or magic-link account an email-confirmed add-password action", async ({
    page,
    context,
  }, testInfo) => {
    desktopOnly(testInfo);
    const { baseURL, cookies } = await authenticatedCookies(testInfo);
    await context.addCookies([
      ...cookies,
      { name: "e2e_has_password", value: "false", url: baseURL },
    ]);
    await page.goto("/settings/account");
    await expect(page.getByRole("heading", { name: "Password", level: 2 })).toBeVisible();
    await expect(page.getByText("No password is set.")).toBeVisible();
    await expect(page.getByLabel("Email address")).toHaveValue("core-loop-fixture@example.test");
    await page.getByRole("button", { name: "Email link to add a password" }).click();
    await expect(page.getByRole("status").filter({ hasText: /sent instructions/ })).toBeVisible();
  });
});
