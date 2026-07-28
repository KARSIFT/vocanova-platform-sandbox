// VOC-031-T09 Lighthouse score thresholds and assertion helpers.
//
// DOC-08 ("Quality standards") records the three performance
// budgets for the VocaNova web app:
//
//   Performance    >= 85
//   Accessibility  >= 95
//   Best Practices >= 90
//
// T09 ("Performance automation") enforces these thresholds
// programmatically across the four screens the T09 acceptance
// criterion names (Home, Discover, Reviews, Progress) at the
// three supported layouts (360px, 430px, one representative
// desktop width >=1024px) - 12 audits in total.
//
// Thresholds are mirrored verbatim here from DOC-08 and from
// the T09 acceptance criterion. The constants are the single
// source of truth in this repository: a change to either DOC-08
// or to the T09 acceptance criterion must be reflected here
// AND in the matching test in
// `scripts/foundation/mock-inventory.test.mjs` (which loads
// `budget.json` and asserts these exact values are present).
// Lowering a threshold here without a matching DOC-08 / T09
// update would be a `never self-approve or weaken a check` event
// per `AGENTS.md` and is exactly the failure mode T09 is
// designed to make visible.

export const DOC_08_THRESHOLDS = Object.freeze({
  performance: 85,
  accessibility: 95,
  "best-practices": 90,
});

export const LIGHTHOUSE_CATEGORIES = Object.freeze([
  "performance",
  "accessibility",
  "best-practices",
]);

/**
 * formatCategoryScoreRow renders one (screen, layout, category)
 * result line in the runner's stdout report. Kept as a separate
 * helper so the runner's per-run and summary output can share
 * one formatter.
 */
export function formatCategoryScoreRow({
  screen,
  layout,
  category,
  score,
  threshold,
  pass,
}) {
  const pct = score === null ? "n/a" : Math.round(score * 100);
  const bar = pass ? "PASS" : "FAIL";
  return `  [${bar}] ${screen.padEnd(8)} ${layout.padEnd(14)} ${category.padEnd(15)} score=${pct} (>=${threshold})`;
}

/**
 * assertScores compares the lighthouse category scores against
 * DOC-08 thresholds and returns the list of failing
 * (category, actual, threshold) entries. An empty list means
 * every category met its threshold.
 *
 * Scores are reported by `lighthouse()` as 0..1 floats; the
 * threshold constants are 0..100 integers, so this helper
 * multiplies before comparing.
 */
export function assertScores({ screen, layout, scores }) {
  const failures = [];
  for (const category of LIGHTHOUSE_CATEGORIES) {
    const raw = scores?.[category];
    const score = typeof raw === "number" ? raw : null;
    const threshold = DOC_08_THRESHOLDS[category];
    if (score === null) {
      failures.push({ category, actual: null, threshold, screen, layout });
      continue;
    }
    const scorePct = score * 100;
    if (scorePct < threshold) {
      failures.push({
        category,
        actual: scorePct,
        threshold,
        screen,
        layout,
      });
    }
  }
  return failures;
}
