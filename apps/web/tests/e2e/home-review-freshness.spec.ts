import { randomUUID } from "node:crypto";
import { expect, test, type Page } from "@playwright/test";
import type { DailyLesson } from "@vocanova/api-client";

async function completedHome(page: Page) {
  const url = test.info().project.use.baseURL!;
  const api = `http://127.0.0.1:${process.env.MOCK_API_PORT ?? 8080}`;
  await page.context().addCookies([
    { name: "vocanova_session", value: `home-midnight-${randomUUID()}`, url },
    { name: "vocanova_csrf", value: "home-freshness", url },
    { name: "e2e_review_fixture_count", value: "6", url },
    { name: "e2e_daily_review_target", value: "5", url },
    { name: "e2e_review_timezone", value: "Pacific/Honolulu", url },
    { name: "e2e_review_local_date", value: "2026-11-15", url },
  ]);
  const post = async (path: string, body = {}) => {
    const response = await page.request.post(`${api}${path}`, { headers: { "X-CSRF-Token": "home-freshness", "Idempotency-Key": randomUUID() }, data: body });
    expect(response.status()).toBe(200);
    return await response.json() as DailyLesson;
  };
  let session = await post("/api/v1/daily-lessons");
  for (let index = 0; index < 5; index++) {
    session = await post(`/api/v1/daily-lessons/${session.id}/actions`, { action: "answer", stepId: session.currentStep!.id, expectedRevision: session.revision, clientActionId: randomUUID(), choiceId: session.currentStep!.choices[0]!.id });
    session = await post(`/api/v1/daily-lessons/${session.id}/actions`, { action: "continue", stepId: session.currentStep!.id, expectedRevision: session.revision, clientActionId: randomUUID() });
  }
  expect(session.goal).toMatchObject({ reviewed: 5, complete: true });
  await page.clock.install({ time: new Date("2026-11-16T09:59:30Z") });
  await page.goto("/home");
  const card = page.getByRole("region", { name: "Review words", exact: true });
  await expect(card.getByText("✓ Today’s review complete", { exact: true })).toBeVisible();
  await expect(card.getByText(/^Reviewed today:/)).toHaveText("Reviewed today: 5 of 5");
  await expect(card.getByRole("button", { name: "Start review", exact: true })).toHaveCount(0);
  return { url, card };
}

test("Home replaces yesterday’s goal after learner-local midnight using only confirmed reads", async ({ page }) => {
  const { url, card } = await completedHome(page);
  const mutations: string[] = [];
  let reads = 0;
  page.on("request", (request) => {
    if (new URL(request.url()).pathname === "/api/v1/daily-lessons/current") reads++;
    if (request.method() !== "GET" && new URL(request.url()).pathname.startsWith("/api/v1/")) mutations.push(request.method());
  });
  await page.context().addCookies([{ name: "e2e_review_local_date", value: "2026-11-16", url }]);
  await page.clock.fastForward(31_000);
  await page.evaluate(() => window.dispatchEvent(new Event("focus")));
  await expect(card.getByText(/^Reviewed today:/)).toHaveText("Reviewed today: 0 of 5");
  await expect(card.getByText("✓ Today’s review complete", { exact: true })).toHaveCount(0);
  await expect(card.getByRole("button", { name: "Start review", exact: true })).toBeVisible();
  expect(reads).toBe(1);
  expect(mutations).toEqual([]);
  await page.clock.fastForward(120_000);
  expect(reads).toBe(1);
});

test("Home keeps confirmed progress readable when an overnight refresh fails and retries explicitly", async ({ page }) => {
  const { url, card } = await completedHome(page);
  let reads = 0;
  const mutations: string[] = [];
  page.on("request", (request) => {
    if (new URL(request.url()).pathname === "/api/v1/daily-lessons/current") reads++;
    if (request.method() !== "GET" && new URL(request.url()).pathname.startsWith("/api/v1/")) mutations.push(request.method());
  });
  await page.context().addCookies([{ name: "e2e_review_local_date", value: "2026-11-16", url }]);
  await page.route("**/api/v1/daily-lessons/current", (route) => route.fulfill({ status: 503, json: { detail: "Unavailable" } }), { times: 1 });
  await page.clock.fastForward(31_000);
  await page.evaluate(() => window.dispatchEvent(new Event("focus")));
  await expect(page.getByText("Progress may be out of date. Refresh to check today’s reviews.", { exact: true })).toBeVisible();
  await expect(card.getByText(/^Reviewed today:/)).toHaveText("Reviewed today: 5 of 5");
  await page.clock.fastForward(120_000);
  expect(reads).toBe(1);
  await page.getByRole("button", { name: "Refresh today’s progress", exact: true }).click();
  await expect(card.getByText(/^Reviewed today:/)).toHaveText("Reviewed today: 0 of 5");
  await expect(card.getByRole("button", { name: "Start review", exact: true })).toBeVisible();
  expect(reads).toBe(2);
  expect(mutations).toEqual([]);
});
