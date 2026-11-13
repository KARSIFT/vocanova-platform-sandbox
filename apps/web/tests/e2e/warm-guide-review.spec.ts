import { randomUUID } from "node:crypto";
import { expect, test, type Locator, type Page } from "@playwright/test";
import type { DailyLesson } from "@vocanova/api-client";

// These regressions catch lost task guidance, an invented next-day promise,
// automatic audio, and presentation changes that alter review request identity.
async function seed(page: Page, count: number, theme: "light" | "dark") {
  const url = test.info().project.use.baseURL!;
  await page.context().addCookies([
    { name: "vocanova_session", value: `warm-review-${randomUUID()}`, url },
    { name: "vocanova_csrf", value: "warm-review-csrf", url },
    { name: "e2e_review_fixture_count", value: String(count), url },
    { name: "vocanova_theme", value: theme, url },
    // A deliberately non-tomorrow server date makes fabricated copy observable.
    { name: "e2e_daily_review_at", value: "2026-10-18T22:30:00.000Z", url },
  ]);
}

async function expectActionFits(page: Page, action: Locator) {
  await expect(action).toBeInViewport();
  expect((await action.boundingBox())!.height).toBeGreaterThanOrEqual(44);
  expect(
    await page.evaluate(() => document.documentElement.scrollWidth),
  ).toBeLessThanOrEqual(page.viewportSize()!.width);
}

async function expectClearReadingSurface(action: Locator) {
  const colors = await action.evaluate((element) => {
    const panel = element.closest("section")!;
    const surface = getComputedStyle(panel);
    const button = getComputedStyle(element);
    return {
      surface: surface.backgroundColor,
      foreground: surface.color,
      action: button.backgroundColor,
      label: button.color,
    };
  });
  const rgb = (value: string) =>
    value
      .match(/[\d.]+/g)!
      .slice(0, 3)
      .map(Number);
  const brightness = (value: string) =>
    rgb(value).reduce((sum, channel, index) => {
      const fraction = channel / 255;
      return (
        sum +
        (fraction <= 0.04045
          ? fraction / 12.92
          : ((fraction + 0.055) / 1.055) ** 2.4) *
          [0.2126, 0.7152, 0.0722][index]!
      );
    }, 0);
  const contrast = (a: string, b: string) =>
    (Math.max(brightness(a), brightness(b)) + 0.05) /
    (Math.min(brightness(a), brightness(b)) + 0.05);
  expect(contrast(colors.foreground, colors.surface)).toBeGreaterThanOrEqual(
    4.5,
  );
  expect(contrast(colors.label, colors.action)).toBeGreaterThanOrEqual(4.5);
  const background = rgb(colors.surface);
  expect(Math.max(...background) - Math.min(...background)).toBeLessThan(35);
  const [red, , blue] = rgb(colors.action);
  expect(blue! - red!).toBeGreaterThan(60);
}

for (const theme of ["light", "dark"] as const) {
  test(`review audio waits for intent and can be enabled (${theme})`, async ({
    page,
  }) => {
    await seed(page, 1, theme);
    let audioRequests = 0;
    await page.route("**/api/v1/pronunciation", async (route) => {
      audioRequests += 1;
      await route.fulfill({
        status: 503,
        contentType: "application/json",
        body: JSON.stringify({ title: "Unavailable" }),
      });
    });
    await page.goto("/review");
    await page
      .getByRole("button", { name: "Start review", exact: true })
      .click();
    await expect(
      page.getByRole("heading", { name: "Review word 1", exact: true }),
    ).toBeVisible();
    const toggle = page.getByRole("button", { name: /^Automatic audio / });
    await expect
      .soft(toggle)
      .toHaveAccessibleName("Automatic audio off", { timeout: 1500 });
    expect.soft(audioRequests).toBe(0);
    audioRequests = 0;
    await toggle.click();
    await expect(toggle).toHaveAccessibleName("Automatic audio on");
    await expect.poll(() => audioRequests).toBe(1);
    await expect(
      page.getByRole("button", {
        name: "Listen to Review word 1",
        exact: true,
      }),
    ).toBeEnabled();
  });

  test(`ready review explains the real queue and has one clear action (${theme})`, async ({
    page,
  }) => {
    await seed(page, 2, theme);
    await page.goto("/review");
    await expect(
      page.getByText(
        "2 words are ready. Choose the meaning, one word at a time.",
        { exact: true },
      ),
    ).toBeVisible();
    const start = page.getByRole("button", {
      name: "Start review",
      exact: true,
    });
    await expectActionFits(page, start);
    await expect(
      page.getByText(
        "Take your time. Choose an answer, then check it before moving on.",
        { exact: true },
      ),
    ).toBeVisible();
    await expectClearReadingSurface(start);
    const started = page.waitForResponse(
      (response) =>
        response.request().method() === "POST" &&
        new URL(response.url()).pathname === "/api/v1/daily-lessons",
    );
    await start.click();
    const lesson = (await (await started).json()) as DailyLesson;
    await expect(page).toHaveURL(
      new RegExp(`/review/session\\?lesson=${lesson.id}$`),
    );
    await expect(
      page.getByRole("heading", { name: "Review word 1", exact: true }),
    ).toBeVisible();
  });

  test(`empty review offers discovery without inventing a return date (${theme})`, async ({
    page,
  }) => {
    await seed(page, 0, theme);
    await page.goto("/review");
    await expect(
      page.getByRole("heading", {
        name: "You’re ready for new words",
        exact: true,
      }),
    ).toBeVisible();
    const discover = page
      .getByRole("main")
      .getByRole("link", { name: "Discover words", exact: true });
    await expectActionFits(page, discover);
    await expect(discover).toHaveAttribute("href", "/discovery");
    await expect(
      page.getByRole("button", { name: "Start review", exact: true }),
    ).toHaveCount(0);
    await expect(
      page.getByText("Come back to see which words are ready.", {
        exact: true,
      }),
    ).toBeVisible();
    await expect(page.getByRole("main")).not.toContainText(
      /come back tomorrow|mastered|earned points/i,
    );
  });

  test(`incorrect feedback keeps its identity and actual schedule through completion (${theme})`, async ({
    page,
  }) => {
    await seed(page, 1, theme);
    await page.goto("/review");
    const started = page.waitForResponse(
      (response) =>
        response.request().method() === "POST" &&
        new URL(response.url()).pathname === "/api/v1/daily-lessons",
    );
    await page
      .getByRole("button", { name: "Start review", exact: true })
      .click();
    const initial = (await (await started).json()) as DailyLesson;
    await expect
      .soft(
        page.getByRole("button", { name: "Automatic audio off", exact: true }),
      )
      .toBeVisible({ timeout: 1500 });
    await expect
      .soft(
        page.getByText(
          "Choose the meaning that fits this word. You can listen if it helps.",
          { exact: true },
        ),
      )
      .toBeVisible({ timeout: 1500 });
    const check = page.getByRole("button", {
      name: "Check answer",
      exact: true,
    });
    await expect(check).toBeDisabled();
    await page
      .getByRole("radio", { name: "A different meaning.", exact: true })
      .check();
    const answerResponse = page.waitForResponse(
      (response) =>
        response.request().method() === "POST" &&
        /\/api\/v1\/daily-lessons\/[^/]+\/actions$/.test(
          new URL(response.url()).pathname,
        ),
    );
    await check.click();
    const response = await answerResponse;
    const answerRequest = response.request();
    const body = answerRequest.postDataJSON();
    expect(body).toEqual({
      stepId: initial.currentStep!.id,
      expectedRevision: 1,
      clientActionId: expect.any(String),
      action: "answer",
      choiceId: "meaning-b",
    });
    expect(answerRequest.headers()["idempotency-key"]).toBe(
      body.clientActionId,
    );
    const answered = (await response.json()) as DailyLesson;
    expect(answered.feedback).toMatchObject({
      stepId: initial.currentStep!.id,
      correct: false,
      selectedChoiceId: "meaning-b",
      answer: "definition for review word 1",
      nextReviewAt: "2026-10-18T22:30:00.000Z",
    });
    await expect(
      page.getByRole("status").filter({ hasText: /^Not quite/ }),
    ).toContainText("definition for review word 1");
    await expect
      .soft(
        page.getByText(
          "Read the meaning, then continue. Your first answer stays recorded.",
          { exact: true },
        ),
      )
      .toBeVisible({ timeout: 1500 });
    const schedule = page.getByRole("button", {
      name: /^Show this word again:/,
    });
    const dateLabel = new Intl.DateTimeFormat("en", {
      month: "short",
      day: "numeric",
      timeZone: await page.evaluate(
        () => Intl.DateTimeFormat().resolvedOptions().timeZone,
      ),
    }).format(new Date("2026-10-18T22:30:00.000Z"));
    await expect(schedule).toHaveAccessibleName(
      `Show this word again: Automatic · ${dateLabel}`,
    );
    await schedule.click();
    await page
      .getByRole("menuitemradio", { name: "In 7 days", exact: true })
      .click();
    const continued = page.waitForResponse(
      (response) =>
        response.request().method() === "POST" &&
        /\/api\/v1\/daily-lessons\/[^/]+\/actions$/.test(
          new URL(response.url()).pathname,
        ),
    );
    const next = page.getByRole("button", { name: "Continue", exact: true });
    await expectActionFits(page, next);
    await next.click();
    const completeResponse = await continued;
    const continueRequest = completeResponse.request();
    const continuedBody = continueRequest.postDataJSON();
    expect(continuedBody).toEqual({
      stepId: initial.currentStep!.id,
      expectedRevision: 2,
      clientActionId: expect.any(String),
      action: "continue",
      nextReview: "7d",
    });
    expect(continueRequest.headers()["idempotency-key"]).toBe(
      continuedBody.clientActionId,
    );
    expect(continuedBody.clientActionId).not.toBe(body.clientActionId);
    expect(await completeResponse.json()).toMatchObject({
      id: initial.id,
      status: "completed",
      firstAnswersCorrect: 0,
      creditedToday: 0,
    });
    await expect(
      page.getByRole("heading", {
        name: "You’ve finished your review",
        exact: true,
      }),
    ).toBeFocused();
    await expect(
      page.getByText("0 correct on the first try.", { exact: true }),
    ).toBeVisible();
    await expect
      .soft(
        page.getByText("Your next review follows each word’s saved schedule.", {
          exact: true,
        }),
      )
      .toBeVisible({ timeout: 1500 });
    await expect
      .soft(
        page.getByText("Come back to see which words are ready.", {
          exact: true,
        }),
      )
      .toBeVisible({ timeout: 1500 });
    await expect(
      page
        .getByRole("main")
        .getByRole("link", { name: "Discover words", exact: true }),
    ).toHaveAttribute("href", "/discovery");
    await expect(
      page.getByRole("main").getByRole("link", { name: /write.*review word/i }),
    ).toHaveCount(0);
    await expect(page.getByRole("main")).not.toContainText(
      /come back tomorrow|mastered|earned points/i,
    );
  });
}
