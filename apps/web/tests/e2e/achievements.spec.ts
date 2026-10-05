import { randomUUID } from "node:crypto";
import { expect, test } from "@playwright/test";
import { scanForAxeViolations, formatViolations } from "./axe-helper";

for (const theme of ["light", "dark"] as const) {
  test(`milestones distinguish participation and independent recall in ${theme} mode`, async ({ page, context, baseURL }, testInfo) => {
    await context.addCookies([
      { name: "vocanova_session", value: randomUUID(), url: baseURL! },
      { name: "vocanova_theme", value: theme, url: baseURL! },
      { name: "e2e_achievements", value: "earned", url: baseURL! },
    ]);
    await page.goto("/progress");
    const section = page.getByRole("region", { name: "Your milestones", exact: true });
    await expect(section.getByText("2 of 8 earned", { exact: true })).toBeVisible();
    const disclosure = section.locator("summary").filter({ hasText: "All milestones" });
    await expect(section.getByRole("heading", { name: "First lesson", exact: true })).toBeHidden();
    await disclosure.focus();
    await page.keyboard.press("Enter");
    await expect(disclosure).toBeFocused();
    await expect(section.getByRole("listitem")).toHaveCount(8);
    const first = section.getByRole("listitem").filter({ has: page.getByRole("heading", { name: "First lesson", exact: true }) });
    await expect(first.getByText("Earned · Participation", { exact: true })).toBeVisible();
    await expect(first.locator("time")).toHaveAttribute("datetime", "2026-10-02T10:00:00Z");
    const recall = section.getByRole("listitem").filter({ has: page.getByRole("heading", { name: "Recall on your own", exact: true }) });
    await expect(recall.getByText("Earned · Recall without help", { exact: true })).toBeVisible();
    await expect(section.getByRole("progressbar", { name: "Three lessons progress", exact: true })).toHaveAttribute("value", "1");
    await expect(section.getByRole("progressbar", { name: "Three lessons progress", exact: true })).toHaveAttribute("max", "3");
    await expect(section.getByRole("link", { name: /^Choose a practice\s*:\s*Keep practising$/ })).toHaveAttribute("href", "/practice");
    expect(await page.evaluate(() => document.documentElement.scrollWidth)).toBeLessThanOrEqual(page.viewportSize()!.width);
    const a11y = await scanForAxeViolations(page);
    expect(a11y.criticalOrSerious, formatViolations(a11y.criticalOrSerious).join("\n")).toEqual([]);
    await page.screenshot({ path: testInfo.outputPath(`milestones-${theme}.png`), fullPage: true });
  });
}

test("unavailable milestones do not pretend that earned history is empty", async ({ page, context, baseURL }) => {
  await context.addCookies([
    { name: "vocanova_session", value: randomUUID(), url: baseURL! },
    { name: "e2e_achievements", value: "unavailable", url: baseURL! },
  ]);
  await page.goto("/progress");
  const main = page.getByRole("main");
  await expect(main.getByText("Your milestones could not load. Your other progress is still available.", { exact: true })).toBeVisible();
  await expect(main.getByText("0 of 8 earned", { exact: true })).toHaveCount(0);
  await expect(main.getByRole("region", { name: "Learning summary", exact: true })).toBeVisible();
});


test("unearned milestones stay compact without removing activity history", async ({ page, context, baseURL }) => {
  await context.addCookies([{ name: "vocanova_session", value: randomUUID(), url: baseURL! }]);
  await page.goto("/progress");
  const section = page.getByRole("region", { name: "Your milestones", exact: true });
  await expect(section.getByText("0 of 8 earned", { exact: true })).toBeVisible();
  await expect(section.getByRole("heading", { name: "First lesson", exact: true })).toBeHidden();
  expect((await section.boundingBox())!.height).toBeLessThan(200);
  await expect(page.getByRole("heading", { name: "Recent activity", exact: true })).toBeVisible();
  const disclosure = section.locator("summary").filter({ hasText: "All milestones" });
  await disclosure.focus();
  await page.keyboard.press("Space");
  await expect(section.getByRole("listitem")).toHaveCount(8);
  await expect(section.getByRole("progressbar", { name: "First lesson progress", exact: true })).toHaveAttribute("value", "0");
  await disclosure.focus();
  await page.keyboard.press("Enter");
  await expect(section.getByRole("heading", { name: "First lesson", exact: true })).toBeHidden();
  await expect(disclosure).toBeFocused();
});
