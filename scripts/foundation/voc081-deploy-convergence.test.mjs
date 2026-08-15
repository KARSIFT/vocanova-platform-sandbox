// VOC-081-T03 — deploy workflow convergence for monitoring + shared-edge.
//
// Runs via `pnpm test` → `node --test scripts/foundation/*.test.mjs`.

import assert from "node:assert/strict";
import { readFileSync } from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";
import test from "node:test";

const repositoryRoot = path.resolve(
  path.dirname(fileURLToPath(import.meta.url)),
  "../..",
);

const deployStagingPath = path.join(
  repositoryRoot,
  ".github/workflows/deploy-staging.yml",
);
const deployProductionPath = path.join(
  repositoryRoot,
  ".github/workflows/deploy-production.yml",
);

const MANUAL_DOCKER_PATTERNS = [
  /docker network connect\b/,
  /ssh\b[^\n]*30-monitor\.conf/,
  /\bdocker rm\b[^\n]*vocanova-shared-edge-nginx/,
];

test("VOC-081-TEST-05: staging deploy converges monitoring from repository source", () => {
  const deployStaging = readFileSync(deployStagingPath, "utf8");

  assert.match(
    deployStaging,
    /cp infra\/docker-compose\.monitoring\.yml/,
    "staging bundle must ship monitoring compose",
  );
  assert.match(
    deployStaging,
    /docker compose -f \/opt\/vocanova\/monitoring\/docker-compose\.monitoring\.yml -p monitoring up -d/,
    "staging deploy must converge the monitoring compose project",
  );
  assert.match(
    deployStaging,
    /docker network inspect vocanova-monitoring-net/,
    "staging deploy must ensure the monitoring network exists",
  );
  assert.match(
    deployStaging,
    /docker network create --driver bridge vocanova-monitoring-net/,
    "staging deploy must idempotently create the monitoring network",
  );

  const networkCreateIndex = deployStaging.indexOf(
    "docker network create --driver bridge vocanova-monitoring-net",
  );
  const monitoringUpIndex = deployStaging.indexOf(
    "docker compose -f /opt/vocanova/monitoring/docker-compose.monitoring.yml -p monitoring up -d",
  );
  const sharedEdgeUpIndex = deployStaging.indexOf(
    "docker compose -f docker-compose.shared-edge.yml -p vocanova-shared-edge up -d",
  );

  assert.ok(
    monitoringUpIndex > networkCreateIndex,
    "monitoring compose must converge after the monitoring network is ensured",
  );
  assert.ok(
    monitoringUpIndex < sharedEdgeUpIndex || sharedEdgeUpIndex < 0,
    "monitoring compose must converge before shared-edge compose bring-up when both are present",
  );

  for (const pattern of MANUAL_DOCKER_PATTERNS) {
    assert.doesNotMatch(
      deployStaging,
      pattern,
      `staging deploy must not rely on manual operator step ${pattern}`,
    );
  }
});

test("VOC-081-TEST-05: first monitoring converge backs up existing data and waits for health", () => {
  const deployStaging = readFileSync(deployStagingPath, "utf8");
  const backupIndex = deployStaging.indexOf(
    'sudo tar -C "$MONITORING_ROOT" -czf "$MONITORING_BACKUP" kuma-data',
  );
  const backupValidationIndex = deployStaging.indexOf(
    'sudo test -s "$MONITORING_BACKUP"',
  );
  const monitoringUpIndex = deployStaging.indexOf(
    "docker compose -f /opt/vocanova/monitoring/docker-compose.monitoring.yml -p monitoring up -d",
  );
  const markerIndex = deployStaging.indexOf('sudo touch "$MONITORING_MARKER"');

  assert.match(
    deployStaging,
    /\.repository-converged-voc081/,
    "first-converge backup must be guarded by a durable marker",
  );
  assert.match(
    deployStaging,
    /docker stop --time 30 vocanova-uptime-kuma/,
    "Kuma must be stopped for a consistent first-converge data backup",
  );
  assert.ok(backupIndex >= 0, "deploy must create a Kuma data backup");
  assert.ok(
    backupValidationIndex > backupIndex,
    "deploy must validate the backup after creating it",
  );
  assert.ok(
    monitoringUpIndex > backupValidationIndex,
    "backup must be validated before monitoring Compose convergence",
  );
  assert.ok(
    markerIndex > monitoringUpIndex,
    "first-converge marker must only be written after Compose convergence",
  );
  assert.match(
    deployStaging,
    /kuma_status[\s\S]*vocanova-uptime-kuma[\s\S]*healthy/,
    "deploy must wait for Kuma health before continuing",
  );
  assert.match(
    deployStaging,
    /chmod 0600 "\$MONITORING_BACKUP"/,
    "backup containing monitoring state must be permission-restricted",
  );
});

test("VOC-081-TEST-05: fail-closed nginx -t precedes shared-edge reload or convergence", () => {
  const deployStaging = readFileSync(deployStagingPath, "utf8");

  assert.match(
    deployStaging,
    /nginx:1\.27-alpine nginx -t/,
    "staging deploy must run disposable nginx -t against the shared-edge mount set",
  );
  assert.match(
    deployStaging,
    /docker exec "\$SHARED_EDGE_CONTAINER" nginx -t/,
    "staging deploy must run fail-closed nginx -t inside the running shared-edge container",
  );

  const disposableTestIndex = deployStaging.indexOf("run_shared_edge_nginx_t");
  const routineReloadIndex = deployStaging.indexOf(
    'docker exec "$SHARED_EDGE_CONTAINER" nginx -s reload',
  );

  assert.ok(
    disposableTestIndex >= 0 && routineReloadIndex > disposableTestIndex,
    "disposable nginx -t must precede shared-edge reload on the routine path",
  );
});

test("VOC-081-TEST-05: production deploy retires stale monitor conf without owning monitoring", () => {
  const deployProduction = readFileSync(deployProductionPath, "utf8");

  assert.match(
    deployProduction,
    /rm -f nginx\/conf\.d\/30-monitor\.conf/,
    "production deploy must remove stale unloaded 30-monitor.conf",
  );
  assert.doesNotMatch(
    deployProduction,
    /docker-compose\.monitoring\.yml/,
    "production deploy must not converge monitoring compose",
  );

  for (const pattern of MANUAL_DOCKER_PATTERNS) {
    assert.doesNotMatch(
      deployProduction,
      pattern,
      `production deploy must not rely on manual operator step ${pattern}`,
    );
  }
});

test("VOC-081-TEST-06: scoped compose ownership and no routine shared-edge force-recreate", () => {
  const deployStaging = readFileSync(deployStagingPath, "utf8");
  const deployProduction = readFileSync(deployProductionPath, "utf8");

  assert.doesNotMatch(
    deployStaging,
    /docker compose[^\n]*--force-recreate/,
    "staging deploy must not force-recreate containers on routine deploys",
  );
  assert.doesNotMatch(
    deployStaging,
    /docker compose -f docker-compose\.yml[^\n]*--remove-orphans/,
    "staging app deploy must not orphan-remove across the staging project",
  );
  assert.doesNotMatch(
    deployProduction,
    /docker compose -f docker-compose\.shared-edge\.yml/,
    "production deploy must not recreate the shared-edge compose project",
  );
  assert.match(
    deployProduction,
    /docker compose -f docker-compose\.production\.yml -p vocanova-production up -d --remove-orphans/,
    "production orphan removal must remain scoped to vocanova-production only",
  );
  assert.doesNotMatch(
    deployStaging,
    /-p\s+monitoring\b[\s\S]*--remove-orphans/,
    "monitoring converge must not use --remove-orphans",
  );
});
