import { openOptionalWriting } from "./optional-writing";
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
        name: /^(?:Review saved|Review set complete|You’ve finished your review)$/,
        exact: true,
      }),
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
    await expect(page.getByText("Learning", { exact: true })).toBeVisible();
    await expect(page.getByText("New", { exact: true })).toHaveCount(0);
    await openOptionalWriting(page);
    await expect(
      page.getByRole("textbox", { name: "Write a sentence using pour" }),
    ).toBeVisible();

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
});
