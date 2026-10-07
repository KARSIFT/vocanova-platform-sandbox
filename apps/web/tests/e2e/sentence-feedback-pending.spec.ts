import { randomUUID } from "node:crypto";

import { expect, test, type BrowserContext, type Page } from "@playwright/test";

import { formatViolations, scanForAxeViolations } from "./axe-helper";

async function prepareFeedback(
  page: Page,
  context: BrowserContext,
  baseURL: string,
): Promise<void> {
  await context.addCookies([
    {
      name: "vocanova_session",
      value: `sentence-pending-${randomUUID()}`,
      url: baseURL,
    },
    { name: "vocanova_csrf", value: `csrf-${randomUUID()}`, url: baseURL },
  ]);
  await page.goto("/discover/ordering-at-a-cafe/pour#sentence-practice");
  await page.getByRole("button", { name: /Save pour:/ }).click();
  await page
    .getByRole("textbox", { name: /Write a sentence using pour/ })
    .fill("I pour coffee before work.");
  await page.getByRole("button", { name: "Check my sentence" }).click();
  await expect(
    page.getByRole("status", { name: "Feedback result: Correct" }),
  ).toBeVisible();
}

async function savedIntent(page: Page, sentence: string) {
  return page.evaluate((sentence) => {
    for (const key of Object.keys(sessionStorage)) {
      if (!key.startsWith("vocanova:sentence-feedback-draft:")) continue;
      const draft = JSON.parse(sessionStorage.getItem(key)!);
      if (draft.sentence === sentence)
        return {
          sentence: draft.sentence as string,
          idempotencyKey: draft.idempotencyKey as string | undefined,
        };
    }
    return null;
  }, sentence);
}

for (const action of ["Revise sentence", "Try another sentence"]) {
  test(`${action} cannot discard a pending sentence or its retry identity`, async ({
    page,
    context,
  }, testInfo) => {
    await prepareFeedback(page, context, testInfo.project.use.baseURL!);
    const requests: Array<{ key: string | undefined; body: unknown }> = [];
    let releaseFailure!: () => void;
    const delayedFailure = new Promise<void>((resolve) => {
      releaseFailure = resolve;
    });
    await page.route("**/api/v1/learner-sentences", async (route) => {
      requests.push({
        key: route.request().headers()["idempotency-key"],
        body: route.request().postDataJSON(),
      });
      if (requests.length === 1) {
        await delayedFailure;
        await route.abort("failed");
      } else {
        await route.continue();
      }
    });

    const draft = "I will pour tea for my friends.";
    const input = page.getByRole("textbox", {
      name: /Write a sentence using pour/,
    });
    await input.fill(draft);
    await page.getByRole("button", { name: "Check my sentence" }).click();
    await expect.poll(() => requests.length).toBe(1);
    const oldAction = page.getByRole("button", { name: action, exact: true });
    try {
      await expect
        .soft(
          page.getByRole("button", { name: "Report a problem", exact: true }),
        )
        .toBeDisabled();
      // Native activation respects disabled controls without Playwright's
      // automatic wait for the button to become enabled after the response.
      await oldAction.evaluate((button) =>
        (button as HTMLButtonElement).click(),
      );
      // Keep checking recovery even if the old control was incorrectly enabled.
      await expect.soft(oldAction).toBeDisabled();
      await expect(input).toHaveValue(draft);
      await expect(input).toBeDisabled();
    } finally {
      releaseFailure();
    }

    await expect(
      page
        .getByRole("alert")
        .filter({ hasText: "Unable to check this sentence" }),
    ).toBeVisible();
    await expect(input).toHaveValue(draft);
    await page.reload();
    await expect(input).toHaveValue(draft);
    await page.getByRole("button", { name: "Check my sentence" }).click();
    await expect(
      page.getByRole("status", { name: "Feedback result: Correct" }),
    ).toBeVisible();
    expect(requests).toHaveLength(2);
    expect(requests[0]!.key).toBeTruthy();
    expect(requests[1]).toEqual(requests[0]);
    await expect(
      page.getByText("Sentence checked", { exact: true }).locator(".."),
    ).toContainText(draft);
    await expect(
      page.getByRole("button", { name: action, exact: true }),
    ).toBeEnabled();
  });
}

for (const theme of ["light", "dark"] as const) {
  test(`a failed session recovery keeps earlier feedback and retries its draft in ${theme} mode`, async ({
    page,
    context,
  }, testInfo) => {
    const baseURL = testInfo.project.use.baseURL!;
    await context.addCookies([
      { name: "vocanova_theme", value: theme, url: baseURL },
    ]);
    await prepareFeedback(page, context, baseURL);
    const draft = "I pour tea for my friends.";
    const input = page.getByRole("textbox", {
      name: /Write a sentence using pour/,
    });
    await input.fill(draft);
    await context.clearCookies({ name: "vocanova_csrf" });
    await page.route("**/api/v1/me", (route) => route.abort("failed"), {
      times: 1,
    });
    let submissions = 0;
    page.on("request", (request) => {
      if (
        request.method() === "POST" &&
        request.url().endsWith("/api/v1/learner-sentences")
      ) {
        submissions += 1;
      }
    });
    await page.getByRole("button", { name: "Check my sentence" }).click();
    await expect(
      page
        .getByRole("alert")
        .filter({ hasText: "Unable to check this sentence" }),
    ).toBeVisible();
    await expect(input).toHaveValue(draft);
    await expect(
      page.getByText("Sentence checked", { exact: true }).locator(".."),
    ).toContainText("I pour coffee before work.");
    expect(submissions).toBe(0);
    const pending = await savedIntent(page, draft);
    expect(pending?.idempotencyKey).toBeTruthy();
    await expect(input).toBeEnabled();
    const feedback = page.getByRole("region", { name: "Practice with pour" });
    await feedback.screenshot({
      path: testInfo.outputPath(`sentence-current-error-${theme}.png`),
    });
    const scan = await scanForAxeViolations(page);
    expect(
      scan.criticalOrSerious,
      formatViolations(scan.criticalOrSerious).join("\n"),
    ).toEqual([]);
    const submitted = page.waitForRequest(
      (request) =>
        request.method() === "POST" &&
        request.url().endsWith("/api/v1/learner-sentences"),
    );
    await page.getByRole("button", { name: "Check my sentence" }).click();
    expect((await submitted).headers()["idempotency-key"]).toBe(
      pending!.idempotencyKey,
    );
    await expect(
      page.getByText("Sentence checked", { exact: true }).locator(".."),
    ).toContainText(draft);
    expect(submissions).toBe(1);
  });
}

test("a current session error is not hidden by retained crisis guidance", async ({
  page,
  context,
}, testInfo) => {
  await prepareFeedback(page, context, testInfo.project.use.baseURL!);
  await page.route("**/api/v1/learner-sentences", async (route) => {
    await route.fulfill({
      contentType: "application/json",
      body: JSON.stringify({
        processingStatus: "skipped",
        originalSentence: route.request().postDataJSON().sentenceText,
        missionCompleted: false,
        reported: false,
        canRetry: false,
        errorCode: "SAFETY_SELF_HARM",
        errorMessage: "We are here to help.",
        crisisResourceMessage:
          "If you need immediate support, contact local emergency services.",
      }),
    });
  });
  const input = page.getByRole("textbox", {
    name: /Write a sentence using pour/,
  });
  await input.fill("I pour tea for my friends.");
  await page.getByRole("button", { name: "Check my sentence" }).click();
  const guidance = page
    .getByRole("alert")
    .filter({ hasText: "If you need immediate support" });
  await expect(guidance).toBeVisible();
  await input.fill("I pour water into a glass.");
  await context.clearCookies({ name: "vocanova_csrf" });
  await page.route("**/api/v1/me", (route) => route.abort("failed"), {
    times: 1,
  });
  await page.getByRole("button", { name: "Check my sentence" }).click();
  await expect(
    page
      .getByRole("alert")
      .filter({ hasText: "Unable to check this sentence" }),
  ).toBeVisible();
  await expect(guidance).toBeVisible();
  await expect(input).toHaveValue("I pour water into a glass.");
});

test("a late report for previous feedback cannot mark the new feedback as reported", async ({
  page,
  context,
}, testInfo) => {
  await prepareFeedback(page, context, testInfo.project.use.baseURL!);
  // Observe the browser fetch promise itself: Playwright's response.finished()
  // can remain pending for an intercepted bodyless 204 in Chromium. Returning
  // the original response keeps the real API client and component handlers in
  // the path; two subsequent paints allow their promise continuations to commit.
  await page.evaluate(() => {
    const fetch = window.fetch.bind(window);
    window.fetch = async (...args: Parameters<typeof window.fetch>) => {
      const response = await fetch(...args);
      if (new URL(response.url).pathname.endsWith("/reports")) {
        requestAnimationFrame(() =>
          requestAnimationFrame(() => {
            document.documentElement.dataset.reportResponseProcessed = String(
              response.status,
            );
          }),
        );
      }
      return response;
    };
  });
  let releaseReport!: () => void;
  const delayedReport = new Promise<void>((resolve) => {
    releaseReport = resolve;
  });
  let reportRequests = 0;
  await page.route("**/api/v1/sentence-feedback/*/reports", async (route) => {
    reportRequests += 1;
    await delayedReport;
    await route.fulfill({ status: 204 });
  });
  await page
    .getByRole("button", { name: "Report a problem", exact: true })
    .click();
  await page
    .getByRole("button", { name: "Already correct", exact: true })
    .click();
  await expect.poll(() => reportRequests).toBe(1);
  try {
    const nextSentence = "I pour tea for my friends.";
    await page
      .getByRole("textbox", { name: /Write a sentence using pour/ })
      .fill(nextSentence);
    await page.getByRole("button", { name: "Check my sentence" }).click();
    await expect(
      page.getByText("Sentence checked", { exact: true }).locator(".."),
    ).toContainText(nextSentence);
    releaseReport();
    await expect(page.locator("html")).toHaveAttribute(
      "data-report-response-processed",
      "204",
    );
    await expect(
      page.getByRole("button", { name: "Report a problem", exact: true }),
    ).toBeVisible();
    await expect(page.getByText("Reported", { exact: true })).toHaveCount(0);
  } finally {
    releaseReport();
  }
});

test("a disappearing CSRF cookie recovers before checking and retains one retry identity", async ({
  page,
  context,
}, testInfo) => {
  await prepareFeedback(page, context, testInfo.project.use.baseURL!);
  const draft = "I pour tea for my friends.";
  const input = page.getByRole("textbox", {
    name: /Write a sentence using pour/,
  });
  await input.fill(draft);
  await context.clearCookies({ name: "vocanova_csrf" });
  let recoveries = 0;
  let releaseRecovery!: () => void;
  const recovery = new Promise<void>((resolve) => {
    releaseRecovery = resolve;
  });
  await page.route("**/api/v1/me", async (route) => {
    recoveries += 1;
    await recovery;
    await route.continue();
  });
  const requests: Array<{ key: string | undefined; body: unknown }> = [];
  await page.route("**/api/v1/learner-sentences", async (route) => {
    requests.push({
      key: route.request().headers()["idempotency-key"],
      body: route.request().postDataJSON(),
    });
    if (requests.length === 1) {
      const response = await route.fetch();
      expect(response.status()).toBe(200);
      await route.abort("failed");
    } else {
      await route.continue();
    }
  });
  const form = input.locator("xpath=ancestor::form");
  await form.evaluate((element) => {
    (element as HTMLFormElement).requestSubmit();
    (element as HTMLFormElement).requestSubmit();
  });
  let pending: Awaited<ReturnType<typeof savedIntent>> = null;
  try {
    await expect.poll(() => recoveries).toBe(1);
    expect(requests).toEqual([]);
    await expect(input).toBeDisabled();
    await expect(input).toHaveValue(draft);
    await expect(
      page.getByRole("button", { name: "Checking...", exact: true }),
    ).toBeDisabled();
    await expect(
      page.getByRole("button", { name: "Revise sentence", exact: true }),
    ).toBeDisabled();
    pending = await savedIntent(page, draft);
    expect(pending?.idempotencyKey).toBeTruthy();
  } finally {
    releaseRecovery();
  }
  await expect(
    page
      .getByRole("alert")
      .filter({ hasText: "Unable to check this sentence" }),
  ).toBeVisible();
  await expect(input).toHaveValue(draft);
  expect(requests).toHaveLength(1);
  expect(requests[0]!.key).toBe(pending!.idempotencyKey);
  await page.getByRole("button", { name: "Check my sentence" }).click();
  await expect(
    page.getByText("Sentence checked", { exact: true }).locator(".."),
  ).toContainText(draft);
  expect(recoveries).toBe(1);
  expect(requests).toHaveLength(2);
  expect(requests[1]).toEqual(requests[0]);
});

test("an expired session during CSRF recovery preserves the intent for same-user return", async ({
  page,
  context,
}, testInfo) => {
  await prepareFeedback(page, context, testInfo.project.use.baseURL!);
  const draft = "I pour water into a glass.";
  const input = page.getByRole("textbox", {
    name: /Write a sentence using pour/,
  });
  await input.fill(draft);
  await context.clearCookies({ name: "vocanova_csrf" });
  let submissions = 0;
  page.on("request", (request) => {
    if (
      request.method() === "POST" &&
      request.url().endsWith("/api/v1/learner-sentences")
    )
      submissions += 1;
  });
  await page.route(
    "**/api/v1/me",
    (route) =>
      route.fulfill({
        status: 401,
        contentType: "application/problem+json",
        body: JSON.stringify({ detail: "Authentication required" }),
      }),
    { times: 1 },
  );
  await page.getByRole("button", { name: "Check my sentence" }).click();
  await expect(page).toHaveURL(/\/login\?returnTo=/);
  const loginURL = new URL(page.url());
  expect(loginURL.searchParams.get("reason")).toBe("session-expired");
  expect(loginURL.searchParams.get("returnTo")).toBe(
    "/discover/ordering-at-a-cafe/pour#sentence-practice",
  );
  expect(submissions).toBe(0);
  const pending = await savedIntent(page, draft);
  expect(pending?.idempotencyKey).toBeTruthy();
  await page.goto("/discover/ordering-at-a-cafe/pour#sentence-practice");
  await expect(input).toHaveValue(draft);
  const submitted = page.waitForRequest(
    (request) =>
      request.method() === "POST" &&
      request.url().endsWith("/api/v1/learner-sentences"),
  );
  await page.getByRole("button", { name: "Check my sentence" }).click();
  expect((await submitted).headers()["idempotency-key"]).toBe(
    pending!.idempotencyKey,
  );
  await expect(
    page.getByText("Sentence checked", { exact: true }).locator(".."),
  ).toContainText(draft);
  expect(submissions).toBe(1);
});
