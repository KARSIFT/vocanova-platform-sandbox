import { randomUUID } from "node:crypto";
import { expect, test } from "@playwright/test";

test.describe("documented route compatibility", () => {
  test("/login renders the passwordless sign-in experience", async ({ page }) => {
    await page.goto("/login");
    await expect(
      page.getByRole("heading", { name: "Sign in to Vocanova" }),
    ).toBeVisible();
  });

  test("/magic-link renders the magic-link result experience", async ({ page }) => {
    await page.goto("/magic-link");
    await expect(page.getByRole("heading", { name: "Sign in link" })).toBeVisible();
  });

  test("documented review routes render the active review screen", async ({
    page,
    context,
    baseURL,
  }) => {
    const csrf = randomUUID();
    await context.addCookies([
      { name: "vocanova_session", value: randomUUID(), url: baseURL! },
      { name: "vocanova_csrf", value: csrf, url: baseURL! },
      { name: "e2e_review_fixture_count", value: "2", url: baseURL! },
    ]);
    const started = await context.request.post(
      `http://127.0.0.1:${process.env.MOCK_API_PORT ?? 8080}/api/v1/daily-lessons`,
      { headers: { "X-CSRF-Token": csrf, "Idempotency-Key": randomUUID() }, data: {} },
    );
    expect(started.ok()).toBe(true);
    const lesson = await started.json();
    for (const route of ["/review", "/review/session"]) {
      await page.goto(route);
      await expect(page.getByRole("heading", { name: "Review word 1", exact: true })).toBeVisible();
      await expect(page.getByRole("radio")).toHaveCount(4);
      await expect(page).toHaveURL(`${baseURL}/review/session?lesson=${lesson.id}`);
    }
  });
});
