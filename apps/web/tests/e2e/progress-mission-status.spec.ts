import { expect, test } from "@playwright/test";

import { scanForAxeViolations } from "./axe-helper";

for (const theme of ["light", "dark"] as const) {
  test(`Progress distinguishes mission completion and streak protection in ${theme} mode`, async ({
    page,
    context,
  }, testInfo) => {
    const baseURL = testInfo.project.use.baseURL;
    if (!baseURL) throw new Error("Expected a configured baseURL.");
    await context.addCookies([
      { name: "e2e_progress_history_fixture", value: "statuses", url: baseURL },
      { name: "vocanova_theme", value: theme, url: baseURL },
    ]);
    await page.goto("/progress");

    const activity = page.getByRole("region", { name: "Recent activity" });
    const rows = activity.getByRole("listitem");
    await expect(rows).toHaveCount(6);
    for (const [index, label] of [
      "Completed",
      "Streak protected",
      "Not complete",
      "Not complete",
      "Completed or protected",
      "Not complete",
    ].entries()) {
      await expect(
        rows.nth(index).getByText(label, { exact: true }),
      ).toBeVisible();
    }
    const completedStyle = await rows
      .nth(0)
      .evaluate((element) => getComputedStyle(element).borderLeftColor);
    const protectedStyle = await rows
      .nth(1)
      .evaluate((element) => getComputedStyle(element).borderLeftColor);
    expect(protectedStyle).not.toBe(completedStyle);
    await expect(page.locator("html")).toHaveAttribute("data-theme", theme);
    expect(
      await page.evaluate(() => document.documentElement.scrollWidth),
    ).toBeLessThanOrEqual(page.viewportSize()!.width);
    const { criticalOrSerious } = await scanForAxeViolations(page);
    expect(criticalOrSerious).toEqual([]);
  });
}
