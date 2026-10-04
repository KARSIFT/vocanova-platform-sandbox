import { randomUUID } from "node:crypto";
import { expect, test, type Page } from "@playwright/test";
import { scanForAxeViolations } from "./axe-helper";
import type { StorySession } from "@vocanova/api-client";

const apiURL = `http://127.0.0.1:${process.env.MOCK_API_PORT ?? 8080}`;
const main = (page: Page) => page.getByRole("main");
test.beforeEach(async ({ context, baseURL }) => {
  if (!baseURL) throw new Error("Missing app URL");
  await context.addCookies([
    { name: "vocanova_session", value: randomUUID(), url: baseURL },
    { name: "vocanova_csrf", value: randomUUID(), url: baseURL },
  ]);
});
async function start(page: Page) {
  await page.goto("/stories/a-quiet-lunch");
  const response = page.waitForResponse(
    (r) =>
      r.request().method() === "POST" &&
      new URL(r.url()).pathname === "/api/v1/story-sessions",
  );
  await main(page)
    .getByRole("button", { name: "Start story", exact: true })
    .click();
  const result = await response;
  expect(result.status()).toBe(200);
  const session = (await result.json()) as StorySession;
  await expect(page).toHaveURL(new RegExp(`/stories/session/${session.id}$`));
  return session;
}
async function act(page: Page, name = "Continue") {
  const response = page.waitForResponse(
    (r) =>
      r.request().method() === "POST" &&
      /\/story-sessions\/[^/]+\/actions$/u.test(new URL(r.url()).pathname),
  );
  await main(page).getByRole("button", { name, exact: true }).click();
  const result = await response;
  expect(result.status()).toBe(200);
  const session = (await result.json()) as StorySession;
  if (session.feedback)
    await expect(
      main(page).getByText(session.feedback.explanation, { exact: true }),
    ).toBeVisible();
  else if (session.currentStep?.kind === "line")
    await expect(
      main(page).getByText(session.currentStep.line!.text, { exact: true }),
    ).toBeVisible();
  else if (session.currentStep)
    await expect(
      main(page).getByText(session.currentStep.prompt!, { exact: true }),
    ).toBeVisible();
  else
    await expect(
      main(page).getByRole("heading", { name: "Story complete", exact: true }),
    ).toBeVisible();
  return session;
}
test("six original situations, full transcript, vocabulary and optional audio", async ({
  page,
  context,
  baseURL,
}) => {
  await context.addCookies([
    { name: "e2e_unit_guides", value: "true", url: baseURL! },
  ]);
  await page.addInitScript(() =>
    Object.defineProperty(window, "speechSynthesis", {
      value: undefined,
      configurable: true,
    }),
  );
  await page.goto("/stories");
  for (const title of [
    "A quiet lunch",
    "The right platform",
    "A clear deadline",
    "A better fit",
    "Plans for Saturday",
    "A room for tonight",
  ])
    await expect(
      main(page).getByRole("heading", { name: title, exact: true }),
    ).toBeVisible();
  await main(page)
    .getByRole("link", { name: "Read and practise", exact: true })
    .first()
    .click();
  await expect(
    main(page).getByRole("heading", { name: "Full story", exact: true }),
  ).toBeVisible();
  await expect(
    main(page).getByRole("button", { name: /^Play audio:/ }),
  ).toHaveCount(8);
  await main(page)
    .getByRole("button", { name: /^Play audio:/ })
    .first()
    .click();
  await expect(
    main(page)
      .getByText(/audio|pronunciation/i)
      .first(),
  ).toBeVisible();
  await main(page).getByText("Vocabulary help", { exact: true }).click();
  const meaningId = "329c4ec9-6562-568e-9e71-9134da2f0d8d";
  const menu = main(page).getByRole("link", { name: "menu", exact: true });
  await expect(menu).toHaveAttribute(
    "href",
    `/vocabulary/menu#meaning-${meaningId}`,
  );
  await menu.click();
  await expect(page).toHaveURL(
    new RegExp(`/vocabulary/menu#meaning-${meaningId}$`),
  );
  await expect(main(page).locator(`[id="meaning-${meaningId}"]`)).toBeVisible();
  expect(
    await page.evaluate(
      () => document.documentElement.scrollWidth <= innerWidth,
    ),
  ).toBe(true);
});
test("wrong correction and correct retry, reload, completion and independent repeat", async ({
  page,
}) => {
  let session = await start(page);
  expect(session.visibleLines).toHaveLength(1);
  for (let i = 0; i < 4; i++) session = await act(page);
  expect(session.currentStep?.kind).toBe("comprehension");
  await main(page)
    .getByRole("radio", { name: "The restaurant is closing now.", exact: true })
    .check();
  session = await act(page, "Check answer");
  expect(session.feedback?.correct).toBe(false);
  await expect(
    main(page).getByRole("button", { name: "Continue", exact: true }),
  ).toHaveCount(0);
  await page.reload();
  await expect(
    main(page).getByText("Let's try that again.", { exact: true }),
  ).toBeVisible();
  await main(page)
    .getByRole("radio", { name: "She has a meeting at two.", exact: true })
    .check();
  session = await act(page, "Check again");
  expect(session.firstAnswersCorrect).toBe(0);
  await act(page);
  for (let i = 0; i < 4; i++) session = await act(page);
  await main(page).getByRole("radio", { name: "bill", exact: true }).check();
  await act(page, "Check answer");
  session = await act(page, "Finish story");
  expect(session).toMatchObject({
    status: "completed",
    firstAnswersCorrect: 1,
    questionsAnswered: 2,
    completedSteps: 10,
  });
  await expect(
    main(page).getByRole("heading", { name: "Story complete", exact: true }),
  ).toBeFocused();
  await expect(
    main(page).getByText("1 of 2 checks correct on your first try.", {
      exact: true,
    }),
  ).toBeVisible();
  await page.reload();
  await expect(
    main(page).getByRole("heading", { name: "Story complete", exact: true }),
  ).toBeVisible();
  const response = page.waitForResponse(
    (r) =>
      r.request().method() === "POST" &&
      new URL(r.url()).pathname === "/api/v1/story-sessions",
  );
  await main(page)
    .getByRole("button", { name: "Practise again", exact: true })
    .click();
  const repeat = (await (await response).json()) as StorySession;
  expect(repeat.id).not.toBe(session.id);
  expect(repeat.revision).toBe(0);
  const history = await page.request.get(
    `${apiURL}/api/v1/story-sessions/${session.id}`,
  );
  expect(await history.json()).toMatchObject({
    status: "completed",
    firstAnswersCorrect: 1,
  });
});
test("lost action response retries the exact request and resumes from library", async ({
  page,
}) => {
  const initial = await start(page);
  const requests: string[] = [];
  let lost = false;
  await page.route("**/api/v1/story-sessions/*/actions", async (route) => {
    requests.push(route.request().postData()!);
    if (!lost) {
      lost = true;
      await route.fetch();
      await route.abort();
    } else await route.continue();
  });
  await main(page)
    .getByRole("button", { name: "Continue", exact: true })
    .click();
  await expect(
    main(page).getByRole("button", { name: "Retry safely", exact: true }),
  ).toBeVisible();
  const next = await act(page, "Retry safely");
  expect(requests).toHaveLength(2);
  expect(requests[0]).toBe(requests[1]);
  expect(next.revision).toBe(1);
  await main(page)
    .getByRole("link", { name: "Back to Stories", exact: true })
    .click();
  const resume = main(page).getByRole("link", {
    name: "Continue story",
    exact: true,
  });
  await expect(resume).toHaveAttribute(
    "href",
    `/stories/session/${initial.id}`,
  );
  await resume.click();
  await expect(
    main(page).getByText("1 of 10 steps completed", { exact: true }),
  ).toBeVisible();
});
test("conflicting progress offers reload and keyboard controls in both themes", async ({
  page,
  context,
  baseURL,
}, testInfo) => {
  await start(page);
  await page.route("**/api/v1/story-sessions/*/actions", async (route) => {
    await route.fulfill({
      status: 409,
      contentType: "application/json",
      body: JSON.stringify({ detail: "Story changed" }),
    });
  });
  await main(page)
    .getByRole("button", { name: "Continue", exact: true })
    .click();
  await expect(
    main(page).getByRole("button", {
      name: "Load saved progress",
      exact: true,
    }),
  ).toBeVisible();
  await page.unroute("**/api/v1/story-sessions/*/actions");
  await main(page)
    .getByRole("button", { name: "Load saved progress", exact: true })
    .click();
  await expect(
    main(page).getByText("Your saved story is up to date.", { exact: true }),
  ).toBeVisible();
  for (const theme of ["light", "dark"]) {
    await context.addCookies([
      { name: "vocanova_theme", value: theme, url: baseURL! },
    ]);
    await page.emulateMedia({ colorScheme: theme as "light" | "dark" });
    await page.reload();
    const button = main(page).getByRole("button", {
      name: "Continue",
      exact: true,
    });
    await button.focus();
    await expect(button).toBeFocused();
    expect((await button.boundingBox())!.height).toBeGreaterThanOrEqual(44);
    expect(
      await page.evaluate(
        () => document.documentElement.scrollWidth <= innerWidth,
      ),
    ).toBe(true);
    expect((await scanForAxeViolations(page)).criticalOrSerious).toEqual([]);
    // Capture from the top so the sticky header keeps its normal position.
    await page.evaluate(() => window.scrollTo(0, 0));
    await page.screenshot({
      path: testInfo.outputPath(`story-${theme}.png`),
      fullPage: true,
    });
  }
  await main(page)
    .getByRole("button", { name: "Continue", exact: true })
    .press("Enter");
  await expect(
    main(page).getByText("1 of 10 steps completed", { exact: true }),
  ).toBeVisible();
});
test("empty and unavailable stories remain recoverable", async ({
  page,
  context,
  baseURL,
}) => {
  await context.addCookies([
    { name: "e2e_stories", value: "empty", url: baseURL! },
  ]);
  await page.goto("/stories");
  await expect(
    main(page).getByText("No stories are available right now.", {
      exact: true,
    }),
  ).toBeVisible();
  await context.addCookies([
    { name: "e2e_stories", value: "unavailable", url: baseURL! },
  ]);
  await page.reload();
  await expect(
    main(page).getByRole("heading", {
      name: "Your story could not load",
      exact: true,
    }),
  ).toBeVisible();
  await expect(
    main(page).getByRole("button", { name: "Try again", exact: true }),
  ).toBeVisible();
});
test("signed-out story screens redirect to sign-in", async ({
  page,
  context,
}) => {
  await context.clearCookies();
  for (const path of [
    "/stories",
    "/stories/a-quiet-lunch",
    "/stories/session/00000000-0000-4000-8000-000000000001",
  ]) {
    await page.goto(path);
    await expect(page).toHaveURL(/\/login\?returnTo=/);
    expect(new URL(page.url()).searchParams.get("returnTo")).toBe(path);
  }
});
