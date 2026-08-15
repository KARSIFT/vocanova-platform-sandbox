// VOC-081-T01 — dedicated monitoring network and topology invariants.
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

const monitoringComposePath = path.join(
  repositoryRoot,
  "infra/docker-compose.monitoring.yml",
);
const sharedEdgeComposePath = path.join(
  repositoryRoot,
  "infra/docker-compose.shared-edge.yml",
);
const productionComposePath = path.join(
  repositoryRoot,
  "infra/docker-compose.production.yml",
);
const stagingComposePath = path.join(
  repositoryRoot,
  "infra/docker-compose.yml",
);
const deployProductionPath = path.join(
  repositoryRoot,
  ".github/workflows/deploy-production.yml",
);
const deployStagingPath = path.join(
  repositoryRoot,
  ".github/workflows/deploy-staging.yml",
);

const MONITORING_NETWORK = "vocanova-monitoring-net";
const HOST_HTTPS_PORT_PATTERN = /["']443:443["']/;
const HOST_HTTP_PORT_PATTERN = /["']80:80["']/;
const BRIDGE_PUBLISH_PATTERNS = [/["']8081:80["']/, /["']8443:443["']/];
const PUBLIC_KUMA_PORT_PATTERNS = [
  /^\s*-\s*["']0\.0\.0\.0:3001:3001["']/m,
  /^\s*-\s*["']\[::\]:3001:3001["']/m,
  /^\s*-\s*["']3001:3001["']/m,
];

function serviceBlock(compose, serviceName) {
  const match = compose.match(
    new RegExp(
      `^\\s{2}${serviceName}:\\s*$\\n([\\s\\S]*?)(?=^\\s{2}\\w|\\n\\w)`,
      "m",
    ),
  );
  assert.ok(match, `${serviceName} service must exist in compose file`);
  return match[1];
}

test("VOC-081-TEST-01: monitoring and shared-edge declare vocanova-monitoring-net", () => {
  const monitoring = readFileSync(monitoringComposePath, "utf8");
  const sharedEdge = readFileSync(sharedEdgeComposePath, "utf8");

  assert.match(
    monitoring,
    new RegExp(`name:\\s*${MONITORING_NETWORK}`),
    "monitoring compose must declare external vocanova-monitoring-net",
  );
  assert.match(
    sharedEdge,
    new RegExp(`name:\\s*${MONITORING_NETWORK}`),
    "shared-edge compose must declare external vocanova-monitoring-net",
  );

  const kumaBlock = serviceBlock(monitoring, "uptime-kuma");
  assert.match(
    kumaBlock,
    new RegExp(`-\\s*${MONITORING_NETWORK}`),
    "uptime-kuma must attach to vocanova-monitoring-net",
  );

  const nginxBlock = serviceBlock(sharedEdge, "nginx");
  assert.match(
    nginxBlock,
    new RegExp(`-\\s*${MONITORING_NETWORK}`),
    "shared-edge nginx must attach to vocanova-monitoring-net",
  );
});

test("VOC-081-TEST-01: Kuma is not on staging/production app networks", () => {
  const monitoring = readFileSync(monitoringComposePath, "utf8");
  const kumaBlock = serviceBlock(monitoring, "uptime-kuma");

  assert.doesNotMatch(
    kumaBlock,
    /-\s*vocanova-net/,
    "uptime-kuma must not attach to vocanova-net",
  );
  assert.doesNotMatch(
    kumaBlock,
    /-\s*vocanova-production-net/,
    "uptime-kuma must not attach to vocanova-production-net",
  );

  const networksSection =
    monitoring.match(/^networks:\s*$\n([\s\S]*)/m)?.[1] ?? "";
  assert.doesNotMatch(
    networksSection,
    /^\s{2}vocanova-net:/m,
    "monitoring compose must not declare vocanova-net",
  );
  assert.doesNotMatch(
    networksSection,
    /^\s{2}vocanova-production-net:/m,
    "monitoring compose must not declare vocanova-production-net",
  );

  assert.doesNotMatch(
    monitoring,
    /\/opt\/vocanova\/(infra|production)\/secrets/,
    "monitoring compose must not mount staging/production secret trees",
  );
  assert.doesNotMatch(
    monitoring,
    /secrets\/(api|postgres)\.env/,
    "monitoring compose must not reference tier env_file secrets",
  );
});

test("VOC-081-TEST-02: only shared-edge publishes host 80/443", () => {
  const sharedEdge = readFileSync(sharedEdgeComposePath, "utf8");
  const monitoring = readFileSync(monitoringComposePath, "utf8");
  const production = readFileSync(productionComposePath, "utf8");
  const staging = readFileSync(stagingComposePath, "utf8");

  assert.match(
    sharedEdge,
    HOST_HTTP_PORT_PATTERN,
    "shared-edge compose must publish host port 80",
  );
  assert.match(
    sharedEdge,
    HOST_HTTPS_PORT_PATTERN,
    "shared-edge compose must publish host port 443",
  );

  for (const [label, source] of [
    ["monitoring", monitoring],
    ["production", production],
    ["staging", staging],
  ]) {
    assert.doesNotMatch(
      source,
      HOST_HTTP_PORT_PATTERN,
      `${label} compose must not publish host port 80`,
    );
    assert.doesNotMatch(
      source,
      HOST_HTTPS_PORT_PATTERN,
      `${label} compose must not publish host port 443`,
    );
    for (const pattern of BRIDGE_PUBLISH_PATTERNS) {
      assert.doesNotMatch(
        source,
        pattern,
        `${label} compose must not publish bridge port mapping ${pattern}`,
      );
    }
  }
});

test("VOC-081-TEST-02: Kuma has no public port 3001 publish", () => {
  const monitoring = readFileSync(monitoringComposePath, "utf8");

  assert.match(
    monitoring,
    /127\.0\.0\.1:3001:3001/,
    "monitoring compose must keep loopback-only 3001 publish for local ops",
  );

  for (const pattern of PUBLIC_KUMA_PORT_PATTERNS) {
    assert.doesNotMatch(
      monitoring,
      pattern,
      `monitoring compose must not publish public Kuma port ${pattern}`,
    );
  }
});

test("VOC-081-TEST-02: Kuma healthcheck invokes the JavaScript probe through Node", () => {
  const monitoring = readFileSync(monitoringComposePath, "utf8");

  assert.match(
    monitoring,
    /test:\s*\["CMD",\s*"node",\s*"extra\/healthcheck\.js"\]/,
    "Kuma 1.x healthcheck script has no shebang and must be invoked through node",
  );
});

test("VOC-081-TEST-01: production deploy does not own the monitoring project", () => {
  const deployProduction = readFileSync(deployProductionPath, "utf8");

  assert.doesNotMatch(
    deployProduction,
    /docker-compose\.monitoring\.yml/,
    "deploy-production must not converge monitoring compose (staging-owned)",
  );
  assert.doesNotMatch(
    deployProduction,
    /-p\s+monitoring\b/,
    "deploy-production must not use compose project name monitoring",
  );
});

test("VOC-081-TEST-01: staging creates the external network before shared-edge bring-up", () => {
  const deployStaging = readFileSync(deployStagingPath, "utf8");
  const inspectIndex = deployStaging.indexOf(
    "docker network inspect vocanova-monitoring-net",
  );
  const createIndex = deployStaging.indexOf(
    "docker network create --driver bridge vocanova-monitoring-net",
  );
  const composeUpIndex = deployStaging.indexOf(
    "docker compose -f docker-compose.shared-edge.yml -p vocanova-shared-edge up -d",
  );

  assert.ok(
    inspectIndex >= 0,
    "staging deploy must inspect the monitoring network",
  );
  assert.ok(
    createIndex > inspectIndex,
    "network creation must follow its idempotent inspect guard",
  );
  assert.ok(
    composeUpIndex > createIndex,
    "monitoring network must exist before shared-edge Compose bring-up",
  );
});
