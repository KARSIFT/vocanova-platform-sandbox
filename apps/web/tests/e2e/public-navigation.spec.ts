import { randomUUID } from "node:crypto";
import { expect, test } from "@playwright/test";
import { pressTab } from "./keyboard-navigation.js";

test("a remembered browser can return from the landing page without signing in again", async ({ page, context, baseURL }) => {
  await context.addCookies([{
    name: "vocanova_session",
    value: randomUUID(),
    url: baseURL!,
    httpOnly: true,
    expires: Math.floor(Date.now() / 1000) + 30 * 24 * 60 * 60,
  }]);

  await page.goto("/");
  await page.getByRole("link", { name: "Start learning", exact: true }).click();
  await expect(page).toHaveURL(`${baseURL}/home`);
  await expect(page.getByRole("heading", { level: 1, name: "Today" })).toBeVisible();

  await page.goto("/signin?returnTo=%2Fprogress");
  await expect(page).toHaveURL(`${baseURL}/progress`);
  await expect(page.getByRole("heading", { level: 1, name: "Progress" })).toBeVisible();
});

for (const theme of ["light", "dark"] as const) {
  test(`invalid deactivation date omits the deadline and sign-in recovery works (${theme})`, async ({ page, baseURL }) => {
    await page.addInitScript(value => localStorage.setItem("vocanova:theme-preference", value), theme);
    await page.goto("/account-deactivated?purgeAfter=not-a-date");
    await expect(page.locator("html")).toHaveAttribute("data-theme", theme);
    const confirmation = page.getByRole("status");
    await expect(confirmation.getByRole("heading", { name: "Your account has been deactivated.", exact: true })).toBeVisible();
    await expect(confirmation).toContainText("You can sign in again before then to reactivate");
    await expect(confirmation.getByText(/permanently anonymize/)).toHaveCount(0);
    await expect(confirmation).not.toContainText("Invalid Date");

    const signIn = confirmation.getByRole("link", { name: "Go to sign in", exact: true });
    await pressTab(page);
    await expect(signIn).toBeFocused();
    await page.keyboard.press("Enter");
    await expect(page).toHaveURL(`${baseURL}/login`);
    await expect(page.getByRole("heading", { level: 1, name: "Sign in to Vocanova", exact: true })).toBeVisible();
    await expect(page.locator("html")).toHaveAttribute("data-theme", theme);
  });
  test(`landing navigation activates each destination by keyboard (${theme})`, async ({ page, baseURL }) => {
    await page.addInitScript(value => localStorage.setItem("vocanova:theme-preference", value), theme);
    for (const [name, destination, heading] of [
      ["VocaNova home", "/", "Learn the words you will actually use."],
      ["Sign in", "/login", "Sign in to Vocanova"],
      ["Start learning", "/login", "Sign in to Vocanova"],
      ["How it works", "/#how-it-works", "Learn the words you will actually use."],
      ["Help and support", "/help", "How can we help?"],
    ] as const) {
      await page.goto("/");
      await expect(page.locator("html")).toHaveAttribute("data-theme", theme);
      const link = page.getByRole("link", { name, exact: true });
      await link.focus();
      await page.keyboard.press("Enter");
      await expect(page).toHaveURL(`${baseURL}${destination}`);
      await expect(page.getByRole("heading", { level: 1, name: heading, exact: true })).toBeVisible();
      if (destination.includes("#")) await expect(page.locator("#how-it-works")).toBeInViewport();
    }
  });
  test(`Help public links work by keyboard and support address is accessible (${theme})`, async ({ page, context, baseURL }) => {
    await page.addInitScript(value => localStorage.setItem("vocanova:theme-preference", value), theme);
    await page.goto("/help");
    await expect(page.locator("html")).toHaveAttribute("data-theme", theme);
    const support = page.locator('a[href^="mailto:"]');
    await expect(support).toHaveCount(1);
    const address = await support.getAttribute("href");
    expect(address).toMatch(/^mailto:[^\s@]+@[^\s@]+\.[^\s@]+$/);
    await expect(support).toHaveText(address!.slice(7));
    await support.focus();
    await expect(support).toBeFocused();
    // Do not launch a mail client or send mail as a test side effect.
    const brand = page.getByRole("link", { name: "VocaNova home", exact: true });
    await brand.focus();
    await page.keyboard.press("Enter");
    await expect(page).toHaveURL(`${baseURL}/`);
    await expect(page.getByRole("heading", { level: 1, name: "Learn the words you will actually use.", exact: true })).toBeVisible();
    await page.goto("/help");
    await page.getByText("Why can’t I sign in?", { exact: true }).click();
    await page.getByRole("link", { name: "Go to sign in", exact: true }).focus();
    await page.keyboard.press("Enter");
    await expect(page).toHaveURL(/\/(login|signin)(?:\?.*)?$/);
    await expect(page.getByRole("heading", { level: 1, name: "Sign in to Vocanova", exact: true })).toBeVisible();
    await context.addCookies([{ name: "vocanova_session", value: randomUUID(), url: baseURL! }]);
    await page.goto("/help");
    await page.getByRole("link", { name: "Go to learning", exact: true }).focus();
    await page.keyboard.press("Enter");
    await expect(page).toHaveURL(`${baseURL}/home`);
    await expect(page.getByRole("heading", { name: "Today", exact: true })).toBeVisible();
  });
}
