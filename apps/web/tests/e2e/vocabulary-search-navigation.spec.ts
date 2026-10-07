import { randomUUID } from "node:crypto";
import { expect, test } from "@playwright/test";

test("word search fields follow Clear filters and browser history", async ({ page, context, baseURL }) => {
  if (!baseURL) throw new Error("A test app URL is required");
  await context.addCookies([{ name: "vocanova_session", value: randomUUID(), url: baseURL }]);
  await page.goto("/vocabulary?q=pour");
  const query = page.getByRole("searchbox", { name: "Search words and meanings", exact: true });
  await expect(query).toHaveValue("pour");
  await query.fill("invite");
  await page.getByRole("button", { name: "Search", exact: true }).click();
  await expect.poll(() => new URL(page.url()).searchParams.get("q")).toBe("invite");
  await expect(page.getByRole("link", { name: "invite", exact: true })).toBeVisible();
  await page.getByRole("link", { name: "Clear filters", exact: true }).click();
  await expect(page).toHaveURL(/\/vocabulary$/);
  await expect(query).toHaveValue("");
  await page.goBack();
  await expect.poll(() => new URL(page.url()).searchParams.get("q")).toBe("invite");
  await expect(query).toHaveValue("invite");
  await page.goBack();
  await expect.poll(() => new URL(page.url()).searchParams.get("q")).toBe("pour");
  await expect(query).toHaveValue("pour");
});
