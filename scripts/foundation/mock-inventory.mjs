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
];

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
  // VOC-031-T01 onboarding read/submit routes, and the
  // VOC-031-T02 settings read/write routes were invented.
  const allowedAPIPaths = [
    /^\/api\/v1\/me$/,
    /^\/api\/v1\/auth(?:\/|$)/,
    /^\/api\/v1\/journey-situations(?:\/[^/]+)?$/,
    /^\/api\/v1\/canonical-words\/[^/]+$/,
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
  ];
  const apiRouteFiles = globSync("**/*.go", { cwd: apiRouteRoot });
  for (const file of apiRouteFiles) {
    const content = readFileSync(path.join(apiRouteRoot, file), "utf8");
    const apiPaths = content.matchAll(/["'](\/api\/v1\/[^"'?\s]*)/g);
    for (const match of apiPaths) {
      const apiPath = match[1];
      if (!allowedAPIPaths.some((allowed) => allowed.test(apiPath))) {
        errors.push(
          `${file}: API path ${apiPath} outside the approved A1/P1/P2/P4-T00/T04/P5-T01/T02 boundary`,
        );
      }
    }
  }

  // Verify no unexpected business modules. VOC-030-T00 introduces the
  // `missions` and `gamification` modules; T01-T03 will wire them into
  // the existing P1/P2/P3 transactions. VOC-031-T00 introduces the
  // `users` module; T03 will add the `accounts` module.
  const allowedBusinessModules = new Set([
    "auth",
    "content",
    "learning",
    "reviews",
    "aifeedback",
    "missions",
    "gamification",
    "users",
  ]);
  for (const entry of readdirSync(apiBusinessRoot, {
    withFileTypes: true,
  })) {
    if (entry.isDirectory() && !allowedBusinessModules.has(entry.name)) {
      errors.push(
        `apps/api/business/${entry.name}: unexpected business module outside the approved A1/P1/P2/P4-T00 boundary`,
      );
    }
  }

  // Verify no unexpected Ent schemas. VOC-030-T00 introduces six new
  // tables: daily_mission_snapshots, daily_activity_summaries,
  // confidence_point_ledger, streak_states, grace_day_ledger, and
  // user_settings. VOC-031-T00 adds user_onboarding_profiles; T03/T04
  // will add email_change_links and account_deletion_requests.
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
  ]);
  for (const entry of readdirSync(apiSchemaRoot, { withFileTypes: true })) {
    if (
      entry.isFile() &&
      entry.name.endsWith(".go") &&
      !allowedSchemaFiles.has(entry.name)
    ) {
      errors.push(
        `apps/api/ent/schema/${entry.name}: unexpected schema outside the approved A1/P1/P2/P4-T00 boundary`,
      );
    }
  }

  // Verify no unexpected migrations. VOC-030-T00 adds the
  // user_settings / mission_tables / gamification_tables migrations in
  // DOC-05 §18 order after ai_feedback_attempts. VOC-031-T00 adds the
  // user_onboarding_profiles migration in DOC-05 §18 order after the
  // P4 gamification_tables migration; T03/T04 will add the
  // email_change_links and account_deletion_requests migrations.
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
        `apps/api/migrations/${entry.name}: unexpected migration outside the approved A1/P1/P2/T00 boundary`,
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

  // VOC-030-T06: P5 is strictly forbidden. No P5 route, table, or
  // behavior may have been invented. The existing API path allow
  // list and schema/migration allow lists already enforce the
  // route/table side; this check enforces the behavior side on
  // the T06 deliverables (a future P5 feature must not be
  // introduced through this file).
  const t06ForbidPatterns = [
    /p5[_-]?leaderboard/i,
    /p5[_-]?badge/i,
    /p5[_-]?reward[_-]?store/i,
  ];
  for (const file of t06TestFiles) {
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
    process.stdout.write("VOC-030-T06 mock inventory validation passed.\n");
  }
}
