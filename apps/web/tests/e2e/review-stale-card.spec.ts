import { randomUUID } from "node:crypto";
import { expect, test, type Page } from "@playwright/test";
import { answerReviewWord, startReviewLesson } from "./review-readiness";

async function seed(page: Page, count: number) {
  const url = test.info().project.use.baseURL!;
  await page.context().addCookies([
    { name: "vocanova_session", value: randomUUID(), url },
    { name: "vocanova_csrf", value: "stale-daily-csrf", url },
    { name: "e2e_review_fixture_count", value: String(count), url },
  ]);
  await page.goto("/reviews");
  await startReviewLesson(page);
}

// A second tab changes server progress. The old client must stop, reload the
// confirmed state, and avoid treating its rejected answer as another credit.
async function conflictFromOtherTab(page: Page, index: number, skip: boolean) {
  const other = await page.context().newPage();
  await other.goto(page.url());
  await expect(
    other.getByRole("heading", { name: `Review word ${index}`, exact: true }),
  ).toBeVisible();
  if (skip)
    await other
      .getByRole("button", { name: "Skip for now", exact: true })
      .click();
  else {
    await other
      .getByRole("radio", {
        name: `definition for review word ${index}`,
        exact: true,
      })
      .check();
    await other
      .getByRole("button", { name: "Check answer", exact: true })
      .click();
  }
  await expect(
    other.getByRole("button", { name: "Continue", exact: true }),
  ).toBeEnabled();
  let rejected = 0;
  page.on("response", (response) => {
    if (
      response.request().method() === "POST" &&
      response.url().includes("/actions") &&
      response.status() === 409
    )
      rejected++;
  });
  await page
    .getByRole("radio", { name: "A different meaning.", exact: true })
    .check();
  await page.getByRole("button", { name: "Check answer", exact: true }).click();
  await expect(page.getByRole("main").getByRole("alert")).toContainText(
    "Your lesson changed",
  );
  expect(rejected).toBe(1);
  await expect(
    page.getByRole("button", { name: "Retry same answer", exact: true }),
  ).toHaveCount(0);
  for (const radio of await page.getByRole("radio").all())
    await expect(radio).toBeDisabled();
  const load = page.getByRole("button", {
    name: "Load saved progress",
    exact: true,
  });
  await load.focus();
  await load.press("Enter");
  await expect(page.getByRole("main").getByRole("alert")).toHaveCount(0);
  await expect(
    page
      .getByRole("status")
      .filter({ hasText: skip ? /^Skipped for now/ : /^Correct/ }),
  ).toBeFocused();
  await other.close();
}

test.describe("Review stale-card recovery", () => {
  test("a last word skipped elsewhere gives no false daily credit", async ({
    page,
  }) => {
    await seed(page, 1);
    await conflictFromOtherTab(page, 1, true);
    await page.getByRole("button", { name: "Continue", exact: true }).click();
    await expect(
      page.getByRole("heading", { name: "Lesson complete", exact: true }),
    ).toBeFocused();
    await expect(
      page.getByText("0 words remembered for today.", { exact: true }),
    ).toBeVisible();
  });
  test("preserves earlier confirmed answers when the next word is skipped elsewhere", async ({
    page,
  }) => {
    await seed(page, 2);
    await answerReviewWord(page, "definition for review word 1");
    await conflictFromOtherTab(page, 2, true);
    await page.getByRole("button", { name: "Continue", exact: true }).click();
    await expect(
      page.getByRole("heading", { name: "Lesson complete", exact: true }),
    ).toBeFocused();
    await expect(
      page.getByText("1 word remembered for today.", { exact: true }),
    ).toBeVisible();
  });
  test("clears stale feedback before later normal answers and credits each word once", async ({
    page,
  }) => {
    await seed(page, 3);
    await conflictFromOtherTab(page, 1, false);
    await page.getByRole("button", { name: "Continue", exact: true }).click();
    await expect(
      page.getByRole("heading", { name: "Review word 2", exact: true }),
    ).toBeFocused();
    await expect(page.getByRole("main").getByRole("alert")).toHaveCount(0);
    await expect(
      page.getByRole("status").filter({ hasText: /^Correct/ }),
    ).toHaveCount(0);
    await answerReviewWord(page, "definition for review word 2");
    await answerReviewWord(page, "definition for review word 3");
    await expect(
      page.getByRole("heading", { name: "Lesson complete", exact: true }),
    ).toBeFocused();
    await expect(
      page.getByText("3 words remembered for today.", { exact: true }),
    ).toBeVisible();
    await expect(
      page.getByText("3 correct on the first try.", { exact: true }),
    ).toBeVisible();
  });
});
