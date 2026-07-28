// VOC-031-T07a axe-core analysis helper.
//
// The Home page accessibility test (and every future screen test
// added in T07b) runs axe-core through @axe-core/playwright's
// AxeBuilder. This helper centralises the rule set so every scan
// uses the same WCAG 2.2 AA baseline DOC-03 §10 / DOC-08 quality
// standards require, and the violation-impact filter (critical +
// serious only) the harness asserts on. The "critical" and
// "serious" impact levels match axe-core's own severity scale and
// the DOC-08 pass bar.
//
// Do not add rule suppressions or rule-set restrictions here
// without updating the T07a acceptance criterion (zero
// critical/serious axe violations) and the corresponding
// documentation - this helper is the single place that defines
// what "zero critical/serious violations" means.

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
