import { randomUUID } from "node:crypto";
import { expect, test, type Page, type TestInfo } from "@playwright/test";
import { formatViolations, scanForAxeViolations } from "./axe-helper";
import { openOptionalWriting } from "./optional-writing";

// Screenshots are review artifacts, not a claim of visual approval. These cases
// visit states reached through real UI actions against isolated synthetic users.
async function capture(page: Page, info: TestInfo, state: string) {
  await expect(page.getByRole("main")).toBeVisible();
  await expect(page.locator("html")).toHaveAttribute("data-theme", state.endsWith("-dark") ? "dark" : "light");
  expect(await page.evaluate(() => document.documentElement.scrollWidth))
    .toBeLessThanOrEqual(page.viewportSize()!.width);
  // Preserve the actual interaction viewport before normalizing the scan.
  // Full-page stitching can position sticky headers mid-document and must not
  // be mistaken for the live viewport shown in the companion artifact.
  if (state.startsWith("saved-writing-")) {
    const focus = await page.evaluate(() => {
      const element = document.activeElement;
      if (!(element instanceof HTMLTextAreaElement)) return { editing: false, unobscured: false };
      const rect = element.getBoundingClientRect();
      const centerX = rect.x + rect.width / 2;
      const centerY = rect.y + rect.height / 2;
      const hit = document.elementFromPoint(centerX, centerY);
      return {
        editing: true,
        unobscured: rect.width > 0 && rect.height > 0 && centerX >= 0 && centerX <= innerWidth && centerY >= 0 && centerY <= innerHeight && Boolean(hit && element.contains(hit)),
      };
    });
    expect(focus.editing).toBe(true);
    expect(focus.unobscured, "The currently edited sentence must be visible above fixed navigation.").toBe(true);
  }
  await page.screenshot({ path: info.outputPath(`${state}-viewport.png`) });
  await page.screenshot({ path: info.outputPath(`${state}.png`), fullPage: true });
  // At an arbitrary editing scroll offset axe can interpret an earlier,
  // offscreen control under the sticky header as an undersized touch target.
  // Run every rule from a stable document origin; keep the genuine editing
  // viewport/focus evidence above and restore it without changing focus.
  const scroll = await page.evaluate(() => ({ x: scrollX, y: scrollY }));
  try {
    await page.evaluate(() => window.scrollTo({ top: 0, left: 0, behavior: "instant" }));
    const scan = await scanForAxeViolations(page);
    expect(scan.criticalOrSerious, formatViolations(scan.criticalOrSerious).join("\n")).toEqual([]);
  } finally {
    await page.evaluate(({ x, y }) => window.scrollTo({ top: y, left: x, behavior: "instant" }), scroll);
  }
}

for (const theme of ["light", "dark"] as const) {
  test.describe(`dynamic visual states (${theme})`, () => {
    test.afterEach(async ({ page }) => {
      await expect(page.locator("html")).toHaveAttribute("data-theme", theme);
    });

    test.beforeEach(async ({ page, context, baseURL }) => {
      await page.addInitScript((preference) => localStorage.setItem("vocanova:theme-preference", preference), theme);
      await context.addCookies([
        { name: "vocanova_session", value: randomUUID(), url: baseURL! },
        { name: "vocanova_csrf", value: randomUUID(), url: baseURL! },
        { name: "vocanova_theme", value: theme, url: baseURL! },
      ]);
    });

    test("guided teaching, question and wrong-answer states", async ({ page }, info) => {
      await page.goto("/learn/conversation-basics");
      const opened = page.waitForResponse((r) => r.request().method() === "POST" && r.url().endsWith("/lessons/conversation-basics/sessions"));
      await page.getByRole("button", { name: "Start lesson", exact: true }).click();
      const { words } = await (await opened).json();
      for (const [index, word] of ["invite", "confirm", "reschedule"].entries()) {
        await expect(page.getByRole("heading", { name: word, exact: true })).toBeVisible();
        if (index === 0) await capture(page, info, `lesson-teaching-${theme}`);
        await page.getByRole("button", { name: "Continue", exact: true }).click();
      }
      await expect(page.getByRole("radio").first()).toBeVisible();
      await capture(page, info, `lesson-question-${theme}`);
      await page.getByRole("radio", { name: words[1].definition, exact: true }).check();
      await page.getByRole("button", { name: "Check answer", exact: true }).click();
      await expect(page.getByRole("heading", { name: "Let’s look again" })).toBeVisible();
      await capture(page, info, `lesson-wrong-${theme}`);
    });

    test("review question, confirmed answer and completion", async ({ page, context, baseURL }, info) => {
      await context.addCookies([{ name: "e2e_review_fixture_count", value: "1", url: baseURL! }]);
      await page.goto("/reviews");
      await page.getByRole("button", { name: "Start review", exact: true }).click();
      await expect(page.getByRole("heading", { name: "Review word 1", exact: true })).toBeVisible();
      await capture(page, info, `review-question-${theme}`);
      await page.getByRole("radio", { name: "definition for review word 1", exact: true }).check();
      await page.getByRole("button", { name: "Check answer", exact: true }).click();
      await expect(page.getByRole("status").filter({ hasText: /^Correct/ })).toBeVisible();
      await capture(page, info, `review-correct-${theme}`);
      await page.getByRole("button", { name: "Continue", exact: true }).click();
      await expect(page.getByRole("heading", { name: /^(?:Review saved|Review set complete|You’ve finished your review)$/ })).toBeVisible();
      await capture(page, info, `review-complete-${theme}`);
    });

    test("typed practice question, assisted feedback and completion", async ({ page }, info) => {
      await page.goto("/practice");
      await page.getByRole("main").getByText("More practice", { exact: true }).click();
      await page.getByRole("button", { name: "Start typed recall", exact: true }).click();
      await expect(page).toHaveURL(/\/practice\/session\//);
      await expect(page.getByRole("textbox", { name: "Your answer" })).toBeVisible();
      await capture(page, info, `practice-question-${theme}`);
      for (let i = 0; i < 3; i++) {
        await page.getByRole("button", { name: "Show answer", exact: true }).click();
        const next = page.getByRole("button", { name: i === 2 ? "Finish practice" : "Continue", exact: true });
        await expect(next).toBeVisible();
        if (i === 0) await capture(page, info, `practice-assisted-${theme}`);
        await next.click();
      }
      await expect(page.getByRole("heading", { name: "Practice complete", exact: true })).toBeVisible();
      await capture(page, info, `practice-complete-${theme}`);
    });

    test("story dialogue, comprehension and wrong-answer states", async ({ page }, info) => {
      await page.goto("/stories/a-quiet-lunch");
      await page.getByRole("button", { name: "Start story", exact: true }).click();
      await expect(page).toHaveURL(/\/stories\/session\//);
      await expect(page.getByRole("button", { name: "Continue", exact: true })).toBeVisible();
      await capture(page, info, `story-dialogue-${theme}`);
      for (let i = 0; i < 4; i++) {
        const response = page.waitForResponse((r) => r.request().method() === "POST" && /\/story-sessions\/[^/]+\/actions$/.test(new URL(r.url()).pathname));
        await page.getByRole("button", { name: "Continue", exact: true }).click();
        const session = await (await response).json();
        await expect(page.getByRole("main").getByText(session.currentStep.kind === "line" ? session.currentStep.line.text : session.currentStep.prompt, { exact: true })).toBeVisible();
      }
      await capture(page, info, `story-question-${theme}`);
      await page.getByRole("radio", { name: "The restaurant is closing now.", exact: true }).check();
      await page.getByRole("button", { name: "Check answer", exact: true }).click();
      await expect(page.getByText("Let's try that again.", { exact: true })).toBeVisible();
      await capture(page, info, `story-wrong-${theme}`);
    });

    test("created list confirmation and saved-detail writing states", async ({ page }, info) => {
      await page.goto("/lists");
      await page.getByLabel("List name", { exact: true }).fill("Words for everyday conversations");
      await page.getByRole("button", { name: "Create list", exact: true }).click();
      await expect(page).toHaveURL(/\/lists\/[^/]+$/);
      await expect(page.getByRole("heading", { name: "Words for everyday conversations", exact: true })).toBeVisible();
      let release!: () => void;
      const pending = new Promise<void>((resolve) => { release = resolve; });
      await page.route("**/api/v1/word-lists/*", async (route) => {
        if (route.request().method() === "PUT") await pending;
        await route.continue();
      });
      await page.locator("summary").filter({ hasText: /^List settings$/ }).click();
      await page.getByLabel("List name", { exact: true }).fill("Everyday conversation plans");
      await page.getByRole("button", { name: "Rename list", exact: true }).click();
      try {
        await expect(page.getByLabel("List name", { exact: true })).toBeDisabled();
        await capture(page, info, `list-pending-${theme}`);
      } finally { release(); }
      await expect(page.getByRole("heading", { name: "Everyday conversation plans", exact: true })).toBeVisible();
      await page.getByRole("button", { name: "Delete list", exact: true }).click();
      await expect(page.getByRole("button", { name: "Keep list", exact: true })).toBeVisible();
      await capture(page, info, `list-confirmation-${theme}`);
      await page.getByRole("button", { name: "Keep list", exact: true }).click();
      await page.goto("/discover/ordering-at-a-cafe/pour");
      await page.getByRole("button", { name: /^Save pour:/ }).click();
      await expect(page.getByRole("button", { name: "Remove pour from saved words" })).toBeVisible();
      // The button confirms the POST before router.refresh() has committed.
      // This server-gated disclosure only mounts once the refreshed meaning
      // has saved=true and its userWordId. Wait for that commit before issuing
      // a separate document navigation, which otherwise aborts it in Firefox.
      await expect(page.getByRole("main").locator("summary").filter({ hasText: /^Optional writing$/ })).toHaveCount(1);
      await page.goto("/words");
      await page.getByRole("link", { name: "Open pour details and sentence practice" }).click();
      await expect(page).toHaveURL(/\/words\/uw-/);
      await expect(page.getByRole("heading", { name: "pour", exact: true })).toBeVisible();
      await capture(page, info, `saved-detail-${theme}`);
      await openOptionalWriting(page);
      const editor = page.getByRole("textbox", { name: "Write a sentence using pour" });
      // WebKit fill() can focus an offscreen textarea without scrolling it.
      // Enter writing through a real pointer interaction before recording its
      // viewport; capture still requires the active editor to be unobscured.
      await editor.click();
      await editor.fill("Could you pour some tea for my friend, please?");
      await capture(page, info, `saved-writing-${theme}`);
    });
  });
}
