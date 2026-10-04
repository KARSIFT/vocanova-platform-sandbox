// Real-staging acceptance for PR #1484's connected learning features.
// Explicitly opt in and run THIS spec with a fresh synthetic session. The core
// journey logs out, so its already-revoked token cannot be reused here.
// No mock server, AI feedback request, vocabulary reset or session-history purge.
import { randomUUID } from "node:crypto";
import { expect, test, type BrowserContext } from "@playwright/test";
import type {
  CurrentUser,
  PracticeSession,
  StorySession,
  WordDetailResponse,
  WordListDetail,
} from "@vocanova/api-client";

const MEANING_ID = "329c4ec9-6562-568e-9e71-9134da2f0d8d";
const STORY_KEY = "a-quiet-lunch";

function requiredEnv(name: string): string {
  const value = process.env[name];
  if (!value)
    throw new Error(
      `${name} is required for the explicitly requested staging maturity journey.`,
    );
  return value;
}
function stagingOrigin(raw: string, hostname: string): string {
  const url = new URL(raw);
  if (
    url.protocol !== "https:" ||
    url.hostname !== hostname ||
    url.port ||
    url.username ||
    url.password ||
    url.pathname !== "/" ||
    url.search ||
    url.hash
  ) {
    throw new Error(
      `The maturity journey only accepts the HTTPS ${hostname} origin.`,
    );
  }
  return url.origin;
}
// Node fetch deliberately avoids Playwright APIRequestContext diagnostic steps,
// which retain authenticated request headers in failed-test HTML/JSON reports.
// The public journey validates exact staging origins before calling this helper.
export async function safeStagingRequest(
  context: Pick<BrowserContext, "cookies">,
  apiOrigin: string,
  rawUrl: string,
  method: "GET" | "DELETE" = "GET",
): Promise<{ status: () => number; json: () => Promise<unknown> }> {
  try {
    const url = new URL(rawUrl);
    if (url.origin !== apiOrigin || !url.pathname.startsWith("/api/v1/")) {
      throw new Error("Request outside the allowed API scope");
    }
    const cookies = (await context.cookies(apiOrigin)).filter(
      (cookie) =>
        cookie.name === "vocanova_session" || cookie.name === "vocanova_csrf",
    );
    if (!cookies.some((cookie) => cookie.name === "vocanova_session")) {
      throw new Error("Session cookie missing");
    }
    const requestHeaders: Record<string, string> = {
      Cookie: cookies
        .map((cookie) => `${cookie.name}=${cookie.value}`)
        .join("; "),
    };
    if (method === "DELETE") {
      const csrf = cookies.find(
        (cookie) => cookie.name === "vocanova_csrf",
      )?.value;
      if (!csrf) throw new Error("CSRF cookie missing");
      requestHeaders["X-CSRF-Token"] = csrf;
      requestHeaders["Idempotency-Key"] = randomUUID();
    }
    const response = await fetch(url, {
      method,
      headers: requestHeaders,
      redirect: "error",
      signal: AbortSignal.timeout(20_000),
    });
    const body = await response.text();
    const status = response.status;
    return {
      status: () => status,
      json: async () => {
        try {
          return JSON.parse(body) as unknown;
        } catch {
          throw new Error("The staging API returned unreadable JSON.");
        }
      },
    };
  } catch {
    // Never attach the original cause, URL, headers, cookie values or response.
    throw new Error("The bounded staging API check failed.");
  }
}

test("connected maturity journey against real staging", async ({
  page,
  context,
  baseURL,
}) => {
  test.skip(
    process.env.E2E_MATURITY_JOURNEY !== "true",
    "Opt-in acceptance: set E2E_MATURITY_JOURNEY=true and mint a fresh synthetic session separately from the core journey.",
  );
  if (!baseURL)
    throw new Error("The staging Playwright config must provide baseURL.");
  const webOrigin = stagingOrigin(baseURL, "staging.vocanova.site");
  const apiOrigin = stagingOrigin(
    process.env.STAGING_API_BASE_URL ?? "https://api-staging.vocanova.site",
    "api-staging.vocanova.site",
  );
  const session = requiredEnv("E2E_SESSION_COOKIE");
  const csrf = requiredEnv("E2E_CSRF_TOKEN");
  const cookieDomain =
    process.env.STAGING_SESSION_COOKIE_DOMAIN ?? ".vocanova.site";
  if (cookieDomain !== ".vocanova.site")
    throw new Error(
      "Staging maturity cookies must use the .vocanova.site parent domain.",
    );
  const expectedEmail =
    process.env.VOCANOVA_SYNTHETIC_SMOKE_TEST_EMAIL ??
    "smoke-test-bot@synthetic.vocanova.invalid";
  if (!expectedEmail.endsWith("@synthetic.vocanova.invalid"))
    throw new Error(
      "Only the designated synthetic .invalid account may run this journey.",
    );
  await context.addCookies([
    {
      name: "vocanova_session",
      value: session,
      domain: cookieDomain,
      path: "/",
      secure: true,
      sameSite: "Lax",
    },
    {
      name: "vocanova_csrf",
      value: csrf,
      domain: cookieDomain,
      path: "/",
      secure: true,
      sameSite: "Lax",
    },
  ]);
  const meResponse = await safeStagingRequest(
    context,
    apiOrigin,
    `${apiOrigin}/api/v1/me`,
  );
  expect(
    meResponse.status(),
    "A fresh minted synthetic session is required.",
  ).toBe(200);
  const me = (await meResponse.json()) as CurrentUser;
  expect(
    me.email === expectedEmail,
    "Refuse mutations on any account except the designated synthetic identity.",
  ).toBe(true);
  expect(me.onboardingStatus).toBe("completed");

  // Pass through to the real API. Refuse an incorrectly deployed frontend's
  // API host before it can write outside staging; never submit AI feedback.
  const forbiddenRequests: string[] = [];
  await page.route("**/api/v1/**", async (route) => {
    const request = route.request();
    const url = new URL(request.url());
    if (
      url.origin !== apiOrigin ||
      (request.method() === "POST" &&
        url.pathname === "/api/v1/learner-sentences")
    ) {
      forbiddenRequests.push("outside-staging-or-AI");
      await route.abort("blockedbyclient");
      return;
    }
    await route.continue();
  });
  const main = page.getByRole("main");
  const readMeaning = async () => {
    const response = await safeStagingRequest(
      context,
      apiOrigin,
      `${apiOrigin}/api/v1/canonical-words/menu`,
    );
    expect(response.status()).toBe(200);
    const body = (await response.json()) as WordDetailResponse;
    expect(body.word.text).toBe("menu");
    const meaning = body.word.meanings.find((item) => item.id === MEANING_ID);
    expect(
      meaning,
      "The reviewed menu meaning must exist in actual staging content.",
    ).toBeDefined();
    return meaning!;
  };
  const vocabularyState = (
    meaning: Awaited<ReturnType<typeof readMeaning>>,
  ) => ({
    saved: meaning.saved,
    selfReportedKnown: meaning.selfReportedKnown,
    userWordId: meaning.userWordId ?? null,
    reviewState: meaning.reviewState ?? null,
  });
  const before = vocabularyState(await readMeaning());
  const listName = `Staging maturity ${randomUUID()}`;
  let ownedListId: string | null = null;
  // Capture the created identity before receiving the response so finally can
  // remove our own list even if its response or subsequent page navigation fails.
  page.on("request", (request) => {
    const url = new URL(request.url());
    if (url.origin !== apiOrigin || request.method() !== "PUT") return;
    const match = url.pathname.match(
      /^\/api\/v1\/word-lists\/([0-9a-f-]{36})$/u,
    );
    if (match && request.postDataJSON()?.name === listName)
      ownedListId = match[1]!;
  });
  try {
    await test.step("create one unique personal list and add the actual menu meaning", async () => {
      await page.goto(`${webOrigin}/lists`);
      await main.getByLabel("List name", { exact: true }).fill(listName);
      const createdResponse = page.waitForResponse(
        (response) =>
          response.request().method() === "PUT" &&
          /^\/api\/v1\/word-lists\/[0-9a-f-]{36}$/u.test(
            new URL(response.url()).pathname,
          ),
      );
      await main
        .getByRole("button", { name: "Create list", exact: true })
        .click();
      const created = await createdResponse;
      expect(created.status()).toBe(200);
      const list = (await created.json()) as WordListDetail;
      expect(list.id).toBe(ownedListId);
      expect(list.name).toBe(listName);
      expect(list.memberCount).toBe(0);
      await expect(page).toHaveURL(`${webOrigin}/lists/${list.id}`);

      await page.goto(`${webOrigin}/vocabulary/menu#meaning-${MEANING_ID}`);
      const meaning = main.locator(`[id="meaning-${MEANING_ID}"]`);
      await meaning
        .getByRole("button", { name: "Choose personal lists", exact: true })
        .click();
      await meaning
        .getByLabel("Choose a list for this meaning", { exact: true })
        .selectOption(list.id);
      const addedResponse = page.waitForResponse(
        (response) =>
          response.request().method() === "PUT" &&
          new URL(response.url()).pathname ===
            `/api/v1/word-lists/${list.id}/members/${MEANING_ID}`,
      );
      await meaning
        .getByRole("button", { name: "Add meaning to list", exact: true })
        .click();
      const added = await addedResponse;
      expect(added.status()).toBe(200);
      const populated = (await added.json()) as WordListDetail;
      expect(populated.members.map((item) => item.meaningId)).toEqual([
        MEANING_ID,
      ]);
      expect(populated.usableMemberCount).toBe(1);
      expect(vocabularyState(await readMeaning())).toEqual(before);
    });

    await test.step("list practice confirms the exact snapshot and reloads a server-graded answer", async () => {
      if (!ownedListId) throw new Error("The test-owned list was not created.");
      const listResponse = await safeStagingRequest(
        context,
        apiOrigin,
        `${apiOrigin}/api/v1/word-lists/${ownedListId}`,
      );
      expect(listResponse.status()).toBe(200);
      const list = (await listResponse.json()) as WordListDetail;
      await page.goto(`${webOrigin}/lists/${list.id}`);
      await expect(
        main.getByRole("heading", { name: listName, exact: true }),
      ).toBeVisible();
      await expect(
        main.getByText(
          "List membership does not save words for scheduled review or change what you marked as known.",
          { exact: true },
        ),
      ).toBeVisible();
      await main
        .getByRole("link", { name: "Practice this list", exact: true })
        .click();
      await expect(
        main.getByLabel("Practice vocabulary", { exact: true }),
      ).toHaveValue(`list:${list.id}`);
      const startedResponse = page.waitForResponse(
        (response) =>
          response.request().method() === "POST" &&
          new URL(response.url()).pathname === "/api/v1/practice-sessions",
      );
      await main
        .getByRole("button", { name: "Start typed recall", exact: true })
        .click();
      const started = await startedResponse;
      expect(started.status()).toBe(200);
      const practice = (await started.json()) as PracticeSession;
      expect(practice).toMatchObject({
        listId: list.id,
        listName,
        listRevision: list.revision,
        totalSteps: 1,
        revision: 0,
      });
      await expect(page).toHaveURL(
        `${webOrigin}/practice/session/${practice.id}`,
      );
      await main
        .getByRole("textbox", { name: "Your answer", exact: true })
        .fill("menu");
      const answerResponse = page.waitForResponse(
        (response) =>
          response.request().method() === "POST" &&
          new URL(response.url()).pathname ===
            `/api/v1/practice-sessions/${practice.id}/actions`,
      );
      await main
        .getByRole("button", { name: "Check answer", exact: true })
        .click();
      const answered = await answerResponse;
      expect(answered.status()).toBe(200);
      const checked = (await answered.json()) as PracticeSession;
      expect(checked.feedback?.correct).toBe(true);
      expect(checked.revision).toBe(1);
      await page.reload();
      await expect(
        main.getByRole("heading", { name: "You remembered it", exact: true }),
      ).toBeVisible();
      const persistedResponse = await safeStagingRequest(
        context,
        apiOrigin,
        `${apiOrigin}/api/v1/practice-sessions/${practice.id}`,
      );
      expect(persistedResponse.status()).toBe(200);
      expect(await persistedResponse.json()).toMatchObject({
        id: practice.id,
        revision: checked.revision,
        feedback: { correct: true },
      });
      await main
        .getByRole("button", { name: "Finish practice", exact: true })
        .click();
      await expect(
        main.getByRole("heading", { name: "Practice complete", exact: true }),
      ).toBeVisible();
      expect(vocabularyState(await readMeaning())).toEqual(before);
    });

    await test.step("read, start, grade and resume a new original story without replacing history", async () => {
      await page.goto(`${webOrigin}/stories/${STORY_KEY}`);
      await expect(
        main.getByRole("heading", { name: "Full story", exact: true }),
      ).toBeVisible();
      await expect(
        main.getByRole("button", { name: /^Play audio: line/ }),
      ).toHaveCount(8);
      const startedResponse = page.waitForResponse(
        (response) =>
          response.request().method() === "POST" &&
          new URL(response.url()).pathname === "/api/v1/story-sessions",
      );
      await main
        .getByRole("button", { name: "Start story", exact: true })
        .click();
      const started = await startedResponse;
      expect(started.status()).toBe(200);
      let story = (await started.json()) as StorySession;
      expect(story).toMatchObject({
        storyKey: STORY_KEY,
        revision: 0,
        totalSteps: 10,
        status: "in_progress",
      });
      await expect(page).toHaveURL(`${webOrigin}/stories/session/${story.id}`);
      for (let line = 0; line < 4; line++) {
        const continuedResponse = page.waitForResponse(
          (response) =>
            response.request().method() === "POST" &&
            new URL(response.url()).pathname ===
              `/api/v1/story-sessions/${story.id}/actions`,
        );
        await main
          .getByRole("button", { name: "Continue", exact: true })
          .click();
        const continued = await continuedResponse;
        expect(continued.status()).toBe(200);
        story = (await continued.json()) as StorySession;
        await expect(
          main.getByText(`${story.completedSteps} of 10 steps completed`, {
            exact: true,
          }),
        ).toBeVisible();
      }
      expect(story.currentStep?.kind).toBe("comprehension");
      await main
        .getByRole("radio", { name: "She has a meeting at two.", exact: true })
        .check();
      const answerResponse = page.waitForResponse(
        (response) =>
          response.request().method() === "POST" &&
          new URL(response.url()).pathname ===
            `/api/v1/story-sessions/${story.id}/actions`,
      );
      await main
        .getByRole("button", { name: "Check answer", exact: true })
        .click();
      const answered = await answerResponse;
      expect(answered.status()).toBe(200);
      story = (await answered.json()) as StorySession;
      expect(story).toMatchObject({
        revision: 5,
        questionsAnswered: 1,
        feedback: { correct: true },
      });
      await page.reload();
      await expect(
        main.getByText("That's right.", { exact: true }),
      ).toBeVisible();
      const savedResponse = await safeStagingRequest(
        context,
        apiOrigin,
        `${apiOrigin}/api/v1/story-sessions/${story.id}`,
      );
      expect(savedResponse.status()).toBe(200);
      expect(await savedResponse.json()).toMatchObject({
        id: story.id,
        revision: story.revision,
        feedback: { correct: true },
      });
      await main
        .getByRole("link", { name: "Back to Stories", exact: true })
        .click();
      const card = main.getByRole("region", {
        name: "A quiet lunch",
        exact: true,
      });
      await expect(
        card.getByRole("link", { name: "Continue story", exact: true }),
      ).toHaveAttribute("href", `/stories/session/${story.id}`);
      await card
        .getByRole("link", { name: "Continue story", exact: true })
        .click();
      await expect(page).toHaveURL(`${webOrigin}/stories/session/${story.id}`);
      await expect(
        main.getByText("That's right.", { exact: true }),
      ).toBeVisible();
    });

    await test.step("open the matching unit guide and original topic prompt without AI submission", async () => {
      await page.goto(`${webOrigin}/stories/${STORY_KEY}`);
      await main
        .getByRole("link", { name: "Open the situation guide", exact: true })
        .click();
      await expect(page).toHaveURL(
        `${webOrigin}/discover/restaurant#unit-guide`,
      );
      const guide = main.getByRole("region", {
        name: "A quick guide",
        exact: true,
      });
      await expect(guide).toBeVisible();
      await guide.getByText("Useful phrases", { exact: true }).click();
      await expect(
        guide.getByText("Could I have the menu, please?", { exact: true }),
      ).toBeVisible();
      await guide
        .getByRole("link", { name: "Write about this situation", exact: true })
        .click();
      await expect(page).toHaveURL(`${webOrigin}/writing?situation=restaurant`);
      await expect(
        main.getByRole("heading", { name: "Topic writing", exact: true }),
      ).toBeVisible();
      await expect(
        main.getByText(
          "You are eating out. Write one sentence asking for something you need or explaining your order.",
          { exact: true },
        ),
      ).toBeVisible();
      await main.getByRole("link", { name: "menu", exact: true }).click();
      await expect(page).toHaveURL(
        `${webOrigin}/writing?${new URLSearchParams({ situation: "restaurant", meaning: MEANING_ID })}`,
      );
      await expect(
        main.getByRole("heading", { name: "Write with menu", exact: true }),
      ).toBeVisible();
      // Open the existing saved-word editor or its honest save-first gate. Never
      // alter the pre-existing synthetic account's vocabulary to satisfy the test.
      if (before.saved)
        await expect(
          main.getByRole("textbox", {
            name: "Write a sentence using menu",
            exact: true,
          }),
        ).toBeVisible();
      else
        await expect(
          main.getByText(
            "Save this meaning first to unlock writing feedback.",
            { exact: false },
          ),
        ).toBeVisible();
      expect(vocabularyState(await readMeaning())).toEqual(before);
    });
    expect(
      forbiddenRequests,
      "No outside-staging API or AI feedback requests are permitted.",
    ).toEqual([]);
  } finally {
    if (ownedListId) {
      const current = await safeStagingRequest(
        context,
        apiOrigin,
        `${apiOrigin}/api/v1/word-lists/${ownedListId}`,
      );
      if (current.status() !== 404) {
        expect(
          current.status(),
          "Read only this test-owned list for cleanup.",
        ).toBe(200);
        const list = (await current.json()) as WordListDetail;
        expect(list.id).toBe(ownedListId);
        expect(list.name).toBe(listName);
        const removed = await safeStagingRequest(
          context,
          apiOrigin,
          `${apiOrigin}/api/v1/word-lists/${ownedListId}?expectedRevision=${list.revision}`,
          "DELETE",
        );
        expect(
          removed.status(),
          "Only the uniquely owned list should be deleted.",
        ).toBe(204);
      }
    }
  }
});
