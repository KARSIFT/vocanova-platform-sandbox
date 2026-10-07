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

test("an unknown lesson query falls back to an explicit full-course mix", async ({
  page,
}) => {
  await page.goto("/practice?lesson=not-a-real-lesson");
  const main = page.getByRole("main");
  await expect(
    main.getByLabel("Practice vocabulary", { exact: true }),
  ).toHaveValue("");
  await main.locator("summary").filter({ hasText: "About this selection" }).focus();
  await page.keyboard.press("Enter");
  await expect(
    main.getByText(
      "A full-course mix can include words you have not studied yet.",
      { exact: false },
    ),
  ).toBeVisible();
  const started = page.waitForRequest(
    (request) =>
      new URL(request.url()).pathname === "/api/v1/practice-sessions" &&
      request.method() === "POST",
  );
  await main
    .getByRole("button", { name: "Start typed recall", exact: true })
    .click();
  expect((await started).postDataJSON()).toEqual({ mode: "typed_recall" });
  await expect(page).toHaveURL(/\/practice\/session\/[^/]+$/);
});
