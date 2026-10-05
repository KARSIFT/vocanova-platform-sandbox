import { expect, test } from "@playwright/test";

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
      await page.goto("/discover");
      await expect(
        page.getByRole("heading", { name: "Journey", level: 1 }),
      ).toBeVisible();
      await delayNextServerRequest(page);

      await page.getByRole("link", { name: /^Choose a practice/ }).focus();
      await page.keyboard.press("Enter");
      const status = page
        .getByRole("status")
        .filter({ hasText: "Loading practice" });
      await expect(status).toBeVisible();
      await expect(status.locator("..")).toHaveAttribute("aria-busy", "true");
      await expect(
        page.getByRole("heading", { name: "Practice your way", level: 1 }),
      ).toBeVisible({ timeout: 10_000 });
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
        await page.goto("/discover");
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
        await link.focus();
        await page.keyboard.press("Enter");
        const status = page
          .getByRole("status")
          .filter({ hasText: "Loading lesson" });
        await expect(status).toBeVisible();
        await expect(status.locator("..")).toHaveAttribute("aria-busy", "true");
        await expect(
          page.getByRole("main").getByRole("heading", { level: 1 }),
        ).toHaveText(resumed ? "confirm" : "Make a plan with a friend", {
          timeout: 10_000,
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
        name: /Today.s practice|Mission complete/,
        exact: true,
      }),
    ).toBeVisible();

    await delayNextServerRequest(page);

    const navigation = page
      .getByRole("navigation", { name: "Primary" })
      .getByRole("link", { name: "Journey" })
      .click();

    const status = page.getByRole("status").filter({
      hasText: "Loading Journey",
    });
    await expect(status).toBeVisible();
    await expect(status.locator("..")).toHaveAttribute("aria-busy", "true");
    await navigation;
    await expect(
      page.getByRole("heading", { name: "Journey", level: 1 }),
    ).toBeVisible({ timeout: 10_000 });
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
        name: /Today.s practice|Mission complete/,
        exact: true,
      }),
    ).toBeVisible();

    await delayNextServerRequest(page);

    const navigation = page.getByRole("link", { name: "Start review" }).click();

    const status = page.getByRole("status").filter({
      hasText: "Loading reviews",
    });
    await expect(status).toBeVisible();
    await expect(status.locator("..")).toHaveAttribute("aria-busy", "true");
    await navigation;
    await expect(
      page.getByRole("heading", { name: "Review", level: 1 }),
    ).toBeVisible({ timeout: 10_000 });
  });

  test("Saved vocabulary shows loading feedback while its data is delayed", async ({
    page,
  }) => {
    await disableViewportPrefetch(page);
    await page.goto("/discover");
    await expect(
      page.getByRole("heading", { name: "Journey", level: 1 }),
    ).toBeVisible();

    await delayNextServerRequest(page);

    const navigation = page.getByRole("link", { name: "Saved words" }).click();
    const status = page.getByRole("status").filter({
      hasText: "Loading saved vocabulary",
    });
    await expect(status).toBeVisible();
    await expect(status.locator("..")).toHaveAttribute("aria-busy", "true");
    await navigation;
    await expect(
      page.getByRole("heading", { name: "Saved vocabulary", level: 1 }),
    ).toBeVisible({ timeout: 10_000 });
  });
});
