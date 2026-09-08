import {
  expect,
  test,
  type Locator,
  type Page,
  type TestInfo,
} from "@playwright/test";

async function seedReviewFixture(
  page: Page,
  count: number,
  testInfo: TestInfo,
) {
  const sessionId = [
    "review-touch-targets",
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
      value: "review-touch-targets-csrf",
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

async function expectMinimumTouchHeight(locator: Locator) {
  await expect(locator).toBeVisible();
  const box = await locator.boundingBox();
  expect(
    box?.height,
    "Expected a rendered touch target.",
  ).toBeGreaterThanOrEqual(44);
}

async function expectNoHorizontalOverflow(page: Page) {
  expect(
    await page.evaluate(() => document.documentElement.scrollWidth),
  ).toBeLessThanOrEqual(page.viewportSize()!.width);
}

test.describe("Review touch targets", () => {
  test("self-check reveal and rating controls meet the 44px minimum", async ({
    page,
  }, testInfo) => {
    await seedReviewFixture(page, 1, testInfo);
    await page.goto("/reviews");

    const showAnswer = page.getByRole("button", { name: "Show answer" });
    await expectMinimumTouchHeight(showAnswer);
    await showAnswer.click();

    for (const rating of ["Again", "Hard", "Good", "Easy"]) {
      await expectMinimumTouchHeight(
        page.getByRole("button", { name: rating }),
      );
    }
    await expectNoHorizontalOverflow(page);
  });

  test("multiple-choice answers and the incorrect-answer continuation meet the 44px minimum", async ({
    page,
  }, testInfo) => {
    await seedReviewFixture(page, 4, testInfo);
    await page.goto("/reviews");

    const options = page.getByRole("button", {
      name: /noun — definition for review word/,
    });
    await expect(options).toHaveCount(4);
    for (let index = 0; index < 4; index += 1) {
      await expectMinimumTouchHeight(options.nth(index));
    }

    await page
      .getByRole("button", { name: "noun — definition for review word 2" })
      .click();
    await expectMinimumTouchHeight(
      page.getByRole("button", { name: "Continue" }),
    );
    await expectNoHorizontalOverflow(page);
  });

  test("retrying a failed queue refresh meets the 44px minimum", async ({
    page,
  }, testInfo) => {
    await seedReviewFixture(page, 1, testInfo);
    await page.goto("/reviews");
    await page.getByRole("button", { name: "Show answer" }).click();

    await page.route("**/api/v1/reviews/due?limit=*", async (route) => {
      await route.fulfill({
        status: 500,
        contentType: "application/json",
        body: JSON.stringify({ error: "temporary_failure" }),
      });
    });
    await page.getByRole("button", { name: "Good" }).click();

    await expectMinimumTouchHeight(
      page.getByRole("button", { name: "Retry loading reviews" }),
    );
    await expectNoHorizontalOverflow(page);
  });
});
