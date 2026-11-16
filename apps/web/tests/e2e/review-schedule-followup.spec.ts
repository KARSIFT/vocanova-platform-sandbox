import { randomUUID } from "node:crypto";
import { expect, test, type Page } from "@playwright/test";
import type { DailyLesson } from "@vocanova/api-client";

const now = "2026-11-16T12:00:00Z";
const step = {
  id: "schedule-step",
  kind: "meaning_choice" as const,
  wordText: "menu",
  wordSlug: "menu",
  partOfSpeech: "noun",
  choices: ["a list of food", "a place to sleep", "a time to leave", "a way to travel"].map((text, index) => ({ id: `choice-${index}`, text })),
};

async function checkedReview(page: Page, {
  nextReviewAt = "2026-11-17T12:00:00Z",
  timeZone = "UTC",
  theme = "light",
}: { nextReviewAt?: string; timeZone?: string | null; theme?: "light" | "dark" } = {}) {
  await page.clock.setFixedTime(new Date(now));
  const baseURL = test.info().project.use.baseURL!;
  await page.context().addCookies([
    { name: "vocanova_session", value: randomUUID(), url: baseURL },
    { name: "vocanova_csrf", value: "review-schedule", url: baseURL },
    { name: "e2e_review_fixture_count", value: "20", url: baseURL },
    { name: "e2e_review_local_date", value: "2026-11-16", url: baseURL },
    { name: "vocanova_theme", value: theme, url: baseURL },
  ]);
  const value: DailyLesson = {
    id: "schedule-lesson",
    attribution: { title: "Word sources", sourceUrl: "https://wordnet.princeton.edu/", licenseUrl: "https://wordnet.princeton.edu/license-and-commercial-use", wordnetLicense: "WordNet" },
    status: "in_progress", revision: 1, completedWords: 0, totalWords: 15,
    firstAnswersCorrect: 0, questionsAnswered: 0, creditedToday: 0,
    goal: { localDate: "2026-11-16", timeZone: timeZone ?? "", target: 15, reviewed: 0, remaining: 15, complete: false },
    currentStep: step, feedback: null, canContinue: false,
    createdAt: now, updatedAt: now,
  };
  const actions: { action: string; nextReview?: string }[] = [];
  await page.route("**/api/v1/daily-lessons", (route) => route.fulfill({ json: value }));
  await page.route("**/api/v1/canonical-words/menu", (route) => route.fulfill({ json: { word: { id: "menu", text: "menu", slug: "menu", wordType: "word", meanings: [{ id: "schedule-meaning", partOfSpeech: "noun", shortDefinition: "a list of food", saved: true, selfReportedKnown: false, examples: [], usageNotes: [] }] } } }));
  await page.route("**/api/v1/meaning-images**", (route) => route.fulfill({ json: { items: [] } }));
  await page.route("**/api/v1/daily-lessons/*/actions", (route) => {
    const action = route.request().postDataJSON() as { action: string; nextReview?: string };
    actions.push(action);
    return route.fulfill({ json: action.action === "continue"
      ? { ...value, revision: 3, completedWords: 1, currentStep: { ...step, id: "next-schedule-step", wordText: "restaurant", wordSlug: "restaurant" } }
      : { ...value, revision: 2, canContinue: true, feedback: { stepId: step.id, correct: true, answer: "a list of food", explanation: "", meaningId: "schedule-meaning", wordText: "menu", credited: true, skipped: false, correctChoiceId: "choice-0", selectedChoiceId: "choice-0", nextReviewAt, reviewStep: 1 } } });
  });
  await page.goto("/review");
  await page.getByRole("button", { name: "Start review", exact: true }).click();
  await page.getByRole("radio").first().check();
  await page.getByRole("button", { name: "Check answer", exact: true }).click();
  const trigger = page.getByRole("button", { name: /^Show this word again:/ });
  await expect(trigger).toBeEnabled();
  return { trigger, actions };
}

for (const theme of ["light", "dark"] as const) {
  test(`schedule popup matches its trigger and reflows at 200 percent text in ${theme}`, async ({ page }) => {
    const { trigger } = await checkedReview(page, { theme });
    await expect(page.locator("html")).toHaveAttribute("data-theme", theme);
    for (const scale of [100, 200]) {
      if (scale === 100) await page.addStyleTag({ content: "html { font-size: 100% !important; }" });
      await trigger.click();
      const menu = page.getByRole("menu", { name: "Next review", exact: true });
      await expect(menu).toBeVisible();
      // Growing text with the popup already open also exercises its live sizing.
      if (scale === 200) await page.addStyleTag({ content: "html { font-size: 200% !important; }" });
      await expect.poll(() => menu.evaluate((element) => {
        const box = element.getBoundingClientRect();
        const header = document.querySelector("header")!.getBoundingClientRect();
        const trigger = document.querySelector(".review-schedule-trigger")!.getBoundingClientRect();
        return box.top >= Math.max(0, header.bottom) + 8 && box.bottom <= trigger.top - 8;
      })).toBe(true);
      const triggerBox = (await trigger.boundingBox())!;
      const menuBox = (await menu.boundingBox())!;
      expect(Math.abs(menuBox.width - triggerBox.width)).toBeLessThanOrEqual(1);
      expect(Math.abs(menuBox.x - triggerBox.x)).toBeLessThanOrEqual(1);
      expect(menuBox.x).toBeGreaterThanOrEqual(0);
      expect(menuBox.x + menuBox.width).toBeLessThanOrEqual(page.viewportSize()!.width);
      const headerBox = (await page.getByRole("banner").boundingBox())!;
      expect(menuBox.y).toBeGreaterThanOrEqual(Math.max(0, headerBox.y + headerBox.height) + 8);
      expect(menuBox.y + menuBox.height).toBeLessThanOrEqual(triggerBox.y - 8);
      for (const option of await menu.getByRole("menuitemradio").all()) {
        // Firefox may report a CSS 44px target as 43.999992px.
        expect((await option.boundingBox())!.height + 0.001).toBeGreaterThanOrEqual(44);
        expect(await option.evaluate((element) => element.scrollWidth <= element.clientWidth + 1)).toBe(true);
      }
      expect(await page.evaluate(() => document.documentElement.scrollWidth)).toBeLessThanOrEqual(page.viewportSize()!.width);
      await page.screenshot({ path: test.info().outputPath(`schedule-${theme}-${scale}.png`) });
      await page.keyboard.press("End");
      const last = menu.getByRole("menuitemradio", { name: "I know this word", exact: true });
      await expect(last).toBeFocused();
      const lastBox = (await last.boundingBox())!;
      const scrolledMenuBox = (await menu.boundingBox())!;
      expect(lastBox.y).toBeGreaterThanOrEqual(scrolledMenuBox.y);
      expect(lastBox.y + lastBox.height).toBeLessThanOrEqual(scrolledMenuBox.y + scrolledMenuBox.height);
      await page.keyboard.press("Escape");
      await expect(trigger).toBeFocused();
    }
  });
}

test("automatic Tomorrow has one checked date and stays server-selected until an explicit choice", async ({ page }) => {
  const { trigger, actions } = await checkedReview(page);
  await trigger.click();
  const menu = page.getByRole("menu", { name: "Next review", exact: true });
  await expect(menu).not.toContainText("Automatic");
  await expect(menu.getByRole("menuitemradio")).toHaveCount(8);
  await expect(menu.getByRole("menuitemradio", { name: /^Tomorrow(?:\s*\.|$)/ })).toHaveCount(1);
  await expect(menu.getByRole("menuitemradio", { checked: true })).toHaveAccessibleName(/^Tomorrow\s*\. November 17, 2026/);
  await page.keyboard.press("Escape");
  await page.getByRole("button", { name: "Continue", exact: true }).click();
  await expect(page.getByRole("heading", { name: "restaurant", exact: true })).toBeVisible();
  expect(actions.at(-1)?.action).toBe("continue");
  expect(actions.at(-1)?.nextReview).toBeUndefined();
});

test("clicking the checked server date preserves its exact time and automatic continuation", async ({ page }) => {
  const { trigger, actions } = await checkedReview(page, { nextReviewAt: "2026-11-17T09:15:00Z" });
  await trigger.click();
  const menu = page.getByRole("menu", { name: "Next review", exact: true });
  await menu.getByRole("menuitemradio", { checked: true }).click();
  await expect(menu).toHaveCount(0);
  await expect(trigger).toBeFocused();
  await expect(trigger).toHaveAccessibleName(/^Show this word again: Tomorrow\s*\. November 17, 2026 at 9:15/);
  await trigger.click();
  await expect(menu.getByRole("menuitemradio", { checked: true })).toHaveCount(1);
  await expect(menu.getByRole("menuitemradio", { checked: true })).toHaveAccessibleName(/^Tomorrow\s*\. November 17, 2026 at 9:15/);
  await page.keyboard.press("Escape");
  await page.getByRole("button", { name: "Continue", exact: true }).click();
  await expect(page.getByRole("heading", { name: "restaurant", exact: true })).toBeVisible();
  expect(actions.at(-1)?.action).toBe("continue");
  expect(actions.at(-1)?.nextReview).toBeUndefined();
});

test("an explicit interval keeps its single check on reopen and supports menu keyboard dismissal", async ({ page }) => {
  // A scheduled Tomorrow at a different time still retains its exact server date.
  const { trigger, actions } = await checkedReview(page, { nextReviewAt: "2026-11-17T09:15:00Z" });
  await trigger.focus();
  await trigger.press("ArrowDown");
  const menu = page.getByRole("menu", { name: "Next review", exact: true });
  const automaticDate = menu.getByRole("menuitemradio", { checked: true });
  await expect(automaticDate).toHaveAccessibleName(/^Tomorrow\s*\. November 17, 2026 at 9:15/);
  await expect(automaticDate).toBeFocused();
  await page.keyboard.press("End");
  await expect(menu.getByRole("menuitemradio", { name: "I know this word", exact: true })).toBeFocused();
  await page.keyboard.press("Home");
  await expect(menu.getByRole("menuitemradio").first()).toBeFocused();
  await page.keyboard.press("ArrowDown");
  await expect(menu.getByRole("menuitemradio", { name: /^In 3 days\s*\./ })).toBeFocused();
  await page.keyboard.press("ArrowRight");
  await expect(menu.getByRole("menuitemradio", { name: /^In 1 week\s*\./ })).toBeFocused();
  await page.keyboard.press("Enter");
  await expect(trigger).toHaveText(/In 1 week/);
  await trigger.click();
  await expect(menu.getByRole("menuitemradio", { checked: true })).toHaveCount(1);
  await expect(menu.getByRole("menuitemradio", { checked: true })).toHaveAccessibleName(/^In 1 week\s*\. November 23, 2026/);
  await page.keyboard.press("Tab");
  await expect(page.getByRole("button", { name: "Continue", exact: true })).toBeFocused();
  await expect(menu).toHaveCount(0);
  await trigger.click();
  await page.keyboard.press("Shift+Tab");
  await expect(trigger).toBeFocused();
  await trigger.click();
  await page.mouse.click(1, 1);
  await expect(menu).toHaveCount(0);
  await page.getByRole("button", { name: "Continue", exact: true }).click();
  await expect(page.getByRole("heading", { name: "restaurant", exact: true })).toBeVisible();
  expect(actions.at(-1)?.nextReview).toBe("7d");
});

test("a server-selected three-day date checks the existing interval without duplicating it", async ({ page }) => {
  const { trigger } = await checkedReview(page, { nextReviewAt: "2026-11-19T12:00:00Z" });
  await trigger.click();
  const menu = page.getByRole("menu", { name: "Next review", exact: true });
  await expect(menu).not.toContainText("Automatic");
  await expect(menu.getByRole("menuitemradio")).toHaveCount(8);
  await expect(menu.getByRole("menuitemradio", { name: /^In 3 days\s*\./ })).toHaveCount(1);
  await expect(menu.getByRole("menuitemradio", { checked: true })).toHaveAccessibleName(/^In 3 days\s*\. November 19, 2026/);
});

test("a server date outside the fixed intervals remains the checked actual date", async ({ page }) => {
  const { trigger, actions } = await checkedReview(page, { nextReviewAt: "2026-11-18T12:00:00Z" });
  await expect(trigger).toHaveText(/In 2 days/);
  await trigger.click();
  const menu = page.getByRole("menu", { name: "Next review", exact: true });
  await expect(menu).not.toContainText("Automatic");
  await expect(menu.getByRole("menuitemradio", { checked: true })).toHaveAccessibleName(/^In 2 days\s*\. November 18, 2026/);
  await expect(menu.getByRole("menuitemradio", { checked: true })).toHaveCount(1);
  await page.keyboard.press("Escape");
  await page.getByRole("button", { name: "Continue", exact: true }).click();
  await expect(page.getByRole("heading", { name: "restaurant", exact: true })).toBeVisible();
  expect(actions.at(-1)?.action).toBe("continue");
  expect(actions.at(-1)?.nextReview).toBeUndefined();
});

test("legacy responses without a learner zone display the real UTC date without inventing Tomorrow", async ({ page }) => {
  const { trigger, actions } = await checkedReview(page, { nextReviewAt: "2026-11-18T09:15:00Z", timeZone: null });
  await trigger.click();
  const menu = page.getByRole("menu", { name: "Next review", exact: true });
  await expect(menu).not.toContainText("Automatic");
  await expect(menu.getByRole("menuitemradio", { checked: true })).toHaveAccessibleName(/November 18, 2026 at 9:15.*UTC/);
  await expect(menu.getByRole("menuitemradio", { checked: true })).toHaveCount(1);
  await page.keyboard.press("Escape");
  await page.getByRole("button", { name: "Continue", exact: true }).click();
  await expect(page.getByRole("heading", { name: "restaurant", exact: true })).toBeVisible();
  expect(actions.at(-1)?.action).toBe("continue");
  expect(actions.at(-1)?.nextReview).toBeUndefined();
});
