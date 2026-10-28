// Within a range, discovery advances from the server-confirmed decision
// instead of re-reading all 100 words after every choice. Crossing a range
// still reads confirmed state (covered in daily-discovery.spec.ts).

import { randomUUID } from "node:crypto";
import { expect, test } from "@playwright/test";

test("each choice in a range is one confirmed request, and the saved state matches", async ({
  page,
  baseURL,
}) => {
  await page.context().addCookies([
    { name: "vocanova_session", value: `advance-${randomUUID()}`, url: baseURL! },
    { name: "vocanova_csrf", value: "discovery-advance", url: baseURL! },
  ]);
  await page.goto("/discovery?band=1&word=1");
  await page.getByRole("button", { name: "Start discovering", exact: true }).click();
  await expect(page.getByRole("heading", { name: "the", exact: true })).toBeFocused();

  const decisions: string[] = [];
  const mapReads: string[] = [];
  page.on("request", (request) => {
    const url = new URL(request.url());
    if (url.pathname === "/api/v1/knowledge-map/decisions" && request.method() === "POST")
      decisions.push(request.postDataJSON().decision);
    if (url.pathname === "/api/v1/knowledge-map" && request.method() === "GET")
      mapReads.push(url.search);
  });

  await page.getByRole("button", { name: "Already know", exact: true }).click();
  await expect(page.getByRole("heading", { name: "be", exact: true })).toBeFocused();
  await expect(page).toHaveURL(/\/discovery\?band=1&word=2$/);
  await page.getByRole("button", { name: "Learn", exact: true }).click();
  await expect(page.getByRole("heading", { name: "Map word 3", exact: true })).toBeFocused();
  // The decided word shows its confirmed state; the next one is still new.
  await expect(page.getByText("New", { exact: true })).toBeVisible();
  expect(decisions).toEqual(["known", "learn"]);
  expect(mapReads).toEqual([]);

  // The server holds exactly what the session showed.
  await page
    .getByRole("banner")
    .getByRole("link", { name: "Back to Journey", exact: true })
    .click();
  await expect(page).toHaveURL(/\/discover\?band=1#range-words$/);
  await expect(page.getByText("1 already know", { exact: true })).toBeVisible();
  await expect(page.getByText("1 learning", { exact: true })).toBeVisible();
  await expect(
    page.getByText("In this range: 1 already known, 1 learning, 98 new.", {
      exact: true,
    }),
  ).toBeVisible();
});

// A lost response is retried with the same key. The API then reports the
// meaning as it is now, which another tab may have changed in between.
for (const elsewhere of ["known", "removed"] as const) {
  test(`a retried Learn that was ${elsewhere === "known" ? "marked known" : "removed"} elsewhere shows the server's state`, async ({
    page,
    baseURL,
  }) => {
    const api = `http://127.0.0.1:${process.env.MOCK_API_PORT ?? 8080}`;
    const csrf = `replay-${elsewhere}`;
    await page.context().addCookies([
      { name: "vocanova_session", value: `replay-${randomUUID()}`, url: baseURL! },
      { name: "vocanova_csrf", value: csrf, url: baseURL! },
    ]);
    await page.goto("/discovery?band=1&word=1");
    await page.getByRole("button", { name: "Start discovering", exact: true }).click();
    await expect(page.getByRole("heading", { name: "the", exact: true })).toBeFocused();

    const attempts: Array<{ key?: string; body: string | null }> = [];
    await page.route("**/api/v1/knowledge-map/decisions", async (route) => {
      attempts.push({
        key: route.request().headers()["idempotency-key"],
        body: route.request().postData(),
      });
      if (attempts.length === 1) {
        await route.fetch(); // The server applies Learn...
        await route.abort("failed"); // ...but the response is lost.
      } else await route.continue();
    });
    await page.getByRole("button", { name: "Learn", exact: true }).click();
    await expect(page.getByRole("button", { name: "Retry choice", exact: true })).toBeVisible();

    // Another tab changes the same meaning before the retry.
    const headers = { "X-CSRF-Token": csrf };
    if (elsewhere === "known") {
      const map = await (await page.request.get(`${api}/api/v1/knowledge-map?band=1`)).json();
      const word = map.currentBand.items[0];
      expect(word.saved).toBe(true);
      const known = await page.request.post(`${api}/api/v1/knowledge-map/decisions`, {
        headers: { ...headers, "Idempotency-Key": randomUUID() },
        data: { meaningId: word.meaningId, decision: "known", expectedRevision: word.revision },
      });
      expect(known.ok()).toBe(true);
    } else {
      const removed = await page.request.delete(`${api}/api/v1/user-words/map-meaning-1`, { headers });
      expect(removed.ok()).toBe(true);
    }

    const mapReads: string[] = [];
    page.on("request", (request) => {
      const url = new URL(request.url());
      if (url.pathname === "/api/v1/knowledge-map" && request.method() === "GET")
        mapReads.push(url.search);
    });
    await page.getByRole("button", { name: "Retry choice", exact: true }).click();
    if (elsewhere === "known") {
      // The word is not shown as Learning; discovery continues from fresh state.
      await expect(page.getByRole("heading", { name: "be", exact: true })).toBeFocused();
    } else {
      // Removed means new again: the same word returns, honestly labelled.
      await expect(page.getByRole("heading", { name: "the", exact: true })).toBeVisible();
      await expect(page.getByRole("main").getByText("New", { exact: true })).toBeVisible();
      await expect(page.getByRole("button", { name: "Learn", exact: true })).toBeEnabled();
    }
    await expect(page.getByRole("button", { name: "Retry choice", exact: true })).toHaveCount(0);
    expect(attempts).toHaveLength(2);
    expect(attempts[1]).toEqual(attempts[0]);
    expect(mapReads.length).toBeGreaterThan(0);

    await page
      .getByRole("banner")
      .getByRole("link", { name: "Back to Journey", exact: true })
      .click();
    await expect(page.getByText(elsewhere === "known" ? "1 already know" : "0 already know", { exact: true })).toBeVisible();
    await expect(page.getByText("0 learning", { exact: true })).toBeVisible();
  });
}

test("decisions stay within thumb reach on a phone", async ({ page, baseURL }, testInfo) => {
  test.skip(testInfo.project.name === "home-desktop-1280", "Phone layouts only.");
  await page.context().addCookies([
    { name: "vocanova_session", value: `advance-${randomUUID()}`, url: baseURL! },
    { name: "vocanova_csrf", value: "discovery-advance", url: baseURL! },
  ]);
  await page.goto("/discovery?band=1&word=1");
  await page.getByRole("button", { name: "Start discovering", exact: true }).click();
  for (const name of ["Learn", "Already know"]) {
    const button = page.getByRole("button", { name, exact: true });
    await expect(button).toBeInViewport({ ratio: 1 });
    const box = (await button.boundingBox())!;
    expect(box.height).toBeGreaterThanOrEqual(48);
  }
});
