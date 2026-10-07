import { randomUUID } from "node:crypto";
import { expect, test } from "@playwright/test";
import type { LessonSession } from "@vocanova/api-client";
import { formatViolations, scanForAxeViolations } from "./axe-helper";

const apiURL = `http://127.0.0.1:${process.env.MOCK_API_PORT ?? 8080}`;

test.beforeEach(async ({ context, baseURL }) => {
  if (!baseURL) throw new Error("Missing app URL");
  await context.addCookies([
    { name: "vocanova_session", value: randomUUID(), url: baseURL },
    { name: "vocanova_csrf", value: randomUUID(), url: baseURL },
  ]);
});

for (const theme of ["light", "dark"] as const) {
  test(`Journey connects activities without a second catalogue in ${theme}`, async ({ page, context, baseURL }, testInfo) => {
    await context.addCookies([{ name: "vocanova_theme", value: theme, url: baseURL! }]);
    await page.goto("/discover");
    const activities = page.getByRole("navigation", { name: "Ways to practise" });
    for (const [name, href] of [["Choose a practice", "/practice"], ["Topic writing", "/writing"], ["Short stories", "/stories"]] as const) {
      const entry = activities.getByRole("link", { name: new RegExp(name) });
      await expect(entry).toBeVisible();
      await expect(entry).toHaveAttribute("href", href);
      await entry.focus();
      await expect(entry).toBeFocused();
      expect((await entry.boundingBox())!.height).toBeGreaterThanOrEqual(44);
    }
    const situations = page.getByRole("region", { name: "Guided lessons", exact: true });
    await expect(situations.getByRole("heading", { name: "Explore by situation" })).toBeVisible();
    await expect(situations.locator("summary").filter({ hasText: "Ordering at a cafe" })).toHaveCount(1);
    await expect(page.getByRole("heading", { name: "Explore by situation" })).toHaveCount(1);
    expect((await activities.boundingBox())!.y).toBeLessThan((await situations.boundingBox())!.y);
    expect(await page.evaluate(() => document.documentElement.scrollWidth)).toBeLessThanOrEqual(page.viewportSize()!.width);
    const scan = await scanForAxeViolations(page);
    expect(scan.criticalOrSerious, formatViolations(scan.criticalOrSerious).join("\n")).toEqual([]);
    await page.screenshot({ path: testInfo.outputPath(`journey-connected-${theme}.png`), fullPage: true });
  });
}

test("Journey resumes the actual saved lesson and recommendation expiry redirects", async ({ page, context, baseURL }) => {
  const csrf = (await context.cookies()).find(cookie => cookie.name === "vocanova_csrf")!.value;
  const started = await context.request.post(apiURL + "/api/v1/lessons/conversation-basics/sessions", {
    headers: { "X-CSRF-Token": csrf, "Idempotency-Key": randomUUID() },
    data: {},
  });
  expect(started.ok()).toBe(true);
  const session = await started.json() as LessonSession;
  await page.goto("/discover");
  const recommendation = page.getByRole("region", { name: "Pick up your lesson", exact: true });
  await expect(recommendation.getByText(`${session.completedSteps} of ${session.totalSteps} steps complete.`, { exact: true })).toBeVisible();
  await recommendation.getByRole("link", { name: /^Continue lesson\s*:/ }).click();
  await expect(page).toHaveURL(/\/learn\/conversation-basics$/);
  await expect(page.getByRole("heading", { name: "invite", exact: true })).toBeVisible();
  await context.addCookies([{ name: "e2e_recommendation_auth", value: "expired", url: baseURL! }]);
  await page.goto("/discover");
  await expect(page).toHaveURL(/\/login\?returnTo=%2Fdiscover$/);
});

test("Journey recommendation failure leads to the available situation catalogue", async ({ page, context, baseURL }) => {
  await context.addCookies([{ name: "e2e_lessons", value: "unavailable", url: baseURL! }]);
  await page.goto("/discover");
  const recommendation = page.getByRole("region", { name: "Your next lesson", exact: true });
  await expect(recommendation.getByRole("status")).toContainText("recommendation is unavailable");
  const browse = recommendation.getByRole("link", { name: "Explore situations", exact: true });
  await expect(browse).toHaveAttribute("href", "#journey-lessons");
  await browse.focus();
  await page.keyboard.press("Enter");
  await expect(page).toHaveURL(/\/discover#journey-lessons$/);
  const catalogue = page.locator("#journey-lessons");
  await expect(catalogue.getByRole("heading", { name: "Explore by situation" })).toBeVisible();
  await expect.poll(async () => {
    const box = await catalogue.boundingBox();
    return Boolean(box && box.y >= 64 && box.y < page.viewportSize()!.height);
  }).toBe(true);
  await expect(catalogue.locator("summary").filter({ hasText: "Ordering at a cafe" })).toHaveCount(1);
});

test("Practice defaults to taught vocabulary while full course stays an explicit choice", async ({ page, context }) => {
  const csrf = (await context.cookies()).find(cookie => cookie.name === "vocanova_csrf")!.value;
  const headers = { "X-CSRF-Token": csrf };
  const start = await context.request.post(apiURL + "/api/v1/lessons/conversation-basics/sessions", {
    headers: { ...headers, "Idempotency-Key": randomUUID() }, data: {},
  });
  expect(start.ok()).toBe(true);
  let session = await start.json() as LessonSession;
  await page.goto("/practice");
  await expect(page.getByLabel("Practice vocabulary", { exact: true })).toHaveValue("");
  // Opening a session alone must not be described as studying its vocabulary.
  for (let index = 0; index < session.words.length; index++) {
    const advanced = await context.request.post(`${apiURL}/api/v1/lesson-sessions/${session.id}/actions`, {
      headers: { ...headers, "Idempotency-Key": randomUUID() },
      data: { action: "continue", clientActionId: randomUUID(), stepId: session.currentStep!.id, expectedRevision: session.revision },
    });
    expect(advanced.ok()).toBe(true);
    session = await advanced.json() as LessonSession;
  }
  await page.reload();
  // Next may retain an inactive subtree during hydration. Interact with the
  // accessible practice region rather than matching its hidden DOM copy.
  const practice = page.getByRole("region", { name: "Remember your words", exact: true });
  const selection = practice.getByLabel("Practice vocabulary", { exact: true });
  await expect(selection).toHaveValue("conversation-basics");
  await selection.selectOption("");
  await practice.locator("summary").filter({ hasText: "About this selection" }).focus();
  await page.keyboard.press("Enter");
  await expect(practice.getByText("A full-course mix can include words you have not studied yet.", { exact: false })).toBeVisible();
  const sent = page.waitForRequest(request => request.method() === "POST" && new URL(request.url()).pathname === "/api/v1/practice-sessions");
  await page.getByRole("button", { name: "Start typed recall", exact: true }).click();
  expect((await sent).postDataJSON()).toEqual({ mode: "typed_recall" });
  await expect(page).toHaveURL(/\/practice\/session\/[^/]+$/);
});
