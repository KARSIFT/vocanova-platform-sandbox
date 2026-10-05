import { randomUUID } from "node:crypto";
import { expect, test } from "@playwright/test";
import { scanForAxeViolations } from "./axe-helper";
test.beforeEach(async ({ context, baseURL }) => {
  await context.addCookies([
    { name: "vocanova_session", value: randomUUID(), url: baseURL! },
    { name: "vocanova_csrf", value: randomUUID(), url: baseURL! },
    { name: "e2e_teaching", value: "true", url: baseURL! },
  ]);
});
for (const theme of ["light", "dark"] as const) {
  test(`canonical picture and optional usage help stay usable in ${theme}`, async ({ page, context, baseURL }, testInfo) => {
    await context.addCookies([{ name: "vocanova_theme", value: theme, url: baseURL! }]);
    await page.goto("/vocabulary/cover-letter");
    const picture = page.getByTestId("meaning-picture").getByRole("img");
    await expect(picture).toBeVisible();
    await expect.poll(() => picture.evaluate((element) => (element as HTMLImageElement).naturalWidth)).toBeGreaterThan(0);
    await expect(picture).toHaveAttribute("alt", /.+/);
    const tips = page.locator("summary").filter({ hasText: "Usage tips" });
    if (await tips.count()) {
      await expect(tips.locator("..")).not.toHaveAttribute("open");
      await tips.focus();
      await page.keyboard.press("Enter");
      await expect(tips.locator("..")).toHaveAttribute("open", "");
      await expect(tips).toBeFocused();
    }
    expect(await page.evaluate(() => document.documentElement.scrollWidth <= innerWidth)).toBe(true);
    await page.evaluate(() => scrollTo(0, 0));
    expect((await scanForAxeViolations(page)).criticalOrSerious).toEqual([]);
    await page.screenshot({ path: testInfo.outputPath(`word-picture-${theme}.png`), fullPage: true });
  });
}
test("broken picture leaves the meaning and save action usable", async ({ page }) => {
  await page.route("**/_next/image?**", (route) => route.abort());
  await page.goto("/vocabulary/cover-letter");
  await expect(page.getByRole("heading", { name: "cover letter", exact: true })).toBeVisible();
  await expect(page.getByTestId("meaning-picture")).toHaveCount(0);
  await expect(page.getByRole("button", { name: /^Save cover letter:/ })).toBeVisible();
  await expect(page.getByRole("heading", { name: "In a sentence" })).toBeVisible();
});
test("a teaching picture disappears before graded lesson questions", async ({ page }) => {
  await page.goto("/learn/conversation-basics");
  await page.getByRole("button", { name: "Start lesson", exact: true }).click();
  const teachingImage = page.getByTestId("meaning-picture").getByRole("img");
  await expect(page.getByRole("heading", { name: "invite", exact: true })).toBeVisible();
  await expect(teachingImage).toBeVisible();
  await expect.poll(() => teachingImage.evaluate((element) => (element as HTMLImageElement).naturalWidth)).toBeGreaterThan(0);
  expect((await teachingImage.boundingBox())!.height).toBeLessThanOrEqual(288);
  for (let step = 0; step < 3; step++) {
    await expect(page.getByRole("button", { name: "Continue", exact: true })).toBeVisible();
    await page.getByRole("button", { name: "Continue", exact: true }).click();
  }
  await expect(page.getByRole("group", { name: "Answer choices" })).toBeVisible();
  await expect(page.getByTestId("meaning-picture")).toHaveCount(0);
});
