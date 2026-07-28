// VOC-031-T07b accessibility scan for /reviews. The fixture data
// the mock API server returns has no due words, so the page
// renders the "all caught up" empty state. That state is the
// real-world landing state for any learner who has finished their
// reviews for the day, and the accessibility bar (axe scan +
// keyboard reachability + non-color-only feedback) applies to it
// the same way it applies to the active review session. The
// active session UI is functionally tested by T08 (the
// DOC-10 §7 full-flow E2E suite), not by the T07b accessibility
// scans.

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
    await page.goto("/reviews");

    // The empty-state heading is the most specific signal that
    // the page has rendered with the mocked data.
    await expect(
      page.getByRole("heading", { name: "Review", level: 1 }),
    ).toBeVisible();
    await expect(
      page.getByRole("heading", { name: "You're all caught up", level: 2 }),
    ).toBeVisible();

    const { criticalOrSerious } = await scanForAxeViolations(page);
    expect(
      criticalOrSerious,
      `Expected zero critical or serious axe-core violations on /reviews; found:\n${formatViolations(
        criticalOrSerious,
      ).join("\n")}`,
    ).toEqual([]);

    // The empty state has the "Back to Home" link as the only
    // focusable element (the header "Back to Home" link is
    // hidden by `dueWords.length === 0 ? <empty> : <session>`
    // so the empty state has exactly one link).
    await assertKeyboardReachable(page, { minFocusable: 1 });

    await assertNonColorOnlyFeedback(page, {
      contextLabel: "/reviews",
      requireText: [
        "text=You're all caught up",
        "text=No words are due for review right now.",
      ],
    });
  });
});
