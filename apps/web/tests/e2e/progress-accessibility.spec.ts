// VOC-031-T07b accessibility scan for /progress.
//
// The progress page is the most state-heavy screen in the core
// loop - the "Done" / "Rest" day pills and the streak text both
// use colour AND text. The non-color-only assertion specifically
// targets those text labels, which is the T07b acceptance
// criterion's "non-color-only feedback is asserted explicitly,
// not only inferred from a clean axe run" requirement for
// mission/streak state.

import { expect, test } from "@playwright/test";

import {
  assertKeyboardReachable,
  assertNonColorOnlyFeedback,
  formatViolations,
  scanForAxeViolations,
} from "./axe-helper.js";

test.describe("Progress accessibility (VOC-031-T07b)", () => {
  test("/progress renders with zero critical/serious axe violations, is keyboard reachable, and uses text-based state", async ({
    page,
  }) => {
    await page.goto("/progress");

    await expect(
      page.getByRole("heading", { name: "Progress", level: 1 }),
    ).toBeVisible();

    const { criticalOrSerious } = await scanForAxeViolations(page);
    expect(
      criticalOrSerious,
      `Expected zero critical or serious axe-core violations on /progress; found:\n${formatViolations(
        criticalOrSerious,
      ).join("\n")}`,
    ).toEqual([]);

    // The progress page is mostly read-only: the "no saved
    // words" empty state has no anchors, and the "all rest"
    // completion history has no controls either. The page may
    // have zero focusable elements, so the floor is 0 - the
    // read-only nature is itself the correct a11y posture.
    await assertKeyboardReachable(page, { minFocusable: 0 });

    await assertNonColorOnlyFeedback(page, {
      contextLabel: "/progress",
      requireText: [
        "text=Confidence Points",
        "text=Your streaks",
        "text=This week",
        // The "Done" / "Rest" labels on the day pills are the
        // most likely place for a colour-only regression
        // (the pills use bg-primary-100 vs bg-neutral-200
        // plus the text label - the text label is the part
        // this assertion checks).
        "text=Done",
        "text=Rest",
      ],
    });
  });
});
