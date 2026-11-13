import { randomUUID } from "node:crypto";
import { expect, test } from "@playwright/test";

for (const theme of ["light", "dark"] as const) {
  test.describe(`manual page hierarchy (${theme})`, () => {
    test.beforeEach(async ({ context, page, baseURL }) => {
      await context.addCookies([
        { name: "vocanova_session", value: randomUUID(), url: baseURL! },
        { name: "vocanova_csrf", value: randomUUID(), url: baseURL! },
        { name: "vocanova_theme", value: theme, url: baseURL! },
      ]);
      await page.addInitScript((value) => {
        localStorage.setItem("vocanova:theme-preference", value);
      }, theme);
    });

    test("sentence history keeps return navigation above its page label", async ({ page }, info) => {
      await page.goto("/progress/sentences");
      const main = page.getByRole("main");
      const back = main.getByRole("link", { name: "Back to Progress", exact: true });
      const label = main.getByText("Your writing practice", { exact: true });
      const heading = main.getByRole("heading", { name: "Sentence history", exact: true });
      await expect(heading).toBeVisible();
      await expect(page.locator("html")).toHaveAttribute("data-theme", theme);

      for (const textSize of ["100%", "200%"] as const) {
        await page.evaluate((size) => { document.documentElement.style.fontSize = size; }, textSize);
        const [backBox, labelBox, headingBox] = await Promise.all([
          back.boundingBox(), label.boundingBox(), heading.boundingBox(),
        ]);
        expect(backBox).not.toBeNull();
        expect(labelBox).not.toBeNull();
        expect(headingBox).not.toBeNull();
        // Inline siblings used to touch on one row. A page label needs its own
        // row with visible space below the complete return-link hit target.
        expect(labelBox!.y - (backBox!.y + backBox!.height)).toBeGreaterThanOrEqual(8);
        expect(Math.abs(labelBox!.x - backBox!.x)).toBeLessThanOrEqual(1);
        expect(headingBox!.y).toBeGreaterThanOrEqual(labelBox!.y + labelBox!.height);
        expect(await page.evaluate(() => document.documentElement.scrollWidth))
          .toBeLessThanOrEqual(page.viewportSize()!.width);
        await page.screenshot({ path: info.outputPath(`sentence-heading-${theme}-${textSize}.png`) });
      }

      await back.focus();
      await expect(back).toBeFocused();
      await page.keyboard.press("Enter");
      await expect(page).toHaveURL(/\/progress$/);
      await expect(main.getByRole("heading", { name: "Progress", exact: true })).toBeVisible();
    });

    test("Home uses interface typography for empty copy and word typography for saved words", async ({ page }, info) => {
      await page.goto("/home");
      const latest = page.getByRole("region", { name: "Your latest word", exact: true });
      const empty = latest.getByText("Nothing saved yet", { exact: true });
      await expect(empty).toBeVisible();
      await expect(page.locator("html")).toHaveAttribute("data-theme", theme);
      const ui = await latest.getByRole("heading").evaluate((element) => ({
        family: getComputedStyle(element).fontFamily,
        baseSize: Number.parseFloat(getComputedStyle(document.body).fontSize),
      }));
      const emptyStyle = await empty.evaluate((element) => ({
        family: getComputedStyle(element).fontFamily,
        size: Number.parseFloat(getComputedStyle(element).fontSize),
      }));
      expect(emptyStyle.family).toBe(ui.family);
      expect(emptyStyle.size).toBeLessThanOrEqual(ui.baseSize * 1.5);
      await latest.screenshot({ path: info.outputPath(`latest-word-empty-${theme}.png`) });

      await latest.getByRole("link", { name: "Search for a word", exact: true }).press("Enter");
      await expect(page).toHaveURL(/\/vocabulary$/);
      await page.goto("/vocabulary/pour");
      await page.getByRole("button", { name: /^Save pour:/ }).click();
      // Server-rendered saved controls confirm the refresh before Home reads
      // the latest word; a changed button label alone precedes that refresh.
      await expect(page.getByRole("main").locator("summary").filter({ hasText: /^Optional writing$/ })).toHaveCount(1);
      await page.goto("/home");
      const word = latest.getByRole("link", { name: "pour", exact: true });
      await expect(word).toBeVisible();
      await expect(empty).toHaveCount(0);
      const wordStyle = await word.evaluate((element) => ({
        family: getComputedStyle(element).fontFamily,
        size: Number.parseFloat(getComputedStyle(element).fontSize),
      }));
      expect(wordStyle.family).not.toBe(ui.family);
      expect(wordStyle.size).toBeGreaterThanOrEqual(ui.baseSize * 1.75);
      await expect(word).toHaveAttribute("href", "/words/uw-mean-pour");
      await latest.screenshot({ path: info.outputPath(`latest-word-saved-${theme}.png`) });
    });
  });
}
