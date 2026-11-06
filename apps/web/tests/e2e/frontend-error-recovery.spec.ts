import { randomUUID } from "node:crypto";
import { expect, test } from "@playwright/test";

const boundaries = [
  { route: "/home", fixture: "home", heading: "Something went wrong", fallback: "Reload page", destination: "/home" },
  { route: "/discover?band=1", fixture: "discover", heading: "Something went wrong", fallback: "Back to Home", destination: "/home" },
  { route: "/reviews", fixture: "reviews", heading: "Something went wrong", fallback: "Back to Home", destination: "/home" },
  { route: "/review", fixture: "reviews", heading: "Something went wrong", fallback: "Back to Home", destination: "/home" },
  { route: "/words", fixture: "words", heading: "Something went wrong", fallback: "Go to Journey", destination: "/discover" },
  { route: "/stories", fixture: "stories", heading: "Your story could not load", fallback: "Back to Stories", destination: "/stories" },
  { route: "/writing", fixture: "writing", heading: "Writing practice could not load", fallback: "Back to Practice", destination: "/practice" },
  { route: "/progress", fixture: "progress", heading: "Something went wrong", fallback: "Reload page", destination: "/progress" },
  { route: "/progress/sentences", fixture: "sentences", heading: "We couldn't load your sentence history", fallback: "Back to Progress", destination: "/progress" },
  { route: "/settings", fixture: "settings", heading: "We couldn't load your settings", fallback: "Back to Home", destination: "/home" },
  { route: "/settings/account", fixture: "account", heading: "We couldn't load your account", fallback: "Back to Settings", destination: "/settings" },
];

const healthyHeadings: Record<string, string | RegExp> = {
  "/home": "Today", "/discover": "Journey", "/words": "Saved vocabulary",
  "/reviews": /Keep your words fresh|You’re ready for new words/,
  "/review": /Keep your words fresh|You’re ready for new words/,
  "/stories": "Everyday stories", "/writing": "Topic writing",
  "/progress": "Progress", "/progress/sentences": "Sentence history",
  "/settings": "Settings", "/settings/account": "Account", "/practice": "Today’s practice",
};

for (const theme of ["light", "dark"] as const) {
  test.describe(`render-boundary recovery (${theme})`, () => {
    test.beforeEach(async ({ page, context, baseURL }) => {
      await page.addInitScript((value) => localStorage.setItem("vocanova:theme-preference", value), theme);
      await context.addCookies([
        { name: "vocanova_session", value: randomUUID(), url: baseURL! },
        { name: "vocanova_csrf", value: randomUUID(), url: baseURL! },
        { name: "vocanova_theme", value: theme, url: baseURL! },
      ]);
    });
    for (const boundary of boundaries) {
      test(`${boundary.route} retries a failed render and its recovery link works`, async ({ page, context, baseURL }) => {
        const fail = () => context.addCookies([{ name: "e2e_render_failure", value: boundary.fixture, url: baseURL! }]);
        const recover = () => context.clearCookies({ name: "e2e_render_failure" });
        await fail();
        await page.goto(boundary.route);
        await expect(page.locator("html")).toHaveAttribute("data-theme", theme);
        const errorHeading = page.getByRole("heading", { name: boundary.heading, exact: true });
        await expect(errorHeading).toBeVisible();
        await recover();
        const retry = page.getByRole("button", { name: "Try again", exact: true });
        await retry.focus();
        await page.keyboard.press("Enter");
        await expect(errorHeading).toHaveCount(0);
        await expect(page.getByRole("main")).toBeVisible();
        await expect(page).toHaveURL(`${baseURL}${boundary.route}`);
        await expect(page.getByRole("heading", { level: 1, name: healthyHeadings[boundary.route.split("?")[0]!]!, exact: true })).toBeVisible();

        await fail();
        await page.reload();
        await expect(errorHeading).toBeVisible();
        await recover();
        const fallback = page.getByRole("main").getByRole("link", { name: boundary.fallback, exact: true });
        await fallback.focus();
        await page.keyboard.press("Enter");
        await expect(page).toHaveURL(`${baseURL}${boundary.destination}`);
        await expect(page.getByRole("heading", { level: 1, name: healthyHeadings[boundary.destination]!, exact: true })).toBeVisible();
        await expect(errorHeading).toHaveCount(0);
        await expect(page.getByRole("main")).toBeVisible();
      });
    }

    test("offline settings save retains the draft and reconnect confirms one persisted edit", async ({ page, context }) => {
      await page.goto("/settings");
        await expect(page.locator("html")).toHaveAttribute("data-theme", theme);
      const displayName = page.getByRole("textbox", { name: "Display name", exact: true });
      await displayName.fill("Offline draft recovery");
      await context.setOffline(true);
      try {
        await page.getByRole("button", { name: "Save settings", exact: true }).click();
        await expect(page.getByRole("alert").filter({ hasText: "We couldn't save your settings. Please try again." })).toBeVisible();
        await expect(displayName).toHaveValue("Offline draft recovery");
        await expect(page.getByText("Your settings have been saved.", { exact: true })).toHaveCount(0);
      } finally {
        await context.setOffline(false);
      }
      const confirmed = page.waitForResponse((response) => response.url().endsWith("/api/v1/settings") && response.request().method() === "PATCH" && response.ok());
      await page.getByRole("button", { name: "Save settings", exact: true }).click();
      await confirmed;
      await expect(page.getByText("Your settings have been saved.", { exact: true })).toBeVisible();
      await page.reload();
      await expect(displayName).toHaveValue("Offline draft recovery");
    });
  });
}
