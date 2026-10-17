import { randomUUID } from "node:crypto";
import { expect, test, type Page } from "@playwright/test";
import type { DailyLesson, KnowledgeMap } from "@vocanova/api-client";
import { scanForAxeViolations } from "./axe-helper";

async function learner(page: Page, count = 2) {
  const baseURL = test.info().project.use.baseURL;
  if (!baseURL) throw new Error("A local test base URL is required");
  await page.context().addCookies([
    { name: "vocanova_session", value: `daily-${randomUUID()}`, url: baseURL },
    { name: "vocanova_csrf", value: "daily-browser-csrf", url: baseURL },
    { name: "e2e_review_fixture_count", value: String(count), url: baseURL },
  ]);
}
async function assessRangeOne(page: Page, keepUnseen: number[] = []) {
  const api = `http://127.0.0.1:${process.env.MOCK_API_PORT ?? 8080}`;
  const initial = await page.request.get(`${api}/api/v1/knowledge-map?band=1`);
  expect(initial.status()).toBe(200);
  const range = (await initial.json()) as KnowledgeMap;
  expect(range.currentBand.items).toHaveLength(100);
  for (const word of range.currentBand.items) {
    if (keepUnseen.includes(word.rank)) continue;
    const assessed = await page.request.post(`${api}/api/v1/knowledge-map/decisions`, {
      headers: { "X-CSRF-Token": "daily-browser-csrf", "Idempotency-Key": randomUUID() },
      data: { meaningId: word.meaningId, decision: "known", expectedRevision: word.revision },
    });
    expect(assessed.status()).toBe(200);
    expect(await assessed.json()).toMatchObject({ meaningId: word.meaningId, decision: "known" });
  }
  const confirmed = await page.request.get(`${api}/api/v1/knowledge-map?band=1`);
  expect(confirmed.status()).toBe(200);
  return (await confirmed.json()) as KnowledgeMap;
}
async function start(page: Page, count = 2) {
  await learner(page, count);
  await page.goto("/reviews");
  await page.getByRole("button", { name: "Start lesson", exact: true }).click();
  await expect(page.getByRole("heading", { name: "Review word 1", exact: true })).toBeVisible();
}
async function correct(page: Page, index: number) {
  await page.getByRole("radio", { name: `definition for review word ${index}`, exact: true }).check();
  await page.getByRole("button", { name: "Check answer", exact: true }).click();
  await expect(page.getByRole("status").filter({ hasText: /^Correct/ })).toBeVisible();
  await page.getByRole("button", { name: "Continue", exact: true }).click();
}
async function otherActorPost(page: Page, endpoint: string, data: unknown = {}) {
  const response = await page.request.post(`http://127.0.0.1:${process.env.MOCK_API_PORT ?? 8080}${endpoint}`, {
    headers: { "X-CSRF-Token": "daily-browser-csrf", "Idempotency-Key": randomUUID() },
    data,
  });
  expect(response.status()).toBe(200);
  return (await response.json()) as DailyLesson;
}

test("Home starts once, choices stay stable through refresh, completion has a durable address", async ({ page }) => {
  await learner(page);
  let starts = 0;
  page.on("request", (request) => { if (request.method() === "POST" && new URL(request.url()).pathname === "/api/v1/daily-lessons") starts += 1; });
  await page.goto("/home");
  await page.getByRole("button", { name: "Start today’s lesson", exact: true }).click();
  await expect(page).toHaveURL(/\/review\/session\?lesson=/);
  const lessonUrl = page.url();
  await expect(page.getByRole("radio")).toHaveCount(4);
  const choices = await page.getByRole("radio").evaluateAll((radios) => radios.map((radio) => radio.parentElement?.textContent));
  await page.reload();
  await expect(page.getByRole("heading", { name: "Review word 1", exact: true })).toBeVisible();
  expect(await page.getByRole("radio").evaluateAll((radios) => radios.map((radio) => radio.parentElement?.textContent))).toEqual(choices);
  await correct(page, 1);
  await correct(page, 2);
  await expect(page.getByRole("heading", { name: "Lesson complete", exact: true })).toBeFocused();
  await expect(page.getByText("2 words remembered for today.", { exact: true })).toBeVisible();
  await page.getByRole("link", { name: "Back to Home", exact: true }).last().click();
  await expect(page.getByRole("progressbar", { name: "Reviews today", exact: true })).toHaveAttribute("aria-valuenow", "2");
  await expect(page.getByText("Currently due: 0 words.", { exact: false })).toBeVisible();
  await page.goto(lessonUrl);
  await expect(page.getByRole("heading", { name: "Lesson complete", exact: true })).toBeVisible();
  expect(page.url()).toBe(lessonUrl);
  expect(starts).toBe(1);
});

test("completion waits for the confirmed remaining words before offering another lesson", async ({ page }) => {
  await start(page, 1);
  await page.getByRole("radio", { name: "definition for review word 1", exact: true }).check();
  await page.getByRole("button", { name: "Check answer", exact: true }).click();
  let release!: () => void;
  let gate = new Promise<void>((resolve) => { release = resolve; });
  await page.route("**/api/v1/daily-lessons/current", async (route) => {
    await gate;
    await route.continue();
  });
  try {
    const checking = page.waitForRequest((request) => new URL(request.url()).pathname === "/api/v1/daily-lessons/current", { timeout: 5000 });
    await page.getByRole("button", { name: "Continue", exact: true }).click();
    await checking;
    await expect(page.getByRole("heading", { name: "Lesson complete", exact: true })).toBeFocused();
    await expect(page.getByRole("button", { name: "Next lesson", exact: true })).toHaveCount(0);
    await expect(page.getByRole("link", { name: "Discover words", exact: true })).toHaveCount(0);
    await expect(page.getByRole("status").filter({ hasText: "Checking for more words…" })).toBeVisible();
    release();
    await expect(page.getByRole("link", { name: "Discover words", exact: true })).toBeVisible();
    await expect(page.getByRole("main").getByRole("alert")).toHaveCount(0);
    // A completed server-rendered session also starts unconfirmed; its old
    // zero count must not show a premature caught-up action before this read.
    gate = new Promise<void>((resolve) => { release = resolve; });
    const rechecking = page.waitForRequest((request) => new URL(request.url()).pathname === "/api/v1/daily-lessons/current", { timeout: 5000 });
    await page.reload();
    await rechecking;
    await expect(page.getByRole("button", { name: "Next lesson", exact: true })).toHaveCount(0);
    await expect(page.getByRole("link", { name: "Discover words", exact: true })).toHaveCount(0);
    release();
    await expect(page.getByRole("link", { name: "Discover words", exact: true })).toBeVisible();
  } finally {
    release();
    await page.unrouteAll({ behavior: "wait" });
  }
});

test("completion offers Next only after a positive remaining-word read and uses a fresh start key", async ({ page }) => {
  const keys: Array<string | undefined> = [];
  page.on("request", (request) => {
    if (request.method() === "POST" && new URL(request.url()).pathname === "/api/v1/daily-lessons") keys.push(request.headers()["idempotency-key"]);
  });
  await start(page, 11);
  for (let word = 1; word < 10; word++) await correct(page, word);
  await page.getByRole("radio", { name: "definition for review word 10", exact: true }).check();
  await page.getByRole("button", { name: "Check answer", exact: true }).click();
  let release!: () => void;
  const gate = new Promise<void>((resolve) => { release = resolve; });
  await page.route("**/api/v1/daily-lessons/current", async (route) => { await gate; await route.continue(); });
  try {
    const checking = page.waitForRequest((request) => new URL(request.url()).pathname === "/api/v1/daily-lessons/current", { timeout: 5000 });
    await page.getByRole("button", { name: "Continue", exact: true }).click();
    await checking;
    await expect(page.getByRole("button", { name: "Next lesson", exact: true })).toHaveCount(0);
    await expect(page.getByRole("link", { name: "Discover words", exact: true })).toHaveCount(0);
    release();
    const next = page.getByRole("button", { name: "Next lesson", exact: true });
    await expect(next).toBeEnabled();
    await next.click();
    await expect(page.getByRole("heading", { name: "Review word 11", exact: true })).toBeFocused();
    expect(keys).toHaveLength(2);
    expect(keys[0]).toBeTruthy();
    expect(keys[1]).toBeTruthy();
    expect(keys[1]).not.toBe(keys[0]);
  } finally {
    release();
    await page.unrouteAll({ behavior: "wait" });
  }
});

test("a failed completion read leaves availability unknown until explicit saved-progress recovery", async ({ page }) => {
  await start(page, 1);
  await page.route("**/api/v1/daily-lessons/current", (route) => route.fulfill({ status: 503, contentType: "application/problem+json", body: JSON.stringify({ detail: "Remaining words could not load." }) }), { times: 1 });
  await correct(page, 1);
  const error = page.getByRole("main").getByRole("alert");
  await expect(error).toContainText("Remaining words could not load.");
  await expect(page.getByRole("button", { name: "Next lesson", exact: true })).toHaveCount(0);
  await expect(page.getByRole("link", { name: "Discover words", exact: true })).toHaveCount(0);
  await error.getByRole("button", { name: "Load saved progress", exact: true }).click();
  await expect(page.getByRole("link", { name: "Discover words", exact: true })).toBeVisible();
  await expect(error).toHaveCount(0);
});

test("a definitive no-due start reconciles another actor finishing the last word", async ({ page }) => {
  await learner(page, 1);
  await page.goto("/reviews");
  await expect(page.getByRole("button", { name: "Start lesson", exact: true })).toBeVisible();
  // Exercise the actual synthetic API with a second actor while this page
  // retains its earlier one-word due snapshot.
  let other = await otherActorPost(page, "/api/v1/daily-lessons");
  other = await otherActorPost(page, `/api/v1/daily-lessons/${other.id}/actions`, { action: "answer", choiceId: "meaning-a", stepId: other.currentStep!.id, expectedRevision: other.revision, clientActionId: randomUUID() });
  other = await otherActorPost(page, `/api/v1/daily-lessons/${other.id}/actions`, { action: "continue", stepId: other.currentStep!.id, expectedRevision: other.revision, clientActionId: randomUUID() });
  expect(other.status).toBe("completed");
  const rejected = page.waitForResponse((response) => response.request().method() === "POST" && new URL(response.url()).pathname === "/api/v1/daily-lessons");
  await page.getByRole("button", { name: "Start lesson", exact: true }).click();
  expect((await rejected).status()).toBe(409);
  await expect(page.getByRole("heading", { name: "You’re ready for new words", exact: true })).toBeVisible();
  await expect(page.getByRole("link", { name: "Discover words", exact: true })).toBeVisible();
  await expect(page.getByRole("main").getByRole("alert")).toHaveCount(0);
});

test("no-due reconciliation resumes an active session even when the fresh due count is zero", async ({ page }) => {
  await learner(page, 1);
  await page.goto("/reviews");
  const active = await otherActorPost(page, "/api/v1/daily-lessons");
  // A concurrent actor can create a saved session after the rejected Start.
  // This bounded transport case supplies that real fixture session with zero
  // currently due words; its snapshot still needs to remain resumable.
  await page.route("**/api/v1/daily-lessons", (route) => route.fulfill({ status: 409, contentType: "application/problem+json", body: JSON.stringify({ detail: "No words are due. Discover a word to start learning." }) }), { times: 1 });
  await page.route("**/api/v1/daily-lessons/current", (route) => route.fulfill({ status: 200, contentType: "application/json", body: JSON.stringify({ session: active, dueCount: 0 }) }), { times: 1 });
  await page.getByRole("button", { name: "Start lesson", exact: true }).click();
  await expect(page.getByRole("heading", { name: "Review word 1", exact: true })).toBeFocused();
  await expect(page).toHaveURL(new RegExp(`/review/session\\?lesson=${active.id}$`));
  await expect(page.getByRole("radio")).toHaveCount(4);
  await expect(page.getByRole("heading", { name: "You’re ready for new words", exact: true })).toHaveCount(0);
});

test("failed no-due reconciliation preserves the exact start key for a later acknowledged retry", async ({ page }) => {
  await learner(page, 1);
  await page.goto("/reviews");
  const keys: Array<string | undefined> = [];
  await page.route("**/api/v1/daily-lessons", async (route) => {
    keys.push(route.request().headers()["idempotency-key"]);
    if (keys.length === 1) await route.fulfill({ status: 409, contentType: "application/problem+json", body: JSON.stringify({ detail: "No words are due. Discover a word to start learning." }) });
    else await route.continue();
  });
  await page.route("**/api/v1/daily-lessons/current", (route) => route.fulfill({ status: 503, contentType: "application/problem+json", body: JSON.stringify({ detail: "Remaining words unavailable." }) }), { times: 1 });
  await page.getByRole("button", { name: "Start lesson", exact: true }).click();
  await expect(page.getByRole("main").getByRole("alert")).toContainText("Remaining words unavailable.");
  await expect(page.getByRole("heading", { name: "You’re ready for new words", exact: true })).toHaveCount(0);
  await page.getByRole("button", { name: "Start lesson", exact: true }).click();
  await expect(page.getByRole("heading", { name: "Review word 1", exact: true })).toBeFocused();
  expect(keys).toHaveLength(2);
  expect(keys[0]).toBeTruthy();
  expect(keys[1]).toBe(keys[0]);
});

for (const failure of ["generic-conflict", "untyped-no-due", "server-error", "lost-response"] as const) {
  test(`an uncertain ${failure} start never becomes a no-due claim or changes its retry key`, async ({ page }) => {
    await learner(page, 1);
    await page.goto("/reviews");
    const keys: Array<string | undefined> = [];
    let reconciliationReads = 0;
    page.on("request", (request) => { if (new URL(request.url()).pathname === "/api/v1/daily-lessons/current") reconciliationReads++; });
    await page.route("**/api/v1/daily-lessons", async (route) => {
      keys.push(route.request().headers()["idempotency-key"]);
      if (keys.length > 1) { await route.continue(); return; }
      if (failure === "lost-response") { await route.fetch(); await route.abort("failed"); return; }
      await route.fulfill({
        status: failure === "server-error" ? 503 : 409,
        contentType: failure === "untyped-no-due" ? "application/json" : "application/problem+json",
        body: JSON.stringify({ detail: failure === "generic-conflict" ? "This lesson changed. Reload it to continue." : failure === "untyped-no-due" ? "No words are due. Discover a word to start learning." : "Temporary failure." }),
      });
    });
    await page.getByRole("button", { name: "Start lesson", exact: true }).click();
    await expect(page.getByRole("main").getByRole("alert")).toBeVisible();
    await expect(page.getByRole("heading", { name: "You’re ready for new words", exact: true })).toHaveCount(0);
    expect(reconciliationReads).toBe(0);
    await page.getByRole("button", { name: "Start lesson", exact: true }).click();
    await expect(page.getByRole("heading", { name: "Review word 1", exact: true })).toBeFocused();
    expect(keys).toHaveLength(2);
    expect(keys[0]).toBeTruthy();
    expect(keys[1]).toBe(keys[0]);
  });
}

test("saved-progress recovery acknowledges a lost Start before a later Next uses a new key", async ({ page }) => {
  await learner(page, 11);
  await page.goto("/reviews");
  const keys: Array<string | undefined> = [];
  await page.route("**/api/v1/daily-lessons", async (route) => {
    keys.push(route.request().headers()["idempotency-key"]);
    if (keys.length === 1) { await route.fetch(); await route.abort("failed"); }
    else await route.continue();
  });
  await page.getByRole("button", { name: "Start lesson", exact: true }).click();
  const error = page.getByRole("main").getByRole("alert");
  await expect(error).toBeVisible();
  await error.getByRole("button", { name: "Load saved progress", exact: true }).click();
  await expect(page.getByRole("heading", { name: "Review word 1", exact: true })).toBeFocused();
  expect(keys).toHaveLength(2);
  expect(keys[0]).toBeTruthy();
  expect(keys[1]).toBe(keys[0]);
  for (let word = 1; word <= 10; word++) await correct(page, word);
  await page.getByRole("button", { name: "Next lesson", exact: true }).click();
  await expect(page.getByRole("heading", { name: "Review word 11", exact: true })).toBeFocused();
  expect(keys).toHaveLength(3);
  expect(keys[2]).toBeTruthy();
  expect(keys[2]).not.toBe(keys[0]);
});

test("saved-progress recovery reads beyond an immutable Start receipt after another actor completes the lesson", async ({ page }) => {
  await learner(page, 1);
  await page.goto("/reviews");
  const keys: Array<string | undefined> = [];
  let initial: DailyLesson;
  await page.route("**/api/v1/daily-lessons", async (route) => {
    keys.push(route.request().headers()["idempotency-key"]);
    if (keys.length === 1) {
      const committed = await route.fetch();
      initial = await committed.json() as DailyLesson;
      await route.abort("failed");
    } else {
      // The real API replays the original receipt, not the mutable session.
      await route.fulfill({ status: 200, contentType: "application/json", body: JSON.stringify(initial) });
    }
  });
  await page.getByRole("button", { name: "Start lesson", exact: true }).click();
  const error = page.getByRole("main").getByRole("alert");
  await expect(error).toBeVisible();
  let other = initial!;
  other = await otherActorPost(page, `/api/v1/daily-lessons/${other.id}/actions`, { action: "answer", choiceId: "meaning-a", stepId: other.currentStep!.id, expectedRevision: other.revision, clientActionId: randomUUID() });
  other = await otherActorPost(page, `/api/v1/daily-lessons/${other.id}/actions`, { action: "continue", stepId: other.currentStep!.id, expectedRevision: other.revision, clientActionId: randomUUID() });
  expect(other.status).toBe("completed");
  // Current and specific reads can observe different snapshots. Even when
  // Current captured this same session before completion, the newer specific
  // revision must win rather than reopening its initial question.
  await page.route("**/api/v1/daily-lessons/current", (route) => route.fulfill({ status: 200, contentType: "application/json", body: JSON.stringify({ session: initial, dueCount: 1 }) }), { times: 1 });
  await error.getByRole("button", { name: "Load saved progress", exact: true }).click();
  await expect(page.getByRole("heading", { name: "Lesson complete", exact: true })).toBeFocused();
  await expect(page.getByText("1 word remembered for today.", { exact: true })).toBeVisible();
  await expect(page.getByRole("link", { name: "Discover words", exact: true })).toBeVisible();
  await expect(page.getByRole("radio")).toHaveCount(0);
  expect(keys).toHaveLength(2);
  expect(keys[0]).toBeTruthy();
  expect(keys[1]).toBe(keys[0]);
});

test("a mistake earns no credit, keeps the chosen answer on refresh, and returns after another word", async ({ page }) => {
  await start(page);
  await page.getByRole("radio", { name: "A different meaning.", exact: true }).check();
  await page.getByRole("button", { name: "Check answer", exact: true }).click();
  await expect(page.getByText("Let’s try this one again later", { exact: true })).toBeVisible();
  await page.reload();
  await expect(page.getByRole("radio", { name: "A different meaning.", exact: true })).toBeChecked();
  await page.getByRole("button", { name: "Continue", exact: true }).click();
  await expect(page.getByRole("heading", { name: "Review word 2", exact: true })).toBeFocused();
  await correct(page, 2);
  await expect(page.getByRole("heading", { name: "Review word 1", exact: true })).toBeVisible();
  await correct(page, 1);
  await expect(page.getByText("2 words remembered for today.", { exact: true })).toBeVisible();
  await expect(page.getByText("1 correct on the first try.", { exact: true })).toBeVisible();
});

test("an unconfirmed answer freezes its exact intent and retries without duplicate credit", async ({ page }) => {
  await start(page, 1);
  const attempts: Array<{ key: string | undefined; body: string | null }> = [];
  await page.route("**/api/v1/daily-lessons/*/actions", async (route) => {
    attempts.push({ key: route.request().headers()["idempotency-key"], body: route.request().postData() });
    if (attempts.length === 1) { await route.fetch(); await route.abort("failed"); }
    else await route.continue();
  });
  await page.getByRole("radio", { name: "definition for review word 1", exact: true }).check();
  await page.getByRole("button", { name: "Check answer", exact: true }).click();
  await expect(page.getByRole("button", { name: "Retry same answer", exact: true })).toBeVisible();
  await expect(page.getByRole("main").getByRole("alert")).toBeFocused();
  await expect(page.getByRole("button", { name: "Retry same answer", exact: true })).toBeInViewport({ ratio: 1 });
  for (const radio of await page.getByRole("radio").all()) await expect(radio).toBeDisabled();
  await page.getByRole("button", { name: "Retry same answer", exact: true }).click();
  await expect(page.getByRole("status").filter({ hasText: /^Correct/ })).toBeVisible();
  expect(attempts).toHaveLength(2);
  expect(attempts[1]).toEqual(attempts[0]);
  await page.getByRole("button", { name: "Continue", exact: true }).click();
  await expect(page.getByText("1 word remembered for today.", { exact: true })).toBeVisible();
});

test("explicit skip is explained honestly and does not claim a remembered word", async ({ page }) => {
  await start(page, 1);
  await page.getByRole("button", { name: "Skip for now", exact: true }).click();
  await expect(page.getByText("Skipped for now", { exact: true })).toBeVisible();
  await page.getByRole("button", { name: "Continue", exact: true }).click();
  await expect(page.getByText("0 words remembered for today.", { exact: true })).toBeVisible();
});

test("all 100 ranges are reachable, discovery confirms choices before advancing, and refresh preserves state", async ({ page }) => {
  await learner(page, 0);
  await page.goto("/knowledge-map");
  await page.getByText("Choose a range · 1–10,000", { exact: true }).click();
  await expect(page.getByRole("navigation", { name: "Word ranges", exact: true }).getByRole("link")).toHaveCount(100);
  await page.getByRole("link", { name: "Words 9901 to 10000, 0 already know, 0 learning, 100 new", exact: true }).click();
  await expect(page.getByRole("heading", { name: "Words 9,901–10,000", exact: true })).toBeVisible();
  await page.goto("/discovery?band=1&word=1");
  await page.getByRole("button", { name: "Start discovering", exact: true }).click();
  await expect(page.getByRole("heading", { name: "the", exact: true })).toBeFocused();
  await page.getByRole("button", { name: "Already know", exact: true }).click();
  await expect(page.getByRole("heading", { name: "be", exact: true })).toBeFocused();
  await page.getByRole("button", { name: "Learn", exact: true }).click();
  await expect(page.getByRole("heading", { name: "Map word 3", exact: true })).toBeVisible();
  await page.getByRole("link", { name: "Knowledge map", exact: true }).click();
  await expect(page.getByText("1 already know", { exact: true })).toBeVisible();
  await expect(page.getByText("1 learning", { exact: true })).toBeVisible();
  await page.reload();
  await expect(page.getByRole("link", { name: "1 the Already know", exact: true })).toBeVisible();
});

test("discovery continues to the next unseen range after 100 acknowledged assessments", async ({ page }) => {
  await learner(page, 0);
  const confirmed = await assessRangeOne(page);
  expect(confirmed.bands[0]).toMatchObject({ knownCount: 100, unseenCount: 0 });

  await page.goto("/home");
  await page.getByRole("link", { name: "Discover words", exact: true }).click();
  await page.getByRole("button", { name: "Start discovering", exact: true }).click();
  await expect(page.getByRole("heading", { name: "Map word 101", exact: true })).toBeFocused();
  await expect(page).toHaveURL(/\/discovery\?band=2&word=101$/);

  await page.goto("/knowledge-map?band=1");
  const discover = page.getByRole("link", { name: "Discover", exact: true });
  await expect(discover).toHaveAttribute("href", "/discovery?band=2");
  await discover.click();
  await page.getByRole("button", { name: "Start discovering", exact: true }).click();
  await expect(page.getByRole("heading", { name: "Map word 101", exact: true })).toBeFocused();

  await page.goto("/discovery?band=1");
  await page.getByRole("button", { name: "Start discovering", exact: true }).click();
  await expect(page.getByRole("heading", { name: "Map word 101", exact: true })).toBeFocused();
  await page.goto("/discovery?band=1&word=1");
  await page.getByRole("button", { name: "Start discovering", exact: true }).click();
  await expect(page.getByRole("heading", { name: "the", exact: true })).toBeFocused();
  await expect(page).toHaveURL(/\/discovery\?band=1&word=1$/);
});

test("guided discovery skips assessed gaps and retains its decision until the next range loads", async ({ page }) => {
  await learner(page, 0);
  const before = await assessRangeOne(page, [1, 3]);
  expect(before.bands[0]).toMatchObject({ knownCount: 98, unseenCount: 2 });
  await page.goto("/discovery?band=1&word=1");
  await page.getByRole("button", { name: "Start discovering", exact: true }).click();
  await page.getByRole("button", { name: "Learn", exact: true }).click();
  await expect(page.getByRole("heading", { name: "Map word 3", exact: true })).toBeFocused();
  await expect(page).toHaveURL(/\/discovery\?band=1&word=3$/);

  const attempts: Array<{ key: string | undefined; body: string | null }> = [];
  page.on("request", (request) => {
    if (request.method() === "POST" && new URL(request.url()).pathname === "/api/v1/knowledge-map/decisions")
      attempts.push({ key: request.headers()["idempotency-key"], body: request.postData() });
  });
  await page.route("**/api/v1/knowledge-map?band=2", (route) => route.abort("failed"), { times: 1 });
  await page.getByRole("button", { name: "Already know", exact: true }).click();
  await expect(page.getByRole("button", { name: "Retry choice", exact: true })).toBeVisible();
  await expect(page.getByRole("heading", { name: "Map word 3", exact: true })).toBeVisible();
  await expect(page.getByRole("button", { name: "Learn", exact: true })).toBeDisabled();
  await expect(page).toHaveURL(/\/discovery\?band=1&word=3$/);
  await page.getByRole("button", { name: "Retry choice", exact: true }).click();
  await expect(page.getByRole("heading", { name: "Map word 101", exact: true })).toBeFocused();
  await expect(page).toHaveURL(/\/discovery\?band=2&word=101$/);
  expect(attempts).toHaveLength(2);
  expect(attempts[0]!.key).toBeTruthy();
  expect(attempts[1]).toEqual(attempts[0]);
});

test("an entirely assessed map has no guided word but keeps explicit revisits", async ({ page }) => {
  await learner(page, 0);
  await page.context().addCookies([
    { name: "e2e_knowledge_map", value: "complete", url: test.info().project.use.baseURL! },
  ]);
  await page.goto("/discovery");
  await expect(page.getByRole("heading", { name: "You’ve explored the whole map", exact: true })).toBeVisible();
  await expect(page.getByRole("button", { name: "Start discovering", exact: true })).toHaveCount(0);
  const map = page.getByRole("link", { name: "Knowledge map", exact: true });
  await expect(map).toHaveAttribute("href", "/knowledge-map?band=1#range-words");
  await map.click();
  await expect(page.getByText("10,000 already know", { exact: true })).toBeVisible();
  await expect(page.getByText("No new words left to discover.", { exact: true })).toBeVisible();
  await expect(page.getByRole("link", { name: "Discover", exact: true })).toHaveCount(0);
  const revisit = page.getByRole("link", { name: "1 the Already know", exact: true });
  await expect(revisit).toHaveAttribute("href", "/discovery?band=1&word=1");
  await revisit.click();
  await page.getByRole("button", { name: "Start discovering", exact: true }).click();
  await expect(page.getByRole("heading", { name: "the", exact: true })).toBeFocused();
  await expect(page).toHaveURL(/\/discovery\?band=1&word=1$/);
  await page.getByRole("button", { name: "Already know", exact: true }).click();
  await expect(page.getByRole("heading", { name: "You’ve explored the whole map", exact: true })).toBeVisible();
  await expect(page.getByRole("link", { name: "View map", exact: true })).toHaveAttribute("href", "/knowledge-map?band=1");
  await page.goto("/discovery?band=100");
  await expect(page.getByRole("heading", { name: "You’ve explored the whole map", exact: true })).toBeVisible();
  await expect(page.getByRole("link", { name: "Knowledge map", exact: true })).toHaveAttribute("href", "/knowledge-map?band=100#range-words");
});

test("a concurrent range completion offers a bounded retry instead of claiming the map is complete", async ({ page }) => {
  await learner(page, 0);
  await page.context().addCookies([
    { name: "e2e_knowledge_map", value: "rollover-race", url: test.info().project.use.baseURL! },
  ]);
  await page.goto("/discovery");
  await expect(page.getByRole("main").getByRole("alert")).toHaveText("New words changed elsewhere. Try again to load your next word.");
  await expect(page.getByRole("heading", { name: "You’ve explored the whole map", exact: true })).toHaveCount(0);
  const retry = page.getByRole("link", { name: "Try again", exact: true });
  await expect(retry).toHaveAttribute("href", "/discovery");
  await retry.click();
  await page.getByRole("button", { name: "Start discovering", exact: true }).click();
  await expect(page.getByRole("heading", { name: "Map word 201", exact: true })).toBeFocused();
  await expect(page).toHaveURL(/\/discovery\?band=3&word=201$/);
});

test("daily choices remain accessible and unobscured in both themes", async ({ page }) => {
  await start(page, 1);
  for (const theme of ["light", "dark"]) {
    await page.evaluate((value) => document.documentElement.setAttribute("data-theme", value), theme);
    // Scan each theme from the same viewport, then independently check focus
    // below. A previous focus scroll can partially clip an unfocused control
    // behind the sticky header and turn the next scan into a different state.
    await page.evaluate(() => window.scrollTo({ top: 0, behavior: "instant" }));
    const { criticalOrSerious } = await scanForAxeViolations(page);
    expect(criticalOrSerious).toEqual([]);
    expect(await page.evaluate(() => document.documentElement.scrollWidth <= window.innerWidth)).toBe(true);
    await page.keyboard.press("Tab");
    await page.getByRole("radio").last().focus();
    const radio = await page.getByRole("radio").last().boundingBox();
    const footer = await page.getByRole("region", { name: "Session actions", exact: true }).boundingBox();
    if (radio && footer) expect(radio.y + radio.height).toBeLessThanOrEqual(footer.y);
    await page.keyboard.press("ArrowDown");
    await expect(page.getByRole("radio").first()).toBeFocused();
    await expect(page.getByRole("radio").first()).toBeChecked();
  }
});

test("discovery retries the same committed choice after a lost response", async ({ page }) => {
  await learner(page, 0);
  await page.goto("/discovery?band=1&word=1");
  await page.getByRole("button", { name: "Start discovering", exact: true }).click();
  const attempts: Array<{ key: string | undefined; body: string | null }> = [];
  await page.route("**/api/v1/knowledge-map/decisions", async (route) => {
    attempts.push({ key: route.request().headers()["idempotency-key"], body: route.request().postData() });
    if (attempts.length === 1) { await route.fetch(); await route.abort("failed"); }
    else await route.continue();
  });
  await page.getByRole("button", { name: "Already know", exact: true }).click();
  await expect(page.getByRole("button", { name: "Retry choice", exact: true })).toBeVisible();
  await expect(page.getByRole("main").getByRole("alert")).toBeFocused();
  await expect(page.getByRole("button", { name: "Retry choice", exact: true })).toBeInViewport({ ratio: 1 });
  await expect(page.getByRole("button", { name: "Learn", exact: true })).toBeDisabled();
  await page.getByRole("button", { name: "Retry choice", exact: true }).click();
  await expect(page.getByRole("heading", { name: "be", exact: true })).toBeVisible();
  expect(attempts).toHaveLength(2);
  expect(attempts[1]).toEqual(attempts[0]);
});

test("discovery waits for confirmed fresh state when conflict recovery itself fails", async ({ page }) => {
  await learner(page, 0);
  await page.goto("/discovery?band=1&word=1");
  await page.getByRole("button", { name: "Start discovering", exact: true }).click();
  await page.route("**/api/v1/knowledge-map/decisions", (route) => route.fulfill({ status: 409, contentType: "application/json", body: JSON.stringify({ detail: "Word changed" }) }), { times: 1 });
  await page.route("**/api/v1/knowledge-map?band=1", (route) => route.fulfill({ status: 503, contentType: "application/json", body: JSON.stringify({ detail: "Temporary read failure" }) }), { times: 1 });
  await page.getByRole("button", { name: "Learn", exact: true }).click();
  await expect(page.getByRole("button", { name: "Load latest word", exact: true })).toBeVisible();
  await expect(page.getByRole("button", { name: "Already know", exact: true })).toBeDisabled();
  await page.getByRole("button", { name: "Load latest word", exact: true }).click();
  await expect(page.getByRole("button", { name: "Learn", exact: true })).toBeEnabled();
  await page.getByRole("button", { name: "Already know", exact: true }).click();
  await expect(page.getByRole("heading", { name: "be", exact: true })).toBeVisible();
});
