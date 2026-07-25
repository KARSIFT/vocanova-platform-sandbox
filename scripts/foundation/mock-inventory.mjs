import { globSync, readFileSync, statSync } from "node:fs";
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
  {
    file: "apps/web/src/app/(app)/discover/page.tsx",
    expectedConstant: "MOCK_DISCOVER_SITUATIONS",
    vocPackage: "VOC-021",
  },
  {
    file: "apps/web/src/app/(app)/discover/[situation]/_lib/mock-word-data.ts",
    expectedConstant: "MOCK_SITUATION_WORD_LISTS",
    vocPackage: "VOC-022",
  },
];

const expectedMockConsumers = [
  {
    file: "apps/web/src/app/(app)/discover/[situation]/page.tsx",
    expectedImport: "MOCK_SITUATION_WORD_LISTS",
  },
  {
    file: "apps/web/src/app/(app)/discover/[situation]/[word]/page.tsx",
    expectedImport: "MOCK_SITUATION_WORD_LISTS",
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

  // Verify each mock file exists and still declares its expected constant.
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

  // Verify the two situation routes still import the shared mock data.
  for (const consumer of expectedMockConsumers) {
    const filePath = path.join(repositoryRoot, consumer.file);
    if (!exists(filePath)) {
      errors.push(`missing mock consumer file: ${consumer.file}`);
      continue;
    }
    const content = readFileSync(filePath, "utf8");
    if (!content.includes(consumer.expectedImport)) {
      errors.push(
        `${consumer.file}: expected import ${consumer.expectedImport} not found`,
      );
    }
  }

  // Verify no API routes beyond A1 auth and the VOC-026 T01 content reads were
  // invented. The P1 allowlist stays deliberately narrow here: user-word
  // writes and later-milestone endpoints remain forbidden until their tasks.
  const allowedAPIPaths = [
    /^\/api\/v1\/me$/,
    /^\/api\/v1\/auth(?:\/|$)/,
    /^\/api\/v1\/journey-situations(?:\/[^/]+)?$/,
    /^\/api\/v1\/canonical-words\/[^/]+$/,
  ];
  const apiRouteFiles = globSync("**/*.go", { cwd: apiRouteRoot });
  for (const file of apiRouteFiles) {
    const content = readFileSync(path.join(apiRouteRoot, file), "utf8");
    const apiPaths = content.matchAll(/["'](\/api\/v1\/[^"'?\s]*)/g);
    for (const match of apiPaths) {
      const apiPath = match[1];
      if (!allowedAPIPaths.some((allowed) => allowed.test(apiPath))) {
        errors.push(
          `${file}: contains API path ${apiPath} outside A1 and VOC-026-T01`,
        );
      }
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
  const appFiles = globSync("**/*.{ts,tsx}", { cwd: appRouteRoot });
  const knownMocks = new Set(
    expectedMocks
      .map((m) => m.expectedConstant)
      .concat("MOCK_SITUATION_WORD_LISTS"),
  );
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
    process.stdout.write("A1 mock inventory validation passed.\n");
  }
}
