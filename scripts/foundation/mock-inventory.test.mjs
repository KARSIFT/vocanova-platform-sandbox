import assert from "node:assert/strict";
import { readFileSync } from "node:fs";
import path from "node:path";
import test from "node:test";
import { fileURLToPath } from "node:url";

import { validateMockInventory } from "./mock-inventory.mjs";

// VOC-031-T03: the protected-boundary allow list now includes the
// T03 email-change routes (`/api/v1/settings/email-change-links`),
// the `accounts` business module, the `email_change_links` Ent
// schema, and the `email_change_links` migration, in addition to
// the previously adopted A1/P1/P2/P4-T00/P5-T01/T02 boundaries.
test("VOC-031-T03 mock inventory accepts the email-change backend boundary", () => {
  assert.deepEqual(validateMockInventory(), []);
});

// VOC-030-T06: the T06 cross-cutting safety test files are present
// and the P5-forbidden invariant holds across the T06 deliverable.
// This test is the same code path as the one above; it is
// separately named to make the T06 acceptance criterion visible in
// the test runner output (the production script's success message
// already prints "VOC-030-T06 mock inventory validation passed").
test("VOC-030-T06 mock inventory accepts the T06 cross-cutting test files", () => {
  assert.deepEqual(validateMockInventory(), []);
});

// VOC-031-T06: the P5 cross-cutting reliability test file
// (apps/api/app/api/core_loop_reliability_test.go) and the
// client-side session-expiry helper (apps/web/src/lib/session.ts)
// are both present and the P5-forbidden invariant holds across
// them. This test re-runs the same code path as the two above; it
// is separately named to make the T06 acceptance criterion
// visible in the test runner output.
test("VOC-031-T06 mock inventory accepts the P5 cross-cutting reliability deliverables", () => {
  assert.deepEqual(validateMockInventory(), []);
});

// VOC-031-T07a: the accessibility-automation scaffolding
// (Playwright config, tests/e2e/ tree, axe-helper, mock API
// server, CI workflow) is present and the P5-forbidden
// invariant continues to hold. This test re-runs the same
// code path as the ones above; it is separately named to
// make the T07a acceptance criterion visible in the test
// runner output.
test("VOC-031-T07a mock inventory accepts the accessibility-automation scaffolding", () => {
  assert.deepEqual(validateMockInventory(), []);
});

// VOC-031-T09: the performance-automation harness
// (apps/web/tests/lighthouse/ tree + CI workflow) is present
// and the P5-forbidden invariant continues to hold across
// it. This test re-runs the same code path as the ones
// above; it is separately named to make the T09 acceptance
// criterion visible in the test runner output.
//
// The T09 acceptance criterion requires that the DOC-08
// thresholds (Performance 85+ / Accessibility 95+ / Best
// Practices 90+) be the single source of truth for the
// runner's assertion logic. The thresholds live in
// `tests/lighthouse/assertions.mjs` (the file the runner
// imports) and are mirrored in `tests/lighthouse/budget.json`
// so the budget file is consumable by a future LHCI
// configuration. To keep the two files in lockstep, this
// test loads `budget.json` and asserts it contains the
// exact values the T09 acceptance criterion names, so a
// drift surfaces here.
test("VOC-031-T09 mock inventory accepts the performance-automation harness", () => {
  assert.deepEqual(validateMockInventory(), []);
});

test("VOC-031-T09 budget.json pins the DOC-08 thresholds verbatim", () => {
  const budgetPath = path.resolve(
    path.dirname(fileURLToPath(import.meta.url)),
    "../../apps/web/tests/lighthouse/budget.json",
  );
  const budget = JSON.parse(readFileSync(budgetPath, "utf8"));
  const scores = budget?.budgets?.[0]?.scores ?? {};
  assert.deepEqual(
    scores,
    { performance: 0.85, accessibility: 0.95, "best-practices": 0.9 },
    "T09 budget.json scores must mirror DOC-08 (Performance 85+ / Accessibility 95+ / Best Practices 90+) exactly",
  );
});

// VOC-031-T11: the P5 mock-inventory completeness check is
// the final piece of the package's "zero legacy mocks, no
// P5-invented route/table/behavior beyond this package's
// own documented scope" guarantee. The static guard rails
// T11 installs are:
//
//  - expectedRouteDirectories now includes the T05 Settings
//    and Settings/account (app) routes;
//  - expectedTopLevelRouteDirectories now includes the
//    T01 /onboarding route (which lives at the top of
//    apps/web/src/app/, not inside (app));
//  - the middleware matcher check now requires /settings
//    and /settings/:path* (the T05 additions);
//  - the D06 SupportedAppLanguages value list is pinned to
//    exactly []string{"en"} at the source level;
//  - the P5 staging-evidence document records the T06/T09/
//    T10 audit findings T11 cites as in-repository evidence.
//
// The first four sub-tests below run the same `validateMockInventory()`
// code path the prior tests run; they are separately
// named so the T11 acceptance criterion is visible in the
// test runner output (the production script's success
// message already prints "VOC-031-T11 mock inventory
// validation passed"). The last sub-test asserts the
// D06 appLanguage invariant directly: the source-level
// declaration is the only canonical source for the
// "no multi-language picker before i18n infrastructure
// exists" guarantee, and a future contributor who
// silently widens the set (e.g. to admit a second
// placeholder locale) would surface here.
test("VOC-031-T11 mock inventory accepts the P5 completeness check", () => {
  assert.deepEqual(validateMockInventory(), []);
});

test("VOC-031-T11 SupportedAppLanguages is restricted to en-only at the source", () => {
  const settingsPath = path.resolve(
    path.dirname(fileURLToPath(import.meta.url)),
    "../../apps/api/business/users/settings.go",
  );
  const content = readFileSync(settingsPath, "utf8");
  assert.match(
    content,
    /SupportedAppLanguages\s*=\s*\[\]string\{"en"\}/,
    'VOC-031-D06: apps/api/business/users/settings.go must declare SupportedAppLanguages = []string{"en"} exactly; admitting a second locale without i18n infrastructure would silently claim a capability the product does not have',
  );
});
