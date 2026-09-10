import { randomUUID } from "node:crypto";

import { expect, test } from "@playwright/test";

test.describe("Sentence feedback retries", () => {
  test("replays an ambiguous submission exactly once and identifies the checked sentence", async ({
    page,
    context,
  }, testInfo) => {
    const baseURL = testInfo.project.use.baseURL!;
    await context.addCookies([
      { name: "vocanova_session", value: `sentence-retry-${randomUUID()}`, url: baseURL },
      { name: "vocanova_csrf", value: `sentence-retry-csrf-${randomUUID()}`, url: baseURL },
    ]);
    await page.goto("/discover/ordering-at-a-cafe/pour");
    await page.getByRole("button", { name: /Save pour:/ }).click();

    const textarea = page.getByRole("textbox", { name: /Write a sentence using pour/ });
    const submit = page.getByRole("button", { name: "Check my sentence" });
    const requests: Array<{ key: string | null; body: unknown }> = [];
    let abortFirst = true;
    await page.route("**/api/v1/learner-sentences", async (route) => {
      requests.push({
        key: route.request().headers()["idempotency-key"] ?? null,
        body: route.request().postDataJSON(),
      });
      if (abortFirst) {
        abortFirst = false;
        await route.abort("failed");
        return;
      }
      await route.continue();
    });

    const original = "I will pour the coffee carefully.";
    await textarea.fill(original);
    await submit.click();
    await expect(page.getByText(/Unable to check this sentence/)).toBeVisible();
    await submit.click();
    await expect(page.getByText("Sentence checked")).toBeVisible();
    const checkedSentence = page.getByText("Sentence checked", { exact: true }).locator("..");
    await expect(checkedSentence.getByText(original, { exact: true })).toBeVisible();
    expect(requests).toHaveLength(2);
    expect(requests[0]!.key).toBeTruthy();
    expect(requests[1]).toEqual(requests[0]);

    await textarea.fill("I pour tea every morning.");
    await expect(checkedSentence.getByText(original, { exact: true })).toBeVisible();
    await submit.click();
    await expect.poll(() => requests.length).toBe(3);
    expect(requests[2]!.key).not.toBe(requests[0]!.key);
    expect(requests[2]!.body).toMatchObject({ sentenceText: "I pour tea every morning." });
  });

  test("blocks rapid duplicate submits while a request is pending", async ({
    page,
    context,
  }, testInfo) => {
    const baseURL = testInfo.project.use.baseURL!;
    await context.addCookies([
      { name: "vocanova_session", value: `sentence-rapid-${randomUUID()}`, url: baseURL },
      { name: "vocanova_csrf", value: `sentence-rapid-csrf-${randomUUID()}`, url: baseURL },
    ]);
    await page.goto("/discover/ordering-at-a-cafe/pour");
    await page.getByRole("button", { name: /Save pour:/ }).click();
    let requestCount = 0;
    let release: (() => void) | undefined;
    const delayed = new Promise<void>((resolve) => { release = resolve; });
    await page.route("**/api/v1/learner-sentences", async (route) => {
      requestCount += 1;
      await delayed;
      await route.continue();
    });
    await page.getByRole("textbox", { name: /Write a sentence using pour/ }).fill("I will pour tea slowly.");
    const form = page.getByRole("textbox", { name: /Write a sentence using pour/ }).locator("xpath=ancestor::form");
    await form.evaluate((element) => {
      const form = element as HTMLFormElement;
      form.requestSubmit();
      form.requestSubmit();
    });
    await expect.poll(() => requestCount).toBe(1);
    await expect(page.getByRole("button", { name: "Checking..." })).toBeDisabled();
    release?.();
    await expect(page.getByText("Sentence checked")).toBeVisible();
  });
});
