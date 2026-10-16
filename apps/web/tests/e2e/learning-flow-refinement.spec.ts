import { answerReviewWord, startReviewLesson } from "./review-readiness";
import { openOptionalWriting } from "./optional-writing";
import { randomUUID } from "node:crypto";

import { expect, test } from "@playwright/test";

test("shows sparse mission history in date order without inventing missing days", async ({
  page,
  context,
  baseURL,
}) => {
  if (!baseURL) throw new Error("A test app URL is required");
  await context.addCookies([
    { name: "vocanova_session", value: randomUUID(), url: baseURL },
    { name: "e2e_progress_history_fixture", value: "sparse", url: baseURL },
  ]);
  await page.goto("/progress");
  await expect(
    page.getByRole("heading", { name: "Recent activity" }),
  ).toBeVisible();
  const dates = page.locator("main time[datetime]");
  await expect(dates).toHaveCount(3);
  expect(
    await dates.evaluateAll((items) =>
      items.map((item) => item.getAttribute("datetime")),
    ),
  ).toEqual(["2026-09-09", "2026-09-10", "2026-09-12"]);
  for (const date of await dates.all()) {
    await expect(date).not.toHaveText("");
  }
});

test("keeps lesson completion simple and optional writing available afterward", async ({
  page,
  context,
  baseURL,
}) => {
  if (!baseURL) throw new Error("A test app URL is required");
  await context.addCookies([
    { name: "vocanova_session", value: randomUUID(), url: baseURL },
    { name: "vocanova_csrf", value: "flow-refinement-csrf", url: baseURL },
  ]);
  await page.goto("/vocabulary/pour");
  await page.getByRole("button", { name: /^Save pour:/ }).click();
  await page.goto("/review");
  await startReviewLesson(page);
  await answerReviewWord(page, "to make liquid flow into a container");
  await expect(
    page.getByRole("heading", { name: "Lesson complete", exact: true }),
  ).toBeVisible();
  await expect(
    page.getByText("1 word remembered for today.", { exact: true }),
  ).toBeVisible();
  await expect(page.getByRole("textbox")).toHaveCount(0);
  await page.goto("/words/uw-mean-pour");
  await openOptionalWriting(page);
  const input = page.getByRole("textbox", {
    name: "Write a sentence using pour",
    exact: true,
  });
  await input.fill("I pour tea every day.");
  await expect(
    page.getByRole("button", { name: "Check my sentence" }),
  ).toBeEnabled();
  expect(
    await page.evaluate(() => document.documentElement.scrollWidth),
  ).toBeLessThanOrEqual(page.viewportSize()!.width);
});
