import { randomUUID } from "node:crypto";
import { expect, test } from "@playwright/test";

for (const theme of ["light", "dark"] as const) {
  test(`the final revisited Learn retains its confirmed meaning through completion in ${theme}`, async ({
    page,
    context,
    baseURL,
  }) => {
    await context.addCookies([
      { name: "vocanova_session", value: randomUUID(), url: baseURL! },
      { name: "vocanova_csrf", value: "completion-bridge", url: baseURL! },
      { name: "vocanova_theme", value: theme, url: baseURL! },
      { name: "e2e_knowledge_map", value: "complete", url: baseURL! },
    ]);
    // Learn still confirms this exact meaning in a completed range. Removing
    // revisit from its completion URL must not erase the confirmed receipt.
    await page.goto("/discovery?band=100&word=9902&active=1&revisit=1");
    const announcement = page.getByRole("status", { name: "Save confirmation", exact: true });
    await expect(announcement).toBeEmpty();
    const liveRegion = await announcement.elementHandle();
    await page
      .getByRole("button", { name: "Previous word", exact: true })
      .click();
    await expect(page).toHaveURL(/band=100&word=9901&active=1&revisit=1$/);
    await expect(
      page.getByRole("heading", { name: "Map word 9901", exact: true }),
    ).toBeFocused();
    const decision = page.waitForResponse(
      (response) =>
        response.request().method() === "POST" &&
        new URL(response.url()).pathname ===
          "/api/v1/knowledge-map/decisions" &&
        response.request().postDataJSON()?.meaningId === "map-meaning-9901",
    );
    await page.getByRole("button", { name: "Learn", exact: true }).click();
    const result = await decision;
    expect(result.status()).toBe(200);
    expect(await result.json()).toMatchObject({
      meaningId: "map-meaning-9901",
      saved: true,
    });
    await expect(page).toHaveURL(/band=100&word=9901&active=1$/);
    await expect(
      page.getByRole("heading", {
        name: "You’ve explored the whole map",
        exact: true,
      }),
    ).toBeVisible();
    const bridge = page.getByRole("region", {
      name: "Saved meaning",
      exact: true,
    });
    await expect(bridge).toContainText("Saved Map word 9901 for review.");
    await expect(announcement).toHaveText("Saved Map word 9901 for review.");
    expect(await liveRegion!.evaluate((node) => node.isConnected)).toBe(true);
    expect(await announcement.evaluate((node, original) => node === original, liveRegion)).toBe(true);
    await expect(bridge.getByRole("status")).toHaveCount(0);
    await expect(
      bridge.getByRole("link", {
        name: "Use Map word 9901 in a sentence",
        exact: true,
      }),
    ).toHaveAttribute(
      "href",
      "/vocabulary/map-word-9901#meaning-map-meaning-9901",
    );
    // A subsequent history selection is a different visit; the receipt must
    // not follow the learner onto another meaning.
    await page.goBack();
    await expect(
      page.getByRole("heading", { name: "Map word 9902", exact: true }),
    ).toBeVisible();
    await expect(bridge).toHaveCount(0);
    await expect(announcement).toBeEmpty();
  });
}
