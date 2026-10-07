import { randomUUID } from "node:crypto";
import { expect, test } from "@playwright/test";
import { scanForAxeViolations } from "./axe-helper";

for (const theme of ["light", "dark"] as const) {
  test(`word teaching stays focused and optional tools retain state in ${theme}`, async ({
    page,
    context,
    baseURL,
  }) => {
    await context.addCookies([
      { name: "vocanova_session", value: randomUUID(), url: baseURL! },
      { name: "vocanova_csrf", value: randomUUID(), url: baseURL! },
      { name: "vocanova_theme", value: theme, url: baseURL! },
    ]);
    await page.goto("/vocabulary/pour");
    await page.getByRole("button", { name: /^Save pour:/ }).click();
    await expect(
      page.getByRole("button", { name: "Remove pour from saved words" }),
    ).toBeVisible();
    const writing = page.getByRole("textbox", {
      name: "Write a sentence using pour",
    });
    const practice = page
      .locator("summary")
      .filter({ hasText: "Practise in a sentence" });
    await expect(writing).toBeHidden();
    await expect(
      page.getByRole("button", { name: "Edit knowledge and note" }),
    ).toBeHidden();
    await expect(
      page.getByText("Could you pour me a cup of coffee?", { exact: true }),
    ).toBeVisible();
    const historyLength = await page.evaluate(() => window.history.length);
    await practice.focus();
    await page.keyboard.press("Enter");
    await expect(writing).toBeVisible();
    await expect(page).toHaveURL(/#sentence-practice$/);
    expect(await page.evaluate(() => window.history.length)).toBe(
      historyLength,
    );
    await writing.fill("I pour coffee before work.");
    await practice.click();
    await expect(writing).toBeHidden();
    await expect(page).not.toHaveURL(/#sentence-practice$/);
    expect(await page.evaluate(() => window.history.length)).toBe(
      historyLength,
    );
    await practice.click();
    await expect(writing).toHaveValue("I pour coffee before work.");
    await page.goto("/words/uw-mean-pour#sentence-practice");
    await expect(writing).toBeVisible();
    await expect(writing).toHaveValue("I pour coffee before work.");
    await expect(page.locator("#sentence-practice")).toHaveAttribute(
      "open",
      "",
    );
    await practice.click();
    await expect(writing).toBeHidden();
    await page.evaluate(() => {
      window.location.hash = "";
    });
    await page.evaluate(() => {
      window.location.hash = "sentence-practice";
    });
    await expect(writing).toBeVisible();
    const tools = page.locator("summary").filter({ hasText: "Personal tools" });
    await tools.focus();
    await page.keyboard.press("Space");
    await expect(
      page.getByRole("button", { name: "Edit knowledge and note" }),
    ).toBeVisible();
    expect(
      await page.evaluate(
        () => document.documentElement.scrollWidth <= innerWidth,
      ),
    ).toBe(true);
    const { criticalOrSerious } = await scanForAxeViolations(page);
    expect(criticalOrSerious).toEqual([]);
  });
}

test("manually opened sentence practice restores the visible draft and exact retry after a 401", async ({
  page,
  context,
  baseURL,
}) => {
  await context.addCookies([
    { name: "vocanova_session", value: randomUUID(), url: baseURL! },
    { name: "vocanova_csrf", value: randomUUID(), url: baseURL! },
  ]);
  await page.goto("/vocabulary/pour?from=home");
  await page.getByRole("button", { name: /^Save pour:/ }).click();
  const writing = page.getByRole("textbox", {
    name: "Write a sentence using pour",
  });
  await expect(writing).toBeHidden();
  await page
    .locator("summary")
    .filter({ hasText: "Practise in a sentence" })
    .click();
  await expect(writing).toBeVisible();
  await expect(page).toHaveURL(
    /\/vocabulary\/pour\?from=home#sentence-practice$/,
  );
  const draft = "I pour water into a glass.";
  await writing.fill(draft);
  let firstRequest: { key: string | undefined; body: unknown } | undefined;
  await page.route(
    "**/api/v1/learner-sentences",
    async (route) => {
      firstRequest = {
        key: route.request().headers()["idempotency-key"],
        body: route.request().postDataJSON(),
      };
      await route.fulfill({
        status: 401,
        contentType: "application/problem+json",
        body: JSON.stringify({ detail: "Authentication required" }),
      });
    },
    { times: 1 },
  );
  await page.getByRole("button", { name: "Check my sentence" }).click();
  await expect(page).toHaveURL(/\/login\?returnTo=/);
  const returnTo = new URL(page.url()).searchParams.get("returnTo");
  expect(returnTo).toBe("/vocabulary/pour?from=home#sentence-practice");
  expect(firstRequest?.key).toBeTruthy();

  // The fixture retains this learner's session; emulate the same-user return
  // destination without claiming this exercises a real identity provider.
  await page.goto(returnTo!);
  await expect(writing).toBeVisible();
  await expect(writing).toHaveValue(draft);
  const submitted = page.waitForRequest(
    (request) =>
      request.method() === "POST" &&
      request.url().endsWith("/api/v1/learner-sentences"),
  );
  await page.getByRole("button", { name: "Check my sentence" }).click();
  const retry = await submitted;
  expect(retry.headers()["idempotency-key"]).toBe(firstRequest!.key);
  expect(retry.postDataJSON()).toEqual(firstRequest!.body);
  await expect(
    page.getByText("Sentence checked", { exact: true }).locator(".."),
  ).toContainText(draft);
});
