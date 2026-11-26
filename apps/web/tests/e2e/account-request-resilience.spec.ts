import { randomUUID } from "node:crypto";
import { expect, test } from "@playwright/test";

const emailEndpoint = "**/api/v1/settings/email-change-links";
const deletionEndpoint = "**/api/v1/account-deletion-requests";

for (const theme of ["light", "dark"] as const) {
  test.describe(`account request resilience (${theme})`, () => {
    test.beforeEach(async ({ page, context, baseURL }) => {
      await page.addInitScript((value) => localStorage.setItem("vocanova:theme-preference", value), theme);
      await context.addCookies([
        { name: "vocanova_session", value: randomUUID(), url: baseURL! },
        { name: "vocanova_csrf", value: randomUUID(), url: baseURL! },
        { name: "vocanova_theme", value: theme, url: baseURL! },
      ]);
    });

    test("unavailable confirmation delivery leaves the current address and account controls visible", async ({
      page,
      context,
      baseURL,
    }) => {
      if (!baseURL) throw new Error("Expected a configured base URL.");
      await context.addCookies([
        { name: "e2e_email_change_enabled", value: "false", url: baseURL },
        { name: "e2e_magic_link_enabled", value: "false", url: baseURL },
        { name: "e2e_password_enabled", value: "false", url: baseURL },
        { name: "e2e_oauth_enabled", value: "true", url: baseURL },
      ]);

      await page.goto("/settings/account");

      const email = page.getByRole("region", { name: "Sign-in email" });
      await expect(email).toContainText("core-loop-fixture@example.test");
      await expect(email.getByText(/unavailable/i)).toBeVisible();
      await expect(email.getByRole("textbox", { name: "New sign-in email" })).toHaveCount(0);
      await expect(email.getByRole("button", { name: "Send confirmation link" })).toHaveCount(0);
      await expect(page.getByRole("button", { name: "Download my data" })).toBeEnabled();
      await expect(page.getByRole("button", { name: "I want to delete my account" })).toBeEnabled();
    });

    test("email request remains pending until acknowledged, recovers from failure and permits another address", async ({ page }) => {
      let release!: () => void;
      const held = new Promise<void>((resolve) => { release = resolve; });
      let requests = 0;
      await page.route(emailEndpoint, async (route) => {
        requests += 1;
        expect(route.request().method()).toBe("POST");
        expect(route.request().headers()["x-csrf-token"]).toBeTruthy();
        expect(route.request().postDataJSON()).toEqual({ newEmail: "new-learner@example.com" });
        if (requests === 1) {
          await held;
          await route.fulfill({ status: 503, contentType: "application/problem+json", json: { detail: "Confirmation delivery is temporarily unavailable." } });
        } else {
          await route.fulfill({ status: 204, body: "" });
        }
      });
      await page.goto("/settings/account");
        await expect(page.locator("html")).toHaveAttribute("data-theme", theme);
      const address = page.getByLabel("New sign-in email", { exact: true });
      const send = page.getByRole("button", { name: "Send confirmation link", exact: true });
      await expect(send).toBeDisabled();
      await address.fill("new-learner@example.com");
      await send.focus();
      await page.keyboard.press("Enter");
      try {
        await expect(page.getByRole("button", { name: "Sending confirmation link...", exact: true })).toBeDisabled();
        await expect(address).toBeDisabled();
        await expect(page.getByRole("status").filter({ hasText: "If this address can be used" })).toHaveCount(0);
        // Pending UI precedes asynchronous CSRF preparation and request dispatch.
        await expect.poll(() => requests).toBe(1);
      } finally {
        release();
      }
      await expect(page.getByRole("alert").filter({ hasText: "Confirmation delivery is temporarily unavailable." })).toBeVisible();
      await expect(address).toHaveValue("new-learner@example.com");
      await expect(send).toBeEnabled();
      await send.click();
      await expect(page.getByRole("status").filter({ hasText: "If this address can be used, we've sent a confirmation link to new-learner@example.com" })).toBeVisible();
      expect(requests).toBe(2);
      await expect(address).toHaveCount(0);
      const another = page.getByRole("button", { name: "Use a different address" });
      await another.focus();
      await page.keyboard.press("Enter");
      await expect(address).toHaveValue("");
      await expect(send).toBeDisabled();
      await expect(page.getByRole("status").filter({ hasText: "If this address can be used" })).toHaveCount(0);
    });

    test("email request expiry returns to account sign-in without claiming delivery", async ({ page }) => {
      await page.route(emailEndpoint, (route) => route.fulfill({ status: 401, json: { error: "unauthorized" } }));
      await page.goto("/settings/account");
        await expect(page.locator("html")).toHaveAttribute("data-theme", theme);
      await page.getByLabel("New sign-in email", { exact: true }).fill("new-learner@example.com");
      await page.getByRole("button", { name: "Send confirmation link", exact: true }).click();
      await expect(page).toHaveURL(/\/login\?/);
      expect(new URL(page.url()).searchParams.get("returnTo")).toBe("/settings/account");
      await expect(page.getByRole("status").filter({ hasText: "If this address can be used" })).toHaveCount(0);
    });

    test("pending deletion locks confirmation controls and expiry never claims deactivation", async ({ page }) => {
      let release!: () => void;
      const held = new Promise<void>((resolve) => { release = resolve; });
      let requests = 0;
      await page.route(deletionEndpoint, async (route) => {
        requests += 1;
        expect(route.request().method()).toBe("POST");
        expect(route.request().headers()["x-csrf-token"]).toBeTruthy();
        expect(route.request().headers()["idempotency-key"]).toBeTruthy();
        await held;
        await route.fulfill({ status: 401, json: { error: "unauthorized" } });
      });
      await page.goto("/settings/account");
        await expect(page.locator("html")).toHaveAttribute("data-theme", theme);
      await page.getByRole("button", { name: "I want to delete my account" }).click();
      const phrase = page.getByLabel("Type the confirmation phrase");
      await phrase.fill("DELETE MY ACCOUNT");
      const deactivate = page.getByRole("button", { name: "Deactivate my account" });
      await expect(deactivate).toBeDisabled();
      await phrase.fill("delete my account");
      await deactivate.click();
      try {
        await expect(page.getByRole("button", { name: "Deactivating...", exact: true })).toBeDisabled();
        await expect(phrase).toBeDisabled();
        await expect(page.getByRole("button", { name: "Cancel", exact: true })).toBeDisabled();
        // Pending UI precedes asynchronous CSRF preparation and request dispatch.
        await expect.poll(() => requests).toBe(1);
        await expect(page).toHaveURL(/\/settings\/account$/);
      } finally {
        release();
      }
      await expect(page).toHaveURL(/\/login\?/);
      expect(new URL(page.url()).searchParams.get("returnTo")).toBe("/settings/account");
      await expect(page.getByRole("heading", { name: "Your account has been deactivated." })).toHaveCount(0);
    });
  });
}
