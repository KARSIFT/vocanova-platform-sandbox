import { randomUUID } from "node:crypto";
import { expect, test, type Locator, type Page } from "@playwright/test";
import { startReviewLesson } from "./review-readiness";
async function target(control: Locator) {
  await expect(control).toBeVisible();
  const box = await control.boundingBox();
  expect(box?.height).toBeGreaterThanOrEqual(44);
  const inside = await control.evaluate((element) => {
    const rect = element.getBoundingClientRect();
    const walker = document.createTreeWalker(element, NodeFilter.SHOW_TEXT);
    const boxes: DOMRect[] = [];
    while (walker.nextNode()) {
      if (!walker.currentNode.textContent?.trim()) continue;
      const range = document.createRange();
      range.selectNodeContents(walker.currentNode);
      boxes.push(...range.getClientRects());
    }
    return boxes.every(
      (box) =>
        box.left >= rect.left - 0.5 &&
        box.right <= rect.right + 0.5 &&
        box.top >= rect.top - 0.5 &&
        box.bottom <= rect.bottom + 0.5,
    );
  });
  expect(inside).toBe(true);
}
async function setup(page: Page, theme = "light") {
  await page.context().addCookies([
    {
      name: "vocanova_session",
      value: randomUUID(),
      url: test.info().project.use.baseURL!,
    },
    {
      name: "vocanova_csrf",
      value: "review-touch",
      url: test.info().project.use.baseURL!,
    },
    {
      name: "e2e_review_fixture_count",
      value: "4",
      url: test.info().project.use.baseURL!,
    },
    {
      name: "vocanova_theme",
      value: theme,
      url: test.info().project.use.baseURL!,
    },
  ]);
  await page.goto("/reviews");
  await target(page.getByRole("button", { name: "Start review", exact: true }));
  await startReviewLesson(page);
}
async function noOverflow(page: Page) {
  expect(
    await page.evaluate(() => document.documentElement.scrollWidth),
  ).toBeLessThanOrEqual(page.viewportSize()!.width);
}
for (const theme of ["light", "dark"] as const) {
  test(`meaning labels and confirmed-answer continuation fit their targets in ${theme}`, async ({
    page,
  }) => {
    await setup(page, theme);
    await expect(page.locator("html")).toHaveAttribute("data-theme", theme);
    const options = page.getByRole("radio");
    await expect(options).toHaveCount(4);
    for (const option of await options.all())
      await target(option.locator(".."));
    await options.first().check();
    await target(
      page.getByRole("button", { name: "Check answer", exact: true }),
    );
    await target(
      page.getByRole("button", { name: "Skip for now", exact: true }),
    );
    await page
      .getByRole("button", { name: "Check answer", exact: true })
      .click();
    await target(page.getByRole("button", { name: "Continue", exact: true }));
    await noOverflow(page);
  });
}
test("incorrect-answer continuation and unconfirmed-answer recovery meet the 44px minimum", async ({
  page,
}) => {
  await setup(page);
  await page
    .getByRole("radio", { name: "A different meaning.", exact: true })
    .check();
  await page.route(
    "**/api/v1/daily-lessons/*/actions",
    (route) => route.abort("failed"),
    { times: 1 },
  );
  await page.getByRole("button", { name: "Check answer", exact: true }).click();
  await target(
    page.getByRole("button", { name: "Retry same answer", exact: true }),
  );
  await target(
    page.getByRole("button", { name: "Load saved progress", exact: true }),
  );
  await page
    .getByRole("button", { name: "Retry same answer", exact: true })
    .click();
  await expect(
    page.getByRole("status").filter({ hasText: /try this one again later/ }),
  ).toBeVisible();
  await target(page.getByRole("button", { name: "Continue", exact: true }));
  await noOverflow(page);
});
