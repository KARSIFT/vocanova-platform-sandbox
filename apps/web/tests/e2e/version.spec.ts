import { expect, test } from "@playwright/test";

test.describe("Public build identity", () => {
  test("exposes only sanitized release metadata without authentication", async ({
    page,
  }) => {
    const response = await page.request.get("/version");
    expect(response.ok()).toBeTruthy();
    expect(response.headers()["cache-control"]).toContain("no-store");
    await expect(response.json()).resolves.toEqual({
      version: expect.stringMatching(
        /^(unknown|\d+\.\d+\.\d+(?:-[0-9A-Za-z.-]+)?)$/,
      ),
      commit: expect.stringMatching(/^(unknown|[0-9a-f]{40})$/),
      environment: expect.stringMatching(
        /^(unknown|development|test|staging|production)$/,
      ),
      builtAt: expect.stringMatching(/^(unknown|\d{4}-\d{2}-\d{2}T.*Z)$/),
    });
  });
});
