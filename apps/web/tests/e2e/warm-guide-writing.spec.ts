import { randomUUID } from "node:crypto";
import { expect, test } from "@playwright/test";

// Catch hidden writing, wrong-sense submission, lost drafts, or sample success
// shown before a real feedback response reaches the editor.
test.beforeEach(async ({ context, baseURL }) => {
  await context.addCookies([
    { name: "vocanova_session", value: randomUUID(), url: baseURL! },
    { name: "vocanova_csrf", value: randomUUID(), url: baseURL! },
  ]);
});

for (const theme of ["light", "dark"] as const) {
  test(`saved writing is directly reachable while personal tools stay optional in ${theme}`, async ({
    page,
    context,
    baseURL,
  }, testInfo) => {
    await context.addCookies([
      { name: "vocanova_theme", value: theme, url: baseURL! },
    ]);
    // This case begins with a confirmed saved meaning. UI saving is covered
    // separately; its RSC refresh must not race this destination navigation.
    const api = `http://127.0.0.1:${process.env.MOCK_API_PORT ?? 8080}`;
    const csrf = (await context.cookies()).find(
      (cookie) => cookie.name === "vocanova_csrf",
    )!.value;
    const saved = await context.request.post(`${api}/api/v1/user-words`, {
      headers: { "X-CSRF-Token": csrf, "Idempotency-Key": randomUUID() },
      data: { meaningId: "mean-pour", source: "journey" },
    });
    expect(saved.status()).toBe(200);
    const savedWord = await saved.json();
    expect(savedWord).toMatchObject({
      userWordId: "uw-mean-pour",
      meaningId: "mean-pour",
      wordSlug: "pour",
      saved: true,
    });
    await page.goto(`/words/${savedWord.userWordId}`);
    const main = page.getByRole("main");
    const entry = main
      .locator("summary")
      .filter({ hasText: "Optional writing" });
    await expect(entry).toBeVisible();
    await expect(
      page.getByRole("button", { name: "Edit knowledge and note" }),
    ).toBeHidden();
    await entry.focus();
    await page.keyboard.press("Enter");
    const writing = page.getByRole("textbox", {
      name: "Write a sentence using pour",
      exact: true,
    });
    await expect(writing).toBeVisible();
    await expect(page).toHaveURL(/\/words\/uw-mean-pour#sentence-practice$/);
    await expect(
      page.getByRole("button", { name: "Edit knowledge and note" }),
    ).toBeHidden();
    await expect(
      main.getByText(/Avoid private details; AI can make mistakes/),
    ).toBeVisible();
    await writing.fill("I pour tea into my cup.");
    await page.reload();
    await expect(writing).toBeVisible();
    await expect(writing).toHaveValue("I pour tea into my cup.");
    expect(
      await page.evaluate(
        () => document.documentElement.scrollWidth <= innerWidth,
      ),
    ).toBe(true);
    await page.screenshot({
      path: testInfo.outputPath(`saved-writing-${theme}.png`),
      fullPage: true,
    });
  });
}

test("a secondary saved meaning opens and submits its exact writing draft", async ({
  page,
  context,
  baseURL,
}) => {
  await context.addCookies([
    { name: "e2e_vocabulary", value: "secondary-saved-meaning", url: baseURL! },
  ]);
  await page.goto("/vocabulary/pour");
  const secondMeaning = page
    .getByRole("main")
    .locator("#meaning-mean-pour-second");
  await secondMeaning.getByRole("button", { name: /^Save pour:/ }).click();
  await expect(
    secondMeaning.getByRole("button", { name: "Remove pour from saved words" }),
  ).toBeVisible();
  const entry = secondMeaning
    .locator("summary")
    .filter({ hasText: "Optional writing" });
  await expect(entry).toBeVisible();
  await entry.click();
  await expect(page).toHaveURL(/#sentence-practice-mean-pour-second$/);
  const writing = secondMeaning.getByRole("textbox", {
    name: "Write a sentence using pour",
    exact: true,
  });
  const draft = "The fans pour into the stadium.";
  await writing.fill(draft);
  await page.reload();
  await expect(writing).toBeVisible();
  await expect(writing).toHaveValue(draft);
  await page.goto("/words/uw-mean-pour-second#sentence-practice");
  await expect(
    page.getByRole("textbox", {
      name: "Write a sentence using pour",
      exact: true,
    }),
  ).toHaveValue(draft);
  await expect(
    page
      .getByRole("main")
      .getByText("to arrive in large numbers", { exact: true })
      .first(),
  ).toBeVisible();
  const request = page.waitForRequest(
    (candidate) =>
      candidate.method() === "POST" &&
      candidate.url().endsWith("/api/v1/learner-sentences"),
  );
  await page
    .getByRole("button", { name: "Check my sentence", exact: true })
    .click();
  expect((await request).postDataJSON()).toMatchObject({
    attemptId: "uw-mean-pour-second",
    sentenceText: draft,
    source: "word_detail",
  });
  await expect(
    page.getByText("Sentence checked", { exact: true }).locator(".."),
  ).toContainText(draft);
});

test("real feedback, revision and failure retain the learner's own sentence", async ({
  page,
}, testInfo) => {
  await page.goto("/vocabulary/pour#sentence-practice");
  await page.getByRole("button", { name: /^Save pour:/ }).click();
  const writing = page.getByRole("textbox", {
    name: "Write a sentence using pour",
    exact: true,
  });
  await expect(writing).toBeVisible();
  const original = "i pour tea!";
  const correction = "I pour tea.";
  const explanation = "Use a capital I and a full stop for this statement.";
  await writing.fill(original);
  let release!: () => void;
  const ready = new Promise<void>((resolve) => {
    release = resolve;
  });
  await page.route(
    "**/api/v1/learner-sentences",
    async (route) => {
      const response = await route.fetch();
      const data = await response.json();
      await ready;
      await route.fulfill({
        response,
        json: {
          ...data,
          processingStatus: "completed",
          status: "needs_improvement",
          originalSentence: original,
          correctedSentence: correction,
          explanation,
        },
      });
    },
    { times: 1 },
  );
  const checking = page.getByRole("button", {
    name: "Check my sentence",
    exact: true,
  });
  try {
    await checking.click();
    await expect(
      page.getByRole("button", { name: "Checking...", exact: true }),
    ).toBeDisabled();
    await expect(
      page.getByRole("region", { name: "Sentence comparison" }),
    ).toHaveCount(0);
    await expect(
      page.getByRole("status", { name: /^Feedback result:/ }),
    ).toHaveCount(0);
  } finally {
    release();
  }
  const comparison = page.getByRole("region", { name: "Sentence comparison" });
  await expect(comparison.getByText(original, { exact: true })).toBeVisible();
  await expect(comparison.getByText(correction, { exact: true })).toBeVisible();
  await expect(
    page.getByRole("status", { name: "Feedback result: Needs improvement" }),
  ).toContainText(explanation);
  await page.screenshot({
    path: testInfo.outputPath("real-sentence-feedback.png"),
    fullPage: true,
  });
  await comparison
    .getByRole("button", { name: "Revise sentence", exact: true })
    .click();
  await expect(writing).toBeFocused();
  await expect(writing).toHaveValue(original);
  await expect(
    page.getByRole("complementary", { name: "Previous feedback" }),
  ).toContainText(correction);
  const revision = "I pour tea before work.";
  await writing.fill(revision);
  await page.route(
    "**/api/v1/learner-sentences",
    (route) => route.abort("failed"),
    { times: 1 },
  );
  await checking.click();
  await expect(
    page
      .getByRole("alert")
      .filter({ hasText: /Unable to check this sentence/ }),
  ).toBeVisible();
  await expect(writing).toHaveValue(revision);
  await expect(
    page.getByRole("status", { name: /^Feedback result:/ }),
  ).toHaveCount(0);
  await expect(comparison).toHaveCount(0);
  await page.reload();
  await expect(writing).toBeVisible();
  await expect(writing).toHaveValue(revision);
});

test("a rejected real fixture response never becomes sample feedback", async ({
  page,
}) => {
  await page.goto("/vocabulary/pour#sentence-practice");
  await page.getByRole("button", { name: /^Save pour:/ }).click();
  const writing = page.getByRole("textbox", {
    name: "Write a sentence using pour",
    exact: true,
  });
  const sentence = "The weather is sunny today.";
  await writing.fill(sentence);
  await page
    .getByRole("button", { name: "Check my sentence", exact: true })
    .click();
  await expect(
    page.getByRole("alert").filter({ hasText: 'Add "pour" to your sentence' }),
  ).toBeVisible();
  await expect(writing).toHaveValue(sentence);
  const comparison = page.getByRole("region", { name: "Sentence comparison" });
  await expect(comparison.getByText(sentence, { exact: true })).toBeVisible();
  await expect(
    comparison.getByText("Your sentence", { exact: true }),
  ).toBeVisible();
  await expect(
    comparison.getByText("Sentence checked", { exact: true }),
  ).toHaveCount(0);
  await expect(
    comparison.getByText("Corrected sentence", { exact: true }),
  ).toHaveCount(0);
  await expect(
    page.getByRole("button", { name: "Revise sentence", exact: true }),
  ).toHaveCount(0);
});
