import { randomUUID } from "node:crypto";
import { expect, test } from "@playwright/test";
for (const theme of ["light", "dark"] as const) {
  test(`overview keeps practice prominent and help keyboard-accessible in ${theme}`, async ({ page, context, baseURL }, testInfo) => {
    if (!baseURL) throw new Error("Missing app URL");
    await context.addCookies([
      { name: "vocanova_session", value: randomUUID(), url: baseURL },
      { name: "vocanova_theme", value: theme, url: baseURL },
    ]);
    await page.goto("/home");
    await expect(page.getByRole("heading", { name: "Today", exact: true })).toBeVisible();
    const mission = page.getByRole("region", { name: "Today’s practice" });
    await expect(mission.getByRole("progressbar", { name: "Reviews today" })).toHaveAttribute("aria-valuenow", "0");
    await expect(mission.getByRole("link")).toBeInViewport();
    await page.screenshot({ path: testInfo.outputPath(`home-${theme}.png`), fullPage: true });
    await page.goto("/discover");
    await expect(page.getByRole("region", { name: "Guided lessons" })).toBeVisible();
    const practice = page.getByRole("navigation", { name: "Ways to practise" }).getByRole("link", { name: /^Choose a practice/ });
    await expect(practice).toBeVisible();
    await practice.focus();
    await expect(practice).toBeFocused();
    await expect(practice).toHaveAttribute("href", "/practice");
    await expect(page.getByRole("link", { name: /^Short stories/ })).toHaveAttribute("href", "/stories");
    await page.evaluate(() => scrollTo(0, 0));
    await page.screenshot({ path: testInfo.outputPath(`journey-${theme}.png`), fullPage: true });
    await page.goto("/progress");
    const summary = page.getByRole("region", { name: "Learning summary" });
    const map = page.getByRole("region", { name: "Your vocabulary map" });
    expect((await map.boundingBox())!.y).toBeLessThan((await summary.boundingBox())!.y);
    await page.screenshot({ path: testInfo.outputPath(`progress-${theme}.png`), fullPage: true });
    const stages = map.locator("summary").filter({ hasText: "About these stages" });
    const explanation = map.getByText("New words become Learning, Reviewing and Mastered through your review history. These stages describe saved words, not your overall English level.");
    await expect(explanation).toBeHidden();
    await stages.focus();
    await page.keyboard.press("Enter");
    await expect(explanation).toBeVisible();
    await expect(stages).toBeFocused();
    expect(await stages.evaluate((element) => element.getBoundingClientRect().height)).toBeGreaterThanOrEqual(44);
    expect(await page.evaluate(() => document.documentElement.scrollWidth <= innerWidth)).toBe(true);
  });
}
