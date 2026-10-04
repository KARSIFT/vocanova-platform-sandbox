import { randomUUID } from "node:crypto";
import { expect, test, type BrowserContext, type Page } from "@playwright/test";
import type { LessonSession, SavedMeaning } from "@vocanova/api-client";
import { scanForAxeViolations } from "./axe-helper";

function apiURL(baseURL: string, path: string) {
  const url = new URL(baseURL);
  url.port = process.env.MOCK_API_PORT ?? "8080";
  return new URL(path, url).toString();
}

async function mutationHeaders(context: BrowserContext) {
  const token = (await context.cookies()).find(
    (cookie) => cookie.name === "vocanova_csrf",
  )?.value;
  if (!token) throw new Error("Missing synthetic CSRF cookie");
  return { "X-CSRF-Token": token, "Idempotency-Key": randomUUID() };
}

// The existing guided-lesson spec exercises the teaching UI. Prepare a real
// completed fixture session here so these tests isolate its review handoff.
async function completedLesson(context: BrowserContext, baseURL: string) {
  const opened = await context.request.post(
    apiURL(baseURL, "/api/v1/lessons/conversation-basics/sessions"),
    { headers: await mutationHeaders(context), data: {} },
  );
  expect(opened.ok()).toBe(true);
  let session = (await opened.json()) as LessonSession;
  while (session.status !== "completed") {
    const step = session.currentStep;
    if (!step) throw new Error("Missing fixture lesson step");
    const answer =
      session.words[session.completedSteps % session.words.length]!;
    const action = session.canContinue ? "continue" : "answer";
    const key = randomUUID();
    const result = await context.request.post(
      apiURL(baseURL, `/api/v1/lesson-sessions/${session.id}/actions`),
      {
        headers: {
          ...(await mutationHeaders(context)),
          "Idempotency-Key": key,
        },
        data: {
          stepId: step.id,
          expectedRevision: session.revision,
          clientActionId: key,
          action,
          ...(action === "answer" ? { choiceId: answer.meaningId } : {}),
        },
      },
    );
    expect(result.ok()).toBe(true);
    session = (await result.json()) as LessonSession;
  }
  return session;
}

function wordRow(page: Page, text: string) {
  return page
    .getByRole("main")
    .getByRole("listitem")
    .filter({ has: page.getByRole("link", { name: text, exact: true }) });
}

test.beforeEach(async ({ context, baseURL }) => {
  if (!baseURL) throw new Error("Missing app URL");
  await context.addCookies([
    { name: "vocanova_session", value: randomUUID(), url: baseURL },
    { name: "vocanova_csrf", value: randomUUID(), url: baseURL },
  ]);
});

for (const theme of ["light", "dark"] as const) {
  test(`completed lesson saves only chosen meanings and respects removal in ${theme} mode`, async ({
    page,
    context,
    baseURL,
  }, testInfo) => {
    await context.addCookies([
      { name: "vocanova_theme", value: theme, url: baseURL! },
    ]);
    const session = await completedLesson(context, baseURL!);
    const invite = session.words[0]!;
    const confirm = session.words[1]!;
    const priorSave = await context.request.post(
      apiURL(baseURL!, "/api/v1/user-words"),
      {
        headers: await mutationHeaders(context),
        data: { meaningId: confirm.meaningId, source: "journey" },
      },
    );
    expect(priorSave.ok()).toBe(true);
    const known = await context.request.patch(
      apiURL(baseURL!, `/api/v1/meaning-knowledge/${invite.meaningId}`),
      {
        headers: await mutationHeaders(context),
        data: { selfReportedKnown: true },
      },
    );
    expect(known.ok()).toBe(true);
    const browserWrites: string[] = [];
    page.on("request", (request) => {
      if (["POST", "PUT", "PATCH", "DELETE"].includes(request.method())) {
        browserWrites.push(new URL(request.url()).pathname);
      }
    });
    await page.goto("/learn/conversation-basics");
    await expect(page.locator("html")).toHaveAttribute("data-theme", theme);
    const inviteRow = wordRow(page, "invite");
    const save = inviteRow.getByRole("button", {
      name: "Save invite for review",
    });
    await expect(save).toBeEnabled();
    await expect(
      wordRow(page, "confirm").getByRole("link", {
        name: "Open saved confirm",
      }),
    ).toBeVisible();
    await expect(
      inviteRow.getByText(
        "Marked as already known. Saving for review is optional.",
      ),
    ).toBeVisible();
    await expect(
      page
        .getByRole("main")
        .getByText("Finishing a lesson does not save its words automatically."),
    ).toBeVisible();
    expect(browserWrites).toEqual([]);
    const bounds = await save.boundingBox();
    expect(bounds?.height).toBeGreaterThanOrEqual(44);
    expect(bounds?.width).toBeGreaterThanOrEqual(44);
    await save.focus();
    await page.keyboard.press("Enter");
    await expect(
      inviteRow.getByRole("link", { name: "Open saved invite" }),
    ).toBeVisible();
    expect(browserWrites).toEqual(["/api/v1/user-words"]);

    const persisted = await context.request.get(
      apiURL(baseURL!, "/api/v1/user-words"),
    );
    const library = await persisted.json();
    expect(
      library.items.map((item: SavedMeaning) => item.meaningId).sort(),
    ).toEqual([invite.meaningId, confirm.meaningId].sort());
    const persistedKnown = await context.request.get(
      apiURL(baseURL!, `/api/v1/meaning-knowledge/${invite.meaningId}`),
    );
    expect((await persistedKnown.json()).selfReportedKnown).toBe(true);
    const unchangedLesson = await context.request.get(
      apiURL(baseURL!, `/api/v1/lesson-sessions/${session.id}`),
    );
    expect(await unchangedLesson.json()).toEqual(session);

    const removed = await context.request.delete(
      apiURL(baseURL!, `/api/v1/user-words/${invite.meaningId}`),
      { headers: await mutationHeaders(context) },
    );
    expect(removed.status()).toBe(204);
    await page.reload();
    await expect(save).toBeEnabled();
    await expect(
      inviteRow.getByRole("link", { name: "Open saved invite" }),
    ).toHaveCount(0);
    await expect(
      wordRow(page, "confirm").getByRole("link", {
        name: "Open saved confirm",
      }),
    ).toBeVisible();
    expect(browserWrites).toEqual(["/api/v1/user-words"]);
    expect(
      await page.evaluate(
        () => document.documentElement.scrollWidth <= window.innerWidth,
      ),
    ).toBe(true);
    await expect(page).toHaveTitle(/Vocanova/i);
    expect((await scanForAxeViolations(page)).criticalOrSerious).toEqual([]);
    await page.screenshot({
      path: testInfo.outputPath(`lesson-review-save-${theme}.png`),
      fullPage: true,
    });
  });
}

test("a completed lesson preserves save identity through CSRF recovery and a lost response", async ({
  page,
  context,
  baseURL,
}) => {
  const session = await completedLesson(context, baseURL!);
  await page.goto("/learn/conversation-basics");
  const row = wordRow(page, "invite");
  const save = row.getByRole("button", { name: "Save invite for review" });
  await expect(save).toBeEnabled();
  const writes: { key: string | undefined; body: string | null }[] = [];
  await page.route("**/api/v1/user-words", async (route) => {
    if (route.request().method() !== "POST") return route.continue();
    writes.push({
      key: route.request().headers()["idempotency-key"],
      body: route.request().postData(),
    });
    if (writes.length === 1) {
      const applied = await route.fetch();
      expect(applied.ok()).toBe(true);
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
  await save.click();
  await expect(row.getByRole("alert")).toContainText(
    "could not confirm this save",
  );
  expect(writes).toEqual([]);
  await expect(save).toBeDisabled();
  await row.getByRole("button", { name: "Retry save" }).click();
  await expect(row.getByRole("alert")).toContainText(
    "could not confirm this save",
  );
  expect(writes).toHaveLength(1);
  await expect(
    row.getByRole("link", { name: "Open saved invite" }),
  ).toHaveCount(0);
  await row.getByRole("button", { name: "Retry save" }).click();
  await expect(
    row.getByRole("link", { name: "Open saved invite" }),
  ).toBeVisible();
  expect(recoveries).toBe(2);
  expect(writes).toHaveLength(2);
  expect(writes[0]!.key).toBeTruthy();
  expect(writes[1]).toEqual(writes[0]);
  expect(JSON.parse(writes[0]!.body!)).toEqual({
    meaningId: session.words[0]!.meaningId,
    source: "journey",
  });
  const library = await context.request.get(
    apiURL(baseURL!, "/api/v1/user-words"),
  );
  expect((await library.json()).totalCount).toBe(1);
});

test("replaying an uncertain save never restores a meaning removed afterwards", async ({
  page,
  context,
  baseURL,
}) => {
  const session = await completedLesson(context, baseURL!);
  const meaningId = session.words[0]!.meaningId;
  await page.goto("/learn/conversation-basics");
  const row = wordRow(page, "invite");
  await expect(
    row.getByRole("button", { name: "Save invite for review" }),
  ).toBeEnabled();
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
      const removed = await context.request.delete(
        apiURL(baseURL!, `/api/v1/user-words/${meaningId}`),
        { headers: await mutationHeaders(context) },
      );
      expect(removed.status()).toBe(204);
      return route.abort("failed");
    }
    if (writes.length === 2) {
      // The real API returns 404 for a receipt whose saved row was removed.
      // The client must check the current overlay instead of replaying forever.
      expect(writes[1]).toEqual(writes[0]);
      return route.fulfill({
        status: 404,
        contentType: "application/json",
        body: JSON.stringify({ detail: "Saved word not found" }),
      });
    }
    return route.continue();
  });
  await row.getByRole("button", { name: "Save invite for review" }).click();
  await expect(row.getByRole("alert")).toContainText(
    "could not confirm this save",
  );
  await row.getByRole("button", { name: "Retry save" }).click();
  await expect(row.getByRole("status")).toContainText(
    "Your earlier save was not reapplied",
  );
  await expect(
    row.getByRole("button", { name: "Save invite for review" }),
  ).toBeEnabled();
  expect(writes).toHaveLength(2);
  const empty = await context.request.get(
    apiURL(baseURL!, "/api/v1/user-words"),
  );
  expect((await empty.json()).totalCount).toBe(0);
  await row.getByRole("button", { name: "Save invite for review" }).click();
  await expect(
    row.getByRole("link", { name: "Open saved invite" }),
  ).toBeVisible();
  expect(writes).toHaveLength(3);
  expect(writes[2]!.key).not.toBe(writes[0]!.key);
});

test("an unavailable current saved state cannot be mistaken for an unsaved word", async ({
  page,
  context,
  baseURL,
}) => {
  const session = await completedLesson(context, baseURL!);
  const saved = await context.request.post(
    apiURL(baseURL!, "/api/v1/user-words"),
    {
      headers: await mutationHeaders(context),
      data: { meaningId: session.words[0]!.meaningId, source: "journey" },
    },
  );
  expect(saved.ok()).toBe(true);
  await page.route(
    "**/api/v1/canonical-words/invite",
    (route) =>
      route.fulfill({
        status: 503,
        contentType: "application/json",
        body: JSON.stringify({ detail: "temporarily unavailable" }),
      }),
    { times: 1 },
  );
  const writes: string[] = [];
  page.on("request", (request) => {
    if (request.method() === "POST")
      writes.push(new URL(request.url()).pathname);
  });
  await page.goto("/learn/conversation-basics");
  const row = wordRow(page, "invite");
  await expect(row.getByRole("alert")).toContainText("could not check");
  await expect(
    row.getByRole("button", { name: "Save invite for review" }),
  ).toHaveCount(0);
  await row.getByRole("button", { name: "Check saved status" }).click();
  await expect(
    row.getByRole("link", { name: "Open saved invite" }),
  ).toBeVisible();
  expect(writes).toEqual([]);
});

for (const state of ["saved", "unsaved", "unavailable"] as const) {
  test(`a save conflict reloads ${state} canonical state without automatically saving`, async ({
    page,
    context,
    baseURL,
  }) => {
    const session = await completedLesson(context, baseURL!);
    const meaningId = session.words[0]!.meaningId;
    await page.goto("/learn/conversation-basics");
    const row = wordRow(page, "invite");
    const save = row.getByRole("button", { name: "Save invite for review" });
    await expect(save).toBeEnabled();
    const writes: { key: string | undefined; body: string | null }[] = [];
    let conflictReturned = false;
    let recoveryReads = 0;
    await page.route("**/api/v1/canonical-words/invite", async (route) => {
      if (!conflictReturned) return route.continue();
      recoveryReads += 1;
      if (state === "unavailable" && recoveryReads === 1) {
        return route.fulfill({
          status: 503,
          contentType: "application/json",
          body: JSON.stringify({ detail: "temporarily unavailable" }),
        });
      }
      return route.continue();
    });
    await page.route("**/api/v1/user-words", async (route) => {
      if (route.request().method() !== "POST") return route.continue();
      writes.push({
        key: route.request().headers()["idempotency-key"],
        body: route.request().postData(),
      });
      if (writes.length !== 1) return route.continue();
      if (state === "saved") {
        // Another confirmed action has already saved this meaning. A conflict
        // must read that current state instead of implying this write succeeded.
        const saved = await context.request.post(
          apiURL(baseURL!, "/api/v1/user-words"),
          {
            headers: await mutationHeaders(context),
            data: { meaningId, source: "journey" },
          },
        );
        expect(saved.ok()).toBe(true);
      }
      conflictReturned = true;
      return route.fulfill({
        status: 409,
        contentType: "application/problem+json",
        body: JSON.stringify({ detail: "saved state changed" }),
      });
    });
    await save.click();
    if (state === "unavailable") {
      await expect(row.getByRole("alert")).toContainText(
        "could not confirm the current saved status",
      );
      await expect(save).toHaveCount(0);
      await expect(
        row.getByRole("link", { name: "Open saved invite" }),
      ).toHaveCount(0);
      expect(recoveryReads).toBe(1);
      expect(writes).toHaveLength(1);
      await row.getByRole("button", { name: "Check saved status" }).click();
    }
    if (state === "saved") {
      await expect(
        row.getByRole("link", { name: "Open saved invite" }),
      ).toBeVisible();
      await expect(row.getByRole("status")).toContainText(
        "The current saved meaning is up to date.",
      );
      await expect(save).toHaveCount(0);
    } else {
      await expect(save).toBeEnabled();
      await expect(
        row.getByRole("link", { name: "Open saved invite" }),
      ).toHaveCount(0);
    }
    expect(recoveryReads).toBe(state === "unavailable" ? 2 : 1);
    expect(writes).toHaveLength(1);
    const library = await context.request.get(
      apiURL(baseURL!, "/api/v1/user-words"),
    );
    expect((await library.json()).totalCount).toBe(state === "saved" ? 1 : 0);
    if (state === "unsaved") {
      await expect(row.getByRole("status")).toContainText(
        "Your earlier save was not reapplied",
      );
      await save.click();
      await expect(
        row.getByRole("link", { name: "Open saved invite" }),
      ).toBeVisible();
      expect(writes).toHaveLength(2);
      expect(writes[1]!.key).not.toBe(writes[0]!.key);
      expect(writes[1]!.body).toBe(writes[0]!.body);
    }
  });
}
