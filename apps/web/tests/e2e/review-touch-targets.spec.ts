import { randomUUID } from "node:crypto";

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
    randomUUID(),
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
  for (const theme of ["light", "dark"] as const) {
    test(`correct multiple-choice rating text stays inside each touch target in ${theme} mode`, async ({
      page,
    }, testInfo) => {
      await page.emulateMedia({ colorScheme: theme });
      await seedReviewFixture(page, 4, testInfo);
      await page.context().addCookies([
        { name: "vocanova_theme", value: theme, domain: "127.0.0.1", path: "/" },
      ]);
      await page.goto("/reviews");
      await expect(page.locator("html")).toHaveAttribute("data-theme", theme);
      await page
        .getByRole("button", { name: "noun — definition for review word 1" })
        .click();

      for (const rating of ["Hard", "Good", "Easy"]) {
        const button = page.getByRole("button", { name: rating, exact: true });
        await expectMinimumTouchHeight(button);
        const geometry = await button.evaluate((element) => {
          const buttonBox = element.getBoundingClientRect();
          const walker = document.createTreeWalker(element, NodeFilter.SHOW_TEXT);
          const textBoxes: { left: number; right: number; top: number; bottom: number }[] = [];
          while (walker.nextNode()) {
            if (!walker.currentNode.textContent?.trim()) continue;
            const range = document.createRange();
            range.selectNodeContents(walker.currentNode);
            for (const rect of range.getClientRects()) {
              textBoxes.push({ left: rect.left, right: rect.right, top: rect.top, bottom: rect.bottom });
            }
          }
          return {
            button: { left: buttonBox.left, right: buttonBox.right, top: buttonBox.top, bottom: buttonBox.bottom },
            textBoxes,
          };
        });
        expect(geometry.textBoxes.length).toBeGreaterThan(0);
        for (const rect of geometry.textBoxes) {
          expect(rect.left, `${rating} text left edge`).toBeGreaterThanOrEqual(geometry.button.left - 0.5);
          expect(rect.right, `${rating} text right edge`).toBeLessThanOrEqual(geometry.button.right + 0.5);
          expect(rect.top, `${rating} text top edge`).toBeGreaterThanOrEqual(geometry.button.top - 0.5);
          expect(rect.bottom, `${rating} text bottom edge`).toBeLessThanOrEqual(geometry.button.bottom + 0.5);
        }
      }
      await expect(page.getByRole("button", { name: "Again", exact: true })).toHaveCount(0);
      await expectNoHorizontalOverflow(page);
    });
  }

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
