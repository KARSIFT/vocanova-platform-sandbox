// VOC-031-T07b accessibility scans for the Discover journey tree:
// /discover, /discover/[situation], and /discover/[situation]/[word]
// (the Word Detail screen, which also embeds the sentence feedback
// widget from sentence-feedback.tsx).
//
// DOC-03 §10 requires the full mobile-first journey to be
// accessible at the three supported layouts. The Word Detail
// screen is the screen where most of the saved-words interaction
// happens (the Save button + the sentence feedback widget), so it
// is also where colour-only feedback and keyboard reachability
// regressions are most likely to land. The non-color-only
// assertion specifically targets the "Saved" badge and the
// meaning-by-meaning card structure on the situation page.

import { expect, test } from "@playwright/test";

import {
  assertKeyboardReachable,
  assertNonColorOnlyFeedback,
  formatViolations,
  scanForAxeViolations,
} from "./axe-helper.js";

test.describe("Discover accessibility (VOC-031-T07b)", () => {
  test("/discover renders with zero critical/serious axe violations, is keyboard reachable, and uses text-based state", async ({
    page,
  }) => {
    await page.goto("/discover");

    await expect(
      page.getByRole("heading", { name: "Journey", level: 1 }),
    ).toBeVisible();

    const { criticalOrSerious } = await scanForAxeViolations(page);
    expect(
      criticalOrSerious,
      `Expected zero critical or serious axe-core violations on /discover; found:\n${formatViolations(
        criticalOrSerious,
      ).join("\n")}`,
    ).toEqual([]);

    // Each situation card is an anchor, so the page has at
    // least one focusable element per card (the fixture serves
    // two cards).
    await assertKeyboardReachable(page, { minFocusable: 2 });

    await assertNonColorOnlyFeedback(page, {
      contextLabel: "/discover",
      requireText: [
        "text=Choose a situation",
        // The two fixture situation cards must each render their
        // title text - the situation grid is the page's primary
        // content.
        "text=Ordering at a cafe",
        "text=Navigating an airport",
      ],
    });
  });

  test("/discover/[situation] renders with zero critical/serious axe violations, is keyboard reachable, and uses text-based state", async ({
    page,
  }) => {
    await page.goto("/discover/ordering-at-a-cafe");

    await expect(
      page.getByRole("heading", { name: "Ordering at a cafe", level: 1 }),
    ).toBeVisible();

    const { criticalOrSerious } = await scanForAxeViolations(page);
    expect(
      criticalOrSerious,
      `Expected zero critical or serious axe-core violations on /discover/[situation]; found:\n${formatViolations(
        criticalOrSerious,
      ).join("\n")}`,
    ).toEqual([]);

    // Each meaning is an anchor; the back link is also focusable.
    // Fixture has two meanings, so at least 3 focusable elements.
    await assertKeyboardReachable(page, { minFocusable: 3 });

    await assertNonColorOnlyFeedback(page, {
      contextLabel: "/discover/[situation]",
      requireText: [
        // The "Saved" indicator on already-saved meanings is the
        // most likely colour-only regression on this screen.
        "text=Saved",
        "text=Back to Journey",
      ],
    });
  });

  test("/discover/[situation]/[word] (Word Detail + sentence feedback) renders with zero critical/serious axe violations, is keyboard reachable, and uses text-based state", async ({
    page,
  }) => {
    await page.goto("/discover/ordering-at-a-cafe/pour");

    await expect(
      page.getByRole("heading", { name: "pour", level: 1 }),
    ).toBeVisible();

    const { criticalOrSerious } = await scanForAxeViolations(page);
    expect(
      criticalOrSerious,
      `Expected zero critical or serious axe-core violations on /discover/[situation]/[word]; found:\n${formatViolations(
        criticalOrSerious,
      ).join("\n")}`,
    ).toEqual([]);

    // Word Detail has a back link, the meaning's Save button, and
    // (when saved) the sentence-feedback form. The unsaved
    // fixture gives us a back link + at least one Save button +
    // the back-to-journey link from the sentence-feedback
    // pattern is not in this fixture. Use a conservative floor.
    await assertKeyboardReachable(page, { minFocusable: 2 });

    await assertNonColorOnlyFeedback(page, {
      contextLabel: "/discover/[situation]/[word]",
      requireText: [
        "text=Meanings",
        "text=to make liquid flow into a container",
        "text=Example sentences",
        "text=Could you pour me a cup of coffee?",
      ],
    });
  });
});
