import { randomUUID } from "node:crypto";
import { readFileSync } from "node:fs";
import { expect, test } from "@playwright/test";
import { scanForAxeViolations } from "./axe-helper";

const license = readFileSync(new URL("../../../api/business/dictionary/LICENSE.wordnet", import.meta.url), "utf8");
const normalizeSpace = (text: string) => text.replace(/\s+/g, " ").trim();

test.beforeEach(async ({ context, baseURL }) => {
  await context.addCookies([{ name: "vocanova_session", value: randomUUID(), url: baseURL! }]);
});

for (const theme of ["light", "dark"] as const) {
  test(`global word search gives a readable dictionary result in ${theme}`, async ({ page, context, baseURL }, testInfo) => {
    await context.addCookies([{ name: "vocanova_theme", value: theme, url: baseURL! }]);
    const writes: string[] = [];
    page.on("request", request => { if (["POST", "PUT", "PATCH", "DELETE"].includes(request.method())) writes.push(request.url()); });
    await page.goto("/home");
    const search = page.getByRole("searchbox", { name: "Search any English word", exact: true });
    await expect(search).toBeVisible();
    await search.focus();
    await search.fill("serendipity");
    await page.keyboard.press("Enter");
    await expect(page).toHaveURL(/\/vocabulary\?q=serendipity/);
    const result = page.getByRole("region", { name: "Dictionary result", exact: true });
    await expect(result.getByRole("heading", { name: "serendipity", exact: true })).toBeVisible();
    await expect(result.getByText("Finding something useful or pleasant when you are not looking for it.", { exact: true })).toBeVisible();
    await expect(result.getByText("“By serendipity, she found the book she needed.”", { exact: true })).toBeVisible();
    await expect(page.getByRole("main").getByRole("button", { name: /^Save / })).toHaveCount(0);
    await expect(page.getByRole("heading", { name: "No matching words yet" })).toHaveCount(0);
    await expect(page.getByRole("combobox", { name: "Situation", exact: true })).toBeHidden();
    await page.screenshot({ path: testInfo.outputPath(`dictionary-${theme}.png`), fullPage: true });
    const source = result.locator("summary").filter({ hasText: "Dictionary source" });
    await source.focus();
    await page.keyboard.press("Enter");
    await expect(source).toBeFocused();
    expect(await source.evaluate(element => {
      const box = element.getBoundingClientRect();
      const hit = document.elementFromPoint(box.x + box.width / 2, box.y + box.height / 2);
      return box.height >= 44 && (hit === element || element.contains(hit));
    })).toBe(true);
    await expect(result.getByRole("link", { name: "Princeton WordNet 3.0", exact: true })).toHaveAttribute("href", "https://wordnet.princeton.edu/");
    const notice = result.locator("p").filter({ hasText: "WordNet Release 3.0" });
    await expect(notice).toBeVisible();
    expect(normalizeSpace(await notice.innerText())).toBe(normalizeSpace(license));
    expect(writes).toEqual([]);
    expect(await page.evaluate(() => document.documentElement.scrollWidth <= innerWidth)).toBe(true);
    // Audit the expanded document at a stable position. At an incidental
    // mid-page scroll, any row passing behind sticky chrome is only partly
    // visible; the focused disclosure's actual hit target is checked above.
    await page.evaluate(() => window.scrollTo(0, 0));
    expect((await scanForAxeViolations(page)).criticalOrSerious).toEqual([]);
  });
}

test("an inflected word shows its base form and optional meanings", async ({ page }) => {
  await page.goto("/vocabulary?q=books");
  const result = page.getByRole("region", { name: "Dictionary result", exact: true });
  await expect(page.getByRole("searchbox", { name: "Search words and meanings" })).toHaveValue("books");
  await expect(result.getByRole("heading", { name: "book", exact: true })).toBeVisible();
  await expect(result.getByText("A written work with pages bound together.", { exact: true })).toBeVisible();
  const extra = result.getByText("Arrange to use a room, seat, or service at a future time.", { exact: true });
  await expect(extra).toBeHidden();
  const more = result.locator("summary").filter({ hasText: "More meanings" });
  await more.focus();
  await page.keyboard.press("Space");
  await expect(extra).toBeVisible();
  await expect(result.getByText("A written record of business accounts.", { exact: true })).toBeVisible();
  await more.focus();
  await page.keyboard.press("Enter");
  await expect(extra).toBeHidden();
  await expect(more).toBeFocused();
});

test("an unknown dictionary word offers another search without inventing content", async ({ page }) => {
  await page.goto("/vocabulary?q=zzqxnonword");
  await expect(page.getByRole("region", { name: "Dictionary result" })).toHaveCount(0);
  await expect(page.getByRole("heading", { name: "No matching words yet" })).toBeVisible();
  await expect(page.getByText("Try another spelling or a shorter search.", { exact: true })).toBeVisible();
  await expect(page.getByRole("link", { name: "Try dictionary again" })).toHaveCount(0);
  await expect(page.getByRole("main").getByRole("button", { name: /^Save / })).toHaveCount(0);
});

for (const [fixture, message] of [
  ["unavailable", "We could not load the dictionary right now."],
  ["rate_limited", "You’ve searched several words quickly. Try again in a minute."],
] as const) {
  test(`dictionary ${fixture} keeps the query and offers a real retry`, async ({ page, context, baseURL }) => {
    await context.addCookies([{ name: "e2e_dictionary", value: fixture, url: baseURL! }]);
    await page.goto("/vocabulary?q=serendipity");
    await expect(page.getByText(message, { exact: true })).toBeVisible();
    await expect(page.getByRole("heading", { name: "No matching words yet" })).toHaveCount(0);
    const retry = page.getByRole("link", { name: "Try dictionary again", exact: true });
    await expect(retry).toHaveAttribute("href", "/vocabulary?q=serendipity");
    await context.clearCookies({ name: "e2e_dictionary" });
    await retry.focus();
    await page.keyboard.press("Enter");
    await expect(page.getByRole("region", { name: "Dictionary result" })).toBeVisible();
    await expect(page.getByText(message, { exact: true })).toHaveCount(0);
  });
}

for (const suffix of ["&view=map", "&category=travel", "&level=a2", "&knowledge=saved"]) {
  test(`dictionary does not bypass lesson filters ${suffix}`, async ({ page, context, baseURL }) => {
    await context.addCookies([{ name: "e2e_dictionary", value: "unavailable", url: baseURL! }]);
    await page.goto(`/vocabulary?q=serendipity${suffix}`);
    await expect(page.getByRole("region", { name: "Dictionary result" })).toHaveCount(0);
    await expect(page.getByRole("link", { name: "Try dictionary again" })).toHaveCount(0);
    await expect(page.getByRole("heading", { name: "No matching words yet" })).toBeVisible();
    if (!suffix.includes("view=map")) await expect(page.getByRole("combobox", { name: "Situation", exact: true })).toBeVisible();
  });
}

test("an exact lesson word retains canonical saving without a duplicate dictionary card", async ({ page, context, baseURL }) => {
  await context.addCookies([{ name: "e2e_dictionary", value: "unavailable", url: baseURL! }]);
  await page.goto("/vocabulary?q=pour");
  await expect(page.getByRole("button", { name: /^Save pour:/ })).toBeVisible();
  await expect(page.getByRole("region", { name: "Dictionary result" })).toHaveCount(0);
  await expect(page.getByRole("link", { name: "Try dictionary again" })).toHaveCount(0);
});

test("an expired dictionary request returns to the same search after sign-in", async ({ page, context, baseURL }) => {
  await context.addCookies([{ name: "e2e_dictionary", value: "unauthorized", url: baseURL! }]);
  await page.goto("/vocabulary?q=serendipity");
  await expect(page).toHaveURL(/\/login\?returnTo=/);
  expect(new URL(page.url()).searchParams.get("returnTo")).toBe("/vocabulary?q=serendipity");
});
