import { randomUUID } from "node:crypto";

import { expect, test } from "@playwright/test";

for (const route of ["/reviews", "/review/session"]) {
  test(`${route} preserves server-rendered choices through hydration and queue refresh`, async ({
    page,
    context,
  }, testInfo) => {
    const baseURL = testInfo.project.use.baseURL;
    if (!baseURL) {
      throw new Error(
        "Expected the Playwright project to configure use.baseURL.",
      );
    }
    await context.addCookies([
      {
        name: "vocanova_session",
        value: `review-hydration-${randomUUID()}`,
        url: baseURL,
      },
      { name: "vocanova_csrf", value: "review-hydration-csrf", url: baseURL },
      { name: "e2e_review_fixture_count", value: "4", url: baseURL },
    ]);

    const hydrationErrors: string[] = [];
    page.on("pageerror", (error) => hydrationErrors.push(error.message));
    page.on("console", (message) => {
      if (
        message.type() === "error" &&
        /hydrat|react error #418/i.test(message.text())
      ) {
        hydrationErrors.push(message.text());
      }
    });

    const response = await page.goto(route);
    expect(response).not.toBeNull();
    const serverChoices = await page.evaluate(
      (html) => {
        const document = new DOMParser().parseFromString(html, "text/html");
        const group = [...document.querySelectorAll("fieldset")].find(
          (fieldset) =>
            fieldset
              .querySelector("legend")
              ?.textContent?.startsWith("Choose the meaning for "),
        );
        return [
          ...(group?.querySelectorAll("button > span:first-child") ?? []),
        ].map((label) => label.textContent ?? "");
      },
      await response!.text(),
    );
    expect(serverChoices).toHaveLength(4);

    const group = page.getByRole("group", {
      name: "Choose the meaning for Review word 1",
    });
    const labels = group.getByRole("button").locator("span:first-child");
    await expect(labels).toHaveText(serverChoices);

    // Selecting a choice proves React has hydrated and attached its handlers.
    await group
      .getByRole("button", {
        name: "noun — definition for review word 1",
        exact: true,
      })
      .click();
    await expect(
      page.getByRole("button", { name: "Good", exact: true }),
    ).toBeVisible();
    await expect(labels).toHaveText(serverChoices);
    expect(hydrationErrors).toEqual([]);

    // Force an authoritative queue reload without changing the fixture cards.
    // The mounted session must keep its seed when the same queue is returned.
    await page.route(
      "**/api/v1/reviews/submissions",
      (request) =>
        request.fulfill({
          status: 404,
          contentType: "application/json",
          body: JSON.stringify({ error: "not_found" }),
        }),
      { times: 1 },
    );
    await page.getByRole("button", { name: "Good", exact: true }).click();
    await expect(page.getByRole("status")).toHaveText(
      "This word was removed. Your review list was updated.",
    );
    await expect(labels).toHaveText(serverChoices);
    await group
      .getByRole("button", {
        name: "noun — definition for review word 1",
        exact: true,
      })
      .click();
    await expect(labels).toHaveText(serverChoices);
    expect(hydrationErrors).toEqual([]);

    await page.getByRole("button", { name: "Good", exact: true }).click();
    await expect(
      page.getByRole("heading", { name: "Review word 2", exact: true }),
    ).toBeVisible();
    await expect(
      page.getByRole("button", { name: "Show answer", exact: true }),
    ).toBeEnabled();
  });
}
