import { randomUUID } from "node:crypto";

import {
  expect,
  test,
  type BrowserContext,
  type Page,
} from "../../apps/web/node_modules/@playwright/test/index.mjs";

import {
  formatViolations,
  scanForAxeViolations,
} from "../../apps/web/tests/e2e/axe-helper";

// Record only reads the application already performs. Reading document.cookie
// at fetch headers would itself change the WebKit cache under investigation.
test.beforeEach(async ({ page }) => {
  await page.addInitScript(() => {
    const events: { at: number; csrfPresent: boolean; visibility: string }[] =
      [];
    (window as unknown as { __cookieReads: typeof events }).__cookieReads =
      events;
    const descriptor = Object.getOwnPropertyDescriptor(
      Document.prototype,
      "cookie",
    )!;
    Object.defineProperty(document, "cookie", {
      configurable: true,
      get() {
        const value = descriptor.get!.call(document) as string;
        events.push({
          at: performance.now(),
          visibility: document.visibilityState,
          csrfPresent: /(?:^|;\s*)vocanova_csrf=/.test(value),
        });
        return value;
      },
      set(value: string) {
        descriptor.set!.call(document, value);
      },
    });
  });
});
test.afterEach(async ({ page, context }, testInfo) => {
  const browser = await page
    .evaluate(() => ({
      events: [
        ...((window as unknown as { __cookieReads?: unknown[] })
          .__cookieReads ?? []),
      ],
      finalCsrfPresent: /(?:^|;\s*)vocanova_csrf=/.test(document.cookie),
      observedAt: performance.now(),
    }))
    .catch(() => null);
  const diagnostic = {
    browser,
    contextCsrfPresent: (await context.cookies()).some(
      (cookie) => cookie.name === "vocanova_csrf",
    ),
    retryEnabled: await page
      .getByRole("button", { name: "Check my sentence", exact: true })
      .isEnabled()
      .catch(() => false),
    errorVisible: await page
      .getByRole("alert")
      .filter({ hasText: "Unable to check this sentence" })
      .isVisible()
      .catch(() => false),
  };
  console.log("COOKIE_READ_DIAGNOSTIC", JSON.stringify(diagnostic));
  await testInfo.attach("cookie-read-diagnostic", {
    body: JSON.stringify(diagnostic),
    contentType: "application/json",
  });
});

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

for (const variant of ["protocol", "native-expiry", "foreground"] as const) {
  for (const theme of ["light", "dark"] as const) {
    test(`a failed session recovery keeps earlier feedback and retries its draft in ${theme} mode (${variant})`, async ({
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
      if (variant === "native-expiry") {
        await page.evaluate(() => {
          document.cookie = "vocanova_csrf=; Max-Age=0; Path=/";
        });
        expect(
          (await context.cookies()).some(
            (cookie) => cookie.name === "vocanova_csrf",
          ),
        ).toBe(false);
      } else {
        await context.clearCookies({ name: "vocanova_csrf" });
      }
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
      if (variant === "foreground") await page.bringToFront();
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
}
