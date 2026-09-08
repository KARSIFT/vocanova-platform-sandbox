import { randomUUID } from "node:crypto";

import { expect, test } from "@playwright/test";

test.describe("Account deletion confirmation", () => {
  test("keeps the server-confirmed deactivation message visible after the session is revoked", async ({
    page,
    context,
  }, testInfo) => {
    const baseURL = testInfo.project.use.baseURL;
    if (!baseURL) {
      throw new Error("Expected Playwright to configure a base URL.");
    }

    await context.addCookies([
      {
        name: "vocanova_session",
        value: `account-deletion-${randomUUID()}`,
        url: baseURL,
      },
      {
        name: "vocanova_csrf",
        value: `account-deletion-csrf-${randomUUID()}`,
        url: baseURL,
      },
    ]);

    await page.goto("/settings/account");
    await expect
      .poll(() => page.evaluate(() => document.cookie))
      .toContain("vocanova_csrf=");
    const trigger = page.getByRole("button", {
      name: "I want to delete my account",
    });
    const confirmation = page.getByLabel("Type the confirmation phrase");
    await trigger.click();
    await expect(confirmation).toBeFocused();
    await page.getByRole("button", { name: "Cancel" }).click();
    await expect(trigger).toBeFocused();

    await trigger.click();
    await expect(confirmation).toBeFocused();
    await confirmation.fill("delete my account");

    // A recoverable API failure must keep the confirmation controls and the
    // typed phrase available for a safe retry; it must not collapse to an
    // error-only view in this irreversible flow.
    const deletionEndpoint = "**/api/v1/account-deletion-requests";
    await page.route(deletionEndpoint, async (route) => {
      await route.fulfill({
        status: 500,
        contentType: "application/json",
        body: JSON.stringify({ error: "temporary_failure" }),
      });
    });
    const deactivate = page.getByRole("button", {
      name: "Permanently deactivate my account",
    });
    await deactivate.click();
    await expect(page.getByRole("alert")).toBeVisible();
    await expect(confirmation).toHaveValue("delete my account");
    await expect(deactivate).toBeVisible();
    await page.unroute(deletionEndpoint);

    await deactivate.click();

    await expect(
      page.getByRole("heading", { name: "Your account has been deactivated." }),
    ).toBeVisible();
    await expect(page).toHaveURL(/\/account-deactivated\?purgeAfter=/);
    const purgeAfter = new URL(page.url()).searchParams.get("purgeAfter");
    expect(purgeAfter).not.toBeNull();
    const expectedPurgeAfter = new Intl.DateTimeFormat("en-US", {
      dateStyle: "long",
      timeZone: "UTC",
    }).format(new Date(purgeAfter!));
    await expect(page.getByText(expectedPurgeAfter)).toBeVisible();
    await expect(
      page.getByRole("link", { name: "Go to sign in" }),
    ).toBeVisible();
    await expect(page.getByRole("button", { name: "Log out" })).toHaveCount(0);

    const cookieNames = (await context.cookies()).map((cookie) => cookie.name);
    expect(cookieNames).not.toContain("vocanova_session");
    expect(cookieNames).not.toContain("vocanova_csrf");
  });
});
