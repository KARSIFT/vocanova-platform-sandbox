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
  test("Journey shows loading feedback while its data is delayed", async ({
    page,
  }) => {
    await disableViewportPrefetch(page);
    await page.goto("/home");
    await expect(
      page.getByRole("heading", { name: "Today's Mission", level: 1 }),
    ).toBeVisible();

    await delayNextServerRequest(page);

    const navigation = page.getByRole("link", { name: "Go to Journey" }).click();

    const status = page.getByRole("status").filter({
      hasText: "Loading Journey",
    });
    await expect(status).toBeVisible();
    await expect(status.locator("..")).toHaveAttribute(
      "aria-busy",
      "true",
    );
    await navigation;
    await expect(
      page.getByRole("heading", { name: "Journey", level: 1 }),
    ).toBeVisible({ timeout: 10_000 });
  });

  test("Review shows loading feedback while its data is delayed", async ({
    page,
  }) => {
    await disableViewportPrefetch(page);
    await page.goto("/home");
    await expect(
      page.getByRole("heading", { name: "Today's Mission", level: 1 }),
    ).toBeVisible();

    await delayNextServerRequest(page);

    const navigation = page.getByRole("link", { name: "Start review" }).click();

    const status = page.getByRole("status").filter({
      hasText: "Loading reviews",
    });
    await expect(status).toBeVisible();
    await expect(status.locator("..")).toHaveAttribute(
      "aria-busy",
      "true",
    );
    await navigation;
    await expect(
      page.getByRole("heading", { name: "Review", level: 1 }),
    ).toBeVisible({ timeout: 10_000 });
  });
});
