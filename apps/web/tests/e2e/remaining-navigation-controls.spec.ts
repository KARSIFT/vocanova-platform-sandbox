import { randomUUID } from "node:crypto";
import { expect, test, type Locator, type Page } from "@playwright/test";

const apiURL = `http://127.0.0.1:${process.env.MOCK_API_PORT ?? 8080}`;

async function activate(page: Page, link: Locator, href: string, heading: string) {
  await expect(link).toHaveAttribute("href", href);
  await link.focus();
  await expect(link).toBeFocused();
  await page.keyboard.press("Enter");
  await expect(page).toHaveURL(new URL(href, test.info().project.use.baseURL!).toString());
  await expect(page.getByRole("main").getByRole("heading", { name: heading, level: 1, exact: true })).toBeVisible();
}

async function writeHeaders(page: Page) {
  const csrf = (await page.context().cookies()).find(cookie => cookie.name === "vocanova_csrf")!.value;
  return { "X-CSRF-Token": csrf, "Idempotency-Key": randomUUID() };
}

for (const theme of ["light", "dark"] as const) {
  test.describe(`remaining navigation (${theme})`, () => {
    test.beforeEach(async ({ page, context, baseURL }) => {
      await context.addCookies([
        { name: "vocanova_session", value: randomUUID(), url: baseURL! },
        { name: "vocanova_csrf", value: randomUUID(), url: baseURL! },
        { name: "vocanova_theme", value: theme, url: baseURL! },
        { name: "e2e_daily_conversation", value: "true", url: baseURL! },
      ]);
      await page.addInitScript(value => localStorage.setItem("vocanova:theme-preference", value), theme);
    });

    async function visit(page: Page, path: string, heading: string) {
      await page.goto(path);
      await expect(page.locator("html")).toHaveAttribute("data-theme", theme);
      await expect(page.getByRole("main").getByRole("heading", { name: heading, level: 1, exact: true })).toBeVisible();
    }

    test("saved words and Journey secondary links reach their healthy destinations", async ({ page }) => {
      for (const [source, sourceHeading, name, href, heading] of [
        ["/words", "Saved vocabulary", "Explore Journey", "/discover", "Journey"],
        ["/words", "Saved vocabulary", "Organize meanings in personal lists", "/lists", "Personal lists"],
        ["/discover", "Journey", "All situations", "/courses", "Situation courses"],
      ] as const) {
        await visit(page, source, sourceHeading);
        await activate(page, page.getByRole("main").getByRole("link", { name, exact: true }), href, heading);
      }
      const canonical = await page.request.get(`${apiURL}/api/v1/canonical-words/invite`);
      expect(canonical.status()).toBe(200);
      const { word } = await canonical.json();
      const saved = await page.request.post(`${apiURL}/api/v1/user-words`, {
        headers: await writeHeaders(page), data: { meaningId: word.meanings[0].id, source: "journey" },
      });
      expect(saved.status()).toBe(200);
      await visit(page, "/discover/daily-conversation", "Daily Conversation");
      const progress = page.getByRole("region", { name: "Situation progress", exact: true });
      await expect(progress).toContainText(/1 of \d+ words saved/);
      await activate(page, progress.getByRole("link", { name: "View saved vocabulary", exact: true }), "/words", "Saved vocabulary");
      await expect(page.getByRole("link", { name: "Open invite details and sentence practice", exact: true })).toBeVisible();
      await expect(page.getByRole("main").getByRole("link", { name: "Explore Journey", exact: true })).toHaveCount(0);
      await activate(page, page.getByRole("main").getByRole("link", { name: "Discover new words", exact: true }), "/discover", "Journey");
    });

    test("empty lists and meaning-list tools open the actual selected list", async ({ page }) => {
      await visit(page, "/lists", "Personal lists");
      await expect(page.getByRole("heading", { name: "Give your words a place", exact: true })).toBeVisible();
      await activate(page, page.getByRole("link", { name: "Find useful words", exact: true }), "/vocabulary", "Word search");
      const listID = randomUUID();
      const created = await page.request.put(`${apiURL}/api/v1/word-lists/${listID}`, {
        headers: await writeHeaders(page), data: { name: "Navigation practice", expectedRevision: 0 },
      });
      expect(created.status()).toBe(200);
      for (const [name, href, heading] of [
        ["Open list", `/lists/${listID}`, "Navigation practice"],
        ["Manage personal lists", "/lists", "Personal lists"],
      ] as const) {
        await visit(page, "/vocabulary/invite", "invite");
        const personal = page.locator("summary").filter({ hasText: "Personal tools" });
        await personal.focus();
        await page.keyboard.press("Enter");
        await page.getByRole("button", { name: "Choose personal lists", exact: true }).press("Enter");
        await page.getByLabel("Choose a list for this meaning", { exact: true }).selectOption(listID);
        await expect(page.getByText("This meaning is not in this list.", { exact: true })).toBeVisible();
        await activate(page, page.getByRole("main").getByRole("link", { name, exact: true }), href, heading);
      }
    });

    test("Practice return and personal-list management links work", async ({ page }) => {
      await visit(page, "/practice", "Today’s practice");
      await activate(page, page.getByRole("link", { name: "Back to Today", exact: true }), "/home", "Today");
      await visit(page, "/practice", "Today’s practice");
      const more = page.getByRole("main").locator("summary").filter({ hasText: "More practice" });
      await more.focus();
      await page.keyboard.press("Enter");
      await activate(page, page.getByRole("link", { name: "Manage personal lists", exact: true }), "/lists", "Personal lists");
    });

    test("vocabulary view links and self-check exits reach the requested view", async ({ page }) => {
      await visit(page, "/vocabulary", "Word search");
      const views = page.getByRole("navigation", { name: "Vocabulary views", exact: true });
      await activate(page, views.getByRole("link", { name: "Knowledge map", exact: true }), "/vocabulary?view=map", "Your vocabulary map");
      await expect(views.getByRole("link", { name: "Knowledge map", exact: true })).toHaveAttribute("aria-current", "page");
      await activate(page, views.getByRole("link", { name: "Word search", exact: true }), "/vocabulary", "Word search");
      await activate(page, views.getByRole("link", { name: "Find your starting words", exact: true }), "/vocabulary/check", "Find your starting words");
      await activate(page, page.getByRole("link", { name: "Back to vocabulary", exact: true }), "/vocabulary", "Word search");
      await activate(page, views.getByRole("link", { name: "Saved vocabulary", exact: true }), "/words", "Saved vocabulary");
      await visit(page, "/vocabulary/check", "Find your starting words");
      await activate(page, page.getByRole("link", { name: "Go to your learning plan", exact: true }), "/plan", "Your learning plan");
    });

    test("password sign-in opens recovery and signup with its intended return", async ({ page, context }) => {
      await context.clearCookies({ name: "vocanova_session" });
      await visit(page, "/login?returnTo=%2Fwords", "Sign in to Vocanova");
      await activate(page, page.getByRole("link", { name: "Forgot password?", exact: true }), "/auth/password/reset", "Reset or add a password");
      await activate(page, page.getByRole("link", { name: "Back to sign in", exact: true }), "/login", "Sign in to Vocanova");
      await visit(page, "/login?returnTo=%2Fwords", "Sign in to Vocanova");
      await activate(page, page.getByRole("link", { name: "Create account", exact: true }), "/signup?returnTo=%2Fwords", "Create your account");
    });

    test("email-change success and rejection offer working account and home exits", async ({ page }) => {
      await visit(page, "/auth/email-change?token=test-token", "Confirm your new email");
      await expect(page.getByRole("status")).toContainText("Your sign-in email is now");
      await activate(page, page.getByRole("link", { name: "Return to account settings", exact: true }), "/settings/account", "Account");
      for (const [name, href, heading] of [
        ["Go to account settings", "/settings/account", "Account"],
        ["Back to home", "/", "Learn the words you will actually use."],
      ] as const) {
        await visit(page, "/auth/email-change?token=invalid-token", "Confirm your new email");
        await expect(page.getByRole("main").getByRole("alert")).toContainText("invalid or has expired");
        await activate(page, page.getByRole("link", { name, exact: true }), href, heading);
      }
    });

    test("signed-out email-change exits preserve its token and magic-only sign-in", async ({ page, context, baseURL }) => {
      await context.addCookies([{ name: "e2e_unauthenticated", value: "1", url: baseURL! }]);
      const source = "/auth/email-change?token=test-token";
      const signin = `/login?${new URLSearchParams({ returnTo: source, magicOnly: "1" })}`;
      for (const [name, href, heading] of [
        ["Sign in", signin, "Sign in to Vocanova"],
        ["Back to home", "/", "Learn the words you will actually use."],
      ] as const) {
        await visit(page, source, "Confirm your new email");
        await expect(page.getByText("Sign in to the account that requested this change", { exact: false })).toBeVisible();
        await activate(page, page.getByRole("link", { name, exact: true }), href, heading);
        if (name === "Sign in") {
          await expect(page.getByRole("button", { name: "Send sign-in link", exact: true })).toBeVisible();
          await expect(page.getByLabel("Password", { exact: true })).toHaveCount(0);
        }
      }
    });

    test("Home Continue review resumes the confirmed review without a second start", async ({ page, context, baseURL }) => {
      await context.addCookies([{ name: "e2e_review_fixture_count", value: "2", url: baseURL! }]);
      let starts = 0;
      page.on("request", request => {
        if (request.method() === "POST" && new URL(request.url()).pathname === "/api/v1/daily-lessons") starts += 1;
      });
      await visit(page, "/home", "Today");
      await page.getByRole("button", { name: "Start review", exact: true }).click();
      await expect(page).toHaveURL(/\/review\/session\?lesson=/);
      await expect(page.getByRole("heading", { name: "Review word 1", level: 1, exact: true })).toBeVisible();
      const lessonURL = page.url();
      await visit(page, "/home", "Today");
      const resume = page.getByRole("link", { name: "Continue review", exact: true });
      await expect(resume).toHaveAttribute("href", "/review");
      await resume.focus();
      await expect(resume).toBeFocused();
      await page.keyboard.press("Enter");
      // The player canonicalizes /review to its durable session address.
      // Require the exact original session, not a transient /review URL.
      await expect(page).toHaveURL(lessonURL);
      await expect(page.getByRole("heading", { name: "Review word 1", level: 1, exact: true })).toBeVisible();
      await expect(page.getByRole("radio", { name: "definition for review word 1", exact: true })).toBeVisible();
      expect(starts).toBe(1);
    });

    test("Writing offers the full situation catalogue when its topic list has more", async ({ page, context, baseURL }) => {
      await context.addCookies([{ name: "e2e_journey_situations", value: "has-more", url: baseURL! }]);
      await visit(page, "/writing", "Topic writing");
      await expect(page.getByRole("navigation", { name: "Writing situations", exact: true }).getByRole("link")).toHaveCount(1);
      await activate(page, page.getByRole("link", { name: "Explore more situations", exact: true }), "/courses", "Situation courses");
    });

    for (const source of ["/home", "/practice"] as const) {
      test(`${source} Open practice fallback independently opens its confirmed session without starting again`, async ({ page, context }) => {
        await visit(page, source, source === "/home" ? "Today" : "Today’s practice");
        let release!: () => void;
        let arrived!: () => void;
        const held = new Promise<void>(resolve => { release = resolve; });
        const navigationStarted = new Promise<void>(resolve => { arrived = resolve; });
        let starts = 0;
        context.on("request", request => {
          if (request.method() === "POST" && new URL(request.url()).pathname === "/api/v1/practice-sessions") starts += 1;
        });
        await page.route("**/practice/session/**", async route => {
          if (route.request().isNavigationRequest()) {
            await route.continue();
            return;
          }
          arrived();
          await held;
          await route.continue();
        });
        const created = page.waitForResponse(response => response.request().method() === "POST" && new URL(response.url()).pathname === "/api/v1/practice-sessions");
        await page.getByRole("button", { name: "Start practice lesson", exact: true }).click();
        const response = await created;
        expect(response.status()).toBe(200);
        const session = await response.json();
        const href = `/practice/session/${session.id}`;
        expect(session.mode).toBe("journey");
        try {
          await navigationStarted;
          // Keep all client-router requests held until ordinary keyboard
          // activation independently loads the confirmed session document.
          // Releasing the automatic navigation cannot satisfy these assertions.
          const fallback = page.getByRole("link", { name: "Open practice", exact: true });
          await expect(fallback).toHaveAttribute("href", href);
          await fallback.focus();
          await expect(fallback).toBeFocused();
          await Promise.all([
            page.waitForRequest(request => request.isNavigationRequest()
              && request.resourceType() === "document"
              && new URL(request.url()).pathname === href, { timeout: 5000 }),
            page.keyboard.press("Enter"),
          ]);
          await expect(page).toHaveURL(new URL(href, test.info().project.use.baseURL!).toString());
          await expect(page.getByRole("heading", { name: "invite", level: 1, exact: true })).toBeVisible();
          await expect(page.getByRole("button", { name: "Continue", exact: true })).toBeEnabled();
          expect(starts).toBe(1);
        } finally {
          release();
        }
        await expect(page).toHaveURL(new URL(href, test.info().project.use.baseURL!).toString());
        await expect(page.getByRole("heading", { name: "invite", level: 1, exact: true })).toBeVisible();
        expect(starts).toBe(1);
      });
    }
  });
}
