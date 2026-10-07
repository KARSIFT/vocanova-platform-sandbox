import { randomUUID } from "node:crypto";

import { expect, test, type BrowserContext, type Page } from "@playwright/test";

const WORD_DETAIL_PATH = "/discover/ordering-at-a-cafe/pour#sentence-practice";

async function prepareSavedWord(
  page: Page,
  context: BrowserContext,
  baseURL: string,
): Promise<void> {
  await context.addCookies([
    {
      name: "vocanova_session",
      value: `sentence-feedback-${randomUUID()}`,
      url: baseURL,
    },
    {
      name: "vocanova_csrf",
      value: `csrf-${randomUUID()}`,
      url: baseURL,
    },
  ]);
  await page.goto(WORD_DETAIL_PATH);
  await page
    .getByRole("button", {
      name: /^Save pour: to make liquid flow into a container$/,
    })
    .click();
  await expect(
    page.getByRole("textbox", { name: /Write a sentence using pour/ }),
  ).toBeVisible();
}

test("revision retains prior feedback and submits a new request identity", async ({
  page,
  context,
}, testInfo) => {
  const requestKeys: string[] = [];
  page.on("request", (request) => {
    if (
      request.method() === "POST" &&
      request.url().includes("/api/v1/learner-sentences")
    ) {
      requestKeys.push(request.headers()["idempotency-key"] ?? "");
    }
  });
  await prepareSavedWord(page, context, testInfo.project.use.baseURL!);

  const sentenceInput = page.getByRole("textbox", {
    name: /Write a sentence using pour/,
  });
  await sentenceInput.fill("I pour teh coffee.");
  await page.getByRole("button", { name: "Check my sentence" }).click();
  await expect(
    page.getByText("Needs improvement", { exact: true }),
  ).toBeVisible();
  await expect(
    page.getByText("I pour the coffee.", { exact: true }),
  ).toBeVisible();

  await page.getByRole("button", { name: "Revise sentence" }).click();
  const previousFeedback = page.getByLabel("Previous feedback");
  await expect(previousFeedback).toBeVisible();
  await expect(
    previousFeedback.getByText("I pour teh coffee.", { exact: true }),
  ).toBeVisible();
  await expect(
    page.getByText("Suggested revision: I pour the coffee."),
  ).toBeVisible();

  await page.reload();
  await expect(sentenceInput).toHaveValue("I pour teh coffee.");
  await sentenceInput.fill("I pour the coffee.");
  await page.getByRole("button", { name: "Check my sentence" }).click();
  await expect(page.getByText("Correct", { exact: true })).toBeVisible();
  await expect(new Set(requestKeys).size).toBe(2);
  expect(requestKeys).toHaveLength(2);
});

test("revision preserves a newer unsent draft through reload", async ({
  page,
  context,
}, testInfo) => {
  const requests: Array<{ key: string; sentence: string }> = [];
  page.on("request", (request) => {
    if (
      request.method() === "POST" &&
      request.url().includes("/api/v1/learner-sentences")
    ) {
      requests.push({
        key: request.headers()["idempotency-key"] ?? "",
        sentence: request.postDataJSON().sentenceText,
      });
    }
  });
  await prepareSavedWord(page, context, testInfo.project.use.baseURL!);

  const checkedSentence = "I pour teh coffee.";
  const newerDraft = "I pour tea for my friends.";
  const sentenceInput = page.getByRole("textbox", {
    name: /Write a sentence using pour/,
  });
  await sentenceInput.fill(checkedSentence);
  await page.getByRole("button", { name: "Check my sentence" }).click();
  await expect(
    page.getByRole("status", { name: "Feedback result: Needs improvement" }),
  ).toBeVisible();
  await expect(sentenceInput).toBeEnabled();

  // No request is pending: the learner edits before using the earlier
  // feedback's revision action. The visible draft must remain recoverable.
  await sentenceInput.fill(newerDraft);
  await page.getByRole("button", { name: "Revise sentence" }).click();
  await expect(sentenceInput).toHaveValue(newerDraft);
  await expect(sentenceInput).toBeFocused();
  const previousFeedback = page.getByLabel("Previous feedback");
  await expect(
    previousFeedback.getByText(checkedSentence, { exact: true }),
  ).toBeVisible();
  await expect(previousFeedback).toContainText(
    "Suggested revision: I pour the coffee.",
  );
  await expect(previousFeedback).not.toContainText(newerDraft);
  expect(requests).toHaveLength(1);

  await page.reload();
  await expect(sentenceInput).toHaveValue(newerDraft);
  // Only the draft is recovered; previous feedback is not stored as a
  // fabricated result for the new sentence.
  await expect(previousFeedback).not.toBeVisible();
  await page.getByRole("button", { name: "Check my sentence" }).click();
  await expect(
    page.getByRole("status", { name: "Feedback result: Correct" }),
  ).toBeVisible();
  expect(requests).toHaveLength(2);
  expect(requests[0]!.key).toBeTruthy();
  expect(requests[1]!.key).toBeTruthy();
  expect(requests[1]!.key).not.toBe(requests[0]!.key);
  expect(requests[1]!.sentence).toBe(newerDraft);
});

test("a word-detail draft survives an auth gate and same-user return", async ({
  page,
  context,
}, testInfo) => {
  const baseURL = testInfo.project.use.baseURL!;
  await prepareSavedWord(page, context, baseURL);
  const sentenceInput = page.getByRole("textbox", {
    name: /Write a sentence using pour/,
  });
  await sentenceInput.fill("I pour coffee before work.");

  await context.addCookies([
    { name: "e2e_unauthenticated", value: "1", url: baseURL },
  ]);
  await page.reload();
  await expect(page).toHaveURL(/\/login\?returnTo=/);
  await expect(
    page.getByRole("heading", { name: "Sign in to Vocanova" }),
  ).toBeVisible();

  await context.clearCookies({ name: "e2e_unauthenticated" });
  await page.goto(WORD_DETAIL_PATH);
  await expect(sentenceInput).toHaveValue("I pour coffee before work.");
});

test("an alternate authenticated user cannot read a prior user's word draft", async ({
  page,
  context,
}, testInfo) => {
  const baseURL = testInfo.project.use.baseURL!;
  await prepareSavedWord(page, context, baseURL);
  const sentenceInput = page.getByRole("textbox", {
    name: /Write a sentence using pour/,
  });
  await sentenceInput.fill("I pour coffee before work.");

  await context.addCookies([
    { name: "e2e_user_id", value: "alternate", url: baseURL },
  ]);
  await page.reload();
  await expect(sentenceInput).toHaveValue("");
});
