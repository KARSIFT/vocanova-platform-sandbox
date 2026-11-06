import { randomUUID } from "node:crypto";
import { expect, test } from "@playwright/test";

test.beforeEach(async ({ context, baseURL, page }) => {
  await context.addCookies([
    { name: "vocanova_session", value: randomUUID(), url: baseURL! },
    { name: "vocanova_csrf", value: randomUUID(), url: baseURL! },
    { name: "e2e_unit_guides", value: "true", url: baseURL! },
  ]);
  await page.addInitScript(() =>
    Object.defineProperty(window, "speechSynthesis", {
      value: undefined,
      configurable: true,
    }),
  );
});

for (const theme of ["light", "dark"] as const) {
  test(`story explains optional audio once and preserves accessible failure feedback (${theme})`, async ({ page, context, baseURL }) => {
    await context.addCookies([{ name: "vocanova_theme", value: theme, url: baseURL! }]);
    await page.emulateMedia({ colorScheme: theme });
    await page.goto("/stories/a-better-fit");
    const main = page.getByRole("main");
    await expect(main.getByText("Audio is optional. You can read every line.", { exact: true })).toHaveCount(1);
    await expect(main.getByText("Optional device audio. You can read the line above.", { exact: true })).toHaveCount(0);
    const audio = main.getByRole("button", { name: /^Play audio: line / });
    await expect(audio).toHaveCount(8);
    await expect(audio.first()).not.toHaveAttribute("aria-describedby");
    await audio.first().focus();
    await page.keyboard.press("Enter");
    const feedback = main.getByRole("status").filter({ hasText: "Pronunciation is not available in this browser." });
    await expect(feedback).toBeVisible();
    await expect(audio.first()).toHaveAttribute("aria-describedby", await feedback.getAttribute("id") as string);
    await expect(audio.nth(1)).not.toHaveAttribute("aria-describedby");
    expect(await page.evaluate(() => document.documentElement.scrollWidth <= innerWidth)).toBe(true);
  });
}

test("guide and active story session explain optional audio once", async ({ page }) => {
  await page.goto("/discover/shopping");
  const guide = page.getByRole("region", { name: "A quick guide", exact: true });
  await guide.getByText("Useful phrases", { exact: true }).click();
  await expect(guide.getByText("Audio is optional. You can read every phrase.", { exact: true })).toHaveCount(1);
  await expect(guide.getByRole("button", { name: /^Play audio: guide phrase / })).toHaveCount(3);
  await page.goto("/stories/a-better-fit");
  await page.getByRole("main").getByRole("button", { name: "Start story", exact: true }).click();
  await expect(page).toHaveURL(/\/stories\/session\//);
  const main = page.getByRole("main");
  await expect(main.getByText("Audio is optional. You can read every line.", { exact: true })).toHaveCount(1);
  await expect(main.getByRole("button", { name: /^Play audio: line / })).toHaveCount(1);
});
