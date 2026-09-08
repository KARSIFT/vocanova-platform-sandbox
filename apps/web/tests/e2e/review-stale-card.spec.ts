import { expect, test, type BrowserContext, type Page } from "@playwright/test";

async function seedReviewSession(
  context: BrowserContext,
  sessionId: string,
  fixtureCount?: number,
  reviewTarget?: number,
) {
  const origins = [
    "http://127.0.0.1:3000",
    "http://127.0.0.1:8080",
    "http://localhost:8080",
  ];
  await context.addCookies(
    origins.flatMap((url) => [
      { name: "vocanova_session", value: sessionId, url },
      { name: "vocanova_csrf", value: "review-stale-card-csrf", url },
      ...(fixtureCount === undefined
        ? []
        : [
            {
              name: "e2e_review_fixture_count",
              value: String(fixtureCount),
              url,
            },
            ...(reviewTarget === undefined
              ? []
              : [
                  {
                    name: "e2e_daily_review_target",
                    value: String(reviewTarget),
                    url,
                  },
                ]),
          ]),
    ]),
  );
}

async function submitFixtureReview(page: Page, index: number) {
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

function fixtureDueWord(index: number) {
  return {
    userWordId: `recovered-user-word-${index}`,
    meaningId: `recovered-meaning-${index}`,
    wordId: `recovered-word-${index}`,
    wordSlug: `recovered-word-${index}`,
    wordText: `Review word ${index}`,
    partOfSpeech: "noun",
    shortDefinition: `definition for review word ${index}`,
    status: "due",
    reviewStep: 0,
  };
}

test.describe("Review stale-card recovery", () => {
  test("reconciles a removed last card instead of retrying its rejected answer", async ({
    page,
    context,
  }, testInfo) => {
    test.skip(
      testInfo.project.name !== "home-desktop-1280",
      "The mutation flow is covered once against the production bundle.",
    );

    const sessionId = [
      "review-stale-card",
      testInfo.testId,
      `retry-${testInfo.retry}`,
    ]
      .map(encodeURIComponent)
      .join("-");
    // Server rendering uses the configured 127.0.0.1 API origin while the
    // production client bundle may use localhost. Give both origins the same
    // mock session so this remains one learner session across the two tabs.
    await seedReviewSession(context, sessionId);

    await page.goto("/discover/ordering-at-a-cafe/pour");
    await page.getByRole("button", { name: "Save" }).click();
    await expect(
      page.getByRole("button", { name: "Remove pour from saved words" }),
    ).toBeVisible();

    await page.goto("/reviews");
    await page.getByRole("button", { name: "Show answer" }).click();

    const otherTab = await context.newPage();
    await otherTab.goto("/discover/ordering-at-a-cafe/pour");
    await otherTab
      .getByRole("button", { name: "Remove pour from saved words" })
      .click();

    let rejectedSubmissionCount = 0;
    await page.route("**/api/v1/reviews/submissions", async (route) => {
      rejectedSubmissionCount += 1;
      await route.fulfill({
        status: 404,
        contentType: "application/json",
        body: JSON.stringify({
          title: "Not Found",
          status: 404,
          detail: "saved word not found",
        }),
      });
    });

    await page.getByRole("button", { name: "Good" }).click();

    await expect(
      page.getByRole("heading", { name: "You're all caught up", level: 2 }),
    ).toBeVisible();
    await expect(
      page.getByText("This word was removed. Your review list was updated."),
    ).toBeVisible();
    await expect(page.getByText(/You reviewed \d+ word/)).toHaveCount(0);
    await expect(page.getByRole("button", { name: "Good" })).toHaveCount(0);
    expect(rejectedSubmissionCount).toBe(1);
  });

  test("preserves earlier confirmed reviews when the next card was removed", async ({
    page,
    context,
  }, testInfo) => {
    test.skip(
      testInfo.project.name !== "home-desktop-1280",
      "The mutation flow is covered once against the production bundle.",
    );

    const sessionId = [
      "review-stale-card-prior-confirmed",
      testInfo.testId,
      `retry-${testInfo.retry}`,
    ]
      .map(encodeURIComponent)
      .join("-");
    await seedReviewSession(context, sessionId, 2, 2);
    await page.goto("/reviews");

    await submitFixtureReview(page, 1);
    await expect(
      page.getByRole("heading", { name: "Review word 2", level: 2 }),
    ).toBeVisible();

    let dueRefreshCount = 0;
    await page.route("**/api/v1/reviews/due?limit=*", async (route) => {
      dueRefreshCount += 1;
      await route.fulfill({
        contentType: "application/json",
        body: JSON.stringify({ items: [], totalCount: 0 }),
      });
    });
    await page.route("**/api/v1/reviews/submissions", async (route) => {
      await route.fulfill({
        status: 404,
        contentType: "application/json",
        body: JSON.stringify({
          title: "Not Found",
          status: 404,
          detail: "saved word not found",
        }),
      });
    });

    await submitFixtureReview(page, 2);

    await expect(
      page.getByRole("heading", { name: "Review complete", level: 2 }),
    ).toBeVisible();
    await expect(page.getByText("You reviewed 1 word.")).toBeVisible();
    await expect(
      page.getByText("This word was removed. Your review list was updated."),
    ).toBeVisible();
    expect(dueRefreshCount).toBe(1);
  });

  test("clears the removed-word status after later normal reviews", async ({
    page,
    context,
  }, testInfo) => {
    test.skip(
      testInfo.project.name !== "home-desktop-1280",
      "The mutation flow is covered once against the production bundle.",
    );

    const sessionId = [
      "review-stale-card-status-lifecycle",
      testInfo.testId,
      `retry-${testInfo.retry}`,
    ]
      .map(encodeURIComponent)
      .join("-");
    await seedReviewSession(context, sessionId, 1, 2);
    await page.goto("/reviews");

    let submissionCount = 0;
    await page.route("**/api/v1/reviews/submissions", async (route) => {
      submissionCount += 1;
      if (submissionCount === 1) {
        await route.fulfill({
          status: 404,
          contentType: "application/json",
          body: JSON.stringify({
            title: "Not Found",
            status: 404,
            detail: "saved word not found",
          }),
        });
        return;
      }
      await route.continue();
    });

    let dueRefreshCount = 0;
    await page.route("**/api/v1/reviews/due?limit=*", async (route) => {
      dueRefreshCount += 1;
      await route.fulfill({
        contentType: "application/json",
        body: JSON.stringify(
          dueRefreshCount === 1
            ? {
                items: [fixtureDueWord(2), fixtureDueWord(3)],
                totalCount: 2,
              }
            : { items: [], totalCount: 0 },
        ),
      });
    });

    await submitFixtureReview(page, 1);
    await expect(
      page.getByRole("heading", { name: "Review word 2", level: 2 }),
    ).toBeVisible();
    await expect(
      page.getByText("This word was removed. Your review list was updated."),
    ).toBeVisible();

    await submitFixtureReview(page, 2);
    await expect(
      page.getByRole("heading", { name: "Review word 3", level: 2 }),
    ).toBeVisible();
    await expect(
      page.getByText("This word was removed. Your review list was updated."),
    ).toHaveCount(0);

    await submitFixtureReview(page, 3);
    await expect(
      page.getByRole("heading", { name: "Review complete", level: 2 }),
    ).toBeVisible();
    await expect(page.getByText("You reviewed 2 words.")).toBeVisible();
    await expect(
      page.getByText("This word was removed. Your review list was updated."),
    ).toHaveCount(0);
    expect(dueRefreshCount).toBe(2);
  });
});
