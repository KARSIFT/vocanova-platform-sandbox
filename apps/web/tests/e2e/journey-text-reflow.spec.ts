import { randomUUID } from "node:crypto";
import { expect, test } from "@playwright/test";

for (const theme of ["light", "dark"] as const) {
  test(`Journey range and map remain readable without overlapping at enlarged text (${theme})`, async ({ page, context, baseURL }) => {
    await context.addCookies([
      { name: "vocanova_session", value: randomUUID(), url: baseURL! },
      { name: "vocanova_csrf", value: randomUUID(), url: baseURL! },
    ]);
    await page.addInitScript((value) => localStorage.setItem("vocanova:theme-preference", value), theme);
    await page.goto("/discover?band=1");
    await expect(page.getByRole("heading", { name: "Journey", exact: true })).toBeVisible();
    await expect(page.locator("html")).toHaveAttribute("data-theme", theme);
    for (const size of ["100%", "200%"]) {
      await page.addStyleTag({ content: `html { font-size: ${size} !important; }` });
      await expect.poll(async () => {
        const range = await page.getByRole("region", { name: "Words 1–100", exact: true }).boundingBox();
        const title = await page.getByRole("heading", { name: "Words 1–100", exact: true }).boundingBox();
        const map = await page.getByRole("region", { name: "Your word map", exact: true }).boundingBox();
        if (!range || !title || !map) return false;
        const titleContained = title.x >= range.x && title.x + title.width <= range.x + range.width + 1;
        const separated = range.x + range.width <= map.x + 1 || map.x + map.width <= range.x + 1 || range.y + range.height <= map.y + 1 || map.y + map.height <= range.y + 1;
        const rootSize = await page.evaluate(() => parseFloat(getComputedStyle(document.documentElement).fontSize));
        // Reserve readable text width, capped by the available phone viewport.
        const readable = range.width >= Math.min(20 * rootSize, page.viewportSize()!.width - 64);
        return titleContained && separated && readable;
      }).toBe(true);
      await expect.poll(() => page.evaluate(() => document.documentElement.scrollWidth))
        .toBeLessThanOrEqual(page.viewportSize()!.width);
    }
  });

  test(`Journey enlarged range controls retain selection and direct navigation (${theme})`, async ({ page, context, baseURL }) => {
    await context.addCookies([
      { name: "vocanova_session", value: randomUUID(), url: baseURL! },
      { name: "vocanova_csrf", value: randomUUID(), url: baseURL! },
    ]);
    await page.addInitScript((value) => localStorage.setItem("vocanova:theme-preference", value), theme);
    await page.goto("/discover?band=1");
    await expect(page.getByRole("heading", { name: "Journey", exact: true })).toBeVisible();
    await expect(page.locator("html")).toHaveAttribute("data-theme", theme);
    await page.addStyleTag({ content: "html { font-size: 200% !important; }" });
    const chooser = page.getByRole("combobox", { name: "Go to a range", exact: true });
    if (page.viewportSize()!.width < 640) {
      await expect(chooser).toBeVisible();
      // Enlarged text needs the whole row for the selected numeric range.
      await expect.poll(() => chooser.evaluate((element) => {
        const select = element.getBoundingClientRect();
        const row = element.parentElement!.getBoundingClientRect();
        return select.width >= row.width - 1;
      })).toBe(true);
      await expect(chooser.locator("option")).toHaveCount(100);
      await chooser.selectOption({ label: "4,901–5,000" });
    } else {
      const range = page.getByRole("link", { name: "Words 4901 to 5000, 0 already know, 0 learning, 100 new", exact: true });
      await range.focus();
      await expect(range).toBeFocused();
      await page.keyboard.press("Enter");
    }
    await expect(page).toHaveURL(/\/discovery\?band=50&word=4901&active=1$/);
    await expect(page.getByRole("heading", { name: "Map word 4901", exact: true })).toBeVisible();
    await expect(page.getByRole("combobox", { name: "Word range", exact: true })).toHaveValue("50");
    await expect.poll(() => page.evaluate(() => document.documentElement.scrollWidth))
      .toBeLessThanOrEqual(page.viewportSize()!.width);
  });
}

test("password reset instructions render an apostrophe rather than HTML entity text", async ({ page }) => {
  await page.goto("/auth/password/reset");
  await expect(page.getByText("We'll send instructions to your verified email. This also lets magic-link and Google users add a password.", { exact: true })).toBeVisible();
  await expect(page.getByText(/We&apos;ll/)).toHaveCount(0);
});
