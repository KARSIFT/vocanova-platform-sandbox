import { expect, test } from "@playwright/test";

import { formatViolations, scanForAxeViolations } from "./axe-helper.js";
import { expectNavigationFeedback } from "./navigation-feedback-readiness";

test.beforeEach(async ({ context, baseURL }) => {
  await context.addCookies([
    { name: "vocanova_session", value: crypto.randomUUID(), url: baseURL! },
    { name: "vocanova_csrf", value: "route-loading-csrf", url: baseURL! },
  ]);
});

const DELAY_COOKIE = {
  name: "e2e_response_delay_ms",
  value: "1500",
};

async function delayNextServerRequest(page: import("@playwright/test").Page) {
  await page.context().addCookies([
    {
      ...DELAY_COOKIE,
      url: page.url(),
    },
  ]);
}

async function disableViewportPrefetch(page: import("@playwright/test").Page) {
  // Keep the delayed transition deterministic: desktop Link components eagerly
  // prefetch visible routes, which otherwise may complete before the fixture
  // delay is installed. This only affects this isolated browser context.
  await page.addInitScript(() => {
    window.IntersectionObserver = class {
      disconnect() {}
      observe() {}
      takeRecords() {
        return [];
      }
      unobserve() {}
    } as unknown as typeof IntersectionObserver;
  });
}

test.describe("Route loading states", () => {
  for (const theme of ["light", "dark"] as const) {
    test(`Practice shows loading feedback during navigation (${theme})`, async ({
      page,
    }) => {
      await page.emulateMedia({ colorScheme: theme });
      await disableViewportPrefetch(page);
      await page.goto("/courses");
      await expect(
        page.getByRole("heading", { name: "Situation courses", level: 1 }),
      ).toBeVisible();
      await delayNextServerRequest(page);

      await expectNavigationFeedback(page, {
        loadingStatusId: "practice-loading-status", heading: "Today’s practice",
      }, async () => {
        await page.getByRole("link", { name: /^Choose a practice/ }).focus();
        await page.keyboard.press("Enter");
      });
      expect(
        await page.evaluate(() => document.documentElement.scrollWidth),
      ).toBeLessThanOrEqual(page.viewportSize()!.width);
    });

    for (const resumed of [false, true]) {
      test(`${resumed ? "Resumed" : "New"} lesson shows loading feedback during navigation (${theme})`, async ({
        page,
        context,
        baseURL,
      }) => {
        await context.addCookies([
          {
            name: "vocanova_session",
            value: crypto.randomUUID(),
            url: baseURL!,
          },
        ]);
        await page.emulateMedia({ colorScheme: theme });
        await disableViewportPrefetch(page);
        if (resumed) {
          await page.goto("/learn/conversation-basics");
          await page
            .getByRole("button", { name: "Start lesson", exact: true })
            .click();
          await expect(
            page.getByRole("heading", { name: "invite", exact: true }),
          ).toBeVisible();
          await page
            .getByRole("button", { name: "Continue", exact: true })
            .click();
          await expect(
            page.getByRole("heading", { name: "confirm", exact: true }),
          ).toBeVisible();
        }
        await page.goto("/courses");
        const recommendation = page.getByRole("region", {
          name: resumed ? "Pick up your lesson" : "Your next lesson",
          exact: true,
        });
        const link = recommendation.getByRole("link", {
          name: resumed ? /^Continue lesson/ : /^Start lesson/,
        });
        await expect(link).toHaveAttribute(
          "href",
          "/learn/conversation-basics",
        );
        await delayNextServerRequest(page);
        await expectNavigationFeedback(page, {
          loadingStatusId: "lesson-loading-status",
          heading: resumed ? "confirm" : "Make a plan with a friend",
        }, async () => {
          await link.focus();
          await page.keyboard.press("Enter");
        });
        expect(
          await page.evaluate(() => document.documentElement.scrollWidth),
        ).toBeLessThanOrEqual(page.viewportSize()!.width);
      });
    }
  }

  test("Journey shows loading feedback while its data is delayed", async ({
    page,
  }) => {
    await disableViewportPrefetch(page);
    await page.goto("/home");
    await expect(
      page.getByRole("region", {
        name: "Review words",
        exact: true,
      }),
    ).toBeVisible();

    await delayNextServerRequest(page);

    await expectNavigationFeedback(page, {
      loadingStatusId: "journey-loading-status", heading: "Journey",
    }, () => page.getByRole("navigation", { name: "Primary" })
      .getByRole("link", { name: "Journey" }).click());
  });

  test("Review shows loading feedback while its data is delayed", async ({
    page,
  }, testInfo) => {
    await disableViewportPrefetch(page);
    await page.context().addCookies([
      {
        name: "e2e_review_fixture_count",
        value: "1",
        url: testInfo.project.use.baseURL!,
      },
    ]);
    await page.goto("/home");
    await expect(
      page.getByRole("region", {
        name: "Review words",
        exact: true,
      }),
    ).toBeVisible();

    await delayNextServerRequest(page);

    const action = page.getByRole("button", {
      name: "Start review",
      exact: true,
    });
    await action.click();
    await expect(
      page.getByRole("button", { name: "Starting…", exact: true }),
    ).toBeVisible();
    await expect(
      page.getByRole("button", { name: "Starting…", exact: true }),
    ).toBeDisabled();
    await expect(
      page.getByRole("heading", { name: "Review word 1", exact: true }),
    ).toBeVisible({ timeout: 15_000 });
  });

  test("Saved vocabulary shows loading feedback while its data is delayed", async ({
    page,
  }) => {
    await disableViewportPrefetch(page);
    await page.goto("/courses");
    await expect(
      page.getByRole("heading", { name: "Situation courses", level: 1 }),
    ).toBeVisible();

    await delayNextServerRequest(page);

    await expectNavigationFeedback(page, {
      loadingStatusId: "saved-words-loading-status", heading: "Saved vocabulary",
    }, () => page.getByRole("link", { name: "Saved words" }).click());
  });
});

test("Home keeps Review ready while only the personal practice read is slow", async ({
  page,
  context,
  baseURL,
}) => {
  await context.addCookies([
    { name: "e2e_review_fixture_count", value: "2", url: baseURL! },
    { name: "e2e_practice_sessions_delay_ms", value: "2500", url: baseURL! },
  ]);
  // A streaming response may commit before its secondary section finishes.
  // Waiting for DOMContentLoaded would conceal that independently useful UI.
  await page.goto("/home", { waitUntil: "commit" });
  const main = page.getByRole("main");
  const review = main.getByRole("region", {
    name: "Review words",
    exact: true,
  });
  await expect(
    review.getByRole("button", { name: "Start review", exact: true }),
  ).toBeVisible();
  const path = main.getByRole("region", {
    name: "Your learning path",
    exact: true,
  });
  const pending = path
    .getByRole("status")
    .filter({ hasText: /^Loading practice…$/ });
  await expect(pending).toBeVisible();
  await expect(path).toHaveAttribute("aria-busy", "true");
  await expect(path.getByText(/^\d+ lessons? completed$/)).toHaveCount(0);
  await expect(pending).toHaveCount(0, { timeout: 10000 });
  await expect(
    path.getByRole("button", { name: "Start practice lesson", exact: true }),
  ).toBeVisible();
  await expect(
    path.getByRole("progressbar", {
      name: "Practice lessons today",
      exact: true,
    }),
  ).toHaveAttribute("value", "0");
  await expect(review.getByText(/^Reviewed today:/)).toHaveText(
    "Reviewed today: 0",
  );
});

test("Home keeps Review usable when its secondary practice read fails and can retry it", async ({
  page,
  context,
  baseURL,
}) => {
  await context.addCookies([
    { name: "e2e_review_fixture_count", value: "2", url: baseURL! },
    { name: "e2e_practice_sessions_status", value: "503", url: baseURL! },
  ]);
  await page.goto("/home");
  const main = page.getByRole("main");
  const review = main.getByRole("region", {
    name: "Review words",
    exact: true,
  });
  const path = main.getByRole("region", {
    name: "Your learning path",
    exact: true,
  });
  await expect(
    review.getByRole("button", { name: "Start review", exact: true }),
  ).toBeEnabled();
  await expect(path.getByRole("status")).toHaveText(
    "We could not load your learning path.",
  );
  await expect(path.getByText(/^\d+ lessons? completed$/)).toHaveCount(0);
  await expect(path.getByRole("progressbar")).toHaveCount(0);
  await context.clearCookies({ name: "e2e_practice_sessions_status" });
  await path.getByRole("button", { name: "Try again", exact: true }).click();
  await expect(
    path.getByRole("button", { name: "Start practice lesson", exact: true }),
  ).toBeVisible();
  await expect(
    path.getByRole("progressbar", {
      name: "Practice lessons today",
      exact: true,
    }),
  ).toHaveAttribute("value", "0");
  await review
    .getByRole("button", { name: "Start review", exact: true })
    .click();
  await expect(
    page.getByRole("heading", { name: "Review word 1", exact: true }),
  ).toBeVisible();
});

test("a streamed practice-only authentication failure still redirects Home to sign in", async ({
  page,
  context,
  baseURL,
}) => {
  await context.addCookies([
    { name: "e2e_review_fixture_count", value: "2", url: baseURL! },
    { name: "e2e_practice_sessions_delay_ms", value: "2500", url: baseURL! },
    { name: "e2e_practice_sessions_status", value: "401", url: baseURL! },
  ]);
  await page.goto("/home", { waitUntil: "commit" });
  await expect(
    page
      .getByRole("region", { name: "Review words", exact: true })
      .getByRole("button", { name: "Start review", exact: true }),
  ).toBeVisible();
  await expect(page).toHaveURL(/\/login\?returnTo=%2Fhome&reason=session-expired$/, {
    timeout: 10000,
  });
  await expect(
    page.getByRole("region", { name: "Review words", exact: true }),
  ).toHaveCount(0);
});

test.describe("loading accessibility", () => {
  const routes = [
    { path: "/home", loading: "Loading home", heading: "Today" },
    { path: "/progress", loading: "Loading progress", heading: "Progress" },
    { path: "/settings", loading: "Loading settings", heading: "Settings" },
    // Profile inherits Settings' route boundary; it reproduced the Firefox bug.
    {
      path: "/settings/profile",
      loading: "Loading settings",
      heading: "Profile",
    },
    {
      path: "/settings/account",
      loading: "Loading account settings",
      heading: "Account",
    },
  ];

  for (const theme of ["light", "dark"] as const) {
    for (const route of routes) {
      test(`${route.path} exposes accessible loading feedback (${theme})`, async ({
        page,
        context,
        baseURL,
      }) => {
        await page.emulateMedia({ colorScheme: theme });
        await context.addCookies([
          // Hold the real SSR data reads, not a replacement loading document.
          // A fresh document also avoids the client router's prefetched cache.
          { ...DELAY_COOKIE, value: "6000", url: baseURL! },
        ]);
        await page.goto(route.path, { waitUntil: "commit" });
        await expect(page.locator("html")).toHaveAttribute("data-theme", theme);
        // getByLabel deliberately also matches the original invalid generic
        // div, so the regression fails on axe's actual ARIA violation.
        const pending = page.getByLabel(route.loading, { exact: true });
        await expect(pending).toBeVisible();
        await expect(pending).toHaveAttribute("aria-busy", "true");
        const { criticalOrSerious } = await scanForAxeViolations(page);
        // Prove the scan happened during loading rather than after it settled.
        await expect(pending).toBeVisible();
        await expect(pending).toHaveAttribute("aria-busy", "true");
        expect(
          criticalOrSerious,
          formatViolations(criticalOrSerious).join("\n"),
        ).toEqual([]);
        await expect(
          page.getByRole("region", { name: route.loading, exact: true }),
        ).toBeVisible();
        await expect(pending.getByRole("status")).toHaveText(route.loading);
        await expect(
          page.getByRole("heading", {
            level: 1,
            name: route.heading,
            exact: true,
          }),
        ).toBeVisible({ timeout: 20_000 });
        await expect(pending).toHaveCount(0);
      });
    }
  }
});
