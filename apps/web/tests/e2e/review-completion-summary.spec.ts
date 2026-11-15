import { randomUUID } from "node:crypto";
import { expect, test, type Page } from "@playwright/test";
import { answerReviewWord, startReviewLesson } from "./review-readiness";

async function seed(page: Page, count: number) {
  await page.context().addCookies([
    {
      name: "vocanova_session",
      value: randomUUID(),
      domain: "127.0.0.1",
      path: "/",
    },
    {
      name: "vocanova_csrf",
      value: "summary-csrf",
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
      value: String(count || 1),
      domain: "127.0.0.1",
      path: "/",
    },
  ]);
  await page.goto("/reviews");
}
const completed = (page: Page) =>
  page.getByRole("heading", {
    name: /^(?:Review saved|Review set complete|You’ve finished your review|All available reviews finished|Today’s review complete)$/,
    level: 1,
  });
const actions = "**/api/v1/daily-lessons/*/actions";

test.describe("Review completion summary", () => {
  test("counts 51 words in a goal-sized lesson only after confirmations", async ({
    page,
  }) => {
    test.setTimeout(90_000);
    await seed(page, 51);
    await startReviewLesson(page);
    for (let index = 1; index <= 51; index++) {
      await expect(
        page.getByRole("heading", {
          name: `Review word ${index}`,
          exact: true,
        }),
      ).toBeVisible();
      await answerReviewWord(page, `definition for review word ${index}`);
      if (index === 51) {
        await expect(completed(page)).toBeFocused();
        await expect(page.getByText("51 of 51 words reviewed today", { exact: true })).toBeVisible();
      } else {
        await expect(completed(page)).toHaveCount(0);
        await expect(page.getByText(`${index} of 51 words reviewed today`, { exact: true })).toBeVisible();
      }
    }
    await page.goto("/home");
    await expect(
      page
        .getByRole("region", { name: "Review words", exact: true })
        .getByText(/^Reviewed today:/),
    ).toHaveText("Reviewed today: 51 of 51");
  });

  test("does not count a rejected answer before its successful keyboard retry", async ({
    page,
  }) => {
    await seed(page, 1);
    await startReviewLesson(page);
    await page.route(
      actions,
      (route) =>
        route.fulfill({
          status: 500,
          contentType: "application/json",
          body: JSON.stringify({ detail: "temporary failure" }),
        }),
      { times: 1 },
    );
    await page
      .getByRole("radio", { name: "definition for review word 1", exact: true })
      .check();
    await page
      .getByRole("button", { name: "Check answer", exact: true })
      .click();
    await expect(page.getByRole("main").getByRole("alert")).toBeVisible();
    await expect(completed(page)).toHaveCount(0);
    for (const radio of await page.getByRole("radio").all())
      await expect(radio).toBeDisabled();
    const retry = page.getByRole("button", { name: "Retry same answer" });
    await expect(retry).toBeInViewport({ ratio: 1 });
    const alert = page.getByRole("main").getByRole("alert");
    await expect
      .poll(() =>
        alert.evaluate(
          (element) =>
            element === document.activeElement ||
            element.contains(document.activeElement),
        ),
      )
      .toBe(true);
    const retryBox = await retry.boundingBox();
    const actionsBox = await page
      .getByRole("region", { name: "Session actions" })
      .boundingBox();
    expect(retryBox!.y + retryBox!.height).toBeLessThanOrEqual(actionsBox!.y);
    await retry.focus();
    await retry.press("Enter");
    await expect(
      page.getByRole("status").filter({ hasText: /^Correct/ }),
    ).toBeVisible();
    await page.getByRole("button", { name: "Continue", exact: true }).click();
    await expect(completed(page)).toBeFocused();
    await expect(
      page.getByText("1 of 20 words reviewed today", { exact: true }),
    ).toBeVisible();
  });

  test("an unconfirmed continuation cannot announce completion", async ({
    page,
  }) => {
    await seed(page, 1);
    await startReviewLesson(page);
    await page
      .getByRole("radio", { name: "definition for review word 1", exact: true })
      .check();
    await page
      .getByRole("button", { name: "Check answer", exact: true })
      .click();
    await expect(
      page.getByRole("status").filter({ hasText: /^Correct/ }),
    ).toBeVisible();
    await page.route(actions, (route) => route.abort("failed"), { times: 1 });
    await page.getByRole("button", { name: "Continue", exact: true }).click();
    await expect(
      page.getByRole("button", { name: "Retry same answer" }),
    ).toBeVisible();
    await expect(completed(page)).toHaveCount(0);
    await page.getByRole("button", { name: "Retry same answer" }).click();
    await expect(completed(page)).toBeFocused();
    await expect(
      page.getByText("1 of 20 words reviewed today", { exact: true }),
    ).toBeVisible();
  });

  test("an initially empty queue does not claim session completion", async ({
    page,
  }) => {
    await seed(page, 0);
    const empty = page.getByRole("heading", {
      name: "You’re ready for new words",
      exact: true,
    });
    await expect(empty).toBeVisible();
    await expect(empty).not.toBeFocused();
    await expect(completed(page)).toHaveCount(0);
    await expect(page.getByText(/words? remembered for today/)).toHaveCount(0);
    await expect(
      page.getByRole("link", { name: "Discover words", exact: true }),
    ).toHaveAttribute("href", "/discovery");
  });
});
