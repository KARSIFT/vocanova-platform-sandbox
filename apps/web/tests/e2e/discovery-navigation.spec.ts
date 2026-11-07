import { randomUUID } from "node:crypto";
import { createServer } from "node:http";
import { expect, test } from "@playwright/test";
import { formatViolations, scanForAxeViolations } from "./axe-helper";

for (const theme of ["light", "dark"] as const) {
  test(`discovery navigation preserves the selected word without learning mutations in ${theme}`, async ({
    page,
    baseURL,
  }) => {
    await page.context().addCookies([
      {
        name: "vocanova_session",
        value: `navigate-${randomUUID()}`,
        url: baseURL!,
      },
      { name: "vocanova_theme", value: theme, url: baseURL! },
    ]);
    const decisions: string[] = [];
    page.on("request", (request) => {
      if (
        request.method() === "POST" &&
        new URL(request.url()).pathname.endsWith("/knowledge-map/decisions")
      )
        decisions.push(request.url());
    });
    await page.goto("/discovery?band=12&word=1101&active=1");
    await expect(
      page.getByRole("heading", { name: "Map word 1101", exact: true }),
    ).toBeFocused();
    await expect(
      page.getByRole("button", { name: "Start discovering", exact: true }),
    ).toHaveCount(0);
    await page
      .getByRole("button", { name: "Choose a word", exact: true })
      .click();
    const words = page.getByRole("combobox", {
      name: "Word in this range",
      exact: true,
    });
    await expect(words.locator("option")).toHaveCount(100);
    await words.selectOption("1150");
    await page.getByRole("button", { name: "Go to word", exact: true }).click();
    await expect(
      page.getByRole("heading", { name: "Map word 1150", exact: true }),
    ).toBeFocused();
    await page
      .getByRole("button", { name: "Previous word", exact: true })
      .click();
    await expect(
      page.getByRole("heading", { name: "Map word 1149", exact: true }),
    ).toBeFocused();
    await page.goBack();
    await expect(
      page.getByRole("heading", { name: "Map word 1150", exact: true }),
    ).toBeVisible();
    await page.reload();
    await expect(
      page.getByRole("heading", { name: "Map word 1150", exact: true }),
    ).toBeVisible();
    await expect(
      page.getByRole("main").getByText("Already know", { exact: true }).first(),
    ).toBeVisible();
    const ranges = page.getByRole("combobox", {
      name: "Word range",
      exact: true,
    });
    expect((await ranges.boundingBox())!.height).toBeGreaterThanOrEqual(44);
    await ranges.selectOption("13");
    await expect(
      page.getByRole("button", { name: "Go to range", exact: true }),
    ).toHaveCount(0);
    await expect(
      page.getByRole("heading", { name: "Map word 1201", exact: true }),
    ).toBeFocused();
    await page.goBack();
    await expect(page).toHaveURL(
      /\/discovery\?band=12&word=1150&active=1&revisit=1$/,
    );
    await expect(
      page.getByRole("heading", { name: "Map word 1150", exact: true }),
    ).toBeVisible();
    await page.goForward();
    await expect(page).toHaveURL(/\/discovery\?band=13&word=1201&active=1$/);
    await expect(
      page.getByRole("heading", { name: "Map word 1201", exact: true }),
    ).toBeVisible();
    await page
      .getByRole("button", { name: "Previous word", exact: true })
      .click();
    await expect(
      page.getByRole("heading", { name: "Map word 1200", exact: true }),
    ).toBeFocused();
    expect(decisions).toEqual([]);
    await expect(page).toHaveTitle("Vocanova");
    const { criticalOrSerious } = await scanForAxeViolations(page);
    expect(
      criticalOrSerious,
      formatViolations(criticalOrSerious).join("\n"),
    ).toEqual([]);
    expect(
      await page.evaluate(() => document.documentElement.scrollWidth),
    ).toBeLessThanOrEqual(page.viewportSize()!.width);
  });
}

test("an uncertain choice locks navigation until its original request is confirmed", async ({
  page,
  baseURL,
}) => {
  await page.context().addCookies([
    {
      name: "vocanova_session",
      value: `pending-nav-${randomUUID()}`,
      url: baseURL!,
    },
    { name: "vocanova_csrf", value: "pending-nav", url: baseURL! },
  ]);
  await page.goto("/discovery?band=1&word=2&active=1");
  await expect(
    page.getByRole("heading", { name: "be", exact: true }),
  ).toBeVisible();
  const attempts: Array<{ key?: string; body: string | null }> = [];
  await page.route("**/api/v1/knowledge-map/decisions", async (route) => {
    attempts.push({
      key: route.request().headers()["idempotency-key"],
      body: route.request().postData(),
    });
    if (attempts.length === 1) {
      await route.fetch();
      await route.abort("failed");
    } else await route.continue();
  });
  await page.getByRole("button", { name: "Learn", exact: true }).click();
  await expect(
    page.getByRole("button", { name: "Retry choice", exact: true }),
  ).toBeVisible();
  for (const name of ["Previous word", "Choose a word"])
    await expect(
      page.getByRole("button", { name, exact: true }),
    ).toBeDisabled();
  await expect(
    page.getByRole("combobox", { name: "Word range", exact: true }),
  ).toBeDisabled();
  await page.getByRole("button", { name: "Retry choice", exact: true }).click();
  await expect(
    page.getByRole("heading", { name: "Map word 3", exact: true }),
  ).toBeVisible();
  expect(attempts).toHaveLength(2);
  expect(attempts[1]).toEqual(attempts[0]);
  await expect(
    page.getByRole("button", { name: "Choose a word", exact: true }),
  ).toBeEnabled();
});

test("word details load only on demand, bind the displayed sense and retry honestly", async ({
  page,
  baseURL,
}) => {
  await page.context().addCookies([
    {
      name: "vocanova_session",
      value: `details-${randomUUID()}`,
      url: baseURL!,
    },
  ]);
  const reads: string[] = [];
  let wrongSense = true;
  let failureStatus = 404;
  await page.route("**/api/v1/canonical-words/*", async (route) => {
    reads.push(new URL(route.request().url()).pathname);
    if (failureStatus) {
      await route.fulfill({
        status: failureStatus,
        contentType: "application/problem+json",
        json: { detail: "Raw provider upstream failure" },
      });
      return;
    }
    await route.fulfill({
      json: {
        word: {
          id: "map-word-1",
          text: "the",
          slug: "map-word-1",
          wordType: "word",
          meanings: [
            {
              id: wrongSense ? "different-sense" : "map-meaning-1",
              partOfSpeech: "article",
              shortDefinition: "The currently selected meaning.",
              saved: false,
              selfReportedKnown: false,
              due: false,
              examples: [
                {
                  id: "ex-current",
                  exampleText: "An example for this meaning.",
                },
              ],
              usageNotes: [],
            },
            {
              id: "other-sense",
              partOfSpeech: "article",
              shortDefinition: "Another genuine meaning.",
              saved: false,
              selfReportedKnown: false,
              due: false,
              examples: [],
              usageNotes: [],
            },
          ],
        },
      },
    });
  });
  await page.goto("/discovery?band=1&word=1&active=1");
  await expect(
    page.getByRole("heading", { name: "the", exact: true }),
  ).toBeVisible();
  expect(reads).toEqual([]);
  const details = page.locator("details").filter({
    has: page.getByText("Word details & other meanings", { exact: true }),
  });
  await details.locator("summary").press("Enter");
  for (const nextFailureStatus of [500, 0]) {
    await expect(details.getByRole("alert")).toContainText(
      "Word details could not load. Try again.",
    );
    await expect(details.getByRole("alert")).not.toContainText(
      /HTTP|Raw provider|upstream/,
    );
    failureStatus = nextFailureStatus;
    await page
      .getByRole("button", { name: "Retry word details", exact: true })
      .click();
  }
  await expect(
    page.getByRole("button", { name: "Retry word details", exact: true }),
  ).toBeVisible();
  await expect(
    page.getByText("Another genuine meaning.", { exact: true }),
  ).toHaveCount(0);
  wrongSense = false;
  await page
    .getByRole("button", { name: "Retry word details", exact: true })
    .click();
  await expect(
    details.getByText("article · This meaning", { exact: true }),
  ).toBeVisible();
  await expect(
    details.getByText("An example for this meaning.", { exact: true }),
  ).toBeVisible();
  await expect(
    details.getByText("Another genuine meaning.", { exact: true }),
  ).toBeVisible();
  await details.locator("summary").press("Enter");
  await details.locator("summary").press("Enter");
  expect(reads).toEqual(Array(4).fill("/api/v1/canonical-words/map-word-1"));
  await page
    .getByRole("button", { name: "Choose a word", exact: true })
    .click();
  await page
    .getByRole("combobox", { name: "Word in this range", exact: true })
    .selectOption("2");
  await page.getByRole("button", { name: "Go to word", exact: true }).click();
  await expect(
    page.getByRole("heading", { name: "be", exact: true }),
  ).toBeVisible();
  await expect(
    page.getByText("An example for this meaning.", { exact: true }),
  ).toHaveCount(0);
  await expect(
    page
      .getByRole("banner")
      .getByRole("link", { name: "Back to Journey", exact: true }),
  ).toHaveAttribute("href", "/discover?band=1#range-words");
});

for (const stalledAt of ["headers", "body"] as const) {
  for (const theme of ["light", "dark"] as const) {
    test(`word details recover from stalled ${stalledAt} after eight seconds in ${theme}`, async ({
      page,
      baseURL,
      browserName,
    }) => {
      await page.context().addCookies([
        {
          name: "vocanova_session",
          value: `timeout-details-${randomUUID()}`,
          url: baseURL!,
        },
        { name: "vocanova_theme", value: theme, url: baseURL! },
      ]);
      let reads = 0;
      let cancelledReads = 0;
      const server = createServer((_request, response) => {
        reads++;
        response.setHeader("Access-Control-Allow-Origin", baseURL!);
        response.setHeader("Access-Control-Allow-Credentials", "true");
        response.setHeader("Content-Type", "application/json");
        if (reads === 1) {
          response.on("close", () => {
            if (!response.writableEnded) cancelledReads++;
          });
          if (stalledAt === "body") {
            response.writeHead(200);
            response.write('{"word":');
          }
          return;
        }
        response.end(
          JSON.stringify({
            word: {
              id: "map-word-1",
              text: "the",
              slug: "map-word-1",
              wordType: "word",
              meanings: [
                {
                  id: "map-meaning-1",
                  partOfSpeech: "article",
                  shortDefinition: "The recovered current meaning.",
                  saved: false,
                  selfReportedKnown: false,
                  due: false,
                  examples: [
                    {
                      id: "recovered-example",
                      exampleText: "An example after manual retry.",
                    },
                  ],
                  usageNotes: [],
                },
              ],
            },
          }),
        );
      });
      await new Promise<void>((resolve) =>
        server.listen(0, "127.0.0.1", resolve),
      );
      try {
        const address = server.address();
        if (!address || typeof address === "string")
          throw new Error("Test server did not expose a port");
        await page.route("**/api/v1/canonical-words/map-word-1", (route) =>
          route.continue({ url: `http://127.0.0.1:${address.port}/word` }),
        );
        await page.addInitScript(() => {
          const observed = window as typeof window & {
            __wordDetailsAborts: number;
          };
          observed.__wordDetailsAborts = 0;
          const nativeFetch = window.fetch.bind(window);
          window.fetch = (input, init) => {
            const url = new URL(
              input instanceof Request ? input.url : String(input),
              window.location.href,
            );
            if (url.pathname === "/api/v1/canonical-words/map-word-1") {
              init?.signal?.addEventListener(
                "abort",
                () => {
                  observed.__wordDetailsAborts++;
                },
                { once: true },
              );
            }
            return nativeFetch(input, init);
          };
        });
        await page.clock.install();
        await page.goto("/discovery?band=1&word=1&active=1");
        await expect(
          page.getByRole("heading", { name: "the", exact: true }),
        ).toBeVisible();
        expect(reads).toBe(0);
        const details = page.locator("details").filter({
          has: page.getByText("Word details & other meanings", { exact: true }),
        });
        const receivedHeaders =
          stalledAt === "body"
            ? page.waitForResponse(
                (response) =>
                  response.url() === `http://127.0.0.1:${address.port}/word` &&
                  response.status() === 200,
              )
            : undefined;
        await details.locator("summary").press("Enter");
        await expect.poll(() => reads).toBe(1);
        if (receivedHeaders) await receivedHeaders;
        await expect(details.getByRole("status")).toHaveText(
          "Loading word details…",
        );
        await page.clock.fastForward(8_000);
        await expect(details.getByRole("alert")).toHaveText(
          /Word details could not load\. Try again\./,
        );
        await expect(details.getByRole("status")).toHaveCount(0);
        await expect
          .poll(() =>
            page.evaluate(
              () =>
                (window as typeof window & { __wordDetailsAborts: number })
                  .__wordDetailsAborts,
            ),
          )
          .toBe(1);
        // Firefox's intercepted fetch rejects on abort without immediately
        // closing this fixture's socket. Verify the native signal everywhere;
        // retain the additional observed socket-close check in Chromium.
        if (browserName === "chromium")
          await expect.poll(() => cancelledReads).toBe(1);
        expect(reads).toBe(1);
        await details
          .getByRole("button", { name: "Retry word details", exact: true })
          .press("Enter");
        await expect(
          details.getByText("article · This meaning", { exact: true }),
        ).toBeVisible();
        await expect(
          details.getByText("An example after manual retry.", { exact: true }),
        ).toBeVisible();
        await expect(details.getByRole("alert")).toHaveCount(0);
        expect(reads).toBe(2);
        await page.clock.fastForward(8_000);
        await expect(
          details.getByText("The recovered current meaning.", { exact: true }),
        ).toBeVisible();
        await expect(details.getByRole("alert")).toHaveCount(0);
        expect(reads).toBe(2);
        expect(
          await page.evaluate(
            () =>
              (window as typeof window & { __wordDetailsAborts: number })
                .__wordDetailsAborts,
          ),
        ).toBe(1);
      } finally {
        server.closeAllConnections();
        await new Promise<void>((resolve, reject) =>
          server.close((error) => (error ? reject(error) : resolve())),
        );
      }
    });
  }
}

test("browser Back cannot detach a held or uncertain choice from its displayed word", async ({
  page,
  baseURL,
}) => {
  await page.context().addCookies([
    {
      name: "vocanova_session",
      value: `history-pending-${randomUUID()}`,
      url: baseURL!,
    },
    { name: "vocanova_csrf", value: "history-pending", url: baseURL! },
  ]);
  await page.goto("/discovery?band=1&word=1&active=1");
  await page
    .getByRole("button", { name: "Choose a word", exact: true })
    .click();
  await page
    .getByRole("combobox", { name: "Word in this range", exact: true })
    .selectOption("2");
  await page.getByRole("button", { name: "Go to word", exact: true }).click();
  await expect(
    page.getByRole("heading", { name: "be", exact: true }),
  ).toBeFocused();
  let release!: () => void;
  const held = new Promise<void>((resolve) => {
    release = resolve;
  });
  const attempts: Array<{ key?: string; body: string | null }> = [];
  await page.route("**/api/v1/knowledge-map/decisions", async (route) => {
    attempts.push({
      key: route.request().headers()["idempotency-key"],
      body: route.request().postData(),
    });
    if (attempts.length === 1) {
      await route.fetch();
      await held;
      await route.abort("failed");
    } else await route.continue();
  });
  await page.getByRole("button", { name: "Learn", exact: true }).click();
  await expect(
    page.getByRole("status").filter({ hasText: "Saving your choice…" }),
  ).toBeVisible();
  await page.goBack();
  await expect(
    page.getByRole("heading", { name: "be", exact: true }),
  ).toBeVisible();
  await expect(page).toHaveURL(
    /\/discovery\?band=1&word=2&active=1&revisit=1$/,
  );
  release();
  await expect(
    page.getByRole("button", { name: "Retry choice", exact: true }),
  ).toBeVisible();
  await page.goBack();
  await expect(
    page.getByRole("heading", { name: "be", exact: true }),
  ).toBeVisible();
  await expect(page).toHaveURL(
    /\/discovery\?band=1&word=2&active=1&revisit=1$/,
  );
  await page.getByRole("button", { name: "Retry choice", exact: true }).click();
  await expect(
    page.getByRole("heading", { name: "Map word 3", exact: true }),
  ).toBeVisible();
  expect(attempts).toHaveLength(2);
  expect(attempts[1]).toEqual(attempts[0]);
  // Once confirmed, explicit navigation and ordinary Back work again.
  await page
    .getByRole("button", { name: "Choose a word", exact: true })
    .click();
  await page
    .getByRole("combobox", { name: "Word in this range", exact: true })
    .selectOption("4");
  await page.getByRole("button", { name: "Go to word", exact: true }).click();
  await expect(
    page.getByRole("heading", { name: "Map word 4", exact: true }),
  ).toBeVisible();
  await page.goBack();
  await expect(
    page.getByRole("heading", { name: "Map word 3", exact: true }),
  ).toBeVisible();
});

test("a finished range advances to the next unseen word in a later range", async ({
  page,
  baseURL,
}) => {
  await page.context().addCookies([
    {
      name: "vocanova_session",
      value: `rollover-${randomUUID()}`,
      url: baseURL!,
    },
    { name: "vocanova_csrf", value: "rollover", url: baseURL! },
    { name: "e2e_knowledge_map", value: "gapped-ranges", url: baseURL! },
  ]);
  await page.goto("/discovery?band=12&active=1");
  await expect(
    page.getByRole("heading", { name: "Map word 1200", exact: true }),
  ).toBeFocused();
  await page.getByRole("button", { name: "Learn", exact: true }).click();
  await expect(
    page.getByRole("heading", { name: "Map word 1303", exact: true }),
  ).toBeFocused();
  await expect(page).toHaveURL(/\/discovery\?band=14&word=1303&active=1$/);

  // Direct links to assessed ranks open the next new word too.
  await page.goto("/discovery?band=14&word=1301&active=1");
  await expect(page).toHaveURL(/\/discovery\?band=14&word=1303&active=1$/);
  await page
    .getByRole("button", { name: "Previous word", exact: true })
    .click();
  await expect(
    page.getByRole("heading", { name: "Map word 1302", exact: true }),
  ).toBeFocused();
  await expect(
    page.getByRole("article").getByText("Learning", { exact: true }),
  ).toBeVisible();
  await page
    .getByRole("combobox", { name: "Word range", exact: true })
    .selectOption("13");
  await expect(
    page.getByRole("heading", { name: "Map word 1303", exact: true }),
  ).toBeFocused();
  await expect(page).toHaveURL(/\/discovery\?band=14&word=1303&active=1$/);
  await page
    .getByRole("button", { name: "Choose a word", exact: true })
    .click();
  const words = page.getByRole("combobox", {
    name: "Word in this range",
    exact: true,
  });
  await words.selectOption("1301");
  await page.getByRole("button", { name: "Go to word", exact: true }).click();
  await expect(
    page.getByRole("heading", { name: "Map word 1301", exact: true }),
  ).toBeFocused();
  await expect(
    page.getByRole("article").getByText("Already know", { exact: true }),
  ).toBeVisible();
  await expect(page).toHaveURL(
    /\/discovery\?band=14&word=1301&active=1&revisit=1$/,
  );
  await page
    .getByRole("button", { name: "Previous word", exact: true })
    .click();
  await expect(
    page.getByRole("heading", { name: "Map word 1300", exact: true }),
  ).toBeFocused();
  await expect(
    page.getByRole("article").getByText("Already know", { exact: true }),
  ).toBeVisible();
});

test("browser Forward skips an assessment made after opening another range", async ({
  page,
  baseURL,
}) => {
  await page.context().addCookies([
    {
      name: "vocanova_session",
      value: `range-history-${randomUUID()}`,
      url: baseURL!,
    },
    { name: "vocanova_csrf", value: "range-history", url: baseURL! },
  ]);
  await page.goto("/discovery?band=12&word=1101&active=1");
  await page
    .getByRole("combobox", { name: "Word range", exact: true })
    .selectOption("13");
  const first = page.getByRole("heading", {
    name: "Map word 1201",
    exact: true,
  });
  const next = page.getByRole("heading", {
    name: "Map word 1202",
    exact: true,
  });
  await expect(first).toBeFocused();
  await page.getByRole("button", { name: "Already know", exact: true }).click();
  await expect(next).toBeFocused();
  await page.goBack();
  await expect(
    page.getByRole("heading", { name: "Map word 1101", exact: true }),
  ).toBeVisible();
  await page.goForward();
  await expect(next).toBeVisible();
  await expect(first).toHaveCount(0);
  await page
    .getByRole("button", { name: "Previous word", exact: true })
    .click();
  await expect(first).toBeFocused();
  await expect(
    page.getByRole("article").getByText("Already know", { exact: true }),
  ).toBeVisible();
});

test("finishing the final range keeps earlier new words an explicit choice", async ({
  page,
  baseURL,
}) => {
  await page.context().addCookies([
    {
      name: "vocanova_session",
      value: `complete-history-${randomUUID()}`,
      url: baseURL!,
    },
    { name: "vocanova_csrf", value: "complete-history", url: baseURL! },
    { name: "e2e_knowledge_map", value: "gapped-ranges", url: baseURL! },
  ]);
  await page.goto("/discovery?band=100&active=1");
  await page
    .getByRole("button", { name: "Choose a word", exact: true })
    .click();
  await page
    .getByRole("combobox", { name: "Word in this range", exact: true })
    .selectOption("10000");
  await page.getByRole("button", { name: "Go to word", exact: true }).click();
  await expect(
    page.getByRole("heading", { name: "Map word 10000", exact: true }),
  ).toBeVisible();
  await page.getByRole("button", { name: "Already know", exact: true }).click();
  await expect(
    page.getByRole("heading", { name: "Map word 9999", exact: true }),
  ).toBeVisible();
  await page.getByRole("button", { name: "Learn", exact: true }).click();
  const finished = page.getByRole("heading", {
    name: "You’ve finished this range",
    exact: true,
  });
  await expect(finished).toBeFocused();
  await expect(page).toHaveURL(/\/discovery\?band=100&word=9999&active=1$/);
  await page.goBack();
  await expect(finished).toBeVisible();
  await expect(
    page.getByRole("button", { name: "Already know", exact: true }),
  ).toHaveCount(0);
  await page
    .getByRole("link", { name: "Continue with earlier words", exact: true })
    .click();
  await expect(
    page.getByRole("heading", { name: "Map word 9801", exact: true }),
  ).toBeFocused();
});

test("browser Back skips an assessment made from a different visit to the same range", async ({
  page,
  baseURL,
}) => {
  await page.context().addCookies([
    {
      name: "vocanova_session",
      value: `range-revisit-${randomUUID()}`,
      url: baseURL!,
    },
    { name: "vocanova_csrf", value: "range-revisit", url: baseURL! },
  ]);
  await page.goto("/discovery?band=13&word=1201&active=1");
  const first = page.getByRole("heading", {
    name: "Map word 1201",
    exact: true,
  });
  const next = page.getByRole("heading", {
    name: "Map word 1202",
    exact: true,
  });
  const other = page.getByRole("heading", {
    name: "Map word 1101",
    exact: true,
  });
  const ranges = page.getByRole("combobox", {
    name: "Word range",
    exact: true,
  });
  await expect(first).toBeFocused();
  await ranges.selectOption("12");
  await expect(other).toBeFocused();
  await ranges.selectOption("13");
  await expect(first).toBeFocused();
  await page.getByRole("button", { name: "Learn", exact: true }).click();
  await expect(next).toBeFocused();
  await page.goBack();
  await expect(other).toBeVisible();
  await page.goBack();
  await expect(next).toBeVisible();
  await expect(first).toHaveCount(0);
});

test("details reject a newer revision of the same meaning until the card is refreshed", async ({
  page,
  baseURL,
}) => {
  await page.context().addCookies([
    {
      name: "vocanova_session",
      value: `revision-details-${randomUUID()}`,
      url: baseURL!,
    },
    { name: "vocanova_csrf", value: "revision-details", url: baseURL! },
  ]);
  const cardRevision = "a".repeat(64);
  const detailsRevision = "b".repeat(64);
  // A real conflict recovery reads current map data; add a catalog revision to
  // this response without changing the shared synthetic fixture.
  await page.route("**/api/v1/knowledge-map?band=1", async (route) => {
    const response = await route.fetch();
    const map = await response.json();
    map.currentBand.items[0].meaningRevision = cardRevision;
    await route.fulfill({ json: map });
  });
  await page.route("**/api/v1/knowledge-map/decisions", (route) =>
    route.fulfill({
      status: 409,
      json: { error: { code: "conflict", message: "Word changed" } },
    }),
  );
  await page.route("**/api/v1/canonical-words/map-word-1", (route) =>
    route.fulfill({
      json: {
        word: {
          id: "map-word-1",
          slug: "map-word-1",
          text: "the",
          wordType: "word",
          meanings: [
            {
              id: "map-meaning-1",
              meaningRevision: detailsRevision,
              partOfSpeech: "article",
              shortDefinition: "Teaching from the requested revision.",
              saved: false,
              selfReportedKnown: false,
              due: false,
              examples: [],
              usageNotes: [],
            },
          ],
        },
      },
    }),
  );
  await page.goto("/discovery?band=1&word=1&active=1");
  await page.getByRole("button", { name: "Learn", exact: true }).click();
  await expect(page.getByRole("main").getByRole("alert")).toContainText(
    "This word changed elsewhere.",
  );
  await page
    .getByText("Word details & other meanings", { exact: true })
    .click();
  const reload = page.getByRole("button", { name: "Reload word", exact: true });
  await expect(reload).toBeVisible();
  await expect(
    page.getByRole("button", { name: "Retry word details", exact: true }),
  ).toHaveCount(0);
  await expect(
    page.getByText("Teaching from the requested revision.", { exact: true }),
  ).toHaveCount(0);
  // This fixture has no SSR meaning revision. The real document reload must
  // replace the stale client card; this proves recovery navigation and a new
  // details read, not that a production catalog revision changed on the server.
  await Promise.all([
    page.waitForNavigation({ waitUntil: "domcontentloaded" }),
    reload.click(),
  ]);
  await expect(
    page.getByRole("heading", { name: "the", exact: true }),
  ).toBeVisible();
  await expect(reload).toHaveCount(0);
  await page
    .getByText("Word details & other meanings", { exact: true })
    .click();
  await expect(
    page.getByText("Teaching from the requested revision.", { exact: true }),
  ).toBeVisible();
});

for (const theme of ["light", "dark"] as const) {
  test(`discovery Start and Sound controls share aligned full-height targets in ${theme}`, async ({
    page,
    baseURL,
  }) => {
    await page.context().addCookies([
      {
        name: "vocanova_session",
        value: `start-alignment-${randomUUID()}`,
        url: baseURL!,
      },
      { name: "vocanova_theme", value: theme, url: baseURL! },
    ]);
    await page.goto("/discovery?band=1&word=1");
    await expect(page.locator("html")).toHaveAttribute("data-theme", theme);
    await expect(
      page.getByText("Listen, then choose Learn or Already know.", {
        exact: true,
      }),
    ).toBeVisible();
    const start = page.getByRole("button", {
      name: "Start discovering",
      exact: true,
    });
    const sound = page.getByRole("button", { name: "Sound on", exact: true });
    await expect(start).toBeVisible();
    await expect(sound).toBeVisible();
    await expect
      .poll(async () => {
        const first = await start.boundingBox();
        const second = await sound.boundingBox();
        return Boolean(
          first &&
          second &&
          first.height >= 44 &&
          second.height >= 44 &&
          Math.abs(first.y - second.y) <= 1 &&
          Math.abs(first.height - second.height) <= 1,
        );
      })
      .toBe(true);
    await sound.focus();
    await page.keyboard.press("Enter");
    await expect(
      page.getByRole("button", { name: "Sound off", exact: true }),
    ).toBeFocused();
    await start.focus();
    await page.keyboard.press("Enter");
    await expect(
      page.getByRole("heading", { name: "the", exact: true }),
    ).toBeFocused();
  });
}

for (const { choice, partial } of [
  { choice: "Learn", partial: false },
  { choice: "Already know", partial: false },
  { choice: "Learn", partial: true },
] as const) {
  test(`${choice} continues after a chosen word and wraps only at the end of its range${partial ? " after a partial response" : ""}`, async ({
    page,
    baseURL,
  }) => {
    await page.context().addCookies([
      {
        name: "vocanova_session",
        value: `chosen-continuation-${randomUUID()}`,
        url: baseURL!,
      },
      { name: "vocanova_csrf", value: "chosen-continuation", url: baseURL! },
    ]);
    if (partial)
      await page.route("**/api/v1/knowledge-map/decisions", async (route) => {
        await route.fetch(); // The API applies the exact original decision.
        await route.fulfill({ json: { applied: true } }); // No usable state receipt.
      });
    await page.goto("/discovery?band=12&word=1101&active=1");
    await page
      .getByRole("button", { name: "Choose a word", exact: true })
      .click();
    await page
      .getByRole("combobox", { name: "Word in this range", exact: true })
      .selectOption("1150");
    await page.getByRole("button", { name: "Go to word", exact: true }).click();
    await expect(
      page.getByRole("heading", { name: "Map word 1150", exact: true }),
    ).toBeFocused();
    await page.reload();
    await expect(
      page.getByRole("heading", { name: "Map word 1150", exact: true }),
    ).toBeVisible();
    await page.getByRole("button", { name: choice, exact: true }).click();
    await expect(
      page.getByRole("heading", { name: "Map word 1151", exact: true }),
    ).toBeFocused();
    await expect(page).toHaveURL(/\/discovery\?band=12&word=1151&active=1$/);
    await page
      .getByRole("button", { name: "Choose a word", exact: true })
      .click();
    await page
      .getByRole("combobox", { name: "Word in this range", exact: true })
      .selectOption("1200");
    await page.getByRole("button", { name: "Go to word", exact: true }).click();
    await expect(
      page.getByRole("heading", { name: "Map word 1200", exact: true }),
    ).toBeFocused();
    await page.getByRole("button", { name: choice, exact: true }).click();
    await expect(
      page.getByRole("heading", { name: "Map word 1101", exact: true }),
    ).toBeFocused();
    await expect(page).toHaveURL(/\/discovery\?band=12&word=1101&active=1$/);
  });
}
