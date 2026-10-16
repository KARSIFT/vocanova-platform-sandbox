import { expect, test } from "@playwright/test";
import AxeBuilder from "@axe-core/playwright";

for (const theme of ["light", "dark"] as const) {
  test(`temporary auth outage preserves the destination and offers recovery (${theme})`, async ({ page }, testInfo) => {
    await page.context().addCookies([
      { name: "e2e_auth_check_unavailable", value: "1", url: testInfo.project.use.baseURL! },
      { name: "vocanova_theme", value: theme, url: testInfo.project.use.baseURL! },
    ]);
    const response = await page.goto("/progress?from=practice");
    expect(response?.status()).toBe(503);
    expect(response?.headers()["cache-control"]).toBe("no-store");
    expect(response?.headers()["content-type"]).toContain("text/html");
    await expect(page.locator("html")).toHaveAttribute("data-theme-preference", theme);
    await expect(page).toHaveURL(/\/progress\?from=practice$/);
    await expect(page.getByRole("heading", { name: "We couldn't connect" })).toBeVisible();
    const retry = page.getByRole("link", { name: "Try again", exact: true });
    await expect(retry).toHaveAttribute("href", "/progress?from=practice");
    await expect(page.getByRole("heading", { name: "Sign in to Vocanova" })).toHaveCount(0);
    expect((await new AxeBuilder({ page }).analyze()).violations).toEqual([]);
    expect(await page.evaluate(() => document.documentElement.scrollWidth <= innerWidth)).toBe(true);
    await retry.focus();
    await expect(retry).toBeFocused();
    await page.context().clearCookies({ name: "e2e_auth_check_unavailable" });
    await page.keyboard.press("Enter");
    await expect(page.getByRole("heading", { name: "Progress", exact: true })).toBeVisible();
    await expect(page).toHaveURL(/\/progress\?from=practice$/);
  });
}

test("expired authentication still goes to sign in", async ({ page }, testInfo) => {
  await page.context().addCookies([{ name: "e2e_unauthenticated", value: "1", url: testInfo.project.use.baseURL! }]);
  await page.goto("/home");
  await expect(page).toHaveURL(/\/login\?returnTo=%2Fhome$/);
  await expect(page.getByRole("heading", { name: "Sign in to Vocanova" })).toBeVisible();
});

test("a fresh in-app navigation during an auth outage falls back to a recoverable 503 document", async ({ page, context, baseURL }) => {
  // Do not let an earlier prefetch satisfy this navigation from a healthy cache.
  await page.route("**/plan*", async (route) => {
    if (route.request().headers()["next-router-prefetch"]) await route.abort();
    else await route.continue();
  });
  await page.goto("/home");
  await context.addCookies([{ name: "e2e_auth_check_unavailable", value: "1", url: baseURL! }]);
  const document = page.waitForResponse((response) =>
    response.request().resourceType() === "document" && new URL(response.url()).pathname === "/plan",
  );
  await page.getByRole("link", { name: "Your plan", exact: true }).click();
  const response = await document;
  expect(response.status()).toBe(503);
  expect(response.headers()["cache-control"]).toBe("no-store");
  await expect(page.getByRole("heading", { name: "We couldn't connect", exact: true })).toBeVisible();
  await expect(page).toHaveURL(/\/plan$/);
  const retry = page.getByRole("link", { name: "Try again", exact: true });
  await expect(retry).toHaveAttribute("href", "/plan");
  await context.clearCookies({ name: "e2e_auth_check_unavailable" });
  await retry.click();
  await expect(page.getByRole("heading", { name: "Your learning plan", exact: true })).toBeVisible();
});

test("outage retry destinations stay inside Vocanova and avoid self loops", async ({ page }) => {
  for (const returnTo of ["https://outside.example/", "//outside.example/", "/connection-error?again=1"]) {
    await page.goto(`/connection-error?${new URLSearchParams({ returnTo })}`);
    await expect(page.getByRole("link", { name: "Try again", exact: true })).toHaveAttribute("href", "/home");
  }
});
