import { randomUUID } from "node:crypto";
import { expect, test } from "@playwright/test";

for (const theme of ["light", "dark"] as const) {
  test(`progress stages and activity text stay inside their own cards at enlarged text (${theme})`, async ({ page, context, baseURL }) => {
    await context.addCookies([
      { name: "vocanova_session", value: randomUUID(), url: baseURL! },
      { name: "vocanova_csrf", value: randomUUID(), url: baseURL! },
    ]);
    await page.addInitScript((value) => localStorage.setItem("vocanova:theme-preference", value), theme);
    await page.goto("/progress");
    await expect(page.getByRole("heading", { name: "Progress", exact: true })).toBeVisible();
    await expect(page.locator("html")).toHaveAttribute("data-theme", theme);
    await expect(page.getByRole("list", { name: "Review stages" }).locator("li")).toHaveCount(4);
    await expect(page.locator(".activity-days li")).toHaveCount(7);
    for (const size of ["100%", "200%"]) {
      await page.addStyleTag({ content: `html { font-size: ${size} !important; }` });
      await expect.poll(() => page.evaluate(() => {
        const cards = [...document.querySelectorAll<HTMLElement>(".stage-flow li, .activity-days li")];
        return cards.every((card) => {
          const box = card.getBoundingClientRect();
          // DOM element boxes alone miss overflowing text glyphs. Measure the
          // rendered text fragments too, including wrapped date/status lines.
          const nodes = document.createTreeWalker(card, NodeFilter.SHOW_TEXT);
          let node;
          while ((node = nodes.nextNode())) {
            if (!node.textContent?.trim()) continue;
            const range = document.createRange();
            range.selectNodeContents(node);
            if ([...range.getClientRects()].some((rect) => rect.left < box.left - 1 || rect.right > box.right + 1)) return false;
          }
          return true;
        });
      })).toBe(true);
    }
  });
}
