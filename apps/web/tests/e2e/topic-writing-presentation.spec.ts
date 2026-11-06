import { randomUUID } from "node:crypto";
import { expect, test, type Page } from "@playwright/test";

test.beforeEach(async ({ context, baseURL }) => {
  await context.addCookies([
    { name: "vocanova_session", value: randomUUID(), url: baseURL! },
    { name: "vocanova_csrf", value: randomUUID(), url: baseURL! },
    { name: "e2e_daily_conversation", value: "true", url: baseURL! },
  ]);
});

async function openInvite(page: Page) {
  await page.goto("/writing?situation=daily-conversation");
  await page.getByRole("main").getByRole("link", { name: "invite", exact: true }).click();
  await expect(page.getByRole("button", { name: "Save and write", exact: true })).toBeEnabled();
}

test("saved topic writing shows one meaning introduction and keeps a labelled form", async ({ page }) => {
  await openInvite(page);
  const selected = page.getByRole("region", { name: "Writing with your selected meaning", exact: true });
  const definition = await selected.locator(":scope > p").first().innerText();
  await selected.getByRole("button", { name: "Save and write", exact: true }).click();
  const box = selected.getByRole("textbox", { name: "Write a sentence using invite" });
  await expect(box).toBeVisible();
  await expect(box).toBeFocused();
  await expect(selected.getByRole("heading", { name: "Practice with invite", exact: true })).toHaveCount(0);
  await expect(selected.getByRole("heading", { name: "Write with invite", exact: true })).toHaveCount(1);
  await expect(selected.getByText(definition, { exact: true })).toHaveCount(1);
  await expect(selected.getByRole("region", { name: "Write with invite", exact: true })).toBeVisible();
});

test("returning to a saved topic keeps its confirmed form and draft without another save", async ({ page }) => {
  await openInvite(page);
  let saves = 0;
  page.on("request", (request) => {
    if (request.method() === "POST" && new URL(request.url()).pathname === "/api/v1/user-words") saves++;
  });
  const originalURL = page.url();
  await page.getByRole("button", { name: "Save and write", exact: true }).click();
  const box = page.getByRole("main").getByRole("textbox", { name: "Write a sentence using invite" });
  await expect(box).toBeVisible();
  await box.fill("I invite you to dinner tomorrow.");
  await page.getByRole("main").getByText("Change topic: Daily Conversation", { exact: true }).click();
  await page.getByRole("navigation", { name: "Writing situations" }).getByRole("link", { name: "Ordering at a cafe", exact: true }).click();
  await expect(page).toHaveURL(/situation=ordering-at-a-cafe/);
  await page.goBack();
  await expect(page).toHaveURL(originalURL);
  await expect(box).toBeVisible();
  await expect(box).toHaveValue("I invite you to dinner tomorrow.");
  await expect(page.getByRole("button", { name: "Save and write", exact: true })).toHaveCount(0);
  expect(saves).toBe(1);
});
