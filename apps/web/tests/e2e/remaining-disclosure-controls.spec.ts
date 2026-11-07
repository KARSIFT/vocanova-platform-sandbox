import { randomUUID } from "node:crypto";
import { expect, test, type Locator, type Page } from "@playwright/test";
import { startReviewLesson } from "./review-readiness";

function disclosure(page: Page, label: string) {
  return page.getByRole("main").locator("details").filter({
    has: page.locator("summary").filter({ hasText: new RegExp(`^${label.replace(/[.*+?^${}()|[\]\\]/g, "\\$&")}$`) }),
  });
}

async function toggleWithKeyboard(page: Page, details: Locator, content: Locator) {
  const summary = details.locator("summary");
  await expect(summary).toBeVisible();
  if (await details.getAttribute("open") !== null) {
    await summary.focus();
    await page.keyboard.press("Enter");
  }
  await expect(details).not.toHaveAttribute("open");
  await expect(content).toBeHidden();
  await summary.focus();
  await page.keyboard.press("Space");
  await expect(details).toHaveAttribute("open", "");
  await expect(summary).toBeFocused();
  await expect(content).toBeVisible();
  await page.keyboard.press("Enter");
  await expect(details).not.toHaveAttribute("open");
  await expect(content).toBeHidden();
  await expect(summary).toBeFocused();
  await page.keyboard.press("Space");
  await expect(details).toHaveAttribute("open", "");
  await expect(content).toBeVisible();
}

async function checkLayout(page: Page) {
  expect(await page.evaluate(() => document.documentElement.scrollWidth))
    .toBeLessThanOrEqual(page.viewportSize()!.width);
}

for (const theme of ["light", "dark"] as const) {
  test.describe(`remaining disclosures (${theme})`, () => {
    test.beforeEach(async ({ page, context, baseURL }) => {
      await context.addCookies([
        { name: "vocanova_session", value: randomUUID(), url: baseURL! },
        { name: "vocanova_csrf", value: randomUUID(), url: baseURL! },
        { name: "vocanova_theme", value: theme, url: baseURL! },
        { name: "e2e_review_fixture_count", value: "1", url: baseURL! },
      ]);
      await page.addInitScript(value => localStorage.setItem("vocanova:theme-preference", value), theme);
    });

    test("Journey ordering and vocabulary filters open, close and retain submitted choices", async ({ page }) => {
      await page.goto("/discover");
      await expect(page.locator("html")).toHaveAttribute("data-theme", theme);
      const ordering = disclosure(page, "Why this order?");
      await toggleWithKeyboard(page, ordering, ordering.getByText("Common words first", { exact: true }));
      await expect(ordering.getByText("Synthetic interface fixture.", { exact: false })).toBeVisible();
      await expect(ordering.getByRole("link", { name: "Word order source (opens a new tab)", exact: true }))
        .toHaveAttribute("href", "https://github.com/hermitdave/FrequencyWords");
      await checkLayout(page);

      await page.goto("/vocabulary");
      const filters = disclosure(page, "Filter lesson words");
      await toggleWithKeyboard(page, filters, filters.getByRole("combobox", { name: "Situation", exact: true }));
      await page.getByRole("searchbox", { name: "Search words and meanings", exact: true }).fill("pour");
      await filters.getByRole("combobox", { name: "Situation", exact: true }).selectOption("travel");
      await filters.getByRole("combobox", { name: "Level", exact: true }).selectOption("a2");
      await page.getByRole("button", { name: "Search", exact: true }).focus();
      await page.keyboard.press("Enter");
      await expect(page).toHaveURL(/category=travel/);
      await expect(page).toHaveURL(/level=a2/);
      await expect(page.getByRole("heading", { name: "No matching words yet", exact: true })).toBeVisible();
      await expect(filters).toHaveAttribute("open", "");
      await expect(filters.getByRole("combobox", { name: "Situation", exact: true })).toHaveValue("travel");
      await expect(filters.getByRole("combobox", { name: "Level", exact: true })).toHaveValue("a2");
      await checkLayout(page);
    });

    test("word sources remain reachable from discovery and disclose on Journey and review", async ({ page }) => {
      for (const path of ["/discover", "/discovery", "/reviews"]) {
        await page.goto(path);
        if (path === "/discovery") {
          await expect(page.getByText("Word sources", { exact: true })).toHaveCount(0);
          await page.getByRole("banner").getByRole("link", { name: "Back to Journey", exact: true }).click();
          await expect(page).toHaveURL(/\/discover\?band=1#range-words$/);
        }
        if (path === "/reviews") {
          await startReviewLesson(page);
          await expect(page.getByRole("radio").first()).toBeVisible();
        }
        await expect(page.locator("html")).toHaveAttribute("data-theme", theme);
        const sources = disclosure(page, "Word sources");
        const title = sources.getByText("FrequencyWords / OpenSubtitles 2018, WordNet 3.0 and Vocanova editorial content", { exact: true });
        await toggleWithKeyboard(page, sources, title);
        await expect(sources.getByRole("link", { name: "Vocabulary source", exact: true }))
          .toHaveAttribute("href", "https://github.com/hermitdave/FrequencyWords");
        await expect(sources.getByRole("link", { name: "Catalog license", exact: true }))
          .toHaveAttribute("href", "https://creativecommons.org/licenses/by-sa/4.0/");
        await expect(sources.locator("pre")).toContainText("WordNet");
        await checkLayout(page);
      }
    });

    test("recent practice and unsupported path words disclose confirmed session information", async ({ page }) => {
      await page.goto("/practice?lesson=conversation-basics");
      await expect(page.getByLabel("Practice vocabulary", { exact: true })).toHaveValue("conversation-basics");
      const start = page.getByRole("button", { name: "Start typed recall", exact: true });
      await start.focus();
      await page.keyboard.press("Enter");
      await expect(page).toHaveURL(/\/practice\/session\/[^/]+$/);
      await expect(page.getByRole("progressbar", { name: "Practice progress", exact: true })).toBeVisible();
      const sessionPath = new URL(page.url()).pathname;
      await page.goto("/practice?lesson=conversation-basics");
      await expect(page.locator("html")).toHaveAttribute("data-theme", theme);
      const recent = disclosure(page, "Recent practice");
      const resume = recent.getByRole("link", { name: /^Resume practice/ });
      await toggleWithKeyboard(page, recent, resume);
      await expect(resume).toHaveAttribute("href", sessionPath);
      await expect(recent.getByText(/0 of \d+ questions completed/)).toBeVisible();

      // The fixture's three canonical words all support practice. Change only
      // this synthetic read response to exercise an additional unsupported
      // meaning; this is not real catalog coverage or persistence evidence.
      let refreshed = 0;
      await page.route("**/api/v1/practice-sessions", async route => {
        if (route.request().method() !== "GET") return route.continue();
        const response = await route.fetch();
        expect(response.ok()).toBe(true);
        const body = await response.json();
        body.journey = { ...body.journey, totalKnownWords: 4, supportedKnownWords: 3 };
        refreshed += 1;
        await route.fulfill({ response, json: body });
      });
      const refresh = page.getByRole("button", { name: "Refresh practice sources", exact: true });
      await refresh.focus();
      await page.keyboard.press("Enter");
      const words = disclosure(page, "Words in this path");
      await toggleWithKeyboard(page, words, words.getByText("3 of your 4 words currently have supported practice exercises.", { exact: false }));
      expect(refreshed).toBe(1);
      await expect(words.getByText("More words remain available in your map.", { exact: false })).toBeVisible();
      await checkLayout(page);
    });
  });
}
