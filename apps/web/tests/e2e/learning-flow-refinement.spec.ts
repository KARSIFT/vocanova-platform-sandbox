import { randomUUID } from "node:crypto";

import { expect, test } from "@playwright/test";

test("shows sparse mission history in date order without inventing missing days", async ({
  page,
  context,
  baseURL,
}) => {
  if (!baseURL) throw new Error("A test app URL is required");
  await context.addCookies([
    { name: "vocanova_session", value: randomUUID(), url: baseURL },
    { name: "e2e_progress_history_fixture", value: "sparse", url: baseURL },
  ]);
  await page.goto("/progress");
  await expect(page.getByRole("heading", { name: "Recent activity" })).toBeVisible();
  const dates = page.locator("main time[datetime]");
  await expect(dates).toHaveCount(3);
  expect(await dates.evaluateAll((items) => items.map((item) => item.getAttribute("datetime"))))
    .toEqual(["2026-09-09", "2026-09-10", "2026-09-12"]);
  for (const date of await dates.all()) {
    await expect(date).not.toHaveText("");
  }
});

test("keeps optional sentence practice together with confirmed review completion", async ({
  page,
  context,
  baseURL,
}) => {
  if (!baseURL) throw new Error("A test app URL is required");
  await context.addCookies([
    { name: "vocanova_session", value: randomUUID(), url: baseURL },
    { name: "vocanova_csrf", value: "flow-refinement-csrf", url: baseURL },
    { name: "e2e_review_fixture_count", value: "1", url: baseURL },
  ]);
  await page.goto("/review");
  await page.getByRole("button", { name: "Show answer" }).click();
  await page.getByRole("button", { name: "Good" }).click();
  await expect(page.getByRole("heading", { name: "Review complete" })).toBeVisible();
  await expect(page.getByText("You reviewed 1 word.")).toBeVisible();
  const input = page.getByRole("textbox", { name: /Write a sentence using/ });
  await expect(input).toBeVisible();
  const completion = page.getByRole("heading", { name: "Review complete" });
  expect(await completion.evaluate((heading) => Boolean(
    heading.compareDocumentPosition(document.querySelector("textarea")!) & Node.DOCUMENT_POSITION_FOLLOWING,
  ))).toBe(true);
  await input.fill("I use review word 1 every day.");
  await expect(page.getByRole("button", { name: "Check my sentence" })).toBeEnabled();
  expect(await page.evaluate(() => document.documentElement.scrollWidth)).toBeLessThanOrEqual(page.viewportSize()!.width);
});
