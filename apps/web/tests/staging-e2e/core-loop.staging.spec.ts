// VOC-050-T02 - the core-loop journey run against a real deployed
// environment (staging), as deploy-staging.yml's post-deploy gate.
//
// Relationship to tests/e2e/core-loop.spec.ts (VOC-031-T08): same
// journey, different backend. That spec runs at PR time against
// mock-api-server.mjs and is untouched by this task. This spec runs
// after a staging deploy against the real Go API and real Postgres,
// so the two differ in the three places where the mock's shortcuts
// have no real-backend equivalent (specification.md's open
// questions 2 and 3):
//
//   1. Authentication. The mock accepts any vocanova_session value;
//      the real API requires a session row created by
//      CreateSession. This spec therefore consumes a session minted
//      server-side by VOC-050-T01's token-gated
//      POST /ops/synthetic-smoke-test/session for the seeded
//      synthetic account (VOC-050-T00), passed in as
//      E2E_SESSION_COOKIE / E2E_CSRF_TOKEN.
//   2. Onboarding. No e2e_onboarding_status override cookie is used
//      or expected - the real backend does not implement one, and
//      this task deliberately does not ask it to. The synthetic
//      account is seeded with onboarding_status = 'completed', so
//      the journey starts on /home. The onboarding form is still
//      driven for real if the app redirects there, so a
//      regression in onboarding is caught rather than hidden.
//   3. Fixture content. The mock serves fixed fixtures; staging
//      serves whatever canonical content is seeded. Every selector
//      here is content-agnostic (first situation, first word,
//      whichever prompt type the review session picks) instead of
//      hard-coding "Ordering at a cafe" / "pour".
//
// Rerun safety: this journey runs on every staging deploy against
// an account that keeps its state between runs. Each step is
// written to be correct whether or not a previous run already left
// the account with saved words, a due-review backlog, or none.

import { expect, test } from "@playwright/test";
import type { Locator, Page } from "@playwright/test";

const SESSION_COOKIE_ENV = "E2E_SESSION_COOKIE";
const CSRF_COOKIE_ENV = "E2E_CSRF_TOKEN";
const SESSION_COOKIE_NAME = "vocanova_session";
const CSRF_COOKIE_NAME = "vocanova_csrf";

// The review session refetches after the last card in its batch, so
// "review everything due" is a loop. The bound keeps a large
// accumulated backlog from consuming the whole deploy budget; when
// it is hit, the journey still proves the review path works and
// records why it stopped short of the caught-up state.
const MAX_REVIEW_CARDS = 8;
const REVIEWED_TODAY_PATTERN = /(\d+) of \d+ words reviewed today/;

// Cookies must be scoped to the shared parent domain, not just the web
// app's own host: the app's server-side rendering/fetch paths send these
// cookies on to api-staging.vocanova.site (a different subdomain from
// staging.vocanova.site), mirroring deploy-production.yml's own
// `SESSION_COOKIE_DOMAIN=.${HOST#*.}` computation and apps/api/.env.example's
// documented `.vocanova.site` convention. Using Playwright's `url`-only
// cookie form (as an earlier revision did) scopes the cookie to exactly
// the web app's host and never reaches the API host, which independent
// review flagged as a High finding blocking the real authenticated
// journey from working as wired.
function cookieDomainFor(baseURL: string): string {
  const override = process.env.STAGING_SESSION_COOKIE_DOMAIN;
  if (override) {
    return override;
  }
  const host = new URL(baseURL).hostname;
  const labels = host.split(".");
  return `.${labels.slice(1).join(".")}`;
}

function requiredEnv(name: string): string {
  const value = process.env[name];
  if (!value) {
    throw new Error(
      `${name} must be set for the staging core-loop journey. deploy-staging.yml supplies it from the VOC-050-T01 session-mint endpoint.`,
    );
  }
  return value;
}

async function readReviewedTodayCount(page: Page): Promise<number> {
  const counter = page.getByText(REVIEWED_TODAY_PATTERN);
  await expect(counter).toBeVisible();
  const text = (await counter.textContent()) ?? "";
  const match = REVIEWED_TODAY_PATTERN.exec(text);
  if (!match) {
    throw new Error(`Could not read the daily-mission counter from: ${text}`);
  }
  return Number(match[1]);
}

async function completeOnboardingIfRedirected(page: Page): Promise<void> {
  if (!page.url().includes("/onboarding")) {
    return;
  }

  await expect(
    page.getByRole("heading", { name: "Welcome to Vocanova", level: 1 }),
  ).toBeVisible();

  await page.getByRole("radio", { name: /A2/ }).check();
  await page.getByRole("button", { name: "Continue" }).click();

  await expect(
    page.getByRole("heading", { name: /What's your native language\?/ }),
  ).toBeVisible();
  await page.getByRole("textbox", { name: "Native language" }).fill("es");
  await page.getByRole("button", { name: "Continue" }).click();

  await expect(
    page.getByRole("radiogroup", {
      name: /What's your main reason for learning\?/,
    }),
  ).toBeVisible();
  await page.getByRole("radio", { name: "General growth" }).check();
  await page.getByRole("button", { name: "Continue" }).click();

  await expect(
    page.getByRole("radiogroup", { name: /Where will you use English most\?/ }),
  ).toBeVisible();
  await page.getByRole("radio", { name: "Daily life" }).check();
  await page.getByRole("button", { name: "Continue" }).click();

  await expect(
    page.getByRole("heading", { name: "Daily review target" }),
  ).toBeVisible();
  await page
    .getByRole("radiogroup", { name: "Daily review target" })
    .getByText("15", { exact: true })
    .click();
  await page.getByRole("button", { name: "Finish setup" }).click();

  await page.goto("/home");
  await expect(page).toHaveURL(/\/home(\?|$)/);
}

// Prefers a word the account has not saved yet: a newly-saved word
// has next_review_at NULL and so becomes due immediately, which is
// what keeps the review, sentence-feedback, and progress steps
// exercised rather than skipped on most runs. Falls back to the
// first word when every word in the situation is already saved.
async function chooseWordLink(page: Page): Promise<Locator> {
  const wordItems = page
    .locator("main ul > li")
    .filter({ has: page.locator("a") });
  const count = await wordItems.count();
  expect(count).toBeGreaterThan(0);

  for (let index = 0; index < count; index++) {
    const item = wordItems.nth(index);
    if ((await item.getByText("Saved").count()) === 0) {
      return item.locator("a").first();
    }
  }
  return wordItems.first().locator("a").first();
}

async function reviewOneCard(page: Page): Promise<void> {
  const showAnswerButton = page.getByRole("button", { name: "Show answer" });
  if (await showAnswerButton.isVisible()) {
    await showAnswerButton.click();
    await page.getByRole("button", { name: "Good" }).click();
    return;
  }

  // Multiple-choice prompt. Which option is correct is not knowable
  // from the rendered page, so the journey answers, then follows
  // whichever branch the app shows - both submit a real attempt.
  await page
    .getByRole("group", { name: /^Choose the meaning for / })
    .getByRole("button")
    .first()
    .click();

  const ratingButton = page.getByRole("button", { name: "Good" });
  const continueButton = page.getByRole("button", { name: "Continue" });
  await expect(ratingButton.or(continueButton)).toBeVisible();
  if (await ratingButton.isVisible()) {
    await ratingButton.click();
    return;
  }
  await continueButton.click();
}

test.describe("Core loop against real staging (VOC-050-T02)", () => {
  test("authenticated journey: home -> discover -> save -> review -> sentence feedback -> progress -> settings -> logout -> auth gate", async ({
    page,
    context,
  }, testInfo) => {
    const sessionCookie = requiredEnv(SESSION_COOKIE_ENV);
    const csrfToken = requiredEnv(CSRF_COOKIE_ENV);
    const baseURL = testInfo.project.use.baseURL;
    if (!baseURL) {
      throw new Error(
        "Expected playwright.staging.config.ts to configure use.baseURL so cookies can be scoped to the deployed origin.",
      );
    }

    await test.step("1. authenticate with the minted synthetic session", async () => {
      const domain = cookieDomainFor(baseURL);
      await context.addCookies([
        {
          name: SESSION_COOKIE_NAME,
          value: sessionCookie,
          domain,
          path: "/",
          secure: true,
          sameSite: "Lax",
        },
        {
          name: CSRF_COOKIE_NAME,
          value: csrfToken,
          domain,
          path: "/",
          secure: true,
          sameSite: "Lax",
        },
      ]);
      await page.goto("/home");
      await completeOnboardingIfRedirected(page);
      await expect(page).toHaveURL(/\/home(\?|$)/);
      await expect(
        page.getByRole("heading", { name: "Today's Mission", level: 1 }),
      ).toBeVisible();
    });

    const reviewedBefore = await test.step("2. read the daily-mission baseline", async () =>
      readReviewedTodayCount(page));

    await test.step("3. discover a situation and open a word", async () => {
      await page.goto("/discover");
      await expect(
        page.getByRole("heading", { name: "Journey", level: 1 }),
      ).toBeVisible();

      const situationLinks = page.locator("main ul > li > a");
      expect(await situationLinks.count()).toBeGreaterThan(0);
      await situationLinks.first().click();
      await expect(page).toHaveURL(/\/discover\/[^/]+(\?|$)/);
      await expect(page.getByRole("link", { name: "Back to Journey" })).toBeVisible();

      const wordLink = await chooseWordLink(page);
      await wordLink.click();
      await expect(page).toHaveURL(/\/discover\/[^/]+\/[^/]+(\?|$)/);
    });

    await test.step("4. save the word", async () => {
      const saveButton = page.locator("button[aria-pressed]").first();
      await expect(saveButton).toBeVisible();
      if ((await saveButton.getAttribute("aria-pressed")) !== "true") {
        await saveButton.click();
      }
      await expect(saveButton).toHaveAttribute("aria-pressed", "true");
    });

    const reviewedCards = await test.step("5. work the review queue", async () => {
      await page.goto("/reviews");
      await expect(
        page.getByRole("heading", { name: "Review", level: 1 }),
      ).toBeVisible();

      const caughtUpHeading = page.getByRole("heading", {
        name: "You're all caught up",
        level: 2,
      });

      const cardCounter = page.getByText(/^Card \d+ of \d+$/);

      let reviewed = 0;
      while (reviewed < MAX_REVIEW_CARDS) {
        if (await caughtUpHeading.isVisible()) {
          break;
        }
        await reviewOneCard(page);
        reviewed++;
        // The submission either advances to the next card or empties
        // the queue; both are settled states, so wait for one of them
        // instead of a fixed delay.
        await expect(caughtUpHeading.or(cardCounter).first()).toBeVisible();
      }
      return reviewed;
    });

    await test.step("6. sentence feedback on the reviewed word", async () => {
      const feedbackHeading = page.getByRole("heading", {
        name: /^Practice with /,
      });
      if (!(await feedbackHeading.isVisible())) {
        // The widget only renders in the caught-up state, for a card
        // reviewed in this same session. A run that started with a
        // backlog deeper than MAX_REVIEW_CARDS legitimately never
        // reaches it.
        testInfo.annotations.push({
          type: "skipped-step",
          description: `sentence feedback not reachable this run (reviewed ${reviewedCards} card(s), queue not emptied)`,
        });
        return;
      }

      const sentenceInput = page.getByRole("textbox", {
        name: /^Write a sentence using /,
      });
      const targetWord = (
        (await sentenceInput.getAttribute("placeholder")) ?? ""
      ).replace(/^Type a sentence using "(.*)"\.\.\.$/, "$1");
      await sentenceInput.fill(
        `Today I practise the word ${targetWord} in a full sentence.`,
      );

      const feedbackResponse = page.waitForResponse(
        (response) =>
          response.url().includes("/api/v1/sentence-feedback") &&
          response.request().method() === "POST",
      );
      await page.getByRole("button", { name: "Check my sentence" }).click();
      expect((await feedbackResponse).status()).toBe(200);

      // The real evaluator is not deterministic, so the assertion is
      // that the widget settled on a rendered outcome - a verdict or
      // an explicit error - not which verdict it was.
      const verdict = page.getByRole("status", { name: /^Feedback result: / });
      await expect(verdict.or(page.getByRole("alert")).first()).toBeVisible();
    });

    await test.step("7. progress reflects the completed reviews", async () => {
      await page.goto("/home");
      const reviewedAfter = await readReviewedTodayCount(page);
      expect(reviewedAfter).toBeGreaterThanOrEqual(reviewedBefore + reviewedCards);
    });

    await test.step("8. change a setting", async () => {
      await page.goto("/settings");
      await expect(
        page.getByRole("heading", { name: "Settings", level: 1 }),
      ).toBeVisible();
      const displayNameInput = page.getByRole("textbox", {
        name: "Display name",
      });
      const displayName = `Synthetic smoke-test ${Date.now()}`;
      await displayNameInput.fill(displayName);
      await page.getByRole("button", { name: "Save settings" }).click();
      await expect(page.getByText("Your settings have been saved.")).toBeVisible();
      await expect(displayNameInput).toHaveValue(displayName);
    });

    await test.step("9. log out", async () => {
      await page.getByRole("button", { name: "Log out" }).click();
      await expect(page).toHaveURL(/\/signin(\?|$)/);
      await expect(
        page.getByRole("heading", { name: "Sign in to Vocanova" }),
      ).toBeVisible();
    });

    await test.step("10. the auth gate rejects the now-invalid session", async () => {
      // No e2e_unauthenticated override cookie: logout revoked the
      // session server-side, so this is the real auth gate reacting
      // to a real 401 from /api/v1/me.
      await page.goto("/home");
      await expect(page).toHaveURL(/\/signin\?returnTo=%2Fhome(\b|$)/);
    });
  });
});
