import { randomUUID } from "node:crypto";
import { expect, test } from "@playwright/test";

test.describe("Primary navigation", () => {
  test("provides one keyboard-usable navigation that adapts to the viewport", async ({
    page,
  }) => {
    await page.goto("/home");
    const navigation = page.getByRole("navigation", { name: "Primary" });
    await expect(navigation).toHaveCount(1);
    await expect(navigation.getByRole("link")).toHaveText([
      "Home",
      "Journey",
      "Progress",
    ]);

    const viewport = page.viewportSize()!;
    const box = await navigation.boundingBox();
    expect(box).not.toBeNull();
    expect(box!.x).toBeGreaterThanOrEqual(0);
    expect(box!.x + box!.width).toBeLessThanOrEqual(viewport.width);
    expect(box!.y + box!.height).toBeLessThanOrEqual(viewport.height);
    if (viewport.width < 1024) {
      expect(box!.y + box!.height).toBe(viewport.height);
    } else {
      expect(box!.y).toBeLessThan(viewport.height / 2);
    }

    const journey = navigation.getByRole("link", { name: "Journey" });
    await journey.focus();
    await expect(journey).toBeFocused();
    await page.keyboard.press("Enter");
    await expect(page).toHaveURL(/\/discover$/);
    await expect(
      navigation.getByRole("link", { name: "Journey" }),
    ).toHaveAttribute("aria-current", "page");
  });

  test("keeps Journey current throughout the discovery flow", async ({
    page,
    context,
    baseURL,
  }) => {
    if (!baseURL) throw new Error("Missing app URL");
    // The guide loads authenticated story links. Legacy /me fixture defaults
    // alone are not a signed-in session for these connected API reads.
    await context.addCookies([
      { name: "vocanova_session", value: randomUUID(), url: baseURL },
    ]);
    for (const path of [
      "/discover",
      "/discover/ordering-at-a-cafe",
      "/discover/ordering-at-a-cafe/pour",
      "/words",
    ]) {
      await page.goto(path);
      await expect(page).toHaveURL(new URL(path, baseURL).toString());

      await expect(
        page.getByRole("navigation", { name: "Primary" }).getByRole("link", {
          name: "Journey",
        }),
      ).toHaveAttribute("aria-current", "page");
    }
  });
});
