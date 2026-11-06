import { randomUUID } from "node:crypto";
import { expect, test, type Page } from "@playwright/test";
import { expectNavigationFeedback } from "./navigation-feedback-readiness";

async function expectReminderReached(page: Page) {
  await expect(page.getByRole("button", { name: "Download calendar reminder", exact: true })).toBeEnabled();
  const heading = page.getByRole("heading", { name: "Calendar reminder", exact: true });
  await expect(heading).toBeInViewport({ ratio: 1 });
  await expect.poll(() => heading.evaluate((element) => {
    const box = element.getBoundingClientRect();
    const top = document.elementFromPoint(box.x + box.width / 2, box.y + box.height / 2);
    return !!top && (element.contains(top) || top.contains(element));
  })).toBe(true);
}

for (const theme of ["light", "dark"] as const) {
  test.describe(`calendar fragment navigation (${theme})`, () => {
    test.beforeEach(async ({ page, context, baseURL }) => {
      await context.addCookies([{ name: "vocanova_session", value: randomUUID(), url: baseURL! }]);
      await page.addInitScript(value => localStorage.setItem("vocanova:theme-preference", value), theme);
    });

    for (const entry of [
      { path: "/plan", label: "Set a calendar reminder", hash: "practice-reminder" },
      { path: "/help", label: "Set up a calendar reminder", hash: "calendar-reminder-heading" },
    ]) {
      test(`${entry.path} reaches the reminder with continuous navigation feedback`, async ({ page, context, baseURL }) => {
        await page.goto(entry.path);
        await expect(page.locator("html")).toHaveAttribute("data-theme", theme);
        if (entry.path === "/help") {
          await page.locator("summary").filter({ hasText: "How do I set a daily reminder?" }).press("Enter");
        }
        // Delay only Settings page data so identity/layout reads can stream the
        // real loading shell before the calendar fragment target is mounted.
        await context.addCookies([{ name: "e2e_settings_delay_ms", value: "3000", url: baseURL! }]);
        const link = page.getByRole("link", { name: entry.label, exact: true });
        await expectNavigationFeedback(page, {
          loadingStatusId: "settings-loading-status", heading: "Settings",
        }, async () => {
          await link.focus();
          await page.keyboard.press("Enter");
        });
        await expect(page).toHaveURL(`${baseURL}/settings#${entry.hash}`);
        await expectReminderReached(page);
      });
    }

    test("direct fragment loads reach the reminder while ordinary settings preserves the top", async ({ page, context, baseURL }) => {
      await page.goto("/settings");
      await expect(page.locator("html")).toHaveAttribute("data-theme", theme);
      await expect(page.getByRole("button", { name: "Download calendar reminder", exact: true })).toBeEnabled();
      await expect(page.getByRole("heading", { level: 1, name: "Settings", exact: true })).toBeInViewport();
      expect(await page.evaluate(() => scrollY)).toBe(0);
      await context.addCookies([{ name: "e2e_settings_delay_ms", value: "3000", url: baseURL! }]);
      for (const hash of ["practice-reminder", "calendar-reminder-heading"]) {
        // Leave Settings first so each goto is a fresh document navigation,
        // not the browser's already-mounted same-document hash behavior.
        await page.goto("/help");
        await page.goto(`/settings#${hash}`);
        await expectReminderReached(page);
      }
    });
  });
}
