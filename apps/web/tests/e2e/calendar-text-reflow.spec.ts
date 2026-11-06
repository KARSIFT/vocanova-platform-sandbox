import { randomUUID } from "node:crypto";
import { expect, test } from "@playwright/test";
import { pressTab } from "./keyboard-navigation.js";

for (const theme of ["light", "dark"] as const) {
  test(`calendar date and time remain readable at enlarged text (${theme})`, async ({ page, context, baseURL }, testInfo) => {
    await context.addCookies([
      { name: "vocanova_session", value: randomUUID(), url: baseURL! },
      { name: "vocanova_csrf", value: randomUUID(), url: baseURL! },
    ]);
    await page.addInitScript(value => localStorage.setItem("vocanova:theme-preference", value), theme);
    await page.clock.setFixedTime(new Date("2030-01-15T12:00:00Z"));
    await page.goto("/settings");
    await expect(page.locator("html")).toHaveAttribute("data-theme", theme);
    await page.addStyleTag({ content: "html { font-size: 200% !important; }" });
    const calendar = page.getByRole("region", { name: "Calendar reminder", exact: true });
    const date = calendar.getByLabel("Start date", { exact: true });
    const time = calendar.getByLabel("Reminder time", { exact: true });
    await date.fill("2030-02-01");
    await time.fill("18:45");

    // A viewport breakpoint previously forced two columns inside a narrow
    // desktop settings column. Check room for the rendered value and native
    // picker icon rather than asserting a particular grid implementation.
    for (const [control, sample] of [[date, "02/01/2030"], [time, "06:45 PM"]] as const) {
      const dimensions = await control.evaluate((element, text) => {
        const style = getComputedStyle(element);
        const canvas = document.createElement("canvas");
        const measure = canvas.getContext("2d")!;
        measure.font = `${style.fontWeight} ${style.fontSize} ${style.fontFamily}`;
        return {
          actual: element.getBoundingClientRect().width,
          minimum: measure.measureText(text).width + parseFloat(style.fontSize)
            + parseFloat(style.paddingLeft) + parseFloat(style.paddingRight) + 8,
        };
      }, sample);
      expect(dimensions.actual, `Complete ${sample} and its picker icon need room`).toBeGreaterThanOrEqual(dimensions.minimum);
    }
    expect(await page.evaluate(() => document.documentElement.scrollWidth)).toBeLessThanOrEqual(page.viewportSize()!.width);

    await date.focus();
    // Native date fields have multiple Tab stops for their segments.
    for (let tabs = 0; tabs < 8; tabs += 1) {
      await pressTab(page);
      if (await time.evaluate(element => element === document.activeElement)) break;
    }
    await expect(time).toBeFocused();
    await expect(date).toHaveValue("2030-02-01");
    await expect(time).toHaveValue("18:45");
    expect(await time.evaluate(element => {
      const rect = element.getBoundingClientRect();
      const style = getComputedStyle(element);
      const ring = parseFloat(style.outlineWidth) + Math.max(0, parseFloat(style.outlineOffset));
      // The whole focused control and ring must clear sticky/fixed chrome.
      // A center-only hit check missed the lower third under mobile navigation.
      const visibleChrome = [...document.querySelectorAll("header, nav")].filter(node => {
        const position = getComputedStyle(node).position;
        const box = node.getBoundingClientRect();
        return ["fixed", "sticky"].includes(position) && box.width > 0 && box.height > 0;
      });
      const clearOfChrome = visibleChrome.every(node => {
        const box = node.getBoundingClientRect();
        return rect.right + ring <= box.left || rect.left - ring >= box.right
          || rect.bottom + ring <= box.top || rect.top - ring >= box.bottom;
      });
      const interiorIsClear = [0.25, 0.75].every(x => [0.25, 0.75].every(y =>
        document.elementFromPoint(rect.left + rect.width * x, rect.top + rect.height * y) === element,
      ));
      return interiorIsClear && rect.top - ring >= 0 && rect.bottom + ring <= innerHeight
        && rect.left - ring >= 0 && rect.right + ring <= innerWidth
        && clearOfChrome
        && document.elementFromPoint(rect.x + rect.width / 2, rect.y + rect.height / 2) === element
        && style.outlineStyle !== "none" && parseFloat(style.outlineWidth) >= 2;
    })).toBe(true);
    await page.screenshot({ path: testInfo.outputPath(`calendar-enlarged-${theme}-viewport.png`), scale: "css" });
    await calendar.screenshot({ path: testInfo.outputPath(`calendar-enlarged-${theme}-card.png`), scale: "css" });
  });
}
