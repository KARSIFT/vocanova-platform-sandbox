import { randomUUID } from "node:crypto";

import { expect, test, type BrowserContext, type Page } from "@playwright/test";

const PREFIX = "vocanova:preferences-draft:";

async function authenticate(context: BrowserContext, baseURL?: string) {
  if (!baseURL) throw new Error("A test app URL is required");
  await context.addCookies([
    { name: "vocanova_session", value: randomUUID(), url: baseURL },
    { name: "vocanova_csrf", value: randomUUID(), url: baseURL },
  ]);
}

async function preferenceDraftCount(page: Page) {
  return page.evaluate((prefix) => Object.keys(sessionStorage).filter((key) => key.startsWith(prefix)).length, PREFIX);
}

async function resumeSameLearner(page: Page, context: BrowserContext, baseURL: string | undefined, returnTo: string) {
  await expect(page).toHaveURL(/\/login\?.*reason=session-expired/);
  // Synthetic reauthentication: retain the fixture's identity and replace
  // the browser-readable CSRF token removed by the real expiry handler.
  await context.addCookies([{ name: "vocanova_csrf", value: randomUUID(), url: baseURL! }]);
  await page.goto(returnTo);
}

for (const theme of ["light", "dark"] as const) {
  test(`onboarding restores answers and final step after full session-expiry navigation (${theme})`, async ({ page, context, baseURL }, testInfo) => {
    await authenticate(context, baseURL);
    await context.addCookies([
      { name: "e2e_onboarding_status", value: "not_started", url: baseURL! },
      { name: "vocanova_theme", value: theme, url: baseURL! },
    ]);
    let expires = true;
    await page.route("**/api/v1/onboarding", async (route) => {
      if (route.request().method() !== "POST") return route.continue();
      if (expires) {
        expires = false;
        return route.fulfill({ status: 401, contentType: "application/json", body: JSON.stringify({ error: "unauthorized" }) });
      }
      const response = await route.fetch();
      await context.clearCookies({ name: "e2e_onboarding_status" });
      await route.fulfill({ response });
    });
    await page.goto("/onboarding");
    await page.getByRole("radio", { name: /B2 — Upper-intermediate/ }).check();
    await page.getByRole("button", { name: "Continue", exact: true }).click();
    const language = "LongCustomLanguageName".repeat(3);
    await page.getByRole("textbox", { name: "Native language" }).fill(language);
    await page.getByRole("button", { name: "Continue", exact: true }).click();
    await page.getByRole("radio", { name: "Conversation", exact: true }).check();
    await page.getByRole("button", { name: "Continue", exact: true }).click();
    await page.getByRole("radio", { name: "Social", exact: true }).check();
    await page.getByRole("button", { name: "Continue", exact: true }).click();
    await page.getByRole("radio", { name: "20", exact: true }).press("Space");
    await page.getByRole("button", { name: "Finish setup", exact: true }).click();
    await resumeSameLearner(page, context, baseURL, "/onboarding");
    const summary = page.getByRole("complementary", { name: "Onboarding summary" });
    await expect(summary).toContainText(language);
    await expect(summary).toContainText("B2 — Upper-intermediate");
    await expect(summary).toContainText("Conversation");
    await expect(summary).toContainText("Social");
    await expect(page.getByRole("radio", { name: "20", exact: true })).toBeChecked();
    await expect(page.getByRole("status")).toContainText("restored");
    await page.addStyleTag({ content: "html { font-size: 200% !important; }" });
    expect(await page.evaluate(() => document.documentElement.scrollWidth)).toBeLessThanOrEqual(page.viewportSize()!.width);
    await page.screenshot({ path: testInfo.outputPath(`onboarding-restored-${theme}.png`) });
    await page.getByRole("button", { name: "Finish setup", exact: true }).click();
    await expect(page).toHaveURL(/\/home(\?|$)/);
    await expect.poll(() => preferenceDraftCount(page)).toBe(0);
  });

  test(`settings retain unsaved changes through full session-expiry navigation and clear on save (${theme})`, async ({ page, context, baseURL }) => {
    await authenticate(context, baseURL);
    await context.addCookies([{ name: "vocanova_theme", value: theme, url: baseURL! }]);
    let expires = true;
    await page.route("**/api/v1/settings", async (route) => {
      if (route.request().method() === "PATCH" && expires) {
        expires = false;
        return route.fulfill({ status: 401, contentType: "application/json", body: JSON.stringify({ error: "unauthorized" }) });
      }
      await route.continue();
    });
    await page.goto("/settings");
    const name = page.getByRole("textbox", { name: "Display name" });
    await name.fill("Recovered learner");
    await page.getByRole("radio", { name: "20", exact: true }).press("Space");
    const emails = page.getByRole("switch", { name: "Product news and tips", exact: true });
    await emails.click();
    await page.getByRole("button", { name: "Save settings", exact: true }).click();
    await resumeSameLearner(page, context, baseURL, "/settings");
    await expect(name).toHaveValue("Recovered learner");
    await expect(page.getByRole("radio", { name: "20", exact: true })).toBeChecked();
    await expect(emails).toHaveAttribute("aria-checked", "true");
    await expect(page.getByRole("form", { name: "Practice settings" }).getByRole("status")).toContainText("unsaved changes");
    await page.addStyleTag({ content: "html { font-size: 200% !important; }" });
    expect(await page.evaluate(() => document.documentElement.scrollWidth)).toBeLessThanOrEqual(page.viewportSize()!.width);
    await expect(page.getByText("Your settings have been saved.")).toHaveCount(0);
    await page.getByRole("button", { name: "Save settings", exact: true }).click();
    await expect(page.getByText("Your settings have been saved.")).toBeVisible();
    await expect.poll(() => preferenceDraftCount(page)).toBe(0);
    await page.reload();
    await expect(name).toHaveValue("Recovered learner");
    await expect(page.getByText(/unsaved changes were restored/)).toHaveCount(0);
  });
}

test("a reverted value survives reload after an earlier save commits but loses its response", async ({ page, context, baseURL }) => {
  await authenticate(context, baseURL);
  let committed = false;
  let release: (() => void) | undefined;
  const responsePending = new Promise<void>((resolve) => { release = resolve; });
  let first = true;
  await page.route("**/api/v1/settings", async (route) => {
    if (route.request().method() === "PATCH" && first) {
      first = false;
      const response = await route.fetch();
      expect(response.ok()).toBe(true);
      committed = true;
      await responsePending;
      return route.abort("failed");
    }
    await route.continue();
  });
  await page.goto("/settings");
  const name = page.getByRole("textbox", { name: "Display name" });
  const originalName = await name.inputValue();
  await name.fill("Earlier saved name");
  await page.getByRole("button", { name: "Save settings", exact: true }).click();
  await expect.poll(() => committed).toBe(true);
  try {
    await name.fill(originalName);
  } finally {
    release?.();
  }
  await expect(page.getByRole("form", { name: "Practice settings" }).getByRole("alert")).toBeVisible();
  await page.reload();
  await expect(name).toHaveValue(originalName);
  await expect(page.getByRole("form", { name: "Practice settings" }).getByRole("status")).toContainText("unsaved changes");
  await page.getByRole("button", { name: "Save settings", exact: true }).click();
  await expect(page.getByText("Your settings have been saved.")).toBeVisible();
  await page.reload();
  await expect(name).toHaveValue(originalName);
});

test("onboarding locks final answers while the submitted setup is pending", async ({ page, context, baseURL }) => {
  await authenticate(context, baseURL);
  await context.addCookies([{ name: "e2e_onboarding_status", value: "not_started", url: baseURL! }]);
  let release: (() => void) | undefined;
  const pending = new Promise<void>((resolve) => { release = resolve; });
  let submitted: Record<string, unknown> | undefined;
  await page.route("**/api/v1/onboarding", async (route) => {
    if (route.request().method() !== "POST") return route.continue();
    submitted = route.request().postDataJSON();
    await pending;
    const response = await route.fetch();
    await context.clearCookies({ name: "e2e_onboarding_status" });
    await route.fulfill({ response });
  });
  await page.goto("/onboarding");
  await page.getByRole("radio", { name: /A2 — Elementary/ }).check();
  await page.getByRole("button", { name: "Continue", exact: true }).click();
  await page.getByRole("textbox", { name: "Native language" }).fill("Persian");
  await page.getByRole("button", { name: "Continue", exact: true }).click();
  await page.getByRole("radio", { name: "Conversation", exact: true }).check();
  await page.getByRole("button", { name: "Continue", exact: true }).click();
  await page.getByRole("radio", { name: "Social", exact: true }).check();
  await page.getByRole("button", { name: "Continue", exact: true }).click();
  await page.getByRole("radio", { name: "20", exact: true }).press("Space");
  await page.getByRole("button", { name: "Finish setup", exact: true }).click();
  await expect.poll(() => submitted?.dailyReviewTarget).toBe(20);
  try {
    await expect(page.getByRole("radio", { name: "30", exact: true })).toBeDisabled();
    await expect(page.getByRole("radio", { name: "20", exact: true })).toBeChecked();
    await expect(page.getByRole("button", { name: "Back", exact: true })).toBeDisabled();
  } finally {
    release?.();
  }
  await expect(page).toHaveURL(/\/home(\?|$)/);
  await expect.poll(() => preferenceDraftCount(page)).toBe(0);
});

test("changing accounts never restores another learner's preferences and logout clears drafts", async ({ page, context, baseURL }) => {
  await authenticate(context, baseURL);
  await page.goto("/settings");
  const name = page.getByRole("textbox", { name: "Display name" });
  await name.fill("Private unsaved name");
  await expect.poll(() => preferenceDraftCount(page)).toBe(1);
  await context.addCookies([{ name: "e2e_user_id", value: "alternate", url: baseURL! }]);
  await page.reload();
  await expect(name).not.toHaveValue("Private unsaved name");
  await expect.poll(() => preferenceDraftCount(page)).toBe(0);
  await name.fill("Another private name");
  await page.getByRole("button", { name: "Log out", exact: true }).click();
  await expect(page).toHaveURL(/\/signin(\?|$)|\/login(\?|$)/);
  await expect.poll(() => preferenceDraftCount(page)).toBe(0);
});
