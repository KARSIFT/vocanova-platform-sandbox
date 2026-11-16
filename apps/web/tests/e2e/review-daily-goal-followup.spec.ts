import { randomUUID } from "node:crypto";
import { expect, test } from "@playwright/test";
import type { DailyLesson } from "@vocanova/api-client";

for (const totalWords of [10, 7]) {
  test(`daily progress uses the goal of 15 inside a frozen ${totalWords}-word set`, async ({ page, baseURL }) => {
    await page.context().addCookies([
      { name: "vocanova_session", value: `goal-${randomUUID()}`, url: baseURL! },
      { name: "vocanova_csrf", value: "goal-followup", url: baseURL! },
      { name: "e2e_review_fixture_count", value: "20", url: baseURL! },
    ]);
    const goal = { localDate: new Date().toISOString().slice(0, 10), timeZone: "UTC", target: 15, reviewed: 8, remaining: 7, complete: false };
    const lesson: DailyLesson = {
      id: "frozen-set", status: "in_progress", revision: 1,
      completedWords: totalWords === 10 ? 9 : 0, totalWords,
      firstAnswersCorrect: 0, questionsAnswered: 9, creditedToday: 0, goal,
      currentStep: { id: "frozen-question", kind: "meaning_choice", wordText: "wit", wordSlug: "wit", partOfSpeech: "noun", choices: ["a", "b", "c", "d"].map((id) => ({ id, text: `Definition ${id}` })) },
      feedback: null, canContinue: false,
      createdAt: "2026-11-15T12:00:00Z", updatedAt: "2026-11-15T12:00:00Z",
      attribution: { title: "Word sources", sourceUrl: "https://wordnet.princeton.edu/", licenseUrl: "https://wordnet.princeton.edu/license-and-commercial-use", wordnetLicense: "WordNet" },
    };
    await page.route("**/api/v1/daily-lessons", (route) => route.fulfill({ json: lesson }));
    await page.goto("/review");
    await page.getByRole("button", { name: "Start review", exact: true }).click();
    await expect(page.getByRole("heading", { name: "wit", exact: true })).toBeVisible();
    await expect(page.getByText("8 / 15", { exact: true })).toBeVisible();
    await expect(page.getByText(`${lesson.completedWords} / ${totalWords}`, { exact: true })).toHaveCount(0);
    const progress = page.getByRole("progressbar", { name: "Review progress", exact: true });
    await expect(progress).toHaveAttribute("max", "15");
    await expect(progress).toHaveAttribute("value", "8");
    await expect(page.getByText("8 of 15 words reviewed today", { exact: true })).toBeVisible();
    await expect(page.getByRole("radio")).toHaveCount(4);
  });
}
