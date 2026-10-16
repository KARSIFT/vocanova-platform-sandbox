import { expect, type Page } from "@playwright/test";

export async function openPersonalTools(page: Page) {
  const tools = page.locator("summary").filter({ hasText: "Personal tools" });
  await expect(tools).toBeVisible();
  if ((await tools.locator("..").getAttribute("open")) === null)
    await tools.click();
}

export async function openOptionalWriting(page: Page) {
  await openPersonalTools(page);
  const writing = page
    .locator("summary")
    .filter({ hasText: "Optional writing" });
  await expect(writing).toBeVisible();
  if ((await writing.locator("..").getAttribute("open")) === null)
    await writing.click();
}
