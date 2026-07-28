// VOC-031-T07a + VOC-031-T07b axe-core / accessibility helpers.
//
// T07a added the WCAG 2.2 AA axe scan + violation formatter. T07b
// extends the helper with explicit keyboard-reachability and
// non-color-only-feedback assertions, because DOC-03 §10 and
// DOC-08 quality standards require more than a clean axe run
// (axe covers most contrast / labelling / structure rules but
// does not exhaustively check WCAG 1.4.1 "Use of Color" or
// keyboard reachability across a page in the way the
// acceptance criterion requires).
//
// What this file owns:
//   - scanForAxeViolations / formatViolations   (T07a, unchanged)
//   - countFocusableElements                    (T07b, keyboard)
//   - assertKeyboardReachable                   (T07b, keyboard)
//   - assertNonColorOnlyFeedback                (T07b, non-color-only)
//
// Do not add rule suppressions or rule-set restrictions to
// scanForAxeViolations without updating the T07a acceptance
// criterion (zero critical/serious axe violations) and the
// corresponding documentation - this helper is the single place
// that defines what "zero critical/serious violations" means.

import type { AxeResults } from "axe-core";
import AxeBuilder from "@axe-core/playwright";
import type { Page } from "@playwright/test";

export type AxeViolation = AxeResults["violations"][number];

export interface AxeScanResult {
  violations: AxeViolation[];
  criticalOrSerious: AxeViolation[];
}

const WCAG_22_AA_TAGS = [
  "wcag2a",
  "wcag2aa",
  "wcag21a",
  "wcag21aa",
  "wcag22aa",
] as const;

/**
 * scanForAxeViolations runs axe-core against the current page with
 * the WCAG 2.2 AA rule set and returns both the raw violations and
 * the subset whose impact is "critical" or "serious" (the T07a
 * acceptance bar).
 *
 * Callers MUST assert the empty-set property on
 * `result.criticalOrSerious` for the test to pass; checking only
 * `result.violations` would let moderate/minor axe findings
 * silently ship, which DOC-03 §10 explicitly disallows at the
 * serious-or-worse threshold.
 */
export async function scanForAxeViolations(
  page: Page,
): Promise<AxeScanResult> {
  const builder = new AxeBuilder({ page }).withTags([...WCAG_22_AA_TAGS]);
  const results = await builder.analyze();
  const criticalOrSerious = results.violations.filter(
    (violation: AxeViolation) =>
      violation.impact === "critical" || violation.impact === "serious",
  );
  return {
    violations: results.violations,
    criticalOrSerious,
  };
}

/**
 * formatViolations renders an axe-core violation as a single line
 * for assertion messages and CI log output. The test report uses
 * it so a failure points at the specific rule, element, and
 * impact level - not just a count.
 */
export function formatViolations(violations: AxeViolation[]): string[] {
  return violations.map((violation: AxeViolation) => {
    const targets = violation.nodes
      .slice(0, 3)
      .map((node) => node.target.join(" "))
      .join("; ");
    const more =
      violation.nodes.length > 3
        ? ` (+${violation.nodes.length - 3} more)`
        : "";
    return `${violation.impact ?? "unknown"} ${violation.id}: ${violation.help} [${targets}${more}]`;
  });
}

// --- T07b: keyboard reachability + non-color-only feedback -----

const FOCUSABLE_SELECTOR = [
  "a[href]",
  "button:not([disabled])",
  "input:not([disabled]):not([type='hidden'])",
  "select:not([disabled])",
  "textarea:not([disabled])",
  "[tabindex]:not([tabindex='-1'])",
].join(",");

export interface KeyboardReachableOptions {
  /**
   * Minimum number of focusable elements expected on the page.
   * Defaults to 1. A screen with no focusable elements at all
   * cannot be operated by keyboard and is by definition
   * inaccessible.
   */
  minFocusable?: number;
  /**
   * Minimum number of distinct Tab stops expected while walking the
   * page. Defaults to `minFocusable`. This is deliberately a
   * separate number: a native `<input type="radio">` group sharing
   * one `name` is a SINGLE Tab stop (the browser moves focus into,
   * not through, the group - options within it are cycled with
   * arrow keys, per native radio semantics), even though every
   * option independently matches the focusable-elements selector
   * `countFocusableElements` counts. A screen built from one or
   * more radio groups therefore has fewer Tab stops than focusable
   * elements without that being an accessibility defect; pass this
   * explicitly on such a screen instead of lowering `minFocusable`
   * (which would also weaken the "do these elements exist at all"
   * check).
   */
  minTabStops?: number;
  /**
   * Maximum number of Tab presses to attempt while looking for
   * focus movement. Defaults to 30 (well above the number of
   * focusable elements on any single core-loop screen).
   */
  maxTabs?: number;
}

/**
 * countFocusableElements returns the number of elements in the
 * current page that can receive keyboard focus. The selector
 * mirrors the set axe-core's own keyboard rule uses, so a
 * discrepancy between this count and axe's own keyboard tabbable
 * audit is a meaningful signal in CI.
 */
export async function countFocusableElements(page: Page): Promise<number> {
  return page.evaluate((selector: string) => {
    return document.querySelectorAll(selector).length;
  }, FOCUSABLE_SELECTOR);
}

/**
 * assertKeyboardReachable asserts that the current page has at
 * least `options.minFocusable` focusable elements AND that at
 * least `options.minTabStops` (defaults to `options.minFocusable`)
 * distinct Tab stops are reachable by pressing Tab sequentially.
 * This is the explicit, screen-by-screen
 * keyboard-reachability check the T07b acceptance criterion
 * requires (axe-core alone does not exhaustively verify
 * tab order across a page).
 *
 * The helper is deliberately conservative: it does not try to
 * verify visual focus indicators (those are covered by axe's
 * `focus-order` / `focus-visible` rules and the screen CSS
 * `focus-visible:outline` pattern every core-loop screen
 * already uses), it only verifies the page is operable from a
 * keyboard.
 */
export async function assertKeyboardReachable(
  page: Page,
  options: KeyboardReachableOptions = {},
): Promise<void> {
  const minFocusable = options.minFocusable ?? 1;
  const minTabStops = options.minTabStops ?? minFocusable;
  const maxTabs = options.maxTabs ?? 30;

  const total = await countFocusableElements(page);
  if (total < minFocusable) {
    throw new Error(
      `Expected at least ${minFocusable} focusable elements on the page; found ${total}. ` +
        `A screen with no keyboard-focusable elements is not operable without a mouse.`,
    );
  }

  // Focus the body first so Tab starts from a known anchor, then
  // press Tab up to `maxTabs` times and record each element that
  // receives focus. The test passes if we observe at least
  // `minFocusable` distinct elements across the press sequence -
  // i.e. focus is moving, not just staying on <body>.
  //
  // Distinctness is tracked by DOM element identity (a Set kept on
  // `window` across evaluate calls), not by a tag+text signature.
  // A tag+text signature collapses radio-group and toggle-row
  // members that carry no text of their own (their label lives in
  // a sibling element) into a single entry, which undercounts a
  // screen that has many same-shaped controls (e.g. a bank of
  // review-target radios) even though every one of them is
  // genuinely, individually reachable by Tab.
  await page.evaluate(() => {
    if (document.activeElement instanceof HTMLElement) {
      document.activeElement.blur();
    }
    document.body.focus();
    (window as unknown as { __e2eSeenFocusable?: Set<Element> }).__e2eSeenFocusable =
      new Set();
  });

  const seenTags: string[] = [];
  for (let i = 0; i < maxTabs; i++) {
    await page.keyboard.press("Tab");
    const tag = await page.evaluate(() => {
      const el = document.activeElement;
      if (!el || el === document.body) {
        return "";
      }
      const seen = (window as unknown as { __e2eSeenFocusable: Set<Element> })
        .__e2eSeenFocusable;
      if (seen.has(el)) {
        return "";
      }
      seen.add(el);
      const text = (el.textContent ?? "").trim().slice(0, 24);
      return `${el.tagName.toLowerCase()}${text ? `:${text}` : ""}`;
    });
    if (tag) {
      seenTags.push(tag);
    }
    if (seenTags.length >= minTabStops) {
      return;
    }
  }

  if (seenTags.length < minTabStops) {
    throw new Error(
      `Tabbing through the page reached only ${seenTags.length} Tab stops; expected at least ${minTabStops}. ` +
        `Seen: ${seenTags.join(", ")}`,
    );
  }
}

export interface NonColorOnlyFeedbackOptions {
  /**
   * Selectors that MUST have non-empty text content. The
   * assertion fails if any matching element's trimmed text
   * content is empty - a status conveyed by color alone fails
   * WCAG 1.4.1 "Use of Color" and is the kind of regression
   * the T07b acceptance criterion calls out explicitly.
   */
  requireText: string[];
  /**
   * Optional human-readable label for the screen being checked,
   * used to make assertion failures self-explanatory in CI logs.
   */
  contextLabel?: string;
}

/**
 * assertNonColorOnlyFeedback walks the selectors in
 * `options.requireText` and asserts that every matching element
 * has non-empty text content. This is the explicit, screen-by-
 * screen non-color-only check the T07b acceptance criterion
 * requires. Axe-core covers `color-contrast` and most labelling
 * rules but does not exhaustively verify that "state conveyed
 * by a colored background" always has a text equivalent, so
 * each core-loop screen names the specific selectors that
 * carry state in its spec and the helper checks each one.
 */
export async function assertNonColorOnlyFeedback(
  page: Page,
  options: NonColorOnlyFeedbackOptions,
): Promise<void> {
  const label = options.contextLabel
    ? ` on ${options.contextLabel}`
    : "";
  for (const selector of options.requireText) {
    const elements = page.locator(selector);
    const count = await elements.count();
    if (count === 0) {
      throw new Error(
        `Non-color-only feedback check${label}: expected at least one element matching "${selector}" but found none. ` +
          `If the screen legitimately no longer renders this element, remove the selector; otherwise this is a regression.`,
      );
    }
    for (let i = 0; i < count; i++) {
      const element = elements.nth(i);
      const text = ((await element.textContent()) ?? "").trim();
      if (text.length === 0) {
        throw new Error(
          `Non-color-only feedback check${label}: element "${selector}" (index ${i}) has no text content. ` +
            `State must be conveyed by text or an accessible icon, not color alone (WCAG 1.4.1).`,
        );
      }
    }
  }
}
