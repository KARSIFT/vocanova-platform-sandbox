import { randomUUID } from "node:crypto";
import { expect, test } from "@playwright/test";
import { formatViolations, scanForAxeViolations } from "./axe-helper";

// This matrix checks entry surfaces, not the session controls covered by the
// lesson, story, review and practice journey specs.
const routes: { path: string; heading: string | RegExp }[] = [
  { path: "/home", heading: "Today" },
  { path: "/discover", heading: "Journey" },
  { path: "/discovery", heading: "Discover your next words" },
  { path: "/reviews", heading: /^(Keep your words fresh|You’re ready for new words)$/ },
  { path: "/progress", heading: "Progress" },
  { path: "/progress/sentences", heading: "Sentence history" },
  { path: "/words", heading: "Saved vocabulary" },
  { path: "/vocabulary", heading: "Word search" },
  { path: "/vocabulary/check", heading: "Find your starting words" },
  { path: "/practice", heading: "Today’s practice" },
  { path: "/stories", heading: "Everyday stories" },
  { path: "/courses", heading: "Situation courses" },
  { path: "/lists", heading: "Personal lists" },
  { path: "/plan", heading: "Your learning plan" },
  { path: "/writing", heading: "Topic writing" },
  { path: "/settings", heading: "Settings" },
  { path: "/settings/profile", heading: "Profile" },
  { path: "/settings/account", heading: "Account" },
  { path: "/help", heading: "How can we help?" },
  { path: "/", heading: "Learn the words you will actually use." },
  { path: "/signin", heading: "Sign in to Vocanova" },
  { path: "/signup", heading: "Create your account" },
  { path: "/auth/password/reset", heading: "Reset or add a password" },
  { path: "/auth/password/verify", heading: "Verify your email" },
  { path: "/auth/magic", heading: "Sign in link" },
  { path: "/auth/email-change", heading: "Confirm your new email" },
  { path: "/connection-error", heading: "We couldn't connect" },
  { path: "/account-deactivated", heading: "Your account has been deactivated." },
];

for (const theme of ["light", "dark"] as const) {
  for (const { path: route, heading } of routes) {
    test(`${route} entry reflows with enlarged text and reduced motion (${theme})`, async ({ page, context, baseURL }, testInfo) => {
      await context.addCookies([
        ...(route === "/signin"
          ? []
          : [{ name: "vocanova_session", value: randomUUID(), url: baseURL! }]),
        { name: "vocanova_csrf", value: randomUUID(), url: baseURL! },
      ]);
      await page.addInitScript((value) => localStorage.setItem("vocanova:theme-preference", value), theme);
      await page.emulateMedia({ reducedMotion: "reduce", colorScheme: theme });
      const errors: string[] = [];
      page.on("pageerror", error => errors.push(error.message));
      await page.goto(route);
      // Discovery canonicalizes its default range in the query string.
      await expect.poll(() => new URL(page.url()).pathname).toBe(route);
      const expectedHeading = page.getByRole("heading", { level: 1, name: heading, exact: true });
      await expect(expectedHeading).toBeVisible();
      await expect(page.locator("html")).toHaveAttribute("data-theme", theme);
      const scan = await scanForAxeViolations(page);
      expect(scan.criticalOrSerious, formatViolations(scan.criticalOrSerious).join("\n")).toEqual([]);
      await page.screenshot({ path: testInfo.outputPath("normal.png"), fullPage: true, scale: "css" });
      // Text resizing is distinct from device emulation and browser zoom.
      await page.addStyleTag({ content: "html { font-size: 200% !important; }" });
      await page.screenshot({ path: testInfo.outputPath("text-200-percent.png"), fullPage: true, scale: "css" });
      await expect.poll(() => page.evaluate(() => document.documentElement.scrollWidth))
        .toBeLessThanOrEqual(page.viewportSize()!.width);
      await expect(expectedHeading).toBeVisible();
      if (route === "/discovery") {
        for (const [trigger, title] of [
          ["Word range", "Choose a range"],
          ["Choose a word", "Choose a word"],
        ]) {
          await page.getByRole("button", { name: trigger, exact: true }).click();
          const dialog = page.getByRole("dialog", { name: title, exact: true });
          await expect(dialog).toBeVisible();
          await expect.poll(() => dialog.evaluate(element => element.scrollWidth <= element.clientWidth))
            .toBe(true);
          await dialog.getByRole("button", { name: "Close picker" }).click();
          await expect(dialog).toHaveCount(0);
        }
      }
      expect(errors).toEqual([]);
    });
  }
}
