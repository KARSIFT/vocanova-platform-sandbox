import { randomUUID } from "node:crypto";
import { expect, test } from "@playwright/test";

for (const view of ["search", "map"] as const) test(`${view} opens the selected later meaning directly`, async ({ page, context, baseURL }) => {
  await context.addCookies([
    { name: "vocanova_session", value: randomUUID(), url: baseURL! },
    { name: "e2e_vocabulary", value: "multiple-meanings", url: baseURL! },
  ]);
  await page.goto(`/vocabulary?q=pour${view === "map" ? "&view=map" : ""}`);
  const result = page.getByRole("main").getByRole("listitem").filter({ hasText: "to arrive in large numbers" });
  await result.getByRole("link").click();
  await expect(page).toHaveURL(/\/vocabulary\/pour(?:\?[^#]*)?#meaning-mean-pour-second$/);
  await expect(page.locator("#meaning-mean-pour-second")).toContainText("to arrive in large numbers");
  await expect(page.locator("#meaning-mean-pour-second")).toBeInViewport();
});

test("word detail returns to the same filtered page of the vocabulary map", async ({
  page,
  context,
  baseURL,
}) => {
  await context.addCookies([
    { name: "vocanova_session", value: randomUUID(), url: baseURL! },
    { name: "e2e_vocabulary", value: "paginated", url: baseURL! },
  ]);
  await page.goto("/vocabulary?category=social&view=map");
  const main = page.getByRole("main");
  const map = main.getByRole("region", { name: "Word knowledge map" });
  await main.getByRole("link", { name: "Next results", exact: true }).click();
  await expect(page).toHaveURL(/(?:\?|&)after=2(?:&|$)/);
  const result = map.getByRole("link").first();
  const resultText = (await result.innerText()).trim();
  await result.press("Enter");

  await main.getByRole("link", { name: /^Back to / }).press("Enter");
  await expect(map).toBeVisible();
  await expect(map.getByRole("link").first()).toHaveText(resultText, {
    useInnerText: true,
  });
  const params = new URL(page.url()).searchParams;
  expect(params.get("category")).toBe("social");
  expect(params.get("view")).toBe("map");
  expect(params.get("after")).toBe("2");
});

test("a situation opens the exact later meaning of a word", async ({
  page,
  context,
  baseURL,
}) => {
  await context.addCookies([
    { name: "vocanova_session", value: randomUUID(), url: baseURL! },
    { name: "vocanova_csrf", value: randomUUID(), url: baseURL! },
    { name: "e2e_vocabulary", value: "secondary-saved-meaning", url: baseURL! },
  ]);
  await page.goto("/discover/ordering-at-a-cafe/pour");
  await page
    .getByRole("button", { name: "Save pour: to make liquid flow into a container" })
    .click();
  await expect(page.getByText("Optional writing", { exact: true })).toBeVisible();

  await page.goto("/discover/ordering-at-a-cafe");
  const words = page.getByRole("list", { name: "Words in this situation" });
  await expect
    .soft(page.getByRole("link", { name: "Continue with pour", exact: true }))
    .toHaveAttribute("href", /#meaning-mean-pour-second$/);
  await words
    .getByRole("link", { name: /pour.*to arrive in large numbers/i })
    .press("Enter");
  await expect(page).toHaveURL(/#meaning-mean-pour-second$/);
  const meaning = page.locator("#meaning-mean-pour-second");
  await expect(meaning).toContainText("to arrive in large numbers");
  await expect(meaning).toBeInViewport();
});

test("a long pasted search is explained as an invalid query, not a lost results page", async ({ page, context, baseURL }) => {
  await context.addCookies([{ name: "vocanova_session", value: randomUUID(), url: baseURL! }]);
  await page.goto("/vocabulary?q=" + "a".repeat(101));
  await expect(page.getByRole("main").getByRole("alert")).toHaveText("Use a search of 100 characters or fewer, without special control characters.");
  await expect(page.getByText("This results page is no longer available.", { exact: true })).toHaveCount(0);
  const search = page.getByRole("searchbox", { name: "Search words and meanings" });
  await expect(search).toHaveAttribute("maxlength", "100");
  await search.fill("pour");
  await page.getByRole("button", { name: "Search", exact: true }).click();
  await expect(page.getByText("1 meaning matching “pour”", { exact: true })).toBeVisible();
});

test("a vocabulary outage offers recovery without claiming empty results", async ({ page, context, baseURL }) => {
  await context.addCookies([
    { name: "vocanova_session", value: randomUUID(), url: baseURL! },
    { name: "e2e_vocabulary", value: "unavailable", url: baseURL! },
  ]);
  await page.goto("/vocabulary?q=pour");
  await expect(page.getByText("We could not load vocabulary right now.", { exact: true })).toBeVisible();
  await expect(page.getByRole("heading", { name: "No matching words yet" })).toHaveCount(0);
  await context.clearCookies({ name: "e2e_vocabulary" });
  await page.getByRole("link", { name: "Try searching again", exact: true }).click();
  await expect(page.getByText("1 meaning matching “pour”", { exact: true })).toBeVisible();
});

test("search, save, detail and knowledge summary form one learner flow", async ({ page, context, baseURL }) => {
  if (!baseURL) throw new Error("Missing app URL");
  await context.addCookies([{ name: "vocanova_session", value: randomUUID(), url: baseURL }, { name: "vocanova_csrf", value: randomUUID(), url: baseURL }]);
  await page.goto("/vocabulary");
  await page.getByRole("searchbox", { name: "Search words and meanings" }).fill("pour");
  await page.getByRole("button", { name: "Search", exact: true }).click();
  await expect(page).toHaveURL(/q=pour/);
  await expect(page.getByText("1 meaning matching “pour”", { exact: true })).toBeVisible();
  await page.getByRole("button", { name: "Save pour: to make liquid flow into a container", exact: true }).click();
  await expect(page.getByRole("button", { name: "Remove pour from saved words", exact: true })).toHaveAttribute("aria-pressed", "true");
  await page.getByRole("link", { name: "pour", exact: true }).click();
  await expect(page.getByRole("heading", { name: "pour", exact: true })).toBeVisible();
  await page.goto("/progress");
  await expect(page.getByRole("main").getByText("1 saved · 1 ready for review", { exact: true })).toBeVisible();
  await expect(page.getByRole("main").getByRole("link", { name: "Start review", exact: true })).toHaveAttribute("href", "/review");
});

test("search filters and empty results survive navigation", async ({ page, context, baseURL }) => {
  if (!baseURL) throw new Error("Missing app URL");
  await context.addCookies([{ name: "vocanova_session", value: randomUUID(), url: baseURL }]);
  await page.goto("/vocabulary?q=pour&category=travel&level=a2");
  await expect(page.getByRole("heading", { name: "No matching words yet" })).toBeVisible();
  await expect(page.getByRole("combobox", { name: "Situation", exact: true })).toHaveValue("travel");
  await expect(page.getByRole("combobox", { name: "Level", exact: true })).toHaveValue("a2");
  await page.reload();
  await expect(page.getByRole("searchbox")).toHaveValue("pour");
  await page.getByRole("link", { name: "Clear filters" }).click();
  await expect(page.getByRole("link", { name: "pour", exact: true })).toBeVisible();
});
