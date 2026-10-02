import { randomUUID } from "node:crypto";
import { expect, test } from "@playwright/test";

import { assertKeyboardReachable, scanForAxeViolations } from "./axe-helper";

const pathway = [
  "greeting",
  "small talk",
  "casual",
  "weekend plans",
  "available",
  "invite",
  "join",
  "suggest",
  "sounds good",
  "arrange",
  "meet up",
  "confirm",
  "on time",
  "reschedule",
  "cancel",
  "catch up",
  "keep in touch",
  "farewell",
];

for (const theme of ["light", "dark"] as const) {
  test(`Daily Conversation presents the real curriculum and saved meaning in ${theme} mode`, async ({
    page,
  }, testInfo) => {
    const url = testInfo.project.use.baseURL!;
    await page.context().addCookies([
      { name: "vocanova_session", value: randomUUID(), url },
      { name: "vocanova_csrf", value: "conversation-fixture", url },
      { name: "e2e_daily_conversation", value: "true", url },
      { name: "vocanova_theme", value: theme, url },
    ]);
    await page.goto("/discover");
    await page.getByRole("link", { name: /Daily Conversation/ }).click();
    await expect(
      page.getByRole("heading", { level: 1, name: "Daily Conversation" }),
    ).toBeVisible();
    await expect(
      page
        .getByRole("list", { name: "Words in this situation" })
        .getByRole("heading", { level: 2 }),
    ).toHaveText(pathway);
    await expect(
      page.getByText("0 of 18 words saved", { exact: true }),
    ).toBeVisible();
    await assertKeyboardReachable(page, { minFocusable: 20 });

    async function checkPageAndCapture(name: string) {
      await expect(page.locator("html")).toHaveAttribute("data-theme", theme);
      expect(
        await page.evaluate(() => document.documentElement.scrollWidth),
      ).toBeLessThanOrEqual(page.viewportSize()!.width);
      const { criticalOrSerious } = await scanForAxeViolations(page);
      expect(criticalOrSerious).toEqual([]);
      await page.screenshot({
        path: testInfo.outputPath(`${name}-${theme}.png`),
        fullPage: true,
      });
    }
    await checkPageAndCapture("daily-conversation");
    await page.getByRole("link", { name: /^keep in touch / }).click();
    await expect(
      page.getByRole("heading", { level: 1, name: "keep in touch" }),
    ).toBeVisible();
    const fullerDefinition = page.getByText(
      "You can keep in touch by calling, messaging or meeting. The phrase is often used when people say goodbye and want to stay connected.",
      { exact: true },
    );
    await expect(fullerDefinition).toBeVisible();
    // Verify both authored examples survive the seed -> API fixture -> page path.
    await expect(
      page.getByText(
        "We kept in touch by sending each other a message every week.",
        { exact: true },
      ),
    ).toBeVisible();
    const examples = page
      .getByRole("heading", { name: "Example sentences", exact: true })
      .locator("..")
      .getByRole("listitem");
    await expect(examples).toHaveCount(2);
    const save = page.getByRole("button", { name: /^Save keep in touch:/ });
    await save.focus();
    await expect(save).toBeFocused();
    const box = await save.boundingBox();
    expect(box?.height).toBeGreaterThanOrEqual(44);
    await checkPageAndCapture("keep-in-touch");
    await page.keyboard.press("Enter");
    await expect(
      page.getByRole("button", {
        name: "Remove keep in touch from saved words",
      }),
    ).toBeVisible();
    await page.goto("/words");
    await page
      .getByRole("link", {
        name: "Open keep in touch details and sentence practice",
        exact: true,
      })
      .click();
    await expect(
      page.getByRole("heading", { level: 1, name: "keep in touch" }),
    ).toBeVisible();
    await expect(examples).toHaveCount(2);
    await expect(fullerDefinition).toBeVisible();
    await expect(
      page.getByRole("region", { name: "Practice with keep in touch" }),
    ).toBeVisible();
    await checkPageAndCapture("saved-keep-in-touch");
    await page.goto("/discover/daily-conversation");
    await expect(
      page.getByText("1 of 18 words saved", { exact: true }),
    ).toBeVisible();
    // Navigation state is derived from the mock API's saved set, not a visual-only toggle.
    await expect(
      page.getByRole("link", { name: /^keep in touch / }),
    ).toContainText("Saved");
    await page.getByRole("link", { name: /^meet up / }).click();
    await expect(
      page.getByText("phrasal verb · A2", { exact: true }),
    ).toBeVisible();
  });
}
