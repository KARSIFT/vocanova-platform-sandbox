import { randomUUID } from "node:crypto";

import {
  expect,
  test,
  type Locator,
  type Page,
  type TestInfo,
} from "@playwright/test";

import { scanForAxeViolations } from "./axe-helper";
import { verifyContextPracticeJourney } from "./context-practice-journey";

const SITUATION_PATH = "/discover/daily-conversation";
const PRACTICE_NAME = "Choose the word for the situation";
const CASES = [
  {
    title: "Ask a friend to dinner",
    correct: /^invite$/i,
    other: /^join$/i,
    explanation: "Invite is what you do when you ask Sam to come.",
    otherExplanation: "Join means taking part with other people.",
    word: "invite",
    slug: "invite",
  },
  {
    title: "Respond to an idea",
    correct: /^sounds good!?$/i,
    other: /^keep in touch!?$/i,
    explanation: "Sounds good is a friendly way to say you like the idea.",
    otherExplanation:
      "Keep in touch means continuing to contact someone over time.",
    word: "sounds good",
    slug: "sounds-good",
  },
  {
    title: "Move lunch to another day",
    correct: /^reschedule$/i,
    other: /^cancel$/i,
    explanation:
      "Reschedule means moving a planned event to a different time or day.",
    otherExplanation: "Cancel means stopping a plan from going ahead.",
    word: "reschedule",
    slug: "reschedule",
  },
];

test.beforeEach(async ({ context, baseURL }) => {
  if (!baseURL) throw new Error("A test app URL is required");
  await context.addCookies([
    { name: "vocanova_session", value: randomUUID(), url: baseURL },
    { name: "vocanova_csrf", value: randomUUID(), url: baseURL },
    { name: "e2e_daily_conversation", value: "true", url: baseURL },
  ]);
});

test("shared deployed context journey works against the local canonical fixture", async ({
  page,
}) => {
  await verifyContextPracticeJourney(page);
});

function collectMutations(page: Page) {
  const mutations: Array<{ method: string; path: string; body: unknown }> = [];
  page.on("request", (request) => {
    const path = new URL(request.url()).pathname;
    if (
      path.startsWith("/api/v1/") &&
      ["POST", "PATCH", "DELETE"].includes(request.method())
    ) {
      mutations.push({
        method: request.method(),
        path,
        body: request.postDataJSON(),
      });
    }
  });
  return mutations;
}

async function inspectAndCapture(
  page: Page,
  region: Locator,
  testInfo: TestInfo,
  name: string,
) {
  expect(
    await page.evaluate(() => document.documentElement.scrollWidth),
  ).toBeLessThanOrEqual(page.viewportSize()!.width);
  for (const control of await region
    .getByRole("button")
    .or(region.getByRole("link"))
    .all()) {
    const box = await control.boundingBox();
    expect(box?.height).toBeGreaterThanOrEqual(44);
    expect(box?.width).toBeGreaterThanOrEqual(44);
  }
  const { criticalOrSerious } = await scanForAxeViolations(page);
  expect(criticalOrSerious).toEqual([]);
  await region.screenshot({ path: testInfo.outputPath(`${name}.png`) });
}

async function completePractice(region: Locator) {
  await region.getByRole("button", { name: "Start context practice" }).click();
  for (const [index, item] of CASES.entries()) {
    await expect(
      region.getByRole("heading", { name: item.title }),
    ).toBeVisible();
    await region.getByRole("button", { name: item.correct }).click();
    await region
      .getByRole("button", {
        name: index === CASES.length - 1 ? "Finish practice" : "Next example",
      })
      .click();
  }
  await expect(
    region.getByRole("heading", { name: "You’ve explored three situations" }),
  ).toBeVisible();
}

for (const theme of ["light", "dark"] as const) {
  test(`context choices explain the meaning and remain local in ${theme} mode`, async ({
    page,
  }, testInfo) => {
    test.setTimeout(60_000);
    await page.emulateMedia({ colorScheme: theme });
    const mutations = collectMutations(page);
    await page.goto(SITUATION_PATH);
    await expect(page.locator("html")).toHaveAttribute("data-theme", theme);
    const region = page.getByRole("region", { name: PRACTICE_NAME });
    await expect(region).toBeVisible();
    await expect(
      page.getByText("0 of 18 words saved", { exact: true }),
    ).toBeVisible();
    await inspectAndCapture(page, region, testInfo, `context-intro-${theme}`);

    const start = region.getByRole("button", {
      name: "Start context practice",
    });
    await start.focus();
    await page.keyboard.press("Enter");

    for (const [index, item] of CASES.entries()) {
      const heading = region.getByRole("heading", { name: item.title });
      await expect(heading).toBeFocused();
      await expect(region).toContainText(`Example ${index + 1} of 3`);
      await expect(region).not.toContainText(item.explanation);
      await expect(region).not.toContainText(item.otherExplanation);
      const wrong = region.getByRole("button", { name: item.other });
      await page.keyboard.press("Tab");
      await expect(
        region.getByRole("group").getByRole("button").first(),
      ).toBeFocused();
      if (index !== 1) await page.keyboard.press("Tab");
      await expect(wrong).toBeFocused();
      await page.keyboard.press("Enter");
      await expect(region.getByRole("status")).toContainText(
        "Not quite. Compare the two choices.",
      );
      await expect(region).toContainText(item.explanation);
      await expect(region).toContainText(item.otherExplanation);
      await expect(
        region.getByRole("button", { name: "Next example" }),
      ).toHaveCount(0);
      await expect(
        region.getByRole("button", { name: "Finish practice" }),
      ).toHaveCount(0);
      if (index === 0) {
        await inspectAndCapture(
          page,
          region,
          testInfo,
          `context-explanation-${theme}`,
        );
      }

      const retry = region.getByRole("button", { name: "Try again" });
      // Answer buttons become disabled after selection. The next Tab must
      // still reach the recovery action without manually moving focus.
      await page.keyboard.press("Tab");
      await expect(retry).toBeFocused();
      await page.keyboard.press("Enter");
      await expect(region).not.toContainText(item.explanation);
      await expect(region).not.toContainText(item.otherExplanation);
      await expect(
        region.getByRole("group").getByRole("button").first(),
      ).toBeFocused();
      if (index === 1) await page.keyboard.press("Tab");
      await expect(
        region.getByRole("button", { name: item.correct }),
      ).toBeFocused();
      await page.keyboard.press("Enter");
      await expect(region.getByRole("status")).toContainText(
        "That fits this situation.",
      );
      await expect(region).toContainText(item.explanation);
      const next = region.getByRole("button", {
        name: index === CASES.length - 1 ? "Finish practice" : "Next example",
      });
      await page.keyboard.press("Tab");
      await expect(next).toBeFocused();
      await page.keyboard.press("Enter");
    }

    await expect(
      region.getByRole("heading", { name: "You’ve explored three situations" }),
    ).toBeFocused();
    for (const item of CASES) {
      await expect(
        region.getByRole("link", { name: `Practice with ${item.word}` }),
      ).toHaveAttribute("href", `${SITUATION_PATH}/${item.slug}`);
    }
    await inspectAndCapture(
      page,
      region,
      testInfo,
      `context-complete-${theme}`,
    );
    expect(mutations).toEqual([]);

    await region.getByRole("button", { name: "Restart practice" }).click();
    await expect(
      region.getByRole("heading", { name: CASES[0]!.title }),
    ).toBeFocused();
    await expect(region).toContainText("Example 1 of 3");
    await expect(region).not.toContainText(CASES[0]!.explanation);
    await region.getByRole("button", { name: "Exit to words" }).click();
    await expect(
      page.getByRole("list", { name: "Words in this situation" }),
    ).toBeFocused();
    await expect(start).toBeVisible();
    await expect(
      page.getByText("0 of 18 words saved", { exact: true }),
    ).toBeVisible();
    expect(mutations).toEqual([]);
  });
}

test("finishing practice leads to an explicit save and existing sentence feedback", async ({
  page,
}) => {
  const mutations = collectMutations(page);
  await page.goto(SITUATION_PATH);
  const region = page.getByRole("region", { name: PRACTICE_NAME });
  await completePractice(region);
  expect(mutations).toEqual([]);
  await region.getByRole("link", { name: "Practice with sounds good" }).click();
  await expect(page).toHaveURL(/\/discover\/daily-conversation\/sounds-good$/);
  await expect(
    page.getByRole("heading", { level: 1, name: "sounds good" }),
  ).toBeVisible();
  await expect(
    page.getByRole("textbox", { name: /Write a sentence using sounds good/ }),
  ).toHaveCount(0);
  expect(mutations).toEqual([]);

  await page.getByRole("button", { name: /^Save sounds good:/ }).click();
  const input = page.getByRole("textbox", {
    name: /Write a sentence using sounds good/,
  });
  await expect(input).toBeVisible();
  expect(mutations).toHaveLength(1);
  expect(mutations[0]).toMatchObject({
    method: "POST",
    path: "/api/v1/user-words",
    body: {
      meaningId: "52b8ebc0-ba65-579d-8749-6ff51b9ce6b5",
      source: "journey",
    },
  });
  await input.fill("That sounds good to me.");
  await page.getByRole("button", { name: "Check my sentence" }).click();
  await expect(
    page.getByRole("status", { name: "Feedback result: Correct" }),
  ).toBeVisible();
  expect(mutations).toHaveLength(2);
  expect(mutations[1]).toMatchObject({
    method: "POST",
    path: "/api/v1/learner-sentences",
    body: { sentenceText: "That sounds good to me.", source: "word_detail" },
  });
});

test("reloading a partially answered activity returns to its local starting state", async ({
  page,
}) => {
  const mutations = collectMutations(page);
  await page.goto(SITUATION_PATH);
  const region = page.getByRole("region", { name: PRACTICE_NAME });
  await region.getByRole("button", { name: "Start context practice" }).click();
  await region.getByRole("button", { name: /^invite$/i }).click();
  await region.getByRole("button", { name: "Next example" }).click();
  await expect(
    region.getByRole("heading", { name: CASES[1]!.title }),
  ).toBeVisible();
  await page.reload();
  await expect(
    region.getByRole("button", { name: "Start context practice" }),
  ).toBeVisible();
  await expect(
    region.getByRole("heading", { name: CASES[1]!.title }),
  ).toHaveCount(0);
  await region.getByRole("button", { name: "Start context practice" }).click();
  await expect(
    region.getByRole("heading", { name: CASES[0]!.title }),
  ).toBeVisible();
  await expect(region).not.toContainText(CASES[0]!.explanation);
  expect(mutations).toEqual([]);
});

test("keeps ordinary browsing available when required content is missing and on other situations", async ({
  page,
  context,
  baseURL,
}) => {
  if (!baseURL) throw new Error("A test app URL is required");
  await context.addCookies([
    {
      name: "e2e_context_practice_missing_meaning",
      value: "true",
      url: baseURL,
    },
  ]);
  await page.goto(SITUATION_PATH);
  await expect(
    page.getByRole("heading", { level: 1, name: "Daily Conversation" }),
  ).toBeVisible();
  await expect(page.getByRole("region", { name: PRACTICE_NAME })).toHaveCount(
    0,
  );
  await expect(
    page.getByRole("button", { name: "Start context practice" }),
  ).toHaveCount(0);
  await expect(
    page
      .getByRole("list", { name: "Words in this situation" })
      .getByRole("listitem"),
  ).toHaveCount(17);
  await expect(
    page.getByText("0 of 17 words saved", { exact: true }),
  ).toBeVisible();
  await page.getByRole("link", { name: /^sounds good / }).click();
  await expect(
    page.getByRole("heading", { level: 1, name: "sounds good" }),
  ).toBeVisible();
  await expect(
    page.getByRole("button", { name: /^Save sounds good:/ }),
  ).toBeVisible();

  await page.goto("/discover/ordering-at-a-cafe");
  await expect(
    page.getByRole("heading", { level: 1, name: "Ordering at a cafe" }),
  ).toBeVisible();
  await expect(page.getByRole("region", { name: PRACTICE_NAME })).toHaveCount(
    0,
  );
  await expect(
    page.getByRole("button", { name: "Start context practice" }),
  ).toHaveCount(0);
});
