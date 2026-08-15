// VOC-085-T00 — production deploy P1 content seed bundling, ordering, fail-closed.
//
// Runs via `pnpm test` → `node --test scripts/foundation/*.test.mjs`.

import assert from "node:assert/strict";
import { execFileSync } from "node:child_process";
import { mkdtempSync, readFileSync, writeFileSync } from "node:fs";
import { tmpdir } from "node:os";
import path from "node:path";
import { fileURLToPath } from "node:url";
import test from "node:test";

const repositoryRoot = path.resolve(
  path.dirname(fileURLToPath(import.meta.url)),
  "../..",
);

const deployProductionPath = path.join(
  repositoryRoot,
  ".github/workflows/deploy-production.yml",
);
const deployStagingPath = path.join(
  repositoryRoot,
  ".github/workflows/deploy-staging.yml",
);
const seedMainPath = path.join(repositoryRoot, "apps/api/cmd/seed/main.go");
const seedMainTestPath = path.join(
  repositoryRoot,
  "apps/api/cmd/seed/main_test.go",
);
const syntheticSeedSqlPath = path.join(
  repositoryRoot,
  "apps/api/scripts/seed-synthetic-smoke-user.sql",
);

const deployProduction = readFileSync(deployProductionPath, "utf8");
const deployStaging = readFileSync(deployStagingPath, "utf8");

function deployHostScriptBlock(workflowYaml) {
  const marker = "name: Deploy to production host";
  const start = workflowYaml.indexOf(marker);
  assert.ok(start >= 0, "missing Deploy to production host step");
  const scriptStart = workflowYaml.indexOf("script: |", start);
  assert.ok(scriptStart >= 0, "missing deploy host script block");
  const scriptBodyStart = workflowYaml.indexOf("\n", scriptStart) + 1;
  const scriptEnd = workflowYaml.indexOf(
    "\n      # VOC-079-T01",
    scriptBodyStart,
  );
  assert.ok(scriptEnd > scriptBodyStart, "could not bound deploy host script");
  return workflowYaml.slice(scriptBodyStart, scriptEnd);
}

test("VOC-085-TEST-00: production bundle builds p1-content-seed with pinned Go toolchain", () => {
  assert.match(
    deployProduction,
    /name: Set up Go for the P1 content seed build/,
    "production deploy must set up Go before bundling the seed binary",
  );
  assert.match(
    deployProduction,
    /uses: actions\/setup-go@b7ad1dad31e06c5925ef5d2fc7ad053ef454303e/,
    "production deploy must SHA-pin setup-go like staging",
  );
  assert.match(
    deployProduction,
    /go-version-file:\s*["']apps\/api\/go\.mod["']/,
    "production deploy must pin Go from apps/api/go.mod",
  );
  assert.match(
    deployProduction,
    /mkdir -p \/tmp\/production-deploy-bundle\/apps\/api\/bin/,
    "production bundle must include apps/api/bin",
  );
  assert.match(
    deployProduction,
    /CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build[\s\S]*-C apps\/api[\s\S]*-o \/tmp\/production-deploy-bundle\/apps\/api\/bin\/p1-content-seed[\s\S]*\.\/cmd\/seed/,
    "production bundle must build p1-content-seed from apps/api/cmd/seed",
  );
  assert.match(
    deployStaging,
    /-o \/tmp\/deploy-bundle\/apps\/api\/bin\/p1-content-seed/,
    "staging bundle must still ship p1-content-seed (no regression)",
  );
});

test("VOC-085-TEST-01: P1 seed runs after migrations and synthetic-user seed, before up -d", () => {
  const hostScript = deployHostScriptBlock(deployProduction);

  const migrateIndex = hostScript.indexOf(
    "sh /opt/vocanova/production/apps/api/scripts/migrate.sh",
  );
  const syntheticIndex = hostScript.indexOf(
    "sh /opt/vocanova/production/apps/api/scripts/seed-synthetic-smoke-user.sh",
  );
  const p1SeedIndex = hostScript.indexOf(
    "/opt/vocanova/production/apps/api/bin/p1-content-seed",
  );
  const upIndex = hostScript.indexOf(
    "docker compose -f docker-compose.production.yml -p vocanova-production up -d --remove-orphans",
  );

  assert.ok(migrateIndex >= 0, "deploy host script must run migrate.sh");
  assert.ok(
    syntheticIndex >= 0,
    "deploy host script must run synthetic-user seed",
  );
  assert.ok(p1SeedIndex >= 0, "deploy host script must run p1-content-seed");
  assert.ok(upIndex >= 0, "deploy host script must run docker compose up -d");

  assert.ok(
    migrateIndex < syntheticIndex,
    "migrations must precede synthetic-user seed",
  );
  assert.ok(
    syntheticIndex < p1SeedIndex,
    "synthetic-user seed must precede P1 content seed",
  );
  assert.ok(
    p1SeedIndex < upIndex,
    "P1 content seed must precede application convergence (up -d)",
  );

  const p1SeedLine = hostScript
    .split("\n")
    .find((line) => line.includes("p1-content-seed"));
  assert.ok(p1SeedLine, "missing p1-content-seed invocation line");
  assert.match(
    p1SeedLine,
    /DATABASE_URL="\$migration_database_url"/,
    "P1 seed must reuse the private Postgres bridge DATABASE_URL",
  );
});

test("VOC-085-TEST-02: P1 seed failure aborts before application convergence", () => {
  const hostScript = deployHostScriptBlock(deployProduction);

  assert.match(
    hostScript,
    /set -euo pipefail/,
    "deploy host script must use fail-closed shell options",
  );
  assert.doesNotMatch(
    deployProduction,
    /continue-on-error:\s*true[\s\S]{0,200}p1-content-seed/,
    "P1 content seed must not use continue-on-error",
  );

  const tempDir = mkdtempSync(path.join(tmpdir(), "voc085-fail-closed-"));
  const failingSeed = path.join(tempDir, "p1-content-seed");
  const upMarker = path.join(tempDir, "up-ran");
  writeFileSync(failingSeed, "#!/bin/sh\nexit 1\n", { mode: 0o755 });

  const disposableScript = `
set -euo pipefail
migration_database_url=postgres://stub
sh -c 'exit 0'
sh -c 'exit 0'
DATABASE_URL="$migration_database_url" "${failingSeed}"
touch "${upMarker}"
docker compose up -d --remove-orphans
`;

  let exitCode = 0;
  try {
    execFileSync("bash", ["-c", disposableScript], {
      encoding: "utf8",
      stdio: "pipe",
    });
  } catch (error) {
    exitCode = error.status ?? 1;
  }

  assert.notEqual(
    exitCode,
    0,
    "failing P1 seed must abort the disposable deploy script",
  );
  assert.throws(
    () => readFileSync(upMarker, "utf8"),
    "application convergence must not run after a failing P1 seed under set -e",
  );
});

test("VOC-085-TEST-03: canonical seed remains idempotent upsert-only", () => {
  const seedMain = readFileSync(seedMainPath, "utf8");
  const seedTests = readFileSync(seedMainTestPath, "utf8");

  assert.match(
    seedMain,
    /ON CONFLICT/,
    "seed tool must upsert by fixed keys (ON CONFLICT)",
  );
  assert.doesNotMatch(
    seedMain,
    /\bDELETE FROM\b/i,
    "seed tool must not delete user-owned learning state",
  );
  assert.match(
    seedTests,
    /TestApplySeedExecutesUpsertStatementsInOrder/,
    "existing seed tests must cover upsert statement execution",
  );
  assert.match(
    seedTests,
    /require\.Len\(t, seed\.JourneySituations, 7\)/,
    "embedded canonical dataset must remain the seven P1 situations",
  );
});

test("VOC-085-TEST-08: synthetic account onboarding remains repository-seed deterministic", () => {
  const syntheticSql = readFileSync(syntheticSeedSqlPath, "utf8");

  assert.match(
    syntheticSql,
    /onboarding_status\s*=\s*'completed'/,
    "synthetic-user seed must set onboarding_status completed",
  );
  assert.match(
    syntheticSql,
    /is_synthetic_test_account\s*=\s*true/,
    "synthetic-user seed must mark the reserved account synthetic",
  );
  assert.doesNotMatch(
    deployProduction,
    /psql[\s\S]{0,120}onboarding_status/,
    "production deploy must not manually edit onboarding in live DB",
  );
});
