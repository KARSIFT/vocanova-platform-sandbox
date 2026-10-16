import { randomUUID } from "node:crypto";
import { expect, test } from "@playwright/test";
import { answerReviewWord, startReviewLesson } from "./review-readiness";
import { openOptionalWriting } from "./optional-writing";

test("optional writing after a lesson restores its unfinished draft on a same-user reload", async ({
  page,
  context,
  baseURL,
}) => {
  await context.addCookies([
    { name: "vocanova_session", value: `draft-${randomUUID()}`, url: baseURL! },
    { name: "vocanova_csrf", value: randomUUID(), url: baseURL! },
  ]);
  await page.goto("/vocabulary/pour");
  await page.getByRole("button", { name: /^Save pour:/ }).click();
  await page.goto("/reviews");
  await startReviewLesson(page);
  await answerReviewWord(page, "to make liquid flow into a container");
  await expect(
    page.getByRole("heading", { name: "Lesson complete", exact: true }),
  ).toBeVisible();
  await expect(page.getByRole("textbox")).toHaveCount(0);
  await page
    .getByRole("link", { name: "Back to Home", exact: true })
    .last()
    .click();
  await page.getByRole("link", { name: "pour", exact: true }).click();
  await openOptionalWriting(page);
  const writing = page.getByRole("textbox", {
    name: "Write a sentence using pour",
    exact: true,
  });
  await writing.fill("I pour coffee at work.");
  // Verify both recovery inputs before leaving: text is saved synchronously,
  // while the fragment restores the editor and its collapsed parent group.
  await expect(page).toHaveURL(/#sentence-practice$/);
  const inputId = await writing.getAttribute("id");
  const attemptId = inputId!.replace(/^sentence-input-/, "");
  await expect
    .poll(() =>
      page.evaluate((expectedAttemptId) => {
        return Object.keys(sessionStorage)
          .filter((key) => key.startsWith("vocanova:sentence-feedback-draft:"))
          .some((key) => {
            const draft = JSON.parse(sessionStorage.getItem(key)!);
            return (
              draft.source === "word_detail" &&
              draft.attemptId === expectedAttemptId &&
              draft.sentence === "I pour coffee at work."
            );
          });
      }, attemptId),
    )
    .toBe(true);
  await page.reload();
  await expect(page).toHaveURL(/#sentence-practice$/);
  await expect(writing).toBeVisible();
  await expect(writing).toHaveValue("I pour coffee at work.");
});
