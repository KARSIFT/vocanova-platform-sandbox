import { randomUUID } from "node:crypto";
import { setTimeout as delay } from "node:timers/promises";

import { expect, test, type BrowserContext, type Page } from "@playwright/test";

import { completeOnboardingIfRedirected } from "./onboarding-journey";

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
  await expect(
    page.getByRole("heading", { name: "Today's Mission", level: 2, exact: true }),
  ).toBeVisible();
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
