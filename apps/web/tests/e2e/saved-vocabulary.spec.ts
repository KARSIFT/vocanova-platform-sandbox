import { answerReviewWord, startReviewLesson } from "./review-readiness";
import { randomUUID } from "node:crypto";

import { expect, test } from "@playwright/test";

import {
  assertKeyboardReachable,
  formatViolations,
  scanForAxeViolations,
} from "./axe-helper.js";

test.describe("Saved vocabulary library", () => {
  test("save -> library -> detail and practice -> remove", async ({
    page,
    context,
  }, testInfo) => {
    const baseURL = testInfo.project.use.baseURL;
    if (!baseURL) {
      throw new Error("Expected Playwright to configure use.baseURL");
    }

    const mockAPI = `http://127.0.0.1:${process.env.MOCK_API_PORT ?? 8080}`;
    const unauthorizedRecord = await page.request.get(
      `${mockAPI}/api/v1/user-words/records/uw-mean-pour`,
      { headers: { Cookie: "e2e_unauthenticated=1" } },
    );
    expect(unauthorizedRecord.status()).toBe(401);

    await context.addCookies([
      {
        name: "vocanova_session",
        value: `saved-library-${randomUUID()}`,
        url: baseURL,
      },
      {
        name: "vocanova_csrf",
        value: `saved-library-csrf-${randomUUID()}`,
        url: baseURL,
      },
    ]);

    await page.goto("/discover/ordering-at-a-cafe/pour");
    await page.getByRole("button", { name: /^Save pour:/ }).click();
    await expect(
      page.getByRole("button", { name: "Remove pour from saved words" }),
    ).toBeVisible();
    // The local Saved label precedes the route refresh. This disclosure is
    // server-gated by saved=true and userWordId, so its mount confirms that
    // refresh committed before the independent library navigation starts.
    await expect(
      page
        .getByRole("main")
        .locator("summary")
        .filter({ hasText: /^Optional writing$/ }),
    ).toHaveCount(1);

    await page.goto("/words");
    await expect(
      page.getByRole("heading", { name: "Saved vocabulary", level: 1 }),
    ).toBeVisible();
    await expect(page.getByRole("heading", { name: "pour" })).toBeVisible();
    await expect(
      page
        .getByRole("list", { name: "Saved vocabulary results" })
        .getByText("New", { exact: true }),
    ).toBeVisible();

    const listScan = await scanForAxeViolations(page);
    expect(
      listScan.criticalOrSerious,
      `Expected zero critical or serious axe violations on /words; found:\n${formatViolations(
        listScan.criticalOrSerious,
      ).join("\n")}`,
    ).toEqual([]);
    await assertKeyboardReachable(page, { minFocusable: 5 });

    await page
      .getByRole("link", { name: "Open pour details and sentence practice" })
      .click();
    await expect(page).toHaveURL(/\/words\/uw-/);
    await expect(
      page.getByRole("heading", { name: "pour", level: 1 }),
    ).toBeVisible();
    await expect(
      page.getByText("Could you pour me a cup of coffee?"),
    ).toBeVisible();
    await expect(page.getByText("New", { exact: true })).toBeVisible();

    // The saved record keeps its legacy "new" status after a review. Both
    // routes must display the server's learner-facing stage instead.
    await page.goto("/reviews");
    await startReviewLesson(page);
    await answerReviewWord(page, "to make liquid flow into a container");
    await expect(
      page.getByRole("heading", {
        name: "All available reviews finished",
        exact: true,
      }),
    ).toBeVisible();
    await expect(
      page.getByText("1 of 20 words reviewed today", { exact: true }),
    ).toBeVisible();
    await expect(
      page.getByText("1 correct on the first try.", { exact: true }),
    ).toBeVisible();
    await page.goto("/words");
    await expect(
      page
        .getByRole("list", { name: "Saved vocabulary results" })
        .getByText("Learning", { exact: true }),
    ).toBeVisible();
    await page
      .getByRole("link", { name: "Open pour details and sentence practice" })
      .click();
    await expect(page).toHaveURL(/\/words\/uw-/);
    const detail = page.getByRole("main");
    await expect(detail.getByRole("heading", { name: "pour", level: 1 })).toBeVisible();
    await expect(detail.getByText("Learning", { exact: true })).toBeVisible();
    await expect(detail.getByText("New", { exact: true })).toHaveCount(0);
    await detail
      .locator("summary")
      .filter({ hasText: /^Optional writing$/ })
      .click();
    const editor = detail.getByRole("textbox", {
      name: "Write a sentence using pour",
    });
    await expect(editor).toBeVisible();
    await editor.click({ trial: true });

    // Check the editor at its interaction position, then scan from a stable
    // origin so unrelated pronunciation controls aren't under sticky chrome.
    await page.evaluate(() =>
      window.scrollTo({ top: 0, left: 0, behavior: "instant" }),
    );
    const detailScan = await scanForAxeViolations(page);
    expect(
      detailScan.criticalOrSerious,
      `Expected zero critical or serious axe violations on /words/[userWordId]; found:\n${formatViolations(
        detailScan.criticalOrSerious,
      ).join("\n")}`,
    ).toEqual([]);
    await assertKeyboardReachable(page, { minFocusable: 5 });

    await page.getByRole("button", { name: "Remove pour" }).click();
    await expect(page).toHaveURL(/\/words(?:\?|$)/);
    await expect(
      page.getByRole("heading", { name: "Your vocabulary starts here" }),
    ).toBeVisible();

    for (const theme of ["light", "dark"] as const) {
      await context.addCookies([
        { name: "vocanova_theme", value: theme, url: baseURL },
      ]);
      await page.goto("/words");
      await expect(page.locator("html")).toHaveAttribute("data-theme", theme);
      const guidance = page
        .getByRole("main")
        .getByText(
          "Save a useful word from Journey and it will appear here, ready for review and sentence practice.",
          { exact: true },
        );
      await expect(guidance).toBeVisible();
      const guidanceBox = await guidance.boundingBox();
      expect(guidanceBox).not.toBeNull();
      // A spacing-token collision previously squeezed this paragraph to 24px.
      expect(guidanceBox!.width).toBeGreaterThanOrEqual(
        Math.min(320, page.viewportSize()!.width - 64),
      );
      expect(
        await page.evaluate(() => document.documentElement.scrollWidth),
      ).toBeLessThanOrEqual(page.viewportSize()!.width);
      const explore = page.getByRole("link", { name: "Explore Journey" });
      await expect(explore).toHaveAttribute("href", "/discover");
      await explore.focus();
      await expect(explore).toBeFocused();
      await page.screenshot({
        path: testInfo.outputPath(`empty-saved-vocabulary-${theme}.png`),
        fullPage: true,
      });
    }
  });

  test.describe("Meaning save confirmation", () => {
    test.beforeEach(async ({ page, context, baseURL }) => {
      await context.addCookies([
        { name: "vocanova_session", value: randomUUID(), url: baseURL! },
        { name: "vocanova_csrf", value: randomUUID(), url: baseURL! },
      ]);
      await page.goto("/vocabulary/pour");
    });

    for (const recoverCsrf of [false, true]) {
      test(`a lost meaning-save response retries the exact request${recoverCsrf ? " after CSRF recovery" : ""}`, async ({
        page,
        context,
      }) => {
        const main = page.getByRole("main");
        const saved = main.getByRole("button", {
          name: "Remove pour from saved words",
        });
        const writes: { key: string | undefined; body: string | null }[] = [];
        await page.route("**/api/v1/user-words", async (route) => {
          const request = route.request();
          if (request.method() !== "POST") return route.continue();
          writes.push({
            key: request.headers()["idempotency-key"],
            body: request.postData(),
          });
          if (writes.length === 1) {
            const accepted = await route.fetch();
            expect(accepted.ok()).toBe(true);
            return route.abort("failed");
          }
          return route.continue();
        });

        await main.getByRole("button", { name: /^Save pour:/ }).click();
        await expect(main.getByRole("alert")).toBeVisible();
        expect(writes).toHaveLength(1);
        await expect(saved).toHaveCount(0);
        if (recoverCsrf) await context.clearCookies({ name: "vocanova_csrf" });
        await main
          .getByRole("button", { name: /^(?:Retry.*save|Save pour:)/i })
          .and(main.locator(':enabled:not([aria-disabled="true"])'))
          .click();
        await expect(saved).toHaveAttribute("aria-pressed", "true");
        await expect(
          main.getByText("Optional writing", { exact: true }),
        ).toBeVisible();
        expect(writes).toHaveLength(2);
        expect(writes[0]!.key).toBeTruthy();
        expect(JSON.parse(writes[0]!.body!)).toEqual({
          meaningId: "mean-pour",
          source: "search",
        });
        expect(writes[1]).toEqual(writes[0]);
      });
    }

    test("a newer removal wins over a delayed successful save receipt", async ({
      page,
      context,
    }) => {
      const main = page.getByRole("main");
      const save = main.getByRole("button", { name: /^Save pour:/ });
      let writes = 0;
      let releaseReceipt!: () => void;
      const heldReceipt = new Promise<void>((resolve) => {
        releaseReceipt = resolve;
      });
      await page.route("**/api/v1/user-words", async (route) => {
        const request = route.request();
        if (request.method() !== "POST") return route.continue();
        writes += 1;
        if (writes !== 1) return route.continue();
        const receipt = await route.fetch();
        expect(receipt.ok()).toBe(true);
        const removed = await context.request.delete(
          new URL("/api/v1/user-words/mean-pour", request.url()).href,
          { headers: { "X-CSRF-Token": request.headers()["x-csrf-token"]! } },
        );
        expect(removed.status()).toBe(204);
        await heldReceipt;
        return route.fulfill({ response: receipt });
      });

      try {
        await save.click();
        await expect.poll(() => writes).toBe(1);
        await expect(save).toHaveAttribute("aria-busy", "true");
      } finally {
        releaseReceipt();
      }

      await expect(save).toBeEnabled();
      await expect(save).toHaveAttribute("aria-pressed", "false");
      await expect(
        main.getByText("Optional writing", { exact: true }),
      ).toHaveCount(0);
      expect(writes).toBe(1);
    });

    test("a receipt for another meaning leaves the requested meaning unsaved", async ({
      page,
    }) => {
      const main = page.getByRole("main");
      await page.route("**/api/v1/user-words", async (route) => {
        const request = route.request();
        if (request.method() !== "POST") return route.continue();
        const response = await route.fetch({
          postData: { ...request.postDataJSON(), meaningId: "mean-pour-second" },
        });
        expect(response.ok()).toBe(true);
        return route.fulfill({ response });
      });

      await main.getByRole("button", { name: /^Save pour:/ }).click();
      await expect(main.getByRole("alert")).toBeVisible();
      await expect(
        main.getByRole("button", { name: "Remove pour from saved words" }),
      ).toHaveCount(0);
      await expect(
        main.getByText("Optional writing", { exact: true }),
      ).toHaveCount(0);
    });
  });
});

test("saved detail returns to the originating filters after back and removal", async ({
  page,
  context,
  baseURL,
}) => {
  if (!baseURL) throw new Error("Missing app URL");
  await context.addCookies([
    { name: "vocanova_session", value: randomUUID(), url: baseURL },
    { name: "vocanova_csrf", value: randomUUID(), url: baseURL },
  ]);

  await page.goto("/vocabulary/greeting");
  await page.getByRole("button", { name: /^Save greeting:/ }).click();
  await expect(
    page.getByRole("button", { name: "Remove greeting from saved words" }),
  ).toBeVisible();

  const returnPath = "/words?q=greet&stage=new";
  await page.goto(returnPath);
  const main = page.getByRole("main");
  await main.getByRole("link", { name: "greeting", exact: true }).click();
  const back = main.getByRole("link", { name: "Saved words", exact: true });
  await expect(back).toHaveAttribute("href", returnPath);
  await back.click();
  await expect(page).toHaveURL(new URL(returnPath, baseURL).href);
  await expect(main.getByRole("searchbox")).toHaveValue("greet");
  await expect(main.getByRole("combobox", { name: "Review stage" })).toHaveValue(
    "new",
  );

  await main
    .getByRole("link", { name: "Open greeting details and sentence practice" })
    .click();
  await expect(back).toHaveAttribute("href", returnPath);
  await main.getByRole("button", { name: "Remove greeting", exact: true }).click();
  await expect(page).toHaveURL(new URL(returnPath, baseURL).href);
  await expect(
    main.getByRole("heading", { name: "No saved words match these filters" }),
  ).toBeVisible();
});
