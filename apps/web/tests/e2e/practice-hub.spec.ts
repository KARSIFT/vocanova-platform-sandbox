import { randomUUID } from "node:crypto";
import { expect, test } from "@playwright/test";

import { formatViolations, scanForAxeViolations } from "./axe-helper";

const mockAPI = `http://127.0.0.1:${process.env.MOCK_API_PORT ?? 8080}`;

for (const theme of ["light", "dark"] as const) {
  test(`Practice connects review, discovery, lesson and listening activities in ${theme} mode`, async ({
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
      main.getByRole("heading", { level: 1, name: "Today’s practice" }),
    ).toBeVisible();
    const activities = main.getByRole("navigation", {
      name: "Practice activities",
    });
    await expect(
      activities.getByRole("link", { name: "Discover words", exact: true }),
    ).toHaveAttribute("href", "/discovery");
    await expect(
      activities.getByRole("link", { name: "Short stories", exact: true }),
    ).toHaveAttribute("href", "/stories");
    await expect(main.getByRole("textbox")).toHaveCount(0);
    // The endpoint returns one item but seven due words. The count must come
    // from totalCount, never from the sampled queue length.
    await expect(
      main.getByText("7 words are due for review.", { exact: true }),
    ).toBeVisible();
    await expect(
      main.getByRole("link", { name: "Start review", exact: true }),
    ).toHaveAttribute("href", "/review");
    await expect(
      main.getByRole("button", { name: "Start practice lesson", exact: true }),
    ).toBeEnabled();
    const currentAction = main.getByRole("button", {
      name: "Start practice lesson",
      exact: true,
    });
    await expect(currentAction).toBeInViewport({ ratio: 1 });
    const actionBounds = await currentAction.boundingBox();
    const navBounds = await page
      .getByRole("navigation", { name: "Primary", exact: true })
      .boundingBox();
    if (navBounds && navBounds.y > page.viewportSize()!.height / 2)
      expect(actionBounds!.y + actionBounds!.height).toBeLessThanOrEqual(
        navBounds.y,
      );

    await expect(
      main.getByRole("link", { name: "Listen to pour", exact: true }),
    ).toHaveAttribute("href", "/vocabulary/pour");
    await main.getByText("More practice", { exact: true }).click();
    // The listening choice explains the activity where the learner starts it.
    await expect(
      main.getByRole("button", {
        name: "Start listening practice",
        exact: true,
      }),
    ).toContainText("No microphone needed.");
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

    // Resume the exact server-confirmed personal practice session.
    const started = await context.request.post(
      `${mockAPI}/api/v1/practice-sessions`,
      {
        headers: { "X-CSRF-Token": csrf, "Idempotency-Key": randomUUID() },
        data: { mode: "journey" },
      },
    );
    expect(started.ok()).toBe(true);
    const session = await started.json();
    await page.reload();
    const resume = main.getByRole("link", {
      name: "Continue practice",
      exact: true,
    });
    await expect(resume).toHaveAttribute(
      "href",
      `/practice/session/${session.id}`,
    );
    await resume.focus();
    await page.keyboard.press("Enter");
    await expect(page).toHaveURL(
      new RegExp(`/practice/session/${session.id}$`),
    );
    await expect(
      main.getByRole("progressbar", { name: "Practice progress" }),
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
      main.getByRole("link", { name: "Open reviews", exact: true }),
    ).toHaveAttribute("href", "/review");
    await expect(
      main.getByRole("button", { name: "Start practice lesson", exact: true }),
    ).toBeEnabled();
    await expect(
      main.getByRole("link", {
        name: "Explore situation lessons",
        exact: true,
      }),
    ).toHaveAttribute("href", "/courses");
    await expect(
      main.getByRole("link", { name: "Discover words", exact: true }),
    ).toHaveAttribute("href", "/discovery");
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
  let response = await context.request.post(`${mockAPI}/api/v1/daily-lessons`, {
    headers: { "X-CSRF-Token": csrf, "Idempotency-Key": randomUUID() },
    data: {},
  });
  expect(response.ok()).toBe(true);
  let lesson = await response.json();
  for (let index = 0; index < 5; index++) {
    for (const action of ["answer", "continue"] as const) {
      const id = randomUUID();
      response = await context.request.post(
        `${mockAPI}/api/v1/daily-lessons/${lesson.id}/actions`,
        {
          headers: { "X-CSRF-Token": csrf, "Idempotency-Key": id },
          data: {
            action,
            clientActionId: id,
            stepId: lesson.currentStep.id,
            expectedRevision: lesson.revision,
            ...(action === "answer" ? { choiceId: "meaning-a" } : {}),
          },
        },
      );
      expect(response.ok()).toBe(true);
      lesson = await response.json();
    }
  }

  await page.goto("/practice");
  const main = page.getByRole("main");
  await expect(
    main.getByText("2 words are due for review.", { exact: true }),
  ).toBeVisible();
  // A daily habit target does not conceal still-due words or an active lesson.
  await main.getByRole("link", { name: "Start review", exact: true }).click();
  await expect(
    main.getByRole("heading", { name: "Review word 6", exact: true }),
  ).toBeVisible();
  await expect(
    main.getByRole("progressbar", { name: "Review progress" }),
  ).toHaveAttribute("value", "5");
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
  await main.getByText("More practice", { exact: true }).click();
  await expect(main.getByRole("status")).toHaveText(
    "Situation lesson choices are unavailable. Your personal practice path is still available.",
  );
  await expect(
    main.getByRole("button", { name: "Start practice lesson", exact: true }),
  ).toBeEnabled();
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
  test(`Practice can reload unavailable lists with ${lessons} lessons and loaded sessions`, async ({
    page,
    context,
    baseURL,
  }) => {
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
    await main.getByText("More practice", { exact: true }).click();
    await expect(
      main.getByLabel("Practice vocabulary", { exact: true }),
    ).toHaveCount(0);
    await expect(
      main.getByText("Personal lists are unavailable. Refresh to try again.", {
        exact: true,
      }),
    ).toBeVisible();
    // A successful empty session read must not hide the list retry action.
    await expect(
      main.getByRole("button", { name: "Load saved sessions", exact: true }),
    ).toHaveCount(0);
    const refresh = main.getByRole("button", {
      name: "Refresh practice sources",
      exact: true,
    });
    await expect(refresh).toBeVisible();
    await context.clearCookies({ name: "e2e_lists" });
    const id = randomUUID();
    const created = await context.request.put(
      `${mockAPI}/api/v1/word-lists/${id}`,
      {
        headers: { "X-CSRF-Token": csrf, "Idempotency-Key": randomUUID() },
        data: { name: "Recovered list", expectedRevision: 0 },
      },
    );
    expect(created.ok()).toBe(true);
    const reloaded = page.waitForResponse(
      (response) =>
        new URL(response.url()).pathname === "/api/v1/word-lists" &&
        response.request().method() === "GET",
    );
    await refresh.focus();
    await page.keyboard.press("Enter");
    expect((await reloaded).status()).toBe(200);
    await expect(
      main.getByText("Personal lists are unavailable. Refresh to try again.", {
        exact: true,
      }),
    ).toHaveCount(0);
    const selection = main.getByLabel("Practice vocabulary", { exact: true });
    await expect(selection).toBeVisible();
    await expect(selection.locator(`option[value="list:${id}"]`)).toHaveText(
      "Recovered list (0 practice meanings)",
    );
    await expect(
      main.getByRole("button", { name: "Start typed recall", exact: true }),
    ).toBeEnabled();
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
