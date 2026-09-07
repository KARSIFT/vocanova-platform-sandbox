import { expect, test } from "@playwright/test";

async function seedReviewFixture(
  page: import("@playwright/test").Page,
  count: number,
) {
  await page.context().addCookies([
    {
      name: "vocanova_session",
      value: `review-summary-${count}`,
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

async function submitCurrentReview(
  page: import("@playwright/test").Page,
  index: number,
) {
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
  }) => {
    await seedReviewFixture(page, 51);
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
  }) => {
    await seedReviewFixture(page, 1);
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
