import { randomUUID } from "node:crypto";
import { expect, test, type Page } from "@playwright/test";
import type { PracticeSession } from "@vocanova/api-client";

const apiURL = `http://127.0.0.1:${process.env.MOCK_API_PORT ?? 8080}`;
const main = (page: Page) => page.getByRole("main");

test.beforeEach(async ({ context, baseURL }) => {
  if (!baseURL) throw new Error("Missing app URL");
  await context.addCookies([
    { name: "vocanova_session", value: randomUUID(), url: baseURL },
    { name: "vocanova_csrf", value: randomUUID(), url: baseURL },
  ]);
});

async function start(page: Page, name = "Start typed recall") {
  await page.goto("/practice");
  const response = page.waitForResponse(
    (r) =>
      r.request().method() === "POST" &&
      new URL(r.url()).pathname === "/api/v1/practice-sessions",
  );
  await main(page).getByRole("button", { name, exact: true }).click();
  const result = await response;
  expect(result.status()).toBe(200);
  const session = (await result.json()) as PracticeSession;
  await expect(page).toHaveURL(new RegExp(`/practice/session/${session.id}$`));
  return session;
}

async function act(page: Page, button: string) {
  const response = page.waitForResponse(
    (r) =>
      r.request().method() === "POST" &&
      /\/practice-sessions\/[^/]+\/actions$/u.test(new URL(r.url()).pathname),
  );
  await main(page).getByRole("button", { name: button, exact: true }).click();
  const result = await response;
  expect(result.status()).toBe(200);
  const session = (await result.json()) as PracticeSession;
  if (session.feedback)
    await expect(
      main(page).getByText(session.feedback.explanation, { exact: true }),
    ).toBeVisible();
  else if (session.currentStep)
    await expect(
      main(page).getByText(session.currentStep.prompt, { exact: true }),
    ).toBeVisible();
  else
    await expect(
      main(page).getByRole("heading", {
        name: "Practice complete",
        exact: true,
      }),
    ).toBeVisible();
  return session;
}

async function saved(page: Page, id: string) {
  const response = await page.request.get(
    `${apiURL}/api/v1/practice-sessions/${id}`,
  );
  expect(response.status()).toBe(200);
  return (await response.json()) as PracticeSession;
}

test("returning through Practice navigation offers the saved session and a fresh start", async ({
  page,
}) => {
  const initial = await start(page);
  await page
    .getByRole("link", { name: "Back to Practice", exact: true })
    .click();
  await expect(page).toHaveURL(/\/practice$/);
  await expect(
    main(page).getByRole("button", { name: "Start typed recall", exact: true }),
  ).toBeEnabled();
  const resume = main(page).getByRole("link", {
    name: /^Resume practice\s*:\s*Type the word$/,
  });
  await expect(resume).toHaveAttribute(
    "href",
    `/practice/session/${initial.id}`,
  );
  await resume.click();
  await expect(page).toHaveURL(new RegExp(`/practice/session/${initial.id}$`));
  await expect(
    main(page).getByRole("textbox", { name: "Your answer" }),
  ).toBeEnabled();
});

test("typed recall saves normalized, wrong and assisted answers without review credit and can repeat", async ({
  page,
}) => {
  const initial = await start(page);
  expect(initial.totalSteps).toBe(3);
  await expect(
    main(page).getByRole("textbox", { name: "Your answer" }),
  ).toBeVisible();
  await main(page)
    .getByRole("textbox", { name: "Your answer" })
    .fill("  InViTe  ");
  let session = await act(page, "Check answer");
  expect(session.feedback).toMatchObject({
    correct: true,
    assisted: false,
    answer: "invite",
  });
  expect(session.firstAnswersCorrect).toBe(1);
  await page.reload();
  await expect(
    main(page).getByText(session.feedback!.explanation, { exact: true }),
  ).toBeVisible();
  session = await act(page, "Continue");
  expect(session.completedSteps).toBe(1);
  await main(page).getByRole("textbox", { name: "Your answer" }).fill("cancel");
  session = await act(page, "Check answer");
  expect(session.feedback).toMatchObject({
    correct: false,
    assisted: false,
    answer: "confirm",
  });
  await page.reload();
  await expect(
    main(page).getByText(session.feedback!.explanation, { exact: true }),
  ).toBeVisible();
  await main(page)
    .getByRole("textbox", { name: "Your answer" })
    .fill("confirm");
  session = await act(page, "Check answer");
  expect(session.feedback).toMatchObject({ correct: true, assisted: true });
  expect(session.questionsAnswered).toBe(2);
  expect(session.firstAnswersCorrect).toBe(1);
  await act(page, "Continue");
  session = await act(page, "Show answer");
  expect(session.feedback).toMatchObject({
    correct: false,
    assisted: true,
    answer: "reschedule",
  });
  session = await act(page, "Finish practice");
  expect(session).toMatchObject({
    status: "completed",
    completedSteps: 3,
    questionsAnswered: 3,
    firstAnswersCorrect: 1,
  });
  await expect(
    main(page).getByText("1 of 3 correct on your first try without help.", {
      exact: true,
    }),
  ).toBeVisible();
  await expect(
    main(page).getByRole("heading", { name: "Practice complete", exact: true }),
  ).toBeFocused();
  await page.reload();
  await expect(
    main(page).getByRole("heading", { name: "Practice complete", exact: true }),
  ).toBeVisible();
  expect(await saved(page, initial.id)).toMatchObject({
    status: "completed",
    firstAnswersCorrect: 1,
    questionsAnswered: 3,
  });
  // Self-study must not save words, award mastery or advance the review target.
  await page.goto("/progress");
  await expect(
    main(page).getByText("0 saved · 0 ready for review", {
      exact: true,
    }),
  ).toBeVisible();
  const mission = await page.request.get(`${apiURL}/api/v1/daily-mission`);
  expect(mission.status()).toBe(200);
  expect((await mission.json()).reviewsCompleted).toBe(0);
  const repeated = await start(page);
  expect(repeated.id).not.toBe(initial.id);
  expect(repeated).toMatchObject({
    revision: 0,
    completedSteps: 0,
    questionsAnswered: 0,
    firstAnswersCorrect: 0,
  });
});

test("a lost practice answer response retries one unchanged action and key", async ({
  page,
}) => {
  const initial = await start(page);
  const requests: { key: string | undefined; body: string | null }[] = [];
  await page.route(
    `**/api/v1/practice-sessions/${initial.id}/actions`,
    async (route) => {
      requests.push({
        key: route.request().headers()["idempotency-key"],
        body: route.request().postData(),
      });
      if (requests.length === 1) {
        const response = await route.fetch();
        expect(response.status()).toBe(200);
        return route.abort("failed");
      }
      return route.continue();
    },
  );
  await main(page).getByRole("textbox", { name: "Your answer" }).fill("invite");
  await main(page)
    .getByRole("button", { name: "Check answer", exact: true })
    .click();
  await expect(main(page).getByRole("alert")).toBeVisible();
  await expect(
    main(page).getByRole("textbox", { name: "Your answer" }),
  ).toBeDisabled();
  expect(await saved(page, initial.id)).toMatchObject({
    revision: 1,
    questionsAnswered: 1,
    firstAnswersCorrect: 1,
  });
  const result = await act(page, "Retry answer");
  expect(result).toMatchObject({
    revision: 1,
    completedSteps: 0,
    questionsAnswered: 1,
    firstAnswersCorrect: 1,
  });
  expect(requests).toHaveLength(2);
  expect(requests[0]!.key).toBeTruthy();
  expect(requests[1]).toEqual(requests[0]);
  expect(JSON.parse(requests[0]!.body!)).toMatchObject({
    typedAnswer: "invite",
    stepId: initial.currentStep!.id,
    expectedRevision: 0,
    clientActionId: requests[0]!.key,
  });
});

test("a stale practice action loads server progress instead of replaying a new answer", async ({
  page,
}) => {
  const initial = await start(page);
  let intercepted = false;
  await page.route(
    `**/api/v1/practice-sessions/${initial.id}/actions`,
    async (route) => {
      if (!intercepted) {
        intercepted = true;
        const otherKey = randomUUID();
        const response = await route.fetch({
          headers: {
            ...route.request().headers(),
            "idempotency-key": otherKey,
          },
          postData: JSON.stringify({
            ...route.request().postDataJSON(),
            clientActionId: otherKey,
            typedAnswer: "invite",
          }),
        });
        expect(response.status()).toBe(200);
      }
      return route.continue();
    },
  );
  await main(page).getByRole("textbox", { name: "Your answer" }).fill("cancel");
  const conflict = page.waitForResponse(
    (r) =>
      r.status() === 409 &&
      r.url().endsWith(`/practice-sessions/${initial.id}/actions`),
  );
  await main(page)
    .getByRole("button", { name: "Check answer", exact: true })
    .click();
  await conflict;
  await expect(main(page).getByRole("alert")).toBeVisible();
  await main(page)
    .getByRole("button", { name: "Load saved progress", exact: true })
    .click();
  const state = await saved(page, initial.id);
  await expect(
    main(page).getByText(state.feedback!.explanation, { exact: true }),
  ).toBeVisible();
  expect(state).toMatchObject({
    revision: 1,
    questionsAnswered: 1,
    firstAnswersCorrect: 1,
  });
  await expect(main(page).getByRole("alert")).toHaveCount(0);
  const next = await act(page, "Continue");
  expect(next.completedSteps).toBe(1);
});

test("past mistakes come from a submitted wrong answer and resolve only in independent practice", async ({
  page,
}) => {
  await page.goto("/practice");
  await expect(
    main(page).getByText("No past mistakes to revisit right now.", {
      exact: true,
    }),
  ).toBeVisible();
  const original = await start(page);
  await main(page).getByRole("textbox", { name: "Your answer" }).fill("cancel");
  await act(page, "Check answer");
  const attempt = await start(page, "Practise past mistakes");
  expect(attempt).toMatchObject({ mode: "mistakes", totalSteps: 1 });
  const assisted = await act(page, "Show answer");
  expect(assisted.feedback).toMatchObject({ correct: false, assisted: true });
  await act(page, "Finish practice");
  await page.goto("/practice");
  await expect(
    main(page).getByText("1 word is ready for another try.", { exact: true }),
  ).toBeVisible();
  const independent = await start(page, "Practise past mistakes");
  expect(independent.id).not.toBe(attempt.id);
  await main(page).getByRole("textbox", { name: "Your answer" }).fill("invite");
  let result = await act(page, "Check answer");
  expect(result.feedback).toMatchObject({ correct: true, assisted: false });
  result = await act(page, "Finish practice");
  expect(result.firstAnswersCorrect).toBe(1);
  await page.goto("/practice");
  await expect(
    main(page).getByText("No past mistakes to revisit right now.", {
      exact: true,
    }),
  ).toBeVisible();
  expect((await saved(page, original.id)).feedback).toMatchObject({
    correct: false,
    assisted: false,
  });
});

test("a lost start response opens the same practice session on retry", async ({
  page,
}) => {
  await page.goto("/practice");
  const requests: { key: string | undefined; body: string | null }[] = [];
  let createdID = "";
  await page.route("**/api/v1/practice-sessions", async (route) => {
    if (route.request().method() !== "POST") return route.continue();
    requests.push({
      key: route.request().headers()["idempotency-key"],
      body: route.request().postData(),
    });
    if (requests.length === 1) {
      const response = await route.fetch();
      expect(response.status()).toBe(200);
      createdID = (await response.json()).id;
      return route.abort("failed");
    }
    return route.continue();
  });
  await main(page)
    .getByRole("button", { name: "Start typed recall", exact: true })
    .click();
  await expect(main(page).getByRole("alert")).toBeVisible();
  await expect(
    main(page).getByRole("button", {
      name: "Start listening practice",
      exact: true,
    }),
  ).toBeDisabled();
  await main(page)
    .getByRole("button", { name: "Retry start", exact: true })
    .click();
  await expect(page).toHaveURL(new RegExp(`/practice/session/${createdID}$`));
  expect(requests).toHaveLength(2);
  expect(requests[0]!.key).toBeTruthy();
  expect(requests[1]).toEqual(requests[0]);
  expect(JSON.parse(requests[0]!.body!)).toEqual({ mode: "typed_recall" });
  const list = await page.request.get(`${apiURL}/api/v1/practice-sessions`);
  expect(list.status()).toBe(200);
  expect(
    (await list.json()).items.map((item: PracticeSession) => item.id),
  ).toEqual([createdID]);
});

async function installAudio(page: Page, unsupported = false) {
  await page.addInitScript((unavailable) => {
    const calls: { text: string; rate: number }[] = [];
    const synthesis = {
      speaking: false,
      pending: false,
      getVoices: () => [{ lang: "en-GB", localService: true, default: true }],
      speak: (utterance: {
        text: string;
        rate: number;
        onstart?: () => void;
        onend?: () => void;
      }) => {
        calls.push({ text: utterance.text, rate: utterance.rate });
        utterance.onstart?.();
        utterance.onend?.();
      },
      cancel: () => undefined,
    };
    Object.defineProperty(window, "__practiceAudioCalls", { value: calls });
    Object.defineProperty(window, "speechSynthesis", {
      configurable: true,
      value: unavailable ? undefined : synthesis,
    });
    Object.defineProperty(window, "SpeechSynthesisUtterance", {
      configurable: true,
      value: unavailable
        ? undefined
        : class {
            constructor(public text: string) {}
          },
    });
  }, unsupported);
}

for (const theme of ["light", "dark"] as const) {
  test(`listening uses explicit generic audio controls without revealing the answer in ${theme} mode`, async ({
    page,
    context,
    baseURL,
  }, testInfo) => {
    if (!baseURL) throw new Error("Missing app URL");
    await context.addCookies([
      { name: "vocanova_theme", value: theme, url: baseURL },
    ]);
    await installAudio(page);
    const session = await start(page, "Start listening practice");
    const calls = () =>
      page.evaluate(
        () =>
          (
            window as unknown as {
              __practiceAudioCalls: { text: string; rate: number }[];
            }
          ).__practiceAudioCalls,
      );
    expect(await calls()).toEqual([]);
    await expect(page.locator("html")).toHaveAttribute("data-theme", theme);
    await expect(
      main(page).getByRole("button", { name: "Play audio", exact: true }),
    ).toBeVisible();
    // Speech text must not leak through headings, accessible names or tooltips.
    expect(await main(page).innerText()).not.toMatch(/\binvite\b/iu);
    expect(await main(page).ariaSnapshot()).not.toMatch(/\binvite\b/iu);
    await main(page)
      .getByRole("button", { name: "Play audio", exact: true })
      .click();
    await expect.poll(calls).toEqual([{ text: "invite", rate: 1 }]);
    await main(page)
      .getByRole("button", { name: "Slower audio", exact: true })
      .click();
    await main(page)
      .getByRole("button", { name: "Play audio", exact: true })
      .click();
    await expect.poll(calls).toEqual([
      { text: "invite", rate: 1 },
      { text: "invite", rate: 0.75 },
    ]);
    expect((await saved(page, session.id)).questionsAnswered).toBe(0);
    const options = main(page).getByRole("group", { name: "Answer choices" });
    await expect(options.getByRole("radio")).toHaveCount(3);
    expect(
      await page.evaluate(
        () => document.documentElement.scrollWidth <= window.innerWidth,
      ),
    ).toBe(true);
    await page.screenshot({
      path: testInfo.outputPath(`listening-${theme}.png`),
      fullPage: true,
    });
    const graded = page.waitForResponse(
      (r) =>
        r.request().method() === "POST" &&
        r.url().endsWith(`/practice-sessions/${session.id}/actions`),
    );
    await options
      .getByRole("radio", {
        name: "To ask someone to come to an event or do something with you.",
        exact: true,
      })
      .check();
    await main(page)
      .getByRole("button", { name: "Check answer", exact: true })
      .click();
    const result = (await (await graded).json()) as PracticeSession;
    expect(result.feedback).toMatchObject({
      correct: true,
      assisted: false,
      answer: "invite",
    });
    await expect(
      main(page).getByText(result.feedback!.explanation, { exact: true }),
    ).toBeVisible();
    await page.reload();
    expect(await calls()).toEqual([]);
    await expect(
      main(page).getByText(result.feedback!.explanation, { exact: true }),
    ).toBeVisible();
  });
}

test("unsupported listening keeps its ungraded session and offers another mode", async ({
  page,
}) => {
  await installAudio(page, true);
  const session = await start(page, "Start listening practice");
  await main(page)
    .getByRole("button", { name: "Play audio", exact: true })
    .click();
  await expect(
    main(page).getByText("Pronunciation is not available in this browser.", {
      exact: true,
    }),
  ).toBeVisible();
  await expect(
    main(page).getByRole("link", { name: "Choose another practice mode" }),
  ).toHaveAttribute("href", "/practice");
  expect(await saved(page, session.id)).toMatchObject({
    revision: 0,
    completedSteps: 0,
    questionsAnswered: 0,
  });
  await page.reload();
  expect(await saved(page, session.id)).toMatchObject({
    revision: 0,
    status: "in_progress",
  });
  await expect(
    main(page).getByRole("button", { name: "Play audio", exact: true }),
  ).toBeVisible();
});
