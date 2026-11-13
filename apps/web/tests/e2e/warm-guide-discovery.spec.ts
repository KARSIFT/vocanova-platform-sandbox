import { randomUUID } from "node:crypto";
import { expect, test, type Page } from "@playwright/test";

async function learner(page: Page, baseURL: string, theme = "light") {
  await page.context().addCookies([
    {
      name: "vocanova_session",
      value: `warm-discovery-${randomUUID()}`,
      url: baseURL,
    },
    { name: "vocanova_csrf", value: "warm-discovery", url: baseURL },
    { name: "vocanova_theme", value: theme, url: baseURL },
  ]);
}

const savedMeaning = (page: Page) =>
  page.getByRole("region", { name: "Saved meaning", exact: true });
const saveConfirmation = (page: Page) =>
  page.getByRole("status", { name: "Save confirmation", exact: true });

for (const scenario of [
  { label: "fresh Start", active: false, preference: "unset" },
  { label: "direct active entry", active: true, preference: "unset" },
  { label: "stored mute", active: true, preference: "true" },
  { label: "unrecognized preference", active: true, preference: "unexpected" },
  { label: "unavailable storage", active: true, preference: "blocked" },
] as const) {
  test(`Discovery audio stays off until Listen with ${scenario.label}`, async ({ page, baseURL }) => {
    await learner(page, baseURL!);
    await page.addInitScript((preference) => {
      if (preference === "blocked") {
        Object.defineProperty(window, "sessionStorage", {
          get() { throw new DOMException("Storage unavailable", "SecurityError"); },
        });
      } else if (preference !== "unset") {
        sessionStorage.setItem("vocanova-discovery-muted", preference);
      }
    }, scenario.preference);
    const audio: unknown[] = [];
    await page.route("**/api/v1/pronunciation", async (route) => {
      audio.push(route.request().postDataJSON());
      await route.fulfill({ status: 503, json: { title: "Unavailable" } });
    });
    await page.goto(`/discovery?band=1&word=2${scenario.active ? "&active=1" : ""}`);
    if (!scenario.active) {
      await expect.soft(page.getByRole("button", { name: "Sound off", exact: true })).toBeVisible();
      await page.getByRole("button", { name: "Start discovering", exact: true }).click();
    }
    await expect.soft(page.getByRole("button", { name: "Automatic sound off", exact: true })).toBeVisible();
    await page.getByRole("button", { name: "Previous word", exact: true }).click();
    await expect(page.getByRole("heading", { name: "the", exact: true })).toBeFocused();
    expect.soft(audio).toEqual([]);
    audio.length = 0;
    await page.getByRole("button", { name: "Listen to the", exact: true }).click();
    await expect.poll(() => audio.length).toBe(1);
    expect(audio[0]).toMatchObject({ text: "the", meaningId: "map-meaning-1" });
    await expect(page.getByRole("button", { name: "Automatic sound off", exact: true })).toBeVisible();
  });
}

test("Discovery retains explicit automatic sound opt-in and saves a later opt-out", async ({ page, baseURL }) => {
  await learner(page, baseURL!);
  await page.addInitScript(() => {
    sessionStorage.setItem("vocanova-discovery-muted", "false");
  });
  const audio: unknown[] = [];
  await page.route("**/api/v1/pronunciation", async (route) => {
    audio.push(route.request().postDataJSON());
    await route.fulfill({ status: 503, json: { title: "Unavailable" } });
  });
  await page.goto("/discovery?band=1&word=2&active=1");
  const toggle = page.getByRole("button", { name: /^Automatic sound / });
  await expect(toggle).toHaveAccessibleName("Automatic sound on");
  await expect.poll(() => audio.length).toBe(1);
  expect(audio[0]).toMatchObject({ text: "be", meaningId: "map-meaning-2" });
  await toggle.click();
  await expect(toggle).toHaveAccessibleName("Automatic sound off");
  expect(await page.evaluate(() => sessionStorage.getItem("vocanova-discovery-muted"))).toBe("true");
  await page.getByRole("button", { name: "Previous word", exact: true }).click();
  await expect(page.getByRole("heading", { name: "the", exact: true })).toBeFocused();
  expect(audio).toHaveLength(1);
  await page.getByRole("button", { name: "Listen to the", exact: true }).click();
  await expect.poll(() => audio.length).toBe(2);
  expect(await page.evaluate(() => sessionStorage.getItem("vocanova-discovery-muted"))).toBe("true");
});

test("Discovery lets the learner enable automatic sound before Start", async ({ page, baseURL }) => {
  await learner(page, baseURL!);
  let audioRequests = 0;
  await page.route("**/api/v1/pronunciation", async (route) => {
    audioRequests += 1;
    await route.fulfill({ status: 503, json: { title: "Unavailable" } });
  });
  await page.goto("/discovery?band=1&word=2");
  const sound = page.getByRole("button", { name: "Sound off", exact: true });
  await expect(sound).toBeVisible();
  expect(audioRequests).toBe(0);
  await sound.click();
  await expect(page.getByRole("button", { name: "Sound on", exact: true })).toBeVisible();
  expect(await page.evaluate(() => sessionStorage.getItem("vocanova-discovery-muted"))).toBe("false");
  await page.getByRole("button", { name: "Start discovering", exact: true }).click();
  await expect.poll(() => audioRequests).toBe(1);
  await expect(page.getByRole("button", { name: "Automatic sound on", exact: true })).toBeVisible();
});

for (const theme of ["light", "dark"]) {
  test(`Discovery explains its choices on a readable entry surface in ${theme}`, async ({
    page,
    baseURL,
  }) => {
    await learner(page, baseURL!, theme);
    await page.goto("/discovery?band=1&word=1");
    const intro = page.getByRole("region", {
      name: "Discover your next words",
      exact: true,
    });
    await expect(intro).toContainText("Learn saves this meaning for review.");
    await expect(intro).toContainText("Already know records your own choice.");
    const colors = await intro.evaluate((element) => ({
      surface: getComputedStyle(element).backgroundColor,
      foreground: getComputedStyle(element).color,
      canvas: getComputedStyle(document.body).backgroundColor,
    }));
    expect(colors.surface).not.toBe("rgba(0, 0, 0, 0)");
    expect(colors.surface).not.toBe(colors.foreground);
    expect(colors.surface).not.toBe(colors.canvas);
    await intro
      .getByRole("button", { name: "Start discovering", exact: true })
      .press("Enter");
    await expect(
      page.getByRole("heading", { name: "the", exact: true }),
    ).toBeFocused();
    await expect(page.getByRole("article")).toContainText("This meaning");
    expect(
      await page.evaluate(() => document.documentElement.scrollWidth),
    ).toBeLessThanOrEqual(page.viewportSize()!.width);
  });

  test(`confirmed Learn offers the saved meaning while advancing normally in ${theme}`, async ({
    page,
    baseURL,
  }) => {
    await learner(page, baseURL!, theme);
    await page.goto("/discovery?band=1&word=2&active=1");
    const announcement = saveConfirmation(page);
    await expect(announcement).toBeEmpty();
    await expect(announcement).toHaveAttribute("aria-live", "polite");
    await expect(announcement).toHaveAttribute("aria-atomic", "true");
    const liveRegion = await announcement.elementHandle();
    const reads: string[] = [];
    const bodies: unknown[] = [];
    page.on("request", (request) => {
      const path = new URL(request.url()).pathname;
      if (
        request.method() === "GET" &&
        (path === "/api/v1/knowledge-map" ||
          path.startsWith("/api/v1/canonical-words/"))
      )
        reads.push(path);
      if (
        request.method() === "POST" &&
        path.endsWith("/knowledge-map/decisions")
      )
        bodies.push(request.postDataJSON());
    });
    await page.getByRole("button", { name: "Learn", exact: true }).click();
    await expect(
      page.getByRole("heading", { name: "Map word 3", exact: true }),
    ).toBeFocused();
    await expect(page).toHaveURL(/\/discovery\?band=1&word=3&active=1$/);
    await expect(savedMeaning(page)).toContainText("Saved be for review.");
    await expect(announcement).toHaveText("Saved be for review.");
    expect(await liveRegion!.evaluate((node) => node.isConnected)).toBe(true);
    expect(await announcement.evaluate((node, original) => node === original, liveRegion)).toBe(true);
    await expect(savedMeaning(page).getByRole("status")).toHaveCount(0);
    const writing = savedMeaning(page).getByRole("link", {
      name: "Use be in a sentence",
      exact: true,
    });
    await expect(writing).toHaveAttribute(
      "href",
      "/vocabulary/map-word-2#meaning-map-meaning-2",
    );
    await expect(savedMeaning(page)).toContainText(
      "A synthetic meaning for this browser fixture.",
    );
    expect(bodies).toEqual([
      {
        meaningId: "map-meaning-2",
        decision: "learn",
        expectedRevision: "0".repeat(64),
      },
    ]);
    expect(reads).toEqual([]);
    await writing.focus();
    await expect(writing).toBeFocused();
    await page
      .getByRole("button", { name: "Already know", exact: true })
      .click();
    await expect(
      page.getByRole("heading", { name: "Map word 4", exact: true }),
    ).toBeFocused();
    await expect(savedMeaning(page)).toHaveCount(0);
    await expect(announcement).toBeEmpty();
    expect(await liveRegion!.evaluate((node) => node.isConnected)).toBe(true);
    expect(await announcement.evaluate((node, original) => node === original, liveRegion)).toBe(true);
    expect(
      await page.evaluate(() => document.documentElement.scrollWidth),
    ).toBeLessThanOrEqual(page.viewportSize()!.width);
  });
}

test("a lost Learn receipt exposes no writing until the original request is confirmed", async ({
  page,
  baseURL,
}) => {
  await learner(page, baseURL!);
  await page.goto("/discovery?band=1&word=2&active=1");
  const announcement = saveConfirmation(page);
  await expect(announcement).toBeEmpty();
  const attempts: Array<{ body: string | null; key?: string }> = [];
  await page.route("**/api/v1/knowledge-map/decisions", async (route) => {
    attempts.push({
      body: route.request().postData(),
      key: route.request().headers()["idempotency-key"],
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
  await expect(savedMeaning(page)).toHaveCount(0);
  await expect(announcement).toBeEmpty();
  await expect(
    page.getByRole("button", { name: "Choose a word", exact: true }),
  ).toBeDisabled();
  await page.getByRole("button", { name: "Retry choice", exact: true }).click();
  await expect(
    page.getByRole("heading", { name: "Map word 3", exact: true }),
  ).toBeFocused();
  await expect(savedMeaning(page).getByRole("link")).toHaveAttribute(
    "href",
    "/vocabulary/map-word-2#meaning-map-meaning-2",
  );
  await expect(announcement).toHaveText("Saved be for review.");
  expect(attempts).toHaveLength(2);
  expect(attempts[1]).toEqual(attempts[0]);
});

for (const changed of ["known", "removed"] as const) {
  test(`a replay changed to ${changed} cannot offer an earlier saved meaning`, async ({
    page,
    baseURL,
  }) => {
    await learner(page, baseURL!);
    await page.goto("/discovery?band=1&word=2&active=1");
    let first = true;
    await page.route("**/api/v1/knowledge-map/decisions", async (route) => {
      if (first) {
        first = false;
        await route.fetch();
        await route.abort("failed");
      } else await route.continue();
    });
    await page.getByRole("button", { name: "Learn", exact: true }).click();
    await expect(
      page.getByRole("button", { name: "Retry choice", exact: true }),
    ).toBeVisible();
    const api = `http://127.0.0.1:${process.env.MOCK_API_PORT ?? 8080}`;
    if (changed === "known") {
      const map = await (
        await page.request.get(`${api}/api/v1/knowledge-map?band=1`)
      ).json();
      const current = map.currentBand.items[1];
      const response = await page.request.post(
        `${api}/api/v1/knowledge-map/decisions`,
        {
          headers: {
            "X-CSRF-Token": "warm-discovery",
            "Idempotency-Key": randomUUID(),
          },
          data: {
            meaningId: "map-meaning-2",
            decision: "known",
            expectedRevision: current.revision,
          },
        },
      );
      expect(response.ok()).toBe(true);
    } else {
      const response = await page.request.delete(
        `${api}/api/v1/user-words/map-meaning-2`,
        { headers: { "X-CSRF-Token": "warm-discovery" } },
      );
      expect(response.ok()).toBe(true);
    }
    const freshRead = page.waitForResponse((response) => {
      const url = new URL(response.url());
      return (
        response.request().method() === "GET" &&
        url.pathname === "/api/v1/knowledge-map" &&
        url.searchParams.get("band") === "1"
      );
    });
    await page
      .getByRole("button", { name: "Retry choice", exact: true })
      .click();
    const reconciled = await (await freshRead).json();
    expect(reconciled.currentBand.items[1]).toMatchObject({
      meaningId: "map-meaning-2",
      saved: false,
      selfReportedKnown: changed === "known",
    });
    // Neither current state confirms Learn. Existing reconciliation starts
    // from the first unseen word (rank1), including when rank2 was removed.
    // Wait for that settled card: hiding Retry only means the request began.
    await expect(
      page.getByRole("heading", { name: "the", exact: true }),
    ).toBeFocused();
    await expect(page).toHaveURL(/\/discovery\?band=1&word=1&active=1$/);
    await expect(
      page.getByRole("button", { name: "Retry choice", exact: true }),
    ).toHaveCount(0);
    await expect(savedMeaning(page)).toHaveCount(0);
    await expect(saveConfirmation(page)).toBeEmpty();
  });
}

test("an incomplete receipt uses the existing fresh read before offering writing", async ({
  page,
  baseURL,
}) => {
  await learner(page, baseURL!);
  await page.goto("/discovery?band=1&word=2&active=1");
  await page.route("**/api/v1/knowledge-map/decisions", async (route) => {
    await route.fetch();
    await route.fulfill({ json: { applied: true } });
  });
  await page.getByRole("button", { name: "Learn", exact: true }).click();
  await expect(
    page.getByRole("heading", { name: "Map word 3", exact: true }),
  ).toBeFocused();
  await expect(savedMeaning(page).getByRole("link")).toHaveAttribute(
    "href",
    "/vocabulary/map-word-2#meaning-map-meaning-2",
  );
  await expect(saveConfirmation(page)).toHaveText("Saved be for review.");
});

test("manual word, history and range navigation retire the previous save bridge", async ({
  page,
  baseURL,
}) => {
  await learner(page, baseURL!);
  await page.goto("/discovery?band=1&word=2&active=1");
  const announcement = saveConfirmation(page);
  await expect(announcement).toBeEmpty();
  await page.getByRole("button", { name: "Learn", exact: true }).click();
  await expect(savedMeaning(page)).toBeVisible();
  await expect(announcement).toHaveText("Saved be for review.");
  await page
    .getByRole("button", { name: "Previous word", exact: true })
    .click();
  await expect(
    page.getByRole("heading", { name: "be", exact: true }),
  ).toBeFocused();
  await expect(savedMeaning(page)).toHaveCount(0);
  await expect(announcement).toBeEmpty();
  await page.goBack();
  await expect(
    page.getByRole("heading", { name: "Map word 3", exact: true }),
  ).toBeVisible();
  await expect(savedMeaning(page)).toHaveCount(0);
  await expect(announcement).toBeEmpty();
  await page.getByRole("button", { name: "Learn", exact: true }).click();
  await expect(savedMeaning(page).getByRole("link")).toHaveAttribute(
    "href",
    "/vocabulary/map-word-3#meaning-map-meaning-3",
  );
  await expect(announcement).toHaveText("Saved Map word 3 for review.");
  await page.getByRole("button", { name: "Word range", exact: true }).click();
  await page
    .getByRole("dialog", { name: "Choose a range" })
    .getByRole("button", { name: "101–200", exact: true })
    .click();
  await expect(
    page.getByRole("heading", { name: "Map word 101", exact: true }),
  ).toBeFocused();
  await expect(savedMeaning(page)).toHaveCount(0);
  await expect(announcement).toBeEmpty();
});
