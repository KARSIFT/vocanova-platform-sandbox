import { randomUUID } from "node:crypto";
import { expect, test, type BrowserContext } from "@playwright/test";
import { formatViolations, scanForAxeViolations } from "./axe-helper";

const apiURL = `http://127.0.0.1:${process.env.MOCK_API_PORT ?? 8080}`;
test("Home redirects when its recommendation read sees an expired session", async ({ page, context, baseURL }) => {
  await context.addCookies([{ name: "e2e_recommendation_auth", value: "expired", url: baseURL! }]);
  await page.goto("/home");
  await expect(page).toHaveURL(/\/login\?returnTo=%2Fhome$/);
});
test.beforeEach(async ({ context, baseURL }) => {
  await context.addCookies([
    { name: "vocanova_session", value: randomUUID(), url: baseURL! },
    { name: "vocanova_csrf", value: randomUUID(), url: baseURL! },
  ]);
});
async function markTargetsKnown(context: BrowserContext) {
  const csrf = (await context.cookies()).find(cookie => cookie.name === "vocanova_csrf")!.value;
  for (const slug of ["invite", "confirm", "reschedule"]) {
    const { word } = await (await context.request.get(`${apiURL}/api/v1/canonical-words/${slug}`)).json();
    expect((await context.request.patch(`${apiURL}/api/v1/meaning-knowledge/${word.meanings[0].id}`, {
      headers: { "X-CSRF-Token": csrf, "Idempotency-Key": randomUUID() },
      data: { selfReportedKnown: true },
    })).ok()).toBe(true);
  }
}
for (const theme of ["light", "dark"] as const) {
  test(`Home and plan use actual target coverage without awarding completion (${theme})`, async ({ page, context, baseURL }) => {
    await context.addCookies([{ name: "vocanova_theme", value: theme, url: baseURL! }]);
    await page.goto("/plan");
    await expect(page.getByRole("main").getByRole("link", { name: /^Start lesson\s*:/ })).toBeVisible();
    await markTargetsKnown(context);
    for (const path of ["/plan", "/home"]) {
      await page.goto(path);
      const region = page.getByRole("main").getByRole("region", { name: "Your next lesson", exact: true });
      await expect(region.getByText(/You have marked the remaining lesson meanings as known/)).toBeVisible();
      await expect(region.getByRole("link", { name: /^Start lesson/ })).toHaveCount(0);
      await expect(region.getByRole("link", { name: "All lessons", exact: true })).toHaveAttribute("href", "/discover");
      await expect(page.locator("html")).toHaveAttribute("data-theme", theme);
      expect(await page.evaluate(() => document.documentElement.scrollWidth)).toBeLessThanOrEqual(page.viewportSize()!.width);
      const scan = await scanForAxeViolations(page);
      expect(scan.criticalOrSerious, formatViolations(scan.criticalOrSerious).join("\n")).toEqual([]);
    }
    const catalog = await (await context.request.get(apiURL + "/api/v1/lessons")).json();
    expect(catalog.items[0]).toMatchObject({ status: "not_started", completedSteps: 0 });
  });
}
test("an unfinished lesson stays resumable when its words are marked known", async ({ page, context }) => {
  const csrf = (await context.cookies()).find(cookie => cookie.name === "vocanova_csrf")!.value;
  expect((await context.request.post(apiURL + "/api/v1/lessons/conversation-basics/sessions", {
    headers: { "X-CSRF-Token": csrf, "Idempotency-Key": randomUUID() }, data: {},
  })).ok()).toBe(true);
  await markTargetsKnown(context);
  await page.goto("/plan");
  const region = page.getByRole("main").getByRole("region", { name: "Pick up your lesson" });
  await expect(region.getByText("0 of 9 steps saved. Continue where you left off.", { exact: true })).toBeVisible();
  await region.getByRole("link", { name: /^Continue lesson\s*:/ }).click();
  await expect(page.getByRole("main").getByRole("heading", { name: "invite", exact: true })).toBeVisible();
});
test("recommendation failure leaves browsing available without claiming all words are known", async ({ page, context, baseURL }) => {
  await context.addCookies([{ name: "e2e_lessons", value: "unavailable", url: baseURL! }]);
  await page.goto("/plan");
  const region = page.getByRole("main").getByRole("region", { name: "Your next lesson" });
  await expect(region.getByRole("status")).toHaveText("Your lesson recommendation is unavailable right now. You can still browse lessons or review your words.");
  await expect(region.getByRole("link", { name: "All lessons", exact: true })).toBeVisible();
  await expect(region.getByText(/You have marked the remaining/)).toHaveCount(0);
});
