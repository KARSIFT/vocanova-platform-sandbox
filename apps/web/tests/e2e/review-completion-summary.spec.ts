import { expect, test, type Page, type TestInfo } from "@playwright/test";

async function seedReviewFixture(
  page: Page,
  count: number,
  testInfo: TestInfo,
) {
  // The mock keeps review state in a process-wide map keyed by this cookie.
  // Every Playwright project and retry therefore needs its own key: the three
  // accessibility projects share the same mock server, and a retry may begin
  // after its earlier attempt has consumed some fixture cards.
  const sessionId = [
    "review-summary",
    testInfo.project.name,
    testInfo.testId,
    `retry-${testInfo.retry}`,
  ]
    .map(encodeURIComponent)
    .join("-");

  await page.context().addCookies([
    {
      name: "vocanova_session",
      value: sessionId,
      domain: "127.0.0.1",
      path: "/",
    },
    {
      name: "vocanova_csrf",
      value: "review-summary-csrf",
      domain: "127.0.0.1",
      path: "/",
    },
    {
      name: "e2e_review_fixture_count",
      value: String(count),
      domain: "127.0.0.1",
      path: "/",
    },
  ]);
}

async function submitCurrentReview(page: Page, index: number) {
  const showAnswer = page.getByRole("button", { name: "Show answer" });
  if (await showAnswer.isVisible()) {
    await showAnswer.click();
  } else {
    await page
      .getByRole("button", {
        name: `noun — definition for review word ${index}`,
      })
      .click();
  }
  await page.getByRole("button", { name: "Good" }).click();
}

test.describe("Review completion summary", () => {
  test("counts a paginated 51-card session after server confirmations", async ({
    page,
  }, testInfo) => {
    await seedReviewFixture(page, 51, testInfo);
    await page.goto("/reviews");

    for (let index = 1; index <= 51; index += 1) {
      await expect(
        page.getByRole("heading", { name: `Review word ${index}`, level: 2 }),
      ).toBeVisible();
      await submitCurrentReview(page, index);
    }

    await expect(
      page.getByRole("heading", { name: "Review complete", level: 2 }),
    ).toBeVisible();
    await expect(page.getByText("You reviewed 51 words.")).toBeVisible();
  });

  test("does not count a rejected submission before its successful retry", async ({
    page,
  }, testInfo) => {
    await seedReviewFixture(page, 1, testInfo);
    await page.goto("/reviews");
    await page.getByRole("button", { name: "Show answer" }).click();

    await page.route("**/api/v1/reviews/submissions", async (route) => {
      await route.fulfill({
        status: 500,
        contentType: "application/json",
        body: JSON.stringify({ error: "temporary_failure" }),
      });
    });
    await page.getByRole("button", { name: "Good" }).click();
    await expect(page.getByText("HTTP 500")).toBeVisible();
    await expect(page.getByText(/You reviewed \d+ word/)).toHaveCount(0);

    await page.unroute("**/api/v1/reviews/submissions");
    await page.getByRole("button", { name: "Good" }).click();
    await expect(page.getByText("You reviewed 1 word.")).toBeVisible();
  });
});
