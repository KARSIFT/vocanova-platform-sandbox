import { expect, test } from "@playwright/test";

test.describe("Magic-link session recovery", () => {
  test("returns the learner to the requested app route after sign-in", async ({
    page,
  }) => {
    await page.goto(
      "/auth/magic?token=test-token&email=user%40example.com&returnTo=%2Freviews%3Fmode%3Ddue",
    );

    await expect(page).toHaveURL(/\/reviews\?mode=due$/);
  });
});
