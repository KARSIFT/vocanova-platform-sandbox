import { randomUUID } from "node:crypto";
import { expect, test, type Page } from "@playwright/test";
import { formatViolations, scanForAxeViolations } from "./axe-helper";

const apiURL = `http://127.0.0.1:${process.env.MOCK_API_PORT ?? 8080}`;
const form = (page: Page) =>
  page
    .getByRole("main")
    .getByRole("form", { name: "Learning direction", exact: true });

test.beforeEach(async ({ context, baseURL }) => {
  if (!baseURL) throw new Error("Missing app URL");
  await context.addCookies([
    { name: "vocanova_session", value: randomUUID(), url: baseURL },
    { name: "vocanova_csrf", value: randomUUID(), url: baseURL },
  ]);
});

async function openDirection(page: Page) {
  await page.goto("/plan");
  await page
    .getByRole("main")
    .getByRole("button", { name: "Change learning direction", exact: true })
    .click();
  await expect(form(page)).toBeVisible();
}

async function choose(page: Page, goal: string, focus: string) {
  await form(page)
    .getByRole("combobox", { name: "Learning goal", exact: true })
    .selectOption(goal);
  await form(page)
    .getByRole("combobox", { name: "Main focus", exact: true })
    .selectOption(focus);
}

test("a failed starting-word search stays unavailable until a successful reload", async ({
  page,
  context,
  baseURL,
}) => {
  await context.addCookies([
    { name: "e2e_vocabulary", value: "unavailable", url: baseURL! },
  ]);
  await page.goto("/vocabulary/check");
  const main = page.getByRole("main");
  await expect(
    main.getByRole("heading", {
      name: "Your starting words are unavailable",
      exact: true,
    }),
  ).toBeVisible();
  await expect(
    main.getByRole("heading", {
      name: "You have explored these words",
      exact: true,
    }),
  ).toHaveCount(0);
  await expect(
    main.getByRole("button", { name: "Already know", exact: true }),
  ).toHaveCount(0);
  await expect(
    main.getByRole("button", { name: "Want to learn", exact: true }),
  ).toHaveCount(0);
  await context.clearCookies({ name: "e2e_vocabulary" });
  await main
    .getByRole("button", { name: "Reload the check", exact: true })
    .click();
  await expect(
    main.getByRole("button", { name: "Already know", exact: true }),
  ).toBeVisible();
  await expect(
    main.getByRole("heading", {
      name: "Your starting words are unavailable",
      exact: true,
    }),
  ).toHaveCount(0);
});

for (const theme of ["light", "dark"] as const) {
  test(`saved learning direction guides starting words without changing original onboarding (${theme})`, async ({
    page,
    context,
    baseURL,
  }, testInfo) => {
    await context.addCookies([
      { name: "vocanova_theme", value: theme, url: baseURL! },
    ]);
    const original = await (
      await context.request.get(`${apiURL}/api/v1/onboarding`)
    ).json();
    // This mock emits a fresh completedAt on each GET. Exact preservation of
    // the real historical timestamp is covered by the PostgreSQL regression.
    delete original.completedAt;
    const settings = await (
      await context.request.get(`${apiURL}/api/v1/settings`)
    ).json();
    await openDirection(page);
    await choose(page, "conversation", "social");
    const save = form(page).getByRole("button", {
      name: "Save learning direction",
      exact: true,
    });
    await form(page).getByRole("combobox", { name: "Main focus", exact: true }).focus();
    await page.keyboard.press("Tab");
    await expect(save).toBeFocused();
    await page.keyboard.press("Enter");
    await expect(form(page).getByRole("status")).toHaveText(
      "Your learning direction is saved.",
    );
    await expect(save).toBeDisabled();
    expect(
      await (
        await context.request.get(`${apiURL}/api/v1/learning-preferences`)
      ).json(),
    ).toEqual({
      learningGoal: "conversation",
      mainUseCase: "social",
      revision: 1,
    });
    const originalAfter = await (
      await context.request.get(`${apiURL}/api/v1/onboarding`)
    ).json();
    delete originalAfter.completedAt;
    expect(originalAfter).toEqual(original);
    expect(
      await (await context.request.get(`${apiURL}/api/v1/settings`)).json(),
    ).toEqual(settings);
    await expect(page.locator("html")).toHaveAttribute("data-theme", theme);
    const scan = await scanForAxeViolations(page);
    expect(
      scan.criticalOrSerious,
      formatViolations(scan.criticalOrSerious).join("\n"),
    ).toEqual([]);
    expect(
      await page.evaluate(() => document.documentElement.scrollWidth),
    ).toBeLessThanOrEqual(page.viewportSize()!.width);
    await form(page).screenshot({
      path: testInfo.outputPath(`learning-direction-${theme}.png`),
    });
    await page.reload();
    await page
      .getByRole("main")
      .getByRole("button", { name: "Change learning direction", exact: true })
      .click();
    await expect(
      form(page).getByRole("combobox", { name: "Learning goal", exact: true }),
    ).toHaveValue("conversation");
    await expect(
      form(page).getByRole("combobox", { name: "Main focus", exact: true }),
    ).toHaveValue("social");
    const expected = await (
      await context.request.get(
        `${apiURL}/api/v1/canonical-words?knowledge=unexplored&category=social&limit=10`,
      )
    ).json();
    expect(expected.items.length).toBeGreaterThan(0);
    await page.goto("/vocabulary/check");
    await expect(
      page
        .getByRole("main")
        .getByText("Starting with your focus: social.", { exact: true }),
    ).toBeVisible();
    await expect(
      page.getByRole("main").getByRole("heading", { level: 2 }),
    ).toHaveText(expected.items[0].wordText);
  });
}

test("an unknown grandfathered direction starts with no invented choices", async ({
  page,
  context,
  baseURL,
}) => {
  await context.addCookies([
    { name: "e2e_learning_preferences", value: "unset", url: baseURL! },
  ]);
  await page.goto("/plan");
  await expect(form(page)).toBeVisible();
  await expect(
    form(page).getByRole("combobox", { name: "Learning goal", exact: true }),
  ).toHaveValue("");
  await expect(
    form(page).getByRole("combobox", { name: "Main focus", exact: true }),
  ).toHaveValue("");
  const save = form(page).getByRole("button", {
    name: "Save learning direction",
    exact: true,
  });
  await expect(save).toBeDisabled();
  expect(
    await (
      await context.request.get(`${apiURL}/api/v1/learning-preferences`)
    ).json(),
  ).toEqual({ learningGoal: null, mainUseCase: null, revision: 0 });
  await form(page)
    .getByRole("combobox", { name: "Learning goal", exact: true })
    .selectOption("work");
  await expect(save).toBeDisabled();
  await form(page)
    .getByRole("combobox", { name: "Main focus", exact: true })
    .selectOption("work");
  await save.click();
  await expect(form(page).getByRole("status")).toHaveText(
    "Your learning direction is saved.",
  );
  expect(
    await (
      await context.request.get(`${apiURL}/api/v1/learning-preferences`)
    ).json(),
  ).toEqual({ learningGoal: "work", mainUseCase: "work", revision: 1 });
});

test("a lost direction response retries the identical intent and revision", async ({
  page,
  context,
}) => {
  await openDirection(page);
  const bodies: string[] = [];
  await page.route("**/api/v1/learning-preferences", async (route) => {
    if (route.request().method() !== "PATCH") return route.continue();
    bodies.push(route.request().postData()!);
    if (bodies.length === 1) {
      const response = await route.fetch();
      expect(response.status()).toBe(200);
      expect(await response.json()).toEqual({
        learningGoal: "conversation",
        mainUseCase: "social",
        revision: 1,
      });
      return route.abort("failed");
    }
    return route.continue();
  });
  await choose(page, "conversation", "social");
  await form(page)
    .getByRole("button", { name: "Save learning direction", exact: true })
    .click();
  await expect(form(page).getByRole("alert")).toBeVisible();
  await expect(
    form(page).getByRole("combobox", { name: "Learning goal", exact: true }),
  ).toBeDisabled();
  await expect(
    form(page).getByRole("combobox", { name: "Main focus", exact: true }),
  ).toBeDisabled();
  await form(page)
    .getByRole("button", { name: "Retry saving direction", exact: true })
    .click();
  await expect(form(page).getByRole("status")).toHaveText(
    "Your learning direction is saved.",
  );
  expect(bodies).toHaveLength(2);
  expect(bodies[1]).toBe(bodies[0]);
  expect(JSON.parse(bodies[0]!)).toEqual({
    learningGoal: "conversation",
    mainUseCase: "social",
    expectedRevision: 0,
  });
  expect(
    await (
      await context.request.get(`${apiURL}/api/v1/learning-preferences`)
    ).json(),
  ).toEqual({
    learningGoal: "conversation",
    mainUseCase: "social",
    revision: 1,
  });
  await choose(page, "work", "work");
  await form(page)
    .getByRole("button", { name: "Save learning direction", exact: true })
    .click();
  await expect(form(page).getByRole("status")).toHaveText(
    "Your learning direction is saved.",
  );
  expect(JSON.parse(bodies[2]!)).toEqual({
    learningGoal: "work",
    mainUseCase: "work",
    expectedRevision: 1,
  });
});

test("a competing direction requires loading current choices before a deliberate new save", async ({
  page,
  context,
}) => {
  await openDirection(page);
  await choose(page, "conversation", "social");
  const csrf = (await context.cookies()).find(
    (cookie) => cookie.name === "vocanova_csrf",
  )!.value;
  const competing = await context.request.patch(
    `${apiURL}/api/v1/learning-preferences`,
    {
      headers: { "X-CSRF-Token": csrf },
      data: {
        learningGoal: "travel",
        mainUseCase: "travel",
        expectedRevision: 0,
      },
    },
  );
  expect(competing.status()).toBe(200);
  const bodies: unknown[] = [];
  page.on("request", (request) => {
    if (
      new URL(request.url()).pathname === "/api/v1/learning-preferences" &&
      request.method() === "PATCH"
    )
      bodies.push(request.postDataJSON());
  });
  await form(page)
    .getByRole("button", { name: "Save learning direction", exact: true })
    .click();
  await expect(form(page).getByRole("alert")).toHaveText(
    "Your choices changed elsewhere. Load your current choices before saving again.",
  );
  await expect(
    form(page).getByRole("button", {
      name: "Retry saving direction",
      exact: true,
    }),
  ).toHaveCount(0);
  await expect(
    form(page).getByRole("combobox", { name: "Main focus", exact: true }),
  ).toBeDisabled();
  await form(page)
    .getByRole("button", { name: "Load current choices", exact: true })
    .click();
  await expect(form(page).getByRole("status")).toHaveText(
    "Current choices loaded. Review them before saving.",
  );
  await expect(
    form(page).getByRole("combobox", { name: "Learning goal", exact: true }),
  ).toHaveValue("travel");
  await expect(
    form(page).getByRole("combobox", { name: "Main focus", exact: true }),
  ).toHaveValue("travel");
  await expect(
    form(page).getByRole("button", {
      name: "Save learning direction",
      exact: true,
    }),
  ).toBeDisabled();
  expect(bodies).toHaveLength(1);
  await choose(page, "conversation", "social");
  await form(page)
    .getByRole("button", { name: "Save learning direction", exact: true })
    .click();
  await expect(form(page).getByRole("status")).toHaveText(
    "Your learning direction is saved.",
  );
  expect(bodies).toEqual([
    {
      learningGoal: "conversation",
      mainUseCase: "social",
      expectedRevision: 0,
    },
    {
      learningGoal: "conversation",
      mainUseCase: "social",
      expectedRevision: 1,
    },
  ]);
  expect(
    await (
      await context.request.get(`${apiURL}/api/v1/learning-preferences`)
    ).json(),
  ).toEqual({
    learningGoal: "conversation",
    mainUseCase: "social",
    revision: 2,
  });
});

test("CSRF preflight failure retains the intended direction for retry", async ({
  page,
  context,
  baseURL,
}) => {
  await openDirection(page);
  await choose(page, "conversation", "social");
  await context.clearCookies({ name: "vocanova_csrf" });
  await page.route("**/api/v1/me", (route) =>
    route.fulfill({
      status: 503,
      contentType: "application/json",
      body: JSON.stringify({ detail: "Synthetic preflight failure" }),
    }),
  );
  const bodies: unknown[] = [];
  page.on("request", (request) => {
    if (
      new URL(request.url()).pathname === "/api/v1/learning-preferences" &&
      request.method() === "PATCH"
    )
      bodies.push(request.postDataJSON());
  });
  await form(page)
    .getByRole("button", { name: "Save learning direction", exact: true })
    .click();
  await expect(form(page).getByRole("alert")).toBeVisible();
  expect(bodies).toEqual([]);
  await expect(
    form(page).getByRole("combobox", { name: "Learning goal", exact: true }),
  ).toHaveValue("conversation");
  await expect(
    form(page).getByRole("combobox", { name: "Main focus", exact: true }),
  ).toBeDisabled();
  await page.unroute("**/api/v1/me");
  await context.addCookies([
    { name: "vocanova_csrf", value: randomUUID(), url: baseURL! },
  ]);
  await form(page)
    .getByRole("button", { name: "Retry saving direction", exact: true })
    .click();
  await expect(form(page).getByRole("status")).toHaveText(
    "Your learning direction is saved.",
  );
  expect(bodies).toEqual([
    {
      learningGoal: "conversation",
      mainUseCase: "social",
      expectedRevision: 0,
    },
  ]);
});
