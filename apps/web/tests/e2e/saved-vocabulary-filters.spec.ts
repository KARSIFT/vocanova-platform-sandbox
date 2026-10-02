import { randomUUID } from "node:crypto";
import { expect, test, type Page } from "@playwright/test";
import { formatViolations, scanForAxeViolations } from "./axe-helper";

const main = (page: Page) => page.getByRole("main");
const filters = (page: Page) =>
  main(page).getByRole("search", {
    name: "Saved vocabulary filters",
    exact: true,
  });
const results = (page: Page) =>
  main(page).getByRole("list", {
    name: "Saved vocabulary results",
    exact: true,
  });

test.beforeEach(async ({ context, baseURL }) => {
  if (!baseURL) throw new Error("Missing app URL");
  await context.addCookies([
    { name: "vocanova_session", value: randomUUID(), url: baseURL },
    { name: "vocanova_csrf", value: randomUUID(), url: baseURL },
    { name: "e2e_saved_words", value: "filters", url: baseURL },
  ]);
});

for (const theme of ["light", "dark"] as const) {
  test(`saved filters show full counts, preserve pagination and report real review stages (${theme})`, async ({
    page,
    context,
    baseURL,
  }, testInfo) => {
    await context.addCookies([
      { name: "vocanova_theme", value: theme, url: baseURL! },
    ]);
    await page.goto("/words");
    await expect(
      main(page).getByText("Showing 20 of 26 saved meanings.", { exact: true }),
    ).toBeVisible();
    await expect(results(page).getByRole("listitem")).toHaveCount(20);
    await filters(page)
      .getByRole("searchbox", {
        name: "Search saved words and meanings",
        exact: true,
      })
      .fill("  SAVED COLLECTION  ");
    await filters(page)
      .getByRole("combobox", { name: "Review stage", exact: true })
      .selectOption("new");
    const due = filters(page).getByRole("checkbox", {
      name: "Due for review now",
      exact: true,
    });
    await due.focus();
    await page.keyboard.press("Space");
    await expect(due).toBeChecked();
    await page.keyboard.press("Tab");
    await expect(
      filters(page).getByRole("button", { name: "Apply filters", exact: true }),
    ).toBeFocused();
    await page.keyboard.press("Enter");
    await expect(
      main(page).getByText("Showing 20 of 24 matching meanings.", {
        exact: true,
      }),
    ).toBeVisible();
    await expect(filters(page).getByRole("searchbox")).toHaveValue(
      "SAVED COLLECTION",
    );
    await expect(results(page).getByText("New", { exact: true })).toHaveCount(
      20,
    );
    await expect(
      results(page).getByText("Due now", { exact: true }),
    ).toHaveCount(20);
    await expect(
      results(page).getByText("Not due now", { exact: true }),
    ).toHaveCount(0);
    const more = main(page).getByRole("link", {
      name: "More saved words",
      exact: true,
    });
    const moreURL = new URL((await more.getAttribute("href"))!, baseURL);
    expect(moreURL.searchParams.get("q")).toBe("SAVED COLLECTION");
    expect(moreURL.searchParams.get("stage")).toBe("new");
    expect(moreURL.searchParams.get("due")).toBe("true");
    expect(moreURL.searchParams.get("after")).toBeTruthy();
    await more.click();
    await expect(
      main(page).getByText("Showing 4 of 24 matching meanings.", {
        exact: true,
      }),
    ).toBeVisible();
    await expect(results(page).getByRole("listitem")).toHaveCount(4);
    await expect(
      main(page).getByRole("link", { name: "More saved words", exact: true }),
    ).toHaveCount(0);
    const back = main(page).getByRole("link", {
      name: "Back to newest",
      exact: true,
    });
    const backURL = new URL((await back.getAttribute("href"))!, baseURL);
    expect(backURL.searchParams.has("after")).toBe(false);
    expect(backURL.searchParams.get("stage")).toBe("new");
    expect(backURL.searchParams.get("due")).toBe("true");
    await back.click();
    await expect(
      main(page).getByText("Showing 20 of 24 matching meanings.", {
        exact: true,
      }),
    ).toBeVisible();
    await filters(page).getByRole("searchbox").fill("");
    await filters(page)
      .getByRole("combobox", { name: "Review stage", exact: true })
      .selectOption("reviewing");
    await filters(page)
      .getByRole("checkbox", { name: "Due for review now", exact: true })
      .uncheck();
    await filters(page)
      .getByRole("button", { name: "Apply filters", exact: true })
      .click();
    await expect(
      main(page).getByText("Showing 1 of 1 matching meaning.", {
        exact: true,
      }),
    ).toBeVisible();
    await expect(
      results(page).getByRole("heading", {
        name: "Fixture reviewing word",
        exact: true,
      }),
    ).toBeVisible();
    await expect(
      results(page).getByText("Reviewing", { exact: true }),
    ).toBeVisible();
    await expect(
      results(page).getByText("Not due now", { exact: true }),
    ).toBeVisible();
    await expect(page.locator("html")).toHaveAttribute("data-theme", theme);
    const scan = await scanForAxeViolations(page);
    expect(
      scan.criticalOrSerious,
      formatViolations(scan.criticalOrSerious).join("\n"),
    ).toEqual([]);
    expect(
      await page.evaluate(() => document.documentElement.scrollWidth),
    ).toBeLessThanOrEqual(page.viewportSize()!.width);
    await page.screenshot({
      path: testInfo.outputPath(`saved-filters-${theme}.png`),
      fullPage: true,
    });
    await filters(page)
      .getByRole("link", { name: "Clear filters", exact: true })
      .click();
    await expect(page).toHaveURL(/\/words$/);
    await expect(filters(page).getByRole("searchbox")).toHaveValue("");
    await expect(
      filters(page).getByRole("combobox", {
        name: "Review stage",
        exact: true,
      }),
    ).toHaveValue("");
    await expect(
      main(page).getByText("Showing 20 of 26 saved meanings.", { exact: true }),
    ).toBeVisible();
  });
}

test("a literal unmatched search is distinct from having no saved vocabulary", async ({
  page,
}) => {
  await page.goto("/words?q=%25_");
  await expect(
    main(page).getByRole("heading", {
      name: "No saved words match these filters",
      exact: true,
    }),
  ).toBeVisible();
  await expect(
    main(page).getByRole("heading", {
      name: "Your vocabulary starts here",
      exact: true,
    }),
  ).toHaveCount(0);
  await expect(results(page)).toHaveCount(0);
  await expect(filters(page).getByRole("searchbox")).toHaveValue("%_");
  await main(page)
    .getByRole("link", { name: "Show all saved words", exact: true })
    .click();
  await expect(
    main(page).getByText("Showing 20 of 26 saved meanings.", { exact: true }),
  ).toBeVisible();
});

test("a cross-filter cursor can restart the same filters without dropping them", async ({
  page,
  baseURL,
}) => {
  await page.goto("/words?q=saved%20collection&stage=new&due=true");
  const href = await main(page)
    .getByRole("link", { name: "More saved words", exact: true })
    .getAttribute("href");
  const stale = new URL(href!, baseURL);
  stale.searchParams.set("stage", "reviewing");
  await page.goto(stale.pathname + stale.search);
  await expect(
    main(page).getByRole("heading", {
      name: "Start from the first page",
      exact: true,
    }),
  ).toBeVisible();
  const restart = main(page).getByRole("link", {
    name: "Restart these results",
    exact: true,
  });
  const reset = new URL((await restart.getAttribute("href"))!, baseURL);
  expect(reset.searchParams.has("after")).toBe(false);
  expect(reset.searchParams.get("q")).toBe("saved collection");
  expect(reset.searchParams.get("stage")).toBe("reviewing");
  expect(reset.searchParams.get("due")).toBe("true");
  await restart.click();
  await expect(
    main(page).getByRole("heading", {
      name: "No saved words match these filters",
      exact: true,
    }),
  ).toBeVisible();
  await expect(
    main(page).getByRole("heading", {
      name: "Start from the first page",
      exact: true,
    }),
  ).toHaveCount(0);
});

test("an unavailable saved library retains filters and reloads without claiming it is empty", async ({
  page,
  context,
  baseURL,
}) => {
  await context.addCookies([
    { name: "e2e_saved_words", value: "unavailable", url: baseURL! },
  ]);
  await page.goto("/words?q=saved%20collection&stage=new&due=true");
  await expect(
    main(page).getByRole("heading", {
      name: "Saved vocabulary is unavailable",
      exact: true,
    }),
  ).toBeVisible();
  await expect(
    main(page).getByRole("heading", {
      name: "Your vocabulary starts here",
      exact: true,
    }),
  ).toHaveCount(0);
  await expect(
    main(page).getByRole("heading", {
      name: "No saved words match these filters",
      exact: true,
    }),
  ).toHaveCount(0);
  await expect(filters(page).getByRole("searchbox")).toHaveValue(
    "saved collection",
  );
  await expect(
    filters(page).getByRole("combobox", { name: "Review stage", exact: true }),
  ).toHaveValue("new");
  await expect(
    filters(page).getByRole("checkbox", {
      name: "Due for review now",
      exact: true,
    }),
  ).toBeChecked();
  await context.addCookies([
    { name: "e2e_saved_words", value: "filters", url: baseURL! },
  ]);
  await main(page)
    .getByRole("button", { name: "Try loading again", exact: true })
    .click();
  await expect(
    main(page).getByText("Showing 20 of 24 matching meanings.", {
      exact: true,
    }),
  ).toBeVisible();
});

test("oversized direct searches and unlisted stages ask for correction instead of hiding saved words", async ({
  page,
}) => {
  await page.goto(`/words?q=${"a".repeat(101)}`);
  await expect(
    main(page).getByRole("heading", {
      name: "Check your filters",
      exact: true,
    }),
  ).toBeVisible();
  await expect(main(page).getByRole("alert")).toContainText("100 characters");
  await expect(
    main(page).getByRole("heading", {
      name: "Your vocabulary starts here",
      exact: true,
    }),
  ).toHaveCount(0);
  await page.goto("/words?stage=not-a-stage");
  await expect(
    main(page).getByRole("heading", {
      name: "Check your filters",
      exact: true,
    }),
  ).toBeVisible();
  await filters(page)
    .getByRole("link", { name: "Clear filters", exact: true })
    .click();
  await expect(
    main(page).getByText("Showing 20 of 26 saved meanings.", { exact: true }),
  ).toBeVisible();
});
