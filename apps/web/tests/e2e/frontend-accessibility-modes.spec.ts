import { randomUUID } from "node:crypto";
import AxeBuilder from "@axe-core/playwright";
import { expect, test, type Page } from "@playwright/test";
import { formatViolations, scanForAxeViolations } from "./axe-helper";
import { pressTab } from "./keyboard-navigation.js";

async function reflow(page: Page) {
  expect(await page.evaluate(() => document.documentElement.scrollWidth)).toBeLessThanOrEqual(page.viewportSize()!.width);
  const scan = await scanForAxeViolations(page);
  expect(scan.criticalOrSerious, formatViolations(scan.criticalOrSerious).join("\n")).toEqual([]);
}

async function keyboardFocusVisible(page: Page) {
  // Actual Tab navigation must scroll the focused element clear of sticky UI.
  for (let i = 0; i < 6; i++) {
    await pressTab(page);
    const result = await page.evaluate(() => {
      const element = document.activeElement as HTMLElement;
      const rect = element.getBoundingClientRect();
      // Composite controls may draw their focus indicator on a focus-within
      // wrapper (the header search does), rather than on the input itself.
      const ownStyle = getComputedStyle(element);
      // Vocanova's focus-visible contract is an outline. A permanent button
      // shadow is decoration and must not make a missing focus ring pass.
      const ownRing = element.matches(":focus-visible") && ownStyle.outlineStyle !== "none" && parseFloat(ownStyle.outlineWidth) >= 2;
      const search = element.closest(".header-search");
      const searchStyle = search?.matches(":focus-within") ? getComputedStyle(search) : null;
      const compositeRing = Boolean(searchStyle && searchStyle.outlineStyle !== "none" && parseFloat(searchStyle.outlineWidth) >= 2);
      const indicator = ownRing || compositeRing;
      const x = rect.x + rect.width / 2;
      const y = rect.y + rect.height / 2;
      const top = document.elementFromPoint(x, y);
      return {
        tag: element.tagName,
        visible: rect.width > 0 && rect.height > 0 && x >= 0 && x <= innerWidth && y >= 0 && y <= innerHeight,
        unobscured: !!top && (element.contains(top) || top.contains(element)),
        indicator,
      };
    });
    expect(result.tag).not.toBe("BODY");
    expect(result.visible).toBe(true);
    expect(result.unobscured).toBe(true);
    expect(result.indicator).toBe(true);
  }
}

for (const theme of ["light", "dark"] as const) {
  test.describe(`alternative access modes (${theme})`, () => {
    test.beforeEach(async ({ page, context, baseURL }) => {
      await page.addInitScript((preference) => localStorage.setItem("vocanova:theme-preference", preference), theme);
      await context.addCookies([
        { name: "vocanova_session", value: randomUUID(), url: baseURL! },
        { name: "vocanova_csrf", value: randomUUID(), url: baseURL! },
        { name: "vocanova_theme", value: theme, url: baseURL! },
      ]);
    });

    test("short landscape preserves navigation, focused controls and profile save", async ({ page }, info) => {
      await page.setViewportSize({ width: 667, height: 375 });
      await page.goto("/settings/profile");
      await expect(page.locator("html")).toHaveAttribute("data-theme", theme);
      const name = page.getByRole("textbox", { name: "Display name" });
      await expect(name).toBeVisible();
      await keyboardFocusVisible(page);
      await name.fill("Landscape learner");
      await page.getByRole("button", { name: "Save profile" }).press("Enter");
      await expect(page.getByRole("status")).toContainText("Your profile has been saved.");
      // Audit the refreshed server presentation and its streamed metadata.
      await expect(page.getByRole("region", { name: "Your profile", exact: true })
        .getByText("LL", { exact: true })).toBeVisible();
      await expect(page).toHaveTitle("Profile — Vocanova");
      await reflow(page);
      await page.screenshot({ path: info.outputPath(`landscape-profile-${theme}.png`), fullPage: true });
      await page.getByRole("navigation", { name: "Primary" }).getByRole("link", { name: "Journey", exact: true }).press("Enter");
      await expect(page).toHaveURL(/\/discover$/);
      await reflow(page);
    });

    test("forced colors and reduced motion preserve review choices and feedback", async ({ page, context, baseURL }, info) => {
      await page.emulateMedia({ forcedColors: "active", reducedMotion: "reduce" });
      await context.addCookies([{ name: "e2e_review_fixture_count", value: "1", url: baseURL! }]);
      await page.goto("/reviews");
      await expect(page.locator("html")).toHaveAttribute("data-theme", theme);
      expect(await page.evaluate(() => matchMedia("(forced-colors: active)").matches)).toBe(true);
      expect(await page.evaluate(() => matchMedia("(prefers-reduced-motion: reduce)").matches)).toBe(true);
      await page.getByRole("button", { name: "Start review", exact: true }).press("Enter");
      const choice = page.getByRole("radio", { name: "definition for review word 1", exact: true });
      await choice.focus();
      await page.keyboard.press("Space");
      await expect(choice).toBeChecked();
      await page.getByRole("button", { name: "Check answer", exact: true }).press("Enter");
      await expect(page.getByRole("status").filter({ hasText: /^Correct/ })).toBeVisible();
      expect(await page.evaluate(() => document.documentElement.scrollWidth)).toBeLessThanOrEqual(page.viewportSize()!.width);
      // Keep all other WCAG checks in system-color mode. Only color-contrast
      // cannot be interpreted here: Chromium's painted system colors differ
      // from the author values returned to axe (see audit evidence below).
      const forcedScan = await new AxeBuilder({ page })
        .withTags(["wcag2a", "wcag2aa", "wcag21a", "wcag21aa", "wcag22aa"])
        .disableRules(["color-contrast"])
        .analyze();
      expect(forcedScan.violations.filter((violation) => ["critical", "serious"].includes(violation.impact ?? ""))).toEqual([]);
      await page.screenshot({ path: info.outputPath(`forced-colors-review-${theme}-viewport.png`) });
      await page.screenshot({ path: info.outputPath(`forced-colors-review-${theme}.png`), fullPage: true });
      // In forced-colors Chromium paints system colors, while axe reads the
      // author's dark foreground tokens against the system Canvas background.
      // Inspect the forced-color artifact separately; run the full unchanged
      // axe rules on the same feedback state with the authored theme restored.
      await page.emulateMedia({ forcedColors: "none" });
      await expect(page.locator("html")).toHaveAttribute("data-theme", theme);
      await reflow(page);
      await page.getByRole("button", { name: "Continue", exact: true }).press("Enter");
      await expect(page.getByRole("heading", { name: /^(?:Review saved|Review set complete|You’ve finished your review)$/ })).toBeFocused();
    });

    test("composed Unicode profile text stays local until explicit save", async ({ page }, info) => {
      await page.goto("/settings/profile");
      await expect(page.locator("html")).toHaveAttribute("data-theme", theme);
      const field = page.getByRole("textbox", { name: "Display name" });
      let writes = 0;
      page.on("request", (request) => {
        if (request.method() === "PATCH" && new URL(request.url()).pathname === "/api/v1/settings") writes += 1;
      });
      await field.fill("");
      await field.dispatchEvent("compositionstart", { data: "" });
      // Synthetic DOM composition checks React's event handling; this does not
      // represent physical OS IME candidate windows or assistive speech output.
      await field.dispatchEvent("compositionupdate", { data: "学習" });
      await page.keyboard.insertText("学習");
      await field.dispatchEvent("compositionend", { data: "学習" });
      await page.keyboard.insertText(" learner e\u0301 🌱");
      await expect(field).toHaveValue("学習 learner e\u0301 🌱");
      expect(writes).toBe(0);
      await page.getByRole("button", { name: "Save profile" }).press("Enter");
      await expect(page.getByRole("status")).toContainText("Your profile has been saved.");
      expect(writes).toBe(1);
      // The saved status precedes router.refresh; the server-rendered avatar
      // confirms that refresh has committed before a separate reload begins.
      await expect(page.getByRole("region", { name: "Your profile", exact: true })
        .getByText("学L", { exact: true })).toBeVisible();
      await page.reload();
      await expect(field).toHaveValue("学習 learner e\u0301 🌱");
      await reflow(page);
      await page.screenshot({ path: info.outputPath(`unicode-profile-${theme}.png`), fullPage: true });
    });
  });
}
