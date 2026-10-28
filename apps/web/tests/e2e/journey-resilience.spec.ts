// Journey's situation cards are optional context. These checks prove that the
// knowledge map and the other ways to learn never depend on, or wait for, the
// situations read, and that the alternatives survive a failed map read.

import { randomUUID } from "node:crypto";
import { expect, test, type Page } from "@playwright/test";

import { formatViolations, scanForAxeViolations } from "./axe-helper";

async function openJourney(
  page: Page,
  cookies: Record<string, string>,
  // A streamed page fires `load` only after every boundary resolves; "commit"
  // lets a test observe the map while optional content is still pending.
  waitUntil: "load" | "commit" = "load",
) {
  const baseURL = test.info().project.use.baseURL!;
  await page.context().addCookies([
    { name: "vocanova_session", value: randomUUID(), url: baseURL },
    ...Object.entries(cookies).map(([name, value]) => ({
      name,
      value,
      url: baseURL,
    })),
  ]);
  await page.goto("/discover", { waitUntil });
}

async function expectOtherWaysToLearn(page: Page) {
  const ways = page.getByRole("navigation", { name: "More ways to learn" });
  for (const [name, href] of [
    ["Situation courses", "/courses"],
    ["Short stories", "/stories"],
    ["Personal lists", "/lists"],
  ] as const) {
    const link = ways.getByRole("link", { name, exact: true });
    await expect(link).toBeVisible();
    await expect(link).toHaveAttribute("href", href);
    expect((await link.boundingBox())!.height).toBeGreaterThanOrEqual(44);
  }
}

for (const theme of ["light", "dark"] as const) {
  test(`Journey keeps its other entry points when the map fails in ${theme}`, async ({
    page,
  }) => {
    await openJourney(page, {
      e2e_knowledge_map: "unavailable",
      vocanova_theme: theme,
    });
    await expect(
      page.getByRole("heading", { name: "Journey", level: 1 }),
    ).toBeVisible();
    await expect(
      page.getByRole("alert").filter({ hasText: "Your map could not load." }),
    ).toBeVisible();
    await expectOtherWaysToLearn(page);
    await expect(
      page.getByRole("link", { name: "Ordering at a cafe", exact: true }),
    ).toHaveAttribute("href", "/discover/ordering-at-a-cafe");
    expect(
      await page.evaluate(() => document.documentElement.scrollWidth),
    ).toBeLessThanOrEqual(page.viewportSize()!.width);
    const { criticalOrSerious } = await scanForAxeViolations(page);
    expect(
      criticalOrSerious,
      formatViolations(criticalOrSerious).join("\n"),
    ).toEqual([]);
  });
}

test("Journey explains unavailable situations and keeps the map", async ({
  page,
}) => {
  await openJourney(page, { e2e_journey_situations: "unavailable" });
  await expect(
    page.getByRole("heading", { name: "Words 1–100", level: 2 }),
  ).toBeVisible();
  await expect(
    page.getByRole("status").filter({ hasText: "Situations could not load" }),
  ).toBeVisible();
  await expect(
    page.getByRole("link", { name: "Ordering at a cafe", exact: true }),
  ).toHaveCount(0);
  await expectOtherWaysToLearn(page);
});

test("Journey shows the map before slow situations arrive", async ({
  page,
}) => {
  await openJourney(page, { e2e_journey_situations: "slow" }, "commit");
  // The fixture holds the situations list for several seconds; the map and
  // alternatives must already be usable while its own fallback is pending.
  await expect(
    page.getByRole("heading", { name: "Words 1–100", level: 2 }),
  ).toBeVisible();
  await expect(
    page.getByRole("status").filter({ hasText: "Loading situations…" }),
  ).toBeAttached();
  await expect(
    page.getByRole("link", { name: "Ordering at a cafe", exact: true }),
  ).toHaveCount(0);
  await expectOtherWaysToLearn(page);
  await expect(
    page.getByRole("link", { name: "Ordering at a cafe", exact: true }),
  ).toBeVisible({ timeout: 10_000 });
  await expect(
    page.getByRole("status").filter({ hasText: "Loading situations…" }),
  ).toHaveCount(0);
});
