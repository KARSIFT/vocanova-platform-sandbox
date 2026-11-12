import { randomUUID } from "node:crypto";

import { expect, test, type Locator, type Page } from "@playwright/test";

import { formatViolations, scanForAxeViolations } from "./axe-helper.js";
import { pressTab } from "./keyboard-navigation.js";

async function expectKeyboardTargetClear(page: Page, target: Locator) {
  await target.focus();
  await pressTab(page);
  await pressTab(page, { backwards: true });
  await expect(target).toBeFocused();
  await expect.poll(() => target.evaluate((element) => {
    const rect = element.getBoundingClientRect();
    const header = document.querySelector("header")?.getBoundingClientRect();
    const bottomNav = Array.from(document.querySelectorAll("nav")).find((nav) =>
      getComputedStyle(nav).position === "fixed" && nav.getBoundingClientRect().height > 0 &&
      nav.getBoundingClientRect().bottom >= innerHeight - 1)?.getBoundingClientRect();
    const hit = document.elementFromPoint(rect.x + rect.width / 2, rect.y + rect.height / 2);
    return rect.top >= Math.max(0, header?.bottom ?? 0) && rect.bottom <= (bottomNav?.top ?? innerHeight) &&
      rect.width >= 44 && rect.height >= 44 && !!hit && element.contains(hit);
  })).toBe(true);
}

for (const theme of ["light", "dark"] as const) {
  test(
    "wrong-meaning feedback supports a retry without inventing a correction (" +
      theme +
      ")",
    async ({ page, context }, testInfo) => {
      const baseURL = testInfo.project.use.baseURL!;
      await context.addCookies([
        {
          name: "vocanova_session",
          value: "sense-" + randomUUID(),
          url: baseURL,
        },
        {
          name: "vocanova_csrf",
          value: "sense-csrf-" + randomUUID(),
          url: baseURL,
        },
      ]);
      await page.addInitScript((selectedTheme) => {
        window.localStorage.setItem("vocanova:theme-preference", selectedTheme);
      }, theme);
      await page.goto("/discover/ordering-at-a-cafe/pour#sentence-practice");
      await expect(page.locator("html")).toHaveAttribute("data-theme", theme);
      await page.getByRole("button", { name: /Save pour:/ }).click();

      const original = "I pour my energy into my work.";
      const explanation =
        "Your sentence uses a figurative meaning of pour. This exercise is about making a liquid flow.";
      const tip = "Try a new sentence about pouring a drink.";
      await page.route("**/api/v1/learner-sentences", async (route) => {
        expect(route.request().postDataJSON()).toMatchObject({
          sentenceText: original,
        });
        await route.fulfill({
          contentType: "application/json",
          body: JSON.stringify({
            feedbackId: randomUUID(),
            sentenceId: randomUUID(),
            attemptId: randomUUID(),
            targetWordId: randomUUID(),
            processingStatus: "completed",
            status: "incorrect",
            originalSentence: original,
            correctedSentence: null,
            headline: "Try the meaning for liquids",
            explanation,
            improvementTip: tip,
            targetWordUsedCorrectly: false,
            grammarAcceptable: true,
            meaningClear: true,
            naturalness: "natural",
            missionCompleted: true,
            canRetry: false,
            reported: false,
          }),
        });
      });

      const feedback = page
        .getByRole("heading", { name: "Practice with pour" })
        .locator("..");
      const textarea = feedback.getByRole("textbox", {
        name: "Write a sentence using pour",
      });
      await textarea.fill(original);
      await feedback.getByRole("button", { name: "Check my sentence" }).click();

      await expect(
        feedback.getByRole("status", { name: "Feedback result: Incorrect" }),
      ).toContainText(explanation);
      await expect(feedback.getByText(tip, { exact: true })).toBeVisible();
      await expect(
        feedback.getByText("Corrected sentence", { exact: true }),
      ).toHaveCount(0);
      await expect(feedback.getByText("Mission completed: Yes")).toBeVisible();
      await expect(
        feedback.getByRole("button", { name: "Report a problem" }),
      ).toBeEnabled();
      expect(
        await page.evaluate(() => document.documentElement.scrollWidth),
      ).toBeLessThanOrEqual(page.viewportSize()!.width);
      // Saving refreshes streamed route metadata; scan the settled document.
      await expect(page).toHaveTitle("Vocanova");
      // Check actual focused controls independently of an arbitrary reading
      // scroll position where a sticky header naturally clips earlier content.
      await expectKeyboardTargetClear(page, page.getByRole("button", { name: "Edit knowledge and note", exact: true }));
      await expectKeyboardTargetClear(page, page.locator("summary").filter({ hasText: "Optional writing" }));
      await expectKeyboardTargetClear(page, feedback.getByRole("button", { name: "Revise sentence", exact: true }));
      await page.evaluate(() => window.scrollTo({ top: 0, behavior: "instant" }));
      const { criticalOrSerious } = await scanForAxeViolations(page);
      expect(
        criticalOrSerious,
        formatViolations(criticalOrSerious).join("\n"),
      ).toEqual([]);
      await feedback
        .getByRole("status", { name: "Feedback result: Incorrect" })
        .scrollIntoViewIfNeeded();
      await page.screenshot({
        path: testInfo.outputPath("wrong-meaning-" + theme + ".png"),
        fullPage: false,
      });

      // A nullable correction must still leave the original available for revision.
      const revise = feedback.getByRole("button", { name: "Revise sentence" });
      await revise.focus();
      await page.keyboard.press("Enter");
      await expect(textarea).toBeFocused();
      await expect(textarea).toHaveValue(original);
      const previous = feedback.getByRole("complementary", {
        name: "Previous feedback",
      });
      await expect(previous).toContainText(explanation);
      await expect(previous.getByText(/Suggested revision:/)).toHaveCount(0);
    },
  );
}
