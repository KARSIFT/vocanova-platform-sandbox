import { randomUUID } from "node:crypto";
import { test, expect, type Page } from "@playwright/test";
import { scanForAxeViolations } from "./axe-helper";

test.beforeEach(async ({ context, baseURL }) => {
  await context.addCookies([
    { name: "vocanova_session", value: randomUUID(), url: baseURL! },
    { name: "vocanova_csrf", value: randomUUID(), url: baseURL! },
    { name: "e2e_daily_conversation", value: "true", url: baseURL! },
  ]);
});

async function selectInvite(page: Page) {
  await page.goto("/writing?situation=daily-conversation");
  await page.getByRole("link", { name: "invite", exact: true }).click();
  await expect(
    page.getByRole("button", { name: "Save and write", exact: true }),
  ).toBeEnabled();
  return page.url();
}

function saveButton(page: Page) {
  return page.getByRole("button", { name: "Save and write", exact: true });
}
function sentenceBox(page: Page) {
  return page.getByRole("textbox", { name: "Write a sentence using invite" });
}

async function openMeaningChooser(page: Page, selectedWord: string) {
  const summary = page.getByText(`Change meaning: ${selectedWord}`, {
    exact: true,
  });
  const details = summary.locator("..");
  // Native details may keep its open state during a client-side navigation.
  // Open the chooser if needed instead of blindly toggling it closed.
  if ((await details.getAttribute("open")) === null) await summary.click();
  await expect(details).toHaveAttribute("open", "");
}

for (const theme of ["light", "dark"]) {
  test(`inline save opens writing only after confirmation and compares exact edits in ${theme}`, async ({
    page,
    context,
    baseURL,
  }, testInfo) => {
    await context.addCookies([
      { name: "vocanova_theme", value: theme, url: baseURL! },
    ]);
    const writes: string[] = [];
    page.on("request", (request) => {
      if (request.method() === "POST")
        writes.push(new URL(request.url()).pathname);
    });
    const selectedURL = await selectInvite(page);
    expect(writes).toEqual([]);
    await expect(sentenceBox(page)).toHaveCount(0);
    await saveButton(page).focus();
    await page.keyboard.press("Enter");
    await expect(sentenceBox(page)).toBeVisible();
    await expect(sentenceBox(page)).toBeFocused();
    expect(page.url()).toBe(selectedURL);
    expect(writes).toEqual(["/api/v1/user-words"]);
    await sentenceBox(page).fill("i invite you!");
    await page.route("**/api/v1/learner-sentences", async (route) => {
      const response = await route.fetch();
      const data = await response.json();
      await route.fulfill({
        response,
        json: {
          ...data,
          originalSentence: "i invite you!",
          correctedSentence: "I invite you.",
          status: "needs_improvement",
        },
      });
    });
    await page
      .getByRole("button", { name: "Check my sentence", exact: true })
      .click();
    const comparison = page.getByRole("region", {
      name: "Sentence comparison",
    });
    await expect(
      comparison.getByText("i invite you!", { exact: true }),
    ).toBeVisible();
    await expect(
      comparison.getByText("I invite you.", { exact: true }),
    ).toBeVisible();
    await expect(comparison.locator("mark")).toHaveText(["i", "!", "I", "."]);
    expect((await scanForAxeViolations(page)).criticalOrSerious).toEqual([]);
    expect(
      await page.evaluate(
        () => document.documentElement.scrollWidth <= window.innerWidth,
      ),
    ).toBe(true);
    await page.screenshot({
      path: testInfo.outputPath(`writing-comparison-${theme}.png`),
      fullPage: true,
    });
    await comparison
      .getByRole("button", { name: "Revise sentence", exact: true })
      .click();
    await expect(sentenceBox(page)).toBeFocused();
    await expect(sentenceBox(page)).toHaveValue("i invite you!");
    await expect(
      page.getByRole("complementary", { name: "Previous feedback" }),
    ).toContainText("I invite you.");
    await page.reload();
    await expect(sentenceBox(page)).toHaveValue("i invite you!");
    expect(writes.filter((path) => path === "/api/v1/user-words")).toHaveLength(
      1,
    );
  });
}

test("CSRF recovery and a lost save response retain the request identity without unlocking feedback early", async ({
  page,
  context,
}) => {
  const selectedURL = await selectInvite(page);
  const writes: { key: string | undefined; body: string | null }[] = [];
  await page.route("**/api/v1/user-words", async (route) => {
    if (route.request().method() !== "POST") return route.continue();
    writes.push({
      key: route.request().headers()["idempotency-key"],
      body: route.request().postData(),
    });
    if (writes.length === 1) {
      const response = await route.fetch();
      expect(response.ok()).toBe(true);
      return route.abort("failed");
    }
    return route.continue();
  });
  let recoveries = 0;
  await page.route("**/api/v1/me", async (route) => {
    recoveries += 1;
    if (recoveries === 1) return route.abort("failed");
    return route.continue();
  });
  await context.clearCookies({ name: "vocanova_csrf" });
  await saveButton(page).click();
  await expect(page.getByRole("main").getByRole("alert")).toContainText(
    "could not confirm this save",
  );
  expect(writes).toEqual([]);
  await expect(sentenceBox(page)).toHaveCount(0);
  await page.getByRole("button", { name: "Retry save", exact: true }).click();
  await expect(page.getByRole("main").getByRole("alert")).toContainText(
    "could not confirm this save",
  );
  await expect(sentenceBox(page)).toHaveCount(0);
  expect(writes).toHaveLength(1);
  await page.getByRole("button", { name: "Retry save", exact: true }).click();
  await expect(sentenceBox(page)).toBeVisible();
  expect(writes).toHaveLength(2);
  expect(writes[0]!.key).toBeTruthy();
  expect(writes[1]).toEqual(writes[0]);
  expect(recoveries).toBe(2);
  expect(page.url()).toBe(selectedURL);
});

for (const status of [404, 409]) {
  test(`save ${status} reconciles an unsaved meaning without automatically saving again`, async ({
    page,
  }) => {
    const selectedURL = await selectInvite(page);
    const writes: { key: string | undefined; body: string | null }[] = [];
    await page.route("**/api/v1/user-words", async (route) => {
      if (route.request().method() !== "POST") return route.continue();
      writes.push({
        key: route.request().headers()["idempotency-key"],
        body: route.request().postData(),
      });
      if (writes.length === 1)
        return route.fulfill({
          status,
          contentType: "application/problem+json",
          body: JSON.stringify({ detail: "Saved state changed" }),
        });
      return route.continue();
    });
    await saveButton(page).click();
    await expect(page.getByRole("main").getByRole("status")).toContainText(
      "Your earlier request was not reapplied",
    );
    await expect(saveButton(page)).toBeEnabled();
    await expect(sentenceBox(page)).toHaveCount(0);
    expect(writes).toHaveLength(1);
    expect(page.url()).toBe(selectedURL);
    await saveButton(page).click();
    await expect(sentenceBox(page)).toBeVisible();
    expect(writes).toHaveLength(2);
    expect(writes[1]!.key).not.toBe(writes[0]!.key);
    expect(writes[1]!.body).toBe(writes[0]!.body);
  });
}

for (const boundary of ["save response", "canonical meaning"]) {
  test(`a mismatched ${boundary} cannot unlock writing feedback`, async ({
    page,
  }) => {
    await selectInvite(page);
    let writes = 0;
    page.on("request", (request) => {
      if (
        request.method() === "POST" &&
        new URL(request.url()).pathname === "/api/v1/user-words"
      )
        writes += 1;
    });
    if (boundary === "save response") {
      await page.route("**/api/v1/user-words", async (route) => {
        const response = await route.fetch();
        const data = await response.json();
        return route.fulfill({
          response,
          json: { ...data, meaningId: "different-meaning" },
        });
      });
    } else {
      await page.route(
        "**/api/v1/canonical-words/invite",
        async (route) => {
          const response = await route.fetch();
          const data = await response.json();
          const selected = new URL(page.url()).searchParams.get("meaning");
          return route.fulfill({
            response,
            json: {
              ...data,
              word: {
                ...data.word,
                meanings: data.word.meanings.map((meaning: { id: string }) =>
                  meaning.id === selected
                    ? { ...meaning, id: "different-meaning" }
                    : meaning,
                ),
              },
            },
          });
        },
        { times: 1 },
      );
    }
    await saveButton(page).click();
    await expect(page.getByRole("main").getByRole("alert")).toContainText(
      "could not confirm its current status",
    );
    await expect(sentenceBox(page)).toHaveCount(0);
    await expect(saveButton(page)).toHaveCount(0);
    expect(writes).toBe(1);
    await page
      .getByRole("button", { name: "Check saved status", exact: true })
      .click();
    await expect(sentenceBox(page)).toBeVisible();
    expect(writes).toBe(1);
  });
}

test("changing the selected meaning while a save is in flight does not leak feedback or drafts", async ({
  page,
}) => {
  const inviteURL = await selectInvite(page);
  let releaseSave!: () => void;
  const release = new Promise<void>((resolve) => {
    releaseSave = resolve;
  });
  let saveApplied!: () => void;
  const applied = new Promise<void>((resolve) => {
    saveApplied = resolve;
  });
  let saveDelivered!: () => void;
  const delivered = new Promise<void>((resolve) => {
    saveDelivered = resolve;
  });
  await page.route(
    "**/api/v1/user-words",
    async (route) => {
      const response = await route.fetch();
      expect(response.ok()).toBe(true);
      saveApplied();
      await release;
      try {
        await route.fulfill({ response });
      } catch (error) {
        // Changing a route may legitimately cancel its old browser request.
        // The server application above remains confirmed in either outcome.
        if (!route.request().failure()) throw error;
      } finally {
        saveDelivered();
      }
    },
    { times: 1 },
  );
  await saveButton(page).click();
  await applied;
  await openMeaningChooser(page, "invite");
  await page.getByRole("link", { name: "confirm", exact: true }).click();
  await expect(
    page.getByRole("heading", { name: "Write with confirm", exact: true }),
  ).toBeVisible();
  await expect(saveButton(page)).toBeEnabled();
  const confirmURL = page.url();
  expect(confirmURL).not.toBe(inviteURL);
  releaseSave();
  await delivered;
  await expect(sentenceBox(page)).toHaveCount(0);
  await expect(page.getByRole("textbox")).toHaveCount(0);
  expect(page.url()).toBe(confirmURL);
  await saveButton(page).click();
  const confirmBox = page.getByRole("textbox", {
    name: "Write a sentence using confirm",
  });
  await expect(confirmBox).toBeVisible();
  await expect(confirmBox).toHaveValue("");
  await confirmBox.fill("Please confirm our appointment on Friday.");
  await openMeaningChooser(page, "confirm");
  await page.getByRole("link", { name: "invite", exact: true }).click();
  await expect(sentenceBox(page)).toBeVisible();
  await expect(sentenceBox(page)).toHaveValue("");
  await openMeaningChooser(page, "invite");
  await page.getByRole("link", { name: "confirm", exact: true }).click();
  await expect(confirmBox).toHaveValue(
    "Please confirm our appointment on Friday.",
  );
});

test("a save response without current confirmation requires a read, rather than a second save", async ({
  page,
}) => {
  await selectInvite(page);
  const writes: string[] = [];
  page.on("request", (request) => {
    if (request.method() === "POST")
      writes.push(new URL(request.url()).pathname);
  });
  await page.route(
    "**/api/v1/canonical-words/invite",
    (route) => route.abort("failed"),
    { times: 1 },
  );
  await saveButton(page).click();
  await expect(page.getByRole("main").getByRole("alert")).toContainText(
    "could not confirm its current status",
  );
  await expect(sentenceBox(page)).toHaveCount(0);
  await expect(saveButton(page)).toHaveCount(0);
  await page
    .getByRole("button", { name: "Check saved status", exact: true })
    .click();
  await expect(sentenceBox(page)).toBeVisible();
  expect(writes).toEqual(["/api/v1/user-words"]);
});

test("save session expiry returns to the exact selected topic and meaning", async ({
  page,
}) => {
  const selectedURL = await selectInvite(page);
  await page.route("**/api/v1/user-words", (route) =>
    route.fulfill({
      status: 401,
      contentType: "application/problem+json",
      body: JSON.stringify({ detail: "Session expired" }),
    }),
  );
  await saveButton(page).click();
  await expect(page).toHaveURL(/\/login\?/);
  const login = new URL(page.url());
  const original = new URL(selectedURL);
  expect(login.searchParams.get("returnTo")).toBe(
    original.pathname + original.search,
  );
  expect(login.searchParams.get("reason")).toBe("session-expired");
});

test("identical provider wording is shown once without invented edits", async ({
  page,
}) => {
  await selectInvite(page);
  await saveButton(page).click();
  await expect(sentenceBox(page)).toBeVisible();
  const sentence = "I invite you to dinner on Friday.";
  await sentenceBox(page).fill(sentence);
  await page.route("**/api/v1/learner-sentences", async (route) => {
    const response = await route.fetch();
    const data = await response.json();
    return route.fulfill({
      response,
      json: {
        ...data,
        originalSentence: sentence,
        correctedSentence: sentence,
      },
    });
  });
  await page
    .getByRole("button", { name: "Check my sentence", exact: true })
    .click();
  const comparison = page.getByRole("region", { name: "Sentence comparison" });
  await expect(comparison).toContainText("No wording changes suggested.");
  await expect(comparison.getByText(sentence, { exact: true })).toHaveCount(1);
  await expect(comparison.locator("mark")).toHaveCount(0);
});

test("feedback without a suggested correction preserves the checked sentence and rewrite action", async ({
  page,
}) => {
  await selectInvite(page);
  await saveButton(page).click();
  await expect(sentenceBox(page)).toBeVisible();
  const sentence = "I invite you to dinner on Friday.";
  await sentenceBox(page).fill(sentence);
  await page.route("**/api/v1/learner-sentences", async (route) => {
    const response = await route.fetch();
    const data = await response.json();
    return route.fulfill({
      response,
      json: { ...data, originalSentence: sentence, correctedSentence: null },
    });
  });
  await page
    .getByRole("button", { name: "Check my sentence", exact: true })
    .click();
  const comparison = page.getByRole("region", { name: "Sentence comparison" });
  await expect(comparison.getByText(sentence, { exact: true })).toBeVisible();
  await expect(
    comparison.getByText("Corrected sentence", { exact: true }),
  ).toHaveCount(0);
  await expect(
    comparison.getByRole("button", { name: "Revise sentence", exact: true }),
  ).toBeEnabled();
});
