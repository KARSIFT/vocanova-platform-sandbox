import { randomUUID } from "node:crypto";
import { expect, test } from "@playwright/test";

type Destination = {
  label: string;
  href: string;
  heading: string;
  navigation?: string;
  disclosure?: string;
  anchorHeading?: string;
};

// These are source-derived controls whose existing specs checked href/focus,
// rather than reaching the destination. Home Practice/Situations and Your plan
// already activate in navigation-prefetch/auth-check-recovery and are excluded.
const groups: { route: string; heading: string; links: Destination[] }[] = [
  {
    route: "/home", heading: "Today", links: [
      { label: "Short stories", href: "/stories", heading: "Everyday stories", navigation: "Explore your English" },
      { label: "Personal lists", href: "/lists", heading: "Personal lists", navigation: "Explore your English" },
      { label: "Search for a word", href: "/vocabulary", heading: "Word search" },
    ],
  },
  {
    route: "/courses", heading: "Situation courses", links: [
      { label: "Find a word", href: "/vocabulary", heading: "Word search" },
      { label: "Saved words", href: "/words", heading: "Saved vocabulary" },
      { label: "Choose a practice", href: "/practice", heading: "Today’s practice", navigation: "Ways to practise" },
      { label: "Short stories", href: "/stories", heading: "Everyday stories", navigation: "Ways to practise" },
      { label: "Personal lists", href: "/lists", heading: "Personal lists" },
      { label: "Your learning plan", href: "/plan", heading: "Your learning plan" },
    ],
  },
  {
    route: "/plan", heading: "Your learning plan", links: [
      { label: "Change daily preferences", href: "/settings", heading: "Settings" },
      { label: "Set a calendar reminder", href: "/settings#practice-reminder", heading: "Settings", anchorHeading: "Calendar reminder" },
      { label: "View reviews", href: "/review", heading: "You’re ready for new words" },
      { label: "Find your starting words", href: "/vocabulary/check", heading: "Find your starting words" },
      { label: "Open your vocabulary map", href: "/vocabulary?view=map", heading: "Your vocabulary map" },
      { label: "Choose a practice", href: "/practice", heading: "Today’s practice" },
    ],
  },
  {
    route: "/progress", heading: "Progress", links: [
      { label: "Saved words", href: "/words", heading: "Saved vocabulary" },
      { label: "Open your word map", href: "/discover", heading: "Journey" },
      { label: "0 meanings marked already known", href: "/vocabulary?view=map&knowledge=known", heading: "Your vocabulary map", disclosure: "About these stages" },
      { label: "Browse courses", href: "/courses", heading: "Situation courses", disclosure: "Your learning path" },
      { label: "Write your first sentence", href: "/writing", heading: "Topic writing", disclosure: "Sentence practice" },
      { label: "View sentence history", href: "/progress/sentences", heading: "Sentence history", disclosure: "Sentence practice" },
    ],
  },
  {
    route: "/settings", heading: "Settings", links: [
      { label: "Back to Home", href: "/home", heading: "Today" },
      { label: "Account security", href: "/settings/account", heading: "Account" },
      { label: "Profile", href: "/settings/profile", heading: "Profile" },
      { label: "Help and support", href: "/help", heading: "How can we help?" },
    ],
  },
];

for (const theme of ["light", "dark"] as const) {
  for (const group of groups) {
    test(`${group.route} secondary links reach healthy destinations by keyboard (${theme})`, async ({ page, context, baseURL }) => {
      if (!baseURL) throw new Error("A local test app URL is required");
      await context.addCookies([
        { name: "vocanova_session", value: randomUUID(), url: baseURL },
        { name: "vocanova_csrf", value: randomUUID(), url: baseURL },
        { name: "vocanova_theme", value: theme, url: baseURL },
      ]);
      await page.addInitScript(value => localStorage.setItem("vocanova:theme-preference", value), theme);
      for (const destination of group.links) {
        await test.step(`${destination.label} → ${destination.href}`, async () => {
          await page.goto(group.route);
          await expect(page.getByRole("heading", { level: 1, name: group.heading, exact: true })).toBeVisible();
          await expect(page.locator("html")).toHaveAttribute("data-theme", theme);
          const main = page.getByRole("main");
          if (destination.disclosure) {
            const disclosure = main.locator("details").filter({
              has: page.locator("summary").filter({ hasText: destination.disclosure }),
            });
            await expect(disclosure).toHaveCount(1);
            if (await disclosure.getAttribute("open") === null) {
              await disclosure.locator("summary").focus();
              await page.keyboard.press("Enter");
            }
            await expect(disclosure).toHaveAttribute("open", "");
          }
          const scope = destination.navigation
            ? main.getByRole("navigation", { name: destination.navigation, exact: true })
            : main;
          const link = scope.getByRole("link", { name: destination.label, exact: true });
          await expect(link).toHaveAttribute("href", destination.href);
          await link.focus();
          await expect(link).toBeFocused();
          await page.keyboard.press("Enter");
          await expect(page).toHaveURL(new URL(destination.href, baseURL).toString());
          await expect(page.getByRole("heading", { level: 1, name: destination.heading, exact: true })).toBeVisible();
          if (destination.anchorHeading) {
            await expect(page.getByRole("heading", { name: destination.anchorHeading, exact: true })).toBeInViewport();
          }
        });
      }
    });
  }
}
