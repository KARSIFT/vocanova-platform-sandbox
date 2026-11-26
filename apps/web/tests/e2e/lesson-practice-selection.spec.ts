import { randomUUID } from "node:crypto";
import { expect, test } from "@playwright/test";

const apiURL = `http://127.0.0.1:${process.env.MOCK_API_PORT ?? 8080}`;

test.beforeEach(async ({ context, baseURL }) => {
  if (!baseURL) throw new Error("Missing app URL");
  await context.addCookies([
    { name: "vocanova_session", value: randomUUID(), url: baseURL },
    { name: "vocanova_csrf", value: randomUUID(), url: baseURL },
  ]);
});

for (const [mode, button] of [
  ["typed_recall", "Start typed recall"],
  ["listening_choice", "Start listening practice"],
] as const) {
  test(`chosen lesson is retained on a lost ${mode} start response`, async ({
    page,
    context,
  }) => {
    await page.goto("/practice?lesson=conversation-basics");
    const main = page.getByRole("main");
    const selection = main.getByLabel("Practice vocabulary", { exact: true });
    await expect(selection).toHaveValue("conversation-basics");
    await selection.selectOption("");
    await selection.selectOption("conversation-basics");
    const requests: { body: unknown; key: string | undefined }[] = [];
    let createdID = "";
    await page.route("**/api/v1/practice-sessions", async (route) => {
      if (route.request().method() !== "POST") return route.continue();
      requests.push({
        body: route.request().postDataJSON(),
        key: route.request().headers()["idempotency-key"],
      });
      if (requests.length === 1) {
        const response = await route.fetch();
        expect(response.status()).toBe(200);
        const session = await response.json();
        expect(session).toMatchObject({
          mode,
          lessonKey: "conversation-basics",
        });
        createdID = session.id;
        return route.abort("failed");
      }
      return route.continue();
    });
    await main.getByRole("button", { name: button, exact: true }).click();
    await expect(main.getByRole("alert")).toBeVisible();
    await expect(selection).toBeDisabled();
    await expect(selection).toHaveValue("conversation-basics");
    await main
      .getByRole("button", { name: "Retry start", exact: true })
      .click();
    await expect(page).toHaveURL(new RegExp(`/practice/session/${createdID}$`));
    expect(requests).toHaveLength(2);
    expect(requests[0]!.key).toBeTruthy();
    expect(requests[1]).toEqual(requests[0]);
    expect(requests[0]!.body).toEqual({
      mode,
      lessonKey: "conversation-basics",
    });
    const list = await (
      await context.request.get(`${apiURL}/api/v1/practice-sessions`)
    ).json();
    expect(list.items).toHaveLength(1);
    expect(list.items[0]).toMatchObject({
      id: createdID,
      mode,
      lessonKey: "conversation-basics",
    });
  });
}

for (const [kind, lessonKey] of [
  ["an unknown", "not-a-real-lesson"],
  ["a blank", ""],
] as const) {
  test(`${kind} lesson needs a new vocabulary choice before focused practice`, async ({
    page,
  }) => {
    await page.goto(`/practice?mode=typed_recall&lesson=${encodeURIComponent(lessonKey)}`);
    const main = page.getByRole("main");
    const selection = main.getByRole("combobox", { name: "Practice vocabulary" });
    const typed = main.getByRole("button", {
      name: "Start typed recall",
      exact: true,
    });
    await expect(
      main.getByRole("option", { name: "Selected lesson is unavailable" }),
    ).toBeAttached();
    await expect(selection).not.toHaveValue("");
    await expect(typed).toBeDisabled();
    await expect(
      main.getByRole("button", { name: "Start listening practice", exact: true }),
    ).toBeDisabled();
    await expect(
      main.getByRole("button", { name: "Start practice lesson", exact: true }),
    ).toBeEnabled();

    await selection.selectOption({ label: "Mix from the full course" });
    const started = page.waitForRequest(
      (request) =>
        new URL(request.url()).pathname === "/api/v1/practice-sessions" &&
        request.method() === "POST",
    );
    await typed.click();
    expect((await started).postDataJSON()).toEqual({ mode: "typed_recall" });
    await expect(page).toHaveURL(/\/practice\/session\/[^/]+$/);
  });
}

test("focused listening retains its lesson through a catalog outage", async ({
  page,
  context,
  baseURL,
}) => {
  if (!baseURL) throw new Error("Missing app URL");
  await context.addCookies([
    { name: "e2e_lessons", value: "unavailable", url: baseURL },
  ]);
  const starts: unknown[] = [];
  page.on("request", (request) => {
    if (
      request.method() === "POST" &&
      new URL(request.url()).pathname === "/api/v1/practice-sessions"
    )
      starts.push(request.postDataJSON());
  });
  await page.goto(
    "/practice?mode=listening_choice&lesson=conversation-basics#practice-mode-listening_choice",
  );
  const main = page.getByRole("main");
  const selection = main.getByRole("combobox", { name: "Practice vocabulary" });
  const listening = main.getByRole("button", {
    name: "Start listening practice",
    exact: true,
  });
  await expect(
    main.getByRole("status").filter({
      hasText: /situation lesson choices are unavailable/i,
    }),
  ).toBeVisible();
  await expect(
    main.getByRole("option", { name: "Selected lesson is unavailable" }),
  ).toBeAttached();
  await expect(selection).not.toHaveValue("");
  await expect(listening).toBeDisabled();

  await context.clearCookies({ name: "e2e_lessons" });
  const refresh = main.getByRole("button", { name: "Refresh practice sources" });
  await refresh.focus();
  await expect(refresh).toBeFocused();
  await page.keyboard.press("Enter");
  await expect(selection).toHaveValue("conversation-basics");
  await expect(listening).toBeEnabled();
  expect(starts).toEqual([]);

  const started = page.waitForResponse(
    (response) =>
      response.request().method() === "POST" &&
      new URL(response.url()).pathname === "/api/v1/practice-sessions",
  );
  await listening.click();
  const response = await started;
  expect(response.status()).toBe(200);
  const session = await response.json();
  expect(starts).toEqual([
    { mode: "listening_choice", lessonKey: "conversation-basics" },
  ]);
  await expect(page).toHaveURL(`/practice/session/${session.id}`);
});

test("a focused mode link keeps its situation and resumes the same saved session", async ({
  page,
  context,
}) => {
  const url =
    "/practice?mode=typed_recall&lesson=conversation-basics#practice-mode-typed_recall";
  await page.goto(url);
  const main = page.getByRole("main");
  const focused = main.locator("#practice-mode-typed_recall");
  await expect(
    main.getByRole("link", { name: "Open situation lesson", exact: true }),
  ).toHaveAttribute("href", "/learn/conversation-basics");
  await expect(
    focused.getByRole("button", { name: "Start typed recall" }),
  ).toBeVisible();
  const before = await (
    await context.request.get(`${apiURL}/api/v1/practice-sessions`)
  ).json();
  expect(before.items).toHaveLength(0);
  const response = page.waitForResponse(
    (r) =>
      r.request().method() === "POST" &&
      new URL(r.url()).pathname === "/api/v1/practice-sessions",
  );
  await focused.getByRole("button").click();
  const saved = await (await response).json();
  expect(saved).toMatchObject({
    mode: "typed_recall",
    lessonKey: "conversation-basics",
  });
  await expect(page).toHaveURL(`/practice/session/${saved.id}`);
  await page.goto(url);
  const resume = focused.getByRole("link", { name: /Resume typed recall/ });
  await expect(resume).toHaveAttribute("href", `/practice/session/${saved.id}`);
  await resume.focus();
  await resume.press("Enter");
  await expect(page).toHaveURL(`/practice/session/${saved.id}`);
  const after = await (
    await context.request.get(`${apiURL}/api/v1/practice-sessions`)
  ).json();
  expect(after.items).toHaveLength(1);
});

test("a learner arriving from Review sees the empty mistake state", async ({
  page,
}) => {
  await page.goto("/practice?mode=mistakes#practice-mode-mistakes");
  const modes = page.locator("#practice-mode-mistakes");
  await expect(modes).toContainText("No past mistakes to revisit right now.");
  await expect(modes.getByRole("button")).toHaveCount(0);
  await expect(modes.locator("..").locator("li").first()).toHaveAttribute(
    "id",
    "practice-mode-mistakes",
  );
});

test("situation practice leads to the lesson in the server catalog", async ({
  page,
}) => {
  await page.goto("/practice");
  const scenarios = page.getByRole("region", { name: "Practise a situation" });
  const lesson = scenarios.getByRole("link", {
    name: /Make a plan with a friend/,
  });
  await expect(lesson).toContainText("3 words · 9 steps");
  await lesson.click();
  await expect(page).toHaveURL("/learn/conversation-basics");
});

test("the practice hub links each available situation to its guided lesson", async ({
  context,
  baseURL,
  page,
}) => {
  await context.addCookies([
    { name: "e2e_lessons", value: "scenarios", url: baseURL! },
  ]);
  await page.goto("/practice");
  const situations = page.getByRole("region", { name: "Practise a situation" });
  for (const [topic, key] of [
    ["Work", "work-meeting"],
    ["Travel", "airport"],
    ["Dining", "restaurant-other-plans"],
    ["Shopping", "shopping-returns"],
    ["Everyday problems", "delivery-parcel"],
  ]) {
    const link = situations.getByRole("link", { name: new RegExp(`^${topic} `) });
    await expect(link).toHaveAttribute("href", `/learn/${key}`);
  }
  const travel = situations.getByRole("link", { name: /^Travel / });
  await travel.focus();
  await travel.press("Enter");
  await expect(page).toHaveURL("/learn/airport");
});
