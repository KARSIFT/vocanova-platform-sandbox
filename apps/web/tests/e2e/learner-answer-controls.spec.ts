import { randomUUID } from "node:crypto";
import { expect, test, type Page } from "@playwright/test";
import type { LessonSession, PracticeSession } from "@vocanova/api-client";
import { scanForAxeViolations } from "./axe-helper";

test.beforeEach(async ({ context, baseURL }) => {
  if (!baseURL) throw new Error("Missing app URL");
  await context.addCookies([
    { name: "vocanova_session", value: randomUUID(), url: baseURL },
    { name: "vocanova_csrf", value: randomUUID(), url: baseURL },
  ]);
});

async function openLessonQuestion(page: Page) {
  await page.goto("/learn/conversation-basics");
  const opened = page.waitForResponse(
    (response) =>
      response.request().method() === "POST" &&
      response.url().endsWith("/lessons/conversation-basics/sessions"),
  );
  await page.getByRole("button", { name: "Start lesson", exact: true }).click();
  let session = (await (await opened).json()) as LessonSession;
  for (const word of session.words) {
    await expect(
      page.getByRole("heading", { name: word.wordText, exact: true }),
    ).toBeVisible();
    const changed = page.waitForResponse(
      (response) =>
        response.request().method() === "POST" &&
        response.url().endsWith("/lesson-sessions/" + session.id + "/actions"),
    );
    await page.getByRole("button", { name: "Continue", exact: true }).click();
    session = (await (await changed).json()) as LessonSession;
  }
  return session;
}

async function openListeningPractice(page: Page) {
  await page.goto("/practice");
  const opened = page.waitForResponse(
    (response) =>
      response.request().method() === "POST" &&
      new URL(response.url()).pathname === "/api/v1/practice-sessions",
  );
  await page
    .getByRole("button", { name: "Start listening practice", exact: true })
    .click();
  const session = (await (await opened).json()) as PracticeSession;
  await expect(page).toHaveURL(
    new RegExp("/practice/session/" + session.id + "$"),
  );
  return session;
}

async function expectActionOnScreen(page: Page, buttonName: string) {
  await page.evaluate(() => window.scrollTo(0, 0));
  const button = page.getByRole("button", { name: buttonName, exact: true });
  await expect(button).toBeVisible();
  const bounds = await button.boundingBox();
  expect(bounds).not.toBeNull();
  expect(bounds!.y).toBeGreaterThanOrEqual(0);
  expect(bounds!.y + bounds!.height).toBeLessThanOrEqual(
    page.viewportSize()!.height,
  );
}

for (const theme of ["light", "dark"] as const) {
  test(
    "lesson selection can change before checking, with truthful retries in " +
      theme,
    async ({ page, context, baseURL }) => {
      await context.addCookies([
        { name: "vocanova_theme", value: theme, url: baseURL! },
      ]);
      const session = await openLessonQuestion(page);
      const actions: unknown[] = [];
      page.on("request", (request) => {
        if (
          request.method() === "POST" &&
          request.url().endsWith("/lesson-sessions/" + session.id + "/actions")
        ) {
          actions.push(request.postDataJSON());
        }
      });
      const choices = page
        .getByRole("group", { name: "Answer choices" })
        .getByRole("radio");
      const check = page.getByRole("button", {
        name: "Check answer",
        exact: true,
      });
      await expect(check).toBeDisabled();
      await choices.first().focus();
      await page.keyboard.press("Space");
      await expect(choices.first()).toBeChecked();
      await page.keyboard.press("ArrowDown");
      await expect(choices.nth(1)).toBeChecked();
      await expect(choices.first()).not.toBeChecked();
      expect(actions).toEqual([]);
      await expect(check).toBeEnabled();
      await expectActionOnScreen(page, "Check answer");
      await page
        .getByRole("radio", { name: session.words[1]!.definition, exact: true })
        .check();
      expect(actions).toEqual([]);
      const graded = page.waitForResponse(
        (response) =>
          response.request().method() === "POST" &&
          response
            .url()
            .endsWith("/lesson-sessions/" + session.id + "/actions"),
      );
      await check.click();
      const first = (await (await graded).json()) as LessonSession;
      expect(first).toMatchObject({
        questionsAnswered: 1,
        firstAnswersCorrect: 0,
        canContinue: false,
      });
      expect(actions).toHaveLength(1);
      await expect(
        page.getByRole("heading", { name: "Let’s look again", exact: true }),
      ).toBeVisible();
      await expect(check).toBeDisabled();
      await expect(
        page.getByRole("button", { name: "Continue", exact: true }),
      ).toHaveCount(0);
      await page
        .getByRole("radio", { name: session.words[0]!.definition, exact: true })
        .check();
      const retried = page.waitForResponse(
        (response) =>
          response.request().method() === "POST" &&
          response
            .url()
            .endsWith("/lesson-sessions/" + session.id + "/actions"),
      );
      await check.click();
      expect((await (await retried).json()) as LessonSession).toMatchObject({
        questionsAnswered: 1,
        firstAnswersCorrect: 0,
        canContinue: true,
      });
      await expectActionOnScreen(page, "Continue");
      await expect(choices.first()).toBeDisabled();
      await expect(page.locator("html")).toHaveAttribute("data-theme", theme);
      expect((await scanForAxeViolations(page)).criticalOrSerious).toEqual([]);
      expect(
        await page.evaluate(() => document.documentElement.scrollWidth),
      ).toBeLessThanOrEqual(page.viewportSize()!.width);
    },
  );

  test(
    "listening selection waits for Check and resets for the next question in " +
      theme,
    async ({ page, context, baseURL }) => {
      await context.addCookies([
        { name: "vocanova_theme", value: theme, url: baseURL! },
      ]);
      const session = await openListeningPractice(page);
      const actions: unknown[] = [];
      page.on("request", (request) => {
        if (
          request.method() === "POST" &&
          request
            .url()
            .endsWith("/practice-sessions/" + session.id + "/actions")
        ) {
          actions.push(request.postDataJSON());
        }
      });
      const radios = page
        .getByRole("group", { name: "Answer choices" })
        .getByRole("radio");
      const check = page.getByRole("button", {
        name: "Check answer",
        exact: true,
      });
      await expect(check).toBeDisabled();
      await radios.nth(1).check();
      await radios.first().focus();
      await page.keyboard.press("Space");
      await expect(radios.first()).toBeChecked();
      await expect(radios.nth(1)).not.toBeChecked();
      expect(actions).toEqual([]);
      await expectActionOnScreen(page, "Check answer");
      const graded = page.waitForResponse(
        (response) =>
          response.request().method() === "POST" &&
          response
            .url()
            .endsWith("/practice-sessions/" + session.id + "/actions"),
      );
      await check.click();
      expect((await (await graded).json()) as PracticeSession).toMatchObject({
        questionsAnswered: 1,
        firstAnswersCorrect: 1,
        canContinue: true,
      });
      expect(actions).toHaveLength(1);
      await expectActionOnScreen(page, "Continue");
      await page.getByRole("button", { name: "Continue", exact: true }).click();
      await expect(check).toBeVisible();
      await expect(check).toBeDisabled();
      await expect(page.getByRole("radio", { checked: true })).toHaveCount(0);
      expect((await scanForAxeViolations(page)).criticalOrSerious).toEqual([]);
    },
  );
}

test("a lost listening answer retries the same confirmed choice without re-selection", async ({
  page,
}) => {
  const session = await openListeningPractice(page);
  const sent: { key: string | undefined; body: string | null }[] = [];
  await page.route(
    "**/api/v1/practice-sessions/" + session.id + "/actions",
    async (route) => {
      sent.push({
        key: route.request().headers()["idempotency-key"],
        body: route.request().postData(),
      });
      if (sent.length === 1) {
        await route.fetch();
        await route.abort("failed");
      } else await route.continue();
    },
  );
  const radios = page
    .getByRole("group", { name: "Answer choices" })
    .getByRole("radio");
  await radios.first().check();
  expect(sent).toEqual([]);
  await page.getByRole("button", { name: "Check answer", exact: true }).click();
  await expect(page.getByRole("main").getByRole("alert")).toBeVisible();
  await expect(radios.first()).toBeChecked();
  await expect(radios.first()).toBeDisabled();
  await expect(
    page.getByRole("button", { name: "Check answer", exact: true }),
  ).toBeDisabled();
  await page.getByRole("button", { name: "Retry answer", exact: true }).click();
  await expect(
    page.getByRole("heading", { name: "You remembered it", exact: true }),
  ).toBeVisible();
  expect(sent).toHaveLength(2);
  expect(sent[0]!.key).toBeTruthy();
  expect(sent[1]).toEqual(sent[0]);
});
