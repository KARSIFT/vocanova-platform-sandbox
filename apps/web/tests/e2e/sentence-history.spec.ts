import { expect, test } from "@playwright/test";

import { formatViolations, scanForAxeViolations } from "./axe-helper.js";

test.describe("Sentence history", () => {
  test("explains an empty history and points to a first practice", async ({
    page,
  }) => {
    await page.goto("/progress/sentences");

    await expect(
      page.getByRole("heading", {
        name: "Your sentence practice will appear here",
      }),
    ).toBeVisible();
    await expect(page.getByRole("link", { name: "Go to Home" })).toBeVisible();
  });

  test("shows a bounded page of feedback and older-sentence navigation", async ({
    page,
  }, testInfo) => {
    await page.context().addCookies([
      {
        name: "e2e_sentence_history_fixture",
        value: "paginated",
        url: testInfo.project.use.baseURL!,
      },
    ]);
    await page.goto("/progress/sentences");

    await expect(page.getByText("Fixture sentence 1 uses pour naturally.")).toBeVisible();
    await expect(page.getByRole("link", { name: "View older sentences" })).toBeVisible();
    await page.getByRole("link", { name: "View older sentences" }).click();
    await expect(page.getByText("Fixture sentence 13 uses pour naturally.")).toBeVisible();
    await expect(page.getByText("Suggested revision")).toBeVisible();
  });

  test("offers a retry after the history service fails", async ({ page }, testInfo) => {
    await page.context().addCookies([
      {
        name: "e2e_sentence_history_fixture",
        value: "error",
        url: testInfo.project.use.baseURL!,
      },
    ]);
    await page.goto("/progress/sentences");
    await expect(
      page.getByRole("heading", {
        name: "We couldn't load your sentence history",
      }),
    ).toBeVisible();

    await page.context().clearCookies({ name: "e2e_sentence_history_fixture" });
    await page.getByRole("button", { name: "Try again" }).click();
    await expect(
      page.getByRole("heading", {
        name: "Your sentence practice will appear here",
      }),
    ).toBeVisible();
  });

  test("keeps a 300-character unbroken sentence within every supported layout", async ({
    page,
  }, testInfo) => {
    await page.context().addCookies([
      {
        name: "e2e_sentence_history_fixture",
        value: "unbroken",
        url: testInfo.project.use.baseURL!,
      },
    ]);
    await page.goto("/progress/sentences");
    await expect(page.getByText("a".repeat(300), { exact: true })).toBeVisible();

    const documentWidth = await page.evaluate(
      () => document.documentElement.scrollWidth,
    );
    expect(documentWidth).toBeLessThanOrEqual(page.viewportSize()!.width);

    const { criticalOrSerious } = await scanForAxeViolations(page);
    expect(
      criticalOrSerious,
      formatViolations(criticalOrSerious).join("\n"),
    ).toEqual([]);
  });
});
