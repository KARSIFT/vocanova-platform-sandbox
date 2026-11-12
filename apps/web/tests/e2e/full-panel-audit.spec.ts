import { randomUUID } from "node:crypto";
import { readFile } from "node:fs/promises";
import { expect, test } from "@playwright/test";
import type { BrowserContext, Page } from "@playwright/test";
import { formatViolations, scanForAxeViolations } from "./axe-helper";

async function signIn(context: BrowserContext, baseURL: string) {
  await context.addCookies([
    { name: "vocanova_session", value: randomUUID(), url: baseURL },
    { name: "vocanova_csrf", value: randomUUID(), url: baseURL },
  ]);
}

async function checkLayout(page: Page) {
  expect(await page.evaluate(() => document.documentElement.scrollWidth))
    .toBeLessThanOrEqual(page.viewportSize()!.width);
  const scan = await scanForAxeViolations(page);
  expect(scan.criticalOrSerious, formatViolations(scan.criticalOrSerious).join("\n")).toEqual([]);
}

const exportFixture = {
  schemaVersion: "1.0",
  profile: { displayName: "Synthetic audit learner" },
  settings: { dailyReviewTarget: 10 },
  onboardingProfile: null,
  savedWords: [{ word: "reservation" }],
  reviewHistory: [],
  sentenceFeedbackHistory: [],
  dailyMissions: [],
  dailyActivity: [],
  confidencePointLedger: [],
  graceDayLedger: [],
  streakState: null,
};

for (const theme of ["light", "dark"] as const) {
  test(`Help disclosures and learning destinations work with the keyboard (${theme})`, async ({ page, context, baseURL }, testInfo) => {
    await signIn(context, baseURL!);
    await context.addCookies([{ name: "vocanova_theme", value: theme, url: baseURL! }]);
    await page.goto("/help");
    await expect(page.getByRole("heading", { name: "How can we help?" })).toBeVisible();
    const disclosures = page.locator("main details");
    await expect(disclosures).toHaveCount(8);
    for (const disclosure of await disclosures.all()) {
      const summary = disclosure.locator("summary");
      if (await disclosure.getAttribute("open") !== null) {
        await summary.focus();
        await page.keyboard.press("Enter");
      }
      await expect(disclosure).not.toHaveAttribute("open");
      await summary.focus();
      await page.keyboard.press("Space");
      await expect(disclosure).toHaveAttribute("open", "");
      await expect(summary).toBeFocused();
      await expect(disclosure.locator("p").first()).toBeVisible();
      await page.keyboard.press("Enter");
      await expect(disclosure).not.toHaveAttribute("open");
      await expect(summary).toBeFocused();
      await page.keyboard.press("Space");
      await expect(disclosure).toHaveAttribute("open", "");
    }
    await checkLayout(page);
    await page.screenshot({ path: testInfo.outputPath(`help-${theme}.png`), fullPage: true });
    // Exercise the public help page's actual links rather than only comparing hrefs.
    for (const [label, destination] of [
      ["Open your learning plan", "/plan"],
      ["See your progress", "/progress"],
      ["Explore your word map", "/discover"],
      ["Choose a practice", "/practice"],
      ["Set up a calendar reminder", "/settings#calendar-reminder-heading"],
      ["Open account security", "/settings/account"],
    ]) {
      await page.goto("/help");
      const link = page.getByRole("link", { name: label, exact: true });
      // Closed native details hide their links from the accessibility tree.
      // Locate the containing disclosure before opening it, then use the visible link.
      const disclosure = page.locator("details").filter({
        has: page.getByRole("link", { name: label, exact: true, includeHidden: true }),
      });
      if (await disclosure.getAttribute("open") === null) {
        await disclosure.locator("summary").focus();
        await page.keyboard.press("Enter");
      }
      await expect(link).toBeVisible();
      await link.focus();
      await page.keyboard.press("Enter");
      await expect(page).toHaveURL(`${baseURL}${destination}`);
      await expect(page.getByRole("main")).toBeVisible();
    }
  });

  test(`personal data export waits for confirmation and downloads usable JSON (${theme})`, async ({ page, context, baseURL }, testInfo) => {
    await signIn(context, baseURL!);
    await context.addCookies([{ name: "vocanova_theme", value: theme, url: baseURL! }]);
    let release!: () => void;
    const pending = new Promise<void>((resolve) => { release = resolve; });
    let requests = 0;
    await page.route("**/api/v1/personal-data-export", async (route) => {
      requests += 1;
      expect(route.request().method()).toBe("POST");
      expect(route.request().headers()["x-csrf-token"]).toBeTruthy();
      expect(route.request().headers()["idempotency-key"]).toBeTruthy();
      await pending;
      await route.fulfill({ json: exportFixture });
    });
    await page.goto("/settings/account");
    const section = page.getByRole("region", { name: "Download your data" });
    const button = section.getByRole("button");
    const downloaded = page.waitForEvent("download");
    await button.focus();
    await page.keyboard.press("Enter");
    try {
      await expect(button).toBeDisabled();
      await expect(button).toHaveAccessibleName("Preparing download...");
      await expect(section.getByRole("status")).toHaveCount(0);
      await expect.poll(() => requests).toBe(1);
    } finally {
      release();
    }
    const download = await downloaded;
    expect(download.suggestedFilename()).toMatch(/^vocanova-personal-data-\d{4}-\d{2}-\d{2}\.json$/);
    expect(await download.failure()).toBeNull();
    const path = await download.path();
    expect(path).not.toBeNull();
    expect(JSON.parse(await readFile(path!, "utf8"))).toEqual(exportFixture);
    await expect(section.getByRole("status")).toHaveText("Your download has started.");
    await expect(button).toBeEnabled();
    await checkLayout(page);
    await section.screenshot({ path: testInfo.outputPath(`export-${theme}.png`) });
    expect(requests).toBe(1);
  });
}

test("failed data export offers a working retry without claiming a download", async ({ page, context, baseURL }) => {
  await signIn(context, baseURL!);
  let requests = 0;
  let downloads = 0;
  page.on("download", () => { downloads += 1; });
  await page.route("**/api/v1/personal-data-export", async (route) => {
    requests += 1;
    await route.fulfill(requests === 1
      ? { status: 503, json: { error: { code: "unavailable", message: "Export temporarily unavailable" } } }
      : { json: exportFixture });
  });
  await page.goto("/settings/account");
  const section = page.getByRole("region", { name: "Download your data" });
  await section.getByRole("button").click();
  await expect(section.getByRole("alert")).toBeVisible();
  await expect(section.getByRole("status")).toHaveCount(0);
  expect(downloads).toBe(0);
  const downloaded = page.waitForEvent("download");
  await section.getByRole("button").click();
  await downloaded;
  await expect(section.getByRole("alert")).toHaveCount(0);
  await expect(section.getByRole("status")).toHaveText("Your download has started.");
  expect(requests).toBe(2);
});

test("expired export returns to account sign-in without creating a file", async ({ page, context, baseURL }) => {
  await signIn(context, baseURL!);
  let downloads = 0;
  page.on("download", () => { downloads += 1; });
  await page.route("**/api/v1/personal-data-export", (route) => route.fulfill({
    status: 401, json: { error: { code: "unauthorized", message: "Session expired" } },
  }));
  await page.goto("/settings/account");
  await page.getByRole("button", { name: "Download my data" }).click();
  await expect(page).toHaveURL(/\/login\?returnTo=%2Fsettings%2Faccount&reason=session-expired/);
  await expect(page.getByText("Your session expired. Sign in again to continue.")).toBeVisible();
  expect(downloads).toBe(0);
});

for (const theme of ["light", "dark"] as const) {
  test(`phone password recovery remains operable after validation (${theme})`, async ({ page, context, baseURL }, testInfo) => {
    test.skip(testInfo.project.name === "home-desktop-1280", "Desktop lifecycle is covered by password-auth.spec.ts.");
    await context.addCookies([{ name: "vocanova_theme", value: theme, url: baseURL! }]);
    await page.goto("/auth/password/reset");
    await page.getByLabel("Email address").fill("learner@example.test");
    const send = page.getByRole("button", { name: "Send password reset link" });
    await send.focus();
    await page.keyboard.press("Enter");
    await expect(page.getByRole("status").filter({ hasText: /sent instructions/ })).toBeVisible();
    await page.goto("/auth/password/reset?token=invalid-reset-token");
    await page.getByRole("textbox", { name: "Password" }).fill("short");
    await page.getByRole("button", { name: "Save new password" }).click();
    await expect(page.getByRole("alert").filter({ hasText: "Use a password between 15 and 128 characters." })).toBeVisible();
    await page.getByRole("textbox", { name: "Password" }).fill("a synthetic password long enough");
    await page.getByRole("button", { name: "Save new password" }).click();
    await expect(page.getByRole("alert").filter({ hasText: /invalid, expired, or has already been used/ })).toBeVisible();
    await checkLayout(page);
    await page.screenshot({ path: testInfo.outputPath(`phone-password-${theme}.png`), fullPage: true });
  });

  test(`phone profile survives reload and explains session expiry (${theme})`, async ({ page, context, baseURL }, testInfo) => {
    test.skip(testInfo.project.name === "home-desktop-1280", "Desktop profile persistence and expiry already covered.");
    await signIn(context, baseURL!);
    await context.addCookies([{ name: "vocanova_theme", value: theme, url: baseURL! }]);
    await page.goto("/settings/profile");
    const name = page.getByRole("textbox", { name: "Display name" });
    await name.fill("  Mobile audit learner  ");
    await page.getByRole("button", { name: "Save profile" }).click();
    await expect(page.getByRole("status").filter({ hasText: "Your profile has been saved." })).toBeVisible();
    // Wait for confirmed server presentation before testing explicit reload.
    await expect(page.getByRole("region", { name: "Your profile", exact: true })
      .getByText("MA", { exact: true })).toBeVisible();
    await page.reload();
    await expect(name).toHaveValue("Mobile audit learner");
    await checkLayout(page);
    await name.fill("Session expired draft");
    await context.addCookies([{ name: "e2e_unauthenticated", value: "1", url: baseURL! }]);
    await page.getByRole("button", { name: "Save profile" }).click();
    await expect(page).toHaveURL(/\/login\?returnTo=%2Fsettings%2Fprofile/);
    await expect(page.getByText("Your session expired. Sign in again to continue.")).toBeVisible();
  });
}
