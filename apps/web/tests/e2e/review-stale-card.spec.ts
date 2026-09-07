import { expect, test } from "@playwright/test";

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
    await context.addCookies(
      [
        "http://127.0.0.1:3000",
        "http://127.0.0.1:8080",
        "http://localhost:8080",
      ].flatMap((url) => [
        { name: "vocanova_session", value: sessionId, url },
        {
          name: "vocanova_csrf",
          value: "review-stale-card-csrf",
          url,
        },
      ]),
    );

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
});
