// The word map lives on Journey. Range links, retries and sign-in returns must
// keep the learner on Journey; the old /knowledge-map address stays an alias.

import { randomUUID } from "node:crypto";
import { expect, test, type Page, type TestInfo } from "@playwright/test";

import { formatViolations, scanForAxeViolations } from "./axe-helper";

async function learner(page: Page, cookies: Record<string, string> = {}) {
  const baseURL = test.info().project.use.baseURL!;
  await page.context().addCookies([
    { name: "vocanova_session", value: `map-${randomUUID()}`, url: baseURL },
    ...Object.entries(cookies).map(([name, value]) => ({
      name,
      value,
      url: baseURL,
    })),
  ]);
}

// Below 40rem the map is a picture with a range chooser; wider screens link
// every square.
const onPhone = (testInfo: TestInfo) =>
  testInfo.project.name.startsWith("mobile-");

test("the old map address opens the same range on Journey", async ({
  page,
}) => {
  await learner(page);
  await page.goto("/knowledge-map?band=7#range-words");
  await expect(page).toHaveURL(/\/discover\?band=7#range-words$/);
  await expect(
    page.getByRole("heading", { name: "Journey", level: 1 }),
  ).toBeVisible();
  await expect(
    page.getByRole("heading", { name: "Words 601–700", level: 2 }),
  ).toBeVisible();
});

test("range, paging and map links stay on Journey", async ({
  page,
}, testInfo) => {
  await learner(page);
  await page.goto("/discover?band=2");
  const ranges = page.getByRole("navigation", { name: "Word ranges", exact: true });
  if (onPhone(testInfo)) {
    await expect(ranges).toBeHidden();
    const choice = page.getByRole("combobox", { name: "Go to a range" });
    await expect(choice.locator("option")).toHaveCount(100);
    await expect(choice).toHaveValue("2");
  } else {
    await expect(ranges.getByRole("link")).toHaveCount(100);
    for (const link of await ranges.getByRole("link").all())
      await expect(link).toHaveAttribute(
        "href",
        /^\/discover\?band=\d+#range-words$/,
      );
    await expect(
      ranges.getByRole("link", { name: /^Words 101 to 200,/ }),
    ).toHaveAttribute("aria-current", "true");
  }
  const paging = page.getByRole("navigation", { name: "Adjacent word ranges" });
  await expect(
    paging.getByRole("link", { name: "Previous 100", exact: true }),
  ).toHaveAttribute("href", "/discover?band=1#range-words");
  await paging.getByRole("link", { name: "Next 100", exact: true }).click();
  await expect(page).toHaveURL(/\/discover\?band=3#range-words$/);
  await expect(
    page.getByRole("heading", { name: "Words 201–300", level: 2 }),
  ).toBeVisible();
  // One filled action: discovering the next word.
  await expect(page.getByRole("main").locator(".vn-btn-primary")).toHaveCount(1);
  await expect(
    page.getByRole("link", { name: "Discover words", exact: true }),
  ).toHaveAttribute("href", "/discovery?band=3&word=201");
});

test("plain Journey opens the first range that still has new words", async ({
  page,
}) => {
  await learner(page, { e2e_knowledge_map: "first-40-known" });
  await page.goto("/discover");
  await expect(
    page.getByRole("heading", { name: "Words 4,001–4,100", level: 2 }),
  ).toBeVisible();
  await expect(
    page
      .getByRole("list", { name: "Map progress" })
      .getByText("4,000 already know", { exact: true }),
  ).toBeVisible();
});

for (const requested of ["", "?band=3"] as const) {
  test(`a failed map read retries on Journey${requested ? " with its range" : ""}`, async ({
    page,
  }) => {
    await learner(page, { e2e_knowledge_map: "unavailable" });
    await page.goto(`/discover${requested}`);
    await expect(
      page.getByRole("alert").filter({ hasText: "Your map could not load." }),
    ).toBeVisible();
    await expect(
      page.getByRole("link", { name: "Try again", exact: true }),
    ).toHaveAttribute("href", `/discover${requested}`);
  });
}

test("an expired map read signs in again and returns to the same Journey range", async ({
  page,
}) => {
  await learner(page, { e2e_knowledge_map: "expired" });
  await page.goto("/discover?band=3");
  await expect
    .poll(() => new URL(page.url()).searchParams.get("returnTo"))
    .toBe("/discover?band=3");
});

test("every way to choose a range is at least 44px", async ({
  page,
}, testInfo) => {
  await learner(page);
  await page.goto("/discover?band=2");
  const controls = onPhone(testInfo)
    ? [
        page.getByRole("combobox", { name: "Go to a range" }),
        page.getByRole("button", { name: "Show", exact: true }),
      ]
    : await page
        .getByRole("navigation", { name: "Word ranges", exact: true })
        .getByRole("link")
        .all();
  controls.push(
    ...(await page
      .getByRole("navigation", { name: "Adjacent word ranges" })
      .getByRole("link")
      .all()),
  );
  for (const control of controls) {
    const box = (await control.boundingBox())!;
    expect(Math.min(box.width, box.height)).toBeGreaterThanOrEqual(44);
  }
});

test("the range chooser opens a range on a phone", async ({ page }, testInfo) => {
  test.skip(!onPhone(testInfo), "Phones choose ranges with the chooser.");
  await learner(page);
  await page.goto("/discover?band=2");
  await page
    .getByRole("combobox", { name: "Go to a range" })
    .selectOption({ label: "Words 2,201–2,300: 0 known, 0 learning" });
  await page.getByRole("button", { name: "Show", exact: true }).click();
  await expect(page).toHaveURL(/\/discover\?band=23#range-words$/);
  await expect(
    page.getByRole("heading", { name: "Words 2,201–2,300", level: 2 }),
  ).toBeVisible();
  // The range the chooser opened is the one scrolled into view.
  await expect(
    page.getByRole("region", { name: "Words 2,201–2,300" }),
  ).toBeInViewport();
});

test("the word map is one Tab stop with arrow-key movement", async ({
  page,
}, testInfo) => {
  test.skip(onPhone(testInfo), "The linked map is shown from 40rem.");
  await learner(page);
  await page.goto("/discover?band=12");
  const ranges = page.getByRole("navigation", { name: "Word ranges", exact: true });
  const current = ranges.getByRole("link", { name: /^Words 1101 to 1200,/ });
  await expect(current).toHaveAttribute("tabindex", "0");
  await expect(
    ranges.locator('a[tabindex="0"]'),
  ).toHaveCount(1);
  await current.focus();
  await page.keyboard.press("ArrowRight");
  await expect(
    ranges.getByRole("link", { name: /^Words 1201 to 1300,/ }),
  ).toBeFocused();
  await page.keyboard.press("ArrowDown");
  await expect(
    ranges.getByRole("link", { name: /^Words 2201 to 2300,/ }),
  ).toBeFocused();
  await page.keyboard.press("Enter");
  await expect(page).toHaveURL(/\/discover\?band=23#range-words$/);
  await expect(
    page.getByRole("heading", { name: "Words 2,201–2,300", level: 2 }),
  ).toBeVisible();
  // Map squares are not prefetched, so the range renders a moment before its
  // streamed metadata and the title is briefly absent. Audit the settled page.
  await expect(page).toHaveTitle("Vocanova");
  const { criticalOrSerious } = await scanForAxeViolations(page);
  expect(criticalOrSerious, formatViolations(criticalOrSerious).join("\n")).toEqual([]);
  expect(
    await page.evaluate(() => document.documentElement.scrollWidth),
  ).toBeLessThanOrEqual(page.viewportSize()!.width);
});
