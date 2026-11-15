import { randomUUID } from "node:crypto";
import { expect, test, type Page } from "@playwright/test";

test("generic discovery re-reads account position after a completed hover prefetch", async ({ page, context, baseURL }, testInfo) => {
  await context.addCookies([
    { name: "vocanova_session", value: `prefetch-${randomUUID()}`, url: baseURL! },
    { name: "vocanova_csrf", value: "prefetch-synthetic", url: baseURL! },
  ]);
  const started = Date.now();
  const events: Array<Record<string, unknown>> = [];
  const observe = (surface: Page, name: string) => {
    surface.on("response", (response) => {
      const request = response.request();
      const url = new URL(response.url());
      if (url.pathname === "/api/v1/discovery-position" && request.method() === "PUT") {
        events.push({ page: name, ms: Date.now() - started, kind: "position-put", rank: request.postDataJSON()?.rank, expectedRevision: request.postDataJSON()?.expectedRevision, status: response.status() });
      }
      if (url.pathname === "/discovery" && url.searchParams.has("_rsc")) {
        events.push({ page: name, ms: Date.now() - started, kind: "discovery-rsc", search: url.search, prefetch: request.headers()["next-router-prefetch"] ?? null, status: response.status() });
      }
    });
  };
  observe(page, "A");
  const other = await context.newPage();
  observe(other, "B");
  let prefetchEvidence: Record<string, unknown> = {};
  try {
    await page.goto("/discovery?band=44&word=4330&active=1&revisit=1");
    await expect(page.getByRole("heading", { name: "Map word 4330", exact: true })).toBeVisible();
    await expect(page.getByRole("status", { name: "Discovery position" })).toHaveText("Your place is saved.");
    await page.getByRole("banner").getByRole("link", { name: "VocaNova home", exact: true }).click();
    await expect(page).toHaveURL("/home");
    const discover = page.getByRole("link", { name: "Discover words", exact: true });
    await expect(discover).toHaveAttribute("href", "/discovery");
    const prefetched = page.waitForResponse((response) => {
      const url = new URL(response.url());
      return url.pathname === "/discovery" && url.searchParams.has("_rsc") && response.request().headers()["next-router-prefetch"] === "1";
    });
    await discover.hover();
    const response = await prefetched;
    const payload = await response.text();
    prefetchEvidence = {
      completedMs: Date.now() - started,
      status: response.status(),
      bytes: Buffer.byteLength(payload),
      selected4330: payload.includes('"initialRank":4330,') || payload.includes('"initialRank":4330}'),
      containsWord4330: payload.includes("Map word 4330"),
      hasLoadingShell: payload.includes("Loading"),
    };
    events.push({ page: "A", kind: "prefetch-complete", ...prefetchEvidence });
    await page.mouse.move(0, 0);

    // B shares the same synthetic requester. Its write starts only after A's
    // prefetch response body has completely arrived and been inspected.
    await other.goto("/discovery?band=44&word=4331&active=1&revisit=1");
    await expect(other.getByRole("heading", { name: "Map word 4331", exact: true })).toBeVisible();
    await expect(other.getByRole("status", { name: "Discovery position" })).toHaveText("Your place is saved.");
    events.push({ page: "B", kind: "4331-confirmed", ms: Date.now() - started });

    await discover.click();
    await expect(page.getByRole("heading", { name: "Map word 4331", exact: true })).toBeVisible();
    await expect(page.getByRole("status", { name: "Discovery position" })).toHaveText("Your place is saved.");
  } finally {
    const result = {
      prefetch: prefetchEvidence,
      events,
      finalURL: page.url(),
      displayedHeading: await page.locator("main h1").allTextContents(),
      positionStatus: await page.getByRole("status", { name: "Discovery position" }).allTextContents(),
    };
    console.log(`DISCOVERY_PREFETCH_EVIDENCE ${JSON.stringify(result)}`);
    await testInfo.attach("discovery-prefetch-evidence.json", { body: JSON.stringify(result, null, 2), contentType: "application/json" });
    await other.close();
  }
});
