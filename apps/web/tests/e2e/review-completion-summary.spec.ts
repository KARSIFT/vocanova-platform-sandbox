import { expect, test, type Page, type TestInfo } from "@playwright/test";

async function seedReviewFixture(
  page: Page,
  count: number,
  testInfo: TestInfo,
  reviewTarget = count,
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
    {
      name: "e2e_daily_review_target",
      value: String(reviewTarget),
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
    await expect(
      page.getByRole("heading", { name: "Review complete", level: 2 }),
    ).toBeFocused();
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
    await expect(
      page.getByRole("heading", { name: "Review complete", level: 2 }),
    ).toHaveCount(0);

    await page.unroute("**/api/v1/reviews/submissions");
    const rating = page.getByRole("button", { name: "Good" });
    await rating.focus();
    await rating.press("Enter");
    await expect(page.getByText("You reviewed 1 word.")).toBeVisible();
    await expect(
      page.getByRole("heading", { name: "Review complete", level: 2 }),
    ).toBeFocused();
  });

  test("focuses completion only after a failed queue refresh succeeds", async ({
    page,
  }, testInfo) => {
    await seedReviewFixture(page, 1, testInfo, 2);
    await page.goto("/reviews");
    let refreshCount = 0;
    let releaseRefresh!: () => void;
    const delayedRefresh = new Promise<void>((resolve) => {
      releaseRefresh = resolve;
    });
    await page.route("**/api/v1/reviews/due?limit=*", async (route) => {
      refreshCount += 1;
      if (refreshCount === 1) {
        await route.fulfill({
          status: 500,
          contentType: "application/json",
          body: JSON.stringify({ error: "temporary_failure" }),
        });
        return;
      }
      await delayedRefresh;
      await route.fulfill({
        contentType: "application/json",
        body: JSON.stringify({ items: [], totalCount: 0 }),
      });
    });
    await submitCurrentReview(page, 1);
    const retry = page.getByRole("button", { name: "Retry loading reviews" });
    await expect(retry).toBeVisible();
    const completion = page.getByRole("heading", {
      name: "Review complete",
      level: 2,
    });
    await expect(completion).toHaveCount(0);
    await retry.focus();
    await retry.press("Enter");
    try {
      await expect.poll(() => refreshCount).toBe(2);
      await expect(page.getByText("Loading next reviews…")).toBeVisible();
      await expect(completion).toHaveCount(0);
    } finally {
      releaseRefresh();
    }
    await expect(completion).toBeFocused();
    await expect(page.getByText("You reviewed 1 word.")).toBeVisible();
  });

  test("an initially empty queue does not claim session completion", async ({
    page,
  }, testInfo) => {
    await seedReviewFixture(page, 0, testInfo, 1);
    await page.goto("/reviews");
    const caughtUp = page.getByRole("heading", {
      name: "You're all caught up",
      level: 2,
    });
    await expect(caughtUp).toBeVisible();
    await expect(caughtUp).not.toBeFocused();
    await expect(page.getByText(/You reviewed \d+ word/)).toHaveCount(0);
    await expect(
      page.getByRole("heading", { name: "Review complete", level: 2 }),
    ).toHaveCount(0);
  });

  test("an emptied stale queue focuses caught-up state without crediting a review", async ({
    page,
  }, testInfo) => {
    await seedReviewFixture(page, 1, testInfo);
    await page.goto("/reviews");
    await page.route("**/api/v1/reviews/submissions", async (route) => {
      await route.fulfill({
        status: 404,
        contentType: "application/json",
        body: JSON.stringify({ error: "saved_word_not_found" }),
      });
    });
    await page.route("**/api/v1/reviews/due?limit=*", async (route) => {
      await route.fulfill({
        contentType: "application/json",
        body: JSON.stringify({ items: [], totalCount: 0 }),
      });
    });
    await submitCurrentReview(page, 1);
    await expect(
      page.getByRole("heading", { name: "You're all caught up", level: 2 }),
    ).toBeFocused();
    await expect(page.getByText(/You reviewed \d+ word/)).toHaveCount(0);
    await expect(
      page.getByRole("heading", { name: "Review complete", level: 2 }),
    ).toHaveCount(0);
  });
});
