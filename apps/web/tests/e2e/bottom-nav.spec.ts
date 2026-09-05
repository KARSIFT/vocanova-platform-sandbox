import { expect, test } from "@playwright/test";

test.describe("Primary navigation", () => {
  test("keeps Journey current throughout the discovery flow", async ({ page }) => {
    for (const path of [
      "/discover",
      "/discover/ordering-at-a-cafe",
      "/discover/ordering-at-a-cafe/pour",
    ]) {
      await page.goto(path);

      await expect(
        page.getByRole("navigation", { name: "Primary" }).getByRole("link", {
          name: "Journey",
        }),
      ).toHaveAttribute("aria-current", "page");
    }
  });
});
