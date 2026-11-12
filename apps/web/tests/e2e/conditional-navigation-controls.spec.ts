import { randomUUID } from "node:crypto";
import { expect, test, type BrowserContext, type Locator, type Page } from "@playwright/test";
import type { LessonSession, StorySession, VocabularySearchItem } from "@vocanova/api-client";

const api = `http://127.0.0.1:${process.env.MOCK_API_PORT ?? 8080}`;

async function headers(context: BrowserContext) {
  const csrf = (await context.cookies()).find(cookie => cookie.name === "vocanova_csrf")!.value;
  return { "X-CSRF-Token": csrf, "Idempotency-Key": randomUUID() };
}

async function activate(page: Page, link: Locator, href: string, heading: string, level = 1) {
  await expect(link).toHaveAttribute("href", href);
  await expect(link).toBeVisible();
  await link.focus();
  await expect(link).toBeFocused();
  await page.keyboard.press("Enter");
  await expect(page).toHaveURL(new URL(href, test.info().project.use.baseURL!).toString());
  await expect(page.getByRole("main").getByRole("heading", { name: heading, exact: true, level })).toBeVisible();
}

// Prepare genuine mock sessions through their mutation endpoints. The browser
// exercises the conditional navigation, while existing specs test the teaching.
async function completeStory(context: BrowserContext) {
  const started = await context.request.post(`${api}/api/v1/story-sessions`, {
    headers: await headers(context), data: { storyKey: "a-quiet-lunch" },
  });
  expect(started.ok()).toBe(true);
  let session = await started.json() as StorySession;
  for (let count = 0; session.status !== "completed" && count < 24; count++) {
    const step = session.currentStep!;
    const action = session.canContinue ? "continue" : "answer";
    const choice = step.choices?.find(item => ["She has a meeting at two.", "bill"].includes(item.text));
    if (action === "answer") expect(choice).toBeDefined();
    const result = await context.request.post(`${api}/api/v1/story-sessions/${session.id}/actions`, {
      headers: await headers(context),
      data: { stepId: step.id, expectedRevision: session.revision, clientActionId: randomUUID(), action,
        ...(action === "answer" ? { choiceId: choice!.id } : {}) },
    });
    expect(result.ok()).toBe(true);
    session = await result.json() as StorySession;
  }
  expect(session.status).toBe("completed");
  return session;
}

async function completeLesson(context: BrowserContext) {
  const started = await context.request.post(`${api}/api/v1/lessons/conversation-basics/sessions`, {
    headers: await headers(context), data: {},
  });
  expect(started.ok()).toBe(true);
  let session = await started.json() as LessonSession;
  for (let count = 0; session.status !== "completed" && count < 24; count++) {
    const step = session.currentStep!;
    const action = session.canContinue ? "continue" : "answer";
    const answer = session.words[session.completedSteps % session.words.length]!;
    const result = await context.request.post(`${api}/api/v1/lesson-sessions/${session.id}/actions`, {
      headers: await headers(context),
      data: { stepId: step.id, expectedRevision: session.revision, clientActionId: randomUUID(), action,
        ...(action === "answer" ? { choiceId: answer.meaningId } : {}) },
    });
    expect(result.ok()).toBe(true);
    session = await result.json() as LessonSession;
  }
  expect(session.status).toBe("completed");
  return session;
}

for (const theme of ["light", "dark"] as const) {
  test.describe(`conditional navigation (${theme})`, () => {
    test.beforeEach(async ({ page, context, baseURL }) => {
      await context.addCookies([
        { name: "vocanova_session", value: randomUUID(), url: baseURL! },
        { name: "vocanova_csrf", value: randomUUID(), url: baseURL! },
        { name: "vocanova_theme", value: theme, url: baseURL! },
      ]);
      await page.addInitScript(value => localStorage.setItem("vocanova:theme-preference", value), theme);
    });
    async function goto(page: Page, path: string) {
      await page.goto(path);
      await expect(page.locator("html")).toHaveAttribute("data-theme", theme);
    }

    test("exhausted saved vocabulary and sentence history return to their newest records", async ({ page, context, baseURL }) => {
      await context.addCookies([{ name: "e2e_saved_words", value: "filters", url: baseURL! }]);
      // Obtain an actual valid cursor, then shrink its collection. This models
      // an exhausted previously valid page without forging a cursor payload.
      const first = await context.request.get(`${api}/api/v1/user-words?limit=20`);
      expect(first.ok()).toBe(true);
      const { nextCursor } = await first.json();
      expect(nextCursor).toBeTruthy();
      await context.clearCookies({ name: "e2e_saved_words" });
      await goto(page, `/words?after=${encodeURIComponent(nextCursor)}`);
      await expect(page.getByRole("heading", { name: "You've reached the end", exact: true })).toBeVisible();
      await context.addCookies([{ name: "e2e_saved_words", value: "filters", url: baseURL! }]);
      await activate(page, page.getByRole("link", { name: "Back to newest words", exact: true }), "/words", "Saved vocabulary");
      await expect(page.getByRole("list", { name: "Saved vocabulary results" }).getByRole("listitem")).toHaveCount(20);

      await context.addCookies([{ name: "e2e_sentence_history_fixture", value: "paginated", url: baseURL! }]);
      await goto(page, "/progress/sentences?after=13");
      await expect(page.getByRole("heading", { name: "You've reached the end of this history", exact: true })).toBeVisible();
      await activate(page, page.getByRole("link", { name: "Back to newest sentences", exact: true }), "/progress/sentences", "Sentence history");
      await expect(page.getByRole("link", { name: "View older sentences", exact: true })).toBeVisible();
    });

    test("empty situation offers other available courses", async ({ page }) => {
      await goto(page, "/discover/navigating-an-airport");
      await expect(page.getByRole("heading", { name: "No words here yet", exact: true })).toBeVisible();
      await activate(page, page.getByRole("link", { name: "Browse other situations", exact: true }), "/courses", "Situation courses");
      await expect(page.getByRole("link", { name: /Start lesson.*Make a plan with a friend/ })).toBeVisible();
    });

    test("vocabulary invalid cursor and empty filters recover actual results", async ({ page, context, baseURL }) => {
      await context.addCookies([{ name: "e2e_vocabulary", value: "invalid-cursor", url: baseURL! }]);
      await goto(page, "/vocabulary?category=social&after=expired");
      await expect(page.getByText("This results page is no longer available.", { exact: true })).toBeVisible();
      await activate(page, page.getByRole("link", { name: "Load results again", exact: true }), "/vocabulary?category=social", "Word search");
      await expect(page.getByRole("link", { name: "invite", exact: true })).toBeVisible();
      await goto(page, "/vocabulary?q=pour&category=travel&level=a2");
      await expect(page.getByRole("heading", { name: "No matching words yet", exact: true })).toBeVisible();
      await activate(page, page.getByRole("link", { name: "Browse all vocabulary", exact: true }), "/vocabulary", "Word search");
      await expect(page.getByRole("link", { name: "pour", exact: true })).toBeVisible();
    });

    test("vocabulary Next results and First results activate real pagination", async ({ page, context, baseURL }) => {
      await context.addCookies([{ name: "e2e_vocabulary", value: "paginated", url: baseURL! }]);
      const responses = await Promise.all([
        context.request.get(`${api}/api/v1/canonical-words?limit=20`),
        context.request.get(`${api}/api/v1/canonical-words?limit=20&after=2`),
      ]);
      for (const response of responses) expect(response.ok()).toBe(true);
      const first = await responses[0]!.json() as { items: VocabularySearchItem[]; nextCursor: string };
      const second = await responses[1]!.json() as { items: VocabularySearchItem[] };
      expect(first.nextCursor).toBe("2");
      expect(first.items.length).toBeGreaterThan(0);
      expect(second.items.length).toBeGreaterThan(0);
      expect(second.items.map(item => item.meaningId)).not.toEqual(first.items.map(item => item.meaningId));
      async function expectRows(items: VocabularySearchItem[]) {
        const rows = page.getByRole("main").getByRole("listitem");
        await expect(rows).toHaveCount(items.length);
        for (const [index, item] of items.entries()) {
          const row = rows.nth(index);
          const word = row.getByRole("link", { name: item.wordText, exact: true });
          await expect(word).toBeVisible();
          await expect(word).toHaveAttribute("href", `/vocabulary/${encodeURIComponent(item.wordSlug)}#meaning-${item.meaningId}`);
          await expect(row.getByText(item.shortDefinition, { exact: true })).toBeVisible();
        }
      }
      await goto(page, "/vocabulary");
      await expectRows(first.items);
      await activate(page, page.getByRole("link", { name: "Next results", exact: true }), "/vocabulary?after=2", "Word search");
      await expectRows(second.items);
      await activate(page, page.getByRole("link", { name: "First results", exact: true }), "/vocabulary", "Word search");
      await expectRows(first.items);
      await expect(page.getByRole("link", { name: "Next results", exact: true })).toBeVisible();
    });

    test("completed story links open full reading, the situation guide and Journey", async ({ page, context, baseURL }) => {
      const session = await completeStory(context);
      await context.addCookies([{ name: "e2e_unit_guides", value: "true", url: baseURL! }]);
      for (const [name, href, heading] of [
        ["Read the full story", "/stories/a-quiet-lunch", "A quiet lunch"],
        ["Open the situation guide", "/discover/restaurant#unit-guide", "Restaurant"],
        ["Return to Journey", "/discover", "Journey"],
      ]) {
        await goto(page, `/stories/session/${session.id}`);
        await expect(page.getByRole("heading", { name: "Story complete", exact: true })).toBeVisible();
        await activate(page, page.getByRole("link", { name: name!, exact: true }), href!, heading!);
        if (name === "Read the full story") await expect(page.getByRole("heading", { name: "Full story", exact: true })).toBeVisible();
        if (name === "Open the situation guide") await expect(page.locator("#unit-guide")).toBeVisible();
      }
    });

    for (const destination of ["saved", "known"] as const) {
      test(`completed self-check opens ${destination} vocabulary`, async ({ page, context, baseURL }) => {
        await context.addCookies([{ name: "e2e_onboarding_focus", value: "work", url: baseURL! }]);
        const response = await context.request.get(`${api}/api/v1/canonical-words?knowledge=unexplored&limit=10`);
        expect(response.ok()).toBe(true);
        const { items } = await response.json() as { items: VocabularySearchItem[] };
        expect(items.length).toBeGreaterThanOrEqual(2);
        await goto(page, "/vocabulary/check");
        for (const [index, word] of items.entries()) {
          await expect(page.getByRole("main").getByRole("heading", { level: 2 })).toHaveText(word.wordText);
          await page.getByRole("button", { name: index === 0 ? "Already know" : index === 1 ? "Want to learn" : "Skip for now", exact: true }).click();
        }
        await expect(page.getByRole("heading", { name: "Your starting words are ready", exact: true })).toBeVisible();
        if (destination === "saved") {
          await activate(page, page.getByRole("link", { name: "Open saved vocabulary", exact: true }), "/words", "Saved vocabulary");
          await expect(page.getByRole("heading", { name: items[1]!.wordText, exact: true, level: 2 })).toBeVisible();
        } else {
          await activate(page, page.getByRole("link", { name: "Review your known words", exact: true }), "/vocabulary?knowledge=known", "Word search");
          await expect(page.getByRole("link", { name: items[0]!.wordText, exact: true })).toBeVisible();
        }
      });
    }

    test("completed lesson opens both newly saved and previously saved meanings", async ({ page, context }) => {
      const session = await completeLesson(context);
      const invite = session.words[0]!;
      const confirm = session.words[1]!;
      const reschedule = session.words[2]!;
      expect([invite.wordText, confirm.wordText, reschedule.wordText]).toEqual(["invite", "confirm", "reschedule"]);
      const saved = await context.request.post(`${api}/api/v1/user-words`, {
        headers: await headers(context), data: { meaningId: confirm.meaningId, source: "journey" },
      });
      expect(saved.ok()).toBe(true);
      await goto(page, "/learn/conversation-basics");
      await page.getByRole("button", { name: "Save invite for review", exact: true }).click();
      for (const word of session.words) {
        await expect(page.getByRole("main").getByRole("link", { name: word.wordText, exact: true }))
          .toHaveAttribute("href", `/vocabulary/${encodeURIComponent(word.wordSlug)}#meaning-${word.meaningId}`);
      }
      for (const word of [invite, confirm]) {
        // A returned lesson streams before its independently loaded saved
        // states. Establish this destination and its saved link before
        // asserting that the untouched sibling still offers saving.
        await expect(page.getByRole("main").getByRole("heading", {
          name: "Lesson complete", exact: true, level: 1,
        })).toBeVisible();
        const savedLink = page.getByRole("main").getByRole("link", {
          name: `Open saved ${word.wordText}`, exact: true,
        });
        await expect(savedLink).toBeVisible();
        await expect(page.getByRole("button", { name: "Save reschedule for review", exact: true })).toBeEnabled();
        await expect(page.getByRole("link", { name: "Open saved reschedule", exact: true })).toHaveCount(0);
        await activate(page, savedLink, `/words/uw-${word.meaningId}`, word.wordText);
        await goto(page, "/learn/conversation-basics");
      }
    });
  });
}
