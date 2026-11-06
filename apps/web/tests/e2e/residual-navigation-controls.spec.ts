import { randomUUID } from "node:crypto";
import { expect, test, type Locator, type Page } from "@playwright/test";

async function activate(page: Page, link: Locator, href: string, heading: string | RegExp, level?: number) {
  await expect(link).toHaveAttribute("href", href);
  await link.focus();
  await expect(link).toBeFocused();
  await page.keyboard.press("Enter");
  await expect(page).toHaveURL(new URL(href, test.info().project.use.baseURL!).toString());
  await expect(page.getByRole("heading", { name: heading, exact: true, ...(level ? { level } : {}) })).toBeVisible();
}

async function saveInvite(page: Page) {
  const csrf = (await page.context().cookies()).find(cookie => cookie.name === "vocanova_csrf")!.value;
  const api = `http://127.0.0.1:${process.env.MOCK_API_PORT ?? 8080}`;
  const canonical = await page.request.get(`${api}/api/v1/canonical-words/invite`);
  expect(canonical.ok()).toBe(true);
  const { word } = await canonical.json();
  const saved = await page.request.post(`${api}/api/v1/user-words`, {
    headers: { "X-CSRF-Token": csrf, "Idempotency-Key": randomUUID() },
    data: { meaningId: word.meanings[0].id, source: "journey" },
  });
  expect(saved.ok()).toBe(true);
  return { api, csrf, wordPath: `/words/uw-${word.meanings[0].id}` };
}

for (const theme of ["light", "dark"] as const) {
  test.describe(`residual navigation (${theme})`, () => {
    async function themedGoto(page: Page, path: string) {
      await page.goto(path);
      await expect(page.locator("html")).toHaveAttribute("data-theme", theme);
    }
    test.beforeEach(async ({ page, context, baseURL }) => {
      await context.addCookies([
        { name: "vocanova_session", value: randomUUID(), url: baseURL! },
        { name: "vocanova_csrf", value: randomUUID(), url: baseURL! },
        { name: "vocanova_theme", value: theme, url: baseURL! },
        { name: "e2e_daily_conversation", value: "true", url: baseURL! },
      ]);
      await page.addInitScript(value => localStorage.setItem("vocanova:theme-preference", value), theme);
    });

    test("magic-link incomplete and rejected results preserve sign-in returnTo and offer home", async ({ page }) => {
      // The shared mock's consume endpoint always succeeds. Intercept only this
      // frontend failure branch; no real token or email is used or sent.
      await page.route("**/api/v1/auth/magic-links/consume", route => route.fulfill({
        status: 401, json: { error: { code: "unauthorized", message: "Synthetic expired link" } },
      }));
      const returnTo = "/reviews?mode=due";
      for (const rejected of [false, true]) {
        const query = new URLSearchParams({ returnTo, ...(rejected ? { token: "synthetic-expired-token", email: "navigation@example.test" } : {}) });
        for (const destination of ["signin", "home"] as const) {
          await themedGoto(page, `/auth/magic?${query}`);
          await expect(page.getByRole("main").getByRole("alert")).toContainText(rejected ? "invalid, expired, or has already been used" : "sign-in link is incomplete");
          const name = destination === "signin" ? "Back to sign in" : "Back to home";
          const href = destination === "signin" ? `/login?${new URLSearchParams({ returnTo })}` : "/";
          await activate(page, page.getByRole("link", { name, exact: true }), href,
            destination === "signin" ? "Sign in to Vocanova" : /Learn the words you will actually use\./, 1);
        }
      }
    });

    test("situation guide opens its writing topic and guided lesson", async ({ page }) => {
      for (const destination of ["writing", "lesson"] as const) {
        await themedGoto(page, "/discover/daily-conversation");
        const guide = page.getByRole("region", { name: "A quick guide", exact: true });
        await expect(guide).toBeVisible();
        if (destination === "writing") {
          await activate(page, guide.getByRole("link", { name: "Write about this situation", exact: true }),
            "/writing?situation=daily-conversation", "Topic writing", 1);
          await expect(page.getByRole("heading", { name: "Daily Conversation", exact: true })).toBeVisible();
          await expect(page.getByText("A plan with a friend has changed.", { exact: false })).toBeVisible();
        } else {
          await activate(page, guide.getByRole("link", { name: "Start guided lesson", exact: true }),
            "/learn/conversation-basics", "Make a plan with a friend", 1);
          await expect(page.getByRole("button", { name: "Start lesson", exact: true })).toBeVisible();
        }
      }
    });

    test("healthy Writing links open the selected situation, saved vocabulary and Practice", async ({ page }) => {
      for (const [name, href, heading] of [
        ["Explore this situation", "/discover/daily-conversation", "Daily Conversation"],
        ["Your saved vocabulary", "/words", "Saved vocabulary"],
        ["Back to Practice", "/practice", "Today’s practice"],
      ] as const) {
        await themedGoto(page, "/writing?situation=daily-conversation");
        const main = page.getByRole("main");
        await expect(main.getByRole("heading", { name: "Topic writing", level: 1, exact: true })).toBeVisible();
        await expect(main.getByRole("heading", { name: "Daily Conversation", level: 2, exact: true })).toBeVisible();
        await expect(main.getByText("A plan with a friend has changed.", { exact: false })).toBeVisible();
        await activate(page, main.getByRole("link", { name, exact: true }), href, heading, 1);
        await expect(page.locator("html")).toHaveAttribute("data-theme", theme);
      }
    });

    test("Home latest-word links open saved vocabulary and reveal optional writing directly", async ({ page }) => {
      const { wordPath } = await saveInvite(page);
      await themedGoto(page, "/home");
      const latest = page.getByRole("region", { name: "Your latest word", exact: true });
      await activate(page, latest.getByRole("link", { name: "All your words", exact: true }), "/words", "Saved vocabulary", 1);
      await themedGoto(page, "/home");
      await activate(page, latest.getByRole("link", { name: "Write a sentence with it", exact: true }), `${wordPath}#sentence-practice`, "invite", 1);
      // Do not open disclosures in the test: the hash link must reveal them.
      await expect(page.getByRole("textbox", { name: "Write a sentence using invite", exact: true })).toBeVisible();
      await expect(page.locator("#sentence-practice")).toHaveAttribute("open", "");
    });

    test("missing saved word and empty sentence history recovery links reach healthy pages", async ({ page }) => {
      await themedGoto(page, `/words/${randomUUID()}`);
      await expect(page.getByRole("heading", { name: "Saved word not found", exact: true })).toBeVisible();
      await activate(page, page.getByRole("link", { name: "Back to saved vocabulary", exact: true }), "/words", "Saved vocabulary", 1);
      await themedGoto(page, "/progress/sentences");
      await expect(page.getByRole("heading", { name: "Your sentence practice will appear here", exact: true })).toBeVisible();
      await activate(page, page.getByRole("link", { name: "Go to Home", exact: true }), "/home", "Today", 1);
    });

    test("Profile secondary links reach settings and account", async ({ page }) => {
      for (const [name, href, heading] of [
        ["Back to Settings", "/settings", "Settings"],
        ["Account security", "/settings/account", "Account"],
        ["Learning preferences", "/settings", "Settings"],
      ]) {
        await themedGoto(page, "/settings/profile");
        await expect(page.getByRole("heading", { name: "Profile", level: 1 })).toBeVisible();
        await activate(page, page.getByRole("main").getByRole("link", { name, exact: true }), href!, heading!, 1);
      }
    });

    test("Progress due review and saved lesson links return to their actual activities", async ({ page }) => {
      const { api, csrf } = await saveInvite(page);
      await themedGoto(page, "/progress");
      await activate(page, page.getByRole("region", { name: "Your saved words", exact: true }).getByRole("link", { name: "Start review", exact: true }),
        "/review", "Keep your words fresh", 1);
      const started = await page.request.post(`${api}/api/v1/lessons/conversation-basics/sessions`, {
        headers: { "X-CSRF-Token": csrf, "Idempotency-Key": randomUUID() }, data: {},
      });
      expect(started.ok()).toBe(true);
      await themedGoto(page, "/progress");
      const main = page.getByRole("main");
      await expect(main.getByRole("heading", { name: "Progress", level: 1, exact: true })).toBeVisible();
      const path = main.locator("summary").filter({ hasText: "Your learning path" });
      await expect(path).toHaveCount(1);
      await expect(path).toBeVisible();
      await path.focus();
      await page.keyboard.press("Enter");
      await activate(page, page.getByRole("main").getByRole("link", { name: "Make a plan with a friend", exact: true }),
        "/learn/conversation-basics", "invite");
    });

    test("Progress empty activity and failed writing links offer usable recovery", async ({ page, context, baseURL }) => {
      await context.addCookies([{ name: "e2e_progress_history_fixture", value: "empty", url: baseURL! }]);
      await themedGoto(page, "/progress");
      await expect(page.getByText("Complete a daily mission to start your history.", { exact: true })).toBeVisible();
      await activate(page, page.getByRole("link", { name: "Go to today's mission", exact: true }), "/home", "Today", 1);
      await context.addCookies([{ name: "e2e_sentence_history_fixture", value: "error", url: baseURL! }]);
      await themedGoto(page, "/progress");
      const retry = page.getByRole("link", { name: "Open sentence history to try again.", exact: true });
      await expect(retry).toBeVisible();
      await context.clearCookies({ name: "e2e_sentence_history_fixture" });
      await activate(page, retry, "/progress/sentences", "Sentence history", 1);
      await expect(page.getByRole("heading", { name: "Your sentence practice will appear here", exact: true })).toBeVisible();
    });

    test("Courses empty catalogue sends the learner to available reviews", async ({ page, context, baseURL }) => {
      await context.addCookies([
        { name: "e2e_lessons", value: "empty", url: baseURL! },
        { name: "e2e_journey_situations", value: "empty", url: baseURL! },
      ]);
      await themedGoto(page, "/courses");
      await expect(page.getByRole("heading", { name: "New situations are on the way", exact: true })).toBeVisible();
      await activate(page, page.getByRole("link", { name: "Go to Reviews", exact: true }), "/review", "You’re ready for new words", 1);
    });
  });
}
