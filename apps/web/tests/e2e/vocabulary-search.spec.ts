import { randomUUID } from "node:crypto";
import { expect, test } from "@playwright/test";

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
  await expect(page.getByRole("main").getByText("1 saved meaning, each at its own stage.", { exact: true })).toBeVisible();
  await expect(page.getByRole("main").getByText("1 ready for review now", { exact: true })).toBeVisible();
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
