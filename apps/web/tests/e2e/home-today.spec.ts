// Home leads with one next step, then the day's real goal. Practice and other
// activities stay available without competing with that step.

import { randomUUID } from "node:crypto";
import { expect, test } from "@playwright/test";

import { formatViolations, scanForAxeViolations } from "./axe-helper";

for (const due of [0, 3] as const) {
  for (const theme of ["light", "dark"] as const) {
    test(`Home has one filled action and the mission's own numbers (${due} due, ${theme})`, async ({
      page,
      baseURL,
    }) => {
      await page.context().addCookies([
        { name: "vocanova_session", value: `today-${randomUUID()}`, url: baseURL! },
        { name: "e2e_review_fixture_count", value: String(due), url: baseURL! },
        { name: "vocanova_theme", value: theme, url: baseURL! },
      ]);
      await page.goto("/home");
      const main = page.getByRole("main");
      await expect(main.locator(".vn-btn-primary")).toHaveCount(1);
      const next = main.getByRole("region", { name: "Review words", exact: true });
      await expect(next.locator(".vn-btn-primary")).toHaveCount(1);
      await expect(
        due
          ? next.getByRole("button", { name: "Start review", exact: true })
          : next.getByRole("link", { name: "Discover words", exact: true }),
      ).toBeInViewport();

      const goal = main.getByRole("region", { name: "Today’s goal", exact: true });
      await expect(goal.getByText("In progress", { exact: true })).toBeVisible();
      for (const [label, value] of [
        ["Reviews", "0 of 20"],
        ["New words", "0 of 5"],
        ["Sentence practice(optional)", "0 of 3"],
      ] as const) {
        const row = goal.locator(".goal-row").filter({ hasText: label });
        await expect(row.locator("dd")).toHaveText(value);
      }
      // The secondary practice path keeps its own action, but not a filled one.
      const path = main.getByRole("region", { name: "Your learning path", exact: true });
      await expect(
        path.getByRole("button", { name: "Start practice lesson", exact: true }),
      ).toBeVisible();
      await expect(path.locator(".vn-btn-primary")).toHaveCount(0);
      // When discovery is already the next step it is not offered twice.
      await expect(
        main.getByRole("link", { name: "Discover", exact: true }),
      ).toHaveCount(due ? 1 : 0);
      const { criticalOrSerious } = await scanForAxeViolations(page);
      expect(criticalOrSerious, formatViolations(criticalOrSerious).join("\n")).toEqual([]);
      expect(
        await page.evaluate(() => document.documentElement.scrollWidth),
      ).toBeLessThanOrEqual(page.viewportSize()!.width);
    });
  }
}

test("an empty review queue points to saving words, and a newly saved word is ready to review", async ({
  page,
  baseURL,
}) => {
  await page.context().addCookies([
    { name: "vocanova_session", value: `today-${randomUUID()}`, url: baseURL! },
    { name: "vocanova_csrf", value: "home-today", url: baseURL! },
  ]);
  await page.goto("/home");
  const goal = page.getByRole("region", { name: "Today’s goal", exact: true });
  const note = goal.getByText(
    "No words are due right now. Save new words in Discover, then check back here for review.",
    { exact: true },
  );
  await expect(note).toBeVisible();
  const next = page.getByRole("region", { name: "Review words", exact: true });
  await expect(
    next.getByRole("link", { name: "Discover words", exact: true }),
  ).toBeVisible();

  // Newly saved words are due straight away (backend contract).
  await page.goto("/vocabulary/invite");
  await page.getByRole("button", { name: /^Save invite:/ }).click();
  await expect(
    page.getByRole("button", { name: "Remove invite from saved words", exact: true }),
  ).toBeVisible();
  await page.goto("/home");
  await expect(next.getByRole("button", { name: "Start review", exact: true })).toBeVisible();
  await expect(next.getByText("1 word is ready to remember.", { exact: true })).toBeVisible();
  await expect(note).toHaveCount(0);
});

test("a learner without saved words gets an invitation, not an empty card", async ({
  page,
  baseURL,
}) => {
  await page.context().addCookies([
    { name: "vocanova_session", value: `today-${randomUUID()}`, url: baseURL! },
  ]);
  await page.goto("/home");
  const latest = page.getByRole("region", { name: "Your latest word", exact: true });
  await expect(latest.getByText("Nothing saved yet", { exact: true })).toBeVisible();
  await expect(
    latest.getByRole("link", { name: "Search for a word", exact: true }),
  ).toHaveAttribute("href", "/vocabulary");
});
