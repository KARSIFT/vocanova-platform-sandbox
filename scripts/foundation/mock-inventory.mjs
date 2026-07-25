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

// Inventory of VOC-010–VOC-024 mocks retained as P4-pending placeholders.
// See specs/changes/VOC-026-begin-milestone-p1-discover-and-save-words-per-doc/mock-inventory.md
const expectedMocks = [
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

// Inventory of VOC-010–VOC-024 mocks decommissioned to real P1 sources.
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
];

const expectedRouteDirectories = [
  "home",
  "discover",
  path.join("discover", "[situation]"),
  path.join("discover", "[situation]", "[word]"),
  "progress",
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

  // Verify no API routes beyond A1 auth and VOC-026 P1 content/learning were invented.
  const allowedAPIPaths = [
    /^\/api\/v1\/me$/,
    /^\/api\/v1\/auth(?:\/|$)/,
    /^\/api\/v1\/journey-situations(?:\/[^/]+)?$/,
    /^\/api\/v1\/canonical-words\/[^/]+$/,
    /^\/api\/v1\/user-words$/,
    /^\/api\/v1\/user-words\/[^/]+$/,
  ];
  const apiRouteFiles = globSync("**/*.go", { cwd: apiRouteRoot });
  for (const file of apiRouteFiles) {
    const content = readFileSync(path.join(apiRouteRoot, file), "utf8");
    const apiPaths = content.matchAll(/["'](\/api\/v1\/[^"'?\s]*)/g);
    for (const match of apiPaths) {
      const apiPath = match[1];
      if (!allowedAPIPaths.some((allowed) => allowed.test(apiPath))) {
        errors.push(
          `${file}: contains API path ${apiPath} outside A1 and VOC-026 P1`,
        );
      }
    }
  }

  // Verify no unexpected business modules (i.e. no invented P2–P4 behavior).
  const allowedBusinessModules = new Set(["auth", "content", "learning"]);
  for (const entry of readdirSync(apiBusinessRoot, {
    withFileTypes: true,
  })) {
    if (entry.isDirectory() && !allowedBusinessModules.has(entry.name)) {
      errors.push(
        `apps/api/business/${entry.name}: unexpected business module outside A1 and VOC-026 P1`,
      );
    }
  }

  // Verify no unexpected Ent schemas (i.e. no invented P2–P4 tables).
  const allowedSchemaFiles = new Set([
    "canonicalword.go",
    "externalidentity.go",
    "journeysituation.go",
    "journeyword.go",
    "magiclink.go",
    "mixins.go",
    "session.go",
    "usagenote.go",
    "user.go",
    "userword.go",
    "wordexample.go",
    "wordmeaning.go",
  ]);
  for (const entry of readdirSync(apiSchemaRoot, { withFileTypes: true })) {
    if (
      entry.isFile() &&
      entry.name.endsWith(".go") &&
      !allowedSchemaFiles.has(entry.name)
    ) {
      errors.push(
        `apps/api/ent/schema/${entry.name}: unexpected schema outside A1 and VOC-026 P1`,
      );
    }
  }

  // Verify no unexpected migrations (i.e. no invented P2–P4 tables).
  const allowedMigrationFiles = new Set([
    "20260724210000_identity_foundation.sql",
    "20260724210001_oauth_state.sql",
    "20260725100000_voc026_p1_content_tables.sql",
    "20260725100001_voc026_p1_idempotency_keys.sql",
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
        `apps/api/migrations/${entry.name}: unexpected migration outside A1 and VOC-026 P1`,
      );
    }
  }

  // Verify the middleware matcher covers all (app) routes.
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
          `${file}: unexpected mock constant ${mock}; add it to the T05 inventory or replace with a real source`,
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
    process.stdout.write("VOC-026 P1 mock inventory validation passed.\n");
  }
}
