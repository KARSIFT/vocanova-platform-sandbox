import { randomUUID } from "node:crypto";

import { expect, test } from "@playwright/test";

import { formatViolations, scanForAxeViolations } from "./axe-helper.js";

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
      await page.goto("/discover/ordering-at-a-cafe/pour");
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
