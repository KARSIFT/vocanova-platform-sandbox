import { randomUUID } from "node:crypto";
import { expect, test, type BrowserContext } from "@playwright/test";
import type { WordListDetail } from "@vocanova/api-client";
import { formatViolations, scanForAxeViolations } from "./axe-helper";

const apiURL = `http://127.0.0.1:${process.env.MOCK_API_PORT ?? 8080}`;

async function seedList(context: BrowserContext) {
  const csrf = (await context.cookies()).find((cookie) => cookie.name === "vocanova_csrf")!.value;
  const headers = () => ({ "X-CSRF-Token": csrf, "Idempotency-Key": randomUUID() });
  const id = randomUUID();
  const created = await context.request.put(`${apiURL}/api/v1/word-lists/${id}`, {
    headers: headers(), data: { name: "Dynamic control audit", expectedRevision: 0 },
  });
  expect(created.status()).toBe(200);
  const list = await created.json() as WordListDetail;
  const word = await (await context.request.get(`${apiURL}/api/v1/canonical-words/invite`)).json();
  const added = await context.request.put(`${apiURL}/api/v1/word-lists/${id}/members/${word.word.meanings[0].id}`, {
    headers: headers(), data: { expectedRevision: list.revision },
  });
  expect(added.status()).toBe(200);
  return await added.json() as WordListDetail;
}

test.beforeEach(async ({ context, baseURL }) => {
    await context.addCookies([
      { name: "vocanova_session", value: randomUUID(), url: baseURL! },
      { name: "vocanova_csrf", value: randomUUID(), url: baseURL! },
    ]);
});

for (const theme of ["light", "dark"] as const) {
  test(`list detail cancel, validation, destinations and confirmed removal (${theme})`, async ({ page, context, baseURL }) => {
    await page.addInitScript((preference) => localStorage.setItem("vocanova:theme-preference", preference), theme);
    await context.addCookies([{ name: "vocanova_theme", value: theme, url: baseURL! }]);
    const list = await seedList(context);
    const path = `/lists/${list.id}`;
    await page.goto(path);
    await expect(page.locator("html")).toHaveAttribute("data-theme", theme);
    const main = page.getByRole("main");
    await main.locator("summary").filter({ hasText: /^List settings$/ }).press("Enter");
    const rename = main.getByRole("button", { name: "Rename list", exact: true });
    const name = main.getByLabel("List name", { exact: true });
    await expect(rename).toBeDisabled();
    for (const invalid of ["   ", "a".repeat(81)]) {
      await name.fill(invalid);
      await expect(rename).toBeDisabled();
    }
    await name.fill(list.name);
    let deletes = 0;
    page.on("request", (request) => {
      if (request.method() === "DELETE" && new URL(request.url()).pathname === `/api/v1/word-lists/${list.id}`) deletes += 1;
    });
    await main.getByRole("button", { name: "Delete list", exact: true }).press("Enter");
    await expect(main.getByRole("button", { name: "Confirm delete list", exact: true })).toBeVisible();
    await main.getByRole("button", { name: "Keep list", exact: true }).press("Space");
    await expect(main.getByRole("button", { name: "Confirm delete list", exact: true })).toHaveCount(0);
    await expect(main.getByRole("button", { name: "Delete list", exact: true })).toBeVisible();
    expect(deletes).toBe(0);
    await page.reload();
    await expect(main.getByRole("heading", { name: list.name, exact: true })).toBeVisible();
    await main.getByRole("link", { name: "Practice this list", exact: true }).press("Enter");
    await expect(page).toHaveURL(`${baseURL}/practice?list=${list.id}`);
    await expect(page.getByLabel("Practice vocabulary", { exact: true })).toHaveValue(`list:${list.id}`);
    await page.goto(path);
    await expect(page.locator("html")).toHaveAttribute("data-theme", theme);
    await main.getByRole("link", { name: "Find words to add", exact: true }).press("Enter");
    await expect(page).toHaveURL(`${baseURL}/vocabulary`);
    await expect(main.getByRole("heading", { level: 1, name: "Word search", exact: true })).toBeVisible();
    await page.goto(path);
    await expect(page.locator("html")).toHaveAttribute("data-theme", theme);
    await main.getByRole("link", { name: "invite", exact: true }).press("Enter");
    await expect(page).toHaveURL(`${baseURL}/vocabulary/invite#meaning-${list.members[0]!.meaningId}`);
    await expect(page.locator(`#meaning-${list.members[0]!.meaningId}`)).toBeInViewport();
    await expect(main.getByRole("heading", { level: 1, name: "invite", exact: true })).toBeVisible();
    await page.goto(path);
    await expect(page.locator("html")).toHaveAttribute("data-theme", theme);
    const search = main.getByLabel("Search words and meanings in this list");
    await search.fill("no-such-meaning");
    await expect(main.getByRole("heading", { name: "No meanings match your search" })).toBeVisible();
    await search.fill("");
    await expect(main.getByRole("list", { name: "List meanings", exact: true }).getByRole("listitem")).toHaveCount(1);
    await main.getByRole("button", { name: "Remove invite from this list", exact: true }).press("Enter");
    await expect(main.getByRole("status")).toContainText("Current list membership confirmed.");
    await expect(main.getByRole("heading", { name: "Your list is empty" })).toBeVisible();
    await expect(main.getByRole("link", { name: "Practice this list", exact: true })).toHaveCount(0);
    await page.reload();
    await expect(main.getByRole("heading", { name: "Your list is empty" })).toBeVisible();
    const persisted = await context.request.get(`${apiURL}/api/v1/word-lists/${list.id}`);
    expect(await persisted.json()).toMatchObject({ memberCount: 0, members: [] });
    const scan = await scanForAxeViolations(page);
    expect(scan.criticalOrSerious, formatViolations(scan.criticalOrSerious).join("\n")).toEqual([]);
    expect(await page.evaluate(() => document.documentElement.scrollWidth)).toBeLessThanOrEqual(page.viewportSize()!.width);
  });

  test(`unavailable list detail retries its own data and missing list offers recovery (${theme})`, async ({ page, context, baseURL }) => {
    await page.addInitScript((preference) => localStorage.setItem("vocanova:theme-preference", preference), theme);
    const list = await seedList(context);
    await context.addCookies([
      { name: "vocanova_theme", value: theme, url: baseURL! },
      { name: "e2e_lists", value: "unavailable", url: baseURL! },
    ]);
    await page.goto(`/lists/${list.id}`);
    await expect(page.locator("html")).toHaveAttribute("data-theme", theme);
    await expect(page.getByRole("heading", { name: "Your list could not be loaded" })).toBeVisible();
    await context.clearCookies({ name: "e2e_lists" });
    await page.getByRole("button", { name: "Try loading list again", exact: true }).press("Enter");
    await expect(page.getByRole("heading", { name: list.name, exact: true })).toBeVisible();
    await expect(page.getByRole("list", { name: "List meanings", exact: true }).getByRole("listitem")).toHaveCount(1);
    await page.goto(`/lists/${randomUUID()}`);
    await expect(page.getByRole("heading", { name: "This list is no longer available" })).toBeVisible();
    await page.getByRole("link", { name: "Choose a personal list", exact: true }).press("Enter");
    await expect(page).toHaveURL(`${baseURL}/lists`);
    await expect(page.getByRole("link", { name: list.name, exact: true })).toBeVisible();
  });
}
