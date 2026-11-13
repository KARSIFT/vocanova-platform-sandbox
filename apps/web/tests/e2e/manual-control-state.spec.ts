import { randomUUID } from "node:crypto";
import { expect, test, type Locator } from "@playwright/test";

async function colors(control: Locator) {
  return control.evaluate((element) => {
    const style = getComputedStyle(element);
    return {
      background: style.backgroundColor,
      border: style.borderTopColor,
    };
  });
}

for (const theme of ["light", "dark"] as const) {
  test.describe(`manual selected control states (${theme})`, () => {
    test.beforeEach(async ({ context, page, baseURL }) => {
      if (!baseURL) throw new Error("Missing app URL");
      await context.addCookies([
        { name: "vocanova_session", value: randomUUID(), url: baseURL },
        { name: "vocanova_csrf", value: randomUUID(), url: baseURL },
        { name: "vocanova_theme", value: theme, url: baseURL },
        { name: "e2e_daily_conversation", value: "true", url: baseURL },
      ]);
      await page.addInitScript((preference) => {
        localStorage.setItem("vocanova:theme-preference", preference);
      }, theme);
    });

    test("slower audio visibly marks its pressed state and restores the default", async ({ page }) => {
      await page.goto("/practice");
      const main = page.getByRole("main");
      await main.getByText("More practice", { exact: true }).click();
      await main.getByRole("button", { name: "Start listening practice", exact: true }).click();
      await expect(page).toHaveURL(/\/practice\/session\/[^/]+$/);
      await expect(page.locator("html")).toHaveAttribute("data-theme", theme);
      const slower = main.getByRole("button", { name: "Slower audio", exact: true });
      await expect(slower).toBeVisible();
      await expect(slower).toHaveAttribute("aria-pressed", "false");
      // Keyboard activation and a parked pointer isolate selection from hover.
      // This test changes the speed preference only; it never starts playback.
      await page.mouse.move(0, 0);
      await slower.focus();
      // Leaving the former pointer position can start a hover-out transition.
      // Compare against the settled default rather than an intermediate color.
      await slower.evaluate(async (element) => {
        await Promise.all(element.getAnimations().map((animation) => animation.finished));
      });
      const normal = await colors(slower);
      await slower.press("Enter");
      await expect(slower).toHaveAttribute("aria-pressed", "true");
      await expect.poll(async () => {
        const selected = await colors(slower);
        return {
          background: selected.background !== normal.background,
          border: selected.border !== normal.border,
        };
      }).toEqual({ background: true, border: true });
      await slower.press("Enter");
      await expect(slower).toHaveAttribute("aria-pressed", "false");
      await expect.poll(() => colors(slower)).toEqual(normal);
    });

    test("writing topic selection has distinct background and border from its neighbor", async ({ page }) => {
      await page.goto("/writing?situation=daily-conversation");
      await expect(page.locator("html")).toHaveAttribute("data-theme", theme);
      await page.getByRole("main").locator("summary")
        .filter({ hasText: "Change topic: Daily Conversation" }).press("Enter");
      const choices = page.getByRole("navigation", { name: "Writing situations", exact: true });
      const selected = choices.locator('a[aria-current="page"]');
      const neighbor = choices.locator("a:not([aria-current])").first();
      await expect(selected).toHaveCount(1);
      await expect(selected).toBeVisible();
      await expect(neighbor).toBeVisible();
      await page.mouse.move(0, 0);
      await expect.poll(async () => {
        const [current, other] = await Promise.all([colors(selected), colors(neighbor)]);
        return {
          background: current.background !== other.background,
          border: current.border !== other.border,
        };
      }).toEqual({ background: true, border: true });
    });
  });
}
