import { randomUUID } from "node:crypto";
import { expect, test } from "@playwright/test";

for (const theme of ["light", "dark"] as const) {
  test(`profile and primary navigation remain readable with enlarged text in ${theme}`, async ({
    page,
    context,
    baseURL,
  }) => {
    if (!baseURL) throw new Error("Expected the configured preview URL");
    await context.addCookies([
      { name: "vocanova_session", value: randomUUID(), url: baseURL },
      { name: "vocanova_csrf", value: randomUUID(), url: baseURL },
      { name: "vocanova_theme", value: theme, url: baseURL },
    ]);
    await page.goto("/settings/profile");
    const input = page.getByRole("textbox", { name: "Display name", exact: true });
    await expect(input).toBeEnabled();
    await expect(page.locator("html")).toHaveAttribute("data-theme", theme);
    await page.evaluate(() => {
      document.documentElement.style.fontSize = "200%";
    });
    const helper = page.getByText("Your initials are used as your avatar.", { exact: true });
    await expect(helper).toBeVisible();
    const header = await helper.evaluate((element) => {
      const paragraph = element.getBoundingClientRect();
      const section = element.closest("section")!;
      const form = section.querySelector("form")!.getBoundingClientRect();
      const field = section.querySelector("input")!.getBoundingClientRect();
      const range = document.createRange();
      range.selectNodeContents(element);
      const fragments = Array.from(range.getClientRects()).map((rect) => ({
        left: rect.left,
        right: rect.right,
        bottom: rect.bottom,
      }));
      return {
        paragraphWidth: paragraph.width,
        paragraphBottom: paragraph.bottom,
        formTop: form.top,
        fieldTop: field.top,
        textEscapes: fragments.some((rect) => rect.left < paragraph.left - 1 || rect.right > paragraph.right + 1),
      };
    });
    // An avatar must not squeeze the accompanying explanation into a column
    // too narrow for its ordinary words or let them run into the form.
    expect.soft(header.paragraphWidth).toBeGreaterThanOrEqual(180);
    expect.soft(header.textEscapes).toBe(false);
    expect.soft(header.paragraphBottom).toBeLessThanOrEqual(header.formTop + 1);
    expect.soft(header.paragraphBottom).toBeLessThan(header.fieldTop);

    const nav = page.locator('nav[aria-label="Primary"]:visible');
    if (await nav.count()) {
      const labels = await nav.getByRole("link").evaluateAll((links) => links.map((link) => {
        const bounds = link.getBoundingClientRect();
        const navBounds = link.closest("nav")!.getBoundingClientRect();
        // Labels are direct text nodes beside their decorative icon spans.
        const text = Array.from(link.childNodes).find((node) => node.nodeType === Node.TEXT_NODE && node.textContent?.trim());
        if (!text) throw new Error("Primary destination must have a text label");
        const range = document.createRange();
        range.selectNodeContents(text);
        const rects = Array.from(range.getClientRects()).filter((rect) => rect.width > 0);
        return {
          label: text.textContent?.trim(),
          lines: rects.length,
          escapes: rects.some((rect) => rect.left < bounds.left - 1 || rect.right > bounds.right + 1 || rect.top < bounds.top - 1 || rect.bottom > bounds.bottom + 1 || rect.top < navBounds.top - 1 || rect.bottom > navBounds.bottom + 1),
        };
      }));
      expect(labels.map((label) => label.label)).toEqual(["Home", "Journey", "Progress"]);
      for (const label of labels) {
        expect.soft(label.lines, `${label.label} must remain a readable word`).toBe(1);
        expect.soft(label.escapes, `${label.label} must fit its destination`).toBe(false);
      }
    }
    expect(await page.evaluate(() => document.documentElement.scrollWidth > innerWidth + 1)).toBe(false);
    await input.focus();
    await expect(input).toBeFocused();
  });
}

for (const range of ["", "?band=3"]) {
  test(`Journey map retry reloads the document${range ? " and preserves its range" : ""}`, async ({ page, context, baseURL }) => {
    await context.addCookies([
      { name: "vocanova_session", value: randomUUID(), url: baseURL! },
      { name: "vocanova_csrf", value: randomUUID(), url: baseURL! },
      { name: "e2e_knowledge_map", value: "unavailable", url: baseURL! },
    ]);
    await page.goto(`/discover${range}`);
    await expect(page.getByRole("alert").filter({ hasText: "Your map could not load." })).toBeVisible();
    await context.clearCookies({ name: "e2e_knowledge_map" });
    const retry = page.getByRole("link", { name: "Try again", exact: true });
    await retry.focus();
    await Promise.all([
      page.waitForRequest((request) => request.isNavigationRequest() && new URL(request.url()).pathname === "/discover", { timeout: 5000 }),
      page.keyboard.press("Enter"),
    ]);
    await expect(page.getByRole("heading", { name: range ? "Words 201–300" : "Words 1–100", level: 2, exact: true })).toBeVisible();
    await expect(page.getByRole("alert").filter({ hasText: "Your map could not load." })).toHaveCount(0);
  });
}
