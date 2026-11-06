import { randomUUID } from "node:crypto";
import { expect, test, type Page } from "@playwright/test";
import { formatViolations, scanForAxeViolations } from "./axe-helper";

// These frontend lifecycle fixtures are isolated per test. The shared mock's
// fixed verification tokens are single-use across the whole browser run.
// They establish UI behavior, not email delivery or server credential storage.
const PASSWORD = "a synthetic password long enough";
const validationMessage = "Use a password between 15 and 128 characters.";

async function checkLayout(page: Page) {
  expect(await page.evaluate(() => document.documentElement.scrollWidth))
    .toBeLessThanOrEqual(page.viewportSize()!.width);
  const scan = await scanForAxeViolations(page);
  expect(scan.criticalOrSerious, formatViolations(scan.criticalOrSerious).join("\n")).toEqual([]);
}

for (const theme of ["light", "dark"] as const) {
  test.describe(`responsive password lifecycle (${theme})`, () => {
    test.afterEach(async ({ page }) => {
      await expect(page.locator("html")).toHaveAttribute("data-theme", theme);
    });

    test.beforeEach(async ({ page, context, baseURL }) => {
      await page.addInitScript((preference) => localStorage.setItem("vocanova:theme-preference", preference), theme);
      await context.addCookies([{ name: "vocanova_theme", value: theme, url: baseURL! }]);
    });

    test("signup validates, retries failure, and waits for confirmation", async ({ page }) => {
      let requests = 0;
      let release!: () => void;
      const pending = new Promise<void>((resolve) => { release = resolve; });
      await page.route("**/api/v1/auth/password/signups", async (route) => {
        requests += 1;
        expect(route.request().postDataJSON()).toEqual({
          email: "learner@example.test", password: PASSWORD, displayName: "Mobile learner",
        });
        if (requests === 1) {
          await route.fulfill({ status: 503, json: { error: "unavailable" } });
        } else {
          await pending;
          await route.fulfill({ status: 204 });
        }
      });
      await page.goto("/signup");
      await page.getByLabel("Email address").fill("learner@example.test");
      await page.getByLabel("Display name").fill("  Mobile learner  ");
      const password = page.getByLabel("Password", { exact: true });
      const submit = page.getByRole("button", { name: "Create account", exact: true });
      for (const invalid of ["short", "a".repeat(129)]) {
        await password.fill(invalid);
        await submit.click();
        await expect(page.getByRole("alert").filter({ hasText: validationMessage })).toBeVisible();
        expect(requests).toBe(0);
      }
      await password.fill(PASSWORD);
      await page.getByRole("button", { name: "Show password" }).press("Enter");
      await expect(password).toHaveAttribute("type", "text");
      await page.getByRole("button", { name: "Hide password" }).press("Space");
      await expect(password).toHaveAttribute("type", "password");
      await submit.click();
      await expect(page.getByRole("alert").filter({ hasText: /Account creation is temporarily unavailable/ })).toBeVisible();
      await expect(password).toHaveValue(PASSWORD);
      await checkLayout(page);
      await submit.press("Enter");
      try {
        await expect(page.getByRole("button", { name: "Creating account..." })).toBeDisabled();
        await expect(password).toBeDisabled();
        await expect(page.getByLabel("Email address")).toBeDisabled();
        await expect.poll(() => requests).toBe(2);
        await expect(page.getByRole("status").filter({ hasText: /verification link/ })).toHaveCount(0);
      } finally { release(); }
      await expect(page.getByRole("status").filter({ hasText: /If this address can create an account/ })).toBeVisible();
      expect(await page.evaluate(() => JSON.stringify({ localStorage, sessionStorage }))).not.toContain(PASSWORD);
      expect(page.url()).not.toContain(encodeURIComponent(PASSWORD));
    });

    test("verification requires consent, rejects replay and leads to working sign-in", async ({ page, context }) => {
      const token = randomUUID();
      let requests = 0;
      let release!: () => void;
      const pending = new Promise<void>((resolve) => { release = resolve; });
      await page.route("**/api/v1/auth/password/signups/verify", async (route) => {
        requests += 1;
        expect(route.request().postDataJSON()).toEqual({ token });
        if (requests === 1) { await pending; await route.fulfill({ status: 204 }); }
        else await route.fulfill({ status: 401, json: { error: "unauthorized" } });
      });
      await page.goto(`/auth/password/verify?token=${token}`);
      await expect(page.locator('meta[name="referrer"][content="no-referrer"]')).toHaveCount(1);
      const verify = page.getByRole("button", { name: "Verify email", exact: true });
      await expect(verify).toBeVisible();
      expect(requests).toBe(0);
      await verify.press("Enter");
      try {
        await expect(page.getByRole("button", { name: "Verifying..." })).toBeDisabled();
        await expect.poll(() => requests).toBe(1);
      } finally { release(); }
      await expect(page.getByRole("status").filter({ hasText: /password is ready to use/ })).toBeVisible();
      await expect(page).toHaveURL(/\/auth\/password\/verify$/);
      expect((await context.cookies()).some((cookie) => cookie.name === "vocanova_session")).toBe(false);
      await checkLayout(page);
      await page.getByRole("link", { name: "Sign in", exact: true }).press("Enter");
      await expect(page.getByLabel("Email address")).toBeVisible();
      await page.goto(`/auth/password/verify?token=${token}`);
      await page.getByRole("button", { name: "Verify email", exact: true }).click();
      await expect(page.getByRole("alert").filter({ hasText: /invalid, expired, or has already been used/ })).toBeVisible();
      await page.getByRole("link", { name: "Create account", exact: true }).press("Enter");
      await expect(page).toHaveURL(/\/signup$/);
      await page.goto("/signin?returnTo=%2Freviews%3Fmode%3Ddue");
      await page.getByLabel("Email address").fill("wrong-password@example.test");
      await page.getByLabel("Password", { exact: true }).fill(PASSWORD);
      await page.getByRole("button", { name: "Sign in", exact: true }).click();
      await expect(page.getByRole("alert").filter({ hasText: /couldn't sign you in/ })).toBeVisible();
      await page.getByLabel("Email address").fill("learner@example.test");
      await page.getByRole("button", { name: "Sign in", exact: true }).press("Enter");
      await expect(page).toHaveURL(/\/reviews\?mode=due$/);
      expect((await context.cookies()).some((cookie) => cookie.name === "vocanova_session")).toBe(true);
    });

    test("password reset recovers from failure, clears token and rejects reuse", async ({ page, context }) => {
      const token = randomUUID();
      let requests = 0;
      let release!: () => void;
      const pending = new Promise<void>((resolve) => { release = resolve; });
      await page.route("**/api/v1/auth/password/resets", async (route) => {
        requests += 1;
        expect(route.request().postDataJSON()).toEqual({ token, password: PASSWORD });
        if (requests === 1) await route.fulfill({ status: 503, json: { error: "unavailable" } });
        else if (requests === 2) { await pending; await route.fulfill({ status: 204 }); }
        else await route.fulfill({ status: 401, json: { error: "unauthorized" } });
      });
      await page.goto(`/auth/password/reset?token=${token}`);
      await expect(page.locator('meta[name="referrer"][content="no-referrer"]')).toHaveCount(1);
      const password = page.getByLabel("Password", { exact: true });
      await password.fill(PASSWORD);
      await page.getByRole("button", { name: "Save new password" }).click();
      await expect(page.getByRole("alert").filter({ hasText: /Password reset is temporarily unavailable/ })).toBeVisible();
      await expect(password).toHaveValue(PASSWORD);
      await checkLayout(page);
      await page.getByRole("button", { name: "Save new password" }).press("Enter");
      try {
        await expect(page.getByRole("button", { name: "Saving password..." })).toBeDisabled();
        await expect(password).toBeDisabled();
        await expect.poll(() => requests).toBe(2);
      } finally { release(); }
      await expect(page.getByRole("status").filter({ hasText: /Your password is ready/ })).toBeVisible();
      await expect(page).toHaveURL(/\/auth\/password\/reset$/);
      expect((await context.cookies()).some((cookie) => cookie.name === "vocanova_session")).toBe(false);
      await page.getByRole("link", { name: "Sign in", exact: true }).press("Enter");
      await expect(page.getByLabel("Email address")).toBeVisible();
      await page.goto(`/auth/password/reset?token=${token}`);
      await page.getByLabel("Password", { exact: true }).fill(PASSWORD);
      await page.getByRole("button", { name: "Save new password" }).click();
      await expect(page.getByRole("alert").filter({ hasText: /invalid, expired, or has already been used/ })).toBeVisible();
    });

    test("account add-password email handles failure and pending confirmation", async ({ page, context, baseURL }) => {
      await context.addCookies([
        { name: "vocanova_session", value: randomUUID(), url: baseURL! },
        { name: "vocanova_csrf", value: randomUUID(), url: baseURL! },
        { name: "e2e_has_password", value: "false", url: baseURL! },
      ]);
      let requests = 0;
      let release!: () => void;
      const pending = new Promise<void>((resolve) => { release = resolve; });
      await page.route("**/api/v1/auth/password/reset-requests", async (route) => {
        requests += 1;
        expect(route.request().postDataJSON()).toEqual({ email: "core-loop-fixture@example.test" });
        if (requests === 1) await route.fulfill({ status: 503, json: { error: "unavailable" } });
        else { await pending; await route.fulfill({ status: 204 }); }
      });
      await page.goto("/settings/account");
      const region = page.getByRole("region", { name: "Password", exact: true });
      await expect(region).toContainText("No password is set.");
      const send = region.getByRole("button", { name: "Email link to add a password" });
      await send.click();
      await expect(region.getByRole("alert")).toContainText("Password email is temporarily unavailable.");
      await send.press("Enter");
      try {
        await expect(region.getByRole("button", { name: "Sending..." })).toBeDisabled();
        await expect(region.getByLabel("Email address")).toBeDisabled();
        await expect(region.getByRole("status")).toHaveCount(0);
        await expect.poll(() => requests).toBe(2);
      } finally { release(); }
      await expect(region.getByRole("status")).toContainText("sent instructions");
      await expect(region.getByRole("button")).toHaveCount(0);
      // Requesting an email must never claim the account already has a password.
      await expect(region).toContainText("No password is set.");
      await checkLayout(page);
      expect(requests).toBe(2);
    });
  });
}
