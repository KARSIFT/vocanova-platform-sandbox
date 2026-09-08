import { randomUUID } from "node:crypto";

import { expect, test } from "@playwright/test";

const MAX_SENTENCE_CHARACTERS = 300;

test.describe("Sentence feedback character-limit accessibility", () => {
  test("describes the count and announces only when the limit is reached", async ({
    page,
    context,
  }, testInfo) => {
    const baseURL = testInfo.project.use.baseURL;
    if (!baseURL) {
      throw new Error("Expected the Playwright project to configure use.baseURL.");
    }

    const sessionValue = `sentence-a11y-${randomUUID()}`;
    const csrfValue = `sentence-a11y-csrf-${randomUUID()}`;
    await context.addCookies([
      { name: "vocanova_session", value: sessionValue, url: baseURL },
      { name: "vocanova_csrf", value: csrfValue, url: baseURL },
    ]);

    await page.goto("/discover/ordering-at-a-cafe/pour");
    await page.getByRole("button", { name: /Save pour:/ }).click();

    const feedback = page
      .getByRole("heading", { name: "Practice with pour" })
      .locator("..");
    const textarea = feedback.getByRole("textbox", {
      name: "Write a sentence using pour",
    });
    const counter = feedback.locator('[id^="sentence-character-count-"]');
    const limitStatus = feedback.locator(
      '[id^="sentence-character-limit-"]',
    );

    await expect(textarea).toHaveAccessibleDescription("0 of 300 characters");
    await expect(counter).toHaveText("0/300");
    await expect(limitStatus).toHaveCount(0);

    await textarea.fill("a");
    await expect(textarea).toHaveAccessibleDescription("1 of 300 characters");
    await expect(counter).toHaveText("1/300");
    await expect(limitStatus).toHaveCount(0);

    const atLimit = "a".repeat(MAX_SENTENCE_CHARACTERS);
    await textarea.fill(atLimit);
    await expect(textarea).toHaveAccessibleDescription(
      "300 of 300 characters You've reached the 300-character limit.",
    );
    await expect(counter).toHaveText("300/300");
    await expect(limitStatus).toHaveCount(1);
    await expect(limitStatus).toHaveAttribute("role", "status");
    await expect(limitStatus).toHaveText(
      "You've reached the 300-character limit.",
    );

    // The controlled input rejects an over-limit edit, so the existing live
    // region remains mounted rather than being recreated and re-announced.
    await limitStatus.evaluate((node) => {
      node.setAttribute("data-e2e-announcement-instance", "at-limit");
    });
    await textarea.fill(`${atLimit}a`);
    await expect(textarea).toHaveValue(atLimit);
    await expect(counter).toHaveText("300/300");
    await expect(limitStatus).toHaveAttribute(
      "data-e2e-announcement-instance",
      "at-limit",
    );
  });
});
