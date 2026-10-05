import { randomUUID } from "node:crypto";
import { expect, test } from "@playwright/test";
import type { LessonAction, LessonSession } from "@vocanova/api-client";

test.beforeEach(async ({ context, baseURL }) => {
  if (!baseURL) throw new Error("Missing app URL");
  await context.addCookies([
    { name: "vocanova_session", value: randomUUID(), url: baseURL },
    { name: "vocanova_csrf", value: randomUUID(), url: baseURL },
  ]);
});

test("a larger curriculum groups lessons without locking other situations", async ({
  page,
  context,
  baseURL,
}) => {
  await context.addCookies([
    { name: "e2e_lessons", value: "expanded", url: baseURL! },
  ]);
  await page.goto("/discover");
  const section = page.getByRole("region", {
    name: "Guided lessons",
    exact: true,
  });
  await expect(
    section.getByText("0 of 30 lessons completed.", { exact: false }),
  ).toBeVisible();
  const travel = section
    .locator("details")
    .filter({
      has: page.locator("summary").filter({ hasText: "Travel basics" }),
    });
  await expect(travel).not.toHaveAttribute("open");
  await travel.locator("summary").focus();
  await page.keyboard.press("Enter");
  await expect(
    travel.getByRole("link", { name: /^Start lesson\s*:\s*Fixture lesson 7$/ }),
  ).toBeVisible();
  await expect(
    travel.getByRole("link", { name: /^Start lesson\s*:\s*Fixture lesson 7$/ }),
  ).toHaveAttribute("href", "/learn/fixture-lesson-6");
  expect(
    await page.evaluate(() => document.documentElement.scrollWidth),
  ).toBeLessThanOrEqual(page.viewportSize()!.width);
  const conversation = section.locator("details").filter({ has: page.locator("summary").filter({ hasText: "Daily Conversation" }) });
  await conversation.locator("summary").focus();
  await page.keyboard.press("Enter");
  await section
    .getByRole("link", {
      name: /^Start lesson\s*:\s*Make a plan with a friend$/,
    })
    .click();
  await expect(
    page.getByRole("main").getByRole("heading", { level: 1 }),
  ).toHaveText("Make a plan with a friend");
});

test("a guided lesson resumes and completes teaching, recall and context", async ({
  page,
}) => {
  await page.goto("/learn/conversation-basics");
  const opened = page.waitForResponse(
    (response) =>
      response.request().method() === "POST" &&
      response.url().endsWith("/lessons/conversation-basics/sessions"),
  );
  await page.getByRole("button", { name: "Start lesson", exact: true }).click();
  const { words } = await (await opened).json();
  await expect(
    page.getByRole("heading", { name: "invite", exact: true }),
  ).toBeVisible();
  await page.getByRole("button", { name: "Continue", exact: true }).click();
  await expect(
    page.getByRole("heading", { name: "confirm", exact: true }),
  ).toBeVisible();
  await page.reload();
  await expect(
    page.getByRole("heading", { name: "confirm", exact: true }),
  ).toBeVisible();
  await page.getByRole("button", { name: "Continue", exact: true }).click();
  await expect(
    page.getByRole("heading", { name: "reschedule", exact: true }),
  ).toBeVisible();
  await page.getByRole("button", { name: "Continue", exact: true }).click();
  // Use fixture response metadata to identify authored options; the UI must
  // wait for server grading rather than infer correctness from this test.
  let wrongAnswerUsed = false;
  for (let index = 0; index < 6; index++) {
    const word = words[index % 3];
    if (!wrongAnswerUsed) {
      await page
        .getByRole("radio", { name: words[1].definition, exact: true })
        .check();
      await page
        .getByRole("button", { name: "Check answer", exact: true })
        .click();
      await expect(
        page.getByRole("heading", { name: "Let’s look again" }),
      ).toBeVisible();
      await expect(
        page.getByRole("button", { name: "Continue", exact: true }),
      ).toHaveCount(0);
      await page.reload();
      await expect(
        page.getByRole("heading", { name: "Let’s look again" }),
      ).toBeVisible();
      wrongAnswerUsed = true;
    }
    await page
      .getByRole("radio", {
        name: index < 3 ? word.definition : word.wordText,
        exact: true,
      })
      .check();
    await page
      .getByRole("button", { name: "Check answer", exact: true })
      .click();
    await expect(
      page.getByRole("heading", { name: "That’s right" }),
    ).toBeVisible();
    await page
      .getByRole("button", {
        name: index === 5 ? "Finish lesson" : "Continue",
        exact: true,
      })
      .click();
  }
  await expect(
    page.getByRole("heading", { name: "Lesson complete", exact: true }),
  ).toBeFocused();
  await expect(
    page.getByText("5 of 6 questions correct on your first try.", {
      exact: false,
    }),
  ).toBeVisible();
  await page.reload();
  await expect(
    page.getByRole("heading", { name: "Lesson complete", exact: true }),
  ).toBeVisible();
  await page.goto("/progress");
  await expect(
    page
      .getByRole("main")
      .getByText("1 of 1 lessons completed", { exact: true }),
  ).toBeVisible();
  await expect(
    page
      .getByRole("main")
      .getByText("0 saved meanings, each at its own stage.", { exact: true }),
  ).toBeVisible();
});

test("an applied lesson action can be retried after its response is lost", async ({
  page,
}) => {
  await page.goto("/learn/conversation-basics");
  await page.getByRole("button", { name: "Start lesson", exact: true }).click();
  await expect(
    page.getByRole("heading", { name: "invite", exact: true }),
  ).toBeVisible();
  const requests: { key: string | undefined; body: string | null }[] = [];
  await page.route("**/api/v1/lesson-sessions/*/actions", async (route) => {
    requests.push({
      key: route.request().headers()["idempotency-key"],
      body: route.request().postData(),
    });
    if (requests.length === 1) {
      await route.fetch();
      await route.abort("failed");
    } else await route.continue();
  });
  await page.getByRole("button", { name: "Continue", exact: true }).click();
  await expect(page.getByRole("main").getByRole("alert")).toContainText(
    "could not confirm",
  );
  await expect(
    page.getByRole("button", { name: "Continue", exact: true }),
  ).toBeDisabled();
  await page.getByRole("button", { name: "Retry step" }).click();
  await expect(
    page.getByRole("heading", { name: "confirm", exact: true }),
  ).toBeVisible();
  expect(requests).toHaveLength(2);
  expect(requests[0]!.key).toBeTruthy();
  expect(requests[1]).toEqual(requests[0]);
  await expect(page.getByText("1 of 9 steps", { exact: true })).toBeVisible();
});

test("a failed CSRF preflight preserves the chosen lesson answer for retry", async ({
  page,
  context,
}) => {
  await page.goto("/learn/conversation-basics");
  const started = page.waitForResponse(
    (response) =>
      response.request().method() === "POST" &&
      new URL(response.url()).pathname ===
        "/api/v1/lessons/conversation-basics/sessions",
  );
  await page.getByRole("button", { name: "Start lesson", exact: true }).click();
  let session = (await (await started).json()) as LessonSession;
  for (const word of session.words) {
    await expect(
      page.getByRole("heading", { name: word.wordText, exact: true }),
    ).toBeVisible();
    const advanced = page.waitForResponse(
      (response) =>
        response.request().method() === "POST" &&
        new URL(response.url()).pathname ===
          `/api/v1/lesson-sessions/${session.id}/actions`,
    );
    await page.getByRole("button", { name: "Continue", exact: true }).click();
    session = (await (await advanced).json()) as LessonSession;
  }
  expect(session.currentStep?.kind).toBe("recall");
  const expectedStep = session.currentStep!.id;
  const expectedRevision = session.revision;
  // Choose a specific wrong option: retry must preserve this selection, not
  // silently choose the target meaning or manufacture a correct result.
  const chosenWord = session.words[1]!;
  const choice = page.getByRole("radio", {
    name: chosenWord.definition,
    exact: true,
  });
  await expect(choice).toBeEnabled();

  const requests: { key: string | undefined; body: LessonAction }[] = [];
  page.on("request", (request) => {
    if (
      request.method() === "POST" &&
      new URL(request.url()).pathname ===
        `/api/v1/lesson-sessions/${session.id}/actions`
    ) {
      requests.push({
        key: request.headers()["idempotency-key"],
        body: request.postDataJSON() as LessonAction,
      });
    }
  });
  let recoveryRequests = 0;
  await page.route("**/api/v1/me", async (route) => {
    if (route.request().method() !== "GET") return route.continue();
    recoveryRequests += 1;
    if (recoveryRequests === 1) return route.abort("failed");
    return route.continue();
  });
  // The helper reads the cookie on every call. Remove it after initial shell
  // recovery and successful lesson actions, so this is the mutation preflight.
  await context.clearCookies({ name: "vocanova_csrf" });
  await choice.check();
  await page.getByRole("button", { name: "Check answer", exact: true }).click();
  const main = page.getByRole("main");
  await expect(main.getByRole("alert")).toContainText("could not confirm");
  expect(recoveryRequests).toBe(1);
  expect(requests).toEqual([]);
  for (const answer of await main
    .getByRole("group", { name: "Answer choices" })
    .getByRole("radio")
    .all()) {
    await expect(answer).toBeDisabled();
  }
  await expect(main.getByText("3 of 9 steps", { exact: true })).toBeVisible();
  const graded = page.waitForResponse(
    (response) =>
      response.request().method() === "POST" &&
      new URL(response.url()).pathname ===
        `/api/v1/lesson-sessions/${session.id}/actions`,
  );
  await main.getByRole("button", { name: "Retry step", exact: true }).click();
  const gradedResponse = await graded;
  expect(gradedResponse.status()).toBe(200);
  const saved = (await gradedResponse.json()) as LessonSession;
  expect(saved.revision).toBe(expectedRevision + 1);
  expect(saved.questionsAnswered).toBe(1);
  expect(saved.firstAnswersCorrect).toBe(0);
  await expect(
    main.getByRole("heading", { name: "Let’s look again", exact: true }),
  ).toBeVisible();
  expect(recoveryRequests).toBe(2);
  expect(requests).toHaveLength(1);
  const sent = requests[0]!;
  expect(sent.key).toBeTruthy();
  expect(sent.body).toEqual({
    stepId: expectedStep,
    expectedRevision,
    action: "answer",
    choiceId: chosenWord.meaningId,
    clientActionId: sent.key,
  });
  await expect(choice).toBeEnabled();
  await expect(
    main.getByRole("button", { name: "Continue", exact: true }),
  ).toHaveCount(0);
  await expect(main.getByRole("alert")).toHaveCount(0);
  // Persistence confirms that the recovery produced one actual answer; it did
  // not merely dismiss the error or show client-only feedback.
  await page.reload();
  await expect(
    main.getByRole("heading", { name: "Let’s look again", exact: true }),
  ).toBeVisible();
  await expect(main.getByText("3 of 9 steps", { exact: true })).toBeVisible();
  expect(requests).toHaveLength(1);
});

test("a stale lesson loads server progress before allowing another answer", async ({
  page,
}) => {
  await page.goto("/learn/conversation-basics");
  await page.getByRole("button", { name: "Start lesson", exact: true }).click();
  await expect(
    page.getByRole("heading", { name: "invite", exact: true }),
  ).toBeVisible();
  await page.route(
    "**/api/v1/lesson-sessions/*/actions",
    async (route) => {
      await route.fetch();
      await route.fulfill({
        status: 409,
        contentType: "application/json",
        body: JSON.stringify({ detail: "stale" }),
      });
    },
    { times: 1 },
  );
  await page.getByRole("button", { name: "Continue", exact: true }).click();
  await expect(page.getByRole("main").getByRole("alert")).toContainText(
    "changed in another request",
  );
  await page.getByRole("button", { name: "Load saved progress" }).click();
  await expect(
    page.getByRole("heading", { name: "confirm", exact: true }),
  ).toBeVisible();
  await expect(
    page.getByRole("button", { name: "Continue", exact: true }),
  ).toBeEnabled();
});
