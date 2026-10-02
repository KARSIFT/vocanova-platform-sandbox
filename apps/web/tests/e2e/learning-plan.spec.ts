import { randomUUID } from "node:crypto";
import { expect, test } from "@playwright/test";
import { scanForAxeViolations, formatViolations } from "./axe-helper";

const mockAPI = `http://127.0.0.1:${process.env.MOCK_API_PORT ?? 8080}`;

test.beforeEach(async ({ context, baseURL }) => {
  if (!baseURL) throw new Error("Missing app URL");
  await context.addCookies([
    { name: "vocanova_session", value: randomUUID(), url: baseURL },
    { name: "vocanova_csrf", value: randomUUID(), url: baseURL },
    // No work meanings in this small fixture: the real page must fall back.
    { name: "e2e_onboarding_focus", value: "work", url: baseURL },
  ]);
});

for (const theme of ["light", "dark"] as const) {
  test(`self-check separates known, saved and skipped choices in ${theme} mode`, async ({ page, context, baseURL }, testInfo) => {
    await context.addCookies([{ name: "vocanova_theme", value: theme, url: baseURL! }]);
    const words = await (await context.request.get(`${mockAPI}/api/v1/canonical-words?knowledge=unexplored&limit=10`)).json();
    expect(words.items.length).toBeGreaterThanOrEqual(3);
    await page.goto("/vocabulary/check");
    const main = page.getByRole("main");
    await expect(main.getByRole("heading", { level: 1 })).toHaveText("Find your starting words");
    await expect(main.getByText("Starting with your focus:", { exact: false })).toHaveCount(0);
    await expect(main.getByRole("heading", { level: 2 })).toHaveText(words.items[0].wordText);
    const a11y = await scanForAxeViolations(page);
    expect(a11y.criticalOrSerious, formatViolations(a11y.criticalOrSerious).join("\n")).toEqual([]);
    expect(await page.evaluate(() => document.documentElement.scrollWidth)).toBeLessThanOrEqual(page.viewportSize()!.width);
    await page.screenshot({ path: testInfo.outputPath(`starting-words-${theme}.png`), fullPage: true });
    await main.getByRole("button", { name: "Already know", exact: true }).focus();
    await page.keyboard.press("Enter");
    await expect(main.getByRole("heading", { level: 2 })).toHaveText(words.items[1].wordText);
    await expect(main.getByRole("heading", { level: 2 })).toBeFocused();
    await main.getByRole("button", { name: "Want to learn", exact: true }).click();
    for (const word of words.items.slice(2)) {
      await expect(main.getByRole("heading", { level: 2 })).toHaveText(word.wordText);
      await main.getByRole("button", { name: "Skip for now", exact: true }).click();
    }
    await expect(main.getByRole("heading", { name: "Your starting words are ready" })).toBeFocused();
    await expect(main.getByText("In this check, you marked 1 meaning as already known and saved 1 to learn.", { exact: true })).toBeVisible();
    const summary = await (await context.request.get(`${mockAPI}/api/v1/knowledge-summary`)).json();
    expect(summary.selfReportedKnown).toBe(1);
    expect(summary.saved).toBe(1);
    expect(summary.mastered).toBe(0);
    const remaining = await (await context.request.get(`${mockAPI}/api/v1/canonical-words?knowledge=unexplored&limit=10`)).json();
    const remainingIds = remaining.items.map((word: { meaningId: string }) => word.meaningId);
    expect(remainingIds).toEqual(expect.arrayContaining(words.items.slice(2).map((word: { meaningId: string }) => word.meaningId)));
    expect(remainingIds).not.toContain(words.items[0].meaningId);
    expect(remainingIds).not.toContain(words.items[1].meaningId);
    await main.getByRole("link", { name: "See your learning plan", exact: true }).click();
    await expect(main.getByRole("heading", { level: 1 })).toHaveText("Your learning plan");
    await expect(page.getByRole("navigation", { name: "Primary" }).getByRole("link", { name: "Journey", exact: true })).toHaveAttribute("aria-current", "page");
    await expect(main.getByRole("link", { name: "Change daily preferences", exact: true })).toHaveAttribute("href", "/settings");
    await expect(main.getByRole("link", { name: /^Start lesson\s*:\s*Make a plan with a friend$/ })).toHaveAttribute("href", "/learn/conversation-basics");
    const planA11y = await scanForAxeViolations(page);
    expect(planA11y.criticalOrSerious, formatViolations(planA11y.criticalOrSerious).join("\n")).toEqual([]);
    expect(await page.evaluate(() => document.documentElement.scrollWidth)).toBeLessThanOrEqual(page.viewportSize()!.width);
    await page.screenshot({ path: testInfo.outputPath(`learning-plan-${theme}.png`), fullPage: true });
  });
}

test("lost self-check response retries its exact choice without overwriting a private note", async ({ page, context }) => {
  const csrf = (await context.cookies()).find((cookie) => cookie.name === "vocanova_csrf")!.value;
  const firstWord = (await (await context.request.get(`${mockAPI}/api/v1/canonical-words?knowledge=unexplored&limit=1`)).json()).items[0];
  const note = "My private example must stay unchanged.";
  expect((await context.request.put(`${mockAPI}/api/v1/meaning-knowledge/${firstWord.meaningId}`, { headers: { "X-CSRF-Token": csrf, "Idempotency-Key": randomUUID() }, data: { selfReportedKnown: false, note } })).ok()).toBe(true);
  const requests: { key?: string; body: string | null }[] = [];
  await page.route("**/api/v1/meaning-knowledge/*", async (route) => {
    if (route.request().method() !== "PATCH") return route.continue();
    requests.push({ key: route.request().headers()["idempotency-key"], body: route.request().postData() });
    if (requests.length === 1) { expect((await route.fetch()).ok()).toBe(true); return route.abort("failed"); }
    return route.continue();
  });
  await page.goto("/vocabulary/check");
  const main = page.getByRole("main");
  await expect(main.getByRole("heading", { level: 2 })).toHaveText(firstWord.wordText);
  await main.getByRole("button", { name: "Already know", exact: true }).click();
  await expect(main.getByRole("alert")).toContainText("could not confirm");
  await expect(main.getByRole("heading", { level: 2 })).toHaveText(firstWord.wordText);
  for (const name of ["Already know", "Want to learn", "Skip for now"]) await expect(main.getByRole("button", { name, exact: true })).toBeDisabled();
  await main.getByRole("button", { name: "Retry choice", exact: true }).click();
  await expect(main.getByRole("heading", { level: 2 })).not.toHaveText(firstWord.wordText);
  expect(requests).toHaveLength(2);
  expect(requests[0]!.key).toBeTruthy();
  expect(requests[1]).toEqual(requests[0]);
  expect(JSON.parse(requests[0]!.body!)).toEqual({ selfReportedKnown: true });
  const persisted = await (await context.request.get(`${mockAPI}/api/v1/meaning-knowledge/${firstWord.meaningId}`)).json();
  expect(persisted).toMatchObject({ selfReportedKnown: true, note });
});

test("self-check handles an empty unexplored catalog without offering fake choices", async ({ page, context }) => {
  const csrf = (await context.cookies()).find((cookie) => cookie.name === "vocanova_csrf")!.value;
  const words = await (await context.request.get(`${mockAPI}/api/v1/canonical-words?limit=50`)).json();
  for (const word of words.items) {
    expect((await context.request.patch(`${mockAPI}/api/v1/meaning-knowledge/${word.meaningId}`, { headers: { "X-CSRF-Token": csrf, "Idempotency-Key": randomUUID() }, data: { selfReportedKnown: true } })).ok()).toBe(true);
  }
  await page.goto("/vocabulary/check");
  const main = page.getByRole("main");
  await expect(main.getByRole("heading", { name: "You have explored these words" })).toBeVisible();
  await expect(main.getByRole("button", { name: "Already know", exact: true })).toHaveCount(0);
  await expect(main.getByRole("link", { name: "See your learning plan", exact: true })).toBeVisible();
});

for (const path of ["/plan", "/vocabulary/check"]) {
  test(`${path} requires a signed-in learner`, async ({ page, context, baseURL }) => {
    await context.addCookies([{ name: "e2e_unauthenticated", value: "1", url: baseURL! }]);
    await page.goto(path);
    await expect(page).toHaveURL(new RegExp(`/login\\?returnTo=${encodeURIComponent(path)}$`));
  });
}
