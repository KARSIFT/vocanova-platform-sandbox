import { randomUUID } from "node:crypto";

import { expect, test, type BrowserContext } from "@playwright/test";

import {
  getEnabledReviewPromptControl,
  getTerminalReviewHeading,
  waitForReviewReadiness,
} from "./review-readiness";

async function seedReviewFixture(
  context: BrowserContext,
  sessionId: string,
  count: number,
  baseURL: string,
): Promise<void> {
  await context.addCookies([
    { name: "vocanova_session", value: sessionId, url: baseURL },
    {
      name: "vocanova_csrf",
      value: `review-readiness-csrf-${randomUUID()}`,
      url: baseURL,
    },
    { name: "e2e_review_fixture_count", value: String(count), url: baseURL },
    { name: "e2e_daily_review_target", value: String(count), url: baseURL },
  ]);
}

async function submitReadyReview(
  page: Parameters<typeof waitForReviewReadiness>[0],
) {
  const promptControl = getEnabledReviewPromptControl(page);
  await expect(promptControl).toBeVisible();
  await promptControl.check();
  await page.getByRole("button", { name: "Check answer", exact: true }).click();
  await page.getByRole("button", { name: "Continue", exact: true }).click();
}

test.describe("Review readiness", () => {
  test("waits for live prompt controls and terminal review state", async ({
    page,
    context,
  }, testInfo) => {
    const baseURL = testInfo.project.use.baseURL;
    if (!baseURL) {
      throw new Error(
        "Expected the Playwright project to configure use.baseURL.",
      );
    }

    const interactiveSession = `review-readiness-controls-${randomUUID()}`;
    await seedReviewFixture(context, interactiveSession, 4, baseURL);
    await page.goto("/reviews");
    expect(await waitForReviewReadiness(page)).toBe("prompt");

    const multipleChoiceOption = page
      .getByRole("group", { name: /^Meaning of / })
      .getByRole("radio", { disabled: false })
      .first();
    await expect(multipleChoiceOption).toBeVisible();
    await submitReadyReview(page);

    expect(await waitForReviewReadiness(page)).toBe("prompt");
    await expect(
      page.getByRole("radio", {
        name: "definition for review word 2",
        disabled: false,
      }),
    ).toBeVisible();

    const terminalSession = `review-readiness-terminal-${randomUUID()}`;
    // The mock accepts daily targets from 5 through 100, matching the real
    // settings constraint. Complete the smallest valid target so a reload
    // can exercise the page-level target-complete state honestly.
    await seedReviewFixture(context, terminalSession, 5, baseURL);
    await page.goto("/reviews");
    for (let completed = 0; completed < 5; completed += 1) {
      expect(await waitForReviewReadiness(page)).toBe("prompt");
      await submitReadyReview(page);
    }

    expect(await waitForReviewReadiness(page)).toBe("terminal");
    await expect(getTerminalReviewHeading(page)).toHaveText(
      "You’ve finished your review",
    );
    await page.goto("/reviews");
    expect(await waitForReviewReadiness(page)).toBe("terminal");
    await expect(getTerminalReviewHeading(page)).toHaveText(
      "You’re ready for new words",
    );
    await page.goto("/home");
    await expect(
      page
        .getByRole("region", { name: "Review words", exact: true })
        .getByText(/^Reviewed today:/),
    ).toHaveText("Reviewed today: 5");
  });
});
