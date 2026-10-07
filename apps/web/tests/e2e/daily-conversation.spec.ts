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
    const situation = page
      .getByRole("region", { name: "Guided lessons", exact: true })
      .locator("details")
      .filter({
        has: page.locator("summary").filter({ hasText: "Daily Conversation" }),
      });
    await expect(situation).toHaveCount(1);
    if ((await situation.getAttribute("open")) === null) {
      await situation.locator("summary").focus();
      await page.keyboard.press("Enter");
      await expect(situation).toHaveAttribute("open", "");
    }
    const openWords = situation.getByRole("link", {
      name: "Explore words in Daily Conversation",
      exact: true,
    });
    await expect(openWords).toBeVisible();
    await expect(openWords).toHaveAttribute("href", "/discover/daily-conversation");
    await openWords.focus();
    await page.keyboard.press("Enter");
    await expect(page).toHaveURL(/\/discover\/daily-conversation$/);
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
      // Keyboard focus may leave the back link clipped under the sticky header.
      // Scan the same stable viewport used by the other teaching a11y tests.
      await page.evaluate(() => window.scrollTo(0, 0));
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
    // Both authored examples remain available; only the first is initially shown.
    async function verifyExamples() {
      const first = page.getByRole("heading", { name: "In a sentence", exact: true }).locator("..").getByRole("listitem");
      await expect(first).toHaveCount(1);
      await expect(first.getByText("Let us keep in touch after the course ends.", { exact: true })).toBeVisible();
      const more = page.locator("summary").filter({ hasText: "More examples" });
      const extra = more.locator("..").getByRole("listitem");
      const second = extra.getByText("We kept in touch by sending each other a message every week.", { exact: true });
      await expect(second).toBeHidden();
      await more.focus();
      await page.keyboard.press("Enter");
      await expect(more).toBeFocused();
      await expect(extra).toHaveCount(1);
      await expect(second).toBeVisible();
    }
    await verifyExamples();
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
    await verifyExamples();
    await expect(fullerDefinition).toBeVisible();
    const writing = page.locator("summary").filter({ hasText: "Practise in a sentence" });
    await writing.focus();
    await page.keyboard.press("Enter");
    await expect(
      page.getByRole("region", { name: "Practice with keep in touch" }),
    ).toBeVisible();
    await checkPageAndCapture("saved-keep-in-touch");
    await page.goto("/discover/daily-conversation");
    await expect(
      page.getByRole("main").getByRole("region", { name: "Situation progress", exact: true }).getByText("1 of 18 words saved", { exact: true }),
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
