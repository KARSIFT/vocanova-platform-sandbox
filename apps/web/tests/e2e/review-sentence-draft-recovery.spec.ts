import { randomUUID } from "node:crypto";

import { expect, test } from "@playwright/test";

test("restores an unfinished review-completion sentence after a same-user reload", async ({
  page,
  context,
}, testInfo) => {
  const baseURL = testInfo.project.use.baseURL!;
  const session = `draft-recovery-${randomUUID()}`;
  await context.addCookies([
    { name: "vocanova_session", value: session, url: baseURL },
    { name: "vocanova_csrf", value: `csrf-${randomUUID()}`, url: baseURL },
    { name: "e2e_review_fixture_count", value: "1", url: baseURL },
  ]);

  await page.goto("/reviews");
  await page.getByRole("button", { name: "Show answer" }).click();
  await page.getByRole("button", { name: "Good" }).click();
  const sentenceInput = page.getByRole("textbox", {
    name: /Write a sentence using Review word 1/,
  });
  await expect(sentenceInput).toBeVisible();
  await sentenceInput.fill("I use Review word 1 at work.");

  await page.reload();
  await expect(
    page.getByRole("heading", { name: "Continue your sentence practice" }),
  ).toBeVisible();
  await expect(
    page.getByRole("textbox", {
      name: /Write a sentence using Review word 1/,
    }),
  ).toHaveValue("I use Review word 1 at work.");
});
