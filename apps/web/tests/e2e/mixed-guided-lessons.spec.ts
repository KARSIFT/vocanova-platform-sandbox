import { randomUUID } from "node:crypto";
import { test, expect } from "@playwright/test";
import { scanForAxeViolations } from "./axe-helper";

test.beforeEach(async ({ context, baseURL }) => {
  await context.addCookies([
    { name: "vocanova_session", value: randomUUID(), url: baseURL! },
    { name: "vocanova_csrf", value: randomUUID(), url: baseURL! },
    { name: "e2e_lesson_format", value: "mixed", url: baseURL! },
  ]);
});
for (const theme of ["light", "dark"])
  test(`mixed lesson typed recall, audio fallback and resume in ${theme}`, async ({
    page,
    context,
    baseURL,
  }) => {
    await context.addCookies([
      { name: "vocanova_theme", value: theme, url: baseURL! },
    ]);
    // No audio availability: text help must permit progress without automatic playback.
    await page.addInitScript(() =>
      Object.defineProperty(window, "speechSynthesis", {
        value: undefined,
        configurable: true,
      }),
    );
    await page.goto("/learn/conversation-basics");
    const start = page.waitForResponse(
      (r) => r.request().method() === "POST" && r.url().endsWith("/sessions"),
    );
    await page
      .getByRole("button", { name: "Start lesson", exact: true })
      .click();
    const session = await (await start).json();
    for (let i = 0; i < 3; i++) {
      await page.getByRole("button", { name: "Continue", exact: true }).click();
    }
    await page
      .getByRole("radio", { name: session.words[0].definition, exact: true })
      .check();
    await page
      .getByRole("button", { name: "Check answer", exact: true })
      .click();
    await page.getByRole("button", { name: "Continue", exact: true }).click();
    const input = page.getByRole("textbox", { name: "Your word or phrase" });
    await input.fill("invite");
    await page
      .getByRole("button", { name: "Check answer", exact: true })
      .click();
    await expect(
      page.getByRole("heading", { name: "Let’s look again", exact: true }),
    ).toBeVisible();
    await expect(
      page.getByRole("button", { name: "Continue", exact: true }),
    ).toHaveCount(0);
    await page.reload();
    await expect(
      page.getByRole("heading", { name: "Let’s look again", exact: true }),
    ).toBeVisible();
    await input.fill("  CONFIRM  ");
    await page
      .getByRole("button", { name: "Check answer", exact: true })
      .click();
    await expect(
      page.getByRole("status").filter({
        has: page.getByRole("heading", { name: "That’s right", exact: true }),
      }),
    ).toBeFocused();
    await page.getByRole("button", { name: "Continue", exact: true }).click();
    await expect(
      page.getByRole("button", { name: "Listen to lesson audio", exact: true }),
    ).toBeVisible();
    await page
      .getByRole("button", { name: "Listen to lesson audio", exact: true })
      .click();
    const help = page.locator("summary").filter({ hasText: "Show audio text" });
    await help.focus();
    await page.keyboard.press("Enter");
    await expect(page.getByText("reschedule", { exact: true })).toBeVisible();
    await expect(
      page.getByText(
        "This lesson records first answers, not unaided listening ability.",
        { exact: false },
      ),
    ).toBeVisible();
    expect(
      await page.evaluate(() => document.documentElement.scrollWidth),
    ).toBeLessThanOrEqual(page.viewportSize()!.width);
    expect((await scanForAxeViolations(page)).criticalOrSerious).toEqual([]);
    await page
      .getByRole("radio", { name: session.words[2].definition, exact: true })
      .check();
    await page
      .getByRole("button", { name: "Check answer", exact: true })
      .click();
    await page.getByRole("button", { name: "Continue", exact: true }).click();
    for (let i = 0; i < 3; i++) {
      await page
        .getByRole("radio", { name: session.words[i].wordText, exact: true })
        .check();
      await page
        .getByRole("button", { name: "Check answer", exact: true })
        .click();
      await page
        .getByRole("button", {
          name: i === 2 ? "Finish lesson" : "Continue",
          exact: true,
        })
        .click();
    }
    await expect(
      page.getByRole("heading", { name: "Lesson complete", exact: true }),
    ).toBeVisible();
    await expect(
      page.getByText("5 of 6 questions correct on your first try.", {
        exact: false,
      }),
    ).toBeVisible();
  });
test("typed lesson retry keeps its exact answer after response loss", async ({
  page,
}) => {
  await page.goto("/learn/conversation-basics");
  const start = page.waitForResponse(
    (r) => r.request().method() === "POST" && r.url().endsWith("/sessions"),
  );
  await page.getByRole("button", { name: "Start lesson", exact: true }).click();
  const session = await (await start).json();
  for (let i = 0; i < 3; i++)
    await page.getByRole("button", { name: "Continue", exact: true }).click();
  await page
    .getByRole("radio", { name: session.words[0].definition, exact: true })
    .check();
  await page.getByRole("button", { name: "Check answer", exact: true }).click();
  await page.getByRole("button", { name: "Continue", exact: true }).click();
  const intents: unknown[] = [];
  let lost = false;
  await page.route("**/api/v1/lesson-sessions/*/actions", async (route) => {
    const body = route.request().postDataJSON();
    if (body.typedAnswer) {
      intents.push([body, route.request().headers()["idempotency-key"]]);
      if (!lost) {
        lost = true;
        await route.fetch();
        await route.abort("failed");
        return;
      }
    }
    await route.continue();
  });
  await page
    .getByRole("textbox", { name: "Your word or phrase" })
    .fill("confirm");
  await page.getByRole("button", { name: "Check answer", exact: true }).click();
  await expect(
    page.getByRole("textbox", { name: "Your word or phrase" }),
  ).toBeDisabled();
  await page.getByRole("button", { name: "Retry step", exact: true }).click();
  await expect(
    page.getByRole("heading", { name: "That’s right", exact: true }),
  ).toBeVisible();
  expect(intents).toHaveLength(2);
  expect(intents[1]).toEqual(intents[0]);
});

test("a confirmed typed validation rejection keeps an editable draft", async ({
  page,
}) => {
  await page.goto("/learn/conversation-basics");
  const opening = page.waitForResponse(
    (r) => r.request().method() === "POST" && r.url().endsWith("/sessions"),
  );
  await page.getByRole("button", { name: "Start lesson", exact: true }).click();
  const session = await (await opening).json();
  for (let i = 0; i < 3; i++)
    await page.getByRole("button", { name: "Continue", exact: true }).click();
  await page
    .getByRole("radio", { name: session.words[0].definition, exact: true })
    .check();
  await page.getByRole("button", { name: "Check answer", exact: true }).click();
  await page.getByRole("button", { name: "Continue", exact: true }).click();
  const draft = page.getByRole("textbox", { name: "Your word or phrase" });
  await draft.fill("é".repeat(101));
  await page.route(
    "**/api/v1/lesson-sessions/*/actions",
    (route) =>
      route.fulfill({
        status: 422,
        contentType: "application/json",
        body: JSON.stringify({ detail: "Answer too long" }),
      }),
    { times: 1 },
  );
  await page.getByRole("button", { name: "Check answer", exact: true }).click();
  await expect(page.getByRole("main").getByRole("alert")).toContainText(
    "your draft is still here",
  );
  await expect(draft).toBeEnabled();
  await expect(draft).toHaveValue("é".repeat(101));
  await draft.fill("confirm");
  await page.getByRole("button", { name: "Check answer", exact: true }).click();
  await expect(
    page.getByRole("heading", { name: "That’s right", exact: true }),
  ).toBeVisible();
});
