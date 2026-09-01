#!/usr/bin/env node
// Verify module boundaries between apps/web, apps/api, and the shared
// packages/* workspaces don't regress.
//
// Adapted from kdlbs/kandev's architecture-lint.yml *pattern* only — their
// scripts/lint-architecture.test.py encodes their own module graph, not a
// portable ruleset. This encodes vocanova's actual, already-implicit
// boundaries instead:
//
//   - packages/api-client must not import from apps/api or apps/web.
//   - packages/design-tokens must not import from apps/*.
//   - apps/web must not import apps/api files directly (only through the
//     @vocanova/api-client package).
//
// No apps/api (Go) rule: Go cannot import TypeScript files through any
// mechanism, so "apps/api must not import apps/web" is not a boundary that
// can actually be crossed - a naive text-reference check against it only
// produces false positives on legitimate cross-references (e.g. a test
// fixture listing apps/web/Dockerfile as an expected repo path).
//
// Scans the whole current tree rather than diffing against a merge-queue
// baseline (kandev's approach) - simpler to get right, and this repo is
// small enough that a full scan is cheap. Revisit if that changes.
//
// Usage: node scripts/foundation/architecture-boundaries.mjs [--root <path>]
// --root overrides the scanned repository root; used by
// architecture-boundaries.test.mjs to point at a disposable fixture tree.

import { readFileSync, readdirSync } from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";

const rootFlagIndex = process.argv.indexOf("--root");
const repositoryRoot =
  rootFlagIndex !== -1 && process.argv[rootFlagIndex + 1]
    ? path.resolve(process.argv[rootFlagIndex + 1])
    : path.resolve(path.dirname(fileURLToPath(import.meta.url)), "../..");

const IMPORT_RE = /(?:from\s+|import\s*\(|require\s*\()\s*["']([^"']+)["']/g;

function walk(dir, extensions) {
  const results = [];
  let entries;
  try {
    entries = readdirSync(dir, { withFileTypes: true });
  } catch {
    return results;
  }
  for (const entry of entries) {
    if (
      entry.name === "node_modules" ||
      entry.name === "dist" ||
      entry.name === ".next"
    ) {
      continue;
    }
    const full = path.join(dir, entry.name);
    if (entry.isDirectory()) {
      results.push(...walk(full, extensions));
    } else if (extensions.some((ext) => entry.name.endsWith(ext))) {
      results.push(full);
    }
  }
  return results;
}

function resolveImport(fromFile, specifier) {
  if (!specifier.startsWith(".")) {
    return null; // package import, not a relative path - not a boundary risk
  }
  return path.resolve(path.dirname(fromFile), specifier);
}

function checkTsBoundary({ label, scanRoot, forbiddenRoots }) {
  const violations = [];
  const files = walk(path.join(repositoryRoot, scanRoot), [
    ".ts",
    ".tsx",
    ".mjs",
    ".js",
  ]);
  for (const file of files) {
    if (file.endsWith(".test.ts") || file.endsWith(".test.tsx")) continue;
    const content = readFileSync(file, "utf8");
    for (const match of content.matchAll(IMPORT_RE)) {
      const resolved = resolveImport(file, match[1]);
      if (!resolved) continue;
      for (const forbiddenRoot of forbiddenRoots) {
        const absoluteForbidden = path.join(repositoryRoot, forbiddenRoot);
        if (
          resolved === absoluteForbidden ||
          resolved.startsWith(absoluteForbidden + path.sep)
        ) {
          violations.push(
            `${label}: ${path.relative(repositoryRoot, file)} imports ` +
              `"${match[1]}" which resolves into ${forbiddenRoot} (forbidden)`,
          );
        }
      }
    }
  }
  return violations;
}

const violations = [
  ...checkTsBoundary({
    label: "packages/api-client",
    scanRoot: "packages/api-client/src",
    forbiddenRoots: ["apps/api", "apps/web"],
  }),
  ...checkTsBoundary({
    label: "packages/design-tokens",
    scanRoot: "packages/design-tokens/src",
    forbiddenRoots: ["apps/api", "apps/web"],
  }),
  ...checkTsBoundary({
    label: "apps/web",
    scanRoot: "apps/web/src",
    forbiddenRoots: ["apps/api"],
  }),
];

if (violations.length > 0) {
  for (const violation of violations) {
    console.error(`::error::${violation}`);
  }
  console.error(
    `\n${violations.length} architecture boundary violation(s) found.`,
  );
  process.exit(1);
}

console.log("✓ No architecture boundary violations found.");
