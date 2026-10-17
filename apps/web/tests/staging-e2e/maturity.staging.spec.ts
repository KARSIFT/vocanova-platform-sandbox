// Real-staging acceptance for connected learning and local dictionary search.
// Explicitly opt in and run THIS spec with a fresh synthetic session. The core
// journey logs out, so its already-revoked token cannot be reused here.
// No mock server, AI feedback request, vocabulary reset or session-history purge.
import { randomUUID } from "node:crypto";
import { writeFile } from "node:fs/promises";
import { expect, test, type BrowserContext } from "@playwright/test";
import type {
  CurrentUser,
  DailyMission,
  DictionaryEntry,
  PracticeSession,
  PracticeSessionsResponse,
  Progress,
  PronunciationAudio,
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
  method: "GET" | "DELETE" | "POST" = "GET",
  body?: { text: "invite" },
): Promise<{ status: () => number; json: () => Promise<unknown> }> {
  try {
    const url = new URL(rawUrl);
    if (url.origin !== apiOrigin || !url.pathname.startsWith("/api/v1/")) {
      throw new Error("Request outside the allowed API scope");
    }
    if (
      (method === "POST" &&
        (url.pathname !== "/api/v1/pronunciation" ||
          body?.text !== "invite")) ||
      (method !== "POST" && body !== undefined)
    ) {
      throw new Error("Request outside the bounded pronunciation acceptance");
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
    if (method === "DELETE" || method === "POST") {
      const csrf = cookies.find(
        (cookie) => cookie.name === "vocanova_csrf",
      )?.value;
      if (!csrf) throw new Error("CSRF cookie missing");
      requestHeaders["X-CSRF-Token"] = csrf;
      if (method === "DELETE") requestHeaders["Idempotency-Key"] = randomUUID();
      if (method === "POST")
        requestHeaders["Content-Type"] = "application/json";
    }
    const response = await fetch(url, {
      method,
      headers: requestHeaders,
      ...(body ? { body: JSON.stringify(body) } : {}),
      redirect: "error",
      signal: AbortSignal.timeout(20_000),
    });
    // Bound the public clip's JSON before parsing; do not retain raw response
    // data in Playwright request diagnostics or error attachments.
    const reader = response.body?.getReader();
    const chunks: Uint8Array[] = [];
    let bytes = 0;
    if (reader) {
      try {
        while (true) {
          const next = await reader.read();
          if (next.done) break;
          bytes += next.value.byteLength;
          if (bytes > 1_500_000) throw new Error("Response exceeded the bound");
          chunks.push(next.value);
        }
      } catch {
        await reader.cancel();
        throw new Error("Unreadable bounded response");
      } finally {
        reader.releaseLock();
      }
    }
    const responseBody = Buffer.concat(chunks).toString("utf8");
    const status = response.status;
    return {
      status: () => status,
      json: async () => {
        try {
          return JSON.parse(responseBody) as unknown;
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
    await test.step("search beyond lesson words using the real local dictionary without saving", async () => {
      await page.goto(`${webOrigin}/home`);
      const search = page.getByRole("search", {
        name: "Word search",
        exact: true,
      });
      await search
        .getByRole("searchbox", {
          name: "Search any English word",
          exact: true,
        })
        .fill("serendipity");
      await search
        .getByRole("button", { name: "Search words", exact: true })
        .click();
      await expect(page).toHaveURL(`${webOrigin}/vocabulary?q=serendipity`);
      const result = main.getByRole("region", {
        name: "Dictionary result",
        exact: true,
      });
      await expect(result).toBeVisible();
      await expect(
        result.getByRole("heading", { name: "serendipity", exact: true }),
      ).toBeVisible();
      await expect(
        result.getByText(
          "good luck in making unexpected and fortunate discoveries",
          { exact: true },
        ),
      ).toBeVisible();
      await expect(result.getByRole("button", { name: /save/i })).toHaveCount(
        0,
      );
      await result.getByText("Dictionary source", { exact: true }).click();
      await expect(
        result.getByRole("link", { name: "WordNet 3.0", exact: true }),
      ).toHaveAttribute("href", "https://wordnet.princeton.edu/");
      await expect(
        result.getByText(/WordNet 3\.0 Copyright 2006 by Princeton University/),
      ).toBeVisible();

      // Read the deployed API through the redacted helper, never Playwright's
      // authenticated request diagnostics or a raw response-body assertion.
      const response = await safeStagingRequest(
        context,
        apiOrigin,
        `${apiOrigin}/api/v1/dictionary?q=serendipity`,
      );
      expect(response.status()).toBe(200);
      const entry = (await response.json()) as DictionaryEntry;
      expect(entry.word).toBe("serendipity");
      expect(entry.attribution.provider).toBe("WordNet 3.0");
      expect(
        entry.attribution.licenses.some((license) =>
          license.text.includes(
            "WordNet 3.0 Copyright 2006 by Princeton University",
          ),
        ),
        "The deployed dictionary must preserve the original source notice.",
      ).toBe(true);
    });

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
      await meaning.getByText("Personal tools", { exact: true }).click();
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
      const practiceLink = main.getByRole("link", {
        name: "Practice this list",
        exact: true,
      });
      await expect(practiceLink).toHaveAttribute(
        "href",
        `/practice?list=${list.id}`,
      );
      await practiceLink.click();
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

    await test.step("complete and resume real mixed personal practice without review credit", async () => {
      const readPractice = async () => {
        const response = await safeStagingRequest(
          context,
          apiOrigin,
          `${apiOrigin}/api/v1/practice-sessions`,
        );
        expect(response.status()).toBe(200);
        const result = (await response.json()) as PracticeSessionsResponse;
        expect(result.journey).toBeDefined();
        return result.journey!;
      };
      const readCredit = async () => {
        const missionResponse = await safeStagingRequest(
          context,
          apiOrigin,
          `${apiOrigin}/api/v1/daily-mission`,
        );
        const progressResponse = await safeStagingRequest(
          context,
          apiOrigin,
          `${apiOrigin}/api/v1/progress`,
        );
        expect(missionResponse.status()).toBe(200);
        expect(progressResponse.status()).toBe(200);
        return {
          mission: (await missionResponse.json()) as DailyMission,
          progress: (await progressResponse.json()) as Progress,
        };
      };
      const baseline = await readPractice();
      const creditBefore = await readCredit();
      const seenKinds = new Set<string>();
      let finishedLessons = 0;
      // An existing lesson may have only one remaining step. At most one new
      // lesson then establishes mixed formats without assuming its vocabulary,
      // a fixed lesson length, a zero history or an empty active session.
      for (let lesson = 0; lesson < 2 && seenKinds.size < 2; lesson++) {
        const currentPath = await readPractice();
        await page.goto(`${webOrigin}/practice`);
        const path = main.getByRole("region", {
          name: "Your learning path",
          exact: true,
        });
        await expect(path).toBeVisible();
        let current: PracticeSession;
        if (currentPath.currentSession) {
          const resume = path.getByRole("link", {
            name: "Continue practice",
            exact: true,
          });
          await expect(resume).toHaveAttribute(
            "href",
            `/practice/session/${currentPath.currentSession.id}`,
          );
          await resume.click();
          const response = await safeStagingRequest(
            context,
            apiOrigin,
            `${apiOrigin}/api/v1/practice-sessions/${currentPath.currentSession.id}`,
          );
          expect(response.status()).toBe(200);
          current = (await response.json()) as PracticeSession;
        } else {
          const startedResponse = page.waitForResponse(
            (response) =>
              response.request().method() === "POST" &&
              new URL(response.url()).origin === apiOrigin &&
              new URL(response.url()).pathname === "/api/v1/practice-sessions",
          );
          await path
            .getByRole("button", { name: "Start practice lesson", exact: true })
            .click();
          const response = await startedResponse;
          expect(response.status()).toBe(200);
          current = (await response.json()) as PracticeSession;
        }
        expect(current.mode).toBe("journey");
        expect(current.status).toBe("in_progress");
        const sessionId = current.id;
        await expect(page).toHaveURL(
          `${webOrigin}/practice/session/${sessionId}`,
        );
        const maximumActions =
          (current.totalSteps - current.completedSteps) * 3 + 4;
        let reloadedProgress = false;
        const act = async (button: ReturnType<typeof main.getByRole>) => {
          const previous = current;
          const actionResponse = page.waitForResponse(
            (response) =>
              response.request().method() === "POST" &&
              new URL(response.url()).origin === apiOrigin &&
              new URL(response.url()).pathname ===
                `/api/v1/practice-sessions/${sessionId}/actions`,
          );
          await button.click();
          const response = await actionResponse;
          expect(response.status()).toBe(200);
          current = (await response.json()) as PracticeSession;
          expect(current.id).toBe(sessionId);
          expect(current.revision).toBe(previous.revision + 1);
          expect(current.totalSteps).toBe(previous.totalSteps);
          expect(current.completedSteps).toBeGreaterThanOrEqual(
            previous.completedSteps,
          );
          if (!reloadedProgress) {
            await page.reload();
            const saved = await safeStagingRequest(
              context,
              apiOrigin,
              `${apiOrigin}/api/v1/practice-sessions/${sessionId}`,
            );
            expect(saved.status()).toBe(200);
            expect(await saved.json()).toMatchObject({
              id: sessionId,
              revision: current.revision,
              completedSteps: current.completedSteps,
              status: current.status,
              feedback: current.feedback,
            });
            await expect(
              main.getByRole("progressbar", {
                name: "Practice progress",
                exact: true,
              }),
            ).toHaveAttribute("value", String(current.completedSteps));
            reloadedProgress = true;
          }
          return current;
        };
        let actions = 0;
        while (current.status !== "completed" && actions < maximumActions) {
          const step = current.currentStep;
          expect(step).not.toBeNull();
          if (!step) throw new Error("A saved practice step is missing.");
          await expect(
            main.getByRole("progressbar", {
              name: "Practice progress",
              exact: true,
            }),
          ).toHaveAttribute("value", String(current.completedSteps));
          const taskHeading =
            step.kind === "teach" ? step.wordText : step.prompt;
          if (!taskHeading)
            throw new Error("The saved practice task has no heading.");
          await expect(
            main.getByRole("heading", {
              level: 1,
              name: taskHeading,
              exact: true,
            }),
          ).toBeVisible();
          seenKinds.add(step.kind);
          actions++;
          if (current.canContinue) {
            await act(
              main.getByRole("button", {
                name:
                  current.completedSteps + 1 === current.totalSteps
                    ? "Finish practice"
                    : "Continue",
                exact: true,
              }),
            );
          } else if (current.feedback) {
            await act(
              main.getByRole("button", {
                name: "Continue with help",
                exact: true,
              }),
            );
            expect(current.feedback?.assisted).toBe(true);
            expect(current.feedback?.correct).toBe(false);
            expect(current.canContinue).toBe(true);
          } else if (step.kind === "listening_choice") {
            // This acceptance does not pretend that a guessed listening answer
            // is hearing comprehension or request extra provider clips.
            const checked = await act(
              main.getByRole("button", {
                name: "Can’t listen now",
                exact: true,
              }),
            );
            expect(checked.feedback?.assisted).toBe(true);
            expect(checked.canContinue).toBe(true);
          } else if (
            step.kind === "meaning_choice" ||
            step.kind === "gap_choice"
          ) {
            const choice = step.choices[0];
            if (!choice)
              throw new Error("This practice has no answer choices.");
            await main
              .getByRole("radio", { name: choice.text, exact: true })
              .check();
            const checked = await act(
              main.getByRole("button", { name: "Check answer", exact: true }),
            );
            expect(checked.feedback?.stepId).toBe(step.id);
          } else if (step.kind === "sentence_build") {
            const available = main.getByRole("group", {
              name: "Available words",
              exact: true,
            });
            // Choose the public displayed token order, then accept the real
            // server result; correction uses an explicit assisted action.
            for (const [index, token] of step.choices.entries()) {
              await available
                .getByRole("button", { name: token.text, exact: true })
                .and(available.locator("button:enabled"))
                .first()
                .click();
              await expect(
                main
                  .getByRole("group", { name: "Your sentence", exact: true })
                  .getByRole("button"),
              ).toHaveCount(index + 1);
            }
            const checked = await act(
              main.getByRole("button", { name: "Check answer", exact: true }),
            );
            expect(checked.feedback?.stepId).toBe(step.id);
          } else {
            throw new Error(
              "The active journey has an unsupported task format.",
            );
          }
        }
        expect(current.status).toBe("completed");
        expect(current.completedSteps).toBe(current.totalSteps);
        await expect(
          main.getByRole("heading", { name: "Practice complete", exact: true }),
        ).toBeVisible();
        finishedLessons++;
        const confirmed = await readPractice();
        expect(confirmed.completedLessons).toBe(
          baseline.completedLessons + finishedLessons,
        );
        expect(confirmed.currentSession).toBeNull();
        await main
          .getByRole("link", { name: "Back to your path", exact: true })
          .click();
        await expect(page).toHaveURL(`${webOrigin}/practice`);
      }
      expect(seenKinds.size).toBeGreaterThanOrEqual(2);
      const persisted = await readPractice();
      expect(persisted.completedLessons).toBe(
        baseline.completedLessons + finishedLessons,
      );
      const creditAfter = await readCredit();
      expect(creditAfter.progress.confidencePointsBalance).toBe(
        creditBefore.progress.confidencePointsBalance,
      );
      expect(creditAfter.mission.reviewsCompleted).toBe(
        creditAfter.mission.localDate === creditBefore.mission.localDate
          ? creditBefore.mission.reviewsCompleted
          : 0,
      );
      expect(vocabularyState(await readMeaning())).toEqual(before);
    });

    await test.step("decode actual bounded natural pronunciation and confirm its shared cache", async () => {
      const getClip = async () => {
        const response = await safeStagingRequest(
          context,
          apiOrigin,
          `${apiOrigin}/api/v1/pronunciation`,
          "POST",
          { text: "invite" },
        );
        expect(
          response.status(),
          "Actual staging natural pronunciation must be available.",
        ).toBe(200);
        const clip = (await response.json()) as PronunciationAudio;
        expect(
          clip.voice === "luna",
          "The configured natural voice must match.",
        ).toBe(true);
        expect(
          clip.mimeType === "audio/mpeg",
          "The clip must be MPEG audio.",
        ).toBe(true);
        expect(typeof clip.cacheHit === "boolean").toBe(true);
        expect(typeof clip.audioBase64 === "string").toBe(true);
        expect(
          clip.audioBase64.length > 0 && clip.audioBase64.length <= 1_398_104,
          "The encoded clip must stay within the one MiB audio bound.",
        ).toBe(true);
        const bytes = Buffer.from(clip.audioBase64, "base64");
        expect(
          bytes.toString("base64") === clip.audioBase64,
          "The clip must contain canonical valid base64.",
        ).toBe(true);
        expect(bytes.byteLength >= 3 && bytes.byteLength <= 1_048_576).toBe(
          true,
        );
        expect(
          bytes.subarray(0, 3).toString("ascii") === "ID3" ||
            (bytes[0] === 0xff && (bytes[1]! & 0xe0) === 0xe0),
          "The returned clip must have an MP3 header.",
        ).toBe(true);
        return clip;
      };
      const first = await getClip();
      const second = await getClip();
      expect(
        second.cacheHit,
        "A successful replay must use the durable cache.",
      ).toBe(true);
      expect(
        second.audioBase64 === first.audioBase64,
        "The cached public clip must be byte-identical.",
      ).toBe(true);
      let decoded: { duration: number; sampleRate: number; channels: number };
      try {
        decoded = await page.evaluate(async (encoded) => {
          const bytes = Uint8Array.from(atob(encoded), (value) =>
            value.charCodeAt(0),
          );
          const audio = new AudioContext();
          try {
            const buffer = await audio.decodeAudioData(bytes.buffer);
            return {
              duration: buffer.duration,
              sampleRate: buffer.sampleRate,
              channels: buffer.numberOfChannels,
            };
          } finally {
            await audio.close();
          }
        }, first.audioBase64);
      } catch {
        throw new Error(
          "The actual public MP3 could not be decoded by the real browser.",
        );
      }
      expect(
        Number.isFinite(decoded.duration) && decoded.duration > 0,
        "The real browser must decode a finite nonempty audio duration.",
      ).toBe(true);
      expect(decoded.sampleRate > 0 && decoded.channels > 0).toBe(true);
      // Only the validated public word clip is retained for owner listening.
      // Never attach API envelopes, cookies, headers or authenticated traces.
      const samplePath = test.info().outputPath("natural-pronunciation-invite.mp3");
      await writeFile(samplePath, Buffer.from(first.audioBase64, "base64"));
      await test.info().attach("natural-pronunciation-invite.mp3", {
        path: samplePath,
        contentType: "audio/mpeg",
      });
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
      await page
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
      const guideLink = main.getByRole("link", {
        name: "Open the situation guide",
        exact: true,
      });
      await expect(guideLink).toHaveAttribute(
        "href",
        "/discover/restaurant#unit-guide",
      );
      await guideLink.click();
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
      const writingLink = guide.getByRole("link", {
        name: "Write about this situation",
        exact: true,
      });
      await expect(writingLink).toHaveAttribute(
        "href",
        "/writing?situation=restaurant",
      );
      await writingLink.click();
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
      else {
        await expect(
          main.getByRole("button", { name: "Save and write", exact: true }),
        ).toBeEnabled();
        await expect(
          main.getByRole("textbox", {
            name: "Write a sentence using menu",
            exact: true,
          }),
        ).toHaveCount(0);
      }
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
