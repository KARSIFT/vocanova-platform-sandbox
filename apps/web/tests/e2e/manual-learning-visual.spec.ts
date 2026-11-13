import { randomUUID } from "node:crypto";
import { expect, test, type Locator, type Page } from "@playwright/test";
import type { PracticeSession } from "@vocanova/api-client";

async function startPractice(page: Page, name: string) {
  await page.goto("/practice");
  if (name === "Start typed recall")
    await page.getByRole("main").getByText("More practice", { exact: true }).click();
  const response = page.waitForResponse((result) =>
    result.request().method() === "POST" &&
    new URL(result.url()).pathname === "/api/v1/practice-sessions",
  );
  await page.getByRole("button", { name, exact: true }).click();
  const result = await response;
  expect(result.status()).toBe(200);
  const session = await result.json() as PracticeSession;
  await expect(page).toHaveURL(new RegExp(`/practice/session/${session.id}$`));
  return session;
}

async function practiceAction(page: Page, name: string) {
  const response = page.waitForResponse((result) =>
    result.request().method() === "POST" &&
    /\/practice-sessions\/[^/]+\/actions$/u.test(new URL(result.url()).pathname),
  );
  await page.getByRole("button", { name, exact: true }).click();
  const result = await response;
  expect(result.status()).toBe(200);
  return await result.json() as PracticeSession;
}

async function expectHeadwordFont(heading: Locator) {
  await expect(heading).toBeVisible();
  expect(await heading.evaluate((element) => getComputedStyle(element).fontFamily))
    .toMatch(/Georgia/);
}

for (const theme of ["light", "dark"] as const) {
  test.describe(`manual learning visual regressions (${theme})`, () => {
    test.beforeEach(async ({ context, baseURL, page }) => {
      if (!baseURL) throw new Error("Missing app URL");
      await context.addCookies([
        { name: "vocanova_session", value: randomUUID(), url: baseURL },
        { name: "vocanova_csrf", value: randomUUID(), url: baseURL },
        { name: "vocanova_theme", value: theme, url: baseURL },
        { name: "e2e_teaching", value: "true", url: baseURL },
      ]);
      await page.addInitScript((value) => localStorage.setItem("vocanova:theme-preference", value), theme);
    });

    test("practice path actions stay within their row with enlarged text", async ({ page }) => {
      for (const name of ["Start practice lesson", "Continue practice"]) {
        await page.goto("/practice");
        await page.addStyleTag({ content: "html { font-size: 200% !important; }" });
        const action = page.getByRole(name === "Start practice lesson" ? "button" : "link", { name, exact: true });
        await expect(action).toBeVisible();
        const bounds = await action.evaluate((element) => {
          const button = element.getBoundingClientRect();
          const row = element.closest("li")!.getBoundingClientRect();
          return { left: button.left >= row.left, right: button.right <= row.right };
        });
        expect(bounds).toEqual({ left: true, right: true });
        expect(await page.evaluate(() => document.documentElement.scrollWidth <= innerWidth)).toBe(true);
        await action.focus();
        await expect(action).toBeFocused();
        if (name === "Start practice lesson") {
          await action.press("Enter");
          await expect(page).toHaveURL(/\/practice\/session\/[^/]+$/u);
        }
      }
    });

    test("typed recall separates its intact prompt from the focused task heading", async ({ page }) => {
      let session = await startPractice(page, "Start typed recall");
      const main = page.getByRole("main");
      const heading = main.getByRole("heading", { level: 1, name: "Remember the word", exact: true });
      await expect(heading).toBeVisible();
      const prompt = main.getByText(session.currentStep!.prompt, { exact: true });
      await expect(prompt).toBeVisible();
      expect(await prompt.evaluate((element) => element.tagName)).toBe("P");
      await expect(prompt).toHaveCSS("font-weight", "400");
      expect(await heading.evaluate((element) => getComputedStyle(element).fontFamily)).not.toMatch(/Georgia/);
      await expect(main.getByText("invite", { exact: true })).toHaveCount(0);
      await expect(main.getByTestId("meaning-picture")).toHaveCount(0);
      const input = main.getByRole("textbox", { name: "Your answer", exact: true });
      await input.focus();
      await expect(input).toBeFocused();
      await expect(input).toBeInViewport({ ratio: 1 });
      await input.fill("invite");
      session = await practiceAction(page, "Check answer");
      expect(session.feedback).toMatchObject({ correct: true, answer: "invite" });
      session = await practiceAction(page, "Continue");
      await expect(heading).toBeFocused();
      await expect(main.getByText(session.currentStep!.prompt, { exact: true })).toBeVisible();
      await page.keyboard.press("Tab");
      await expect(input).toBeFocused();
      await expect(input).toBeInViewport({ ratio: 1 });
      expect(await page.evaluate(() => document.documentElement.scrollWidth <= innerWidth)).toBe(true);
    });

    test("practice and lesson teaching words use the shared dictionary typeface", async ({ page }) => {
      const session = await startPractice(page, "Start practice lesson");
      expect(session.currentStep?.kind).toBe("teach");
      await expectHeadwordFont(page.getByRole("heading", { level: 1, name: session.currentStep!.wordText, exact: true }));
      await page.goto("/learn/conversation-basics");
      await page.getByRole("button", { name: "Start lesson", exact: true }).click();
      await expectHeadwordFont(page.getByRole("heading", { level: 1, name: "invite", exact: true }));
    });

    test("self-check words use dictionary type while completion keeps interface type", async ({ page }) => {
      await page.goto("/vocabulary/check");
      const main = page.getByRole("main");
      await expectHeadwordFont(main.getByRole("heading", { level: 2 }));
      expect(await main.getByRole("heading", { level: 1 }).evaluate((element) => getComputedStyle(element).fontFamily)).not.toMatch(/Georgia/);
      await main.getByRole("button", { name: "Skip for now", exact: true }).press("Enter");
      const complete = main.getByRole("heading", { level: 2, name: "Your starting words are ready", exact: true });
      await expect(complete).toBeFocused();
      expect(await complete.evaluate((element) => getComputedStyle(element).fontFamily)).not.toMatch(/Georgia/);
      await expect(main.getByRole("link", { name: "See your learning plan", exact: true })).toBeVisible();
      expect(await page.evaluate(() => document.documentElement.scrollWidth <= innerWidth)).toBe(true);
    });
  });
}
