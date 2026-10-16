// The empty daily lesson remains useful and keyboard accessible without claiming completion.

import { randomUUID } from "node:crypto";
import { expect, test } from "@playwright/test";

import {
  assertKeyboardReachable,
  assertNonColorOnlyFeedback,
  formatViolations,
  scanForAxeViolations,
} from "./axe-helper.js";

test.describe("Reviews accessibility (VOC-031-T07b)", () => {
  test("/reviews renders with zero critical/serious axe violations, is keyboard reachable, and uses text-based state", async ({
    page,
  }) => {
    await page.context().addCookies([
      { name: "vocanova_session", value: randomUUID(), url: test.info().project.use.baseURL! },
      { name: "e2e_review_fixture_count", value: "0", url: test.info().project.use.baseURL! },
    ]);
    await page.goto("/reviews");

    // The empty-state heading is the most specific signal that
    // the page has rendered with the mocked data.
    await expect(
      page.getByRole("heading", { name: "You’re ready for new words", level: 1 }),
    ).toBeVisible();
    await expect(page.getByRole("link", { name: "Discover words", exact: true })).toBeVisible();
    await expect(page.getByText(/words? remembered for today/)).toHaveCount(0);

    const { criticalOrSerious } = await scanForAxeViolations(page);
    expect(
      criticalOrSerious,
      `Expected zero critical or serious axe-core violations on /reviews; found:\n${formatViolations(
        criticalOrSerious,
      ).join("\n")}`,
    ).toEqual([]);

    // Discovery is available from the empty lesson state.
    await assertKeyboardReachable(page, { minFocusable: 1 });

    await assertNonColorOnlyFeedback(page, {
      contextLabel: "/reviews",
      requireText: [
        "text=You’re ready for new words",
        "text=Discover a word and choose whether to learn it or mark it as known.",
      ],
    });
  });
});
