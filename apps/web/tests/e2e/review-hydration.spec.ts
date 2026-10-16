import { randomUUID } from "node:crypto";
import { expect, test } from "@playwright/test";
import { answerReviewWord } from "./review-readiness";

for (const route of ["/reviews", "/review/session"]) {
  test(`${route} preserves server-rendered choices through hydration and saved-progress refresh`, async ({
    page,
    context,
    baseURL,
  }) => {
    const csrf = "review-hydration-csrf";
    await context.addCookies([
      {
        name: "vocanova_session",
        value: `hydration-${randomUUID()}`,
        url: baseURL!,
      },
      { name: "vocanova_csrf", value: csrf, url: baseURL! },
      { name: "e2e_review_fixture_count", value: "4", url: baseURL! },
    ]);
    const started = await context.request.post(
      `http://127.0.0.1:${process.env.MOCK_API_PORT ?? 8080}/api/v1/daily-lessons`,
      {
        headers: { "X-CSRF-Token": csrf, "Idempotency-Key": randomUUID() },
        data: {},
      },
    );
    expect(started.ok()).toBe(true);
    const lesson = await started.json();
    const hydrationErrors: string[] = [];
    page.on("pageerror", (error) => hydrationErrors.push(error.message));
    page.on("console", (message) => {
      if (
        message.type() === "error" &&
        /hydrat|react error #418/i.test(message.text())
      )
        hydrationErrors.push(message.text());
    });
    const response = await page.goto(
      route === "/reviews" ? route : `${route}?lesson=${lesson.id}`,
    );
    expect(response).not.toBeNull();
    const serverChoices = await page.evaluate(
      (html) => {
        const document = new DOMParser().parseFromString(html, "text/html");
        return [
          ...document.querySelectorAll("fieldset label > span:first-of-type"),
        ].map((label) => label.textContent ?? "");
      },
      await response!.text(),
    );
    expect(serverChoices).toHaveLength(4);
    const group = page.getByRole("group", {
      name: "Meaning of Review word 1",
      exact: true,
    });
    const labels = group.locator("label > span:first-of-type");
    await expect(group.getByRole("radio")).toHaveCount(4);
    await expect(labels).toHaveText(serverChoices);
    await group
      .getByRole("radio", { name: "definition for review word 1", exact: true })
      .check();
    await page.route(
      "**/api/v1/daily-lessons/*/actions",
      (route) =>
        route.fulfill({
          status: 409,
          contentType: "application/json",
          body: JSON.stringify({ detail: "stale" }),
        }),
      { times: 1 },
    );
    await page
      .getByRole("button", { name: "Check answer", exact: true })
      .click();
    await expect(page.getByRole("main").getByRole("alert")).toContainText(
      "lesson changed",
    );
    await page
      .getByRole("button", { name: "Load saved progress", exact: true })
      .click();
    await expect(group.getByRole("radio").first()).toBeEnabled();
    await expect(labels).toHaveText(serverChoices);
    await answerReviewWord(page, "definition for review word 1");
    await expect(
      page.getByRole("heading", { name: "Review word 2", exact: true }),
    ).toBeVisible();
    await expect(page.getByRole("radio").first()).toBeEnabled();
    expect(hydrationErrors).toEqual([]);
  });
}
