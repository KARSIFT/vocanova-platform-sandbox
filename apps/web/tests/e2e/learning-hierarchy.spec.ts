import { randomUUID } from "node:crypto";
import { expect, test, type BrowserContext, type Page } from "@playwright/test";
import type { StorySession } from "@vocanova/api-client";
import { formatViolations, scanForAxeViolations } from "./axe-helper";

const apiURL = `http://127.0.0.1:${process.env.MOCK_API_PORT ?? 8080}`;

async function headers(context: BrowserContext) {
  const csrf = (await context.cookies()).find((cookie) => cookie.name === "vocanova_csrf")?.value;
  if (!csrf) throw new Error("Missing fixture CSRF cookie");
  return { "X-CSRF-Token": csrf, "Idempotency-Key": randomUUID() };
}

async function createList(context: BrowserContext) {
  const id = randomUUID();
  const created = await context.request.put(`${apiURL}/api/v1/word-lists/${id}`, {
    headers: await headers(context),
    data: { name: "Words for conversations", expectedRevision: 0 },
  });
  expect(created.status()).toBe(200);
  const list = await created.json();
  const word = await (await context.request.get(`${apiURL}/api/v1/canonical-words/invite`)).json();
  const added = await context.request.put(`${apiURL}/api/v1/word-lists/${id}/members/${word.word.meanings[0].id}`, {
    headers: await headers(context),
    data: { expectedRevision: list.revision },
  });
  expect(added.status()).toBe(200);
  return id;
}

async function accessibleWithoutOverflow(page: Page) {
  await page.evaluate(() => window.scrollTo(0, 0));
  expect(await page.evaluate(() => document.documentElement.scrollWidth <= window.innerWidth)).toBe(true);
  const scan = await scanForAxeViolations(page);
  expect(scan.criticalOrSerious, formatViolations(scan.criticalOrSerious).join("\n")).toEqual([]);
}

test.beforeEach(async ({ context, baseURL, page }) => {
  if (!baseURL) throw new Error("Missing app URL");
  await context.addCookies([
    { name: "vocanova_session", value: randomUUID(), url: baseURL },
    { name: "vocanova_csrf", value: randomUUID(), url: baseURL },
  ]);
  await page.addInitScript(() => Object.defineProperty(window, "speechSynthesis", { value: undefined, configurable: true }));
});

for (const theme of ["light", "dark"] as const) {
  test.describe(`learning hierarchy (${theme})`, () => {
    test.beforeEach(async ({ context, baseURL, page }) => {
      await context.addCookies([{ name: "vocanova_theme", value: theme, url: baseURL! }]);
      await page.addInitScript((value) => localStorage.setItem("vocanova:theme-preference", value), theme);
    });

    test("list meanings lead, while settings remain available by keyboard", async ({ page, context }) => {
      const id = await createList(context);
      await page.goto(`/lists/${id}`);
      const main = page.getByRole("main");
      await expect(main.getByRole("link", { name: "Practice this list", exact: true })).toBeVisible();
      await expect(main.getByRole("list", { name: "List meanings", exact: true })).toBeVisible();
      const settings = main.locator("summary").filter({ hasText: /^List settings$/ });
      await expect(settings).toBeVisible();
      await expect(main.getByLabel("List name", { exact: true })).toBeHidden();
      expect(await main.getByRole("list", { name: "List meanings", exact: true }).evaluate((list) => Boolean(list.compareDocumentPosition(document.querySelector("#list-settings")!) & Node.DOCUMENT_POSITION_FOLLOWING))).toBe(true);
      await settings.focus();
      await page.keyboard.press("Enter");
      await expect(main.getByLabel("List name", { exact: true })).toBeVisible();
      await page.keyboard.press("Tab");
      await expect(main.getByLabel("List name", { exact: true })).toBeFocused();
      await settings.press("Space");
      await expect(main.getByLabel("List name", { exact: true })).toBeHidden();
      await page.evaluate(() => document.documentElement.style.fontSize = "200%");
      await accessibleWithoutOverflow(page);
    });

    test("existing lists lead before the optional create form", async ({ page, context }) => {
      await createList(context);
      await page.goto("/lists");
      const main = page.getByRole("main");
      const lists = main.getByRole("list", { name: "Personal lists", exact: true });
      await expect(lists).toBeVisible();
      await expect(main.getByLabel("List name", { exact: true })).toBeHidden();
      const create = main.locator("summary").filter({ hasText: /^Create a list$/ });
      await expect(create).toBeVisible();
      expect(await lists.evaluate((list) => Boolean(list.compareDocumentPosition(document.querySelector("#create-list")!) & Node.DOCUMENT_POSITION_FOLLOWING))).toBe(true);
      await create.press("Enter");
      await expect(main.getByLabel("List name", { exact: true })).toBeVisible();
      await accessibleWithoutOverflow(page);
    });

    test("a failed rename reopens settings and focuses recovery without changing the retry", async ({ page, context }) => {
      const id = await createList(context);
      await page.goto(`/lists/${id}`);
      const main = page.getByRole("main");
      const settings = main.locator("summary").filter({ hasText: /^List settings$/ });
      await settings.press("Enter");
      await main.getByLabel("List name", { exact: true }).fill("My revised plans");
      let release!: () => void;
      let pending!: () => void;
      const released = new Promise<void>((resolve) => { release = resolve; });
      const started = new Promise<void>((resolve) => { pending = resolve; });
      const writes: { key: string | undefined; body: string | null }[] = [];
      await page.route(`**/api/v1/word-lists/${id}`, async (route) => {
        if (route.request().method() !== "PUT") return route.continue();
        writes.push({ key: route.request().headers()["idempotency-key"], body: route.request().postData() });
        if (writes.length === 1) {
          pending();
          await released;
          return route.fulfill({ status: 503, contentType: "application/problem+json", body: JSON.stringify({ detail: "Temporary interruption" }) });
        }
        return route.continue();
      });
      await main.getByRole("button", { name: "Rename list", exact: true }).click();
      await started;
      await settings.press("Space");
      await expect(main.getByLabel("List name", { exact: true })).toBeHidden();
      release();
      const alert = main.getByRole("alert");
      await expect(alert).toBeVisible();
      await expect(alert).toBeFocused();
      await expect(alert).toBeInViewport({ ratio: 1 });
      await expect(main.getByLabel("List name", { exact: true })).toBeVisible();
      await main.getByRole("button", { name: "Retry same change", exact: true }).press("Enter");
      await expect(main.getByRole("heading", { name: "My revised plans", exact: true })).toBeVisible();
      expect(writes).toHaveLength(2);
      expect(writes[1]).toEqual(writes[0]);
    });

    test("an ongoing story offers continuation as its primary action", async ({ page, context }) => {
      const started = await context.request.post(`${apiURL}/api/v1/story-sessions`, {
        headers: await headers(context), data: { storyKey: "a-quiet-lunch" },
      });
      expect(started.status()).toBe(200);
      const session = await started.json();
      await page.goto("/stories");
      const card = page.getByRole("main").locator("section").filter({ has: page.getByRole("heading", { name: "A quiet lunch", exact: true }) });
      const resume = card.getByRole("link", { name: "Continue story", exact: true });
      await expect(resume).toHaveAttribute("href", `/stories/session/${session.id}`);
      await expect(resume).toHaveClass(/bg-primary-700/);
      await expect(card.getByRole("link", { name: "Read and practise", exact: true })).toHaveCount(0);
      await expect(card.getByRole("link", { name: "Story details", exact: true })).toHaveAttribute("href", "/stories/a-quiet-lunch");
      await accessibleWithoutOverflow(page);
      await resume.press("Enter");
      await expect(page).toHaveURL(new RegExp(`/stories/session/${session.id}$`));
    });

    test("new and completed stories offer one clear reading action", async ({ page, context }) => {
      await page.goto("/stories");
      const card = page.getByRole("main").locator("section").filter({ has: page.getByRole("heading", { name: "A quiet lunch", exact: true }) });
      await expect(card.getByRole("link", { name: "Read and practise", exact: true })).toHaveClass(/bg-primary-700/);
      const started = await context.request.post(`${apiURL}/api/v1/story-sessions`, {
        headers: await headers(context), data: { storyKey: "a-quiet-lunch" },
      });
      expect(started.status()).toBe(200);
      let session = await started.json() as StorySession;
      for (let step = 0; session.status !== "completed" && step < 20; step++) {
        const choiceId = session.canContinue
          ? undefined
          : session.feedback
            ? session.currentStep!.choices!.find((choice) => choice.text === session.feedback!.answer)!.id
            : session.currentStep!.choices![0]!.id;
        const response = await context.request.post(`${apiURL}/api/v1/story-sessions/${session.id}/actions`, {
          headers: await headers(context),
          data: { action: session.canContinue ? "continue" : "answer", stepId: session.currentStep!.id, expectedRevision: session.revision, clientActionId: randomUUID(), ...(choiceId ? { choiceId } : {}) },
        });
        expect(response.status()).toBe(200);
        session = await response.json();
      }
      expect(session.status).toBe("completed");
      await page.reload();
      await expect(card.getByRole("link", { name: "Continue story", exact: true })).toHaveCount(0);
      const repeat = card.getByRole("link", { name: "Read and practise again", exact: true });
      await expect(repeat).toHaveClass(/bg-primary-700/);
      await expect(repeat).toHaveAttribute("href", "/stories/a-quiet-lunch");
      await accessibleWithoutOverflow(page);
    });

    test("an empty writing catalog explains the state and offers useful alternatives", async ({ page, context, baseURL }) => {
      await context.addCookies([{ name: "e2e_journey_situations", value: "empty", url: baseURL! }]);
      await page.goto("/writing");
      const main = page.getByRole("main");
      await expect(main.getByRole("heading", { name: "No writing topics are available", exact: true })).toBeVisible();
      await expect(main.getByText("Choose a situation to begin", { exact: true })).toHaveCount(0);
      await expect(main.locator("summary").filter({ hasText: "Choose a writing topic" })).toHaveCount(0);
      await expect(main.getByRole("link", { name: "Practise with saved words", exact: true })).toHaveAttribute("href", "/words");
      await accessibleWithoutOverflow(page);
      await main.getByRole("link", { name: "Practise with saved words", exact: true }).press("Enter");
      await expect(page).toHaveURL(/\/words$/);
    });

    test("a writing catalog outage offers retry rather than claiming no topics exist", async ({ page, context, baseURL }) => {
      await context.addCookies([{ name: "e2e_journey_situations", value: "unavailable", url: baseURL! }]);
      await page.goto("/writing");
      const main = page.getByRole("main");
      await expect(main.getByRole("heading", { name: "Writing practice could not load", exact: true })).toBeVisible();
      await expect(main.getByRole("heading", { name: "No writing topics are available", exact: true })).toHaveCount(0);
      await expect(main.getByRole("button", { name: "Try again", exact: true })).toBeVisible();
      await context.clearCookies({ name: "e2e_journey_situations" });
      await main.getByRole("button", { name: "Try again", exact: true }).press("Enter");
      await expect(main.getByRole("heading", { name: "Topic writing", exact: true })).toBeVisible();
      await expect(main.locator("summary").filter({ hasText: "Choose a writing topic" })).toBeVisible();
    });
  });
}
