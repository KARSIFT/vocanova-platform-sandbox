import { expect, test } from "@playwright/test";
import { scanForAxeViolations } from "./axe-helper";

const enabled = process.env.VOCANOVA_COMPONENT_GALLERY === "true";

test("component gallery is unavailable unless explicitly enabled", async ({ page }) => {
  test.skip(enabled, "The explicit local gallery run verifies the enabled variant.");
  const response = await page.goto("/component-gallery");
  expect(response?.status()).toBe(404);
  await expect(page.getByRole("heading", { name: "Vocanova component gallery" })).toHaveCount(0);
});

for (const theme of ["light", "dark"]) {
  test(`shared controls preserve native form behavior and accessibility in ${theme}`, async ({ page }) => {
    test.skip(!enabled, "Gallery interactions require the explicit local preview flag.");
    await page.goto("/component-gallery");
    await page.getByLabel("Appearance", { exact: true }).selectOption(theme);
    await expect(page.locator("html")).toHaveAttribute("data-theme", theme);
    await page.getByRole("button", { name: "Focus sentence", exact: true }).click();
    await expect(page.getByLabel("Your sentence", { exact: true })).toBeFocused();
    await expect(page.getByRole("status", { name: "Preview submissions" })).toHaveText("0");
    await page.getByRole("button", { name: "Check preview sentence", exact: true }).click();
    await expect(page.getByRole("main").getByRole("alert")).toContainText("Write a sentence");
    await expect(page.getByLabel("Your sentence", { exact: true })).toHaveAttribute("aria-invalid", "true");
    await page.getByLabel("Your sentence", { exact: true }).fill("I want to invite my friends to dinner.");
    await page.getByRole("button", { name: "Check preview sentence", exact: true }).click();
    await expect(page.getByRole("status", { name: "Preview submissions" })).toHaveText("1");
    await expect(page.getByRole("main").getByRole("alert")).toHaveCount(0);
    await page.getByRole("button", { name: "Simulate waiting", exact: true }).click();
    await expect(page.getByRole("button", { name: "Check preview sentence", exact: true })).toBeDisabled();
    await page.getByRole("button", { name: "End simulated wait", exact: true }).click();
    await expect(page.getByRole("status", { name: "Preview submissions" })).toHaveText("1");
    await expect(page.getByRole("button", { name: "Check preview sentence", exact: true })).toBeEnabled();
    await page.getByRole("link", { name: "Go to form example" }).focus();
    await page.keyboard.press("Enter");
    await expect(page).toHaveURL(/#gallery-form$/);
    await page.evaluate(() => { document.documentElement.style.fontSize = "200%"; });
    const overflow = await page.evaluate(() => document.documentElement.scrollWidth > window.innerWidth + 1);
    expect(overflow).toBe(false);
    expect((await scanForAxeViolations(page)).criticalOrSerious).toEqual([]);
  });
}
