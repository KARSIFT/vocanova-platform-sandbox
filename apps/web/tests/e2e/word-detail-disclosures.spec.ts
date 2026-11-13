import { openOptionalWriting, openPersonalTools } from "./optional-writing";
import { randomUUID } from "node:crypto";
import { expect, test } from "@playwright/test";
import { scanForAxeViolations } from "./axe-helper";

for (const theme of ["light", "dark"] as const) {
  test(`word details show only available difficulty labels in ${theme}`, async ({
    page,
    context,
    baseURL,
  }) => {
    await context.addCookies([
      { name: "vocanova_session", value: randomUUID(), url: baseURL! },
      { name: "vocanova_theme", value: theme, url: baseURL! },
    ]);
    for (const difficulty of ["unknown", "missing", "default"]) {
      await context.addCookies([
        { name: "e2e_word_difficulty", value: difficulty, url: baseURL! },
      ]);
      for (const path of [
        "/vocabulary/pour",
        "/discover/ordering-at-a-cafe/pour",
      ]) {
        await page.goto(path);
        await expect(page.locator("html")).toHaveAttribute("data-theme", theme);
        // Streamed route payloads can retain an inactive copy outside main.
        const main = page.getByRole("main");
        const heading = main.getByRole("heading", { name: "pour", exact: true });
        await expect(heading).toBeVisible();
        const wordHeader = heading.locator("..");
        await expect(wordHeader.getByText("verb", { exact: true })).toBeVisible();
        await expect(wordHeader.getByText(/^unknown$/i)).toHaveCount(0);
        await expect(wordHeader.getByText("A2", { exact: true })).toHaveCount(
          difficulty === "default" ? 1 : 0,
        );
        await expect(
          main.getByText("to make liquid flow into a container", { exact: true }),
        ).toBeVisible();
        expect(
          await page.evaluate(() => document.documentElement.scrollWidth <= innerWidth),
        ).toBe(true);
      }
    }
  });
}

test("native writing opened before hydration preserves its reload target", async ({
  page,
  context,
  baseURL,
}) => {
  await context.addCookies([
    { name: "vocanova_session", value: randomUUID(), url: baseURL! },
    { name: "vocanova_csrf", value: randomUUID(), url: baseURL! },
  ]);
  await page.goto("/vocabulary/pour");
  await page.getByRole("button", { name: /^Save pour:/ }).click();
  await expect(
    page.getByRole("button", { name: "Remove pour from saved words" }),
  ).toBeVisible();

  // A separate document has no previously running React state. Routing its
  // chunks also disables HTTP cache, so this interaction must use native HTML.
  const nativePage = await context.newPage();
  let releaseChunks!: () => void;
  const chunksReleased = new Promise<void>((resolve) => {
    releaseChunks = resolve;
  });
  let blockedChunks = 0;
  let deadlineReleased = false;
  const deadline = setTimeout(() => {
    deadlineReleased = true;
    releaseChunks();
  }, 15_000);
  await nativePage.route("**/_next/static/**/*.js", async (route) => {
    blockedChunks += 1;
    await chunksReleased;
    await route.continue();
  });
  try {
    await nativePage.goto("/words/uw-mean-pour", { waitUntil: "commit" });
    await expect.poll(() => blockedChunks).toBeGreaterThan(0);
    await openOptionalWriting(nativePage);
    const writing = nativePage.getByRole("textbox", {
      name: "Write a sentence using pour",
      exact: true,
    });
    await expect(writing).toBeVisible();
    // React has not attached its activation handler yet.
    await expect(nativePage).not.toHaveURL(/#sentence-practice$/);
    expect(deadlineReleased).toBe(false);
    const historyLength = await nativePage.evaluate(() => history.length);
    releaseChunks();
    await expect(nativePage).toHaveURL(/#sentence-practice$/);
    expect(await nativePage.evaluate(() => history.length)).toBe(historyLength);
    await writing.fill("I pour tea into my cup.");
    await nativePage.reload();
    await expect(writing).toBeVisible();
    await expect(writing).toHaveValue("I pour tea into my cup.");
    await expect(nativePage.getByRole("main").locator("#sentence-practice")).toHaveAttribute(
      "open",
      "",
    );
  } finally {
    clearTimeout(deadline);
    releaseChunks();
    await nativePage.unrouteAll({ behavior: "wait" });
    await nativePage.close();
  }
});

for (const theme of ["light", "dark"] as const) {
  test(`word teaching stays focused and optional tools retain state in ${theme}`, async ({
    page,
    context,
    baseURL,
  }, testInfo) => {
    await context.addCookies([
      { name: "vocanova_session", value: randomUUID(), url: baseURL! },
      { name: "vocanova_csrf", value: randomUUID(), url: baseURL! },
      { name: "vocanova_theme", value: theme, url: baseURL! },
    ]);
    await page.goto("/vocabulary/pour");
    await page.getByRole("button", { name: /^Save pour:/ }).click();
    await expect(
      page.getByRole("button", { name: "Remove pour from saved words" }),
    ).toBeVisible();
    const writing = page.getByRole("textbox", {
      name: "Write a sentence using pour",
    });
    const practice = page
      .getByRole("main")
      .locator("summary")
      .filter({ hasText: "Optional writing" });
    await expect(writing).toBeHidden();
    await expect(
      page.getByRole("button", { name: "Edit knowledge and note" }),
    ).toBeHidden();
    await expect(
      page.getByText("Could you pour me a cup of coffee?", { exact: true }),
    ).toBeVisible();
    await openPersonalTools(page);
    const historyLength = await page.evaluate(() => window.history.length);
    await practice.focus();
    await page.keyboard.press("Enter");
    await expect(writing).toBeVisible();
    await expect(page).toHaveURL(/#sentence-practice$/);
    expect(await page.evaluate(() => window.history.length)).toBe(
      historyLength,
    );
    await writing.fill("I pour coffee before work.");
    await practice.click();
    await expect(writing).toBeHidden();
    await expect(page).not.toHaveURL(/#sentence-practice$/);
    expect(await page.evaluate(() => window.history.length)).toBe(
      historyLength,
    );
    await practice.click();
    await expect(writing).toHaveValue("I pour coffee before work.");
    await page.goto("/words/uw-mean-pour#sentence-practice");
    await expect(writing).toBeVisible();
    await expect(writing).toHaveValue("I pour coffee before work.");
    await expect(page.getByRole("main").locator("#sentence-practice")).toHaveAttribute(
      "open",
      "",
    );
    // Close using real native activation, then follow a new fragment in the
    // same task. The queued close toggle must not erase that newer target
    // before hashchange opens the writing disclosure again.
    const immediate = await page.evaluate(() => {
      const disclosure =
        document.querySelector<HTMLDetailsElement>("#main-content #sentence-practice")!;
      const wasOpen = disclosure.open;
      disclosure.querySelector("summary")!.click();
      window.location.hash = "sentence-practice";
      return { wasOpen, closed: !disclosure.open, hash: window.location.hash };
    });
    expect(immediate).toEqual({
      wasOpen: true,
      closed: true,
      hash: "#sentence-practice",
    });
    await expect(writing).toBeVisible();
    await expect(page).toHaveURL(/#sentence-practice$/);
    await expect(writing).toHaveValue("I pour coffee before work.");
    const tools = page.getByRole("main").locator("summary").filter({ hasText: "Personal tools" });
    // A writing deep link opens writing independently of optional organization.
    await expect(
      page.getByRole("button", { name: "Edit knowledge and note" }),
    ).toBeHidden();
    await tools.focus();
    await page.keyboard.press("Space");
    await expect(
      page.getByRole("button", { name: "Edit knowledge and note" }),
    ).toBeVisible();
    await page.keyboard.press("Space");
    await expect(
      page.getByRole("button", { name: "Edit knowledge and note" }),
    ).toBeHidden();
    await page.keyboard.press("Space");
    await expect(
      page.getByRole("button", { name: "Edit knowledge and note" }),
    ).toBeVisible();
    expect(
      await page.evaluate(
        () => document.documentElement.scrollWidth <= innerWidth,
      ),
    ).toBe(true);
    // Preserve the actual focused-control evidence before normalizing scroll.
    // Earlier audio controls can be partly behind the sticky header at this
    // offset; they are not the control the learner is currently operating.
    await expect(tools).toBeFocused();
    expect(await tools.evaluate(element => {
      const rect = element.getBoundingClientRect();
      const style = getComputedStyle(element);
      const points = [
        [rect.left + 2, rect.top + 2], [rect.right - 2, rect.top + 2],
        [rect.left + 2, rect.bottom - 2], [rect.right - 2, rect.bottom - 2],
        [rect.left + rect.width / 2, rect.top + rect.height / 2],
      ];
      return rect.width > 0 && rect.height > 0 && rect.top >= 0
        && rect.bottom <= innerHeight && rect.left >= 0 && rect.right <= innerWidth
        && style.outlineStyle !== "none" && parseFloat(style.outlineWidth) >= 2
        && points.every(([x, y]) => {
          const hit = document.elementFromPoint(x!, y!);
          return hit !== null && element.contains(hit);
        });
    }), "Focused Personal tools must remain visible and unobscured").toBe(true);
    await page.screenshot({ path: testInfo.outputPath(`personal-tools-focused-${theme}.png`) });
    const scroll = await page.evaluate(() => ({ x: scrollX, y: scrollY }));
    try {
      await page.evaluate(() => window.scrollTo({ top: 0, left: 0, behavior: "instant" }));
      const { criticalOrSerious } = await scanForAxeViolations(page);
      expect(criticalOrSerious).toEqual([]);
    } finally {
      await page.evaluate(({ x, y }) => window.scrollTo({ top: y, left: x, behavior: "instant" }), scroll);
    }
  });
}

test("manually opened sentence practice restores the visible draft and exact retry after a 401", async ({
  page,
  context,
  baseURL,
}) => {
  await context.addCookies([
    { name: "vocanova_session", value: randomUUID(), url: baseURL! },
    { name: "vocanova_csrf", value: randomUUID(), url: baseURL! },
  ]);
  await page.goto("/vocabulary/pour?from=home");
  await page.getByRole("button", { name: /^Save pour:/ }).click();
  const writing = page.getByRole("textbox", {
    name: "Write a sentence using pour",
  });
  await expect(writing).toBeHidden();
  await openPersonalTools(page);
  await page.getByRole("main").locator("summary").filter({ hasText: "Optional writing" }).click();
  await expect(writing).toBeVisible();
  await expect(page).toHaveURL(
    /\/vocabulary\/pour\?from=home#sentence-practice$/,
  );
  const draft = "I pour water into a glass.";
  await writing.fill(draft);
  // Exercise the fixture's HTTP auth boundary without changing browser
  // request interception while the 401 starts a full-document navigation.
  await context.addCookies([
    { name: "e2e_unauthenticated", value: "1", url: baseURL! },
  ]);
  const rejectedSubmission = page.waitForResponse(
    (response) =>
      response.request().method() === "POST" &&
      response.url().endsWith("/api/v1/learner-sentences"),
  );
  await page.getByRole("button", { name: "Check my sentence" }).click();
  const rejectedResponse = await rejectedSubmission;
  expect(rejectedResponse.status()).toBe(401);
  const firstRequest = {
    key: rejectedResponse.request().headers()["idempotency-key"],
    body: rejectedResponse.request().postDataJSON(),
  };
  await expect(page).toHaveURL(/\/login\?returnTo=/);
  const returnTo = new URL(page.url()).searchParams.get("returnTo");
  expect(returnTo).toBe("/vocabulary/pour?from=home#sentence-practice");
  expect(firstRequest.key).toBeTruthy();

  // The fixture retains this learner's session; emulate the same-user return
  // destination without claiming this exercises a real identity provider.
  await context.clearCookies({ name: "e2e_unauthenticated" });
  await page.goto(returnTo!);
  await expect(writing).toBeVisible();
  await expect(writing).toHaveValue(draft);
  const submitted = page.waitForRequest(
    (request) =>
      request.method() === "POST" &&
      request.url().endsWith("/api/v1/learner-sentences"),
  );
  await page.getByRole("button", { name: "Check my sentence" }).click();
  const retry = await submitted;
  expect(retry.headers()["idempotency-key"]).toBe(firstRequest.key);
  expect(retry.postDataJSON()).toEqual(firstRequest.body);
  await expect(
    page.getByText("Sentence checked", { exact: true }).locator(".."),
  ).toContainText(draft);
});
