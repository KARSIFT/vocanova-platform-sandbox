import { randomUUID } from "node:crypto";
import { expect, test } from "@playwright/test";

import { formatViolations, scanForAxeViolations } from "./axe-helper";

const mockAPI = `http://127.0.0.1:${process.env.MOCK_API_PORT ?? 8080}`;

for (const theme of ["light", "dark"] as const) {
  test(`Practice connects real review, lesson, writing and listening activities in ${theme} mode`, async ({
    page,
    context,
    baseURL,
  }, testInfo) => {
    if (!baseURL) throw new Error("A test app URL is required");
    const csrf = randomUUID();
    await context.addCookies([
      { name: "vocanova_session", value: randomUUID(), url: baseURL },
      { name: "vocanova_csrf", value: csrf, url: baseURL },
      { name: "vocanova_theme", value: theme, url: baseURL },
      { name: "e2e_review_fixture_count", value: "7", url: baseURL },
    ]);
    const saved = await context.request.post(`${mockAPI}/api/v1/user-words`, {
      headers: { "X-CSRF-Token": csrf },
      data: { meaningId: "mean-pour", source: "journey" },
    });
    expect(saved.ok()).toBe(true);

    await page.goto("/practice");
    const main = page.getByRole("main");
    await expect(page).toHaveTitle("Practice — Vocanova");
    await expect(page.locator("html")).toHaveAttribute("data-theme", theme);
    await expect(
      main.getByRole("heading", { level: 1, name: "Practice" }),
    ).toBeVisible();
    const activities = main.getByRole("navigation", { name: "Practice activities" });
    await expect(activities.getByRole("link", { name: "Topic writing", exact: true })).toHaveAttribute("href", "/writing");
    await expect(activities.getByRole("link", { name: "Short stories", exact: true })).toHaveAttribute("href", "/stories");
    await expect(main.getByRole("heading", { name: "Write a sentence", exact: true })).toHaveAttribute("id", "practice-writing-heading");
    // The endpoint returns one item but seven due words. The count must come
    // from totalCount, never from the sampled queue length.
    await expect(
      main.getByText("7 words are due for review.", { exact: true }),
    ).toBeVisible();
    await expect(
      main.getByRole("link", { name: "Start review", exact: true }),
    ).toHaveAttribute("href", "/review");
    await expect(
      main.getByRole("link", { name: "Start lesson", exact: true }),
    ).toHaveAttribute("href", "/learn/conversation-basics");
    await expect(
      main.getByRole("link", { name: "Write with pour", exact: true }),
    ).toHaveAttribute("href", "/words/uw-mean-pour#sentence-practice");
    await expect(
      main.getByRole("link", { name: "Listen to pour", exact: true }),
    ).toHaveAttribute("href", "/vocabulary/pour");
    // The listening choice explains the activity where the learner starts it.
    await expect(main.getByRole("button", { name: "Start listening practice", exact: true })).toContainText("No microphone needed.");
    await expect(
      main.getByRole("link", { name: "Open sentence history", exact: true }),
    ).toHaveAttribute("href", "/progress/sentences");
    await expect(
      page.getByRole("navigation", { name: "Primary" }).getByRole("link"),
    ).toHaveCount(3);
    await expect(
      page
        .getByRole("navigation", { name: "Primary" })
        .getByRole("link", { name: "Journey", exact: true }),
    ).toHaveAttribute("aria-current", "page");

    for (const link of await main.getByRole("link").all()) {
      const box = await link.boundingBox();
      expect(box).not.toBeNull();
      expect(box!.height).toBeGreaterThanOrEqual(44);
    }
    expect(
      await page.evaluate(() => document.documentElement.scrollWidth),
    ).toBeLessThanOrEqual(page.viewportSize()!.width);
    const accessibility = await scanForAxeViolations(page);
    expect(
      accessibility.criticalOrSerious,
      formatViolations(accessibility.criticalOrSerious).join("\n"),
    ).toEqual([]);
    await page.screenshot({
      path: testInfo.outputPath(`practice-${theme}.png`),
      fullPage: true,
    });

    // Start through the API to establish saved progress, then exercise the
    // hub's resume link against the same server-owned session.
    const started = await context.request.post(
      `${mockAPI}/api/v1/lessons/conversation-basics/sessions`,
      {
        headers: { "X-CSRF-Token": csrf, "Idempotency-Key": randomUUID() },
      },
    );
    expect(started.ok()).toBe(true);
    const session = await started.json();
    await page.reload();
    await expect(
      main.getByText(
        `${session.completedSteps} of ${session.totalSteps} steps saved`,
        { exact: true },
      ),
    ).toBeVisible();
    const resume = main.getByRole("link", {
      name: "Continue lesson",
      exact: true,
    });
    await resume.focus();
    await page.keyboard.press("Enter");
    await expect(page).toHaveURL(/\/learn\/conversation-basics$/);
    await expect(main.getByRole("heading", { level: 1 })).toHaveText(
      session.currentStep.word.wordText,
    );
    await expect(
      main.getByRole("progressbar", { name: "Lesson progress" }),
    ).toHaveAttribute("value", String(session.completedSteps));
  });

  test(`Practice offers useful first steps without saved words or lessons in ${theme} mode`, async ({
    page,
    context,
    baseURL,
  }, testInfo) => {
    if (!baseURL) throw new Error("A test app URL is required");
    await context.addCookies([
      { name: "vocanova_session", value: randomUUID(), url: baseURL },
      { name: "vocanova_theme", value: theme, url: baseURL },
      { name: "e2e_lessons", value: "empty", url: baseURL },
    ]);
    await page.goto("/practice");
    const main = page.getByRole("main");
    await expect(
      main.getByText("No words are due right now.", { exact: true }),
    ).toBeVisible();
    await expect(
      main.getByRole("link", { name: "Start review", exact: true }),
    ).toHaveCount(0);
    await expect(
      main.getByRole("link", { name: "View reviews", exact: true }),
    ).toHaveAttribute("href", "/review");
    await expect(
      main.getByText(
        "No guided lessons are available yet.",
        { exact: true },
      ),
    ).toBeVisible();
    await expect(
      main.getByRole("link", { name: "Explore situations", exact: true }),
    ).toHaveAttribute("href", "/discover");
    await expect(
      main.getByText(
        "Save a word first, then use it in your own sentence.",
        { exact: true },
      ),
    ).toBeVisible();
    await expect(
      main.getByRole("link", { name: "Find a word to save", exact: true }),
    ).toHaveAttribute("href", "/vocabulary");
    await expect(
      main.getByRole("link", { name: "Find a word to hear", exact: true }),
    ).toHaveAttribute("href", "/vocabulary");
    await expect(
      main.getByRole("link", { name: "Open sentence history", exact: true }),
    ).toBeVisible();
    await expect(
      main.locator("[aria-disabled='true'], button:disabled"),
    ).toHaveCount(0);
    expect(
      await page.evaluate(() => document.documentElement.scrollWidth),
    ).toBeLessThanOrEqual(page.viewportSize()!.width);
    await page.screenshot({
      path: testInfo.outputPath(`practice-empty-${theme}.png`),
      fullPage: true,
    });
  });
}

test("Practice keeps scheduled count truthful after today's review target is complete", async ({
  page,
  context,
  baseURL,
}) => {
  if (!baseURL) throw new Error("A test app URL is required");
  const csrf = randomUUID();
  await context.addCookies([
    { name: "vocanova_session", value: randomUUID(), url: baseURL },
    { name: "vocanova_csrf", value: csrf, url: baseURL },
    { name: "e2e_review_fixture_count", value: "7", url: baseURL },
    { name: "e2e_daily_review_target", value: "5", url: baseURL },
  ]);
  for (let index = 1; index <= 5; index += 1) {
    const response = await context.request.post(
      `${mockAPI}/api/v1/reviews/submissions`,
      {
        headers: { "X-CSRF-Token": csrf, "Idempotency-Key": randomUUID() },
        data: {
          userWordId: `fixture-user-word-${index}`,
          meaningId: `fixture-meaning-${index}`,
          promptType: "self_check",
          result: "correct",
          rating: "good",
          answeredAt: new Date().toISOString(),
          clientAttemptId: randomUUID(),
        },
      },
    );
    expect(response.ok()).toBe(true);
  }
  await page.goto("/practice");
  const main = page.getByRole("main");
  await expect(
    main.getByText("2 words are due for review.", { exact: true }),
  ).toBeVisible();
  await expect(
    main.getByText(
      "Today’s review target is complete. Your next review session is tomorrow.",
      { exact: true },
    ),
  ).toBeVisible();
  await expect(
    main.getByRole("link", { name: "Start review", exact: true }),
  ).toHaveCount(0);
  await main.getByRole("link", { name: "View reviews", exact: true }).click();
  await expect(
    main.getByRole("heading", { name: "Today's review target is complete" }),
  ).toBeVisible();
});

test("Practice distinguishes unavailable lessons from an empty catalog without hiding other activities", async ({
  page,
  context,
  baseURL,
}) => {
  if (!baseURL) throw new Error("A test app URL is required");
  await context.addCookies([
    { name: "vocanova_session", value: randomUUID(), url: baseURL },
    { name: "e2e_lessons", value: "unavailable", url: baseURL },
    { name: "e2e_review_fixture_count", value: "3", url: baseURL },
  ]);
  await page.goto("/practice");
  const main = page.getByRole("main");
  await expect(main.getByRole("status")).toHaveText(
    "We could not load your lessons. You can still practise with your vocabulary.",
  );
  await expect(
    main.getByText("3 words are due for review.", { exact: true }),
  ).toBeVisible();
  await expect(
    main.getByRole("link", { name: "Start review", exact: true }),
  ).toBeVisible();
  await expect(
    main.getByRole("link", { name: "Find a word to hear", exact: true }),
  ).toBeVisible();
  await expect(
    main.getByRole("link", { name: "Open sentence history", exact: true }),
  ).toBeVisible();
});

for (const lessons of ["empty", "unavailable"]) {
  test(`Practice can reload unavailable lists with ${lessons} lessons and loaded sessions`, async ({ page, context, baseURL }) => {
    if (!baseURL) throw new Error("A test app URL is required");
    const csrf = randomUUID();
    await context.addCookies([
      { name: "vocanova_session", value: randomUUID(), url: baseURL },
      { name: "vocanova_csrf", value: csrf, url: baseURL },
      { name: "e2e_lessons", value: lessons, url: baseURL },
      { name: "e2e_lists", value: "unavailable", url: baseURL },
    ]);
    await page.goto("/practice");
    const main = page.getByRole("main");
    await expect(main.getByLabel("Practice vocabulary", { exact: true })).toHaveCount(0);
    await expect(main.getByText("Personal lists are unavailable. Refresh to try again.", { exact: true })).toBeVisible();
    // A successful empty session read must not hide the list retry action.
    await expect(main.getByRole("button", { name: "Load saved sessions", exact: true })).toHaveCount(0);
    const refresh = main.getByRole("button", { name: "Refresh practice sources", exact: true });
    await expect(refresh).toBeVisible();
    await context.clearCookies({ name: "e2e_lists" });
    const id = randomUUID();
    const created = await context.request.put(`${mockAPI}/api/v1/word-lists/${id}`, {
      headers: { "X-CSRF-Token": csrf, "Idempotency-Key": randomUUID() },
      data: { name: "Recovered list", expectedRevision: 0 },
    });
    expect(created.ok()).toBe(true);
    const reloaded = page.waitForResponse(response => new URL(response.url()).pathname === "/api/v1/word-lists" && response.request().method() === "GET");
    await refresh.focus();
    await page.keyboard.press("Enter");
    expect((await reloaded).status()).toBe(200);
    await expect(main.getByText("Personal lists are unavailable. Refresh to try again.", { exact: true })).toHaveCount(0);
    const selection = main.getByLabel("Practice vocabulary", { exact: true });
    await expect(selection).toBeVisible();
    await expect(selection.locator(`option[value="list:${id}"]`)).toHaveText("Recovered list (0 practice meanings)");
    await expect(main.getByRole("button", { name: "Start typed recall", exact: true })).toBeEnabled();
  });
}

test("Practice asks signed-out learners to sign in and preserves their destination", async ({
  page,
  context,
  baseURL,
}) => {
  if (!baseURL) throw new Error("A test app URL is required");
  await context.addCookies([
    { name: "e2e_unauthenticated", value: "1", url: baseURL },
  ]);
  await page.goto("/practice");
  await expect(page).toHaveURL(/\/login\?returnTo=%2Fpractice$/);
  await expect(
    page.getByRole("main").getByRole("heading", { name: "Practice" }),
  ).toHaveCount(0);
});
