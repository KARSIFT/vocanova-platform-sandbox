import { randomUUID } from "node:crypto";
import { expect, test, type Locator, type Page, type Request } from "@playwright/test";
import { pressTab } from "./keyboard-navigation.js";

async function tabTo(page: Page, target: Locator) {
  for (let step = 0; step < 60; step++) {
    if (await target.evaluate((element) => element === document.activeElement)) return;
    await pressTab(page);
  }
  throw new Error("Control could not be reached using Tab");
}
async function keyboardActivate(page: Page, target: Locator) {
  await tabTo(page, target);
  await page.keyboard.press("Enter");
}

for (const theme of ["light", "dark"] as const) {
  test(`onboarding completes using only keyboard controls (${theme})`, async ({ page, context, baseURL }) => {
    await context.addCookies([
      { name: "vocanova_session", value: randomUUID(), url: baseURL! },
      { name: "vocanova_csrf", value: randomUUID(), url: baseURL! },
      { name: "e2e_onboarding_status", value: "not_started", url: baseURL! },
    ]);
    await page.addInitScript((value) => localStorage.setItem("vocanova:theme-preference", value), theme);
    let submitted: unknown;
    await page.route("**/api/v1/onboarding", async (route) => {
      if (route.request().method() !== "POST") return route.continue();
      submitted = route.request().postDataJSON();
      const response = await route.fetch();
      expect(response.status()).toBe(200);
      await context.clearCookies({ name: "e2e_onboarding_status" });
      await route.fulfill({ response });
    });
    await page.goto("/onboarding");
    await expect(page.locator("html")).toHaveAttribute("data-theme", theme);
    await tabTo(page, page.getByRole("radio", { name: /A1/ }));
    await page.keyboard.press("ArrowDown");
    await expect(page.getByRole("radio", { name: /A2/ })).toBeChecked();
    const next = page.getByRole("button", { name: "Continue", exact: true });
    await keyboardActivate(page, next);
    const language = page.getByRole("textbox", { name: "Native language", exact: true });
    await tabTo(page, language);
    await page.keyboard.press("ControlOrMeta+A");
    await page.keyboard.type("Spanish");
    await keyboardActivate(page, next);
    await keyboardActivate(page, page.getByRole("button", { name: "Back", exact: true }));
    await expect(language).toHaveValue("Spanish");
    await keyboardActivate(page, next);
    await tabTo(page, page.getByRole("radio", { name: "General growth", exact: true }));
    await page.keyboard.press("Space");
    await keyboardActivate(page, next);
    await tabTo(page, page.getByRole("radio", { name: "Daily life", exact: true }));
    await page.keyboard.press("Space");
    await keyboardActivate(page, next);
    await tabTo(page, page.getByRole("radio", { name: "5", exact: true }));
    await page.keyboard.press("ArrowRight");
    await page.keyboard.press("ArrowRight");
    await expect(page.getByRole("radio", { name: "15", exact: true })).toBeChecked();
    await keyboardActivate(page, page.getByRole("button", { name: "Finish setup", exact: true }));
    await expect(page).toHaveURL(/\/home$/);
    expect(submitted).toMatchObject({ englishLevel: "a2", nativeLanguage: "es", learningGoal: "general", mainUseCase: "daily_life", dailyReviewTarget: 15 });
    await expect(page.getByRole("heading", { name: "Today", exact: true })).toBeVisible();
  });
}

test("malformed direct Journey bands select the safe first range and links use canonical values", async ({ page, context, baseURL }) => {
  await context.addCookies([{ name: "vocanova_session", value: randomUUID(), url: baseURL! }]);
  for (const query of ["band=0", "band=-1", "band=101", "band=1.5", "band=invalid", "band=2&band=3"]) {
    await page.goto(`/discover?${query}`);
    await expect(page.getByRole("heading", { name: "Words 1–100", exact: true })).toBeVisible();
    const next = page.getByRole("navigation", { name: "Adjacent word ranges" }).getByRole("link", { name: "Next 100", exact: true });
    await expect(next).toHaveAttribute("href", "/discover?band=2#range-words");
    await expect(page.getByRole("navigation", { name: "Adjacent word ranges" }).getByRole("link", { name: "Previous 100", exact: true })).toHaveCount(0);
  }
  await keyboardActivate(page, page.getByRole("navigation", { name: "Adjacent word ranges" }).getByRole("link", { name: "Next 100", exact: true }));
  await expect(page).toHaveURL(/\/discover\?band=2#range-words$/);
  await expect(page.getByRole("heading", { name: "Words 101–200", exact: true })).toBeVisible();
});

test("a late older word-search response cannot replace the newer search", async ({ page, context, baseURL }) => {
  await context.addCookies([{ name: "vocanova_session", value: randomUUID(), url: baseURL! }]);
  await page.goto("/vocabulary");
  let release!: () => void;
  let arrived!: () => void;
  let settled!: () => void;
  let browserSettled!: () => void;
  let olderRequest: Request | undefined;
  const held = new Promise<void>((resolve) => { release = resolve; });
  const started = new Promise<void>((resolve) => { arrived = resolve; });
  const completed = new Promise<void>((resolve) => { settled = resolve; });
  const browserCompleted = new Promise<void>((resolve) => { browserSettled = resolve; });
  const observeCompletion = (request: Request) => {
    if (request === olderRequest) browserSettled();
  };
  page.on("requestfinished", observeCompletion);
  page.on("requestfailed", observeCompletion);
  await page.route("**/vocabulary?**", async (route) => {
    if (new URL(route.request().url()).searchParams.get("q") !== "pour") return route.continue();
    olderRequest = route.request();
    try {
      const response = await route.fetch();
      arrived();
      await held;
      await route.fulfill({ response });
    } finally { settled(); }
  });
  const input = page.getByRole("searchbox", { name: "Search words and meanings", exact: true });
  await input.fill("pour");
  await page.getByRole("button", { name: "Search", exact: true }).click();
  await started;
  try {
    await input.fill("invite");
    await page.getByRole("button", { name: "Search", exact: true }).click();
    await expect.poll(() => new URL(page.url()).searchParams.get("q")).toBe("invite");
    await expect(page.getByRole("link", { name: "invite", exact: true })).toBeVisible();
  } finally { release(); }
  await completed;
  // Fulfilling the route alone only hands bytes to the browser. Wait until
  // that request has finished (or the newer navigation canceled it), then
  // allow the browser to paint any React update before inspecting the result.
  await browserCompleted;
  await page.evaluate(() => new Promise<void>((resolve) => {
    requestAnimationFrame(() => requestAnimationFrame(() => resolve()));
  }));
  page.off("requestfinished", observeCompletion);
  page.off("requestfailed", observeCompletion);
  await expect(input).toHaveValue("invite");
  await expect(page.getByRole("link", { name: "invite", exact: true })).toBeVisible();
  expect(new URL(page.url()).searchParams.get("q")).toBe("invite");
  await expect(page.getByRole("link", { name: "pour", exact: true })).toHaveCount(0);
});
