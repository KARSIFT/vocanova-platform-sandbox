// Regression coverage for architecture-boundaries.mjs. Builds disposable
// fixture trees under a temp dir and runs the real script (via --root)
// against them, so this proves the checker actually fails closed rather
// than just asserting its source looks right.

import assert from "node:assert/strict";
import { mkdtempSync, mkdirSync, writeFileSync, rmSync } from "node:fs";
import { tmpdir } from "node:os";
import path from "node:path";
import { fileURLToPath } from "node:url";
import { spawnSync } from "node:child_process";
import test from "node:test";

const repositoryRoot = path.resolve(
  path.dirname(fileURLToPath(import.meta.url)),
  "../..",
);
const scriptPath = path.join(
  repositoryRoot,
  "scripts/foundation/architecture-boundaries.mjs",
);

function runAgainst(fixtureRoot) {
  return spawnSync(process.execPath, [scriptPath, "--root", fixtureRoot], {
    encoding: "utf8",
  });
}

function writeFile(root, relativePath, content) {
  const fullPath = path.join(root, relativePath);
  mkdirSync(path.dirname(fullPath), { recursive: true });
  writeFileSync(fullPath, content);
}

test("architecture-boundaries: passes on a clean tree", () => {
  const root = mkdtempSync(path.join(tmpdir(), "arch-boundaries-clean-"));
  try {
    writeFile(
      root,
      "packages/api-client/src/index.ts",
      `import { z } from "zod";\nexport const ok = z.object({});\n`,
    );
    writeFile(root, "apps/web/src/lib/thing.ts", `export const thing = 1;\n`);
    const result = runAgainst(root);
    assert.equal(result.status, 0, result.stderr);
    assert.match(result.stdout, /No architecture boundary violations/);
  } finally {
    rmSync(root, { recursive: true, force: true });
  }
});

test("architecture-boundaries: fails closed on packages/api-client importing apps/web", () => {
  const root = mkdtempSync(path.join(tmpdir(), "arch-boundaries-violation-"));
  try {
    writeFile(
      root,
      "packages/api-client/src/index.ts",
      `import { thing } from "../../../apps/web/src/lib/thing";\nexport { thing };\n`,
    );
    writeFile(root, "apps/web/src/lib/thing.ts", `export const thing = 1;\n`);
    const result = runAgainst(root);
    assert.equal(result.status, 1);
    assert.match(result.stderr, /packages\/api-client.*apps\/web/s);
  } finally {
    rmSync(root, { recursive: true, force: true });
  }
});

test("architecture-boundaries: fails closed on apps/web importing apps/api", () => {
  const root = mkdtempSync(path.join(tmpdir(), "arch-boundaries-web-api-"));
  try {
    writeFile(
      root,
      "apps/web/src/lib/thing.ts",
      `import { handler } from "../../../api/some/handler";\nexport { handler };\n`,
    );
    writeFile(root, "apps/api/some/handler.ts", `export const handler = 1;\n`);
    const result = runAgainst(root);
    assert.equal(result.status, 1);
    assert.match(result.stderr, /apps\/web.*apps\/api/s);
  } finally {
    rmSync(root, { recursive: true, force: true });
  }
});

test("architecture-boundaries: ignores package-name imports (not relative paths)", () => {
  const root = mkdtempSync(path.join(tmpdir(), "arch-boundaries-pkg-import-"));
  try {
    writeFile(
      root,
      "packages/design-tokens/src/index.ts",
      `import { thing } from "@vocanova/web-internal";\nexport { thing };\n`,
    );
    const result = runAgainst(root);
    assert.equal(result.status, 0, result.stderr);
  } finally {
    rmSync(root, { recursive: true, force: true });
  }
});
