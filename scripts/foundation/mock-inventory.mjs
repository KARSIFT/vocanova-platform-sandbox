import { globSync, readFileSync, readdirSync, statSync } from "node:fs";
import path from "node:path";
import process from "node:process";
import { fileURLToPath } from "node:url";

const repositoryRoot = path.resolve(
  path.dirname(fileURLToPath(import.meta.url)),
  "../..",
);

const appRouteRoot = path.join(repositoryRoot, "apps/web/src/app/(app)");
const webSrcRoot = path.join(repositoryRoot, "apps/web/src");
const apiRouteRoot = path.join(repositoryRoot, "apps/api/app/api");
const apiBusinessRoot = path.join(repositoryRoot, "apps/api/business");
const apiSchemaRoot = path.join(repositoryRoot, "apps/api/ent/schema");
const apiMigrationRoot = path.join(repositoryRoot, "apps/api/migrations");

// Inventory of VOC-010–VOC-024 mocks retained as P4-pending placeholders,
// reconciled at the VOC-028-T00 P3 boundary. See
// specs/changes/VOC-027-begin-milestone-p2-review-saved-words/mock-inventory.md
const expectedMocks = [];

// Inventory of VOC-010–VOC-024 mocks decommissioned to real sources.
const decommissionedMocks = [
  {
    file: "apps/web/src/app/(app)/discover/page.tsx",
    expectedConstant: "MOCK_DISCOVER_SITUATIONS",
    vocPackage: "VOC-021",
  },
  {
    file: "apps/web/src/app/(app)/discover/[situation]/page.tsx",
    expectedConstant: "MOCK_SITUATION_WORD_LISTS",
    vocPackage: "VOC-022",
  },
  {
    file: "apps/web/src/app/(app)/discover/[situation]/[word]/page.tsx",
    expectedConstant: "MOCK_SITUATION_WORD_LISTS",
    vocPackage: "VOC-022",
  },
  {
    file: "apps/web/src/app/(app)/discover/[situation]/_lib/mock-word-data.ts",
    expectedConstant: "MOCK_SITUATION_WORD_LISTS",
    vocPackage: "VOC-022",
    mustNotExist: true,
  },
  // VOC-030-T05: Home and Progress P4-pending mocks retired to real
  // `GET /api/v1/daily-mission` and `GET /api/v1/progress` reads.
  {
    file: "apps/web/src/app/(app)/home/page.tsx",
    expectedConstant: "MOCK_HOME_STATE",
    vocPackage: "VOC-019",
  },
  {
    file: "apps/web/src/app/(app)/progress/page.tsx",
    expectedConstant: "MOCK_PROGRESS_STATE",
    vocPackage: "VOC-020",
  },
];

const expectedRouteDirectories = [
  "home",
  "discover",
  path.join("discover", "[situation]"),
  path.join("discover", "[situation]", "[word]"),
  "progress",
  "reviews",
  // VOC-031-T05: the two new (app) routes P5 introduces
  // (Settings + Settings/account, per DOC-08). They live
  // inside the authenticated (app) group so the existing
  // auth/CSRF middleware matcher already covers them; the
  // expectedRouteDirectories presence check confirms the
  // route directories were actually created on disk by T05
  // and were not silently dropped by a future contributor.
  "settings",
  path.join("settings", "account"),
];

// VOC-031-T01 introduces the /onboarding route at the top
// level of apps/web/src/app/ (NOT inside the (app) group -
// per DOC-03 §3's flow, the gate redirects a learner to
// /onboarding before they reach any (app) route, so it
// must live outside the (app) group). The expectedRouteDirectories
// check above only covers (app) routes; the onboarding route
// gets its own presence check below.
const expectedTopLevelRouteDirectories = ["onboarding"];

export function validateMockInventory() {
  const errors = [];

  // Verify each retained mock file exists and still declares its expected constant.
  for (const mock of expectedMocks) {
    const filePath = path.join(repositoryRoot, mock.file);
    if (!exists(filePath)) {
      errors.push(`missing mock file: ${mock.file}`);
      continue;
    }
    const content = readFileSync(filePath, "utf8");
    if (!content.includes(mock.expectedConstant)) {
      errors.push(
        `${mock.file}: expected constant ${mock.expectedConstant} not found`,
      );
    }
    if (!content.includes(mock.vocPackage)) {
      errors.push(
        `${mock.file}: expected VOC package reference ${mock.vocPackage} not found`,
      );
    }
  }

  // VOC-027-D05: Home's dueReviewWords was decommissioned to the real P2 due-queue
  // count and must no longer be a hardcoded field in MOCK_HOME_STATE.
  const homePath = path.join(
    repositoryRoot,
    "apps/web/src/app/(app)/home/page.tsx",
  );
  if (exists(homePath)) {
    const homeContent = readFileSync(homePath, "utf8");
    const homeMockMatch = homeContent.match(
      /const\s+MOCK_HOME_STATE\s*=\s*\{[\s\S]*?\}\s*as\s+const/,
    );
    if (homeMockMatch && homeMockMatch[0].includes("dueReviewWords")) {
      errors.push(
        "apps/web/src/app/(app)/home/page.tsx: MOCK_HOME_STATE still contains the decommissioned dueReviewWords field",
      );
    }
    if (
      !/const\s+dueReviewWords\s*=\s*dueResponse\.data\.totalCount/.test(
        homeContent,
      )
    ) {
      errors.push(
        "apps/web/src/app/(app)/home/page.tsx: dueReviewWords is not wired to the real due-queue totalCount",
      );
    }
  }

  // Verify the decommissioned P1 mocks are gone from their original files and,
  // where marked, that the source file itself has been removed.
  for (const mock of decommissionedMocks) {
    const filePath = path.join(repositoryRoot, mock.file);
    const fileExists = exists(filePath);
    if (mock.mustNotExist && fileExists) {
      errors.push(
        `${mock.file}: decommissioned mock source file must be removed`,
      );
      continue;
    }
    if (fileExists) {
      const content = readFileSync(filePath, "utf8");
      if (content.includes(mock.expectedConstant)) {
        errors.push(
          `${mock.file}: decommissioned constant ${mock.expectedConstant} still present`,
        );
      }
    }
  }

  // Global scan: no decommissioned mock constant may appear anywhere in (app).
  const appFiles = globSync("**/*.{ts,tsx}", { cwd: appRouteRoot });
  const decommissionedConstants = new Set(
    decommissionedMocks.map((m) => m.expectedConstant),
  );
  for (const file of appFiles) {
    const content = readFileSync(path.join(appRouteRoot, file), "utf8");
    for (const constant of decommissionedConstants) {
      if (content.includes(constant)) {
        errors.push(
          `${file}: decommissioned constant ${constant} still present`,
        );
      }
    }
  }

  // Verify no API routes beyond A1 auth, VOC-026 P1 content/learning, the
  // VOC-027 P2 review routes (due-queue read and submission), the
  // VOC-028-T04 sentence-feedback write/report routes, the
  // VOC-030-T04 daily-mission/progress reads, the
  // VOC-031-T01 onboarding read/submit routes, the
  // VOC-031-T02 settings read/write routes, the
  // VOC-031-T03 email-change request/consume routes, and the
  // VOC-031-T04 account-deletion request route were invented.
  const allowedAPIPaths = [
    /^\/api\/v1\/me$/,
    /^\/api\/v1\/auth(?:\/|$)/,
    /^\/api\/v1\/journey-situations(?:\/[^/]+)?$/,
    /^\/api\/v1\/canonical-words(?:\/[^/]+)?$/,
    /^\/api\/v1\/user-words$/,
    /^\/api\/v1\/user-words\/[^/]+$/,
    /^\/api\/v1\/reviews\/due$/,
    /^\/api\/v1\/reviews\/submissions$/,
    /^\/api\/v1\/sentence-feedback$/,
    /^\/api\/v1\/sentence-feedback\/[^/]+\/reports$/,
    /^\/api\/v1\/daily-mission$/,
    /^\/api\/v1\/progress$/,
    /^\/api\/v1\/onboarding$/,
    /^\/api\/v1\/settings$/,
    /^\/api\/v1\/settings\/email-change-links$/,
    /^\/api\/v1\/settings\/email-change-links\/consume$/,
    /^\/api\/v1\/account-deletion-requests$/,
  ];
  const apiRouteFiles = globSync("**/*.go", { cwd: apiRouteRoot });
  for (const file of apiRouteFiles) {
    const content = readFileSync(path.join(apiRouteRoot, file), "utf8");
    const apiPaths = content.matchAll(/["'](\/api\/v1\/[^"'?\s]*)/g);
    for (const match of apiPaths) {
      const apiPath = match[1];
      if (!allowedAPIPaths.some((allowed) => allowed.test(apiPath))) {
        errors.push(
          `${file}: API path ${apiPath} outside the approved A1/P1/P2/P4-T00/T04/P5-T01/T02/T03 boundary`,
        );
      }
    }
  }

  // Verify no unexpected business modules. VOC-030-T00 introduces the
  // `missions` and `gamification` modules; T01-T03 will wire them into
  // the existing P1/P2/P3 transactions. VOC-031-T00 introduces the
  // `users` module; T03 adds the `accounts` module.
  const allowedBusinessModules = new Set([
    "auth",
    "content",
    "learning",
    "reviews",
    "aifeedback",
    "missions",
    "gamification",
    "users",
    "accounts",
  ]);
  for (const entry of readdirSync(apiBusinessRoot, {
    withFileTypes: true,
  })) {
    if (entry.isDirectory() && !allowedBusinessModules.has(entry.name)) {
      errors.push(
        `apps/api/business/${entry.name}: unexpected business module outside the approved A1/P1/P2/P4-T00/P5-T03 boundary`,
      );
    }
  }

  // Verify no unexpected Ent schemas. VOC-030-T00 introduces six new
  // tables: daily_mission_snapshots, daily_activity_summaries,
  // confidence_point_ledger, streak_states, grace_day_ledger, and
  // user_settings. VOC-031-T00 adds user_onboarding_profiles; T03
  // adds email_change_links; T04 adds account_deletion_requests.
  const allowedSchemaFiles = new Set([
    "canonicalword.go",
    "externalidentity.go",
    "journeysituation.go",
    "journeyword.go",
    "learnersentence.go",
    "magiclink.go",
    "mixins.go",
    "reviewattempt.go",
    "aifeedbackattempt.go",
    "session.go",
    "usagenote.go",
    "user.go",
    "userword.go",
    "wordexample.go",
    "wordmeaning.go",
    "dailymissionsnapshot.go",
    "dailyactivitysummary.go",
    "confidencepointledger.go",
    "streakstate.go",
    "gracedayledger.go",
    "usersettings.go",
    "useronboardingprofile.go",
    "emailchangelink.go",
    "accountdeletionrequest.go",
  ]);
  for (const entry of readdirSync(apiSchemaRoot, { withFileTypes: true })) {
    if (
      entry.isFile() &&
      entry.name.endsWith(".go") &&
      !allowedSchemaFiles.has(entry.name)
    ) {
      errors.push(
        `apps/api/ent/schema/${entry.name}: unexpected schema outside the approved A1/P1/P2/P4-T00/P5-T03 boundary`,
      );
    }
  }

  // Verify no unexpected migrations. VOC-030-T00 adds the
  // user_settings / mission_tables / gamification_tables migrations in
  // DOC-05 §18 order after ai_feedback_attempts. VOC-031-T00 adds the
  // user_onboarding_profiles migration in DOC-05 §18 order after the
  // P4 gamification_tables migration; T03 adds the
  // email_change_links migration; T04 adds the
  // account_deletion_requests migration.
  const allowedMigrationFiles = new Set([
    "20260724210000_identity_foundation.sql",
    "20260724210001_oauth_state.sql",
    "20260725100000_voc026_p1_content_tables.sql",
    "20260725100001_voc026_p1_idempotency_keys.sql",
    "20260725110000_voc027_p2_review_attempts.sql",
    "20260725120000_voc028_p3_learner_sentences.sql",
    "20260725120001_voc028_p3_ai_feedback_attempts.sql",
    "20260725130000_voc030_p4_user_settings.sql",
    "20260725130001_voc030_p4_mission_tables.sql",
    "20260725130002_voc030_p4_gamification_tables.sql",
    "20260725140000_voc031_p5_user_onboarding_profiles.sql",
    "20260725140001_voc031_p5_email_change_links.sql",
    "20260725140002_voc031_p5_account_deletion_requests.sql",
  ]);
  for (const entry of readdirSync(apiMigrationRoot, {
    withFileTypes: true,
  })) {
    if (
      entry.isFile() &&
      entry.name.endsWith(".sql") &&
      !allowedMigrationFiles.has(entry.name)
    ) {
      errors.push(
        `apps/api/migrations/${entry.name}: unexpected migration outside the approved A1/P1/P2/T00/P5-T03 boundary`,
      );
    }
  }

  // Verify the T05 evaluation, observability, and AI-disable/cost-ceiling
  // deliverables are present in the aifeedback module.
  const t05ExpectedFiles = ["evaluation.go", "metrics.go", "gate.go"];
  for (const file of t05ExpectedFiles) {
    const filePath = path.join(apiBusinessRoot, "aifeedback", file);
    if (!exists(filePath)) {
      errors.push(
        `apps/api/business/aifeedback/${file}: T05 expected file is missing`,
      );
    }
  }

  // Verify the middleware matcher covers all (app) routes and the
  // VOC-031-T01 onboarding gate.
  const middlewarePath = path.join(webSrcRoot, "middleware.ts");
  const middlewareContent = readFileSync(middlewarePath, "utf8");
  const matcherMatch = middlewareContent.match(/matcher:\s*\[([^\]]+)\]/s);
  if (!matcherMatch) {
    errors.push("apps/web/src/middleware.ts: could not parse matcher array");
  } else {
    const matcherText = matcherMatch[1];
    const requiredPatterns = [
      '"/home"',
      '"/discover"',
      '"/discover/:path*"',
      '"/progress"',
      '"/reviews"',
      '"/reviews/:path*"',
      '"/onboarding"',
      // VOC-031-T05: the two new Settings routes must be
      // covered by the auth/CSRF middleware matcher, or
      // an unauthenticated request would reach them
      // without the auth-gate redirect. The exact strings
      // mirror what the adopted T05 PR added to
      // apps/web/src/middleware.ts.
      '"/settings"',
      '"/settings/:path*"',
    ];
    for (const pattern of requiredPatterns) {
      if (!matcherText.includes(pattern)) {
        errors.push(
          `apps/web/src/middleware.ts: matcher is missing required pattern ${pattern}`,
        );
      }
    }
  }

  // Verify the expected route directories exist in the filesystem.
  for (const relative of expectedRouteDirectories) {
    const dirPath = path.join(appRouteRoot, relative);
    if (!exists(dirPath)) {
      errors.push(`missing (app) route directory: ${relative}`);
    }
  }

  // Verify the expected top-level (non-(app)) route directories
  // exist in the filesystem. The T01 onboarding route is the
  // only P5 route that lives outside the (app) group (per
  // DOC-03 §3's "redirect to /onboarding before any (app) route"
  // flow), so this list is intentionally short. A regression
  // that deleted the directory would silently re-open the
  // "no onboarding screen" gap.
  const topLevelAppRoot = path.join(repositoryRoot, "apps/web/src/app");
  for (const relative of expectedTopLevelRouteDirectories) {
    const dirPath = path.join(topLevelAppRoot, relative);
    if (!exists(dirPath)) {
      errors.push(
        `missing top-level route directory: ${relative} (expected at apps/web/src/app/${relative})`,
      );
    }
  }

  // Verify no new MOCK_ or placeholder data sources were introduced in (app).
  const knownMocks = new Set(expectedMocks.map((m) => m.expectedConstant));
  for (const file of appFiles) {
    const content = readFileSync(path.join(appRouteRoot, file), "utf8");
    const mockMatches = content.match(/\bMOCK_[A-Z_]+\b/g) ?? [];
    for (const mock of mockMatches) {
      if (!knownMocks.has(mock)) {
        errors.push(
          `${file}: unexpected mock constant ${mock}; add it to the VOC-027 inventory or replace with a real source`,
        );
      }
    }
  }

  // VOC-030-T05 field retirements: the retired MOCK_HOME_STATE
  // object and MOCK_PROGRESS_STATE object must be gone from the
  // (app) routes. T05 replaced both with real
  // GET /api/v1/daily-mission and GET /api/v1/progress reads; a
  // regression here would mean the web app is still reading from
  // the (deleted) mock state. We check by looking for the mock
  // object names, not for individual field names — the page may
  // legitimately use the real API field names (e.g.
  // reviewTarget, reviewsCompleted) as local aliases.
  for (const file of appFiles) {
    const content = readFileSync(path.join(appRouteRoot, file), "utf8");
    if (/\bMOCK_HOME_STATE\b/.test(content)) {
      errors.push(
        `${file}: MOCK_HOME_STATE is still present; T05 retired it in favor of GET /api/v1/daily-mission`,
      );
    }
    if (/\bMOCK_PROGRESS_STATE\b/.test(content)) {
      errors.push(
        `${file}: MOCK_PROGRESS_STATE is still present; T05 retired it in favor of GET /api/v1/progress`,
      );
    }
  }

  // VOC-030-T06: the P3 aifeedback test fixture must wire the
  // real missions.MissionUpdater (replacing the StubMissionUpdater
  // seam) — the T03 acceptance criterion. A regression that
  // removes the real-updater wiring would silently re-open the
  // always-false missionCompleted behavior in production.
  const aifeedbackP4TestPath = path.join(
    apiBusinessRoot,
    "aifeedback",
    "aifeedback_p4_test.go",
  );
  if (exists(aifeedbackP4TestPath)) {
    const testContent = readFileSync(aifeedbackP4TestPath, "utf8");
    if (!/missions\.NewMissionUpdater\s*\(/.test(testContent)) {
      errors.push(
        "apps/api/business/aifeedback/aifeedback_p4_test.go: T03 acceptance requires the test fixture to wire missions.NewMissionUpdater() (replacing NewStubMissionUpdater)",
      );
    }
  }

  // VOC-030-T06: the cross-cutting safety test files added in T06
  // must exist (their presence is part of the T06 acceptance
  // criteria for the duplicate/failed/unauthorized-safety
  // guarantees). A regression that removes either file would
  // silently re-open the cross-cutting safety holes.
  const t06TestFiles = [
    "apps/api/business/missions/cross_cutting_safety_test.go",
    "apps/api/app/api/missions_cross_cutting_test.go",
  ];
  for (const file of t06TestFiles) {
    const filePath = path.join(repositoryRoot, file);
    if (!exists(filePath)) {
      errors.push(
        `${file}: T06 cross-cutting safety test file is missing; T06 acceptance requires it`,
      );
    }
  }

  // VOC-031-T06: the P5 cross-cutting reliability test file
  // exists, alongside the P5 client-side session-expiry helper.
  // A regression that drops either would silently re-open the
  // session-expiry mid-flow gap and the no-fabricated-fallback
  // contract guard.
  const t06P5Deliverables = [
    "apps/api/app/api/core_loop_reliability_test.go",
    "apps/web/src/lib/session.ts",
  ];
  for (const file of t06P5Deliverables) {
    const filePath = path.join(repositoryRoot, file);
    if (!exists(filePath)) {
      errors.push(
        `${file}: T06 P5 cross-cutting reliability deliverable is missing; T06 acceptance requires it`,
      );
    }
  }

  // VOC-031-T06: the no-fabricated-fallback static guard.
  // Every (app) route's render code must consume data
  // exclusively through the API client. A regression that
  // hardcoded a data value, a default count, or a placeholder
  // string in place of real API data would let a learner
  // continue interacting with a screen that is showing
  // fabricated state — exactly the failure mode T06 forbids.
  //
  // The check is deliberately conservative: it scans (app)
  // for any non-comment data-object or array literal that
  // is not a known-empty-state literal (the empty array `[]`
  // is a legitimate render of "the API returned no items",
  // not a fabricated data value). A fabricated value would
  // look like `const SAVED_WORDS = [{ wordText: "..." }]` or
  // `const FALLBACK_STREAK = 5`. The regex below catches
  // the common patterns by name and shape; a regression
  // that introduced a fabrication under a different name
  // would be caught by the test plan's TEST-30 contract
  // check in apps/api/app/api/core_loop_reliability_test.go
  // (the API never returns a placeholder value), so the two
  // checks together cover the contract side and the client
  // side of the no-fabricated-fallback guarantee.
  const fabricatedDataPatterns = [
    /\bconst\s+MOCK_[A-Z_]+\s*=\s*\[/,
    /\bconst\s+MOCK_[A-Z_]+\s*=\s*\{/,
    /\bconst\s+FAKE_[A-Z_]+\s*=/,
    /\bconst\s+FALLBACK_[A-Z_]+\s*=/,
    /\bconst\s+PLACEHOLDER_[A-Z_]+\s*=/,
    /\bconst\s+DEMO_[A-Z_]+\s*=\s*\[/,
    /\bconst\s+DEMO_[A-Z_]+\s*=\s*\{/,
  ];
  for (const file of appFiles) {
    const content = readFileSync(path.join(appRouteRoot, file), "utf8");
    for (const pattern of fabricatedDataPatterns) {
      if (pattern.test(content)) {
        errors.push(
          `${file}: a hardcoded fabricated data literal matched ${pattern}; T06 forbids client-fabricated fallback values (use the real API instead)`,
        );
      }
    }
  }

  // VOC-030-T06: P5 is strictly forbidden. No P5 route, table, or
  // behavior may have been invented. The existing API path allow
  // list and schema/migration allow lists already enforce the
  // route/table side; this check enforces the behavior side on
  // the T06 deliverables (a future P5 feature must not be
  // introduced through this file).
  //
  // VOC-031-T06 extends the same behavior-allow-list to the
  // P5 cross-cutting reliability deliverables (the
  // core_loop_reliability_test.go file and the session.ts
  // helper) so a future contributor cannot introduce
  // out-of-scope P5 behavior through a T06 file either.
  const t06ForbidPatterns = [
    /p5[_-]?leaderboard/i,
    /p5[_-]?badge/i,
    /p5[_-]?reward[_-]?store/i,
  ];
  for (const file of [...t06TestFiles, ...t06P5Deliverables]) {
    const filePath = path.join(repositoryRoot, file);
    if (!exists(filePath)) continue;
    const content = readFileSync(filePath, "utf8");
    for (const pattern of t06ForbidPatterns) {
      if (pattern.test(content)) {
        errors.push(
          `${file}: P5 behavior detected (matched ${pattern}); P5 is strictly forbidden by the adopted D00 and must not be invented in T06`,
        );
      }
    }
  }

  // VOC-031-T07a: the accessibility-automation scaffolding must
  // remain present for the duration of the package. T07b
  // (multi-screen / multi-viewport coverage) and T08 (full
  // DOC-10 §7 core-loop E2E) both build on top of this
  // harness, so a regression that drops any of the
  // scaffolding files would silently re-open the
  // "no accessibility test in CI" gap the package exists to
  // close. The check intentionally stops at the scaffolding
  // surface (config + test directory + CI workflow) and does
  // not pin the specific test files inside tests/e2e/ - T07b
  // is free to add, rename, or remove individual specs.
  const t07aScaffolding = [
    "apps/web/playwright.config.ts",
    "apps/web/tests/e2e/README.md",
    "apps/web/tests/e2e/home-accessibility.spec.ts",
    "apps/web/tests/e2e/axe-helper.ts",
    "apps/web/tests/e2e/mock-api-server.mjs",
    ".github/workflows/accessibility.yml",
  ];
  for (const file of t07aScaffolding) {
    const filePath = path.join(repositoryRoot, file);
    if (!exists(filePath)) {
      errors.push(
        `${file}: T07a accessibility-automation scaffolding is missing; T07a acceptance requires the harness`,
      );
    }
  }

  // VOC-031-T09: the performance-automation harness (Lighthouse
  // CI, net-new in this repository per VOC-031-D00) must
  // remain present for the duration of the package. The
  // check is the same surface-level pattern the T07a
  // scaffolding check uses: directory + entry-point runner
  // + threshold source-of-truth + CI workflow. A regression
  // that drops any of these would silently re-open the
  // "no performance budget in CI" gap the package exists
  // to close - the DOC-08 quality-standards thresholds
  // (Performance 85+ / Accessibility 95+ / Best Practices
  // 90+) would go un-enforced.
  //
  // The check is intentionally NOT extended to pin
  // individual `apps/web/lighthouse-reports/*.json` files:
  // those are runtime artifacts, written per-audit, and
  // gitignored - a future T09 hardening is free to add,
  // remove, or rename them.
  const t09Scaffolding = [
    "apps/web/tests/lighthouse/README.md",
    "apps/web/tests/lighthouse/runner.mjs",
    "apps/web/tests/lighthouse/assertions.mjs",
    "apps/web/tests/lighthouse/budget.json",
    ".github/workflows/lighthouse.yml",
  ];
  for (const file of t09Scaffolding) {
    const filePath = path.join(repositoryRoot, file);
    if (!exists(filePath)) {
      errors.push(
        `${file}: T09 performance-automation scaffolding is missing; T09 acceptance requires the harness`,
      );
    }
  }

  // VOC-031-T09: the T09 runner and assertions are the T09
  // deliverables. The T06-forbid-patterns check is mirrored
  // here so a future contributor cannot introduce out-of-
  // scope P5 behavior (leaderboards, badges, rewards store)
  // through a T09 file either.
  const t09ForbidPatterns = [
    /p5[_-]?leaderboard/i,
    /p5[_-]?badge/i,
    /p5[_-]?reward[_-]?store/i,
  ];
  for (const file of t09Scaffolding) {
    const filePath = path.join(repositoryRoot, file);
    if (!exists(filePath)) continue;
    const content = readFileSync(filePath, "utf8");
    for (const pattern of t09ForbidPatterns) {
      if (pattern.test(content)) {
        errors.push(
          `${file}: P5 behavior detected (matched ${pattern}); P5 is strictly forbidden by the adopted D00 and must not be invented in T09`,
        );
      }
    }
  }

  // VOC-031-T11: the P5 mock-inventory completeness check.
  // T11 is the final task in the P5 ordered PR sequence
  // (T00→T11); its job is to confirm that the package's
  // documentation-level claims ("zero legacy mocks", "no
  // P5-invented route/table/behavior beyond this package's
  // own documented scope", "D06 appLanguage is restricted
  // to en-only") hold in the actual committed code. The
  // checks below are the static guard rails:
  //
  // 1. The settings appLanguage accepted-value set is
  //    exactly `["en"]` per VOC-031-D06. Any future
  //    contributor who silently widens the set to admit
  //    another language (without the i18n infrastructure
  //    D06 explicitly defers to a future package) would
  //    be presenting a UI affordance the product does not
  //    actually deliver. The check pins the value list
  //    via the same exported identifier the service-layer
  //    Validate() function reads, so the test, the
  //    service, and the doc are all in lockstep.
  // 2. The P5 staging-evidence document records the
  //    T06/T09/T10 audit findings (the T06 reliability
  //    pass, the T09 Lighthouse harness install, the T10
  //    UX-consistency audit). T11 cites them as
  //    in-repository evidence; a future contributor who
  //    silently truncated the audit findings would
  //    surface here.
  // 3. The mock-inventory check itself pins the package
  //    self-reference (T11 is a self-referential check
  //    by construction: T11 is the only task whose
  //    deliverable is the check itself).
  const settingsLanguagePath = path.join(
    apiBusinessRoot,
    "users",
    "settings.go",
  );
  if (exists(settingsLanguagePath)) {
    const settingsContent = readFileSync(settingsLanguagePath, "utf8");
    if (
      !/SupportedAppLanguages\s*=\s*\[\]string\{"en"\}/.test(settingsContent)
    ) {
      errors.push(
        'apps/api/business/users/settings.go: SupportedAppLanguages must be exactly []string{"en"} per VOC-031-D06 (no UI language picker before i18n infrastructure exists)',
      );
    }
  }
  const t11StagingEvidencePath = path.join(
    repositoryRoot,
    "specs/changes/VOC-031-begin-milestone-p5-integrated-core-loop/staging-evidence.md",
  );
  if (exists(t11StagingEvidencePath)) {
    const evidenceContent = readFileSync(t11StagingEvidencePath, "utf8");
    // T11 pins the existence (not the precise content) of the
    // T06/T09/T10 audit findings as the in-repository evidence
    // T11 cites. A future contributor who truncated them would
    // surface here.
    if (!/## `T06` audit findings/.test(evidenceContent)) {
      errors.push(
        "specs/changes/VOC-031-begin-milestone-p5-integrated-core-loop/staging-evidence.md: T06 audit findings section is missing; T11 evidence requires the cross-cutting reliability pass audit be present",
      );
    }
    if (!/## `T09` audit findings/.test(evidenceContent)) {
      errors.push(
        "specs/changes/VOC-031-begin-milestone-p5-integrated-core-loop/staging-evidence.md: T09 audit findings section is missing; T11 evidence requires the Lighthouse-CI install audit be present",
      );
    }
    if (!/## `T10` audit findings/.test(evidenceContent)) {
      errors.push(
        "specs/changes/VOC-031-begin-milestone-p5-integrated-core-loop/staging-evidence.md: T10 audit findings section is missing; T11 evidence requires the UX-consistency audit be present",
      );
    }
  }

  // VOC-031-T11: the package's own no-P5-invented-behavior
  // sweep, applied to T11's deliverable surface (this file
  // plus the staging-evidence.md / mock-inventory.md docs).
  // T11 is the final task; nothing in the P5 ordered PR
  // sequence depends on T11 producing behavior, so the
  // forbid-patterns list applies only to the T11 docs +
  // this script. A future contributor who added a leaderboard
  // or a rewards store as a "T11 follow-up" without going
  // through the change-control gate would surface here.
  const t11ForbidPatterns = [
    /p5[_-]?leaderboard/i,
    /p5[_-]?badge/i,
    /p5[_-]?reward[_-]?store/i,
  ];
  const t11DeliverableFiles = [
    "scripts/foundation/mock-inventory.mjs",
    "specs/changes/VOC-031-begin-milestone-p5-integrated-core-loop/mock-inventory.md",
    "specs/changes/VOC-031-begin-milestone-p5-integrated-core-loop/staging-evidence.md",
  ];
  for (const file of t11DeliverableFiles) {
    const filePath = path.join(repositoryRoot, file);
    if (!exists(filePath)) continue;
    const content = readFileSync(filePath, "utf8");
    for (const pattern of t11ForbidPatterns) {
      if (pattern.test(content)) {
        errors.push(
          `${file}: P5 behavior detected (matched ${pattern}); P5 is strictly forbidden by the adopted D00 and must not be invented in T11`,
        );
      }
    }
  }

  return errors;
}

function exists(filePath) {
  try {
    statSync(filePath);
    return true;
  } catch {
    return false;
  }
}

if (process.argv[1] === fileURLToPath(import.meta.url)) {
  const errors = validateMockInventory();
  if (errors.length) {
    process.stderr.write(
      `Mock inventory validation failed:\n${errors.map((e) => `  - ${e}`).join("\n")}\n`,
    );
    process.exitCode = 1;
  } else {
    process.stdout.write(
      "VOC-031-T11 mock inventory validation passed (T07a accessibility scaffolding + T06 cross-cutting reliability deliverables + T09 performance-automation scaffolding + T11 P5 completeness check present).\n",
    );
  }
}
