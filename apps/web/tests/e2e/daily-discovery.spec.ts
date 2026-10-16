import { randomUUID } from "node:crypto";
import { expect, test, type Page } from "@playwright/test";
import type { KnowledgeMap } from "@vocanova/api-client";
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
