import { randomUUID } from "node:crypto";
import { setTimeout as delay } from "node:timers/promises";

import { expect, test, type BrowserContext, type Page } from "@playwright/test";

import { completeOnboardingIfRedirected } from "./onboarding-journey";
import { scanForAxeViolations } from "./axe-helper";

async function startOnboarding(
  page: Page,
  context: BrowserContext,
  baseURL: string | undefined,
) {
  if (!baseURL) throw new Error("Expected a configured base URL.");
  await context.addCookies([
    {
      name: "vocanova_session",
      value: `onboarding-${randomUUID()}`,
      url: baseURL,
    },
    { name: "vocanova_csrf", value: `csrf-${randomUUID()}`, url: baseURL },
    { name: "e2e_onboarding_status", value: "not_started", url: baseURL },
  ]);
  await page.goto("/home");
  await expect(page).toHaveURL(/\/onboarding(\?|$)/);
}

test("staging onboarding helper waits for delayed completion and lets the app reach home", async ({
  page,
  context,
  baseURL,
}) => {
  await startOnboarding(page, context, baseURL);
  let completionPending = false;
  let prematureHomeNavigation = false;
  let onboardingResponse: unknown;
  let submissionCount = 0;
  page.on("request", (request) => {
    if (
      completionPending &&
      request.isNavigationRequest() &&
      request.frame() === page.mainFrame() &&
      new URL(request.url()).pathname === "/home"
    ) {
      prematureHomeNavigation = true;
    }
  });
  await page.route("**/api/v1/onboarding", async (route) => {
    if (route.request().method() !== "POST") {
      await route.continue();
      return;
    }
    submissionCount++;
    completionPending = true;
    // Simulate a slow mutation before forwarding it to the existing backend.
    // This delay is transport behavior, not synchronization of the UI test.
    await delay(500);
    const response = await route.fetch();
    expect(response.status()).toBe(200);
    onboardingResponse = await response.json();
    // The mock's cookie override must end only after a successful completion.
    // No browser navigation is performed here: the app handles its own result.
    await context.clearCookies({ name: "e2e_onboarding_status" });
    completionPending = false;
    await route.fulfill({ response });
  });

  const helperError = await completeOnboardingIfRedirected(page).then(
    () => undefined,
    (error: unknown) => error,
  );
  expect(
    prematureHomeNavigation,
    "the helper must not leave while onboarding completion is pending",
  ).toBe(false);
  expect(helperError).toBeUndefined();
  expect(submissionCount).toBe(1);
  expect(onboardingResponse).toMatchObject({
    status: "completed",
    englishLevel: "a2",
    nativeLanguage: "es",
    learningGoal: "general",
    mainUseCase: "daily_life",
    dailyReviewTarget: 15,
  });
  await expect(page).toHaveURL(/\/home(\?|$)/);
  const mission = page.getByRole("region", {
    name: "Review words",
    exact: true,
  });
  await expect(mission).toBeVisible();
  await expect(
    mission.getByRole("link", { name: "Discover words", exact: true }),
  ).toHaveAttribute("href", "/discovery");
  await expect(
    page
      .getByRole("region", { name: "Review words", exact: true })
      .getByText(/^Reviewed today:/),
  ).toHaveText("Reviewed today: 0");
});

test("staging onboarding helper rejects a failed save and preserves the retry form", async ({
  page,
  context,
  baseURL,
}) => {
  await startOnboarding(page, context, baseURL);
  await page.route("**/api/v1/onboarding", async (route) => {
    if (route.request().method() !== "POST") {
      await route.continue();
      return;
    }
    await delay(500);
    await route.fulfill({
      status: 503,
      contentType: "application/problem+json",
      body: JSON.stringify({
        detail: "We couldn't save your answers. Please try again.",
      }),
    });
  });

  const helperError = await completeOnboardingIfRedirected(page).then(
    () => undefined,
    (error: unknown) => error,
  );
  expect(
    helperError,
    "a failed completion must fail the journey",
  ).toBeInstanceOf(Error);
  await expect(page).toHaveURL(/\/onboarding(\?|$)/);
  await expect(page.getByRole("main").getByRole("alert")).toHaveText(
    "We couldn't save your answers. Please try again.",
  );
  await expect(
    page.getByRole("heading", { name: "Daily review target" }),
  ).toBeVisible();
  await expect(
    page.getByRole("button", { name: "Finish setup" }),
  ).toBeEnabled();
});

for (const language of [
  { query: "فارسی", code: "fa", label: "Persian", suggestion: true },
  { query: "日本語", code: "ja", label: "Japanese", suggestion: true },
  { query: "Tigrinya", code: "Tigrinya", label: "Tigrinya", suggestion: false },
]) {
  for (const theme of ["light", "dark"] as const) {
  test(`native language search saves ${language.label} and survives Back in ${theme}`, async ({
    page,
    context,
    baseURL,
  }) => {
    if (!baseURL) throw new Error("Expected a configured base URL.");
    await context.addCookies([{ name: "vocanova_theme", value: theme, url: baseURL }]);
    await startOnboarding(page, context, baseURL);
    await page.getByRole("radio", { name: /A2 — Elementary/ }).check();
    await page.getByRole("button", { name: "Continue", exact: true }).click();
    const field = page.getByRole("textbox", { name: "Native language" });
    await expect(field).toBeFocused();
    await expect(page.getByRole("button", { name: /^Persian/ })).toBeVisible();
    await field.fill(language.query);
    if (language.suggestion) {
      const choice = page.getByRole("button", {
        name: new RegExp(`^${language.label}`),
      });
      await field.press("Tab");
      await expect(choice).toBeFocused();
      await choice.press("Enter");
      await expect(choice).toHaveAttribute("aria-pressed", "true");
      await expect(field).toHaveValue(language.label);
    }
    const { criticalOrSerious } = await scanForAxeViolations(page);
    expect(criticalOrSerious).toEqual([]);
    const overflow = await page.evaluate(
      () => document.documentElement.scrollWidth > window.innerWidth,
    );
    expect(overflow).toBe(false);
    await page.screenshot({ path: test.info().outputPath(`native-language-${theme}.png`) });
    await page.getByRole("button", { name: "Continue", exact: true }).click();
    await page.getByRole("button", { name: "Back", exact: true }).click();
    await expect(field).toHaveValue(language.label);
    await page.getByRole("button", { name: "Continue", exact: true }).click();
    await page.getByRole("radio", { name: "General growth", exact: true }).check();
    await page.getByRole("button", { name: "Continue", exact: true }).click();
    await page.getByRole("radio", { name: "Daily life", exact: true }).check();
    await page.getByRole("button", { name: "Continue", exact: true }).click();
    const summary = page.getByRole("complementary", { name: "Onboarding summary" });
    await expect(summary).toContainText(language.label);
    const responsePromise = page.waitForResponse(
      (response) =>
        response.request().method() === "POST" &&
        new URL(response.url()).pathname === "/api/v1/onboarding",
    );
    await page.getByRole("button", { name: "Finish setup", exact: true }).click();
    const response = await responsePromise;
    expect(response.status()).toBe(200);
    expect(response.request().postDataJSON()).toMatchObject({
      nativeLanguage: language.code,
    });
    expect(await response.json()).toMatchObject({ nativeLanguage: language.code });
  });
  }
}
