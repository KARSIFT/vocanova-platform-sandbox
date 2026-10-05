import { randomUUID } from "node:crypto";
import { expect, test } from "@playwright/test";
import { scanForAxeViolations, formatViolations } from "./axe-helper";

const apiURL = `http://127.0.0.1:${process.env.MOCK_API_PORT ?? 8080}`;

for (const theme of ["light", "dark"] as const) {
  test(`Home exposes productive writing and Progress leads with learning evidence in ${theme}`, async ({ page, context, baseURL }, testInfo) => {
    const csrf = randomUUID();
    await context.addCookies([
      { name: "vocanova_session", value: randomUUID(), url: baseURL! },
      { name: "vocanova_csrf", value: csrf, url: baseURL! },
      { name: "vocanova_theme", value: theme, url: baseURL! },
      { name: "e2e_daily_conversation", value: "true", url: baseURL! },
    ]);
    const wordResponse = await context.request.get(`${apiURL}/api/v1/canonical-words/invite`);
    expect(wordResponse.ok()).toBe(true);
    const { word } = await wordResponse.json();
    const save = await context.request.post(`${apiURL}/api/v1/user-words`, {
      headers: { "X-CSRF-Token": csrf, "Idempotency-Key": randomUUID() },
      data: { meaningId: word.meanings[0].id, source: "journey" },
    });
    expect(save.ok()).toBe(true);
    const mutations: string[] = [];
    page.on("request", request => { if (request.method() === "POST") mutations.push(request.url()); });
    await page.goto("/home");
    const writing = page.getByRole("region", { name: "Use a word", exact: true });
    const textbox = writing.getByRole("textbox", { name: "Write a sentence using invite" });
    await expect(textbox).toBeVisible();
    await expect(writing.locator("details")).toHaveCount(0);
    await textbox.focus();
    await expect(textbox).toBeFocused();
    await textbox.fill("I invite my friend to dinner.");
    expect(mutations).toEqual([]);
    await expect(page.getByRole("navigation", { name: "Explore your English" }).getByRole("link", { name: /^Read a short story/ })).toHaveAttribute("href", "/stories");
    await page.screenshot({ path: testInfo.outputPath(`home-workspace-${theme}.png`), fullPage: true });
    await page.goto("/progress");
    const map = page.getByRole("region", { name: "Your vocabulary map" });
    const writingHistory = page.getByRole("region", { name: "Sentence practice", exact: true });
    const rewards = page.getByRole("region", { name: "Learning summary" });
    expect((await map.boundingBox())!.y).toBeLessThan((await rewards.boundingBox())!.y);
    expect((await writingHistory.boundingBox())!.y).toBeLessThan((await rewards.boundingBox())!.y);
    await expect(rewards.getByText("Confidence Points", { exact: true })).toBeVisible();
    const scan = await scanForAxeViolations(page);
    expect(scan.criticalOrSerious, formatViolations(scan.criticalOrSerious).join("\n")).toEqual([]);
    expect(await page.evaluate(() => document.documentElement.scrollWidth <= innerWidth)).toBe(true);
    await page.screenshot({ path: testInfo.outputPath(`progress-evidence-${theme}.png`), fullPage: true });
  });

  test(`public preview shows real teaching and clearly authored feedback in ${theme}`, async ({ page, context, baseURL }, testInfo) => {
    await context.addCookies([{ name: "vocanova_theme", value: theme, url: baseURL! }]);
    await page.goto("/");
    const preview = page.getByRole("region", { name: "Learning example" });
    await expect(preview.getByRole("heading", { name: "invite", exact: true })).toBeVisible();
    const image = preview.getByRole("img");
    await expect(image).toBeVisible();
    await expect.poll(() => image.evaluate((element: HTMLImageElement) => element.complete && element.naturalWidth > 0)).toBe(true);
    await page.screenshot({ path: testInfo.outputPath(`public-teaching-${theme}.png`), fullPage: true });
    const button = preview.getByRole("button", { name: "See example feedback", exact: true });
    await button.focus();
    await page.keyboard.press("Enter");
    await expect(preview.getByText("I want invite my friend.", { exact: true })).toBeVisible();
    await expect(preview.getByText("I want to invite my friend.", { exact: true })).toBeVisible();
    await expect(preview.getByText("Example feedback for the sentence shown.")).toBeVisible();
    await expect(preview.getByRole("textbox")).toHaveCount(0);
    await expect(preview.getByTestId("meaning-picture")).toHaveCount(0);
    await preview.getByRole("button", { name: "Back to the word", exact: true }).click();
    await expect(image).toBeVisible();
    const scan = await scanForAxeViolations(page);
    expect(scan.criticalOrSerious, formatViolations(scan.criticalOrSerious).join("\n")).toEqual([]);
    expect(await page.evaluate(() => document.documentElement.scrollWidth <= innerWidth)).toBe(true);
  });
}

for (const fixture of ["paginated", "unbroken", "error", "empty"] as const) {
  test(`Progress shows server writing with truthful ${fixture} state`, async ({ page, context, baseURL }) => {
    await context.addCookies([
      { name: "vocanova_session", value: randomUUID(), url: baseURL! },
      { name: "e2e_sentence_history_fixture", value: fixture, url: baseURL! },
    ]);
    await page.goto("/progress");
    const writing = page.getByRole("region", { name: "Sentence practice", exact: true });
    if (fixture === "paginated") {
      const recent = writing.getByRole("list", { name: "Recent sentences" });
      await expect(recent.getByRole("listitem")).toHaveCount(2);
      await expect(recent.getByText("Fixture sentence 1 uses pour naturally.", { exact: true })).toBeVisible();
      await expect(recent.getByText("Fixture sentence 2 uses pour naturally.", { exact: true })).toBeVisible();
    } else if (fixture === "error") {
      await expect(writing.getByRole("status")).toHaveText("Your recent writing could not load. Open sentence history to try again.");
      await expect(writing.getByRole("link", { name: "Write your first sentence" })).toHaveCount(0);
    } else if (fixture === "empty") {
      await expect(writing.getByRole("link", { name: "Write your first sentence" })).toHaveAttribute("href", "/writing");
    } else {
      await expect(writing.getByRole("list", { name: "Recent sentences" }).getByRole("listitem")).not.toHaveCount(0);
    }
    await expect(writing.getByRole("link", { name: "View sentence history" })).toHaveAttribute("href", "/progress/sentences");
    expect(await page.evaluate(() => document.documentElement.scrollWidth <= innerWidth)).toBe(true);
  });
}
