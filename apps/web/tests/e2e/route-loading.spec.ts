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

test.describe("Route loading states", () => {
  test("Journey shows loading feedback while its data is delayed", async ({
    page,
  }) => {
    await page.goto("/home");
    await expect(
      page.getByRole("heading", { name: "Today's Mission", level: 1 }),
    ).toBeVisible();

    await delayNextServerRequest(page);
    const navigation = page.getByRole("link", { name: "Go to Journey" }).click();

    await expect(page.getByLabel("Loading Journey")).toBeVisible();
    await expect(page.getByLabel("Loading Journey")).toHaveAttribute(
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
    await page.goto("/home");
    await expect(
      page.getByRole("heading", { name: "Today's Mission", level: 1 }),
    ).toBeVisible();

    await delayNextServerRequest(page);
    const navigation = page.getByRole("link", { name: "Start review" }).click();

    await expect(page.getByLabel("Loading reviews")).toBeVisible();
    await expect(page.getByLabel("Loading reviews")).toHaveAttribute(
      "aria-busy",
      "true",
    );
    await navigation;
    await expect(
      page.getByRole("heading", { name: "Review", level: 1 }),
    ).toBeVisible({ timeout: 10_000 });
  });
});
