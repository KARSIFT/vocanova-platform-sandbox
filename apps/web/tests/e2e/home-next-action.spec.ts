import { randomUUID } from "node:crypto";
import { expect, test } from "@playwright/test";

test.describe("Daily learning entry", () => {
  for (const hasReviews of [false, true]) {
    test(`keeps the ${hasReviews ? "review" : "first-word"} action visible without scrolling`, async ({
      page,
      context,
    }, testInfo) => {
      const baseURL = testInfo.project.use.baseURL!;
      await context.addCookies([
        { name: "vocanova_session", value: randomUUID(), url: baseURL },
        {
          name: "e2e_review_fixture_count",
          value: hasReviews ? "2" : "0",
          url: baseURL,
        },
      ]);
      await page.goto("/home");
      await expect(page.getByRole("heading", { level: 1 })).toHaveCount(1);
      const action = page.getByRole("link", {
        name: hasReviews ? "Start review" : "Start your Journey",
        exact: true,
      });
      await expect(action).toBeVisible();
      await expect(action).toHaveAttribute(
        "href",
        hasReviews ? "/review" : "/discover",
      );
      const actionBox = await action.boundingBox();
      const navigationBox = await page
        .getByRole("navigation", { name: "Primary" })
        .boundingBox();
      expect(actionBox).not.toBeNull();
      expect(navigationBox).not.toBeNull();
      expect(actionBox!.height).toBeGreaterThanOrEqual(44);
      expect(actionBox!.y).toBeGreaterThanOrEqual(0);
      expect(actionBox!.y + actionBox!.height).toBeLessThanOrEqual(
        page.viewportSize()!.height,
      );
      // Navigation moves away from the bottom edge on desktop. The learning
      // action must remain in the viewport and unobscured in either layout.
      const overlapsNavigation =
        actionBox!.x < navigationBox!.x + navigationBox!.width &&
        actionBox!.x + actionBox!.width > navigationBox!.x &&
        actionBox!.y < navigationBox!.y + navigationBox!.height &&
        actionBox!.y + actionBox!.height > navigationBox!.y;
      expect(overlapsNavigation).toBe(false);
      expect(
        await page.evaluate(() => document.documentElement.scrollWidth),
      ).toBeLessThanOrEqual(page.viewportSize()!.width);
      await expect(page.getByRole("textbox")).toHaveCount(0);
    });
  }
});
